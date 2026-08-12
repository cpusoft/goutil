package websubutil

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cpusoft/goutil/websubutil/handler"
	"github.com/cpusoft/goutil/websubutil/model"
	"github.com/cpusoft/goutil/websubutil/store"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

// Validator is a function to validate a subscription request.
// If error is not nil, hub.mode=verify will be called with the error.
type Validator func(model.Subscription) error

// ContentProvider is a function to extract content out of the specific content topic.
type ContentProvider func(topic string) ([]byte, string, error)

// Option represents a Hub option.
type Option func(h *Hub)

// Hub represents a WebSub hub.
type Hub struct {
	*handler.Handler

	client          *http.Client
	store           store.Store
	validator       Validator
	contentProvider ContentProvider
	worker          Worker
	hasher          string
	url             string
	maxLease        time.Duration
}

var (
	v = validator.New()
)

// WithValidator sets the subscription validator.
func WithValidator(validator Validator) Option {
	return func(h *Hub) {
		h.validator = validator
	}
}

// WithContentProvider sets the content provider for external hub.mode=publish requests.
func WithContentProvider(provider ContentProvider) Option {
	return func(h *Hub) {
		h.contentProvider = provider
	}
}

// WithHasher lets you set other hmac hashers/types (like sha256, sha384, sha512, etc)
func WithHasher(hasher string) Option {
	switch hasher {
	case "sha1", "sha256", "sha384", "sha512":
	default:
		hasher = "sha256"
	}
	return func(h *Hub) {
		h.hasher = hasher
	}
}

// WithWorker lets you set the worker used to distribute subscription responses.
// This can be done with any number of systems, such as Amazon SQS, Beanstalk, etc.
func WithWorker(worker Worker) Option {
	return func(h *Hub) {
		h.worker = worker
	}
}

// WithURL lets you set the hub url.
// By default, this is auto detected on first request for ease of use.
func WithURL(url string) Option {
	return func(h *Hub) {
		h.url = url
	}
}

// WithMaxLease lets you set the hub's max lease time.
// By default, this is 24 hours.
func WithMaxLease(maxLease time.Duration) Option {
	return func(h *Hub) {
		h.maxLease = maxLease
	}
}

// New creates a new WebSub Hub instance.
// store is required to store all of the subscriptions.
func New(store store.Store, opts ...Option) *Hub {
	h := &Hub{
		Handler: handler.New(),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		store:           store,
		contentProvider: HttpContent,
		hasher:          "sha256",
		maxLease:        24 * time.Hour,
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.worker == nil {
		h.worker = NewGoWorker(h, runtime.NumCPU())
		h.worker.Start()
	}

	return h
}

// ServeHTTP is a generic webserver handler for websub.
// It takes in "hub.mode" from the form, and passes it to the appropriate handlers.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hubMode := r.FormValue("hub.mode")

	if hubMode == "" {
		http.Error(w, "missing hub.mode parameter", http.StatusBadRequest)
		return
	}

	// If url is not set, set to something we can "guess"
	if h.url == "" {
		proto := "http"

		// Usually X-Forwarded cannot be trusted, but in this case it's the first request that defines it.
		// For our case, this simply sets the hub url via "auto detection".
		// it is STRONGLY advised to set the url using WithURL beforehand.
		if r.Header.Get("X-Forwarded-Proto") == "https" {
			proto = r.Header.Get("X-Forwarded-Proto")
		}

		u := &url.URL{
			Scheme: proto,
			Host:   r.Host,
			Path:   r.URL.Path, // 仅取路径，丢弃查询参数 r.RequestURI,
		}

		h.url = strings.TrimRight(u.String(), "/")
	}

	switch hubMode {
	case model.ModeSubscribe:
		var req model.SubscribeRequest

		if err := DecodeForm(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := h.HandleSubscribe(req)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	case model.ModeUnsubscribe:
		var req model.UnsubscribeRequest

		if err := DecodeForm(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := h.HandleUnsubscribe(req)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	case model.ModePublish:
		var req model.PublishRequest

		if err := DecodeForm(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := h.HandlePublish(req)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	default:
		http.Error(w, "hub.mode not recognized", http.StatusBadRequest)
	}
}

// HandleSubscribe handles a hub.mode=subscribe request.
func (h *Hub) HandleSubscribe(req model.SubscribeRequest) error {
	// validate for required fields
	if err := v.Struct(req); err != nil {
		return err
	}

	// Default lease
	leaseDuration := h.maxLease //240 * time.Hour

	if req.LeaseSeconds > 0 {
		if req.LeaseSeconds < 60 || time.Duration(req.LeaseSeconds)*time.Second > h.maxLease {
			return errors.New("invalid hub.lease_seconds value")
		} else {
			leaseDuration = time.Duration(req.LeaseSeconds) * time.Second
		}
	}

	sub := model.Subscription{
		Topic:     req.Topic,
		Callback:  req.Callback,
		Secret:    req.Secret,
		Expires:   time.Now().Add(leaseDuration),
		LeaseTime: leaseDuration, // 设置 LeaseTime
	}

	if h.validator != nil {
		err := h.validator(sub)

		if err != nil {
			sub.Reason = err

			return h.Verify(model.ModeDenied, sub)
		}
	}

	existingSub, err := h.store.Get(req.Topic, req.Callback)

	if existingSub != nil && err == nil {
		// Update existingSub instead.
		// TODO: Can Secret be updated?
		sub = *existingSub
		sub.LeaseTime = leaseDuration // 续订更新时也同步更新 LeaseTime
		sub.Expires = time.Now().Add(leaseDuration)
		if req.Secret != "" {
			sub.Secret = req.Secret // 允许更新 Secret
		}
	}

	go func(hubMode string, sub model.Subscription) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in subscribe verification: %v", r)
			}
		}()

		err := h.Verify(hubMode, sub)

		if err != nil {
			h.Call(&VerificationFailed{
				Subscription: sub,
				Error:        err,
			})
		} else {
			h.Call(&Verified{
				Subscription: sub,
			})
		}
	}(req.Mode, sub)

	return nil
}

// HandleUnsubscribe handles a hub.mode=unsubscribe
func (h *Hub) HandleUnsubscribe(req model.UnsubscribeRequest) error {
	// validate for required fields
	if err := v.Struct(req); err != nil {
		return err
	}

	sub := model.Subscription{
		Topic:    req.Topic,
		Callback: req.Callback,
	}

	if h.validator != nil {
		err := h.validator(sub)

		if err != nil {
			sub.Reason = err

			return h.Verify(model.ModeDenied, sub)
		}
	}

	go func(hubMode string, sub model.Subscription) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in unsubscribe verification: %v", r)
			}
		}()
		err := h.Verify(hubMode, sub)

		if err != nil {
			log.Println("Error:", err)
		}
	}(req.Mode, sub)

	return nil
}

// Verify sends a response to a subscription model with the specified data.
// If the subscription failed, Reason can be set to send hub.reason in the callback.
func (h *Hub) Verify(mode string, sub model.Subscription) error {
	u, err := url.Parse(sub.Callback)

	if err != nil {
		return err
	}

	challenge := uuid.New().String()

	q := u.Query()
	q.Set("hub.mode", mode)
	q.Set("hub.topic", sub.Topic)

	if mode != model.ModeDenied {
		q.Set("hub.challenge", challenge)
		q.Set("hub.lease_seconds", strconv.Itoa(int(sub.LeaseTime/time.Second)))
	} else if sub.Reason != nil {
		q.Set("hub.reason", sub.Reason.Error())
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)

	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Go WebSub 1.0 ("+runtime.Version()+")")

	res, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	//if res.StatusCode != 200 {
	if res.StatusCode < 200 || res.StatusCode > 299 {
		// Uh oh!
		return errors.New("unexpected status code")
	}

	defer res.Body.Close()

	if mode == model.ModeDenied {
		io.Copy(io.Discard, res.Body)
		return nil
	}

	// Read max of challenge size bytes
	data := make([]byte, len(challenge))

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if string(body) != challenge {
		// Nope.
		return errors.New(fmt.Sprint("verification: challenge did not match for "+u.Host+", expected: ", challenge, " actual: ", string(data)))
	}

	if mode == model.ModeSubscribe {
		// Update the subscription and set it as verified
		// time.Now().Add(time.Duration(leaseSeconds) * time.Second), topic, callback
		err = h.store.Add(sub)
	} else if mode == model.ModeUnsubscribe {
		// Delete the subscription
		err = h.store.Remove(sub)
	}

	return err
}

// HandlePublish handles a request to publish from a publisher.
func (h *Hub) HandlePublish(req model.PublishRequest) error {
	if err := v.Struct(req); err != nil {
		return err
	}

	data, contentType, err := h.contentProvider(req.Topic)

	if err != nil {
		return err
	}

	return h.Publish(req.Topic, contentType, data)
}

// Publish queues responses to the worker for a publish.
func (h *Hub) Publish(topic, contentType string, data []byte) error {
	subs, err := h.store.All(topic)

	if err != nil {
		if err == store.ErrNotFound {
			subs = []model.Subscription{} // 无订阅者视为正常空状态
		} else {
			return err // 仅真实错误（如数据库连接失败）才向上返回
		}
	}

	h.Call(&Publish{
		Topic:       topic,
		ContentType: contentType,
		Data:        data,
	})

	hub := model.Hub{
		Hasher: h.hasher,
		URL:    h.url,
	}

	for _, sub := range subs {
		h.worker.Add(PublishJob{
			Hub:          hub,
			Subscription: sub,
			ContentType:  contentType,
			Data:         data,
		})
	}

	return nil
}

// NewHasher takes a string and returns a hash.Hash based on type.
func NewHasher(hasher string) func() hash.Hash {
	switch hasher {
	case "sha1":
		return sha1.New
	case "sha256":
		return sha256.New
	case "sha384":
		return sha512.New384
	case "sha512":
		return sha512.New
	default:
		return sha256.New
	}

}

// DecodeForm decodes a request form into a struct using the mapstructure package.
func DecodeForm(r *http.Request, dest interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "form",
		Result:  dest,
		// This hook is a trick to allow us to map from []string -> string in the case of elements.
		// This is only required because we're mapping from r.Form -> struct.
		DecodeHook: func(from reflect.Kind, to reflect.Kind, v interface{}) (interface{}, error) {
			if from == reflect.Slice && (to == reflect.String || to == reflect.Int) {
				switch s := v.(type) {
				case []string:
					if len(s) < 1 {
						return "", nil
					}

					// Switch statement seems wasteful here, but if we want to add uint/etc we can easily.
					switch to {
					case reflect.Int:
						return strconv.Atoi(s[0])
					}

					return s[0], nil
				}

				return v, nil
			}

			return v, nil
		},
	})

	if err != nil {
		return err
	}

	return decoder.Decode(r.Form)
}

package websubutil

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/websubutil/model"
)

// PublishJob represents a job to publish data to a subscription.
type PublishJob struct {
	Hub          model.Hub          `json:"hub"`
	Subscription model.Subscription `json:"subscription"`
	ContentType  string             `json:"contentType"`
	Data         []byte             `json:"data"`
}

// Worker is an interface to allow other types of workers to be created.
type Worker interface {
	Add(f PublishJob)
	Start()
	Stop()
}

// NewGoWorker creates a new worker from the specified hub and worker count.
func NewGoWorker(h *Hub, workerCount int) *GoWorker {
	return &GoWorker{
		hub:         h,
		workerCount: workerCount,
		jobCh:       make(chan PublishJob, 1000),
	}
}

// GoWorker is a basic Goroutine-based worker.
// It will start workerCount workers and process jobs from a channel.
type GoWorker struct {
	hub         *Hub
	workerCount int
	jobCh       chan PublishJob
}

// Add will add a job to the queue.
func (w *GoWorker) Add(job PublishJob) {
	belogs.Debug("GoWorker.Add(): adding job:", job)
	w.jobCh <- job
}

// Start will start the worker routines.
func (w *GoWorker) Start() {
	belogs.Debug("GoWorker.Start(): starting workers, workerCount:", w.workerCount)
	for i := 0; i < w.workerCount; i++ {
		go w.run()
	}
}

// Stop will close the job channel, causing each worker routine to exit.
func (w *GoWorker) Stop() {
	close(w.jobCh)
	belogs.Debug("GoWorker.Stop(): jobCh closed")
}

// run pulls jobs off the job channel and processes them.
func (w *GoWorker) run() {
	for {
		job, ok := <-w.jobCh

		if !ok {
			belogs.Debug("GoWorker.run(): jobCh closed, exiting")
			return
		}

		publishResponse, err := Notify(w.hub.client, job)

		// TODO: Log errors
		if err != nil {
			belogs.Error("GoWorker.run(): Notify fail, publishResponse:", jsonutil.MarshalJson(publishResponse), "job:", job, err)
			w.hub.store.PublishResult(job.Subscription, job.Data, publishResponse)
			continue
		}
		belogs.Debug("GoWorker.run(): Notify ok, publishResponse:", jsonutil.MarshalJson(publishResponse), "job:", job)
		w.hub.store.PublishResult(job.Subscription, job.Data, publishResponse)

		// Remove failed subscriptions
		//if !sent {
		//	w.hub.store.Remove(job.Subscription)
		//}
	}
}

package badgedb

import (
	"errors"
	"github.com/goccy/go-json"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/dgraph-io/badger/v4"
)

func (b *BadgeDBImpl) Get(key string) ([]byte, bool, error) {
	var result []byte
	var exists bool

	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // Key not found, but not considered an error
			}
			return err
		}

		exists = true

		result, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return result, exists, err
	}

	return result, exists, nil
}

func (b *BadgeDBImpl) GetWithTxn(txn *badger.Txn, key string) ([]byte, bool, error) {
	var result []byte
	var exists bool

	item, err := txn.Get([]byte(key))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return result, exists, nil // Key not found, but not considered an error
		}
		return result, exists, err
	}

	exists = true

	result, err = item.ValueCopy(nil)
	if err != nil {
		return result, exists, err
	}

	return result, exists, nil
}

func (b *BadgeDBImpl) MGet(keys []string) (map[string][]byte, map[string]error, error) {
	results := make(map[string][]byte)
	resultsErrors := make(map[string]error)

	err := b.db.View(func(txn *badger.Txn) error {
		for _, key := range keys {
			var result []byte
			item, err := txn.Get([]byte(key))
			if err != nil {
				if errors.Is(err, badger.ErrKeyNotFound) {
					resultsErrors[key] = badger.ErrKeyNotFound
				} else {
					resultsErrors[key] = err
				}
				continue
			}
			result, convErr := item.ValueCopy(nil)
			if convErr != nil {
				resultsErrors[key] = convErr
				continue
			}
			results[key] = result
		}
		return nil
	})

	if err != nil {
		return results, resultsErrors, err
	}

	return results, resultsErrors, nil
}

func (b *BadgeDBImpl) MGetWithTxn(txn *badger.Txn, keys []string) (map[string][]byte, map[string]error, error) {
	results := make(map[string][]byte)
	resultsErrors := make(map[string]error)

	for _, key := range keys {
		var result []byte
		item, err := txn.Get([]byte(key))

		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				resultsErrors[key] = badger.ErrKeyNotFound
			} else {
				resultsErrors[key] = err
			}
			continue
		}

		result, err = item.ValueCopy(nil)
		if err != nil {
			resultsErrors[key] = err
			continue
		}

		results[key] = result
	}

	return results, resultsErrors, nil
}

// Insert generic function to insert key-value pairs into the database
func (b *BadgeDBImpl) Insert(key string, value any) error {
	// Convert value to byte slice
	valueBytes, err := marshalValue(value)
	if err != nil {
		return err
	}

	err = b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), valueBytes)
	})

	if err != nil {
		log.Fatal(err)
	}
	return err
}

func (b *BadgeDBImpl) InsertWithTxn(txn *badger.Txn, key string, value any) error {
	// Convert value to byte slice
	valueBytes, err := marshalValue(value)
	if err != nil {
		return err
	}
	err = txn.Set([]byte(key), valueBytes)

	if err != nil {
		log.Fatal(err)
	}
	return err
}
func (b *BadgeDBImpl) Delete(key string) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	if err != nil {
		log.Fatal(err)
	}
	return err
}

func (b *BadgeDBImpl) DeleteWithTxn(txn *badger.Txn, key string) error {
	err := txn.Delete([]byte(key))
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func (b *BadgeDBImpl) BatchInsert(data map[string]any) map[string]error {
	if len(data) <= 0 {
		return nil
	}
	errs := make(map[string]error)
	wb := b.db.NewWriteBatch() // Create a batch write object
	defer wb.Cancel()          // Cancel or commit on error or completion

	for key, value := range data {
		// Convert value to byte slice
		valueBytes, err := marshalValue(value)
		if err != nil {
			errs[key] = err
			belogs.Error("BatchInsert, marshalValue failed, but continue, key:", key)
			continue
		}
		err = wb.Set([]byte(key), valueBytes) // Add key-value to be written
		if err != nil {
			errs[key] = err
			belogs.Error("BatchInsert, Set failed, but continue, key:", key)
			continue
		}
	}

	// Commit batch write
	err := wb.Flush()
	if err != nil {
		for key := range data {
			errs[key] = err // Mark all keys that were not successfully written
		}
		return errs
	}

	return nil
}

// Transaction guarantee

func (b *BadgeDBImpl) BatchInsertWithTxn(txn *badger.Txn, data map[string]any) error {
	const defaultBatchSize = 100

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys); i += defaultBatchSize {
		for j := i; j < i+defaultBatchSize && j < len(keys); j++ {
			key := keys[j]
			valueBytes, err := marshalValue(data[key])
			if err != nil {
				belogs.Error("marshalValue failed, err", err)
				return err
			}
			if err := txn.Set([]byte(key), valueBytes); err != nil {
				belogs.Error("Set failed, err", err)
				return err
			}

		}
	}

	return nil
}

func (b *BadgeDBImpl) QueryByCompositeKey(entity string, columns map[string]string) ([]byte, bool, error) {
	compositeKey := b.buildCompositeKey(columns)
	result, exists, err := b.Get(compositeKey)
	if err != nil {
		return result, false, err
	}
	if !exists {
		return result, false, nil
	}
	return result, true, nil
}

func (b *BadgeDBImpl) QueryByCompositeKeyWithTxn(txn *badger.Txn, entity string, columns map[string]string) ([]byte, bool, error) {
	var result []byte
	compositeKey := b.buildCompositeKey(columns)

	result, exists, err := b.GetWithTxn(txn, compositeKey)
	if err != nil {
		return result, false, err
	}
	if !exists {
		return result, false, nil
	}
	return result, true, nil
}

func (b *BadgeDBImpl) BatchQueryByPrefix(prefix string) (map[string][]byte, map[string]error, error) {
	results := make(map[string][]byte)
	errMaps := make(map[string]error)

	err := b.db.View(func(txn *badger.Txn) error {
		// Set iterator options, query by prefix
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		// Iterate and query
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			key := string(item.Key())

			// Get value and parse to T
			val, err := item.ValueCopy(nil)
			if err != nil {
				errMaps[key] = err
				continue
			}

			// Store in result map
			results[key] = val
		}

		return nil
	})

	if err != nil {
		return results, errMaps, err
	}

	return results, errMaps, nil
}

func (b *BadgeDBImpl) BatchQueryByPrefixWithTxn(txn *badger.Txn, prefix string) (map[string][]byte, map[string]error, error) {
	results := make(map[string][]byte)
	errMaps := make(map[string]error)

	// Set iterator options, query by prefix
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = true
	it := txn.NewIterator(opts)
	defer it.Close()

	// Iterate and query
	for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
		item := it.Item()
		key := string(item.Key())

		// Get value and parse to T
		val, err := item.ValueCopy(nil)
		if err != nil {
			errMaps[key] = err
			continue
		}

		// Store in result map
		results[key] = val
	}
	return results, errMaps, nil
}

func (b *BadgeDBImpl) buildBaseKey(columns ...string) string {
	sort.Strings(columns)
	return strings.Join(columns, ":")
}

func (b *BadgeDBImpl) buildCompositeKey(columns map[string]string) string {
	var parts []string

	keys := make([]string, 0, len(columns))
	for k := range columns {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		parts = append(parts, key+":"+columns[key])
	}

	return strings.Join(parts, ":")
}

func (b *BadgeDBImpl) convertToType(data []byte) (any, error) {
	strData := string(data)

	// 处理简单类型
	if isNumeric(strData) {
		if i, err := strconv.ParseInt(strData, 10, 64); err == nil {
			return i, nil
		} else if f, err := strconv.ParseFloat(strData, 64); err == nil {
			return f, nil
		}
	} else if b, err := strconv.ParseBool(strData); err == nil {
		return b, nil
	}

	// 如果是 JSON 数据
	if json.Valid(data) {
		var result any
		err := json.Unmarshal(data, &result)
		if err != nil {
			return result, err
		}
		return result, nil
	}

	// 返回字符串数据
	return strData, nil
}

// 判断是否是数字
func isNumeric(data string) bool {
	_, err := strconv.ParseFloat(data, 64)
	return err == nil
}

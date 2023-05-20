package global

import (
	"crypto/md5"
	"encoding/hex"
	"sync"
	"time"
)

type HashEntry struct {
	Hash      string
	Value     string
	Timestamp time.Time
}

type HashStorage struct {
	sync.RWMutex
	Data       map[string]HashEntry
	Expiration time.Duration
	ForceSave  bool
}

func NewHashStorage(expiration time.Duration, forceSave bool) *HashStorage {
	return &HashStorage{
		Data:       make(map[string]HashEntry),
		Expiration: expiration,
		ForceSave:  forceSave,
	}
}

func (h *HashStorage) AddHash(query string) string {
	if h.ForceSave {
		return h.saveValue(query)
	}

	// we only need to hash for more than 64 chars by default
	if len(query) <= 64 {
		return query
	}

	return h.saveValue(query)
}

func (h *HashStorage) saveValue(query string) string {
	h.Lock()
	defer h.Unlock()

	md5Hash := md5.Sum([]byte(query))
	md5HashString := hex.EncodeToString(md5Hash[:])

	entry := HashEntry{
		Hash:      md5HashString,
		Value:     query,
		Timestamp: time.Now(),
	}

	h.Data[md5HashString] = entry

	return md5HashString
}

func (h *HashStorage) GetValue(hash string) string {
	h.RLock()
	defer h.RUnlock()

	entry, exists := h.Data[hash]
	if !exists {
		return hash
	}
	return entry.Value
}

func (h *HashStorage) RemoveExpiredHashes() {
	h.Lock()
	defer h.Unlock()

	now := time.Now()

	for hash, entry := range h.Data {
		if now.Sub(entry.Timestamp) > h.Expiration {
			delete(h.Data, hash)
		}
	}
}

func (h *HashStorage) Reset() {
	h.Lock()
	defer h.Unlock()

	h.Data = make(map[string]HashEntry)
}

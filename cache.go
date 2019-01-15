package rewledis

import (
	"sync"
	"time"
)

type CacheEntryState int8

// Constants which a CacheEntryState value can assume.
const (
	CacheEntryStateNone CacheEntryState = iota
	CacheEntryStateLoading
	CacheEntryStateExists
	CacheEntryStateDeleted
	CacheEntryStateError
)

type cacheEntry struct {
	RWMutex     sync.RWMutex
	WrittenAt   time.Time
	DoneLoading <-chan struct{}
	State       CacheEntryState
	Type        LedisType
}

type CacheEntryData struct {
	entry       *cacheEntry
	WrittenAt   time.Time
	Key         string
	DoneLoading <-chan struct{}
	State       CacheEntryState
	Type        LedisType
}

func (c *CacheEntryData) Refresh() {
	c.entry.RWMutex.RLock()

	c.copyFrom(c.entry)

	c.entry.RWMutex.RUnlock()
}

func (c *CacheEntryData) copyFrom(entry *cacheEntry) {
	c.WrittenAt = c.entry.WrittenAt
	c.DoneLoading = c.entry.DoneLoading
	c.State = c.entry.State
	c.Type = c.entry.Type
}

type CacheEntrySetter struct {
	entry       *cacheEntry
	Key         string
	doneLoading chan struct{}
}

// Set sets the Type field of an entry along with the State and WrittenAt
// fields. It also signals completion of the loading process by closing the
// DoneLoading channel. state must be either CacheEntryStateExists,
// CacheEntryStateDeleted or CacheEntryStateError and indicates the success
// state of the type resolution. If state is not CacheEntryStateExists the
// keyType value is ignored.
func (c *CacheEntrySetter) Set(state CacheEntryState, keyType LedisType) {
	if c.doneLoading == nil {
		// Panic now. Otherwise close() on the channel would panic, this gives
		// a more useful message.
		panic("CacheEntrySetter: Set() called a second time")
	}

	switch state {
	case CacheEntryStateExists:
	case CacheEntryStateDeleted:
		keyType = LedisTypeNone
	case CacheEntryStateError:
		keyType = LedisTypeNone
	default:
		panic("CacheEntryState: Set() called with invalid state parameter")
	}

	c.entry.RWMutex.Lock()

	c.entry.WrittenAt = time.Now()
	c.entry.DoneLoading = nil
	c.entry.State = state
	c.entry.Type = keyType

	c.entry.RWMutex.Unlock()

	close(c.doneLoading)

	c.doneLoading = nil
}

type Cache struct {
	entries sync.Map
}

func (c *Cache) LoadType(key string) (LedisType, bool) {
	entry, ok := c.loadEntry(key)
	if !ok {
		return LedisTypeNone, false
	}

	entry.RWMutex.RLock()

	state := entry.State
	keyType := entry.Type

	entry.RWMutex.RUnlock()

	if state != CacheEntryStateExists {
		return LedisTypeNone, false
	}

	return keyType, true
}

// TrySetEntry tries to update the entry for the given key with given state
// and keyType information. TrySetEntry does not wait for an ongoing loading
// procedure and returns without setting in this case. TrySetEntry returns
// true, if the entry was updated and false if not.
func (c *Cache) TrySetEntry(key string, state CacheEntryState, keyType LedisType) bool {
	switch state {
	case CacheEntryStateExists:
	case CacheEntryStateDeleted:
		keyType = LedisTypeNone
	case CacheEntryStateError:
		keyType = LedisTypeNone
	default:
		panic("Cache: TrySetEntry() called with invalid state parameter")
	}

	entry, ok := c.loadEntry(key)
	if ok {
		return c.trySetEntry(entry, state, keyType)
	}

	entry = &cacheEntry{
		DoneLoading: nil,
		State:       state,
		Type:        keyType,
		WrittenAt:   time.Now(),
	}

	loadedEntryIntf, loaded := c.entries.LoadOrStore(key, entry)
	if loaded {
		entry := loadedEntryIntf.(*cacheEntry)
		return c.trySetEntry(entry, state, keyType)
	}

	return true
}

func (c *Cache) trySetEntry(entry *cacheEntry, state CacheEntryState, keyType LedisType) bool {
	entry.RWMutex.Lock()

	if entry.State == CacheEntryStateLoading {
		entry.RWMutex.Unlock()

		return false
	} else {
		entry.DoneLoading = nil
		entry.State = state
		entry.Type = keyType
		entry.WrittenAt = time.Now()

		entry.RWMutex.Unlock()

		return true
	}
}

// LoadOrCreateEntry is the primary access method for the cache. For a given
// key, it returns either entry data or a setter. If a setter is returned, the
// caller must fulfill the setter by calling .Set().
//
// The bool returned indicates whether the entry exists and is in the Loading
// or Exists state (this means the returned CacheEntryData is valid). If the
// bool is false, the returned CacheEntrySetter is valid.
func (c *Cache) LoadOrCreateEntry(key string) (CacheEntryData, CacheEntrySetter, bool) {
	entry, ok := c.loadEntry(key)
	if ok {
		return c.prepareEntry(key, entry)
	}

	doneLoading := make(chan struct{})
	entry = &cacheEntry{
		DoneLoading: doneLoading,
		State:       CacheEntryStateLoading,
		Type:        LedisTypeNone,
	}

	loadedEntryIntf, loaded := c.entries.LoadOrStore(key, entry)
	if loaded {
		entry := loadedEntryIntf.(*cacheEntry)
		return c.prepareEntry(key, entry)
	}

	return CacheEntryData{}, CacheEntrySetter{
		entry:       entry,
		Key:         key,
		doneLoading: doneLoading,
	}, false
}

func (c *Cache) prepareEntry(key string, entry *cacheEntry) (
	CacheEntryData, CacheEntrySetter, bool,
) {
	entry.RWMutex.RLock()

	if entry.State == CacheEntryStateExists || entry.State == CacheEntryStateLoading {
		entryData := CacheEntryData{
			entry: entry,
			Key:   key,
		}
		entryData.copyFrom(entry)

		entry.RWMutex.RUnlock()

		return entryData, CacheEntrySetter{}, true
	} else {
		// Switch to write lock: Release read lock, acquire write lock
		entry.RWMutex.RUnlock()
		entry.RWMutex.Lock()

		// Re-check the state
		if entry.State == CacheEntryStateExists || entry.State == CacheEntryStateLoading {
			entryData := CacheEntryData{
				entry: entry,
				Key:   key,
			}
			entryData.copyFrom(entry)

			entry.RWMutex.Unlock()

			return entryData, CacheEntrySetter{}, true
		} else {
			doneLoading := make(chan struct{})

			entry.DoneLoading = doneLoading
			entry.State = CacheEntryStateLoading
			entry.Type = LedisTypeNone

			entry.RWMutex.Unlock()

			return CacheEntryData{}, CacheEntrySetter{
				entry:       entry,
				Key:         key,
				doneLoading: doneLoading,
			}, false
		}
	}
}

func (c *Cache) loadEntry(key string) (*cacheEntry, bool) {
	entryIntf, ok := c.entries.Load(key)
	if !ok {
		return nil, false
	}

	entry := entryIntf.(*cacheEntry)

	return entry, true
}

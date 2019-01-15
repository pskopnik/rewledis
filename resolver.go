package rewledis

import (
	"context"
	"errors"

	"github.com/gomodule/redigo/redis"
)

// Error variables related to Resolver.
var (
	ErrUnexpectedCacheEntryState = errors.New("encountered cache entry with unexpected state")
	ErrErrorCacheEntryState      = errors.New("encountered cache entry with Error state")
)

type TypeInfo struct {
	Key  string
	Type LedisType
}

// Resolver provides functionality to resolve the type of keys while using a
// Cache instance.
type Resolver struct {
	Cache *Cache
	Pool  *redis.Pool
}

func (r *Resolver) ResolveOne(ctx context.Context, key string) (LedisType, error) {
	var typesInfo [1]TypeInfo
	keys := [1]string{key}
	_, err := r.ResolveAppend(typesInfo[:0], ctx, keys[:])
	if err != nil {
		return LedisTypeNone, err
	}

	return typesInfo[0].Type, nil
}

func (r *Resolver) ResolveAppend(typesInfo []TypeInfo, ctx context.Context, keys []string) ([]TypeInfo, error) {
	// Array for pre-allocated on-stack slices
	var entriesDataArray [4]CacheEntryData
	var entrySettersArray [4]CacheEntrySetter

	entriesData := entriesDataArray[:0]
	entrySetters := entrySettersArray[:0]
	inputTypesInfo := typesInfo

	for _, key := range keys {
		entryData, entrySetter, exists := r.Cache.LoadOrCreateEntry(key)
		if exists {
			if entryData.State == CacheEntryStateExists {
				typesInfo = append(typesInfo, TypeInfo{
					Key:  key,
					Type: entryData.Type,
				})
			} else if entryData.State == CacheEntryStateLoading {
				entriesData = append(entriesData, entryData)
			} else {
				// This should not occur, if the entry's state is not one of
				// the above, Cache returns exists = false.
				for i := range entrySetters {
					entrySetters[i].Set(CacheEntryStateError, LedisTypeNone)
				}

				return inputTypesInfo, ErrUnexpectedCacheEntryState
			}
		} else {
			entrySetters = append(entrySetters, entrySetter)
		}
	}

	beginIndex := len(typesInfo)
	typesInfo = append(typesInfo, make([]TypeInfo, len(entrySetters))...)
	for i := beginIndex; i < len(typesInfo); i++ {
		typesInfo[i].Key = entrySetters[i].Key
	}
	err := r.activeResolve(ctx, entrySetters, typesInfo[beginIndex:])
	if err != nil {
		// activeResolve sets all entrySetters in case of error
		return inputTypesInfo, err
	}

	beginIndex = len(typesInfo)
	typesInfo = append(typesInfo, make([]TypeInfo, len(entriesData))...)
	for i := beginIndex; i < len(typesInfo); i++ {
		typesInfo[i].Key = entriesData[i].Key
	}
	err = r.waitResolve(ctx, entriesData, typesInfo[beginIndex:])
	if err != nil {
		return inputTypesInfo, err
	}

	return typesInfo, nil
}

func (r *Resolver) activeResolve(ctx context.Context, entrySetters []CacheEntrySetter, typesInfo []TypeInfo) error {
	ledisTypes := [...]LedisType{
		LedisTypeKV,
		LedisTypeList,
		LedisTypeHash,
		LedisTypeSet,
		LedisTypeZSet,
	}

	noneTypesInfo := typesInfo
	noneEntrySetters := entrySetters

	for _, ledisType := range ledisTypes {
		err := r.checkType(ctx, ledisType, noneTypesInfo)
		if err != nil {
			for i := range noneEntrySetters {
				noneEntrySetters[i].Set(CacheEntryStateError, LedisTypeNone)
			}

			return err
		}

		noneBegin := r.sortApartCoSortEntrySetters(noneTypesInfo, noneEntrySetters)

		for i := 0; i < noneBegin; i++ {
			noneEntrySetters[i].Set(CacheEntryStateExists, noneTypesInfo[i].Type)
		}

		noneTypesInfo = noneTypesInfo[noneBegin:]
		noneEntrySetters = noneEntrySetters[noneBegin:]

		if len(noneTypesInfo) == 0 {
			return nil
		}
	}

	for i := range noneEntrySetters {
		noneEntrySetters[i].Set(CacheEntryStateDeleted, LedisTypeNone)
	}

	return nil
}

func (r *Resolver) waitResolve(ctx context.Context, entriesData []CacheEntryData, typesInfo []TypeInfo) error {
	done := ctx.Done()

	for i := range entriesData {
		entryData := &entriesData[i]
	L:
		for {
			select {
			case <-entryData.DoneLoading:
				entryData.Refresh()
				if entryData.State == CacheEntryStateExists {
					typesInfo[i].Type = entryData.Type
					break L
				} else if entryData.State == CacheEntryStateDeleted {
					typesInfo[i].Type = LedisTypeNone
					break L
				} else if entryData.State == CacheEntryStateError {
					return ErrErrorCacheEntryState
				} else if entryData.State == CacheEntryStateLoading {
					continue L
				} else {
					return ErrUnexpectedCacheEntryState
				}
			case <-done:
				return ctx.Err()
			}
		}
	}

	return nil
}

// sortApart sorts the passed typesInfo slice. The first k entries have a type
// != LedisTypeNone, the remaining entries have type == LedisTypeNone. The
// value of k is returned.
func (r *Resolver) sortApart(typesInfo []TypeInfo) int {
	noneBegin := len(typesInfo)

	for i := 0; i < noneBegin; i++ {
		for i < noneBegin {
			if typesInfo[i].Type != LedisTypeNone {
				break
			}

			noneBegin--

			// TODO potential self-assignment: should this be if-ed out?
			typesInfo[i], typesInfo[noneBegin] = typesInfo[noneBegin], typesInfo[i]
		}
	}

	return noneBegin
}

// sortApartCoSortEntrySetters does the same as sortApart, but applies the
// same permutations to the entrySetters slice. This means, that for all i:
// there exists j: before(typesInfo[i]) == typesInfo[j] and
// before(entrySetters[i]) == entrySetters[j].
func (r *Resolver) sortApartCoSortEntrySetters(typesInfo []TypeInfo, entrySetters []CacheEntrySetter) int {
	noneBegin := len(typesInfo)

	for i := 0; i < noneBegin; i++ {
		for i < noneBegin {
			if typesInfo[i].Type != LedisTypeNone {
				break
			}

			noneBegin--

			// TODO potential self-assignment: should this be if-ed out?
			typesInfo[i], typesInfo[noneBegin] = typesInfo[noneBegin], typesInfo[i]
			entrySetters[i], entrySetters[noneBegin] = entrySetters[noneBegin], entrySetters[i]
		}
	}

	return noneBegin
}

func (r *Resolver) checkType(ctx context.Context, checkType LedisType, typesInfo []TypeInfo) error {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	command, err := r.existsCommandForType(checkType)
	if err != nil {
		return err
	}

	for i := range typesInfo {
		err = conn.Send(command, typesInfo[i].Key)
		if err != nil {
			return err
		}
	}

	err = conn.Flush()
	if err != nil {
		return err
	}

	for i := range typesInfo {
		existsCount, err := redis.Int(conn.Receive())
		if err != nil {
			return err
		}
		if existsCount == 1 {
			typesInfo[i].Type = checkType
		}
	}

	return nil
}

func (r *Resolver) existsCommandForType(ledisType LedisType) (string, error) {
	switch ledisType {
	case LedisTypeKV:
		return "EXISTS", nil
	case LedisTypeList:
		return "LKEYEXISTS", nil
	case LedisTypeHash:
		return "HKEYEXISTS", nil
	case LedisTypeSet:
		return "SKEYEXISTS", nil
	case LedisTypeZSet:
		return "ZKEYEXISTS", nil
	case LedisTypeNone:
		return "", ErrInvalidLedisType
	default:
		return "", ErrInvalidLedisType
	}
}

package rewledis

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownLedisTypeString   = errors.New("input string does not represent a known LedisType value")
	ErrInvalidLedisType         = errors.New("input LedisType value is unknown or otherwise invalid for this operation")
	ErrNoCorrespondingLedisType = errors.New("there is no LedisType value corresponding to the input RedisType value")
)

type LedisType int8

const (
	LedisTypeNone LedisType = iota
	LedisTypeKV
	LedisTypeList
	LedisTypeHash
	LedisTypeSet
	LedisTypeZSet
)

func (l LedisType) String() string {
	switch l {
	case LedisTypeNone:
		return "None"
	case LedisTypeKV:
		return "KV"
	case LedisTypeList:
		return "List"
	case LedisTypeHash:
		return "Hash"
	case LedisTypeSet:
		return "Set"
	case LedisTypeZSet:
		return "ZSet"
	default:
		return fmt.Sprintf("LedisType(%d)", l)
	}
}

func ParseLedisType(str string) (LedisType, error) {
	switch str {
	case "None":
		return LedisTypeNone, nil
	case "KV":
		return LedisTypeKV, nil
	case "List":
		return LedisTypeList, nil
	case "Hash":
		return LedisTypeHash, nil
	case "Set":
		return LedisTypeSet, nil
	case "ZSet":
		return LedisTypeZSet, nil
	default:
		return LedisTypeNone, ErrUnknownLedisTypeString
	}

}
func ParseLedisTypeFromLedis(str string) (LedisType, error) {
	switch str {
	case "KV":
		return LedisTypeKV, nil
	case "LIST":
		return LedisTypeList, nil
	case "HASH":
		return LedisTypeHash, nil
	case "SET":
		return LedisTypeSet, nil
	case "ZSET":
		return LedisTypeZSet, nil
	default:
		return LedisTypeNone, ErrUnknownLedisTypeString
	}
}

func LedisTypeFromRedisType(redisType RedisType) (LedisType, error) {
	switch redisType {
	case RedisTypeNone:
		return LedisTypeNone, nil
	case RedisTypeString:
		return LedisTypeKV, nil
	case RedisTypeList:
		return LedisTypeList, nil
	case RedisTypeHash:
		return LedisTypeHash, nil
	case RedisTypeSet:
		return LedisTypeSet, nil
	case RedisTypeZSet:
		return LedisTypeZSet, nil
	case RedisTypeGeneric:
		return LedisTypeNone, ErrNoCorrespondingLedisType
	default:
		return LedisTypeNone, ErrInvalidRedisType
	}
}

var (
	ErrUnknownRedisTypeString = errors.New("input string does not represent a known RedisType value")
	ErrInvalidRedisType       = errors.New("input RedisType value is unknown or otherwise invalid for this operation")
)

type RedisType int8

const (
	RedisTypeNone RedisType = iota
	RedisTypeString
	RedisTypeList
	RedisTypeHash
	RedisTypeSet
	RedisTypeZSet

	RedisTypeGeneric = -2
)

func (r RedisType) String() string {
	switch r {
	case RedisTypeNone:
		return "None"
	case RedisTypeString:
		return "String"
	case RedisTypeList:
		return "List"
	case RedisTypeHash:
		return "Hash"
	case RedisTypeSet:
		return "Set"
	case RedisTypeZSet:
		return "ZSet"
	case RedisTypeGeneric:
		return "Generic"
	default:
		return fmt.Sprintf("RedisType(%d)", r)
	}
}

func ParseRedisType(str string) (RedisType, error) {
	switch str {
	case "None":
		return RedisTypeNone, nil
	case "String":
		return RedisTypeString, nil
	case "List":
		return RedisTypeList, nil
	case "Hash":
		return RedisTypeHash, nil
	case "Set":
		return RedisTypeSet, nil
	case "ZSet":
		return RedisTypeZSet, nil
	case "Generic":
		return RedisTypeGeneric, nil
	default:
		return 0, ErrUnknownRedisTypeString
	}

}
func ParseRedisTypeFromRedis(str string) (RedisType, error) {
	switch str {
	case "string":
		return RedisTypeString, nil
	case "list":
		return RedisTypeList, nil
	case "hash":
		return RedisTypeHash, nil
	case "set":
		return RedisTypeSet, nil
	case "zset":
		return RedisTypeZSet, nil
	default:
		return 0, ErrUnknownRedisTypeString
	}
}

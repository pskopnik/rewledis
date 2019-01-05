package rewledis

import (
	"strings"
)

// RedisCommand variables describing the Redis commands operating on
// strings (RedisTypeString).
//
//     https://redis.io/commands#string
var (
	RedisCommandAPPEND = RedisCommand{
		Name:          "APPEND",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "APPEND key value",
	}

	RedisCommandBITCOUNT = RedisCommand{
		Name:          "BITCOUNT",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "BITCOUNT key [start end]",
	}

	// BITFIELD command is not implemented in LedisDB.

	// RedisCommandBITOP contains information about the BITOP Redis command.
	// BITOP writes the result of operation to destkey. However, it treats
	// unset keys as "\x00" * max(len key). I.e. the commands succeeds even if
	// some keys are not set.
	//
	// Perhaps KeyExtractor should be KeysAtIndices(1) instead, as only
	// destkey is a string.
	// If no keys are set the result of applying the operation is a 0 length
	// string. In this case Redis sets destkey to nil (does not exist).
	RedisCommandBITOP = RedisCommand{
		Name:          "BITOP",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysFromIndex(1),
		TransformFunc: NoneTransformer(),
		Syntax:        "BITOP operation destkey key [key ...]",
	}

	RedisCommandBITPOS = RedisCommand{
		Name:          "BITPOS",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "BITPOS key bit [start] [end]",
	}

	RedisCommandDECR = RedisCommand{
		Name:          "DECR",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "DECR key",
	}

	RedisCommandDECRBY = RedisCommand{
		Name:          "DECRBY",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "DECRBY key decrement",
	}

	RedisCommandGET = RedisCommand{
		Name:          "GET",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "GET key",
	}

	RedisCommandGETBIT = RedisCommand{
		Name:          "GETBIT",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "GETBIT key offset",
	}

	RedisCommandGETRANGE = RedisCommand{
		Name:          "GETRANGE",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "GETRANGE key start end",
	}

	RedisCommandGETSET = RedisCommand{
		Name:          "GETSET",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "GETSET key value",
	}

	RedisCommandINCR = RedisCommand{
		Name:          "INCR",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "INCR key",
	}

	RedisCommandINCRBY = RedisCommand{
		Name:          "INCRBY",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "INCRBY key increment",
	}

	// INCRBYFLOAT command is not implemented in LedisDB.

	RedisCommandMGET = RedisCommand{
		Name:          "MGET",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysFromIndex(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "MGET key [key ...]",
	}

	RedisCommandMSET = RedisCommand{
		Name:          "MSET",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysFromIndex(0, 1),
		TransformFunc: NoneTransformer(),
		Syntax:        "MSET key value [key value ...]",
	}

	// MSETNX command is not implemented in LedisDB.

	// PSETEX command is not implemented in LedisDB.

	RedisCommandSET = RedisCommand{
		Name:          "SET",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SET key value [expiration EX seconds|PX milliseconds] [NX|XX]",
	}

	RedisCommandSETBIT = RedisCommand{
		Name:          "SETBIT",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SETBIT key offset value",
	}

	RedisCommandSETEX = RedisCommand{
		Name:          "SETEX",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SETEX key seconds value",
	}

	RedisCommandSETNX = RedisCommand{
		Name:          "SETNX",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SETNX key value",
	}

	RedisCommandSETRANGE = RedisCommand{
		Name:          "SETRANGE",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SETRANGE key offset value",
	}

	RedisCommandSTRLEN = RedisCommand{
		Name:          "STRLEN",
		KeyType:       RedisTypeString,
		KeyExtractor:  KeysAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "STRLEN key",
	}
)

// RedisCommand variables describing the Redis commands operating on
// hashes (RedisTypeHash).
//
//     https://redis.io/commands#hash
var (
	RedisCommandHDEL = RedisCommand{
		Name:         "HDEL",
		KeyType:      RedisTypeHash,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "HDEL key field [field ...]",
	}

	RedisCommandHEXISTS = RedisCommand{
		Name:         "HEXISTS",
		KeyType:      RedisTypeHash,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "HEXISTS key field",
	}

	RedisCommandHGET = RedisCommand{
		Name:         "HGET",
		KeyType:      RedisTypeHash,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "HGET key field",
	}

	RedisCommandHGETALL = RedisCommand{
		Name:         "HGETALL",
		KeyType:      RedisTypeHash,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "HGETALL key",
	}

	RedisCommandHINCRBY = RedisCommand{
		Name:         "HINCRBY",
		KeyType:      RedisTypeHash,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "HINCRBY key field increment",
	}

	// HINCRBYFLOAT command is not implemented in LedisDB.

	RedisCommandHKEYS = RedisCommand{
		Name:         "HKEYS",
		KeyType:      RedisTypeHash,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "HKEYS key",
	}

	RedisCommandHLEN = RedisCommand{
		Name:         "HLEN",
		KeyType:      RedisTypeHash,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "HLEN key",
	}

	RedisCommandHMGET = RedisCommand{
		Name:         "HMGET",
		KeyType:      RedisTypeHash,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "HMGET key field [field ...]",
	}

	RedisCommandHMSET = RedisCommand{
		Name:         "HMSET",
		KeyType:      RedisTypeHash,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "HMSET key field value [field value ...]",
	}

	RedisCommandHSET = RedisCommand{
		Name:         "HSET",
		KeyType:      RedisTypeHash,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "HSET key field value",
	}

	// HSETNX command is not implemented in LedisDB.

	// HSTRLEN command is not implemented in LedisDB.

	RedisCommandHVALS = RedisCommand{
		Name:         "HVALS",
		KeyType:      RedisTypeHash,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "HVALS key",
	}

	// HSCAN command is not implemented in LedisDB.
)

// RedisCommand variables describing the Redis commands operating on
// lists (RedisTypeList).
//
//     https://redis.io/commands#list
var (
	RedisCommandBLPOP = RedisCommand{
		Name:         "BLPOP",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysFromUntilIndex(0, -1),
		Syntax:       "BLPOP key [key ...] timeout",
	}

	RedisCommandBRPOP = RedisCommand{
		Name:         "BRPOP",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysFromUntilIndex(0, -1),
		Syntax:       "BRPOP key [key ...] timeout",
	}

	RedisCommandBRPOPLPUSH = RedisCommand{
		Name:         "BRPOPLPUSH",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysAtIndices(0, 1),
		Syntax:       "BRPOPLPUSH source destination timeout",
	}

	RedisCommandLINDEX = RedisCommand{
		Name:         "LINDEX",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "LINDEX key index",
	}

	// LINSERT command is not implemented in LedisDB.

	RedisCommandLLEN = RedisCommand{
		Name:         "LLEN",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "LLEN key index",
	}

	RedisCommandLPOP = RedisCommand{
		Name:         "LPOP",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "LPOP key",
	}

	RedisCommandLPUSH = RedisCommand{
		Name:         "LPUSH",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "LPUSH key value [value ...]",
	}

	// LPUSHX command is not implemented in LedisDB.

	RedisCommandLRANGE = RedisCommand{
		Name:         "LRANGE",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "LRANGE key start stop",
	}

	// LREM command is not implemented in LedisDB.

	// LSET command is not implemented in LedisDB.

	RedisCommandLTRIM = RedisCommand{
		Name:         "LTRIM",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "LTRIM key start stop",
	}

	RedisCommandRPOP = RedisCommand{
		Name:         "RPOP",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "RPOP key",
	}

	RedisCommandRPOPLPUSH = RedisCommand{
		Name:         "RPOPLPUSH",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysAtIndices(0, 1),
		Syntax:       "RPOPLPUSH source destination",
	}

	RedisCommandRPUSH = RedisCommand{
		Name:         "RPUSH",
		KeyType:      RedisTypeList,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "RPUSH key value [value ...]",
	}

	// RPUSHX command is not implemented in LedisDB.
)

// RedisCommand variables describing the Redis commands operating on
// sets (RedisTypeSet).
//
//     https://redis.io/commands#set
var (
	RedisCommandSADD = RedisCommand{
		Name:         "SADD",
		KeyType:      RedisTypeSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "SADD key member [member ...]",
	}

	RedisCommandSCARD = RedisCommand{
		Name:         "SCARD",
		KeyType:      RedisTypeSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "SCARD key",
	}

	RedisCommandSDIFF = RedisCommand{
		Name:         "SDIFF",
		KeyType:      RedisTypeSet,
		KeyExtractor: KeysFromIndex(0),
		Syntax:       "SDIFF key [key ...]",
	}

	RedisCommandSDIFFSTORE = RedisCommand{
		Name:         "SDIFFSTORE",
		KeyType:      RedisTypeSet,
		KeyExtractor: KeysFromIndex(0),
		Syntax:       "SDIFFSTORE destination key [key ...]",
	}

	RedisCommandSINTER = RedisCommand{
		Name:         "SINTER",
		KeyType:      RedisTypeSet,
		KeyExtractor: KeysFromIndex(0),
		Syntax:       "SINTER key [key ...]",
	}

	RedisCommandSINTERSTORE = RedisCommand{
		Name:         "SINTERSTORE",
		KeyType:      RedisTypeSet,
		KeyExtractor: KeysFromIndex(0),
		Syntax:       "SINTERSTORE destination key [key ...]",
	}

	RedisCommandSISMEMBER = RedisCommand{
		Name:         "SISMEMBER",
		KeyType:      RedisTypeSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "SISMEMBER key member",
	}

	RedisCommandSMEMBERS = RedisCommand{
		Name:         "SMEMBERS",
		KeyType:      RedisTypeSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "SMEMBERS key",
	}

	// SMOVE command is not implemented in LedisDB.

	// SPOP command is not implemented in LedisDB.

	// SRANDMEMBER command is not implemented in LedisDB.

	RedisCommandSREM = RedisCommand{
		Name:         "SREM",
		KeyType:      RedisTypeSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "SREM key member [member ...]",
	}

	RedisCommandSUNION = RedisCommand{
		Name:         "SUNION",
		KeyType:      RedisTypeSet,
		KeyExtractor: KeysFromIndex(0),
		Syntax:       "SUNION key [key ...]",
	}

	RedisCommandSUNIONSTORE = RedisCommand{
		Name:         "SUNIONSTORE",
		KeyType:      RedisTypeSet,
		KeyExtractor: KeysFromIndex(0),
		Syntax:       "SUNIONSTORE destination key [key ...]",
	}
)

// RedisCommand variables describing the Redis commands operating on
// sorted sets (RedisTypeZSet).
//
//     https://redis.io/commands#sorted_set
var (
	// BZPOPMIN command is not implemented in LedisDB.

	// BZPOPMAX command is not implemented in LedisDB.

	RedisCommandZADD = RedisCommand{
		Name:         "ZADD",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZADD key [NX|XX] [CH] [INCR] score member [score member ...]",
	}

	RedisCommandZCARD = RedisCommand{
		Name:         "ZCARD",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZCARD key",
	}

	RedisCommandZCOUNT = RedisCommand{
		Name:         "ZCOUNT",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZCOUNT key min max",
	}

	RedisCommandZINCRBY = RedisCommand{
		Name:         "ZINCRBY",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZINCRBY key increment member",
	}

	// RedisCommandZINTERSTORE contains information about the ZINTERSTORE Redis
	// command.
	//
	// The KeyExtractor only extracts the destination key. Further keys could
	// be extracted by introducing a new KeyExtractor implementation.
	// However, Redis interprets non-existing key as empty keys.
	RedisCommandZINTERSTORE = RedisCommand{
		Name:         "ZINTERSTORE",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZINTERSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]",
	}

	RedisCommandZLEXCOUNT = RedisCommand{
		Name:         "ZLEXCOUNT",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZLEXCOUNT key min max",
	}

	// ZPOPMAX command is not implemented in LedisDB.

	// ZPOPMIN command is not implemented in LedisDB.

	RedisCommandZRANGE = RedisCommand{
		Name:         "ZRANGE",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZRANGE key start stop [WITHSCORES]",
	}

	RedisCommandZRANGEBYLEX = RedisCommand{
		Name:         "ZRANGEBYLEX",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZRANGEBYLEX key min max [LIMIT offset count]",
	}

	RedisCommandZRANGEBYSCORE = RedisCommand{
		Name:         "ZRANGEBYSCORE",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]",
	}

	RedisCommandZRANK = RedisCommand{
		Name:         "ZRANK",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZRANK key member",
	}

	RedisCommandZREM = RedisCommand{
		Name:         "ZREM",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZREM key member [member ...]",
	}

	RedisCommandZREMRANGEBYLEX = RedisCommand{
		Name:         "ZREMRANGEBYLEX",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZREMRANGEBYLEX key min max",
	}

	RedisCommandZREMRANGEBYRANK = RedisCommand{
		Name:         "ZREMRANGEBYRANK",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZREMRANGEBYRANK key start stop",
	}

	RedisCommandZREMRANGEBYSCORE = RedisCommand{
		Name:         "ZREMRANGEBYSCORE",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZREMRANGEBYSCORE key min max",
	}

	RedisCommandZREVRANGE = RedisCommand{
		Name:         "ZREVRANGE",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZREVRANGE key start stop [WITHSCORES]",
	}

	// ZREVRANGEBYLEX command is not implemented in LedisDB.

	RedisCommandZREVRANGEBYSCORE = RedisCommand{
		Name:         "ZREVRANGEBYSCORE",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZREVRANGEBYSCORE key max min [WITHSCORES] [LIMIT offset count]",
	}

	RedisCommandZREVRANK = RedisCommand{
		Name:         "ZREVRANK",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZREVRANK key member",
	}

	RedisCommandZSCORE = RedisCommand{
		Name:         "ZSCORE",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZSCORE key member",
	}

	// RedisCommandZUNIONSTORE contains information about the ZUNIONSTORE Redis
	// command.
	//
	// The KeyExtractor only extracts the destination key. Further keys could
	// be extracted by introducing a new KeyExtractor implementation.
	// However, Redis interprets non-existing key as empty keys.
	RedisCommandZUNIONSTORE = RedisCommand{
		Name:         "ZUNIONSTORE",
		KeyType:      RedisTypeZSet,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "ZUNIONSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]",
	}

	// ZSCAN command is not implemented in LedisDB.
)

// RedisCommand variables describing the Redis commands operating on
// keys / generics (RedisTypeGeneric).
//
//     https://redis.io/commands#generic
var (
	RedisCommandDEL = RedisCommand{
		Name:         "DEL",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: KeysFromIndex(0),
		TransformFunc: TypeSpecificBulkTransformer(&TypeSpecificBulkTransformerConfig{
			Commands:   TypeSpecificCommands{
				KV   :"DEL",
				List :"LMCLEAR",
				Hash :"HMCLEAR",
				Set  :"SMCLEAR",
				ZSet :"ZMCLEAR",
			},
			Aggregator: AggregatorSum,
		}),
		Syntax:       "DEL key [key ...]",
	}

	RedisCommandDUMP = RedisCommand{
		Name:         "DUMP",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "DUMP key",
	}

	RedisCommandEXISTS = RedisCommand{
		Name:         "EXISTS",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "EXISTS key",
	}

	RedisCommandEXPIRE = RedisCommand{
		Name:         "EXPIRE",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "EXPIRE key seconds",
	}

	RedisCommandEXPIREAT = RedisCommand{
		Name:         "EXPIREAT",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "EXPIREAT key timestamp",
	}

	// KEYS command is not implemented in LedisDB.

	// MIGRATE command is not implemented in LedisDB.

	// MOVE command is not implemented in LedisDB.

	// OBJECT command is not implemented in LedisDB.

	RedisCommandPERSIST = RedisCommand{
		Name:         "PERSIST",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "PERSIST key",
	}

	// PEXPIRE command is not implemented in LedisDB.

	// PEXPIREAT command is not implemented in LedisDB.

	// PTTL command is not implemented in LedisDB.

	// RANDOMKEY command is not implemented in LedisDB.

	// RENAME command is not implemented in LedisDB.

	// RENAMENX command is not implemented in LedisDB.

	// RESTORE command is not implemented in LedisDB.

	// SORT command is not implemented in LedisDB.

	// TOUCH command is not implemented in LedisDB.

	RedisCommandTTL = RedisCommand{
		Name:         "TTL",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: KeysAtIndices(0),
		Syntax:       "TTL key",
	}

	// TYPE command is not implemented in LedisDB.

	// UNLINK command is not implemented in LedisDB.

	// WAIT command is not implemented in LedisDB.

	// SCAN command is not implemented in LedisDB.
)

func RedisCommandFromName(name string) (*RedisCommand, error) {
	switch strings.ToUpper(name) {
	case "APPEND":
		return &RedisCommandAPPEND, nil
	case "BITCOUNT":
		return &RedisCommandBITCOUNT, nil
	case "BITOP":
		return &RedisCommandBITOP, nil
	case "BITPOS":
		return &RedisCommandBITPOS, nil
	case "DECR":
		return &RedisCommandDECR, nil
	case "DECRBY":
		return &RedisCommandDECRBY, nil
	case "GET":
		return &RedisCommandGET, nil
	case "GETBIT":
		return &RedisCommandGETBIT, nil
	case "GETRANGE":
		return &RedisCommandGETRANGE, nil
	case "GETSET":
		return &RedisCommandGETSET, nil
	case "INCR":
		return &RedisCommandINCR, nil
	case "INCRBY":
		return &RedisCommandINCRBY, nil
	case "MGET":
		return &RedisCommandMGET, nil
	case "MSET":
		return &RedisCommandMSET, nil
	case "SET":
		return &RedisCommandSET, nil
	case "SETBIT":
		return &RedisCommandSETBIT, nil
	case "SETEX":
		return &RedisCommandSETEX, nil
	case "SETNX":
		return &RedisCommandSETNX, nil
	case "SETRANGE":
		return &RedisCommandSETRANGE, nil
	case "STRLEN":
		return &RedisCommandSTRLEN, nil
	case "HDEL":
		return &RedisCommandHDEL, nil
	case "HEXISTS":
		return &RedisCommandHEXISTS, nil
	case "HGET":
		return &RedisCommandHGET, nil
	case "HGETALL":
		return &RedisCommandHGETALL, nil
	case "HINCRBY":
		return &RedisCommandHINCRBY, nil
	case "HKEYS":
		return &RedisCommandHKEYS, nil
	case "HLEN":
		return &RedisCommandHLEN, nil
	case "HMGET":
		return &RedisCommandHMGET, nil
	case "HMSET":
		return &RedisCommandHMSET, nil
	case "HSET":
		return &RedisCommandHSET, nil
	case "HVALS":
		return &RedisCommandHVALS, nil
	case "BLPOP":
		return &RedisCommandBLPOP, nil
	case "BRPOP":
		return &RedisCommandBRPOP, nil
	case "BRPOPLPUSH":
		return &RedisCommandBRPOPLPUSH, nil
	case "LINDEX":
		return &RedisCommandLINDEX, nil
	case "LLEN":
		return &RedisCommandLLEN, nil
	case "LPOP":
		return &RedisCommandLPOP, nil
	case "LPUSH":
		return &RedisCommandLPUSH, nil
	case "LRANGE":
		return &RedisCommandLRANGE, nil
	case "LTRIM":
		return &RedisCommandLTRIM, nil
	case "RPOP":
		return &RedisCommandRPOP, nil
	case "RPOPLPUSH":
		return &RedisCommandRPOPLPUSH, nil
	case "RPUSH":
		return &RedisCommandRPUSH, nil
	case "SADD":
		return &RedisCommandSADD, nil
	case "SCARD":
		return &RedisCommandSCARD, nil
	case "SDIFF":
		return &RedisCommandSDIFF, nil
	case "SDIFFSTORE":
		return &RedisCommandSDIFFSTORE, nil
	case "SINTER":
		return &RedisCommandSINTER, nil
	case "SINTERSTORE":
		return &RedisCommandSINTERSTORE, nil
	case "SISMEMBER":
		return &RedisCommandSISMEMBER, nil
	case "SMEMBERS":
		return &RedisCommandSMEMBERS, nil
	case "SREM":
		return &RedisCommandSREM, nil
	case "SUNION":
		return &RedisCommandSUNION, nil
	case "SUNIONSTORE":
		return &RedisCommandSUNIONSTORE, nil
	case "ZADD":
		return &RedisCommandZADD, nil
	case "ZCARD":
		return &RedisCommandZCARD, nil
	case "ZCOUNT":
		return &RedisCommandZCOUNT, nil
	case "ZINCRBY":
		return &RedisCommandZINCRBY, nil
	case "ZINTERSTORE":
		return &RedisCommandZINTERSTORE, nil
	case "ZLEXCOUNT":
		return &RedisCommandZLEXCOUNT, nil
	case "ZRANGE":
		return &RedisCommandZRANGE, nil
	case "ZRANGEBYLEX":
		return &RedisCommandZRANGEBYLEX, nil
	case "ZRANGEBYSCORE":
		return &RedisCommandZRANGEBYSCORE, nil
	case "ZRANK":
		return &RedisCommandZRANK, nil
	case "ZREM":
		return &RedisCommandZREM, nil
	case "ZREMRANGEBYLEX":
		return &RedisCommandZREMRANGEBYLEX, nil
	case "ZREMRANGEBYRANK":
		return &RedisCommandZREMRANGEBYRANK, nil
	case "ZREMRANGEBYSCORE":
		return &RedisCommandZREMRANGEBYSCORE, nil
	case "ZREVRANGE":
		return &RedisCommandZREVRANGE, nil
	case "ZREVRANGEBYSCORE":
		return &RedisCommandZREVRANGEBYSCORE, nil
	case "ZREVRANK":
		return &RedisCommandZREVRANK, nil
	case "ZSCORE":
		return &RedisCommandZSCORE, nil
	case "ZUNIONSTORE":
		return &RedisCommandZUNIONSTORE, nil
	case "DEL":
		return &RedisCommandDEL, nil
	case "DUMP":
		return &RedisCommandDUMP, nil
	case "EXISTS":
		return &RedisCommandEXISTS, nil
	case "EXPIRE":
		return &RedisCommandEXPIRE, nil
	case "EXPIREAT":
		return &RedisCommandEXPIREAT, nil
	case "PERSIST":
		return &RedisCommandPERSIST, nil
	case "TTL":
		return &RedisCommandTTL, nil
	default:
		return nil, ErrUnknownRedisCommandName
	}
}

// Ledis commands:

// HCLEAR key
// HDUMP key
// HKEYEXISTS key
// HEXPIRE key seconds
// HEXPIREAT key timestamp
// HMCLEAR key [key...]
// HPERSIST key
// HTTL key

// LCLEAR key
// LDUMP key

// LEXPIRE key seconds
// LEXPIREAT key timestamp
// LKEYEXISTS key
// LMCLEAR key [key ...]
// LPERSIST key
// LTTL key

// SCLEAR key
// SDUMP key
// SEXPIRE key seconds
// SEXPIREAT key timestamp
// SKEYEXISTS key
// SMCLEAR key [key ...]
// SPERSIST key
// STTL key

// ZCLEAR key
// ZDUMP key
// ZEXPIRE key seconds
// ZEXPIREAT key timestamp
// ZKEYEXISTS key
// ZMCLEAR key [key ...]
// ZPERSIST key
// ZTTL key

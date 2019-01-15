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
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "APPEND key value",
	}

	RedisCommandBITCOUNT = RedisCommand{
		Name:          "BITCOUNT",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "BITCOUNT key [start end]",
	}

	// BITFIELD command is not implemented in LedisDB.

	// RedisCommandBITOP contains information about the BITOP Redis command.
	// BITOP writes the result of operation to destkey. However, it treats
	// unset keys as "\x00" * max(len key). I.e. the commands succeeds even if
	// some keys are not set.
	//
	// Perhaps KeyExtractor should be ArgsAtIndices(1) instead, as only
	// destkey is a string.
	// If no keys are set the result of applying the operation is a 0 length
	// string. In this case Redis sets destkey to nil (does not exist).
	//
	// TODO: Define the exact meaning of KeyType / KeyExtractor.
	RedisCommandBITOP = RedisCommand{
		Name:          "BITOP",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsFromIndex(1),
		TransformFunc: NoneTransformer(),
		Syntax:        "BITOP operation destkey key [key ...]",
	}

	RedisCommandBITPOS = RedisCommand{
		Name:          "BITPOS",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "BITPOS key bit [start] [end]",
	}

	RedisCommandDECR = RedisCommand{
		Name:          "DECR",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "DECR key",
	}

	RedisCommandDECRBY = RedisCommand{
		Name:          "DECRBY",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "DECRBY key decrement",
	}

	RedisCommandGET = RedisCommand{
		Name:          "GET",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "GET key",
	}

	RedisCommandGETBIT = RedisCommand{
		Name:          "GETBIT",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "GETBIT key offset",
	}

	RedisCommandGETRANGE = RedisCommand{
		Name:          "GETRANGE",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "GETRANGE key start end",
	}

	RedisCommandGETSET = RedisCommand{
		Name:          "GETSET",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "GETSET key value",
	}

	RedisCommandINCR = RedisCommand{
		Name:          "INCR",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "INCR key",
	}

	RedisCommandINCRBY = RedisCommand{
		Name:          "INCRBY",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "INCRBY key increment",
	}

	// INCRBYFLOAT command is not implemented in LedisDB.

	RedisCommandMGET = RedisCommand{
		Name:          "MGET",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsFromIndex(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "MGET key [key ...]",
	}

	RedisCommandMSET = RedisCommand{
		Name:          "MSET",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsFromIndex(0, 1),
		TransformFunc: NoneTransformer(),
		Syntax:        "MSET key value [key value ...]",
	}

	// MSETNX command is not implemented in LedisDB.

	// PSETEX command is not implemented in LedisDB.

	RedisCommandSET = RedisCommand{
		Name:          "SET",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: SetCommandTransformer,
		Syntax:        "SET key value [expiration EX seconds|PX milliseconds] [NX|XX]",
	}

	RedisCommandSETBIT = RedisCommand{
		Name:          "SETBIT",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SETBIT key offset value",
	}

	RedisCommandSETEX = RedisCommand{
		Name:          "SETEX",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SETEX key seconds value",
	}

	RedisCommandSETNX = RedisCommand{
		Name:          "SETNX",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SETNX key value",
	}

	RedisCommandSETRANGE = RedisCommand{
		Name:          "SETRANGE",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SETRANGE key offset value",
	}

	RedisCommandSTRLEN = RedisCommand{
		Name:          "STRLEN",
		KeyType:       RedisTypeString,
		KeyExtractor:  ArgsAtIndices(0),
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
		Name:          "HDEL",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HDEL key field [field ...]",
	}

	RedisCommandHEXISTS = RedisCommand{
		Name:          "HEXISTS",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HEXISTS key field",
	}

	RedisCommandHGET = RedisCommand{
		Name:          "HGET",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HGET key field",
	}

	RedisCommandHGETALL = RedisCommand{
		Name:          "HGETALL",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HGETALL key",
	}

	RedisCommandHINCRBY = RedisCommand{
		Name:          "HINCRBY",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HINCRBY key field increment",
	}

	// HINCRBYFLOAT command is not implemented in LedisDB.

	RedisCommandHKEYS = RedisCommand{
		Name:          "HKEYS",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HKEYS key",
	}

	RedisCommandHLEN = RedisCommand{
		Name:          "HLEN",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HLEN key",
	}

	RedisCommandHMGET = RedisCommand{
		Name:          "HMGET",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HMGET key field [field ...]",
	}

	RedisCommandHMSET = RedisCommand{
		Name:          "HMSET",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HMSET key field value [field value ...]",
	}

	RedisCommandHSET = RedisCommand{
		Name:          "HSET",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HSET key field value",
	}

	// HSETNX command is not implemented in LedisDB.

	// HSTRLEN command is not implemented in LedisDB.

	RedisCommandHVALS = RedisCommand{
		Name:          "HVALS",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HVALS key",
	}

	RedisCommandHSCAN = RedisCommand{
		Name:          "HSCAN",
		KeyType:       RedisTypeHash,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "HSCAN key cursor [MATCH pattern] [COUNT count]",
	}
)

// RedisCommand variables describing the Redis commands operating on
// lists (RedisTypeList).
//
//     https://redis.io/commands#list
var (
	RedisCommandBLPOP = RedisCommand{
		Name:          "BLPOP",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsFromUntilIndex(0, -1),
		TransformFunc: NoneTransformer(),
		Syntax:        "BLPOP key [key ...] timeout",
	}

	RedisCommandBRPOP = RedisCommand{
		Name:          "BRPOP",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsFromUntilIndex(0, -1),
		TransformFunc: NoneTransformer(),
		Syntax:        "BRPOP key [key ...] timeout",
	}

	RedisCommandBRPOPLPUSH = RedisCommand{
		Name:          "BRPOPLPUSH",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsAtIndices(0, 1),
		TransformFunc: NoneTransformer(),
		Syntax:        "BRPOPLPUSH source destination timeout",
	}

	RedisCommandLINDEX = RedisCommand{
		Name:          "LINDEX",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "LINDEX key index",
	}

	// LINSERT command is not implemented in LedisDB.

	RedisCommandLLEN = RedisCommand{
		Name:          "LLEN",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "LLEN key index",
	}

	RedisCommandLPOP = RedisCommand{
		Name:          "LPOP",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "LPOP key",
	}

	RedisCommandLRANGE = RedisCommand{
		Name:          "LRANGE",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "LRANGE key start stop",
	}

	RedisCommandLPUSH = RedisCommand{
		Name:          "LPUSH",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "LPUSH key value [value ...]",
	}

	// LPUSHX command is not implemented in LedisDB.

	// LREM command is not implemented in LedisDB.

	// LSET command is not implemented in LedisDB.

	RedisCommandLTRIM = RedisCommand{
		Name:          "LTRIM",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "LTRIM key start stop",
	}

	RedisCommandRPOP = RedisCommand{
		Name:          "RPOP",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "RPOP key",
	}

	RedisCommandRPOPLPUSH = RedisCommand{
		Name:          "RPOPLPUSH",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsAtIndices(0, 1),
		TransformFunc: NoneTransformer(),
		Syntax:        "RPOPLPUSH source destination",
	}

	RedisCommandRPUSH = RedisCommand{
		Name:          "RPUSH",
		KeyType:       RedisTypeList,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "RPUSH key value [value ...]",
	}

	// RPUSHX command is not implemented in LedisDB.
)

// RedisCommand variables describing the Redis commands operating on
// sets (RedisTypeSet).
//
//     https://redis.io/commands#set
var (
	RedisCommandSADD = RedisCommand{
		Name:          "SADD",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SADD key member [member ...]",
	}

	RedisCommandSCARD = RedisCommand{
		Name:          "SCARD",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SCARD key",
	}

	RedisCommandSDIFF = RedisCommand{
		Name:          "SDIFF",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsFromIndex(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SDIFF key [key ...]",
	}

	RedisCommandSDIFFSTORE = RedisCommand{
		Name:          "SDIFFSTORE",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsFromIndex(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SDIFFSTORE destination key [key ...]",
	}

	RedisCommandSINTER = RedisCommand{
		Name:          "SINTER",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsFromIndex(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SINTER key [key ...]",
	}

	RedisCommandSINTERSTORE = RedisCommand{
		Name:          "SINTERSTORE",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsFromIndex(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SINTERSTORE destination key [key ...]",
	}

	RedisCommandSISMEMBER = RedisCommand{
		Name:          "SISMEMBER",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SISMEMBER key member",
	}

	RedisCommandSMEMBERS = RedisCommand{
		Name:          "SMEMBERS",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SMEMBERS key",
	}

	// SMOVE command is not implemented in LedisDB.

	// SPOP command is not implemented in LedisDB.

	// SRANDMEMBER command is not implemented in LedisDB.

	RedisCommandSREM = RedisCommand{
		Name:          "SREM",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SREM key member [member ...]",
	}

	RedisCommandSSCAN = RedisCommand{
		Name:          "SSCAN",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SSCAN key cursor [MATCH pattern] [COUNT count]",
	}

	RedisCommandSUNION = RedisCommand{
		Name:          "SUNION",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsFromIndex(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SUNION key [key ...]",
	}

	RedisCommandSUNIONSTORE = RedisCommand{
		Name:          "SUNIONSTORE",
		KeyType:       RedisTypeSet,
		KeyExtractor:  ArgsFromIndex(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "SUNIONSTORE destination key [key ...]",
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
		Name:          "ZADD",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: ZaddCommandTransformer,
		Syntax:        "ZADD key [NX|XX] [CH] [INCR] score member [score member ...]",
	}

	RedisCommandZCARD = RedisCommand{
		Name:          "ZCARD",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZCARD key",
	}

	RedisCommandZCOUNT = RedisCommand{
		Name:          "ZCOUNT",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZCOUNT key min max",
	}

	RedisCommandZINCRBY = RedisCommand{
		Name:          "ZINCRBY",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZINCRBY key increment member",
	}

	// RedisCommandZINTERSTORE contains information about the ZINTERSTORE Redis
	// command.
	//
	// The KeyExtractor only extracts the destination key. Further keys could
	// be extracted by introducing a new KeyExtractor implementation.
	// However, Redis interprets non-existing key as empty keys.
	RedisCommandZINTERSTORE = RedisCommand{
		Name:          "ZINTERSTORE",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZINTERSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]",
	}

	RedisCommandZLEXCOUNT = RedisCommand{
		Name:          "ZLEXCOUNT",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZLEXCOUNT key min max",
	}

	// ZPOPMAX command is not implemented in LedisDB.

	// ZPOPMIN command is not implemented in LedisDB.

	RedisCommandZRANGE = RedisCommand{
		Name:          "ZRANGE",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZRANGE key start stop [WITHSCORES]",
	}

	RedisCommandZRANGEBYLEX = RedisCommand{
		Name:          "ZRANGEBYLEX",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZRANGEBYLEX key min max [LIMIT offset count]",
	}

	RedisCommandZRANGEBYSCORE = RedisCommand{
		Name:          "ZRANGEBYSCORE",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]",
	}

	RedisCommandZRANK = RedisCommand{
		Name:          "ZRANK",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZRANK key member",
	}

	RedisCommandZREM = RedisCommand{
		Name:          "ZREM",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZREM key member [member ...]",
	}

	RedisCommandZREMRANGEBYLEX = RedisCommand{
		Name:          "ZREMRANGEBYLEX",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZREMRANGEBYLEX key min max",
	}

	RedisCommandZREMRANGEBYRANK = RedisCommand{
		Name:          "ZREMRANGEBYRANK",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZREMRANGEBYRANK key start stop",
	}

	RedisCommandZREMRANGEBYSCORE = RedisCommand{
		Name:          "ZREMRANGEBYSCORE",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZREMRANGEBYSCORE key min max",
	}

	RedisCommandZREVRANGE = RedisCommand{
		Name:          "ZREVRANGE",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZREVRANGE key start stop [WITHSCORES]",
	}

	// ZREVRANGEBYLEX command is not implemented in LedisDB.

	RedisCommandZREVRANGEBYSCORE = RedisCommand{
		Name:          "ZREVRANGEBYSCORE",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZREVRANGEBYSCORE key max min [WITHSCORES] [LIMIT offset count]",
	}

	RedisCommandZREVRANK = RedisCommand{
		Name:          "ZREVRANK",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZREVRANK key member",
	}

	RedisCommandZSCAN = RedisCommand{
		Name:          "ZSCAN",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZSCAN key cursor [MATCH pattern] [COUNT count]",
	}

	RedisCommandZSCORE = RedisCommand{
		Name:          "ZSCORE",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZSCORE key member",
	}

	// RedisCommandZUNIONSTORE contains information about the ZUNIONSTORE Redis
	// command.
	//
	// The KeyExtractor only extracts the destination key. Further keys could
	// be extracted by introducing a new KeyExtractor implementation.
	// However, Redis interprets non-existing key as empty keys.
	RedisCommandZUNIONSTORE = RedisCommand{
		Name:          "ZUNIONSTORE",
		KeyType:       RedisTypeZSet,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: NoneTransformer(),
		Syntax:        "ZUNIONSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]",
	}
)

// RedisCommand variables describing the Redis commands operating on
// keys / generics (RedisTypeGeneric).
//
//     https://redis.io/commands#generic
var (
	RedisCommandDEL = RedisCommand{
		Name:         "DEL",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: ArgsFromIndex(0),
		TransformFunc: TypeSpecificBulkTransformer(&TypeSpecificBulkTransformerConfig{
			Commands: TypeSpecificCommands{
				KV:   "DEL",
				List: "LMCLEAR",
				Hash: "HMCLEAR",
				Set:  "SMCLEAR",
				ZSet: "ZMCLEAR",
			},
			Aggregation: AggregationSum,
		}),
		Syntax: "DEL key [key ...]",
	}

	RedisCommandDUMP = RedisCommand{
		Name:         "DUMP",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: ArgsAtIndices(0),
		TransformFunc: TypeSpecificBulkTransformer(&TypeSpecificBulkTransformerConfig{
			Commands: TypeSpecificCommands{
				KV:   "DUMP",
				List: "LDUMP",
				Hash: "HDUMP",
				Set:  "SDUMP",
				ZSet: "ZDUMP",
			},
			Aggregation: AggregationFirst,
		}),
		Syntax: "DUMP key",
	}

	RedisCommandEXISTS = RedisCommand{
		Name:         "EXISTS",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: ArgsAtIndices(0),
		TransformFunc: TypeSpecificBulkTransformer(&TypeSpecificBulkTransformerConfig{
			Commands: TypeSpecificCommands{
				KV:   "EXISTS",
				List: "LKEYEXISTS",
				Hash: "HKEYEXISTS",
				Set:  "SKEYEXISTS",
				ZSet: "ZKEYEXISTS",
			},
			Debulk:      true,
			Aggregation: AggregationSum,
		}),
		Syntax: "EXISTS key [key ...]",
	}

	RedisCommandEXPIRE = RedisCommand{
		Name:         "EXPIRE",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: ArgsAtIndices(0),
		TransformFunc: TypeSpecificBulkTransformer(&TypeSpecificBulkTransformerConfig{
			Commands: TypeSpecificCommands{
				KV:   "EXPIRE",
				List: "LEXPIRE",
				Hash: "HEXPIRE",
				Set:  "SEXPIRE",
				ZSet: "ZEXPIRE",
			},
			Aggregation:         AggregationSum,
			AppendArgsExtractor: ArgsAtIndices(1),
		}),
		Syntax: "EXPIRE key seconds",
	}

	RedisCommandEXPIREAT = RedisCommand{
		Name:         "EXPIREAT",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: ArgsAtIndices(0),
		TransformFunc: TypeSpecificBulkTransformer(&TypeSpecificBulkTransformerConfig{
			Commands: TypeSpecificCommands{
				KV:   "EXPIREAT",
				List: "LEXPIREAT",
				Hash: "HEXPIREAT",
				Set:  "SEXPIREAT",
				ZSet: "ZEXPIREAT",
			},
			Aggregation:         AggregationSum,
			AppendArgsExtractor: ArgsAtIndices(1),
		}),
		Syntax: "EXPIREAT key timestamp",
	}

	// KEYS command is not implemented in LedisDB.

	// MIGRATE command is not implemented in LedisDB.

	// MOVE command is not implemented in LedisDB.

	// OBJECT command is not implemented in LedisDB.

	RedisCommandPERSIST = RedisCommand{
		Name:         "PERSIST",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: ArgsAtIndices(0),
		TransformFunc: TypeSpecificBulkTransformer(&TypeSpecificBulkTransformerConfig{
			Commands: TypeSpecificCommands{
				KV:   "PERSIST",
				List: "LPERSIST",
				Hash: "HPERSIST",
				Set:  "SPERSIST",
				ZSet: "ZPERSIST",
			},
			Aggregation: AggregationSum,
		}),
		Syntax: "PERSIST key",
	}

	// PEXPIRE command is not implemented in LedisDB.

	// PEXPIREAT command is not implemented in LedisDB.

	// PTTL command is not implemented in LedisDB.

	// RANDOMKEY command is not implemented in LedisDB.

	// RENAME command is not implemented in LedisDB.

	// RENAMENX command is not implemented in LedisDB.

	RedisCommandRESTORE = RedisCommand{
		Name:          "RESTORE",
		KeyType:       RedisTypeGeneric,
		KeyExtractor:  ArgsAtIndices(0),
		TransformFunc: RestoreCommandTransformer,
		Syntax:        "RESTORE key ttl serialized-value [REPLACE] [ABSTTL] [IDLETIME seconds] [FREQ frequency]",
	}

	RedisCommandSORT = RedisCommand{
		Name:         "SORT",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: ArgsAtIndices(0),
		TransformFunc: TypeSpecificBulkTransformer(&TypeSpecificBulkTransformerConfig{
			Commands: TypeSpecificCommands{
				List: "XLSORT",
				Set:  "XSSORT",
				ZSet: "XZSORT",
			},
			Aggregation: AggregationFirst,
		}),
		Syntax: "SORT key [BY pattern] [LIMIT offset count] [GET pattern [GET pattern ...]] [ASC|DESC] [ALPHA] [STORE destination]",
	}

	// TOUCH command is not implemented in LedisDB.

	RedisCommandTTL = RedisCommand{
		Name:         "TTL",
		KeyType:      RedisTypeGeneric,
		KeyExtractor: ArgsAtIndices(0),
		TransformFunc: TypeSpecificBulkTransformer(&TypeSpecificBulkTransformerConfig{
			Commands: TypeSpecificCommands{
				KV:   "TTL",
				List: "LTTL",
				Hash: "HTTL",
				Set:  "STTL",
				ZSet: "ZTTL",
			},
			Aggregation: AggregationSum,
		}),
		Syntax: "TTL key",
	}

	// TYPE command is not implemented in LedisDB.

	// UNLINK command is not implemented in LedisDB.

	// WAIT command is not implemented in LedisDB.

	// SCAN command is not implemented in LedisDB.
	// However XSCAN is available. By adding an additional caching structure
	// to Rewriter, SCAN could be emulated in most circumstances. XSCAN
	// requires specifying the type of the keyspace to be scanned in addition
	// to the parameters required by Redis.
)

// RedisCommand variables describing the Redis commands for altering server
// configuration and perform server-side, global actions.
//
//     https://redis.io/commands#server
var (
// Redis

// BGREWRITEAOF
// BGSAVE
// CLIENT ID
// CLIENT KILL [ip:port] [ID client-id] [TYPE normal|master|slave|pubsub] [ADDR ip:port] [SKIPME yes/no]
// CLIENT LIST [TYPE normal|master|replica|pubsub]
// CLIENT GETNAME
// CLIENT PAUSE timeout
// CLIENT REPLY ON|OFF|SKIP
// CLIENT SETNAME connection-name
// CLIENT UNBLOCK client-id [TIMEOUT|ERROR]
// COMMAND
// COMMAND COUNT
// COMMAND GETKEYS
// COMMAND INFO command-name [command-name ...]
// CONFIG GET parameter
// CONFIG REWRITE
// CONFIG SET parameter value
// CONFIG RESETSTAT
// DBSIZE
// DEBUG OBJECT key
// DEBUG SEGFAULT
// FLUSHALL [ASYNC]
// FLUSHDB [ASYNC]
// INFO [section]
// LASTSAVE
// MEMORY DOCTOR
// MEMORY HELP
// MEMORY MALLOC-STATS
// MEMORY PURGE
// MEMORY STATS
// MEMORY USAGE key [SAMPLES count]
// MONITOR
// ROLE
// SAVE
// SHUTDOWN [NOSAVE|SAVE]
// SLAVEOF host port
// REPLICAOF host port
// SLOWLOG subcommand [argument]
// SYNC
// TIME

// LedisDB (Server & Replication Sections)

// CONFIG REWRITE
// FLUSHALL
// FLUSHDB
// FULLSYNC [NEW]
// INFO [section]
// ROLE
// SLAVEOF host port [RESTART] [READONLY]
// SYNC logid
// TIME
)

// RedisCommand variables describing the Redis commands for managing the
// connection: Utilities for probing and altering connection properties.
//
//     https://redis.io/commands#connection
var (
	RedisCommandAUTH = RedisCommand{
		Name:          "AUTH",
		KeyType:       RedisTypeGeneric,
		KeyExtractor:  ArgsAtIndices(),
		TransformFunc: NoneTransformer(),
		Syntax:        "AUTH password",
	}

	RedisCommandECHO = RedisCommand{
		Name:          "ECHO",
		KeyType:       RedisTypeGeneric,
		KeyExtractor:  ArgsAtIndices(),
		TransformFunc: NoneTransformer(),
		Syntax:        "ECHO message",
	}

	RedisCommandPING = RedisCommand{
		Name:          "PING",
		KeyType:       RedisTypeGeneric,
		KeyExtractor:  ArgsAtIndices(),
		TransformFunc: PingCommandTransformer,
		Syntax:        "PING [message]",
	}

	// QUIT command is not implemented in LedisDB.

	RedisCommandSELECT = RedisCommand{
		Name:          "SELECT",
		KeyType:       RedisTypeGeneric,
		KeyExtractor:  ArgsAtIndices(),
		TransformFunc: NoneTransformer(),
		Syntax:        "SELECT index",
	}

	// SWAPDB command is not implemented in LedisDB.
)

// RedisCommand variables describing the Redis commands for working with lua
// scripts.
//
//     https://redis.io/commands#scripting
var (
	RedisCommandEVAL = RedisCommand{
		Name:          "EVAL",
		KeyType:       RedisTypeGeneric,
		KeyExtractor:  ArgsAtIndices(),
		TransformFunc: NoneTransformer(),
		Syntax:        "EVAL script numkeys key [key ...] arg [arg ...]",
	}

	RedisCommandEVALSHA = RedisCommand{
		Name:          "EVALSHA",
		KeyType:       RedisTypeGeneric,
		KeyExtractor:  ArgsAtIndices(),
		TransformFunc: NoneTransformer(),
		Syntax:        "EVALSHA sha1 numkeys key [key ...] arg [arg ...]",
	}

	RedisCommandSCRIPT = RedisCommand{
		Name:          "SCRIPT",
		KeyType:       RedisTypeGeneric,
		KeyExtractor:  ArgsAtIndices(),
		TransformFunc: ScriptCommandTransformer,
		Syntax:        "SCRIPT subcommand [arg ...]",
	}
)

// RedisCommand variable describing the rewledis specific UNSAFE command.
// Similarly to the homonymous Go package: Do not use this unless you know
// what you are doing.
var (
	RedisCommandUNSAFE = RedisCommand{
		Name:          "UNSAFE",
		KeyType:       RedisTypeGeneric,
		KeyExtractor:  ArgsAtIndices(),
		TransformFunc: UnsafeCommandTransformer,
		Syntax:        "UNSAFE subcommand [arg ...]",
	}
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
	case "HSCAN":
		return &RedisCommandHSCAN, nil
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
	case "SSCAN":
		return &RedisCommandSSCAN, nil
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
	case "ZSCAN":
		return &RedisCommandZSCAN, nil
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
	case "RESTORE":
		return &RedisCommandRESTORE, nil
	case "SORT":
		return &RedisCommandSORT, nil
	case "TTL":
		return &RedisCommandTTL, nil
	case "AUTH":
		return &RedisCommandAUTH, nil
	case "ECHO":
		return &RedisCommandECHO, nil
	case "PING":
		return &RedisCommandPING, nil
	case "SELECT":
		return &RedisCommandSELECT, nil
	case "EVAL":
		return &RedisCommandEVAL, nil
	case "EVALSHA":
		return &RedisCommandEVALSHA, nil
	case "SCRIPT":
		return &RedisCommandSCRIPT, nil
	case "UNSAFE":
		return &RedisCommandUNSAFE, nil
	default:
		return nil, ErrUnknownRedisCommandName
	}
}

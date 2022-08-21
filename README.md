# rewledis

[![Project Status: Abandoned – Initial development has started, but there has not yet been a stable, usable release; the project has been abandoned and the author(s) do not intend on continuing development.](https://www.repostatus.org/badges/latest/abandoned.svg)](https://www.repostatus.org/#abandoned)
[![Go Reference](https://pkg.go.dev/badge/github.com/pskopnik/rewledis.svg)](https://pkg.go.dev/github.com/pskopnik/rewledis)

`rewledis` provides wrappers for the [redigo][redigo] [Redis][redis] client library and transparently rewrites Redis commands to [LedisDB][ledisdb] commands.
Existing applications developed to target Redis are thus enabled to use LedisDB without changing any code.

The simplest entrypoint to this module is `rewledis.NewPool()`, which returns a `redis.Pool` to be treated (almost) like a regular pool processing Redis commands.
Commands are internally rewritten to work on LedisDB.

The motivation for developing this package was that I wanted to use an established Work-Queue library which targeted Redis, but instead use LedisDB embedded in one of the application's components.
Transparently rewriting commands on each client would serve as the compatibility layer between the library expecting to communicate with Redis and the LedisDB instance.

**Caveats**

 * LedisDB seems to be unmaintained.
 * LedisDB's memory model/transaction models differs from Redis.

**This project has been discontinued.**
During development and testing I ran into problems related to the execution of scripts:
On Redis these are executed as a single unit/transaction while LedisDB executes each command in the script individually.
This made LedisDB unsuitable for the use case I had in mind.
I abandoned the project in a half-ready state, so some things may work while others don't.

[redis]: https://redis.io/
[ledisdb]: https://github.com/ledisdb/ledisdb
[redigo]: https://github.com/gomodule/redigo

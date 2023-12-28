# gRedis
gRedis is a Go implementation Redis server. It implemented [RESP](https://redis.io/docs/reference/protocol-spec/) for communication with standard Redis Clients.

## How to run
gRedis server:
```bash
go run main.go
```
Usage for flag options:
```bash
Usage of ./gRedis:
  -config string
        Set a config file
  -host string
        Set a server host to listen (default "127.0.0.1")
  -logdir string
        Set a log directory (default "./")
  -loglevel string
        Set a log level (default "info")
  -port int
        Set a server prot to listen (default 6379)
  -segnum int
        Set a segmentation number for cache database (default 100)
```

## Communication with gRedis server
use `redis-cli` to communicate with gRedis server.
```bash
% redis-cli 
127.0.0.1:6379> EXISTS mykey
(integer) 0
127.0.0.1:6379> APPEND mykey "Hello"
(integer) 5
127.0.0.1:6379> APPEND mykey " World"
(integer) 11
127.0.0.1:6379> GET mykey
"Hello World"
127.0.0.1:6379> 
```

## Support Redis Commands(Unfinished...)
You can find usage for [Redis Commands](https://redis.io/commands/). All commands below are supported.
| key     | string      | hash         |
|---------|-------------|--------------|
| del     | set         | hdel         |
| exists  | get         | hexists      |
| keys    | getrange    | hget         |
| expire  | setrange    | hgetall      |
| persist | mget        | hincrby      |
| ttl     | mset        | hincrbyfloat |
| rename  | setex       | hkeys        |
|         | setnx       | hlen         |
|         | strlen      | hmget        |
|         | incr        | hmset        |
|         | incrby      | hset         |
|         | decr        | hsetnx       |
|         | decrby      | hvals        |
|         | incrbyfloat | hstrlen      |
|         | append      | hrandfield   |

## Todo
+ [] list, set, hash, channel, sorted set, stream commands
+ [] Cluster Mode
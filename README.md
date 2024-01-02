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

## Support Redis Commands
You can find usage for [Redis Commands](https://redis.io/commands/). All commands below are supported.
| key     | string      | hash         | list    | set         | 
|---------|-------------|--------------|---------|-------------|   
| del     | set         | hdel         | lindex  | sadd        |
| exists  | get         | hexists      | linsert | scard       |
| keys    | getrange    | hget         | llen    | sdiff       |
| expire  | setrange    | hgetall      | lmove   | sdiffstore  | 
| persist | mget        | hincrby      | lpop    | sinter      |
| ttl     | mset        | hincrbyfloat | lpos    | sinterstore |
| rename  | setex       | hkeys        | lpush   | sismember   |
|         | setnx       | hlen         | lpushx  | smembers    |
|         | strlen      | hmget        | lrange  | smove       |
|         | incr        | hmset        | lrem    | spop        |
|         | incrby      | hset         | lset    | srandmember |
|         | decr        | hsetnx       | ltrim   | srem        |
|         | decrby      | hvals        | rpop    | sunion      |
|         | incrbyfloat | hstrlen      | rpush   | sunionstore | 
|         | append      | hrandfield   | rpushx  |             |

## Todo
+ [] Channel, sorted set commands
+ [] Cluster Mode
+ [] RDB, AOF (data persistence)
+ [] Testings
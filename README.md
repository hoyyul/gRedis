# gRedis
gRedis is a Go implementation Redis server. It implemented [RESP](https://redis.io/docs/reference/protocol-spec/) for communication with standard Redis Clients.

## How to run
gRedis server:
```bash
go run main.go
```
Usage for flag options:
```bash
go run main.go
Usage of ./gRedis:
  -config string
        Select a config file
  -host string
        Bind a server host (default "127.0.0.1")
  -logdir string
        Set log directory (default "./")
  -loglevel string
        Set log level (default "info")
  -port int
        Bind a server port (default 6379)
  -segnum int
        Set database number (default 100)
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
| key     | string      | 
|---------|-------------|
| del     | set         |
| exists  | get         |
| keys    | getrange    |
| expire  | setrange    |
| persist | mget        |
| ttl     | mset        |
| rename  | setex       |
|         | setnx       |
|         | strlen      |
|         | incr        |
|         | incrby      |
|         | decr        | 
|         | decrby      | 
|         | incrbyfloat | 
|         | append      | 

## Todo
+ [] list, set, hash, channel, sorted set, stream commands
+ [] Cluster Mode
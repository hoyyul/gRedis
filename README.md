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
  -dbnum int
        Set database number for cache storage (default 16)
  -host string
        Set a server host to listen (default "127.0.0.1")
  -logdir string
        Set a log directory (default "./")
  -loglevel string
        Set a log level (default "info")
  -port int
        Set a server prot to listen (default 6379)
  -segnum int
        Set a segmentation number for a cache database (default 100)
```

## Communication with gRedis server
use `redis-cli` to communicate with gRedis server.
```bash
% redis-cli
127.0.0.1:6379> ping
PONG
127.0.0.1:6379> select 1
OK
127.0.0.1:6379[1]> MSET firstname Jack lastname Stuntman age 35
OK
127.0.0.1:6379[1]> KEYS *name*
1) "firstname"
2) "lastname"
127.0.0.1:6379[1]> KEYS a??
1) "age"
127.0.0.1:6379[1]> LPUSH mylist 1 2 3 4 5
(integer) 5
127.0.0.1:6379[1]> LRANGE mylist 0 -1
1) "5"
2) "4"
3) "3"
4) "2"
5) "1"
127.0.0.1:6379[1]> type mylist
list
127.0.0.1:6379[1]> EXPIRE mylist 20
(integer) 1
127.0.0.1:6379[1]> ttl mylist
(integer) 13
127.0.0.1:6379[1]> ttl mylist
(integer) 0
127.0.0.1:6379[1]> LRANGE mylist 0 -1
(nil)
127.0.0.1:6379[1]> select 0
OK
127.0.0.1:6379> KEYS *name*
(empty array)
```

## Benchmark
You can find details about the benchmark tool in [redis-benchmark](https://redis.io/docs/management/optimization/benchmarks/).
The testing is conducted on MacBook Pro 2019 with 2.6 GHz 6-Core Intel Core i7 processor, 32.0 GB RAM, and macOS Ventura.

On average, we can say gRedis can reach **85-90%** of the orginal C redis performance.

`redis-benchmark -c 50 -n 200000 -t [get|set|...] -q -p 6379`

gRedis:
```text
SET: 88613.20 requests per second, p50=0.279 msec                   
GET: 92592.59 requests per second, p50=0.271 msec                   
INCR: 89686.10 requests per second, p50=0.279 msec                   
LPUSH: 92165.90 requests per second, p50=0.271 msec                   
RPUSH: 92421.44 requests per second, p50=0.279 msec                   
LPOP: 92936.80 requests per second, p50=0.271 msec                   
RPOP: 92936.80 requests per second, p50=0.271 msec                   
SADD: 91785.23 requests per second, p50=0.279 msec                   
HSET: 90212.00 requests per second, p50=0.279 msec                   
SPOP: 93153.23 requests per second, p50=0.271 msec                   
MSET (10 keys): 73502.39 requests per second, p50=0.319 msec   
```

Redis:
```text
SET: 92293.49 requests per second, p50=0.279 msec                   
GET: 81833.06 requests per second, p50=0.271 msec                   
INCR: 96246.39 requests per second, p50=0.271 msec                   
LPUSH: 94876.66 requests per second, p50=0.271 msec                   
RPUSH: 95831.34 requests per second, p50=0.271 msec                   
LPOP: 96571.70 requests per second, p50=0.271 msec                   
RPOP: 96946.20 requests per second, p50=0.271 msec                   
SADD: 96432.02 requests per second, p50=0.263 msec                   
HSET: 97181.73 requests per second, p50=0.263 msec                    
SPOP: 98570.72 requests per second, p50=0.263 msec                    
MSET (10 keys): 95147.48 requests per second, p50=0.295 msec    
```

Rough comparation:
```text
SET: 88613.20 / 92293.49 = 0.9612（96.12%）
GET: 92592.59 / 81833.06 = 1.1319（113.19%）
INCR: 89686.10 / 96246.39 = 0.9314（93.14%）
LPUSH: 92165.90 / 94876.66 = 0.9715（97.15%）
RPUSH: 92421.44 / 95831.34 = 0.9644（96.44%）
LPOP: 92936.80 / 96571.70 = 0.9625（96.25%）
RPOP: 92936.80 / 96946.20 = 0.9591（95.91%）
SADD: 91785.23 / 96432.02 = 0.9517（95.17%）
HSET: 90212.00 / 97181.73 = 0.9285（92.85%）
SPOP: 93153.23 / 98570.72 = 0.9451（94.51%）
MSET (10 keys): 73502.39 / 95147.48 = 0.7713（77.13%）

(0.9612 + 1.1319 + 0.9314 + 0.9715 + 0.9644 
+ 0.9625 + 0.9591 + 0.9517 + 0.9285 + 0.9451 + 0.7713) / 11 ≈ 0.9462(94.62%)
```


## Support Redis Commands
You can find usage for [Redis Commands](https://redis.io/commands/). All commands below are supported.
| key     | string      | hash         | list    | set         | general |
|---------|-------------|--------------|---------|-------------|---------|   
| del     | set         | hdel         | lindex  | sadd        | select  |
| exists  | get         | hexists      | linsert | scard       |         |
| keys    | getrange    | hget         | llen    | sdiff       |         |
| expire  | setrange    | hgetall      | lmove   | sdiffstore  |         |
| persist | mget        | hincrby      | lpop    | sinter      |         |
| ttl     | mset        | hincrbyfloat | lpos    | sinterstore |         |
| rename  | setex       | hkeys        | lpush   | sismember   |         |
| type    | setnx       | hlen         | lpushx  | smembers    |         |
|         | strlen      | hmget        | lrange  | smove       |         |
|         | incr        | hmset        | lrem    | spop        |         |
|         | incrby      | hset         | lset    | srandmember |         |
|         | decr        | hsetnx       | ltrim   | srem        |         |
|         | decrby      | hvals        | rpop    | sunion      |         |
|         | incrbyfloat | hstrlen      | rpush   | sunionstore |         |
|         | append      | hrandfield   | rpushx  |             |         |

## Todo
+ [] Channel, sorted set commands
+ [] Cluster Mode
+ [] RDB, AOF (data persistence)
+ [] Testings
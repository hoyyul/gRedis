package memdb

type MemDb struct {
	dict    *ConcurrentMap // memory cache db
	expires *ConcurrentMap // all ttl keys
}

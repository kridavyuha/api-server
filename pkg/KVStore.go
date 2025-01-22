package kvstore

type KVStore interface {
	Get(key string) (string, error)
	Set(key string, value interface{}) error
	Delete(key string) error
	LPush(key string, values ...interface{}) error
	RPush(key string, values ...interface{}) error
	LPop(key string) (string, error)
	RPop(key string) (string, error)
	LLen(key string) (int64, error)
	LIndex(key string, index int64) (string, error)
	LRange(key string, start, stop int64) ([]string, error)
	INCR(key string) (int64, error)
	DECR(key string) (int64, error)
	LRem(key string, count int64, value interface{}) error
	Keys(pattern string) ([]string, error)
	Del(keys ...string) error
	HSet(key, field string, value interface{}) error
	HGet(key, field string) (string, error)
	HGetAll(key string) (map[string]string, error)
}

package cache

import (
	"RedditDownloaderBot/pkg/reddit"
	"RedditDownloaderBot/pkg/util"
	"context"
	"encoding/json"
	"github.com/go-faster/errors"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var _ Interface = RedisCache{}

// Define the prefixes of keys
const redisMediaCachePrefix = "media:"
const redisAlbumCachePrefix = "album:"

// RedisCache satisfies Interface backed by a Redis server
type RedisCache struct {
	ttl    time.Duration
	client *redis.Client
}

// NewRedisCache will create a new redis cache
func NewRedisCache(address, password string, ttl time.Duration) (RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
	})
	return RedisCache{
		client: rdb,
		ttl:    ttl,
	}, rdb.Ping(context.Background()).Err()
}

func (r RedisCache) SetMediaCache(key string, value CallbackDataCached) error {
	return r.client.
		Set(context.Background(), redisMediaCachePrefix+key, util.ToJsonString(value), r.ttl).
		Err()
}

func (r RedisCache) GetAndDeleteMediaCache(key string) (CallbackDataCached, error) {
	return parseRedisJson[CallbackDataCached](
		r.client.
			GetDel(context.Background(), redisMediaCachePrefix+key).
			Result(),
	)
}

func (r RedisCache) SetAlbumCache(key string, value reddit.FetchResultAlbum) error {
	return r.client.
		Set(context.Background(), redisAlbumCachePrefix+key, util.ToJsonString(value), r.ttl).
		Err()
}

func (r RedisCache) GetAndDeleteAlbumCache(key string) (reddit.FetchResultAlbum, error) {
	return parseRedisJson[reddit.FetchResultAlbum](
		r.client.
			GetDel(context.Background(), redisAlbumCachePrefix+key).
			Result(),
	)
}

func (r RedisCache) Close() error {
	return r.client.Close()
}

// parseRedisJson will get the value of a key which is in redis + the error message of redis.
// Then, it tries to parse the data as the generic struct passed to it.
func parseRedisJson[T any](val string, err error) (T, error) {
	// Check errors
	var result T
	if err == redis.Nil {
		return result, NotFoundErr
	} else if err != nil {
		return result, errors.Wrap(err, "Unable to fetch data from Redis")
	}
	// Parse the json
	err = json.NewDecoder(strings.NewReader(val)).Decode(&result)
	if err != nil {
		return result, errors.Wrap(err, "Unable to parse JSON")
	}
	return result, nil
}

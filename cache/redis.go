package cache

import (
	"context"
	"encoding/json"
	"github.com/HirbodBehnam/RedditDownloaderBot/reddit"
	"github.com/HirbodBehnam/RedditDownloaderBot/util"
	"github.com/go-faster/errors"
	"github.com/go-redis/redis/v9"
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

func parseRedisJson[T any](val string, err error) (T, error) {
	// Check errors
	var result T
	if err == redis.Nil {
		return result, NotFoundErr
	} else if err != nil {
		return result, err
	}
	// Parse the json
	err = json.NewDecoder(strings.NewReader(val)).Decode(&result)
	if err != nil {
		return result, errors.Wrap(err, "cannot parse json")
	}
	return result, nil
}

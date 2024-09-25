package cache

import (
	"github.com/go-faster/errors"
)

// NotFoundErr will be returned if the key does not exist in database
var NotFoundErr = errors.New("key not found")

// Interface provides the interface to interface with a cache
type Interface interface {
	// SetMediaCache sets a key and overwrite any old entries in cache. It should be used for
	// storing media data in cache
	SetMediaCache(key string, value CallbackDataCached) error
	// GetAndDeleteMediaCache will atomically get a media cache and delete it from cache.
	// If it does not exist, returns NotFoundErr as error
	GetAndDeleteMediaCache(key string) (CallbackDataCached, error)
	// SetAlbumCache sets a key and overwrite any old entries in cache. It should be used for
	// storing album data in cache
	SetAlbumCache(key string, value CallbackAlbumCached) error
	// GetAndDeleteAlbumCache will atomically get an album cache and delete it from cache.
	// If it does not exist, returns NotFoundErr as error
	GetAndDeleteAlbumCache(key string) (CallbackAlbumCached, error)
	// Close must close the underlying database connection
	Close() error
}

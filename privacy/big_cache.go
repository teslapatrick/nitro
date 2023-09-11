package privacy

import (
	"context"
	"github.com/allegro/bigcache"
	"github.com/offchainlabs/nitro/das/dastree"
	flag "github.com/spf13/pflag"
	"time"
)

type BigCacheConfig struct {
	Enable     bool          `koanf:"enable"`
	Expiration time.Duration `koanf:"expiration"`
	MaxEntries int
}

var BigCacheConfigDefault = BigCacheConfig{
	Enable:     false,
	Expiration: time.Hour * 24,
	MaxEntries: 1000,
}

func BigCacheConfigAddOptions(prefix string, f *flag.FlagSet) {
	f.Bool(prefix+".enable", BigCacheConfigDefault.Enable, "Enable BigCache")
	f.Duration(prefix+".expiration", BigCacheConfigDefault.Expiration, "Expiration of BigCache")
}

const CheckStatusKey = "check status"

type BigCacheStorageService struct {
	config   BigCacheConfig
	bigCache *bigcache.BigCache
}

// NewBigCache generates a new BigCache storage
func NewBigCache(config BigCacheConfig) (ICacheService, error) {
	conf := bigcache.DefaultConfig(config.Expiration)
	bc, err := bigcache.NewBigCache(conf)
	if err != nil {
		return nil, err
	}
	bcService := &BigCacheStorageService{
		config:   config,
		bigCache: bc,
	}
	err = bcService.Set(context.Background(), CheckStatusKey, []byte{1}, uint64(0))
	if err != nil {
		return nil, err
	}
	return bcService, nil
}

// Set sets the key-value pair for the auth service
func (bc *BigCacheStorageService) Set(ctx context.Context, key string, value []byte, expiration uint64) (err error) {
	//log.Printf("BigCacheStorageService.Set(%s, %v, %d)\n", key, value, expiration)
	select {
	case <-ctx.Done():
		// in case the context is cancelled, return the error
		err = ctx.Err()
	default:
		// write the key-value pair
		err = bc.bigCache.Set(key, value)
	}
	return
}

// Get returns the value for the given key
func (bc *BigCacheStorageService) Get(ctx context.Context, key string) (res []byte, err error) {
	//fmt.Printf("BigCacheStorageService.Get(%s)\n", key)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// get the key-value pair
		res, err = bc.bigCache.Get(key)
		if err != nil {
			return nil, err
		}
	}
	return
}

func (bc *BigCacheStorageService) HealthCheck(ctx context.Context) bool {
	_, err := bc.bigCache.Get(dastree.Hash([]byte(CheckStatusKey)).String())
	return err == nil
}

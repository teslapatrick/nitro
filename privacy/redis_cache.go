package privacy

import (
	"context"
	"time"

	flag "github.com/spf13/pflag"
)

type RedisCacheConfig struct {
	Enable  bool          `koanf:"enable"`
	Url     string        `koanf:"url"`
	Key     string        `koanf:"key"`
	Refresh time.Duration `koanf:"refresh-duration"`
}

var RedisCacheConfigDefault = RedisCacheConfig{
	Enable:  false,
	Url:     "redis://127.0.0.1:6379",
	Key:     "",
	Refresh: time.Hour * 24,
}

func RedisCacheConfigAddOptions(prefix string, f *flag.FlagSet) {
	f.Bool(prefix+".enable", RedisCacheConfigDefault.Enable, "enable redis cache")
	f.String(prefix+".url", RedisCacheConfigDefault.Url, "redis url")
	f.String(prefix+".key", RedisCacheConfigDefault.Key, "redis cache key")
	f.Duration(prefix+".refresh-duration", RedisCacheConfigDefault.Refresh, "redis cache refresh duration")
}

func (c *RedisCacheConfig) Set(ctx context.Context, key string, value []byte, expiration uint64) (err error) {
	return nil
}

func (c *RedisCacheConfig) Get(ctx context.Context, key string) (res []byte, err error) {
	return nil, nil
}

func (c *RedisCacheConfig) HealthCheck(ctx context.Context) bool {
	return false
}

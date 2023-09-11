package privacy

import "context"

type ICacheService interface {
	Set(ctx context.Context, key string, value []byte, expiration uint64) (err error)
	Get(ctx context.Context, key string) (res []byte, err error)
	HealthCheck(ctx context.Context) bool
}

type IPrivacyAPI interface {
	//CacheForTest() string
	SetToken(key string, value interface{}) error
	GetToken(key string) (interface{}, error)
}

package privacy

import (
	"context"
)

type PrivacyAPI struct {
	wrapper *PrivacyWrapper
}

func NewPrivacyAPI(wrapper *PrivacyWrapper) *PrivacyAPI {
	return &PrivacyAPI{
		wrapper: wrapper,
	}
}

func (p *PrivacyAPI) SetToken(key string, value interface{}) error {
	if !p.healthCheck() {
		return NewApiServiceError("PrivacyAPI service is not ok")
	}
	// todo parse
	return p.wrapper.cache.Set(context.Background(), key, value.([]byte), uint64(0))
}

func (p *PrivacyAPI) GetToken(key string) (interface{}, error) {
	return p.wrapper.cache.Get(context.Background(), key)
}

func (p *PrivacyAPI) healthCheck() bool {
	return p.wrapper.cache.HealthCheck(context.Background())
}

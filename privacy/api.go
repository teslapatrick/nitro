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

func (p *PrivacyAPI) SetToken(ctx context.Context, token string, addresses []string) (interface{}, error) {
	if !p.healthCheck(ctx) {
		return nil, NewApiServiceError("PrivacyAPI service is not ok")
	}
	if token == "" || len(addresses) == 0 {
		return nil, NewApiServiceError("token or address is empty")
	}
	for _, addr := range addresses {
		if addr == "" {
			return nil, NewSetTokenFailedError("address is empty")
		}
		if err := p.wrapper.cache.Set(ctx, addr, []byte(token), uint64(0)); err != nil {
			return nil, NewSetTokenFailedError("PrivacyAPI: set token failed")
		}
	}
	return "Set token successfully", nil
}

func (p *PrivacyAPI) UpdateToken(ctx context.Context, token string, address string) (interface{}, error) {
	if !p.healthCheck(ctx) {
		return nil, NewApiServiceError("PrivacyAPI service is not ok")
	}
	if token == "" || address == "" {
		return nil, NewApiServiceError("token or address is empty")
	}
	if err := p.wrapper.cache.Set(ctx, address, []byte(token), uint64(0)); err != nil {
		return nil, NewSetTokenFailedError("PrivacyAPI: set token failed")
	}
	return "Update token successfully", nil
}

func (p *PrivacyAPI) GetToken(ctx context.Context, address string) (interface{}, error) {
	//if !p.healthCheck() {
	//	return nil, NewApiServiceError("PrivacyAPI service is not ok")
	//}
	//res, err := p.wrapper.cache.Get(ctx, address)
	//if err != nil {
	//	return nil, NewGetTokenFailedError("PrivacyAPI: get token failed")
	//}
	//return string(res), nil
	return nil, NewGetTokenFailedError("PrivacyAPI: method not allowed")
}

func (p *PrivacyAPI) healthCheck(ctx context.Context) bool {
	//return false
	return p.wrapper.cache.HealthCheck(ctx)
}

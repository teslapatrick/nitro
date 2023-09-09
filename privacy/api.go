package privacy

type PrivacyAPI struct {
	wrapper *PrivacyWrapper
}

func NewPrivacyAPI(wrapper *PrivacyWrapper) *PrivacyAPI {
	return &PrivacyAPI{
		wrapper: wrapper,
	}
}

func (p *PrivacyAPI) SetToken(key string, value interface{}, timeout string) error {
	if !p.healthCheck() {
		return NewApiServiceError("PrivacyAPI service is not ok")
	}
	return p.wrapper.cache.Set(key, value)
}

func (p *PrivacyAPI) GetToken(key string) (interface{}, error) {
	return p.wrapper.cache.Get(key)
}

func (p *PrivacyAPI) healthCheck() bool {
	//return false
	return p.wrapper.cache.Check()
}

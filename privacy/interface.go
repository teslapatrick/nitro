package privacy

type IPrivacyCache interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Check() bool
}

type IPrivacyAPI interface {
	//CacheForTest() string
	SetToken(key string, value interface{}) error
	GetToken(key string) (interface{}, error)
}

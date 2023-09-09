package privacy

import (
	flag "github.com/spf13/pflag"
)

type PrivacyConfig struct {
	Enable     bool             `koanf:"enable"`
	API        []string         `koanf:"api"`
	LocalCache BigCacheConfig   `koanf:"cache"`
	RedisCache RedisCacheConfig `koanf:"redis"`
	Das        DASConfig        `koanf:"das"`
	JwtSecret  string           `koanf:"jwtsecret"`
}

var PrivacyRPCConfigDefault = PrivacyConfig{
	Enable:     false,
	API:        []string{"eth", "web3"},
	JwtSecret:  "",
	Das:        DASConfigDefaults,
	LocalCache: BigCacheConfigDefault,
	RedisCache: RedisCacheConfigDefault,
}

func PrivacyRPCConfigAddOptions(prefix string, f *flag.FlagSet) {
	f.Bool(prefix+".enable", PrivacyRPCConfigDefault.Enable, "enable the privacy router")
	f.StringSlice(prefix+".api", PrivacyRPCConfigDefault.API, "api list to support")
	f.String(prefix+".jwtsecret", PrivacyRPCConfigDefault.JwtSecret, "jwt secret")
	DASConfigAddOptions(prefix+".das", f)
	BigCacheConfigAddOptions(prefix+".cache", f)
	RedisCacheConfigAddOptions(prefix+".redis", f)
}

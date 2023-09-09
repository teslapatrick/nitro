package privacy

import (
	flag "github.com/spf13/pflag"
)

type PrivacyConfig struct {
	Enable     bool             `koanf:"enable"`
	API        []string         `koanf:"api"`
	LocalCache BigCacheConfig   `koanf:"local-cache"`
	RedisCache RedisCacheConfig `koanf:"redis-cache"`
	DasEnable  bool             `koanf:"das-enable"`
	JwtSecret  string           `koanf:"jwtsecret"`
}

var PrivacyRPCConfigDefault = PrivacyConfig{
	Enable:    false,
	API:       []string{"eth", "web3"},
	JwtSecret: "",
	DasEnable: false,
}

func PrivacyRPCConfigAddOptions(prefix string, f *flag.FlagSet) {
	f.Bool(prefix+".enable", PrivacyRPCConfigDefault.Enable, "enable the privacy router")
	f.StringSlice(prefix+".api", PrivacyRPCConfigDefault.API, "api list to support")
	f.Bool(prefix+".das.enable", PrivacyRPCConfigDefault.DasEnable, "enable the das reader, if enabled, will use the das config")
	f.String(prefix+".jwtsecret", PrivacyRPCConfigDefault.JwtSecret, "jwt secret")
	BigCacheConfigAddOptions(prefix+".cache", f)
	RedisCacheConfigAddOptions(prefix+".redis", f)
}

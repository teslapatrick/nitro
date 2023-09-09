package privacy

import (
	flag "github.com/spf13/pflag"
	"time"
)

type BigCacheConfig struct {
	Enable     bool          `koanf:"enable"`
	Expiration time.Duration `koanf:"expiration"`
}

var BigCacheConfigDefault = BigCacheConfig{
	Enable:     false,
	Expiration: time.Hour * 24,
}

func BigCacheConfigAddOptions(prefix string, f *flag.FlagSet) {
	f.Bool(prefix+".enable", BigCacheConfigDefault.Enable, "Enable BigCache")
	f.Duration(prefix+".expiration", BigCacheConfigDefault.Expiration, "Expiration of BigCache")
}

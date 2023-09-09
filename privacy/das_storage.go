package privacy

import (
	flag "github.com/spf13/pflag"
)

type DASConfig struct {
	Enable   bool     `koanf:"enable"`
	Backends []string `koanf:"backends"`
}

var DASConfigDefaults = DASConfig{
	Enable:   false,
	Backends: []string{},
}

func DASConfigAddOptions(prefix string, f *flag.FlagSet) {
	f.Bool(prefix+".enable", DASConfigDefaults.Enable, "enable the das reader, if enabled, will use the das config")
	f.StringSlice(prefix+".backends", DASConfigDefaults.Backends, "das storage backends")
}

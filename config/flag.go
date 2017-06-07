package config

type Flag struct {
	True bool
	set bool
}

func (flag *Flag) UnmarshalYAML(unmarshal func(interface {}) error) error {
	if err := unmarshal(&flag.True); err != nil {
		return err
	}

	flag.set = true

	return nil
}

func (flag *Flag) Merge(flag2 Flag) {
	if flag2.set {
		flag.True = flag2.True
		flag.set = true
	}
}

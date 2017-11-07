package config

// Flag represents a nullable boolean value that is considered null until
// either parsed from YAML or merged in from another Flag value.
//
type Flag struct {
	True bool
	set  bool
}

// UnmarshalYAML implements yaml.Unmarshaler to parse the underlying boolean
// value and detect that the Flag should no longer be considered null.
//
func (flag *Flag) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&flag.True); err != nil {
		return err
	}

	flag.set = true

	return nil
}

// Merge takes another flag and, if set, merged its boolean value into this
// one.
//
func (flag *Flag) Merge(flag2 Flag) {
	if flag2.set {
		flag.True = flag2.True
		flag.set = true
	}
}

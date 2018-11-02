package config
import "strconv"

// Flag represents a nullable boolean value that is considered null until
// either parsed from YAML or merged in from another Flag value.
//
type Flag struct {
	True bool
	set  bool
}

// UnmarshalJSON implements json.Unmarshaler to parse the underlying boolean
// value and detect that the Flag should no longer be considered null.
//
func (flag *Flag) UnmarshalJSON(unmarshal []byte) error {
	var err error
	flag.True, err = strconv.ParseBool(string(unmarshal))
	if err != nil {
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

package config

// UserConfig holds configuration fields related to a user account.
//
type UserConfig struct {
	As  string `yaml:"as" validate:"omitempty,username"` // user name
	UID uint   `yaml:"uid"`                              // user ID
	GID uint   `yaml:"gid"`                              // group ID
}

// Merge takes another UserConfig and overwrites this struct's fields.
//
func (user *UserConfig) Merge(user2 UserConfig) {
	if user2.As != "" {
		user.As = user2.As
	}

	if user2.UID != 0 {
		user.UID = user2.UID
	}

	if user2.GID != 0 {
		user.GID = user2.GID
	}
}

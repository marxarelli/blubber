package config

// CurrentVersion declares the currently supported config version.
//
const CurrentVersion string = "v3"

// VersionConfig contains a single field that allows for validation of the
// config version independent from an entire Config struct.
//
type VersionConfig struct {
	Version string `json:"version" validate:"required,currentversion"`
}

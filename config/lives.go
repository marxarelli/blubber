package config

import (
	"gerrit.wikimedia.org/r/blubber/build"
)

// LocalLibPrefix declares the shared directory into which application level
// dependencies will be installed.
//
const LocalLibPrefix = "/opt/lib"

// LivesConfig holds configuration fields related to the livesship of
// installed dependencies and application files.
//
type LivesConfig struct {
	In         string `json:"in" validate:"omitempty,abspath"` // application directory
	UserConfig `json:",inline"`
}

// Merge takes another LivesConfig and overwrites this struct's fields.
//
func (lives *LivesConfig) Merge(lives2 LivesConfig) {
	if lives2.In != "" {
		lives.In = lives2.In
	}

	lives.UserConfig.Merge(lives2.UserConfig)
}

// InstructionsForPhase injects build instructions related to creation of the
// application lives.
//
// PhasePrivileged
//
// Creates LocalLibPrefix directory and application lives's user home
// directory, creates the lives user and its group, and sets up directory
// permissions.
//
func (lives LivesConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	switch phase {
	case build.PhasePrivileged:
		return []build.Instruction{
			build.NewStringArg("LIVES_AS", lives.As),
			build.NewUintArg("LIVES_UID", lives.UID),
			build.NewUintArg("LIVES_GID", lives.GID),
			build.RunAll{append(
				build.CreateUser("$LIVES_AS", "$LIVES_UID", "$LIVES_GID"),
				build.CreateDirectory(lives.In),
				build.Chown("$LIVES_UID", "$LIVES_GID", lives.In),
				build.CreateDirectory(LocalLibPrefix),
				build.Chown("$LIVES_UID", "$LIVES_GID", LocalLibPrefix),
			)},
		}
	case build.PhasePrivilegeDropped:
		return []build.Instruction{
			build.WorkingDirectory{lives.In},
		}
	}

	return []build.Instruction{}
}

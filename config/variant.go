package config

import (
	"gerrit.wikimedia.org/r/blubber/build"
)

// VariantConfig holds configuration fields for each defined build variant.
//
type VariantConfig struct {
	Includes     []string     `json:"includes" validate:"dive,variantref"`
	Copies       CopiesConfig `json:"copies" validate:"omitempty,unique,dive"`
	CommonConfig `json:",inline"`
}

// Merge takes another VariantConfig and overwrites this struct's fields.
//
func (vc *VariantConfig) Merge(vc2 VariantConfig) {
	vc.Copies.Merge(vc2.Copies)
	vc.CommonConfig.Merge(vc2.CommonConfig)
}

// InstructionsForPhase injects build instructions related to dropping
// priviledge and the application entrypoint, then it delegates to its common
// and copies configurations. It also enforces the correct UID/GID on all copy
// instructions returned from deeper config structs.
//
// PhasePrivileged
//
// Ensure the process and file owner is root.
//
// PhasePrivilegeDropped
//
// Ensure the process and file owner is the "lives.as" user.
//
// PhasePreInstall
//
// Ensure the process and file owner is the "lives.as" user.
//
// PhaseInstall
//
// Ensure the process and file owner is the "lives.as" user.
//
// PhasePostInstall
//
// Ensure the process and file owner is the "runs.as" user, unless configured
// to run insecurely as the "lives.as" user. Finally, sets the application
// entrypoint.
//
func (vc *VariantConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	instructions := vc.CommonConfig.InstructionsForPhase(phase)

	var switchUser string
	var uid, gid uint

	switch phase {
	case build.PhasePrivileged:
		switchUser = "root"

	case build.PhasePrivilegeDropped:
		switchUser = vc.Lives.As
		uid, gid = vc.Lives.UID, vc.Lives.GID

	case build.PhasePreInstall:
		uid, gid = vc.Lives.UID, vc.Lives.GID

	case build.PhaseInstall:
		uid, gid = vc.Lives.UID, vc.Lives.GID

	case build.PhasePostInstall:
		if vc.Runs.Insecurely.True {
			uid, gid = vc.Lives.UID, vc.Lives.GID
		} else {
			switchUser = vc.Runs.As
			uid, gid = vc.Runs.UID, vc.Runs.GID
		}

		if len(vc.EntryPoint) > 0 {
			instructions = append(instructions, build.EntryPoint{vc.EntryPoint})
		}
	}

	// CopiesConfig may not implement InstructionsForPhase for all possible
	// phases, which makes the expansion of it here less than efficient, but to
	// assume which phases it does implement would result in gross coupling
	instructions = append(instructions, vc.Copies.Expand(vc.Lives.In).InstructionsForPhase(phase)...)

	if switchUser != "" {
		instructions = append(
			[]build.Instruction{
				build.User{switchUser},
				build.Home(switchUser),
			},
			instructions...,
		)
	}

	if uid != 0 {
		instructions = build.ApplyUser(uid, gid, instructions)
	}

	return instructions
}

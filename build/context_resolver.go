package build

import (
	"context"

	"github.com/moby/buildkit/client/llb"
)

// ContextResolver returns an initialzed llb.State for a build context.
type ContextResolver func(context.Context) (*llb.State, error)

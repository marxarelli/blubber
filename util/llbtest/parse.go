package llbtest

import (
	"testing"

	"github.com/moby/buildkit/solver/pb"
	digest "github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

// ParseDef parses the given [llb.Definition] in a [pb.Op] slice and map of
// [digest.Digest] to ops.
func ParseDef(t *testing.T, def [][]byte) (map[digest.Digest]pb.Op, []pb.Op) {
	t.Helper()

	m := map[digest.Digest]pb.Op{}
	arr := make([]pb.Op, 0, len(def))

	for _, dt := range def {
		var op pb.Op
		err := (&op).Unmarshal(dt)
		require.NoError(t, err)
		dgst := digest.FromBytes(dt)
		m[dgst] = op
		arr = append(arr, op)
		// fmt.Printf(":: %T %+v\n", op.Op, op)
	}

	return m, arr
}

// LastOp returns the final Op in the given ordered slice
func LastOp(t *testing.T, arr []pb.Op) (digest.Digest, int) {
	t.Helper()

	require.True(t, len(arr) > 1)

	op := arr[len(arr)-1]
	require.Equal(t, 1, len(op.Inputs))
	return op.Inputs[0].Digest, int(op.Inputs[0].Index)
}

package imagefs

import (
	"context"
	"io/fs"
)

// lazyFS implements FS, providing a pattern for JIT initialization
// (initialization upon the first call to Open).
type lazyFS struct {
	ctx  context.Context
	init func(context.Context) (FS, error)
	fs   FS
}

func (lfs *lazyFS) Open(name string) (fs.File, error) {
	if lfs.fs == nil {
		fs, err := lfs.init(lfs.ctx)

		if err != nil {
			return nil, err
		}

		lfs.fs = fs
	}

	return lfs.fs.Open(name)
}

func (lfs *lazyFS) Close() error {
	if lfs.fs == nil {
		return nil
	}

	return lfs.fs.Close()
}

func (lfs *lazyFS) WithContext(ctx context.Context) FS {
	if lfs.fs != nil {
		return &lazyFS{ctx: ctx, init: lfs.init, fs: lfs.fs.WithContext(ctx)}
	}

	// NOTE it is very important here to return the current pointer rather than
	// creating a new *lazyFS since we don't yet have a pointer to the real
	// underlying filesystem. Otherwise, the initialization of the real
	// underlying filesystem is effectively lost between calls to `WithContext`,
	// resulting in orphaned FS objects for every call to `Open` and no garbage
	// collection of the temporary files that may be created during
	// initialization, not to mention tons of duplicate initialize processing.
	return lfs
}

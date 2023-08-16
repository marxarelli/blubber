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

	return &lazyFS{ctx: ctx, init: lfs.init}
}

package imagefs

import (
	"context"
	"io/fs"

	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/types"
	"github.com/pkg/errors"
)

// FS provides an fs.FS interface to the underlying image layers as well as a
// Close() method to remove temporary resources that are created upon first
// access.
type FS interface {
	fs.FS

	// Close will remove any underlying temporary resources that have been created
	// as a result of accessing the image filesystem. If the filesystem was never
	// accessed, this should be a noop.
	Close() error

	// WithContext returns a copy of the filesystem with a newly given context.
	WithContext(context.Context) FS
}

// New returns a new FS that will provide JIT access to the given image's
// layered filesystem.
func New(ctx context.Context, ref types.ImageReference, sys *types.SystemContext, cache types.BlobInfoCache) FS {
	return &lazyFS{
		ctx: ctx,
		init: func(ctx context.Context) (FS, error) {
			// initialize the image source and resolve sub filesystems for all blobs
			// based on their mediatypes
			source, err := ref.NewImageSource(ctx, sys)

			if err != nil {
				return nil, errors.Wrap(err, "failed to initialize image source")
			}

			image, err := image.FromSource(ctx, sys, source)

			if err != nil {
				return nil, errors.Wrap(err, "failed to intialize image from source")
			}

			blobinfos := image.LayerInfos()
			layers := make([]FS, len(blobinfos))

			for i, bi := range blobinfos {
				newBlobFS, supported := newBlobFSs[bi.MediaType]

				if !supported {
					return nil, errors.Errorf("media type %s of layer %s is not supported", bi.MediaType, bi.Digest)
				}

				bfs, err := newBlobFS(ctx, source, bi, cache)

				if err != nil {
					return nil, err
				}

				// add blob filesystems in the order they should be searched (last/outer first)
				layers[len(layers)-1-i] = bfs
			}

			return &imageFS{
				ctx:    ctx,
				ref:    ref,
				sys:    sys,
				cache:  cache,
				source: source,
				image:  image,
				layers: layers,
			}, nil
		},
	}
}

type imageFS struct {
	ctx    context.Context
	ref    types.ImageReference
	sys    *types.SystemContext
	cache  types.BlobInfoCache
	image  types.ImageCloser
	source types.ImageSource
	layers []FS
}

// Open attempts to open the given file from each of the image's layers,
// starting from the outer layer first, and returning the first file
// successfully open or the first error if it is any other than
// fs.ErrNotExist.
func (ifs *imageFS) Open(name string) (fs.File, error) {
	for _, layer := range ifs.layers {
		file, err := layer.Open(name)

		if err == nil || !errors.Is(err, fs.ErrNotExist) {
			return file, err
		}
	}

	return nil, fs.ErrNotExist
}

// Close removes and closes temporary resources that were created/opened as a
// result of accessing the filesystem.
//
// The last encountered error will be returned.
func (ifs *imageFS) Close() error {
	if ifs.source != nil {
		defer ifs.source.Close()
	}

	if ifs.image != nil {
		defer ifs.image.Close()
	}

	var err error

	for _, layer := range ifs.layers {
		e := layer.Close()

		if e != nil {
			err = e
		}
	}

	return err
}

// WithContext returns a new imageFS that uses the given context during layer
// retrieval.
func (ifs *imageFS) WithContext(ctx context.Context) FS {
	return &imageFS{
		ctx:    ctx,
		ref:    ifs.ref,
		sys:    ifs.sys,
		cache:  ifs.cache,
		image:  ifs.image,
		source: ifs.source,
		layers: ifs.layers,
	}
}

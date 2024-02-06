package imagefs

import (
	"context"
	"io"

	"github.com/containers/image/v5/types"
	"github.com/pkg/errors"
)

var (
	newBlobFSs = map[string]newBlobFS{}
)

type newBlobFS func(context.Context, types.ImageSource, types.BlobInfo, types.BlobInfoCache) (FS, error)

// BlobOpener is a function that, given an open io.ReadCloser and size for an
// image layer, returns a valid FS.
type BlobOpener func(ctx context.Context, reader io.ReadCloser, size int64) (FS, error)

// RegisterBlobOpener registers a new BlobOpener that will be used to initialize
// an FS from blobs of the given media type.
//
// This function is not thread safe and should only be called from init().
func RegisterBlobOpener(mediaType string, opener BlobOpener) {
	newBlobFSs[mediaType] = newBlobFS(
		func(ctx context.Context, src types.ImageSource, binfo types.BlobInfo, cache types.BlobInfoCache) (FS, error) {
			return &lazyFS{
				ctx: ctx,
				init: func(ctx context.Context) (FS, error) {
					reader, size, err := src.GetBlob(ctx, binfo, cache)

					if err != nil {
						return nil, errors.Wrapf(
							err,
							"failed to get blob %s from image %s",
							binfo.Digest,
							src.Reference().StringWithinTransport(),
						)
					}

					fs, err := opener(ctx, reader, size)

					if err != nil {
						return nil, errors.Wrapf(err, "failed to open blob")
					}

					return fs, nil
				},
			}, nil
		},
	)
}

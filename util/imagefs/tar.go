package imagefs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/fs"

	oci_v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

func openTarFS(ctx context.Context, reader io.ReadCloser, _ int64) (FS, error) {
	// read all files into memory
	tr := tar.NewReader(reader)

	files := map[string]*tarFile{}

readRecords:
	for {
		select {
		case <-ctx.Done():
			return nil, errors.Errorf("context cancelled")
		default:
			header, err := tr.Next()

			if err == io.EOF {
				break readRecords
			}

			if err != nil {
				return nil, errors.Wrap(err, "error reading tar")
			}

			var data []byte

			if header.Typeflag == tar.TypeReg {
				data, err = io.ReadAll(tr)

				if err != nil {
					return nil, errors.Wrapf(err, "error reading tar entry %s", header.Name)
				}

			}

			files[header.Name] = &tarFile{data, header}
		}
	}

	return &tarFS{files: files}, nil
}

func openTarGzipFS(ctx context.Context, reader io.ReadCloser, size int64) (FS, error) {
	tarReader, err := gzip.NewReader(reader)

	if err != nil {
		return nil, errors.Wrap(err, "failed to open gzip stream")
	}

	return openTarFS(ctx, tarReader, size)
}

type tarFS struct {
	files map[string]*tarFile
}

func (tfs *tarFS) Open(name string) (fs.File, error) {
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}

	if file, ok := tfs.files[name]; ok {
		return file.Open(), nil
	}

	return nil, fs.ErrNotExist
}

func (tfs *tarFS) Close() error {
	return nil
}

func (tfs *tarFS) WithContext(_ context.Context) FS {
	return tfs
}

type tarFile struct {
	data   []byte
	header *tar.Header
}

func (tf *tarFile) Open() fs.File {
	return &openTarFile{bytes.NewReader(tf.data), tf.header}
}

type openTarFile struct {
	*bytes.Reader
	header *tar.Header
}

func (otf *openTarFile) Stat() (fs.FileInfo, error) {
	return otf.header.FileInfo(), nil
}

func (otf *openTarFile) Close() error {
	return nil
}

func init() {
	RegisterBlobOpener(oci_v1.MediaTypeImageLayer, openTarFS)
	RegisterBlobOpener(oci_v1.MediaTypeImageLayerGzip, openTarGzipFS)
	RegisterBlobOpener(oci_v1.MediaTypeImageLayerNonDistributable, openTarFS)
	RegisterBlobOpener(oci_v1.MediaTypeImageLayerNonDistributableGzip, openTarGzipFS)
}

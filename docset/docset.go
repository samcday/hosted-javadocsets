package docset

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/cihub/seelog"
	"github.com/samcday/go-dash-javadocset"
	"github.com/samcday/hosted-javadocsets/mavencentral"
)

func Create(groupId, artifactId, version string, w io.WriteCloser) error {
	defer w.Close()

	log.Infof("Creating docset for %s:%s:%s", groupId, artifactId, version)
	stream, err := mavencentral.GetArtifact(groupId, artifactId, version, "javadoc")
	if err != nil {
		return err
	}
	defer stream.Close()
	file, err := ioutil.TempFile("", "javadocsetartifact")
	if err != nil {
		return err
	}
	defer file.Close()
	defer os.Remove(file.Name())

	io.Copy(file, stream)

	javadocDir, err := ioutil.TempDir("", "javadocsetartifact")
	if err != nil {
		return err
	}
	defer os.RemoveAll(javadocDir)

	docsetDir, err := ioutil.TempDir("", "javadocset")
	if err != nil {
		return err
	}
	defer os.RemoveAll(docsetDir)

	log.Debugf("Unzipping javadoc from %s into %s", file.Name(), javadocDir)
	if err := unzip(file.Name(), javadocDir); err != nil {
		return err
	}

	log.Debug("Building javadocset in ", docsetDir)
	if err := javadocset.Build(javadocDir, docsetDir, artifactId); err != nil {
		return err
	}

	gw := gzip.NewWriter(w)
	defer gw.Close()

	return tarDir(filepath.Join(docsetDir, artifactId+".docset"), gw)
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				return err
			}
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			_, err = io.Copy(f, rc)
			f.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// tar walks directory at given path and writes a tar archive to provided Writer
func tarDir(root string, w io.Writer) error {
	log.Info("Tarring ", root)
	tw := tar.NewWriter(w)
	defer tw.Close()
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Tar archives don't need directory entries.
		if info.IsDir() {
			return nil
		}
		tarFilename, _ := filepath.Rel(root, path)
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = filepath.Join(filepath.Base(root), tarFilename)
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := writeTarData(tw, f); err != nil {
			return err
		}
		return nil
	})
}

func writeTarData(tw *tar.Writer, f *os.File) error {
	log.Tracef("Writing %s to tar archive", f.Name())
	b := make([]byte, 1024)
	for {
		read, err := f.Read(b)
		if _, werr := tw.Write(b[0:read]); werr != nil {
			return err
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}

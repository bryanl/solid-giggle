package importer

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/bryanl/keps/pkg/kep"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var reIsKep = regexp.MustCompile(`^\d+.*?\.md$`)

// Import imports keps from a disk location.
func Import(root string) error {
	if err := os.RemoveAll("site/content/keps"); err != nil {
		return errors.Wrap(err, "clear keps content")
	}

	kepDir, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	count := 0
	failed := 0

	err = filepath.Walk(kepDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "unable to access path %s", path)
		}

		if !isKep(fi) {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("paniced while reading kep %q from disk: %v\n", path, r)
				failed++
			}
		}()

		k, err := kep.Read(f)
		if err != nil {
			logrus.Errorf("unable to process %s", path)
			failed++
			return nil
		}

		logrus.Infof("processed %s", k.Title)
		count++

		destDir := filepath.Join("site/content/keps", k.OwningSIG)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return errors.Wrapf(err, "create sig dir")
		}

		dest := filepath.Join(destDir, fi.Name())
		data, err := k.String()
		if err != nil {
			return errors.Wrapf(err, "convert %s to hugo doc", path)
		}

		if err := ioutil.WriteFile(dest, []byte(data), 0644); err != nil {
			return errors.Wrapf(err, "write %s as hugo doc", path)
		}

		return nil
	})

	if err != nil {
		return errors.Wrap(err, "walk kep dir")
	}

	fmt.Printf("processed %d keps and %d failures\n", count, failed)

	return nil

}

func isKep(fi os.FileInfo) bool {
	return !fi.IsDir() &&
		filepath.Ext(fi.Name()) == ".md" &&
		reIsKep.MatchString(fi.Name())
}

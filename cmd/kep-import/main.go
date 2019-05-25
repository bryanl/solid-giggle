package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/bryanl/keps/pkg/kep"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func main() {
	kepDir := "../enhancements/keps"
	if err := run(kepDir); err != nil {
		logrus.WithError(err).Error("unable to generate")
		os.Exit(1)
	}
}

var reIsKep = regexp.MustCompile(`^\d+.*?\.md$`)

func run(kepDir string) error {
	kepDir, err := filepath.Abs(kepDir)
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

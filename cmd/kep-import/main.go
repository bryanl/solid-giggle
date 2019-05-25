package main

import (
	"os"

	"github.com/bryanl/keps/pkg/importer"
	"github.com/sirupsen/logrus"
)

func main() {
	kepDir := "../enhancements/keps"
	if err := run(kepDir); err != nil {
		logrus.WithError(err).Error("unable to generate")
		os.Exit(1)
	}
}

func run(kepDir string) error {
	return importer.Import(kepDir)
}

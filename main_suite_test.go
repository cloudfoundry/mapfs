package main

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var binaryPath string
var binaryName string

func TestLocalbroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mapfs Main Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	binaryPath, err = gexec.Build("code.cloudfoundry.org/mapfs", "-race")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binaryPath)
}, func(bytes []byte) {
	binaryPath = string(bytes)
	binaryName = filepath.Base(binaryPath)
})

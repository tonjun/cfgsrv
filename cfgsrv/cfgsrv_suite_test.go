package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfgsrv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cfgsrv Suite")
}

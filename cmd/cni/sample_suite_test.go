// The boilerplate needed for Ginkgo

package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSample(t *testing.T) {
	t.Skip()
	RegisterFailHandler(Fail)
	RunSpecs(t, "plugins/sample")
}

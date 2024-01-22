package navaros_test

import (
	"net/http/httptest"

	"github.com/RobertWHurst/navaros"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Context", func() {

	Describe("NewContext", func() {
		It("should return a context", func() {
			res := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/a/b/c", nil)

			ctx := navaros.NewContext(res, req, nil)

			Expect(ctx).ToNot(BeNil())
		})
	})
})

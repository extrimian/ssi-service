package testutil

import (
	"os"

	"github.com/extrimian/ssi-sdk/schema"
)

func EnableSchemaCaching() {
	s, err := schema.GetAllLocalSchemas()
	if err != nil {
		println(err)
		os.Exit(1)
	}
	l, err := schema.NewCachingLoader(s)
	if err != nil {
		println(err)
		os.Exit(1)
	}
	l.EnableHTTPCache()
}

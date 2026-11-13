package data_test

import (
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/data"
	"github.com/larvit/fejkdata/internal/fejktest"
)

var onTestDay = fejkdata.WithToday(fejktest.TestDay)

func withShipped() fejkdata.Option { return fejkdata.WithDataFS(data.Modules()...) }

func shipped(t *testing.T, opts ...fejkdata.Option) *fejkdata.Generator {
	t.Helper()
	f, err := fejkdata.New(append([]fejkdata.Option{withShipped(), fejkdata.WithSeed(1), onTestDay}, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

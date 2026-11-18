package main

import (
	"slices"
	"testing"

	"github.com/Salim8519/Peregrine/engine"
)

func TestAfterDoneMethods(t *testing.T) {
	cases := []struct {
		ad   engine.AfterDone
		want []string
	}{
		{engine.AfterDone{}, nil},
		{engine.AfterDone{Popup: true}, nil},
		{engine.AfterDone{OpenFile: true}, []string{"openFile"}},
		{engine.AfterDone{OpenFolder: true}, []string{"showInFolder"}},
		{engine.AfterDone{OpenFile: true, OpenFolder: true, Popup: true}, []string{"openFile", "showInFolder"}},
	}
	for _, c := range cases {
		if got := afterDoneMethods(c.ad); !slices.Equal(got, c.want) {
			t.Errorf("%+v: got %v, want %v", c.ad, got, c.want)
		}
	}
}

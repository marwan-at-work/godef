package main

import (
	"bytes"
	"go/token"
	"io/ioutil"
	"log"
	"testing"

	"golang.org/x/tools/go/packages/packagestest"
)

var (
	startMarker     = []byte("/*#")
	endMarker       = []byte("*/")
	directiveMarker = []byte(":")
)

func TestGoDef(t *testing.T) { packagestest.TestAll(t, testGoDef) }
func testGoDef(t *testing.T, exporter packagestest.Exporter) {

	cfg, files, cleanup := packagestest.Write(t, exporter, map[string]string{
		"github.com/rogpeppe/godef#tests/a.go": "LINK:testdata/a.go",
	})
	defer cleanup()

	anchors := map[string]token.Position{}
	type testEntry struct {
		filename string
		offset   int
		marker   string
	}
	tests := []testEntry{}
	// search for all the anchor points
	for _, fullpath := range files {
		content, err := ioutil.ReadFile(fullpath)
		if err != nil {
			log.Fatalf("Could not read test file: %v", err)
		}
		next := 0
		for _, line := range bytes.SplitAfter(content, []byte("\n")) {
			offset := next
			next += len(line)
			pos := 0
			//does this line have any markers
			for {
				start := bytes.Index(line[pos:], startMarker)
				if start < 0 {
					break
				}
				start = pos + start + len(startMarker)
				end := bytes.Index(line[start:], endMarker)
				if end < 0 {
					break
				}
				end = start + end
				anchor := string(line[start:end])
				pos = end + len(endMarker)
				log.Printf("Found anchor %q at %d", anchor, offset)
			}
		}
	}

	for _, test := range tests {
		fset, obj, err := godef(cfg, test.filename, nil, test.offset)
		if err != nil {
			t.Errorf("Failed: %v", err)
			continue
		}
		pos := fset.Position(obj.Pos())
		target := anchors[test.marker]
		if pos.String() != target.String() {
			t.Errorf("Got %v expected %v", pos, target)
			continue
		}
	}
	t.Errorf("No way dude")
}

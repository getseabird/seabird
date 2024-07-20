package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Must convert strokes to fills when adding icons from Lucide
// npm i -g oslllo-svg-fixer svgo
// sed -i 's/stroke-width="2"/stroke-width="3"/' lucide/*.svg
// oslllo-svg-fixer -s lucide -d lucide
// svgo -f lucide -o lucide
// TODO increase stroke with to 2.5

func main() {
	xml :=
		`<?xml version="1.0" encoding="UTF-8"?>
	<gresources>
		<gresource prefix="/hicolor/symbolic/actions">
			<file alias="edit-find-symbolic.svg">seabird.svg</file>
		</gresource>
		<gresource prefix="/dev/skynomads/Seabird/icons/scalable/actions/">
			<file preprocess="xml-stripblanks">seabird.svg</file>
			<file alias="edit-find-symbolic.svg">seabird.svg</file> <!-- TODO doesn't work -->
`

	for _, dir := range []string{"lucide"} {
		files, err := os.ReadDir(dir)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			if f.Type().IsDir() || !strings.HasSuffix(f.Name(), ".svg") {
				continue
			}
			xml += fmt.Sprintf(
				`			<file alias="%s-symbolic.svg" preprocess="xml-stripblanks">%s/%s</file>
`,
				strings.TrimSuffix(strings.TrimSuffix(f.Name(), ".svg"), "-symbolic"), dir, f.Name(),
			)
		}
	}

	xml += `		</gresource>
	</gresources>
	`

	os.WriteFile("gresource.xml", []byte(xml), os.ModePerm)
}

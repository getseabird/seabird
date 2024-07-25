package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	xml :=
		`<?xml version="1.0" encoding="UTF-8"?>
	<gresources>
		<gresource prefix="/dev/skynomads/Seabird/icons/scalable/actions/">
			<file preprocess="xml-stripblanks">seabird.svg</file>
`

	icons := map[string]string{}

	for _, root := range []string{"icon-development-kit"} {
		files, err := os.ReadDir(root)
		if err != nil {
			log.Fatal(err)
		}
		for _, f1 := range files {
			if f1.Type().IsDir() {
				files, err := os.ReadDir(path.Join(root, f1.Name()))
				if err != nil {
					log.Fatal(err)
				}
				for _, f2 := range files {
					if f2.Type().IsDir() || !strings.HasSuffix(f2.Name(), ".svg") {
						continue
					}
					icons[strings.TrimSuffix(strings.TrimSuffix(f2.Name(), ".svg"), "-symbolic")] = path.Join(root, f1.Name(), f2.Name())
				}
			} else {
				if !strings.HasSuffix(f1.Name(), ".svg") {
					continue
				}
				icons[strings.TrimSuffix(strings.TrimSuffix(f1.Name(), ".svg"), "-symbolic")] = path.Join(root, f1.Name())
			}
		}
	}

	for name, path := range icons {
		xml += fmt.Sprintf(
			`			<file alias="%s-symbolic.svg" preprocess="xml-stripblanks">%s</file>
`,
			strings.TrimSuffix(strings.TrimSuffix(name, ".svg"), "-symbolic"), path,
		)
	}

	xml += `		</gresource>
	</gresources>
	`

	os.WriteFile("gresource.xml", []byte(xml), os.ModePerm)
}

package py

import (
	"log"
	"os"
	"path/filepath"
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("PYTHONPATH", filepath.Join(wd, "testdata"))
	Initialize()
}

package main

import (
	"encoding/json"
	"os"
)

func main() {
	if err := list(); err != nil {
		print("Error: ", err)
		os.Exit(1)
	}
}

func list() error {
	var contents []string
	for _, arg := range os.Args[1:] {
		buf, err := os.ReadFile(arg)
		if err != nil {
			return err
		}

		contents = append(contents, string(buf))
	}

	buf, err := json.Marshal(contents)
	if err != nil {
		return err
	}

	os.Stdout.Write(buf)
	return nil
}

package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	pwd, err := openAsString("./signal.backup.password")
	if err != nil {
		fmt.Println("fuck you, password can't be read:", err.Error())
		os.Exit(1)
	}

	bf, err := newBackupFile("signal.backup", pwd)
	if err != nil {
		fmt.Printf("got error: %s", err)
	}

	f, _ := bf.frame()
	fmt.Println("got frame:", f)
}

func openAsString(path string) (string, error) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %s", err)
	}

	return string(bs), nil
}

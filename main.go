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
	var as int

	for {
		f, err := bf.frame()
		if err != nil {
			break
		}

		if a := f.GetAttachment(); a != nil {
			fmt.Println("attachment get:", a)
			file, err := os.OpenFile(fmt.Sprintf("out%v.jpg", as), os.O_CREATE|os.O_WRONLY, os.ModePerm)
			if err != nil {
				fmt.Println("nope:", err.Error())
				break
			}
			as++
			if _, err = bf.decryptAttachment(a, file); err != nil {
				fmt.Println("nope att:", err.Error())
			}
		}

		if s := f.GetStatement(); s != nil {
			_, err := extractImage(s)
			if err != nil {
				// fmt.Println("nope:", err.Error())
			}
		}

		// fmt.Println("got frame:", f)
	}
}

func openAsString(path string) (string, error) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %s", err)
	}

	return string(bs), nil
}

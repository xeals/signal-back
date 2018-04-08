package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

var aEncs map[uint64]string

func init() {
	aEncs = make(map[uint64]string)
}

func extractAttachments(bf *backupFile) error {
	for {
		f, err := bf.frame()
		if err != nil {
			return nil // TODO This should be specific to an EOF-type error
		}

		ps := f.GetStatement().GetParameters()
		if len(ps) == 25 { // Contains blob information
			aEncs[*ps[19].IntegerParameter] = *ps[3].StringParamter
		}

		if a := f.GetAttachment(); a != nil {
			var ext string
			switch enc := aEncs[*a.AttachmentId]; enc {
			case "image/jpeg":
				ext = "jpg"
			default:
				return errors.Errorf("encoding `%s` not recognised. create a PR or issue if you think it should be", enc)
			}

			fileName := fmt.Sprintf("%v.%s", *a.AttachmentId, ext)
			file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "failed to open output file")
			}
			if _, err = bf.decryptAttachment(a, file); err != nil {
				return errors.Wrap(err, "failed to decrypt attachment")
			}
		}
	}
}

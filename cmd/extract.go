package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/xeals/signal-back/types"
)

// Extract fulfils the `extract` subcommand.
var Extract = cli.Command{
	Name:               "extract",
	Usage:              "Retrieve attachments from the backup",
	UsageText:          "Decrypt files embedded in the backup.",
	CustomHelpTemplate: SubcommandHelp,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "log, l",
			Usage: "write logging output to `FILE`",
		},
		cli.StringFlag{
			Name:  "password, p",
			Usage: "use `PASS` as password for backup file",
		},
		cli.StringFlag{
			Name:  "pwdfile, P",
			Usage: "read password from `FILE`",
		},
		cli.StringFlag{
			Name:  "outdir, o",
			Usage: "output attachments to `DIRECTORY`",
		},
	},
	Action: func(c *cli.Context) error {
		bf, err := setup(c)
		if err != nil {
			return err
		}

		if path := c.String("outdir"); path != "" {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return errors.Wrap(err, "unable to create output directory")
			}
			err = os.Chdir(path)
			if err != nil {
				return errors.Wrap(err, "unable to change working directory")
			}
		}
		if err = ExtractAttachments(bf); err != nil {
			return errors.Wrap(err, "failed to extract attachment")
		}

		return nil
	},
}

// ExtractAttachments pulls only the attachments out of the backup file and
// outputs them in the current working directory.
func ExtractAttachments(bf *types.BackupFile) error {
	aEncs := make(map[uint64]string)
	for {
		f, err := bf.Frame()
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
			if err = bf.DecryptAttachment(a, file); err != nil {
				return errors.Wrap(err, "failed to decrypt attachment")
			}
		}
	}
}

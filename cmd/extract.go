package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/h2non/filetype"
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
	Flags: append([]cli.Flag{
		cli.StringFlag{
			Name:  "outdir, o",
			Usage: "output attachments to `DIRECTORY`",
		},
	}, coreFlags...),
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
	defer func() {
		if r := recover(); r != nil {
			log.Println("Panicked during extraction:", r)
		}
	}()
	defer bf.Close()

	for {
		f, err := bf.Frame()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return errors.Wrap(err, "extraction")
		}

		ps := f.GetStatement().GetParameters()
		if len(ps) == 25 { // Contains blob information
			aEncs[*ps[19].IntegerParameter] = *ps[3].StringParameter
			log.Printf("found attachment metadata %v: `%v`\n", *ps[19].IntegerParameter, ps)
		}

		if a := f.GetAttachment(); a != nil {
			log.Printf("found attachment binary %v\n", *a.AttachmentId)
			id := *a.AttachmentId

			mime, hasMime := aEncs[id]
			ext := getExt(mime, id)

			fileName := fmt.Sprintf("%v%s", id, ext)
			file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)

			if err != nil {
				return errors.Wrap(err, "failed to open output file")
			}
			if err = bf.DecryptAttachment(a.GetLength(), file); err != nil {
				return errors.Wrap(err, "failed to decrypt attachment")
			}
			if err = file.Close(); err != nil {
				return errors.Wrap(err, "failed to close output file")
			}

			if !hasMime { // Time to look into the file itself and guess.
				buf, err := ioutil.ReadFile(fileName)
				if err != nil {
					return errors.Wrap(err, "failed to read output file for MIME detection")
				}
				kind, err := filetype.Match(buf)
				if err != nil {
					log.Printf("unable to detect file type: %s\n", err.Error())
				}
				if err = os.Rename(fileName, fileName+"."+kind.Extension); err != nil {
					log.Println("unknown file type")
					return errors.Wrap(err, "unable to rename output file")
				}
				log.Println("found file type:", kind.MIME)
			}
		}
	}
}

func getExt(mime string, file uint64) string {
	// List taken from https://github.com/h2non/filetype
	switch mime {
	// IMAGE
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/x-canon-cr2":
		return ".cr2"
	case "image/tiff":
		return ".tif"
	case "image/bmp":
		return ".bmp"
	case "image/vnd.ms-photo":
		return ".jxr"
	case "image/vnd.adobe.photoshop":
		return ".psd"
	case "image/x-icon":
		return ".ico"

		// VIDEO
	case "video/mp4":
		return ".mp4"
	case "video/x-m4v":
		return ".m4v"
	case "video/x-matroska":
		return ".mkv"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	case "video/x-msvideo":
		return ".avi"
	case "video/x-ms-wmv":
		return ".wmv"
	case "video/mpeg":
		return ".mpg"
	case "video/x-flv":
		return ".flv"

		// AUDIO
	case "audio/midi":
		return ".mid"
	case "audio/mpeg":
		return ".mp3"
	case "audio/m4a":
		return ".m4a"
	case "audio/ogg":
		return ".ogg"
	case "audio/x-flac":
		return ".flac"
	case "audio/x-wav":
		return ".wav"
	case "audio/amr":
		return ".amr"

		// ARCHIVE
	case "application/epub+zip":
		return ".epub"
	case "application/zip":
		return ".zip"
	case "application/x-tar":
		return ".tar"
	case "application/x-rar-compressed":
		return ".rar"
	case "application/gzip":
		return ".gz"
	case "application/x-bzip2":
		return ".bz2"
	case "application/x-7z-compressed":
		return ".7z"
	case "application/x-xz":
		return ".xz"
	case "application/pdf":
		return ".pdf"
	case "application/x-msdownload":
		return ".exe"
	case "application/x-shockwave-flash":
		return ".swf"
	case "application/rtf":
		return ".rtf"
	case "application/octet-stream":
		return ".eot"
	case "application/postscript":
		return ".ps"
	case "application/x-sqlite3":
		return ".sqlite"
	case "application/x-nintendo-nes-rom":
		return ".nes"
	case "application/x-google-chrome-extension":
		return ".crx"
	case "application/vnd.ms-cab-compressed":
		return ".cab"
	case "application/x-deb":
		return ".deb"
	case "application/x-unix-archive":
		return ".ar"
	case "application/x-compress":
		return ".Z"
	case "application/x-lzip":
		return ".lz"
	case "application/x-rpm":
		return ".rpm"
	case "application/x-executable":
		return ".elf"

		// DOCUMENTS
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "application/vnd.ms-excel":
		return ".xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	case "application/vnd.ms-powerpoint":
		return ".ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return ".pptx"

		// FONTS
	case "application/font-woff":
		warnExt(file, "woff2")
		return ".woff"
	case "application/font-sfnt":
		warnExt(file, "otf")
		return ".ttf"

	case "":
		log.Printf("file `%v` has no associated SQL entry; going to have to guess at its encoding", file)
		return ""

	default:
		log.Printf("encoding `%s` not recognised. create a PR or issue if you think it should be\n", mime)
		log.Printf("if you can provide details on the file `%v` as well, it would be appreciated", file)
		return ""
	}
}

func warnExt(file uint64, mime string) {
	log.Printf("note that file `%v` should possibly have file extension `%s`", file, mime)
}

package types

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/xeals/signal-back/signal"
	"golang.org/x/crypto/hkdf"
)

// ATTACHMENT_BUFFER_SIZE is the size of the buffer in bytes used for decrypting attachments. Larger
// values of this will consume more memory, but may decrease the overall time taken to decrypt an
// attachment.
const ATTACHMENT_BUFFER_SIZE = 8192

// ProtoCommitHash is the commit hash of the Signal Protobuf spec.
var ProtoCommitHash = "d6610f0"

// BackupFile stores information about a given backup file.
//
// Decrypting a backup is done by consuming the underlying file buffer. Attemtping to read from a
// BackupFile after it is consumed will return an error.
//
// Closing the underlying file handle is the responsibilty of the programmer if implementing the
// iteration manually, or is done as part of the Consume method.
type BackupFile struct {
	file      *os.File
	FileSize  int64
	CipherKey []byte
	MacKey    []byte
	Mac       hash.Hash
	IV        []byte
	Counter   uint32
}

// NewBackupFile initialises a backup file for reading using the provided path
// and password.
func NewBackupFile(path, password string) (*BackupFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open backup file")
	}
	size := info.Size()

	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open backup file")
	}

	headerLengthBytes := make([]byte, 4)
	_, err = io.ReadFull(file, headerLengthBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read headerLengthBytes")
	}
	headerLength := bytesToUint32(headerLengthBytes)

	headerFrame := make([]byte, headerLength)
	_, err = io.ReadFull(file, headerFrame)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read headerFrame")
	}
	frame := &signal.BackupFrame{}
	if err = proto.Unmarshal(headerFrame, frame); err != nil {
		return nil, errors.Wrap(err, "failed to decode header")
	}

	iv := frame.Header.Iv
	if len(iv) != 16 {
		return nil, errors.New("No IV in header")
	}

	key := backupKey(password, frame.Header.Salt)
	derived := deriveSecrets(key, []byte("Backup Export"))
	cipherKey := derived[:32]
	macKey := derived[32:]

	return &BackupFile{
		file:      file,
		FileSize:  size,
		CipherKey: cipherKey,
		MacKey:    macKey,
		Mac:       hmac.New(crypto.SHA256.New, macKey),
		IV:        iv,
		Counter:   bytesToUint32(iv),
	}, nil
}

// Frame returns the next frame in the file.
func (bf *BackupFile) Frame() (*signal.BackupFrame, error) {
	length := make([]byte, 4)
	_, err := io.ReadFull(bf.file, length)
	if err != nil {
		return nil, err
	}

	frameLength := bytesToUint32(length)
	frame := make([]byte, frameLength)

	io.ReadFull(bf.file, frame)

	theirMac := frame[:len(frame)-10]

	bf.Mac.Reset()
	bf.Mac.Write(frame)
	ourMac := bf.Mac.Sum(nil)

	if bytes.Equal(theirMac, ourMac) {
		return nil, errors.New("Bad MAC")
	}

	uint32ToBytes(bf.IV, bf.Counter)
	bf.Counter++

	aesCipher, err := aes.NewCipher(bf.CipherKey)
	if err != nil {
		return nil, errors.New("Bad cipher")
	}
	stream := cipher.NewCTR(aesCipher, bf.IV)

	output := make([]byte, len(frame)-10)
	stream.XORKeyStream(output, frame[:len(frame)-10])

	decoded := new(signal.BackupFrame)
	proto.Unmarshal(output, decoded)

	return decoded, nil
}

// DecryptAttachment reads the attachment immediately next in the file's bytes, using a streaming
// intermediate buffer of size ATTACHMENT_BUFFER_SIZE.
func (bf *BackupFile) DecryptAttachment(length uint32, out io.Writer) error {
	if length == 0 {
		return errors.New("can't read attachment of length 0")
	}

	uint32ToBytes(bf.IV, bf.Counter)
	bf.Counter++

	aesCipher, err := aes.NewCipher(bf.CipherKey)
	if err != nil {
		return errors.New("Bad cipher")
	}
	stream := cipher.NewCTR(aesCipher, bf.IV)
	bf.Mac.Write(bf.IV)

	buf := make([]byte, ATTACHMENT_BUFFER_SIZE)
	output := make([]byte, len(buf))

	for length > 0 {
		// Go can't read an arbitrary number of bytes,
		// so we have to downsize the containing buffer instead.
		if length < ATTACHMENT_BUFFER_SIZE {
			buf = make([]byte, length)
		}
		n, err := bf.file.Read(buf)
		if err != nil {
			return errors.Wrap(err, "failed to read att")
		}
		bf.Mac.Write(buf)

		stream.XORKeyStream(output, buf)
		if _, err = out.Write(output); err != nil {
			return errors.Wrap(err, "can't write to output")
		}

		length -= uint32(n)
	}

	theirMac := make([]byte, 10)
	io.ReadFull(bf.file, theirMac)
	ourMac := bf.Mac.Sum(nil)

	if bytes.Equal(theirMac, ourMac) {
		return errors.New("Bad MAC")
	}

	return nil
}

// ConsumeFuncs stores parameters for a Consume operation.
type ConsumeFuncs struct {
	AttachmentFunc func(*signal.Attachment) error
	AvatarFunc     func(*signal.Avatar) error
	StatementFunc  func(*signal.SqlStatement) error
}

func DiscardConsumeFuncs(bf *BackupFile) ConsumeFuncs {
	return ConsumeFuncs{
		AttachmentFunc: func(a *signal.Attachment) error {
			return bf.DecryptAttachment(a.GetLength(), ioutil.Discard)
		},
		AvatarFunc: func(a *signal.Avatar) error {
			return bf.DecryptAttachment(a.GetLength(), ioutil.Discard)
		},
		StatementFunc: func(s *signal.SqlStatement) error {
			return nil
		},
	}
}

// Consume iterates over the backup file using the fields in the provided ConsumeFuncs. When a
// BackupFrame is encountered, the matching function will run.
//
// If any image-related functions are nil (e.g., AttachmentFunc) the default will be to discard the
// next *n* bytes, where n is the Attachment.Length.
//
// The underlying file is closed at the end of the method, and the backup file should be considered
// spent.
func (bf *BackupFile) Consume(fns ConsumeFuncs) error {
	var (
		f       *signal.BackupFrame
		err     error
		discard = DiscardConsumeFuncs(bf)
	)

	defer bf.Close()

	if fns.AttachmentFunc == nil {
		fns.AttachmentFunc = discard.AttachmentFunc
	}
	if fns.AvatarFunc == nil {
		fns.AvatarFunc = discard.AvatarFunc
	}
	if fns.StatementFunc == nil {
		fns.StatementFunc = discard.StatementFunc
	}

	for {
		f, err = bf.Frame()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if a := f.GetAttachment(); a != nil {
			if err = fns.AttachmentFunc(a); err != nil {
				return errors.Wrap(err, "consume [attachment]")
			}
		}
		if a := f.GetAvatar(); a != nil {
			if err = fns.AvatarFunc(a); err != nil {
				return errors.Wrap(err, "consume [avatar]")
			}
		}
		if stmt := f.GetStatement(); stmt != nil {
			if err = fns.StatementFunc(stmt); err != nil {
				return errors.Wrap(err, "consume [statement]")
			}
		}
	}

	return nil
}

// Slurp consumes the entire BackupFile and returns a list of all frames
// contained in the file. Note that after calling this function, the underlying
// file buffer will be empty and the file should be considered dropped. Calling
// any function on the backup file after calling Slurp will fail.
//
// Note that any attachments in the backup file will not be handled.
//
// Closes the underlying file handler afterwards. The backup file should be considered exhausted.
func (bf *BackupFile) Slurp() ([]*signal.BackupFrame, error) {
	frames := []*signal.BackupFrame{}
	defer bf.Close()

	for {
		f, err := bf.Frame()
		if err == io.EOF {
			return frames, nil
		} else if err != nil {
			return nil, err
		}

		frames = append(frames, f)

		// Remove images
		if a := f.GetAttachment(); a != nil {
			if err = bf.DecryptAttachment(a.GetLength(), ioutil.Discard); err != nil {
				return nil, errors.Wrap(err, "failed to remove attachment")
			}
		}
		if a := f.GetAvatar(); a != nil {
			if err = bf.DecryptAttachment(a.GetLength(), ioutil.Discard); err != nil {
				return nil, errors.Wrap(err, "failed to remove avatar")
			}
		}
	}
}

func (bf *BackupFile) Close() error {
	return bf.file.Close()
}

func backupKey(password string, salt []byte) []byte {
	digest := crypto.SHA512.New()
	input := []byte(strings.Replace(strings.TrimSpace(password), " ", "", -1))
	hash := input

	if salt != nil {
		digest.Write(salt)
	}

	for i := 0; i < 250000; i++ {
		digest.Write(hash)
		digest.Write(input)
		hash = digest.Sum(nil)
		digest.Reset()
	}

	return hash[:32]
}

func deriveSecrets(input, info []byte) []byte {
	sha := crypto.SHA256.New
	salt := make([]byte, sha().Size())
	okm := make([]byte, 64)

	hkdf := hkdf.New(sha, input, salt, info)
	_, err := io.ReadFull(hkdf, okm)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to generate hashes:", err.Error())
	}

	return okm
}

func bytesToUint32(b []byte) (val uint32) {
	val |= uint32(b[3])
	val |= uint32(b[2]) << 8
	val |= uint32(b[1]) << 16
	val |= uint32(b[0]) << 24
	return
}

func uint32ToBytes(b []byte, val uint32) {
	b[3] = byte(val)
	b[2] = byte(val >> 8)
	b[1] = byte(val >> 16)
	b[0] = byte(val >> 24)
	return
}

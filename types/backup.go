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

// ProtoCommitHash is the commit hash of the Signal Protobuf spec.
var ProtoCommitHash = "d6610f0"

// BackupFile holds the internal state of decryption of a Signal backup.
type BackupFile struct {
	File      *bytes.Buffer
	FileSize  int
	CipherKey []byte
	MacKey    []byte
	Mac       hash.Hash
	IV        []byte
	Counter   uint32
}

// NewBackupFile initialises a backup file for reading using the provided path
// and password.
func NewBackupFile(path, password string) (*BackupFile, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open backup file")
	}

	fileBuf := bytes.NewBuffer(fileBytes)
	size := fileBuf.Len()

	headerLengthBytes := make([]byte, 4)
	_, err = io.ReadFull(fileBuf, headerLengthBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read headerLengthBytes")
	}
	headerLength := bytesToUint32(headerLengthBytes)

	headerFrame := make([]byte, headerLength)
	_, err = io.ReadFull(fileBuf, headerFrame)
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
		File:      fileBuf,
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
	if bf.File.Len() == 0 {
		return nil, errors.New("Nothing left to decode")
	}

	length := make([]byte, 4)
	io.ReadFull(bf.File, length)
	frameLength := bytesToUint32(length)
	defer rescue(fmt.Sprintf("frame: starting at %v, size %v; %v remaining in file", bf.FileSize-bf.File.Len(), frameLength, bf.File.Len()))

	frame := make([]byte, frameLength)
	io.ReadFull(bf.File, frame)

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

// DecryptAttachment reads the attachment immediately next in the file's bytes.
func (bf *BackupFile) DecryptAttachment(a *signal.Attachment, out io.Writer) error {
	uint32ToBytes(bf.IV, bf.Counter)
	bf.Counter++

	aesCipher, err := aes.NewCipher(bf.CipherKey)
	if err != nil {
		return errors.New("Bad cipher")
	}
	stream := cipher.NewCTR(aesCipher, bf.IV)
	bf.Mac.Write(bf.IV)

	buf := make([]byte, *a.Length)
	n, err := io.ReadFull(bf.File, buf)
	if err != nil {
		return errors.Wrap(err, "failed to read att")
	}
	if n != len(buf) {
		return errors.Errorf("didn't read enough bytes: %v, %v\n", n, len(buf))
	}
	bf.Mac.Write(buf)

	output := make([]byte, *a.Length)
	stream.XORKeyStream(output, buf)
	if _, err = out.Write(output); err != nil {
		return errors.Wrap(err, "can't write to output")
	}

	theirMac := make([]byte, 10)
	io.ReadFull(bf.File, theirMac)
	ourMac := bf.Mac.Sum(nil)

	if bytes.Equal(theirMac, ourMac) {
		return errors.New("Bad MAC")
	}

	return nil
}

// Slurp consumes the entire BackupFile and returns a list of all frames
// contained in the file. Note that after calling this function, the underlying
// file buffer will be empty and the file should be considered dropped. Calling
// any function on the backup file after calling Slurp will fail.
//
// Note that any attachments in the backup file will not be handled.
func (bf *BackupFile) Slurp() ([]*signal.BackupFrame, error) {
	frames := []*signal.BackupFrame{}
	for {
		f, err := bf.Frame()
		if err != nil {
			return frames, nil // TODO error matching
		}

		frames = append(frames, f)

		// Attachment needs removing
		if a := f.GetAttachment(); a != nil {
			err := bf.DecryptAttachment(a, ioutil.Discard)
			if err != nil {
				return nil, errors.Wrap(err, "unable to chew through attachment")
			}
		}
	}
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

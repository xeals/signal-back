package main

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
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/xeals/signal-back/signal"
	"golang.org/x/crypto/hkdf"
)

var protoCommitHash = "d6610f0"

type backupFile struct {
	File      *bytes.Buffer
	CipherKey []byte
	MacKey    []byte
	Mac       hash.Hash
	IV        []byte
	Counter   uint32
}

func newBackupFile(path, password string) (*backupFile, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open backup file")
	}

	fileBuf := bytes.NewBuffer(fileBytes)

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

	return &backupFile{
		File:      fileBuf,
		CipherKey: cipherKey,
		MacKey:    macKey,
		Mac:       hmac.New(crypto.SHA256.New, macKey),
		IV:        iv,
		Counter:   bytesToUint32(iv),
	}, nil
}

func (bf *backupFile) frame() (*signal.BackupFrame, error) {
	if bf.File.Len() == 0 {
		return nil, errors.New("Nothing left to decode")
	}

	length := make([]byte, 4)
	io.ReadFull(bf.File, length)
	frameLength := bytesToUint32(length)

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

func (bf *backupFile) decryptAttachment(a *signal.Attachment, out io.Writer) error {
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
		fmt.Println("failed to generate hashes:", err.Error())
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

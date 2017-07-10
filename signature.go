package rdiff

import (
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/md4"
)

const (
	BLAKE2_SUM_LENGTH = 32
	MD4_SUM_LENGTH    = 16
)

type SignatureType struct {
	sigType    MagicNumber
	blockLen   uint32
	strongLen  uint32
	strongSigs [][]byte
	weak2block map[uint32]int
}

func CalcStrongSum(data []byte, sigType MagicNumber, strongLen uint32) ([]byte, error) {
	switch sigType {
	case BLAKE2_SIG_MAGIC:
		d := blake2b.Sum256(data)
		return d[:strongLen], nil
	case MD4_SIG_MAGIC:
		d := md4.New()
		d.Write(data)
		return d.Sum(nil)[:strongLen], nil
	}
	return nil, fmt.Errorf("Invalid sigType %#x", sigType)
}

func Signature(input io.Reader, output io.Writer, blockLen, strongLen uint32, sigType MagicNumber) error {
	var maxStrongLen uint32

	switch sigType {
	case BLAKE2_SIG_MAGIC:
		maxStrongLen = BLAKE2_SUM_LENGTH
	case MD4_SIG_MAGIC:
		maxStrongLen = MD4_SUM_LENGTH
	default:
		return fmt.Errorf("invalid sigType %#x", sigType)
	}

	if strongLen > maxStrongLen {
		return fmt.Errorf("invalid strongLen %d for sigType %#x", strongLen, sigType)
	}

	err := binary.Write(output, binary.BigEndian, sigType)
	if err != nil {
		return err
	}
	err = binary.Write(output, binary.BigEndian, blockLen)
	if err != nil {
		return err
	}
	err = binary.Write(output, binary.BigEndian, strongLen)
	if err != nil {
		return err
	}

	block := make([]byte, blockLen)

	for {
		n, err := io.ReadFull(input, block)
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		data := block[:n]

		err = binary.Write(output, binary.BigEndian, WeakChecksum(data))
		if err != nil {
			return err
		}

		strong, _ := CalcStrongSum(data, sigType, strongLen)
		output.Write(strong)
	}

	return nil
}

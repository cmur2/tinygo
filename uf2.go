// Converts firmware files from BIN to UF2 format before flashing.
//
// For more information about the UF2 firmware file format, please see:
// https://github.com/Microsoft/uf2
//
//
package main

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
)

// ConvertELFFileToUF2File converts an ELF file to a UF2 file.
func ConvertELFFileToUF2File(infile, outfile string) error {
	// Read the .text segment.
	_, data, err := ExtractROM(infile)
	if err != nil {
		return err
	}

	output, _ := ConvertBinToUF2(data)
	return ioutil.WriteFile(outfile, output, 0644)
}

// ConvertBinToUF2 converts the binary bytes in input to UF2 formatted data.
func ConvertBinToUF2(input []byte) ([]byte, int) {
	blocks := split(input, 256)
	output := make([]byte, 0)

	bl := NewUF2Block()
	bl.SetNumBlocks(len(blocks))

	for i := 0; i < len(blocks); i++ {
		bl.SetBlockNo(i)
		bl.SetData(blocks[i])

		output = append(output, bl.Bytes()...)
		bl.IncrementAddress(bl.payloadSize)
	}

	return output, len(blocks)
}

const (
	uf2MagicStart0  = 0x0A324655 // "UF2\n"
	uf2MagicStart1  = 0x9E5D5157 // Randomly selected
	uf2MagicEnd     = 0x0AB16F30 // Ditto
	uf2StartAddress = 0x4000     // TODO: must allow this to be set via the board's json file
)

// UF2Block is the structure used for each UF2 code block sent to device.
type UF2Block struct {
	magicStart0 uint32
	magicStart1 uint32
	flags       uint32
	targetAddr  uint32
	payloadSize uint32
	blockNo     uint32
	numBlocks   uint32
	familyID    uint32
	data        []uint8
	magicEnd    uint32
}

// NewUF2Block returns a new UF2Block struct that has been correctly populated
func NewUF2Block() *UF2Block {
	return &UF2Block{magicStart0: uf2MagicStart0,
		magicStart1: uf2MagicStart1,
		magicEnd:    uf2MagicEnd,
		targetAddr:  uf2StartAddress,
		flags:       0x0,
		familyID:    0x0,
		payloadSize: 256,
		data:        make([]byte, 476),
	}
}

// Bytes converts the UF2Block to a slice of bytes that can be written to file.
func (b *UF2Block) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	binary.Write(buf, binary.LittleEndian, b.magicStart0)
	binary.Write(buf, binary.LittleEndian, b.magicStart1)
	binary.Write(buf, binary.LittleEndian, b.flags)
	binary.Write(buf, binary.LittleEndian, b.targetAddr)
	binary.Write(buf, binary.LittleEndian, b.payloadSize)
	binary.Write(buf, binary.LittleEndian, b.blockNo)
	binary.Write(buf, binary.LittleEndian, b.numBlocks)
	binary.Write(buf, binary.LittleEndian, b.familyID)
	binary.Write(buf, binary.LittleEndian, b.data)
	binary.Write(buf, binary.LittleEndian, b.magicEnd)

	return buf.Bytes()
}

// IncrementAddress moves the target address pointer forward by count bytes.
func (b *UF2Block) IncrementAddress(count uint32) {
	b.targetAddr += b.payloadSize
}

// SetData sets the data to be used for the current block.
func (b *UF2Block) SetData(d []byte) {
	b.data = make([]byte, 476)
	copy(b.data[:], d)
}

// SetBlockNo sets the current block number to be used.
func (b *UF2Block) SetBlockNo(bn int) {
	b.blockNo = uint32(bn)
}

// SetNumBlocks sets the total number of blocks for this UF2 file.
func (b *UF2Block) SetNumBlocks(total int) {
	b.numBlocks = uint32(total)
}

// split splits a slice of bytes into a slice of byte slices of a specific size limit.
func split(input []byte, limit int) [][]byte {
	var block []byte
	output := make([][]byte, 0, len(input)/limit+1)
	for len(input) >= limit {
		block, input = input[:limit], input[limit:]
		output = append(output, block)
	}
	if len(input) > 0 {
		output = append(output, input[:len(input)])
	}
	return output
}

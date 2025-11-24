package qrcode

import (
	"errors"
)

// Mode indicators
const (
	ModeNumeric      = 1
	ModeAlphanumeric = 2
	ModeByte         = 4
	ModeKanji        = 8
	ModeECI          = 7
)

// ECC Levels
const (
	LevelL = 1 // 7%
	LevelM = 0 // 15%
	LevelQ = 3 // 25%
	LevelH = 2 // 30%
)

// VersionInfo Version 1-40 info
type VersionInfo struct {
	TotalCodewords int
	ECCodewords    int
	Blocks         int // Number of blocks in group 1 (simplified for V1-V2)
	// For larger versions, there are groups. We will start with support for small versions.
	// We will implement dynamic lookup or just support V1 and V2 for "create and read again".
}

// Simplified table for Version 1 and 2, Level L/M
// Ref: https://www.thonky.com/qr-code-tutorial/error-correction-table
var versionTable = map[int]map[int]VersionInfo{
	1: {
		LevelL: {TotalCodewords: 26, ECCodewords: 7, Blocks: 1},
		LevelM: {TotalCodewords: 26, ECCodewords: 10, Blocks: 1},
		LevelQ: {TotalCodewords: 26, ECCodewords: 13, Blocks: 1},
		LevelH: {TotalCodewords: 26, ECCodewords: 17, Blocks: 1},
	},
	2: {
		LevelL: {TotalCodewords: 44, ECCodewords: 10, Blocks: 1},
		LevelM: {TotalCodewords: 44, ECCodewords: 16, Blocks: 1},
		LevelQ: {TotalCodewords: 44, ECCodewords: 22, Blocks: 1},
		LevelH: {TotalCodewords: 44, ECCodewords: 28, Blocks: 1},
	},
	3: {
		LevelL: {TotalCodewords: 70, ECCodewords: 15, Blocks: 1},
		LevelM: {TotalCodewords: 70, ECCodewords: 26, Blocks: 1},
		LevelQ: {TotalCodewords: 70, ECCodewords: 36, Blocks: 2}, // split not implemented
		LevelH: {TotalCodewords: 70, ECCodewords: 44, Blocks: 2}, // split not implemented
	},
	4: {
		LevelL: {TotalCodewords: 100, ECCodewords: 20, Blocks: 1},
		LevelM: {TotalCodewords: 100, ECCodewords: 36, Blocks: 2}, // split not implemented
		LevelQ: {TotalCodewords: 100, ECCodewords: 52, Blocks: 2}, // split not implemented
		LevelH: {TotalCodewords: 100, ECCodewords: 64, Blocks: 4}, // split not implemented
	},
	// Add more if needed.
}

type QRCode struct {
	Version int
	Level   int
	Size    int // Dimension (21 + 4*(V-1))
	Modules [][]bool
}

// NewQRCode creates a matrix for the given string.
// Currently defaults to Byte Mode.
func NewQRCode(content string, level int) (*QRCode, error) {
	// Analyze data and choose version.
	// Start with V1, if not fit, go V2.
	data := []byte(content)

	var v int
	var vInfo VersionInfo
	found := false

	// Try versions 1 to 4
	for ver := 1; ver <= 4; ver++ {
		info := versionTable[ver][level]
		if info.Blocks > 1 {
			// Skip versions requiring interleaving for this simplified implementation
			continue
		}

		// Capacity check
		// Byte mode: 4 bits mode + 8 bits count (V1-9) + 8*len
		// V1-9 count indicator is 8 bits.
		totalDataBits := 4 + 8 + len(data)*8
		if totalDataBits <= (info.TotalCodewords-info.ECCodewords)*8 {
			v = ver
			vInfo = info
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("content too long or requires block interleaving (not implemented)")
	}

	// Data Encoding
	bitBuffer := NewBitBuffer()
	bitBuffer.Put(ModeByte, 4)
	bitBuffer.Put(len(data), 8) // 8 bits for count in V1-V9
	for _, b := range data {
		bitBuffer.Put(int(b), 8)
	}

	// Terminator and Padding
	dataCapacityBits := (vInfo.TotalCodewords - vInfo.ECCodewords) * 8
	if bitBuffer.Len() < dataCapacityBits {
		// Terminator (up to 4 zeros)
		term := 4
		if bitBuffer.Len()+term > dataCapacityBits {
			term = dataCapacityBits - bitBuffer.Len()
		}
		bitBuffer.Put(0, term)
	}

	// Byte alignment
	if bitBuffer.Len()%8 != 0 {
		bitBuffer.Put(0, 8-(bitBuffer.Len()%8))
	}

	// Pad bytes
	padBytes := []int{0xEC, 0x11}
	padIdx := 0
	for bitBuffer.Len() < dataCapacityBits {
		bitBuffer.Put(padBytes[padIdx], 8)
		padIdx = (padIdx + 1) % 2
	}

	// Error Correction Coding
	dataCodewords := make([]int, 0)
	// Convert bits to bytes
	bits := bitBuffer.Bits
	for i := 0; i < len(bits); i += 8 {
		val := 0
		for j := 0; j < 8; j++ {
			if i+j < len(bits) && bits[i+j] {
				val |= 1 << (7 - j)
			}
		}
		dataCodewords = append(dataCodewords, val)
	}

	ecCodewords := CalculateECCodewords(dataCodewords, vInfo.ECCodewords)

	finalMessage := append(dataCodewords, ecCodewords...)

	// Matrix Construction
	qr := &QRCode{
		Version: v,
		Level:   level,
		Size:    21 + 4*(v-1),
	}
	qr.Modules = make([][]bool, qr.Size)
	for i := range qr.Modules {
		qr.Modules[i] = make([]bool, qr.Size)
	}

	// We need a way to track which modules are function patterns (reserved)
	isFunction := make([][]bool, qr.Size)
	for i := range isFunction {
		isFunction[i] = make([]bool, qr.Size)
	}

	// Finder Patterns
	addFinderPattern := func(r, c int) {
		for i := 0; i < 7; i++ {
			for j := 0; j < 7; j++ {
				if r+i >= qr.Size || c+j >= qr.Size || r+i < 0 || c+j < 0 {
					continue
				}
				isFunction[r+i][c+j] = true
				if i == 0 || i == 6 || j == 0 || j == 6 || (i >= 2 && i <= 4 && j >= 2 && j <= 4) {
					qr.Modules[r+i][c+j] = true
				} else {
					qr.Modules[r+i][c+j] = false
				}
			}
		}
	}

	addFinderPattern(0, 0)
	addFinderPattern(0, qr.Size-7)
	addFinderPattern(qr.Size-7, 0)

	// Separators (white space around finders)
	// Top Left
	for i := 0; i < 8; i++ {
		if i < qr.Size && 7 < qr.Size {
			isFunction[i][7] = true
			qr.Modules[i][7] = false
		}
		if i < qr.Size && 7 < qr.Size {
			isFunction[7][i] = true
			qr.Modules[7][i] = false
		}
	}
	// Top Right
	for i := 0; i < 8; i++ {
		if i < qr.Size && qr.Size-8 >= 0 {
			isFunction[i][qr.Size-8] = true
			qr.Modules[i][qr.Size-8] = false
		}
		if qr.Size-1-i >= 0 && 7 < qr.Size {
			isFunction[7][qr.Size-1-i] = true
			qr.Modules[7][qr.Size-1-i] = false
		}
	}
	// Bottom Left
	for i := 0; i < 8; i++ {
		if qr.Size-1-i >= 0 && 7 < qr.Size {
			isFunction[qr.Size-1-i][7] = true
			qr.Modules[qr.Size-1-i][7] = false
		}
		if i < qr.Size && qr.Size-8 >= 0 {
			isFunction[qr.Size-8][i] = true
			qr.Modules[qr.Size-8][i] = false
		}
	}

	// Alignment Patterns (For V2+)
	if v >= 2 {
		// Locations depend on version. Simplified for V2-V4.
		// V2: 6, 18
		// V3: 6, 22
		// V4: 6, 26
		// Note: The '6' is implicitly handled by finder patterns exclusion usually, but we need to place at intersections.
		// Locations list includes 6 but 6 overlaps with finder.

		var locs []int
		switch v {
		case 2:
			locs = []int{6, 18}
		case 3:
			locs = []int{6, 22}
		case 4:
			locs = []int{6, 26}
		}

		for _, cx := range locs {
			for _, cy := range locs {
				// If overlaps with finder patterns, skip.
				// Finders are 0..8 (inclusive of separator)
				if (cx < 9 && cy < 9) || (cx < 9 && cy > qr.Size-9) || (cx > qr.Size-9 && cy < 9) {
					continue
				}

				// Draw Alignment Pattern 5x5
				for i := -2; i <= 2; i++ {
					for j := -2; j <= 2; j++ {
						r, c := cy+i, cx+j
						if !isFunction[r][c] {
							isFunction[r][c] = true
							if i == -2 || i == 2 || j == -2 || j == 2 || (i == 0 && j == 0) {
								qr.Modules[r][c] = true
							} else {
								qr.Modules[r][c] = false
							}
						}
					}
				}
			}
		}
	}

	// Timing Patterns
	for i := 8; i < qr.Size-8; i++ {
		if !isFunction[6][i] {
			isFunction[6][i] = true
			qr.Modules[6][i] = (i%2 == 0)
		}
		if !isFunction[i][6] {
			isFunction[i][6] = true
			qr.Modules[i][6] = (i%2 == 0)
		}
	}

	// Dark Module
	isFunction[qr.Size-8][8] = true
	qr.Modules[qr.Size-8][8] = true

	// Reserve Format Information areas
	// Around Top-Left Finder
	for i := 0; i < 9; i++ {
		if !isFunction[8][i] {
			isFunction[8][i] = true
		} // Horizontal
		if !isFunction[i][8] {
			isFunction[i][8] = true
		} // Vertical
	}
	// Below Top-Right Finder
	for i := 0; i < 8; i++ {
		if !isFunction[8][qr.Size-1-i] {
			isFunction[8][qr.Size-1-i] = true
		}
	}
	// Right of Bottom-Left Finder
	for i := 0; i < 7; i++ {
		if !isFunction[qr.Size-1-i][8] {
			isFunction[qr.Size-1-i][8] = true
		}
	}

	// Place Data
	// Zig-zag pattern
	idx := 0
	totalBits := len(finalMessage) * 8

	// Simple Mask Pattern 0: (row + col) % 2 == 0 (Checkerboard)
	// We will use mask 0 strictly for now to simplify.
	maskPattern := 0

	// Helper to get bit from message
	getBit := func(k int) bool {
		byteIdx := k / 8
		bitIdx := 7 - (k % 8)
		return (finalMessage[byteIdx]>>bitIdx)&1 == 1
	}

	for col := qr.Size - 1; col > 0; col -= 2 {
		if col == 6 {
			col--
		} // Skip timing pattern

		for rowIter := 0; rowIter < qr.Size; rowIter++ {
			r := rowIter
			if ((col+1)/2)%2 == 0 { // Upwards
				r = qr.Size - 1 - rowIter
			}

			for c := col; c > col-2; c-- {
				if !isFunction[r][c] {
					bit := false
					if idx < totalBits {
						bit = getBit(idx)
						idx++
					}
					// Apply mask 0: (row + column) % 2 == 0
					mask := (r+c)%2 == 0
					if mask {
						bit = !bit
					}
					qr.Modules[r][c] = bit
				}
			}
		}
	}

	// Format Information
	// ECC Level (2 bits) + Mask Pattern (3 bits)
	// L=01, M=00, Q=11, H=10. Re-mapped:
	// LevelL is 1 -> 01
	// LevelM is 0 -> 00
	// LevelQ is 3 -> 11
	// LevelH is 2 -> 10
	ecBits := 0
	switch level {
	case LevelL:
		ecBits = 1
	case LevelM:
		ecBits = 0
	case LevelQ:
		ecBits = 3
	case LevelH:
		ecBits = 2
	}

	formatData := (ecBits << 3) | maskPattern
	formatPoly := calculateBCHFormat(formatData)

	// Place Format bits (15 bits)
	// Bit 0 is LSB, Bit 14 is MSB.
	// Standard Placement (Top Left):
	// (8,0) -> Bit 14
	// (8,1) -> Bit 13
	// (8,2) -> Bit 12
	// (8,3) -> Bit 11
	// (8,4) -> Bit 10
	// (8,5) -> Bit 9
	// (8,7) -> Bit 8 (Skip 6)
	// (8,8) -> Bit 7
	// (7,8) -> Bit 6
	// (5,8) -> Bit 5
	// (4,8) -> Bit 4
	// (3,8) -> Bit 3
	// (2,8) -> Bit 2
	// (1,8) -> Bit 1
	// (0,8) -> Bit 0

	setModule := func(r, c int, b bool) {
		qr.Modules[r][c] = b
	}

	for i := 0; i < 15; i++ {
		bit := (formatPoly>>i)&1 == 1 // i is bit index (0=LSB)

		// Top Left
		switch i {
		case 0:
			setModule(0, 8, bit)
		case 1:
			setModule(1, 8, bit)
		case 2:
			setModule(2, 8, bit)
		case 3:
			setModule(3, 8, bit)
		case 4:
			setModule(4, 8, bit)
		case 5:
			setModule(5, 8, bit)
		case 6:
			setModule(7, 8, bit) // Skip timing (6,8)
		case 7:
			setModule(8, 8, bit)
		case 8:
			setModule(8, 7, bit) // Skip timing (8,6)
		case 9:
			setModule(8, 5, bit)
		case 10:
			setModule(8, 4, bit)
		case 11:
			setModule(8, 3, bit)
		case 12:
			setModule(8, 2, bit)
		case 13:
			setModule(8, 1, bit)
		case 14:
			setModule(8, 0, bit)
		}

		// Copies
		// Bits 0-7: (8, Size-1) -> Bit 0 ... (8, Size-8) -> Bit 7
		// Bits 8-14: (Size-8, 8) -> Bit 8 ... (Size-1, 8) -> Bit 14

		if i < 8 {
			setModule(8, qr.Size-1-i, bit)
		} else {
			setModule(qr.Size-8+(i-8), 8, bit)
		}
	}
	// Dark Module fixed at [Size-8][8] is already set

	return qr, nil
}

func calculateBCHFormat(data int) int {
	d := data << 10
	// Generator 10100110111 (0x537)
	g := 0x537

	// Simple division
	for i := 4; i >= 0; i-- {
		if (d>>(i+10))&1 == 1 {
			d ^= (g << i)
		}
	}

	// Mask string 101010000010010 (0x5412)
	return ((data << 10) | d) ^ 0x5412
}

// BitBuffer helper
type BitBuffer struct {
	Bits []bool
}

func NewBitBuffer() *BitBuffer {
	return &BitBuffer{Bits: []bool{}}
}

func (b *BitBuffer) Put(num, length int) {
	for i := 0; i < length; i++ {
		b.Bits = append(b.Bits, ((num>>(length-1-i))&1) == 1)
	}
}

func (b *BitBuffer) Len() int {
	return len(b.Bits)
}

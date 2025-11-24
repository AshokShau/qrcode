package qrcode

// Galois Field (256) logic for QR Code Reed-Solomon error correction.
// Primitive Polynomial: x^8 + x^4 + x^3 + x^2 + 1 (0x11D or 285)

var (
	expTable [256]int
	logTable [256]int
)

func init() {
	val := 1
	for i := 0; i < 255; i++ {
		expTable[i] = val
		logTable[val] = i
		val *= 2
		if val >= 256 {
			val ^= 0x11D
		}
	}
	// logTable[0] is technically undefined but often unused or handled separately
}

func gfAdd(x, y int) int {
	return x ^ y
}

func gfSub(x, y int) int {
	return x ^ y
}

func gfMul(x, y int) int {
	if x == 0 || y == 0 {
		return 0
	}
	return expTable[(logTable[x]+logTable[y])%255]
}

func gfDiv(x, y int) int {
	if y == 0 {
		panic("division by zero")
	}
	if x == 0 {
		return 0
	}
	return expTable[(logTable[x]+255-logTable[y])%255]
}

func gfPolyMul(p, q []int) []int {
	res := make([]int, len(p)+len(q)-1)
	for i := 0; i < len(p); i++ {
		for j := 0; j < len(q); j++ {
			res[i+j] ^= gfMul(p[i], q[j])
		}
	}
	return res
}

// GenerateGeneratorPoly creates a generator polynomial for the given number of error correction codewords.
func GenerateGeneratorPoly(numECCodewords int) []int {
	gen := []int{1}
	for i := 0; i < numECCodewords; i++ {
		gen = gfPolyMul(gen, []int{1, expTable[i]})
	}
	return gen
}

// CalculateECCodewords generates error correction codewords for the given data.
func CalculateECCodewords(data []int, numECCodewords int) []int {
	generator := GenerateGeneratorPoly(numECCodewords)

	// Initialize message polynomial: data padded with zeros for ECC
	msg := make([]int, len(data)+numECCodewords)
	copy(msg, data)

	// Polynomial division: msg / generator
	// We only care about the remainder.

	// We perform division by XORing the generator scaled to the leading coefficient
	// The "remainder" will end up in the last numECCodewords positions of what would be the register.
	// But simpler: standard polynomial division simulation.

	// Copy data to a working buffer that is large enough to hold the result (which is the remainder)
	// Actually, standard algorithm usually works on a buffer of size max(len(gen), len(data)) or similar.
	// Let's use the standard shift-register approach.

	remainder := make([]int, len(data)+numECCodewords)
	copy(remainder, data)

	for i := 0; i < len(data); i++ {
		coef := remainder[i]
		if coef != 0 {
			for j := 0; j < len(generator); j++ {
				// We don't need to modify the part we've already passed (index i),
				// but for the math to be strictly correct in a full poly div we would.
				// Optimization: only XOR the terms that overlap.
				// generator[0] is always 1, so we are effectively zeroing remainder[i].
				// But we can just process the part that affects the remainder.
				remainder[i+j] ^= gfMul(generator[j], coef)
			}
		}
	}

	// The remainder is the last numECCodewords bytes
	return remainder[len(data):]
}

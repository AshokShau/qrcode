# QR Code Generator and Decoder in Go

A pure Go library for generating and decoding QR codes. This implementation supports various QR code versions, error correction levels, and encoding modes.

## Features

- 🚀 Generate QR codes in different versions (1-40)
- 🔒 Multiple error correction levels (L, M, Q, H)
- 🔄 Supports different encoding modes (Numeric, Alphanumeric, Byte, Kanji)
- 🔍 QR code decoding from images
- 🛠 Pure Go implementation with no external dependencies
- 🎯 Simple and easy-to-use API

## Installation

```bash
go get github.com/ashokshau/qrcode
```

## Quick Start

### Generating a QR Code

```go
package main

import (
	"os"
	
	"github.com/ashokshau/qrcode"
)

func main() {
	// Create a new QR code with content and error correction level
	qr, err := qrcode.NewQRCode("https://example.com", qrcode.LevelM)
	if err != nil {
		panic(err)
	}

	// Save the QR code as a PNG file
	file, err := os.Create("qrcode.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := qr.EncodePNG(file); err != nil {
		panic(err)
	}
}
```

### Decoding a QR Code

```go
package main

import (
	"fmt"
	"os"
	
	"github.com/ashokshau/qrcode"
)

func main() {
	file, err := os.Open("qrcode.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	content, err := qrcode.Decode(file)
	if err != nil {
		panic(err)
	}

	fmt.Println("Decoded content:", content)
}
```

## Error Correction Levels

| Level | Error Correction Capability  |
|-------|------------------------------|
| L     | ~7% of data can be restored  |
| M     | ~15% of data can be restored |
| Q     | ~25% of data can be restored |
| H     | ~30% of data can be restored |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- QR Code is a registered trademark of Denso Wave Incorporated
- This implementation is based on the QR Code specification (ISO/IEC 18004:2015)

package main

import (
	"fmt"
	"os"

	"github.com/ashokshau/qrcode"
)

func main() {
	// The content to encode
	content := "https://www.google.com"
	filename := "test_qr.png"

	fmt.Printf("Generating QR code for: %s\n", content)

	// Create the QR Code object
	// LevelM is a good balance (15% error correction)
	qr, err := qrcode.NewQRCode(content, qrcode.LevelM)
	if err != nil {
		fmt.Printf("Error creating QR: %v\n", err)
		return
	}

	// Open file for writing
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	// Write PNG to file
	// Scale 10 means each module (dot) is 10x10 pixels
	err = qr.WritePNG(file, 10)
	if err != nil {
		fmt.Printf("Error writing PNG: %v\n", err)
		return
	}

	fmt.Printf("Successfully saved QR code to %s\n", filename)

	// Verify by reading it back
	fmt.Println("Verifying by reading the file back...")

	readFile, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer readFile.Close()

	decoded, err := qrcode.Decode(readFile)
	if err != nil {
		fmt.Printf("Error decoding QR: %v\n", err)
		return
	}

	fmt.Printf("Decoded content: %s\n", decoded)

	if decoded == content {
		fmt.Println("SUCCESS: Decoded content matches original!")
	} else {
		fmt.Println("FAILURE: Content mismatch.")
	}
}

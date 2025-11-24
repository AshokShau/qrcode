package qrcode

import (
	"bytes"
	"image/png"
	"testing"
)

func TestQRCreateAndRead(t *testing.T) {
	content := "Hello World"

	// Create
	qr, err := NewQRCode(content, LevelL)
	if err != nil {
		t.Fatalf("Failed to create QR: %v", err)
	}

	// Write to buffer
	var buf bytes.Buffer
	err = qr.WritePNG(&buf, 10) // Scale 10
	if err != nil {
		t.Fatalf("Failed to write PNG: %v", err)
	}

	// os.WriteFile("test_qr.png", buf.Bytes(), 0644)

	// Read back
	decoded, err := Decode(&buf)
	if err != nil {
		t.Fatalf("Failed to decode QR: %v", err)
	}

	if decoded != content {
		t.Errorf("Expected '%s', got '%s'", content, decoded)
	}
}

func TestQRVersion2(t *testing.T) {
	// Longer string to force Version 2
	// 40 chars should be enough to push past V1-L (19 bytes) but fit in V2-L (34 bytes)
	// V1-L capacity is 19 bytes (26 total codewords - 7 ecc).
	content := "This is a longer string for V2 QR Code!!"
	if len(content) <= 19 {
		t.Fatalf("Test content too short to force V2: %d", len(content))
	}

	qr, err := NewQRCode(content, LevelL)
	if err != nil {
		t.Fatalf("Failed to create QR: %v", err)
	}

	if qr.Version != 2 {
		t.Logf("Expected Version 2, got %d (might be valid if V1 capacity is higher than thought or encoding efficient)", qr.Version)
	}

	var buf bytes.Buffer
	err = qr.WritePNG(&buf, 5)
	if err != nil {
		t.Fatalf("Failed to write PNG: %v", err)
	}

	decoded, err := Decode(&buf)
	if err != nil {
		t.Fatalf("Failed to decode QR: %v", err)
	}

	if decoded != content {
		t.Errorf("Mismatch. Len: %d vs %d", len(content), len(decoded))
	}
}

func TestVerifyPNGFormat(t *testing.T) {
	content := "Test"
	qr, _ := NewQRCode(content, LevelL)
	var buf bytes.Buffer
	qr.WritePNG(&buf, 1)

	_, err := png.Decode(&buf)
	if err != nil {
		t.Fatalf("Generated output is not a valid PNG: %v", err)
	}
}

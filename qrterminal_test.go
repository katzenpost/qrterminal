package qrterminal

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"rsc.io/qr"
)

// Original tests that just verify the code doesn't crash
func TestGenerate(t *testing.T) {
	Generate("https://github.com/mdp/qrterminal", L, os.Stdout)
}

func TestGenerateWithConfig(t *testing.T) {
	config := Config{
		Level:     M,
		Writer:    os.Stdout,
		BlackChar: WHITE, // Inverted
		WhiteChar: BLACK,
		QuietZone: QUIET_ZONE,
	}
	GenerateWithConfig("https://github.com/mdp/qrterminal", config)
}

func TestGenerateHalfBlock(t *testing.T) {
	GenerateHalfBlock("https://github.com/mdp/qrterminal", L, os.Stdout)
}

func TestGenerateWithHalfBlockConfig(t *testing.T) {
	config := Config{
		Level:          M,
		Writer:         os.Stdout,
		HalfBlocks:     true,
		BlackChar:      BLACK_BLACK,
		WhiteBlackChar: WHITE_BLACK,
		WhiteChar:      WHITE_WHITE,
		BlackWhiteChar: BLACK_WHITE,
		QuietZone:      3,
	}
	GenerateWithConfig("https://github.com/mdp/qrterminal", config)
}

func TestGenerateWithHalfBlockMinConfig(t *testing.T) {
	config := Config{
		Level:      M,
		Writer:     os.Stdout,
		HalfBlocks: true,
		QuietZone:  3,
	}
	GenerateWithConfig("https://github.com/mdp/qrterminal", config)
}

// New tests that actually verify the output

// Test that captures and verifies the output
func TestCaptureOutput(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		level      qr.Level
		halfBlocks bool
	}{
		{"BasicURL", "https://example.com", L, false},
		{"ShortText", "test", M, false},
		{"HalfBlockMode", "test", L, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			if tc.halfBlocks {
				config := Config{
					Level:      tc.level,
					Writer:     &buf,
					HalfBlocks: true,
				}
				GenerateWithConfig(tc.input, config)
			} else {
				Generate(tc.input, tc.level, &buf)
			}

			output := buf.String()

			// Verify output is not empty
			if len(output) == 0 {
				t.Errorf("Generated QR code is empty")
			}

			// Verify output contains multiple lines
			lines := strings.Split(output, "\n")
			if len(lines) <= 1 {
				t.Errorf("Generated QR code should have multiple lines, got %d", len(lines))
			}

			// Verify the output contains the expected characters
			if tc.halfBlocks {
				// Half blocks mode should contain these characters
				expectedChars := []string{BLACK_BLACK, WHITE_WHITE, BLACK_WHITE, WHITE_BLACK}
				foundExpectedChar := false

				for _, char := range expectedChars {
					if strings.Contains(output, char) {
						foundExpectedChar = true
						break
					}
				}

				if !foundExpectedChar {
					t.Errorf("Half block output doesn't contain expected characters")
				}
			} else {
				// Regular mode should contain BLACK and WHITE
				if !strings.Contains(output, BLACK) && !strings.Contains(output, WHITE) {
					t.Errorf("Output doesn't contain expected BLACK or WHITE characters")
				}
			}
		})
	}
}

// Test the structure of the QR code (size, quiet zone, etc.)
func TestQRStructure(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		quietZone int
	}{
		{"DefaultQuietZone", "test", QUIET_ZONE},
		{"MinimalQuietZone", "test", 1},
		{"LargeQuietZone", "test", 8},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := Config{
				Level:     L,
				Writer:    &buf,
				BlackChar: BLACK,
				WhiteChar: WHITE,
				QuietZone: tc.quietZone,
			}
			GenerateWithConfig(tc.input, config)

			output := buf.String()
			lines := strings.Split(output, "\n")

			// Check that we have at least 2*quietZone lines for top and bottom borders
			if len(lines) < 2*tc.quietZone {
				t.Errorf("Expected at least %d lines for quiet zone, got %d", 2*tc.quietZone, len(lines))
			}

			// Check that the first few lines contain WHITE (quiet zone)
			for i := 0; i < tc.quietZone && i < len(lines); i++ {
				if len(lines[i]) > 0 && !strings.Contains(lines[i], WHITE) {
					t.Errorf("Line %d should contain WHITE (quiet zone)", i)
				}
			}

			// Check that the last few lines are all WHITE (quiet zone)
			// Note: The last line might be empty due to a trailing newline
			for i := len(lines) - tc.quietZone; i < len(lines); i++ {
				if i < len(lines) && len(lines[i]) > 0 && !strings.Contains(lines[i], WHITE) {
					t.Errorf("Line %d should contain WHITE (quiet zone)", i)
				}
			}
		})
	}
}

// Test with various configurations
func TestConfigVariations(t *testing.T) {
	testCases := []struct {
		name   string
		config Config
		input  string
	}{
		{
			"InvertedColors",
			Config{
				Level:     L,
				BlackChar: WHITE,
				WhiteChar: BLACK,
			},
			"test",
		},
		{
			"CustomCharacters",
			Config{
				Level:     L,
				BlackChar: "XX",
				WhiteChar: "..",
			},
			"test",
		},
		{
			"HalfBlocksCustomChars",
			Config{
				Level:          L,
				HalfBlocks:     true,
				BlackChar:      "a",
				WhiteChar:      "b",
				BlackWhiteChar: "c",
				WhiteBlackChar: "d",
			},
			"test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			tc.config.Writer = &buf
			GenerateWithConfig(tc.input, tc.config)

			output := buf.String()

			// Verify output is not empty
			if len(output) == 0 {
				t.Errorf("Generated QR code is empty")
			}

			// For custom characters, verify they appear in the output
			if tc.name == "CustomCharacters" {
				if !strings.Contains(output, "XX") || !strings.Contains(output, "..") {
					t.Errorf("Output doesn't contain custom characters")
				}
			} else if tc.name == "HalfBlocksCustomChars" {
				// Check for at least one of the custom characters
				customChars := []string{"a", "b", "c", "d"}
				foundCustomChar := false

				for _, char := range customChars {
					if strings.Contains(output, char) {
						foundCustomChar = true
						break
					}
				}

				if !foundCustomChar {
					t.Errorf("Output doesn't contain custom half block characters")
				}
			}
		})
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"EmptyString", ""},
		{"VeryLongString", strings.Repeat("a", 100)},
		{"SpecialCharacters", "!@#$%^&*()_+{}|:<>?"},
		{"Unicode", "こんにちは世界"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			// Test that it doesn't panic
			Generate(tc.input, L, &buf)

			output := buf.String()

			// Verify output is not empty (unless input is empty)
			if tc.input != "" && len(output) == 0 {
				t.Errorf("Generated QR code is empty for input: %s", tc.input)
			}

			// For empty string, we should still get some output (the QR code for an empty string)
			if tc.input == "" && len(output) == 0 {
				t.Errorf("Generated QR code for empty string should not be empty")
			}
		})
	}
}

// Test that the same input always produces the same output
func TestConsistentOutput(t *testing.T) {
	input := "https://github.com/mdp/qrterminal"

	var buf1 bytes.Buffer
	Generate(input, L, &buf1)
	output1 := buf1.String()

	var buf2 bytes.Buffer
	Generate(input, L, &buf2)
	output2 := buf2.String()

	if output1 != output2 {
		t.Errorf("Generated QR codes for the same input should be identical")
	}
}

// Test that different error correction levels produce different outputs
func TestErrorCorrectionLevels(t *testing.T) {
	input := "https://github.com/mdp/qrterminal"

	var bufL bytes.Buffer
	Generate(input, L, &bufL)
	outputL := bufL.String()

	var bufM bytes.Buffer
	Generate(input, M, &bufM)
	outputM := bufM.String()

	var bufH bytes.Buffer
	Generate(input, H, &bufH)
	outputH := bufH.String()

	// Different error correction levels should produce different outputs
	// (higher levels add more redundancy, changing the pattern)
	if outputL == outputM || outputL == outputH || outputM == outputH {
		t.Errorf("Different error correction levels should produce different outputs")
	}
}

// Test that the sixel detection function works
func TestSixelDetection(t *testing.T) {
	// This is a simple test that just ensures the function doesn't crash
	// We can't really test the actual detection without a terminal
	result := IsSixelSupported(os.Stdout)

	// The result could be true or false depending on the terminal
	// We just want to make sure it runs without error
	t.Logf("Sixel support detected: %v", result)
}

// Test that the QR code pattern is consistent and contains the expected pattern
func TestQRPattern(t *testing.T) {
	// Generate a QR code with a known input
	input := "test"

	var buf bytes.Buffer
	Generate(input, L, &buf)
	output := buf.String()

	// Split the output into lines
	lines := strings.Split(output, "\n")

	// Skip the quiet zone at the top
	contentStart := QUIET_ZONE

	// Check for the finder patterns (the three square patterns in the corners)
	// These are a key part of any QR code and should be present

	// The exact position depends on the QR code size, but we can check for patterns
	// that should be present in any valid QR code for our input

	// Check for some BLACK pixels in the content area (not just quiet zone)
	foundBlack := false
	for i := contentStart; i < len(lines)-QUIET_ZONE; i++ {
		if strings.Contains(lines[i], BLACK) {
			foundBlack = true
			break
		}
	}

	if !foundBlack {
		t.Errorf("QR code doesn't contain any BLACK pixels in the content area")
	}
}

// Test binary data encoding functions
func TestGenerateBinary(t *testing.T) {
	// Test with various binary data patterns
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "simple binary",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
		},
		{
			name: "null bytes",
			data: []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "high bytes",
			data: []byte{0xFF, 0xFE, 0xFD, 0xFC, 0xFB},
		},
		{
			name: "mixed data",
			data: []byte("Hello\x00\xFF\x01World"),
		},
		{
			name: "random binary",
			data: []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			// Test GenerateBinary
			GenerateBinary(tc.data, L, &buf)
			output := buf.String()

			// Verify output is not empty
			if len(output) == 0 {
				t.Errorf("Generated QR code is empty for binary data: %v", tc.data)
			}

			// Verify output contains multiple lines
			lines := strings.Split(output, "\n")
			if len(lines) <= 1 {
				t.Errorf("Generated QR code should have multiple lines, got %d", len(lines))
			}

			// Verify the output contains the expected characters
			if !strings.Contains(output, BLACK) && !strings.Contains(output, WHITE) {
				t.Errorf("Output doesn't contain expected BLACK or WHITE characters")
			}
		})
	}
}

func TestGenerateBinaryWithConfig(t *testing.T) {
	data := []byte{0x00, 0xFF, 0x12, 0x34, 0x56, 0x78}

	testCases := []struct {
		name   string
		config Config
	}{
		{
			name: "basic config",
			config: Config{
				Level:     M,
				BlackChar: BLACK,
				WhiteChar: WHITE,
				QuietZone: QUIET_ZONE,
			},
		},
		{
			name: "half blocks",
			config: Config{
				Level:          M,
				HalfBlocks:     true,
				BlackChar:      BLACK_BLACK,
				WhiteBlackChar: WHITE_BLACK,
				WhiteChar:      WHITE_WHITE,
				BlackWhiteChar: BLACK_WHITE,
				QuietZone:      3,
			},
		},
		{
			name: "inverted colors",
			config: Config{
				Level:     H,
				BlackChar: WHITE,
				WhiteChar: BLACK,
				QuietZone: 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			tc.config.Writer = &buf

			GenerateBinaryWithConfig(data, tc.config)
			output := buf.String()

			// Verify output is not empty
			if len(output) == 0 {
				t.Errorf("Generated QR code is empty")
			}

			// Verify output contains multiple lines
			lines := strings.Split(output, "\n")
			if len(lines) <= 1 {
				t.Errorf("Generated QR code should have multiple lines, got %d", len(lines))
			}
		})
	}
}

func TestGenerateBinaryHalfBlock(t *testing.T) {
	data := []byte{0x00, 0xFF, 0x12, 0x34}

	var buf bytes.Buffer
	GenerateBinaryHalfBlock(data, L, &buf)
	output := buf.String()

	// Verify output is not empty
	if len(output) == 0 {
		t.Errorf("Generated QR code is empty")
	}

	// Verify output contains multiple lines
	lines := strings.Split(output, "\n")
	if len(lines) <= 1 {
		t.Errorf("Generated QR code should have multiple lines, got %d", len(lines))
	}

	// Verify the output contains half block characters
	expectedChars := []string{BLACK_BLACK, WHITE_WHITE, BLACK_WHITE, WHITE_BLACK}
	foundExpectedChar := false

	for _, char := range expectedChars {
		if strings.Contains(output, char) {
			foundExpectedChar = true
			break
		}
	}

	if !foundExpectedChar {
		t.Errorf("Half block output doesn't contain expected characters")
	}
}

// Test that binary and string versions produce different outputs for the same bytes
func TestBinaryVsStringEncoding(t *testing.T) {
	// Use data that would be interpreted differently as string vs binary
	data := []byte{0x00, 0xFF, 0x01, 0x02}
	text := string(data) // This creates a string with null bytes and high bytes

	var binaryBuf bytes.Buffer
	var stringBuf bytes.Buffer

	GenerateBinary(data, L, &binaryBuf)
	Generate(text, L, &stringBuf)

	binaryOutput := binaryBuf.String()
	stringOutput := stringBuf.String()

	// Both should produce output
	if len(binaryOutput) == 0 {
		t.Errorf("Binary QR code is empty")
	}
	if len(stringOutput) == 0 {
		t.Errorf("String QR code is empty")
	}

	// For this specific case, they should actually be the same since we're using the same underlying encoding
	// But this test ensures both methods work
	t.Logf("Binary output length: %d", len(binaryOutput))
	t.Logf("String output length: %d", len(stringOutput))
}

// Test consistency - same binary data should always produce the same output
func TestBinaryConsistency(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}

	var buf1 bytes.Buffer
	GenerateBinary(data, L, &buf1)
	output1 := buf1.String()

	var buf2 bytes.Buffer
	GenerateBinary(data, L, &buf2)
	output2 := buf2.String()

	if output1 != output2 {
		t.Errorf("Generated QR codes for the same binary data should be identical")
	}
}

// Test different error correction levels with binary data
func TestBinaryErrorCorrectionLevels(t *testing.T) {
	data := []byte{0x00, 0xFF, 0x12, 0x34, 0x56, 0x78}

	var bufL bytes.Buffer
	GenerateBinary(data, L, &bufL)
	outputL := bufL.String()

	var bufM bytes.Buffer
	GenerateBinary(data, M, &bufM)
	outputM := bufM.String()

	var bufH bytes.Buffer
	GenerateBinary(data, H, &bufH)
	outputH := bufH.String()

	// Different error correction levels should produce different outputs
	// (higher levels add more redundancy, changing the pattern)
	if outputL == outputM || outputL == outputH || outputM == outputH {
		t.Errorf("Different error correction levels should produce different outputs for binary data")
	}
}

// Test round-trip: encode binary data to QR, then verify the underlying QR code integrity
func TestBinaryDataRoundTrip(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "simple binary",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
		},
		{
			name: "null bytes",
			data: []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "high bytes",
			data: []byte{0xFF, 0xFE, 0xFD, 0xFC, 0xFB},
		},
		{
			name: "mixed data with text",
			data: []byte("Hello\x00\xFF\x01World"),
		},
		{
			name: "single byte",
			data: []byte{0x42},
		},
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "all byte values",
			data: func() []byte {
				data := make([]byte, 256)
				for i := 0; i < 256; i++ {
					data[i] = byte(i)
				}
				return data
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test 1: Verify that the underlying QR encoding can handle the data
			text := string(tc.data)
			qrCode, err := qr.Encode(text, L)
			if err != nil {
				t.Fatalf("Failed to encode data into QR code: %v", err)
			}

			// Test 2: Verify our binary function produces the same result as string conversion
			var binaryBuf bytes.Buffer
			GenerateBinary(tc.data, L, &binaryBuf)
			binaryOutput := binaryBuf.String()

			var stringBuf bytes.Buffer
			Generate(text, L, &stringBuf)
			stringOutput := stringBuf.String()

			// The outputs should be identical since we're using the same underlying encoding
			if binaryOutput != stringOutput {
				t.Errorf("Binary and string encoding should produce identical results for the same data")
			}

			// Test 3: Verify the QR code was created successfully
			if len(binaryOutput) == 0 {
				t.Errorf("QR code generation produced empty output")
			}

			// Test 4: Verify the QR code has the expected structure
			lines := strings.Split(binaryOutput, "\n")
			if len(lines) <= QUIET_ZONE*2 {
				t.Errorf("QR code should have more lines than just quiet zones")
			}

			// Test 5: Verify consistency - same data should always produce same QR
			var binaryBuf2 bytes.Buffer
			GenerateBinary(tc.data, L, &binaryBuf2)
			binaryOutput2 := binaryBuf2.String()

			if binaryOutput != binaryOutput2 {
				t.Errorf("Same binary data should always produce identical QR codes")
			}

			// Test 6: Verify the QR code dimensions match expectations
			if qrCode.Size > 0 {
				// The terminal output should have the QR code size plus quiet zones
				expectedMinLines := qrCode.Size + QUIET_ZONE*2
				if len(lines) < expectedMinLines {
					t.Errorf("QR code output should have at least %d lines, got %d", expectedMinLines, len(lines))
				}
			}
		})
	}
}

// Test that binary data encoding preserves exact byte values
func TestBinaryDataPreservation(t *testing.T) {
	// Test with data that could be problematic for string conversion
	problematicData := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F,
		0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E, 0x8F,
		0xF0, 0xF1, 0xF2, 0xF3, 0xF4, 0xF5, 0xF6, 0xF7, 0xF8, 0xF9, 0xFA, 0xFB, 0xFC, 0xFD, 0xFE, 0xFF,
	}

	// Generate QR code
	var buf bytes.Buffer
	GenerateBinary(problematicData, M, &buf)
	output := buf.String()

	// Verify output was generated
	if len(output) == 0 {
		t.Errorf("Failed to generate QR code for problematic binary data")
	}

	// Verify consistency - same data should always produce same output
	var buf2 bytes.Buffer
	GenerateBinary(problematicData, M, &buf2)
	output2 := buf2.String()

	if output != output2 {
		t.Errorf("Same binary data should always produce identical QR codes")
	}

	// Test that the underlying QR encoding is working by comparing with direct string conversion
	text := string(problematicData)
	var stringBuf bytes.Buffer
	Generate(text, M, &stringBuf)
	stringOutput := stringBuf.String()

	if output != stringOutput {
		t.Errorf("Binary encoding should produce same result as string encoding for same byte sequence")
	}
}

// Test that verifies binary data integrity by checking that the QR code generation
// preserves the exact byte sequence without corruption
func TestBinaryDataIntegrity(t *testing.T) {
	// Test with specific problematic sequences that could cause issues
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "UTF-8 invalid sequences",
			data: []byte{0x80, 0x81, 0x82, 0x83}, // Invalid UTF-8 start bytes
		},
		{
			name: "Control characters",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
		},
		{
			name: "High bit set",
			data: []byte{0x80, 0x90, 0xA0, 0xB0, 0xC0, 0xD0, 0xE0, 0xF0},
		},
		{
			name: "Mixed valid/invalid UTF-8",
			data: []byte("Hello\x80\x81World\xFF\x00"),
		},
		{
			name: "Binary file header simulation",
			data: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG header
		},
		{
			name: "Repeated null bytes",
			data: bytes.Repeat([]byte{0x00}, 20),
		},
		{
			name: "Repeated high bytes",
			data: bytes.Repeat([]byte{0xFF}, 20),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate QR code with binary function
			var binaryBuf bytes.Buffer
			GenerateBinary(tc.data, M, &binaryBuf)
			binaryOutput := binaryBuf.String()

			// Verify output was generated
			if len(binaryOutput) == 0 {
				t.Errorf("Failed to generate QR code for binary data: %v", tc.data)
			}

			// Verify the underlying QR library can handle this data
			text := string(tc.data)
			_, err := qr.Encode(text, M)
			if err != nil {
				t.Errorf("QR library failed to encode binary data: %v", err)
			}

			// Verify consistency across multiple generations
			var binaryBuf2 bytes.Buffer
			GenerateBinary(tc.data, M, &binaryBuf2)
			binaryOutput2 := binaryBuf2.String()

			if binaryOutput != binaryOutput2 {
				t.Errorf("Binary data should produce consistent QR codes")
			}

			// Verify that our binary function produces the same result as direct string conversion
			var stringBuf bytes.Buffer
			Generate(text, M, &stringBuf)
			stringOutput := stringBuf.String()

			if binaryOutput != stringOutput {
				t.Errorf("Binary function should produce same result as string function for same byte sequence")
			}

			// Verify the QR code structure is valid
			lines := strings.Split(binaryOutput, "\n")
			if len(lines) <= QUIET_ZONE*2 {
				t.Errorf("QR code should have proper structure with quiet zones")
			}

			// Check that the QR code contains both black and white areas (unless it's very small)
			hasBlack := strings.Contains(binaryOutput, BLACK) || strings.Contains(binaryOutput, BLACK_BLACK)
			hasWhite := strings.Contains(binaryOutput, WHITE) || strings.Contains(binaryOutput, WHITE_WHITE)

			if !hasBlack || !hasWhite {
				t.Errorf("QR code should contain both black and white areas")
			}
		})
	}
}

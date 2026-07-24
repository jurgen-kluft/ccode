package embedded

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestXXDCFormatUsesCCoreAlignment(t *testing.T) {
	previousDumpType := dumpType
	dumpType = dumpCformat
	t.Cleanup(func() { dumpType = previousDumpType })

	var output bytes.Buffer
	if err := xxd(bytes.NewReader([]byte{0x01, 0xa5, 0xff}), &output, "program.image"); err != nil {
		t.Fatalf("xxd failed: %v", err)
	}

	generated := output.String()
	for _, expected := range []string{
		"#include \"ccore/c_target.h\"\n",
		"CC_ALIGN(4) unsigned char program_image[] = {",
		"0x01, 0xA5, 0xFF",
		"unsigned int program_image_len = 3;",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated output does not contain %q:\n%s", expected, generated)
		}
	}
}

func TestBinaryEncode(t *testing.T) {
	dst := make([]byte, 8)
	binaryEncode(dst, []byte{0xA5})
	if got, want := string(dst), "10100101"; got != want {
		t.Fatalf("binaryEncode() = %q, want %q", got, want)
	}
}

func TestBinaryDecode(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		want    byte
		wantPos int
	}{
		{name: "valid", src: "10100101", want: 0xA5, wantPos: -1},
		{name: "leading space", src: " 0000000", wantPos: 1},
		{name: "separator", src: "1010 101", wantPos: 4},
		{name: "invalid", src: "010x0000", wantPos: 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dst := []byte{0}
			if got := binaryDecode(dst, []byte(test.src)); got != test.wantPos {
				t.Fatalf("binaryDecode() position = %d, want %d", got, test.wantPos)
			}
			if test.wantPos == -1 && dst[0] != test.want {
				t.Fatalf("binaryDecode() byte = %#x, want %#x", dst[0], test.want)
			}
		})
	}
}

func TestCfmtEncode(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		wantText string
	}{
		{name: "lowercase", table: ldigits, wantText: "0xaf"},
		{name: "uppercase", table: udigits, wantText: "0xAF"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dst := make([]byte, 4)
			cfmtEncode(dst, []byte{0xAF}, test.table)
			if got := string(dst); got != test.wantText {
				t.Fatalf("cfmtEncode() = %q, want %q", got, test.wantText)
			}
		})
	}
}

func TestHexEncode(t *testing.T) {
	dst := make([]byte, 2)
	hexEncode(dst, []byte{0xAF}, ldigits)
	if got, want := string(dst), "af"; got != want {
		t.Fatalf("hexEncode() = %q, want %q", got, want)
	}

	hexEncode(dst, []byte{0xAF}, udigits)
	if got, want := string(dst), "AF"; got != want {
		t.Fatalf("hexEncode() = %q, want %q", got, want)
	}
}

func TestHexDecode(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		want    byte
		wantPos int
	}{
		{name: "plain", src: "aF ", want: 0xAF, wantPos: 0},
		{name: "prefixed lowercase", src: "0xaf", want: 0xAF, wantPos: 0},
		{name: "prefixed uppercase", src: "0XAF", want: 0xAF, wantPos: 0},
		{name: "space", src: " af", wantPos: -1},
		{name: "consecutive spaces", src: "  a", wantPos: -2},
		{name: "invalid high nibble", src: "zf ", wantPos: -1},
		{name: "invalid low nibble", src: "fz ", wantPos: -1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dst := []byte{0}
			if got := hexDecode(dst, []byte(test.src)); got != test.wantPos {
				t.Fatalf("hexDecode() = %d, want %d", got, test.wantPos)
			}
			if test.wantPos == 0 && dst[0] != test.want {
				t.Fatalf("hexDecode() byte = %#x, want %#x", dst[0], test.want)
			}
		})
	}
}

func TestFromHexChar(t *testing.T) {
	tests := []struct {
		input byte
		want  byte
		ok    bool
	}{
		{input: '0', want: 0, ok: true},
		{input: '9', want: 9, ok: true},
		{input: 'a', want: 10, ok: true},
		{input: 'f', want: 15, ok: true},
		{input: 'A', want: 10, ok: true},
		{input: 'F', want: 15, ok: true},
		{input: 'g', ok: false},
	}

	for _, test := range tests {
		got, ok := fromHexChar(test.input)
		if got != test.want || ok != test.ok {
			t.Errorf("fromHexChar(%q) = (%d, %v), want (%d, %v)", test.input, got, ok, test.want, test.ok)
		}
	}
}

func TestEmpty(t *testing.T) {
	zeroes := []byte{0, 0, 0}
	if !empty(&zeroes) {
		t.Fatal("empty() = false for zero-filled slice")
	}

	nonzero := []byte{0, 1, 0}
	if empty(&nonzero) {
		t.Fatal("empty() = true for non-zero slice")
	}
}

func TestParseSpecifier(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{input: "", want: 0},
		{input: "k", want: 1024},
		{input: "M", want: 1048576},
		{input: "g", want: 1073741824},
		{input: "kb", want: 0.0078125},
		{input: "Mb", want: 7.62939453125e-06},
		{input: "gb", want: 7.45058059692383e-09},
		{input: "kB", want: 1024},
		{input: "MB", want: 1048576},
		{input: "GB", want: 1073741824},
		{input: "bytes", want: 1},
	}

	for _, test := range tests {
		if got := parseSpecifier(test.input); got != test.want {
			t.Errorf("parseSpecifier(%q) = %g, want %g", test.input, got, test.want)
		}
	}
}

func TestParseSeek(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{input: "7", want: 7},
		{input: "20", want: 20},
		{input: "100", want: 100},
		{input: "2k", want: 2048},
		{input: "1.5kB", want: 1536},
		{input: "16kb", want: 0},
	}

	for _, test := range tests {
		if got := parseSeek(test.input); got != test.want {
			t.Errorf("parseSeek(%q) = %d, want %d", test.input, got, test.want)
		}
	}
}

func TestIsSpace(t *testing.T) {
	for _, input := range []byte{' ', '\t', '\f'} {
		if !isSpace(input) {
			t.Errorf("isSpace(%q) = false", input)
		}
	}
	for _, input := range []byte{'\n', 'a'} {
		if isSpace(input) {
			t.Errorf("isSpace(%q) = true", input)
		}
	}
}

func TestIsPrefix(t *testing.T) {
	for _, input := range []string{"0x", "0X"} {
		if !isPrefix([]byte(input)) {
			t.Errorf("isPrefix(%q) = false", input)
		}
	}
	if isPrefix([]byte("1x")) {
		t.Fatal("isPrefix(\"1x\") = true")
	}
}

func TestXXD(t *testing.T) {
	tests := []struct {
		name     string
		mode     int
		input    []byte
		filename string
		want     string
	}{
		{
			name:     "hex",
			mode:     dumpHex,
			input:    []byte("AB"),
			filename: "data.bin",
			want:     "0000000: 4142" + strings.Repeat(" ", 32) + "AB\n",
		},
		{
			name:     "binary",
			mode:     dumpBinary,
			input:    []byte{'A'},
			filename: "data.bin",
			want:     "0000000: 01000001" + strings.Repeat(" ", 101) + "A\n",
		},
		{
			name:     "C format",
			mode:     dumpCformat,
			input:    []byte{0x00, 0xAF},
			filename: "data.bin",
			want:     "#include \"ccore/c_target.h\"\n\nCC_ALIGN(4) unsigned char data_bin[] = {\n  0x00, 0xAF\n};\nunsigned int data_bin_len = 2;\n",
		},
		{
			name:     "PostScript",
			mode:     dumpPostscript,
			input:    []byte{0x00, 0xAF},
			filename: "data.bin",
			want:     "00AF\n",
		},
	}

	oldDumpType := dumpType
	t.Cleanup(func() { dumpType = oldDumpType })

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dumpType = test.mode
			var output bytes.Buffer
			if err := xxd(bytes.NewReader(test.input), &output, test.filename); err != nil {
				t.Fatalf("xxd() error = %v", err)
			}
			if got := output.String(); got != test.want {
				t.Fatalf("xxd() output = %q, want %q", got, test.want)
			}
		})
	}
}

func TestXXDAutoskip(t *testing.T) {
	oldDumpType := dumpType
	t.Cleanup(func() { dumpType = oldDumpType })
	dumpType = dumpHex

	input := make([]byte, 36)
	input[35] = 'A'
	var output bytes.Buffer
	if err := xxd(bytes.NewReader(input), &output, "data.bin"); err != nil {
		t.Fatalf("xxd() error = %v", err)
	}
	if got := output.String(); !strings.Contains(got, "\n*\n0000020:") {
		t.Fatalf("xxd() did not autoskip repeated zero lines: %q", got)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestXXDReadError(t *testing.T) {
	oldDumpType := dumpType
	t.Cleanup(func() { dumpType = oldDumpType })
	dumpType = dumpHex

	err := xxd(failingReader{}, io.Discard, "data.bin")
	if err == nil || err.Error() != "read failed" {
		t.Fatalf("xxd() error = %v, want read failed", err)
	}
}

func TestFileNameWithoutExtension(t *testing.T) {
	tests := map[string]string{
		"file.txt":     "file",
		"archive.tar":  "archive",
		"no-extension": "no-extension",
	}
	for input, want := range tests {
		if got := fileNameWithoutExtension(input); got != want {
			t.Errorf("fileNameWithoutExtension(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestWriteEmbedded(t *testing.T) {
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	oldDumpType := dumpType
	t.Cleanup(func() {
		dumpType = oldDumpType
		if err := os.Chdir(oldWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	root := t.TempDir()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}

	WriteEmbedded()
	if _, err := os.Stat(filepath.Join(root, "asset.cpp")); !os.IsNotExist(err) {
		t.Fatalf("WriteEmbedded() generated output without embedded directory")
	}

	embeddedDirectory := filepath.Join(root, "embedded", "assets")
	if err := os.MkdirAll(embeddedDirectory, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "assets"), 0o755); err != nil {
		t.Fatalf("os.Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(embeddedDirectory, "logo.bin"), []byte{0x01, 0xAB}, 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(embeddedDirectory, ".DS_Store"), []byte("Finder metadata"), 0o644); err != nil {
		t.Fatalf("os.WriteFile(.DS_Store) error = %v", err)
	}

	WriteEmbedded()
	if _, err := os.Stat(filepath.Join(root, "assets", ".cpp")); !os.IsNotExist(err) {
		t.Fatalf("WriteEmbedded() generated output for .DS_Store")
	}
	outputPath := filepath.Join(root, "assets", "logo.cpp")
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", outputPath, err)
	}
	want := "#include \"ccore/c_target.h\"\n\nCC_ALIGN(4) unsigned char logo[] = {\n  0x01, 0xAB\n};\nunsigned int logo_len = 2;\n"
	if got := string(output); got != want {
		t.Fatalf("generated output = %q, want %q", got, want)
	}
}

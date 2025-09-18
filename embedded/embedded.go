package embedded

import (
	"bufio"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// constants used
const ebcdicOffset = 0x40
const upper = true
const columns = -1
const group = -1
const length = -1
const autoskip = true

// dumpType enum
const (
	dumpHex = iota
	dumpBinary
	dumpCformat
	dumpPostscript
)

// hex lookup table for hex encoding
const (
	ldigits = "0123456789abcdef"
	udigits = "0123456789ABCDEF"
)

// variables used
var (
	dumpType int

	space        = []byte(" ")
	doubleSpace  = []byte("  ")
	dot          = []byte(".")
	newLine      = []byte("\n")
	zeroHeader   = []byte("0000000: ")
	unsignedChar = []byte("unsigned char ")
	unsignedInt  = []byte("};\nunsigned int ")
	lenEquals    = []byte("_len = ")
	brackets     = []byte("[] = {")
	asterisk     = []byte("*")
	commaSpace   = []byte(", ")
	comma        = []byte(",")
	semiColonNl  = []byte(";\n")
	bar          = []byte("|")
)

// convert a byte into its binary representation
func binaryEncode(dst, src []byte) {
	d := uint(0)
	_, _ = src[0], dst[7]
	for i := 7; i >= 0; i-- {
		if src[0]&(1<<d) == 0 {
			dst[i] = '0'
		} else {
			dst[i] = '1'
		}
		d++
	}
}

// returns -1 on success
// returns k > -1 if space found where k is index of space byte
func binaryDecode(dst, src []byte) int {
	var v, d byte

	for i := 0; i < len(src); i++ {
		v, d = src[i], d<<1
		if isSpace(v) { // found a space, so between groups
			if i == 0 {
				return 1
			}
			return i
		}
		if v == '1' {
			d ^= 1
		} else if v != '0' {
			return i // will catch issues like "000000: "
		}
	}

	dst[0] = d
	return -1
}

func cfmtEncode(dst, src []byte, hextable string) {
	b := src[0]
	dst[3] = hextable[b&0x0f]
	dst[2] = hextable[b>>4]
	dst[1] = 'x'
	dst[0] = '0'
}

// copied from encoding/hex package in order to add support for uppercase hex
func hexEncode(dst, src []byte, hextable string) {
	b := src[0]
	dst[1] = hextable[b&0x0f]
	dst[0] = hextable[b>>4]
}

// copied from encoding/hex package
// returns -1 on bad byte or space (\t \s \n)
// returns -2 on two consecutive spaces
// returns 0 on success
func hexDecode(dst, src []byte) int {
	_, _ = src[2], dst[0]

	if isSpace(src[0]) {
		if isSpace(src[1]) {
			return -2
		}
		return -1
	}

	if isPrefix(src[0:2]) {
		src = src[2:]
	}

	for i := 0; i < len(src)/2; i++ {
		a, ok := fromHexChar(src[i*2])
		if !ok {
			return -1
		}
		b, ok := fromHexChar(src[i*2+1])
		if !ok {
			return -1
		}

		dst[0] = (a << 4) | b
	}
	return 0
}

// copied from encoding/hex package
func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}

	return 0, false
}

// check if entire line is full of empty []byte{0} bytes (nul in C)
func empty(b *[]byte) bool {
	for i := 0; i < len(*b); i++ {
		if (*b)[i] != 0 {
			return false
		}
	}
	return true
}

// quick binary tree check
// probably horribly written idk it's late at night
func parseSpecifier(b string) float64 {
	lb := len(b)
	if lb == 0 {
		return 0
	}

	var b0, b1 byte
	if lb < 2 {
		b0 = b[0]
		b1 = '0'
	} else {
		b1 = b[1]
		b0 = b[0]
	}

	if b1 != '0' {
		if b1 == 'b' { // bits, so convert bytes to bits for os.Seek()
			if b0 == 'k' || b0 == 'K' {
				return 0.0078125
			}

			if b0 == 'm' || b0 == 'M' {
				return 7.62939453125e-06
			}

			if b0 == 'g' || b0 == 'G' {
				return 7.45058059692383e-09
			}
		}

		if b1 == 'B' { // kilo/mega/giga- bytes are assumed
			if b0 == 'k' || b0 == 'K' {
				return 1024
			}

			if b0 == 'm' || b0 == 'M' {
				return 1048576
			}

			if b0 == 'g' || b0 == 'G' {
				return 1073741824
			}
		}
	} else { // kilo/mega/giga- bytes are assumed for single b, k, m, g
		if b0 == 'k' || b0 == 'K' {
			return 1024
		}

		if b0 == 'm' || b0 == 'M' {
			return 1048576
		}

		if b0 == 'g' || b0 == 'G' {
			return 1073741824
		}
	}

	return 1 // assumes bytes as fallback
}

// parses *seek input
func parseSeek(s string) int64 {
	var (
		sl    = len(s)
		split int
	)

	switch {
	case sl >= 2:
		if sl == 2 {
			split = 1
		} else {
			split = 2
		}
	case sl != 0:
		split = 0
	default:
		log.Fatalln("seek string somehow has len of 0")
	}

	mod := parseSpecifier(s[sl-split:])
	ret, err := strconv.ParseFloat(s[:sl-split], 64) // 64 bit float
	if err != nil {
		log.Fatalln(err)
	}

	return int64(ret * mod)
}

// is byte a space? (\t, \n, \s)
func isSpace(b byte) bool {
	switch b {
	case 32, 12, 9:
		return true
	default:
		return false
	}
}

// are the two bytes hex prefixes? (0x or 0X)
func isPrefix(b []byte) bool {
	return b[0] == '0' && (b[1] == 'x' || b[1] == 'X')
}

func xxd(r io.Reader, w io.Writer, fname string) error {
	var (
		lineOffset int64
		hexOffset  = make([]byte, 6)
		groupSize  int
		cols       int
		octs       int
		caps       = ldigits
		doCHeader  = true
		doCEnd     bool
		// enough room for "unsigned char NAME_FORMAT[] = {"
		varDeclChar = make([]byte, 14+len(fname)+6)
		// enough room for "unsigned int NAME_FORMAT = "
		varDeclInt = make([]byte, 16+len(fname)+7)
		nulLine    int64
		totalOcts  int64
	)

	// Generate the first and last line in the -i output:
	// e.g. unsigned char foo_txt[] = { and unsigned int foo_txt_len =
	if dumpType == dumpCformat {
		// copy over "unnsigned char " and "unsigned int"
		_ = copy(varDeclChar[0:14], unsignedChar[:])
		_ = copy(varDeclInt[0:16], unsignedInt[:])

		for i := 0; i < len(fname); i++ {
			if fname[i] != '.' {
				varDeclChar[14+i] = fname[i]
				varDeclInt[16+i] = fname[i]
			} else {
				varDeclChar[14+i] = '_'
				varDeclInt[16+i] = '_'
			}
		}
		// copy over "[] = {" and "_len = "
		_ = copy(varDeclChar[14+len(fname):], brackets[:])
		_ = copy(varDeclInt[16+len(fname):], lenEquals[:])
	}

	// Switch between upper- and lower-case hex chars
	if upper {
		caps = udigits
	}

	// xxd -bpi FILE outputs in binary format
	// xxd -b -p -i FILE outputs in C format
	// simply catch the last option since that's what I assume the author
	// wanted...
	if columns == -1 {
		cols = 12
	} else {
		cols = columns
	}

	// See above comment
	switch dumpType {
	case dumpBinary:
		octs = 8
		groupSize = 1
	case dumpPostscript:
		octs = 0
	case dumpCformat:
		octs = 4
	default:
		octs = 2
		groupSize = 2
	}

	if group != -1 {
		groupSize = group
	}

	// If -l is smaller than the number of cols just truncate the cols
	if length != -1 {
		if length < int64(cols) {
			cols = int(length)
		}
	}

	if octs < 1 {
		octs = cols
	}

	// These are bumped down from the beginning of the function in order to
	// allow for their sizes to be allocated based on the user's speficiations
	var (
		line = make([]byte, cols)
		char = make([]byte, octs)
	)

	c := int64(0) // number of characters
	nl := int64(0)
	r = bufio.NewReader(r)

	var (
		n   int
		err error
	)

	for {
		n, err = io.ReadFull(r, line)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}
		// Speed it up a bit ;)
		if dumpType == dumpPostscript && n != 0 {
			// Post script values
			// Basically just raw hex output
			for i := 0; i < n; i++ {
				hexEncode(char, line[i:i+1], caps)
				w.Write(char)
				c++
			}
			continue
		}

		if n == 0 {
			if dumpType == dumpPostscript {
				w.Write(newLine)
			}

			if dumpType == dumpCformat {
				doCEnd = true
			} else {
				return nil // Hidden return!
			}
		}

		if length != -1 {
			if totalOcts == length {
				break
			}
			totalOcts += length
		}

		if autoskip && empty(&line) {
			if nulLine == 1 {
				w.Write(asterisk)
				w.Write(newLine)
			}

			nulLine++

			if nulLine > 1 {
				lineOffset++ // continue to increment our offset
				continue
			}
		}

		if dumpType <= dumpBinary { // either hex or binary
			// Line offset
			hexOffset = strconv.AppendInt(hexOffset[0:0], lineOffset, 16)
			w.Write(zeroHeader[0:(6 - len(hexOffset))])
			w.Write(hexOffset)
			w.Write(zeroHeader[6:])
			lineOffset++
		} else if doCHeader {
			w.Write(varDeclChar)
			w.Write(newLine)
			doCHeader = false
		}

		if dumpType == dumpBinary {
			// Binary values
			for i, k := 0, octs; i < n; i, k = i+1, k+octs {
				binaryEncode(char, line[i:i+1])
				w.Write(char)
				c++

				if k == octs*groupSize {
					k = 0
					w.Write(space)
				}
			}
		} else if dumpType == dumpCformat {
			// C values
			if !doCEnd {
				w.Write(doubleSpace)
			}
			for i := 0; i < n; i++ {
				cfmtEncode(char, line[i:i+1], caps)
				w.Write(char)
				c++

				// don't add spaces to EOL
				if i != n-1 {
					w.Write(commaSpace)
				} else if n == cols {
					w.Write(comma)
				}
			}
		} else {
			// Hex values -- default xxd FILE output
			for i, k := 0, octs; i < n; i, k = i+1, k+octs {
				hexEncode(char, line[i:i+1], caps)
				w.Write(char)
				c++

				if k == octs*groupSize {
					k = 0 // reset counter
					w.Write(space)
				}
			}
		}

		if doCEnd {
			w.Write(varDeclInt)
			w.Write([]byte(strconv.FormatInt(c, 10)))
			w.Write(semiColonNl)
			return nil
		}

		if n < len(line) && dumpType <= dumpBinary {
			for i := n * octs; i < len(line)*octs; i++ {
				w.Write(space)

				if i%octs == 1 {
					w.Write(space)
				}
			}
		}

		if dumpType != dumpCformat {
			w.Write(space)
		}

		if dumpType <= dumpBinary {
			// Character values
			b := line[:n]

			// ASCII
			{
				var v byte
				for i := 0; i < len(b); i++ {
					v = b[i]
					if v > 0x1f && v < 0x7f {
						w.Write(line[i : i+1])
					} else {
						w.Write(dot)
					}
				}
			}
		}
		w.Write(newLine)
		nl++
	}
	return nil
}

func fileNameWithoutExtension(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func WriteEmbedded() {

	// Collect all files in 'embedded' folder.
	// Anything under the 'embedded' folder will be considered to be embedded.
	// The directory hierarchy will be put over the root, e.g. :
	// embedded/source/main/cpp/data.txt -> source/main/cpp/data.cpp

	root_dir, err := os.Getwd()
	if err != nil {
		corepkg.LogError(err)
	}

	embedded_dir := filepath.Join(root_dir, "embedded")

	// Does the embedded directory exist, if not then we're done
	_, err = os.Stat(embedded_dir)
	if err != nil {
		return
	}

	err = filepath.WalkDir(embedded_dir,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				subdir, filename := filepath.Split(path)
				subdir = subdir[len(embedded_dir)+1:]

				arrayName := fileNameWithoutExtension(filename)
				inFilename := path
				outFilename := filepath.Join(root_dir, subdir, arrayName+".cpp")
				corepkg.LogPrintf("input: %s", inFilename)
				corepkg.LogPrintf("output: %s", outFilename)

				var inFile *os.File
				inFile, err = os.Open(inFilename)
				if err != nil {
					log.Fatalln(err)
				}
				defer inFile.Close()

				var outFile *os.File
				outFile, err = os.Create(outFilename)
				if err != nil {
					log.Fatalln(err)
				}
				defer outFile.Close()

				out := bufio.NewWriter(outFile)
				defer out.Flush()

				dumpType = dumpCformat
				if err = xxd(inFile, out, arrayName); err != nil {
					log.Fatalln(err)
				}
			}

			return nil
		})

	if err != nil {
		corepkg.LogError(err)
	}

}

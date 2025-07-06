package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/katzenpost/qrterminal/v3"
	"github.com/mattn/go-colorable"
	"rsc.io/qr"
)

var verboseFlag bool
var levelFlag string
var quietZoneFlag int
var sixelDisableFlag bool
var binaryFlag bool

func getLevel(s string) qr.Level {
	switch l := strings.ToLower(s); l {
	case "l":
		return qr.L
	case "m":
		return qr.M
	case "h":
		return qr.H
	default:
		return -1
	}
}

func main() {
	flag.BoolVar(&verboseFlag, "v", false, "Output debugging information")
	flag.StringVar(&levelFlag, "l", "L", "Error correction level")
	flag.IntVar(&quietZoneFlag, "q", 2, "Size of quietzone border")
	flag.BoolVar(&sixelDisableFlag, "s", false, "disable sixel format for output")
	flag.BoolVar(&binaryFlag, "b", false, "treat input as binary data (preserves exact byte values)")

	flag.Parse()
	level := getLevel(levelFlag)

	if level < 0 {
		fmt.Fprintf(os.Stderr, "Invalid error correction level: %s\n", levelFlag)
		fmt.Fprintf(os.Stderr, "Valid options are [L, M, H]\n")
		os.Exit(1)
	}

	var content string
	var binaryData []byte
	var err error

	args := flag.Args()
	if len(args) < 1 {
		// Get input from stdin until EOF
		binaryData, err = io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		if !binaryFlag {
			content = string(binaryData)
		}
	} else {
		content = strings.Join(args, " ")
		if binaryFlag {
			binaryData = []byte(content)
		}
	}

	cfg := qrterminal.Config{
		Level:     level,
		Writer:    os.Stdout,
		QuietZone: quietZoneFlag,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
	}
	if !sixelDisableFlag {
		cfg.WithSixel = qrterminal.IsSixelSupported(os.Stdout)
	}
	if verboseFlag {
		fmt.Fprintf(os.Stdout, "Level: %s \n", levelFlag)
		fmt.Fprintf(os.Stdout, "Quietzone Border Size: %d \n", quietZoneFlag)
		fmt.Fprintf(os.Stdout, "Binary mode: %t \n", binaryFlag)
		if binaryFlag {
			fmt.Fprintf(os.Stdout, "Encoded data: %d bytes of binary data \n", len(binaryData))
		} else {
			fmt.Fprintf(os.Stdout, "Encoded data: %s \n", strings.Join(flag.Args(), "\n"))
		}
		fmt.Println("")
	}

	if runtime.GOOS == "windows" {
		cfg.Writer = colorable.NewColorableStdout()
		cfg.BlackChar = qrterminal.BLACK
		cfg.WhiteChar = qrterminal.WHITE
	}

	fmt.Fprint(os.Stdout, "\n")

	if binaryFlag {
		qrterminal.GenerateBinaryWithConfig(binaryData, cfg)
	} else {
		qrterminal.GenerateWithConfig(content, cfg)
	}
}

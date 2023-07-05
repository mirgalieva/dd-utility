package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Options struct {
	From      string
	To        string
	Offset    int64
	Limit     int64
	BlockSize int
	Conv      string
}
type OptionsReadWriter struct {
	reader io.Reader
	writer io.Writer
}

func ParseFlags() (*Options, error) {
	var opts Options

	flag.StringVar(&opts.From, "from", "", "file to read. by default - stdin")
	flag.StringVar(&opts.To, "to", "", "file to write. by default - stdout")
	flag.Int64Var(&opts.Offset, "offset", 0, "number of bytes skip")
	flag.Int64Var(&opts.Limit, "limit", -1, "maximum number of bytes")
	flag.IntVar(&opts.BlockSize, "block-size", 0, "size of single block")
	flag.StringVar(&opts.Conv, "conv", "", "convert to upper, lower or trim")

	flag.Parse()

	return &opts, nil
}

func (e OptionsReadWriter) ReadFile() (error, []byte) {
	var buf []byte
	line := make([]byte, 64)
	for {
		n, err := e.reader.Read(line)
		if err == io.EOF {
			return nil, buf
		}
		if err != nil {
			return fmt.Errorf("can not read"), buf
		}
		buf = append(buf, line[:n]...)
	}
}

func (e OptionsReadWriter) PrintFile(buf []byte) error {
	_, err := e.writer.Write(buf)
	if err != nil {
		return fmt.Errorf("can not write")
	}
	return nil
}

func Offset(skip int64, buf []byte) (error, []byte) {
	if skip >= int64(len(buf)) {
		return fmt.Errorf("can not write"), buf
	}
	return nil, buf[skip:]
}
func Convert(conversions string, buf []byte) (error, []byte) {
	if conversions == "" {
		return nil, buf
	}
	conv := strings.Split(conversions, ",")
	hasRegisterConv := false
	for i := 0; i < len(conv); i++ {
		switch arg := conv[i]; {
		case arg == "upper_case":
			if !hasRegisterConv {
				buf = bytes.ToUpper(buf)
				hasRegisterConv = true
			} else {
				return fmt.Errorf("conv arguments are not correct"), buf
			}
		case arg == "lower_case":
			if !hasRegisterConv {
				buf = bytes.ToLower(buf)
				hasRegisterConv = true
			} else {
				return fmt.Errorf("conv arguments are not correct"), buf
			}
		case arg == "trim_spaces":
			buf = bytes.TrimSpace(buf)
		default:
			return fmt.Errorf("conv arguments are not correct"), buf
		}
	}
	return nil, buf
}

func main() {
	opts, err := ParseFlags()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "can not parse flags:", err)
		os.Exit(1)
	}
	var f *os.File
	var optionRW OptionsReadWriter
	fileFrom := opts.From
	fileTo := opts.To
	limit := opts.Limit
	offset := opts.Offset

	if fileFrom == "" {
		f = os.Stdin
	} else {
		f, err = os.Open(fileFrom)
		if err != nil {
			log.Fatalf("unable to read file: %v", err)
		}
		defer f.Close()

	}
	if limit != -1 {
		optionRW.reader = io.LimitReader(f, limit+offset)
	} else {
		optionRW.reader = f
	}
	if fileTo == "" {
		optionRW.writer = os.Stdout
	} else {
		_, err := os.Open(fileTo)
		if err == nil {
			_, _ = fmt.Fprintln(os.Stderr, "file exists:", err)
			os.Exit(1)
		}
		f, err := os.Create(fileTo)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "can not create file:", err)
			os.Exit(1)
		}
		optionRW.writer = f
		defer f.Close()
	}
	err, buf := OptionsReadWriter.ReadFile(optionRW)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "can not read:", err)
		os.Exit(1)
	}
	err, buf = Offset(opts.Offset, buf)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "offset argument is not correct", err)
		os.Exit(1)
	}
	err, buf = Convert(opts.Conv, buf)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "convert argument is not correct", err)
		os.Exit(1)
	}
	err = OptionsReadWriter.PrintFile(optionRW, buf)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "can not write", err)
		os.Exit(1)
	}
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	Index    int
	From, To time.Duration
	PosInfo  *string
	Text     []*string
}

func ReadBlock(sc *bufio.Scanner) (*Block, error) {
	// Ignore empty lines
	for {
		if len(sc.Bytes()) != 0 {
			break
		}

		if ok := sc.Scan(); !ok && sc.Err() != nil {
			return nil, fmt.Errorf("Error occured while reading file: %v", sc.Err())
		} else if !ok && sc.Err() == nil {
			return nil, nil
		}
	}

	// Read the index
	indexStr := sc.Text()
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return nil, fmt.Errorf("Could not parse block index: %v", err)
	}

	// Read the timestamp
	if ok := sc.Scan(); !ok {
		return nil, fmt.Errorf("Could not read timestamp: %v", err)
	}

	// Parse the timestamps
	from, to, posInfo, err := ParseTimestamp(sc.Text())
	if err != nil {
		return nil, fmt.Errorf("Could not parse timestamp: %v", err)
	}

	// Read the actual subtitle text
	text := []*string{}
	for {
		if ok := sc.Scan(); !ok && sc.Err() != nil {
			return nil, fmt.Errorf("Could not read text: %v", sc.Err())
		} else if !ok && sc.Err() == nil {
			// EOF
			break
		}
		if len(sc.Bytes()) == 0 {
			break
		}

		str := sc.Text()
		text = append(text, &str)
	}

	return &Block{
		Index:   index,
		From:    *from,
		To:      *to,
		PosInfo: posInfo,
		Text:    text,
	}, nil
}

func (b *Block) WriteBlock(wr io.Writer) error {
	_, err := fmt.Fprintf(wr, "%d\r\n", b.Index)
	if err != nil {
		return fmt.Errorf("Could not write index: %v", err)
	}

	_, err = fmt.Fprintf(wr, "%s --> %s", FormatTimestamp(b.From), FormatTimestamp(b.To))
	if err != nil {
		return fmt.Errorf("Could not write timestamp: %v", err)
	}

	if b.PosInfo != nil {
		fmt.Fprintf(wr, " %s\r\n", *b.PosInfo)
	} else {
		fmt.Fprint(wr, "\r\n")
	}

	for _, text := range b.Text {
		_, err = fmt.Fprintf(wr, "%s\r\n", *text)
		if err != nil {
			return fmt.Errorf("Could not write text: %v", err)
		}
	}

	return nil
}

func (b *Block) Stretch(fpsSrc, fpsDst float64) {
	b.From = StretchTimestamp(b.From, fpsSrc, fpsDst)
	b.To = StretchTimestamp(b.To, fpsSrc, fpsDst)
}

const TIMESTAMP_REGEXP_STRING = `(\d{2}):(\d{2}):(\d{2}),(\d{3})\s-->\s(\d{2}):(\d{2}):(\d{2}),(\d{3})\s*(.*)$`

var TIMESTAMP_REGEXP = regexp.MustCompile(TIMESTAMP_REGEXP_STRING)

func ParseTimestamp(input string) (*time.Duration, *time.Duration, *string, error) {
	matches := TIMESTAMP_REGEXP.FindStringSubmatch(input)
	if len(matches) != 10 {
		return nil, nil, nil, fmt.Errorf("Could not parse timestamp header, expected 10 matches, got %d", len(matches))
	}

	from, err := Timestamp(matches[1], matches[2], matches[3], matches[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Could not parse from: %v", err)
	}

	to, err := Timestamp(matches[5], matches[6], matches[7], matches[8])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Could not parse to: %v", err)
	}

	var posInfo *string
	if len(matches[9]) != 0 {
		posInfo = &matches[9]
	}

	return from, to, posInfo, nil
}

func Timestamp(hrs, min, sec, msec string) (*time.Duration, error) {
	hrsInt, err := strconv.Atoi(hrs)
	if err != nil {
		return nil, fmt.Errorf("Could not parse hrs: %v", err)
	}

	minInt, err := strconv.Atoi(min)
	if err != nil {
		return nil, fmt.Errorf("Could not parse mins: %v", err)
	}

	secInt, err := strconv.Atoi(sec)
	if err != nil {
		return nil, fmt.Errorf("Could not parse secs: %v", err)
	}

	msecInt, err := strconv.Atoi(msec)
	if err != nil {
		return nil, fmt.Errorf("Could not parse msec: %v", err)
	}

	dur := (time.Hour * time.Duration(hrsInt)) + (time.Minute * time.Duration(minInt)) + (time.Second * time.Duration(secInt)) + (time.Millisecond * time.Duration(msecInt))
	return &dur, nil
}

func StretchTimestamp(dur time.Duration, fpsSrc, fpsDst float64) time.Duration {
	coeff := fpsSrc / fpsDst
	newNsec := float64(dur) * coeff
	newDur := time.Nanosecond * time.Duration(newNsec)
	return newDur
}

func FormatTimestamp(dur time.Duration) string {
	// Use the given duration as an offset to zero unix time to format it via time.Format
	timestamp := time.Unix(0, dur.Nanoseconds())
	timestamp = timestamp.In(time.UTC)
	return strings.Replace(timestamp.Format("15:04:05.000"), ".", ",", -1)
}

func main() {
	inFile := flag.String("in", "", "Filename to read the srt from")
	outFile := flag.String("out", "", "Filename to write the new srt to")
	fpsIn := flag.Float64("infps", 0, "FPS of the source srt")
	fpsOut := flag.Float64("outfps", 0, "New FPS for the destination srt")

	flag.Parse()

	if inFile == nil || outFile == nil || fpsIn == nil || fpsOut == nil {
		fmt.Println("-in, -fpsin, -out and -outfps are all mandatory")
		return
	}

	if *fpsIn < 1.0 || *fpsIn > 120.0 {
		fmt.Println("infps should be between 1 and 120")
		return
	}

	if *fpsOut < 1.0 || *fpsOut > 120.0 {
		fmt.Println("outfps should be between 1 and 120")
		return
	}

	in, err := os.Open(*inFile)
	if err != nil {
		fmt.Println("Could not open in file :%v", err)
		return
	}
	defer in.Close()
	bufIn := bufio.NewReader(in)

	out, err := os.Create(*outFile)
	if err != nil {
		fmt.Printf("Could not create file :%v\n", err)
		return
	}
	defer out.Close()

	bufOut := bufio.NewWriter(out)
	defer bufOut.Flush()

	// Stupid checks for BOM
	b, err := bufIn.Peek(3)
	if err != nil || len(b) != 3 {
		fmt.Printf("Could not read from file: %v", err)
		return
	}

	// If we found a BOM, write it to the output file
	if b[0] == 239 && b[1] == 187 && b[2] == 191 {
		_, err = bufOut.Write(b)
		if err != nil {
			fmt.Printf("Could not write to out file: %v", err)
			return
		}
		// We're only interested in the text following the BOM
		bufIn.Discard(3)
	}

	scanner := bufio.NewScanner(bufIn)
	for {
		block, err := ReadBlock(scanner)

		if err != nil {
			fmt.Printf("Could not read block: %v\n", err)
			return
		}

		if block == nil {
			break
		}

		block.Stretch(*fpsIn, *fpsOut)
		err = block.WriteBlock(bufOut)
		if err != nil {
			fmt.Printf("Could not write modified block: %v\n", err)
			return
		}
		bufOut.WriteString("\r\n")
	}
}

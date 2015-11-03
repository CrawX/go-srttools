package main

import (
	"bufio"
	"bytes"
	"testing"
	"time"
)

const (
	NEWLINE = "\r\n"
	BLOCK_1 = "1\r\n00:00:23,065 --> 00:00:25,363\r\nJennifer, this is Carrie\r\nMathison and Peter Quinn."

	BLOCK_2 = "2\r\n00:00:25,442 --> 00:00:28,116  X1:63 X2:223 Y1:43 Y2:58\r\nThey were there with Sandy\r\nwhen it happened."

	SRT_INPUT = BLOCK_1 + NEWLINE + NEWLINE + NEWLINE + BLOCK_2 + NEWLINE
)

const (
	TIMESTAMP_INPUT     = `01:12:23,442 --> 01:12:28,116`
	TIMESTAMP_POS_INPUT = `01:12:23,442 --> 01:12:28,116  X1:63 X2:223 Y1:43 Y2:58`

	TIMESTAMP_FROM = 4343442 * time.Millisecond
	TIMESTAMP_TO   = 4348116 * time.Millisecond
	TIMESTAMP_POS  = `X1:63 X2:223 Y1:43 Y2:58`
)

func TestReadBlock(t *testing.T) {
	sc := bufio.NewScanner(bytes.NewReader([]byte(SRT_INPUT)))

	block, err := ReadBlock(sc)
	if err != nil {
		t.Fatal(err)
	}

	if block == nil {
		t.Fatal("Block nil")
	}

	if block.Index != 1 {
		t.Fatal("Index mismatch")
	}

	if block.From != 23065*time.Millisecond || block.To != 25363*time.Millisecond {
		t.Fatal("Timestamp mismatch")
	}

	if block.PosInfo != nil {
		t.Fatal("PosInfo mismatch")
	}

	if len(block.Text) != 2 || *block.Text[0] != "Jennifer, this is Carrie" || *block.Text[1] != "Mathison and Peter Quinn." {
		t.Fatal("Text mismatch")
	}

	block, err = ReadBlock(sc)
	if err != nil {
		t.Fatal(err)
	}

	if block == nil {
		t.Fatal("Block nil")
	}

	if block.Index != 2 {
		t.Fatal("Index mismatch")
	}

	if block.From != 25442*time.Millisecond || block.To != 28116*time.Millisecond {
		t.Fatal("Timestamp mismatch")
	}

	if block.PosInfo == nil || *block.PosInfo != "X1:63 X2:223 Y1:43 Y2:58" {
		t.Fatal("PosInfo mismatch")
	}

	if len(block.Text) != 2 || *block.Text[0] != "They were there with Sandy" || *block.Text[1] != "when it happened." {
		t.Fatal("Text mismatch")
	}

	block, err = ReadBlock(sc)
	if err != nil {
		t.Fatal(err)
	}

	if block != nil {
		t.Fatal("Unexpected block")
	}
}

func TestWriteBlock(t *testing.T) {
	sc := bufio.NewScanner(bytes.NewReader([]byte(SRT_INPUT)))

	block, err := ReadBlock(sc)
	if err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	err = block.WriteBlock(buf)

	if buf.String() != BLOCK_1+NEWLINE {
		t.Fatal("Out mismatch")
	}
}

func TestParseTimestampNoPos(t *testing.T) {
	from, to, posInfo, err := ParseTimestamp(TIMESTAMP_INPUT)
	if err != nil {
		t.Fatal(err)
	}

	if *from != TIMESTAMP_FROM {
		t.Fatal("From mismatch")
	}

	if *to != TIMESTAMP_TO {
		t.Fatal("To mismatch")
	}

	if posInfo != nil {
		t.Fatal("PosInfo should be nil")
	}
}

func TestParseTimestampPos(t *testing.T) {
	from, to, posInfo, err := ParseTimestamp(TIMESTAMP_POS_INPUT)
	if err != nil {
		t.Fatal(err)
	}

	if *from != TIMESTAMP_FROM {
		t.Fatal("From mismatch")
	}

	if *to != TIMESTAMP_TO {
		t.Fatal("To mismatch")
	}

	if posInfo == nil || *posInfo != TIMESTAMP_POS {
		t.Fatal("posInfo mismatch")
	}
}

func TestStretch23967_25(t *testing.T) {
	dur := StretchTimestamp(23065*time.Millisecond, 23.967, 25)
	if dur != time.Duration(22111954200*time.Nanosecond) {
		t.Fatal("Duration mismatch")
	}
}

// Copyright 2014 Wav Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package wav implements wave file reader and writer.
package wav

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	markerFmt  = binary.BigEndian.Uint32([]byte("fmt "))
	markerData = binary.BigEndian.Uint32([]byte("data"))
	markerRiff = binary.BigEndian.Uint32([]byte("RIFF"))
	formatWave = binary.BigEndian.Uint32([]byte("WAVE"))
)

var (
	ErrFileFormat = errors.New("Unexpected data in header")
	ErrClosed     = errors.New("Closed")
)

type Header struct {
	Channels   uint8
	SampleRate uint32
	BPS        uint16
}

// Reader is a wav file reader.
type Reader struct {
	Header
	r          io.Reader
	closed     bool
	size       uint32
	format     uint32
	fmtSize    uint32
	dataSize   uint32
	audioFmt   uint16
	byteRate   uint32
	blockAlign uint16
	buf        [100]byte
}

// NewReader creates wav reader from Reader.
func NewReader(r io.Reader) (*Reader, error) {
	m := new(Reader)
	m.r = r
	m.closed = false
	if err := m.readHeader(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Reader) Read(p []byte) (n int, err error) {
	if m.closed {
		return 0, ErrClosed
	}
	return m.r.Read(p)
}

func (m *Reader) Close() error {
	m.closed = true
	return nil
}

func (m *Reader) readHeader() (err error) {
	_, err = io.ReadFull(m.r, m.buf[0:12])
	if err != nil {
		return err
	}

	if binary.BigEndian.Uint32(m.buf[0:4]) != markerRiff {
		return ErrFileFormat
	}
	m.size = binary.LittleEndian.Uint32(m.buf[4:8])
	m.format = binary.BigEndian.Uint32(m.buf[8:12])

	if m.format != formatWave {
		return ErrFileFormat
	}

	_, err = io.ReadFull(m.r, m.buf[0:8])
	if err != nil {
		return err
	}
	if binary.BigEndian.Uint32(m.buf[0:4]) != markerFmt {
		return ErrFileFormat
	}
	m.fmtSize = binary.LittleEndian.Uint32(m.buf[4:8])
	if m.fmtSize < 16 {
		return ErrFileFormat
	}

	_, err = io.ReadFull(m.r, m.buf[0:m.fmtSize])
	if err != nil {
		return err
	}

	m.audioFmt = binary.LittleEndian.Uint16(m.buf[0:2])
	m.Header.Channels = uint8(binary.LittleEndian.Uint16(m.buf[2:4]))
	m.Header.SampleRate = binary.LittleEndian.Uint32(m.buf[4:8])
	m.byteRate = binary.LittleEndian.Uint32(m.buf[8:12])
	m.blockAlign = binary.LittleEndian.Uint16(m.buf[12:14])
	m.Header.BPS = binary.LittleEndian.Uint16(m.buf[14:16])

	_, err = io.ReadFull(m.r, m.buf[0:8])
	if err != nil {
		return err
	}
	if binary.BigEndian.Uint32(m.buf[0:4]) != markerData {
		return ErrFileFormat
	}
	m.dataSize = binary.LittleEndian.Uint32(m.buf[4:8])
	return nil
}

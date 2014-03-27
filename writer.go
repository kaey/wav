// Copyright 2014 Wav Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wav

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Writer struct {
	Header
	w          io.WriteSeeker
	closed     bool
	skipped    bool
	size       uint32
	fmtSize    uint32
	dataSize   uint32
	audioFmt   uint16
	byteRate   uint32
	blockAlign uint16
	buf        [50]byte
}

func NewWriter(w io.WriteSeeker, h Header) (*Writer, error) {
	m := new(Writer)
	m.w = w
	m.size = 44
	m.fmtSize = 16
	m.audioFmt = 0xfffe
	m.closed = false
	m.skipped = false
	m.Channels = h.Channels
	m.SampleRate = h.SampleRate
	m.BPS = h.BPS

	return m, nil
}

func (m *Writer) Write(p []byte) (n int, err error) {
	if m.closed {
		return 0, ErrClosed
	}
	if !m.skipped {
		m.w.Write(m.buf[0:44])
		m.skipped = true
	}
	n, err = m.w.Write(p)
	m.size += uint32(n)
	m.dataSize += uint32(n)
	return
}

func (m *Writer) Close() (err error) {
	err = m.writeHeader()
	if err != nil {
		return err
	}
	m.closed = true
	return nil
}

func (m *Writer) writeHeader() (err error) {
	binary.BigEndian.PutUint32(m.buf[0:4], markerRiff)
	binary.LittleEndian.PutUint32(m.buf[4:8], m.size)
	binary.BigEndian.PutUint32(m.buf[8:12], formatWave)
	binary.BigEndian.PutUint32(m.buf[12:16], markerFmt)
	binary.LittleEndian.PutUint32(m.buf[16:20], m.fmtSize)
	binary.LittleEndian.PutUint16(m.buf[20:22], m.audioFmt)
	binary.LittleEndian.PutUint16(m.buf[22:24], uint16(m.Header.Channels))
	binary.LittleEndian.PutUint32(m.buf[24:28], m.Header.SampleRate)
	binary.LittleEndian.PutUint32(m.buf[28:32], m.byteRate)
	binary.LittleEndian.PutUint16(m.buf[32:34], m.blockAlign)
	binary.LittleEndian.PutUint16(m.buf[34:36], m.Header.BPS)
	binary.BigEndian.PutUint32(m.buf[36:40], markerData)
	binary.BigEndian.PutUint32(m.buf[40:44], m.dataSize)

	_, err = m.w.Seek(0, 0)
	if err != nil {
		return
	}
	_, err = m.w.Write(m.buf[0:44])
	if err != nil {
		return
	}
	fmt.Printf("%#v\n", m)
	return nil
}

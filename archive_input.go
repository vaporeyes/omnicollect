// ABOUTME: Bounds ZIP directory allocation before archive/zip parses attacker-controlled headers.
// ABOUTME: Rejects ZIP64/multidisk archives unnecessary under the application's size/count limits.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// archive/zip allocates entries before callers can check len(Reader.File).
// Scan the bounded central directory first, including the actual record count;
// merely trusting the end record's declared count would still allow exhaustion.
func preflightZIP(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Size() > maxArchiveCompressedBytes {
		return fmt.Errorf("backup exceeds 256 MB compressed limit")
	}
	tailSize := info.Size()
	if tailSize > 65557 {
		tailSize = 65557
	}
	tail := make([]byte, int(tailSize))
	if _, err := file.ReadAt(tail, info.Size()-tailSize); err != nil && err != io.EOF {
		return err
	}
	end := -1
	for offset := len(tail) - 22; offset >= 0; offset-- {
		if bytes.Equal(tail[offset:offset+4], []byte{'P', 'K', 5, 6}) && offset+22+int(binary.LittleEndian.Uint16(tail[offset+20:])) == len(tail) {
			end = offset
			break
		}
	}
	if end < 0 {
		return fmt.Errorf("invalid ZIP directory")
	}
	footer := tail[end:]
	count := int(binary.LittleEndian.Uint16(footer[10:]))
	if count > maxArchiveEntries {
		return fmt.Errorf("backup exceeds entry limit")
	}
	if binary.LittleEndian.Uint16(footer[4:]) != 0 || binary.LittleEndian.Uint16(footer[6:]) != 0 || int(binary.LittleEndian.Uint16(footer[8:])) != count {
		return fmt.Errorf("multidisk backups are unsupported")
	}
	if end >= 20 && bytes.Equal(tail[end-20:end-16], []byte{'P', 'K', 6, 7}) {
		return fmt.Errorf("ZIP64 backups exceed supported limits")
	}
	size := int64(binary.LittleEndian.Uint32(footer[12:]))
	offset := int64(binary.LittleEndian.Uint32(footer[16:]))
	if size > 8<<20 || offset+size != info.Size()-tailSize+int64(end) {
		return fmt.Errorf("invalid or oversized ZIP directory")
	}
	directory := make([]byte, int(size))
	if _, err := file.ReadAt(directory, offset); err != nil && err != io.EOF {
		return err
	}
	records := 0
	for len(directory) > 0 {
		if len(directory) < 46 || !bytes.Equal(directory[:4], []byte{'P', 'K', 1, 2}) {
			return fmt.Errorf("invalid ZIP directory record")
		}
		length := 46 + int(binary.LittleEndian.Uint16(directory[28:])) + int(binary.LittleEndian.Uint16(directory[30:])) + int(binary.LittleEndian.Uint16(directory[32:]))
		if length > len(directory) {
			return fmt.Errorf("truncated ZIP directory record")
		}
		records++
		if records > maxArchiveEntries {
			return fmt.Errorf("backup exceeds entry limit")
		}
		directory = directory[length:]
	}
	if records != count {
		return fmt.Errorf("ZIP entry count mismatch")
	}
	return nil
}

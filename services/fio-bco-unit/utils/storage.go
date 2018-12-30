// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"syscall"
	"unsafe"
)

var defaultBufferSize int

func init() {
	defaultBufferSize = 2 * os.Getpagesize()
}

func nameFromDirent(dirent *syscall.Dirent) []byte {
	reg := int(uint64(dirent.Reclen) - uint64(unsafe.Offsetof(syscall.Dirent{}.Name)))

	var name []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&name))
	header.Cap = reg
	header.Len = reg
	header.Data = uintptr(unsafe.Pointer(&dirent.Name[0]))

	if index := bytes.IndexByte(name, 0); index >= 0 {
		header.Cap = index
		header.Len = index
	}

	return name
}

// ListDirectory returns sorted slice of item names in given absolute path
// default sorting is ascending
func ListDirectory(absPath string, ascending bool) ([]string, error) {
	v := make([]string, 0)

	dh, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}

	fd := int(dh.Fd())

	scratchBuffer := make([]byte, defaultBufferSize)

	var de *syscall.Dirent

	for {
		n, err := syscall.ReadDirent(fd, scratchBuffer)
		if err != nil {
			_ = dh.Close()
			return nil, err
		}
		if n <= 0 {
			break
		}
		buf := scratchBuffer[:n]
		for len(buf) > 0 {
			de = (*syscall.Dirent)(unsafe.Pointer(&buf[0]))
			buf = buf[de.Reclen:]

			if de.Ino == 0 {
				continue
			}

			nameSlice := nameFromDirent(de)
			namlen := len(nameSlice)
			if (namlen == 0) || (namlen == 1 && nameSlice[0] == '.') || (namlen == 2 && nameSlice[0] == '.' && nameSlice[1] == '.') {
				continue
			}
			v = append(v, string(nameSlice))
		}
	}

	if err = dh.Close(); err != nil {
		return nil, err
	}

	if ascending {
		sort.Slice(v, func(i, j int) bool {
			return v[i] < v[j]
		})
	} else {
		sort.Slice(v, func(i, j int) bool {
			return v[i] > v[j]
		})
	}

	return v, nil
}

// CountFiles returns number of items in directory
func CountFiles(absPath string) int {
	dh, err := os.Open(absPath)
	if err != nil {
		return -1
	}

	nodes := 0
	fd := int(dh.Fd())

	scratchBuffer := make([]byte, defaultBufferSize)

	var de *syscall.Dirent

	for {
		n, err := syscall.ReadDirent(fd, scratchBuffer)
		if err != nil {
			_ = dh.Close()
			return -1
		}
		if n <= 0 {
			break
		}
		buf := scratchBuffer[:n]
		for len(buf) > 0 {
			de = (*syscall.Dirent)(unsafe.Pointer(&buf[0]))
			buf = buf[de.Reclen:]

			if de.Ino == 0 || de.Type != syscall.DT_REG {
				continue
			}

			nodes++
		}
	}

	return nodes
}

// Exists returns true if absolute path exists
func Exists(absPath string) bool {
	_, err := os.Stat(absPath)
	return !os.IsNotExist(err)
}

// TouchFile creates files given absolute path if file does not already exist
func TouchFile(absPath string) bool {
	if err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm); err != nil {
		return false
	}

	f, err := os.OpenFile(absPath, os.O_RDONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		return false
	}
	defer f.Close()

	return true
}

// ReadFileFully reads whole file given absolute path
func ReadFileFully(absPath string) ([]byte, error) {
	f, err := os.OpenFile(absPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, fi.Size())
	_, err = f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return buf, nil
}

// WriteFile writes data given absolute path to a file if that file does not
// already exists
func WriteFile(absPath string, data []byte) bool {
	if err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm); err != nil {
		return false
	}

	f, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return false
	}

	return true
}

// DeleteFile removes file given absolute path if that file does exists
func DeleteFile(absPath string) bool {
	if err := os.Remove(absPath); err != nil {
		return false
	}

	return true
}

// UpdateFile rewrite file with data given absolute path to a file if that file
// exist
func UpdateFile(absPath string, data []byte) bool {
	f, err := os.OpenFile(absPath, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return false
	}

	return true
}

// AppendFile appens data given absolute path to a file, creates it if it does
// not exist
func AppendFile(absPath string, data []byte) bool {
	if err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm); err != nil {
		return false
	}

	f, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return false
	}

	return true
}
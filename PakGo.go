package PakGo

import (
	"bytes"
	"encoding/binary"
	"errors"

	"os"
	"reflect"
	"strings"
)

type pakHeader struct {
	Id     [4]byte
	Offset uint32
	Size   uint32
}

type pakFileEntry struct {
	Name   [56]byte
	Offset uint32
	Size   uint32
}

type pakFileEntryNative struct {
	Name   string
	Offset int64
	Size   int
}

type PakFile struct {
	filetable  []pakFileEntryNative
	filehandle *os.File
}

func (pak PakFile) fileid(filename string, filetable []pakFileEntryNative) (int, error) {
	for id, file := range filetable {
		if filename == file.Name {
			return id, nil
		}
	}
	return -1, errors.New("the system cannot find the file specified")
}

func (pak PakFile) ReadFile(filename string) ([]byte, error) {
	i, err := pak.fileid(filename, pak.filetable)
	if err != nil {
		return nil, err
	}

	data := make([]byte, pak.filetable[i].Size)
	pak.filehandle.Seek(pak.filetable[i].Offset, 0)
	pak.filehandle.Read(data)

	return data, nil
}

func (pak PakFile) PakClose() {
	pak.filehandle.Close()
}

func PakLoad(filename string) (PakFile, error) {
	var filetable []pakFileEntryNative

	f, err := os.Open(filename)
	if err != nil {
		return PakFile{}, err
	}

	header := pakHeader{}
	binary.Read(f, binary.LittleEndian, &header.Id)
	binary.Read(f, binary.LittleEndian, &header.Offset)
	binary.Read(f, binary.LittleEndian, &header.Size)

	validheader := []byte{'P', 'A', 'C', 'K'}
	valid := bytes.Compare(validheader, header.Id[:])
	if valid != 0 {
		return PakFile{}, errors.New("not a valid .PAK file")
	}

	size := reflect.TypeOf(pakFileEntry{}).Size()
	filecount := int(header.Size / uint32(size))

	f.Seek(int64(header.Offset), 0)

	for i := 1; i <= filecount; i++ {
		file := pakFileEntry{}

		binary.Read(f, binary.LittleEndian, &file.Name)
		binary.Read(f, binary.LittleEndian, &file.Offset)
		binary.Read(f, binary.LittleEndian, &file.Size)

		filename := string(file.Name[:])
		filename = string(file.Name[:strings.Index(filename, "\x00")])

		nativefile := pakFileEntryNative{
			Name:   filename,
			Offset: int64(file.Offset),
			Size:   int(file.Size),
		}
		filetable = append(filetable, nativefile)

	}

	return PakFile{
		filetable:  filetable,
		filehandle: f,
	}, nil
}

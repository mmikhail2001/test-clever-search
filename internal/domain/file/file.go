package file

import "os"

type File struct {
	ID       uint64
	Filename string
	Size     uint64
	File     *os.File
}

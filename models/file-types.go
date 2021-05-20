package models

type Suite int

type (
	fileType uint
)

const (
	_ fileType = iota
	uploadFile
	downloadFile
	chunkFile
	templatesFrameFile
	outputFile
)

var (
	fileTypes = map[fileType]string{
		uploadFile:         "upload",
		downloadFile:       "download",
		chunkFile:          "chunk",
		templatesFrameFile: "template",
		outputFile:         "output",
	}
)

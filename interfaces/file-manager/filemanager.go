package filemanager

import (
	"os"
	"sync"

	"github.com/cavaliercoder/grab"
)

type (
	Temperory struct {
		DirName string
	}
	FileManager interface {
		CreateTempDir() error
		DeleteTempDir() error
	}
)

func (t *Temperory) CreateTempDir() error {
	err := os.Mkdir("/tmp/"+t.DirName, 0777)
	return err
}

func (t *Temperory) DeleteTempDir() error {
	err := os.Remove("/tmp/" + t.DirName)
	return err
}

func DownloadFile(wg *sync.WaitGroup, url string, path string) (bool, string) {
	client := grab.NewClient()
	req, err := grab.NewRequest(path, url)
	resp := client.Do(req)
	defer wg.Done()
	return err == nil, resp.HTTPResponse.Status
}

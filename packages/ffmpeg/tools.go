package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	filemanager "github.com/amupxm/go-video-concat/interfaces/file-manager"
	"github.com/amupxm/go-video-concat/interfaces/frame"
	"github.com/amupxm/go-video-concat/interfaces/splash"
	"github.com/amupxm/go-video-concat/packages/cache"

	mediainfo "github.com/amupxm/go-video-concat/packages/mediaInfo"
	"github.com/amupxm/go-video-concat/packages/s3"
	"github.com/gofrs/uuid"
	"github.com/minio/minio-go/v7"
)

func (f *FFmpeg_Generator) GenerateChunks() {
	var wg sync.WaitGroup
	wg.Add(len(f.Recipe.Chunks))
	for chunkIndex := range f.Recipe.Chunks {
		switch f.Recipe.Chunks[chunkIndex].Type {
		case "image":
			go f.img2vid(chunkIndex, &wg)
		case "video":
			go f.trimVid(chunkIndex, &wg)
		}
	}
	wg.Wait()
}

// add color back ground as same size as frame requirements for Splash and place it in center.
func (f *FFmpeg_Generator) FitSplash() {
	var splashMediaInfo mediainfo.MediaInfoStruct
	err := splashMediaInfo.GetVideoInfo(f.Dir + "splash-file")
	if err != nil {
		f.Error.Status = err == nil
		f.Error.Message = "mediaInfo splash error"
		return
	}
	splashCorrectionOperation := []string{
		"-i",
		f.Dir + "splash-file",
		"-f",
		"lavfi",
		"-i",
		fmt.Sprintf("color=red:s=%dx%d", MakeEvenNumber(f.Frame.Width), MakeEvenNumber(f.Frame.Height)),
		"-filter_complex",
		fmt.Sprintf("[1]scale=%d:-1[inner];[inner][0]overlay=%d:0:shortest=1[out]", MakeEvenNumber(f.Frame.Width), MakeEvenNumber((f.Frame.Width-int(splashMediaInfo.Width))/2)),
		"-map",
		"[out]:v",
		"-map",
		"0:a",
		f.Dir + "splash.final.mp4",
	}
	err = Execute(splashCorrectionOperation...)
	if err != nil {
		f.Error.Status = err == nil
		f.Error.Message = "mediaInfo splash error"
	}
}

// get splash file using splash from s3 and write to tmp dir
func (f *FFmpeg_Generator) GetSplashFile(wg *sync.WaitGroup) {
	var objectStorage = &s3.ObjectStorage
	err := objectStorage.Client.FGetObject(
		context.Background(),
		"splash",
		f.Recipe.Splash,
		f.Dir+"splash-file",
		minio.GetObjectOptions{})
	if err != nil {
		f.Error.Status = err == nil
		f.Error.Message = "object splash error"
	}
	wg.Done()
}

// download all files (chunk + external audio id exists) and write to tmp dir.
func (f *FFmpeg_Generator) DownloadFiles() {
	var wg sync.WaitGroup

	// add extra to wait group if has external audio (because it will download audio as external file from url)
	if f.Recipe.ExternalAudio {
		wg.Add(len(f.Recipe.Chunks) + 1)

	} else {
		wg.Add(len(f.Recipe.Chunks) + 1)
	}

	innerResponseChannel := make(chan *FFmpeg_Message)

	for chunkIndex := range f.Recipe.Chunks {
		go func(chunkIndex int) {
			// set a uuid as name for chunk (audio or video file)
			if newChunkName, err := uuid.NewV4(); err == nil {
				f.Recipe.Chunks[chunkIndex].Name = newChunkName.String()
			} else {
				innerResponseChannel <- &FFmpeg_Message{
					Status:  false,
					Message: "internal error",
				}
			}

			status, statusCode := filemanager.DownloadFile(
				&wg,
				f.Recipe.Chunks[chunkIndex].Url,
				f.Dir+f.Recipe.Chunks[chunkIndex].Name)
			innerResponseChannel <- &FFmpeg_Message{
				Status:  status,
				Message: "url responded withCode" + statusCode,
			}
		}(chunkIndex)

	}
	// download audio if exists
	if f.Recipe.ExternalAudio {
		go func() {
			audioName, _ := uuid.NewV4()
			filemanager.DownloadFile(
				&wg,
				f.Recipe.Audio,
				f.Dir+audioName.String())
			f.Recipe.Audio = audioName.String()
		}()

	}
	wg.Wait()

	isAllDownloaded := true
	count := 0
	for responseFromInnerChannel := range innerResponseChannel {
		isAllDownloaded = isAllDownloaded && responseFromInnerChannel.Status
		log.Println("Downloaded to :  " + f.Dir + f.Recipe.Chunks[count].Name)
		count++
		if count == len(f.Recipe.Chunks) {
			close(innerResponseChannel)
		}
	}

	if !isAllDownloaded {
		f.Error.Status = false
		f.Error.Message = "error while downloading files"
	}
}

// get frame object from s3 and write to temp dir
func (f *FFmpeg_Generator) GetFrameFile(wg *sync.WaitGroup) {
	var objectStorage = &s3.ObjectStorage
	err := objectStorage.Client.FGetObject(
		context.Background(),
		"frame",
		f.Frame.FileCode,
		f.Dir+f.Recipe.FrameCode,
		minio.GetObjectOptions{})
	if err != nil {
		f.Error.Status = false
		f.Error.Message = "s3 error"
	}
	wg.Done()

}

// send FFMPEG_MESSAGE to response channel. use to send errors to api
func (f *FFmpeg_Generator) ResponseError() {
	cache.UpdateStatus(f.UUID, f.Error.Message, false)
}

// get Splash details from database and validate it.
// set error on invalid frame.
func (f *FFmpeg_Generator) ValidateSplash() {
	var splashFinder splash.Splash
	tmpSplash, status := splashFinder.FindSplash(f.Recipe.Splash)
	f.Error.Status = !status
	f.Error.Message = "splash file error"
	f.Splash = &tmpSplash
}

// get frame details from database and validate it.
// set error on invalid frame.
func (f *FFmpeg_Generator) ValidateFrame() {
	var frameFinder frame.Frame
	tmpFrame, status := frameFinder.FindFrame(f.Recipe.FrameCode)
	f.Error.Status = !status
	f.Error.Message = "template file error"
	f.Frame = &tmpFrame
}

// modify Temp directory
// with operation (init)
func (f *FFmpeg_Generator) TmpDir(operation string) {
	fm := filemanager.Temperory{DirName: f.UUID}
	if operation == "init" {
		f.Dir = "/tmp/" + f.UUID + "/"
		err := fm.CreateTempDir()
		if err != nil {
			f.Error.Status = false
			f.Error.Message = "file operation error"
		}
	}
}

func (f *FFmpeg_Generator) GetFromS3(code string) (*minio.Object, string, error) {

	var s3 = &s3.ObjectStorage
	var tmp *minio.Object

	reader, err := s3.Client.GetObject(context.Background(), "amupxm", code, minio.GetObjectOptions{})
	if err != nil {
		return tmp, "", err
	}

	if err != nil {
		return tmp, "", err
	}
	return reader, "", nil
}

//write results to minio
func (f *FFmpeg_Generator) WriteTos3() error {
	var s3 = &s3.ObjectStorage
	_, err := s3.Client.FPutObject(
		context.Background(),
		"amupxm",
		f.UUID,
		f.Dir+"x.mp4",
		minio.PutObjectOptions{},
	)
	if err != nil {
		log.Println(111, err)
		return err
	}
	log.Println(111)
	return nil
}

// send callback status
func (f *FFmpeg_Generator) SendResponseCallback() error {
	var callBack = []byte(`{"status":true , "url" : "url" }`)
	req, err := http.NewRequest(
		"POST",
		f.Recipe.CallBackAddress,
		bytes.NewBuffer(callBack),
	)
	if err != nil {
		return errors.New("invalid request generation")
	}
	req.Header.Set("Content-Type", "application/json")
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return errors.New("call back failed ")
	}
	if resp.Status != "200 OK" {
		return errors.New("callback responded with not 200 code")
	}
	defer resp.Body.Close()
	return nil
}

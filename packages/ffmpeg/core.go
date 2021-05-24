package ffmpeg

import (
	"sync"

	"github.com/amupxm/go-video-concat/interfaces/frame"
	"github.com/amupxm/go-video-concat/interfaces/splash"
	"github.com/amupxm/go-video-concat/models"
	"github.com/amupxm/go-video-concat/packages/cache"
)

type (
	FFmpeg_Audio struct {
		Format   string
		Input    string
		Output   string
		Duration int
	}
	FFmpeg_Splash struct {
		Audio  string
		Video  string
		Output string
	}
	FFmpeg_Generator struct {
		Recipe *models.VideoRecipe
		Frame  *frame.Frame
		Splash *splash.Splash
		Error  *FFmpeg_Message
		UUID   string
		Dir    string
	}
	FFmpeg_Message struct {
		Status  bool
		Message string
		UUID    *string
	}
	FFmpegInterface interface {
		Generator(f *FFmpeg_Generator, responseChannel chan *FFmpeg_Message)
		GenerateSplash(splash *splash.Splash) *FFmpeg_Message
		img2vid(index int, wg *sync.WaitGroup)
		trimVid(index int, wg *sync.WaitGroup)
		CreateSplash(splash FFmpeg_Splash)
		GetSplashFile(wg *sync.WaitGroup)
		GetFRameFile(wg *sync.WaitGroup)
		TrimAudio(audio FFmpeg_Audio)
		Execute(args ...string) error
		MakeEvenNumber(i int) int
		ValidateSplash()
		GenerateChunks()
		ResponseError()
		DownloadFiles()
		FitSplash()
		Overlay()
		Concat()
		TmpDir()
	}
)

func Generator(f *FFmpeg_Generator) {
	// init redis status
	cache.NewProccess(f.UUID, len(f.Recipe.Chunks))
	f.Error.Status = false
	f.Error.Message = ""

	// validate frame
	cache.UpdateStatus(f.UUID, "validate frame", true)
	f.ValidateFrame()
	if f.Error.Status {
		f.ResponseError()
		return
	}
	// validate splash
	cache.UpdateStatus(f.UUID, "validate Splash", true)
	f.ValidateSplash()
	if f.Error.Status {
		f.ResponseError()
		return
	}

	// create tmp file
	cache.UpdateStatus(f.UUID, "init directory", true)

	f.TmpDir("init")
	if f.Error.Status {
		f.ResponseError()
		return
	}

	// download files
	cache.UpdateStatus(f.UUID, "download files", true)
	f.DownloadFiles()
	if f.Error.Status {
		f.ResponseError()
		return
	}

	// get splash and frame file
	cache.UpdateStatus(f.UUID, "connecting to s3", true)
	var fileControllerWaitGroup sync.WaitGroup
	fileControllerWaitGroup.Add(2)
	go f.GetSplashFile(&fileControllerWaitGroup)
	go f.GetFrameFile(&fileControllerWaitGroup)
	fileControllerWaitGroup.Wait()
	if f.Error.Status {
		f.ResponseError()
		return
	}
	// make splash in correct size
	cache.UpdateStatus(f.UUID, "generate splash", true)
	f.FitSplash()
	// trim video  chunks in correct duration
	cache.UpdateStatus(f.UUID, "prepare", true)
	f.GenerateChunks()
	// connect chunks (+ sound overlay)
	cache.UpdateStatus(f.UUID, "concatting", true)
	f.Concat()
	//overLay
	cache.UpdateStatus(f.UUID, "overlaying", true)
	f.OverLay()
	cache.UpdateStatus(f.UUID, "done", true)
	f.WriteTos3()
	defer f.SendResponseCallback()
}

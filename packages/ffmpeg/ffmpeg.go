package ffmpeg

import (
	"context"
	"fmt"
	"strings"
	"sync"

	filemanager "github.com/amupxm/go-video-concat/interfaces/file-manager"
	"github.com/amupxm/go-video-concat/interfaces/splash"
	"github.com/amupxm/go-video-concat/packages/cache"
	"github.com/amupxm/go-video-concat/packages/s3"
	"github.com/minio/minio-go/v7"
)

func (f *FFmpeg_Generator) img2vid(index int, wg *sync.WaitGroup) {

	img2vidArgs := []string{
		"-r",
		"30",
		"-framerate",
		fmt.Sprintf("1/%d", f.Recipe.Chunks[index].End-f.Recipe.Chunks[index].Start),
		"-i",
		f.Dir + f.Recipe.Chunks[index].Name,
		"-f", "lavfi", "-i", "anullsrc",
		"-filter_complex",
		fmt.Sprintf("[0:v]split=2[blur][vid];[blur]scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d,boxblur=luma_radius=min(h\\,w)/20:luma_power=1:chroma_radius=min(cw\\,ch)/20:chroma_power=1[bg];[vid]scale=%d:%d:force_original_aspect_ratio=decrease[ov];[bg][ov]overlay=(W-w)/2:(H-h)/2[out]",
			f.Frame.Width, f.Frame.Height,
			f.Frame.Width, f.Frame.Height,
			f.Frame.Width, f.Frame.Height,
		),
		"-vcodec", "libx264",
		"-pix_fmt", "yuv420p",
		"-map",
		"[out]:v", "-map", "1:a", "-shortest",
		fmt.Sprintf("%s/%d.final.mp4", f.Dir, index),
	}
	err := Execute(img2vidArgs...)

	if err != nil {
		f.Error.Status = false
		f.Error.Message = "img 2 vid failure"
	}
	wg.Done()

}

func (f *FFmpeg_Generator) trimVid(index int, wg *sync.WaitGroup) {
	vidTrimArgs := []string{
		"-i",
		f.Dir + f.Recipe.Chunks[index].Name,
		"-ss",
		"0",
		"-t",
		fmt.Sprintf("%d", f.Recipe.Chunks[index].End-f.Recipe.Chunks[index].Start),
		"-async",
		"1",
		"-filter_complex",
		fmt.Sprintf("[0:v]split=2[blur][vid];[blur]scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d,boxblur=luma_radius=min(h\\,w)/20:luma_power=1:chroma_radius=min(cw\\,ch)/20:chroma_power=1[bg];[vid]scale=%d:%d:force_original_aspect_ratio=decrease[ov];[bg][ov]overlay=(W-w)/2:(H-h)/2",
			f.Frame.Width, f.Frame.Height,
			f.Frame.Width, f.Frame.Height,
			f.Frame.Width, f.Frame.Height),
		"-c:a",
		"copy",
		"-avoid_negative_ts", "make_zero", "-fflags", "+genpts",
		fmt.Sprintf("%s/%d.final.mp4", f.Dir, index),
	}
	err := Execute(vidTrimArgs...)
	if err != nil {
		f.Error.Status = false
		f.Error.Message = "trim video failure"
	}
	wg.Done()
}
func (f *FFmpeg_Generator) Concat() bool {
	count := len(f.Recipe.Chunks)
	inputsStr := []string{}

	inputsStr = append(inputsStr, "-i", f.Dir+"splash.final.mp4")
	filterStr := "[0:v:0] [0:a:0] "

	for i := 0; i < count; i++ {
		inputsStr = append(inputsStr, "-i", fmt.Sprintf("%s%d.final.mp4", f.Dir, i))
		filterStr = filterStr + fmt.Sprintf("[%d:v:0] [%d:a:0] ", i+1, i+1)
	}
	concatArgs := fmt.Sprintf("concat=n=%d:v=1:a=1:unsafe=1 [v] [a]", count+1)
	mapArgs := []string{
		"-map", "[v]", "-map", "[a]",
	}
	args := []string{}
	args = append(args, inputsStr...)
	args = append(args, "-filter_complex")
	args = append(args, fmt.Sprintf("%s%s", filterStr, concatArgs))
	args = append(args, mapArgs...)
	args = append(args, f.Dir+"final.output.mp4")

	err := Execute(args...)

	return err == nil

}

func (f *FFmpeg_Generator) OverLay() {
	var args []string
	if f.Recipe.ExternalAudio {
		args = []string{"-loop", "1", "-i", f.Dir + f.Frame.FileCode, "-i", f.Dir + "final.output.mp4",
			"-i", f.Dir + f.Recipe.Audio,
			"-filter_complex",
			fmt.Sprintf("[1]scale=%d:%d[inner];[0][inner]overlay=0:%d:shortest=1[out];[1:a]aformat=sample_fmts=fltp:sample_rates=44100:channel_layouts=stereo,volume=1[a1];[2:a]aformat=sample_fmts=fltp:sample_rates=44100:channel_layouts=stereo,volume=0.%d[a2];[a1][a2]amerge=inputs=2,pan=stereo:c0<c0+c2:c1<c1+c3[audiooutput]",
				f.Frame.Width, f.Frame.Height, f.Frame.StartOffset,
				f.Recipe.SoundLevel),
			"-map", "[out]:v", "-map", "[audiooutput]", "-c:a", "libmp3lame", "-shortest",
			f.Dir + "x.mp4"}
	} else {
		args = []string{"-loop", "1", "-i", f.Dir + f.Frame.FileCode, "-i", f.Dir + "final.output.mp4",
			"-filter_complex",
			fmt.Sprintf("[1]scale=%d:%d[inner];[0][inner]overlay=0:%d:shortest=1[out];",
				f.Frame.Width, f.Frame.Height, f.Frame.StartOffset),
			"-map", "[out]:v", "-shortest",
			f.Dir + "x.mp4"}
	}
	Execute(args...)
}

//=========================

func TrimAudio(audio FFmpeg_Audio) {
	args := []string{
		"-i", audio.Input,
		"-ss",
		"0",
		"-t",
		fmt.Sprint(audio.Duration),
		"-c",
		"copy",
		audio.Output,
	}
	Execute(args...)
}

func CreateSplash(splash FFmpeg_Splash) {
	args := []string{
		"-i",
		splash.Video,
		"-i",
		splash.Audio,
		"-map",
		"0:v",
		"-map",
		"1:a",
		"-c:v",
		"copy",
		"-shortest",
		splash.Output,
	}
	Execute(args...)
}

func GenerateSplash(splash *splash.Splash) *FFmpeg_Message {
	var storage = &s3.ObjectStorage
	filemanager := filemanager.Temperory{DirName: splash.FileCode}
	object, err := storage.Client.StatObject(context.Background(), "splash-base", splash.BaseFile, minio.StatObjectOptions{})
	if err != nil {
		return &FFmpeg_Message{Message: "error on connecting to storage", Status: false}
	}
	//check file is image or video
	if strings.Split(object.ContentType, "/")[0] == "image" {
		return &FFmpeg_Message{Message: "only video as base are allowed", Status: false}
	}
	// create redis files
	cache.NewProccess(splash.FileCode, 2)
	if err != nil {
		return &FFmpeg_Message{Message: "cache error", Status: false}
	}
	// create tmp folder
	err = filemanager.CreateTempDir()
	if err != nil {
		return &FFmpeg_Message{Message: "fileManager error", Status: false}
	}
	// download splash filed to tmp
	err = storage.Client.FGetObject(context.Background(), "splash-base", splash.BaseFile, "/tmp/"+splash.FileCode+"/"+splash.BaseFile, minio.GetObjectOptions{})
	if err != nil {
		return &FFmpeg_Message{Message: "Object error", Status: false}
	}
	err = storage.Client.FGetObject(context.Background(), "splash-audio", splash.BaseAudio, "/tmp/"+splash.FileCode+"/"+splash.BaseAudio, minio.GetObjectOptions{})
	if err != nil {
		return &FFmpeg_Message{Message: "Object error", Status: false}
	}
	// concat and trim

	//1=> trim audio ::
	_, err = storage.Client.StatObject(context.Background(), "splash-audio", splash.BaseAudio, minio.StatObjectOptions{})
	if err != nil {
		return &FFmpeg_Message{Message: "Object error", Status: false}
	}
	var audio = FFmpeg_Audio{Input: "/tmp/" + splash.FileCode + "/" + splash.BaseAudio, Output: "/tmp/" + splash.FileCode + "/" + "output.mp3", Duration: int(splash.MaxLength)}
	TrimAudio(audio)

	var video = FFmpeg_Splash{Audio: "/tmp/" + splash.FileCode + "/" + "output.mp3", Video: "/tmp/" + splash.FileCode + "/" + splash.BaseFile, Output: "/tmp/" + splash.FileCode + "/out.mp4"}
	CreateSplash(video)

	xx, err := storage.Client.FPutObject(context.Background(), "splash", splash.FileCode, video.Output, minio.PutObjectOptions{})
	fmt.Print("\n\n\n\n\n\n\n\n", xx)
	if err != nil {
		return &FFmpeg_Message{Message: "Object error", Status: false}
	}
	// ffmpeg.New().Input()
	return &FFmpeg_Message{Message: "done", Status: true}

}

func MakeEvenNumber(i int) int {
	if i%2 == 0 {
		return i
	}
	return i + 1
}

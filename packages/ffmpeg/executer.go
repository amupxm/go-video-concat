package ffmpeg

import (
	"os"
	"os/exec"
)

type Config struct {
	FfmpegBinPath   string
	FfprobeBinPath  string
	ProgressEnabled bool
	Verbose         bool
}
type Progress struct {
	FramesProcessed string
	CurrentTime     string
	CurrentBitrate  string
	Progress        float64
	Speed           string
}

const (
	FfmpegBinPath   = "/usr/bin/ffmpeg"
	FfprobeBinPath  = "/usr/bin/ffprobe"
	ProgressEnabled = true
	Verbose         = true
)

// Spawn chield proccess
func Execute(args ...string) error {
	cmd := exec.Command(FfmpegBinPath, args...)
	if Verbose {
		cmd.Stderr = os.Stdout
	}
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

// TODO : add transcoder for executable

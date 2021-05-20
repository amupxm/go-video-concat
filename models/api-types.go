package models

import "time"

type (
	VideoRecipe struct {
		VideoName       string         `json:"video_name" binding:"required"`
		FrameCode       string         `json:"frame_code" binding:"required"`
		CallBackAddress string         `json:"call_back_address" binding:"required"`
		SoundLevel      uint           `json:"sound_level" binding:"required"`
		Audio           string         `json:"audio_file" binding:"required" `
		Splash          string         `json:"splash_file" binding:"required" `
		ExternalAudio   bool           `json:"has_audio" binding:"required"`
		Chunks          []RecipeChunks `json:"inputs"`
		Status          bool
		Message         string
		UUID            string
		StartedAt       time.Time
		EndedAt         *time.Time
	}
	RecipeChunks struct {
		FinalFile string
		Type      string `json:"type" binding:"required"`
		Url       string `json:"path"  binding:"required"`
		Start     int    `json:"start"  binding:"required"`
		Name      string
		End       int `json:"end"  binding:"required"`
	}
)

package models

type SplashBaseType int

const (
	Image SplashBaseType = iota
	Video
)

func (splash SplashBaseType) String() string {
	return [...]string{"Image", "Video"}[splash]
}

type (
	Splash struct {
		ID       uint   `gorm:"primaryKey;index"`
		FileCode string `gorm:"uniqueIndex;not null"`
		Status   bool
		// requirements as json input
		BaseFile   string `json:"base_file" binding:"required"`
		BaseAudio  string `json:"base_audio"  binding:"required"`
		SplashBase SplashBaseType
		Name       string `json:"name"  binding:"required"`
		MaxLength  uint   `json:"max_length"  binding:"required"`
		BaseColor  string `json:"base_color"  binding:"required"`
	}
)

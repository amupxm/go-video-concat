package models

type (
	Frame struct {
		ID          uint   `gorm:"primaryKey;index"`
		Name        string `json:"name" binding:"required"`
		FileCode    string `json:"file_code" binding:"required" gorm:"uniqueIndex;not null"`
		Height      int    `json:"height" binding:"required"`
		Width       int    `json:"width" binding:"required"`
		StartOffset int    `json:"start_offset" binding:"required"`
	}
)

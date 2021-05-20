package mediainfo

import mediainfo "github.com/lbryio/go_mediainfo"

type (
	MediaInfoStruct struct {
		Height   int64
		Width    int64
		Duration int64
	}
	MediaInfoInterface interface {
		getVideoInfo() error
	}
)

func (m *MediaInfoStruct) GetVideoInfo(path string) error {
	mi := mediainfo.NewMediaInfo()
	err := mi.OpenFile(path)
	if err != nil {
		return err
	}
	defer mi.Close()

	m.Duration = mi.GetInt(mediainfo.MediaInfo_Stream_General, "Duration")
	m.Width = mi.GetInt(mediainfo.MediaInfo_Stream_Video, "Width")
	m.Height = mi.GetInt(mediainfo.MediaInfo_Stream_Video, "Height")
	return nil
}

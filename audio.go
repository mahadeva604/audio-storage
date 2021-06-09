package storage

import "errors"

const FileExt = ".aac"

type Audio struct {
	Id       int    `json:"id"`
	UserId   int    `json:"user_id"`
	Title    string `json:"title"`
	Duration int    `json:"duration"`
	FilePath string `json:"-"`
}

type DownloadAudio struct {
	Title    string `db:"title"`
	FilePath string `db:"file_path"`
}

type UpdateAudio struct {
	Title    *string `json:"title"`
	Duration *int    `json:"duration"`
}

type AudioListParam struct {
	Limit     *int   `json:"limit" form:"limit" binding:"required"`
	Offset    *int   `json:"offset" form:"offset" binding:"required"`
	OrderType string `json:"order_type" form:"order_type" binding:"required,oneof='owner' 'alphabet'" enums:"owner,alphabet"`
}

type ShareList struct {
	UserId int    `json:"id" db:"shared_to_id"`
	Name   string `json:"name" db:"shared_to_name"`
}

type AudioList struct {
	Id      int          `json:"id" db:"audio_id"`
	Title   string       `json:"name" db:"title"`
	IsOwner bool         `json:"is_owner" db:"is_owner"`
	Owner   int          `json:"owner_id" db:"user_id"`
	Name    string       `json:"owner_name" db:"name"`
	Shares  *[]ShareList `json:"shared_to,omitempty""`
}

type AudioListJson struct {
	TotalCount int         `json:"total_count"`
	Records    []AudioList `json:"records"`
}

type AudioListDb struct {
	Count int `db:"full_count"`
	AudioList
	ShareList
}

func (i UpdateAudio) Validate() error {
	if i.Title == nil && i.Duration == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

type Share struct {
	UserId  int `json:"user_id"`
	AudioId int `json:"audio_id"`
}

func Aac(buf []byte) bool {
	return len(buf) > 1 &&
		((buf[0] == 0xFF && buf[1] == 0xF1) ||
			(buf[0] == 0xFF && buf[1] == 0xF9))
}

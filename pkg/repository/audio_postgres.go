package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	storage "github.com/mahadeva604/audio-storage"
)

type AudioPostgres struct {
	db *sqlx.DB
}

func NewAudioPostgres(db *sqlx.DB) *AudioPostgres {
	return &AudioPostgres{db: db}
}

func (r *AudioPostgres) UploadFile(userId int, path string) (int, error) {
	var audioId int
	query := fmt.Sprintf(`INSERT INTO %s (user_id, title, duration, file_path)
							VALUES ($1, '', 0, $2) RETURNING audio_id`, audiosTable)
	err := r.db.Get(&audioId, query, userId, path)

	return audioId, err
}

func (r *AudioPostgres) DownloadFile(userID, audioId int) (storage.DownloadAudio, error) {
	var audio storage.DownloadAudio
	query := fmt.Sprintf("SELECT title, file_path FROM %s a LEFT JOIN %s r USING (audio_id) WHERE audio_id = $1 and (a.user_id = $2 or r.user_id = $2)", audiosTable, sharesTable)
	err := r.db.Get(&audio, query, audioId, userID)

	if err == sql.ErrNoRows {
		err = storage.FileNotFound
	}

	return audio, err
}

func (r *AudioPostgres) AddDescription(userID, audioId int, input storage.UpdateAudio) error {
	query := fmt.Sprintf("UPDATE %s SET title = $1, duration = $2 WHERE user_id = $3 and audio_id = $4", audiosTable)

	result, err := r.db.Exec(query, input.Title, input.Duration, userID, audioId)

	if err != nil {
		return err
	}

	if rowsAff, err := result.RowsAffected(); rowsAff == 0 && err == nil {
		return storage.NotOwner
	}

	return err
}

func (r *AudioPostgres) GetAudioList(userID int, input storage.AudioListParam) (storage.AudioListJson, error) {

	var orderType string
	if input.OrderType == "owner" {
		orderType = "is_owner DESC, name, title"
	} else if input.OrderType == "alphabet" {
		orderType = "title"
	} else {
		return storage.AudioListJson{}, errors.New("unknown order type")
	}

	query := fmt.Sprintf(`SELECT full_count, audio_id, title, is_owner, o.user_id, o.name,
						COALESCE(r.user_id, 0) AS shared_to_id, COALESCE(u.name, '') AS shared_to_name
						FROM
						(SELECT
    						count(*) OVER() AS full_count, audio_id, title,
    						CASE WHEN user_id = $1 THEN true ELSE false END AS is_owner,
    						user_id, name
						FROM %s
						JOIN users USING (user_id)
						WHERE user_id = $1
						OR audio_id IN (SELECT audio_id FROM shares WHERE user_id = $1)
						ORDER BY %[2]s
						OFFSET $2 LIMIT $3) o
						LEFT JOIN shares r USING (audio_id)
						LEFT JOIN users u ON r.user_id = u.user_id
						ORDER BY %[2]s`, audiosTable, orderType)

	resultOut := make([]storage.AudioList, 0)

	var result storage.AudioListDb
	var lastAudio storage.AudioList
	var totalCount int

	rows, err := r.db.Queryx(query, userID, input.Offset, input.Limit)
	if err != nil {
		return storage.AudioListJson{}, err
	}
	for rows.Next() {

		err := rows.StructScan(&result)
		if err != nil {
			return storage.AudioListJson{}, err
		}
		totalCount = result.Count
		if lastAudio == result.AudioList {
			lastIndex := len(resultOut) - 1
			emptyShare := storage.ShareList{}
			if result.ShareList != emptyShare {
				*resultOut[lastIndex].Shares = append(*resultOut[lastIndex].Shares, result.ShareList)
			}
		} else {
			resultOut = append(resultOut, result.AudioList)
			lastIndex := len(resultOut) - 1
			emptyShare := storage.ShareList{}
			if result.ShareList != emptyShare {
				newShares := make([]storage.ShareList, 0, 1)
				resultOut[lastIndex].Shares = &newShares
				*resultOut[lastIndex].Shares = append(*resultOut[lastIndex].Shares, result.ShareList)
			}
		}
		lastAudio = result.AudioList
	}

	return storage.AudioListJson{TotalCount: totalCount, Records: resultOut}, err
}

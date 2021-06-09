package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	storage "github.com/mahadeva604/audio-storage"
)

type SharePostgres struct {
	db *sqlx.DB
}

func NewSharePostgres(db *sqlx.DB) *SharePostgres {
	return &SharePostgres{db: db}
}

func (r *SharePostgres) ShareAudio(userID, audioId, shareId int) error {
	query := fmt.Sprintf(`INSERT INTO %s SELECT audio_id, $1 FROM %s
								WHERE audio_id = $2 and user_id = $3`, sharesTable, audiosTable)
	result, err := r.db.Exec(query, shareId, audioId, userID)

	if _, ok := err.(*pq.Error); ok {
		switch err.(*pq.Error).Code {
		case "23505":
			return storage.ShareExists
		case "23503":
			return storage.ShareUserNotExists
		}
	}

	if err != nil {
		return err
	}

	if rowsAff, err := result.RowsAffected(); rowsAff == 0 && err == nil {
		return storage.NotOwner
	}

	return err
}

func (r *SharePostgres) UnshareAudio(userID, audioId, shareId int) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE audio_id = (SELECT audio_id FROM %s
								WHERE audio_id = $1 and user_id = $2) AND user_id = $3`, sharesTable, audiosTable)
	result, err := r.db.Exec(query, audioId, userID, shareId)

	if err != nil {
		return err
	}

	if rowsAff, err := result.RowsAffected(); rowsAff == 0 && err == nil {
		return storage.NotOwner
	}

	return err
}

func (r *SharePostgres) GetSharedList(input storage.ShareListParam) (storage.ShareListJson, error) {

	query := fmt.Sprintf(`SELECT count(*) OVER() AS full_count, a.user_id, name, count(*) AS count
								FROM %s s
								JOIN %s a USING (audio_id)
								JOIN %s u ON a.user_id = u.user_id
								GROUP BY a.user_id, name ORDER BY name
								OFFSET $1 LIMIT $2`, sharesTable, audiosTable, usersTable)

	rows, err := r.db.Queryx(query, input.Offset, input.Limit)
	if err != nil {
		return storage.ShareListJson{}, err
	}
	var result storage.ShareListDb
	var totalCount int
	shareList := make([]storage.ShareListCount, 0)

	for rows.Next() {

		err := rows.StructScan(&result)
		if err != nil {
			return storage.ShareListJson{}, err
		}
		totalCount = result.Count
		shareList = append(shareList, result.ShareListCount)
	}

	return storage.ShareListJson{Count: totalCount, Users: shareList}, err
}

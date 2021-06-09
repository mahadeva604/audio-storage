package storage

type ShareInput struct {
	ShareTo int `json:"share_to" binding:"required"`
}

type ShareListParam struct {
	Limit  *int `json:"limit" form:"limit" binding:"required"`
	Offset *int `json:"offset" form:"offset" binding:"required"`
}

type ShareListCount struct {
	UserId     int    `json:"id" db:"user_id"`
	Name       string `json:"name" db:"name"`
	ShareCount int    `json:"shared_records" db:"count"`
}

type ShareListDb struct {
	Count int `db:"full_count"`
	ShareListCount
}

type ShareListJson struct {
	Count int              `json:"total_count"`
	Users []ShareListCount `json:"users"`
}

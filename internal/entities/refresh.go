package entities

import "time"

type RefreshToken struct {
	UserID  string
	Hash    string
	Expires time.Time
}

func NewRefresh(userID string, hash string, exp time.Time) RefreshToken {
	return RefreshToken{
		UserID:  userID,
		Hash:    hash,
		Expires: exp,
	}
}

package entities

import "time"

type RefreshToken struct {
	UserID  string
	Token   string
	Expires time.Time
}

func NewRefresh(userID string, token string, exp time.Time) RefreshToken {
	return RefreshToken{
		UserID:  userID,
		Token:   token,
		Expires: exp,
	}
}

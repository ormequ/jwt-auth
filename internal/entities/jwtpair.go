package entities

type JWTPair struct {
	Access  string
	Refresh string
}

func NewPair(access string, refresh string) JWTPair {
	return JWTPair{
		Access:  access,
		Refresh: refresh,
	}
}

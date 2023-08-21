package bcrypt

import "context"

type BCrypt struct {
	cost int
}

func (b BCrypt) Generate(ctx context.Context, token string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func New(cost int) BCrypt {
	return BCrypt{cost: cost}
}

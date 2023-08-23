package tests

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	client := setupClient(time.Second*2, time.Second*4)
	usr := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	p1, err := client.generate(usr)
	require.NoError(t, err, "correct generation")
	accUsr, err := decodeAccess(p1.Access)
	require.NoError(t, err, "correct generation")
	require.Equal(t, usr, accUsr, "correct generation")

	_, err = client.generate("")
	require.ErrorIs(t, err, ErrBadRequest, "empty user id")

	_, err = client.generate("замена__-на__-русс-кие_-символы_____")
	require.ErrorIs(t, err, ErrBadRequest, "non-uuid")

	time.Sleep(time.Second)
	p3, err := client.generate(usr)
	require.NoError(t, err, "repeat generation")
	require.NotEqual(t, p1.Access, p3.Access, "repeat generation")
	require.NotEqual(t, p1.Refresh, p3.Refresh, "repeat generation")
}

func TestRefresh(t *testing.T) {
	client := setupClient(time.Second*2, time.Second*4)
	usr := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	notFound := "f47ac10b-58cc-4372-a567-0e02b2c3d479"

	gen, err := client.generate(usr)
	require.NoError(t, err, "correct generating")
	time.Sleep(time.Second)

	ref, err := client.refresh(gen.Access, gen.Refresh)
	require.NoError(t, err, "correct refreshing")
	require.NotEqual(t, gen.Access, ref.Access, "correct refreshing")
	require.NotEqual(t, gen.Refresh, ref.Refresh, "correct refreshing")

	_, err = client.refresh(gen.Access, gen.Refresh)
	require.ErrorIs(t, err, ErrForbidden, "refreshing with old refresh")

	_, err = client.refresh("not-access-token", ref.Refresh)
	require.ErrorIs(t, err, ErrBadRequest, "refreshing with invalid access token")

	_, err = client.refresh(ref.Access, "не-рефреш-токен")
	require.ErrorIs(t, err, ErrBadRequest, "refreshing with invalid refresh token")

	_, err = client.refresh(encodeToken(notFound, time.Minute, []byte(accessSecret)), ref.Refresh)
	require.ErrorIs(t, err, ErrNotFound, "refreshing with not found user")

	time.Sleep(time.Second)
	ref, err = client.refresh(gen.Access, ref.Refresh)
	require.NoError(t, err, "refreshing with old expired access")

	time.Sleep(time.Second)
	gen2, err := client.generate(usr)
	require.NoError(t, err)
	_, err = client.refresh(gen2.Access, ref.Refresh)
	require.ErrorIs(t, err, ErrForbidden, "refreshing after generating new pair")

	time.Sleep(time.Second * 5)
	_, err = client.refresh(gen2.Access, gen2.Refresh)
	require.ErrorIs(t, err, ErrForbidden, "refreshing with expired refresh token")
}

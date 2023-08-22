package tests

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestTokenInformation(t *testing.T) {
	client := setupClient(time.Second*2, time.Second*10)
	usr := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	// Обычная генерация
	p1, err := client.generate(usr)
	require.NoError(t, err)
	acc1, err := decodeAccess(p1.Access)
	require.NoError(t, err)
	require.Equal(t, strings.Split(p1.Refresh, ".")[2], acc1.Refresh)
	require.Equal(t, usr, acc1.User)
	ref1, err := decodeRefresh(p1.Refresh)
	require.NoError(t, err)
	require.Equal(t, usr, ref1.User)

	// Пустая строка
	_, err = client.generate("")
	require.ErrorIs(t, ErrBadRequest, err)

	// Non-UUID
	_, err = client.generate("замена__-на__-русс-кие_-символы_____")
	require.ErrorIs(t, ErrBadRequest, err)

	// Повторная генерация
	time.Sleep(time.Second)
	p3, err := client.generate(usr)
	require.NoError(t, err)
	require.NotEqual(t, p1.Access, p3.Access)
	require.NotEqual(t, p1.Refresh, p3.Refresh)

}

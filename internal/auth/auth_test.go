package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	cdn_go "animakuro/cdn"
	"github.com/cristalhq/jwt/v4"
	"github.com/stretchr/testify/require"
)

func TestValidateToken(t *testing.T) {
	key := "abcd"

	signer, _ := jwt.NewSignerHS(jwt.HS256, []byte(key))

	builder := jwt.NewBuilder(signer)

	t.Run("valid payload", func(t *testing.T) {

		t.Parallel()

		// Payload inside token
		payload := &Claims{
			Bucket: "bucket",
			FileID: "1234",
		}

		expectedPayload := &Claims{
			Bucket: "bucket",
			FileID: "1234",
		}

		token, _ := builder.Build(payload)

		ok, err := ValidateToken(token.Bytes(), []string{key}, expectedPayload)
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("invalid payload", func(t *testing.T) {
		t.Parallel()

		// Payload inside token
		payload := map[string]interface{}{
			"some-bullshit": "jomama",
		}

		expectedPayload := &Claims{
			Bucket: "bucket",
			FileID: "1234",
		}

		token, _ := builder.Build(payload)
		ok, err := ValidateToken(token.Bytes(), []string{key}, expectedPayload)
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("invalid key", func(t *testing.T) {
		t.Parallel()

		// Payload inside token
		payload := &Claims{
			Bucket: "bucket",
			FileID: "1234",
		}

		expectedPayload := &Claims{
			Bucket: "bucket",
			FileID: "1234",
		}

		token, _ := builder.Build(payload)
		ok, err := ValidateToken(token.Bytes(), []string{"some-bullshit-key"}, expectedPayload)
		require.NoError(t, err)
		require.False(t, ok)
	})

}

func TestParseToken(t *testing.T) {
	key := "abcd"

	signer, _ := jwt.NewSignerHS(jwt.HS256, []byte(key))

	builder := jwt.NewBuilder(signer)

	token, err := builder.Build(nil)
	require.NoError(t, err)

	t.Run("operation get. Get token from url", func(t *testing.T) {
		t.Parallel()
		fullPath := fmt.Sprintf("random-host.com/abdc-efgh?auth=%s", token)

		parsedUrl, err := url.Parse(fullPath)
		require.NoError(t, err)

		// Can set tokenSource to empty because operation is get
		parsedToken, err := ParseToken(cdn_go.OperationGet, parsedUrl, "" /* token source */)
		require.NoError(t, err)
		require.EqualValues(t, token.Bytes(), parsedToken)
	})

	t.Run("operation get. Missing auth key", func(t *testing.T) {
		t.Parallel()
		// Without auth=...
		fullPath := fmt.Sprintf("random-host.com/abdc-efgh")

		parsedUrl, err := url.Parse(fullPath)
		require.NoError(t, err)

		// Can set tokenSource to empty because operation is get
		parsedToken, err := ParseToken(cdn_go.OperationGet, parsedUrl, "" /* token source */)
		require.Error(t, err)
		require.Equal(t, ErrMissingAuthKey.Error(), err.Error())
		require.Nil(t, parsedToken)
	})

	t.Run("operation other than get. Get token from header", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodPost, "google.com", nil)
		require.NoError(t, err)

		// Set token
		req.Header.Set("Authorization", "bearer "+token.String())

		// Can set url to nil because operation is not get
		parsedToken, err := ParseToken(cdn_go.OperationPost, nil /* url */, req.Header.Get("Authorization") /* tokenSource */)
		require.NoError(t, err)
		require.EqualValues(t, token.Bytes(), parsedToken)
	})

	t.Run("invalid authorization header (no bearer)", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodPost, "google.com", nil)
		require.NoError(t, err)

		// Set token
		req.Header.Set("Authorization", token.String())

		// Can set url to nil because operation is not get
		parsedToken, err := ParseToken(cdn_go.OperationPost, nil /* url */, req.Header.Get("Authorization") /* tokenSource */)
		require.Error(t, err)
		require.Equal(t, ErrInvalidAuthHeader.Error(), err.Error())
		require.Nil(t, parsedToken)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodPost, "google.com", nil)
		require.NoError(t, err)

		// Dont set token
		//req.Header.Set("Authorization", token.String())

		// Can set url to nil because operation is not get
		parsedToken, err := ParseToken(cdn_go.OperationPost, nil /* url */, req.Header.Get("Authorization") /* tokenSource */)
		require.Error(t, err)
		require.Equal(t, ErrMissingAuthHeader.Error(), err.Error())
		require.Nil(t, parsedToken)
	})

}

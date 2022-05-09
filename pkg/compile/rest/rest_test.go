package rest

import (
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

type bytesReadCloser struct {
	bytes []byte
}

func (b *bytesReadCloser) Read(p []byte) (n int, err error) {
	copy(p, b.bytes)
	return len(b.bytes), nil
}

func (b *bytesReadCloser) Close() error {
	return nil
}

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockPing(nil)
		defer monkey.UnpatchAll()
		_, err := New()
		assert.Nil(t, err)
	})
}

func TestPing(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := newClient(http.DefaultClient)
		mockPost(client, nil)
		defer monkey.UnpatchAll()
		err := Ping(client)
		assert.Nil(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		client := newClient(http.DefaultClient)
		mockPost(client, assert.AnError)
		defer monkey.UnpatchAll()
		err := Ping(client)
		assert.NotNil(t, err)
	})
}

func TestGetPostPutDelete(t *testing.T) {
	t.Run("t1", func(t *testing.T) {
		client := newClient(http.DefaultClient)
		mockDo(client)
		defer monkey.UnpatchAll()
		_, err := client.Get("/", &struct{ Data string }{})
		assert.Nil(t, err)
		_, err = client.Post("/", &struct{ Data string }{}, &struct{ Data string }{})
		assert.Nil(t, err)
		_, err = client.Put("/", &struct{ Data string }{}, &struct{ Data string }{})
		assert.Nil(t, err)
		_, err = client.Delete("/", &struct{ Data string }{}, &struct{ Data string }{})
		assert.Nil(t, err)
	})
}

func TestCompile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := newClient(http.DefaultClient)
		mockPost(client, nil)
		defer monkey.UnpatchAll()
		_, err := client.Compile(nil)
		assert.Nil(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		client := newClient(http.DefaultClient)
		mockPost(client, assert.AnError)
		defer monkey.UnpatchAll()
		_, err := client.Compile(nil)
		assert.NotNil(t, err)
	})
}

func mockPing(mockErr error) {
	monkey.Patch(Ping, func(_ *Client) error {
		return mockErr
	})
}

func mockPost(client *Client, mockErr error) {
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Post", func(_ *Client, _ string, _, _ interface{}) (*http.Response, error) {
		return nil, mockErr
	})
}

func mockDo(client *Client) {
	monkey.PatchInstanceMethod(reflect.TypeOf(client.client), "Do", func(_ *http.Client, _ *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       &bytesReadCloser{bytes: []byte("{\"data\":\"test\"}")},
		}, nil
	})
}

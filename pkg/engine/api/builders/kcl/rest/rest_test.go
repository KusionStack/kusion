package rest

import (
	"net/http"
	"testing"

	"github.com/bytedance/mockey"
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
	mockey.PatchConvey("success", t, func() {
		mockPing(nil)
		_, err := New()
		assert.Nil(t, err)
	})
}

func TestPing(t *testing.T) {
	mockey.PatchConvey("success", t, func() {
		client := newClient(http.DefaultClient)
		mockPost(client, nil)
		err := Ping(client)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("failed", t, func() {
		client := newClient(http.DefaultClient)
		mockPost(client, assert.AnError)
		err := Ping(client)
		assert.NotNil(t, err)
	})
}

func TestGetPostPutDelete(t *testing.T) {
	mockey.PatchConvey("t1", t, func() {
		client := newClient(http.DefaultClient)
		mockDo(client)
		resp1, err := client.Get("/", &struct{ Data string }{})
		assert.Nil(t, err)
		defer resp1.Body.Close()

		resp2, err := client.Post("/", &struct{ Data string }{}, &struct{ Data string }{})
		assert.Nil(t, err)
		defer resp2.Body.Close()

		resp3, err := client.Put("/", &struct{ Data string }{}, &struct{ Data string }{})
		assert.Nil(t, err)
		defer resp3.Body.Close()

		resp4, err := client.Delete("/", &struct{ Data string }{}, &struct{ Data string }{})
		assert.Nil(t, err)
		defer resp4.Body.Close()
	})
}

func TestCompile(t *testing.T) {
	mockey.PatchConvey("success", t, func() {
		client := newClient(http.DefaultClient)
		mockPost(client, nil)
		_, err := client.Compile(nil)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("failed", t, func() {
		client := newClient(http.DefaultClient)
		mockPost(client, assert.AnError)
		_, err := client.Compile(nil)
		assert.NotNil(t, err)
	})
}

func mockPing(mockErr error) {
	mockey.Mock(Ping).To(func(_ *Client) error {
		return mockErr
	}).Build()
}

func mockPost(client *Client, mockErr error) {
	mockey.Mock(mockey.GetMethod(client, "Post")).To(func(_ *Client, _ string, _, _ interface{}) (*http.Response, error) {
		return nil, mockErr
	}).Build()
}

func mockDo(client *Client) {
	mockey.Mock(mockey.GetMethod(client.client, "Do")).To(func(_ *http.Client, _ *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       &bytesReadCloser{bytes: []byte("{\"data\":\"test\"}")},
		}, nil
	}).Build()
}

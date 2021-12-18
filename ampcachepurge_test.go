package ampcachepurge

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHttpClient struct {
	mock.Mock
}

func (m *MockHttpClient) Get(url string) (*http.Response, error) {

	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)

}

func TestCheckCacheExistsReturnsTrueWhen200(t *testing.T) {
	httpClient := new(MockHttpClient)
	httpClient.On("Get", mock.Anything).Return(&http.Response{
		StatusCode: 200,
	}, nil)
	res := checkCacheExists("/some-url", httpClient)
	assert := assert.New(t)
	assert.True(res)
}

func TestCheckCacheExistsReturnsFalseWhen404(t *testing.T) {
	httpClient := new(MockHttpClient)
	httpClient.On("Get", mock.Anything).Return(&http.Response{
		StatusCode: 404,
	}, nil)
	res := checkCacheExists("/some-url", httpClient)
	assert := assert.New(t)
	assert.False(res)
}
func TestCheckCacheExistsReturnsFalseWhen500(t *testing.T) {
	httpClient := new(MockHttpClient)
	httpClient.On("Get", mock.Anything).Return(&http.Response{
		StatusCode: 500,
	}, nil)
	res := checkCacheExists("/some-url", httpClient)
	assert := assert.New(t)
	assert.False(res)
}

func TestMakePurgeRequestSuccess(t *testing.T) {
	httpClient := new(MockHttpClient)
	body := ioutil.NopCloser(bytes.NewReader([]byte("OK")))
	httpClient.On("Get", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       body,
	}, nil)

	err := makePurgeRequest("/some-url", httpClient)
	assert.Nil(t, err)
}

func TestMakePurgeRequestError(t *testing.T) {
	httpClient := new(MockHttpClient)
	body := ioutil.NopCloser(bytes.NewReader([]byte("OK")))
	httpClient.On("Get", mock.Anything).Return(&http.Response{
		StatusCode: 403,
		Body:       body,
	}, nil)

	err := makePurgeRequest("/some-url", httpClient)
	assert.NotNil(t, err)
}

func TestPanicWhenPrivateKeyCantBeLoaded(t *testing.T) {
	location := ""
	password := ""
	assert.Panics(t, func() { loadPrivateKey(location, password) })
}

func TestPreparePurgingUrl(t *testing.T) {
	parsedUrl, err := url.Parse("https://www.example.com/amp/test-amp-page")
	if err != nil {
		fmt.Println(err.Error())
	}
	ampCDNUrl := "https://www-example-com.cdn.ampproject.org"
	var timestamp int64 = 1637421562

	result := preparePurgingUrl(ampCDNUrl, "c", parsedUrl, timestamp)
	expected := "https://www-example-com.cdn.ampproject.org/update-cache/c/s/www.example.com/amp/test-amp-page?amp_action=flush&amp_ts=1637421562&amp_url_signature=B1qX89cO_i4emMiC9T6ITiBwSNwJXL9y_86AoE0hbI4EwkAKW89TmwLqI0rvZG9hwAdYvYsfMw2vAg7ygrvWfHN18hrhWiZg_AL8NEV0Jk_IRrAYfah7s1_QDOLC5FLbDE-z9Lo-NnaEfYjlA-Cc7-jnFDa5GN6CSS_tcb-suBkmPKWjr_E9eSVfFcoNuEuVChEnlzDft1IMUJCZ3kFMSRtp4ZbGclJxqYJwVmvCkEY7HXptVi4kGGjl_qJtc3Qt3sDWmIzXiXwG5OA7U8xAcyF1b_XkoPpxp5D8ZT437OGswdPlJ6T7elx1rycccozpyZCS3fId6mNSAGd5SdNoUQ"
	assert.Equal(t, result, expected)
}

func TestUrlSignature(t *testing.T) {
	expected := "B1qX89cO_i4emMiC9T6ITiBwSNwJXL9y_86AoE0hbI4EwkAKW89TmwLqI0rvZG9hwAdYvYsfMw2vAg7ygrvWfHN18hrhWiZg_AL8NEV0Jk_IRrAYfah7s1_QDOLC5FLbDE-z9Lo-NnaEfYjlA-Cc7-jnFDa5GN6CSS_tcb-suBkmPKWjr_E9eSVfFcoNuEuVChEnlzDft1IMUJCZ3kFMSRtp4ZbGclJxqYJwVmvCkEY7HXptVi4kGGjl_qJtc3Qt3sDWmIzXiXwG5OA7U8xAcyF1b_XkoPpxp5D8ZT437OGswdPlJ6T7elx1rycccozpyZCS3fId6mNSAGd5SdNoUQ"
	path := "/update-cache/c/s/www.example.com/amp/test-amp-page?amp_action=flush&amp_ts=1637421562"
	result := signPath(path)
	assert.Equal(t, result, expected)
}

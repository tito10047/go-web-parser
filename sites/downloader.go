package sites

import (
	"net/http"
	"bytes"
	"io"
	"io/ioutil"
	"crypto/tls"
	"strconv"
	"errors"
	"encoding/json"
)

type contentType string

const(
	UrlEncoded contentType = "application/x-www-form-urlencoded"
)

type downloader struct {
	Request *http.Request
}

func NewDownloader(method, url string) (*downloader, error) {

	r, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("User-Agent","Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36")
	d := &downloader{r}
	return d, nil
}

func (r *downloader) Param(body map[string]string, contentType contentType) {
	requestBody := ""
	if body != nil && len(body) > 0 {
		for key, value := range body {
			requestBody += key + "=" + value + "&"
		}
		requestBody = requestBody[0:len(requestBody)-1]
	}
	passBytes := []byte(requestBody)
	var b io.Reader
	b = bytes.NewBuffer(passBytes)
	rc, ok := b.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(b)
	}
	r.Request.Body = rc

	switch v := b.(type) {
	case *bytes.Buffer:
		r.Request.ContentLength = int64(v.Len())
		buf := v.Bytes()
		r.Request.GetBody = func() (io.ReadCloser, error) {
			r := bytes.NewReader(buf)
			return ioutil.NopCloser(r), nil
		}
	default:
		panic(errors.New("bad type"))
	}
	length :=strconv.Itoa(len(passBytes))

	r.Request.Header.Set("Content-Type", string(contentType))
	r.Request.Header.Set("content-length", length)
}

func (r *downloader) Download() (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(r.Request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, response.Body)
	return buf.String(), nil
}

func (r *downloader) DownloadJson(object interface{}) error {
	jsonStream, err := r.Download()
	if err!=nil {
		return err
	}
	err = json.Unmarshal([]byte(jsonStream),object)
	if err!=nil {
		return err
	}
	return nil
}

func (r *downloader) AcceptJson() {
	r.Request.Header.Add("Accept", "application/json")
}

package ossapi

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	//"bytes"
	"encoding/base64"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	//"errors"
)

const (
	DefaultContentType     = "application/octet-stream"
	SelfDefineHeaderPrefix = "x-oss-"
	DefaultHost            = "http://oss.aliyuncs.com"
)

type OSS struct {
	AccessId   string
	AccessKey  string
	signedAuth string
	Client     *http.Client
}

func NewOSS(id string, key string) *OSS {
	return &OSS{id, key, "", &http.Client{}}
}

func (oss *OSS) SetAccessId(id string) {
	oss.AccessId = id
}

func (oss *OSS) SetAccessKey(id string) {
	oss.AccessKey = id
}

func (oss *OSS) getAssign(method string, headers http.Header, resource string) string {

	contentMd5 := headers.Get("Content-Md5")
	contentType := headers.Get("Content-Type")
	date := headers.Get("Date")
	canonicalizedResource := resource
	CanonicalizeHeader(headers)

	var canonicalizedHeaders []string = make([]string, 0)

	for k, _ := range headers {
		if strings.HasPrefix(k, SelfDefineHeaderPrefix) {
			canonicalizedHeaders = append(canonicalizedHeaders, k)
		}
	}

	sort.Strings(canonicalizedHeaders)

	var canonicalizedHeader string

	for _, v := range canonicalizedHeaders {
		canonicalizedHeader += v + ":" + headers.Get(v) + "\n"
	}

	stringToSign := method + "\n" + contentMd5 + "\n" + contentType + "\n" + date + "\n" + canonicalizedHeader + canonicalizedResource
	h := hmac.New(sha1.New, []byte(oss.AccessKey))
	log.Printf("stringToSing: %s", stringToSign)

	h.Write([]byte(stringToSign))

	return strings.TrimSpace(base64.StdEncoding.EncodeToString(h.Sum(nil)))
}

func (oss *OSS) SignUrlAuthWithExpireTime(method, urladdr string, headers http.Header, resource string, timeout int) string {
	sendTime := headers.Get("Date")
	if sendTime == "" {
		sendTime = fmt.Sprint("%d", time.Now().Unix())
	}
	headers.Add("Date", sendTime)
	auth := oss.getAssign(method, headers, resource)
	params := make(url.Values)
	params.Add("OSSAccessKeyId", oss.AccessId)
	params.Add("Expires", sendTime)
	params.Add("Signature", auth)
	return url.QueryEscape(urladdr + "?" + params.Encode())
}

func (oss *OSS) createSignForNormalAuth(method string, headers http.Header, resource string) string {
	return "OSS " + oss.AccessId + ":" + oss.getAssign(method, headers, resource)
}

//Have to break the abstraction to append keys with lower case.
func CanonicalizeHeader(headers http.Header) {
	for k, _ := range headers {
		if lower := strings.ToLower(k); strings.HasPrefix(lower, SelfDefineHeaderPrefix) {
			copy(headers[lower], headers[k])
			delete(headers, k)
		}
	}
}

func (oss *OSS) BucketOp(method, bucket string, headers http.Header, params url.Values) *http.Response {
	if headers == nil {
		headers = make(http.Header)
	}

	if params == nil {
		params = make(url.Values)
	}
	url := "/" + bucket + "?" + params.Encode()
	headers.Add("Date", GetGMTime())
	headers.Add("Host", DefaultHost)
	var resource string
	if acl := params.Get("acl"); acl != "" {
		resource = "/" + bucket + "/?acl"
	} else {
		resource = "/" + bucket
	}
	return oss.Do(method, url, resource, strings.NewReader(""), headers)
}

func (oss *OSS) GetService() *http.Response {
	method := "GET"
	url := "/"
	date := GetGMTime()
	headers := make(http.Header)

	headers.Add("Date", date)
	headers.Add("Host", DefaultHost)
	resource := "/"
	return oss.Do(method, url, resource, strings.NewReader(""), headers)
}

func (oss *OSS) GetBucketAcl(bucket string) *http.Response {
	headers := make(http.Header)
	params := make(url.Values)
	params.Add("acl", "")
	return oss.BucketOp("GET", bucket, headers, params)
}

func (oss *OSS) GetBucket(bucket, prefix, marker, delimiter, maxkeys string, headers http.Header) *http.Response {
	return oss.ListBucket(bucket, prefix, marker, delimiter, maxkeys, headers)
}

func (oss *OSS) ListBucket(bucket, prefix, marker, delimiter, maxkeys string, headers http.Header) *http.Response {
	params := make(url.Values)
	if prefix != "" {
		params.Add("prefix", prefix)
	}
	if marker != "" {
		params.Add("marker", marker)
	}
	if delimiter != "" {
		params.Add("delimiter", delimiter)
	}
	if maxkeys != "" {
		params.Add("max-keys", maxkeys) //TODO: check max-keys <= 1000?
	}
	return oss.BucketOp("GET", bucket, headers, params)
}

func (oss *OSS) PutBucket(bucket, acl string, headers http.Header) *http.Response {
	if acl != "" {
		headers.Add("x-oss-acl", acl)
	}
	return oss.BucketOp("PUT", bucket, headers, nil)
}

func (oss *OSS) CreateBucket(bucket, acl string, headers http.Header) *http.Response {
	return oss.PutBucket(bucket, acl, headers)
}

func (oss *OSS) DeleteBucket(bucket string) *http.Response {
	return oss.BucketOp("DELETE", bucket, nil, nil)
}

func (oss *OSS) ObjectOp(method, bucket, object string, headers http.Header, data io.Reader) *http.Response {
	if headers == nil {
		headers = make(http.Header)
	}
	resource := "/" + bucket + "/" + object
	urladdr := resource
	date := GetGMTime()
	headers.Add("Date", date)
	headers.Add("Host", DefaultHost)

	return oss.Do(method, urladdr, resource, data, headers)
}

func (oss *OSS) PutObjectFromFile(bucket, object string, headers http.Header, file *os.File) *http.Response {
	if headers == nil {
		headers = make(http.Header)
	}
	if dotPos := strings.LastIndex(file.Name(), "."); dotPos == -1 {
		headers.Add("Content-Type", DefaultContentType)
	} else {
		if mimeType := mime.TypeByExtension(file.Name()[dotPos:]); mimeType == "" {
			headers.Add("Content-Type", DefaultContentType)
		} else {
			headers.Add("Content-Type", mimeType)
		}
	}
	stats, err := file.Stat()
	if err != nil {
		log.Panicf("Unable to read file %s stats.", file.Name())
		return nil
	}
	headers.Add("Content-Length", strconv.FormatInt(stats.Size(), 10))
	headers.Add("Expect", "100-Continue") //TODO: what's this?

	log.Printf("Header in file put: %v", headers)
	return oss.ObjectOp("PUT", bucket, object, headers, file)
}

func (oss *OSS) GetObject(bucket, object string, headers http.Header) *http.Response {
	if headers == nil {
		headers = make(http.Header)
	}
	return oss.ObjectOp("GET", bucket, object, headers, nil)
}

func (oss *OSS) GetObjectToFile(bucket, object, filename string, headers http.Header) error {
	res := oss.GetObject(bucket, object, headers)
	defer res.Body.Close()
	f, err := os.Create(filename)
	if err != nil {
		log.Panicf("Error creating file: %s", filename)
		return err
	}
	written, err := io.Copy(f, res.Body)
	if err != nil {
		log.Panicf("Unable to write to file :%s", filename)
		return err
	}
	log.Printf("Wrote to file %s : %d Bytes", filename, written)
	f.Close()
	return nil
}

func (oss *OSS) Do(method, url, resource string, body io.Reader, headers http.Header) *http.Response {
	if oss.AccessKey != "" {
		headers.Add("Authorization", oss.createSignForNormalAuth(method, headers, resource))
	} else {
		headers.Add("Authorization", oss.AccessId)
	}

	var host string
	if h := headers.Get("Host"); h != "" {
		host = h
	} else {
		host = DefaultHost
	}
	req, err := http.NewRequest(method, host+url, body)
	if err != nil {
		fmt.Print("Error Creating Request: \n", err)
	}
	req.Header = headers
	if contentLenStr := headers.Get("Content-Length"); contentLenStr != "" {
		if cLen, err := strconv.ParseInt(contentLenStr, 10, 64); err != nil {
			req.ContentLength = -1
		} else {
			req.ContentLength = cLen
		}
	}

	log.Printf("Request: %v\n", req)
	resp, err := oss.Client.Do(req)

	log.Printf("Response: %v\n", resp)
	if err != nil {
		fmt.Print("Error getting the resonse.\n")
	}

	return resp
}

func CopyHeader(header http.Header) (newHeader http.Header) {
	newHeader = make(http.Header)
	for k, v := range header {
		newSlice := make([]string, len(v))
		copy(newSlice, v)
		newHeader[k] = newSlice
	}
	return
}

func GetGMTime() string {
	return (time.Now().UTC()).Format(http.TimeFormat)
}

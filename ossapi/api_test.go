package ossapi

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var (
	accessId   = "*******"
	accessKey  = "*******"
	testBucket = "wliao"
	oss        = NewOSS(accessId, accessKey)
	testObject = "api_handler.go"
)

func TestPutObjectFromFile(t *testing.T) {
	file, err := os.Open("../README.md")
	if err != nil {
		log.Println(err)
		return
	}
	response := oss.PutObjectFromFile(testBucket, testObject, nil, file)
	if response.StatusCode != 200 {
		t.Error("Unable to put Object")
	}
}

func TestGetObject(t *testing.T) {
	response := oss.GetObject(testBucket, testObject, nil)
	if response.StatusCode != 200 {
		t.Error("Unable to get object")
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(contents))
}

func TestDelObject(t *testing.T) {
	response := oss.DelObject(testBucket, testObject, nil)
	if response.StatusCode != 204 {
		t.Error("Unable to del object")
		log.Println(response)
	}
}

func TestOssStruct(t *testing.T) {
	if oss.AccessId != accessId || oss.AccessKey != accessKey {
		t.Error("OSS struct not correctly initialized")
	}
}

func TestGetService(t *testing.T) {
	resp := oss.GetService()
	if resp.StatusCode != 200 {
		t.Error("Unable to get service")
	}
}

func TestGetBucket(t *testing.T) {
	resp := oss.GetBucket(testBucket, "", "", "", "", nil)
	if resp.StatusCode != 200 {
		t.Error("Unable to list Bucket with no params.")
	}
	resp = oss.GetBucket(testBucket, "", "", "", "10", nil)
	if resp.StatusCode != 200 {
		t.Error("Unable to list Bucket with 10 maxkeys")
	}
}

func TestPutBucket(t *testing.T) {
	resp := oss.PutBucket("wliao_1924", "", nil)
	if resp.StatusCode != 200 {
		t.Error("Unable to create a new bucket with no acl specified")
	}
	resp = oss.PutBucket("wliao_1925", "private", nil)
	if resp.StatusCode != 200 {
		t.Error("Unable to create a new bucket with private acl")
	}
}

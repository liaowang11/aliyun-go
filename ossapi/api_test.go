package ossapi

import (
	"testing"
)

var (
	accessId   = "*******"
	accessKey  = "*******"
	testBucket = "wliao"
	oss        = ossapi.NewOSS(accessId, accessKey)
)

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

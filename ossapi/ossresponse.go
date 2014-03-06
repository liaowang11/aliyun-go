package ossapi

import (
	"encoding/xml"
	"io"
)

type BucketList struct {
	Owner   Owner    `xml:"Owner"`
	Buckets []Bucket `xml:"Buckets>Bucket"`
}
type Owner struct {
	ID          string
	DisplayName string
}
type Bucket struct {
	Name         string
	CreationDate string
}
type ObjectList struct {
	BucketName  string `xml:"Name"`
	Prefix      string
	Marker      string
	MaxKeys     int
	Delimiter   string
	IsTruncated bool
	Object      []Object `xml:"Contents"`
}
type Object struct {
	Key          string
	LastModified string
	ETag         string
	Type         string
	Size         int
	StorageClass string
	Owner        Owner
}
type AccessControlPolicy struct {
	Owner  Owner
	Grants []string `xml:AccessControlList>Grant"`
}
type CopyObjectResult struct {
	LastModified string
	ETag         string
}

//Check status code before using any of the funcs here.
func ParseListBucket(content io.ReadCloser) (bucketList *BucketList, err error) {
	defer content.Close()
	bucketList = &BucketList{}
	decoder := xml.NewDecoder(content)
	err = decoder.Decode(bucketList)
	return
}

func ParseListObject(content io.ReadCloser) (objectList *ObjectList, err error) {
	defer content.Close()
	objectList = &ObjectList{}
	decoder := xml.NewDecoder(content)
	err = decoder.Decode(objectList)
	return
}

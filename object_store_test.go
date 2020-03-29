package main

var _ ObjectStore = new(MockObjectStore)

type MockObjectStore struct {
	putObject   func(bucketName string, object *Object) error
	listObjects func(bucketName string) []Object
	downloadURL func(bucketName string, objectName string) (string, error)
}

func (x MockObjectStore) PutObject(bucketName string, object *Object) error {
	return x.putObject(bucketName, object)
}

func (x MockObjectStore) ListObjects(bucketName string) []Object {
	return x.listObjects(bucketName)
}

func (x MockObjectStore) DownloadURL(bucketName string, objectName string) (string, error) {
	return x.downloadURL(bucketName, objectName)
}

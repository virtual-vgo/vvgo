package main

var _ ObjectStore = new(MockObjectStore)

type MockObjectStore struct {
	putObject   func(bucketName string, object *Object) error
	listObjects func(bucketName string) []Object
}

func (x MockObjectStore) PutObject(bucketName string, object *Object) error {
	return x.putObject(bucketName, object)
}

func (x MockObjectStore) ListObjects(bucketName string) []Object {
	return x.listObjects(bucketName)
}

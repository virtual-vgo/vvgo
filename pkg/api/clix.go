package api

import (
	"github.com/virtual-vgo/vvgo/pkg/clix"
	"net/http"
)

const ClixBucketName = "clix"
const ClixLockerKey = "clix.lock"

type ClixHandler struct {
	clix.Clix
}

func (x ClixHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) { notImplemented(w) }

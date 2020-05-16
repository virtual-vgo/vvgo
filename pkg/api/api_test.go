package api

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func init() {
	PublicFiles = "../../public"
}

var lrand = rand.New(rand.NewSource(time.Now().UnixNano()))

func newParts() *parts.RedisParts {
	return parts.NewParts("testing" + strconv.Itoa(lrand.Int()))
}

func newBucket(t *testing.T) *storage.Bucket {
	bucket, err := storage.NewBucket(context.Background(), "testing"+strconv.Itoa(lrand.Int()))
	require.NoError(t, err, "storage.NewBucket()")
	return bucket
}

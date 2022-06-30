package data

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jwvictor/cubby/pkg/types"
	"math/rand"
	"testing"
	"time"
)

func randomBlob(existing []*types.Blob) *types.Blob {
	data := fmt.Sprintf("%d", rand.Intn(999999999))
	title := fmt.Sprintf("%d", rand.Intn(999999999))
	parentId := ""
	if rand.Intn(2) == 1 && len(existing) > 2 {
		parentId = existing[rand.Intn(len(existing))].Id
	}
	return &types.Blob{
		Id:       uuid.New().String(),
		Title:    title,
		Data:     data,
		OwnerId:  "test",
		ParentId: parentId,
	}
}

func TestStaticFileProvider_GetBlob(t *testing.T) {
	ctx, cxl := context.WithCancel(context.Background())
	defer cxl()
	sfp := NewStaticFileProvider(ctx, "/tmp/test_out.dat")
	var blobs []*types.Blob
	N := 50000
	//t0 := time.Now()
	for i := 0; i < N; i++ {
		blob := randomBlob(blobs)
		blobs = append(blobs, blob)
		sfp.PutBlob(blob)
		fmt.Printf("Done putting blob %s\n", blob.Id)
		time.Sleep(1 * time.Millisecond)
	}
	//t1 := time.Now()
	t2 := time.Now()

	for i := 0; i < N; i++ {
		id := blobs[rand.Intn(len(blobs))].Id
		sfp.GetBlob(id, "test")
	}

	t3 := time.Now()
	t.Errorf("Took %d milliseconds to get %d blobs.\n", t3.Sub(t2).Milliseconds(), N)
}

package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jwvictor/cubby/pkg/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type CubbyDataProvider interface {
	GetBlob(blobId, userId string) *types.Blob
	DeleteBlob(blobId, userId string) *types.Blob
	GetBlobByPath(path, userId string) *types.Blob
	QueryBlobs(blobId, userId string) []*types.Blob
	ListBlobs(userId string) []*types.BlobSkeleton
	PutBlob(blob *types.Blob) error

	PutPost(post *types.Post) error
	GetPost(ownerId, postId string) *types.Post
	DeletePost(ownerId, postId string) bool
}

type StaticFileProvider struct {
	outputFilename string
	data           map[string]map[string]*types.Blob
	postData       map[string]map[string]*types.Post
	blobChannel    chan *types.Blob
	postChannel    chan *types.Post
	deleteChannel  chan *types.Blob
	lock           *sync.RWMutex
}

func loadUserData(filename string) map[string]*types.Blob {
	m := &StaticFileUserDump{}
	jsonFile, err := os.Open(filename)
	defer jsonFile.Close()
	if err == nil {
		bytes, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal(bytes, &m)
	}
	data := map[string]*types.Blob{}
	if m != nil && m.Blobs != nil && len(m.Blobs) > 0 {
		data = m.Blobs
	}
	return data
}

func loadPostData(dataPath string) (map[string]map[string]*types.Post, error) {
	filename := filepath.Join(dataPath, "posts.json")
	data := map[string]map[string]*types.Post{}
	jsonFile, err := os.Open(filename)
	defer jsonFile.Close()
	dump := &StaticPostsDump{}
	if err == nil {
		bytes, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal(bytes, &dump)
	}
	if dump != nil && dump.Posts != nil {
		data = dump.Posts
	}
	return data, nil
}

func loadBlobData(dataPath string) (map[string]map[string]*types.Blob, error) {
	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		return nil, err
	}
	data := map[string]map[string]*types.Blob{}
	for _, file := range files {
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".json")) {
			continue
		}
		fname := file.Name()
		uid := fname[:len(fname)-len(".json")]
		log.Printf("Loading user %s from file %s...\n", uid, fname)
		userDat := loadUserData(filepath.Join(dataPath, fname))
		data[uid] = userDat
	}
	return data, nil
}

func NewStaticFileProvider(context context.Context, outputFilename string) CubbyDataProvider {
	data, err := loadBlobData(outputFilename)
	if err != nil {
		log.Printf("Failed to load blob data: %s\n", err.Error())
		return nil
	}
	postData, err := loadPostData(outputFilename)
	if err != nil {
		log.Printf("Failed to load post data: %s\n", err.Error())
		return nil
	}
	sfp := &StaticFileProvider{outputFilename: outputFilename, data: data, postChannel: make(chan *types.Post, 32), deleteChannel: make(chan *types.Blob, 32), blobChannel: make(chan *types.Blob, 32), lock: &sync.RWMutex{}, postData: postData}
	go sfp.Run(context)
	return sfp
}

func findBlobByTitleSliceUnsafe(data []*types.Blob, title string) *types.Blob {
	for _, v := range data {
		if v.Deleted {
			continue
		}
		if strings.HasPrefix(strings.ToLower(v.Title), strings.ToLower(title)) {
			return v
		}
	}
	return nil
}

func findBlobByTitleMapUnsafe(data map[string]*types.Blob, title string, searchChildren bool) *types.Blob {
	for _, v := range data {
		if v.Deleted {
			continue
		}
		if !searchChildren {
			if v.ParentId != "" {
				continue
			}
		}
		if strings.HasPrefix(strings.ToLower(v.Title), strings.ToLower(title)) {
			return v
		}
	}
	return nil
}

func (t *StaticFileProvider) _getPathUnsafe(path string, userId string, data map[string]*types.Blob) *types.Blob {
	segments := strings.Split(path, ":")
	var cur *types.Blob
	for _, seg := range segments {
		if cur == nil {
			next := findBlobByTitleMapUnsafe(data, seg, false)
			if next == nil {
				break
			}
			cur = t.resolveChildren(next)
		} else {
			next := findBlobByTitleSliceUnsafe(cur.Children, seg)
			if next == nil {
				break
			}
			cur = t.resolveChildren(next)
		}
	}
	return cur
}

func (t *StaticFileProvider) DeletePost(ownerId, postId string) bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	if userData, ok0 := t.postData[ownerId]; ok0 {
		// user has posts
		_, exists := userData[postId]
		if exists {
			delete(userData, postId)
			t.postData[ownerId] = userData // dont think this is necessary
			return true
		}
		return false
	}
	return false
}

func (t *StaticFileProvider) DeleteBlob(blobId, userId string) *types.Blob {
	userBlobs := t.getUserBlobs(userId)

	t.lock.RLock()

	if blob, ok := userBlobs[blobId]; ok {
		t.lock.RUnlock()
		t.deleteChannel <- &types.Blob{OwnerId: userId, Id: blobId}
		for _, child := range blob.ChildIds {
			t.DeleteBlob(child, userId)
		}
		return blob
	} else {
		t.lock.RUnlock()
		return nil
	}
}

func (t *StaticFileProvider) getUserBlobs(userId string) map[string]*types.Blob {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if x, ok := t.data[userId]; ok {
		return x
	} else {
		t.data[userId] = map[string]*types.Blob{}
		return t.data[userId]
	}
}

func (t *StaticFileProvider) resolveChildren(blob *types.Blob) *types.Blob {
	b := blob.Clone()
	for _, id := range b.ChildIds {
		b.Children = append(b.Children, t.resolveChildren(t.GetBlob(id, blob.OwnerId)))
	}
	return b
}

func blobToSkeleton(blob *types.Blob) *types.BlobSkeleton {
	b := &types.BlobSkeleton{
		Id:         blob.Id,
		Title:      blob.Title,
		Type:       blob.Type,
		Importance: blob.Importance,
		Tags:       blob.Tags,
		Deleted:    blob.Deleted,
		Children:   nil,
		OwnerId:    blob.OwnerId,
		ParentId:   blob.ParentId,
	}
	for _, child := range blob.Children {
		if child.Deleted == true {
			continue
		}
		b.Children = append(b.Children, blobToSkeleton(child))
	}
	return b
}

func (t *StaticFileProvider) ListBlobs(userId string) []*types.BlobSkeleton {
	data := t.getUserBlobs(userId)
	t.lock.RLock()
	defer t.lock.RUnlock()
	var result []*types.BlobSkeleton
	for _, v := range data {
		// Non-deleted, root blobs only
		if v.Deleted == true || v.ParentId != "" {
			continue
		}
		blob := t.resolveChildren(v)
		result = append(result, blobToSkeleton(blob))
		if len(result) >= MaxSearchResults {
			break
		}
	}
	return result
}

const (
	MaxSearchResults = 50
)

func (t *StaticFileProvider) QueryBlobs(searchString, userId string) []*types.Blob {
	data := t.getUserBlobs(userId)
	t.lock.RLock()
	defer t.lock.RUnlock()
	var result []*types.Blob
	for _, v := range data {
		if v.Matches(searchString) && !v.Deleted {
			result = append(result, v)
		}
		if len(result) >= MaxSearchResults {
			break
		}
	}
	return result
}

func (t *StaticFileProvider) GetPost(ownerId, postId string) *types.Post {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if userData, ok0 := t.postData[ownerId]; ok0 {
		if x, ok := userData[postId]; ok {
			return x
		}
	}
	return nil
}

func (t *StaticFileProvider) GetBlobByPath(path, userId string) *types.Blob {
	data := t.getUserBlobs(userId)
	t.lock.RLock()
	defer t.lock.RUnlock()
	if x := t._getPathUnsafe(path, userId, data); x != nil {
		return x
	}
	return nil
}

func (t *StaticFileProvider) GetBlob(blobId, userId string) *types.Blob {
	data := t.getUserBlobs(userId)
	t.lock.RLock()
	defer t.lock.RUnlock()
	if x, ok := data[blobId]; ok {
		return t.resolveChildren(x)
	}

	return nil
}

type StaticFileUserDump struct {
	UserId    string                 `json:"user_id"`
	Blobs     map[string]*types.Blob `json:"blobs"`
	Timestamp time.Time              `json:"timestamp"`
}

type StaticFileDump struct {
	Blobs     map[string]map[string]*types.Blob `json:"blobs"`
	Timestamp time.Time                         `json:"timestamp"`
}

type StaticPostsDump struct {
	Posts     map[string]map[string]*types.Post `json:"posts"`
	Timestamp time.Time                         `json:"timestamp"`
}

func (t *StaticFileProvider) _dumpUserUnsafe(userId string) error {
	outFilename := filepath.Join(t.outputFilename, fmt.Sprintf("%s.json", userId))
	userData := t.data[userId]
	bs, err := json.Marshal(&StaticFileUserDump{UserId: userId, Blobs: userData, Timestamp: time.Now()})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outFilename, bs, 0644)
}

func (t *StaticFileProvider) _dumpPostsUnsafe() error {
	outFilename := filepath.Join(t.outputFilename, "posts.json")
	bs, err := json.Marshal(&StaticPostsDump{
		Posts:     t.postData,
		Timestamp: time.Now(),
	})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outFilename, bs, 0644)
}

func (t *StaticFileProvider) Dump() error {
	t.lock.RLock()
	defer t.lock.RUnlock()
	for uid, _ := range t.data {
		err := t._dumpUserUnsafe(uid)
		if err != nil {
			return err
		}
	}
	err := t._dumpPostsUnsafe()
	if err != nil {
		return err
	}
	return nil
}

func (t *StaticFileProvider) _getBlobByIdUnsafe(userId, blobId string) *types.Blob {
	if userData, ok := t.data[userId]; ok {
		if blob, ok2 := userData[blobId]; ok2 {
			return blob
		}
	}
	return nil
}

func (t *StaticFileProvider) PutPost(post *types.Post) error {
	if !types.ValidateCustomPostId(post.Id) {
		return errors.New("InvalidPostId")
	}
	t.postChannel <- post
	return nil
}

func (t *StaticFileProvider) PutBlob(blob *types.Blob) error {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if blob.Id == "" {
		// This is a put, not a set
		err := t._checkNewBlobUniquenessUnsafe(blob)
		if err != nil {
			return err
		}
		blob.Id = uuid.New().String()
	}
	t.blobChannel <- blob
	return nil
}

func (t *StaticFileProvider) _checkNewBlobUniquenessUnsafe(blob *types.Blob) error {
	if blob.ParentId != "" {
		// Check uniqueness under parent
		parent := t._getBlobByIdUnsafe(blob.OwnerId, blob.ParentId)
		for _, child := range parent.ChildIds {
			childBlob := t._getBlobByIdUnsafe(blob.OwnerId, child)
			if childBlob.Title == blob.Title && (!childBlob.Deleted) {
				return errors.New("NameCollision")
			}
		}
	} else {
		// Check uniqueness at root
		userData := t.data[blob.OwnerId]
		for _, other := range userData {
			if blob.Title == other.Title && (!other.Deleted) {
				return errors.New("NameCollision")
			}
		}
	}
	return nil
}

func (t *StaticFileProvider) _markDeletedUnsafe(userId, blobId string) bool {
	blobs, ok0 := t.data[userId]
	if !ok0 {
		return false
	}
	if blob, ok := blobs[blobId]; ok {
		blob.Deleted = true
		return true
	} else {
		return false
	}
}

func (t *StaticFileProvider) _markDeleted(userId, blobId string) bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t._markDeletedUnsafe(userId, blobId)
}

func arrayHas(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func (t *StaticFileProvider) _putChild(userId, parentBlobId, childBlobId string) bool {
	blobs := t.getUserBlobs(userId)
	t.lock.Lock()
	defer t.lock.Unlock()
	if parent, ok := blobs[parentBlobId]; ok {
		if !arrayHas(parent.ChildIds, childBlobId) {
			parent.ChildIds = append(parent.ChildIds, childBlobId)
		}
		return true
	} else {
		return false
	}
}

func (t *StaticFileProvider) _expireBlobUnsafe(userId, blobId string) {
	userData := t.data[userId]
	if v, ok := userData[blobId]; ok {
		childIds := v.ChildIds
		t._markDeletedUnsafe(userId, blobId)
		for _, childId := range childIds {
			t._expireBlobUnsafe(userId, childId)
		}
	}
}

func (t *StaticFileProvider) _expireBlobsForUserUnsafe(userId string) {
	// Iterate through root blobs
	for blobId, v := range t.data[userId] {
		if v.ExpireTime == nil || v.Deleted {
			continue
		}
		if time.Now().After(*v.ExpireTime) {
			// Expire this blob
			log.Printf("Expiring blob %s (for user %s)\n", blobId, userId)
			t._expireBlobUnsafe(userId, blobId)
		}
	}
}

func (t *StaticFileProvider) expireBlobs() {
	t.lock.Lock()
	defer t.lock.Unlock()
	for user, _ := range t.data {
		t._expireBlobsForUserUnsafe(user)
	}
}

func scrubBlobForHistory(blob *types.Blob) *types.Blob {
	histBlob := blob.Clone()
	histBlob.Children = nil
	histBlob.VersionHistory = nil
	return histBlob
}

func (t *StaticFileProvider) _putBlob(userId, blobId string, blob *types.Blob) {
	toInsert := blob.Clone()
	blobs := t.getUserBlobs(userId)
	t.lock.Lock()
	defer t.lock.Unlock()
	if currentVersion, ok := blobs[blobId]; ok {
		toInsert.VersionHistory = currentVersion.VersionHistory
	}
	toInsert.VersionHistory = append(toInsert.VersionHistory, &types.BlobHistoryItem{
		AsOf:      time.Now(),
		BlobValue: scrubBlobForHistory(blob),
	})
	toInsert.Children = nil
	//fmt.Printf("INSERTING THIS : \n%+v\n\n", toInsert)
	blobs[blobId] = toInsert
}

func (t *StaticFileProvider) _putPost(p *types.Post) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if _, ok := t.postData[p.OwnerId]; !ok {
		t.postData[p.OwnerId] = map[string]*types.Post{}
	}
	t.postData[p.OwnerId][p.Id] = p
}

func (t *StaticFileProvider) Run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			break
		case b := <-t.deleteChannel:
			t._markDeleted(b.OwnerId, b.Id)
		case b := <-t.blobChannel:
			if b.ParentId != "" {
				if success := t._putChild(b.OwnerId, b.ParentId, b.Id); success {
					t._putBlob(b.OwnerId, b.Id, b)
				} else {
					log.Printf("Invalid parent: %s\n", b.ParentId)
				}
			} else {
				log.Printf("Putting blob with ID %s (owner %s)\n", b.Id, b.OwnerId)
				t._putBlob(b.OwnerId, b.Id, b)
			}
		case p := <-t.postChannel:
			log.Printf("Putting post with ID %s (owner %s)...\n", p.Id, p.OwnerId)
			t._putPost(p)
		case <-ticker.C:
			t.expireBlobs()
			log.Printf("Dumping %d users' blobs as JSON...\n", len(t.data))
			t.Dump()
		}
	}
}

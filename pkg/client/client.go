package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jwvictor/cubby/pkg/types"
	"io"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"os"
	"time"
)

type CubbyClient struct {
	httpClient *http.Client
	host       string
	port       int
	jwtTokens  *types.AuthTokens
	userEmail  string
	userPass   string
}

func NewCubbyClient(host string, port int, userEmail, userPass string) *CubbyClient {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	return &CubbyClient{
		httpClient: client,
		host:       host,
		port:       port,
		userEmail:  userEmail,
		userPass:   userPass,
	}
}

func (c *CubbyClient) CheckVersions() (bool, string, error) {
	versions, err := c.Versions()
	if err != nil {
		return true, versions.UpgradeScriptUrl, err
	}
	if !types.IsVersionMin(versions) {
		fmt.Fprintf(os.Stderr, "Version %s is less than the minimum client version %s. Please upgrade with `cubby upgrade`.\n", types.ClientVersion, versions.MinClientVersion)
		os.Exit(1)
		return false, versions.UpgradeScriptUrl, nil
	} else if types.IsVersionLess(versions) {
		fmt.Fprintf(os.Stderr, "Version %s is behind latest client version %s. Please upgrade with `cubby upgrade`.\n", types.ClientVersion, versions.MinClientVersion)
		return true, versions.UpgradeScriptUrl, nil
	}
	return true, versions.UpgradeScriptUrl, nil
}

func (c *CubbyClient) FetchInstallScript(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("CouldNotGetScript")
	}
	return body, nil
}

func (c *CubbyClient) Versions() (*types.VersionResponse, error) {
	url := fmt.Sprintf("%s:%d/v1/version", c.host, c.port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("CouldNotGetVersions")
	}
	var versions *types.VersionResponse
	err = json.Unmarshal(body, &versions)
	if err != nil {
		return nil, err
	} else {
		return versions, nil
	}
}

func (c *CubbyClient) SearchUser(emailOrDisplayName string) (*types.UserResponse, error) {
	url := fmt.Sprintf("%s:%d/v1/users/search/%s", c.host, c.port, url2.QueryEscape(emailOrDisplayName))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.jwtTokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("CouldNotGetUser")
	}
	var user *types.UserResponse
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

func (c *CubbyClient) UserProfile() (*types.UserResponse, error) {
	url := fmt.Sprintf("%s:%d/v1/users/profile", c.host, c.port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.jwtTokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("CouldNotGetProfile")
	}
	var user *types.UserResponse
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

func (c *CubbyClient) SignUp(displayName string) error {
	url := fmt.Sprintf("%s:%d/v1/users/signup", c.host, c.port)
	jsonStr, err := json.Marshal(&types.BasicAuthCredentials{
		UserEmail:    c.userEmail,
		UserPassword: c.userPass,
		DisplayName:  displayName,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Sign up failed: %s.", string(body)))
	}
	var tokens *types.UserResponse
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		return err
	}
	fmt.Printf("Sign up successful! You don't need it, but your user ID is: %s\n", tokens.Id)
	return nil
}

func (c *CubbyClient) Authenticate() error {
	url := fmt.Sprintf("%s:%d/v1/users/authenticate", c.host, c.port)
	jsonStr, err := json.Marshal(&types.BasicAuthCredentials{
		UserEmail:    c.userEmail,
		UserPassword: c.userPass,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Token not obtained: %s.", string(body)))
	}
	var tokens *types.AuthTokens
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		return err
	}
	c.jwtTokens = tokens
	return nil
}

func (c *CubbyClient) checkAuthExists() bool {
	return c.jwtTokens != nil && c.jwtTokens.AccessToken != ""
}

func (c *CubbyClient) PutPost(blob *types.Post) (string, error) {
	if !c.checkAuthExists() {
		return "", errors.New("NotAuthenticated")
	}
	url := fmt.Sprintf("%s:%d/v1/posts", c.host, c.port)
	jsonStr, err := json.Marshal(blob)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", err
	}
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Authorization", "Bearer "+c.jwtTokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("PostNotCreated: %d", resp.StatusCode))
	}
	var newBlob *types.PostResponse
	err = json.Unmarshal(body, &newBlob)
	if err != nil {
		return "", err
	} else if newBlob.Posts[0].Id != "" {
		return newBlob.Posts[0].Id, nil
	} else {
		return "", fmt.Errorf("PostNotCreated: %d", resp.StatusCode)
	}
}

func (c *CubbyClient) PutBlob(blob *types.Blob) (string, error) {
	if !c.checkAuthExists() {
		return "", errors.New("NotAuthenticated")
	}
	url := fmt.Sprintf("%s:%d/v1/blobs", c.host, c.port)
	jsonStr, err := json.Marshal(blob)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", err
	}
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Authorization", "Bearer "+c.jwtTokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", errors.New(fmt.Sprintf("BlobNotCreated: %s", string(body)))
	}
	var newBlob *types.BlobResponse
	err = json.Unmarshal(body, &newBlob)
	if err != nil {
		return "", err
	} else if newBlob.Blobs[0].Id != "" {
		return newBlob.Blobs[0].Id, nil
	} else {
		return "", fmt.Errorf("BlobNotCreated: %d", resp.StatusCode)
	}
}

func (c *CubbyClient) DeletePost(ownerId, id string) (*types.Post, error) {
	if !c.checkAuthExists() {
		return nil, errors.New("NotAuthenticated")
	}
	url := fmt.Sprintf("%s:%d/v1/posts/%s/%s", c.host, c.port, ownerId, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.jwtTokens.AccessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New("CouldNotDeletePost")
	}
	//log.Printf("Body: %s\n", string(body))
	var post *types.PostResponse
	err = json.Unmarshal(body, &post)
	if err != nil {
		return nil, err
	}
	if post.Posts == nil {
		return nil, errors.New("CouldNotDeletePost")
	}
	return post.Posts[0], nil
}

func (c *CubbyClient) ListPublishedBlobs() ([]*types.Post, error) {
	if !c.checkAuthExists() {
		return nil, errors.New("NotAuthenticated")
	}
	url := fmt.Sprintf("%s:%d/v1/posts/list", c.host, c.port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.jwtTokens.AccessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New("CouldNotGetPosts")
	}
	var post *types.PostResponse
	err = json.Unmarshal(body, &post)
	if err != nil {
		return nil, err
	}
	if post.Posts == nil {
		return nil, errors.New("CouldNotGetPosts")
	}
	return post.Posts, nil
}

func (c *CubbyClient) GetPostById(ownerId, id string) (*types.PostResponse, error) {
	if !c.checkAuthExists() {
		return nil, errors.New("NotAuthenticated")
	}
	url := fmt.Sprintf("%s:%d/v1/posts/%s/%s", c.host, c.port, ownerId, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.jwtTokens.AccessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New("CouldNotGetPost")
	}
	//log.Printf("Body: %s\n", string(body))
	var post *types.PostResponse
	err = json.Unmarshal(body, &post)
	if err != nil {
		return nil, err
	}
	if post.Posts == nil {
		return nil, errors.New("CouldNotGetPost")
	}
	return post, nil
}

func (c *CubbyClient) GetBlobById(id string) (*types.Blob, error) {
	if !c.checkAuthExists() {
		return nil, errors.New("NotAuthenticated")
	}
	url := fmt.Sprintf("%s:%d/v1/blobs/%s", c.host, c.port, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.jwtTokens.AccessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New("CouldNotGetBlob")
	}
	//log.Printf("Body: %s\n", string(body))
	var blob *types.BlobResponse
	err = json.Unmarshal(body, &blob)
	if err != nil {
		return nil, err
	}
	if blob.Blobs == nil {
		return nil, errors.New("CouldNotGetBlob")
	}
	return blob.Blobs[0], nil
}

func (c *CubbyClient) DeleteBlob(blobId string) (*types.Blob, error) {
	if !c.checkAuthExists() {
		return nil, errors.New("NotAuthenticated")
	}
	url := fmt.Sprintf("%s:%d/v1/blobs/%s", c.host, c.port, blobId)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.jwtTokens.AccessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, errors.New("NotFound")
		}
		return nil, err
	}
	//log.Printf("Body: %s\n", string(body))
	var blob *types.BlobResponse
	err = json.Unmarshal(body, &blob)
	if err != nil {
		return nil, err
	}
	if blob.Blobs == nil {
		return nil, errors.New("ZeroBlobsMatch")
	}
	return blob.Blobs[0], nil
}

func (c *CubbyClient) SearchBlob(query string) ([]*types.Blob, error) {
	if !c.checkAuthExists() {
		return nil, errors.New("NotAuthenticated")
	}
	url := fmt.Sprintf("%s:%d/v1/blobs/search/%s", c.host, c.port, query)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.jwtTokens.AccessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, errors.New("NotFound")
		}
		return nil, err
	}
	//log.Printf("Body: %s\n", string(body))
	var blob *types.BlobResponse
	err = json.Unmarshal(body, &blob)
	if err != nil {
		return nil, err
	}
	if blob.Blobs == nil {
		return nil, errors.New("ZeroBlobsMatch")
	}
	return blob.Blobs, nil
}

func (c *CubbyClient) ListBlobs() ([]*types.BlobSkeleton, error) {
	if !c.checkAuthExists() {
		return nil, errors.New("NotAuthenticated")
	}
	url := fmt.Sprintf("%s:%d/v1/blobs/list", c.host, c.port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.jwtTokens.AccessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, errors.New("NotFound")
		}
		return nil, err
	}
	//log.Printf("Body: %s\n", string(body))
	var blob *types.BlobList
	err = json.Unmarshal(body, &blob)
	if err != nil {
		return nil, err
	}
	if blob.RootBlobs == nil {
		return nil, errors.New("ZeroBlobsMatch")
	}
	return blob.RootBlobs, nil
}

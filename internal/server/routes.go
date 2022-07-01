package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jwvictor/cubby/internal/data"
	"github.com/jwvictor/cubby/internal/users"
	"github.com/jwvictor/cubby/pkg/types"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
)

type Server struct {
	portNum       int
	router        *chi.Mux
	dataProvider  data.CubbyDataProvider
	jwtAuth       *jwtauth.JWTAuth
	userProvider  users.UserProvider
	staticData    staticData
	adminPassword string
}

type staticData struct {
	shareHtml  []byte
	splashHtml []byte
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

func mimeType(fn string) string {
	fnl := strings.ToLower(fn)
	if strings.HasSuffix(fnl, ".js") {
		return "application/json"
	} else if strings.HasSuffix(fnl, ".css") {
		return "text/css"
	} else if strings.HasSuffix(fnl, ".html") {
		return "text/html"
	} else if strings.HasSuffix(fnl, ".png") {
		return "image/png"
	}

	return "text/plain"
}

func NewServer(portNum int, adminPass string) *Server {
	usersStore, err := users.NewUsersFileStore("cubby-users.json")
	if err != nil {
		panic(err)
	}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - nothing here"))
	})

	// TODO - parameterize this? Infer path of binary?
	shareHtml, err := ioutil.ReadFile("./static/share.html")
	if err != nil {
		panic(err)
	}
	splashHtml, err := ioutil.ReadFile("./static/splash.html")
	if err != nil {
		panic(err)
	}

	server := &Server{
		portNum:       portNum,
		router:        router,
		dataProvider:  data.NewStaticFileProvider(context.Background(), "./data"),
		jwtAuth:       tokenAuth,
		userProvider:  usersStore,
		staticData:    staticData{shareHtml: shareHtml, splashHtml: splashHtml},
		adminPassword: adminPass,
	}
	router.Get("/", server.SplashPage)
	router.Route("/static", func(staticRouter chi.Router) {
		staticRouter.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			fname := "./static/" + strings.TrimPrefix(r.URL.Path, "/static/")
			bs, err := ioutil.ReadFile(fname)
			if err != nil {
				log.Printf("Failed to open file: %s\n", fname)
				w.WriteHeader(404)
				w.Write([]byte("not found: " + fname))
			}
			mt := mimeType(fname)
			//log.Printf("MIME type for %s: %s\n", fname, mt)
			w.Header().Add("Content-Type", mt)
			w.Write(bs)
		})
	})
	router.Route("/v1", func(v1Router chi.Router) {

		// Version route
		v1Router.Get("/version", server.GetVersion)
		v1Router.Post("/stats", server.GetStats)

		// Authenticated post routes
		v1Router.Route("/posts", func(postsRouter chi.Router) {
			postsRouter.Use(jwtauth.Verifier(tokenAuth))
			postsRouter.Use(jwtauth.Authenticator)
			postsRouter.Post("/", server.CreatePost)
			postsRouter.Get("/list", server.ListPosts)
			postsRouter.Route("/{ownerName}/{postId}", func(postRouter chi.Router) {
				postRouter.Use(server.PostCtx)
				postRouter.Get("/", server.GetPost)
				postRouter.Delete("/", server.DeletePost)
			})
		})

		// Unauthenticated post routes
		v1Router.Route("/post", func(postRouter chi.Router) {
			postRouter.Route("/{ownerName}/{postId}", func(postRouter chi.Router) {
				postRouter.Use(server.PostCtx)
				postRouter.Get("/", server.GetPost)
				postRouter.Get("/view", server.ViewPost)
			})
		})
		v1Router.Route("/blobs", func(blobsRouter chi.Router) {
			blobsRouter.Use(jwtauth.Verifier(tokenAuth))
			blobsRouter.Use(jwtauth.Authenticator)
			blobsRouter.Post("/", server.CreateBlob)
			blobsRouter.Get("/list", server.ListBlobs)
			blobsRouter.Route("/search/{query}", func(searchRouter chi.Router) {
				searchRouter.Use(server.BlobsCtx)
				searchRouter.Get("/", SearchBlobs)
			})
			blobsRouter.Route("/{blobId}", func(blobRouter chi.Router) {
				blobRouter.Use(server.BlobCtx)
				blobRouter.Get("/", GetBlob)
				blobRouter.Delete("/", server.DeleteBlob)
			})
		})
		v1Router.Route("/users", func(usersRouter chi.Router) {
			usersRouter.Post("/signup", server.SignUp)
			usersRouter.Post("/authenticate", server.Authenticate)
			usersRouter.Route("/search/{query}", func(searchUserRouter chi.Router) {
				searchUserRouter.Get("/", server.SearchUser)
			})
			usersRouter.Route("/profile", func(profileRouter chi.Router) {
				profileRouter.Use(jwtauth.Verifier(tokenAuth))
				profileRouter.Use(jwtauth.Authenticator)
				profileRouter.Get("/", server.UserProfile)

			})
		})
	})
	return server
}

type BlobRequest struct {
	*types.Blob
}

type PostRequest struct {
	*types.Post
}

func (b *PostRequest) Bind(r *http.Request) error {
	if b.Post == nil {
		return errors.New("missing required Post fields")
	}
	return nil
}

func (b *BlobRequest) Bind(r *http.Request) error {
	if b.Blob == nil {
		return errors.New("missing required Blob fields")
	}
	return nil
}

type StatsRequest struct {
	*types.AdminRequest
}

func (b *StatsRequest) Bind(r *http.Request) error {
	if b.AdminPassword == "" {
		return errors.New("missing required auth fields")
	}
	return nil
}

type BasicAuthRequest struct {
	*types.BasicAuthCredentials
}

func (b *BasicAuthRequest) Bind(r *http.Request) error {
	if b.UserEmail == "" {
		return errors.New("missing required auth fields")
	}
	return nil
}

func (s *Server) SignUp(w http.ResponseWriter, r *http.Request) {
	data := &BasicAuthRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	user, err := s.userProvider.SignUp(data.UserEmail, data.UserPassword, data.DisplayName)
	if user != nil && err == nil {
		resp := &types.UserResponse{
			Id:          user.Id,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}
	} else {
		render.Render(w, r, ErrNotFound)
		return
	}
}

func (s *Server) SearchUser(w http.ResponseWriter, r *http.Request) {
	if query := chi.URLParam(r, "query"); query != "" {
		log.Printf("Searching for user: %s\n", query)
		user, err := s.userProvider.GetByEmail(query)
		if err != nil || user == nil {
			// Try display name
			user, err = s.userProvider.GetByDisplayName(query)
		}
		if err != nil || user == nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		resp := &types.UserResponse{
			Id:          user.Id,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		}
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}
	} else {
		log.Printf("No user supplied: %s\n", query)
		render.Render(w, r, ErrNotFound)
		return
	}
}

func (s *Server) UserProfile(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId, _ := claims["user_id"].(string)
	user, err := s.userProvider.GetById(userId)
	if user != nil && err == nil {
		resp := &types.UserResponse{
			Id:          user.Id,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}
	} else {
		render.Render(w, r, ErrNotFound)
		return
	}
}

func (s *Server) Authenticate(w http.ResponseWriter, r *http.Request) {
	data := &BasicAuthRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	user, err := s.userProvider.Authenticate(data.UserEmail, data.UserPassword)
	if user != nil && err == nil {
		_, tokenString, _ := s.jwtAuth.Encode(map[string]interface{}{"user_id": user.Id})
		resp := &types.AuthTokens{
			AccessToken:  tokenString,
			RefreshToken: "",
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}
	} else {
		render.Render(w, r, ErrNotFound)
		return
	}
}

func (s *Server) ListBlobs(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId, _ := claims["user_id"].(string)
	data := s.dataProvider.ListBlobs(userId)
	resp := &types.BlobList{RootBlobs: data}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}
}

func (s *Server) ListPosts(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId, _ := claims["user_id"].(string)
	if userId == "" {
		render.Render(w, r, ErrNotFound)
		return
	}
	data := s.dataProvider.ListPosts(userId)
	resp := &types.PostResponse{Posts: data}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	data := &PostRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	post := data.Post
	if post.Id == "" {
		post.Id = uuid.New().String()
	}

	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	userId, ok := claims["user_id"].(string)
	if !ok {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	post.OwnerId = userId

	err = s.dataProvider.PutPost(post)
	if err != nil {
		log.Printf("PutPost error: %s\n", err.Error())
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else {
		log.Printf("PutPost created\n")
		render.Status(r, http.StatusCreated)
		json.NewEncoder(w).Encode(&types.PostResponse{Posts: []*types.Post{post}})
	}
}

func (s *Server) CreateBlob(w http.ResponseWriter, r *http.Request) {
	data := &BlobRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	blob := data.Blob
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId, _ := claims["user_id"].(string)
	blob.OwnerId = userId
	log.Printf("Creating blob: %v\n", blob)
	if blob.Size() > 10000000 {
		render.Render(w, r, ErrInvalidRequest(errors.New("MaxFileSizeExceeded")))
		return
	}
	putErr := s.dataProvider.PutBlob(blob)
	if putErr != nil {
		render.Status(r, http.StatusConflict)
		render.Render(w, r, ErrNameConflict)
	} else {
		render.Status(r, http.StatusCreated)
		render.Render(w, r, &BlobResponse{[]*types.Blob{blob}})
	}
}

func (s *Server) DeleteBlob(w http.ResponseWriter, r *http.Request) {
	blob := r.Context().Value("blob").(*types.Blob)
	s.dataProvider.DeleteBlob(blob.Id, blob.OwnerId)
	if err := render.Render(w, r, &BlobResponse{[]*types.Blob{blob}}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func (s *Server) DeletePost(w http.ResponseWriter, r *http.Request) {
	post := r.Context().Value("post").(*types.Post)
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId, _ := claims["user_id"].(string)
	if userId != post.OwnerId {
		// Only owner can delete
		render.Render(w, r, ErrUnauthorized)
		return
	}

	success := s.dataProvider.DeletePost(post.OwnerId, post.Id)
	if success {
		post = &types.Post{}
	}

	if err := render.Render(w, r, &types.PostResponse{Posts: []*types.Post{post}}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func (s *Server) SplashPage(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(s.staticData.splashHtml))
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func (s *Server) ViewPost(w http.ResponseWriter, r *http.Request) {
	post := r.Context().Value("post").(*types.Post)
	relBlob := s.dataProvider.GetBlob(post.BlobId, post.OwnerId) // this has been shared to me so this is ok
	log.Printf("Got blob for post (view): %v\n", relBlob)
	if relBlob == nil {
		render.Render(w, r, ErrNotFound)
		return
	}
	_, err := w.Write([]byte(s.staticData.shareHtml))
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func (s *Server) GetStats(w http.ResponseWriter, r *http.Request) {
	data := &StatsRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if data.AdminPassword != s.adminPassword {
		render.Render(w, r, ErrUnauthorized)
		return
	}
	ct, users, err := s.userProvider.GetStats()
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return

	}
	var resps []types.UserResponse
	for _, user := range users {
		resp := types.UserResponse{
			Id:          user.Id,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		}
		resps = append(resps, resp)
	}
	response := types.AdminResponse{
		NumUsers:  ct,
		SomeUsers: resps,
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) GetVersion(w http.ResponseWriter, r *http.Request) {
	response := types.VersionResponse{
		ServerVersion:       types.ServerVersion,
		LatestClientVersion: types.ClientVersion,
		MinClientVersion:    types.ClientVersion,
		UpgradeScriptUrl:    "https://www.cubbycli.com/static/install.sh",
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) GetPost(w http.ResponseWriter, r *http.Request) {
	post := r.Context().Value("post").(*types.Post)
	relBlob := s.dataProvider.GetBlob(post.BlobId, post.OwnerId) // this has been shared to me so this is ok
	if relBlob == nil {
		render.Render(w, r, ErrNotFound)
		return
	}
	var body string
	var encBody []byte
	if relBlob.IsEncryptedAndEmpty() {
		encBody = relBlob.EncryptedBody().Data
	} else {
		body = relBlob.Data
	}

	if err := render.Render(w, r, &types.PostResponse{Blobs: []*types.Blob{relBlob}, Posts: []*types.Post{post}, Body: body, EncryptedBody: encBody}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func GetBlob(w http.ResponseWriter, r *http.Request) {
	blob := r.Context().Value("blob").(*types.Blob)
	if err := render.Render(w, r, &BlobResponse{[]*types.Blob{blob}}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func SearchBlobs(w http.ResponseWriter, r *http.Request) {
	blobs := r.Context().Value("blobs").([]*types.Blob)
	if err := render.Render(w, r, &BlobResponse{blobs}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

type BlobResponse struct {
	Blobs []*types.Blob `json:"blobs"`
}

func (b *BlobResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) userIdFromCtx(ctx context.Context) string {
	_, claims, _ := jwtauth.FromContext(ctx)
	if x, ok := claims["user_id"]; ok {
		if s, ok := x.(string); ok {
			return s
		}
	}
	return ""
}

func (s *Server) BlobsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var blobs []*types.Blob
		userId := s.userIdFromCtx(r.Context())
		if userId == "" {
			render.Render(w, r, ErrUnauthorized)
			return
		}
		if queryStr := chi.URLParam(r, "query"); queryStr != "" {
			blobs = s.dataProvider.QueryBlobs(queryStr, userId)
			for _, b := range blobs {
				if b.OwnerId != userId {
					render.Render(w, r, ErrNotFound)
					return
				}
			}
			if blobs == nil {
				render.Render(w, r, ErrNotFound)
				return
			}
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}

		if blobs == nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "blobs", blobs)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) PostCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var post *types.Post

		if postId := chi.URLParam(r, "postId"); postId != "" {
			if ownerName := chi.URLParam(r, "ownerName"); ownerName != "" {
				ownerUser, err := s.userProvider.GetByDisplayName(ownerName)
				if err != nil {
					render.Render(w, r, ErrNotFound)
					return
				}
				ownerId := ownerUser.Id
				userId := s.userIdFromCtx(r.Context())
				// `userId` will be empty if no user ID supplied
				post = s.dataProvider.GetPost(ownerId, postId)
				if post == nil {
					render.Render(w, r, ErrNotFound)
					return
				}
				hasPerm := post.HasPermission(userId) || post.OwnerId == userId
				if !hasPerm {
					log.Printf("User not authorized: %s\n%v\n", userId, post)
					render.Render(w, r, ErrUnauthorized)
					return
				}
			} else {
				render.Render(w, r, ErrNotFound)
				return
			}
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "post", post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) BlobCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var blob *types.Blob
		userId := s.userIdFromCtx(r.Context())
		if userId == "" {
			render.Render(w, r, ErrNotFound)
			return
		}
		if blobId := chi.URLParam(r, "blobId"); blobId != "" {
			blob = s.dataProvider.GetBlob(blobId, userId)
			if blob == nil {
				blob = s.dataProvider.GetBlobByPath(blobId, userId)
			}
			log.Printf("Returning blob %s for requesting user %s...\n", blobId, userId)
			if blob == nil || blob.OwnerId != userId {
				render.Render(w, r, ErrNotFound)
				return
			}
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}

		if blob == nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "blob", blob)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}
var ErrNameConflict = &ErrResponse{HTTPStatusCode: 409, StatusText: "Name conflict prevents insert."}
var ErrUnauthorized = &ErrResponse{HTTPStatusCode: 401, StatusText: "User not authorized."}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func (s *Server) Serve() {
	log.Printf("Serving on %d...\n", s.portNum)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.portNum), s.router)
	if err != nil {
		fmt.Printf("HTTP listen error: %s\n", err.Error())
	}
}

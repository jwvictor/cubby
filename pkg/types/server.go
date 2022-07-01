package types

type AdminRequest struct {
	AdminPassword string `json:"admin_password"`
}

type AdminResponse struct {
	NumUsers  int            `json:"num_users"`
	SomeUsers []UserResponse `json:"some_users"`
}

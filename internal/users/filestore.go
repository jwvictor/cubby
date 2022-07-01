package users

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"os"
	"sync"
)

type UserRecord struct {
	Id           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"pass"`
	DisplayName  string `json:"display_name"`
}

type UsersFileStore struct {
	outputFilename string
	users          map[string]*UserRecord
	lock           *sync.RWMutex
}

func NewUsersFileStore(filename string) (*UsersFileStore, error) {
	jsonFile, err := os.Open(filename)
	defer jsonFile.Close()
	m := map[string]*UserRecord{}
	if err == nil {
		bytes, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal(bytes, &m)
	}
	return &UsersFileStore{
		outputFilename: filename,
		users:          m,
		lock:           &sync.RWMutex{},
	}, nil
}

func (p *UsersFileStore) _dumpUnsafe() error {
	bs, err := json.Marshal(p.users)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(p.outputFilename, bs, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (p *UsersFileStore) _checkDisplayNameUnsafe(name string) bool {
	for _, user := range p.users {
		if user.DisplayName == name {
			return false
		}
	}
	return true
}

func (p *UsersFileStore) GetStats() (int, []User, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	var usrs []User
	ct := 0
	for _, user := range p.users {
		u := user.ToUser()
		usrs = append(usrs, *u)
		ct += 1
		if ct >= 100 {
			break
		}
	}
	return len(p.users), usrs, nil
}

func (p *UsersFileStore) GetByEmail(email string) (*User, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	for _, user := range p.users {
		if user.Email == email {
			return user.ToUser(), nil
		}
	}
	return nil, errors.New("NotFound")
}

func (p *UsersFileStore) GetByDisplayName(displayName string) (*User, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	for _, user := range p.users {
		if user.DisplayName == displayName {
			return user.ToUser(), nil
		}
	}
	return nil, errors.New("NotFound")
}

func (p *UsersFileStore) SignUp(userEmail, userPass, displayName string) (*User, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.users[userEmail]; ok {
		return nil, errors.New("UserAlreadyExists")
	} else {
		if !p._checkDisplayNameUnsafe(displayName) {
			return nil, errors.New("DisplayNameAlreadyExists")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(userPass), bcrypt.MinCost)
		if err != nil {
			return nil, err
		}
		hashHex := hex.EncodeToString(hash)
		p.users[userEmail] = &UserRecord{
			Id:           uuid.New().String(),
			Email:        userEmail,
			PasswordHash: hashHex,
			DisplayName:  displayName,
		}
		p._dumpUnsafe()
		return &User{
			Id:          p.users[userEmail].Id,
			Email:       userEmail,
			DisplayName: displayName,
		}, nil
	}
}

func (r *UserRecord) ToUser() *User {
	return &User{Id: r.Id, Email: r.Email, DisplayName: r.DisplayName}
}

func (p *UsersFileStore) GetById(id string) (*User, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	for _, user := range p.users {
		if user.Id == id {
			return user.ToUser(), nil
		}
	}
	return nil, errors.New("NotFound")
}

func (p *UsersFileStore) Authenticate(userEmail string, userPass string) (*User, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if user, ok := p.users[userEmail]; ok {
		passHash, err := hex.DecodeString(user.PasswordHash)
		if err != nil {
			return nil, err
		}
		err = bcrypt.CompareHashAndPassword(passHash, []byte(userPass))
		if err != nil {
			return nil, errors.New("BadPassword")
		} else {
			return &User{
				Id:          user.Id,
				Email:       user.Email,
				DisplayName: user.DisplayName,
			}, nil
		}
	} else {
		return nil, errors.New("NoSuchUser")
	}
}

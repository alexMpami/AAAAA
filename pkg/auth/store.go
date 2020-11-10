package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/finitum/AAAAA/pkg/models"
	"github.com/finitum/AAAAA/pkg/store"
	"github.com/go-chi/jwtauth"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type StoreAuth struct {
	db  store.Store
	jwt *jwtauth.JWTAuth
}

func NewStoreAuth(db store.Store, jwtkey string) *StoreAuth {
	return &StoreAuth{
		db:  db,
		jwt: jwtauth.New(jwt.SigningMethodHS384.Name, []byte(jwtkey), nil),
	}
}

func (s *StoreAuth) Verify(token string) (Claims, bool) {
	dec, err := s.jwt.Decode(token)
	if err != nil {
		return Claims{}, false
	}

	if !dec.Valid {
		return Claims{}, false
	}

	claims := dec.Claims.(*jwt.StandardClaims)

	return Claims{
		Username: claims.Subject,
		RawToken: token,
	}, dec.Valid
}

func (s *StoreAuth) Update(user *models.User, _ string) error {
	_, err := s.db.GetUser(user.Username)
	if err == store.ErrNotExists {
		return errors.New("user doesn't exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "bcrypt generate")
	}
	user.Password = string(hash)

	if err := s.db.AddUser(user); err != nil {
		return errors.Wrap(err, "adding user to db")
	}

	return nil
}


func (s *StoreAuth) Login(user *models.User) (string, error) {
	dbUser, err := s.db.GetUser(user.Username)
	if err != nil {
		return "", errors.Wrap(err, "user not in database")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		return "", errors.Wrap(err, "password wrong or invalid")
	}

	_, tokenString, err := s.jwt.Encode(jwt.StandardClaims{Subject: dbUser.Username, Audience: "user"})
	if err != nil {
		return "", errors.Wrap(err, "couldn't encode jwt token")

	}

	return tokenString, nil
}

func (s StoreAuth) Register(user *models.User) error {
	_, err := s.db.GetUser(user.Username)
	if err != store.ErrNotExists {
		return errors.New("user exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "bcrypt generate")
	}

	user.Password = string(hash)

	if err := s.db.AddUser(user); err != nil {
		return errors.Wrap(err, "adding user to db")
	}

	return nil
}

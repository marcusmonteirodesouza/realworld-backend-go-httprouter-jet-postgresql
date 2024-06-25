package services

import (
	"context"
	"database/sql"
	"fmt"
	"net/mail"
	"time"

	"cloud.google.com/go/logging"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/golang-jwt/jwt/v5"
	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/.gen/realworld/public/model"
	. "github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/.gen/realworld/public/table"
)

type UsersService struct {
	db     *sql.DB
	jwt    UsersServiceJWT
	logger *logging.Logger
}

type UsersServiceJWT struct {
	iss             string
	key             []byte
	validForSeconds int
}

func NewUsersService(db *sql.DB, jwt UsersServiceJWT, logger *logging.Logger) *UsersService {
	return &UsersService{
		db:     db,
		jwt:    jwt,
		logger: logger,
	}
}

func NewUsersServiceJWT(iss string, key []byte, validForSeconds int) *UsersServiceJWT {
	return &UsersServiceJWT{
		iss:             iss,
		key:             key,
		validForSeconds: validForSeconds,
	}
}

type RegisterUser struct {
	Email    string
	Username string
	Password string
}

func NewRegisterUser(email string, username string, password string) RegisterUser {
	return RegisterUser{
		Email:    email,
		Username: username,
		Password: password,
	}
}

func (usersService *UsersService) RegisterUser(ctx context.Context, registerUser RegisterUser) (*model.Users, error) {
	usersService.logger.StandardLogger(logging.Info).Printf("Registering user. Email: %s, Username: %s", registerUser.Email, registerUser.Username)

	err := usersService.validateEmail(ctx, registerUser.Email)
	if err != nil {
		return nil, err
	}

	err = usersService.validateUsername(ctx, registerUser.Username)
	if err != nil {
		return nil, err
	}

	passwordHash, err := usersService.hashPassword(ctx, registerUser.Password)
	if err != nil {
		return nil, err
	}

	user := model.Users{
		Email:        registerUser.Email,
		Username:     registerUser.Username,
		PasswordHash: *passwordHash,
	}

	insertStmt := Users.INSERT(Users.Email, Users.Username, Users.PasswordHash).MODEL(user).RETURNING(Users.AllColumns)

	err = insertStmt.QueryContext(ctx, usersService.db, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (usersService *UsersService) GetToken(user *model.Users) (*string, error) {
	now := time.Now()

	exp := now.Add(time.Second * time.Duration(usersService.jwt.validForSeconds))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": exp.Unix(),
		"iat": now.Unix(),
		"iss": usersService.jwt.iss,
		"sub": user.ID,
	})

	jwt, err := token.SignedString(usersService.jwt.key)

	if err != nil {
		return nil, err
	}

	return &jwt, nil
}

func (usersService *UsersService) hashPassword(ctx context.Context, password string) (*string, error) {
	maxPasswordLength := 72

	if len(password) > maxPasswordLength {
		return nil, &InvalidArgumentError{msg: fmt.Sprintf("Password length must be less than or equal to %d", maxPasswordLength)}
	}

	var dest struct {
		PasswordHash string
	}

	hashPasswordStmt := SELECT(StringExp(Func("crypt", String(password), Func("gen_salt", String("bf")))))

	err := hashPasswordStmt.QueryContext(ctx, usersService.db, &dest)
	if err != nil {
		return nil, err
	}

	return &dest.PasswordHash, nil
}

func (usersService *UsersService) validateEmail(ctx context.Context, email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return err
	}

	var dest struct {
		EmailExists bool
	}

	emailExistsStmt := SELECT(EXISTS(Users.SELECT(Users.ID).WHERE(Users.Email.EQ(String(email)))).AS("email_exists"))
	err = emailExistsStmt.QueryContext(ctx, usersService.db, &dest)
	if err != nil {
		return err
	}

	if dest.EmailExists {
		return &AlreadyExistsError{msg: "Email is taken"}
	}

	return nil
}

func (usersService *UsersService) validateUsername(ctx context.Context, username string) error {
	var dest struct {
		UsernameExists bool
	}

	usernameExistsStmt := SELECT(EXISTS(Users.SELECT(Users.ID).WHERE(Users.Username.EQ(String(username)))).AS("username_exists"))
	err := usernameExistsStmt.QueryContext(ctx, usersService.db, &dest)
	if err != nil {
		return err
	}

	if dest.UsernameExists {
		return &AlreadyExistsError{msg: "Username is taken"}
	}

	return nil
}

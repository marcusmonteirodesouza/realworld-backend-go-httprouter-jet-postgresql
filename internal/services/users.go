package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"time"

	"cloud.google.com/go/logging"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/.gen/realworld/public/model"
	. "github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/.gen/realworld/public/table"
)

type UsersService struct {
	db     *sql.DB
	jwt    *UsersServiceJWT
	logger *logging.Logger
}

type UsersServiceJWT struct {
	iss             string
	key             []byte
	parser          jwt.Parser
	signinMethod    jwt.SigningMethod
	validForSeconds int
}

func NewUsersService(db *sql.DB, jwt *UsersServiceJWT, logger *logging.Logger) UsersService {
	return UsersService{
		db:     db,
		jwt:    jwt,
		logger: logger,
	}
}

func NewUsersServiceJWT(iss string, key []byte, validForSeconds int) UsersServiceJWT {
	return UsersServiceJWT{
		iss:             iss,
		key:             key,
		parser:          *jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithIssuer(iss)),
		signinMethod:    jwt.SigningMethodHS256,
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

type ListUsers struct {
	UserIDs *[]uuid.UUID
}

type UpdateUser struct {
	Email    *string
	Username *string
	Password *string
	Bio      *string
	Image    *string
}

func (usersService *UsersService) RegisterUser(ctx context.Context, registerUser RegisterUser) (*model.Users, error) {
	type registerUserForLogging struct {
		Email    string
		Username string
	}

	if registerUserForLoggingBytes, err := json.Marshal(registerUserForLogging{
		Email:    registerUser.Email,
		Username: registerUser.Username,
	}); err == nil {
		usersService.logger.StandardLogger(logging.Info).Printf("Registering user. %s", string(registerUserForLoggingBytes))
	}

	err := usersService.validateEmail(ctx, registerUser.Email)
	if err != nil {
		return nil, err
	}

	if err = usersService.validateUsername(ctx, registerUser.Username); err != nil {
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

	if err = insertStmt.QueryContext(ctx, usersService.db, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (usersService *UsersService) GetUserById(ctx context.Context, userId uuid.UUID) (*model.Users, error) {
	user := model.Users{}

	selectStmt := SELECT(Users.AllColumns).FROM(Users).WHERE(Users.ID.EQ(UUID(userId)))

	err := selectStmt.QueryContext(ctx, usersService.db, &user)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, &NotFoundError{msg: fmt.Sprintf("User %s not found", userId)}
		}
		return nil, err
	}

	return &user, nil
}

func (usersService *UsersService) GetUserByEmail(ctx context.Context, email string) (*model.Users, error) {
	user := model.Users{}

	selectStmt := SELECT(Users.AllColumns).FROM(Users).WHERE(Users.Email.EQ(String(email)))

	err := selectStmt.QueryContext(ctx, usersService.db, &user)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, &NotFoundError{msg: fmt.Sprintf("Email %s not found", email)}
		}
		return nil, err
	}

	return &user, nil
}

func (usersService *UsersService) GetUserByUsername(ctx context.Context, username string) (*model.Users, error) {
	user := model.Users{}

	selectStmt := SELECT(Users.AllColumns).FROM(Users).WHERE(Users.Username.EQ(String(username)))

	err := selectStmt.QueryContext(ctx, usersService.db, &user)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, &NotFoundError{msg: fmt.Sprintf("Username %s not found", username)}
		}
		return nil, err
	}

	return &user, nil
}

func (usersService *UsersService) GetUserByToken(ctx context.Context, token string) (*model.Users, error) {
	jwt, err := usersService.jwt.parser.Parse(token, func(t *jwt.Token) (interface{}, error) { return usersService.jwt.key, nil })
	if err != nil {
		return nil, err
	}

	userIdString, err := jwt.Claims.GetSubject()
	if err != nil {
		return nil, err
	}

	userId, err := uuid.Parse(userIdString)
	if err != nil {
		return nil, err
	}

	user, err := usersService.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (usersService *UsersService) ListUsers(ctx context.Context, listUsers ListUsers) (*[]model.Users, error) {
	condition := Bool(true)

	if listUsers.UserIDs != nil {
		if len(*listUsers.UserIDs) > 0 {
			var sqlUserIds []Expression

			for _, userId := range *listUsers.UserIDs {
				sqlUserIds = append(sqlUserIds, UUID(userId))
			}

			condition = condition.AND(Users.ID.IN(sqlUserIds...))
		} else {
			condition = condition.AND(Bool(false))
		}
	}

	var users []model.Users

	listUsersStmt := SELECT(Users.AllColumns).FROM(Users).WHERE(condition)

	err := listUsersStmt.QueryContext(ctx, usersService.db, &users)
	if err != nil {
		return nil, err
	}

	return &users, nil
}

func (usersService *UsersService) UpdateUser(ctx context.Context, userId uuid.UUID, updateUser UpdateUser) (*model.Users, error) {
	type updateUserForLogging struct {
		Email            *string
		Username         *string
		UpdatingPassword bool
		Bio              *string
		Image            *string
	}

	if updateUserForLoggingBytes, err := json.Marshal(updateUserForLogging{
		Email:            updateUser.Email,
		Username:         updateUser.Username,
		UpdatingPassword: updateUser.Password != nil,
		Bio:              updateUser.Bio,
		Image:            updateUser.Image,
	}); err == nil {
		usersService.logger.StandardLogger(logging.Info).Printf("Updating user %s. %s", userId, string(updateUserForLoggingBytes))
	}

	user, err := usersService.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}

	if updateUser.Email != nil && *updateUser.Email != user.Email {
		if err = usersService.validateEmail(ctx, *updateUser.Email); err != nil {
			return nil, err
		}

		user.Email = *updateUser.Email
	}

	if updateUser.Username != nil && *updateUser.Username != user.Username {
		if err = usersService.validateUsername(ctx, *updateUser.Username); err != nil {
			return nil, err
		}

		user.Username = *updateUser.Username
	}

	if updateUser.Password != nil {
		passwordHash, err := usersService.hashPassword(ctx, *updateUser.Password)
		if err != nil {
			return nil, err
		}

		user.PasswordHash = *passwordHash
	}

	if updateUser.Bio != user.Bio {
		user.Bio = updateUser.Bio
	}

	if updateUser.Image != user.Image {
		if err = usersService.validateImage(*updateUser.Image); err != nil {
			return nil, err
		}

		user.Image = updateUser.Image
	}

	now := time.Now().UTC()

	user.UpdatedAt = &now

	updateStmt := Users.UPDATE(Users.Email, Users.Username, Users.PasswordHash, Users.Bio, Users.Image, Users.UpdatedAt).MODEL(user).WHERE(Users.ID.EQ(UUID(user.ID))).RETURNING(Users.AllColumns)

	if err = updateStmt.QueryContext(ctx, usersService.db, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (usersService *UsersService) CheckPassword(ctx context.Context, userId uuid.UUID, password string) (*bool, error) {
	var dest struct {
		IsCorrectPassword bool
	}

	checkPasswordStmt := SELECT(Users.PasswordHash.EQ(StringExp(Func("crypt", String(password), Users.PasswordHash))).AS("is_correct_password")).FROM(Users).WHERE(Users.ID.EQ(UUID(userId)))

	err := checkPasswordStmt.QueryContext(ctx, usersService.db, &dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, &NotFoundError{msg: fmt.Sprintf("User %s not found", userId)}
		}
		return nil, err
	}

	return &dest.IsCorrectPassword, nil
}

func (usersService *UsersService) GetToken(user *model.Users) (*string, error) {
	now := time.Now()

	exp := now.Add(time.Second * time.Duration(usersService.jwt.validForSeconds))

	token := jwt.NewWithClaims(usersService.jwt.signinMethod, jwt.MapClaims{
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

	hashPasswordStmt := SELECT(StringExp(Func("crypt", String(password), Func("gen_salt", String("bf")))).AS("password_hash"))

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

	var emailExistsDest struct {
		EmailExists bool
	}

	emailExistsStmt := SELECT(EXISTS(Users.SELECT(Users.ID).WHERE(Users.Email.EQ(String(email)))).AS("email_exists"))

	if err = emailExistsStmt.QueryContext(ctx, usersService.db, &emailExistsDest); err != nil {
		return err
	}

	if emailExistsDest.EmailExists {
		return &AlreadyExistsError{msg: "Email is taken"}
	}

	return nil
}

func (usersService *UsersService) validateUsername(ctx context.Context, username string) error {
	var usernameExistsDest struct {
		UsernameExists bool
	}

	usernameExistsStmt := SELECT(EXISTS(Users.SELECT(Users.ID).WHERE(Users.Username.EQ(String(username)))).AS("username_exists"))

	err := usernameExistsStmt.QueryContext(ctx, usersService.db, &usernameExistsDest)
	if err != nil {
		return err
	}

	if usernameExistsDest.UsernameExists {
		return &AlreadyExistsError{msg: "Username is taken"}
	}

	return nil
}

func (usersService *UsersService) validateImage(image string) error {
	_, err := url.ParseRequestURI(image)
	if err != nil {
		return &InvalidArgumentError{msg: fmt.Sprintf("Invalid image URL %s", image)}
	}

	return nil
}

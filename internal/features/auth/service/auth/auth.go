package service_auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/poponyas/AuthService/internal/core/domain"
	core_logger "github.com/poponyas/AuthService/internal/core/logger"
	service_storage "github.com/poponyas/AuthService/internal/features/auth/service/storage"
	core_jwt "github.com/poponyas/AuthService/pkg/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppId       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
)

type Auth struct {
	logger      *core_logger.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenttl    time.Duration
	signingKey  []byte
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (domain.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (domain.App, error)
}

func New(
	logger *core_logger.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	ttl time.Duration,
	signingKey []byte,
) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		logger:      logger,
		appProvider: appProvider,
		tokenttl:    ttl,
		signingKey:  signingKey,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "auth.Login"
	log := a.logger.With(
		zap.String("op", op),
	)

	log.Info("attemping to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, service_storage.ErrUserNotFound) {
			a.logger.Warn("user not found", zap.Error(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.logger.Error("failed to get user", zap.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.logger.Info("invalid credentials", zap.Error(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, service_storage.ErrAppNotFound) {
			return "", fmt.Errorf("%s: %w", op, ErrInvalidAppId)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := core_jwt.NewToken(user, app, a.tokenttl, a.signingKey)
	if err != nil {
		a.logger.Error("failed to generate token", zap.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	pass string,
) (int64, error) {
	const op = "auth.RegisterNewUser"
	log := a.logger.With(
		zap.String("op", op),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", zap.Error(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, service_storage.ErrUserExists) {
			log.Warn("user already exists", zap.Error(err))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", zap.Error(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return id, nil
}

func (a *Auth) IsAdmin(
	ctx context.Context,
	userID int64,
) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.logger.With(
		zap.String("op", op),
		zap.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, service_storage.ErrUserNotFound) {
			log.Warn("user not found", zap.Error(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppId)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", zap.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

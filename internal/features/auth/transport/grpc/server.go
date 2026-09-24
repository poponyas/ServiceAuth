package transport_grpc_auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
	"net/mail"
	"strconv"
	"strings"

	grpc_auth "github.com/poponyas/AuthService/gen/grpc/auth"
	service_auth "github.com/poponyas/AuthService/internal/features/auth/service/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	grpc_auth.UnimplementedAuthServer
	auth Auth
	key  []byte
}

// регистрируем сервер для обработчиков
func Register(gRPC *grpc.Server, auth Auth, key []byte) {
	grpc_auth.RegisterAuthServer(gRPC, &serverAPI{auth: auth, key: key})
}

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)

	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)

	IsAdmin(
		ctx context.Context,
		userID int64,
	) (bool, error)
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *grpc_auth.LoginRequest,
) (*grpc_auth.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, service_auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		if errors.Is(err, service_auth.ErrInvalidAppId) {
			return nil, status.Error(codes.InvalidArgument, "invalid app")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &grpc_auth.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *grpc_auth.RegisterRequest,
) (*grpc_auth.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, service_auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &grpc_auth.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *grpc_auth.IsAdminRequest,
) (*grpc_auth.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get("authorization")) != 1 || !strings.HasPrefix(md.Get("authorization")[0], "Bearer ") {
		return nil, status.Error(codes.Unauthenticated, "Bearer token required")
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(strings.TrimPrefix(md.Get("authorization")[0], "Bearer "), claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("invalid algorithm")
		}
		return s.key, nil
	}, jwt.WithIssuer("service-auth"), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	sub, err := claims.GetSubject()
	if err != nil || sub != strconv.FormatInt(req.GetUserId(), 10) {
		return nil, status.Error(codes.PermissionDenied, "cannot inspect another user")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &grpc_auth.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func validateLogin(req *grpc_auth.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == 0 {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func validateRegister(req *grpc_auth.RegisterRequest) error {
	address, err := mail.ParseAddress(req.GetEmail())
	if err != nil || address.Address != req.GetEmail() || strings.ContainsAny(req.GetEmail(), " \t\n") || len(req.GetEmail()) > 254 {
		return status.Error(codes.InvalidArgument, "valid email is required")
	}

	if len(req.GetPassword()) < 12 || len(req.GetPassword()) > 72 {
		return status.Error(codes.InvalidArgument, "password must be 12-72 bytes")
	}

	return nil
}

func validateIsAdmin(req *grpc_auth.IsAdminRequest) error {
	if req.GetUserId() == 0 {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	return nil
}

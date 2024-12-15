package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"squad-maker/database"
	pb "squad-maker/generated/auth"
	"squad-maker/models"
	"squad-maker/utils/env"
	grpcUtils "squad-maker/utils/grpc"
	jwtUtils "squad-maker/utils/jwt"
	"squad-maker/utils/utfpr"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

var (
	sessionTokenMaxAge int32
)

func init() {
	var err error
	sessionTokenMaxAge, err = env.GetInt32("SESSION_TOKEN_MAX_AGE")
	if err != nil {
		fmt.Println("failed to parse session token max age from env")
	}
}

func (s *AuthServiceServer) CreateToken(ctx context.Context, req *pb.CreateTokenRequest) (*pb.TokenResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	user, err := s.verifyCredentials(ctx, dbCon, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	session := &models.Session{}

	// gera um novo token
	// https://golang.org/pkg/crypto/rand/#Read
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// https://gist.github.com/nicklaw5/9d2d76b04d345152364d9b8cb4b554e9#file-ramdom_sha1-go-L32
	hash := sha1.New()
	hash.Write(b)
	bs := hash.Sum(nil)

	token := fmt.Sprintf("%x", bs)

	session = &models.Session{
		UserId:      user.Id,
		Token:       token,
		LastRefresh: time.Now(),
		DoNotExpire: req.KeepConnected,
	}
	r := dbCon.Create(session)
	if r.Error != nil {
		return nil, status.Error(codes.Internal, r.Error.Error())
	}

	tokenString, expiresIn, err := jwtUtils.GenerateAuthToken(session.UserId, user.Type, session.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.TokenResponse{
		AccessToken:  tokenString,
		RefreshToken: session.Token,
		TokenType:    "bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *AuthServiceServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.TokenResponse, error) {
	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	session := &models.Session{}

	// geralmente isso vem de um cookie, não de um metadata
	// mas pra não aumentar a complexidade, deixei no metadata mesmo
	// se fosse fazer cookie com todos os padrões de segurança que utilizo,
	// seria necessário um proxy para acessar a API a partir do frontend
	// e seria necessário SSL
	sessionToken := grpcUtils.GetCurrentSessionTokenFromMetadata(ctx)
	if sessionToken == "" {
		return nil, status.Error(codes.InvalidArgument, "session token not found")
	}

	r := dbCon.Where(models.Session{
		Token: sessionToken,
	}).First(session)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.InvalidArgument, "invalid session token")
		}
		return nil, status.Error(codes.Internal, r.Error.Error())
	}

	if !session.DoNotExpire && time.Since(session.LastRefresh) > time.Duration(sessionTokenMaxAge)*time.Second {
		return nil, status.Error(codes.InvalidArgument, "invalid session token")
	}

	// verifica se o user existe
	user := &models.User{}
	r = dbCon.First(user, session.UserId)
	if r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.Internal, r.Error.Error())
	}

	if user.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid session token")
	}

	// gera o novo token
	tokenString, expiresIn, err := jwtUtils.GenerateAuthToken(session.UserId, user.Type, session.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// atualiza o último uso da sessão
	r = dbCon.Model(session).Updates(models.Session{
		LastRefresh: time.Now(),
	})
	if r.Error != nil {
		return nil, status.Error(codes.Internal, r.Error.Error())
	}

	return &pb.TokenResponse{
		AccessToken:  tokenString,
		RefreshToken: session.Token,
		TokenType:    "bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

// TODO verificar, pois se eu não me engano, não tá funcionando (nem chega aqui com os dados necessários)
func (s *AuthServiceServer) InvalidateToken(ctx context.Context, req *pb.InvalidateTokenRequest) (*pb.InvalidateTokenResponse, error) {
	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	sessionId := grpcUtils.GetCurrentSessionIdFromMetadata(ctx)

	if sessionId > 0 {
		r := dbCon.Where(models.Session{
			BaseModelWithSoftDelete: database.BaseModelWithSoftDelete{
				BaseModel: database.BaseModel{
					Id: sessionId,
				},
			},
		}).Delete(&models.Session{})
		if r.Error != nil {
			return nil, status.Error(codes.Internal, r.Error.Error())
		}
	}

	return &pb.InvalidateTokenResponse{}, nil
}

func (s *AuthServiceServer) verifyCredentials(ctx context.Context, dbCon *gorm.DB, username string, password string) (*models.User, error) {
	// chama a API pra verificar as credenciais
	profile, err := utfpr.GetProfileFromLogin(ctx, username, password)
	if err != nil {
		if errors.Is(err, utfpr.ErrInvalidCredentials) {
			return nil, status.Error(codes.NotFound, "no match for informed credentials")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !profile.Active {
		return nil, status.Error(codes.FailedPrecondition, "user is not active")
	}

	// verifica se já existe no banco
	// se não existir, cria
	user := &models.User{}
	r := dbCon.Where(models.User{UtfprUsername: strings.ToLower(username)}).First(user)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			// cria o usuário
			var t pb.UserType
			if profile.ProfileTypes.Contains(utfpr.ProfileType_Student) {
				t = pb.UserType_utStudent
			} else if profile.ProfileTypes.Contains(utfpr.ProfileType_Professor) {
				t = pb.UserType_utProfessor
			} else {
				return nil, status.Error(codes.FailedPrecondition, "user is not a student or a professor")
			}

			user = &models.User{
				UtfprUsername: strings.ToLower(username),
				Type:          t,
				Name:          profile.Name,
				Email:         profile.Email,
			}

			r = dbCon.Create(user)
			if r.Error != nil {
				return nil, status.Error(codes.Internal, r.Error.Error())
			}
		} else {
			return nil, status.Error(codes.Internal, r.Error.Error())
		}
	}

	return user, nil
}

func (s *AuthServiceServer) Me(ctx context.Context, req *pb.MeRequest) (*pb.MeResponse, error) {
	userId := grpcUtils.GetCurrentUserIdFromMetadata(ctx)

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	user := &models.User{}
	r := dbCon.First(user, userId)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, r.Error.Error())
	}

	return &pb.MeResponse{
		Id:    user.Id,
		Email: user.Email,
		Type:  user.Type,
		Name:  user.Name,
	}, nil
}

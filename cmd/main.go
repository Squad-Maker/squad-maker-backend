package main

import (
	_ "github.com/joho/godotenv/autoload"

	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"squad-maker/cmd/reflection"
	"squad-maker/database"
	pbAuth "squad-maker/generated/auth"
	pbSquad "squad-maker/generated/squad"
	"squad-maker/grpc/auth"
	"squad-maker/grpc/squad"
	"squad-maker/migrations"
	"squad-maker/models"
	grpcUtils "squad-maker/utils/grpc"
	jwtUtils "squad-maker/utils/jwt"
	"strconv"
	"strings"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func getNetListener(port uint) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	return lis
}

func authUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	ctx, err := authorize(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

// não vou fazer um interceptor de stream pois não vou precisar

func authorize(ctx context.Context, methodFullName string) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "unable to get metadata from context")
	}

	retFunc := func(err error) (context.Context, error) {
		if err == nil || grpcUtils.IsPublicRoute(methodFullName) {
			return metadata.NewIncomingContext(ctx, md), nil
		}
		return nil, err
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return retFunc(status.Error(codes.Internal, err.Error()))
	}

	// pega o header de autenticação
	authHeader, ok := md["authorization"]
	if !ok {
		return retFunc(status.Error(codes.Unauthenticated, "authorization header malformed or not provided"))
	}

	// extrai o token
	splitHeader := strings.Split(authHeader[len(authHeader)-1], "Bearer ")
	if len(splitHeader) != 2 {
		return retFunc(status.Error(codes.Unauthenticated, "authorization header malformed or not provided"))
	}

	// valida o token
	token, claims := jwtUtils.GetAuthTokenAndClaimsIfValid(splitHeader[1])
	if token == nil || claims == nil {
		return retFunc(status.Error(codes.Unauthenticated, "invalid token"))
	}

	// valida os dados do token
	// claims.Subject -> user id
	// claims.Type -> user type
	// claims.SessionId -> session id

	// carrega o usuário
	userId, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		fmt.Printf("failed to parse user id; claims: %+v\n", claims)
		return retFunc(status.Error(codes.Unauthenticated, "invalid token"))
	}

	user := &models.User{}
	r := dbCon.First(user, userId)
	if r.Error != nil {
		return retFunc(status.Error(codes.Unauthenticated, "invalid token"))
	}

	if user.Type != claims.UserType {
		return retFunc(status.Error(codes.Unauthenticated, "invalid token"))
	}

	// carrega a sessão
	session := &models.Session{}
	r = dbCon.First(session, claims.SessionId)
	if r.Error != nil {
		return retFunc(status.Error(codes.Unauthenticated, "invalid token"))
	}

	if session.UserId != user.Id {
		return retFunc(status.Error(codes.Unauthenticated, "invalid token"))
	}

	// verifica se o usuário pode acessar a rota
	err = grpcUtils.CheckRoutePermission(methodFullName, user.Type)
	if err != nil {
		return retFunc(err)
	}

	md.Set("session-id", strconv.FormatInt(session.Id, 10))
	md.Set("user-id", claims.Subject)
	md.Set("user-type", strconv.FormatInt(int64(user.Type), 10))

	return metadata.NewIncomingContext(ctx, md), nil
}

func main() {
	ctx := context.Background()
	authService := auth.NewAuthServiceServer()
	squadService := squad.NewSquadServiceServer()
	s := grpc.NewServer(grpc.UnaryInterceptor(authUnaryInterceptor))

	pbAuth.RegisterAuthServiceServer(s, authService)
	pbSquad.RegisterSquadServiceServer(s, squadService)
	reflection.MaybeEnableReflection(s)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			s.GracefulStop()

			<-ctx.Done()
		}
	}()

	err := migrations.RunMigrations(ctx)
	if err != nil {
		panic(err)
	}

	listenWeb := getNetListener(9080)
	listenGrpc := getNetListener(9090)

	wrappedServer := grpcweb.WrapServer(
		s,
		grpcweb.WithAllowedRequestHeaders([]string{"*"}),
		grpcweb.WithOriginFunc(func(origin string) bool {
			// ignore cors
			return true
		}),
	)

	go http.Serve(listenWeb, wrappedServer)

	s.Serve(listenGrpc)
}

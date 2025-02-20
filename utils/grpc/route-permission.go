package grpcUtils

import (
	pbAuth "squad-maker/generated/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	mapRoutesToUserType = map[string][]pbAuth.UserType{
		"/auth.AuthService/Me": {pbAuth.UserType_utProfessor, pbAuth.UserType_utStudent},

		"/squad.SquadService/ReadCompetenceLevel":     {pbAuth.UserType_utProfessor, pbAuth.UserType_utStudent},
		"/squad.SquadService/CreateCompetenceLevel":   {pbAuth.UserType_utProfessor},
		"/squad.SquadService/UpdateCompetenceLevel":   {pbAuth.UserType_utProfessor},
		"/squad.SquadService/DeleteCompetenceLevel":   {pbAuth.UserType_utProfessor},
		"/squad.SquadService/ReadAllCompetenceLevels": {pbAuth.UserType_utProfessor, pbAuth.UserType_utStudent},

		"/squad.SquadService/ReadPosition":     {pbAuth.UserType_utProfessor, pbAuth.UserType_utStudent},
		"/squad.SquadService/CreatePosition":   {pbAuth.UserType_utProfessor},
		"/squad.SquadService/UpdatePosition":   {pbAuth.UserType_utProfessor},
		"/squad.SquadService/DeletePosition":   {pbAuth.UserType_utProfessor},
		"/squad.SquadService/ReadAllPositions": {pbAuth.UserType_utProfessor, pbAuth.UserType_utStudent},

		"/squad.SquadService/ReadProject":     {pbAuth.UserType_utProfessor, pbAuth.UserType_utStudent},
		"/squad.SquadService/CreateProject":   {pbAuth.UserType_utProfessor},
		"/squad.SquadService/UpdateProject":   {pbAuth.UserType_utProfessor},
		"/squad.SquadService/DeleteProject":   {pbAuth.UserType_utProfessor},
		"/squad.SquadService/ReadAllProjects": {pbAuth.UserType_utProfessor, pbAuth.UserType_utStudent},

		"/squad.SquadService/GetStudentSubjectData":    {pbAuth.UserType_utStudent},
		"/squad.SquadService/UpdateStudentSubjectData": {pbAuth.UserType_utStudent},
		"/squad.SquadService/ReadAllStudentsInSubject": {pbAuth.UserType_utProfessor},
		"/squad.SquadService/GenerateAllTeams":         {pbAuth.UserType_utProfessor},
		"/squad.SquadService/GenerateTeam":             {pbAuth.UserType_utProfessor},
		"/squad.SquadService/AddStudentToTeam":         {pbAuth.UserType_utProfessor},
		"/squad.SquadService/RemoveStudentFromTeam":    {pbAuth.UserType_utProfessor},
		"/squad.SquadService/RequestTeamRevaluation":   {pbAuth.UserType_utStudent},
	}

	mapIgnorePermissions = map[string]struct{}{
		"/auth.AuthService/CreateToken":     {},
		"/auth.AuthService/RefreshToken":    {},
		"/auth.AuthService/InvalidateToken": {},
	}
)

func CheckRoutePermission(methodFullName string, userType pbAuth.UserType) error {
	// se é o admin, então acessa tudo
	if userType == pbAuth.UserType_utAdmin {
		return nil
	}

	types, ok := mapRoutesToUserType[methodFullName]
	if !ok {
		// rota somente para administrador
		return status.Error(codes.PermissionDenied, "permission denied")
	}

	// verifica se o tipo do usuário é permitido para a rota
	for _, t := range types {
		if t == userType {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, "permission denied")
}

func IsPublicRoute(methodFullName string) bool {
	_, ok := mapIgnorePermissions[methodFullName]
	return ok
}

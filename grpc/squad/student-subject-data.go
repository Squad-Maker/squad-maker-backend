package squad

import (
	"context"
	"errors"
	"squad-maker/database"
	pbSquad "squad-maker/generated/squad"
	"squad-maker/models"
	grpcUtils "squad-maker/utils/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (s *SquadServiceServer) GetStudentSubjectData(ctx context.Context, req *pbSquad.GetStudentSubjectDataRequest) (*pbSquad.GetStudentSubjectDataResponse, error) {
	studentId := grpcUtils.GetCurrentUserIdFromMetadata(ctx)
	subjectId := grpcUtils.GetCurrentSubjectIdFromMetadata(ctx)

	if subjectId == 0 {
		return nil, status.Error(codes.InvalidArgument, "subject id cannot be zero")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pbSquad.GetStudentSubjectDataResponse{}

	studentSubjectData := &models.StudentSubjectData{}
	r := dbCon.
		InnerJoins("Student").
		InnerJoins("Subject").
		Joins("PositionOption1").
		Joins("PositionOption2").
		Joins("PreferredProject").
		Joins("CompetenceLevel").
		Where(models.StudentSubjectData{
			StudentId: studentId,
			SubjectId: subjectId,
		}, "StudentId", "SubjectId").
		First(studentSubjectData)
	if r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.Internal, r.Error.Error())
	}

	if studentSubjectData.CompetenceLevel != nil {
		resp.CompetenceLevel, err = studentSubjectData.CompetenceLevel.ConvertToProtobufMessage(dbCon)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if studentSubjectData.PositionOption1 != nil {
		resp.PositionOption_1, err = studentSubjectData.PositionOption1.ConvertToProtobufMessage(dbCon)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if studentSubjectData.PositionOption2 != nil {
		resp.PositionOption_2, err = studentSubjectData.PositionOption2.ConvertToProtobufMessage(dbCon)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if studentSubjectData.PreferredProject != nil {
		resp.PreferredProject, err = studentSubjectData.PreferredProject.ConvertToProtobufMessage(dbCon)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	resp.HadFirstUpdate = studentSubjectData.HadFirstUpdate
	resp.Tools = studentSubjectData.Tools

	return resp, nil
}

func (s *SquadServiceServer) UpdateStudentSubjectData(ctx context.Context, req *pbSquad.UpdateStudentSubjectDataRequest) (*pbSquad.UpdateStudentSubjectDataResponse, error) {
	subjectId := grpcUtils.GetCurrentSubjectIdFromMetadata(ctx)
	studentId := grpcUtils.GetCurrentUserIdFromMetadata(ctx)

	if subjectId == 0 {
		return nil, status.Error(codes.InvalidArgument, "subject id cannot be zero")
	}

	if studentId == 0 {
		return nil, status.Error(codes.InvalidArgument, "student id cannot be zero")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = dbCon.Transaction(func(tx *gorm.DB) error {
		studentSubjectData := &models.StudentSubjectData{}
		r := tx.Clauses(database.GetLockForUpdateClause(tx.Dialector.Name(), false)).
			Where(models.StudentSubjectData{
				StudentId: studentId,
				SubjectId: subjectId,
			}, "StudentId", "SubjectId").
			First(studentSubjectData)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				studentSubjectData.StudentId = studentId
				studentSubjectData.SubjectId = subjectId
			} else {
				return status.Error(codes.Internal, r.Error.Error())
			}
		}

		studentSubjectData.HadFirstUpdate = true
		studentSubjectData.Tools = req.Tools
		studentSubjectData.CompetenceLevelId = req.CompetenceLevelId
		studentSubjectData.PositionOption1Id = req.PositionOption_1Id
		studentSubjectData.PositionOption2Id = req.PositionOption_2Id
		studentSubjectData.PreferredProjectId = req.PreferredProjectId

		r = tx.Save(studentSubjectData)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.UpdateStudentSubjectDataResponse{}, nil
}

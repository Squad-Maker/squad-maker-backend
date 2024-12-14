package squad

import (
	"context"
	"errors"
	"squad-maker/database"
	pbCommon "squad-maker/generated/common"
	pbSquad "squad-maker/generated/squad"
	"squad-maker/models"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (s *SquadServiceServer) ReadCompetenceLevel(ctx context.Context, req *pbSquad.ReadCompetenceLevelRequest) (*pbSquad.CompetenceLevel, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id cannot be zero")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	competenceLevel := &models.CompetenceLevel{}
	r := dbCon.First(competenceLevel, req.Id)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "competence level not found")
		}
		return nil, status.Error(codes.Internal, r.Error.Error())
	}

	pbCompetenceLevel, err := competenceLevel.ConvertToProtobufMessage(dbCon)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return pbCompetenceLevel, nil
}

func (s *SquadServiceServer) CreateCompetenceLevel(ctx context.Context, req *pbSquad.CreateCompetenceLevelRequest) (*pbSquad.CreateCompetenceLevelResponse, error) {
	if req.SubjectId == 0 {
		return nil, status.Error(codes.InvalidArgument, "subject id cannot be zero")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name cannot be empty")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	competenceLevel := &models.CompetenceLevel{
		SubjectId: req.SubjectId,
		Name:      req.Name,
	}
	err = dbCon.Transaction(func(tx *gorm.DB) error {
		r := tx.Create(competenceLevel)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.CreateCompetenceLevelResponse{
		Id: competenceLevel.Id,
	}, nil
}

func (s *SquadServiceServer) UpdateCompetenceLevel(ctx context.Context, req *pbSquad.UpdateCompetenceLevelRequest) (*pbSquad.UpdateCompetenceLevelResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id cannot be zero")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name cannot be empty")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = dbCon.Transaction(func(tx *gorm.DB) error {
		competenceLevel := &models.CompetenceLevel{}
		r := tx.Clauses(database.GetLockForUpdateClause(tx.Dialector.Name(), false)).First(competenceLevel, req.Id)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "competence level not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		competenceLevel.Name = req.Name

		r = tx.Save(competenceLevel)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.UpdateCompetenceLevelResponse{}, nil
}

func (s *SquadServiceServer) DeleteCompetenceLevel(ctx context.Context, req *pbSquad.DeleteCompetenceLevelRequest) (*pbSquad.DeleteCompetenceLevelResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id cannot be zero")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = dbCon.Transaction(func(tx *gorm.DB) error {
		competenceLevel := &models.CompetenceLevel{}
		r := tx.Clauses(database.GetLockForUpdateClause(tx.Dialector.Name(), false)).First(competenceLevel, req.Id)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "competence level not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		r = tx.Delete(competenceLevel)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.DeleteCompetenceLevelResponse{}, nil
}

func (s *SquadServiceServer) ReadAllCompetenceLevels(ctx context.Context, req *pbCommon.ReadAllRequest) (*pbSquad.ReadAllCompetenceLevelsResponse, error) {
	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tx, err := database.PrepareWithFilters(dbCon, req.Filters, models.CompetenceLevel{}, competenceLevelHandleUnknownFilters)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return database.GetPaginatedResult[pbSquad.CompetenceLevel, pbSquad.ReadAllCompetenceLevelsResponse](ctx, tx, req.Pagination, models.CompetenceLevel{}, competenceLevelHandleUnknownOrderByFields)
}
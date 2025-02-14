package squad

import (
	"context"
	"errors"
	"squad-maker/database"
	pbCommon "squad-maker/generated/common"
	pbSquad "squad-maker/generated/squad"
	"squad-maker/models"
	grpcUtils "squad-maker/utils/grpc"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// TODO quando implementar ownership do subject, tem que validar em tudo

func (s *SquadServiceServer) ReadPosition(ctx context.Context, req *pbSquad.ReadPositionRequest) (*pbSquad.Position, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id cannot be zero")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	position := &models.Position{}
	r := dbCon.First(position, req.Id)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "position not found")
		}
		return nil, status.Error(codes.Internal, r.Error.Error())
	}

	pbPosition, err := position.ConvertToProtobufMessage(dbCon)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return pbPosition, nil
}

func (s *SquadServiceServer) CreatePosition(ctx context.Context, req *pbSquad.CreatePositionRequest) (*pbSquad.CreatePositionResponse, error) {
	subjectId := grpcUtils.GetCurrentSubjectIdFromMetadata(ctx)

	if subjectId == 0 {
		return nil, status.Error(codes.InvalidArgument, "subject id cannot be zero")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name cannot be empty")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	position := &models.Position{
		SubjectId: subjectId,
		Name:      req.Name,
	}
	err = dbCon.Transaction(func(tx *gorm.DB) error {
		// TODO verificar se o user é dono do subject da position
		r := tx.Create(position)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.CreatePositionResponse{
		Id: position.Id,
	}, nil
}

func (s *SquadServiceServer) UpdatePosition(ctx context.Context, req *pbSquad.UpdatePositionRequest) (*pbSquad.UpdatePositionResponse, error) {
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
		// TODO verificar se o user é dono do subject da position
		position := &models.Position{}
		r := tx.Clauses(database.GetLockForUpdateClause(tx.Dialector.Name(), false)).First(position, req.Id)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "position not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		position.Name = req.Name

		r = tx.Save(position)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.UpdatePositionResponse{}, nil
}

func (s *SquadServiceServer) DeletePosition(ctx context.Context, req *pbSquad.DeletePositionRequest) (*pbSquad.DeletePositionResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id cannot be zero")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = dbCon.Transaction(func(tx *gorm.DB) error {
		// TODO verificar se o user é dono do subject da position
		position := &models.Position{}
		r := tx.Clauses(database.GetLockForUpdateClause(tx.Dialector.Name(), false)).First(position, req.Id)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "position not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		r = tx.Delete(position)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.DeletePositionResponse{}, nil
}

func (s *SquadServiceServer) ReadAllPositions(ctx context.Context, req *pbCommon.ReadAllRequest) (*pbSquad.ReadAllPositionsResponse, error) {
	subjectId := grpcUtils.GetCurrentSubjectIdFromMetadata(ctx)

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var filters []*pbCommon.FilterData
	if req.Filters != nil {
		filters = append(filters, &pbCommon.FilterData{
			Filter: &pbCommon.FilterData_Grouped{
				Grouped: &pbCommon.GroupedFilterData{Filters: req.Filters},
			},
		})
	}
	filters = append(filters, &pbCommon.FilterData{
		Filter: &pbCommon.FilterData_Simple{
			Simple: &pbCommon.SimpleFilterData{
				FilterKey: "subjectId",
				Value:     strconv.FormatInt(subjectId, 10),
				Operator:  pbCommon.FilterOperator_foEqual,
			},
		},
	})

	tx, err := database.PrepareWithFilters(dbCon, filters, models.Position{}, positionHandleUnknownFilters)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return database.GetPaginatedResult[pbSquad.Position, pbSquad.ReadAllPositionsResponse](ctx, tx, req.Pagination, models.Position{}, positionHandleUnknownOrderByFields)
}

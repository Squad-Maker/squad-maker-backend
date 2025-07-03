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
// TODO quando implementar mais de um subject, tem que validar tudo pelo subject do contexto (vindo do metadata)
// recomendação: utilizar Scopes do gorm (validar em todos os arquivos necessários, não só neste)

func (s *SquadServiceServer) ReadProject(ctx context.Context, req *pbSquad.ReadProjectRequest) (*pbSquad.Project, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id cannot be zero")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	project := &models.Project{}
	r := dbCon.First(project, req.Id)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.Internal, r.Error.Error())
	}

	pbProject, err := project.ConvertToProtobufMessage(dbCon)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return pbProject, nil
}

func (s *SquadServiceServer) CreateProject(ctx context.Context, req *pbSquad.CreateProjectRequest) (*pbSquad.CreateProjectResponse, error) {
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

	project := &models.Project{
		SubjectId:   subjectId,
		Name:        req.Name,
		Description: req.Description,
		Tools:       req.Tools,
	}
	err = dbCon.Transaction(func(tx *gorm.DB) error {
		// TODO verificar se o user é dono do subject do project
		r := tx.Create(project)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		for _, position := range req.Positions {
			if position.Count <= 0 {
				continue
			}

			ppo := &models.ProjectPosition{
				ProjectId:  project.Id,
				PositionId: position.Id,
				Count:      position.Count,
			}

			r = tx.Create(ppo)
			if r.Error != nil {
				return status.Error(codes.Internal, r.Error.Error())
			}
		}

		for _, competenceLevel := range req.CompetenceLevels {
			if competenceLevel.Count <= 0 {
				continue
			}

			pcl := &models.ProjectCompetenceLevel{
				ProjectId:         project.Id,
				CompetenceLevelId: competenceLevel.Id,
				Count:             competenceLevel.Count,
			}

			r = tx.Create(pcl)
			if r.Error != nil {
				return status.Error(codes.Internal, r.Error.Error())
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.CreateProjectResponse{
		Id: project.Id,
	}, nil
}

func (s *SquadServiceServer) UpdateProject(ctx context.Context, req *pbSquad.UpdateProjectRequest) (*pbSquad.UpdateProjectResponse, error) {
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
		// TODO verificar se o user é dono do subject do project
		project := &models.Project{}
		r := tx.Clauses(database.GetLockForUpdateClause(tx.Dialector.Name(), false)).First(project, req.Id)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "project not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		project.Name = req.Name
		project.Description = req.Description
		project.Tools = req.Tools

		r = tx.Save(project)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		var positionIds []int64
		for _, position := range req.Positions {
			if position.Count <= 0 {
				// delete
				r = tx.Where(models.ProjectPosition{
					ProjectId:  project.Id,
					PositionId: position.Id,
				}, "ProjectId", "PositionId").Delete(&models.ProjectPosition{})
				if r.Error != nil {
					return status.Error(codes.Internal, r.Error.Error())
				}
				continue
			}

			ppo := &models.ProjectPosition{
				ProjectId:  project.Id,
				PositionId: position.Id,
				Count:      position.Count,
			}

			r = tx.Save(ppo)
			if r.Error != nil {
				return status.Error(codes.Internal, r.Error.Error())
			}

			positionIds = append(positionIds, ppo.PositionId)
		}

		r = tx.Where(models.ProjectPosition{
			ProjectId: project.Id,
		}, "ProjectId")
		if len(positionIds) > 0 {
			r = r.Where("ppo_position_id NOT IN (?)", positionIds)
		}
		r = r.Delete(&models.ProjectPosition{})
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		var competenceLevelIds []int64
		for _, competenceLevel := range req.CompetenceLevels {
			if competenceLevel.Count <= 0 {
				// delete
				r = tx.Where(models.ProjectCompetenceLevel{
					ProjectId:         project.Id,
					CompetenceLevelId: competenceLevel.Id,
				}, "ProjectId", "CompetenceLevelId").Delete(&models.ProjectCompetenceLevel{})
				if r.Error != nil {
					return status.Error(codes.Internal, r.Error.Error())
				}
				continue
			}

			pcl := &models.ProjectCompetenceLevel{
				ProjectId:         project.Id,
				CompetenceLevelId: competenceLevel.Id,
				Count:             competenceLevel.Count,
			}
			r = tx.Save(pcl)
			if r.Error != nil {
				return status.Error(codes.Internal, r.Error.Error())
			}
			competenceLevelIds = append(competenceLevelIds, pcl.CompetenceLevelId)
		}
		r = tx.Where(models.ProjectCompetenceLevel{
			ProjectId: project.Id,
		}, "ProjectId")
		if len(competenceLevelIds) > 0 {
			r = r.Where("pcl_competence_level_id NOT IN (?)", competenceLevelIds)
		}
		r = r.Delete(&models.ProjectCompetenceLevel{})
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.UpdateProjectResponse{}, nil
}

func (s *SquadServiceServer) DeleteProject(ctx context.Context, req *pbSquad.DeleteProjectRequest) (*pbSquad.DeleteProjectResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id cannot be zero")
	}

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = dbCon.Transaction(func(tx *gorm.DB) error {
		// TODO verificar se o user é dono do subject do project
		project := &models.Project{}
		r := tx.Clauses(database.GetLockForUpdateClause(tx.Dialector.Name(), false)).First(project, req.Id)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "project not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		r = tx.Delete(project)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.DeleteProjectResponse{}, nil
}

func (s *SquadServiceServer) ReadAllProjects(ctx context.Context, req *pbCommon.ReadAllRequest) (*pbSquad.ReadAllProjectsResponse, error) {
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

	tx, err := database.PrepareWithFilters(dbCon, filters, models.Project{}, projectHandleUnknownFilters)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return database.GetPaginatedResult[pbSquad.Project, pbSquad.ReadAllProjectsResponse](ctx, tx, req.Pagination, models.Project{}, projectHandleUnknownOrderByFields)
}

package squad

import (
	"context"
	"squad-maker/database"
	pbCommon "squad-maker/generated/common"
	pbSquad "squad-maker/generated/squad"
	"squad-maker/models"
	grpcUtils "squad-maker/utils/grpc"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

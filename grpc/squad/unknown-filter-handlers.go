package squad

import (
	"fmt"
	pbCommon "squad-maker/generated/common"
	grpcUtils "squad-maker/utils/grpc"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func competenceLevelHandleUnknownFilters(tx *gorm.DB, filter *pbCommon.FilterData) (*gorm.DB, clause.Expression, error) {
	fmt.Printf("competenceLevelHandleUnknownFilters received filter: %+v\n", filter)
	return tx, nil, nil
}

func positionHandleUnknownFilters(tx *gorm.DB, filter *pbCommon.FilterData) (*gorm.DB, clause.Expression, error) {
	fmt.Printf("positionHandleUnknownFilters received filter: %+v\n", filter)
	return tx, nil, nil
}

func projectHandleUnknownFilters(tx *gorm.DB, filter *pbCommon.FilterData) (*gorm.DB, clause.Expression, error) {
	ctx := tx.Statement.Context
	userId := grpcUtils.GetCurrentUserIdFromMetadata(ctx)

	switch f := filter.Filter.(type) {
	case *pbCommon.FilterData_Simple:
		if f.Simple.FilterKey == "inProject" {
			// se usar isso pra user professor... não vai retornar nada igual

			// só vai funcionar com igual e diferente
			not := ""
			if f.Simple.Operator == pbCommon.FilterOperator_foNotEqual && f.Simple.Value == "true" ||
				f.Simple.Operator == pbCommon.FilterOperator_foEqual && f.Simple.Value == "false" {
				not = "NOT "
			}

			conds := clause.And(tx.Statement.BuildCondition(not+"EXISTS (SELECT 1 FROM project_students WHERE pst_student_id = ? AND pst_project_id = pro_id)", userId)...)

			if filter.IsOr {
				conds = clause.Or(conds)
			}

			return tx, conds, nil
		}
	case *pbCommon.FilterData_In:
		// nada por enquanto
	}
	fmt.Printf("projectHandleUnknownFilters received filter: %+v\n", filter)
	return tx, nil, nil
}

func studentInSubjectHandleUnknownFilters(tx *gorm.DB, filter *pbCommon.FilterData) (*gorm.DB, clause.Expression, error) {
	switch f := filter.Filter.(type) {
	case *pbCommon.FilterData_Simple:
		if f.Simple.FilterKey == "inProject" {
			// se usar isso pra user professor... não vai retornar nada igual

			// só vai funcionar com igual e diferente
			not := ""
			if f.Simple.Operator == pbCommon.FilterOperator_foNotEqual && f.Simple.Value == "true" ||
				f.Simple.Operator == pbCommon.FilterOperator_foEqual && f.Simple.Value == "false" {
				not = "NOT "
			}

			conds := clause.And(tx.Statement.BuildCondition(not + `EXISTS (
				SELECT 1 
				FROM project_students 
				JOIN projects ON pro_id = pst_project_id AND pro_deleted_at IS NULL 
				WHERE pst_student_id = ssd_student_id AND pro_subject_id = ssd_subject_id
			)`)...)

			if filter.IsOr {
				conds = clause.Or(conds)
			}

			return tx, conds, nil
		}
	case *pbCommon.FilterData_In:
		// nada por enquanto
	}
	fmt.Printf("studentInSubjectHandleUnknownFilters received filter: %+v\n", filter)
	return tx, nil, nil
}

package squad

import (
	"fmt"
	pbCommon "squad-maker/generated/common"

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

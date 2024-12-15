package squad

import (
	"fmt"
	pbCommon "squad-maker/generated/common"

	"gorm.io/gorm"
)

func competenceLevelHandleUnknownOrderByFields(tx *gorm.DB, sortOption *pbCommon.SortOption) (*gorm.DB, error) {
	fmt.Printf("competenceLevelHandleUnknownOrderByFields received order by: %+v\n", sortOption)
	return tx, nil
}

func positionHandleUnknownOrderByFields(tx *gorm.DB, sortOption *pbCommon.SortOption) (*gorm.DB, error) {
	fmt.Printf("positionHandleUnknownOrderByFields received order by: %+v\n", sortOption)
	return tx, nil
}

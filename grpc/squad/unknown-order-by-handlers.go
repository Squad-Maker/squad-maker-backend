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

func projectHandleUnknownOrderByFields(tx *gorm.DB, sortOption *pbCommon.SortOption) (*gorm.DB, error) {
	fmt.Printf("projectHandleUnknownOrderByFields received order by: %+v\n", sortOption)
	return tx, nil
}

func studentInSubjectHandleUnknownOrderByFields(tx *gorm.DB, sortOption *pbCommon.SortOption) (*gorm.DB, error) {
	fmt.Printf("studentInSubjectHandleUnknownOrderByFields received order by: %+v\n", sortOption)
	return tx, nil
}

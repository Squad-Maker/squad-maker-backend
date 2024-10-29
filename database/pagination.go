package database

import (
	"context"
	"errors"
	"reflect"
	"strings"

	pb "squad-maker/generated/common"

	"gorm.io/gorm"
)

type HandleOrderFieldNotFound func(tx *gorm.DB, sortOption *pb.SortOption) (*gorm.DB, error)

type ModelWithProtobufMessageMapper[M any] interface{}

type PModelWithProtobufMessageMapper[M any, D ModelWithProtobufMessageMapper[M]] interface {
	*D
	ConvertToProtobufMessage(tx *gorm.DB) (*M, error)
}

func PreparePaginatedQuery(ctx context.Context, tx *gorm.DB, pagination *pb.PaginationData, dest any, applyLimitOffset bool, handleExtra HandleOrderFieldNotFound) (*gorm.DB, error) {
	if tx == nil || pagination == nil || dest == nil {
		return nil, errors.New("nil argument")
	}

	destType := reflect.TypeOf(dest)
	for destType.Kind() == reflect.Ptr || destType.Kind() == reflect.Slice {
		destType = destType.Elem()
	}

	if destType.Kind() != reflect.Struct {
		return nil, errors.New("dest is not a struct")
	}

	tableName := tx.NamingStrategy.TableName(destType.Name())

	for _, sortOption := range pagination.SortOptions {

		if sortOption.Sort == "" {
			continue
		}

		// garantir que a primeira letra é maiúscula
		s := strings.ToUpper(sortOption.Sort[:1]) + sortOption.Sort[1:]

		// verifica se o campo existe na struct
		_, ok := destType.FieldByName(s)

		if !ok {
			if handleExtra != nil {
				var err error
				tx, err = handleExtra(tx, sortOption)
				if err != nil {
					return nil, err
				}
			}
			continue
		}

		dbName := tx.NamingStrategy.ColumnName(tableName, s)

		if dbName != "" {
			var order string
			if sortOption.Order == pb.Order_ordDescending {
				order = " DESC"
			}
			tx = tx.Order(dbName + order)
		}
	}

	if applyLimitOffset {
		limit := pagination.Limit
		if limit > 1000 {
			limit = 1000
		} else if limit < 1 {
			limit = 50
		}
		if pagination.Page == 0 {
			pagination.Page = 1
		}

		if limit > 0 {
			tx = tx.Limit(int(limit))
			if pagination.Page > 1 {
				offset := (pagination.Page - 1) * limit
				tx = tx.Offset(int(offset))
			}
		}
	}

	return tx, nil
}

func GetPaginatedResult[M any, L any, D ModelWithProtobufMessageMapper[M], PD PModelWithProtobufMessageMapper[M, D]](ctx context.Context, tx *gorm.DB, pagination *pb.PaginationData, dest D, handleExtra HandleOrderFieldNotFound) (*L, error) {
	if pagination == nil {
		pagination = &pb.PaginationData{
			Limit: 10,
		}
	}

	tx, err := PreparePaginatedQuery(ctx, tx, pagination, dest, false, handleExtra)
	if err != nil {
		return nil, err
	}

	limit := pagination.Limit
	if limit > 1000 {
		limit = 1000
	} else if limit < 1 {
		limit = 50
	}
	if pagination.Page == 0 {
		pagination.Page = 1
	}

	response := new(L)

	responseMetadata := &pb.PaginatedResponseMetadata{}
	responseMetadata.Limit = limit
	responseMetadata.Page = pagination.Page

	var count int64

	if tx.Statement.Model == nil {
		tx = tx.Model(dest)
	}

	r := tx.Session(&gorm.Session{}).Count(&count)
	if r.Error != nil {
		return nil, r.Error
	}

	if count == 0 {
		// não tem nada pra retornar
		// nem precisa fazer outra query

		return response, nil
	}

	responseMetadata.Total = uint64(count)

	reflect.ValueOf(response).Elem().FieldByName("Metadata").Set(reflect.ValueOf(responseMetadata))

	// código duplicado
	destType := reflect.TypeOf(dest)
	for destType.Kind() == reflect.Ptr || destType.Kind() == reflect.Slice {
		destType = destType.Elem()
	}
	// fim código duplicado

	sliceType := reflect.SliceOf(reflect.PointerTo(destType))
	sliceInterface := reflect.New(sliceType).Interface()

	tx = tx.Limit(int(limit))
	if pagination.Page > 1 {
		offset := (pagination.Page - 1) * limit
		tx = tx.Offset(int(offset))
	}

	r = tx.Find(sliceInterface)
	if r.Error != nil {
		return nil, r.Error
	}

	dataSlice := sliceInterface.(*[]PD)
	data := make([]*M, 0, len(*dataSlice))
	for _, result := range *dataSlice {
		d, err := result.ConvertToProtobufMessage(tx.Session(&gorm.Session{
			NewDB: true,
		}))
		if err != nil {
			return nil, err
		}
		data = append(data, d)
	}

	reflect.ValueOf(response).Elem().FieldByName("Data").Set(reflect.ValueOf(data))

	return response, nil
}

func GetPaginatedResultWithViewModel[M any, L any, D ModelWithProtobufMessageMapper[M], PD PModelWithProtobufMessageMapper[M, D]](ctx context.Context, tx *gorm.DB, pagination *pb.PaginationData, dest any, viewModel D, handleExtra HandleOrderFieldNotFound) (*L, error) {
	if pagination == nil {
		pagination = &pb.PaginationData{
			Limit: 10,
		}
	}

	tx, err := PreparePaginatedQuery(ctx, tx, pagination, dest, false, handleExtra)
	if err != nil {
		return nil, err
	}

	if tx.Statement.Model == nil {
		tx = tx.Model(dest)
	}

	tx, err = PrepareForViewModel(tx, viewModel)
	if err != nil {
		return nil, err
	}

	limit := pagination.Limit
	if limit > 1000 {
		limit = 1000
	} else if limit < 1 {
		limit = 50
	}
	if pagination.Page == 0 {
		pagination.Page = 1
	}

	response := new(L)

	responseMetadata := &pb.PaginatedResponseMetadata{}
	responseMetadata.Limit = limit
	responseMetadata.Page = pagination.Page

	var count int64
	r := tx.Model(dest).Session(&gorm.Session{}).Count(&count)
	if r.Error != nil {
		return nil, r.Error
	}

	if count == 0 {
		// não tem nada pra retornar
		// nem precisa fazer outra query

		return response, nil
	}

	responseMetadata.Total = uint64(count)

	reflect.ValueOf(response).Elem().FieldByName("Metadata").Set(reflect.ValueOf(responseMetadata))

	// código duplicado
	destType := reflect.TypeOf(dest)
	for destType.Kind() == reflect.Ptr || destType.Kind() == reflect.Slice {
		destType = destType.Elem()
	}
	// fim código duplicado
	viewModelType := reflect.TypeOf(viewModel)
	for viewModelType.Kind() == reflect.Ptr || viewModelType.Kind() == reflect.Slice {
		viewModelType = viewModelType.Elem()
	}

	destInterface := reflect.New(destType).Interface()

	sliceType := reflect.SliceOf(reflect.PointerTo(viewModelType))
	sliceInterface := reflect.New(sliceType).Interface()

	if limit > 0 {
		tx = tx.Limit(int(limit))
		if pagination.Page > 1 {
			offset := (pagination.Page - 1) * limit
			tx = tx.Offset(int(offset))
		}
	}

	r = tx.Model(destInterface).Scan(sliceInterface)
	if r.Error != nil {
		return nil, r.Error
	}

	dataSlice := sliceInterface.(*[]PD)
	data := make([]*M, 0, len(*dataSlice))
	for _, result := range *dataSlice {
		d, err := result.ConvertToProtobufMessage(tx.Session(&gorm.Session{
			NewDB: true,
		}))
		if err != nil {
			return nil, err
		}
		data = append(data, d)
	}

	reflect.ValueOf(response).Elem().FieldByName("Data").Set(reflect.ValueOf(data))

	return response, nil
}

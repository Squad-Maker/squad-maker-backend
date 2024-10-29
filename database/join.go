package database

import (
	"errors"
	"reflect"
	"strconv"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type (
	joinedTables  map[string]struct{}
	joinedSchemas []*schema.Schema
)

const (
	alreadyJoinedKey string = "join-already-joined"
	joinedSchemasKey string = "join-joined-schemas"
)

var (
	deletedAtType reflect.Type

	cacheStore = &sync.Map{}
)

func init() {
	deletedAtType = reflect.TypeOf(gorm.DeletedAt{})
}

func Join(tx *gorm.DB, modelFromJoin, modelToJoin interface{}, propertyFromSource, propertyFromDest, extraJoinFilters string) (*gorm.DB, error) {
	return join("", tx, modelFromJoin, modelToJoin, propertyFromSource, propertyFromDest, extraJoinFilters)
}

func LeftJoin(tx *gorm.DB, modelFromJoin, modelToJoin interface{}, propertyFromSource, propertyFromDest, extraJoinFilters string) (*gorm.DB, error) {
	return join("LEFT", tx, modelFromJoin, modelToJoin, propertyFromSource, propertyFromDest, extraJoinFilters)
}

// TODO trocar implementação do LeftJoin pra uma que use o schema.Parse como base, pra evitar problemas
// do jeito que está agora, somente problemas bem específicos (que provavelmente nunca vão ocorrer) podem ocorrer

// tx (with model), model to join, property from source, property from dest
// modelFromJoin must be the same as tx.Statement.Model or tx.Statement.Model == nil or modelFromJoin was already joined with this function
func join(joinType string, tx *gorm.DB, modelFromJoin, modelToJoin interface{}, propertyFromSource, propertyFromDest, extraJoinFilters string) (*gorm.DB, error) {
	iAlreadyJoined, ok := tx.Get(alreadyJoinedKey)

	var alreadyJoined joinedTables

	if ok {
		alreadyJoined, ok = iAlreadyJoined.(joinedTables)
		if !ok {
			alreadyJoined = make(joinedTables)
		}
	} else {
		alreadyJoined = make(joinedTables)
	}
	iAlreadyJoined = nil

	iJoinedSchemas, ok := tx.Get(joinedSchemasKey)

	var joinedSchemasArr joinedSchemas

	if ok {
		joinedSchemasArr, ok = iJoinedSchemas.(joinedSchemas)
		if !ok {
			joinedSchemasArr = make(joinedSchemas, 0, 2)
		}
	} else {
		joinedSchemasArr = make(joinedSchemas, 0, 2)
	}
	iJoinedSchemas = nil

	// validate models
	modelFromJoinType := reflect.TypeOf(modelFromJoin)
	for modelFromJoinType.Kind() == reflect.Ptr || modelFromJoinType.Kind() == reflect.Slice {
		modelFromJoinType = modelFromJoinType.Elem()
	}

	if modelFromJoinType.Kind() != reflect.Struct {
		return nil, errors.New("modelFromJoin is not a struct")
	}

	modelToJoinType := reflect.TypeOf(modelToJoin)
	for modelToJoinType.Kind() == reflect.Ptr || modelToJoinType.Kind() == reflect.Slice {
		modelToJoinType = modelToJoinType.Elem()
	}

	if modelToJoinType.Kind() != reflect.Struct {
		return nil, errors.New("modelToJoin is not a struct")
	}
	// end validate models

	// init schemas
	modelFromJoinSchema, err := schema.Parse(modelFromJoin, cacheStore, tx.NamingStrategy)
	if err != nil {
		return nil, err
	}

	modelToJoinSchema, err := schema.Parse(modelToJoin, cacheStore, tx.NamingStrategy)
	if err != nil {
		return nil, err
	}
	// end init schemas

	if tx.Statement.Model != nil {
		modelType := reflect.TypeOf(tx.Statement.Model)
		for modelType.Kind() == reflect.Ptr || modelType.Kind() == reflect.Slice {
			modelType = modelType.Elem()
		}

		if modelFromJoinType != modelType {
			if _, ok := alreadyJoined[modelFromJoinSchema.Table]; !ok {
				return nil, errors.New("model to join from is not valid")
			}
		}
	} else {
		modelFromJoinPtr := reflect.New(modelFromJoinType).Interface()
		tx = tx.Model(modelFromJoinPtr)
	}

	if _, ok := alreadyJoined[modelFromJoinSchema.Table]; !ok {
		alreadyJoined[modelFromJoinSchema.Table] = struct{}{}
		joinedSchemasArr = append(joinedSchemasArr, modelFromJoinSchema)
	}

	if _, ok = alreadyJoined[modelToJoinSchema.Table]; ok {
		// TODO suporte a fazer join da mesma tabela mais de uma vez
		// isso inclui o próprio model do statement (não pode fazer join com ele mesmo por aqui)
		// por enquanto vai ficar sem, pois envolveria aliases
		return nil, errors.New("model already joined")
	}

	f := modelFromJoinSchema.LookUpField(propertyFromSource)

	var columnFromJoin string
	if f != nil {
		columnFromJoin = f.DBName
	} else {
		// se não encontrar a propriedade, vai direto (pra permitir informar direto uma coluna ou expressão)
		// TODO permitir isso mesmo? pode ser fonte de SQL injection caso não cuide muito...
		// por enquanto vou permitir, pois user inputs não serão passados como parâmetro pra cá

		columnFromJoin = propertyFromSource
	}

	f = modelToJoinSchema.LookUpField(propertyFromDest)

	var columnToJoin string
	if f != nil {
		columnToJoin = f.DBName
	} else {
		// se não encontrar a propriedade, vai direto (pra permitir informar direto uma coluna ou expressão)
		// TODO permitir isso mesmo? pode ser fonte de SQL injection caso não cuide muito...
		// por enquanto vou permitir, pois user inputs não serão passados como parâmetro pra cá

		columnToJoin = propertyFromDest
	}

	joinString := joinType + " JOIN " + modelToJoinSchema.Table + " ON " + modelFromJoinSchema.Table + "." + columnFromJoin + " = " + modelToJoinSchema.Table + "." + columnToJoin

	f = getFieldByIndirectType(modelToJoinSchema, deletedAtType)
	if f != nil {
		joinString = joinString + " AND " + modelToJoinSchema.Table + "." + f.DBName + " IS NULL"
	}

	if extraJoinFilters != "" {
		joinString = joinString + " AND " + extraJoinFilters
	}

	tx = tx.Joins(joinString)

	alreadyJoined[modelToJoinSchema.Table] = struct{}{}
	joinedSchemasArr = append(joinedSchemasArr, modelToJoinSchema)

	tx = tx.Set(alreadyJoinedKey, alreadyJoined)
	tx = tx.Set(joinedSchemasKey, joinedSchemasArr)

	return tx, nil
}

// tx (with model), viewmodel
func PrepareForViewModel(tx *gorm.DB, viewModel interface{}) (*gorm.DB, error) {
	// pega o model do tx -> se não existir, erro
	if tx.Statement.Model == nil {
		return nil, errors.New("tx model cannot be nil")
	}

	// valida o viewModel
	viewModelType := reflect.TypeOf(viewModel)
	for viewModelType.Kind() == reflect.Ptr || viewModelType.Kind() == reflect.Slice {
		viewModelType = viewModelType.Elem()
	}

	if viewModelType.Kind() != reflect.Struct {
		return nil, errors.New("viewModel is not a struct")
	}

	viewModelTableName := tx.NamingStrategy.TableName(viewModelType.Name())

	// verifica se existe a lista de campos processados na tx
	iJoinedSchemas, ok := tx.Get(joinedSchemasKey)

	var selectString string

	if ok {
		joinedSchemasArr, ok := iJoinedSchemas.(joinedSchemas)
		iJoinedSchemas = nil
		if ok && len(joinedSchemasArr) > 0 {
			// se sim, usar ela pra fazer o select pro viewModel

			var fieldsAdded map[string]int
			fieldsAdded = make(map[string]int)

			for _, tableSchema := range joinedSchemasArr {

				// TODO talvez mudar a implementação toda pra não ser um Type, mas ser uma instância, ou até mesmo o próprio schema
				// esse pedaço de código foi feito depois de praticamente toda implementação ter sido terminada, por isso não ficou totalmente certo
				// tableSchema, err := schema.Parse(reflect.New(tableType).Interface(), cacheStore, tx.NamingStrategy)
				// if err != nil {
				// 	return nil, err
				// }

				for _, f := range tableSchema.Fields {
					if f.DBName == "" {
						continue
					}

					originalName := f.DBName
					toName := tx.NamingStrategy.ColumnName(viewModelTableName, f.Name)

					i := fieldsAdded[f.Name]
					i += 1
					fieldsAdded[f.Name] = i
					if i != 1 {
						toName = toName + strconv.FormatInt(int64(i), 10)
					}

					selectString = selectString + tableSchema.Table + "." + originalName + " AS " + toName + ", "
				}
			}

			fieldsAdded = nil

			// warning: not "UTF-8-aware"
			selectString = selectString[:len(selectString)-2]

			tx = tx.Select(selectString)
			return tx, nil
		}
	}
	iJoinedSchemas = nil

	// se não, usar somente o que tiver do tx.Statement.Model pra pegar as colunas e renomear
	if tx.Statement.Schema == nil {
		err := tx.Statement.Parse(tx.Statement.Model)
		if err != nil {
			return nil, err
		}
	}

	modelTableName := tx.Statement.Schema.Table

	for _, f := range tx.Statement.Schema.Fields {
		if f.DBName == "" {
			continue
		}

		originalName := f.DBName
		toName := tx.NamingStrategy.ColumnName(viewModelTableName, f.Name)

		selectString = selectString + modelTableName + "." + originalName + " AS " + toName + ", "
	}

	// warning: not "UTF-8-aware"
	selectString = selectString[:len(selectString)-2]

	tx = tx.Select(selectString)
	return tx, nil
}

func getFieldByIndirectType(t *schema.Schema, indirectType reflect.Type) *schema.Field {
	for _, f := range t.Fields {
		if f.IndirectFieldType == indirectType {
			return f
		}
	}
	return nil
}

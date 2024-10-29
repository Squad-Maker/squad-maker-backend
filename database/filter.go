package database

import (
	"errors"
	"reflect"
	"strings"

	pb "squad-maker/generated/common"

	"github.com/jinzhu/now"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type rawValue struct {
	value string
}

type HandleFilterFieldNotFound func(tx *gorm.DB, filter *pb.FilterData) (*gorm.DB, clause.Expression, error)

func PrepareWithFilters(tx *gorm.DB, filters []*pb.FilterData, dest interface{}, handleExtra HandleFilterFieldNotFound) (*gorm.DB, error) {
	if filters == nil {
		return tx, nil
	}

	tx2, conds, err := BuildFilterExprs(tx, filters, dest, handleExtra)
	if err != nil {
		return nil, err
	}
	if len(conds) > 0 {
		if len(conds) == 1 {
			if orConds, ok := conds[0].(clause.OrConditions); ok && len(orConds.Exprs) == 1 {
				conds = []clause.Expression{
					orConds.Exprs[0],
				}
			}
		}

		tx2.Statement.AddClause(clause.Where{Exprs: []clause.Expression{clause.And(conds...)}})
	}
	return tx2, nil
}

func BuildFilterExprs(tx *gorm.DB, filters []*pb.FilterData, dest interface{}, handleExtra HandleFilterFieldNotFound) (*gorm.DB, []clause.Expression, error) {
	if tx == nil || filters == nil || dest == nil {
		return nil, nil, errors.New("nil argument")
	}

	destType := reflect.TypeOf(dest)
	for destType.Kind() == reflect.Ptr || destType.Kind() == reflect.Slice {
		destType = destType.Elem()
	}

	if destType.Kind() != reflect.Struct {
		return nil, nil, errors.New("dest is not a struct")
	}

	tableName := tx.NamingStrategy.TableName(destType.Name())

	var err error

	var conds []clause.Expression

	for _, filterData := range filters {
		// checar se o campo do filtro existe no model
		// se não existir, chamar a função passada por parâmetro

		if err != nil {
			return nil, nil, err
		}

		switch f := filterData.Filter.(type) {
		case *pb.FilterData_Simple:
			var expr clause.Expression

			// (╯°□°）╯︵ ┻━┻
			if f.Simple.FilterKey == "" {
				// ┬─┬ ノ( ゜-゜ノ)
				continue
			}
			// garantir que a primeira letra é maiúscula (espera-se PascalCase, mas o ngx-grpc gera tudo como camelCase)
			s := strings.ToUpper(f.Simple.FilterKey[:1]) + f.Simple.FilterKey[1:]

			_, ok := destType.FieldByName(s)
			if !ok {
				if handleExtra == nil {
					continue
				}

				tx, expr, err = handleExtra(tx, filterData)
				if err != nil {
					return nil, nil, err
				}
			} else {

				dbName := tx.NamingStrategy.ColumnName(tableName, s)

				var underlyingExpr clause.Expression
				if f.Simple.IsDynamic {
					underlyingExpr, err = getDynamicFilterValue(tx, dbName, f.Simple)
					if err != nil {
						return nil, nil, err
					}
				} else {
					underlyingExpr, ok = BuildUnderlyingFilterClause(tx.Statement, dbName, f.Simple.Value, f.Simple.Operator)
					if !ok {
						return nil, nil, errors.New("underlyingExpr not ok")
					}
				}

				if filterData.IsOr {
					expr = clause.Or(underlyingExpr)
				} else {
					expr = underlyingExpr
				}
			}

			if expr != nil {
				conds = append(conds, expr)
			}
		case *pb.FilterData_In:
			var expr clause.Expression

			// garantir que a primeira letra é maiúscula (espera-se PascalCase, mas o ngx-grpc gera tudo como camelCase)
			s := strings.ToUpper(f.In.FilterKey[:1]) + f.In.FilterKey[1:]

			_, ok := destType.FieldByName(s)
			if !ok {
				if handleExtra == nil {
					continue
				}

				tx, expr, err = handleExtra(tx, filterData)
				if err != nil {
					return nil, nil, err
				}
			} else {
				dbName := tx.NamingStrategy.ColumnName(tableName, s)

				var underlyingExprs []clause.Expression
				if f.In.IsNot {
					underlyingExprs = tx.Statement.BuildCondition(dbName+" NOT IN ?", f.In.Value)
				} else {
					underlyingExprs = tx.Statement.BuildCondition(dbName+" IN ?", f.In.Value)
				}

				if underlyingExprs == nil {
					return nil, nil, errors.New("underlyingExprs nil")
				}

				var underlyingExpr clause.Expression

				if len(underlyingExprs) > 1 {
					underlyingExpr = clause.And(underlyingExprs...)
				} else if len(underlyingExprs) == 1 {
					underlyingExpr = underlyingExprs[0]
				}

				if underlyingExpr == nil {
					return nil, nil, errors.New("underlyingExpr nil")
				}

				if filterData.IsOr {
					expr = clause.Or(underlyingExpr)
				} else {
					expr = underlyingExpr
				}
			}

			if expr != nil {
				conds = append(conds, expr)
			}
		case *pb.FilterData_Grouped:
			if f.Grouped.Filters == nil {
				// ignora o grupo caso ele não contenha filtros
				continue
			}

			var groupedConds []clause.Expression
			tx, groupedConds, err = BuildFilterExprs(tx, f.Grouped.Filters, dest, handleExtra)
			if err != nil {
				return nil, nil, err
			}
			if groupedConds != nil {
				if filterData.IsOr {
					conds = append(conds, clause.Or(clause.And(groupedConds...)))
				} else {
					conds = append(conds, clause.And(groupedConds...))
				}
			}
		}
	}

	return tx, conds, nil
}

func BuildUnderlyingFilterClause(statement *gorm.Statement, dbFieldName string, value interface{}, operator pb.FilterOperator) (clause.Expression, bool) {
	var op string

	if dbFieldName == "" {
		return nil, false
	}

	checkValueEmpty := true
	switch operator {
	case pb.FilterOperator_foEqual:
		op = "="
	case pb.FilterOperator_foNotEqual:
		op = "<>"
	case pb.FilterOperator_foLike:
		op = "LIKE"
	case pb.FilterOperator_foLikeInsensitive:
		op = "ILIKE"
	case pb.FilterOperator_foIsNull:
		op = "IS NULL"
		value = nil // ensure value is empty
		checkValueEmpty = false
	case pb.FilterOperator_foIsNotNull:
		op = "IS NOT NULL"
		value = nil // ensure value is empty
		checkValueEmpty = false
	case pb.FilterOperator_foLessThan:
		op = "<"
	case pb.FilterOperator_foLessThanEqual:
		op = "<="
	case pb.FilterOperator_foGreaterThan:
		op = ">"
	case pb.FilterOperator_foGreaterThanEqual:
		op = ">="
	}

	if checkValueEmpty && value == nil {
		return nil, false
	}

	var exprs []clause.Expression
	if value != nil {
		rawValue, ok := value.(rawValue)
		if ok {
			// é provável que eu pudesse passar um clause.Expr como value e fazer isso funcionar sem um tipo específico
			// mas fica mais fácil de desenvolver agora com um tipo específico
			exprs = statement.BuildCondition(dbFieldName + " " + op + " " + rawValue.value)
		} else {
			exprs = statement.BuildCondition(dbFieldName+" "+op+" ?", value)
		}
	} else {
		exprs = statement.BuildCondition(dbFieldName + " " + op)
	}

	var expr clause.Expression
	if len(exprs) > 1 {
		expr = clause.And(exprs...)
	} else if len(exprs) == 1 {
		expr = exprs[0]
	}
	return expr, expr != nil
}

func getDynamicFilterValue(tx *gorm.DB, dbName string, simple *pb.SimpleFilterData) (clause.Expression, error) {
	// não vou fazer isso ser compatível com o SQL Server por enquanto

	if !simple.IsDynamic {
		// essa é uma função interna, então só vai ser usada em coisas controladas pelo dev
		// se o dev chamar essa func com um filtro que não é dinâmico, então:
		return nil, errors.New("alguém fez uma cagadinha... HEHE")
		// basicamente nunca vai retornar esse erro
	}

	// por enquanto só tem filtros dinâmicos pra data
	// então vou assumir que só vai chegar aqui pra campos de data
	// não é nada inseguro, o máximo q vai acontecer é um erro no banco
	// vou assumir também que o dbName já está tratado (isso sim poderia ser inseguro, mas do jeito que essa função é usada, não é)
	switch simple.Value {
	case "now":
		expr, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, rawValue{"CURRENT_TIMESTAMP"}, simple.Operator)
		if !ok {
			return nil, errors.New("expr not ok")
		}
		return expr, nil
	case "today":
		expr, ok := BuildUnderlyingFilterClause(tx.Statement, dbName+"::DATE", rawValue{"CURRENT_DATE"}, simple.Operator)
		if !ok {
			return nil, errors.New("expr not ok")
		}
		return expr, nil
	case "last_week":
		// expr, ok := BuildUnderlyingFilterClause(tx.Statement, dbName+"::DATE", rawValue{"CURRENT_DATE - INTERVAL '7 DAYS'"}, simple.Operator)
		// if !ok {
		// 	return nil, errors.New("expr not ok")
		// }
		// return expr, nil

		startDate := now.BeginningOfDay().AddDate(0, 0, -7)
		endDate := now.EndOfDay()

		expr, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, startDate, pb.FilterOperator_foGreaterThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		expr2, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, endDate, pb.FilterOperator_foLessThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		return clause.And(expr, expr2), nil
	case "this_week":
		startDate := now.BeginningOfWeek()
		endDate := now.EndOfWeek()

		expr, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, startDate, pb.FilterOperator_foGreaterThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		expr2, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, endDate, pb.FilterOperator_foLessThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		return clause.And(expr, expr2), nil
	case "current_month":
		startDate := now.BeginningOfMonth()
		endDate := now.EndOfMonth()

		expr, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, startDate, pb.FilterOperator_foGreaterThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		expr2, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, endDate, pb.FilterOperator_foLessThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		return clause.And(expr, expr2), nil
	case "last_30days":
		// expr, ok := BuildUnderlyingFilterClause(tx.Statement, dbName+"::DATE", rawValue{"CURRENT_DATE - INTERVAL '30 DAYS'"}, simple.Operator)
		// if !ok {
		// 	return nil, errors.New("expr not ok")
		// }
		// return expr, nil

		startDate := now.BeginningOfDay().AddDate(0, 0, -30)
		endDate := now.EndOfDay()

		expr, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, startDate, pb.FilterOperator_foGreaterThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		expr2, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, endDate, pb.FilterOperator_foLessThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		return clause.And(expr, expr2), nil
	case "current_quarter":
		startDate := now.BeginningOfQuarter()
		endDate := now.EndOfQuarter()

		expr, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, startDate, pb.FilterOperator_foGreaterThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		expr2, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, endDate, pb.FilterOperator_foLessThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		return clause.And(expr, expr2), nil
	case "current_year":
		startDate := now.BeginningOfYear()
		endDate := now.EndOfYear()

		expr, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, startDate, pb.FilterOperator_foGreaterThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		expr2, ok := BuildUnderlyingFilterClause(tx.Statement, dbName, endDate, pb.FilterOperator_foLessThanEqual)
		if !ok {
			return nil, errors.New("expr not ok")
		}

		return clause.And(expr, expr2), nil
		// default:
	}
	return nil, errors.New("invalid dynamic filter")
}

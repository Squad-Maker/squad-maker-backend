package database

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/hints"
)

func GetLockForUpdateClause(dialectorName string, skipLocked bool) clause.Expression {
	var c clause.Expression
	switch dialectorName {
	case "sqlserver":
		hint := WithHint{Keys: []string{"ROWLOCK", "UPDLOCK"}}
		if skipLocked {
			hint.Keys = append(hint.Keys, "READPAST")
		}
		c = hint
	default:
		hint := clause.Locking{Strength: "NO KEY UPDATE"}
		if skipLocked {
			hint.Options = "SKIP LOCKED"
		}
		hint.Table = clause.Table{Name: clause.CurrentTable}
		c = hint
	}
	return c
}

// Based on: https://github.com/go-gorm/hints/blob/133344403bd0cb4a71d6b3014149e77ff4ceecab/index_hint.go

type WithHint struct {
	Keys []string
}

func (indexHint WithHint) ModifyStatement(stmt *gorm.Statement) {
	clause := stmt.Clauses["FROM"]

	if clause.AfterExpression == nil {
		clause.AfterExpression = indexHint
	} else {
		clause.AfterExpression = hints.Exprs{clause.AfterExpression, indexHint}
	}

	stmt.Clauses["FROM"] = clause
}

func (indexHint WithHint) Build(builder clause.Builder) {
	if len(indexHint.Keys) > 0 {
		builder.WriteString("WITH")
		builder.WriteByte('(')
		for idx, key := range indexHint.Keys {
			if idx > 0 {
				builder.WriteByte(',')
			}
			builder.WriteString(key)
		}
		builder.WriteByte(')')
	}
}

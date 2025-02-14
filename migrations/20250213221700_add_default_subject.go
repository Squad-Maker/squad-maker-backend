package migrations

import (
	"squad-maker/models"

	"github.com/ottomillrath/goose/v2"
	"gorm.io/gorm"
)

func init() {
	goose.AddMigration(service, upAddDefaultSubject, downAddDefaultSubject)
}

func upAddDefaultSubject(tx *gorm.DB) error {
	subject := &models.Subject{
		Name: "Fábrica de Software",
	}

	r := tx.Create(subject)

	return r.Error
}

func downAddDefaultSubject(tx *gorm.DB) error {
	r := tx.Where(models.Subject{
		Name: "Fábrica de Software",
	}, "Name").Delete(&models.Subject{})
	return r.Error
}

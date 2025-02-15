package migrations

import (
	"squad-maker/models"

	"github.com/ottomillrath/goose/v2"
	"gorm.io/gorm"
)

func init() {
	goose.AddMigration(service, upAddDefaultPositions, downAddDefaultPositions)
}

func upAddDefaultPositions(tx *gorm.DB) error {
	// assume-se que o subject criado na migração anterior é ID 1
	// se não for, alterar aqui ou mudar essa func pra pegar pelo nome

	positions := []*models.Position{
		{
			SubjectId: 1,
			Name:      "Backend",
		},
		{
			SubjectId: 1,
			Name:      "Frontend",
		},
		{
			SubjectId: 1,
			Name:      "QA",
		},
		{
			SubjectId: 1,
			Name:      "PM",
		},
		{
			SubjectId: 1,
			Name:      "Full Stack",
		},
		{
			SubjectId: 1,
			Name:      "UX/UI",
		},
	}

	for _, position := range positions {
		r := tx.Create(position)
		if r.Error != nil {
			return r.Error
		}
	}

	return nil
}

func downAddDefaultPositions(tx *gorm.DB) error {
	// também assume-se que o subject ID é 1
	return tx.Where(models.Position{
		SubjectId: 1,
	}, "SubjectId").Delete(&models.Position{}).Error
}

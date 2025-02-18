package migrations

import (
	"squad-maker/models"

	"github.com/ottomillrath/goose/v2"
	"gorm.io/gorm"
)

func init() {
	goose.AddMigration(service, upAddDefaultCompetenceLevels, downAddDefaultCompetenceLevels)
}

func upAddDefaultCompetenceLevels(tx *gorm.DB) error {
	// assume-se que o subject criado na migração anterior é ID 1
	// se não for, alterar aqui ou mudar essa func pra pegar pelo nome

	competenceLevels := []*models.CompetenceLevel{
		{
			SubjectId: 1,
			Name:      "Júnior",
		},
		{
			SubjectId: 1,
			Name:      "Pleno",
		},
		{
			SubjectId: 1,
			Name:      "Sênior",
		},
	}

	for _, competenceLevel := range competenceLevels {
		r := tx.Create(competenceLevel)
		if r.Error != nil {
			return r.Error
		}
	}

	return nil
}

func downAddDefaultCompetenceLevels(tx *gorm.DB) error {
	// também assume-se que o subject ID é 1
	return tx.Where(models.CompetenceLevel{
		SubjectId: 1,
	}, "SubjectId").Delete(&models.CompetenceLevel{}).Error
}

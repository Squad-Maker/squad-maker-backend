package migrations

import (
	pbAuth "squad-maker/generated/auth"
	"squad-maker/models"

	"github.com/ottomillrath/goose/v2"
	"gorm.io/gorm"
)

func init() {
	goose.AddMigration(service, upAddProfessorDummyUser, downAddProfessorDummyUser)
}

func upAddProfessorDummyUser(tx *gorm.DB) error {
	user := &models.User{
		Name:          "Professor Dummy",
		Type:          pbAuth.UserType_utProfessor,
		UtfprUsername: "dummy",
		Email:         "professor@example.com",
	}

	r := tx.Create(user)

	return r.Error
}

func downAddProfessorDummyUser(tx *gorm.DB) error {
	r := tx.Where(models.User{
		Email: "professor@example.com",
	}, "Email").Delete(&models.User{})
	return r.Error
}

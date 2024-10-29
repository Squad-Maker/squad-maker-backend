package migrations

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"squad-maker/database"
	"squad-maker/models"

	"github.com/ottomillrath/goose/v2"
)

// goose precisa de um path pra executar migrations do tipo sql
// essa versão modificada é baseada numa versão mais antiga do goose,
// que precisa da pasta ainda
// o executável ficará na pasta bin deste projeto, então a pasta migrations
// fica nesse path relativamente ao executável
const (
	migrationsPath = "../migrations"

	// essa versão modificada do goose foi feita pra suportar múltiplos
	// serviços utilizando um único banco, mas com suas próprias migrações
	// essa string aqui pode ser qualquer coisa
	service = "default"
)

func RunMigrations(ctx context.Context) error {
	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return err
	}

	err = dbCon.AutoMigrate(
		&models.User{},
		&models.Session{},
	)
	if err != nil {
		return err
	}

	// https://stackoverflow.com/a/18537419
	ex, err := os.Executable()
	if err != nil {
		return err
	}
	exPath := filepath.ToSlash(filepath.Dir(ex))
	fmt.Println("ex path", exPath)
	fmt.Println("migrations path", migrationsPath)

	err = goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	exPath = path.Join(exPath, migrationsPath)
	fmt.Println("migrations dir", exPath)
	return goose.Run("up", dbCon, service, filepath.FromSlash(exPath))
}

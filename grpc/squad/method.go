package squad

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"squad-maker/database"
	pbSquad "squad-maker/generated/squad"
	"squad-maker/models"

	"github.com/mroth/weightedrand/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// GenerateTeam gera um time pertencente a um subject.
//
// Se não informar project id (para time pré-cadastrado), um novo project será criado.
// Esta função somente gera um time/projeto por execução.
func (s *SquadServiceServer) GenerateTeam(ctx context.Context, req *pbSquad.GenerateTeamRequest) (*pbSquad.GenerateTeamResponse, error) {
	// req.SubjectId
	// *req.ProjectId
	// TODO req.OverrideConfigs (necessário quando não informar project id)
	// TODO req.weights

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var projectId int64
	err = dbCon.Transaction(func(tx *gorm.DB) error {
		// pega o subject
		// verifica se o usuário é dono do subject
		// se não for, not found
		// se houver id de projeto informado, tenta pegar o projeto
		// se não encontrar, not found
		// se não houver, cria um projeto
		// pega os dados de configuração de times/projetos do project
		// pega os dados de cargos já existentes no projeto
		// pega todos os estudantes do subject que ainda não estão em um projeto
		// repassa tudo pra função de geração e a mágica vai ocorrer lá

		subject := &models.Subject{}
		// TODO aqui sempre carrega todos os estudantes do subject, mesmo que já estejam no project
		// e mesmo que o usuário esteja deletado (o que provavelmente não vai ocorrer muito)
		// se necessário, separar o carregamento dos students, pegando somente os que realmente existem e que não estão no projeto
		// (às vezes, soft delete só complica... mas geralmente é necessário para auditoria)
		// TODO algum tipo de lock
		r := tx.
			Preload("Students").
			Preload("Students.Student").
			First(subject, req.SubjectId)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "subject not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		// TODO quando for feito pra permitir mais de uma disciplina/subject, vai ter
		// uma tabela com quais professores possuem acesso a quais disciplinas/subject
		// verificar aqui

		project := &models.Project{}
		if req.ProjectId != nil {
			r := tx.
				Preload("Positions").
				Preload("Positions.Position").
				Preload("Students").
				Preload("Students.Student").
				First(project, *req.ProjectId)
			if r.Error != nil {
				if errors.Is(r.Error, gorm.ErrRecordNotFound) {
					return status.Error(codes.NotFound, "project not found")
				}
				return status.Error(codes.Internal, r.Error.Error())
			}
		} else {
			project = &models.Project{
				SubjectId: req.SubjectId,
				Name:      "Novo time/projeto",
			}
			r := tx.Create(project)
			if r.Error != nil {
				return status.Error(codes.Internal, r.Error.Error())
			}

			// TODO se cair aqui, vai ter o override sempre preenchido
			// salvar como config do projeto e carregar do banco pra repassar depois pra função de geração
		}
		projectId = project.Id

		mapNewStudents, err := generateProjectTeam(subject, project)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// adiciona os estudantes ao projeto
		for student, position := range mapNewStudents {
			ps := &models.ProjectStudent{
				ProjectId:  project.Id,
				StudentId:  student.StudentId,
				PositionId: position.Id,
			}

			r := tx.Create(ps)
			if r.Error != nil {
				return status.Error(codes.Internal, r.Error.Error())
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.GenerateTeamResponse{
		ProjectId: projectId,
	}, nil
}

// TODO mover isso pra uma interface
// permitindo fazer implementações diferentes com as mesmas entradas e saídas
// por exemplo, permitindo implementar com alguma IA/LLM
func generateProjectTeam(subject *models.Subject, project *models.Project) (map[*models.StudentSubjectData]*models.Position, error) {
	// considera que subject.Students está carregado e que o Student dentro de Students existe e não é nil
	// mesma coisa para o project.Students e project.Positions

	// TODO considerar também em quantos projetos o student já está (dentro do subject)
	// mapeia os estudantes que não estão em nenhum projeto, considerando os projetos favoritos dos estudantes como peso para seleção

	possibleStudents := map[*models.StudentSubjectData]weightedrand.Choice[*models.StudentSubjectData, int64]{}
	for _, student := range subject.Students {
		if student.Student == nil {
			// ignora usuários deletados
			continue
		}

		// verifica se já está no projeto
		alreadyInProject := false
		for _, p := range project.Students {
			if p.StudentId == student.StudentId {
				alreadyInProject = true
				break
			}
		}
		if alreadyInProject {
			continue
		}

		var weight int64
		weight = 5
		if student.PreferredProjectId != nil {
			// prioriza os estudantes que escolheram o projeto
			if *student.PreferredProjectId == project.Id {
				weight += 10 // TODO esse tipo de config deve ser parametrizada
			} else {
				weight -= 4
			}
		}
		possibleStudents[student] = weightedrand.NewChoice(student, weight)
	}

	if len(possibleStudents) == 0 {
		// todos os estudantes já estão em um projeto
		return nil, fmt.Errorf("all students are already in a project")
	}

	// mapeia os cargos/positions que ainda precisam ser preenchidos
	mapPossiblePositionsCount := map[*models.ProjectPosition]int64{}
	for _, position := range project.Positions {
		if position.Position == nil {
			// ignora cargos deletados
			continue
		}

		// verifica se já está preenchido
		var countFilled int64
		for _, ps := range project.Students {
			if ps.Student == nil {
				continue
			}

			if ps.PositionId == position.PositionId {
				countFilled++
			}

			if countFilled >= position.Count {
				break
			}
		}

		if countFilled < position.Count {
			mapPossiblePositionsCount[position] = position.Count - countFilled
		}
	}

	if len(mapPossiblePositionsCount) == 0 {
		// todos os cargos já estão preenchidos
		return nil, fmt.Errorf("all positions are already filled")
	}

	mapSelectedStudentToPosition := map[*models.StudentSubjectData]*models.Position{}
	// para cada cargo/position que ainda falta preencher (conforme configuração), seleciona um student para preencher o cargo
	for position, count := range mapPossiblePositionsCount {
		// copia a lista de students para modificar os pesos conforme cargo
		positionWeightedStudents := maps.Clone(possibleStudents)
		for i, student := range positionWeightedStudents {
			if student.Item.PositionOption1Id == position.PositionId {
				student.Weight += 5
			} else if student.Item.PositionOption2Id != nil && *student.Item.PositionOption2Id == position.PositionId {
				student.Weight += 3
			}
			positionWeightedStudents[i] = student
		}

		for i := int64(0); i < count; i++ {
			if len(positionWeightedStudents) == 0 {
				// não tem mais estudantes para preencher o cargo
				break
			}

			chooser, err := weightedrand.NewChooser(slices.Collect(maps.Values(positionWeightedStudents))...)
			if err != nil {
				return nil, err
			}
			student := chooser.Pick()

			mapSelectedStudentToPosition[student] = position.Position
			delete(possibleStudents, student)
			delete(positionWeightedStudents, student)
		}
	}

	// TODO retornar alguma coisa pra indicar balanceamento de conhecimento do time
	// baseado nas infos de nível de competência dos estudantes escolhidos
	// provavelmente usar isso pra gerar N times e escolher o mais balanceado

	return mapSelectedStudentToPosition, nil
}

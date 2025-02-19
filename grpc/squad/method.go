package squad

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"squad-maker/database"
	pbAuth "squad-maker/generated/auth"
	pbSquad "squad-maker/generated/squad"
	"squad-maker/models"
	grpcUtils "squad-maker/utils/grpc"
	mailUtils "squad-maker/utils/mail"
	otherUtils "squad-maker/utils/other"
	"strings"

	"github.com/mroth/weightedrand/v2"
	mail "github.com/xhit/go-simple-mail/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// TODO quando implementar ownership do subject, tem que validar em tudo

var (
	ErrAllStudentsAlreadyInProject = errors.New("all students are already in a project")
	ErrAllPositionsAlreadyFilled   = errors.New("all positions are already filled")
)

func (s *SquadServiceServer) GenerateAllTeams(ctx context.Context, req *pbSquad.GenerateAllTeamsRequest) (*pbSquad.GenerateAllTeamsResponse, error) {
	subjectId := grpcUtils.GetCurrentSubjectIdFromMetadata(ctx)

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = dbCon.Transaction(func(tx *gorm.DB) error {
		// carrega todos os projetos do subject que ainda possuem cargos vagos
		// não precisa validar nada do subject aqui pois vai validar na outra função
		var projectIds []int64
		//@raw_sql
		r := tx.Raw(`
			with needed as (select ppo_project_id as project_id, ppo_position_id as position_id, ppo_count as c from project_positions group by ppo_project_id, ppo_position_id),
			filled as (select pst_project_id as project_id, pst_position_id as position_id, count(*) as c from project_students group by pst_project_id, pst_position_id)
			select distinct pro_id from projects
			left join needed on needed.project_id = pro_id
			left join filled on filled.project_id = pro_id
			where pro_deleted_at is null
			and pro_subject_id = ?
			and coalesce(needed.c, 0) - coalesce(filled.c, 0) > 0
			order by pro_id
		`, subjectId).Scan(&projectIds)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		// TODO fazer passar no request uma config de quantidade de cargos
		// para utilizar caso sobre alunos, para criar projetos novos
		for _, projectId := range projectIds {
			_, err := s.generateTeam(ctx, &pbSquad.GenerateTeamRequest{
				ProjectInfo: &pbSquad.GenerateTeamRequest_ProjectId{
					ProjectId: projectId,
				},
			}, tx)
			if err != nil {
				if strings.Contains(err.Error(), ErrAllStudentsAlreadyInProject.Error()) {
					break
				} else if strings.Contains(err.Error(), ErrAllPositionsAlreadyFilled.Error()) {
					continue
				} else {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.GenerateAllTeamsResponse{}, nil
}

// GenerateTeam gera um time pertencente a um subject.
//
// Se não informar project id (para time pré-cadastrado), um novo project será criado.
// Esta função somente gera um time/projeto por execução.
func (s *SquadServiceServer) GenerateTeam(ctx context.Context, req *pbSquad.GenerateTeamRequest) (*pbSquad.GenerateTeamResponse, error) {
	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.generateTeam(ctx, req, dbCon)
}

func (s *SquadServiceServer) generateTeam(ctx context.Context, req *pbSquad.GenerateTeamRequest, dbCon *gorm.DB) (*pbSquad.GenerateTeamResponse, error) {
	// subjectId pega do metadata
	// *req.ProjectId
	// TODO req.OverrideConfigs (necessário quando não informar project id ou info de projeto)
	// TODO req.weights
	subjectId := grpcUtils.GetCurrentSubjectIdFromMetadata(ctx)

	var projectId int64
	err := dbCon.Transaction(func(tx *gorm.DB) error {
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
		// TODO algum tipo de lock
		r := tx.
			Preload("Students", `NOT EXISTS (
				SELECT 1
				FROM project_students
				JOIN projects ON pro_id = pst_project_id AND pro_deleted_at IS NULL
				WHERE pst_student_id = ssd_student_id
			)`).
			Preload("Students.Student").
			First(subject, subjectId)
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
		switch c := req.ProjectInfo.(type) {
		case *pbSquad.GenerateTeamRequest_ProjectId:
			r := tx.
				Preload("Positions").
				Preload("Positions.Position").
				Preload("Students").
				Preload("Students.Student").
				First(project, c.ProjectId)
			if r.Error != nil {
				if errors.Is(r.Error, gorm.ErrRecordNotFound) {
					return status.Error(codes.NotFound, "project not found")
				}
				return status.Error(codes.Internal, r.Error.Error())
			}
		case *pbSquad.GenerateTeamRequest_NewProject:
			project = &models.Project{
				SubjectId:   subjectId,
				Name:        c.NewProject.Name,
				Description: c.NewProject.Description,
			}
			r := tx.Create(project)
			if r.Error != nil {
				return status.Error(codes.Internal, r.Error.Error())
			}

			for _, position := range c.NewProject.Positions {
				pp := &models.ProjectPosition{
					ProjectId:  project.Id,
					PositionId: position.Id,
					Count:      position.Count,
				}

				r := tx.Create(pp)
				if r.Error != nil {
					return status.Error(codes.Internal, r.Error.Error())
				}
			}

			r = tx.
				Preload("Positions").
				Preload("Positions.Position").
				Preload("Students").
				Preload("Students.Student").
				First(project, project.Id)
			if r.Error != nil {
				if errors.Is(r.Error, gorm.ErrRecordNotFound) {
					return status.Error(codes.NotFound, "project not found")
				}
				return status.Error(codes.Internal, r.Error.Error())
			}
		default:
			return status.Error(codes.Unimplemented, "generating team without project parameters is not implemented yet (overrides not implemented)")

			// project = &models.Project{
			// 	SubjectId: subjectId,
			// 	Name:      "Novo time/projeto",
			// }
			// r := tx.Create(project)
			// if r.Error != nil {
			// 	return status.Error(codes.Internal, r.Error.Error())
			// }

			// // TODO se cair aqui, vai ter o override sempre preenchido
			// // salvar como config do projeto e carregar do banco pra repassar depois pra função de geração
		}
		projectId = project.Id

		// carrega os níveis de competência para utilizar no balanceamento
		var competenceLevels []*models.CompetenceLevel
		r = tx.Where(models.CompetenceLevel{
			SubjectId: subjectId,
		}, "SubjectId").Find(&competenceLevels)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		var mapNewStudents map[*models.StudentSubjectData]*models.Position
		var currentDiff uint64
		targetDiff := uint64(4) // TODO parametrizar (via request)
		for i := 0; i < 100; i++ {
			generatedTeam, err := generateProjectTeam(subject, project)
			if err != nil {
				if errors.Is(err, ErrAllStudentsAlreadyInProject) || errors.Is(err, ErrAllPositionsAlreadyFilled) {
					return status.Error(codes.AlreadyExists, err.Error())
				}
				return status.Error(codes.Internal, err.Error())
			}

			diff := avaliarBalanceamento(generatedTeam, competenceLevels)
			if math.Abs(float64(diff-targetDiff)) < math.Abs(float64(currentDiff-targetDiff)) || i == 0 {
				currentDiff = diff
				mapNewStudents = generatedTeam
			}

			if diff == targetDiff {
				break
			}
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

func (s *SquadServiceServer) AddStudentToTeam(ctx context.Context, req *pbSquad.AddStudentToTeamRequest) (*pbSquad.AddStudentToTeamResponse, error) {
	subjectId := grpcUtils.GetCurrentSubjectIdFromMetadata(ctx)

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = dbCon.Transaction(func(tx *gorm.DB) error {
		// pega o subject
		// faz lock for update
		// pega o estudante
		// pega o projeto
		// pega o cargo
		// adiciona o estudante ao projeto
		subject := &models.Subject{}
		r := tx.Clauses(database.GetLockForUpdateClause(tx.Dialector.Name(), false)).
			First(subject, subjectId)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "subject not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		student := &models.StudentSubjectData{}
		r = tx.Where(models.StudentSubjectData{
			StudentId: req.StudentId,
			SubjectId: subjectId,
		}, "StudentId", "SubjectId").First(student)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "student not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		project := &models.Project{}
		r = tx.First(project, req.ProjectId)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "project not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		position := &models.Position{}
		r = tx.First(position, req.PositionId)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "position not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		ps := &models.ProjectStudent{
			ProjectId:  project.Id,
			StudentId:  student.StudentId,
			PositionId: position.Id,
		}
		r = tx.Save(ps)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.AddStudentToTeamResponse{}, nil
}

func (s *SquadServiceServer) RemoveStudentFromTeam(ctx context.Context, req *pbSquad.RemoveStudentFromTeamRequest) (*pbSquad.RemoveStudentFromTeamResponse, error) {
	subjectId := grpcUtils.GetCurrentSubjectIdFromMetadata(ctx)

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = dbCon.Transaction(func(tx *gorm.DB) error {
		// pega o subject
		// faz lock for update
		// pega o estudante
		// pega o projeto
		// remove o estudante do projeto
		subject := &models.Subject{}
		r := tx.Clauses(database.GetLockForUpdateClause(tx.Dialector.Name(), false)).
			First(subject, subjectId)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "subject not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		student := &models.StudentSubjectData{}
		r = tx.Where(models.StudentSubjectData{
			StudentId: req.StudentId,
			SubjectId: subjectId,
		}, "StudentId", "SubjectId").First(student)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "student not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		project := &models.Project{}
		r = tx.First(project, req.ProjectId)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "project not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		ps := &models.ProjectStudent{
			ProjectId: project.Id,
			StudentId: student.StudentId,
		}
		r = tx.Delete(ps)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pbSquad.RemoveStudentFromTeamResponse{}, nil
}

func (s *SquadServiceServer) RequestTeamRevaluation(ctx context.Context, req *pbSquad.RequestTeamRevaluationRequest) (*pbSquad.RequestTeamRevaluationResponse, error) {
	if req.ProjectId == 0 {
		return nil, status.Error(codes.InvalidArgument, "project id cannot be zero")
	}

	if req.Reason == "" {
		return nil, status.Error(codes.InvalidArgument, "reason cannot be empty")
	}

	studentId := grpcUtils.GetCurrentUserIdFromMetadata(ctx)
	subjectId := grpcUtils.GetCurrentSubjectIdFromMetadata(ctx)

	dbCon, err := database.GetConnectionWithContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var emailsToSend []*mail.Email
	err = dbCon.Transaction(func(tx *gorm.DB) error {
		// verifica se o subject realmente existe
		subject := &models.Subject{}
		r := tx.First(subject, subjectId)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "subject not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		// verifica se o estudante realmente está no projeto
		projectStudent := &models.ProjectStudent{}
		r = tx.InnerJoins("Project").InnerJoins("Student").Where(models.ProjectStudent{
			ProjectId: req.ProjectId,
			StudentId: studentId,
		}).First(projectStudent)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return status.Error(codes.NotFound, "project not found")
			}
			return status.Error(codes.Internal, r.Error.Error())
		}

		var desiredProject *models.Project
		// se o estudante selecionou um projeto preferido, verifica se o mesmo existe
		if req.DesiredProjectId != nil {
			desiredProject = &models.Project{}
			r = tx.First(desiredProject, *req.DesiredProjectId)
			if r.Error != nil {
				if errors.Is(r.Error, gorm.ErrRecordNotFound) {
					return status.Error(codes.NotFound, "desired project not found")
				}
				return status.Error(codes.Internal, r.Error.Error())
			}
		}

		// envia o email
		// TODO como não tem implementação de owner de subject ainda, envia para TODOS os professores
		var professors []*models.User
		r = tx.Where(models.User{
			Type: pbAuth.UserType_utProfessor,
		}, "Type").Find(&professors)
		if r.Error != nil {
			return status.Error(codes.Internal, r.Error.Error())
		}

		// TODO fazer com html template ou outra coisa que faça escape dos inputs pra não ter html injection
		for _, professor := range professors {
			body := fmt.Sprintf(`<p>Prezado(a) Professor(a) %s,</p>
				<p>Informamos que o aluno <strong>%s</strong> fez uma solicitação de mudança de projeto na disciplina <strong>%s</strong>.</p>
				<br>
				<strong>Projeto atual:</strong>
				<p>%s</p>
				<br>
				<strong>Motivo informado pelo aluno:</strong>
				<p>%s</p>
				<br>`, professor.Name, projectStudent.Student.Name, subject.Name, projectStudent.Project.Name, req.Reason)
			if desiredProject != nil {
				body += fmt.Sprintf(`<strong>Projeto desejado:</strong>
					<p>%s</p>`, desiredProject.Name)
			}

			body += `<br>
				<p>Atenciosamente,</p>
				<p>Sistema Squad Maker</p>`

			emailsToSend = append(emailsToSend, mailUtils.PrepareNewMail(professor.Name, professor.Email, "Solicitação de mudança de projeto - "+projectStudent.Student.Name, body, mail.TextHTML))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	client, err := mailUtils.GetNewSmtpClient()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, email := range emailsToSend {
		err = email.Send(client)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pbSquad.RequestTeamRevaluationResponse{}, nil
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
		return nil, ErrAllStudentsAlreadyInProject
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
		return nil, ErrAllPositionsAlreadyFilled
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

	return mapSelectedStudentToPosition, nil
}

func avaliarBalanceamento(time map[*models.StudentSubjectData]*models.Position, senioridadesPossiveis []*models.CompetenceLevel) uint64 {
	// reimplementação da função que estava no primeiro projeto, que era em python
	// mas utilizando o peso das senioridades ao invés de senioridades fixas
	// também só faz de 1 time por vez
	// deixei até em português para indicar que é uma reimplementação, não algo pensado por nós

	// outra coisa que vou mudar é a forma com que os times são comparados
	// serão gerados de forma aleatória e comparados entre si
	// além disso, o 'melhor time' não é o que tem melhor balanceamento (ex: todos sêniors), mas sim que têm balanceamento em um número arbitário
	// atualmente fixo no código

	// faz uma combinação 'n escolhe 2' das senioridades
	// calcula a diferença absoluta entre a quantidade de senioridades
	// (eu tinha uma ideia de usar 'pesos', mas nem vai precisar, já que usa o count...)
	// este algoritmo considera como se cada senioridade fosse igualmente importante, fazendo com que senioridades não utilizadas contem muito no diffAbs

	// talvez tenha alguma forma melhor de otimizar este código...

	var diffAbs uint64
	iter := &otherUtils.Iterator{N: len(senioridadesPossiveis), K: 2}
	for iter.Next() {
		senioridade1 := senioridadesPossiveis[iter.Comb[0]]
		senioridade2 := senioridadesPossiveis[iter.Comb[1]]

		var count1, count2 uint64

		for ssd := range time {
			if ssd.CompetenceLevelId == senioridade1.Id {
				count1++
			} else if ssd.CompetenceLevelId == senioridade2.Id {
				count2++
			}
		}

		diffAbs += uint64(math.Abs(float64(count1 - count2)))
	}

	return diffAbs
}

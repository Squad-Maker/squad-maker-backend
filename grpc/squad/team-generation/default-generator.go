package teamGeneration

import (
	"maps"
	"math"
	"slices"
	"squad-maker/models"
	otherUtils "squad-maker/utils/other"

	"github.com/mroth/weightedrand/v2"
)

type DefaultGenerator struct{}

func (g *DefaultGenerator) GenerateTeam(subject *models.Subject, project *models.Project) (map[*models.StudentSubjectData]*models.Position, error) {
	// considera que subject.Students está carregado e que o Student dentro de Students existe e não é nil
	// mesma coisa para o project.Students e project.Positions

	// os students dentro de subject.Students são os considerados para preenchimento no projeto

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

func (g *DefaultGenerator) CalculateBalancingScore(time map[*models.StudentSubjectData]*models.Position, senioridadesPossiveis []*models.CompetenceLevel) uint64 {
	// reimplementação da função que estava no primeiro projeto, que era em python
	// mas utilizando o peso das senioridades ao invés de senioridades fixas
	// também só faz de 1 time por vez

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

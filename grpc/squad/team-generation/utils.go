package teamGeneration

import (
	"squad-maker/models"
)

func getPossibleStudents(subject *models.Subject, project *models.Project) []*models.StudentSubjectData {
	var possibleStudents []*models.StudentSubjectData
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

		possibleStudents = append(possibleStudents, student)
	}
	return possibleStudents
}

func getPossiblePositionsCount(project *models.Project) map[*models.ProjectPosition]int64 {
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
	return mapPossiblePositionsCount
}

func getPossibleCompetenceLevelsCount(possibleStudents []*models.StudentSubjectData, project *models.Project) map[*models.ProjectCompetenceLevel]int64 {
	mapPossibleCompetenceLevelsCount := map[*models.ProjectCompetenceLevel]int64{}
	for _, competenceLevel := range project.CompetenceLevels {
		if competenceLevel.CompetenceLevel == nil {
			// ignora cargos deletados
			continue
		}

		// verifica se já está preenchido
		var countFilled int64
		for _, ps := range project.Students {
			if ps.Student == nil {
				continue
			}

			var student *models.StudentSubjectData
			for _, s := range possibleStudents {
				if s.StudentId == ps.StudentId {
					student = s
					break
				}
			}
			if student == nil {
				continue
			}

			if student.CompetenceLevelId == competenceLevel.CompetenceLevelId {
				countFilled++
			}

			if countFilled >= competenceLevel.Count {
				break
			}
		}

		if countFilled < competenceLevel.Count {
			mapPossibleCompetenceLevelsCount[competenceLevel] = competenceLevel.Count - countFilled
		}
	}
	return mapPossibleCompetenceLevelsCount
}

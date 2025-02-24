package teamGeneration

import (
	"errors"
	pbSquad "squad-maker/generated/squad"
	"squad-maker/models"
)

var (
	ErrAllStudentsAlreadyInProject = errors.New("all students are already in a project")
	ErrAllPositionsAlreadyFilled   = errors.New("all positions are already filled")
)

type TeamGenerator interface {
	// GenerateTeam generates a team for a given subject and project.
	//
	// It expects that the students to be considered are loaded in the subject
	// and the current positions and students are loaded in the project.
	//
	// Returns a map with the students and their respective positions to be added to the project.
	GenerateTeam(subject *models.Subject, project *models.Project) (map[*models.StudentSubjectData]*models.Position, error)

	// CalculateBalancingScore calculates the balancing score for a given team.
	// This number is used to know how balanced the team is and is compared against a target score.
	// The closest it is to the target score, the better the team is balanced.
	//
	// Returns the calculated score.
	CalculateBalancingScore(team map[*models.StudentSubjectData]*models.Position, allCompetenceLevelsInSubject []*models.CompetenceLevel) uint64
}

func GetTeamGenerator(generatorType pbSquad.TeamGeneratorType) TeamGenerator {
	switch generatorType {
	case pbSquad.TeamGeneratorType_tgtDefault:
		return &DefaultGenerator{}
	default:
		return &DefaultGenerator{}
	}
}

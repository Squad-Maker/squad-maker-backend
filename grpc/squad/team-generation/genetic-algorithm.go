package teamGeneration

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"math/rand"
	"slices"
	"sort"
	"squad-maker/models"
	otherUtils "squad-maker/utils/other"
	sliceUtils "squad-maker/utils/slice"
	"strconv"
	"strings"
	"time"
)

// https://dev.to/stefanalfbo/genetic-algorithms-with-go-1b5b

type GeneticAlgorithmGenerator struct {
	PopulationSize int
	Generations    int
	MaxRetries     int

	currentProjectId                 int64
	studentsInTheTeam                int64
	possibleStudents                 []*models.StudentSubjectData
	mapPossiblePositionsCount        map[*models.ProjectPosition]int64
	mapPossibleCompetenceLevelsCount map[*models.ProjectCompetenceLevel]int64
	tools                            []string
}

// representa um estudante
// é o índice do estudante no slice de estudantes possíveis
type gene int64

// representa um time
type individual struct {
	chromosome       []gene
	fitness          float64
	mappedToPosition map[*models.StudentSubjectData]*models.Position // mapeia o estudante para o cargo que ele vai ocupar
}

func (i *individual) String() string {
	// retorna uma representação do cromossomo como uma string
	// e o fitness do indivíduo
	genes := make([]string, len(i.chromosome))
	for j, g := range i.chromosome {
		genes[j] = strconv.FormatInt(int64(g), 10)
	}
	return "Genes: [" + strings.Join(genes, ", ") + "], Fitness: " + strconv.FormatFloat(i.fitness, 'f', 4, 64)
}

func (i *individual) PrintLog() {
	fmt.Println(i.String())
}

// o mapeamento para os cargos já é feito aqui, mesmo não seguindo o princípio da responsabilidade única
// pois vai evitar processamento desnecessário
func (i *individual) calculateFitness(ga *GeneticAlgorithmGenerator) float64 {
	// quanto menor for o fitness, melhor é o time
	// poderá ser menor que zero caso os alunos tiverem selecionado o projeto como preferido

	i.mappedToPosition = make(map[*models.StudentSubjectData]*models.Position)

	// inclui tudo o que é necessário e vai diminuindo conforme encontrar
	fitness := float64(ga.studentsInTheTeam)
	for _, c := range ga.mapPossibleCompetenceLevelsCount {
		fitness += float64(c)
	}
	fitness += float64(len(ga.tools))

	tools := slices.Clone(ga.tools)

	mapPositionsToFill := maps.Clone(ga.mapPossiblePositionsCount)
	mapCompetenceLevelsToFill := maps.Clone(ga.mapPossibleCompetenceLevelsCount)
	var studentsToAssignRandomlyLater []*models.StudentSubjectData

	for _, g := range i.chromosome {
		student := ga.possibleStudents[g]
		if student == nil {
			continue // ignora se o gene não tiver um estudante válido
		}

		// considera as duas preferências de cargo do estudante
		var positionOption1, positionOption2 *models.ProjectPosition
		for position := range mapPositionsToFill {
			if position.PositionId == student.PositionOption1Id {
				positionOption1 = position
			} else if student.PositionOption2Id != nil && position.PositionId == *student.PositionOption2Id {
				positionOption2 = position
			}

			if positionOption1 != nil {
				// se encontrou a primeira opção, não precisa continuar procurando
				break
			}
		}

		if positionOption1 != nil && mapPositionsToFill[positionOption1] > 0 {
			mapPositionsToFill[positionOption1]--
			fitness--
			i.mappedToPosition[student] = positionOption1.Position
		} else if positionOption2 != nil && mapPositionsToFill[positionOption2] > 0 {
			mapPositionsToFill[positionOption2]--
			fitness -= 0.5 // se preencheu a segunda opção, considera que o fitness é menor, mas ainda assim preencheu um cargo
			i.mappedToPosition[student] = positionOption2.Position
		} else {
			studentsToAssignRandomlyLater = append(studentsToAssignRandomlyLater, student)
			continue
		}

		// verifica os níveis de competência
		for competenceLevel := range mapCompetenceLevelsToFill {
			if competenceLevel.CompetenceLevelId == student.CompetenceLevelId && mapCompetenceLevelsToFill[competenceLevel] > 0 {
				mapCompetenceLevelsToFill[competenceLevel]--
				fitness--
				break
			}
		}

		studentTools := make([]string, 0, len(student.Tools))
		for _, tool := range student.Tools {
			studentTools = append(studentTools, strings.ToLower(tool))
		}
		studentTools = sliceUtils.RemoveDuplicates(studentTools)

		lenTools := len(tools)
		tools = sliceUtils.Difference(ga.tools, studentTools)
		fitness -= float64(lenTools - len(tools))

		if student.PreferredProjectId != nil {
			if *student.PreferredProjectId == ga.currentProjectId {
				fitness -= 0.2
			} else {
				// se o estudante escolheu outro projeto como preferido, aumenta o fitness
				fitness += 0.1
			}
		}
	}

	for _, student := range studentsToAssignRandomlyLater {
		positionFound := false
		for position := range mapPositionsToFill {
			if mapPositionsToFill[position] > 0 {
				mapPositionsToFill[position]--
				i.mappedToPosition[student] = position.Position
				positionFound = true
				break
			}
		}
		if !positionFound {
			// se não encontrou nenhum cargo, ignora o estudante
			continue
		}

		// verifica os níveis de competência
		for competenceLevel := range mapCompetenceLevelsToFill {
			if competenceLevel.CompetenceLevelId == student.CompetenceLevelId && mapCompetenceLevelsToFill[competenceLevel] > 0 {
				mapCompetenceLevelsToFill[competenceLevel]--
				fitness--
				break
			}
		}

		studentTools := make([]string, 0, len(student.Tools))
		for _, tool := range student.Tools {
			studentTools = append(studentTools, strings.ToLower(tool))
		}
		studentTools = sliceUtils.RemoveDuplicates(studentTools)

		lenTools := len(tools)
		tools = sliceUtils.Difference(ga.tools, studentTools)
		fitness -= float64(lenTools - len(tools))
	}

	fitness = math.Round(fitness*10000) / 10000 // arredonda para 4 casas decimais
	i.fitness = fitness

	return fitness
}

func (ga *GeneticAlgorithmGenerator) newChromosome(size int64) []gene {
	chromosome := make([]gene, 0, size)
	randSource := &otherUtils.MapRand[gene]{}
	randSource.InitializeInterval(0, gene(len(ga.possibleStudents)-1))
	for range size {
		newGene, ok := randSource.GetRandomAndPop()
		if !ok {
			// se não conseguir gerar um gene, retorna o cromossomo atual
			// geração parcial
			return chromosome
		}

		for slices.Contains(chromosome, newGene) {
			// garante que o gene é único no cromossomo
			newGene, ok = randSource.GetRandomAndPop()
			if !ok {
				// se não conseguir gerar um gene, retorna o cromossomo atual
				// geração parcial
				return chromosome
			}
		}
		chromosome = append(chromosome, newGene)
	}
	return chromosome
}

func (ga *GeneticAlgorithmGenerator) newPopulation() []*individual {
	population := make([]*individual, ga.PopulationSize)
	for i := range ga.PopulationSize {
		chromosome := ga.newChromosome(ga.studentsInTheTeam)
		ind := &individual{chromosome: chromosome}
		ind.calculateFitness(ga)
		population[i] = ind
	}
	return population
}

// orderCrossoverHelper é uma função auxiliar que realiza a lógica principal do OX1
// para criar um único filho a partir de dois pais e pontos de corte definidos.
func (ga *GeneticAlgorithmGenerator) orderCrossoverHelper(p1, p2 *individual, start, end int) *individual {
	geneCount := len(p1.chromosome)
	// Inicializa o slice de genes do filho com zeros (ou valor nulo).
	childGenes := make([]gene, geneCount)

	// Mapa para rastrear os genes que já foram adicionados ao filho.
	// Usar um mapa oferece performance O(1) para verificação, o que é muito eficiente.
	genesInChild := make(map[gene]bool)

	// Passo 1: Copiar a subsequência do pai 1 para o filho.
	for i := start; i <= end; i++ {
		gene := p1.chromosome[i]
		childGenes[i] = gene
		genesInChild[gene] = true
	}

	// Passo 2: Preencher os genes restantes a partir do pai 2.
	childInsertIndex := (end + 1) % geneCount

	// Inicia a busca no pai 2 a partir do ponto de corte final.
	for i := range geneCount {
		p2GeneIndex := (end + 1 + i) % geneCount
		geneFromP2 := p2.chromosome[p2GeneIndex]

		// Se o gene do pai 2 AINDA NÃO estiver no filho...
		if !genesInChild[geneFromP2] {
			// Adiciona o gene no próximo local de inserção disponível.
			childGenes[childInsertIndex] = geneFromP2
			// Move o ponteiro de inserção do filho para a próxima posição.
			childInsertIndex = (childInsertIndex + 1) % geneCount
		}
	}

	return &individual{chromosome: childGenes}
}

// OrderCrossover realiza o crossover de ordem (OX1) entre dois pais.
// Ele retorna dois filhos válidos.
func (ga *GeneticAlgorithmGenerator) orderCrossover(parent1, parent2 *individual) (*individual /* , *individual */, error) {
	if len(parent1.chromosome) != len(parent2.chromosome) {
		return &individual{} /* &individual{}, */, errors.New("os pais devem ter o mesmo número de genes")
	}

	geneCount := len(parent1.chromosome)
	if geneCount < 2 {
		// Não é possível fazer crossover se não houver pelo menos 2 genes.
		// Retorna os pais originais.
		return parent1 /*, parent2 */, nil
	}

	// Para garantir aleatoriedade real a cada execução.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Gera dois pontos de corte aleatórios e garante que start < end.
	cutPoint1 := r.Intn(geneCount)
	cutPoint2 := r.Intn(geneCount)

	start := min(cutPoint1, cutPoint2)
	end := max(cutPoint1, cutPoint2)

	// Se os pontos forem iguais, o filho seria uma cópia exata do pai.
	// Para um crossover mais significativo, garantimos que a janela tenha tamanho de pelo menos 1.
	if start == end {
		end = (end + 1) % geneCount
		if start > end {
			start, end = end, start // mantém start < end se o wrap around acontecer
		}
	}

	// fmt.Printf("Pontos de corte aleatórios: Início=%d, Fim=%d\n", start, end)
	// fmt.Println("-----------------------------------------------------")

	// Gera os dois filhos.
	// Filho 1 herda a subsequência do Pai 1 e a ordem do Pai 2.
	child1 := ga.orderCrossoverHelper(parent1, parent2, start, end)
	// // Filho 2 herda a subsequência do Pai 2 e a ordem do Pai 1.
	// child2 := ga.orderCrossoverHelper(parent2, parent1, start, end)

	return child1 /* child2, */, nil
}

func (ga *GeneticAlgorithmGenerator) crossover(population []*individual) *individual {
	topHalf := population[:len(population)/2]
	parent1 := topHalf[rand.Intn(len(topHalf))]
	parent2 := topHalf[rand.Intn(len(topHalf))]

	i, err := ga.orderCrossover(parent1, parent2)
	if err != nil {
		panic(err)
	}
	i.calculateFitness(ga)

	return i
}

func (ga *GeneticAlgorithmGenerator) maybeMutate(i *individual) *individual {
	chromosome := make([]gene, len(i.chromosome))
	copy(chromosome, i.chromosome)

	mutationRate := 0.01 // 1% chance to mutate each gene

	randSource := &otherUtils.MapRand[gene]{}
	randSource.InitializeInterval(0, gene(len(ga.possibleStudents)-1))

	for index := range chromosome {
		if rand.Float64() < mutationRate {
			newGene, ok := randSource.GetRandomAndPop()
			if !ok {
				// se não conseguir gerar um gene, retorna o cromossomo atual
				// mutação impossível
				return i
			}

			for slices.ContainsFunc(chromosome, func(g gene) bool { return g == newGene && g != chromosome[index] }) {
				newGene, ok = randSource.GetRandomAndPop()
				if !ok {
					// se não conseguir gerar um gene, retorna o cromossomo atual
					// mutação impossível
					return i
				}
			}
			chromosome[index] = newGene
		}
	}

	ind := &individual{chromosome: chromosome}
	ind.calculateFitness(ga)

	return ind
}

func (ga *GeneticAlgorithmGenerator) evolveGeneration(population []*individual) []*individual {
	eliteSize := len(population) / 10 // 10% of the population

	sort.Slice(population, func(i, j int) bool {
		return population[i].fitness < population[j].fitness
	})

	newPopulation := make([]*individual, 0, len(population))
	newPopulation = append(newPopulation, population[:eliteSize]...)

	for i := eliteSize; i < len(population); i++ {
		newPopulation = append(newPopulation, ga.maybeMutate(ga.crossover(population)))
	}

	sort.Slice(newPopulation, func(i, j int) bool {
		return newPopulation[i].fitness < newPopulation[j].fitness
	})

	return newPopulation
}

func (ga *GeneticAlgorithmGenerator) evolve(population []*individual) (bestIndividual *individual) {
	newPopulation := make([]*individual, 0, len(population))

	for _, ind := range population {
		newIndividual := &individual{}
		*newIndividual = *ind // copia o indivíduo atual
		// newIndividual.PrintLog()
		newPopulation = append(newPopulation, newIndividual)
	}

	lastFitness := math.MaxFloat64
	var countSinceLastImprovement int
	for range ga.Generations {
		newPopulation = ga.evolveGeneration(newPopulation)

		bestIndividual = newPopulation[0]
		// bestIndividual.PrintLog()

		// if bestIndividual.fitness == 0 {
		// 	return
		// }

		if bestIndividual.fitness < lastFitness {
			lastFitness = bestIndividual.fitness
			countSinceLastImprovement = 0
		} else {
			countSinceLastImprovement++
		}
		if countSinceLastImprovement >= ga.MaxRetries {
			// se não houver melhoria no fitness por várias gerações, considera que encontrou o melhor time possível
			// e encerra a evolução
			return bestIndividual
		}
	}

	return
}

func (ga *GeneticAlgorithmGenerator) clear() {
	ga.studentsInTheTeam = 0
	ga.possibleStudents = nil
	ga.mapPossiblePositionsCount = nil
	ga.mapPossibleCompetenceLevelsCount = nil
	ga.tools = nil
}

func (ga *GeneticAlgorithmGenerator) GenerateTeam(subject *models.Subject, project *models.Project) (map[*models.StudentSubjectData]*models.Position, error) {
	// considera que subject.Students está carregado e que o Student dentro de Students existe e não é nil
	// mesma coisa para o project.Students e project.Positions

	// os students dentro de subject.Students são os considerados para preenchimento no projeto

	defer ga.clear()

	ga.currentProjectId = project.Id

	ga.possibleStudents = getPossibleStudents(subject, project)
	if len(ga.possibleStudents) == 0 {
		// todos os estudantes já estão em um projeto
		return nil, ErrAllStudentsAlreadyInProject
	}

	// mapeia os cargos/positions que ainda precisam ser preenchidos
	ga.mapPossiblePositionsCount = getPossiblePositionsCount(project)
	if len(ga.mapPossiblePositionsCount) == 0 {
		// todos os cargos já estão preenchidos
		return nil, ErrAllPositionsAlreadyFilled
	}

	for _, c := range ga.mapPossiblePositionsCount {
		ga.studentsInTheTeam += c
	}
	ga.studentsInTheTeam = min(ga.studentsInTheTeam, int64(len(ga.possibleStudents)))

	if ga.studentsInTheTeam <= 0 {
		return nil, ErrAllStudentsAlreadyInProject
	}

	ga.tools = make([]string, 0, len(project.Tools))
	for _, tool := range project.Tools {
		ga.tools = append(ga.tools, strings.ToLower(tool))
	}
	ga.tools = sliceUtils.RemoveDuplicates(ga.tools)

	// pega quais níveis de competência ainda faltam ser preenchidos
	ga.mapPossibleCompetenceLevelsCount = getPossibleCompetenceLevelsCount(ga.possibleStudents, project)

	if ga.PopulationSize <= 0 {
		ga.PopulationSize = 500
	}

	if ga.Generations <= 0 {
		ga.Generations = 250
	}

	if ga.MaxRetries <= 0 {
		ga.MaxRetries = 100
	}

	population := ga.newPopulation()
	bestTeam := ga.evolve(population)

	// temos o melhor time encontrado, agora precisamos associar cada estudante ao seu cargo
	// TODO fazer isso diretamente na hora de calcular o fitness, para não precisar fazer isso aqui
	return bestTeam.mappedToPosition, nil
}

func (ga *GeneticAlgorithmGenerator) CalculateBalancingScore(team map[*models.StudentSubjectData]*models.Position, allCompetenceLevelsInSubject []*models.CompetenceLevel) uint64 {
	// função não utilizada, pois já é feito o balanceamento no algoritmo genético
	// retorna 4, que é o valor fixo esperado de melhor solução
	return 4
}

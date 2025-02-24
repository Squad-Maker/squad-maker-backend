# Ideias não implementadas ou parcialmente implementadas

## Multidisciplina

A estrutura do backend foi feita pensando em um cadastro de disciplina `Subject`, onde cada disciplina é de um ou mais professores.
Os cargos (`positions`) e senioridades (`competence levels`) são cadastráveis por disciplina.

Atualmente, o metadata de `subject-id` está fixo `1` para sempre ser da disciplina da fábrica de software.

Da parte do banco de dados e lógica, não foi feito nada de dados específicos da disciplina (hoje só tem um cabeçalho com nome), nem a parte de quem é _dono_ (relação com usuários de professores).

Ao pensar nisso, surgem as seguintes dúvidas:

- Como associar um professor a uma disciplina, tendo em conta que o usuário só é cadastrado depois do primeiro login?
- Existe a disciplina padrão da Fábrica de Software, adicionada via migração; como e quem associar como dono dessa disciplina?

Por conta desta estrutura multidisciplina, o cadastro dos alunos é separado em duas partes:
- Cadastro geral (informações vindas da API, como nome e email; outras infos como curso etc); dados obtidos via `Me`
- Cadastro da disciplina (opção de cargo 1, opção de cargo 2, senioridade etc); dados obtidos via `GetStudentSubjectData`

A edição da primeira parte não foi feita nem no front, nem no back. Os dados existentes são os obtidos da API do COENS-DV, que podem estar incorretos ou incompletos.
Levando em conta isso, temos nosso próximo item...

## Edição do cadastro central do usuário

Conforme mencionado acima, o usuário possui um cadastro fixo com dados vindos da API do COENS-DV. Estes dados podem estar incorretos ou incompletos e atualmente não são editáveis.

O dado mais importante de ser editado atualmente é o email, que é utilizado para envio de notificação para usuários professores.

### Tipo de usuário não utilizado

Existe um tipo `Admin` que não é utilizado. Este tipo de usuário não respeita validações de permissão e passa direto por tudo. Neste projeto, não utilizamos, mas vale a menção.

## Parametrização da geração de times

O código de geração de times atual trabalha com pesos. Todos os pesos estão fixos no código. A ideia seria permitir parametrizar isso e passar em cada requisição, ou até mesmo em alguma configuração da disciplina ou do time, quais pesos utilizar. Outras implementações podem não trabalhar com pesos, mas a ideia seria permitir passar os parâmetros necessários de qualquer forma.

## Função mais específica para geração de todos os times

Atualmente, a função de gerar de todos os times pega os projetos da disciplina ordenados por ID e roda a função de geração individual para cada um deles, em ordem.

Isso faz com que exista um viés, onde muitas vezes os alunos não estarão em nenhum dos projetos desejados (pois o algoritmo de geração em si é aleatório, com pesos, mas ainda aleatório...).

Deve ser desenvolvido um algoritmo próprio para geração de todos os times ao mesmo tempo, onde priorize os projetos desejados dos estudantes antes de distribuir os mesmos para outros projetos.

Uma ideia para este desenvolvimento seria fazer essa distribuição em passos. Pode ser utilizada a interface de gerador de times para criar um gerador que só considere os alunos que possuem desejo por algum projeto, fazendo a geração em si em 2 passos: um neste gerador (não implementado atualmente) e outro no gerador atual, completando os times conforme necessário.

## Outros algoritmos de geração de time

Implementar outros algoritmos para geração de time. Por exemplo, utilizar alguma IA para gerar o time.

Isso pode ser feito utilizando a interface `TeamGenerator`, presente no arquivo `grpc/squad/team-generation/types.go`. Deve ser também registrado um novo elemento no enum `TeamGeneratorType`, presente no arquivo `proto/api/squad/enum.proto`, modificando a função `GetTeamGenerator`, presente no mesmo arquivo da interface, para retornar a instância do gerador novo.

## Chat interno do projeto

Implementação de um chat interno do projeto, com notificações, estilo grupo do WhatsApp Web.

Esta implementação exigiria muita complexidade, trabalhando com fila de notificações (provavelmente utilizando `rabbitmq`) e streams sempre abertos com o frontend para fluxo de notificações e dados de mensagens.

## Fluxo de Refresh do Token

Não foi feito este fluxo no frontend. No backend, está completo.

No frontend, a ideia inicial era utilizar um `Service Worker` para injetar o token em todas as requisições, incluindo imagens e outros assets (que poderiam ser _privados_), fazendo com que o refresh fosse gerenciado também somente por uma aba aberta do sistema por vez, utilizando mutex (com alguma biblioteca que utiliza o `IndexedDB`) e dados do `localStorage` para identificar se ainda precisa fazer o refresh do token.

Este fluxo é complexo e não foi implementado. O frontend está usando somente o `accessToken` e não possui implementação do refresh.

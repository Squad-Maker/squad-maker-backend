# Fluxo de dados

Esta documentação busca demonstrar o fluxo de dados do projeto, desde a autenticação até a chegada em uma função.

## Portas de comunicação

As portas de comunicação são definidas no arquivo `.env`. Por padrão:

- A porta `9080` é utilizada para comunicação do `grpc-web`.
- A porta `9090` é utilizada para comunicação nativa do `grpc`.

## Autenticação e Autorização

O fluxo de autenticação se baseia nas funções `CreateToken`, `RefreshToken` e `InvalidateToken`.

`RefreshToken` e `InvalidateToken` esperam que o `session-token` (`refreshToken` retornado na resposta do `CreateToken`) seja passado como header/metadata da requisição, pois utilizam a sessão pra pegar o novo token, ou invalidá-la.

Quando uma requisição chega no backend, por qualquer das portas de comunicação, ela vai passar pelo middleware/_interceptor_ `authorize`, definido no arquivo `cmd/main.go`.

Com exceção das requisições marcadas como ignoradas no arquivo `utils/grpc/route-permissions.go`, todas as requisições esperam que o token seja passado via header/metadata `authorization`, como se fosse uma requisição HTTP comum, utilizando `Bearer {token}` como seu valor.

O interceptor `authorize` valida o token, a sessão e os dados do usuário, além das permissões conforme definido no arquivo `utils/grpc/route-permissions.go`.
Também faz a gravação de alguns metadados (metadata) na requisição antes de repassar ao gRPC, como o ID do usuário, ID da sessão e ID da disciplina (implementação parcial; veja [a documetação de ideias não implementadas](unimplemented.md) para mais informações). Estes metadados podem ser acessados utilizando as funções utilitárias definidas no arquivo `utils/grpc/metadata.go`.

Depois de passar pela validação de autorização, a chamada é passada à implementação da função executada, conforme implementado em seu devido arquivo na pasta `grpc`.

## Fluxo provável no client

No client, incluindo o frontend, é provável que o seguinte fluxo seja utilizado:

- `CreateToken` ao realizar login
- Chamada da função `Me` para obter os dados do usuário
- Chamada da função `ReadAllSubjects` para listagem de disciplinas disponíveis*
- Injeção do ID da disciplina selecionada em todas as requisições via header/metadata `subject-id`*
- Chamada da função `GetStudentSubjectData` para obter os dados do estudante dentro da disciplina.

\* Implementação parcial; atualmente, o ID da disciplina está fixo `1` e não é esperado que o frontend envie esta informação. Vide [a documetação de ideias não implementadas](unimplemented.md) para mais informações.

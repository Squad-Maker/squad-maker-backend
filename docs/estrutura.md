# Estrutura

Esta documentação busca explicar o que cada pasta do backend faz e significa.

## cmd

Este pacote contém o código do executável do backend (sua função `main`) e os middlewares necessários para autorização (validação da autenticação).

Provavelmente só vai precisar ser mexido se precisar alterar as portas do backend, ou se for adicionado algum outro serviço do gRPC.

A implementação do grpc-web utilizada é a da [improbable-eng](https://github.com/improbable-eng/grpc-web), pois possui suporte ao uso de websockets para comunicação de stream bidirecional. Não utilizamos isso neste projeto, mas pode ser útil. Além disso, a implementação de streams em geral é mais estável que a do projeto oficial (que possui diversos _memory leaks_ há anos que não foram resolvidos pois estão aguardando a API de `streams` oficial nos navegadores...).

## database

Contém tudo o que é necessário para gerenciar conexões do banco de dados. Além disso, contém utilitários e funções genéricas utilizados nas funções de `ReadAll` (paginação, filtros etc) e outras coisas que não são específicas (ex: utilitário para pegar cláusula de _lock for update_ e convenção de nomes).

## grpc

Contém a implementação dos métodos dos serviços definidos no protobuf.

O padrão de nomes seguidos são os mesmos da pasta `proto`. Por exemplo:

- Função `GenerateTeam` possui suas mensagens de request (`GenerateTeamRequest`) e response (`GenerateTeamResponse`) definidas no arquivo `proto/api/squad/method.proto`.
- Sua implementação estará então no arquivo `grpc/squad/method.go`.

## migrations

Contém as migrações do banco de dados.

Utilizamos um fork do `goose/v2` que permite criar grupos de migração (ex: para uso em diversos serviços que podem usar o mesmo banco de dados).
Neste projeto, não utilizamos desta funcionalidade específica.

Outra modificação neste fork é o uso de transações do GORM ao invés de conexões da biblioteca padrão da Golang, motivo principal do uso do fork.

Por conta deste fork, o utilitário de geração de migrações do goose não funciona, pois gera os arquivos sem um parâmetro necessário (nome do _serviço_) e gera com a conexão padrão da Golang.

Recomenda-se copiar alguma migração já existente, modificando o nome e conteúdo, para criação de novas migrações.

Também neste pacote, está o arquivo que gerencia as migrações (fluxo de execução das mesmas). A migração de **estrutura** se dá utilizando o `AutoMigrate` do GORM, enquanto a migração de **dados** se dá utilizando as migrações do goose.

## models

Pasta que contém as estruturas de models do banco de dados.

Cada model pode conter funções associadas a ele. Por exemplo, models utilizados em listagens (`ReadAll`) possuem a função (`ConvertToProtobufMessage`), que executa na função genérica de listagem para gerar a resposta da função de listagem.

## proto

Pasta que contém os arquivos do protobuf, que são as definições de mensagens de _Request_ e _Response_, além das definições das funções e serviços do gRPC.

Toda a comunicação entre client e server é definida aqui.

A pasta `third_party` dentro desta pasta `proto` contém as definições padrão de mensagem do google que não vêm junto com o protobuf. Não deve ser modificada.

A pasta `api` contém toda a definição relacionada ao projeto em si.

**Importante:** Esta pasta é um submódulo e, portanto, alterações nela devem considerar isso.

## utils

Pasta que contém qualquer código reutilizável que não se encaixe nas outras pastas citadas neste documento.

### env

Pasta com utilitários para pegar variáveis de ambiente.

As funções disponíveis neste pacote são _wrappers_ da função padrão da Golang (`os.Getenv`), com conversão de dados e retorno de erro.

### grpc (pacote `grpcUtils`)

Contém utilitários relacionados às implementações do gRPC.

Contém os dados de permissão das funções do gRPC.

### jwt (pacote `jwtUtils`)

Contém a definição dos `claims` do token JWT da autenticação e funções para validação do mesmo.

### mail (pacote `mailUtils`)

Contém utilitários para instanciar client de email e preparação de um novo email.

### other (pacote `otherUtils`)

Contém funções cuja categoria não é bem definida, mas são genéricas o suficiente para serem utilizadas em diversos locais do projeto.

### utfpr

Contém a implementação da comunicação com a API do COENS-DV, utilizada para validação do login.


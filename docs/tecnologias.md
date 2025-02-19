# Tecnologias utilizadas

## Linguagem de programação

Este projeto foi desenvolvido utilizando [Go](https://go.dev).

## Comunicação cliente/servidor

A comunicação client/server foi feita utilizando [protobuf](https://protobuf.dev/) e [grpc](https://grpc.io).

Mais informações podem ser encontradas na [documentação da estrutura do projeto](estrutura.md).

## ORM

O ORM utilizado neste projeto é o [GORM](https://github.com/go-gorm/gorm).

## Autenticação e autorização

O fluxo de autenticação e autorização se baseia na API do COENS-DV. Ao efetuar login, um usuário é criado no banco de dados para que seja gerenciado o token de forma interna ao projeto.

O retorno da autenticação consiste em um `accessToken` (JWT), um `refreshToken` (identificador da sessão no banco de dados) e `expiresIn` (segundos para expiração do token).

Mais informações podem ser encontradas na [documentação do fluxo de dados do projeto](fluxo.md).

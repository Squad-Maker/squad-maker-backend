# Configuração do ambiente

Prepare um ambiente linux para a compilação e execução do projeto.

Este documento considera o uso do `ubuntu`, mas qualquer distribuição linux pode ser utilizada.

## instalar o GO no terminal do ubuntu

`sudo apt-get update && sudo apt-get -y install golang-go`
`sudo apt install golang`

## instalar o GRPC, protogen-go e protoC no terminal do projeto no VS Code

`go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`
`go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
`sudo apt install -y protobuf-compiler`

## Instalar também o plugin no terminal do projeto

`go install github.com/Hellysonrp/protoc-gen-go-db-enum@master`

## Configuração do Postgres

###  instalar e rodar o Postgres no ubuntu

`sudo apt install postgresql postgresql-contrib`
`sudo systemctl enable PostgreSQL`
`sudo systemctl start PostgreSQL`

Para checar se está rodando:

`sudo systemctl status PostgreSQL`

Logue no usuário padrão do postgres:

`sudo -i -u postgres`

Entre no CLI do posgres pra executar comandos:

`psql`

Defina uma senha de sua escolha para o usuário:

`ALTER USER postgres PASSWORD '123';`

Crie o banco de dados:

`CREATE DATABASE squadmaker;`

## instalar o multilog no projeto:

`sudo apt update`
`sudo apt-get install daemontools`

# Configuração do ambiente no Windows

Caso esteja utilizando Windows, recomenda-se a utilização de uma instância do WSL para compilação e execução do projeto.

## Instalar o WSL

`wsl --installd`
Acessar então o aplicativo Ubuntu que será instalado. Ele irá solicitar um usuário e senha. Preencher com o que preferir

## Configurar o VSCode

Pelo VSCode no Windows, instalar a extensão `WSL` da Microsoft.com
Para acessar o VSCode com WSL, clicar no botão no canto inferior esquerdo `Open a Remote Window` e selecionar `Connect to WSL`
Com isso pode ser feito o processo de clonar o projeto pelo VS Code e utilizá-lo normalmente

# Configuração inicial do projeto

Vide a [documentação de compilação](docs/build.md) para mais informações do que precisa ser feito no projeto antes de sua primeira execução.

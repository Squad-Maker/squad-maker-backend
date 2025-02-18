# Configuração do ambiente no Windows

Caso esteja utilizando Windows, pode ser utilizado o WSL para rodar o projeto

## Instalar o WSL
wsl --installd
Acessar então o aplicativo Ubuntu que será instalado. Ele irá solicitar um usuário e senha. Preencher com o que preferir

## Configurar o VSCode
Pelo VSCode no Windows, instalar a extensão "WSL" da Microsoft.com
Para acessar o VSCode com WSL, clicar no botão no canto inferior esquerdo "Open a Remote Window" e selecionar connect to WSL
Com isso pode ser feito o processo de clonar o projeto pelo VS Code e utilizá-lo normalmente

## instalar o GO no terminal do ubuntu
sudo apt-get update && sudo apt-get -y install golang-go 
sudo apt install golang

## instalar o GRPC, protogen-go e protoC no terminal do projeto no VS Code
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
sudo apt install -y protobuf-compiler

## Instalar também o plugin no terminal do projeto
go install github.com/Hellysonrp/protoc-gen-go-db-enum@master

## Configuração do Postgres
###  instalar e rodar o Postgres no ubuntu
    sudo apt install postgresql postgresql-contrib
    sudo systemctl enable PostgreSQL
    sudo systemctl start PostgreSQL
###  para checar se está rodando:
    sudo systemctl status PostgreSQL
###  logar no usuário padrão do postgres:
    sudo -i -u postgres
###  entrar no CLI do posgres pra executar comandos:
    psql
###  setar senha da sua escolha para o usuário:
    ALTER USER postgres PASSWORD '123';
###  criar database:
    CREATE DATABASE squadmaker;

## Conexão com o banco no projeto
copiar o arquivo .env.example no projeto na pasta bin, com o nome .env, e então preencher os dados conforme o que foi criado anteriormente
###  exemplo:
    DB_HOST="127.0.0.1"
    DB_PORT="5432"
    DB_NAME=squadmaker
    DB_USER=postgres
    DB_PASS="123"
### Valores para os outros campos:
    JWT_SECRET=CHAVESECRETAJWT
    JWT_EXPIRES_IN="5400"
    SESSION_TOKEN_MAX_AGE="10800"

## instalar o multilog no projeto:
sudo apt update
sudo apt-get install daemontools

## instalar o goose no projeto
go install github.com/pressly/goose/v3/cmd/goose@latest

## Rodar no terminal do projeto:
git submodule update --init --recursive
## Por fim, para rodar:
make start 
ou
make debug
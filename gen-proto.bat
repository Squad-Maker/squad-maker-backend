@echo off
setlocal enabledelayedexpansion

echo Iniciando > generation_log.txt

set files=

REM Itera sobre cada subdiretório em src\grpc\proto\api
for /f %%F in ('dir /b /ad src\grpc\proto\api') do (
    REM Itera sobre cada arquivo .proto dentro do subdiretório
    for /f %%G in ('dir /b src\grpc\proto\api\%%F\*.proto') do (
        REM Adiciona o caminho completo do arquivo .proto à variável files
        set files=!files! src\grpc\proto\api\%%F\%%G
        echo Encontrado arquivo: src\grpc\proto\api\%%F\%%G >> generation_log.txt
    )
)

echo Executando o comando protoc com os arquivos encontrados: >> generation_log.txt
echo protoc -Iproto\\api -Iproto\\third_party --go_out=paths=source_relative:generated --go-grpc_out=paths=source_relative:generated --go-db-enum_out=paths=source_relative:generated !files! >> generation_log.txt

REM Executa o comando protoc com a lista de arquivos .proto acumulada
echo protoc -Iproto\\api -Iproto\\third_party --go_out=paths=source_relative:generated --go-grpc_out=paths=source_relative:generated --go-db-enum_out=paths=source_relative:generated !files! >> generation_log.txt 2>&1

echo Concluido >> generation_log.txt

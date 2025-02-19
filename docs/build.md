# Compilação (building)

Este documento busca explicar como funciona a compilação e o que deve ser feito para primeira execução do projeto.

O foco deste documento é executar o projeto em debug, considerando o uso durante o desenvolvimento. Para uso em produção, o que muda é a função utilizada para build (ao invés de ser em debug, efetuar build sem a flag de debug) e o local que o executável ficará (pois projetos em produção não vão ficar dentro da pasta `bin` do código-fonte...)

## Preparação do ambiente

Vide a [documentação de contribuição](../CONTRIBUTING.md) para informações de como configurar o ambiente.

## Preparação do projeto

Este projeto possui um submódulo com as definições do protobuf. Inicialize o submódulo com o comando `git submodule update --init --recursive`.

Execute o comando `make gen-proto` para gerar todos os arquivos do protobuf e gRPC necessários para o projeto.

Copie o arquivo `.env.example` para a pasta `bin` (crie se não existir), renomeando-o para `.env` e preencha os dados necessários.

Execute o projeto com o comando `make start`.

Vide `Makefile` para outros comandos que podem ser úteis.

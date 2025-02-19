Este documento busca demonstrar como efetuar os testes de API deste backend.

## Ferramentas

É possível utilizar a ferramenta [Postman](https://www.postman.com/) ou a ferramenta [Insomnia](https://insomnia.rest/). Outras ferramentas podem existir, mas este documento se baseará no funcionamento destas duas.

## Execução

Ao compilar e executar o backend em `debug`, um serviço de `reflection` do gRPC é ativado. Este serviço é necessário para que as ferramentas consigam listar os métodos disponíveis no backend.

Ambas as ferramentas mencionadas possuem tipos de requisição próprios para o gRPC, geralmente chamada de `gRPC`.

Ao criar uma requisição do gRPC, insira o endereço da porta de comunicação nativa do gRPC na URL da requisição. Exemplo: `localhost:9090`

A ferramenta Postman já obtém automaticamente a lista de métodos do backend. Na ferramenta Insomnia, você deve clicar no botão com ícone de _refresh_ para listar os métodos.

Ambas as ferramentas possuem um botão para inserir um exemplo de corpo de requisição. Utilize desta funcionalidade para auxiliar no preenchimento manual das requisições.

Utilize o método `CreateToken` para autenticar um usuário. Copie o `accessToken` retornado e utilize ele como header `Authorization`, com valor `Bearer {accessToken}`, em qualquer requisição autenticada.

Ao executar a requisição, a resposta é convertida em um JSON e exibida. Vale notar que parâmetros `int64` na resposta, pelo menos na ferramenta Insomnia, são exibidos como um objeto com dois inteiros dentro, pois JSON tecnicamente não suporta `int64`.

# Deployment recomendado

Para evitar problemas de CORS, recomenda-se utilizar o backend atrás de um proxy reverso, como o `nginx`.

Redirecione todas as requisições de `/api` para a porta do gRPC Web do backend, enquanto o resto das requisições vai para o frontend.

## Exemplo de configuração

Pseudo-configuração do nginx:

```nginx
upstream squad-back {
    server localhost:9080;
}

upstream squad-front {
    server localhost:5173; # exemplo com a porta de dev; não deve ser usado - utilizar algum server de arquivos estáticos com o código compilado (ou o próprio nginx)
}

# map para aceitar requisições de websocket
# pode ser útil para utilizar streams do gRPC
map $http_upgrade $connection_upgrade {
    default upgrade;
    '' '';
}

...

server {
    listen ... ;

    server_name ... ;

    ...

    location / {
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Host $http_host;
        proxy_set_header X-Forwarded-Host $server_name;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_redirect     off;
        proxy_http_version 1.1;
        proxy_pass http://squad-front;

        ... ou servir os arquivos do front de forma estática
    }

    location /api/ {
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Host $http_host;
        proxy_set_header X-Forwarded-Host $server_name;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_redirect     off;
        proxy_http_version 1.1;
        proxy_pass http://squad-back/;
    }

    ...
}
```

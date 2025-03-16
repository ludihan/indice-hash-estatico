# Índice hash estático
Este repositório contém uma implementação de um índice hash e uma interface gráfica para o seu uso

## Build
1. Instale a linguagem go
2. Instale as dependencias do Gio (https://gioui.org/)
3. Execute isto no terminal:
```sh
cd src
go run .
```

4. Caso esteja usando windows e queira desabilitar o terminal execute isto:
```sh
go run -ldflags="-H windowsgui" .
```

5. Pra criar um binário final use "build" ao invés de "run" nos comandos

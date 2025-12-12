# 1. Escolhemos uma imagem base leve do Go (Alpine Linux)
FROM golang:1.23-alpine

# 2. Cria uma pasta dentro do container para nosso app
WORKDIR /app

# 3. Copia os arquivos de dependência primeiro (go.mod)
COPY go.mod ./

# 4. Copia todo o restante do código fonte e a pasta static
COPY . .

# 5. Compila o programa Go gerando um executável chamado "servidor"
RUN go build -o servidor main.go

# 6. Informa ao Docker que o container deve expor portas (informativo)
EXPOSE 8080

# 7. O comando que roda quando o container inicia
CMD ["./servidor"]
# Dockerfile generico para qualquer binario em hub/cmd/<nome>,
# selecionado pelo build arg SERVICE. Usado por todos os seis
# servicos Go da referencia local (ARQ-04: Go em todas as
# APIs/workers).
FROM golang:1.24-bookworm AS build
ARG SERVICE
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/service /service
ENTRYPOINT ["/service"]

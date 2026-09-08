# Dockerfile generico para qualquer binario em hub/cmd/<nome>,
# selecionado pelo build arg SERVICE. Usado por todos os seis
# servicos Go da referencia local (ARQ-04: Go em todas as
# APIs/workers).
FROM golang:1.24-bookworm@sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac AS build
ARG SERVICE
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian12@sha256:d75cdd72874d4790092fcb1b058493ecf6bb5bf2b2b897045b00ff01d91843f2
COPY --from=build /out/service /service
USER 65532:65532
ENTRYPOINT ["/service"]

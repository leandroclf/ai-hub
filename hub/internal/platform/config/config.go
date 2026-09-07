// Package config le configuracao de ambiente para os servicos do hub.
// Nesta referencia local/dev, a configuracao vem de variaveis de
// ambiente (ver hub/deploy/docker-compose.yml); em ppd/prd ela viria de
// projecoes publicadas pelo Atlas (CFG-01/CFG-02), o que permanece
// pendente de P-01/P-11.
package config

import (
	"os"
	"strconv"
	"time"
)

// Env retorna o valor da variavel de ambiente key, ou def se ausente/vazia.
func Env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// EnvInt retorna o valor inteiro da variavel de ambiente key, ou def se
// ausente/invalida.
func EnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// EnvDurationSeconds le uma variavel de ambiente inteira em segundos
// (convencao normativa de CFG-04: todo prazo configuravel e inteiro em
// segundos) e retorna como time.Duration.
func EnvDurationSeconds(key string, defSeconds int) time.Duration {
	return time.Duration(EnvInt(key, defSeconds)) * time.Second
}

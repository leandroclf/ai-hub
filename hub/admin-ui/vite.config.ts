import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Configuracao Vite da interface administrativa do Atlas (ARQ-04).
//
// Em dev, o servidor Vite faz proxy de "/api" para o Atlas real
// (porta 8081, conforme hub/cmd/atlas/main.go e hub/deploy/docker-compose.yml).
// O backend Go do Atlas nao configura CORS (e nao deve — nao alteramos
// o backend para isso), entao todo o frontend chama sempre o prefixo
// relativo "/api/v1/..." (ver src/api/atlasClient.ts) em vez de uma URL
// absoluta. Isso funciona tanto:
//   - em dev, via este proxy; quanto
//   - atras de um gateway (ex.: Kong, hub/deploy/kong/kong.yml) que um dia
//     sirva esta UI e a API sob o mesmo host/porta, bastando reescrever
//     "/api" -> "http://atlas:8081" na borda.
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8081",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
});

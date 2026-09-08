import {defineConfig} from 'vite';
import react from '@vitejs/plugin-react';
const services:Record<string,number>={atlas:8081,orbita:8082,cometa:8083,pulsar:8084,libra:8085};
export default defineConfig({plugins:[react()],server:{proxy:Object.fromEntries(Object.entries(services).map(([domain,port])=>[`/api/${domain}`,{target:`http://localhost:${port}`,changeOrigin:true,rewrite:(path:string)=>path.replace(`/api/${domain}`,'')}]))}});

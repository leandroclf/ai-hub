import {createHmac} from 'node:crypto';
import {writeFile} from 'node:fs/promises';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

// Mede a admissão ASYNC e o GET unificado usando a mesma autenticação OIDC
// exercitada pelo portal. O bearer e o payload não são persistidos no artefato.
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const orbitaOrigin=process.env.ORBITA_URL||'http://localhost:18080';
const samples=Number(process.env.R2_SLO_SAMPLES||30);
const concurrency=Number(process.env.R2_SLO_CONCURRENCY||5);
const payloadBytes=Number(process.env.R2_SLO_PAYLOAD_BYTES||60000);
const timeoutMs=Number(process.env.R2_SLO_TIMEOUT_MS||10000);
const evidencePath=resolve('hub/evidence/r2/execution/slo-load-latest.json');
const modulePath=resolve(process.env.PLAYWRIGHT_MODULE||'hub/evidence/screenshots/node_modules/playwright-core/index.js');
const pw=await import(pathToFileURL(modulePath).href);
const chromium=pw.chromium||pw.default?.chromium;
if(!chromium) throw new Error('playwright-core sem export chromium');
if(!Number.isInteger(samples)||samples<10||samples>200) throw new Error('R2_SLO_SAMPLES deve estar entre 10 e 200');
if(!Number.isInteger(concurrency)||concurrency<1||concurrency>20) throw new Error('R2_SLO_CONCURRENCY deve estar entre 1 e 20');
if(!Number.isInteger(payloadBytes)||payloadBytes<1024||payloadBytes>64*1024) throw new Error('R2_SLO_PAYLOAD_BYTES deve estar entre 1024 e 65536');

function otp(){
 const counter=Buffer.alloc(8);
 counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)));
 const digest=createHmac('sha1',Buffer.from('AI-HUB-R2-MFA-KEY-01')).update(counter).digest();
 const offset=digest[digest.length-1]&15;
 return ((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');
}

async function authenticatedToken(){
 const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/usr/bin/google-chrome'});
 const page=await browser.newPage();
 let token='';
 page.on('request',request=>{
  const authorization=request.headers().authorization||'';
  if(authorization.startsWith('Bearer ')) token=authorization.slice(7);
 });
 try{
  await page.goto(`${adminOrigin}/services`,{waitUntil:'networkidle'});
  await page.getByRole('button',{name:'Entrar',exact:true}).click();
  await page.locator('#username').fill('operadora-a');
  await page.locator('#password').fill('R2-fixture-password!');
  await page.locator('#kc-login').click();
  await page.locator('#otp').waitFor();
  let authenticated=false;
  for(let attempt=0;attempt<8&&!authenticated;attempt++){
   if(await page.locator('#otp').count()){
    await page.locator('#otp').fill(otp());
    await page.locator('#kc-login').click();
   }
   try{
    await page.waitForURL(url=>new URL(url).origin===new URL(adminOrigin).origin&&new URL(url).pathname==='/services',{timeout:3500});
    authenticated=true;
   }catch{
    authenticated=await page.getByRole('heading',{name:'Serviços',exact:true}).count()>0;
    if(!authenticated) await page.waitForTimeout(500);
   }
  }
  if(!authenticated) throw new Error(`autenticacao OIDC com OTP nao concluida (url=${page.url()})`);
  await page.evaluate(()=>fetch('/api/atlas/admin/v1/clients?limit=1'));
  if(!token) throw new Error('token administrativo nao capturado apos autenticacao');
  return {browser,token};
 }catch(error){
  await browser.close();
  throw error;
 }
}

function percentile(values, q){
 const sorted=[...values].sort((a,b)=>a-b);
 return sorted[Math.max(0,Math.ceil(sorted.length*q)-1)];
}

function stats(observations, targets){
 const latencies=observations.map(item=>item.duration_ms);
 const statuses={};
 for(const item of observations) statuses[item.http_status]=Number(statuses[item.http_status]||0)+1;
 const result={
  samples:observations.length,
  success_count:observations.filter(item=>item.ok).length,
  http_statuses:statuses,
  latency_ms:{
   min:Math.min(...latencies),
   p50:percentile(latencies,.50),
   p95:percentile(latencies,.95),
   p99:percentile(latencies,.99),
   max:Math.max(...latencies)
  },
  target_ms:targets,
  target_result:latencies.length===observations.length&&
   observations.every(item=>item.ok)&&
   percentile(latencies,.95)<=targets.p95&&
   percentile(latencies,.99)<=targets.p99?'PASS':'FAIL'
 };
 return result;
}

async function request(url, options){
 const controller=new AbortController();
 const timer=setTimeout(()=>controller.abort(),timeoutMs);
 const started=process.hrtime.bigint();
 try{
  const response=await fetch(url,{...options,signal:controller.signal});
  const body=await response.json().catch(()=>({}));
  const duration_ms=Number(process.hrtime.bigint()-started)/1e6;
  return {http_status:response.status,duration_ms,body,ok:response.status>=200&&response.status<300};
 }catch(error){
  const duration_ms=Number(process.hrtime.bigint()-started)/1e6;
  return {http_status:0,duration_ms,body:{},ok:false,error:error.name};
 }finally{clearTimeout(timer);}
}

async function inBatches(items, fn){
 const results=[];
 for(let index=0;index<items.length;index+=concurrency){
  const batch=items.slice(index,index+concurrency);
  results.push(...await Promise.all(batch.map(fn)));
 }
 return results;
}

const {browser,token}=await authenticatedToken();
const startedAt=new Date().toISOString();
const prefix=process.env.R2_SLO_PREFIX||`r4-slo-${Date.now()}`;
const filler='x'.repeat(Math.max(0,payloadBytes-140));
const requests=Array.from({length:samples},(_,index)=>({
 idempotencyKey:`${prefix}-${String(index+1).padStart(3,'0')}`,
 body:{mode:'ASYNC',provider_account_id:'prov-poll-1',service_code:'protocolo-assincrono',service_version:1,input:{cpf:filler}}
}));
const requestBodyBytes=Buffer.byteLength(JSON.stringify(requests[0].body));

try{
 const admission=await inBatches(requests,item=>request(`${orbitaOrigin}/v1/protocols`,{
  method:'POST',
  headers:{'content-type':'application/json',authorization:`Bearer ${token}`,'X-Tenant-Id':'acme','Idempotency-Key':item.idempotencyKey},
  body:JSON.stringify(item.body)
 }));
 const protocolIds=admission.map(item=>item.body.protocol_id).filter(id=>typeof id==='string'&&id.length>0);
 const gets=await inBatches(protocolIds,id=>request(`${orbitaOrigin}/v1/protocols/${encodeURIComponent(id)}`,{
  method:'GET',headers:{authorization:`Bearer ${token}`,'X-Tenant-Id':'acme'}
 }));
 const admissionStats=stats(admission,item=>item);
 const getStats=stats(gets,{p95:100,p99:250});
 // A admissão tem metas próprias da R2-OPE-09-S01; mantém a forma explícita
 // no relatório para evitar confundir uma medição com aprovação de produto.
 admissionStats.target_ms={p95:250,p99:500};
 admissionStats.target_result=admissionStats.success_count===admissionStats.samples&&
  admissionStats.latency_ms.p95<=250&&admissionStats.latency_ms.p99<=500?'PASS':'FAIL';
 const result={
  status:admissionStats.target_result==='PASS'&&getStats.target_result==='PASS'&&protocolIds.length===samples?'PASS':'FAIL',
  profile:'r2-ope-09-s01-reference',
  started_at:startedAt,
  finished_at:new Date().toISOString(),
  origin:{orbita:orbitaOrigin,admin:adminOrigin},
  load:{samples,concurrency,payload_bytes:requestBodyBytes,payload_limit_bytes:64*1024,mode:'ASYNC',provider_account_id:'prov-poll-1',service_code:'protocolo-assincrono'},
  admission:admissionStats,
  get:getStats,
  protocol_ids_observed:protocolIds.length,
  failures:{admission:admission.filter(item=>!item.ok).map(item=>({http_status:item.http_status,error:item.error||null})),get:gets.filter(item=>!item.ok).map(item=>({http_status:item.http_status,error:item.error||null}))}
 };
 await writeFile(evidencePath,`${JSON.stringify(result,null,2)}\n`);
 console.log(JSON.stringify(result,null,2));
 if(result.status!=='PASS') process.exitCode=1;
}finally{await browser.close();}

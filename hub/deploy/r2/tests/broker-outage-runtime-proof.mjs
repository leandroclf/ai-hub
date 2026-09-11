import {createHmac} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

// Exercita somente o Compose oficial: broker fora, reinício controlado de
// Orbita/Cometa, SYNC direto e GET. O runner sempre tenta restaurar o broker.
const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const composeFile=resolve(process.env.R2_COMPOSE_FILE||'hub/deploy/r2/compose.yaml');
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const orbitaOrigin=process.env.ORBITA_URL||'http://localhost:18080';
const cometaOrigin=process.env.COMETA_URL||'http://localhost:18082';
const postgresContainer=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const evidencePath=resolve('hub/evidence/r2/execution/broker-outage-runtime-latest.json');
const modulePath=resolve(process.env.PLAYWRIGHT_MODULE||'hub/evidence/screenshots/node_modules/playwright-core/index.js');
const pw=await import(pathToFileURL(modulePath).href);
const chromium=pw.chromium||pw.default?.chromium;
if(!chromium) throw new Error('playwright-core sem export chromium');

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

function compose(args){
 const env={
  ...process.env,
  R2_PROVIDER_CIDR:`${fixtureIP('provider-sim')}/32`,
  R2_SINK_CIDR:`${fixtureIP('webhook-sink')}/32`,
  R2_IDENTITY_CIDR:`${fixtureIP('identity')}/32`
 };
 return execFileSync('docker',['compose','--project-name',project,'--file',composeFile,...args],{encoding:'utf8',stdio:['ignore','pipe','pipe'],env});
}

function fixtureIP(service){
 const container=`${project}-${service}-1`;
 return execFileSync('docker',['inspect',container,'--format','{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'],{encoding:'utf8'}).trim();
}

function sql(statement){
 return execFileSync('docker',['exec',postgresContainer,'psql','-U','hub','-d','hub_core','-At','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim();
}

async function waitReady(url,timeoutMs=45000){
 const deadline=Date.now()+timeoutMs;
 let last='';
 while(Date.now()<deadline){
  try{
   const response=await fetch(url);
   if(response.status===200) return;
   last=`HTTP ${response.status}`;
  }catch(error){last=error.code||error.name||'request_failed';}
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 throw new Error(`${url} nao ficou pronto: ${last}`);
}

async function request(url,options){
 const started=process.hrtime.bigint();
 try{
  const response=await fetch(url,{...options,signal:AbortSignal.timeout(15000)});
  const body=await response.json().catch(()=>({}));
  return {status:response.status,duration_ms:Number(process.hrtime.bigint()-started)/1e6,body};
 }catch(error){
  return {status:0,duration_ms:Number(process.hrtime.bigint()-started)/1e6,error:error.name};
 }
}

const {browser,token}=await authenticatedToken();
const startedAt=new Date().toISOString();
const prefix=process.env.R2_BROKER_OUTAGE_PREFIX||`r4-broker-outage-${Date.now()}`;
const key=`${prefix}-sync-001`;
let result;
let brokerRestored=false;
try{
 compose(['stop','localstack']);
 compose(['up','-d','--force-recreate','--no-deps','orbita','cometa']);
 await waitReady(`${orbitaOrigin}/healthz/ready`);
 await waitReady(`${cometaOrigin}/healthz/ready`);

 const admission=await request(`${orbitaOrigin}/v1/protocols`,{
  method:'POST',
  headers:{'content-type':'application/json',authorization:`Bearer ${token}`,'X-Tenant-Id':'acme','Idempotency-Key':key},
  body:JSON.stringify({mode:'SYNC',provider_account_id:'prov-sync-1',service_code:'consulta-cadastral',service_version:1,input:{cpf:'44444444444'}})
 });
 const protocolID=admission.body?.protocol_id||'';
 const get=protocolID?await request(`${orbitaOrigin}/v1/protocols/${encodeURIComponent(protocolID)}`,{method:'GET',headers:{authorization:`Bearer ${token}`,'X-Tenant-Id':'acme'}}):{status:0,body:{},duration_ms:0};
 const persisted=sql(`SELECT status FROM protocols WHERE idempotency_key='${key}' AND tenant_id='acme' LIMIT 1`);
 const finalizedOutbox=protocolID?sql(`SELECT count(*) FROM outbox WHERE aggregate_id='${protocolID}' AND event_type='protocol.finalized'`):'0';
 const logs=compose(['logs','--no-color','--since','90s','orbita','cometa']);
 const brokerUnavailable=/broker bootstrap unavailable|falha ao localizar fila|falha ao criar fila/i.test(logs);
 const brokerNotFatal=!/panic:|fatal error|server stopped/i.test(logs);
 result={
  status:admission.status===200&&admission.body?.status==='SUCCEEDED'&&get.status===200&&get.body?.protocol_id===protocolID&&persisted==='SUCCEEDED'&&finalizedOutbox==='1'&&brokerUnavailable&&brokerNotFatal?'PASS':'FAIL',
  profile:'r2-ope-06-s01-broker-outage',
  started_at:startedAt,
  finished_at:new Date().toISOString(),
  compose_project:project,
  broker:{service:'localstack',stopped_before_restart:true,unavailable_observed:brokerUnavailable,restored_after_proof:false},
  workloads:{orbita_ready:true,cometa_ready:true,recreated_without_dependencies:true},
  sync:{http_status:admission.status,status:admission.body?.status||null,duration_ms:admission.duration_ms,protocol_id_observed:Boolean(protocolID)},
  get:{http_status:get.status,status:get.body?.status||null,duration_ms:get.duration_ms,protocol_id_observed:get.body?.protocol_id===protocolID},
  persistence:{protocol_status:persisted,finalized_outbox:Number(finalizedOutbox)},
  oracle:{broker_error_observed:brokerUnavailable,process_remained_healthy:brokerNotFatal}
 };
}finally{
 try{
  compose(['up','-d','localstack']);
  await waitReady(`${process.env.LOCALSTACK_URL||'http://localhost:14566'}/_localstack/health`,45000);
  compose(['up','-d','--no-deps','orbita','cometa']);
  brokerRestored=true;
 }catch(error){
  if(result) result.cleanup_error=error.message;
 }
 if(result){
  result.broker.restored_after_proof=brokerRestored;
  if(!brokerRestored) result.status='FAIL';
  result.finished_at=new Date().toISOString();
  await writeFile(evidencePath,`${JSON.stringify(result,null,2)}\n`);
  console.log(JSON.stringify(result,null,2));
  if(result.status!=='PASS') process.exitCode=1;
 }
 await browser.close();
}

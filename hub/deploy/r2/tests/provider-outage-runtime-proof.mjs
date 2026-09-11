import {createHmac} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

// Exercita o limite de uma capacidade externa: provider-sim fora, sem
// reiniciar os demais workloads, sem anunciar sucesso fictício e com
// reconciliação posterior somente quando o oráculo local prova ausência.
const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const composeFile=resolve(process.env.R2_COMPOSE_FILE||'hub/deploy/r2/compose.yaml');
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const orbitaOrigin=process.env.ORBITA_URL||'http://localhost:18080';
const providerOrigin=process.env.PROVIDER_URL||'http://localhost:18090';
const postgres=`${project}-postgres-1`;
const evidencePath=resolve('hub/evidence/r2/execution/provider-outage-runtime-latest.json');
const reconciliationPath=resolve('hub/evidence/r2/execution/provider-outage-reconciliation-latest.log');
const modulePath=resolve(process.env.PLAYWRIGHT_MODULE||'hub/evidence/screenshots/node_modules/playwright-core/index.js');
const pw=await import(pathToFileURL(modulePath).href);
const chromium=pw.chromium||pw.default?.chromium;
if(!chromium) throw new Error('playwright-core sem export chromium');

function otp(){
 const counter=Buffer.alloc(8);
 counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)));
 const digest=createHmac('sha1',Buffer.from('JBSWY3DPEHPK3PXP')).update(counter).digest();
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

function fixtureIP(service){
 return execFileSync('docker',['inspect',`${project}-${service}-1`,'--format','{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'],{encoding:'utf8'}).trim();
}

function compose(args){
 const env={...process.env,R2_PROVIDER_CIDR:`${fixtureIP('provider-sim')}/32`,R2_SINK_CIDR:`${fixtureIP('webhook-sink')}/32`,R2_IDENTITY_CIDR:`${fixtureIP('identity')}/32`};
 return execFileSync('docker',['compose','--project-name',project,'--file',composeFile,...args],{encoding:'utf8',stdio:['ignore','pipe','pipe'],env});
}

function sql(statement){
 return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-F','|','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim();
}

async function request(url,options){
 const started=process.hrtime.bigint();
 try{
  const response=await fetch(url,{...options,signal:AbortSignal.timeout(15000)});
  const body=await response.json().catch(()=>({}));
  return {status:response.status,duration_ms:Number(process.hrtime.bigint()-started)/1e6,body};
 }catch(error){return {status:0,duration_ms:Number(process.hrtime.bigint()-started)/1e6,error:error.name};}
}

function containerID(service){return execFileSync('docker',['inspect',`${project}-${service}-1`,'--format','{{.Id}}'],{encoding:'utf8'}).trim();}

const {browser,token}=await authenticatedToken();
const baselineServices=['atlas','orbita','cometa','pulsar','libra','webhook-sink'];
const baselineIDs=Object.fromEntries(baselineServices.map(service=>[service,containerID(service)]));
const prefix=process.env.R2_PROVIDER_OUTAGE_PREFIX||`r4-provider-outage-${Date.now()}`;
const key=`${prefix}-sync-001`;
const startedAt=new Date().toISOString();
let result;
let providerRestored=false;
try{
 compose(['stop','provider-sim']);
 let providerDown=false;
 try{await fetch(`${providerOrigin}/__qualification/protocols/absent`,{signal:AbortSignal.timeout(2000)});}catch{providerDown=true;}
 const admission=await request(`${orbitaOrigin}/v1/protocols`,{method:'POST',headers:{'content-type':'application/json',authorization:`Bearer ${token}`,'X-Tenant-Id':'acme','Idempotency-Key':key},body:JSON.stringify({mode:'SYNC',provider_account_id:'prov-sync-1',service_code:'consulta-cadastral',service_version:1,input:{cpf:'55555555555'}})});
 const protocolID=admission.body?.protocol_id||'';
 const get=protocolID?await request(`${orbitaOrigin}/v1/protocols/${encodeURIComponent(protocolID)}`,{method:'GET',headers:{authorization:`Bearer ${token}`,'X-Tenant-Id':'acme'}}):{status:0,body:{},duration_ms:0};
 const persisted=protocolID?sql(`SELECT p.status,o.state,COALESCE(cp.transport_open::text,''),COALESCE(cp.pending_external::text,'') FROM protocols p JOIN operations o ON o.protocol_id=p.protocol_id LEFT JOIN capacity_permits cp ON cp.permit_id::text=o.operation_id::text WHERE p.protocol_id='${protocolID}' LIMIT 1`):'';
 const currentIDs=Object.fromEntries(baselineServices.map(service=>[service,containerID(service)]));
 const unaffected=baselineServices.every(service=>currentIDs[service]===baselineIDs[service]);
 const [protocolStatus,operationState,transportOpen,pendingExternal]=persisted.split('|');
 result={status:providerDown&&admission.status===504&&get.status===200&&Boolean(protocolID)&&operationState==='UNKNOWN'&&transportOpen==='false'&&pendingExternal==='true'&&unaffected?'PASS':'FAIL',profile:'r2-ope-03-s03-provider-outage',started_at:startedAt,finished_at:new Date().toISOString(),compose_project:project,provider:{service:'provider-sim',stopped_before_request:true,unavailable_observed:providerDown,restored_after_proof:false},request:{http_status:admission.status,protocol_id_observed:Boolean(protocolID),duration_ms:admission.duration_ms},get:{http_status:get.status,status:get.body?.status||null,protocol_id_observed:get.body?.protocol_id===protocolID,duration_ms:get.duration_ms},persistence:{protocol_status:protocolStatus||null,operation_state:operationState||null,transport_open:transportOpen||null,pending_external:pendingExternal||null},workloads:{unaffected_container_ids:unaffected,checked:baselineServices}};
}finally{
 try{
  compose(['up','-d','provider-sim']);
  let ready=false;
  for(let attempt=0;attempt<45&&!ready;attempt++){
   try{ready=(await fetch(`${providerOrigin}/healthz/ready`)).status===200;}catch{}
   if(!ready) await new Promise(resolve=>setTimeout(resolve,1000));
  }
  providerRestored=ready;
  if(providerRestored){
   execFileSync('bash',['hub/deploy/r2/tests/reconcile-local-pending.sh'],{encoding:'utf8',env:{...process.env,R2_LOCAL_RECONCILIATION_CONFIRM:'I_UNDERSTAND_LOCAL_FIXTURE',R2_RECONCILIATION_EVIDENCE_FILE:reconciliationPath}});
  }
 }catch(error){if(result) result.cleanup_error=error.message;}
 if(result){
  result.provider.restored_after_proof=providerRestored;
  if(!providerRestored) result.status='FAIL';
  result.finished_at=new Date().toISOString();
  await writeFile(evidencePath,`${JSON.stringify(result,null,2)}\n`);
  console.log(JSON.stringify(result,null,2));
  if(result.status!=='PASS') process.exitCode=1;
 }
 await browser.close();
}

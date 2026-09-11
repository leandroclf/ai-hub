import {createHmac} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

// Exercita a recuperação de uma intenção ASYNC imediatamente após o aceite.
// A janela exata entre o commit local e a publicação no broker não é
// observável externamente; por isso o oráculo exige identidade única,
// recuperação após reinício e ausência de segundo efeito.
const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const composeFile=resolve(process.env.R2_COMPOSE_FILE||'hub/deploy/r2/compose.yaml');
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const orbitaOrigin=process.env.ORBITA_URL||'http://localhost:18080';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const providerSim=process.env.PROVIDER_SIM_URL||'http://localhost:18090';
const evidencePath=resolve('hub/evidence/r2/execution/crash-after-acceptance-runtime-latest.json');
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
  if(!authenticated) throw new Error(`autenticação OIDC não concluída (url=${page.url()})`);
  await page.evaluate(()=>fetch('/api/atlas/admin/v1/clients?limit=1'));
  if(!token) throw new Error('token administrativo não capturado após autenticação');
  return {browser,token};
 }catch(error){await browser.close();throw error}
}

function compose(args){
 const fixtureIP=service=>execFileSync('docker',['inspect',`${project}-${service}-1`,'--format','{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'],{encoding:'utf8'}).trim();
 return execFileSync('docker',['compose','--project-name',project,'--file',composeFile,...args],{
  encoding:'utf8',
  stdio:['ignore','pipe','pipe'],
  env:{...process.env,R2_PROVIDER_CIDR:`${fixtureIP('provider-sim')}/32`,R2_SINK_CIDR:`${fixtureIP('webhook-sink')}/32`,R2_IDENTITY_CIDR:`${fixtureIP('identity')}/32`}
 });
}

function sql(statement){
 return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-F','\t','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim();
}

async function waitReady(timeoutMs=45000){
 const deadline=Date.now()+timeoutMs;
 let last='';
 while(Date.now()<deadline){
  try{
   const response=await fetch(`${orbitaOrigin}/healthz/ready`);
   if(response.status===200)return;
   last=`HTTP ${response.status}`;
  }catch(error){last=error.code||error.name||'request_failed'}
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 throw new Error(`Orbita não ficou pronta: ${last}`);
}

async function request(token,path,options={}){
 const response=await fetch(`${orbitaOrigin}${path}`,{...options,headers:{'content-type':'application/json',authorization:`Bearer ${token}`,'X-Tenant-Id':'acme',...(options.headers||{})}});
 const body=await response.json().catch(()=>({}));
 return {status:response.status,body};
}

const prefix=process.env.R2_CRASH_ACCEPTANCE_PREFIX||`r4-crash-acceptance-${Date.now()}`;
const key=`${prefix}-async-001`;
const {browser,token}=await authenticatedToken();
let result={status:'FAIL',profile:'r2-exe-01-s01-crash-after-acceptance',compose_project:project,key};
try{
 const effectsBefore=await fetch(`${providerSim}/__qualification/effects`).then(response=>response.json());
 const admission=await request(token,'/v1/protocols',{method:'POST',headers:{'Idempotency-Key':key},body:JSON.stringify({mode:'ASYNC',provider_account_id:'prov-poll-1',service_code:'protocolo-assincrono',service_version:1,input:{delay_ms:3000}})});
 if(admission.status!==202||!admission.body.protocol_id)throw new Error(`aceite ASYNC inesperado: HTTP ${admission.status} ${JSON.stringify(admission.body)}`);
 const protocolID=admission.body.protocol_id;
 const accepted=sql(`SELECT count(*),max(status) FROM protocols WHERE tenant_id='acme' AND idempotency_key='${key}' AND protocol_id='${protocolID}'`);
 if(!['1\tACCEPTED','1\tQUEUED','1\tWAITING_PROVIDER'].includes(accepted))throw new Error(`aceite não ficou durável antes do restart: ${accepted}`);

 compose(['stop','orbita']);
 compose(['up','-d','--no-deps','orbita']);
 await waitReady();

 let current=admission.body;
 const deadline=Date.now()+30000;
 do{
  await new Promise(resolve=>setTimeout(resolve,750));
  current=(await request(token,`/v1/protocols/${encodeURIComponent(protocolID)}`)).body;
  if(['SUCCEEDED','FAILED','EXPIRED'].includes(current.status))break;
 }while(Date.now()<deadline);
 const durable=sql(`SELECT (SELECT count(*) FROM protocols WHERE tenant_id='acme' AND idempotency_key='${key}')::text||E'\\t'||(SELECT count(*) FROM operations WHERE tenant_id='acme' AND protocol_id='${protocolID}')::text||E'\\t'||(SELECT count(*) FROM outbox WHERE aggregate_id='${protocolID}' AND event_type='protocol.finalized')`);
 const effectsAfter=await fetch(`${providerSim}/__qualification/effects`).then(response=>response.json());
 const [protocols,operations,finalized]=durable.split('\t').map(Number);
 result={...result,status:current.status==='SUCCEEDED'&&protocols===1&&operations===1&&finalized===1&&effectsAfter.effects-effectsBefore.effects===1?'PASS':'FAIL',protocol_id:protocolID,accepted_status:accepted,restart:{stopped_after_durable_acceptance:true,orbita_ready:true},recovered_status:current.status,durable:{protocols,operations,finalized_outbox:finalized},external_effects:{before:effectsBefore.effects,after:effectsAfter.effects,delta:effectsAfter.effects-effectsBefore.effects},limitation:'A janela interna entre commit e publicação no broker não é observável externamente; o ensaio comprova recuperação e unicidade após reinício imediato.'};
}catch(error){result.error=error.message.split('\n')[0];process.exitCode=1}
finally{
 try{compose(['up','-d','--no-deps','orbita']);await waitReady(45000)}catch(error){result.cleanup_error=error.message.split('\n')[0];result.status='FAIL'}
 result.finished_at=new Date().toISOString();
 await writeFile(evidencePath,`${JSON.stringify(result,null,2)}\n`);
 console.log(JSON.stringify(result,null,2));
 await browser.close();
}

import {createHmac} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
import {connect} from 'node:net';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

// Simula a perda da conexão do cliente depois de enviar o aceite, antes de
// receber a resposta. A queda não é injetada dentro do commit PostgreSQL;
// portanto o resultado qualifica recuperação/idempotência na fronteira HTTP,
// mantendo essa limitação no artefato.
const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const orbitaOrigin=process.env.ORBITA_URL||'http://localhost:18080';
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const evidencePath=resolve('hub/evidence/r2/execution/commit-uncertain-runtime-latest.json');
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
  if(authorization.startsWith('Bearer '))token=authorization.slice(7);
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
    if(!authenticated)await page.waitForTimeout(500);
   }
  }
  if(!authenticated)throw new Error(`autenticação OIDC não concluída (url=${page.url()})`);
  await page.evaluate(()=>fetch('/api/atlas/admin/v1/clients?limit=1'));
  if(!token)throw new Error('token administrativo não capturado após autenticação');
  return {browser,token};
 }catch(error){await browser.close();throw error}
}

function sql(statement){
 return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-F','\t','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim();
}

function sendAndClose(token,key,payload){
 return new Promise((resolvePromise,reject)=>{
  const body=JSON.stringify(payload);
  const url=new URL(orbitaOrigin);
  const socket=connect(Number(url.port||80),url.hostname,()=>{
   socket.write(`POST /v1/protocols HTTP/1.1\r\nHost: ${url.host}\r\nContent-Type: application/json\r\nAuthorization: Bearer ${token}\r\nX-Tenant-Id: acme\r\nIdempotency-Key: ${key}\r\nContent-Length: ${Buffer.byteLength(body)}\r\nConnection: close\r\n\r\n${body}`,()=>{
    setTimeout(()=>{
     socket.destroy();
     resolvePromise({client_closed_before_response:true,post_send_grace_ms:100});
    },100);
   });
  });
  socket.on('error',error=>{
   if(error.code==='ECONNRESET'||error.code==='EPIPE')resolvePromise({client_closed_before_response:true,transport_error:error.code});
   else reject(error);
  });
 });
}

async function request(token,key,payload){
 const response=await fetch(`${orbitaOrigin}/v1/protocols`,{method:'POST',headers:{'content-type':'application/json',authorization:`Bearer ${token}`,'X-Tenant-Id':'acme','Idempotency-Key':key},body:JSON.stringify(payload)});
 const body=await response.json().catch(()=>({}));
 return {status:response.status,body};
}

const prefix=process.env.R2_COMMIT_UNCERTAIN_PREFIX||`r4-commit-uncertain-${Date.now()}`;
const key=`${prefix}-async-001`;
const payload={mode:'ASYNC',provider_account_id:'prov-poll-1',service_code:'protocolo-assincrono',service_version:1,input:{delay_ms:2500}};
const {browser,token}=await authenticatedToken();
let result={status:'FAIL',profile:'r2-exe-01-s03-commit-uncertain',compose_project:project,key};
try{
 const dropped=await sendAndClose(token,key,payload);
 let accepted='';
 const deadline=Date.now()+10000;
 while(Date.now()<deadline){
  accepted=sql(`SELECT count(*)::text||E'\\t'||COALESCE((SELECT protocol_id::text FROM protocols WHERE tenant_id='acme' AND idempotency_key='${key}' ORDER BY accepted_at DESC LIMIT 1),'')||E'\\t'||COALESCE((SELECT status FROM protocols WHERE tenant_id='acme' AND idempotency_key='${key}' ORDER BY accepted_at DESC LIMIT 1),'') FROM protocols WHERE tenant_id='acme' AND idempotency_key='${key}'`);
  if(accepted.startsWith('1\t'))break;
  await new Promise(resolve=>setTimeout(resolve,400));
 }
 const [count,protocolID,acceptedStatus]=accepted.split('\t');
 if(count!=='1'||!protocolID)throw new Error(`aceite não foi recuperado após desconexão: ${accepted}`);
 const replay=await request(token,key,payload);
 const cardinality=sql(`SELECT (SELECT count(*) FROM protocols WHERE tenant_id='acme' AND idempotency_key='${key}')::text||E'\\t'||(SELECT count(*) FROM operations WHERE tenant_id='acme' AND protocol_id='${protocolID}')::text`);
 const [protocols,operations]=cardinality.split('\t').map(Number);
 if(replay.status!==200||replay.body.protocol_id!==protocolID||protocols!==1)throw new Error(`replay não recuperou a identidade: HTTP ${replay.status} ${JSON.stringify(replay.body)} cardinalidade=${cardinality}`);
 result={...result,status:operations<=1?'PASS':'FAIL',transport:dropped,accepted:{status:acceptedStatus,protocol_id:protocolID},replay:{http_status:replay.status,protocol_id:replay.body.protocol_id,status:replay.body.status},durable:{protocols,operations},limitation:'A conexão foi encerrada após o envio do request, mas não há injeção no ponto interno do commit PostgreSQL; o ensaio comprova a recuperação/idempotência observável na fronteira HTTP.'};
}catch(error){result.error=error.message.split('\n')[0];process.exitCode=1}
finally{result.finished_at=new Date().toISOString();await writeFile(evidencePath,`${JSON.stringify(result,null,2)}\n`);console.log(JSON.stringify(result,null,2));await browser.close()}

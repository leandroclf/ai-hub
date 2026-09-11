import {createHmac,createHash,randomUUID} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const localstack=process.env.R2_LOCALSTACK_CONTAINER||`${project}-localstack-1`;
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const sinkOrigin=process.env.WEBHOOK_SINK_URL||'http://localhost:18091';
const modulePath=resolve(process.env.PLAYWRIGHT_MODULE||'hub/evidence/screenshots/node_modules/playwright-core/index.js');
const pw=await import(pathToFileURL(modulePath).href);
const chromium=pw.chromium||pw.default?.chromium;
if(!chromium)throw new Error('playwright-core sem export chromium');

function otp(){
 const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)));
 const digest=createHmac('sha1',Buffer.from('JBSWY3DPEHPK3PXP')).update(counter).digest();
 const offset=digest[digest.length-1]&15;
 return ((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');
}
function sql(query){return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-F','|','-v','ON_ERROR_STOP=1','-c',query],{encoding:'utf8'}).trim()}
function canonical(value){
 if(Array.isArray(value))return value.map(canonical);
 if(value&&typeof value==='object')return Object.fromEntries(Object.entries(value).sort(([a],[b])=>a.localeCompare(b)).map(([key,item])=>[key,canonical(item)]));
 return value;
}
async function authenticate(page){
 let token='';
 page.on('request',request=>{const authorization=request.headers().authorization||'';if(authorization.startsWith('Bearer '))token=authorization.slice(7)});
 await page.goto(`${adminOrigin}/services`,{waitUntil:'networkidle'});
 await page.getByRole('button',{name:'Entrar',exact:true}).click();
 await page.locator('#username').fill('operadora-a');
 await page.locator('#password').fill('R2-fixture-password!');
 await page.locator('#kc-login').click();
 await page.locator('#otp').waitFor();
 for(let attempt=0;attempt<8;attempt++){
  if(await page.locator('#otp').count()){await page.locator('#otp').fill(otp());await page.locator('#kc-login').click()}
  try{await page.waitForURL(url=>new URL(url).origin===new URL(adminOrigin).origin&&new URL(url).pathname==='/services',{timeout:3500});break}catch{await page.waitForTimeout(500)}
 }
 await page.evaluate(()=>fetch('/api/atlas/admin/v1/clients?limit=1'));
 if(!token)throw new Error('token OIDC não foi capturado após a primeira chamada autenticada');
 return token;
}
async function request(token,path,options={}){
 const response=await fetch(`${adminOrigin}/api/orbita${path}`,{...options,headers:{'content-type':'application/json',authorization:`Bearer ${token}`,...(options.headers||{})}});
 const raw=await response.text();
 let body={};try{body=JSON.parse(raw)}catch{}
 return {status:response.status,raw,body};
}
async function ensureApplicationDestination(token){
 const url=`${adminOrigin}/api/pulsar/admin/v1/destinations?tenant_id=acme`;
 const secretVersion=execFileSync('docker',['exec',localstack,'awslocal','secretsmanager','get-secret-value','--secret-id','r2/provider/fixture','--query','VersionId','--output','text'],{encoding:'utf8'}).trim();
 const current=await fetch(url,{headers:{authorization:`Bearer ${token}`}});
 if(!current.ok)throw new Error(`consulta de destinos administrativos falhou: HTTP ${current.status}`);
 const page=await current.json();
 const destination=(page.items||[]).find(item=>item.application_id==='app-acme'&&item.state==='ACTIVE');
 if(destination&&destination.timeout_seconds<=2&&destination.secret_ref==='r2/provider/fixture'&&destination.secret_version===secretVersion&&destination.url.endsWith('/webhook'))return;
 const response=await fetch(url,{method:'POST',headers:{'content-type':'application/json',authorization:`Bearer ${token}`},body:JSON.stringify({
  id:destination?.id||randomUUID(),version:(destination?.version||0)+1,application_id:'app-acme',url:'http://webhook-sink:8091/webhook',
  secret_ref:'r2/provider/fixture',secret_version:secretVersion,state:'ACTIVE',max_attempts:3,timeout_seconds:2,
  reason:'destino da prova de representação R4'
 })});
 const body=await response.text();
 if(!response.ok)throw new Error(`publicação do destino administrativo falhou: HTTP ${response.status} ${body}`);
}
async function waitDelivery(protocolID){
 const deadline=Date.now()+15000;
 while(Date.now()<deadline){
  const row=sql(`SELECT delivery_id || '|' || encode(representation,'base64') || '|' || body_sha256 FROM deliveries WHERE protocol_id='${protocolID}' ORDER BY created_at DESC LIMIT 1`);
  if(row)return row.split('|');
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 throw new Error(`Pulsar não criou entrega para ${protocolID}`);
}
async function waitSink(protocolID){
 const deadline=Date.now()+15000;
 while(Date.now()<deadline){
  const response=await fetch(`${sinkOrigin}/received`);
  if(response.ok){
   const items=await response.json();
   const found=(items||[]).find(item=>item?.body?.protocol_id===protocolID);
   if(found)return found.body;
  }
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 throw new Error(`webhook-sink não recebeu ${protocolID}`);
}

const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/usr/bin/google-chrome'});
const page=await browser.newPage();
const evidence=[];
try{
 const token=await authenticate(page);
 await ensureApplicationDestination(token);
 const key=`r4-representation-${Date.now()}`;
 const payload={mode:'SYNC',provider_account_id:'prov-sync-1',service_code:'consulta-cadastral',service_version:1,input:{cpf:'55555555555'}};
 const post=await request(token,'/v1/protocols',{method:'POST',headers:{'Idempotency-Key':key},body:JSON.stringify(payload)});
 if(post.status!==200||post.body.status!=='SUCCEEDED')throw new Error(`POST final inesperado: HTTP ${post.status} ${post.raw}`);
 const protocolID=post.body.protocol_id;
 evidence.push({check:'POST SYNC produz representação final durável',status:'PASS',protocol_id:protocolID,http_status:post.status});
 const get=await request(token,`/v1/protocols/${protocolID}`);
 if(get.status!==200||get.raw!==post.raw)throw new Error(`POST/GET divergentes: post=${post.raw} get=${get.raw}`);
 evidence.push({check:'GET reutiliza exatamente os bytes do POST',status:'PASS',body_sha256:createHash('sha256').update(Buffer.from(get.raw)).digest('hex')});
 const [deliveryID,encoded,storedHash]=await waitDelivery(protocolID);
 const deliveryBytes=Buffer.from(encoded,'base64');
 const postHash=createHash('sha256').update(Buffer.from(post.raw)).digest('hex');
 if(storedHash!==postHash||deliveryBytes.toString()!==post.raw)throw new Error(`custódia Pulsar divergente: hash=${storedHash} esperado=${postHash}`);
 const webhook=await waitSink(protocolID);
 if(JSON.stringify(canonical(webhook))!==JSON.stringify(canonical(post.body)))throw new Error('webhook-sink recebeu corpo diferente do POST final');
 evidence.push({check:'Pulsar/webhook-sink preserva wrapper, result_version e bytes',status:'PASS',delivery_id:deliveryID,body_sha256:postHash,webhook_json_equal:true});
 console.log(JSON.stringify({status:'PASS',protocol_id:protocolID,evidence},null,2));
}catch(error){evidence.push({check:'jornada pública POST/GET/webhook',status:'FAIL',error:error.message.split('\n')[0]});console.log(JSON.stringify({status:'FAIL',evidence},null,2));process.exitCode=1}
finally{await writeFile('hub/evidence/r2/execution/representation-runtime-latest.json',JSON.stringify(evidence,null,2)+'\n');await browser.close()}

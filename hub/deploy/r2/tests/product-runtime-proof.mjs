import {createHmac} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

// Qualifica o produto pela fronteira HTTP pública da Órbita. O catálogo e o
// provedor continuam sendo fixtures sintéticas; nenhuma leitura direta de
// operation_steps substitui a admissão ou a consulta pública.
const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const providerSim=process.env.PROVIDER_SIM_URL||'http://localhost:18090';
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
   try{await page.waitForURL(url=>new URL(url).origin===new URL(adminOrigin).origin&&new URL(url).pathname==='/services',{timeout:3500});authenticated=true}catch{await page.waitForTimeout(500)}
  }
  if(!authenticated) throw new Error(`autenticação OIDC não concluída (url=${page.url()})`);
  // A primeira consulta administrativa é a fronteira que materializa o
  // bearer usado também pela API pública da Órbita neste probe.
  await page.evaluate(()=>fetch('/api/atlas/admin/v1/clients?limit=1'));
  if(!token) throw new Error('token administrativo não capturado após autenticação');
  return {browser,token};
 }catch(error){await browser.close();throw error}
}

async function request(token,path,options={}){
 const response=await fetch(`${adminOrigin}/api/orbita${path}`,{...options,headers:{'content-type':'application/json',authorization:`Bearer ${token}`,...(options.headers||{})}});
 const body=await response.json().catch(()=>({}));
 return {status:response.status,body};
}

async function providerEffects(){
 const response=await fetch(`${providerSim}/__qualification/effects`);
 if(!response.ok) throw new Error(`oráculo do provider-sim: HTTP ${response.status}`);
 return response.json();
}

function sql(protocolID){
 const query=`SELECT (SELECT count(*) FROM operation_plans WHERE protocol_id='${protocolID}') AS plans,(SELECT max_parallel FROM operation_plans WHERE protocol_id='${protocolID}') AS max_parallel,(SELECT count(*) FROM operation_steps WHERE protocol_id='${protocolID}' AND state='SUCCEEDED') AS succeeded_steps,(SELECT count(*) FROM operations WHERE protocol_id='${protocolID}') AS operations,(SELECT count(*) FROM protocols WHERE protocol_id='${protocolID}') AS protocols,(SELECT COALESCE(ROUND(EXTRACT(EPOCH FROM (max(created_at)-min(created_at)))*1000)::int,0) FROM operations WHERE protocol_id='${protocolID}') AS operation_start_spread_ms`;
 const raw=execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-F','\\t','-v','ON_ERROR_STOP=1','-c',query],{encoding:'utf8'}).trim();
 const [plans,max_parallel,succeeded_steps,operations,protocols,operation_start_spread_ms]=raw.split('\\t').map(Number);
 return {plans,max_parallel,succeeded_steps,operations,protocols,operation_start_spread_ms};
}

const key=process.env.R2_PRODUCT_IDEMPOTENCY_KEY||`r4-product-http-${Date.now()}`;
const payload={mode:'ASYNC',provider_account_id:'prov-poll-1',service_code:'produto-duplo-r4',service_version:1,input:{cpf:'44444444444',delay_ms:1500}};
const {browser,token}=await authenticatedToken();
const evidence=[];
try{
 const effectsBefore=await providerEffects();
 const first=await request(token,'/v1/protocols',{method:'POST',headers:{'Idempotency-Key':key},body:JSON.stringify(payload)});
 if(first.status!==202||!first.body.protocol_id) throw new Error(`admissão do produto: HTTP ${first.status} ${JSON.stringify(first.body)}`);
 const protocolID=first.body.protocol_id;
 evidence.push({check:'admissão HTTP do produto composto',status:'PASS',protocol_id:protocolID,initial_status:first.status});

 let current=first.body;
 const deadline=Date.now()+30000;
 do{
  await new Promise(resolve=>setTimeout(resolve,750));
  const result=await request(token,`/v1/protocols/${encodeURIComponent(protocolID)}`);
  if(result.status!==200) throw new Error(`consulta do produto: HTTP ${result.status} ${JSON.stringify(result.body)}`);
  current=result.body;
  if(current.status==='SUCCEEDED'||current.status==='FAILED'||current.status==='EXPIRED')break;
 }while(Date.now()<deadline);
 if(current.status!=='SUCCEEDED') throw new Error(`produto não consolidou: ${JSON.stringify(current)}`);
 evidence.push({check:'consulta pública retorna consolidação terminal do produto',status:'PASS',final_status:current.status,final_body:current.final_body||null});

 const durable=sql(protocolID);
 if(durable.plans!==1||durable.max_parallel!==2||durable.succeeded_steps!==2||durable.operations!==2||durable.protocols!==1||durable.operation_start_spread_ms>700) throw new Error(`oráculo DAG/paralelismo inconsistente: ${JSON.stringify(durable)}`);
 evidence.push({check:'oráculo PostgreSQL confirma plano, duas etapas e duas operações',status:'PASS',durable});
 const effectsAfter=await providerEffects();
 if(effectsAfter.effects-effectsBefore.effects!==2||effectsAfter.protocols-effectsBefore.protocols!==2) throw new Error(`efeitos externos não separados por etapa: antes=${JSON.stringify(effectsBefore)} depois=${JSON.stringify(effectsAfter)}`);
 evidence.push({check:'oráculo do provider-sim confirma dois efeitos externos independentes',status:'PASS',before:effectsBefore,after:effectsAfter});

 const duplicate=await request(token,'/v1/protocols',{method:'POST',headers:{'Idempotency-Key':key},body:JSON.stringify(payload)});
 if(duplicate.status!==200||duplicate.body.protocol_id!==protocolID) throw new Error(`idempotência do produto: HTTP ${duplicate.status} ${JSON.stringify(duplicate.body)}`);
 const afterDuplicate=sql(protocolID);
 if(JSON.stringify(afterDuplicate)!==JSON.stringify(durable)) throw new Error(`repetição alterou cardinalidade: antes=${JSON.stringify(durable)} depois=${JSON.stringify(afterDuplicate)}`);
 const effectsAfterDuplicate=await providerEffects();
 if(JSON.stringify(effectsAfterDuplicate)!==JSON.stringify(effectsAfter)) throw new Error(`repetição alterou o oráculo externo: antes=${JSON.stringify(effectsAfter)} depois=${JSON.stringify(effectsAfterDuplicate)}`);
 evidence.push({check:'repetição da chave não cria novo plano ou efeito',status:'PASS',same_protocol:true,durable_after_duplicate:afterDuplicate,external_after_duplicate:effectsAfterDuplicate});
 console.log(JSON.stringify({status:'PASS',key,evidence},null,2));
}catch(error){evidence.push({check:'jornada HTTP do produto composto',status:'FAIL',error:error.message.split('\n')[0]});console.log(JSON.stringify({status:'FAIL',key,evidence},null,2));process.exitCode=1}
finally{await writeFile('hub/evidence/r2/execution/product-http-latest.log',JSON.stringify({key,evidence},null,2)+'\n');await browser.close()}

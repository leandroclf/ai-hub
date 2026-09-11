import {createHmac,randomUUID} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
const playwright=await import(process.env.PLAYWRIGHT_MODULE||'playwright-core');
const chromium=playwright.chromium||playwright.default?.chromium;
if(!chromium)throw new Error('playwright-core sem export chromium');
const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/usr/bin/google-chrome'});
const page=await browser.newPage({viewport:{width:1365,height:900}});
const postgres=`${process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual'}-postgres-1`;
const evidence=[];
let lifecycleProtocol='',lifecycleOperation='',lifecycleCommand='',validClient='',pendingClient='';
function sql(statement){return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim()}
async function authenticate(username,expectedURL){
 if(!await page.locator('#otp').count()){
  await page.getByRole('button',{name:'Entrar',exact:true}).click();
  await page.locator('#username').fill(username);await page.locator('#password').fill('R2-fixture-password!');await page.locator('#kc-login').click();await page.locator('#otp').waitFor();
 }
 for(let attempt=0;attempt<3;attempt++)for(const skew of [0,-1,1]){
  if(!await page.locator('#otp').count())break;
  const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)+skew));
  const digest=createHmac('sha1',Buffer.from('AI-HUB-R2-MFA-KEY-01')).update(counter).digest();
  const offset=digest[digest.length-1]&15;const otp=((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');
  await page.locator('#otp').fill(otp);await page.locator('#kc-login').click();
  try{await page.waitForURL(expectedURL,{timeout:1500});return}catch{}
 }
 throw new Error('OIDC OTP não foi aceito no ensaio curto de ciclo de cliente');
}
async function createClient(code,capacity){
 await page.getByRole('link',{name:'Clientes',exact:true}).click();await page.waitForURL('http://localhost:13000/clients');await page.getByRole('link',{name:'Criar rascunho',exact:true}).click();
 await page.getByLabel('Nome',{exact:true}).fill(`Lifecycle ${code}`);await page.getByLabel('Código estável',{exact:true}).fill(code);await page.locator('#field-environment').selectOption('local');await page.locator('#field-cell_id').fill('r2-cell-a');await page.locator('#field-capacity_units').fill(String(capacity));await page.locator('#field-isolation_class').selectOption('SHARED');
 await page.getByRole('button',{name:'Salvar rascunho',exact:true}).click();await page.waitForURL(`http://localhost:13000/clients/${code}/1`);
}
try{
 await page.goto('http://localhost:13000/services');await authenticate('operadora-a','http://localhost:13000/services');await page.getByRole('heading',{name:'Serviços',exact:true}).waitFor();
 lifecycleProtocol=randomUUID();lifecycleOperation=randomUUID();lifecycleCommand=randomUUID();
 sql(`INSERT INTO protocols(protocol_id,tenant_id,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,accepted_at,client_deadline_at,application_id,cell_id) VALUES('${lifecycleProtocol}','acme','browser-lifecycle-${Date.now()}','browser-lifecycle-hash','{}','ASYNC','QUEUED','${lifecycleCommand}','RUNNING',clock_timestamp(),clock_timestamp()+interval '1 hour','app-acme','r2-cell-a')`);
 sql(`INSERT INTO operations(operation_id,protocol_id,provider_account_id,state,tenant_id,application_id,cell_id) VALUES('${lifecycleOperation}','${lifecycleProtocol}','prov-poll-1','WAITING_FINAL','acme','app-acme','r2-cell-a')`);
 pendingClient=`browser-pending-${Date.now()}`;await createClient(pendingClient,999);await page.getByRole('button',{name:'Validar e simular sem efeitos externos',exact:true}).click();await page.getByRole('alert').filter({hasText:'capacidade qualificada insuficiente; solicite provisionamento'}).waitFor();await page.getByRole('status').filter({hasText:'Validação com pendências'}).waitFor();
 evidence.push({check:'Admin client activation exposes qualified-capacity pending state',status:'PASS',resource:pendingClient});
 validClient=`browser-active-${Date.now()}`;await createClient(validClient,1);await page.getByRole('button',{name:'Validar e simular sem efeitos externos',exact:true}).click();await page.getByRole('status').filter({hasText:'Validação aprovada'}).waitFor();await page.getByLabel('Motivo da ação',{exact:true}).fill('ativação do cliente sintético');await page.getByRole('button',{name:'Publicar versão 1',exact:true}).click();await page.getByRole('status').filter({hasText:'Versão 1 publicada.'}).waitFor();
 const inProgressProtocols=Number(sql("SELECT count(*) FROM protocols WHERE tenant_id='acme' AND status IN ('ACCEPTED','RUNNING','WAITING_PROVIDER','RECONCILING','UNKNOWN')"));if(inProgressProtocols<1)throw new Error('fixture não preservou protocolo em andamento');
 await page.getByLabel('Motivo da ação',{exact:true}).fill('suspensão operacional com histórico preservado');await page.getByRole('button',{name:'Suspender novas admissões',exact:true}).click();await page.getByRole('status').filter({hasText:'Novas admissões suspensas. Histórico preservado.'}).waitFor();await page.getByText(/SUSPENDED · Revisão/).waitFor();
 evidence.push({check:'Admin suspension confirms impact, preserves reason and keeps history readable',status:'PASS',resource:validClient,inProgressProtocols});
 console.log(JSON.stringify(evidence,null,2));
}catch(error){evidence.push({check:'admin client lifecycle',status:'FAIL',url:page.url(),error:error.message.split('\n')[0],bodyText:(await page.locator('body').innerText().catch(()=>'' )).slice(0,800)});console.log(JSON.stringify(evidence,null,2));process.exitCode=1}
finally{
 if(lifecycleOperation)try{sql(`DELETE FROM operations WHERE operation_id='${lifecycleOperation}'`)}catch{}
 if(lifecycleProtocol)try{sql(`DELETE FROM protocols WHERE protocol_id='${lifecycleProtocol}'`)}catch{}
 await writeFile('hub/evidence/r2/execution/admin-client-lifecycle-smoke.json',JSON.stringify(evidence,null,2)+'\n');await browser.close();
}

import {createHash,createHmac,randomUUID} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';

const playwright=await import(process.env.PLAYWRIGHT_MODULE||'playwright-core');
const chromium=playwright.chromium||playwright.default?.chromium;
if(!chromium)throw new Error('playwright-core sem export chromium');
const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/usr/bin/google-chrome'});
const page=await browser.newPage({viewport:{width:1365,height:900}});
const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const evidence=[];

function sql(statement){return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim()}
function controlSQL(statement){return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_control','-At','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim()}
function otp(){
 const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)));
 const digest=createHmac('sha1',Buffer.from('AI-HUB-R2-MFA-KEY-01')).update(counter).digest();
 const offset=digest[digest.length-1]&15;
 return ((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');
}
async function authenticate(){
 if(!await page.locator('#otp').count()){
  await page.getByRole('button',{name:'Entrar',exact:true}).click();
  await page.locator('#username').fill('operadora-a');
  await page.locator('#password').fill('R2-fixture-password!');
  await page.locator('#kc-login').click();
  await page.locator('#otp').waitFor(); await page.waitForTimeout((30-(Math.floor(Date.now()/1000)%30))*1000+250);
 }
 for(let attempt=0;attempt<5;attempt++){
  if(await page.locator('#otp').count()){
   await page.locator('#otp').fill(otp());
   await page.locator('#kc-login').click();
  }
  try{
   await page.locator('#otp').waitFor({state:'detached',timeout:4000});
   await page.getByRole('button',{name:'Sair',exact:true}).waitFor({timeout:4000});
   return;
  }catch{await page.waitForTimeout(250)}
 }
 throw new Error('autenticação OIDC com OTP não concluída');
}
async function openCatalog(kind,id){
 await page.evaluate(path=>{history.pushState({},'',path);window.dispatchEvent(new PopStateEvent('popstate'))},`/${kind}/${id}/1`);
 await page.getByRole('heading',{name:kind==='technical-profiles'?'Perfis técnicos':'Ofertas',exact:true}).waitFor();
 const persisted=page.getByText('Conteúdo persistido desta revisão',{exact:true});
 await persisted.click();
 return page.locator('body').innerText();
}
async function createDelivery(protocolID,representation){
 const deliveryID=randomUUID();
 const eventID=randomUUID();
 const hash=createHash('sha256').update(Buffer.from(representation)).digest('hex');
 const encoded=Buffer.from(representation).toString('hex');
 sql(`INSERT INTO deliveries(delivery_id,protocol_id,event_id,destination_url,state,next_attempt_at,tenant_id,cell_id,representation,body_sha256) VALUES('${deliveryID}','${protocolID}','${eventID}','http://legacy-contract-fixture.invalid/webhook','DELIVERED',clock_timestamp(),'acme','r2-cell-a',decode('${encoded}','hex'),'${hash}')`);
 return {deliveryID,eventID};
}
function canonical(value){
 if(Array.isArray(value))return value.map(canonical);
 if(value&&typeof value==='object')return Object.fromEntries(Object.entries(value).sort(([a],[b])=>a.localeCompare(b)).map(([key,item])=>[key,canonical(item)]));
 return value;
}
async function detailFields(){
 return await page.locator('dl.definition-list > div').evaluateAll(rows=>Object.fromEntries(rows.map(row=>[row.querySelector('dt')?.textContent?.trim()||'',row.querySelector('pre')?.textContent||row.querySelector('dd')?.textContent?.trim()||''])));
}
try{
 await page.goto('http://localhost:13000/technical-profiles');
 await authenticate();
 const profileText=await openCatalog('technical-profiles','profile-r4-json');
 if(!profileText.includes('cpf')||!profileText.includes('provider_mode')||!profileText.includes('async_poll'))throw new Error('perfil técnico não exibiu contrato JSON/polling persistido');
 evidence.push({check:'Perfil técnico expõe input/output JSON e polling persistidos',status:'PASS',resource:'profile-r4-json/1'});

 const offerText=await openCatalog('offers','offer-r4-async');
 for(const route of ['prov-poll-1','prov-poll-2','prov-callback-1'])if(!offerText.includes(route))throw new Error(`oferta não exibiu rota ${route}`);
 const accountModes=controlSQL("SELECT string_agg(data->>'provider_mode',',' ORDER BY id) FROM catalog_resources WHERE kind='provider-accounts' AND id IN ('prov-callback-1','prov-poll-1','prov-poll-2') AND tenant_id='acme'");
 if(!accountModes.includes('async_callback')||!accountModes.includes('async_poll'))throw new Error(`modalidades de conta não comprovam polling/callback: ${accountModes}`);
 evidence.push({check:'Oferta publicada conserva rotas polling e callback',status:'PASS',resource:'offer-r4-async/1',routes:['prov-poll-1','prov-poll-2','prov-callback-1'],provider_modes:accountModes.split(',')});

 const source=sql("SELECT protocol_id || E'\\t' || encode(final_representation,'base64') FROM protocols WHERE tenant_id='acme' AND status='SUCCEEDED' AND final_representation IS NOT NULL ORDER BY accepted_at DESC LIMIT 1");
 if(!source)throw new Error('nenhum protocolo materializado disponível para a prova GET/webhook');
 const [protocolID,encodedRepresentation]=source.split('\t');
 const representation=Buffer.from(encodedRepresentation,'base64').toString();
 const {deliveryID}=await createDelivery(protocolID,representation);
 try{
  await page.evaluate(path=>{history.pushState({},'',path);window.dispatchEvent(new PopStateEvent('popstate'))},`/protocols/${protocolID}`);
  await page.getByRole('heading',{name:'Detalhe persistido',exact:true}).waitFor();
  const protocolText=await page.locator('body').innerText();
  const protocolFields=await detailFields();
  await page.evaluate(path=>{history.pushState({},'',path);window.dispatchEvent(new PopStateEvent('popstate'))},`/deliveries/${deliveryID}`);
  await page.getByRole('heading',{name:'Detalhe persistido',exact:true}).waitFor();
  const deliveryText=await page.locator('body').innerText();
  const deliveryFields=await detailFields();
  const finalRaw=String(protocolFields['final body']||protocolFields['final representation']||protocolFields.final_representation||'');
  const webhookRaw=String(deliveryFields.representation||'');
  let finalBody,webhookBody;
  try{finalBody=JSON.parse(finalRaw);webhookBody=JSON.parse(webhookRaw)}catch(error){throw new Error(`representações não são JSON: final=${finalRaw.slice(0,160)} webhook=${webhookRaw.slice(0,160)} causa=${error.message}`)}
  const marker=representation.replace(/[{}\[\]",:]/g,' ').split(/\s+/).filter(Boolean).find(value=>value.length>=8);
  if(!marker||!protocolText.includes(marker)||!deliveryText.includes(marker)||JSON.stringify(canonical(finalBody))!==JSON.stringify(canonical(webhookBody)))throw new Error(`corpo final divergente entre GET/webhook: marcador=${marker}`);
  evidence.push({check:'GET do protocolo e detalhe do webhook expõem o mesmo corpo final',status:'PASS',protocol:protocolID,delivery:deliveryID,body_sha256:createHash('sha256').update(Buffer.from(representation)).digest('hex'),body_equality:'canonical-json-equal',marker});
 }finally{
  sql(`DELETE FROM webhook_audit WHERE resource='${deliveryID}'`);
  sql(`DELETE FROM deliveries WHERE delivery_id='${deliveryID}'`);
 }
 console.log(JSON.stringify(evidence,null,2));
}catch(error){
 evidence.push({check:'contrato legado com polling/callback',status:'FAIL',url:page.url(),error:error.message.split('\n')[0],bodyText:(await page.locator('body').innerText().catch(()=>'' )).slice(0,1200)});
 console.log(JSON.stringify(evidence,null,2));
 process.exitCode=1;
}finally{
 await writeFile('hub/evidence/r2/execution/admin-legacy-contract-smoke.json',JSON.stringify(evidence,null,2)+'\n');
 await browser.close();
}

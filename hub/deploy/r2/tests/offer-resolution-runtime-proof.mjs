import {createHmac} from 'node:crypto';
import {writeFile} from 'node:fs/promises';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

// Qualifica somente a resolução de oferta já publicada. O probe não cria
// catálogo nem promove validação client-side: usa o Atlas real e verifica a
// resposta completa, a rota selecionada e a recusa de conta fora da oferta.
const atlas=process.env.ATLAS_URL||'http://localhost:18081';
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const evidencePath=resolve('hub/evidence/r2/execution/offer-resolution-runtime-latest.json');
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

async function sessionToken(){
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

async function request(token,query){
 const response=await fetch(`${atlas}/v1/offers/resolve?${new URLSearchParams(query)}`,{headers:{authorization:`Bearer ${token}`}});
 const body=await response.json().catch(()=>({}));
 return {status:response.status,body};
}

async function adminRequest(token,path,options={}){
 const response=await fetch(`${atlas}${path}`,{...options,headers:{'content-type':'application/json',authorization:`Bearer ${token}`,...(options.headers||{})}});
 const body=await response.json().catch(()=>({}));
 return {status:response.status,body};
}

async function ensureDraftOffer(token,id,data){
 const path=`/admin/v1/offers/${encodeURIComponent(id)}/1`;
 let current=await adminRequest(token,path);
 if(current.status===404){
  current=await adminRequest(token,'/admin/v1/offers',{method:'POST',body:JSON.stringify({id,version:1,tenant_id:'acme',name:`Fixture negativa ${id}`,data})});
  if(current.status!==201)throw new Error(`criação da fixture negativa ${id}: HTTP ${current.status} ${JSON.stringify(current.body)}`);
 }else if(current.status===200){
  if(current.body.state!=='DRAFT')throw new Error(`fixture negativa ${id} não está em DRAFT: ${current.body.state}`);
  current=await adminRequest(token,path,{method:'PATCH',headers:{'if-match':`"${current.body.revision}"`},body:JSON.stringify({name:`Fixture negativa ${id}`,data})});
  if(current.status!==200)throw new Error(`atualização da fixture negativa ${id}: HTTP ${current.status} ${JSON.stringify(current.body)}`);
 }else throw new Error(`consulta da fixture negativa ${id}: HTTP ${current.status}`);
 const validation=await adminRequest(token,`${path}/validate`,{method:'POST',body:'{}'});
 const state=await adminRequest(token,path);
 return {validation,state};
}

const {browser,token}=await sessionToken();
const checks=[];
try{
 const selected=await request(token,{tenant_id:'acme',application_id:'app-acme',service_code:'protocolo-assincrono',service_version:'1',provider_account_id:'prov-poll-2'});
 if(selected.status!==200||selected.body.selected_route?.provider_account_id!=='prov-poll-2'||selected.body.selected_route?.priority!==2||!selected.body.content_hash||selected.body.selection_reason!=='eligible-published-offer-then-priority'){
  throw new Error(`rota elegível inesperada: HTTP ${selected.status} ${JSON.stringify(selected.body)}`);
 }
 checks.push({check:'oferta publicada seleciona conta explicitamente elegível',status:'PASS',offer_id:selected.body.offer?.id,route:selected.body.selected_route,content_hash:selected.body.content_hash});

 const outside=await request(token,{tenant_id:'acme',application_id:'app-acme',service_code:'protocolo-assincrono',service_version:'1',provider_account_id:'prov-sync-1'});
 if(outside.status!==403||outside.body.error!=='offer_not_eligible')throw new Error(`conta fora da oferta não foi recusada: HTTP ${outside.status} ${JSON.stringify(outside.body)}`);
 checks.push({check:'conta escolhida fora da oferta é recusada antes do efeito',status:'PASS',http_status:outside.status,error_code:outside.body.error});

 const crossTenant=await request(token,{tenant_id:'beta',application_id:'app-acme',service_code:'protocolo-assincrono',service_version:'1',provider_account_id:'prov-poll-2'});
 if(crossTenant.status!==403)throw new Error(`resolução cross-tenant retornou HTTP ${crossTenant.status}`);
 checks.push({check:'resolução não amplia escopo para outro tenant',status:'PASS',http_status:crossTenant.status});

 const invalidCommon={input_schema:{type:'object',properties:{},additionalProperties:false},output_schema:{type:'object',properties:{status:{type:'string'}},required:['status'],additionalProperties:false},data_class:'SYNTHETIC',qualification_id:'r4-fixture-qualification',application_id:'app-acme',target_kind:'services',target_id:'consulta-cadastral',target_version:1,technical_profile_id:'profile-r4-json',technical_profile_version:1,purchase_contract_id:'purchase-r4',purchase_contract_version:1,sale_contract_id:'sale-r4',sale_contract_version:1,client_sla_seconds:30,provider_sla_seconds:5,provider_sla_policy:'MONITOR_ONLY',retry_ttl_seconds:10,finalization_reserve_seconds:5};
 const invalidSLA=await ensureDraftOffer(token,'offer-r4-invalid-sla',{...invalidCommon,modes:['SYNC'],provider_mode:'sync',sync_http_budget_seconds:10,provider_sla_seconds:8,routes:[{provider_account_id:'prov-sync-1',provider_account_version:1,binding_id:'bind-r4-sync',binding_version:1,priority:1,capacity_domain:'r4-sync'}]});
 if(invalidSLA.validation.status!==200||invalidSLA.validation.body.valid!==false||!invalidSLA.validation.body.field_errors?.modes||invalidSLA.state.body.state!=='DRAFT')throw new Error(`bloqueio de SLA/modalidade não foi específico ou a oferta foi ativada: ${JSON.stringify(invalidSLA)}`);
 checks.push({check:'publicação bloqueia SLA/modalidade incompatíveis com erro específico',status:'PASS',offer_id:'offer-r4-invalid-sla',field_errors:invalidSLA.validation.body.field_errors,state:invalidSLA.state.body.state});

 const invalidBinding=await ensureDraftOffer(token,'offer-r4-invalid-binding',{...invalidCommon,modes:['SYNC'],provider_mode:'sync',sync_http_budget_seconds:15,routes:[{provider_account_id:'prov-sync-1',provider_account_version:1,binding_id:'bind-r4-poll-1',binding_version:1,priority:1,capacity_domain:'r4-sync'}]});
 if(invalidBinding.validation.status!==200||invalidBinding.validation.body.valid!==false||!invalidBinding.validation.body.field_errors?.['routes/prov-sync-1']||invalidBinding.state.body.state!=='DRAFT')throw new Error(`bloqueio de vínculo não foi específico ou a oferta foi ativada: ${JSON.stringify(invalidBinding)}`);
 checks.push({check:'publicação bloqueia vínculo de credencial divergente',status:'PASS',offer_id:'offer-r4-invalid-binding',field_errors:invalidBinding.validation.body.field_errors,state:invalidBinding.state.body.state});
 console.log(JSON.stringify({status:'PASS',checks},null,2));
 await writeFile(evidencePath,`${JSON.stringify({status:'PASS',checks},null,2)}\n`);
}catch(error){
 const result={status:'FAIL',checks,error:error.message.split('\n')[0]};
 await writeFile(evidencePath,`${JSON.stringify(result,null,2)}\n`);
 console.log(JSON.stringify(result,null,2));
 process.exitCode=1;
}finally{await browser.close()}

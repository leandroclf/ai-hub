import {createHmac} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

// Seed exclusivamente sintético para a stack local. As entidades de catálogo
// passam pelo contrato admin/v1; somente a qualificação e a célula de
// capacidade, que não possuem endpoint administrativo nesta baseline, são
// provisionadas na tabela de autoridade do laboratório.
const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const atlas=process.env.ATLAS_URL||'http://localhost:18081';
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
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

async function sessionToken(){
 const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/usr/bin/google-chrome'});
 const page=await browser.newPage();
 let token='';
 page.on('request',request=>{
  const authorization=request.headers().authorization||'';
  if(authorization.startsWith('Bearer ')) token=authorization.slice(7);
 });
 try{
  await page.goto(`${adminOrigin}/services`);
  await page.getByRole('button',{name:'Entrar',exact:true}).click();
  await page.locator('#username').fill('operadora-a');
  await page.locator('#password').fill('R2-fixture-password!');
  await page.locator('#kc-login').click();
  await page.locator('#otp').waitFor();
  let authenticated=false;
  for(let attempt=0;attempt<5&&!authenticated;attempt++){
   if(await page.locator('#otp').count()){
    await page.locator('#otp').fill(otp());
    await page.locator('#kc-login').click();
   }
   try{
    await page.waitForURL(`${adminOrigin}/services`,{timeout:4000});
    authenticated=true;
   }catch{
    await page.waitForTimeout(250);
   }
  }
  if(!authenticated) throw new Error('autenticação OIDC com OTP não concluída');
  await page.evaluate(()=>fetch('/api/atlas/admin/v1/clients?limit=1'));
  if(!token) throw new Error('token da sessão administrativa não capturado');
  return {browser,token};
 }catch(error){
  await browser.close();
  throw error;
 }
}

function provisionLabAuthority(){
 const sql=`
INSERT INTO catalog_qualifications(id,adapter_id,environment,evidence_ref,valid_until,state)
VALUES ('r4-fixture-qualification','provider-sim','local','r4/catalog-seed-v1',clock_timestamp()+interval '365 days','QUALIFIED')
ON CONFLICT (id) DO UPDATE SET adapter_id=EXCLUDED.adapter_id,environment=EXCLUDED.environment,evidence_ref=EXCLUDED.evidence_ref,valid_until=EXCLUDED.valid_until,state='QUALIFIED';
INSERT INTO catalog_capacity_cells(cell_id,environment,state,total_units,reserved_units,qualification_ref)
VALUES ('r2-cell-a','local','READY',100,0,'r4-fixture-qualification')
ON CONFLICT (cell_id) DO UPDATE SET environment='local',state='READY',total_units=GREATEST(catalog_capacity_cells.total_units,100),qualification_ref='r4-fixture-qualification';`;
 execFileSync('docker',['exec',postgres,'psql','-U','hub_runtime','-d','hub_control','-v','ON_ERROR_STOP=1','-X','-c',sql],{stdio:['ignore','pipe','pipe']});
}

const schemaInput={type:'object',properties:{cpf:{type:'string'},delay_ms:{type:'integer'}},required:[],additionalProperties:false};
const failureInput={type:'object',properties:{force_fail:{type:'boolean'}},required:['force_fail'],additionalProperties:false};
const schemaOutput={type:'object',properties:{status:{type:'string'},provider_request_id:{type:'string'}},required:['status'],additionalProperties:false};
const common={input_schema:schemaInput,output_schema:schemaOutput,data_class:'SYNTHETIC',qualification_id:'r4-fixture-qualification',client_sla_seconds:30,provider_sla_seconds:5,retry_ttl_seconds:10,finalization_reserve_seconds:5};
const resources=[
 {kind:'applications',id:'app-acme',tenant_id:'acme',name:'Aplicação fixture acme',data:{}},
 {kind:'provider-accounts',id:'prov-sync-1',tenant_id:'acme',name:'Provider sync fixture',data:{provider_account_id:'prov-sync-1',provider_id:'provider-sim',environment:'local',base_url:'http://provider-sim:8090',provider_mode:'sync',auth_type:'NONE',token_ttl_seconds:90}},
 {kind:'provider-accounts',id:'prov-poll-1',tenant_id:'acme',name:'Provider polling fixture A',data:{provider_account_id:'prov-poll-1',provider_id:'provider-sim',environment:'local',base_url:'http://provider-sim:8090',provider_mode:'async_poll',auth_type:'NONE',token_ttl_seconds:90}},
 {kind:'provider-accounts',id:'prov-poll-2',tenant_id:'acme',name:'Provider polling fixture B',data:{provider_account_id:'prov-poll-2',provider_id:'provider-sim',environment:'local',base_url:'http://provider-sim:8090',provider_mode:'async_poll',auth_type:'NONE',token_ttl_seconds:90}},
 {kind:'provider-accounts',id:'prov-callback-1',tenant_id:'acme',name:'Provider callback fixture',data:{provider_account_id:'prov-callback-1',provider_id:'provider-sim',environment:'local',base_url:'http://provider-sim:8090',provider_mode:'async_callback',auth_type:'NONE',token_ttl_seconds:90}},
 {kind:'credential-bindings',id:'bind-r4-sync',tenant_id:'acme',name:'Binding sync fixture',data:{credential_mode:'SHARED_HUB',provider_account_id:'prov-sync-1',secret_ref:'vault://r4/fixture/sync',secret_version:'v1',settlement_party:'HUB'}},
 {kind:'credential-bindings',id:'bind-r4-poll-1',tenant_id:'acme',name:'Binding polling fixture A',data:{credential_mode:'SHARED_HUB',provider_account_id:'prov-poll-1',secret_ref:'vault://r4/fixture/poll-1',secret_version:'v1',settlement_party:'HUB'}},
 {kind:'credential-bindings',id:'bind-r4-poll-2',tenant_id:'acme',name:'Binding polling fixture B',data:{credential_mode:'SHARED_HUB',provider_account_id:'prov-poll-2',secret_ref:'vault://r4/fixture/poll-2',secret_version:'v1',settlement_party:'HUB'}},
 {kind:'credential-bindings',id:'bind-r4-callback',tenant_id:'acme',name:'Binding callback fixture',data:{credential_mode:'SHARED_HUB',provider_account_id:'prov-callback-1',secret_ref:'vault://r4/fixture/callback',secret_version:'v1',settlement_party:'HUB'}},
 {kind:'services',id:'consulta-cadastral',tenant_id:'acme',name:'Consulta cadastral fixture',data:{...common,modes:['SYNC'],adapter_id:'provider-sim',provider_mode:'sync',sync_http_budget_seconds:15}},
 {kind:'services',id:'consulta-cadastral-failure',tenant_id:'acme',name:'Consulta cadastral falha fixture',data:{...common,input_schema:failureInput,modes:['SYNC'],adapter_id:'provider-sim',provider_mode:'sync',sync_http_budget_seconds:15}},
 {kind:'services',id:'protocolo-assincrono',tenant_id:'acme',name:'Protocolo assíncrono fixture',data:{...common,modes:['ASYNC','AUTO'],adapter_id:'provider-sim',provider_mode:'async_poll',auto_wait_seconds:5,sync_http_budget_seconds:0}},
 {kind:'products',id:'produto-duplo-r4',tenant_id:'acme',name:'Produto composto duplo fixture',data:{...common,modes:['ASYNC','AUTO'],provider_mode:'async_poll',auto_wait_seconds:5,steps:[{id:'etapa_a',service_id:'protocolo-assincrono',service_version:1,depends_on:[],required:true,input_mapping:{}},{id:'etapa_b',service_id:'protocolo-assincrono',service_version:1,depends_on:[],required:true,input_mapping:{}}],max_parallel:2,allow_partial:false,consolidation:'ALL_REQUIRED',failure_policy:'STOP'}},
 {kind:'technical-profiles',id:'profile-r4-json',tenant_id:'acme',name:'Perfil JSON fixture',data:{...common,modes:['ASYNC'],provider_mode:'async_poll',media_type:'application/json',input_mapping:{},output_mapping:{}}},
 {kind:'contracts',id:'purchase-r4',tenant_id:'acme',name:'Contrato de compra fixture',data:{contract_kind:'PURCHASE',currency:'BRL',unit_price:'0.10',settlement_party:'HUB',meter:'SUCCESS',unit_scope:'PROTOCOL',strict_balance:false}},
 {kind:'contracts',id:'sale-r4',tenant_id:'acme',name:'Contrato de venda fixture',data:{contract_kind:'SALE',currency:'BRL',unit_price:'1.00',settlement_party:'HUB',meter:'SUCCESS',unit_scope:'PROTOCOL',strict_balance:false}},
 {kind:'clients',id:'client-acme',tenant_id:'acme',name:'Cliente fixture acme',data:{cell_id:'r2-cell-a',capacity_units:1}},
 {kind:'offers',id:'offer-r4-sync',tenant_id:'acme',name:'Oferta consulta cadastral fixture',data:{...common,modes:['SYNC'],provider_mode:'sync',sync_http_budget_seconds:15,application_id:'app-acme',target_kind:'services',target_id:'consulta-cadastral',target_version:1,technical_profile_id:'profile-r4-json',technical_profile_version:1,purchase_contract_id:'purchase-r4',purchase_contract_version:1,sale_contract_id:'sale-r4',sale_contract_version:1,routes:[{provider_account_id:'prov-sync-1',provider_account_version:1,binding_id:'bind-r4-sync',binding_version:1,priority:1,capacity_domain:'r4-sync'}]}},
 {kind:'offers',id:'offer-r4-sync-failure',tenant_id:'acme',name:'Oferta consulta falha fixture',data:{...common,modes:['SYNC'],provider_mode:'sync',sync_http_budget_seconds:15,application_id:'app-acme',target_kind:'services',target_id:'consulta-cadastral-failure',target_version:1,technical_profile_id:'profile-r4-json',technical_profile_version:1,purchase_contract_id:'purchase-r4',purchase_contract_version:1,sale_contract_id:'sale-r4',sale_contract_version:1,routes:[{provider_account_id:'prov-sync-1',provider_account_version:1,binding_id:'bind-r4-sync',binding_version:1,priority:1,capacity_domain:'r4-sync-failure'}]}},
 {kind:'offers',id:'offer-r4-async',tenant_id:'acme',name:'Oferta protocolo assíncrono fixture',data:{...common,modes:['ASYNC','AUTO'],provider_mode:'async_poll',auto_wait_seconds:5,application_id:'app-acme',target_kind:'services',target_id:'protocolo-assincrono',target_version:1,technical_profile_id:'profile-r4-json',technical_profile_version:1,purchase_contract_id:'purchase-r4',purchase_contract_version:1,sale_contract_id:'sale-r4',sale_contract_version:1,routes:[{provider_account_id:'prov-poll-1',provider_account_version:1,binding_id:'bind-r4-poll-1',binding_version:1,priority:1,equivalence_id:'r4-async',capacity_domain:'r4-poll-1'},{provider_account_id:'prov-poll-2',provider_account_version:1,binding_id:'bind-r4-poll-2',binding_version:1,priority:2,equivalence_id:'r4-async',capacity_domain:'r4-poll-2'},{provider_account_id:'prov-callback-1',provider_account_version:1,binding_id:'bind-r4-callback',binding_version:1,priority:3,equivalence_id:'r4-async',capacity_domain:'r4-callback'}]}},
 {kind:'offers',id:'offer-r4-product',tenant_id:'acme',name:'Oferta produto composto fixture',data:{...common,modes:['ASYNC','AUTO'],provider_mode:'async_poll',auto_wait_seconds:5,application_id:'app-acme',target_kind:'products',target_id:'produto-duplo-r4',target_version:1,technical_profile_id:'profile-r4-json',technical_profile_version:1,purchase_contract_id:'purchase-r4',purchase_contract_version:1,sale_contract_id:'sale-r4',sale_contract_version:1,routes:[{provider_account_id:'prov-poll-1',provider_account_version:1,binding_id:'bind-r4-poll-1',binding_version:1,priority:1,equivalence_id:'r4-product',capacity_domain:'r4-poll-1'}]}}
];

async function request(token,path,options={}){
 const response=await fetch(`${atlas}${path}`,{...options,headers:{'content-type':'application/json',authorization:`Bearer ${token}`,...(options.headers||{})}});
 const body=await response.json().catch(()=>({}));
 return {status:response.status,body,headers:response.headers};
}

async function ensureResource(token,r){
 const path=`/admin/v1/${r.kind}/${encodeURIComponent(r.id)}/1`;
 const existing=await request(token,path);
 if(existing.status===200){
  if(existing.body.state==='PUBLISHED') return 'existing';
  if(existing.body.state!=='DRAFT') throw new Error(`${r.kind}/${r.id} existe em estado ${existing.body.state}; limpeza manual não é permitida`);
  const patched=await request(token,path,{method:'PATCH',headers:{'if-match':`"${existing.body.revision}"`},body:JSON.stringify({name:r.name,data:r.data})});
  if(patched.status!==200) throw new Error(`atualização ${r.kind}/${r.id}: HTTP ${patched.status} ${JSON.stringify(patched.body)}`);
  existing.body=patched.body;
 } else {
  if(existing.status!==404) throw new Error(`consulta ${r.kind}/${r.id}: HTTP ${existing.status}`);
  const created=await request(token,`/admin/v1/${r.kind}`,{method:'POST',body:JSON.stringify({id:r.id,version:1,tenant_id:r.tenant_id,name:r.name,data:r.data})});
  if(created.status!==201) throw new Error(`criação ${r.kind}/${r.id}: HTTP ${created.status} ${JSON.stringify(created.body)}`);
  existing.body=created.body;
 }
 const revision=existing.body.revision;
 const validation=await request(token,`${path}/validate`,{method:'POST',body:'{}'});
 if(validation.status!==200||!validation.body.valid) throw new Error(`validação ${r.kind}/${r.id}: HTTP ${validation.status} ${JSON.stringify(validation.body)}`);
 const published=await request(token,`${path}/publish`,{method:'POST',headers:{'if-match':`"${revision}"`},body:JSON.stringify({reason:'seed sintético versionado para qualificação R4',content_hash:validation.body.content_hash})});
 if(published.status!==200||published.body.state!=='PUBLISHED') throw new Error(`publicação ${r.kind}/${r.id}: HTTP ${published.status} ${JSON.stringify(published.body)}`);
 return 'created';
}

provisionLabAuthority();
const {browser,token}=await sessionToken();
try{
 const result=[];
 for(const r of resources) result.push(`${r.kind}/${r.id}:${await ensureResource(token,r)}`);
 console.log(JSON.stringify({status:'PASS',project,resources:result},null,2));
}finally{await browser.close();}

import {execFileSync} from 'node:child_process';
import {createHash,randomUUID} from 'node:crypto';
import {writeFile} from 'node:fs/promises';

const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const orbitaOrigin=process.env.ORBITA_URL||'http://localhost:18080';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const evidencePath='hub/evidence/r2/execution/catalog-legacy-contracts-runtime-latest.json';

function token(){return execFileSync('python3',['hub/deploy/r2/scripts/token.py','operadora-a'],{encoding:'utf8'}).trim()}
function sql(statement){return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-F','\t','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim()}
async function request(url,bearer,options={}){
 const response=await fetch(url,{...options,headers:{'content-type':'application/json',authorization:`Bearer ${bearer}`,...(options.headers||{})}});
 return {status:response.status,body:await response.json().catch(()=>({}))};
}
async function publishResource(bearer,kind,id,data,name){
 const base=`${adminOrigin}/api/atlas/admin/v1/${kind}`;
 const created=await request(base,bearer,{method:'POST',body:JSON.stringify({id,version:1,tenant_id:'acme',name,data})});
 if(created.status!==201)throw new Error(`criação ${kind}/${id}: HTTP ${created.status} ${JSON.stringify(created.body)}`);
 const path=`${base}/${encodeURIComponent(id)}/1`;
 const validation=await request(`${path}/validate`,bearer,{method:'POST',body:'{}'});
 if(validation.status!==200||!validation.body.valid)throw new Error(`validação ${kind}/${id}: ${JSON.stringify(validation.body)}`);
 const published=await request(`${path}/publish`,bearer,{method:'POST',headers:{'If-Match':`"${created.body.revision}"`},body:JSON.stringify({reason:'qualificação de contratos legados equivalentes',content_hash:validation.body.content_hash})});
 if(published.status!==200||published.body.state!=='PUBLISHED')throw new Error(`publicação ${kind}/${id}: ${JSON.stringify(published.body)}`);
 return published.body;
}
async function suspendResource(bearer,resource){
 const path=`${adminOrigin}/api/atlas/admin/v1/${resource.kind}/${encodeURIComponent(resource.id)}/${resource.version}/suspend`;
 const suspended=await request(path,bearer,{method:'POST',headers:{'If-Match':`"${resource.revision}"`},body:JSON.stringify({reason:'encerramento de fixture legada local'})});
 if(suspended.status!==200||suspended.body.state!=='SUSPENDED')throw new Error(`suspensão ${resource.kind}/${resource.id}: HTTP ${suspended.status} ${JSON.stringify(suspended.body)}`);
 return suspended.body;
}
async function waitProtocol(bearer,id){
 const deadline=Date.now()+30000;
 let current={};
 while(Date.now()<deadline){
  current=(await request(`${orbitaOrigin}/v1/protocols/${encodeURIComponent(id)}`,bearer,{headers:{'X-Tenant-Id':'acme'}})).body;
  if(['SUCCEEDED','FAILED','EXPIRED','PARTIALLY_SUCCEEDED'].includes(current.status))return current;
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 throw new Error(`protocolo não terminou: ${JSON.stringify(current)}`);
}
function createDelivery(protocolID,representation){
 const deliveryID=randomUUID();
 const eventID=randomUUID();
 const hash=createHash('sha256').update(Buffer.from(representation)).digest('hex');
 const encoded=Buffer.from(representation).toString('hex');
 sql(`INSERT INTO deliveries(delivery_id,protocol_id,event_id,destination_url,state,next_attempt_at,tenant_id,cell_id,representation,body_sha256) VALUES('${deliveryID}','${protocolID}','${eventID}','http://legacy-contract-fixture.invalid/webhook','DELIVERED',clock_timestamp(),'acme','r2-cell-a',decode('${encoded}','hex'),'${hash}')`);
 return deliveryID;
}
function canonical(value){
 if(Array.isArray(value))return value.map(canonical);
 if(value&&typeof value==='object')return Object.fromEntries(Object.entries(value).sort(([a],[b])=>a.localeCompare(b)).map(([key,item])=>[key,canonical(item)]));
 return value;
}

const bearer=token();
const suffix=Date.now();
const profileA=`legacy-profile-a-${suffix}`;
const profileB=`legacy-profile-b-${suffix}`;
const accountA=`legacy-account-a-${suffix}`;
const accountB=`legacy-account-b-${suffix}`;
const bindingA=`legacy-binding-a-${suffix}`;
const bindingB=`legacy-binding-b-${suffix}`;
const offerA=`legacy-offer-a-${suffix}`;
const offerB=`legacy-offer-b-${suffix}`;
const result={status:'FAIL',profile:'r2-cat-05-s01-legacy-contracts',suffix};
const publishedLegacyOffers=[];
try{
 const inputTarget={type:'object',properties:{cpf:{type:'string'},delay_ms:{type:'integer'}},required:[],additionalProperties:false};
 const outputTarget={type:'object',properties:{status:{type:'string'},provider_request_id:{type:'string'}},required:['status'],additionalProperties:false};
 const profileBase={modes:['ASYNC'],provider_mode:'async_poll',adapter_id:'provider-sim',client_sla_seconds:30,provider_sla_seconds:5,provider_sla_policy:'MONITOR_ONLY',retry_ttl_seconds:10,finalization_reserve_seconds:5,auto_wait_seconds:5,sync_http_budget_seconds:0,data_class:'SYNTHETIC',media_type:'application/json',output_schema:outputTarget};
 const dataA={...profileBase,input_schema:{type:'object',properties:{cpf_legado:{type:'string'},espera_ms:{type:'integer'}},required:[],additionalProperties:false},input_mapping:{cpf:'cpf_legado',delay_ms:'espera_ms'},output_mapping:{}};
 const dataB={...profileBase,input_schema:{type:'object',properties:{documento:{type:'string'},atraso:{type:'integer'}},required:[],additionalProperties:false},input_mapping:{cpf:'documento',delay_ms:'atraso'},output_mapping:{}};
 const publishedA=await publishResource(bearer,'technical-profiles',profileA,dataA,'Perfil legado A');
 const publishedB=await publishResource(bearer,'technical-profiles',profileB,dataB,'Perfil legado B');
 const publishedAccountA=await publishResource(bearer,'provider-accounts',accountA,{provider_account_id:accountA,provider_id:'provider-sim',environment:'local',base_url:'http://provider-sim:8090',provider_mode:'async_poll',auth_type:'NONE',token_ttl_seconds:90},'Conta externa legado A');
 const publishedAccountB=await publishResource(bearer,'provider-accounts',accountB,{provider_account_id:accountB,provider_id:'provider-sim',environment:'local',base_url:'http://provider-sim:8090',provider_mode:'async_poll',auth_type:'NONE',token_ttl_seconds:90},'Conta externa legado B');
 const publishedBindingA=await publishResource(bearer,'credential-bindings',bindingA,{credential_mode:'SHARED_HUB',provider_account_id:accountA,secret_ref:`vault://r4/legacy/${suffix}/a`,secret_version:'v1',settlement_party:'HUB'},'Binding legado A');
 const publishedBindingB=await publishResource(bearer,'credential-bindings',bindingB,{credential_mode:'SHARED_HUB',provider_account_id:accountB,secret_ref:`vault://r4/legacy/${suffix}/b`,secret_version:'v1',settlement_party:'HUB'},'Binding legado B');
 const offerBase={modes:['ASYNC'],provider_mode:'async_poll',client_sla_seconds:30,provider_sla_seconds:5,provider_sla_policy:'MONITOR_ONLY',retry_ttl_seconds:10,finalization_reserve_seconds:5,auto_wait_seconds:5,sync_http_budget_seconds:0,data_class:'SYNTHETIC',application_id:'app-acme',target_kind:'services',target_id:'protocolo-assincrono',target_version:1,purchase_contract_id:'purchase-r4',purchase_contract_version:1,sale_contract_id:'sale-r4',sale_contract_version:1};
 const offerDataA={...offerBase,technical_profile_id:profileA,technical_profile_version:1,routes:[{provider_account_id:accountA,provider_account_version:1,binding_id:bindingA,binding_version:1,priority:1,equivalence_id:'legacy-a',capacity_domain:'r4-poll-1'}]};
 const offerDataB={...offerBase,technical_profile_id:profileB,technical_profile_version:1,routes:[{provider_account_id:accountB,provider_account_version:1,binding_id:bindingB,binding_version:1,priority:1,equivalence_id:'legacy-b',capacity_domain:'r4-poll-2'}]};
 const publishedOfferA=await publishResource(bearer,'offers',offerA,offerDataA,'Oferta contrato legado A');
 const publishedOfferB=await publishResource(bearer,'offers',offerB,offerDataB,'Oferta contrato legado B');
 publishedLegacyOffers.push(publishedOfferA,publishedOfferB);
 const admissions=[];
 for(const [profile,account,input] of [[profileA,accountA,{cpf_legado:'11111111111',espera_ms:3000}],[profileB,accountB,{documento:'11111111111',atraso:3000}]]){
  const admission=await request(`${orbitaOrigin}/v1/protocols`,bearer,{method:'POST',headers:{'Idempotency-Key':`r2-cat-05-s01-${profile}`,'X-Tenant-Id':'acme'},body:JSON.stringify({mode:'ASYNC',provider_account_id:account,service_code:'protocolo-assincrono',service_version:1,input})});
  if(admission.status!==202||!admission.body.protocol_id)throw new Error(`aceite ${profile} inesperado: ${JSON.stringify(admission.body)}`);
  const protocolID=admission.body.protocol_id;
  const finalProtocol=await waitProtocol(bearer,protocolID);
  const persisted=sql(`SELECT command->'request_body',encode((SELECT final_representation FROM protocols WHERE protocol_id='${protocolID}'),'hex'),(SELECT status FROM protocols WHERE protocol_id='${protocolID}') FROM command_intents WHERE protocol_id='${protocolID}'`);
  const [commandBody,representation,status]=persisted.split('\t');
  if(status!=='SUCCEEDED'||finalProtocol.status!=='SUCCEEDED'||!commandBody.includes('"cpf"')||!commandBody.includes('"delay_ms"')||commandBody.includes('cpf_legado')||commandBody.includes('documento'))throw new Error(`normalização ${profile} inesperada: persisted=${persisted} final=${JSON.stringify(finalProtocol)}`);
  const representationBody=JSON.parse(Buffer.from(representation,'hex').toString());
  const deliveryID=createDelivery(protocolID,JSON.stringify(representationBody));
  try{
   const protocolView=await request(`${adminOrigin}/api/orbita/admin/v1/protocols/${encodeURIComponent(protocolID)}?tenant_id=acme`,bearer);
   const deliveryView=await request(`${adminOrigin}/api/pulsar/admin/v1/deliveries/${encodeURIComponent(deliveryID)}?tenant_id=acme`,bearer);
   if(protocolView.status!==200||deliveryView.status!==200)throw new Error(`GET administrativo falhou: protocolo=${protocolView.status} entrega=${deliveryView.status}`);
   if(JSON.stringify(canonical(protocolView.body.final_representation))!==JSON.stringify(canonical(JSON.parse(deliveryView.body.representation))))throw new Error(`corpo final divergente entre GET e webhook para ${profile}`);
   admissions.push({profile,account,protocol_id:protocolID,client_input:input,provider_input:JSON.parse(commandBody),final_status:finalProtocol.status,delivery_id:deliveryID,get_webhook_body_equality:'canonical-json-equal',representation_sha256:createHash('sha256').update(Buffer.from(representation,'hex')).digest('hex')});
  }finally{
   sql(`DELETE FROM webhook_audit WHERE resource='${deliveryID}'`);
   sql(`DELETE FROM deliveries WHERE delivery_id='${deliveryID}'`);
  }
 }
 if(admissions.length!==2)throw new Error('os dois contratos não foram executados');
 Object.assign(result,{status:'PASS',profiles:{a:{id:profileA,content_hash:publishedA.content_hash,input_fields:['cpf_legado','espera_ms'],mapping:dataA.input_mapping},b:{id:profileB,content_hash:publishedB.content_hash,input_fields:['documento','atraso'],mapping:dataB.input_mapping}},bindings:{a:{account:accountA,binding:bindingA,account_hash:publishedAccountA.content_hash,binding_hash:publishedBindingA.content_hash},b:{account:accountB,binding:bindingB,account_hash:publishedAccountB.content_hash,binding_hash:publishedBindingB.content_hash}},offers:{a:offerA,b:offerB},admissions,semantic_equivalence:'ambos os contratos normalizaram para cpf/delay_ms e terminaram SUCCEEDED'});
}catch(error){result.error=error.message.split('\n')[0];process.exitCode=1}
finally{
 try{
  for(const resource of publishedLegacyOffers) await suspendResource(bearer,resource);
  result.fixture_cleanup='published legacy offers suspended';
 }catch(error){
  result.cleanup_error=error.message.split('\n')[0];result.status='FAIL';process.exitCode=1;
 }
 result.finished_at=new Date().toISOString();await writeFile(evidencePath,`${JSON.stringify(result,null,2)}\n`);console.log(JSON.stringify(result,null,2))
}

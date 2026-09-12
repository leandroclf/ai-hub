import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';

const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const orbitaOrigin=process.env.ORBITA_URL||'http://localhost:18080';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const evidencePath='hub/evidence/r2/execution/catalog-snapshot-upgrade-runtime-latest.json';

function token(){return execFileSync('python3',['hub/deploy/r2/scripts/token.py','operadora-a'],{encoding:'utf8'}).trim()}
function sql(statement,database='hub_core'){return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d',database,'-At','-F','\t','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim()}
async function request(url,bearer,options={}){
 const response=await fetch(url,{...options,headers:{'content-type':'application/json',authorization:`Bearer ${bearer}`,...(options.headers||{})}});
 return {status:response.status,body:await response.json().catch(()=>({}))};
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
async function waitFinance(protocolID){
 const deadline=Date.now()+30000;
 let facts={count:0,cost:0,revenue:0,contracts:''};let ledger=0;
 while(Date.now()<deadline){
  const financeFacts=sql(`SELECT count(*),count(*) FILTER (WHERE kind='COST'),count(*) FILTER (WHERE kind='REVENUE'),coalesce(string_agg(DISTINCT kind||':'||coalesce(contract_id,'')||':'||coalesce(contract_version::text,''),','),'') FROM economic_facts WHERE protocol_id='${protocolID}'`,'hub_finance');
  const [count,cost,revenue,contracts]=financeFacts.split('\t');facts={count:Number(count),cost:Number(cost),revenue:Number(revenue),contracts};
  ledger=Number(sql(`SELECT count(*) FROM ledger_entries WHERE origin_protocol_id='${protocolID}'`,'hub_finance'));
  if(facts.count>=1&&facts.revenue>=1&&ledger>=2)return {facts,ledger};
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 throw new Error(`custódia financeira não convergiu: facts=${JSON.stringify(facts)} ledger=${ledger}`);
}

const bearer=token();
const servicePath=`${adminOrigin}/api/atlas/admin/v1/services/protocolo-assincrono`;
const key=`r2-cat-06-s01-${Date.now()}`;
const result={status:'FAIL',profile:'r2-cat-06-s01-catalog-snapshot-upgrade',key};
try{
 const v1=await request(`${servicePath}/1?tenant_id=acme`,bearer);
 if(v1.status!==200||v1.body.state!=='PUBLISHED')throw new Error(`serviço v1 indisponível: HTTP ${v1.status} ${JSON.stringify(v1.body)}`);
 const admission=await request(`${orbitaOrigin}/v1/protocols`,bearer,{method:'POST',headers:{'Idempotency-Key':key,'X-Tenant-Id':'acme'},body:JSON.stringify({mode:'ASYNC',provider_account_id:'prov-poll-1',service_code:'protocolo-assincrono',service_version:1,input:{delay_ms:3000}})});
 if(admission.status!==202||!admission.body.protocol_id)throw new Error(`aceite v1 inesperado: HTTP ${admission.status} ${JSON.stringify(admission.body)}`);
 const protocolID=admission.body.protocol_id;
 const v2Data=JSON.parse(JSON.stringify(v1.body.data));
 v2Data.client_sla_seconds=45;
 v2Data.provider_sla_seconds=8;
 const existingV2=await request(`${servicePath}/2?tenant_id=acme`,bearer);
 let v2=existingV2;
 if(existingV2.status===404){
  const created=await request(`${adminOrigin}/api/atlas/admin/v1/services?tenant_id=acme`,bearer,{method:'POST',body:JSON.stringify({id:'protocolo-assincrono',version:2,tenant_id:'acme',name:'Protocolo assíncrono fixture v2',data:v2Data})});
  if(created.status!==201)throw new Error(`criação da v2 recusada: HTTP ${created.status} ${JSON.stringify(created.body)}`);
  const validation=await request(`${servicePath}/2/validate?tenant_id=acme`,bearer,{method:'POST',body:'{}'});
  if(validation.status!==200||!validation.body.valid)throw new Error(`validação da v2 recusada: ${JSON.stringify(validation.body)}`);
  v2=await request(`${servicePath}/2/publish?tenant_id=acme`,bearer,{method:'POST',headers:{'If-Match':`"${created.body.revision}"`},body:JSON.stringify({reason:'migração de contrato durante processamento',content_hash:validation.body.content_hash})});
  if(v2.status!==200||v2.body.state!=='PUBLISHED')throw new Error(`publicação da v2 recusada: HTTP ${v2.status} ${JSON.stringify(v2.body)}`);
 }else if(existingV2.status!==200||existingV2.body.state!=='PUBLISHED')throw new Error(`v2 existente inválida: HTTP ${existingV2.status} ${JSON.stringify(existingV2.body)}`);
 const finalProtocol=await waitProtocol(bearer,protocolID);
 const persisted=sql(`SELECT status,config_snapshot->'target'->>'version',config_snapshot->'purchase_contract'->>'version',config_snapshot->'technical_profile'->>'version',CASE WHEN final_representation IS NULL THEN 'false' ELSE 'true' END FROM protocols WHERE protocol_id='${protocolID}' AND tenant_id='acme'`);
 const [status,targetVersion,purchaseVersion,profileVersion,representation]=persisted.split('\t');
 const finance=await waitFinance(protocolID);const financeFacts=finance.facts;const ledgerEntries=finance.ledger;
 if(v1.body.version!==1||v1.body.state!=='PUBLISHED'||v2.body.version!==2||v2.body.state!=='PUBLISHED'||status!=='SUCCEEDED'||targetVersion!=='1'||purchaseVersion!=='1'||profileVersion!=='1'||representation!=='true'||finalProtocol.status!=='SUCCEEDED')throw new Error(`snapshot/result/economia v1 não preservado: v1=${JSON.stringify(v1.body)} v2=${JSON.stringify(v2.body)} persisted=${persisted} final=${JSON.stringify(finalProtocol)} finance=${JSON.stringify(financeFacts)} ledger=${ledgerEntries}`);
 Object.assign(result,{status:'PASS',protocol_id:protocolID,admission_status:admission.status,version_1:{state:v1.body.state,content_hash:v1.body.content_hash},version_2:{state:v2.body.state,version:v2.body.version,changed_fields:['client_sla_seconds','provider_sla_seconds']},snapshot:{target_version:targetVersion,purchase_contract_version:purchaseVersion,technical_profile_version:profileVersion},final_status:finalProtocol.status,representation_materialized:representation==='true',finance:{facts:financeFacts.count,cost_facts:financeFacts.cost,revenue_facts:financeFacts.revenue,contracts:financeFacts.contracts,ledger_entries:ledgerEntries}});
}catch(error){result.error=error.message.split('\n')[0];process.exitCode=1}
finally{result.finished_at=new Date().toISOString();await writeFile(evidencePath,`${JSON.stringify(result,null,2)}\n`);console.log(JSON.stringify(result,null,2))}

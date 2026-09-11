import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';

const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const orbitaOrigin=process.env.ORBITA_URL||'http://localhost:18080';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const evidencePath='hub/evidence/r2/execution/catalog-schema-upgrade-runtime-latest.json';

function token(){return execFileSync('python3',['hub/deploy/r2/scripts/token.py','operadora-a'],{encoding:'utf8'}).trim()}
function sql(statement){return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-F','\t','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim()}
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

const bearer=token();
const servicePath=`${adminOrigin}/api/atlas/admin/v1/services/protocolo-assincrono`;
const key=`r2-cat-05-s03-${Date.now()}`;
const result={status:'FAIL',profile:'r2-cat-05-s03-catalog-schema-upgrade',key};
try{
 const v1=await request(`${servicePath}/1?tenant_id=acme`,bearer);
 if(v1.status!==200||v1.body.state!=='PUBLISHED')throw new Error(`serviço v1 indisponível: HTTP ${v1.status} ${JSON.stringify(v1.body)}`);
 const admission=await request(`${orbitaOrigin}/v1/protocols`,bearer,{method:'POST',headers:{'Idempotency-Key':key,'X-Tenant-Id':'acme'},body:JSON.stringify({mode:'ASYNC',provider_account_id:'prov-poll-1',service_code:'protocolo-assincrono',service_version:1,input:{delay_ms:3000}})});
 if(admission.status!==202||!admission.body.protocol_id)throw new Error(`aceite v1 inesperado: HTTP ${admission.status} ${JSON.stringify(admission.body)}`);
 const protocolID=admission.body.protocol_id;

 let upgradeVersion=2;
 for(let version=2;version<=20;version++){
  const existing=await request(`${servicePath}/${version}?tenant_id=acme`,bearer);
  if(existing.status===404){upgradeVersion=version;break}
  if(existing.status!==200)throw new Error(`consulta da versão ${version} falhou: HTTP ${existing.status}`);
  upgradeVersion=version+1;
 }
 const incompatible=JSON.parse(JSON.stringify(v1.body.data));
 delete incompatible.input_schema.properties.delay_ms;
 incompatible.output_schema.required=[];
 delete incompatible.output_schema.properties.provider_request_id;
 const created=await request(`${adminOrigin}/api/atlas/admin/v1/services?tenant_id=acme`,bearer,{method:'POST',body:JSON.stringify({id:'protocolo-assincrono',version:upgradeVersion,tenant_id:'acme',name:`Protocolo assíncrono schema v${upgradeVersion}`,data:incompatible})});
 if(created.status!==201)throw new Error(`criação da versão ${upgradeVersion} recusada: HTTP ${created.status} ${JSON.stringify(created.body)}`);
 const validation=await request(`${servicePath}/${upgradeVersion}/validate?tenant_id=acme`,bearer,{method:'POST',body:'{}'});
 if(validation.status!==200||!validation.body.valid)throw new Error(`validação da versão ${upgradeVersion} recusada: ${JSON.stringify(validation.body)}`);
 const published=await request(`${servicePath}/${upgradeVersion}/publish?tenant_id=acme`,bearer,{method:'POST',headers:{'If-Match':`"${created.body.revision}"`},body:JSON.stringify({reason:'upgrade incompatível durante processamento v1',content_hash:validation.body.content_hash})});
 if(published.status!==200||published.body.state!=='PUBLISHED')throw new Error(`publicação da versão ${upgradeVersion} recusada: HTTP ${published.status} ${JSON.stringify(published.body)}`);

 const finalProtocol=await waitProtocol(bearer,protocolID);
 const persisted=sql(`SELECT status,config_snapshot->'target'->>'version',config_snapshot->'technical_profile'->>'version',CASE WHEN final_representation IS NULL THEN 'false' ELSE 'true' END,final_body->'result'->>'provider_request_id' FROM protocols WHERE protocol_id='${protocolID}' AND tenant_id='acme'`);
 const [status,targetVersion,profileVersion,representation,providerRequestID]=persisted.split('\t');
 if(v1.body.version!==1||v1.body.state!=='PUBLISHED'||published.body.version!==upgradeVersion||published.body.state!=='PUBLISHED'||status!=='SUCCEEDED'||targetVersion!=='1'||profileVersion!=='1'||representation!=='true'||!providerRequestID||finalProtocol.status!=='SUCCEEDED')throw new Error(`resultado v1 não preservado durante upgrade: v1=${JSON.stringify(v1.body)} v${upgradeVersion}=${JSON.stringify(published.body)} persisted=${persisted} final=${JSON.stringify(finalProtocol)}`);
 Object.assign(result,{status:'PASS',protocol_id:protocolID,admission_status:admission.status,version_1:{state:v1.body.state,content_hash:v1.body.content_hash},upgrade:{version:upgradeVersion,state:published.body.state,content_hash:published.body.content_hash,removed_fields:['input.delay_ms','output.provider_request_id']},snapshot:{target_version:targetVersion,technical_profile_version:profileVersion},final_status:finalProtocol.status,provider_request_id:providerRequestID,representation_materialized:representation==='true'});
}catch(error){result.error=error.message.split('\n')[0];process.exitCode=1}
finally{result.finished_at=new Date().toISOString();await writeFile(evidencePath,`${JSON.stringify(result,null,2)}\n`);console.log(JSON.stringify(result,null,2))}

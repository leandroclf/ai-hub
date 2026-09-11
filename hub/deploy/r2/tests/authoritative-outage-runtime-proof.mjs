import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';

const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const composeFile=process.env.R2_COMPOSE_FILE||'hub/deploy/r2/compose.yaml';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const orbitaOrigin=process.env.ORBITA_URL||'http://localhost:18080';
const evidencePath='hub/evidence/r2/execution/authoritative-outage-runtime-latest.json';

function token(){
 for(let attempt=0;attempt<30;attempt++){
  try{return execFileSync('python3',['hub/deploy/r2/scripts/token.py','operadora-a'],{encoding:'utf8',stdio:['ignore','pipe','ignore']}).trim()}catch{execFileSync('sleep',['1'])}
 }
 throw new Error('não foi possível emitir token do IdP de laboratório após 12 tentativas');
}
function fixtureIP(service){
 return execFileSync('docker',['inspect',`${project}-${service}-1`,'--format','{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'],{encoding:'utf8'}).trim();
}
function compose(args){
 const env={...process.env,R2_PROVIDER_CIDR:`${fixtureIP('provider-sim')}/32`,R2_SINK_CIDR:`${fixtureIP('webhook-sink')}/32`,R2_IDENTITY_CIDR:`${fixtureIP('identity')}/32`};
 return execFileSync('docker',['compose','--project-name',project,'--file',composeFile,...args],{encoding:'utf8',env});
}
function sql(statement){return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim()}
async function request(bearer,key){
 const response=await fetch(`${orbitaOrigin}/v1/protocols`,{method:'POST',headers:{'content-type':'application/json',authorization:`Bearer ${bearer}`,'X-Tenant-Id':'acme','Idempotency-Key':key},body:JSON.stringify({mode:'ASYNC',provider_account_id:'prov-callback-1',service_code:'protocolo-assincrono',service_version:1,input:{delay_ms:1000}}),signal:AbortSignal.timeout(5000)});
 return {status:response.status,body:await response.json().catch(()=>({}))};
}
async function get(bearer,id){
 const response=await fetch(`${orbitaOrigin}/v1/protocols/${encodeURIComponent(id)}`,{headers:{authorization:`Bearer ${bearer}`,'X-Tenant-Id':'acme'},signal:AbortSignal.timeout(5000)});
 return {status:response.status,body:await response.json().catch(()=>({}))};
}
async function waitReady(){
 const deadline=Date.now()+30000;
 while(Date.now()<deadline){
  try{if(execFileSync('docker',['inspect',postgres,'--format','{{.State.Health.Status}}'],{encoding:'utf8'}).trim()==='healthy'&&sql('SELECT 1')==='1')return true}catch{}
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 return false;
}
async function waitServices(){
 const checks=['http://127.0.0.1:18081/healthz/ready','http://127.0.0.1:18080/healthz/ready','http://127.0.0.1:18082/healthz/ready','http://127.0.0.1:18085/realms/ai-hub-r2'];
 const deadline=Date.now()+30000;
 while(Date.now()<deadline){
  try{if((await Promise.all(checks.map(url=>fetch(url,{signal:AbortSignal.timeout(1000)})))).every(response=>response.status===200))return true}catch{}
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 return false;
}
async function waitFinal(bearer,id){
 const deadline=Date.now()+30000;
 let current={};
 while(Date.now()<deadline){
  current=(await get(bearer,id)).body;
  if(['SUCCEEDED','FAILED','EXPIRED','PARTIALLY_SUCCEEDED'].includes(current.status))return current;
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 return current;
}

let bearer=token();
const key=`r2-ope-06-s03-authority-${Date.now()}`;
const result={status:'FAIL',profile:'r2-ope-06-s03-authoritative-outage',key};
let restored=false;
try{
 compose(['stop','postgres']);
 const unavailable=await request(bearer,key);
 const unavailableCode=unavailable.body.code||unavailable.body.error||null;
 const rejectedWithoutAuthority=unavailable.status===503&&unavailableCode==='admission_unavailable'&&!unavailable.body.protocol_id;
 compose(['up','-d','postgres']);
 restored=await waitReady();
 if(!restored)throw new Error('PostgreSQL não voltou ao estado healthy após a recuperação');
 compose(['up','-d','--force-recreate','--no-deps','identity','atlas','orbita','cometa','admin-ui']);
 if(!await waitServices())throw new Error('serviços dependentes não voltaram ao estado ready após a recuperação');
 bearer=token();
 let recovered={status:0,body:{}};
 for(let attempt=0;attempt<10;attempt++){
  recovered=await request(bearer,key);
  if(recovered.status===202)break;
  await new Promise(resolve=>setTimeout(resolve,1000));
 }
 const protocolID=recovered.body.protocol_id||'';
 const final=protocolID?await waitFinal(bearer,protocolID):{status:''};
 const durable=protocolID?sql(`SELECT count(*),min(status) FROM protocols WHERE tenant_id='acme' AND protocol_id='${protocolID}' AND idempotency_key='${key}'`):'0|';
 const [protocolCount,storedStatus]=durable.split('|');
 result.status=rejectedWithoutAuthority&&recovered.status===202&&protocolCount==='1'&&storedStatus==='SUCCEEDED'&&final.status==='SUCCEEDED'?'PASS':'FAIL';
 Object.assign(result,{outage:{postgres_stopped:true,http_status:unavailable.status,code:unavailableCode,rejected_without_authority:rejectedWithoutAuthority,response:unavailable.body},recovery:{postgres_healthy:restored,retry_http_status:recovered.status,response:recovered.body,protocol_id:protocolID,final_status:final.status,durable_protocol_count:Number(protocolCount),durable_status:storedStatus||null,same_idempotency_key:true}});
}catch(error){result.error=error.message.split('\n')[0];process.exitCode=1}
finally{
 try{if(!restored){compose(['up','-d','postgres']);restored=await waitReady()}}catch(error){result.cleanup_error=error.message.split('\n')[0]}
 result.postgres_restored=restored;
 if(result.status==='PASS'&&!restored)result.status='FAIL';
 result.finished_at=new Date().toISOString();
 await writeFile(evidencePath,`${JSON.stringify(result,null,2)}\n`);
 console.log(JSON.stringify(result,null,2));
}

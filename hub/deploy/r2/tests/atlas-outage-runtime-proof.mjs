import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
import {randomUUID} from 'node:crypto';

const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const composeFile='hub/deploy/r2/compose.yaml';
const orbita=process.env.ORBITA_URL||'http://127.0.0.1:18080';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const evidence=[];
let atlasStopped=false;
function compose(args){return execFileSync('docker',['compose','--project-name',project,'--file',composeFile,...args],{encoding:'utf8',stdio:['ignore','pipe','pipe']})}
function sql(statement){return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-F','\t','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim()}
function token(){return execFileSync('python3',['hub/deploy/r2/scripts/token.py','operadora-a'],{encoding:'utf8'}).trim()}
async function request(bearer,path,options={}){
 const response=await fetch(`${orbita}${path}`,{...options,headers:{'content-type':'application/json',authorization:`Bearer ${bearer}`,'X-Tenant-Id':'acme',...(options.headers||{})}});
 return {status:response.status,body:await response.json().catch(()=>({}))};
}
async function waitTerminal(bearer,protocolID,timeoutMs=20000){
 const deadline=Date.now()+timeoutMs;
 let current={};
 while(Date.now()<deadline){
  current=(await request(bearer,`/v1/protocols/${encodeURIComponent(protocolID)}`)).body;
  if(['SUCCEEDED','FAILED','EXPIRED','CANCELLED','PARTIALLY_SUCCEEDED'].includes(current.status))return current;
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 throw new Error(`protocolo ${protocolID} não terminou: ${JSON.stringify(current)}`);
}
async function waitReady(){
 const deadline=Date.now()+30000;
 while(Date.now()<deadline){
  try{if((await fetch(`${orbita}/healthz/ready`)).status===200)return; }catch{}
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 throw new Error('Orbita não ficou pronta');
}
try{
 const bearer=token();
 const warmKey=`r2-atlas-cache-warm-${Date.now()}`;
 const warm=await request(bearer,'/v1/protocols',{method:'POST',headers:{'Idempotency-Key':warmKey},body:JSON.stringify({mode:'ASYNC',provider_account_id:'prov-poll-1',service_code:'protocolo-assincrono',service_version:1,input:{delay_ms:1000}})});
 if(![200,202].includes(warm.status)||!warm.body.protocol_id)throw new Error(`aquecimento recusado: ${warm.status} ${JSON.stringify(warm.body)}`);
 const warmFinal=await waitTerminal(bearer,warm.body.protocol_id);
 evidence.push({check:'Warm-up stores a valid offer projection through the real admission path',status:'PASS',protocol: warm.body.protocol_id,final_status:warmFinal.status});

 compose(['stop','atlas']);
 atlasStopped=true;
 const cachedKey=`r2-atlas-cache-hit-${Date.now()}`;
 const cached=await request(bearer,'/v1/protocols',{method:'POST',headers:{'Idempotency-Key':cachedKey},body:JSON.stringify({mode:'ASYNC',provider_account_id:'prov-poll-1',service_code:'protocolo-assincrono',service_version:1,input:{delay_ms:1000}})});
 if(![200,202].includes(cached.status)||!cached.body.protocol_id)throw new Error(`admissão com Atlas fora não usou projeção: ${cached.status} ${JSON.stringify(cached.body)}`);
 const cachedDurable=sql(`SELECT status FROM protocols WHERE tenant_id='acme' AND idempotency_key='${cachedKey}' AND protocol_id='${cached.body.protocol_id}'`);
 if(!['ACCEPTED','QUEUED','WAITING_PROVIDER','SUCCEEDED','FAILED','EXPIRED','PARTIALLY_SUCCEEDED'].includes(cachedDurable))throw new Error(`aceite com projeção não ficou durável: ${cachedDurable}`);
 evidence.push({check:'Eligible admission reuses a valid local projection while Atlas is unavailable',status:'PASS',protocol:cached.body.protocol_id,http_status:cached.status,accepted_status:cached.body.status||null,durable_status:cachedDurable,limitation:'A continuidade qualificada neste cenário é a resolução/admissão com projeção válida; a execução posterior ainda depende da política de credencial sem cache.'});

 await new Promise(resolve=>setTimeout(resolve,31000));
 const expiredKey=`r2-atlas-cache-expired-${Date.now()}`;
 const expired=await request(bearer,'/v1/protocols',{method:'POST',headers:{'Idempotency-Key':expiredKey},body:JSON.stringify({mode:'ASYNC',provider_account_id:'prov-poll-1',service_code:'protocolo-assincrono',service_version:1,input:{delay_ms:1000}})});
 const expiredCode=expired.body.error||expired.body.code;
 if(expired.status!==403||expiredCode!=='offer_not_eligible')throw new Error(`snapshot expirado foi aceito ou retornou erro incorreto: ${expired.status} ${JSON.stringify(expired.body)}`);
 evidence.push({check:'Expired projection is rejected while Atlas remains unavailable',status:'PASS',http_status:expired.status,error_code:expiredCode});
 console.log(JSON.stringify(evidence,null,2));
}catch(error){
 evidence.push({check:'Atlas outage projection cache',status:'FAIL',error:error.message.split('\n')[0]});
 console.log(JSON.stringify(evidence,null,2));
 process.exitCode=1;
}finally{
 if(atlasStopped){try{compose(['start','atlas']);await waitReady()}catch(error){console.error(`falha ao restaurar Atlas: ${error.message}`);process.exitCode=1}}
 await writeFile('hub/evidence/r2/execution/atlas-outage-runtime-latest.json',JSON.stringify(evidence,null,2)+'\n');
}

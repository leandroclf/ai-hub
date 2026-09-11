import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';

const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const composeFile=process.env.R2_COMPOSE_FILE||'hub/deploy/r2/compose.yaml';
const orbita='http://127.0.0.1:18080';
const admin='http://127.0.0.1:13000';
const evidencePath='hub/evidence/r2/execution/traceability-runtime-latest.json';

function token(){
 for(let attempt=0;attempt<30;attempt++){
  try{return execFileSync('python3',['hub/deploy/r2/scripts/token.py','operadora-a'],{encoding:'utf8',stdio:['ignore','pipe','ignore']}).trim()}catch{execFileSync('sleep',['1'])}
 }
 throw new Error('não foi possível emitir token do IdP de laboratório');
}
function fixtureIP(service){
 return execFileSync('docker',['inspect',`${project}-${service}-1`,'--format','{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'],{encoding:'utf8'}).trim();
}
function sql(statement){
 return execFileSync('docker',['exec',`${project}-postgres-1`,'psql','-U','hub','-d','hub_core','-At','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim();
}
function logs(service){return execFileSync('docker',['logs','--since','45s',`${project}-${service}-1`],{encoding:'utf8',stdio:['ignore','pipe','pipe']})}
async function call(url,bearer,options={}){
 const response=await fetch(url,{...options,headers:{authorization:`Bearer ${bearer}`,'X-Tenant-Id':'acme',...(options.headers||{})},signal:AbortSignal.timeout(10000)});
 return {status:response.status,headers:Object.fromEntries(response.headers),body:await response.json().catch(()=>({}))};
}
async function waitFinal(bearer,id){
 const deadline=Date.now()+30000;
 let current={};
 while(Date.now()<deadline){
  current=(await call(`${orbita}/v1/protocols/${encodeURIComponent(id)}`,bearer)).body;
  if(['SUCCEEDED','FAILED','EXPIRED','PARTIALLY_SUCCEEDED'].includes(current.status))return current;
  await new Promise(resolve=>setTimeout(resolve,500));
 }
 return current;
}

const bearer=token();
const key=`r2-ope-07-s01-trace-${Date.now()}`;
const result={status:'FAIL',profile:'r2-ope-07-s01-traceability',key};
try{
 const before=Date.now();
 const admission=await call(`${orbita}/v1/protocols`,bearer,{method:'POST',headers:{'content-type':'application/json','Idempotency-Key':key},body:JSON.stringify({mode:'ASYNC',provider_account_id:'prov-callback-1',service_code:'protocolo-assincrono',service_version:1,input:{delay_ms:1000}})});
 const protocolID=admission.body.protocol_id||'';
 const traceID=admission.headers['x-trace-id']||'';
 const final=protocolID?await waitFinal(bearer,protocolID):{};
 const persisted=protocolID?sql(`SELECT p.status,COALESCE(ci.command->>'trace_id',''),COALESCE(ci.command->>'protocol_id',''),COALESCE(ci.command->>'tenant_id','') FROM protocols p JOIN command_intents ci ON ci.protocol_id=p.protocol_id WHERE p.protocol_id='${protocolID}'`):'';
 const [storedStatus,storedTrace,storedProtocol,storedTenant]=persisted.split('|');
 const timeline=protocolID?await call(`${admin}/api/orbita/admin/v1/protocols/${encodeURIComponent(protocolID)}/timeline?tenant_id=acme`,bearer):{status:0,body:{}};
 const metrics=await (await fetch(`${orbita}/metrics`,{signal:AbortSignal.timeout(5000)})).text();
 const forbiddenLabel=/\{[^\n]*(?:protocol_id|tenant_id|operation_id|attempt_id|idempotency_key)=/i.test(metrics);
 const histogram=metrics.split('\n').filter(line=>line.startsWith('hub_http_duration_seconds_bucket')&&line.includes('le=')).length>0;
 const orbitaLogs=logs('orbita');
 const cometaLogs=logs('cometa');
 const traceInOrbita=storedTrace!==''&&orbitaLogs.includes(storedTrace);
 const traceInCometa=storedTrace!==''&&cometaLogs.includes(storedTrace);
 const elapsedMs=Date.now()-before;
 result.status=admission.status===202&&final.status==='SUCCEEDED'&&protocolID!==''&&traceID!==''&&storedStatus==='SUCCEEDED'&&storedTrace!==''&&storedProtocol===protocolID&&storedTenant==='acme'&&timeline.status===200&&Array.isArray(timeline.body.items)&&timeline.body.items.length>=1&&!forbiddenLabel&&histogram&&traceInOrbita&&traceInCometa?'PASS':'FAIL';
 Object.assign(result,{admission:{status:admission.status,http_trace_id_present:traceID!=='',protocol_id:protocolID},final_status:final.status,durable:{status:storedStatus,trace_id:storedTrace,protocol_id:storedProtocol,tenant_id:storedTenant},timeline:{status:timeline.status,events:Array.isArray(timeline.body.items)?timeline.body.items.length:0},metrics:{histogram_series_present:histogram,forbidden_business_id_labels:forbiddenLabel},logs:{trace_in_orbita:traceInOrbita,trace_in_cometa:traceInCometa},elapsed_ms:elapsedMs});
}catch(error){result.error=error.message.split('\n')[0];process.exitCode=1}
finally{result.finished_at=new Date().toISOString();await writeFile(evidencePath,`${JSON.stringify(result,null,2)}\n`);console.log(JSON.stringify(result,null,2))}

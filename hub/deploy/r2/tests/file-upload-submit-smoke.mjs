import {createHash,createHmac} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const localstack=process.env.R2_LOCALSTACK_CONTAINER||`${project}-localstack-1`;
const origin=process.env.ORBITA_URL||'http://localhost:18080';
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const modulePath=resolve(process.env.PLAYWRIGHT_MODULE||'hub/evidence/screenshots/node_modules/playwright-core/index.js');
const pw=await import(pathToFileURL(modulePath).href);
const chromium=pw.chromium||pw.default?.chromium;
if(!chromium)throw new Error('playwright-core sem export chromium');
const evidence=[];
const suffix=`${Date.now()}`;
const policy=`SYNTHETIC-INPUT-${suffix}`;
let fileID='';
let browser;

function sql(statement){return execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-v','ON_ERROR_STOP=1','-c',statement],{encoding:'utf8'}).trim()}
function otp(skew=0){
 const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)+skew));
 const digest=createHmac('sha1',Buffer.from('JBSWY3DPEHPK3PXP')).update(counter).digest();
 const offset=digest[digest.length-1]&15;
 return ((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');
}
async function authenticatedToken(){
 const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/usr/bin/google-chrome'});
 const page=await browser.newPage();let bearer='';
 page.on('request',request=>{const value=request.headers().authorization||'';if(value.startsWith('Bearer '))bearer=value.slice(7)});
 try{
  await page.goto(`${adminOrigin}/services`);
  await page.getByRole('button',{name:'Entrar',exact:true}).click();
  await page.locator('#username').fill('operadora-a');
  await page.locator('#password').fill('R2-fixture-password!');
  await page.locator('#kc-login').click();
  await page.locator('#otp').waitFor();
  for(let attempt=0;attempt<8&&!bearer;attempt++)for(const skew of [0,-1,1]){
   if(!await page.locator('#otp').count())break;
   await page.locator('#otp').fill(otp(skew));await page.locator('#kc-login').click();
   try{await page.getByRole('button',{name:'Sair',exact:true}).waitFor({timeout:1500});break}catch{}
  }
  if(!bearer)throw new Error('bearer OIDC não capturado');
  return {bearer,browser};
 }catch(error){await browser.close();throw error}
}
async function request(bearer,path,options={}){
 const response=await fetch(`${origin}${path}`,{...options,headers:{authorization:`Bearer ${bearer}`,'content-type':'application/json',...(options.headers||{})}});
 const body=await response.json().catch(()=>({}));
 return {status:response.status,body};
}
try{
 const session=await authenticatedToken();
 browser=session.browser;
 const {bearer}=session;
 sql(`INSERT INTO object_retention_policies(class,region,purpose,retention_seconds,max_bytes,allowed_types,approved_by) VALUES('${policy}','fixture-local','INPUT',3600,1048576,'["application/octet-stream"]','file-upload-smoke')`);
 const content=Buffer.from(`ai-hub-file-upload-${suffix}`);
 const sha=createHash('sha256').update(content).digest('hex');
 const created=await request(bearer,'/v1/files',{method:'POST',body:JSON.stringify({size_bytes:content.length,sha256:sha,content_type:'application/octet-stream',purpose:'INPUT',class:policy,region:'fixture-local'})});
 if(created.status!==201||!created.body.file_ref?.file_id||!created.body.uploads?.length)throw new Error(`criação da sessão: HTTP ${created.status} ${JSON.stringify(created.body)}`);
 fileID=created.body.file_ref.file_id;
 const upload=created.body.uploads[0];
 const put=await fetch(upload.url,{method:'PUT',headers:upload.headers,body:content});
 if(put.status!==200)throw new Error(`upload direto: HTTP ${put.status}`);
 const completed=await request(bearer,`/v1/files/${fileID}/complete`,{method:'POST',body:JSON.stringify({parts:[]})});
 if(completed.status!==200||completed.body.state!=='READY'||completed.body.sha256!==sha)throw new Error(`confirmação: HTTP ${completed.status} ${JSON.stringify(completed.body)}`);
 evidence.push({check:'upload HTTP direto e confirmação de FileRef',status:'PASS',file_ref:fileID,state:completed.body.state,sha256:sha});
 const key=`file-upload-submit-${suffix}`;
 const admitted=await request(bearer,'/v1/protocols',{method:'POST',headers:{'Idempotency-Key':key},body:JSON.stringify({file_refs:[fileID],mode:'SYNC',provider_account_id:'prov-sync-1',service_code:'consulta-cadastral',service_version:1,input:{cpf:'11111111111'}})});
 if(![200,202,504].includes(admitted.status)||!admitted.body.protocol_id)throw new Error(`admissão com FileRef: HTTP ${admitted.status} ${JSON.stringify(admitted.body)}`);
 const pinned=sql(`SELECT count(*) FROM object_retention_pins WHERE file_id='${fileID}' AND obligation_id='${admitted.body.protocol_id}'`);
 if(pinned!=='1')throw new Error(`admissão não criou pin de custódia: protocolo=${admitted.body.protocol_id} pins=${pinned}`);
 evidence.push({check:'admissão real referencia FileRef confirmado',status:'PASS',protocol_id:admitted.body.protocol_id,http_status:admitted.status,acceptance_preserved:true,retention_pin:pinned});
 console.log(JSON.stringify(evidence,null,2));
}catch(error){
 evidence.push({check:'upload HTTP e submit com FileRef',status:'FAIL',error:error.message.split('\n')[0]});
 console.log(JSON.stringify(evidence,null,2));
 process.exitCode=1;
}finally{
 if(fileID){
  try{
   const object=sql(`SELECT object_key,COALESCE(object_version,'') FROM file_refs WHERE id='${fileID}'`);
   if(object){const [key,version]=object.split('|');execFileSync('docker',['exec',localstack,'awslocal','s3api','delete-object','--bucket','r2-custody','--key',key,...(version?[ '--version-id',version]:[])],{stdio:'ignore'});}
   sql(`DELETE FROM object_retention_pins WHERE file_id='${fileID}'`);
   sql(`DELETE FROM file_refs WHERE id='${fileID}'`);
  }catch(error){console.error(`cleanup FileRef falhou: ${error.message}`)}
 }
 try{sql(`DELETE FROM object_retention_policies WHERE class='${policy}'`)}catch(error){console.error(`cleanup policy falhou: ${error.message}`)}
 await writeFile('hub/evidence/r2/execution/file-upload-submit-smoke.json',JSON.stringify(evidence,null,2)+'\n');
 if(browser)await browser.close();
}

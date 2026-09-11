import {createHmac} from 'node:crypto';
import {execFileSync, spawnSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
import {resolve} from 'node:path';
import {pathToFileURL} from 'node:url';

// Executa a carga de evidencia pela mesma autenticacao OIDC usada pelo
// navegador. O bearer fica somente em memoria e no ambiente do processo
// filho; nunca e impresso nem persistido no artefato de evidencia.
const project=process.env.R2_COMPOSE_PROJECT||'ai_hub_r3qual';
const postgres=process.env.R2_POSTGRES_CONTAINER||`${project}-postgres-1`;
const adminOrigin=process.env.ADMIN_UI_URL||'http://localhost:13000';
const traffic=resolve('hub/evidence/generate_traffic.sh');
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

async function authenticatedToken(){
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
  if(!authenticated) throw new Error(`autenticacao OIDC com OTP nao concluida (url=${page.url()})`);
  await page.evaluate(()=>fetch('/api/atlas/admin/v1/clients?limit=1'));
  if(!token) throw new Error('token administrativo nao capturado apos autenticacao');
  return {browser,token};
 }catch(error){
  await browser.close();
  throw error;
 }
}

function authorityRows(prefix){
 const query=`SELECT idempotency_key,protocol_id,status FROM protocols WHERE idempotency_key LIKE '${prefix}-%' ORDER BY idempotency_key`;
 const raw=execFileSync('docker',['exec',postgres,'psql','-U','hub','-d','hub_core','-At','-F','\t','-v','ON_ERROR_STOP=1','-c',query],{encoding:'utf8'});
 return raw.trim()?raw.trim().split('\n').map(line=>{
  const [idempotency_key,protocol_id,status]=line.split('\t');
  return {idempotency_key,protocol_id,status};
 }):[];
}

const prefix=process.env.R2_EVIDENCE_PREFIX||`r4-authorized-${Date.now()}`;
const expected={
 [`${prefix}-sync-ok-001`]: 'SUCCEEDED',
 [`${prefix}-sync-fail-001`]: 'FAILED',
 [`${prefix}-async-poll-a-001`]: 'SUCCEEDED',
 [`${prefix}-async-poll-b-001`]: 'SUCCEEDED',
 [`${prefix}-async-callback-001`]: 'SUCCEEDED',
 [`${prefix}-auto-001`]: 'SUCCEEDED',
 [`${prefix}-strict-ok-001`]: 'SUCCEEDED',
 [`${prefix}-dedicated-001`]: 'SUCCEEDED'
};

const {browser,token}=await authenticatedToken();
try{
 const child=spawnSync('bash',[traffic],{
  env:{...process.env,R2_BEARER_TOKEN:token,R2_EVIDENCE_PREFIX:prefix},
  encoding:'utf8',
  maxBuffer:4*1024*1024
 });
 const output=child.stdout||'';
 const errorOutput=child.stderr||'';
 await writeFile('hub/evidence/r2/execution/authorized-load-latest.log',output+errorOutput);
 if(child.error) throw child.error;
 if(child.status!==0) throw new Error(`carga terminou com codigo ${child.status}`);

 let rows=[];
 const deadline=Date.now()+30000;
 do{
  rows=authorityRows(prefix);
  const complete=Object.entries(expected).every(([key,status])=>rows.some(row=>row.idempotency_key===key&&row.status===status));
  if(complete) break;
  await new Promise(resolve=>setTimeout(resolve,1000));
 }while(Date.now()<deadline);

 const byKey=Object.fromEntries(rows.map(row=>[row.idempotency_key,row]));
 const checks=Object.entries(expected).map(([key,status])=>({key,expected:status,observed:byKey[key]?.status||'MISSING',protocol_id:byKey[key]?.protocol_id||null,status:(byKey[key]?.status===status?'PASS':'FAIL')}));
 const duplicateIds=[...output.matchAll(/"protocol_id"\s*:\s*"([0-9a-f-]{36})"/g)].map(match=>match[1]);
 const duplicateCount=duplicateIds.filter(id=>id===byKey[`${prefix}-sync-ok-001`]?.protocol_id).length;
 const result={status:checks.every(check=>check.status==='PASS')&&duplicateCount>=2?'PASS':'FAIL',prefix,checks,idempotency:{same_protocol_observations:duplicateCount,expected_at_least:2},rows};
 console.log(JSON.stringify(result,null,2));
 if(result.status!=='PASS') process.exitCode=1;
}finally{await browser.close();}

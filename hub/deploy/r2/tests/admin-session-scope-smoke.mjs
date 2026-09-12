import {createHmac} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
const playwright=await import(process.env.PLAYWRIGHT_MODULE||'playwright-core');
const chromium=playwright.chromium||playwright.default?.chromium;
if(!chromium)throw new Error('playwright-core sem export chromium');
const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/usr/bin/google-chrome'});
const page=await browser.newPage({viewport:{width:1365,height:900}});const evidence=[];
async function authenticate(username,expectedURL){
 if(!await page.locator('#otp').count()){await page.getByRole('button',{name:'Entrar',exact:true}).click();await page.locator('#username').fill(username);await page.locator('#password').fill('R2-fixture-password!');await page.locator('#kc-login').click();await page.locator('#otp').waitFor();await page.waitForTimeout((30-(Math.floor(Date.now()/1000)%30))*1000+250)}
 for(let attempt=0;attempt<3;attempt++)for(const skew of [0,-1,1]){if(!await page.locator('#otp').count())break;const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)+skew));const digest=createHmac('sha1',Buffer.from('AI-HUB-R2-MFA-KEY-01')).update(counter).digest();const offset=digest[digest.length-1]&15;const otp=((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');await page.locator('#otp').fill(otp);await page.locator('#kc-login').click();try{await page.waitForURL(expectedURL,{timeout:1500});return}catch{}}
 throw new Error(`OIDC OTP não foi aceito para ${username}`);
}
try{
 const token=execFileSync('python3',['hub/deploy/r2/scripts/token.py','leitor-a'],{encoding:'utf8'}).trim();const denied=await fetch('http://localhost:18081/admin/v1/applications?tenant_id=beta&limit=1',{headers:{Authorization:`Bearer ${token}`}});if(denied.status!==403)throw new Error(`consulta cross-tenant retornou HTTP ${denied.status}`);
 await page.goto('http://localhost:13000/applications');await authenticate('leitor-a','http://localhost:13000/applications');await page.getByRole('heading',{name:'Aplicações',exact:true}).waitFor();const tenant=await page.getByText('acme',{exact:true}).first().innerText();const crossTenantInput=await page.getByLabel('Cliente em consulta',{exact:true}).count();const body=await page.locator('body').innerText();if(tenant!=='acme'||crossTenantInput||body.includes('beta'))throw new Error(`UI expôs contexto indevido: tenant=${tenant} cross_input=${crossTenantInput}`);
 evidence.push({check:'Admin denies tenant reader URL outside scope at authority and keeps UI context scoped',status:'PASS',authority_status:denied.status,visible_tenant:tenant,cross_tenant_input:crossTenantInput});console.log(JSON.stringify(evidence,null,2));
}catch(error){evidence.push({check:'admin session scope',status:'FAIL',url:page.url(),error:error.message.split('\n')[0],bodyText:(await page.locator('body').innerText().catch(()=>'' )).slice(0,1000)});console.log(JSON.stringify(evidence,null,2));process.exitCode=1}
finally{await writeFile('hub/evidence/r2/execution/admin-session-scope-smoke.json',JSON.stringify(evidence,null,2)+'\n');await browser.close()}

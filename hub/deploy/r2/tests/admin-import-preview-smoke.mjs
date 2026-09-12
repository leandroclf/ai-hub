import {createHmac} from 'node:crypto';
import {writeFile} from 'node:fs/promises';
const playwright=await import(process.env.PLAYWRIGHT_MODULE||'playwright-core');
const chromium=playwright.chromium||playwright.default?.chromium;
if(!chromium)throw new Error('playwright-core sem export chromium');
const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/usr/bin/google-chrome'});
const page=await browser.newPage({viewport:{width:1365,height:900}});
const evidence=[];
async function authenticate(username,expectedURL){
 if(!await page.locator('#otp').count()){await page.getByRole('button',{name:'Entrar',exact:true}).click();await page.locator('#username').fill(username);await page.locator('#password').fill('R2-fixture-password!');await page.locator('#kc-login').click();await page.locator('#otp').waitFor();await page.waitForTimeout((30-(Math.floor(Date.now()/1000)%30))*1000+250)}
 for(let attempt=0;attempt<3;attempt++)for(const skew of [0,-1,1]){if(!await page.locator('#otp').count())break;const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)+skew));const digest=createHmac('sha1',Buffer.from('AI-HUB-R2-MFA-KEY-01')).update(counter).digest();const offset=digest[digest.length-1]&15;const otp=((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');await page.locator('#otp').fill(otp);await page.locator('#kc-login').click();try{await page.waitForURL(expectedURL,{timeout:1500});return}catch{}}
 throw new Error('OIDC OTP não foi aceito no ensaio de importação');
}
try{
 await page.goto('http://localhost:13000/imports');await authenticate('operadora-a','http://localhost:13000/imports');await page.getByRole('heading',{name:'Importações em staging',exact:true}).waitFor();
 await page.locator('input[type=file]').setInputFiles('hub/deploy/r2/tests/fixtures/admin-import-preview.json');await page.getByRole('button',{name:'Sanitizar e criar preview',exact:true}).click();await page.getByRole('status').filter({hasText:'2 operações importadas'}).waitFor();await page.getByRole('status').filter({hasText:'STAGED'}).waitFor();
 const tableText=await page.locator('table').innerText();
 if(!tableText.includes('GET')||!tableText.includes('/analise')||!tableText.includes('BEARER')||!tableText.includes('IMPORTED_NOT_EXECUTABLE'))throw new Error('preview não expôs estado/importação sanitizada');
 if(tableText.includes('segredo-que-nao-pode-ser-retido'))throw new Error('preview reteve valor sensível do fixture');
 evidence.push({check:'Admin import preview shows sanitized existing/new endpoint diff without publication',status:'PASS',operations:2,state:'STAGED'});
 evidence.push({check:'Admin imported endpoint remains explicitly unavailable without adapter',status:'PASS',state:'IMPORTED_NOT_EXECUTABLE'});
 console.log(JSON.stringify(evidence,null,2));
}catch(error){evidence.push({check:'admin import preview',status:'FAIL',url:page.url(),error:error.message.split('\n')[0],bodyText:(await page.locator('body').innerText().catch(()=>'' )).slice(0,800)});console.log(JSON.stringify(evidence,null,2));process.exitCode=1}
finally{await writeFile('hub/evidence/r2/execution/admin-import-preview-smoke.json',JSON.stringify(evidence,null,2)+'\n');await browser.close()}

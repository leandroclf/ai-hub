import {createHmac} from 'node:crypto';
import {writeFile} from 'node:fs/promises';
const playwright=await import(process.env.PLAYWRIGHT_MODULE||'playwright-core');
const chromium=playwright.chromium||playwright.default?.chromium;
if(!chromium)throw new Error('playwright-core sem export chromium');
const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/usr/bin/google-chrome'});
const page=await browser.newPage({viewport:{width:1365,height:900}});const evidence=[];
async function authenticate(username,expectedURL){
 if(!await page.locator('#otp').count()){await page.getByRole('button',{name:'Entrar',exact:true}).click();await page.locator('#username').fill(username);await page.locator('#password').fill('R2-fixture-password!');await page.locator('#kc-login').click();await page.locator('#otp').waitFor()}
 for(let attempt=0;attempt<3;attempt++)for(const skew of [0,-1,1]){if(!await page.locator('#otp').count())break;const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)+skew));const digest=createHmac('sha1',Buffer.from('AI-HUB-R2-MFA-KEY-01')).update(counter).digest();const offset=digest[digest.length-1]&15;const otp=((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');await page.locator('#otp').fill(otp);await page.locator('#kc-login').click();try{await page.waitForURL(expectedURL,{timeout:1500});return}catch{}}
 throw new Error('OIDC OTP não foi aceito no ensaio SLA');
}
try{
 await page.goto('http://localhost:13000/sla-reports');await authenticate('operadora-a','http://localhost:13000/sla-reports');await page.getByRole('heading',{name:'SLA bilateral',exact:true}).waitFor();await page.getByRole('status').filter({hasText:'watermark'}).waitFor();const freshness=await page.getByRole('status').filter({hasText:'watermark'}).innerText();if(!freshness.includes('generated')&&!freshness.includes('Atualização da projeção'))throw new Error(`watermark sem atualização: ${freshness}`);if(!freshness.includes('atraso observado'))throw new Error(`watermark sem atraso: ${freshness}`);evidence.push({check:'Admin SLA panel shows generated time, watermark and observed lag',status:'PASS',freshness});
 const downloadPromise=page.waitForEvent('download');await page.getByRole('button',{name:'Exportar CSV limitado',exact:true}).click();const download=await downloadPromise;const exportStatus=await page.getByRole('status').filter({hasText:'Exportação limitada'}).innerText();if(download.suggestedFilename()!=='sla-report.csv'||!exportStatus.includes('auditada'))throw new Error(`exportação sem recorte/auditoria visual: arquivo=${download.suggestedFilename()} status=${exportStatus}`);evidence.push({check:'Admin SLA export is limited to authorized filters and records an audit action',status:'PASS',filename:download.suggestedFilename(),exportStatus});
 console.log(JSON.stringify(evidence,null,2));
}catch(error){evidence.push({check:'admin SLA freshness',status:'FAIL',url:page.url(),error:error.message.split('\n')[0],bodyText:(await page.locator('body').innerText().catch(()=>'' )).slice(0,1000)});console.log(JSON.stringify(evidence,null,2));process.exitCode=1}
finally{await writeFile('hub/evidence/r2/execution/admin-sla-freshness-smoke.json',JSON.stringify(evidence,null,2)+'\n');await browser.close()}

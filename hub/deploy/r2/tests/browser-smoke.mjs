import {createHmac} from 'node:crypto';
import {writeFile} from 'node:fs/promises';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'/tmp/ai-hub-r2-browser/node_modules/playwright/index.mjs');
const browser=await chromium.launch({headless:true});
const page=await browser.newPage({viewport:{width:1365,height:900}});
const evidence=[];
try {
 await page.goto('http://localhost:13000/services');
 await page.getByRole('button',{name:'Entrar',exact:true}).click();
 await page.locator('#username').fill('operadora-a');
 await page.locator('#password').fill('R2-fixture-password!');
 await page.locator('#kc-login').click();
 await page.locator('#otp').waitFor();
 const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)));
 const digest=createHmac('sha1',Buffer.from('JBSWY3DPEHPK3PXP')).update(counter).digest();
 const offset=digest[digest.length-1]&15;const otp=((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');
 await page.locator('#otp').fill(otp);await page.locator('#kc-login').click();
 await page.waitForURL('http://localhost:13000/services');
 await page.getByRole('heading',{name:'Serviços',exact:true}).waitFor();
 evidence.push({check:'OIDC Authorization Code PKCE + password + OTP in Chromium',status:'PASS'});
 await page.getByRole('link',{name:'Provedores',exact:true}).click();
 await page.waitForURL('http://localhost:13000/providers');
 await page.getByRole('heading',{name:'Provedores',exact:true}).waitFor();
 evidence.push({check:'URL navigation and real authenticated provider list',status:'PASS'});
 const storage=await page.evaluate(()=>({local:Object.keys(localStorage),session:Object.keys(sessionStorage)}));
 if(storage.local.length||storage.session.includes('atlas.pkce'))throw new Error('Unexpected persisted session material');
 evidence.push({check:'No persistent token and PKCE transaction removed',status:'PASS',storageKeys:storage});
 await page.screenshot({path:'hub/evidence/r2/execution/admin-providers.png',fullPage:true});
 await page.setViewportSize({width:390,height:844});
 const overflow=await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth);
 const dimensions=await page.evaluate(()=>Array.from(document.querySelectorAll('.app-shell,.app-header,.app-nav,.app-main')).map(e=>({element:e.className,width:e.getBoundingClientRect().width,right:e.getBoundingClientRect().right,display:getComputedStyle(e).display,direction:getComputedStyle(e).flexDirection})));
 evidence.push({check:'390px viewport has no page overflow',status:overflow?'FAIL':'PASS',dimensions});
 if(process.env.R2_CSS_PROBE==='1'){await page.addStyleTag({content:'.app-shell{flex-direction:column}'});const after=await page.evaluate(()=>({width:document.documentElement.scrollWidth,viewport:innerWidth}));evidence.push({check:'Single-property flex-direction hypothesis',status:after.width<=after.viewport?'PASS':'FAIL',after})}
 if(overflow&&process.env.R2_CSS_PROBE!=='1')process.exitCode=1;
 await page.getByRole('button',{name:'Sair',exact:true}).click();
 await page.waitForTimeout(1000);
 evidence.push({check:'Logout navigation',status:'OBSERVED',origin:new URL(page.url()).origin});
 console.log(JSON.stringify(evidence,null,2));
} catch(error) {evidence.push({check:'browser flow',status:'FAIL',error:error.message.split('\n')[0]});console.log(JSON.stringify(evidence,null,2));process.exitCode=1}
finally {await writeFile('hub/evidence/r2/execution/browser-smoke.json',JSON.stringify(evidence,null,2)+'\n');await browser.close()}

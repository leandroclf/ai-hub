import {createHmac} from 'node:crypto';
import {writeFile} from 'node:fs/promises';
const playwright=await import(process.env.PLAYWRIGHT_MODULE||'playwright-core');
const chromium=playwright.chromium||playwright.default?.chromium;
if(!chromium)throw new Error('playwright-core sem export chromium');
const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/usr/bin/google-chrome'});
const page=await browser.newPage({viewport:{width:1365,height:900}});
const evidence=[];
let accessToken='';
page.on('request',request=>{const authorization=request.headers().authorization||'';if(authorization.startsWith('Bearer '))accessToken=authorization.slice(7)});
async function authenticate(username,expectedURL){
 if(!await page.locator('#otp').count()){await page.getByRole('button',{name:'Entrar',exact:true}).click();await page.locator('#username').fill(username);await page.locator('#password').fill('R2-fixture-password!');await page.locator('#kc-login').click();await page.locator('#otp').waitFor();await page.waitForTimeout((30-(Math.floor(Date.now()/1000)%30))*1000+250)}
 for(let attempt=0;attempt<3;attempt++)for(const skew of [0,-1,1]){if(!await page.locator('#otp').count())break;const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)+skew));const digest=createHmac('sha1',Buffer.from('AI-HUB-R2-MFA-KEY-01')).update(counter).digest();const offset=digest[digest.length-1]&15;const otp=((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');await page.locator('#otp').fill(otp);await page.locator('#kc-login').click();try{await page.waitForURL(expectedURL,{timeout:1500});return}catch{}}
 throw new Error('OIDC OTP não foi aceito no ensaio de produto');
}
async function providerEffects(){const response=await fetch('http://127.0.0.1:18090/__qualification/effects');if(!response.ok)throw new Error(`provider-sim effects HTTP ${response.status}`);return await response.json()}
const productCode=`browser-product-simulation-${Date.now()}`;
try{
 await page.goto('http://localhost:13000/products');await authenticate('operadora-a','http://localhost:13000/products');await page.getByRole('heading',{name:'Produtos',exact:true}).waitFor();await page.getByRole('link',{name:'Criar rascunho',exact:true}).click();await page.waitForURL('http://localhost:13000/products/new');
 await page.getByLabel('Nome',{exact:true}).fill(`Produto simulado ${productCode}`);await page.getByLabel('Código estável',{exact:true}).fill(productCode);await page.locator('#field-max_parallel').fill('2');
 await page.getByRole('button',{name:'Adicionar passo',exact:true}).click();await page.getByRole('button',{name:'Adicionar passo',exact:true}).click();await page.getByRole('button',{name:'Adicionar passo',exact:true}).click();
 await page.waitForFunction(()=>Array.from(document.querySelectorAll('select')).some(select=>Array.from(select.options).some(option=>option.value==='protocolo-assincrono@1')));
 const names=page.getByLabel('Nome do passo',{exact:true});await names.nth(0).fill('etapa_a');await names.nth(1).fill('etapa_b');await names.nth(2).fill('etapa_c');
 for(let i=0;i<3;i++)await page.locator('div.step').nth(i).locator('select').nth(0).selectOption('protocolo-assincrono@1');
 await page.getByLabel('Depende dos passos (separados por vírgula)',{exact:true}).nth(2).fill('etapa_a,etapa_b');
 await page.getByRole('button',{name:'Salvar rascunho',exact:true}).click();await page.waitForURL(new RegExp(`/products/${productCode}/1$`));
 const before=await providerEffects();await page.getByRole('button',{name:'Validar e simular sem efeitos externos',exact:true}).click();const approved=page.getByRole('status').filter({hasText:'Validação aprovada'});await approved.waitFor();const graph=await page.locator('svg[aria-label="Grafo por grupos de execução"]').count();const tableText=await page.locator('ol').innerText();if(!graph||!tableText.includes('etapa_a')||!tableText.includes('etapa_b')||!tableText.includes('etapa_c'))throw new Error(`grafo/tabela não expôs camadas: ${tableText}`);const after=await providerEffects();if(Number(after.effects)!==Number(before.effects)||Number(after.protocols)!==Number(before.protocols))throw new Error(`simulação alterou efeitos reais: antes=${JSON.stringify(before)} depois=${JSON.stringify(after)}`);
 evidence.push({check:'Admin product simulation shows parallel/dependent graph and tabular order without provider effects',status:'PASS',product:productCode,layers:tableText,fixture:(await approved.innerText()).split('\n')[0],provider_effects_before:before,provider_effects_after:after});
 const depends=page.getByLabel('Depende dos passos (separados por vírgula)',{exact:true});await depends.nth(0).fill('etapa_c');await page.getByRole('button',{name:'Salvar rascunho',exact:true}).click();await page.getByRole('button',{name:'Validar e simular sem efeitos externos',exact:true}).click();await page.getByRole('alert').filter({hasText:'ciclo entre passos'}).waitFor();const alertText=await page.getByRole('alert').innerText();if(!alertText.includes('etapa_a')||!alertText.includes('etapa_c'))throw new Error(`ciclo não identificou passos: ${alertText}`);
 evidence.push({check:'Admin product validation identifies cycle members and blocks publication',status:'PASS',product:productCode,error:alertText.match(/ciclo entre passos:[^\n]*/)?.[0]||'ciclo entre passos'});
 await depends.nth(0).fill('');await page.getByRole('button',{name:'Salvar rascunho',exact:true}).click();
 const fanoutCode=`browser-product-fanout-${Date.now()}`;
 const fanoutSteps=Array.from({length:21},(_,index)=>({id:`etapa_${index+1}`,service_id:'protocolo-assincrono',service_version:1,depends_on:[],required:true,input_mapping:{}}));
 const fanoutData={modes:['ASYNC'],provider_mode:'async_poll',client_sla_seconds:30,provider_sla_seconds:5,provider_sla_policy:'MONITOR_ONLY',retry_ttl_seconds:10,finalization_reserve_seconds:5,auto_wait_seconds:5,sync_http_budget_seconds:15,input_schema:{type:'object',properties:{marker:{type:'string'}},required:[]},output_schema:{type:'object',properties:{marker:{type:'string'}},required:[]},data_class:'SYNTHETIC',target_kind:'products',target_id:fanoutCode,target_version:1,technical_profile_id:'profile-r4-json',technical_profile_version:1,purchase_contract_id:'purchase-r4',purchase_contract_version:1,sale_contract_id:'sale-r4',sale_contract_version:1,max_parallel:5,allow_partial:false,consolidation:'ALL_REQUIRED',failure_policy:'STOP',steps:fanoutSteps};
 const fanoutCreate=await page.evaluate(async ({code,data,token})=>{const response=await fetch('/api/atlas/admin/v1/products?tenant_id=acme',{method:'POST',headers:{'content-type':'application/json',authorization:`Bearer ${token}`},body:JSON.stringify({kind:'products',id:code,version:1,tenant_id:'acme',name:'Fan-out abusivo '+code,data})});return {status:response.status,body:await response.json().catch(()=>({}))}},{code:fanoutCode,data:fanoutData,token:accessToken});
 if(fanoutCreate.status!==201)throw new Error(`rascunho de fan-out não foi criado para validação: HTTP ${fanoutCreate.status}`);
 const fanoutValidation=await page.evaluate(async ({code,token})=>{const response=await fetch(`/api/atlas/admin/v1/products/${encodeURIComponent(code)}/1/simulate?tenant_id=acme`,{method:'POST',headers:{'content-type':'application/json',authorization:`Bearer ${token}`},body:'{}'});return {status:response.status,body:await response.json().catch(()=>({}))}},{code:fanoutCode,token:accessToken});
 const fanoutError=String(fanoutValidation.body?.field_errors?.steps||'');
 if(fanoutValidation.status!==200||!fanoutError.includes('máximo de 20 passos'))throw new Error(`fan-out abusivo não foi bloqueado antes da publicação: ${JSON.stringify(fanoutValidation)}`);
 const afterFanout=await providerEffects();
 if(Number(afterFanout.effects)!==Number(before.effects)||Number(afterFanout.protocols)!==Number(before.protocols))throw new Error('validação de fan-out gerou efeito externo');
 evidence.push({check:'Admin rejects product fan-out above the authorized limit before publication or external effects',status:'PASS',product:fanoutCode,steps:fanoutSteps.length,error:fanoutError,provider_effects_before:before,provider_effects_after:afterFanout});
 console.log(JSON.stringify(evidence,null,2));
}catch(error){evidence.push({check:'admin product simulation',status:'FAIL',url:page.url(),error:error.message.split('\n')[0],bodyText:(await page.locator('body').innerText().catch(()=>'' )).slice(0,800)});console.log(JSON.stringify(evidence,null,2));process.exitCode=1}
finally{await writeFile('hub/evidence/r2/execution/admin-product-simulation-smoke.json',JSON.stringify(evidence,null,2)+'\n');await browser.close()}

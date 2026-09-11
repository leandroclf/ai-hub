import {createHmac} from 'node:crypto';
import {writeFile} from 'node:fs/promises';
const playwright=await import(process.env.PLAYWRIGHT_MODULE||'playwright-core');
const chromium=playwright.chromium||playwright.default?.chromium;
if(!chromium) throw new Error('playwright-core sem export chromium');
const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/usr/bin/google-chrome'});
const page=await browser.newPage({viewport:{width:1365,height:900}});
const evidence=[];
async function authenticate(expectedURL){
 await page.getByRole('button',{name:'Entrar',exact:true}).click();
 await page.locator('#username').fill('operadora-a');
 await page.locator('#password').fill('R2-fixture-password!');
 await page.locator('#kc-login').click();
 await page.locator('#otp').waitFor();
 for(let attempt=0;attempt<8;attempt++){
  if(await page.locator('#otp').count()){
   const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)));
   const digest=createHmac('sha1',Buffer.from('JBSWY3DPEHPK3PXP')).update(counter).digest();
   const offset=digest[digest.length-1]&15;
   const otp=((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');
   await page.locator('#otp').fill(otp);await page.locator('#kc-login').click();
  }
  try{await page.waitForURL(expectedURL,{timeout:3500});return}catch{await page.waitForTimeout(250)}
 }
 throw new Error('OIDC OTP não foi aceito após oito tentativas');
}
try {
 await page.goto('http://localhost:13000/services');
 await authenticate('http://localhost:13000/services');
 await page.waitForURL('http://localhost:13000/services');
 await page.getByRole('heading',{name:'Serviços',exact:true}).waitFor();
 evidence.push({check:'OIDC Authorization Code PKCE + password + OTP in Chromium',status:'PASS'});
 const clientCode=`browser-client-${Date.now()}`;
 await page.getByRole('link',{name:'Clientes',exact:true}).click();
 await page.getByRole('heading',{name:'Clientes',exact:true}).waitFor();
 await page.getByRole('link',{name:'Criar rascunho',exact:true}).click();
 await page.getByLabel('Nome',{exact:true}).fill(`Browser ${clientCode}`);
 await page.getByLabel('Código estável',{exact:true}).fill(clientCode);
 await page.locator('#field-environment').selectOption('local');
 await page.locator('#field-cell_id').fill('r2-cell-a');
 await page.locator('#field-capacity_units').fill('1');
 await page.locator('#field-isolation_class').selectOption('SHARED');
 await page.getByRole('button',{name:'Salvar rascunho',exact:true}).click();
 await page.waitForURL(/http:\/\/localhost:13000\/clients\/[^/]+\/1/);
 await page.reload({waitUntil:'networkidle'});
 await page.waitForTimeout(60500-(Date.now()%30000));
 await authenticate(/http:\/\/localhost:13000\/clients\/[^/]+\/1/);
 await page.waitForURL(/http:\/\/localhost:13000\/clients\/[^/]+\/1/);
 await page.locator(`input[value="Browser ${clientCode}"]`).waitFor();
 evidence.push({check:'Admin resource save, reload and durable readback',status:'PASS',resource:clientCode});
 const productCode=`browser-product-${Date.now()}`;
 await page.getByRole('link',{name:'Produtos',exact:true}).click();
 await page.waitForURL('http://localhost:13000/products');
 await page.getByRole('link',{name:'Criar rascunho',exact:true}).click();
 await page.getByRole('heading',{name:'Produtos',exact:true}).waitFor();
 await page.getByLabel('Nome',{exact:true}).fill(`Produto ${productCode}`);
 await page.getByLabel('Código estável',{exact:true}).fill(productCode);
 await page.locator('#field-max_parallel').fill('2');
 await page.getByRole('button',{name:'Adicionar passo',exact:true}).click();
 await page.getByRole('button',{name:'Adicionar passo',exact:true}).click();
 const stepNames=page.getByLabel('Nome do passo',{exact:true});
 await page.waitForFunction(()=>Array.from(document.querySelectorAll('select')).some(select=>Array.from(select.options).some(option=>option.value==='protocolo-assincrono@1')));
 await stepNames.nth(0).fill('etapa_a');
 await stepNames.nth(1).fill('etapa_b');
 await page.locator('div.step').nth(0).locator('select').nth(0).selectOption('protocolo-assincrono@1');
 await page.locator('div.step').nth(1).locator('select').nth(0).selectOption('protocolo-assincrono@1');
 await page.getByLabel('Depende dos passos (separados por vírgula)',{exact:true}).nth(1).fill('etapa_a');
 await page.locator('#step-1-input-mapping').fill('{"marker":"etapa_a.marker"}');
 await page.getByRole('button',{name:'Salvar rascunho',exact:true}).click();
 await page.waitForURL(/http:\/\/localhost:13000\/products\/[^/]+\/1/);
 await page.reload({waitUntil:'networkidle'});
 await page.waitForTimeout(60500-(Date.now()%30000));
 await authenticate(/http:\/\/localhost:13000\/products\/[^/]+\/1/);
 if(await page.locator('#step-1-input-mapping').inputValue()!=='{\n  "marker": "etapa_a.marker"\n}')throw new Error('input_mapping do produto não reapareceu no readback');
 evidence.push({check:'Admin product editor persists dependent input mapping',status:'PASS',resource:productCode});
 await page.getByRole('link',{name:'Provedores',exact:true}).click();
 await page.waitForURL('http://localhost:13000/providers');
 await page.getByRole('heading',{name:'Provedores',exact:true}).waitFor();
 evidence.push({check:'URL navigation and real authenticated provider list',status:'PASS'});
 await page.getByRole('link',{name:'SLA bilateral',exact:true}).click();
 await page.waitForURL('http://localhost:13000/sla-reports');
 await page.getByRole('heading',{name:'SLA bilateral',exact:true}).waitFor();
 await page.locator('table').waitFor();
 evidence.push({check:'Admin bilateral SLA report reads persisted protocol snapshots',status:'PASS'});
 await page.getByRole('link',{name:'Destinos webhook',exact:true}).click();
 await page.waitForURL('http://localhost:13000/destinations');
 await page.getByRole('heading',{name:'Destinos de webhook versionados',exact:true}).waitFor();
 await page.getByRole('heading',{name:'Destinos persistidos',exact:true}).waitFor();
 await page.waitForFunction(()=>!document.querySelector('button[type="submit"]')?.hasAttribute('disabled'));
 await page.getByLabel('Aplicação (vazio = tenant)').fill('app-a');
 await page.getByLabel('URL autorizada').fill('http://webhook-sink:8091/events');
 await page.getByRole('button',{name:'Publicar versão',exact:true}).click();
 const publishStatus=page.locator('.status-ok');
 await publishStatus.waitFor();
 if(!/^Destino .* v1 publicado\.$/.test(await publishStatus.innerText()))throw new Error('publicação de destino sem recibo de sucesso');
 await page.getByRole('link',{name:'Serviços',exact:true}).click();
 await page.waitForURL('http://localhost:13000/services');
 await page.getByRole('link',{name:'Destinos webhook',exact:true}).click();
 await page.waitForURL('http://localhost:13000/destinations');
 await page.getByRole('heading',{name:'Destinos persistidos',exact:true}).waitFor();
 const destinationRow=page.locator('tbody tr').filter({hasText:'app-a'}).first();
 await destinationRow.waitFor();
 if(!await destinationRow.innerText().then(text=>text.includes('webhook-sink:8091/events')))throw new Error('destino publicado não reapareceu na consulta persistida');
 evidence.push({check:'Admin destination publish, reload and durable readback',status:'PASS'});
 evidence.push({check:'Admin versioned webhook destinations reads persisted application snapshot policy',status:'PASS'});
 await page.getByRole('link',{name:'Financeiro',exact:true}).click();
 await page.waitForURL('http://localhost:13000/finance');
 await page.getByRole('heading',{name:'Financeiro e conciliação',exact:true}).waitFor();
 await page.locator('dl.definition-list').waitFor();
 await page.getByRole('link',{name:'Compra e venda',exact:true}).click();
 await page.waitForURL('http://localhost:13000/finance?view=facts');
 await page.locator('dl.definition-list').waitFor();
 await page.getByRole('link',{name:'Journal',exact:true}).click();
 await page.waitForURL('http://localhost:13000/finance?view=journal');
 await page.locator('dl.definition-list').waitFor();
 evidence.push({check:'Admin finance accounts, facts and journal read persisted Libra data',status:'PASS'});
 await page.getByRole('link',{name:'Fechamentos',exact:true}).click();
 await page.waitForURL('http://localhost:13000/finance?view=periods');
 await page.getByLabel('Início',{exact:true}).fill('2020-01-01');
 await page.getByLabel('Fim (exclusivo)',{exact:true}).fill('2020-01-02');
 await page.getByRole('button',{name:'Verificar e fechar período',exact:true}).click();
 await page.getByRole('alert').filter({hasText:'financial_obligations_open'}).waitFor();
 evidence.push({check:'Admin finance close refuses unresolved obligations with explicit error',status:'PASS'});
 const storage=await page.evaluate(()=>({local:Object.keys(localStorage),session:Object.keys(sessionStorage)}));
 if(storage.local.length||storage.session.includes('atlas.pkce'))throw new Error('Unexpected persisted session material');
 evidence.push({check:'No persistent token and PKCE transaction removed',status:'PASS',storageKeys:storage});
 await page.screenshot({path:'hub/evidence/r2/execution/admin-providers.png',fullPage:true});
 await page.setViewportSize({width:390,height:844});
 const overflowInfo=await page.evaluate(()=>({
  viewport:innerWidth,
  scrollWidth:document.documentElement.scrollWidth,
  scrollContainers:Array.from(document.querySelectorAll('*')).map(element=>{const rect=element.getBoundingClientRect();const style=getComputedStyle(element);return {element:element.tagName.toLowerCase(),className:typeof element.className==='string'?element.className:'',clientWidth:element.clientWidth,scrollWidth:element.scrollWidth,width:Math.round(rect.width),right:Math.round(rect.right),overflowX:style.overflowX}}).filter(item=>item.scrollWidth>item.clientWidth+1).slice(0,12)
 }));
 const overflow=overflowInfo.scrollWidth>overflowInfo.viewport;
 const dimensions=await page.evaluate(()=>Array.from(document.querySelectorAll('.app-shell,.app-header,.app-nav,.app-main')).map(e=>({element:e.className,width:e.getBoundingClientRect().width,right:e.getBoundingClientRect().right,display:getComputedStyle(e).display,direction:getComputedStyle(e).flexDirection})));
 evidence.push({check:'390px viewport has no page overflow',status:overflow?'FAIL':'PASS',dimensions,overflowInfo});
 if(overflow)process.exitCode=1;
 await page.getByRole('button',{name:'Sair',exact:true}).click();
 await page.waitForTimeout(1000);
 const unauthenticatedStatus=await page.request.get('http://localhost:13000/api/atlas/admin/v1/clients?tenant_id=acme').then(response=>response.status());
 if(unauthenticatedStatus!==401)throw new Error(`API administrativa sem sessão retornou HTTP ${unauthenticatedStatus}`);
 evidence.push({check:'Logout revokes browser access to administrative API',status:'PASS',httpStatus:unauthenticatedStatus});
 evidence.push({check:'Logout navigation',status:'OBSERVED',origin:new URL(page.url()).origin});
 console.log(JSON.stringify(evidence,null,2));
} catch(error) {evidence.push({check:'browser flow',status:'FAIL',error:error.message.split('\n')[0]});console.log(JSON.stringify(evidence,null,2));process.exitCode=1}
finally {await writeFile('hub/evidence/r2/execution/browser-smoke.json',JSON.stringify(evidence,null,2)+'\n');await browser.close()}

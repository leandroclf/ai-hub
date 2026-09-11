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
 if(!await page.locator('#otp').count()){await page.getByRole('button',{name:'Entrar',exact:true}).click();await page.locator('#username').fill(username);await page.locator('#password').fill('R2-fixture-password!');await page.locator('#kc-login').click();await page.locator('#otp').waitFor()}
 for(let attempt=0;attempt<3;attempt++)for(const skew of [0,-1,1]){if(!await page.locator('#otp').count())break;const counter=Buffer.alloc(8);counter.writeBigUInt64BE(BigInt(Math.floor(Date.now()/30000)+skew));const digest=createHmac('sha1',Buffer.from('JBSWY3DPEHPK3PXP')).update(counter).digest();const offset=digest[digest.length-1]&15;const otp=((digest.readUInt32BE(offset)&0x7fffffff)%1000000).toString().padStart(6,'0');await page.locator('#otp').fill(otp);await page.locator('#kc-login').click();try{await page.waitForURL(expectedURL,{timeout:1500});return}catch{}}
 throw new Error('OIDC OTP não foi aceito no ensaio de publicação');
}
const serviceCode=`browser-publication-${Date.now()}`;
try{
 await page.goto('http://localhost:13000/services');await authenticate('operadora-a','http://localhost:13000/services');await page.getByRole('heading',{name:'Serviços',exact:true}).waitFor();
 await page.getByRole('link',{name:'Criar rascunho',exact:true}).click();await page.waitForURL('http://localhost:13000/services/new');
 await page.getByLabel('Nome',{exact:true}).fill(`Serviço publicado ${serviceCode}`);await page.getByLabel('Código estável',{exact:true}).fill(serviceCode);
 await page.locator('#field-adapter_id').fill('provider-sim');await page.locator('#field-qualification_id').fill('r4-fixture-qualification');await page.locator('#field-data_class').selectOption('SYNTHETIC');
 await page.getByRole('button',{name:'Salvar rascunho',exact:true}).click();await page.waitForURL(new RegExp(`/services/${serviceCode}/1$`));
 await page.getByRole('button',{name:'Validar e simular sem efeitos externos',exact:true}).click();const validation=page.getByRole('status').filter({hasText:'Validação aprovada'});await validation.waitFor();const validationText=await validation.innerText();if(!validationText.includes('r4-fixture-qualification')&&!validationText.includes('synthetic-validation-no-provider-calls'))throw new Error(`qualificação/fixture ausente: ${validationText}`);
 evidence.push({check:'Admin service validation uses current qualification and no-provider-call fixture',status:'PASS',service:serviceCode,validation:validationText});
 await page.getByLabel('Motivo da ação',{exact:true}).fill('homologação integrada do serviço');await page.getByRole('button',{name:/Publicar versão 1/,exact:true}).click();const published=page.getByRole('status').filter({hasText:'Versão 1 publicada'});await published.waitFor();
 if(!(await page.locator('body').innerText()).includes('PUBLISHED'))throw new Error('serviço publicado sem estado PUBLISHED visível');
 evidence.push({check:'Admin publishes qualified service version and presents immutable state',status:'PASS',service:serviceCode,version:1});
 if(!accessToken)throw new Error('token administrativo não foi observado na chamada autenticada');
 const currentResponse=await fetch(`http://localhost:13000/api/atlas/admin/v1/services/${encodeURIComponent(serviceCode)}/1?tenant_id=acme`,{headers:{Authorization:`Bearer ${accessToken}`}});const current=await currentResponse.json();
 const mutationResponse=await fetch(`http://localhost:13000/api/atlas/admin/v1/services/${encodeURIComponent(serviceCode)}/1?tenant_id=acme`,{method:'PATCH',headers:{Authorization:`Bearer ${accessToken}`,'Content-Type':'application/json','If-Match':`"${current.revision}"`},body:JSON.stringify({name:`alteração indevida ${serviceCode}`,data:current.data})});
 if(mutationResponse.status!==409)throw new Error(`mutação de publicação retornou HTTP ${mutationResponse.status}, esperado 409`);
 const afterResponse=await fetch(`http://localhost:13000/api/atlas/admin/v1/services/${encodeURIComponent(serviceCode)}/1?tenant_id=acme`,{headers:{Authorization:`Bearer ${accessToken}`}});const after=await afterResponse.json();
 if(after.name!==current.name||after.content_hash!==current.content_hash||after.state!=='PUBLISHED')throw new Error('conteúdo publicado mudou após tentativa de PATCH');
 evidence.push({check:'Published service rejects mutation and preserves content hash',status:'PASS',service:serviceCode,version:1,mutation_status:mutationResponse.status,state:after.state});
 await page.getByRole('button',{name:'Sair',exact:true}).click();
 await page.goto(`http://localhost:13000/services/${serviceCode}/1`);
 await authenticate('leitor-a',new RegExp(`/services/${serviceCode}/1$`));
 await page.getByText(/Cliente: acme · PUBLISHED/).waitFor();
 const readerText=await page.locator('body').innerText();
 const readerCode=await page.getByLabel('Código estável',{exact:true}).inputValue();
 const readerResponse=await fetch(`http://localhost:13000/api/atlas/admin/v1/services/${encodeURIComponent(serviceCode)}/1?tenant_id=acme`,{headers:{Authorization:`Bearer ${accessToken}`}});
 const reader=await readerResponse.json();
 if(readerResponse.status!==200||readerCode!==serviceCode||!readerText.includes('PUBLISHED')||reader.content_hash!==current.content_hash||reader.revision!==current.revision)throw new Error(`segundo operador não recuperou a mesma revisão publicada após refresh: status=${readerResponse.status} código=${readerCode} revisão=${reader.revision} hash=${reader.content_hash}`);
 evidence.push({check:'Second same-tenant operator refreshes the published URL and reads the persisted version',status:'PASS',service:serviceCode,version:1,reader:'leitor-a',revision:reader.revision,content_hash:reader.content_hash});
 console.log(JSON.stringify(evidence,null,2));
}catch(error){evidence.push({check:'admin service publication',status:'FAIL',url:page.url(),error:error.message.split('\n')[0],bodyText:(await page.locator('body').innerText().catch(()=>'' )).slice(0,800)});console.log(JSON.stringify(evidence,null,2));process.exitCode=1}
finally{await writeFile('hub/evidence/r2/execution/admin-service-publication-smoke.json',JSON.stringify(evidence,null,2)+'\n');await browser.close()}

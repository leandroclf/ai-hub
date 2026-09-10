export interface Principal { subject:string; tenant_id:string; application_id:string; environment:string; scopes:string[]; roles:string[]; mfa:boolean; expires_at:string }
export interface Resource {kind:string;id:string;version:number;tenant_id:string;name:string;state:string;revision:number;data:Record<string,unknown>;content_hash:string;author:string;updated_at:string}
export interface Page<T>{items:T[];next_cursor:string}
export interface Validation {valid:boolean;field_errors:Record<string,string>;layers:string[][];effective_retry_seconds:number;content_hash:string;fixture:string}
export class APIError extends Error { constructor(public status:number,public body:Record<string,unknown>){super(String(body.message??body.error??`HTTP ${status}`))} }
let token='';
export function setToken(value:string){token=value}
export async function api<T>(path:string,init:RequestInit={}):Promise<T>{
 const domain=path.startsWith("/admin/v1/finance")?"libra":path.startsWith("/admin/v1/deliveries")?"pulsar":path.startsWith("/admin/v1/protocols")||path.startsWith("/admin/v1/sla-reports")?"orbita":path.startsWith("/admin/v1/capacity-domains")?"cometa":"atlas";
 const response=await fetch(`/api/${domain}${path}`,{...init,headers:{'Content-Type':'application/json',...(token?{Authorization:`Bearer ${token}`} : {}),...init.headers},cache:'no-store'});
 const data=await response.json().catch(()=>({message:'Resposta indisponível. Tente novamente.'}));
 if(!response.ok){
  if(response.status===401){
   token='';
   window.dispatchEvent(new Event('session-expired'));
  }
  throw new APIError(response.status,data);
 }
 return data as T;
}
export const resourceURL=(r:Pick<Resource,'kind'|'id'|'version'>)=>`/admin/v1/${r.kind}/${encodeURIComponent(r.id)}/${r.version}`;
export async function command<T>(path:string,body:unknown,revision?:number,idempotencyKey=crypto.randomUUID()):Promise<T>{
 const request=()=>api<T>(path,{method:'POST',headers:{'Idempotency-Key':idempotencyKey,...(revision?{'If-Match':`"${revision}"`}: {})},body:JSON.stringify(body)});
 try{return await request()}catch(error){
  // Timeout/5xx pode ocorrer depois do efeito externo. A segunda tentativa
  // reutiliza a mesma intenção para que a autoridade durável deduplique-a.
  if(error instanceof APIError&&error.status<500)throw error;
  await new Promise(resolve=>setTimeout(resolve,150));
  return request();
 }
}

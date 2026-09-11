import {useEffect,useState} from 'react';
import {api,command,Page,Principal} from '../api/admin';
import {navigate,useRoute} from '../navigation';

type RecordData=Record<string,unknown>;

export default function OperationsPage({kind,tenant,principal}:{kind:string;tenant:string;principal:Principal}){
 const route=useRoute();
 const query=new URLSearchParams(route.split('?')[1]);
 const parts=route.split('?')[0].split('/').filter(Boolean);
 const id=parts[1]||'';
 const [items,setItems]=useState<RecordData[]>([]);
 const [freshness,setFreshness]=useState<RecordData|null>(null);
 const [next,setNext]=useState('');
 const [detail,setDetail]=useState<RecordData|null>(null);
 const [timeline,setTimeline]=useState<RecordData[]>([]);
 const [error,setError]=useState('');
 const [busy,setBusy]=useState(false);
 const [reason,setReason]=useState('');
 const [message,setMessage]=useState('');
 const [refresh,setRefresh]=useState(0);
 const crossTenant=principal.scopes.includes('admin:cross_tenant')&&principal.mfa&&tenant!==principal.tenant_id;
 const requiresReadReason=crossTenant&&(kind==='protocols'||kind==='sla-reports');
 const scopeParams=new URLSearchParams({tenant_id:tenant});
 if(query.get('reason'))scopeParams.set('reason',query.get('reason')!);
 const scope=scopeParams.toString();
 const title=kind==='protocols'?'Protocolos':kind==='deliveries'?'Entregas':'SLA bilateral';

 useEffect(()=>{
  const abort=new AbortController();
  setBusy(true);setError('');setDetail(null);setTimeline([]);
  if(requiresReadReason&&(query.get('reason')||'').trim().length<8){setBusy(false);setError('Informe a finalidade da consulta cross-tenant antes de consultar.');return()=>abort.abort()}
  if(id){
   void api<RecordData>(`/admin/v1/${kind}/${encodeURIComponent(id)}?${scope}`,{signal:abort.signal}).then(async value=>{
    setDetail(value);
    if(kind==='protocols'){
     const result=await api<Page<RecordData>>(`/admin/v1/protocols/${encodeURIComponent(id)}/timeline?${scope}`,{signal:abort.signal});
     setTimeline(result.items||[]);
    }
   }).catch(e=>{if(!abort.signal.aborted)setError(String(e))}).finally(()=>{if(!abort.signal.aborted)setBusy(false)});
  }else{
   const params=new URLSearchParams(query);params.set('tenant_id',tenant);params.set('limit','25');
   void api<Page<RecordData>>(`/admin/v1/${kind}?${params}`,{signal:abort.signal}).then(value=>{setItems(value.items||[]);setNext(value.next_cursor||'');setFreshness(kind==='sla-reports'?value as unknown as RecordData:null)}).catch(e=>{if(!abort.signal.aborted)setError(`Consulta indisponível; dados anteriores podem estar desatualizados. ${e}`)}).finally(()=>{if(!abort.signal.aborted)setBusy(false)});
  }
  return()=>abort.abort();
 },[kind,tenant,route,refresh]);

 useEffect(()=>{setMessage('')},[kind,tenant,route]);

 async function act(action:string){
  if(!id||busy)return;
  setBusy(true);setError('');setMessage('');
  try{const receipt=await command<RecordData>(`/admin/v1/${kind}/${encodeURIComponent(id)}/${action}?${scope}`,{reason});setMessage(`Ação registrada: ${JSON.stringify(receipt)}`);setRefresh(value=>value+1)}
  catch(e){setError(String(e))}finally{setBusy(false)}
 }

 async function exportSLA(){
  if(kind!=='sla-reports'||id||busy)return;
  setBusy(true);setError('');setMessage('');
  try{
   const params=new URLSearchParams(query);params.set('tenant_id',tenant);params.set('limit','100');
   const report=await api<RecordData>(`/admin/v1/sla-reports/export?${params}`);
   const content=String(report.csv||'');
   const blob=new Blob([content],{type:'text/csv;charset=utf-8'});
   const url=URL.createObjectURL(blob);const link=document.createElement('a');
   link.href=url;link.download=String(report.filename||'sla-report.csv');link.click();URL.revokeObjectURL(url);
   setMessage(`Exportação limitada a ${String(report.rows||0)} registros e auditada pela autoridade.`);
  }catch(e){setError(String(e))}finally{setBusy(false)}
 }

 return <section>
  <h2>{title}</h2>
  {error&&<p role="alert">{error}</p>}
  {message&&<p role="status">{message}</p>}
  {busy&&<p role="status">Consultando autoridade do recurso…</p>}
  <form className="filters" onSubmit={event=>{event.preventDefault();const form=new FormData(event.currentTarget);const resource=String(form.get('resource_id')||'');const values=new URLSearchParams({status:String(form.get('status')||''),from:String(form.get('from')||''),to:String(form.get('to')||'')});const readReason=String(form.get('read_reason')||'').trim();if(readReason)values.set('reason',readReason);if(resource)navigate(`/${kind}/${encodeURIComponent(resource)}?${values}`);else navigate(`/${kind}?${values}`)}}>
   <label>Identificador<input name="resource_id" defaultValue={id}/></label>
   <label>Estado<select name="status" defaultValue={query.get('status')||''}><option value="">Todos</option>{(kind==='deliveries'?['PENDING','DELIVERED','EXHAUSTED','SUSPENDED']:['ACCEPTED','RUNNING','WAITING_PROVIDER','RECONCILING','SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED','EXPIRED','UNKNOWN']).map(value=><option key={value}>{value}</option>)}</select></label>
   <label>De<input name="from" type="date" defaultValue={query.get('from')||''}/></label>
   <label>Até<input name="to" type="date" defaultValue={query.get('to')||''}/></label>
   {requiresReadReason&&<label>Finalidade da consulta<input name="read_reason" minLength={8} defaultValue={query.get('reason')||''}/></label>}
   <button disabled={busy}>Consultar</button><button type="button" onClick={()=>setRefresh(value=>value+1)}>Atualizar</button>{kind==='sla-reports'&&!id&&<button type="button" disabled={busy} onClick={()=>void exportSLA()}>Exportar CSV limitado</button>}
  </form>
  {kind==='sla-reports'&&freshness&&<p role="status">Atualização da projeção: {String(freshness.generated_at||'')} · watermark: {String(freshness.watermark_at||'')} · atraso observado: {Number(freshness.watermark_lag_seconds||0).toFixed(3)}s. Dados atrasados continuam identificados; não significam ausência de violações.</p>}
  {kind==='sla-reports'&&freshness&&<div className="sla-cohort" aria-label="Resumo da coorte de SLA"><strong>Coorte:</strong> elegíveis {Number((freshness.cohort as RecordData|undefined)?.eligible||0)} · abertos {Number((freshness.cohort as RecordData|undefined)?.open||0)} · cumpridos {Number((freshness.cohort as RecordData|undefined)?.fulfilled||0)} · vencidos {Number((freshness.cohort as RecordData|undefined)?.expired||0)} · exclusões {Number((freshness.cohort as RecordData|undefined)?.excluded||0)}. Protocolos abertos não são contabilizados como sucesso.</div>}
  {!id&&<div className="table-scroll" tabIndex={0}>
   <table><thead><tr>{kind==='sla-reports'?<><th>Protocolo</th><th>Cliente</th><th>Estado</th><th>SLA cliente</th><th>SLA provedor</th><th>Observado em</th></>:<><th>Identificador</th><th>Cliente</th><th>Estado</th><th>Data</th></>}</tr></thead>
    <tbody>{items.map((row,index)=>{const key=String(row.protocol_id||row.delivery_id||row.id||index);return <tr key={key}>
     <td><a href={`/${kind}/${key}${query.get('reason')?`?reason=${encodeURIComponent(query.get('reason')!)}`:''}`} onClick={event=>{event.preventDefault();navigate(`/${kind}/${key}${query.get('reason')?`?reason=${encodeURIComponent(query.get('reason')!)}`:''}`)}}>{key}</a></td>
     <td>{String(row.tenant_id||'')}</td><td>{String(row.status||row.state||'')}</td>
     {kind==='sla-reports'?<><td>{row.client_sla_breached?'violado':'dentro do prazo'}</td><td>{row.provider_sla_breached?'violado':'dentro do prazo'}</td><td>{String(row.observed_at||'')}</td></>:<td>{String(row.accepted_at||row.created_at||'')}</td>}
    </tr>})}</tbody>
   </table>
  </div>}
  {!id&&!items.length&&!busy&&!error&&<p>Nenhum registro no filtro autorizado.</p>}
  {!id&&next&&<button onClick={()=>navigate(`/${kind}?${new URLSearchParams({...Object.fromEntries(query),cursor:next})}`)}>Próxima página</button>}
  {detail&&<><h3>{kind==='sla-reports'?'Apuração bilateral persistida':'Detalhe persistido'}</h3><DefinitionList value={detail}/>
   {String(detail.status||detail.state)==='UNKNOWN'&&<p>A operação tem efeito externo incerto. Novo envio permanece bloqueado; use reconciliação autorizada.</p>}
   {timeline.length>0&&<><h3>Timeline</h3><ol className="timeline">{timeline.map((item,index)=><li key={index}><DefinitionList value={item}/></li>)}</ol></>}
   {id&&((kind==='deliveries'&&principal.scopes.includes('deliveries:write'))||(kind==='protocols'&&principal.scopes.includes('protocols:reconcile')))&&<div><label>Justificativa<input minLength={8} value={reason} onChange={event=>setReason(event.target.value)}/></label>{kind==='deliveries'&&String(detail.status||detail.state)==='EXHAUSTED'&&<button disabled={busy||reason.trim().length<8} onClick={()=>void act('redeliver')}>Reentregar mesmos bytes do webhook</button>}{kind==='protocols'&&<button disabled={busy||reason.trim().length<8} onClick={()=>void act('reconcile')}>Solicitar reconciliação</button>}</div>}
  </>}
 </section>;
}

export function DefinitionList({value}:{value:RecordData}){return <dl className="definition-list">{Object.entries(value).map(([key,value])=><div key={key}><dt>{key.replace(/_/g,' ')}</dt><dd>{typeof value==='object'?<pre>{JSON.stringify(value,null,2)}</pre>:String(value??'Não informado')}</dd></div>)}</dl>}

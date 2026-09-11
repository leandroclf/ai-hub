import {useEffect,useState} from 'react';
import {api,command,Principal} from '../api/admin';
import {DefinitionList} from './OperationsPage';
import {navigate,useRoute} from '../navigation';

type RecordData=Record<string,unknown>;

const views=[['accounts','Contas e saldos'],['facts','Compra e venda'],['journal','Journal'],['periods','Fechamentos'],['adjustments','Ajustes'],['disputes','Divergências'],['exports','Exportações']];

export default function FinancePage({tenant,principal}:{tenant:string;principal:Principal}){
 const route=useRoute();
 const params=new URLSearchParams(route.split('?')[1]);
 const view=params.get('view')||'accounts';
 const [result,setResult]=useState<RecordData|null>(null);
 const [error,setError]=useState('');
 const [message,setMessage]=useState('');
 const [busy,setBusy]=useState(false);
 const [refresh,setRefresh]=useState(0);
 const [reason,setReason]=useState('');
 const [origin,setOrigin]=useState('');
 const [from,setFrom]=useState('');
 const [to,setTo]=useState('');
 const [exportID,setExportID]=useState('');
 const [factID,setFactID]=useState('');
 const [amount,setAmount]=useState('');
 const [evidence,setEvidence]=useState('');

 useEffect(()=>{
 const abort=new AbortController();
  setBusy(true);setError('');setResult(null);
  if(view==='exports'&&!exportID){setBusy(false);return()=>abort.abort();}
  const query=new URLSearchParams({tenant_id:tenant,limit:'25',cursor:params.get('cursor')||''});
  const path=view==='exports'&&exportID?`exports/${encodeURIComponent(exportID)}`:view;
  void api<RecordData>(`/admin/v1/finance/${path}?${query}`,{signal:abort.signal}).then(setResult).catch(e=>{if(!abort.signal.aborted)setError(String(e))}).finally(()=>{if(!abort.signal.aborted)setBusy(false)});
  return()=>abort.abort();
 },[route,tenant,refresh,view,exportID]);

 async function action(path:string,body:unknown){
  setBusy(true);setError('');setMessage('');
  try{const receipt=await command<RecordData>(`/admin/v1/finance/${path}?tenant_id=${encodeURIComponent(tenant)}`,body);setMessage(`Ação persistida: ${JSON.stringify(receipt)}`);setRefresh(value=>value+1)}
  catch(e){setError(String(e))}finally{setBusy(false)}
 }

 const canWrite=principal.scopes.includes('finance:write');
 const canApprove=principal.scopes.includes('finance:approve');
 return <section>
  <h2>Financeiro e conciliação</h2>
  <p>Valores e saldos são apurados pelo servidor e exibidos com sua precisão contratual.</p>
  <nav aria-label="Consultas financeiras">{views.map(([id,label])=><a key={id} href={`/finance?view=${id}`} aria-current={view===id?'page':undefined} onClick={event=>{event.preventDefault();navigate(`/finance?view=${id}`)}}>{label}</a>)}</nav>
  {busy&&<p role="status">Consultando financeiro…</p>}{error&&<p role="alert">{error}</p>}{message&&<p role="status">{message}</p>}
  <button onClick={()=>setRefresh(value=>value+1)}>Atualizar</button>
  {view==='exports'&&<label>Identificador da exportação<input value={exportID} onChange={event=>setExportID(event.target.value)}/></label>}
  {result&&<><DefinitionList value={result}/>{view==='adjustments'&&Array.isArray(result.items)&&<ul aria-label="Ajustes persistidos">{(result.items as RecordData[]).map(item=>{const preparedBy=String(item.prepared_by||'');const ownPreparation=preparedBy===principal.subject;return <li key={String(item.id)}>{String(item.id)} · {String(item.state)} · {String(item.reason)} {String(item.state)==='PREPARED'&&ownPreparation?<span role="note">Aguardando aprovador distinto</span>:canApprove&&String(item.state)==='PREPARED'&&<button onClick={()=>void action(`adjustments/${encodeURIComponent(String(item.id))}/approve`,{})}>Aprovar</button>}</li>})}</ul>}{!!result.next_cursor&&<button onClick={()=>navigate(`/finance?${new URLSearchParams({view,cursor:String(result.next_cursor)})}`)}>Próxima página</button>}</>}
  {canWrite&&<fieldset disabled={busy}>
   <legend>{view==='periods'?'Solicitar fechamento':view==='adjustments'?'Preparar ajuste compensatório':view==='disputes'?'Registrar divergência':'Ações financeiras'}</legend>
   {view==='periods'&&<><label>Início<input type="date" value={from} onChange={event=>setFrom(event.target.value)}/></label><label>Fim (exclusivo)<input type="date" value={to} onChange={event=>setTo(event.target.value)}/></label><p>Pendências e UNKNOWN financeiros impedem o fechamento. As datas são instantes UTC, com início inclusivo e fim exclusivo.</p><button disabled={!from||!to} onClick={()=>void action('periods',{period_start:`${from}T00:00:00.000Z`,period_end:`${to}T00:00:00.000Z`})}>Verificar e fechar período</button></>}
   {view==='adjustments'&&<><label>Lançamento de origem<input value={origin} onChange={event=>setOrigin(event.target.value)}/></label><label>Razão do ajuste<input minLength={8} value={reason} onChange={event=>setReason(event.target.value)}/></label><p>O ator é derivado da identidade autenticada; a aprovação exige outro operador.</p><button disabled={!origin||reason.trim().length<8} onClick={()=>void action('adjustments',{origin_batch_id:origin,reason})}>Preparar ajuste</button>{canApprove&&<p>A aprovação é feita na ação da API com permissão segregada e não pode ser autoaprovada.</p>}</>}
   {view==='disputes'&&<><label>ID do fato<input value={factID} onChange={event=>setFactID(event.target.value)}/></label><label>Valor contestado<input value={amount} onChange={event=>setAmount(event.target.value)}/></label><label>Evidência<input value={evidence} onChange={event=>setEvidence(event.target.value)}/></label><label>Razão<input minLength={8} value={reason} onChange={event=>setReason(event.target.value)}/></label><button disabled={!factID||!amount||!evidence||reason.trim().length<8} onClick={()=>void action('disputes',{fact_id:Number(factID),amount,evidence_id:evidence,reason})}>Abrir divergência</button></>}
   {view==='exports'&&<><label>Recibo externo<input value={evidence} onChange={event=>setEvidence(event.target.value)}/></label><label>Checksum confirmado<input value={amount} onChange={event=>setAmount(event.target.value)}/></label><button disabled={!exportID||!evidence||!amount} onClick={()=>void action(`exports/${encodeURIComponent(exportID)}/receipts`,{receipt_id:evidence,checksum:amount})}>Registrar recibo</button></>}
  </fieldset>}
  {view==='adjustments'&&canApprove&&<p role="note">Para aprovar, abra o ajuste persistido e use o endpoint de aprovação com identidade diferente do preparador.</p>}
 </section>;
}

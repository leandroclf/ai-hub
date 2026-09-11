#!/usr/bin/env python3
"""Renderiza fixture local-kind com dependências reais do Compose R2.
O manifesto gerado contém exclusivamente senhas de ensaio; não usar em nuvem.
"""
import json, os, subprocess, pathlib, yaml
root=pathlib.Path(__file__).resolve().parents[1];namespace='ai-hub-local-kind';project=os.environ.get('R2_COMPOSE_PROJECT','ai_hub_r3qual');compose=['docker','compose','-p',project,'-f',str(root/'compose.yaml')];docs=[{'apiVersion':'v1','kind':'Namespace','metadata':{'name':namespace}}]
for name,port in [('postgres',5432),('localstack',4566),('identity',8080),('alloy',4318),('provider-sim',8090),('webhook-sink',8091)]:
 container=subprocess.check_output([*compose,'ps','-q',name],text=True).strip()
 if not container: raise RuntimeError(f'container Compose ausente para {name} no projeto {project}')
 result=json.loads(subprocess.check_output(['docker','inspect',container]))[0]
 network=next((n for n in result['NetworkSettings']['Networks'] if n.startswith(project+'_')),None)
 if not network: raise RuntimeError(f'container {name} sem rede Compose do projeto {project}')
 address=result['NetworkSettings']['Networks'][network]['IPAddress']
 docs.extend([{'apiVersion':'v1','kind':'Service','metadata':{'name':name,'namespace':namespace},'spec':{'ports':[{'name':'http','port':port,'targetPort':port}]}},{'apiVersion':'v1','kind':'Endpoints','metadata':{'name':name,'namespace':namespace},'subsets':[{'addresses':[{'ip':address}],'ports':[{'name':'http','port':port}]}]}])
data={'ENVIRONMENT':'local','CELL_ID':'r2-cell-a','OIDC_ISSUER':'http://localhost:18085/realms/ai-hub-r2','OIDC_JWKS_URL':'http://identity:8080/realms/ai-hub-r2/protocol/openid-connect/certs','OIDC_AUDIENCE':'ai-hub','OIDC_TOKEN_URL':'http://identity:8080/realms/ai-hub-r2/protocol/openid-connect/token','ATLAS_URL':'http://atlas:8081','ORBITA_URL':'http://orbita:8080','COMETA_URL':'http://cometa:8082','LIBRA_URL':'http://libra:8084','QUEUE_ENDPOINT':'http://localstack:4566','QUEUE_REGION':'us-east-1','QUEUE_NAMESPACE':'r2-kind-cell-a','AWS_ACCESS_KEY_ID':'local','AWS_SECRET_ACCESS_KEY':'local','AWS_REGION':'us-east-1','AWS_ENDPOINT_URL':'http://localstack:4566','REDIS_ADDR':'','OTEL_EXPORTER_OTLP_ENDPOINT':'http://alloy:4318','LOG_LEVEL':'info'}
data['CAPACITY_DOMAINS']='r4-sync,r4-sync-failure,r4-poll-1,r4-poll-2,r4-callback,r4-webhook'
data['CAPACITY_WEBHOOK_DOMAIN']='r4-webhook'
data['CAPACITY_POLICY_JSON']='{"version":"r4-fixture-v1","evidence_ref":"r4-local-capacity-fixture","valid_until":"2099-12-31T23:59:59Z","max_concurrent":8,"min_concurrent":5,"reconciliation_reserve":1,"max_pending":8,"rate_per_window":200,"window_millis":1000,"lease_millis":5000,"stable_millis":100,"latency_threshold_millis":100,"tenant_limits":{"acme":2,"acme-strict":1,"acme-dedicated":1},"tenant_pending_limits":{"acme":4,"acme-strict":2,"acme-dedicated":2},"tenant_rate_limits":{"acme":90,"acme-strict":50,"acme-dedicated":50}}'
data['EGRESS_HTTP_ORIGINS']='http://provider-sim:8090,http://webhook-sink:8091,http://identity:8080'
data['EGRESS_PRIVATE_RULES']='provider-sim:8090=10.96.0.0/12;webhook-sink:8091=10.96.0.0/12;identity:8080=10.96.0.0/12'
docs.append({'apiVersion':'v1','kind':'ConfigMap','metadata':{'name':'hub-runtime','namespace':namespace},'data':data})
docs.append({'apiVersion':'v1','kind':'Secret','metadata':{'name':'hub-runtime','namespace':namespace},'stringData':{domain.upper()+'_DSN':f'postgres://hub:r2-local-fixture@postgres:5432/hub_{domain}_kind?sslmode=disable' for domain in ['control','core','finance']}})
for name in ['atlas','orbita','cometa','pulsar','libra']:
 docs.append({'apiVersion':'v1','kind':'Secret','metadata':{'name':name+'-workload','namespace':namespace},'stringData':{'client-secret':(root/'identity'/f'{name}-secret.txt').read_text().strip()}})
print(yaml.safe_dump_all(docs,sort_keys=False))

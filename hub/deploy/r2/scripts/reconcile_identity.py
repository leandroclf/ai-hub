#!/usr/bin/env python3
"""Reconcilia a fixture OIDC local sem apagar o banco do realm.

O import de realm do Keycloak só cria credenciais na criação inicial. Este
script torna a fixture repetível em um realm já existente: perfis, mapper de
subject, usuários, senha, papel e OTP são verificados/recriados apenas para
os usuários sintéticos declarados abaixo.
"""
import base64, json, os, time, urllib.error, urllib.parse, urllib.request

BASE=os.environ.get("KEYCLOAK_URL","http://identity:8080")
REALM=os.environ.get("KEYCLOAK_REALM","ai-hub-r2")
ADMIN_USER=os.environ.get("KEYCLOAK_ADMIN_USER","r2-bootstrap")
ADMIN_PASSWORD=os.environ.get("KEYCLOAK_ADMIN_PASSWORD","r2-bootstrap-fixture")
PASSWORD=os.environ.get("R2_FIXTURE_PASSWORD","R2-fixture-password!")
OTP_SECRET=os.environ.get("R2_FIXTURE_OTP","JBSWY3DPEHPK3PXP")
USERS=[
    ("operadora-a","acme","app-acme","r2-cell-a","hub_admin"),
    ("leitor-a","acme","app-acme","r2-cell-a","tenant_reader"),
    ("operador-b","beta","app-beta","r2-cell-a","tenant_operator"),
    ("auditor-global","","app-","r2-cell-a","hub_protocol_reader"),
]

def request(path, method="GET", body=None, token=None):
    data=None if body is None else json.dumps(body).encode()
    headers={"Content-Type":"application/json"}
    if token: headers["Authorization"]="Bearer "+token
    req=urllib.request.Request(BASE+path,data=data,headers=headers,method=method)
    with urllib.request.urlopen(req,timeout=15) as res:
        raw=res.read(); return res.status, (json.loads(raw) if raw else None)

def token():
    body=urllib.parse.urlencode({"client_id":"admin-cli","username":ADMIN_USER,"password":ADMIN_PASSWORD,"grant_type":"password"}).encode()
    req=urllib.request.Request(BASE+"/realms/master/protocol/openid-connect/token",data=body,method="POST")
    with urllib.request.urlopen(req,timeout=15) as res: return json.loads(res.read())["access_token"]

def users(t, username):
    _, data=request("/admin/realms/%s/users?username=%s&exact=true"%(REALM,urllib.parse.quote(username)),token=t)
    return data

def partial_user(t, username, tenant, app, cell, role):
    user={"username":username,"enabled":True,"emailVerified":True,"firstName":username,"lastName":"Fixture R2","email":username+"@r2.invalid","attributes":{"tenant_id":[tenant],"application_id":[app],"cell_id":[cell]},"credentials":[{"type":"password","value":PASSWORD,"temporary":False},{"type":"otp","userLabel":"R2 synthetic TOTP","secretData":json.dumps({"value":OTP_SECRET}),"credentialData":json.dumps({"subType":"totp","digits":6,"counter":0,"period":30,"algorithm":"HmacSHA1"})}]}
    request("/admin/realms/%s/partialImport"%REALM,"POST",{"ifResourceExists":"FAIL","users":[user]},t)

def reconcile_user(t, username, tenant, app, cell, role):
    found=users(t,username)
    if found and not found[0].get("totp",False):
        request("/admin/realms/%s/users/%s"%(REALM,found[0]["id"]),"DELETE",token=t)
        partial_user(t,username,tenant,app,cell,role); found=users(t,username)
    if not found: partial_user(t,username,tenant,app,cell,role); found=users(t,username)
    uid=found[0]["id"]
    request("/admin/realms/%s/users/%s"%(REALM,uid),"PUT",{"username":username,"enabled":True,"emailVerified":True,"firstName":username,"lastName":"Fixture R2","email":username+"@r2.invalid","attributes":{"tenant_id":[tenant],"application_id":[app],"cell_id":[cell]}},t)
    request("/admin/realms/%s/users/%s/reset-password"%(REALM,uid),"PUT",{"type":"password","value":PASSWORD,"temporary":False},t)
    roles=request("/admin/realms/%s/roles/%s"%(REALM,role),token=t)[1]
    request("/admin/realms/%s/users/%s/role-mappings/realm"%(REALM,uid),"POST",[roles],t)

def main():
    for _ in range(30):
        try: t=token(); break
        except (urllib.error.URLError,urllib.error.HTTPError): time.sleep(2)
    else: raise SystemExit("Keycloak indisponível para reconciliação")
    profile=request("/admin/realms/%s/users/profile"%REALM,token=t)[1]
    attrs={a["name"]:a for a in profile.get("attributes",[])}
    for name, label in (("tenant_id","Tenant"),("application_id","Application"),("cell_id","Cell")):
        attrs[name]={"name":name,"displayName":label,"permissions":{"view":["admin","user"],"edit":["admin"]},"multivalued":False}
    profile["attributes"]=list(attrs.values()); request("/admin/realms/%s/users/profile"%REALM,"PUT",profile,t)
    clients=request("/admin/realms/%s/clients?clientId=ai-hub-admin"%REALM,token=t)[1]; client=clients[0]
    if not any(m.get("name")=="subject" for m in client.get("protocolMappers",[])):
        request("/admin/realms/%s/clients/%s/protocol-mappers/models"%(REALM,client["id"]),"POST",{"name":"subject","protocol":"openid-connect","protocolMapper":"oidc-usermodel-property-mapper","config":{"user.attribute":"id","claim.name":"sub","jsonType.label":"String","access.token.claim":"true","id.token.claim":"true"}},t)
    for user in USERS: reconcile_user(t,*user)
    print("identity reconciliation: PASS")
if __name__=="__main__": main()

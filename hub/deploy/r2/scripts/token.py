#!/usr/bin/env python3
"""Emite token somente do IdP de ensaio, com senha+OTP sintéticos; não salva token."""
import argparse, base64, hashlib, hmac, json, struct, time, urllib.parse, urllib.request
p=argparse.ArgumentParser();p.add_argument('username',nargs='?',default='operadora-a');p.add_argument('--claims',action='store_true');args=p.parse_args()
public_secret='IFES2SCVIIWVEMRNJVDECLKLIVMS2MBR'
key=base64.b32decode(public_secret+'='*((8-len(public_secret)%8)%8))
counter=struct.pack('>Q',int(time.time())//30)
digest=hmac.new(key,counter,hashlib.sha1).digest();offset=digest[-1]&15
otp=str((struct.unpack('>I',digest[offset:offset+4])[0]&0x7fffffff)%1000000).zfill(6)
body=urllib.parse.urlencode({'client_id':'ai-hub-fixture','grant_type':'password','username':args.username,'password':'R2-fixture-password!','totp':otp}).encode()
try:
 with urllib.request.urlopen(urllib.request.Request('http://localhost:18085/realms/ai-hub-r2/protocol/openid-connect/token',data=body),timeout=10) as response: token=json.load(response)['access_token']
except urllib.error.HTTPError as error:
 raise SystemExit(error.read().decode())
if args.claims:
 claims=json.loads(base64.urlsafe_b64decode(token.split('.')[1]+'=='))
 print(json.dumps({k:claims.get(k) for k in ['iss','aud','tenant_id','application_id','cell_id','scope','amr','realm_access']},indent=2))
else:print(token)

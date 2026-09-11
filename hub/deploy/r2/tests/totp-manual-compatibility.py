#!/usr/bin/env python3
"""Prova que o segredo Base32 entregue ao usuário autentica no Keycloak local."""
import base64
import hashlib
import hmac
import json
import struct
import time
import urllib.error
import urllib.parse
import urllib.request

BASE_URL = "http://localhost:18085"
REALM = "ai-hub-r2"
CLIENT_ID = "ai-hub-fixture"
USERNAME = "operadora-a"
PASSWORD = "R2-fixture-password!"
PUBLIC_SECRET = "IFES2SCVIIWVEMRNJVDECLKLIVMS2MBR"


def totp(secret, skew=0):
    counter = struct.pack(">Q", int(time.time()) // 30 + skew)
    digest = hmac.new(secret, counter, hashlib.sha1).digest()
    offset = digest[-1] & 15
    return f"{(struct.unpack('>I', digest[offset:offset + 4])[0] & 0x7fffffff) % 1000000:06d}"


def main():
    secret = base64.b32decode(PUBLIC_SECRET + "=" * ((8 - len(PUBLIC_SECRET) % 8) % 8))
    if secret != b"AI-HUB-R2-MFA-KEY-01":
        raise SystemExit("FAIL: segredo Base32 não corresponde à chave do fixture")
    body = urllib.parse.urlencode(
        {
            "client_id": CLIENT_ID,
            "grant_type": "password",
            "username": USERNAME,
            "password": PASSWORD,
            "totp": totp(secret),
        }
    ).encode()
    request = urllib.request.Request(
        f"{BASE_URL}/realms/{REALM}/protocol/openid-connect/token",
        data=body,
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=15) as response:
            if response.status != 200:
                raise SystemExit(f"FAIL: Keycloak retornou HTTP {response.status}")
    except urllib.error.HTTPError as error:
        raise SystemExit(f"FAIL: Keycloak rejeitou o OTP manual (HTTP {error.code})") from error
    print(json.dumps({"check": "manual Base32 TOTP", "status": "PASS"}))


if __name__ == "__main__":
    main()

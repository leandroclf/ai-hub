#!/usr/bin/env python3
"""Qualifica a fronteira local do leitor global sem persistir tokens."""
import base64
import json
import subprocess
import urllib.error
import urllib.request


TOKEN_SCRIPT = "hub/deploy/r2/scripts/token.py"


def fixture_token(username):
    return subprocess.check_output(["python3", TOKEN_SCRIPT, username], text=True).strip()


def claims(token):
    return json.loads(base64.urlsafe_b64decode(token.split(".")[1] + "=="))


def request(token, method, url, body=None):
    payload = None if body is None else json.dumps(body).encode()
    headers = {"Authorization": "Bearer " + token}
    if body is not None:
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url, data=payload, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=10) as response:
            return response.status, json.loads(response.read())
    except urllib.error.HTTPError as error:
        raw = error.read()
        return error.code, json.loads(raw) if raw else {}


def has_secret_key(value):
    if isinstance(value, dict):
        if any(key in value for key in ("secret", "client_secret", "access_token", "refresh_token")):
            return True
        return any(has_secret_key(item) for item in value.values())
    if isinstance(value, list):
        return any(has_secret_key(item) for item in value)
    return False


def main():
    auditor = fixture_token("auditor-global")
    auditor_claims = claims(auditor)
    required_roles = "hub_protocol_reader" in auditor_claims.get("realm_access", {}).get("roles", [])
    required_scope = "admin:cross_tenant" in auditor_claims.get("scope", "").split()
    if not auditor_claims.get("sub") or auditor_claims.get("amr") != ["pwd", "otp"] or not required_roles or not required_scope:
        raise SystemExit("FAIL: claims globais não atendem sujeito, MFA, papel e concessão explícita")

    atlas_status, atlas_body = request(auditor, "GET", "http://localhost:18081/admin/v1/applications?tenant_id=acme&limit=1")
    orbita_status, orbita_body = request(auditor, "GET", "http://localhost:18080/admin/v1/protocols?tenant_id=acme&reason=diagnostico+operacional")
    if atlas_status != 200 or orbita_status != 200 or has_secret_key(atlas_body) or has_secret_key(orbita_body):
        raise SystemExit("FAIL: leitura global não retornou projeções administrativas sem segredo")

    write_status, _ = request(
        auditor,
        "POST",
        "http://localhost:18080/admin/v1/protocols/nao-executar/reconcile?tenant_id=acme",
        {"reason": "não elevar leitor global"},
    )
    if write_status != 403:
        raise SystemExit("FAIL: leitor global recebeu permissão de escrita")

    reader = fixture_token("leitor-a")
    reader_status, _ = request(reader, "GET", "http://localhost:18081/admin/v1/applications?tenant_id=beta&limit=1")
    if reader_status != 403:
        raise SystemExit("FAIL: leitor de tenant cruzou o próprio escopo")

    print("IDENTITY_SCOPE_PROOF=PASS global_read=200/200 global_write=403 tenant_cross_read=403")


if __name__ == "__main__":
    main()

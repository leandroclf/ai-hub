"""Cenário exploratório somente leitura para o console administrativo R4.

Pré-condição: operador já autenticado com uma conta de fixture no navegador.
Senha, MFA, cookies e tokens não são tratados pelo cenário.
"""

import os


base = os.environ.get("R2_ADMIN_URL", "http://localhost:13000")
routes = ("services", "clients", "applications", "providers", "provider-accounts")

new_tab(f"{base}/services")
wait_for_load()
print(page_info())

for route in routes:
    goto_url(f"{base}/{route}")
    wait_for_load()
    info = page_info()
    print({"route": route, "page": info})

    body = js("document.body.innerText || ''")
    if len(body.strip()) < 20:
        raise AssertionError(f"tela sem conteúdo útil: {route}")

    js("window.scrollTo(0, 0)")
    viewport = js("({width: window.innerWidth, scrollWidth: document.documentElement.scrollWidth})")
    if viewport["scrollWidth"] > viewport["width"]:
        raise AssertionError(f"overflow horizontal em {route}: {viewport}")

print({"status": "PASS-EXPLORATORY", "routes": list(routes), "viewport": 390})

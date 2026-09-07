import { chromium } from "playwright-core";

const browser = await chromium.launch({ executablePath: "/usr/bin/google-chrome", headless: true });
const page = await browser.newPage({ viewport: { width: 1440, height: 950 } });
const checks = [];

async function check(label, action, expected) {
  await action();
  const text = await page.locator("body").innerText();
  if (!text.includes(expected)) throw new Error(`${label}: texto esperado ausente: ${expected}`);
  checks.push(`${label}: PASS`);
}

async function lookup(tabLabel, values, expected) {
  await page.getByRole("button", { name: new RegExp(tabLabel, "i") }).click();
  const form = page.locator("form").filter({ has: page.getByRole("button", { name: /buscar|resolver/i }) }).first();
  const inputs = form.locator("input");
  for (let i = 0; i < values.length; i++) await inputs.nth(i).fill(values[i]);
  await form.getByRole("button", { name: /buscar|resolver/i }).click();
  await page.waitForTimeout(600);
  const text = await page.locator("body").innerText();
  if (!text.includes(expected)) throw new Error(`${tabLabel}: texto esperado ausente: ${expected}`);
  checks.push(`${tabLabel}: PASS`);
}

await page.goto("http://localhost:5173/", { waitUntil: "networkidle" });
await lookup("Catálogo de serviços", ["hiveplace-01-token-request", "1"], "hiveplace-01-token-request");
await lookup("Contas de provedor", ["provider-hiveplace-hml"], "provider-hiveplace-hml");
await check("credenciais", async () => {
  await page.getByRole("button", { name: /Vínculos de credencial/i }).click();
  const form = page.locator("form").filter({ has: page.getByRole("button", { name: /resolver/i }) }).first();
  await form.locator("input").nth(0).fill("hiveplace-sandbox");
  await form.locator("input").nth(1).fill("provider-hiveplace-hml");
  await form.getByRole("button", { name: /resolver/i }).click();
  await page.waitForTimeout(500);
}, "SHARED_HUB");
await lookup("Contratos", ["hiveplace-sandbox"], "hiveplace-sandbox");
await page.goto("http://localhost:8092/", { waitUntil: "networkidle" });
await check("swagger", () => page.waitForTimeout(500), "Orbita - API publica");

console.log(checks.join("\n"));
await browser.close();

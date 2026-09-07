import { chromium } from "playwright-core";
import path from "node:path";

const outDir = path.dirname(new URL(import.meta.url).pathname);

const browser = await chromium.launch({
  executablePath: "/usr/bin/google-chrome",
  headless: true,
});
const page = await browser.newPage({ viewport: { width: 1440, height: 950 } });
page.on("console", (m) => console.log("[browser]", m.type(), m.text()));

async function lookupAndShoot(tabLabel, lookupValues, fileSlug) {
  const btn = page.getByRole("button", { name: new RegExp(tabLabel, "i") }).first();
  await btn.click();
  await page.waitForTimeout(200);

  const form = page.locator("form").filter({ has: page.getByRole("button", { name: /buscar|resolver/i }) }).first();
  const inputs = form.locator("input");
  const count = await inputs.count();
  for (let i = 0; i < Math.min(count, lookupValues.length); i++) {
    await inputs.nth(i).fill(String(lookupValues[i]));
  }
  await form.getByRole("button", { name: /buscar|resolver/i }).click();
  await page.waitForTimeout(600);
  await page.screenshot({ path: `${outDir}/${fileSlug}.png`, fullPage: true });
  console.log(`captured ${fileSlug}`);
}

await page.goto("http://localhost:5173/", { waitUntil: "networkidle" });
await page.screenshot({ path: `${outDir}/admin-ui-00-inicial.png`, fullPage: true });
console.log("captured admin-ui-00-inicial");

await lookupAndShoot("Catálogo de serviços", ["consulta-cadastral", "1"], "admin-ui-01-servicos");
await lookupAndShoot("Contas de provedor", ["prov-oauth-1"], "admin-ui-02-contas-provedor");
await lookupAndShoot("Vínculos de credencial", ["acme", "prov-sync-1"], "admin-ui-03-credenciais");
await lookupAndShoot("Contratos", ["acme"], "admin-ui-04-contratos");

// --- Swagger UI ---
await page.goto("http://localhost:8092/", { waitUntil: "networkidle" });
await page.waitForTimeout(1000);
await page.screenshot({ path: `${outDir}/swagger-ui-01-overview.png`, fullPage: true });
console.log("captured swagger-ui overview");

const opBlock = page.locator(".opblock").first();
if (await opBlock.count() > 0) {
  await opBlock.click();
  await page.waitForTimeout(500);
  await page.screenshot({ path: `${outDir}/swagger-ui-02-endpoint-expanded.png`, fullPage: true });
  console.log("captured swagger-ui endpoint detail");
}

await browser.close();
console.log("done");

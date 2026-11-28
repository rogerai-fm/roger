// Real-browser flows against the built site (dist/ by default, or WEB_DIST). The broker is
// intercepted at the network layer in each test, so the hard-coded broker origin stays.
import { defineConfig } from "@playwright/test";
import { fileURLToPath } from "node:url";
// WEB_E2E_PORT: CI picks a free port; by hand it is 4178
const PORT = Number(process.env.WEB_E2E_PORT) || 4178;
const DIST = process.env.WEB_DIST || fileURLToPath(new URL("./dist", import.meta.url));
export default defineConfig({
  testDir: "e2e",
  timeout: 30_000,
  use: { baseURL: `http://127.0.0.1:${PORT}`, headless: true },
  // the list on the terminal, and an html report (playwright-report/, gitignored) CI keeps when a flow fails
  reporter: [["list"], ["html", { open: "never", outputFolder: "playwright-report" }]],
  // locally a server already on the port is reused; CI always starts its own
  webServer: { command: `python3 -m http.server ${PORT} --bind 127.0.0.1 --directory ${DIST}`, url: `http://127.0.0.1:${PORT}/index.html`, reuseExistingServer: !process.env.CI },
  projects: [{ name: "chromium", use: { browserName: "chromium" } }],
});

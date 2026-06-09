import { spawnSync } from "node:child_process";
import { mkdirSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = dirname(fileURLToPath(import.meta.url));
const appDir = resolve(scriptDir, "..");
const repoRoot = resolve(appDir, "../..");
const binariesDir = resolve(appDir, "src-tauri/binaries");

function targetTriple() {
  if (process.platform !== "darwin") {
    throw new Error(`macOS app sidecar builds require darwin, got ${process.platform}`);
  }
  if (process.arch === "arm64") {
    return "aarch64-apple-darwin";
  }
  if (process.arch === "x64") {
    return "x86_64-apple-darwin";
  }
  throw new Error(`Unsupported macOS architecture: ${process.arch}`);
}

mkdirSync(binariesDir, { recursive: true });

const goBin = process.env.GO_BIN || "go";
const outputPath = resolve(binariesDir, `jini-sidecar-${targetTriple()}`);
const result = spawnSync(goBin, ["build", "-o", outputPath, "./cmd/jini"], {
  cwd: repoRoot,
  env: { ...process.env, CGO_ENABLED: process.env.CGO_ENABLED || "0" },
  stdio: "inherit"
});

if (result.error) {
  throw result.error;
}
if (result.status !== 0) {
  process.exit(result.status ?? 1);
}

console.log(`Prepared Go sidecar: ${outputPath}`);

import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const appDir = resolve(dirname(fileURLToPath(import.meta.url)), "..");

const tauriConfig = JSON.parse(readFileSync(resolve(appDir, "src-tauri/tauri.conf.json"), "utf8"));
const capabilities = readFileSync(resolve(appDir, "src-tauri/capabilities/default.json"), "utf8");
const sidecar = readFileSync(resolve(appDir, "src/sidecar.js"), "utf8");

const failures = [];

if (tauriConfig.productName !== "Jini") {
  failures.push("tauri.conf.json must use productName Jini");
}
if (tauriConfig.identifier !== "ai.maridlabs.jini") {
  failures.push("tauri.conf.json must use ai.maridlabs.jini");
}
if (!tauriConfig.bundle?.externalBin?.includes("binaries/jini-sidecar")) {
  failures.push("bundle.externalBin must include binaries/jini-sidecar");
}
for (const required of ["shell:allow-spawn", "shell:allow-stdin-write", "binaries/jini-sidecar"]) {
  if (!capabilities.includes(required)) {
    failures.push(`capabilities/default.json must include ${required}`);
  }
}
for (const forbidden of ["shell:default", "fs:default", "fs:allow-write", "dialog:default"]) {
  if (capabilities.includes(forbidden)) {
    failures.push(`capabilities/default.json must not include broad permission ${forbidden}`);
  }
}
for (const required of ["Command.sidecar(\"binaries/jini-sidecar\"", "macos-app-v1", "idempotency_key"]) {
  if (!sidecar.includes(required)) {
    failures.push(`sidecar bridge must include ${required}`);
  }
}

if (failures.length > 0) {
  for (const failure of failures) {
    console.error(failure);
  }
  process.exit(1);
}

console.log("macOS shell contract ok");

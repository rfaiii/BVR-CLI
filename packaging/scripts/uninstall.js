import fs from "node:fs";
import path from "node:path";

const TARGET = process.argv[1] || path.join(process.cwd(), "node_modules", ".bin", "bvr-cli");
fs.rmSync(TARGET, { force: true });
console.log("Removed bvr-cli");

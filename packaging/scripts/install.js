import https from "node:https";
import http from "node:http";
import fs from "node:fs";
import path from "node:path";
import os from "node:os";
import { execSync } from "node:child_process";

const GITHUB_REPO = "richavery/bvr-cli-main";
const VERSION = process.env.npm_package_version || "latest";
const GITHUB_TOKEN = process.env.GITHUB_TOKEN || process.env.nPm_CONFIG_GITHUB_TOKEN || "";
const HOME = os.homedir();
const INSTALL_DIR = process.env.npm_config_global
  ? path.join(execSync("npm root -g").toString().trim(), ".bin")
  : path.join(process.cwd(), "node_modules", ".bin");
const TARGET = path.join(INSTALL_DIR, "bvr-cli");

function ensureBinDir() {
  if (!fs.existsSync(INSTALL_DIR)) {
    fs.mkdirSync(INSTALL_DIR, { recursive: true });
  }
}

function api(url) {
  return new Promise((resolve, reject) => {
    const headers = { "User-Agent": "bvr-cli-npm-installer" };
    if (GITHUB_TOKEN) headers["Authorization"] = `Bearer ${GITHUB_TOKEN}`;
    https.get(url, { headers }, (res) => {
      let data = "";
      res.on("data", (chunk) => (data += chunk));
      res.on("end", () => {
        try {
          resolve(JSON.parse(data));
        } catch (err) {
          reject(err);
        }
      });
    }).on("error", reject);
  });
}

function getRelease() {
  return api(`https://api.github.com/repos/${GITHUB_REPO}/releases/latest`);
}

function download(url, target) {
  return new Promise((resolve, reject) => {
    const mod = url.startsWith("https") ? https : http;
    const file = fs.createWriteStream(target || TARGET);
    mod.get(url, (res) => {
      if (res.statusCode && res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        download(res.headers.location, target).then(resolve).catch(reject);
        res.resume();
        return;
      }
      res.pipe(file);
      file.on("finish", () => {
        file.close();
        if (!target || target === TARGET) fs.chmodSync(TARGET, 0o755);
        resolve();
      });
    }).on("error", (err) => {
      fs.rmSync(target || TARGET, { force: true });
      reject(err);
    });
  });
}

function resolveAssetName(release, platform, arch) {
  const rawTag = release.tag_name || `v${VERSION}`;
  const tag = rawTag.startsWith("v") ? rawTag.slice(1) : rawTag;
  const osName = platform === "win32" ? "Windows" : platform === "darwin" ? "darwin" : "linux";
  const archName = arch === "x64" ? "amd64" : arch;
  const base = `bvr-cli_${tag}_${osName}_${archName}`;
  return platform === "win32" ? `${base}.exe` : base;
}

function tryFindAsset(release, assetName) {
  if (!release.assets || !Array.isArray(release.assets)) return null;
  return release.assets.find((item) => item.name === assetName) || null;
}

function buildFallbackUrl(tag, assetName) {
  const version = tag.startsWith("v") ? tag.slice(1) : tag;
  return `https://github.com/${GITHUB_REPO}/releases/download/v${version}/${assetName}`;
}

async function main() {
  console.log(`Installing bvr-cli@${VERSION || "latest"}...`);
  ensureBinDir();

  let release;
  try {
    release = await getRelease();
  } catch (err) {
    console.warn(`Release lookup failed: ${err.message}`);
    release = { tag_name: `v${VERSION}`, assets: [] };
  }

  const platform = os.platform();
  const arch = os.arch();
  const assetName = resolveAssetName(release, platform, arch);
  const tag = release.tag_name || `v${VERSION}`;
  let asset = tryFindAsset(release, assetName);
  let downloadUrl = asset?.browser_download_url || buildFallbackUrl(tag, assetName);

  const tempTarget = TARGET + ".tmp";
  await download(downloadUrl, tempTarget);

  if (!fs.existsSync(tempTarget) || fs.statSync(tempTarget).size === 0) {
    fs.rmSync(tempTarget, { force: true });
    console.error(
      `Downloaded file is empty or missing for ${assetName}. ` +
        `Install the binary manually from https://github.com/${GITHUB_REPO}/releases/tag/${tag}`
    );
    process.exit(1);
  }

  const buffer = fs.readFileSync(tempTarget);
  if (platform === "win32" && !buffer.toString("ascii", 0, 2).startsWith("MZ")) {
    fs.rmSync(tempTarget, { force: true });
    console.error(
      `Downloaded file does not look like a Windows executable for ${assetName}. ` +
        `Install the binary manually from https://github.com/${GITHUB_REPO}/releases/tag/${tag}`
    );
    process.exit(1);
  }

  if (platform !== "win32" && buffer.length > 0 && buffer[0] === 0x7f && buffer[1] === 0x45 && buffer[2] === 0x4c && buffer[3] === 0x46) {
    fs.renameSync(tempTarget, TARGET);
    fs.chmodSync(TARGET, 0o755);
    console.log(`Installed bvr-cli to ${TARGET}`);
    return;
  }

  fs.rmSync(tempTarget, { force: true });
  console.error(
    `Downloaded file for ${assetName} is not a valid binary. ` +
      `Install the binary manually from https://github.com/${GITHUB_REPO}/releases/tag/${tag}`
  );
  process.exit(1);
}

main().catch((err) => {
  console.error(err.message);
  process.exit(1);
});

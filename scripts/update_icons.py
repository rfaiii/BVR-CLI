#!/usr/bin/env python3
"""Install a designer-supplied app-icon set across every icon surface.

Usage: scripts/update_icons.py <source-dir>

<source-dir> must contain the iOS asset-catalog names (Icon-20.png ...
Icon-1024.png). Everything else is derived from the 1024 master: the
desktop .icns/.ico, the tool-style resources/icons/ios set, the packaging
staging copies, and the PNG embedded for OSC 99 notifications.

iOS app icons are flattened onto black with no alpha channel -- the App
Store rejects a 1024 image with transparency, and the source exports put
a transparent border outside the rounded square.
"""

import shutil
import struct
import subprocess
import sys
import tempfile
from pathlib import Path

from PIL import Image

ROOT = Path(__file__).resolve().parent.parent
BG = (0, 0, 0)

# iOS asset-catalog names -> pixel size, used to validate the source set.
CATALOG = {
    "Icon-20.png": 20, "Icon-20@2x.png": 40, "Icon-20@3x.png": 60,
    "Icon-29.png": 29, "Icon-29@2x.png": 58, "Icon-29@3x.png": 87,
    "Icon-40.png": 40, "Icon-40@2x.png": 80, "Icon-40@3x.png": 120,
    "Icon-60@2x.png": 120, "Icon-60@3x.png": 180,
    "Icon-76.png": 76, "Icon-76@2x.png": 152, "Icon-83.5@2x.png": 167,
    "Icon-1024.png": 1024,
}

# iconutil expects these names in a .iconset folder.
ICONSET = {
    "icon_16x16.png": 16, "icon_16x16@2x.png": 32,
    "icon_32x32.png": 32, "icon_32x32@2x.png": 64,
    "icon_128x128.png": 128, "icon_128x128@2x.png": 256,
    "icon_256x256.png": 256, "icon_256x256@2x.png": 512,
    "icon_512x512.png": 512, "icon_512x512@2x.png": 1024,
}

ICO_SIZES = "16,24,32,48,64,128,256"


def png_size(path: Path) -> tuple[int, int]:
    with path.open("rb") as fh:
        head = fh.read(24)
    if head[:8] != b"\x89PNG\r\n\x1a\n":
        raise SystemExit(f"{path} is not a PNG")
    return struct.unpack(">II", head[16:24])


def load_master(src_dir: Path) -> Image.Image:
    missing = [n for n in CATALOG if not (src_dir / n).is_file()]
    if missing:
        raise SystemExit(f"source set is missing: {', '.join(missing)}")
    for name, want in CATALOG.items():
        got = png_size(src_dir / name)
        if got != (want, want):
            raise SystemExit(f"{name} is {got[0]}x{got[1]}, expected {want}x{want}")
    return flatten(Image.open(src_dir / "Icon-1024.png"))


def flatten(im: Image.Image) -> Image.Image:
    """Composite onto black and drop the alpha channel."""
    rgba = im.convert("RGBA")
    out = Image.new("RGB", rgba.size, BG)
    out.paste(rgba, mask=rgba.getchannel("A"))
    return out


def save(im: Image.Image, path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    im.save(path, optimize=True)
    print(f"  {path.relative_to(ROOT)}")


def resized(master: Image.Image, px: int) -> Image.Image:
    if px == master.width:
        return master
    return master.resize((px, px), Image.LANCZOS)


def main() -> None:
    if len(sys.argv) != 2:
        raise SystemExit(__doc__)
    src_dir = Path(sys.argv[1]).expanduser().resolve()
    if not src_dir.is_dir():
        raise SystemExit(f"no such directory: {src_dir}")

    master = load_master(src_dir)
    print(f"master: {src_dir / 'Icon-1024.png'}")

    print("\niOS asset catalog")
    for name in CATALOG:
        save(flatten(Image.open(src_dir / name)),
             ROOT / "mobile-cli/ios/BvrCompanion/BvrCompanion/Assets.xcassets/AppIcon.appiconset" / name)

    print("\npackaging staging")
    for name in CATALOG:
        save(flatten(Image.open(src_dir / name)), ROOT / "packaging/icons" / name)
    save(master, ROOT / "packaging/icons/bvr-cli-1024.png")

    print("\ndesktop master + iOS generator set")
    for name in ("bvr-logo-1024.png", "bvr-icon-00.png"):
        save(master, ROOT / "resources/icons" / name)
    for path in sorted((ROOT / "resources/icons/ios").glob("Icon-iOS-Default-*.png")):
        spec = path.name.removeprefix("Icon-iOS-Default-").removesuffix(".png")
        size, scale = spec.split("@")
        w = float(size.split("x")[0])
        save(resized(master, round(w * int(scale.rstrip("x")))), path)

    print("\nembedded notification icon")
    # Kept small on purpose: OSC 99 base64-encodes this PNG into every
    # notification it writes to the terminal, so a 1024 master would put
    # roughly half a megabyte of escape sequence on the pty per notice.
    for path in sorted((ROOT / "internal/ui/notification").glob("bvr-icon*.png")):
        save(resized(master, 128), path)

    print("\n.icns / .ico")
    iconset = Path(tempfile.mkdtemp(prefix="bvr-iconset-")) / "bvr-icon.iconset"
    iconset.mkdir()
    for name, px in ICONSET.items():
        resized(master, px).save(iconset / name)
    icns = ROOT / "resources/icons/bvr-icon.icns"
    subprocess.run(["iconutil", "-c", "icns", str(iconset), "-o", str(icns)], check=True)
    shutil.rmtree(iconset.parent)
    print(f"  {icns.relative_to(ROOT)}")
    ico = ROOT / "resources/icons/bvr-icon.ico"
    subprocess.run(
        ["magick", str(ROOT / "resources/icons/bvr-logo-1024.png"),
         "-define", f"icon:auto-resize={ICO_SIZES}", str(ico)],
        check=True,
    )
    print(f"  {ico.relative_to(ROOT)}")


if __name__ == "__main__":
    main()

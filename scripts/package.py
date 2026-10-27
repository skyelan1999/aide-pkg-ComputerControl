"""Package public plugin and already-verified Windows native hosts. No publication."""
from pathlib import Path
import hashlib,json,zipfile
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'dist/release';OUT.mkdir(parents=True,exist_ok=True)
version=json.loads((ROOT/'manifest.json').read_text())['version']
plugin=OUT/f'aide-pkg-ComputerControl-v{version}.zip'
with zipfile.ZipFile(plugin,'w',zipfile.ZIP_DEFLATED) as z:
 for name in ('index.js','manifest.json','LICENSE','README.md'):z.write(ROOT/name,name)
for arch in ('amd64','386','arm64'):
 with zipfile.ZipFile(OUT/f'AideComputerBridge-v{version}-windows-{arch}.zip','w',zipfile.ZIP_DEFLATED) as z:
  z.write(ROOT/f'dist/AideComputerBridge-windows-{arch}.exe','AideComputerBridge.exe')
  z.write(ROOT/'dist/AideComputerDesktop.exe','AideComputerDesktop.exe')
  z.write(ROOT/'windows/start.cmd','start.cmd')
  z.write(ROOT/'docs/windows.md','README.md')
  z.write(ROOT/'LICENSE','LICENSE')
  z.write(ROOT/'windows/GO-LICENSE','GO-LICENSE')
files=sorted(OUT.glob('*.zip'))
for p in files:
 with zipfile.ZipFile(p) as z:assert z.testzip() is None
(OUT/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in files))
print('Packaged',*[p.name for p in files],sep='\n')

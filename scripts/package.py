"""Package public plugin and already-verified Windows native hosts. No publication."""
from pathlib import Path
import hashlib,json,zipfile,struct,subprocess,os
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'dist/release';OUT.mkdir(parents=True,exist_ok=True)
version=json.loads((ROOT/'manifest.json').read_text())['version']
def add(z,path,name):
 info=zipfile.ZipInfo(name,(2026,1,1,0,0,0));info.compress_type=zipfile.ZIP_DEFLATED;info.external_attr=0o100644<<16
 z.writestr(info,path.read_bytes())
plugin=OUT/f'aide-pkg-ComputerControl-v{version}.zip'
with zipfile.ZipFile(plugin,'w',zipfile.ZIP_DEFLATED) as z:
 for name in ('index.js','manifest.json','LICENSE','README.md'):add(z,ROOT/name,name)
for arch in ('amd64','386','arm64'):
 data=(ROOT/f'dist/AideComputerBridge-windows-{arch}.exe').read_bytes()
 pe=struct.unpack_from('<I',data,0x3c)[0]
 assert data[pe:pe+4]==b'PE\0\0' and struct.unpack_from('<H',data,pe+4)[0]=={'amd64':0x8664,'386':0x14c,'arm64':0xaa64}[arch]
 with zipfile.ZipFile(OUT/f'AideComputerBridge-v{version}-windows-{arch}.zip','w',zipfile.ZIP_DEFLATED) as z:
  add(z,ROOT/f'dist/AideComputerBridge-windows-{arch}.exe','AideComputerBridge.exe')
  add(z,ROOT/'dist/AideComputerDesktop.exe','AideComputerDesktop.exe')
  add(z,ROOT/'windows/start.cmd','start.cmd')
  add(z,ROOT/'docs/windows.md','README.md')
  add(z,ROOT/'LICENSE','LICENSE')
  add(z,ROOT/'windows/GO-LICENSE','GO-LICENSE')
files=sorted(OUT.glob('*.zip'))
proof={'version':version,'commit':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'runner':os.environ.get('RUNNER_OS','local'),'runnerImage':os.environ.get('ImageOS','local'),'assets':{p.name:{'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in files},'nativeBinaries':{p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((ROOT/'dist').glob('AideComputer*.exe'))},'nativeAcceptance':'Required Windows workflow step succeeded before uploading this artifact','powershellDependency':False}
(OUT/'RELEASE-PROOF.json').write_text(json.dumps(proof,indent=2)+'\n')
for p in files:
 with zipfile.ZipFile(p) as z:assert z.testzip() is None
(OUT/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in files+[OUT/'RELEASE-PROOF.json']))
print('Packaged',*[p.name for p in files],sep='\n')

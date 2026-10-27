"""Exercise real Win32/UIA in a disposable WinForms fixture; never user's apps."""
import base64,json,os,pathlib,subprocess,tempfile,time,urllib.request,io,argparse
from PIL import Image
ROOT=pathlib.Path(__file__).resolve().parents[1]
parser=argparse.ArgumentParser();parser.add_argument('--transport',choices=('amd64','386'),default='amd64');args=parser.parse_args()
helper=ROOT/'dist/AideComputerDesktop.exe'
with tempfile.TemporaryDirectory(prefix='aide desktop fixture ') as tmp:
 info=pathlib.Path(tmp)/'fixture.json'
 proc=subprocess.Popen([str(ROOT/'dist/DesktopFixture.exe'),str(info)])
 pathlib.Path(tmp,'compose.yaml').write_text('services: {}')
 token='t'*64;pathlib.Path(tmp,'.env').write_text('AIDE_COMPUTER_BRIDGE_TOKEN='+token+'\n')
 server=subprocess.Popen([str(ROOT/('dist/AideComputerBridge-windows-'+args.transport+'.exe')),'--aide-dir',tmp],stdout=subprocess.DEVNULL)
 time.sleep(1)
 def state():
  for _ in range(50):
   try:return json.loads(info.read_text(encoding='utf-8-sig'))
   except (OSError,ValueError):time.sleep(.1)
  raise AssertionError('Fixture did not start in an interactive desktop')
 def request(action,**extra):
  req={'allowedApps':[state()['process']],**extra}
  http=urllib.request.Request('http://127.0.0.1:17778/v1/'+action,data=json.dumps(req).encode(),headers={'Authorization':'Bearer '+token,'Content-Type':'application/json'})
  with urllib.request.urlopen(http,timeout=12) as response:result=json.load(response)
  assert result.get('ok'),result
  return result
 try:
  time.sleep(2);data=state()
  denied=subprocess.run([str(helper),'--request'],input=json.dumps({'action':'inspect','allowedApps':['definitely-unrelated']}),text=True,encoding='utf-8',capture_output=True,timeout=12)
  assert not json.loads(denied.stdout)['ok']
  result=request('inspect');assert 'fixture-password-secret' not in json.dumps(result)
  assert 'fixture initial' in json.dumps(result),result
  editor=next(c for c in result['controls'] if c.get('text')=='fixture initial')
  point={k:int(round(editor['center'][k])) for k in ('x','y')}
  result=request('snapshot');png=base64.b64decode(result['imageBase64']);assert png.startswith(b'\x89PNG\r\n\x1a\n')
  with Image.open(io.BytesIO(png)) as image:
   assert sum(1 for pixel in image.convert('RGB').getdata() if pixel==(24,73,112))>1000,'Window paint was not captured'
  request('click',x=point['x'],y=point['y'],approved=True)
  request('type',text=' 中文 Ω',approved=True)
  for _ in range(50):
   if '中文 Ω' in state()['text']:break
   time.sleep(.1)
  else:raise AssertionError({'fixture':state(),'controls':request('inspect')})
  before=state()['text'];request('key',key='BACKSPACE',approved=True)
  for _ in range(50):
   if state()['text']!=before:break
   time.sleep(.1)
  else:raise AssertionError('Backspace was not delivered to the fixture')
  outside=subprocess.run([str(helper),'--request'],input=json.dumps({'action':'click','allowedApps':[data['process']],'approved':True,'x':-30000,'y':-30000}),text=True,encoding='utf-8',capture_output=True,timeout=12)
  assert not json.loads(outside.stdout)['ok']
  print('PASS: real Windows fixture inspect/password exclusion/window PNG/click/Unicode input/key/out-of-window denial')
 finally:
  server.terminate();server.wait(timeout=10)
  proc.terminate();proc.wait(timeout=10)

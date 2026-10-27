using System;
using System.Collections;
using System.Collections.Generic;
using System.Diagnostics;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading;
using System.Web.Script.Serialization;
using System.Windows.Automation;
using System.Windows.Forms;

public static class DesktopBridge {
 [StructLayout(LayoutKind.Sequential)] public struct Rect {public int Left,Top,Right,Bottom;}
 [StructLayout(LayoutKind.Sequential)] public struct Point {public int X,Y; public Point(int x,int y){X=x;Y=y;}}
 [StructLayout(LayoutKind.Sequential)] public struct Mouse {public int X,Y;public uint Data,Flags,Time;public UIntPtr Extra;}
 [StructLayout(LayoutKind.Sequential)] public struct Keyboard {public ushort Key,Scan;public uint Flags,Time;public UIntPtr Extra;}
 [StructLayout(LayoutKind.Explicit)] public struct Union {[FieldOffset(0)]public Mouse Mouse;[FieldOffset(0)]public Keyboard Keyboard;}
 [StructLayout(LayoutKind.Sequential)] public struct Input {public uint Type;public Union Value;}
 [DllImport("user32.dll")]static extern IntPtr GetForegroundWindow();
 [DllImport("user32.dll")]static extern uint GetWindowThreadProcessId(IntPtr h,out uint id);
 [DllImport("user32.dll")]static extern bool GetWindowRect(IntPtr h,out Rect r);
 [DllImport("user32.dll")]static extern bool PrintWindow(IntPtr h,IntPtr dc,uint flags);
 [DllImport("user32.dll")]static extern bool SetCursorPos(int x,int y);
 [DllImport("user32.dll")]static extern IntPtr WindowFromPoint(Point p);
 [DllImport("user32.dll")]static extern IntPtr GetAncestor(IntPtr h,uint flags);
 [DllImport("user32.dll")]static extern uint SendInput(uint count,Input[] input,int size);
 [DllImport("user32.dll")]static extern bool SetProcessDPIAware();
 [DllImport("shcore.dll")]static extern int SetProcessDpiAwareness(int value);
 [DllImport("user32.dll")]static extern IntPtr OpenInputDesktop(uint flags,bool inherit,uint access);
 [DllImport("user32.dll",CharSet=CharSet.Unicode)]static extern bool GetUserObjectInformation(IntPtr h,int type,StringBuilder data,int bytes,out int needed);
 [DllImport("user32.dll")]static extern bool CloseDesktop(IntPtr h);
 static string Norm(string s){return s.EndsWith(".exe",StringComparison.OrdinalIgnoreCase)?s.Substring(0,s.Length-4):s;}
 static void Check(IntPtr h){if(h==IntPtr.Zero||h!=GetForegroundWindow())throw new InvalidOperationException("foreground_changed");}
 static void Interactive(){
  if(Process.GetCurrentProcess().SessionId==0)throw new InvalidOperationException("interactive_session_required");
  IntPtr d=OpenInputDesktop(0,false,1);if(d==IntPtr.Zero)throw new InvalidOperationException("desktop_locked_or_unavailable");
  try{var name=new StringBuilder(128);int need;if(!GetUserObjectInformation(d,2,name,256,out need)||!String.Equals(name.ToString(),"Default",StringComparison.OrdinalIgnoreCase))throw new InvalidOperationException("secure_desktop_not_supported");}finally{CloseDesktop(d);}
 }
 static string Text(Dictionary<string,object> r,string key){object v;return r.TryGetValue(key,out v)?v as string:null;}
 static bool Approved(Dictionary<string,object> r){object v;return r.TryGetValue("approved",out v)&&v is bool&&(bool)v;}
 static void Inject(IntPtr h,Input[] inputs){Check(h);if(SendInput((uint)inputs.Length,inputs,Marshal.SizeOf(typeof(Input)))!=inputs.Length)throw new InvalidOperationException("input_blocked_by_windows_integrity_or_session");}
 static Input Key(ushort key,ushort scan,uint flags){var i=new Input();i.Type=1;i.Value.Keyboard.Key=key;i.Value.Keyboard.Scan=scan;i.Value.Keyboard.Flags=flags;return i;}
 public static Dictionary<string,object> Execute(Dictionary<string,object> r){
  Interactive();IntPtr h=GetForegroundWindow();Check(h);uint pid;GetWindowThreadProcessId(h,out pid);
  string app=Process.GetProcessById((int)pid).ProcessName;object raw;bool allowed=false;
  if(r.TryGetValue("allowedApps",out raw)&&raw is ArrayList){foreach(object rule in (ArrayList)raw){if(rule is string&&String.Equals(Norm((string)rule),app,StringComparison.OrdinalIgnoreCase))allowed=true;}}
  // JavaScriptSerializer may use object[] depending on the deserialization entry.
  if(raw is object[]){foreach(object rule in (object[])raw){if(rule is string&&String.Equals(Norm((string)rule),app,StringComparison.OrdinalIgnoreCase))allowed=true;}}
  if(!allowed)throw new InvalidOperationException("foreground_application_not_allowed");
  Rect bounds;if(!GetWindowRect(h,out bounds)||bounds.Right<=bounds.Left||bounds.Bottom<=bounds.Top)throw new InvalidOperationException("window_bounds_unavailable");
  string action=Text(r,"action");var result=new Dictionary<string,object>{{"ok",true},{"app",app},{"processId",pid},{"windowId",h.ToInt64().ToString()},{"coordinateSpace","physical-screen"}};
  if(action=="inspect"){
   var controls=new List<object>();var root=AutomationElement.FromHandle(h);Walk(root,0,controls);Check(h);result["controls"]=controls;result["truncated"]=controls.Count>=300;return result;
  }
  if(action=="snapshot"){
   int width=bounds.Right-bounds.Left,height=bounds.Bottom-bounds.Top;
   if(width>8192||height>8192||(long)width*height>32000000)throw new InvalidOperationException("window_too_large");
   using(var bmp=new Bitmap(width,height,PixelFormat.Format32bppArgb)){
    using(var g=Graphics.FromImage(bmp)){IntPtr dc=g.GetHdc();try{Check(h);if(!PrintWindow(h,dc,2))throw new InvalidOperationException("window_capture_unavailable");}finally{g.ReleaseHdc(dc);}}
    Check(h);using(var stream=new MemoryStream()){bmp.Save(stream,ImageFormat.Png);if(stream.Length>14*1024*1024)throw new InvalidOperationException("capture_too_large");result["mimeType"]="image/png";result["imageBase64"]=Convert.ToBase64String(stream.ToArray());}
   }
   result["width"]=width;result["height"]=height;result["origin"]=new {x=bounds.Left,y=bounds.Top};return result;
  }
  if(!Approved(r))throw new InvalidOperationException("aide_confirmation_required");
  if(action=="click"){
   int x=Convert.ToInt32(r["x"]),y=Convert.ToInt32(r["y"]);
   if(x<bounds.Left||x>=bounds.Right||y<bounds.Top||y>=bounds.Bottom||GetAncestor(WindowFromPoint(new Point(x,y)),2)!=h)throw new InvalidOperationException("point_outside_allowed_foreground_window");
   Check(h);if(!SetCursorPos(x,y))throw new InvalidOperationException("cursor_unavailable");
   var down=new Input();down.Value.Mouse.Flags=2;var up=new Input();up.Value.Mouse.Flags=4;Inject(h,new[]{down,up});result["clicked"]=true;return result;
  }
  if(action=="type"){
   string text=Text(r,"text");if(text==null||text.Length>2000)throw new InvalidOperationException("invalid_text_length");
   foreach(char c in text){Inject(h,new[]{Key(0,c,4),Key(0,c,6)});}result["typed"]=true;return result;
  }
  if(action=="key"){
   var keys=new Dictionary<string,ushort>{{"ENTER",13},{"TAB",9},{"ESCAPE",27},{"BACKSPACE",8},{"DELETE",46},{"UP",38},{"DOWN",40},{"LEFT",37},{"RIGHT",39},{"SPACE",32}};
   string key=Text(r,"key");if(key==null||!keys.ContainsKey(key))throw new InvalidOperationException("unsupported_key");Inject(h,new[]{Key(keys[key],0,0),Key(keys[key],0,2)});result["key"]=key;return result;
  }
  throw new InvalidOperationException("unsupported_operation");
 }
 static void Walk(AutomationElement node,int depth,List<object> controls){
  if(depth>8||controls.Count>=300)return;
  try{
   var c=node.Current;if(c.IsPassword)return;
   var b=c.BoundingRectangle;
   string value=null;object pattern;
   if(node.TryGetCurrentPattern(ValuePattern.Pattern,out pattern))value=((ValuePattern)pattern).Current.Value;
   if(!b.IsEmpty&&!Double.IsInfinity(b.X)&&!Double.IsInfinity(b.Y))controls.Add(new {name=Limit(c.Name),automationId=Limit(c.AutomationId),role=c.ControlType.ProgrammaticName,text=Limit(value),enabled=c.IsEnabled,rect=new{x=b.X,y=b.Y,width=b.Width,height=b.Height},center=new{x=b.X+b.Width/2,y=b.Y+b.Height/2}});
   var child=TreeWalker.ControlViewWalker.GetFirstChild(node);
   while(child!=null&&controls.Count<300){Walk(child,depth+1,controls);child=TreeWalker.ControlViewWalker.GetNextSibling(child);}
  }catch(ElementNotAvailableException){}catch(InvalidOperationException){}
 }
 static string Limit(string s){return s==null?null:s.Substring(0,Math.Min(s.Length,500));}
 [STAThread]public static int Main(string[] args){
  Console.InputEncoding=new UTF8Encoding(false);Console.OutputEncoding=new UTF8Encoding(false);
  try{try{SetProcessDpiAwareness(2);}catch(DllNotFoundException){SetProcessDPIAware();}catch(EntryPointNotFoundException){SetProcessDPIAware();}
   var json=new JavaScriptSerializer();json.MaxJsonLength=20*1024*1024;
   if(args.Length==1&&args[0]=="--pick-folder"){
    using(var picker=new FolderBrowserDialog()){picker.Description="Pair with Aide: select its installation folder. This adds a desktop bridge token to .env. Restart Aide afterwards.";if(picker.ShowDialog()!=DialogResult.OK)return 1;Console.Write(json.Serialize(new{directory=picker.SelectedPath}));return 0;}
   }
   if(args.Length!=1||args[0]!="--request")return 2;
   string input=Console.In.ReadToEnd();if(input.Length>16384)return 2;
   var request=json.Deserialize<Dictionary<string,object>>(input);object result=null;
   var thread=new Thread(delegate(){try{result=Execute(request);}catch(Exception e){result=new {ok=false,error=e is InvalidOperationException?e.Message:"windows_desktop_unavailable"};}});
   thread.IsBackground=true;thread.SetApartmentState(ApartmentState.MTA);thread.Start();
   if(!thread.Join(8000))result=new {ok=false,error="windows_automation_timeout"};
   Console.Write(json.Serialize(result));return 0;
  }catch{Console.Write("{\"ok\":false,\"error\":\"windows_desktop_helper_unavailable\"}");return 1;}
 }
}

using System;using System.IO;using System.Windows.Forms;using System.Web.Script.Serialization;
class DesktopFixture {
 [STAThread]static void Main(string[] args){
 var form=new Form{Text="Aide isolated desktop fixture",Width=600,Height=340,TopMost=true};
 var edit=new TextBox{Name="editor",Text="fixture initial",Left=30,Top=40,Width=400};
 var password=new TextBox{Name="password",Text="fixture-password-secret",UseSystemPasswordChar=true,Left=30,Top=100,Width=400};
 form.Controls.Add(edit);form.Controls.Add(password);var timer=new Timer{Interval=100};
 timer.Tick+=delegate{var point=edit.PointToScreen(new System.Drawing.Point(10,10));try{File.WriteAllText(args[0],new JavaScriptSerializer().Serialize(new {process=System.Diagnostics.Process.GetCurrentProcess().ProcessName,handle=form.Handle.ToInt64().ToString(),x=point.X,y=point.Y,text=edit.Text}));}catch(IOException){}};
 form.Shown+=delegate{form.Activate();edit.Focus();timer.Start();};Application.Run(form);
 }
}

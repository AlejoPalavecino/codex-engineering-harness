package manager

import (
 "context"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "runtime"
 "testing"
)

func TestCompare(t *testing.T){
 cases:=[]struct{a,b string;want int}{
 {"v0.1.3","0.1.2",1},
 {"v1.0.0","1.0.0",0},
 {"0.2.0","v0.1.12",1},
 {"0.1.2","0.1.11",-1},
 }
 for _,c:=range cases{
  got,err:=Compare(c.a,c.b)
  if err!=nil||got!=c.want{t.Errorf("Compare(%s,%s)=%d %v; want %d",c.a,c.b,got,err,c.want)}
 }
 for _,bad:=range []string{"main","v0.1.3-rc","../../main","v0.1.2;whoami"}{
  if _,err:=Compare(bad,"v0.1.2");err==nil{t.Errorf("accepted unsafe version %q",bad)}
 }
}
func TestReleaseJSON(t *testing.T){
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Header.Get("User-Agent")==""{t.Error("missing User-Agent")}
  _,_=w.Write([]byte("`{\"tag_name\":\"v0.1.3\",\"draft\":false,\"prerelease\":false}`"))
 }))
 defer server.Close()
 m:=Manager{Client:server.Client(),LatestURL:server.URL}
 rel,err:=m.Latest(context.Background())
 if err!=nil||rel.Tag!="v0.1.3"{t.Fatalf("latest %+v %v",rel,err)}
}
func TestRejectPrerelease(t *testing.T){
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  _,_=w.Write([]byte("`{\"tag_name\":\"v0.1.3\",\"prerelease\":true}`"))
 }))
 defer server.Close()
 m:=Manager{Client:server.Client(),LatestURL:server.URL}
 if _,err:=m.Latest(context.Background());err==nil{t.Fatal("accepted prerelease")}
}
func TestReadInstalled(t *testing.T){
 m:=Manager{Home:t.TempDir()}
 if v,err:=m.Installed();v!=""||err!=nil{t.Fatalf("not installed: %q %v",v,err)}
 dir:=filepath.Join(m.Home,"harness-install")
 if err:=os.MkdirAll(dir,0700);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(dir,"state.json"),[]byte("{\"version\":\"0.1.2\"}"),0600);err!=nil{t.Fatal(err)}
 if v,err:=m.Installed();v!="0.1.2"||err!=nil{t.Fatalf("installed: %q %v",v,err)}
 if err:=os.WriteFile(filepath.Join(dir,"state.json"),[]byte("bad"),0600);err!=nil{t.Fatal(err)}
 if _,err:=m.Installed();err==nil{t.Fatal("accepted broken manifest")}
}
func TestLocalSource(t *testing.T){
 root:=t.TempDir()
 m:=Manager{Source:root}
 if _,err:=m.Local();err==nil{t.Fatal("accepted incomplete checkout")}
 if err:=os.MkdirAll(filepath.Join(root,"scripts"),0700);err!=nil{t.Fatal(err)}
 for _,name:=range []string{"install.ps1","verify.ps1","rollback.ps1"}{
  if err:=os.WriteFile(filepath.Join(root,"scripts",name),[]byte("# script"),0600);err!=nil{t.Fatal(err)}
 }
 if err:=os.WriteFile(filepath.Join(root,"VERSION"),[]byte("0.1.3\n"),0600);err!=nil{t.Fatal(err)}
 if path,err:=m.Local();err!=nil||path!=root{t.Fatalf("local %s %v",path,err)}
}
func TestRejectUnsupportedOS(t *testing.T){
 if runtime.GOOS!="windows"{
  m:=Manager{}
  if _,err:=m.Script(context.Background(),".","install.ps1",true);err==nil{t.Fatal("PowerShell should not run")}
 }
}

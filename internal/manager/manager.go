// Package manager invokes the existing reviewed installer without rewriting its rules.
package manager

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net/http"
 "os"
 "os/exec"
 "path/filepath"
 "regexp"
 "runtime"
 "strconv"
 "strings"
 "time"
)

const Repository = "AlejoPalavecino/codex-engineering-harness"
const remoteURL = "https://github.com/" + Repository + ".git"
const githubAPI = "https://api.github.com/repos/" + Repository + "/releases/latest"

var releaseTag = regexp.MustCompile("^v[0-9]+\\.[0-9]+\\.[0-9]+$")
var version = regexp.MustCompile("^v?([0-9]+)\\.([0-9]+)\\.([0-9]+)$")

type Manager struct {
 Home string
 Source string
 Client *http.Client
 LatestURL string // test injection only; never exposed through CLI
}

type Release struct {
 Tag string
 URL string
}

type Prepared struct {
 Root string
 Temporary bool
 Version string
 Action string
 Preview string
}

func New(source string) (Manager,error) {
 home:=os.Getenv("CODEX_HOME")
 if home=="" {
  user,err:=os.UserHomeDir()
  if err!=nil { return Manager{},err }
  home=filepath.Join(user,".codex")
 }
 full,err:=filepath.Abs(home)
 if err!=nil{return Manager{},err}
 return Manager{Home:full,Source:source},nil
}

func Compare(a,b string)(int,error){
 aa,bb:=version.FindStringSubmatch(strings.TrimSpace(a)),version.FindStringSubmatch(strings.TrimSpace(b))
 if aa==nil||bb==nil{return 0,fmt.Errorf("versión no reconocida: %q / %q",a,b)}
 for i:=1;i<=3;i++{
  x,_:=strconv.Atoi(aa[i]);y,_:=strconv.Atoi(bb[i])
  if x<y{return -1,nil};if x>y{return 1,nil}
 }
 return 0,nil
}

func (m Manager) Installed() (string,error) {
 raw,err:=os.ReadFile(filepath.Join(m.Home,"harness-install","state.json"))
 if errors.Is(err,os.ErrNotExist){return "",nil}
 if err!=nil{return "",err}
 var obj struct{Version string}
 if err:=json.Unmarshal(raw,&obj);err!=nil{return "",fmt.Errorf("state.json inválido: %w",err)}
 if _,err:=Compare(obj.Version,obj.Version);err!=nil{return "",err}
 return obj.Version,nil
}

func (m Manager) Latest(ctx context.Context)(Release,error){
 url:=m.LatestURL
 if url==""{url=githubAPI}
 req,err:=http.NewRequestWithContext(ctx,"GET",url,nil)
 if err!=nil{return Release{},err}
 req.Header.Set("Accept","application/vnd.github+json")
 req.Header.Set("User-Agent","Chuli-Manager")
 client:=m.Client
 if client==nil{client=&http.Client{Timeout:15*time.Second}}
 resp,err:=client.Do(req)
 if err!=nil{return Release{},err}
 defer resp.Body.Close()
 if resp.StatusCode!=200{return Release{},fmt.Errorf("GitHub Releases: HTTP %d",resp.StatusCode)}
 var v struct{
  Tag string `json:"tag_name"`
  URL string `json:"html_url"`
  Draft bool `json:"draft"`
  Prerelease bool `json:"prerelease"`
 }
 if err:=json.NewDecoder(io.LimitReader(resp.Body,65536)).Decode(&v);err!=nil{return Release{},err}
 if !releaseTag.MatchString(v.Tag)||v.Draft||v.Prerelease {return Release{},errors.New("release no estable o tag no permitido")}
 return Release{v.Tag,v.URL},nil
}

func ValidateSource(root string)(string,error){
 for _,file:=range []string{"scripts/install.ps1","scripts/verify.ps1","scripts/rollback.ps1","VERSION"}{
  if _,err:=os.Stat(filepath.Join(root,filepath.FromSlash(file)));err!=nil{return "",fmt.Errorf("fuente incompleta (%s): %w",file,err)}
 }
 raw,err:=os.ReadFile(filepath.Join(root,"VERSION"))
 if err!=nil{return "",err}
 v:=strings.TrimSpace(string(raw))
 if _,err:=Compare(v,v);err!=nil{return "",err}
 return v,nil
}

func (m Manager) Local() (string,error){
 root:=m.Source
 if root==""{root,_=os.Getwd()}
 abs,err:=filepath.Abs(root)
 if err!=nil{return "",err}
 if _,err=ValidateSource(abs);err!=nil{return "",fmt.Errorf("no se encontró un checkout válido; usá --source RUTA: %w",err)}
 return abs,nil
}

func (m Manager) Download(ctx context.Context,tag string)(string,error){
 if !releaseTag.MatchString(tag){return "",errors.New("tag inválido")}
 if _,err:=exec.LookPath("git");err!=nil{return "",errors.New("Git no está disponible en PATH")}
 root,err:=os.MkdirTemp("","chuli-release-")
 if err!=nil{return "",err}
 cmd:=exec.CommandContext(ctx,"git","clone","--quiet","--depth","1","--branch",tag,"--single-branch",remoteURL,root)
 cmd.Env=append(os.Environ(),"GIT_TERMINAL_PROMPT=0")
 output,err:=cmd.CombinedOutput()
 if err!=nil{os.RemoveAll(root);return "",fmt.Errorf("falló descarga de release: %w (%s)",err,trunc(output))}
 v,err:=ValidateSource(root)
 if err!=nil||"v"+v!=tag{os.RemoveAll(root);return "",fmt.Errorf("paquete distinto al tag %s: %v",tag,err)}
 return root,nil
}

func shell() (string,error){
 if runtime.GOOS!="windows"{return "",errors.New("instalación disponible en Windows; otros sistemas todavía no validados")}
 for _,name:=range []string{"pwsh","powershell.exe"}{
  if path,err:=exec.LookPath(name);err==nil{return path,nil}
 }
 return "",errors.New("se requiere PowerShell")
}

func trunc(b []byte) string{
 if len(b)>65536{return "[salida truncada]\n"+string(b[len(b)-65536:])}
 return string(b)
}

func (m Manager) Script(ctx context.Context,root,file string,apply bool)(string,error){
 if file!="install.ps1"&&file!="rollback.ps1"&&file!="verify.ps1"{return "",errors.New("script no permitido")}
 psh,err:=shell()
 if err!=nil{return "",err}
 if _,err=ValidateSource(root);err!=nil{return "",err}
 args:=[]string{"-NoProfile","-NonInteractive","-File",filepath.Join(root,"scripts",file),"-CodexHome",m.Home}
 if apply{
  if file=="verify.ps1"{return "",errors.New("verificador de solo lectura")}
  args=append(args,"-Apply")
 }
 cmd:=exec.CommandContext(ctx,psh,args...)
 output,err:=cmd.CombinedOutput()
 if err!=nil{return trunc(output),fmt.Errorf("%s falló: %w",file,err)}
 return trunc(output),nil
}

func (m Manager) Prepare(ctx context.Context,action string)(Prepared,error){
 var root string
 var temporary bool
 var v string
 var err error
 switch action{
 case "local":
  root,err=m.Local()
  if err!=nil{return Prepared{},err}
  v,_=ValidateSource(root)
  cmp,_:=Compare(v,"v0.1.2")
  if cmp<0{return Prepared{},errors.New("checkout local anterior a v0.1.2; instalador rechazado por riesgos conocidos")}
 case "latest":
  rel,e:=m.Latest(ctx)
  if e!=nil{return Prepared{},e}
  installed,e:=m.Installed()
  if e!=nil{return Prepared{},e}
  if installed!=""{
   cmp,e:=Compare(rel.Tag,installed)
   if e!=nil{return Prepared{},e}
   if cmp<=0{return Prepared{},fmt.Errorf("sin actualización: instalada %s; última release %s",installed,rel.Tag)}
  }
  cmp,_:=Compare(rel.Tag,"v0.1.2")
  if cmp<0{return Prepared{},errors.New("la última release es anterior a v0.1.2 y su instalador presenta riesgos conocidos")}
  root,err=m.Download(ctx,rel.Tag)
  if err!=nil{return Prepared{},err}
  temporary=true
  v=strings.TrimPrefix(rel.Tag,"v")
 case "rollback":
  installed,e:=m.Installed()
  if e!=nil{return Prepared{},e}
  if installed==""{return Prepared{},errors.New("no existe instalación registrada")}
  cmp,_:=Compare(installed,"v0.1.2")
  if cmp<0{return Prepared{},errors.New("rollback de v0.1.1 bloqueado por fallo conocido; recuperar manualmente")}
  root,err=m.Local()
  if err!=nil{return Prepared{},fmt.Errorf("rollback requiere checkout local con scripts: %w",err)}
  v,_=ValidateSource(root)
 default:
  return Prepared{},errors.New("acción no soportada")
 }
 script:="install.ps1"
 operation:="install"
 if action=="rollback"{script="rollback.ps1";operation="rollback"}
 preview,err:=m.Script(ctx,root,script,false)
 if err!=nil{
  if temporary{os.RemoveAll(root)}
  return Prepared{},fmt.Errorf("preflight rechazado: %w\n%s",err,preview)
 }
 return Prepared{root,temporary,v,operation,preview},nil
}

func (m Manager) Apply(ctx context.Context,p Prepared)(string,error){
 file:="install.ps1"
 if p.Action=="rollback"{file="rollback.ps1"}
 log,err:=m.Script(ctx,p.Root,file,true)
 if err!=nil{return log,err}
 if p.Action=="install"{
  out,e:=m.Script(ctx,p.Root,"verify.ps1",false)
  log+="\n--- Verificación ---\n"+out
  if e!=nil{return log,fmt.Errorf("instalado pero verificación fallida: %w",e)}
 }
 return log,nil
}

func (p Prepared) Cleanup(){
 if p.Temporary&&p.Root!=""{_ = os.RemoveAll(p.Root)}
}

package providerauth
import("testing";"context";"errors";"time";"strings";"crypto/sha256";"encoding/hex";"net/http")
type r4DownVault struct{}
func(r4DownVault) Resolve(context.Context,string,string)(Secret,error){return Secret{},errors.New("synthetic vault unavailable")}
func TestR4WarmTokenSurvivesVaultOutage(t *testing.T){
 cfg:=Config{Environment:"local",TenantID:"t",BindingID:"b",SecretVersion:"v1",TokenURL:"https://fixture.invalid/token",ClientID:"client",ClientSecretRef:"fixture"}
 sum:=sha256.Sum256([]byte(strings.Join([]string{cfg.Environment,cfg.TenantID,cfg.BindingID,"account",cfg.TokenURL,cfg.ClientID,cfg.SecretVersion},"\x1f")))
 key:="hub:provider-token:"+hex.EncodeToString(sum[:])
 c:=&TokenCache{Resolver:r4DownVault{},tokens:map[string]cachedToken{key:{"synthetic-valid",time.Now().Add(time.Minute)}}}
 got,err:=c.bearer(context.Background(),&http.Client{},"account",cfg)
 if err!=nil||got!="synthetic-valid"{t.Fatalf("L1 válido não foi utilizado: err=%v",err)}
}

package orbita
import (
 "encoding/json"
 "testing"
 "ai-hub/hub/internal/atlas"
 "ai-hub/hub/internal/dispatch"
)
func TestR6ProductMustRejectOrFullyResolveDifferentProvider(t *testing.T) {
 data,_:=json.Marshal(atlas.CatalogData{AdapterID:"rest-json-v1",InputSchema:json.RawMessage(`{"type":"object"}`),Routes:[]atlas.Route{{ProviderAccountID:"account-B",BindingID:"binding-B"}}})
 product,_:=json.Marshal(atlas.CatalogData{MaxParallel:1,Steps:[]atlas.Step{{ID:"s",ServiceID:"service-B",ServiceVersion:1}}})
 snapshot:=atlas.OfferSnapshot{Target:atlas.Resource{Kind:"products",ID:"p",Version:1,Data:product},Account:atlas.Resource{ID:"account-A"},Binding:atlas.Resource{ID:"binding-A"},SelectedRoute:atlas.Route{ProviderAccountID:"account-A",BindingID:"binding-A"},Services:[]atlas.Resource{{Kind:"services",ID:"service-B",Version:1,Data:data}}}
 _,commands,err:=BuildProductPlan(snapshot,dispatch.Command{ProviderAccountID:"account-A"},json.RawMessage(`{}`))
 if err!=nil {return} // Refusing an unresolvable route is safe.
 var child atlas.OfferSnapshot
 if err=json.Unmarshal(commands[0].ConfigSnapshot,&child);err!=nil{t.Fatal(err)}
 if child.SelectedRoute.ProviderAccountID!=child.Account.ID || child.SelectedRoute.BindingID!=child.Binding.ID || commands[0].ProviderAccountID!=child.SelectedRoute.ProviderAccountID {
 t.Fatalf("accepted incoherent step: route=%s account=%s binding=%s/%s commandAccount=%s",child.SelectedRoute.ProviderAccountID,child.Account.ID,child.SelectedRoute.BindingID,child.Binding.ID,commands[0].ProviderAccountID)
 }
}

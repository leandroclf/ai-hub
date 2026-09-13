package atlasclient
import("context";"net/http";"net/http/httptest";"testing";"time";"ai-hub/hub/internal/atlas")
func TestR5OfferDenialMustNotUseCache(t *testing.T){
 s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.WriteHeader(403);w.Write([]byte(`{"error":"offer_not_eligible"}`))}));defer s.Close()
 c:=New(s.URL,time.Minute);c.http=s.Client();c.setCached("offer:t:a:svc:1:p",atlas.OfferSnapshot{Hash:"previously-authorized",ValidUntil:time.Now().Add(time.Minute)})
 _,err:=c.Offer(context.Background(),"t","a","svc",1,"p");if err==nil{t.Fatal("negação explícita 403 foi substituída por oferta em cache")}
}

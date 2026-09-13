package orbita
import("encoding/json";"testing")
func TestR5OperationFactPreservesInteger(t *testing.T){
 var f operationFact
 if err:=json.Unmarshal([]byte(`{"response_body":{"id":9007199254740993}}`),&f);err!=nil{t.Fatal(err)}
 out,err:=json.Marshal(f.ResponseBody);if err!=nil{t.Fatal(err)}
 if string(out)!=`{"id":9007199254740993}`{t.Fatalf("fato recebido perdeu precisão: %s",out)}
}

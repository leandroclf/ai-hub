package atlas
import("testing";"encoding/json";"bytes")
func TestR3PreservesLargeInteger(t *testing.T){
 input:=json.RawMessage(`{"id":9007199254740993}`)
 output,err:=TransformJSON(input,nil,json.RawMessage(`{"type":"object","properties":{"id":{"type":"integer"}}}`))
 if err!=nil{t.Fatal(err)}
 if !bytes.Equal(input,output){t.Fatalf("contrato alterado: input=%s output=%s",input,output)}
}
func TestR3RejectsEnumViolation(t *testing.T){
 _,err:=TransformJSON(json.RawMessage(`{"status":"INVALID"}`),nil,json.RawMessage(`{"type":"object","properties":{"status":{"type":"string","enum":["OK"]}}}`))
 if err==nil{t.Fatal("schema enum ignorado: INVALID foi aceito")}
}

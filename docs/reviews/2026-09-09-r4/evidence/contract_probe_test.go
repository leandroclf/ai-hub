package atlas
import("testing";"encoding/json")
func TestR4ContractAdversarial(t *testing.T){
 cases:=[]struct{name,input,schema string}{
 {"null_as_string", "{\"v\":null}", "{\"type\":\"object\",\"properties\":{\"v\":{\"type\":\"string\"}}}"},
 {"quoted_integer","{\"v\":\"123\"}","{\"type\":\"object\",\"properties\":{\"v\":{\"type\":\"integer\"}}}"},
 {"root_null","null","{\"type\":\"object\",\"properties\":{}}"},
 {"trailing_document","{\"v\":1} {\"second\":true}","{\"type\":\"object\",\"properties\":{\"v\":{\"type\":\"integer\"}}}"},
 {"ignored_minimum","{\"v\":-1}","{\"type\":\"object\",\"properties\":{\"v\":{\"type\":\"integer\",\"minimum\":0}}}"},
 }
 for _,c:=range cases{t.Run(c.name,func(t *testing.T){out,err:=TransformJSON(json.RawMessage(c.input),nil,json.RawMessage(c.schema));if err==nil{t.Fatalf("input ou regra deveria ser recusado: output=%s",out)}})}
}
func TestR4PreviousFixes(t *testing.T){
 input:=json.RawMessage("{\"id\":9007199254740993}")
 out,err:=TransformJSON(input,nil,json.RawMessage("{\"type\":\"object\",\"properties\":{\"id\":{\"type\":\"integer\"}}}"))
 if err!=nil||string(out)!=string(input){t.Fatalf("precisão: out=%s err=%v",out,err)}
 _,err=TransformJSON(json.RawMessage("{\"s\":\"BAD\"}"),nil,json.RawMessage("{\"type\":\"object\",\"properties\":{\"s\":{\"type\":\"string\",\"enum\":[\"OK\"]}}}"))
 if err==nil{t.Fatal("enum violado")}
}

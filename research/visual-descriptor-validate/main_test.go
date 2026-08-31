package main
import("path/filepath";"testing")
func TestRepositoryBundle(t *testing.T){root:=filepath.Join("..","..");if err:=validateBundle(filepath.Join(root,"research","visual_descriptors"));err!=nil{t.Fatal(err)}}
func TestForbiddenColumns(t *testing.T){if forbiddenColumns([]string{"page_id","token_count"})==nil{t.Fatal("expected rejection")}}
func TestLeafMapping(t *testing.T){for in,want:=range map[string]string{"f67r1":"f67","f116v":"f116","fRos":"f85-f86-foldout"}{if got:=wantLeaf(in);got!=want{t.Fatalf("%s: %s",in,got)}}}

package backup

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"os"
	"testing"
)

func TestGetIdHash(t *testing.T) {
	cStr := "baidvkeravnxpwlrx2n7u6kjsqmbnchi7ka3jtbvzh3vvmt66vqenmupfqidq"
	binary, err := uuid.New().MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	u := base64.StdEncoding.EncodeToString(binary)
	fmt.Println(u)
	hash1, err := GetIdHash(cStr, u)
	hash2, err := GetIdHash(cStr, u)
	if err != nil {
		t.Fatal(err)
	}
	if hash2 != hash1 {
		t.Fatal("idHash 不一致")
	}
	fmt.Println(hash1)
}

func TestProtoc(t *testing.T) {
	info := Info{
		IdHashPin:      map[string]bool{"1": true},
		IdHashUnpin:    map[string]string{"2": "111"},
		TargetPeerList: map[string]struct{}{"p1": {}, "p2": {}, "p3": {}, "p4": {}},
	}
	marshal, err := Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(marshal))
	bytes, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(bytes))
	res := Info{}
	err = Unmarshal(marshal, &res)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", res)
}

func TestQ(t *testing.T) {
	getenv := os.Getenv("GOLOG_FILE")
	fmt.Println(getenv)

}

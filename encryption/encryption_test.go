package encryption

import (
	"log"
	"testing"
)

func TestEncryption(t *testing.T){
	log.Println(ProtectPANandCVV("4111111111111111","997",1, pieStruct{
		L:      6,
		E:      4,
		K:      "50B46C729E19D39888B14B1E4623C381",
		Key_id: "51408751",
		Phase:  0,
	}))
}

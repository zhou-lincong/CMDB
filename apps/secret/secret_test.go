package secret_test

import (
	"testing"

	"github.com/zhou-lincong/CMDB/apps/secret"
)

var (
	encryptKey = "sdfsdfsfdsfd"
)

func TestSecretEncrypt(t *testing.T) {
	ins := secret.NewDefaultSecret()
	ins.Data.ApiSecret = "123456"
	ins.Data.EncryptAPISecret(encryptKey)
	t.Log(ins.Data.ApiSecret)

	ins.Data.DecryptAPISecret(encryptKey)
	t.Log(ins.Data.ApiSecret)

	// === RUN   TestSecretEncrypt
	//     e:\goproject\CMDB\apps\secret\secret_test.go:17: @ciphered@iepZcOL6WoBJyOToPoWoRCfpZioImUmw5ZfNtP8ZstQ=
	//     e:\goproject\CMDB\apps\secret\secret_test.go:20: 123456
	// --- PASS: TestSecretEncrypt (0.00s)
	// PASS
	// ok  	github.com/zhou-lincong/CMDB/apps/secret	1.114s
}

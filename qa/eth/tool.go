package eth

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"

	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

// E18 ethereum 1后面18个0
const E18 = 1000000000000000000

// AddrInfo 私钥、公钥、地址
type AddrInfo struct {
	PrivkHex, PubkHex, Address string
}

// ToECDSAKey .
func (ad *AddrInfo) ToECDSAKey() *ecdsa.PrivateKey {
	k, err := crypto.HexToECDSA(ad.PrivkHex)
	if err != nil {
		panic(err)
	}
	return k
}

// ToAddress .
func (ad *AddrInfo) ToAddress() common.Address {
	return common.HexToAddress(ad.Address)
}

// GenAddr 生成地址
func GenAddr() *AddrInfo {
	key, _ := crypto.GenerateKey()
	pubKHex := hexutil.Encode(crypto.FromECDSAPub(&key.PublicKey))
	address := crypto.PubkeyToAddress(key.PublicKey).Hex()
	return &AddrInfo{
		PrivkHex: hexutil.Encode(crypto.FromECDSA(key))[2:],
		PubkHex:  pubKHex[2:],
		Address:  address,
	}
}

// PrepareFunds4address 为addr准备一定量的eth,
func PrepareFunds4address(t *testing.T, rpcHost, addr string, funds int64) {
	rq := require.New(t)
	rpcClient, err := rpc.DialContext(context.Background(), rpcHost)
	rq.Nil(err, "Failed to dial rpc")
	defer rpcClient.Close()

	client := ethclient.NewClient(rpcClient)

	var hexedAccounts []string
	err = rpcClient.Call(&hexedAccounts, "eth_accounts")
	// fmt.Println(hexedAccounts)
	rq.Nil(err, "Fail on get accounts")

	fromAccount := ""
	for _, acc := range hexedAccounts {
		bal, err := client.BalanceAt(context.Background(), common.HexToAddress(acc), nil)
		rq.Nil(err, "Failed to get balance of account")
		if bal.Cmp(big.NewInt(E18*funds+E18)) > 0 {
			fromAccount = acc
		}
	}

	if fromAccount == "" {
		t.Fatal("余额不足，无法充值")
	}

	tx := map[string]interface{}{
		"from": fromAccount,
		"to":   addr,
		// "gas": "0x76c0", // 30400
		"gasPrice": "0x9184e72a000", // 10000000000000
		"value":    E18 * funds,
	}
	var txHash string
	err = rpcClient.Call(&txHash, "eth_sendTransaction", tx)
	rq.Nil(err, "转账失败")
	fmt.Println("PrepareFunds4addr done", addr, funds)
}

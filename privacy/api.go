package privacy

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type PrivacyAPI struct {
	wrapper  *PrivacyWrapper
	backends []backend
	pubKeys  []*ecdsa.PublicKey
}

type backend struct {
	Name   string `json:"name"`
	PubKey string `json:"pubKey"`
	Mask   uint64 `json:"mask"`
}

func NewPrivacyAPI(wrapper *PrivacyWrapper) *PrivacyAPI {
	var backends []backend
	_ = json.Unmarshal([]byte(wrapper.config.Backends), &backends)
	pubKeys := make([]*ecdsa.PublicKey, len(backends))

	for i, b := range backends {
		d, _ := hexutil.Decode(b.PubKey)
		pubKeys[i], _ = crypto.UnmarshalPubkey(d)
	}
	return &PrivacyAPI{
		wrapper:  wrapper,
		backends: backends,
		pubKeys:  pubKeys,
	}
}

func (p *PrivacyAPI) SetToken(ctx context.Context, token string, addresses []string, sig string) (interface{}, error) {
	if !p.healthCheck(ctx) {
		return nil, NewApiServiceError("PrivacyAPI service is not ok")
	}
	if token == "" || len(addresses) == 0 {
		return nil, NewApiServiceError("token or address is empty")
	}

	// check signature
	var addressesBytes = make([]byte, len(addresses)*common.AddressLength)
	for i, addr := range addresses {
		//fmt.Println("addr", common.HexToAddress(addr).Bytes())
		copy(addressesBytes[i*common.AddressLength:], common.HexToAddress(addr).Bytes())
	}

	hash := crypto.Keccak256([]byte(token), addressesBytes)
	if !p.validSignature(ctx, hash, sig) {
		return nil, NewSignatureVerificationFailedError("signature is not valid")
	}
	//valid := p.checkSignature(ctx, hash, sig)
	return p.set(ctx, token, addresses)
}

func (p *PrivacyAPI) UpdateToken(ctx context.Context, token string, addresses []string, sig string) (interface{}, error) {
	if !p.healthCheck(ctx) {
		return nil, NewApiServiceError("PrivacyAPI service is not ok")
	}
	if token == "" || len(addresses) == 0 {
		return nil, NewApiServiceError("token or address is empty")
	}
	return p.set(ctx, token, addresses)
}

func (p *PrivacyAPI) set(ctx context.Context, token string, addresses []string) (interface{}, error) {
	for _, addr := range addresses {
		if addr == "" {
			return nil, NewSetTokenFailedError("address is empty")
		}
		addr = common.HexToAddress(addr).String()
		if err := p.wrapper.cache.Set(ctx, addr, []byte(token), uint64(0)); err != nil {
			return nil, NewSetTokenFailedError("PrivacyAPI: set token failed")
		}
	}
	return "Set token successfully", nil
}

func (p *PrivacyAPI) GetToken(ctx context.Context, token string, addresses []string) (interface{}, error) {
	//if !p.healthCheck() {
	//	return nil, NewApiServiceError("PrivacyAPI service is not ok")
	//}
	//res, err := p.wrapper.cache.Get(ctx, address)
	//if err != nil {
	//	return nil, NewGetTokenFailedError("PrivacyAPI: get token failed")
	//}
	//return string(res), nil
	return nil, NewGetTokenFailedError("PrivacyAPI: method not allowed")
}

func (p *PrivacyAPI) healthCheck(ctx context.Context) bool {
	//return false
	return p.wrapper.cache.HealthCheck(ctx)
}

func (p *PrivacyAPI) validSignature(ctx context.Context, hash []byte, sig string) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		sigBytes, err := hexutil.Decode(sig)
		pub, err := crypto.Ecrecover(hash, sigBytes)
		pubkey, err := crypto.UnmarshalPubkey(pub)
		if err != nil {
			return false
		}
		for _, key := range p.pubKeys {
			if key.Equal(pubkey) {
				return true
			}
		}
		return false
	}
}

package infra

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"math/big"

	"github.com/Hyperledger-TWGC/tape/internal/fabric/bccsp/utils"
	"github.com/Hyperledger-TWGC/tape/internal/fabric/common/crypto"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"
)

type CryptoConfig struct {
	MSPID      string
	PrivKey    string
	SignCert   string
	TLSCACerts []string
}

type ECDSASignature struct {
	R, S *big.Int
}

type Crypto struct {
	Creator  []byte
	PrivKey  *ecdsa.PrivateKey
	SignCert *x509.Certificate
}

func (s *Crypto) Sign(message []byte) ([]byte, error) {
	ri, si, err := ecdsa.Sign(rand.Reader, s.PrivKey, digest(message))
	if err != nil {
		return nil, err
	}

	si, _, err = utils.ToLowS(&s.PrivKey.PublicKey, si)
	if err != nil {
		return nil, err
	}

	return asn1.Marshal(ECDSASignature{ri, si})
}

func (s *Crypto) Serialize() ([]byte, error) {
	return s.Creator, nil
}

func (s *Crypto) NewSignatureHeader() (*common.SignatureHeader, error) {
	creator, err := s.Serialize()
	if err != nil {
		return nil, err
	}
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, err
	}

	return &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}, nil
}

func digest(in []byte) []byte {
	h := sha256.New()
	h.Write(in)
	return h.Sum(nil)
}

func toPEM(in []byte) ([]byte, error) {
	d := make([]byte, base64.StdEncoding.DecodedLen(len(in)))
	n, err := base64.StdEncoding.Decode(d, in)
	if err != nil {
		return nil, err
	}
	return d[:n], nil
}

func GetPrivateKey(f string) (*ecdsa.PrivateKey, error) {
	in, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	k, err := utils.PEMtoPrivateKey(in, []byte{})
	if err != nil {
		return nil, err
	}

	key, ok := k.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.Errorf("expecting ecdsa key")
	}

	return key, nil
}

func GetCertificate(f string) (*x509.Certificate, []byte, error) {
	in, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, nil, err
	}

	block, _ := pem.Decode(in)

	c, err := x509.ParseCertificate(block.Bytes)
	return c, in, err
}

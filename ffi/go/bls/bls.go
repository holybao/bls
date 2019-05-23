package bls

/*
#cgo bn256 CFLAGS:-DMCLBN_FP_UNIT_SIZE=4
#cgo bn256 LDFLAGS:-lbls256
#cgo bn384 CFLAGS:-DMCLBN_FP_UNIT_SIZE=6
#cgo bn384 LDFLAGS:-lbls384
#cgo bn384_256 CFLAGS:-DMCLBN_FP_UNIT_SIZE=6 -DMCLBN_FR_UNIT_SIZE=4
#cgo bn384_256 LDFLAGS:-lbls384_256
#cgo LDFLAGS:-lcrypto -lgmp -lgmpxx -lstdc++
typedef unsigned int (*ReadRandFunc)(void *, void *, unsigned int);
int wrapReadRandCgo(void *self, void *buf, unsigned int n);
#include <bls/bls.h>
*/
import "C"
import "fmt"
import "unsafe"
import "io"
import "encoding/hex"

func hex2byte(s string) ([]byte, error) {
	if (len(s) & 1) == 1 {
		return nil, fmt.Errorf("odd length")
	}
	return hex.DecodeString(s)
}

// Init --
// call this function before calling all the other operations
// this function is not thread safe
func Init(curve int) error {
	err := C.blsInit(C.int(curve), C.MCLBN_COMPILED_TIME_VAR)
	if err != 0 {
		return fmt.Errorf("ERR Init curve=%d", curve)
	}
	return nil
}

// ID --
type ID struct {
	v Fr
}

// getPointer --
func (id *ID) getPointer() (p *C.blsId) {
	// #nosec
	return (*C.blsId)(unsafe.Pointer(id))
}

// Serialize --
func (id *ID) Serialize() []byte {
	return id.v.Serialize()
}

// Deserialize --
func (id *ID) Deserialize(buf []byte) error {
	return id.v.Deserialize(buf)
}

// GetLittleEndian -- alias of Serialize
func (id *ID) GetLittleEndian() []byte {
	return id.Serialize()
}

// SetLittleEndian --
func (id *ID) SetLittleEndian(buf []byte) error {
	return id.v.SetLittleEndian(buf)
}

// GetHexString --
func (id *ID) GetHexString() string {
	return id.v.GetString(16)
}

// GetDecString --
func (id *ID) GetDecString() string {
	return id.v.GetString(10)
}

// SetHexString --
func (id *ID) SetHexString(s string) error {
	return id.v.SetString(s, 16)
}

// SetDecString --
func (id *ID) SetDecString(s string) error {
	return id.v.SetString(s, 10)
}

// IsEqual --
func (id *ID) IsEqual(rhs *ID) bool {
	if id == nil || rhs == nil {
		return false
	}
	return id.v.IsEqual(&rhs.v)
}

// SecretKey --
type SecretKey struct {
	v Fr
}

// getPointer --
func (sec *SecretKey) getPointer() (p *C.blsSecretKey) {
	// #nosec
	return (*C.blsSecretKey)(unsafe.Pointer(sec))
}

// Serialize --
func (sec *SecretKey) Serialize() []byte {
	return sec.v.Serialize()
}

// Deserialize --
func (sec *SecretKey) Deserialize(buf []byte) error {
	return sec.v.Deserialize(buf)
}

// GetLittleEndian -- alias of Serialize
func (sec *SecretKey) GetLittleEndian() []byte {
	return sec.Serialize()
}

// SetLittleEndian --
func (sec *SecretKey) SetLittleEndian(buf []byte) error {
	return sec.v.SetLittleEndian(buf)
}

// SerializeToHexStr --
func (sec *SecretKey) SerializeToHexStr() string {
	return hex.EncodeToString(sec.Serialize())
}

// DeserializeHexStr --
func (sec *SecretKey) DeserializeHexStr(s string) error {
	a, err := hex2byte(s)
	if err != nil {
		return err
	}
	return sec.Deserialize(a)
}

// GetHexString --
func (sec *SecretKey) GetHexString() string {
	return sec.v.GetString(16)
}

// GetDecString --
func (sec *SecretKey) GetDecString() string {
	return sec.v.GetString(10)
}

// SetHexString --
func (sec *SecretKey) SetHexString(s string) error {
	return sec.v.SetString(s, 16)
}

// SetDecString --
func (sec *SecretKey) SetDecString(s string) error {
	return sec.v.SetString(s, 10)
}

// IsEqual --
func (sec *SecretKey) IsEqual(rhs *SecretKey) bool {
	if sec == nil || rhs == nil {
		return false
	}
	return sec.v.IsEqual(&rhs.v)
}

// SetByCSPRNG --
func (sec *SecretKey) SetByCSPRNG() {
	sec.v.SetByCSPRNG()
}

// Add --
func (sec *SecretKey) Add(rhs *SecretKey) {
	FrAdd(&sec.v, &sec.v, &rhs.v)
}

// GetMasterSecretKey --
func (sec *SecretKey) GetMasterSecretKey(k int) (msk []SecretKey) {
	msk = make([]SecretKey, k)
	msk[0] = *sec
	for i := 1; i < k; i++ {
		msk[i].SetByCSPRNG()
	}
	return msk
}

// GetMasterPublicKey --
func GetMasterPublicKey(msk []SecretKey) (mpk []PublicKey) {
	n := len(msk)
	mpk = make([]PublicKey, n)
	for i := 0; i < n; i++ {
		mpk[i] = *msk[i].GetPublicKey()
	}
	return mpk
}

// Set --
func (sec *SecretKey) Set(msk []SecretKey, id *ID) error {
	// #nosec
	return FrEvaluatePolynomial(&sec.v, *(*[]Fr)(unsafe.Pointer(&msk)), &id.v)
}

// Recover --
func (sec *SecretKey) Recover(secVec []SecretKey, idVec []ID) error {
	// #nosec
	return FrLagrangeInterpolation(&sec.v, *(*[]Fr)(unsafe.Pointer(&idVec)), *(*[]Fr)(unsafe.Pointer(&secVec)))
}

// GetPop --
func (sec *SecretKey) GetPop() (sign *Sign) {
	sign = new(Sign)
	C.blsGetPop(sign.getPointer(), sec.getPointer())
	return sign
}

// PublicKey --
type PublicKey struct {
	v C.blsPublicKey
}

// Serialize --
func (pub *PublicKey) Serialize() []byte {
	buf := make([]byte, 2048)
	// #nosec
	n := C.blsPublicKeySerialize(unsafe.Pointer(&buf[0]), C.mclSize(len(buf)), &pub.v)
	if n == 0 {
		panic("err blsPublicKeySerialize")
	}
	return buf[:n]
}

// Deserialize --
func (pub *PublicKey) Deserialize(buf []byte) error {
	// #nosec
	err := C.blsPublicKeyDeserialize(&pub.v, unsafe.Pointer(&buf[0]), C.mclSize(len(buf)))
	if err == 0 {
		return fmt.Errorf("err blsPublicKeyDeserialize %x", buf)
	}
	return nil
}

// SerializeToHexStr --
func (pub *PublicKey) SerializeToHexStr() string {
	return hex.EncodeToString(pub.Serialize())
}

// DeserializeHexStr --
func (pub *PublicKey) DeserializeHexStr(s string) error {
	a, err := hex2byte(s)
	if err != nil {
		return err
	}
	return pub.Deserialize(a)
}

// GetHexString -- alias of SerializeToHexStr
func (pub *PublicKey) GetHexString() string {
	return pub.SerializeToHexStr()
}

// SetHexString -- alias of DeserializeHexStr
func (pub *PublicKey) SetHexString(s string) error {
	return pub.DeserializeHexStr(s)
}

// IsEqual --
func (pub *PublicKey) IsEqual(rhs *PublicKey) bool {
	if pub == nil || rhs == nil {
		return false
	}
	return C.blsPublicKeyIsEqual(&pub.v, &rhs.v) == 1
}

// Add --
func (pub *PublicKey) Add(rhs *PublicKey) {
	C.blsPublicKeyAdd(&pub.v, &rhs.v)
}

// Set --
func (pub *PublicKey) Set(mpk []PublicKey, id *ID) error {
	// #nosec
	ret := C.blsPublicKeyShare(&pub.v, &mpk[0].v, (C.mclSize)(len(mpk)), (*C.blsId)(&id.v))
	if ret != 0 {
		return fmt.Errorf("err blsPublicKeyShare")
	}
	return nil
}

// Recover --
func (pub *PublicKey) Recover(pubVec []PublicKey, idVec []ID) error {
	if len(pubVec) != len(idVec) {
		return fmt.Errorf("err PublicKey.Recover bad size")
	}
	// #nosec
	ret := C.blsPublicKeyRecover(&pub.v, &pubVec[0].v, (*C.blsId)(&idVec[0].v), (C.mclSize)(len(pubVec)))
	if ret != 0 {
		return fmt.Errorf("err blsPublicKeyShare")
	}
	return nil
}

// Sign  --
type Sign struct {
	v G1
}

// getPointer --
func (sign *Sign) getPointer() (p *C.blsSignature) {
	// #nosec
	return (*C.blsSignature)(unsafe.Pointer(sign))
}

// Serialize --
func (sign *Sign) Serialize() []byte {
	return sign.v.Serialize()
}

// Deserialize --
func (sign *Sign) Deserialize(buf []byte) error {
	return sign.v.Deserialize(buf)
}

// SerializeToHexStr --
func (sign *Sign) SerializeToHexStr() string {
	return sign.v.GetString(IoSerializeHexStr)
}

// DeserializeHexStr --
func (sign *Sign) DeserializeHexStr(s string) error {
	return sign.v.SetString(s, IoSerializeHexStr)
}

// GetHexString --
func (sign *Sign) GetHexString() string {
	return sign.v.GetString(16)
}

// SetHexString --
func (sign *Sign) SetHexString(s string) error {
	return sign.v.SetString(s, 16)
}

// IsEqual --
func (sign *Sign) IsEqual(rhs *Sign) bool {
	if sign == nil || rhs == nil {
		return false
	}
	return sign.v.IsEqual(&rhs.v)
}

// GetPublicKey --
func (sec *SecretKey) GetPublicKey() (pub *PublicKey) {
	pub = new(PublicKey)
	C.blsGetPublicKey(&pub.v, sec.getPointer())
	return pub
}

// Sign -- Constant Time version
func (sec *SecretKey) Sign(m string) (sign *Sign) {
	sign = new(Sign)
	buf := []byte(m)
	// #nosec
	C.blsSign(sign.getPointer(), sec.getPointer(), unsafe.Pointer(&buf[0]), C.mclSize(len(buf)))
	return sign
}

// Add --
func (sign *Sign) Add(rhs *Sign) {
	C.blsSignatureAdd(sign.getPointer(), rhs.getPointer())
}

// Recover --
func (sign *Sign) Recover(signVec []Sign, idVec []ID) error {
	// #nosec
	return G1LagrangeInterpolation(&sign.v, *(*[]Fr)(unsafe.Pointer(&idVec)), *(*[]G1)(unsafe.Pointer(&signVec)))
}

// Verify --
func (sign *Sign) Verify(pub *PublicKey, m string) bool {
	buf := []byte(m)
	// #nosec
	return C.blsVerify(sign.getPointer(), &pub.v, unsafe.Pointer(&buf[0]), C.mclSize(len(buf))) == 1
}

// VerifyPop --
func (sign *Sign) VerifyPop(pub *PublicKey) bool {
	if pub == nil {
		return false
	}
	return C.blsVerifyPop(sign.getPointer(), &pub.v) == 1
}

// DHKeyExchange --
func DHKeyExchange(sec *SecretKey, pub *PublicKey) (out PublicKey) {
	C.blsDHKeyExchange(&out.v, sec.getPointer(), &pub.v)
	return out
}

// HashAndMapToSignature --
func HashAndMapToSignature(buf []byte) *Sign {
	sig := new(Sign)
	err := sig.v.HashAndMapTo(buf)
	if err == nil {
		return sig
	} else {
		return nil
	}
}

// VerifyPairing --
func VerifyPairing(X *Sign, Y *Sign, pub *PublicKey) bool {
	if X.getPointer() == nil || Y.getPointer() == nil || pub == nil {
		return false
	}
	return C.blsVerifyPairing(X.getPointer(), Y.getPointer(), &pub.v) == 1
}

// SignHash --
func (sec *SecretKey) SignHash(hash []byte) (sign *Sign) {
	sign = new(Sign)
	// #nosec
	err := C.blsSignHash(sign.getPointer(), sec.getPointer(), unsafe.Pointer(&hash[0]), C.mclSize(len(hash)))
	if err == 0 {
		return sign
	} else {
		return nil
	}
}

// VerifyHash --
func (sign *Sign) VerifyHash(pub *PublicKey, hash []byte) bool {
	if pub == nil {
		return false
	}
	// #nosec
	return C.blsVerifyHash(sign.getPointer(), &pub.v, unsafe.Pointer(&hash[0]), C.mclSize(len(hash))) == 1
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// VerifyAggregateHashes --
func (sign *Sign) VerifyAggregateHashes(pubVec []PublicKey, hash [][]byte) bool {
	if pubVec == nil {
		return false
	}
	hashByte := GetOpUnitSize() * 8
	n := len(hash)
	h := make([]byte, n*hashByte)
	for i := 0; i < n; i++ {
		hn := len(hash[i])
		copy(h[i*hashByte:(i+1)*hashByte], hash[i][0:Min(hn, hashByte)])
	}
	return C.blsVerifyAggregatedHashes(sign.getPointer(), &pubVec[0].v, unsafe.Pointer(&h[0]), C.mclSize(hashByte), C.mclSize(n)) == 1
}

///

var s_randReader io.Reader

func createSlice(buf *C.char, n C.uint) []byte {
	size := int(n)
	return (*[1 << 30]byte)(unsafe.Pointer(buf))[:size:size]
}

// this function can't be put in callback.go
//export wrapReadRandGo
func wrapReadRandGo(buf *C.char, n C.uint) C.uint {
	slice := createSlice(buf, n)
	ret, err := s_randReader.Read(slice)
	if ret == int(n) && err == nil {
		return n
	}
	return 0
}

// SetRandFunc --
func SetRandFunc(randReader io.Reader) {
	s_randReader = randReader
	if randReader != nil {
		C.blsSetRandFunc(nil, C.ReadRandFunc(unsafe.Pointer(C.wrapReadRandCgo)))
	} else {
		// use default random generator
		C.blsSetRandFunc(nil, C.ReadRandFunc(unsafe.Pointer(nil)))
	}
}

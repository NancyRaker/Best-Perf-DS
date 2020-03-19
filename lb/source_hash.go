package lb

import (
	"crypto/md5"
	"hash"
	"log"
)

type sourceHash struct {
	h   hash.Hash
	ips []string
}

func newSourceHash(ips []string) *sourceHash {
	return &sourceHash{ips: ips, h:md5.New()}
}

func (sh *sourceHash) GetServer(clientIp ...string) string {
	hValue := sh.getHashDigest(clientIp[0])
	hCode := sh.getHashCode(hValue)
	return sh.ips[hCode%uint32(len(sh.ips))]
}

func (sh *sourceHash)getHashDigest(ip string) []byte {
	sh.h.Write([]byte(ip))
	value := sh.h.Sum(nil)
	sh.h.Reset()
	return value
}

func (sh *sourceHash)getHashCode(hb []byte) uint32 {
	if len(hb) < 4 {
		log.Fatalf("hash key %+v is invalid", hb)
		return 0
	}

	return (uint32(hb[3]) << 24) | (uint32(hb[2]) << 16) | (uint32(hb[1]) << 8) | uint32(hb[0])
}
package encoding

import "testing"

func TestGenerateSignature(t *testing.T) {
	s := GenerateSignature(`{"a":1}`)
	if s[:17] != "signed_body=SIGNA" {
		t.Fatalf("unexpected prefix %q", s[:20])
	}
}

func TestGenerateJazoest(t *testing.T) {
	if got := GenerateJazoest("abc"); got != "2294" {
		t.Fatalf("got %s", got)
	}
}

func TestInstagramIDCodec(t *testing.T) {
	n := int64(12345)
	enc := InstagramIDEncode(n)
	dec, err := InstagramIDDecode(enc)
	if err != nil || dec != n {
		t.Fatalf("roundtrip %d -> %s -> %d err %v", n, enc, dec, err)
	}
}

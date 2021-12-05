package ampcachepurge

import (
	"testing"
)

func TestPanicWhenPrivateKeyCantBeLoaded(t *testing.T) {
	location := ""
	password := ""
	shouldPanic(t, func() { loadPrivateKey(location, password) })
}

func TestUrlSignature(t *testing.T) {
	expected := "B1qX89cO_i4emMiC9T6ITiBwSNwJXL9y_86AoE0hbI4EwkAKW89TmwLqI0rvZG9hwAdYvYsfMw2vAg7ygrvWfHN18hrhWiZg_AL8NEV0Jk_IRrAYfah7s1_QDOLC5FLbDE-z9Lo-NnaEfYjlA-Cc7-jnFDa5GN6CSS_tcb-suBkmPKWjr_E9eSVfFcoNuEuVChEnlzDft1IMUJCZ3kFMSRtp4ZbGclJxqYJwVmvCkEY7HXptVi4kGGjl_qJtc3Qt3sDWmIzXiXwG5OA7U8xAcyF1b_XkoPpxp5D8ZT437OGswdPlJ6T7elx1rycccozpyZCS3fId6mNSAGd5SdNoUQ"
	path := "/update-cache/c/s/www.example.com/amp/test-amp-page?amp_action=flush&amp_ts=1637421562"
	result := signPath(path)
	assertEquals(t, result, expected)
}

func shouldPanic(t *testing.T, f func()) {
	defer func() { recover() }()
	f()
	t.Errorf("should have panicked")
}

func assertEquals(t *testing.T, actual, expected string) {
	if actual != expected {
		t.Fatalf("\nExpected result: %s\nActual   result: %s\n", expected, actual)
	}
}

package main

import "testing"

func TestConfig_Password(t *testing.T) {
	val := ""
	pwd := &CfgPwd{value: &val}
	err := pwd.Set("Password123")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("encrypted password:", val)
	t.Log("decrypted password:", pwd.Get())
}

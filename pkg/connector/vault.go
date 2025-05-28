package connector

import (
	"ResuMatch/internal/vault"
	l "ResuMatch/pkg/logger"
	"os"
)

func GetVaultClient() *vault.VaultClient {
	vaultCfg := &vault.VaultConfig{
		Scheme: os.Getenv("VAULT_CLIENT_SCHEME"),
		Host:   os.Getenv("VAULT_CLIENT_HOST"),
		Port:   os.Getenv("VAULT_CLIENT_PORT"),
		Token:  os.Getenv("VAULT_TOKEN"),
		Prefix: os.Getenv("VAULT_CLIENT_PREFIX"),
	}

	vaultClient, err := vault.NewVaultClient(vaultCfg)
	if err != nil {
		l.Log.Fatal(err.Error())
		return nil
	}

	if vaultClient == nil {
		l.Log.Fatal("vault client is nil")
		return nil
	}
	return vaultClient
}

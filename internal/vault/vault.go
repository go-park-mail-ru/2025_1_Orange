package vault

import (
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/vault/api"
	"net/http"
	"strings"
	"time"
)

type VaultConfig struct {
	Scheme string `yaml:"scheme"`
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`
	Token  string `yaml:"token"`
	Prefix string `yaml:"prefix"`
}

type VaultClient struct {
	Client *api.Client
	Config *VaultConfig
}

func NewVaultClient(cfg *VaultConfig) (*VaultClient, error) {
	vc := &VaultClient{Config: cfg}
	if err := vc.Connect(); err != nil {
		return nil, err
	}
	return vc, nil
}

func (v *VaultClient) Connect() error {
	c := api.DefaultConfig()
	c.Address = v.Config.Scheme + "://" + v.Config.Host + ":" + v.Config.Port

	tlsTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	c.HttpClient = &http.Client{
		Transport: tlsTransport,
		Timeout:   10 * time.Second,
	}

	client, err := api.NewClient(c)
	if err != nil {
		return fmt.Errorf("не удалось создать Vault client: %w", err)
	}

	if strings.TrimSpace(v.Config.Token) != "" {
		client.SetToken(v.Config.Token)
	}

	v.Client = client
	return nil
}

func (v *VaultClient) GetSecret(path string) (*api.Secret, error) {
	fullPath := v.fullSecretPath(path)

	secret, err := v.Client.Logical().Read(fullPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать секрет из Vault")
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("не удалось найти секрет по адресу %s", fullPath)
	}
	return secret, nil
}

func (v *VaultClient) fullSecretPath(path string) string {
	if v.Config.Prefix != "" {
		return fmt.Sprintf("%s/%s", v.Config.Prefix, path)
	}
	return path
}

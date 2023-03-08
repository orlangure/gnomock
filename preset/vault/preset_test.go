package vault_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/vault"
	"github.com/stretchr/testify/require"
)

func TestPreset(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"1.13", "1.12.4", "1.11.8", "1.10.11", "latest"} {
		tmpfile, err := os.CreateTemp("", "token")
		require.NoError(t, err)
		t.Run(version, testPreset(version, tmpfile.Name()))
	}
}

func testPreset(version string, filePath string) func(t *testing.T) {
	const policy = `
path "sys/mounts" {
  capabilities = ["list", "read"]
}

path "secret/*" {
  capabilities = ["list", "create"]
}

path "secret/data/*" {
  capabilities = ["create", "read"]
}

path "secret/metadata/*" {
  capabilities = ["list"]
}
`

	p := vault.Preset(
		vault.WithVersion(version),
		vault.WithAuthToken("root-token"),
		vault.WithAdditionalToken(vault.TokenCreate{
			FilePath: filePath,
			Policies: []string{"default"},
		}),
		vault.WithAuth([]vault.Auth{
			{
				Path: "k8s_cluster1",
				Type: "kubernetes",
			},
		}),
		vault.WithPolicies([]vault.Policy{
			{
				Name: "policy1",
				Data: policy,
			},
			{
				Name: "policy2",
				Data: "{}",
			},
		}),
	)

	return func(t *testing.T) {
		defer func() {
			_ = os.Remove(filePath)
		}()

		container, err := gnomock.Start(p)
		require.NoError(t, err)

		defer func() { require.NoError(t, gnomock.Stop(container)) }()

		vaultConfig := api.DefaultConfig()
		vaultConfig.Address = fmt.Sprintf("http://%s", container.DefaultAddress())

		cli, err := api.NewClient(vaultConfig)
		require.NoError(t, err)
		cli.SetToken("root-token")

		_, err = cli.Sys().Health()
		require.NoError(t, err)

		auth, err := cli.Sys().ListAuth()
		require.NoError(t, err)

		found := false

		for _, a := range auth {
			if a.Type == "kubernetes" {
				found = true

				break
			}
		}

		require.True(t, found, "kubenetes auth was not created")

		policies, err := cli.Sys().ListPolicies()
		require.NoError(t, err)

		found = false

		for _, p := range policies {
			if p == "policy1" {
				found = true

				break
			}
		}

		require.True(t, found, "policy %s was not created", "policy1")

		_, err = os.Stat(filePath)
		require.NoError(t, err, "file %s does not exist", filePath)

		d, err := os.ReadFile(filePath)
		require.NoError(t, err, "could not read file %s", filePath)
		require.True(t, strings.HasPrefix(string(d), "hvs."))

		cli.SetToken(string(d))
		_, err = cli.Sys().Health()
		require.NoError(t, err, "generated token is not valid")
	}
}

func TestPreset_withDefaults(t *testing.T) {
	t.Parallel()

	p := vault.Preset()
	container, err := gnomock.Start(
		p,
		gnomock.WithContainerName("vault-default"),
	)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = fmt.Sprintf("http://%s", container.DefaultAddress())

	cli, err := api.NewClient(vaultConfig)
	require.NoError(t, err)
	cli.SetToken("gnomock-vault-token")

	_, err = cli.Sys().Health()
	require.NoError(t, err)
}

package vault_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/vault"
	"github.com/stretchr/testify/require"
)

func TestPreset(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"1.13", "1.12.4", "1.11.8", "1.10.11", "latest"} {
		// tmpfile, err := os.CreateTemp("", "token")
		t.Run(version, testPreset(version, "root-token"))
	}
}

func testPreset(version, rootToken string) func(t *testing.T) {
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
		vault.WithAuthToken(rootToken),
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
		container, err := gnomock.Start(p)
		require.NoError(t, err)

		defer func() { require.NoError(t, gnomock.Stop(container)) }()

		vaultConfig := api.DefaultConfig()
		vaultConfig.Address = fmt.Sprintf("http://%s", container.DefaultAddress())

		cli, err := vault.Client(container, rootToken)
		require.NoError(t, err)

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

	cli, err := vault.Client(container, "gnomock-vault-token")
	require.NoError(t, err)

	_, err = cli.Sys().Health()
	require.NoError(t, err)
}

func TestCeateToken(t *testing.T) {
	t.Parallel()

	p := vault.Preset()
	container, err := gnomock.Start(
		p,
		gnomock.WithContainerName("vault-create-token"),
	)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	s, err := vault.CreateToken(container, "gnomock-vault-token", "default")
	require.NoError(t, err)

	cli, err := vault.Client(container, s)
	require.NoError(t, err)

	_, err = cli.Sys().Health()
	require.NoError(t, err)
}

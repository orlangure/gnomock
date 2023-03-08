# Gnomock Vault

Gnomock vault is a [Gnomock](https://github.com/orlangure/gnomock) preset
for running tests against a real vault container, without mocks.

The test below starts a vault server with:

* with a policy `policy1` configured
* with `root-token` set as root token
* with an additional token written in a temporary file that has only the `default` policy
* with an additional `kubernetes` secrets engine mounted on `k8s_cluster1`

```go
package vault_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/vault"
	"github.com/stretchr/testify/require"
)

func TestVault(t *testing.T) {
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

	tmpFile, err := os.CreateTemp("", "token")
	require.NoError(t, err)

	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	p := vault.Preset(
		vault.WithVersion("latest"),
		vault.WithAuthToken("root-token"),
		vault.WithTokenCreate(vault.TokenCreate{
			FilePath: tmpFile.Name(),
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
		}),
	)

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
}
```



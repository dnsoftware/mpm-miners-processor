package certmanager

import (
	"testing"

	"github.com/dnsoftware/mpm-save-get-shares/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestCertManager(t *testing.T) {
	path, err := utils.GetProjectRoot(".env")
	require.NoError(t, err)

	certMan, err := NewCertManager(path + "/certs")
	require.NoError(t, err)
	require.NotNil(t, certMan)

	credServer, err := certMan.GetServerCredentials()
	require.NoError(t, err)
	require.NotNil(t, credServer)

	credClient, err := certMan.GetClientCredentials()
	require.NoError(t, err)
	require.NotNil(t, credClient)

}

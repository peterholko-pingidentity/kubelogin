package cmd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/int128/kubelogin/pkg/oidc"
	"github.com/int128/kubelogin/pkg/testing/logger"
	"github.com/int128/kubelogin/pkg/tlsclientconfig"
	"github.com/int128/kubelogin/pkg/usecases/authentication"
	"github.com/int128/kubelogin/pkg/usecases/authentication/authcode"
	"github.com/int128/kubelogin/pkg/usecases/credentialplugin"
	"github.com/int128/kubelogin/pkg/usecases/credentialplugin/mock_credentialplugin"
	"github.com/int128/kubelogin/pkg/usecases/standalone"
	"github.com/int128/kubelogin/pkg/usecases/standalone/mock_standalone"
)

func TestCmd_Run(t *testing.T) {
	const executable = "kubelogin"
	const version = "HEAD"

	t.Run("root", func(t *testing.T) {
		tests := map[string]struct {
			args []string
			in   standalone.Input
		}{
			"Defaults": {
				args: []string{executable},
				in: standalone.Input{
					GrantOptionSet: authentication.GrantOptionSet{
						AuthCodeBrowserOption: &authcode.BrowserOption{
							BindAddress:           defaultListenAddress,
							AuthenticationTimeout: defaultAuthenticationTimeoutSec * time.Second,
							RedirectURLHostname:   "localhost",
						},
					},
				},
			},
			"FullOptions": {
				args: []string{executable,
					"--kubeconfig", "/path/to/kubeconfig",
					"--context", "hello.k8s.local",
					"--user", "google",
					"-v1",
				},
				in: standalone.Input{
					KubeconfigFilename: "/path/to/kubeconfig",
					KubeconfigContext:  "hello.k8s.local",
					KubeconfigUser:     "google",
					GrantOptionSet: authentication.GrantOptionSet{
						AuthCodeBrowserOption: &authcode.BrowserOption{
							BindAddress:           defaultListenAddress,
							AuthenticationTimeout: defaultAuthenticationTimeoutSec * time.Second,
							RedirectURLHostname:   "localhost",
						},
					},
				},
			},
		}
		for name, c := range tests {
			t.Run(name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				ctx := context.TODO()
				mockStandalone := mock_standalone.NewMockInterface(ctrl)
				mockStandalone.EXPECT().
					Do(ctx, c.in)
				cmd := Cmd{
					Root: &Root{
						Standalone: mockStandalone,
						Logger:     logger.New(t),
					},
					Logger: logger.New(t),
				}
				exitCode := cmd.Run(ctx, c.args, version)
				if exitCode != 0 {
					t.Errorf("exitCode wants 0 but %d", exitCode)
				}
			})
		}

		t.Run("TooManyArgs", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cmd := Cmd{
				Root: &Root{
					Standalone: mock_standalone.NewMockInterface(ctrl),
					Logger:     logger.New(t),
				},
				Logger: logger.New(t),
			}
			exitCode := cmd.Run(context.TODO(), []string{executable, "some"}, version)
			if exitCode != 1 {
				t.Errorf("exitCode wants 1 but %d", exitCode)
			}
		})
	})

	t.Run("get-token", func(t *testing.T) {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir error: %s", err)
		}

		tests := map[string]struct {
			args []string
			in   credentialplugin.Input
		}{
			"Defaults": {
				args: []string{executable,
					"get-token",
					"--oidc-issuer-url", "https://issuer.example.com",
					"--oidc-client-id", "YOUR_CLIENT_ID",
				},
				in: credentialplugin.Input{
					TokenCacheDir: userHomeDir + "/.kube/cache/oidc-login",
					Provider: oidc.Provider{
						IssuerURL: "https://issuer.example.com",
						ClientID:  "YOUR_CLIENT_ID",
					},
					GrantOptionSet: authentication.GrantOptionSet{
						AuthCodeBrowserOption: &authcode.BrowserOption{
							BindAddress:           defaultListenAddress,
							AuthenticationTimeout: defaultAuthenticationTimeoutSec * time.Second,
							RedirectURLHostname:   "localhost",
						},
					},
				},
			},
			"FullOptions": {
				args: []string{executable,
					"get-token",
					"--oidc-issuer-url", "https://issuer.example.com",
					"--oidc-client-id", "YOUR_CLIENT_ID",
					"--oidc-client-secret", "YOUR_CLIENT_SECRET",
					"--oidc-extra-scope", "email",
					"--oidc-extra-scope", "profile",
					"-v1",
				},
				in: credentialplugin.Input{
					TokenCacheDir: userHomeDir + "/.kube/cache/oidc-login",
					Provider: oidc.Provider{
						IssuerURL:    "https://issuer.example.com",
						ClientID:     "YOUR_CLIENT_ID",
						ClientSecret: "YOUR_CLIENT_SECRET",
						ExtraScopes:  []string{"email", "profile"},
					},
					GrantOptionSet: authentication.GrantOptionSet{
						AuthCodeBrowserOption: &authcode.BrowserOption{
							BindAddress:           defaultListenAddress,
							AuthenticationTimeout: defaultAuthenticationTimeoutSec * time.Second,
							RedirectURLHostname:   "localhost",
						},
					},
				},
			},
			"HomedirExpansion": {
				args: []string{executable,
					"get-token",
					"--oidc-issuer-url", "https://issuer.example.com",
					"--oidc-client-id", "YOUR_CLIENT_ID",
					"--certificate-authority", "~/.kube/ca.crt",
					"--local-server-cert", "~/.kube/oidc-server.crt",
					"--local-server-key", "~/.kube/oidc-server.key",
					"--token-cache-dir", "~/.kube/oidc-cache",
				},
				in: credentialplugin.Input{
					TokenCacheDir: userHomeDir + "/.kube/oidc-cache",
					Provider: oidc.Provider{
						IssuerURL: "https://issuer.example.com",
						ClientID:  "YOUR_CLIENT_ID",
					},
					GrantOptionSet: authentication.GrantOptionSet{
						AuthCodeBrowserOption: &authcode.BrowserOption{
							BindAddress:           defaultListenAddress,
							AuthenticationTimeout: defaultAuthenticationTimeoutSec * time.Second,
							RedirectURLHostname:   "localhost",
							LocalServerCertFile:   userHomeDir + "/.kube/oidc-server.crt",
							LocalServerKeyFile:    userHomeDir + "/.kube/oidc-server.key",
						},
					},
					TLSClientConfig: tlsclientconfig.Config{
						CACertFilename: []string{userHomeDir + "/.kube/ca.crt"},
					},
				},
			},
		}
		for name, c := range tests {
			t.Run(name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				ctx := context.TODO()
				getToken := mock_credentialplugin.NewMockInterface(ctrl)
				getToken.EXPECT().
					Do(ctx, c.in)
				cmd := Cmd{
					Root: &Root{
						Logger: logger.New(t),
					},
					GetToken: &GetToken{
						GetToken: getToken,
						Logger:   logger.New(t),
					},
					Logger: logger.New(t),
				}
				exitCode := cmd.Run(ctx, c.args, version)
				if exitCode != 0 {
					t.Errorf("exitCode wants 0 but %d", exitCode)
				}
			})
		}

		t.Run("MissingMandatoryOptions", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := context.TODO()
			cmd := Cmd{
				Root: &Root{
					Logger: logger.New(t),
				},
				GetToken: &GetToken{
					GetToken: mock_credentialplugin.NewMockInterface(ctrl),
					Logger:   logger.New(t),
				},
				Logger: logger.New(t),
			}
			exitCode := cmd.Run(ctx, []string{executable, "get-token"}, version)
			if exitCode != 1 {
				t.Errorf("exitCode wants 1 but %d", exitCode)
			}
		})

		t.Run("TooManyArgs", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := context.TODO()
			cmd := Cmd{
				Root: &Root{
					Logger: logger.New(t),
				},
				GetToken: &GetToken{
					GetToken: mock_credentialplugin.NewMockInterface(ctrl),
					Logger:   logger.New(t),
				},
				Logger: logger.New(t),
			}
			exitCode := cmd.Run(ctx, []string{executable, "get-token", "foo"}, version)
			if exitCode != 1 {
				t.Errorf("exitCode wants 1 but %d", exitCode)
			}
		})
	})
}

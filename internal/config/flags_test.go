package config

// import (
// 	"flag"
// 	"io"
// 	"os"
// 	"testing"
// )

// func resetFlags() {
// 	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)
// 	flag.CommandLine.SetOutput(io.Discard)
// }

// func TestParseFlags(t *testing.T) {
// 	originalArgs := os.Args
// 	originalEnvServerHost := os.Getenv("SERVER_ADDRESS")
// 	originalEnvBaseURL := os.Getenv("BASE_URL")

// 	t.Cleanup(func() {
// 		os.Args = originalArgs
// 		os.Setenv("SERVER_ADDRESS", originalEnvServerHost)
// 		os.Setenv("BASE_URL", originalEnvBaseURL)
// 		resetFlags()
// 	})

// 	testCases := []struct {
// 		name               string
// 		envServerHost      string
// 		envBaseURL         string
// 		cliArgs            []string
// 		expectedServerHost string
// 		expectedBaseURL    string
// 	}{
// 		{
// 			name:               "1. Default values (no env, no cli args)",
// 			envServerHost:      "",
// 			envBaseURL:         "",
// 			cliArgs:            []string{"./testprogram"},
// 			expectedServerHost: "localhost:8080",
// 			expectedBaseURL:    "http://localhost:8080",
// 		},
// 		{
// 			name:               "2. Environment variables set, no cli args",
// 			envServerHost:      "env_host:9000",
// 			envBaseURL:         "http://env_base",
// 			cliArgs:            []string{"./testprogram"},
// 			expectedServerHost: "env_host:9000",
// 			expectedBaseURL:    "http://env_base",
// 		},
// 		{
// 			name:               "3. Command-line args set, no env vars",
// 			envServerHost:      "",
// 			envBaseURL:         "",
// 			cliArgs:            []string{"./testprogram", "-a", "cli_host:9000", "-b", "http://cli_base"},
// 			expectedServerHost: "cli_host:9000",
// 			expectedBaseURL:    "http://cli_base",
// 		},
// 		{
// 			name:               "4. Environment variables and command-line args (env takes precedence if set)",
// 			envServerHost:      "env_host_primary:8000",
// 			envBaseURL:         "http://env_base_primary",
// 			cliArgs:            []string{"./testprogram", "-a", "cli_host_secondary:9000", "-b", "http://cli_base_secondary"},
// 			expectedServerHost: "env_host_primary:8000",
// 			expectedBaseURL:    "http://env_base_primary",
// 		},
// 		{
// 			name:               "5. Mixed: one env, one cli arg (env empty for base URL)",
// 			envServerHost:      "mixed_env_host:7000",
// 			envBaseURL:         "",
// 			cliArgs:            []string{"./testprogram", "-a", "cli_host_ignored:9000", "-b", "http://mixed_cli_base"},
// 			expectedServerHost: "mixed_env_host:7000",
// 			expectedBaseURL:    "http://mixed_cli_base",
// 		},
// 		{
// 			name:               "6. Mixed: one cli arg, one env (cli empty for server host)",
// 			envServerHost:      "",
// 			envBaseURL:         "http://mixed_env_base",
// 			cliArgs:            []string{"./testprogram", "-a", "mixed_cli_host:6000", "-b", "http://cli_base_ignored"},
// 			expectedServerHost: "mixed_cli_host:6000",
// 			expectedBaseURL:    "http://mixed_env_base",
// 		},
// 		{
// 			name:               "7. CLI args override default, but not empty env var (if env.Parse makes it empty)",
// 			envServerHost:      "", // Env делает его пустым
// 			envBaseURL:         "", // Env делает его пустым
// 			cliArgs:            []string{"./testprogram", "-a", "cli_override_default:1000", "-b", "http://cli_override_default"},
// 			expectedServerHost: "cli_override_default:1000",
// 			expectedBaseURL:    "http://cli_override_default",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// **Внутри каждого t.Run** нужно выполнить очистку состояния
// 			// Переменные окружения и os.Args сбрасываются в t.Cleanup() родительской функции
// 			// Но flag.CommandLine нужно сбрасывать здесь, чтобы каждый подтест начинался с чистого листа.
// 			resetFlags()

// 			// Устанавливаем переменные окружения для текущего тест-кейса
// 			// os.Setenv устанавливает переменную для текущего процесса и всех дочерних,
// 			// поэтому ее нужно обязательно очищать после каждого теста.
// 			if tc.envServerHost != "" {
// 				os.Setenv("SERVER_ADDRESS", tc.envServerHost)
// 			} else {
// 				os.Unsetenv("SERVER_ADDRESS")
// 			}
// 			if tc.envBaseURL != "" {
// 				os.Setenv("BASE_URL", tc.envBaseURL)
// 			} else {
// 				os.Unsetenv("BASE_URL")
// 			}

// 			os.Args = tc.cliArgs

// 			cfg := &Config{}
// 			cfg.ParseFlags()

// 			// Проверяем результаты
// 			if cfg.ServerHost != tc.expectedServerHost {
// 				t.Errorf("ServerHost: ожидалось %q, получено %q", tc.expectedServerHost, cfg.ServerHost)
// 			}
// 			if cfg.BaseURL != tc.expectedBaseURL {
// 				t.Errorf("BaseURL: ожидалось %q, получено %q", tc.expectedBaseURL, cfg.BaseURL)
// 			}
// 		})
// 	}
// }

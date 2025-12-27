package config

import (
	"flag"
	"io"
	"os"
	"testing"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)
	flag.CommandLine.SetOutput(io.Discard)
}

func TestParseFlags(t *testing.T) {
	// Сохраняем оригинальные os.Args и переменные окружения для восстановления после тестов
	originalArgs := os.Args
	originalEnvServerHost := os.Getenv("SERVER_ADDRESS")
	originalEnvBaseURL := os.Getenv("BASE_URL")

	// Регистрируем функцию очистки, которая будет вызвана после каждого теста.
	// Это критически важно для изоляции тестов.
	t.Cleanup(func() {
		os.Args = originalArgs                             // Восстанавливаем оригинальные аргументы командной строки
		os.Setenv("SERVER_ADDRESS", originalEnvServerHost) // Восстанавливаем переменную окружения
		os.Setenv("BASE_URL", originalEnvBaseURL)          // Восстанавливаем переменную окружения
		resetFlags()                                       // Сбрасываем состояние пакета flag
	})

	testCases := []struct {
		name               string
		envServerHost      string
		envBaseURL         string
		cliArgs            []string
		expectedServerHost string
		expectedBaseURL    string
	}{
		{
			name:               "1. Default values (no env, no cli args)",
			envServerHost:      "",
			envBaseURL:         "",
			cliArgs:            []string{"./testprogram"},
			expectedServerHost: "localhost:8080",
			expectedBaseURL:    "http://localhost:8080",
		},
		{
			name:               "2. Environment variables set, no cli args",
			envServerHost:      "env_host:9000",
			envBaseURL:         "http://env_base",
			cliArgs:            []string{"./testprogram"},
			expectedServerHost: "env_host:9000",
			expectedBaseURL:    "http://env_base",
		},
		{
			name:               "3. Command-line args set, no env vars",
			envServerHost:      "",
			envBaseURL:         "",
			cliArgs:            []string{"./testprogram", "-a", "cli_host:9000", "-b", "http://cli_base"},
			expectedServerHost: "cli_host:9000",
			expectedBaseURL:    "http://cli_base",
		},
		{
			name:               "4. Environment variables and command-line args (env takes precedence if set)",
			envServerHost:      "env_host_primary:8000",
			envBaseURL:         "http://env_base_primary",
			cliArgs:            []string{"./testprogram", "-a", "cli_host_secondary:9000", "-b", "http://cli_base_secondary"},
			expectedServerHost: "env_host_primary:8000",
			expectedBaseURL:    "http://env_base_primary",
		},
		{
			name:               "5. Mixed: one env, one cli arg (env empty for base URL)",
			envServerHost:      "mixed_env_host:7000",
			envBaseURL:         "",
			cliArgs:            []string{"./testprogram", "-a", "cli_host_ignored:9000", "-b", "http://mixed_cli_base"},
			expectedServerHost: "mixed_env_host:7000",
			expectedBaseURL:    "http://mixed_cli_base",
		},
		{
			name:               "6. Mixed: one cli arg, one env (cli empty for server host)",
			envServerHost:      "",
			envBaseURL:         "http://mixed_env_base",
			cliArgs:            []string{"./testprogram", "-a", "mixed_cli_host:6000", "-b", "http://cli_base_ignored"},
			expectedServerHost: "mixed_cli_host:6000",
			expectedBaseURL:    "http://mixed_env_base",
		},
		{
			name:               "7. CLI args override default, but not empty env var (if env.Parse makes it empty)",
			envServerHost:      "", // Env делает его пустым
			envBaseURL:         "", // Env делает его пустым
			cliArgs:            []string{"./testprogram", "-a", "cli_override_default:1000", "-b", "http://cli_override_default"},
			expectedServerHost: "cli_override_default:1000",
			expectedBaseURL:    "http://cli_override_default",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// **Внутри каждого t.Run** нужно выполнить очистку состояния
			// Переменные окружения и os.Args сбрасываются в t.Cleanup() родительской функции
			// Но flag.CommandLine нужно сбрасывать здесь, чтобы каждый подтест начинался с чистого листа.
			resetFlags()

			// Устанавливаем переменные окружения для текущего тест-кейса
			// os.Setenv устанавливает переменную для текущего процесса и всех дочерних,
			// поэтому ее нужно обязательно очищать после каждого теста.
			if tc.envServerHost != "" {
				os.Setenv("SERVER_ADDRESS", tc.envServerHost)
			} else {
				os.Unsetenv("SERVER_ADDRESS")
			}
			if tc.envBaseURL != "" {
				os.Setenv("BASE_URL", tc.envBaseURL)
			} else {
				os.Unsetenv("BASE_URL")
			}

			// Устанавливаем аргументы командной строки для текущего тест-кейса
			os.Args = tc.cliArgs

			// Создаем новый экземпляр Config для каждого теста
			cfg := &Config{}
			cfg.ParseFlags()

			// Проверяем результаты
			if cfg.ServerHost != tc.expectedServerHost {
				t.Errorf("ServerHost: ожидалось %q, получено %q", tc.expectedServerHost, cfg.ServerHost)
			}
			if cfg.BaseURL != tc.expectedBaseURL {
				t.Errorf("BaseURL: ожидалось %q, получено %q", tc.expectedBaseURL, cfg.BaseURL)
			}
		})
	}
}

// // TestParseFlags_EnvParseError имитирует ситуацию, когда env.Parse() вызывает log.Fatal.
// // Для этого нужно перехватить вызов os.Exit(1), который делает log.Fatal.
// // Это более сложный тест, т.к. требует временного переопределения os.Exit.
// func TestParseFlags_EnvParseError(t *testing.T) {
// 	// Сохраняем оригинальные os.Args и os.Exit
// 	originalArgs := os.Args
// 	originalExit := os.Exit
// 	originalEnv := os.Environ()

// 	// Используем буфер для захвата вывода log.Fatal
// 	var logOutput bytes.Buffer
// 	// Сохраняем оригинальный вывод log и перенаправляем его
// 	originalLogOutput := flag.CommandLine.Output()
// 	flag.CommandLine.SetOutput(&logOutput)

// 	// Регистрируем функцию очистки
// 	t.Cleanup(func() {
// 		os.Args = originalArgs
// 		os.Exit = originalExit
// 		// Восстанавливаем оригинальные переменные окружения
// 		for _, e := range originalEnv {
// 			parts := strings.SplitN(e, "=", 2)
// 			os.Setenv(parts[0], parts[1])
// 		}
// 		resetFlags()                                  // Сбрасываем flag.CommandLine
// 		flag.CommandLine.SetOutput(originalLogOutput) // Восстанавливаем вывод log
// 	})

// 	// Устанавливаем переменную окружения, которая вызовет ошибку (если бы был int тип)
// 	// В данном случае, с двумя строковыми полями ServerHost и BaseURL,
// 	// сложно вызвать ошибку env.Parse. Если бы Config содержал:
// 	// Port int `env:"PORT"`
// 	// Тогда os.Setenv("PORT", "not_a_number") вызвало бы ошибку.
// 	// Для текущей структуры Config, env.Parse() для строк не вызовет Fatal,
// 	// но мы все равно покажем, как это тестировать, если бы такая ошибка была.
// 	// Например, мы можем создать фиктивную ошибку для env.Parse (для этого
// 	// потребовалось бы изменить функцию ParseFlags, чтобы она могла принимать
// 	// mock env.Parser, что является хорошей практикой для тестируемости).
// 	// В данном случае мы просто покажем, что log.Fatal будет вызван, если err != nil.

// 	// Для демонстрации, предположим, что env.Parse каким-то образом возвращает ошибку.
// 	// В реальном коде с `caarlos0/env` и двумя строками, это сложно,
// 	// но если бы был `Port int env:"PORT"`, то `os.Setenv("PORT", "abc")`
// 	// вызвало бы ошибку парсинга.

// 	// Создаем "фиктивную" ошибку, переопределяя поведение os.Exit
// 	exitCalled := false
// 	os.Exit = func(code int) {
// 		exitCalled = true
// 		if code != 1 {
// 			t.Errorf("Ожидался os.Exit(1), получено os.Exit(%d)", code)
// 		}
// 		panic("os.Exit called") // Паникуем, чтобы прервать выполнение теста
// 	}

// 	cfg := &config.Config{}
// 	// Имитируем условие, при котором env.Parse вернет ошибку (например, если бы был `int` поле)
// 	// В данном случае, это просто заглушка для демонстрации, так как с текущими полями Config
// 	// env.Parse не выдаст ошибку для строк.
// 	//
// 	// Если бы Config был:
// 	// type Config struct {
// 	//  ServerHost string `env:"SERVER_ADDRESS"`
// 	//  BaseURL    string `env:"BASE_URL"`
// 	//  Port       int    `env:"PORT"` // Новое поле
// 	// }
// 	// Тогда мы могли бы сделать:
// 	// os.Setenv("PORT", "not_a_number")

// 	// Поскольку у нас нет такого поля, этот тест просто покажет структуру.
// 	// Если бы в ParseFlags была какая-то другая логика, которая могла бы привести к log.Fatal
// 	// до env.Parse, то это было бы актуальнее.
// 	// Для чистоты данного теста, мы не можем легко заставить env.Parse вернуть ошибку
// 	// с текущей структурой `Config`. Тест-кейс для `log.Fatal` будет более уместен,
// 	// если `ParseFlags` будет иметь более сложную логику, которая может падать.

// 	// Тест будет "паниковать", если `log.Fatal` будет вызван.
// 	// Мы можем обернуть вызов в `recover`.
// 	defer func() {
// 		if r := recover(); r == nil {
// 			// t.Error("Ожидалось паника (вызов log.Fatal / os.Exit), но ее не было.")
// 			// Если env.Parse не вызывает Fatal для текущей Config, то это нормально.
// 		} else if !exitCalled {
// 			t.Errorf("Паника произошла, но os.Exit не был вызван. Ошибка: %v", r)
// 		}
// 	}()

// 	cfg.ParseFlags() // Вызываем функцию, которая может вызвать log.Fatal

// 	// Проверяем, что os.Exit был вызван
// 	if !exitCalled {
// 		// Это произойдет, если env.Parse не смог вызвать Fatal.
// 		// Для `string` полей `caarlos0/env` всегда успешен.
// 		// Чтобы этот тест был полезным, нужно изменить `Config` и спровоцировать ошибку.
// 		t.Skip("Пропуск: env.Parse не вызывает Fatal для строковых полей в этой конфигурации.")
// 	}

// 	// Проверяем, что log.Fatal вывел сообщение
// 	if !strings.Contains(logOutput.String(), "Error parsing environment variables:") {
// 		// Ожидалось бы сообщение об ошибке парсинга env
// 	}
// }

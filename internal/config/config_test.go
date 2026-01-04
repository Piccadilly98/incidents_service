package config

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestValidtionPort(t *testing.T) {
	testCases := []struct {
		name         string
		envName      string
		valueEnv     string
		expectError  bool
		expectedPort string
		defaultPort  string
	}{
		{
			name:         "valid_port_5433",
			envName:      "DB_PORT",
			valueEnv:     "5433",
			expectError:  false,
			expectedPort: "5433",
			defaultPort:  "5432",
		},
		{
			name:         "valid_min_port_1",
			envName:      "APP_PORT",
			valueEnv:     "1",
			expectError:  false,
			expectedPort: "1",
			defaultPort:  "8080",
		},
		{
			name:         "valid_max_port_65535",
			envName:      "REDIS_PORT",
			valueEnv:     "65535",
			expectError:  false,
			expectedPort: "65535",
			defaultPort:  "6379",
		},
		{
			name:         "valid_common_port_8080",
			envName:      "PORT",
			valueEnv:     "8080",
			expectError:  false,
			expectedPort: "8080",
			defaultPort:  "3000",
		},

		{
			name:         "empty_env_uses_default",
			envName:      "DB_PORT",
			valueEnv:     "",
			expectError:  false,
			expectedPort: "5432",
			defaultPort:  "5432",
		},
		{
			name:         "env_not_set_uses_default",
			envName:      "UNSET_VAR",
			valueEnv:     "",
			expectError:  false,
			expectedPort: "8080",
			defaultPort:  "8080",
		},

		{
			name:         "invalid_not_a_number",
			envName:      "PORT",
			valueEnv:     "abc",
			expectError:  true,
			expectedPort: "",
			defaultPort:  "8080",
		},
		{
			name:         "invalid_negative_number",
			envName:      "PORT",
			valueEnv:     "-1",
			expectError:  true,
			expectedPort: "",
			defaultPort:  "8080",
		},
		{
			name:         "invalid_zero_port",
			envName:      "PORT",
			valueEnv:     "0",
			expectError:  true,
			expectedPort: "",
			defaultPort:  "8080",
		},
		{
			name:         "invalid_too_large_port",
			envName:      "PORT",
			valueEnv:     "65536",
			expectError:  true,
			expectedPort: "",
			defaultPort:  "8080",
		},
		{
			name:         "invalid_port_with_spaces",
			envName:      "PORT",
			valueEnv:     " 8080 ",
			expectedPort: "8080",
			defaultPort:  "3000",
		},
		{
			name:         "invalid_port_decimal",
			envName:      "PORT",
			valueEnv:     "8080.5",
			expectError:  true,
			expectedPort: "",
			defaultPort:  "3000",
		},

		{
			name:         "valid_with_leading_zeros",
			envName:      "PORT",
			valueEnv:     "08080",
			expectError:  false,
			expectedPort: "08080",
			defaultPort:  "3000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := os.Setenv(tc.envName, tc.valueEnv)
			if err != nil {
				t.Fatalf("unexpected error: %s\n", err.Error())
			}

			res, err := validationPort(tc.envName, tc.defaultPort)
			if err != nil {
				if !tc.expectError {
					t.Errorf("unexpected error: %s\n", err.Error())
				}
			}

			if tc.expectError {
				if res != "" {
					t.Errorf("unexpected port: %s\n", res)
				}
			} else {
				if res != tc.expectedPort {
					t.Errorf("RESULT: got: %s, expect: %s\n", res, tc.expectedPort)
				}
			}
		})
	}
}

func TestNewConfig_NoEnv(t *testing.T) {
	testCases := []struct {
		name               string
		serverAddrValue    string
		serverPortValue    string
		nameDbValue        string
		dbSslValue         string
		dbHostValue        string
		dbPortValue        string
		dbUserValue        string
		dbPasswordValue    string
		webhookURLValue    string
		webhookMethodValue string

		expectCfg             bool
		expectedError         error
		expectedWebhookURL    string
		expectedWebhookMethod string
		expectedServerAddr    string
		expectedServerPort    string
		expectConnStr         string
	}{
		// === ВАЛИДНЫЕ СЦЕНАРИИ ===
		{
			name:               "valid_all_env_vars_set",
			serverAddrValue:    "localhost",
			serverPortValue:    "8080",
			nameDbValue:        "postgres",
			dbSslValue:         "disable",
			dbHostValue:        "localhost",
			dbPortValue:        "5432",
			dbUserValue:        "postgres",
			dbPasswordValue:    "1234",
			webhookURLValue:    "localhost:9090",
			webhookMethodValue: "POST",

			expectCfg:             true,
			expectedWebhookURL:    "localhost:9090",
			expectedServerAddr:    "localhost",
			expectedServerPort:    "8080",
			expectedWebhookMethod: http.MethodPost,
			expectConnStr:         "user=postgres port=5432 password=1234 dbname=postgres host=localhost sslmode=disable",
		},
		{
			name:               "valid_with_defaults_for_optional",
			serverAddrValue:    "",
			serverPortValue:    "3000",
			nameDbValue:        "testdb",
			dbSslValue:         "",
			dbHostValue:        "",
			dbPortValue:        "5433",
			dbUserValue:        "admin",
			dbPasswordValue:    "admin123",
			webhookURLValue:    "",
			webhookMethodValue: "",

			expectCfg:             true,
			expectedWebhookURL:    DefaultWebhookURL,
			expectedServerAddr:    DefaultServerAddress,
			expectedServerPort:    "3000",
			expectedWebhookMethod: DefaultWebhookMethod,
			expectConnStr:         "user=admin port=5433 password=admin123 dbname=testdb host=" + DefaultDbHost + " sslmode=" + DefaultDbSSLMode,
		},
		{
			name:               "valid_webhook_method_get",
			serverPortValue:    "8080",
			nameDbValue:        "test",
			dbPortValue:        "5432",
			dbUserValue:        "user",
			dbPasswordValue:    "pass",
			webhookURLValue:    "http://webhook:9090",
			webhookMethodValue: "GET",

			expectCfg:             true,
			expectedServerAddr:    DefaultServerAddress,
			expectedWebhookURL:    "http://webhook:9090",
			expectedServerPort:    "8080",
			expectedWebhookMethod: "GET",
			expectConnStr:         "user=user port=5432 password=pass dbname=test host=" + DefaultDbHost + " sslmode=" + DefaultDbSSLMode,
		},
		{
			name:            "valid_minimal_required_fields",
			serverPortValue: "8080",
			nameDbValue:     "minimal",
			dbPortValue:     "5432",
			dbUserValue:     "minimal",
			dbPasswordValue: "minimal",

			expectCfg:             true,
			expectedWebhookURL:    DefaultWebhookURL,
			expectedServerAddr:    DefaultServerAddress,
			expectedServerPort:    "8080",
			expectedWebhookMethod: DefaultWebhookMethod,
			expectConnStr:         "user=minimal port=5432 password=minimal dbname=minimal host=" + DefaultDbHost + " sslmode=" + DefaultDbSSLMode,
		},
		{
			name:            "valid_edge_port_values",
			serverPortValue: "1",
			nameDbValue:     "test",
			dbPortValue:     "65535",
			dbUserValue:     "user",
			dbPasswordValue: "pass",

			expectCfg:             true,
			expectedServerAddr:    DefaultServerAddress,
			expectedWebhookURL:    DefaultWebhookURL,
			expectedWebhookMethod: DefaultWebhookMethod,
			expectedServerPort:    "1",
			expectConnStr:         "user=user port=65535 password=pass dbname=test host=" + DefaultDbHost + " sslmode=" + DefaultDbSSLMode,
		},

		// === НЕВАЛИДНЫЕ СЦЕНАРИИ (ожидаем ошибки) ===
		{
			name:            "error_missing_db_name",
			serverPortValue: "8080",
			nameDbValue:     "",
			dbPortValue:     "5432",
			dbUserValue:     "user",
			dbPasswordValue: "pass",

			expectCfg:     false,
			expectedError: fmt.Errorf("db_name cannot be empty"),
		},
		{
			name:            "error_missing_db_user",
			serverPortValue: "8080",
			nameDbValue:     "test",
			dbPortValue:     "5432",
			dbUserValue:     "",
			dbPasswordValue: "pass",

			expectCfg:     false,
			expectedError: fmt.Errorf("db_user cannot be empty"),
		},
		{
			name:            "error_missing_db_password",
			serverPortValue: "8080",
			nameDbValue:     "test",
			dbPortValue:     "5432",
			dbUserValue:     "user",
			dbPasswordValue: "",

			expectCfg:     false,
			expectedError: fmt.Errorf("db_password cannot be empty"),
		},
		{
			name:            "error_invalid_server_port",
			serverPortValue: "99999",
			nameDbValue:     "test",
			dbPortValue:     "5432",
			dbUserValue:     "user",
			dbPasswordValue: "pass",

			expectCfg:     false,
			expectedError: fmt.Errorf("port 99999 out of range (1-65535)"),
		},
		{
			name:            "error_invalid_db_port",
			serverPortValue: "8080",
			nameDbValue:     "test",
			dbPortValue:     "0",
			dbUserValue:     "user",
			dbPasswordValue: "pass",

			expectCfg:     false,
			expectedError: fmt.Errorf("port 0 out of range (1-65535)"),
		},
		{
			name:            "error_invalid_server_port_format",
			serverPortValue: "abc",
			nameDbValue:     "test",
			dbPortValue:     "5432",
			dbUserValue:     "user",
			dbPasswordValue: "pass",

			expectCfg:     false,
			expectedError: fmt.Errorf("invalid port 'abc'"),
		},
		{
			name:            "error_negative_server_port",
			serverPortValue: "-1",
			nameDbValue:     "test",
			dbPortValue:     "5432",
			dbUserValue:     "user",
			dbPasswordValue: "pass",

			expectCfg:     false,
			expectedError: fmt.Errorf("port -1 out of range (1-65535)"),
		},

		// === WEBHOOK МЕТОДЫ ===
		{
			name:               "webhook_method_invalid_falls_back_to_default",
			serverPortValue:    "8080",
			nameDbValue:        "test",
			dbPortValue:        "5432",
			dbUserValue:        "user",
			dbPasswordValue:    "pass",
			webhookMethodValue: "PUT",

			expectCfg:             true,
			expectedServerAddr:    DefaultServerAddress,
			expectedWebhookURL:    DefaultWebhookURL,
			expectedWebhookMethod: DefaultWebhookMethod,
			expectedServerPort:    "8080",
			expectConnStr:         "user=user port=5432 password=pass dbname=test host=" + DefaultDbHost + " sslmode=" + DefaultDbSSLMode,
		},
		{
			name:               "webhook_method_case_insensitive_should_fail",
			serverPortValue:    "8080",
			nameDbValue:        "test",
			dbPortValue:        "5432",
			dbUserValue:        "user",
			dbPasswordValue:    "pass",
			webhookMethodValue: "post",

			expectCfg:             true,
			expectedWebhookMethod: DefaultWebhookMethod,
			expectedServerAddr:    DefaultServerAddress,
			expectedWebhookURL:    DefaultWebhookURL,
			expectedServerPort:    "8080",
			expectConnStr:         "user=user port=5432 password=pass dbname=test host=" + DefaultDbHost + " sslmode=disable",
		},
		{
			name:               "webhook_method_lowercase_get_should_fail",
			serverPortValue:    "8080",
			nameDbValue:        "test",
			dbPortValue:        "5432",
			dbUserValue:        "user",
			dbPasswordValue:    "pass",
			webhookMethodValue: "get",

			expectCfg:             true,
			expectedWebhookMethod: DefaultWebhookMethod,
			expectedServerPort:    "8080",
			expectedServerAddr:    DefaultServerAddress,
			expectedWebhookURL:    DefaultWebhookURL,
			expectConnStr:         "user=user port=5432 password=pass dbname=test host=" + DefaultDbHost + " sslmode=disable",
		},

		// === СПЕЦИАЛЬНЫЕ СЛУЧАИ ===
		{
			name:               "empty_strings_treated_as_empty_not_default",
			serverAddrValue:    "",
			serverPortValue:    "",
			nameDbValue:        "test",
			dbSslValue:         "",
			dbHostValue:        "",
			dbPortValue:        "",
			dbUserValue:        "user",
			dbPasswordValue:    "pass",
			webhookURLValue:    "",
			webhookMethodValue: "",

			expectCfg:             true,
			expectedWebhookURL:    DefaultWebhookURL,
			expectedServerAddr:    DefaultServerAddress,
			expectedServerPort:    DefaultServerPort,
			expectedWebhookMethod: DefaultWebhookMethod,
			expectConnStr:         "user=user port=" + DefaultDbPort + " password=pass dbname=test host=" + DefaultDbHost + " sslmode=" + DefaultDbSSLMode,
		},
		{
			name:            "special_characters_in_password",
			serverPortValue: "8080",
			nameDbValue:     "test",
			dbPortValue:     "5432",
			dbUserValue:     "user",
			dbPasswordValue: "p@ssw0rd!123#",

			expectCfg:             true,
			expectConnStr:         "user=user port=5432 password=p@ssw0rd!123# dbname=test host=" + DefaultDbHost + " sslmode=" + DefaultDbSSLMode,
			expectedServerAddr:    DefaultServerAddress,
			expectedWebhookURL:    DefaultWebhookURL,
			expectedWebhookMethod: DefaultWebhookMethod,
			expectedServerPort:    "8080",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.Setenv(EnvNameServerAddr, tc.serverAddrValue); err != nil {
				t.Fatalf("FAIL TO LOAD SERVER_ADDR: %s\n", err.Error())
			}
			if err := os.Setenv(EnvNameServerPort, tc.serverPortValue); err != nil {
				t.Fatalf("FAIL TO LOAD SERVER_PORT: %s\n", err.Error())
			}

			if err := os.Setenv(EnvNameDbHost, tc.dbHostValue); err != nil {
				t.Fatalf("FAIL TO LOAD DB_HOST: %s\n", err.Error())
			}
			if err := os.Setenv(EnvNameDbPort, tc.dbPortValue); err != nil {
				t.Fatalf("FAIL TO LOAD DB_PORT: %s\n", err.Error())
			}
			if err := os.Setenv(EnvNameDbName, tc.nameDbValue); err != nil {
				t.Fatalf("FAIL TO LOAD DB_NAME: %s\n", err.Error())
			}
			if err := os.Setenv(EnvNameDbPassword, tc.dbPasswordValue); err != nil {
				t.Fatalf("FAIL TO LOAD DB_PASSWORD: %s\n", err.Error())
			}
			if err := os.Setenv(EnvNameDbUser, tc.dbUserValue); err != nil {
				t.Fatalf("FAIL TO LOAD DB_USER: %s\n", err.Error())
			}
			if err := os.Setenv(EnvNameDbSSlMode, tc.dbSslValue); err != nil {
				t.Fatalf("FAIL TO LOAD DB_SSLMODE: %s\n", err.Error())
			}

			if err := os.Setenv(EnvNameWebHookURL, tc.webhookURLValue); err != nil {
				t.Fatalf("FAIL TO LOAD WEBHOOK_URL: %s\n", err.Error())
			}
			if err := os.Setenv(EnvNameWebhookMethod, tc.webhookMethodValue); err != nil {
				t.Fatalf("FAIL TO LOAD WEBHOOK_METHOD: %s\n", err.Error())
			}
			// tc.expectConnStr = fmt.Sprintf("user=%s port=%s password=%s dbname=%s host=%s sslmode=%s", tc.dbUserValue, tc.dbPortValue, tc.dbPasswordValue, tc.nameDbValue, tc.dbHostValue, tc.dbSslValue)
			cfg, err := NewConfig(false)
			if err != nil {
				if tc.expectedError != nil {
					if !strings.Contains(err.Error(), tc.expectedError.Error()) {
						t.Errorf("ERROR: got: %s, expect: %s\n", err.Error(), tc.expectedError.Error())
					}
				} else {
					t.Errorf("unexpected error: %s\n", err.Error())
				}
				if cfg != nil {
					t.Errorf("unexpectd cfg\n")
				}
			} else {
				if tc.expectedError != nil {
					t.Errorf("ERROR: got: nil, expect: %s\n", tc.expectedError.Error())
					return
				}
				if cfg == nil {
					t.Errorf("cfg==nil, expect != nil\n")
					return
				}
				if cfg.ConnectionStr != tc.expectConnStr {
					t.Errorf("CONN STR: got: %s, expect: %s\n", cfg.ConnectionStr, tc.expectConnStr)
				}
				if cfg.ServerAddr != tc.expectedServerAddr {
					t.Errorf("SERV ADDR: got: %s, expect: %s\n", cfg.ServerAddr, tc.expectedServerAddr)
				}
				if cfg.ServerPort != tc.expectedServerPort {
					t.Errorf("SERV PORT: got: %s, expect: %s\n", cfg.ServerPort, tc.expectedServerPort)
				}
				if cfg.WebhookURL != tc.expectedWebhookURL {
					t.Errorf("WEBHOOK URL: got: %s, expect: %s\n", cfg.WebhookURL, tc.expectedWebhookURL)
				}
				if cfg.WebhookMethod != tc.expectedWebhookMethod {
					t.Errorf("WEBHOOK METHOD: got: %s, expect: %s\n", cfg.WebhookMethod, tc.expectedWebhookMethod)
				}
			}
		})
	}
}

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

func TestNewConfig_AdditionalFields(t *testing.T) {
	defer func() {
		os.Unsetenv(EnvNameDefaultIncidentRadius)
		os.Unsetenv(EnvNameMaxIncidentRadius)
		os.Unsetenv(EnvMaxRowsInPage)
		os.Unsetenv(EnvNameStatsTime)
		os.Unsetenv(EnvNameLoggingUserError)

		os.Setenv(EnvNameDbName, "testdb")
		os.Setenv(EnvNameDbUser, "user")
		os.Setenv(EnvNameDbPassword, "pass")
	}()

	testCases := []struct {
		name             string
		defaultRadiusEnv string
		maxRadiusEnv     string
		maxRowsEnv       string
		statsTimeEnv     string
		loggingUserError string

		expectedDefaultRadius int
		expectedMaxRadius     int
		expectedMaxRows       int
		expectedStatsTime     int
		expectedLoggingError  bool

		expectError           bool
		expectedErrorContains string
	}{
		{
			name:             "all_additional_fields_custom_valid",
			defaultRadiusEnv: "10000",
			maxRadiusEnv:     "100000",
			maxRowsEnv:       "50",
			statsTimeEnv:     "30",
			loggingUserError: "true",

			expectedDefaultRadius: 10000,
			expectedMaxRadius:     100000,
			expectedMaxRows:       50,
			expectedStatsTime:     30,
			expectedLoggingError:  true,
			expectError:           false,
		},
		{
			name:             "all_additional_fields_default",
			defaultRadiusEnv: "",
			maxRadiusEnv:     "",
			maxRowsEnv:       "",
			statsTimeEnv:     "",
			loggingUserError: "",

			expectedDefaultRadius: DefaultRadius,
			expectedMaxRadius:     DefaultMaxRadius,
			expectedMaxRows:       DefaultMaxRowsInPage,
			expectedStatsTime:     DefaultStatsTime,
			expectedLoggingError:  DefaultLoggingUserError,
			expectError:           false,
		},
		{
			name:             "logging_user_error_variants",
			loggingUserError: "1",

			expectedDefaultRadius: DefaultRadius,
			expectedMaxRadius:     DefaultMaxRadius,
			expectedMaxRows:       DefaultMaxRowsInPage,
			expectedStatsTime:     DefaultStatsTime,
			expectedLoggingError:  true,
			expectError:           false,
		},
		{
			name:                  "logging_user_error_false_variants",
			loggingUserError:      "false",
			expectedDefaultRadius: DefaultRadius,
			expectedMaxRadius:     DefaultMaxRadius,
			expectedMaxRows:       DefaultMaxRowsInPage,
			expectedStatsTime:     DefaultStatsTime,
			expectedLoggingError:  false,
			expectError:           false,
		},
		{
			name:             "invalid_default_radius_not_integer",
			defaultRadiusEnv: "abc",

			expectError:           true,
			expectedErrorContains: "invalid DEFAULT_INCIDENT_RADIUS: not integer",
		},
		{
			name:             "invalid_default_radius_zero_or_negative",
			defaultRadiusEnv: "-500",

			expectError:           true,
			expectedErrorContains: "invalid DEFAULT_INCIDENT_RADIUS: <= 0",
		},
		{
			name:             "default_radius_greater_than_max_radius",
			defaultRadiusEnv: "60000",
			maxRadiusEnv:     "50000",

			expectError:           true,
			expectedErrorContains: "invalid DEFAULT_INCIDENT_RADIUS: value > MAX_INCIDENT_RADIUS",
		},
		{
			name:         "invalid_max_radius_not_integer",
			maxRadiusEnv: "notnumber",

			expectError:           true,
			expectedErrorContains: "invalid MAX_INCIDENT_RADIUS: not integer",
		},
		{
			name:         "invalid_max_radius_<0",
			maxRadiusEnv: "-100",

			expectError:           true,
			expectedErrorContains: "invalid MAX_INCIDENT_RADIUS: <=",
		},
		{
			name:       "invalid_max_rows_not_integer",
			maxRowsEnv: "no",

			expectError:           true,
			expectedErrorContains: "invalid MAX_ROWS_IN_PAGE",
		},
		{
			name:       "invalid_max_rows_negative",
			maxRowsEnv: "-10",

			expectError:           true,
			expectedErrorContains: "invalid MAX_ROWS_IN_PAGE: <= 0",
		},
		{
			name:         "invalid_stats_time_not_integer",
			statsTimeEnv: "thirty",

			expectError:           true,
			expectedErrorContains: "invalid STATS_TIME_WINDOW_MINUTES: not integer",
		},
		{
			name:         "invalid_stats_time_zero",
			statsTimeEnv: "0",

			expectError:           true,
			expectedErrorContains: "invalid STATS_TIME_WINDOW_MINUTES: <= 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setEnv := func(name, value string) {
				if value != "" {
					os.Setenv(name, value)
				} else {
					os.Unsetenv(name)
				}
			}

			setEnv(EnvNameDefaultIncidentRadius, tc.defaultRadiusEnv)
			setEnv(EnvNameMaxIncidentRadius, tc.maxRadiusEnv)
			setEnv(EnvMaxRowsInPage, tc.maxRowsEnv)
			setEnv(EnvNameStatsTime, tc.statsTimeEnv)
			setEnv(EnvNameLoggingUserError, tc.loggingUserError)

			os.Setenv(EnvNameDbName, "testdb")
			os.Setenv(EnvNameDbUser, "user")
			os.Setenv(EnvNameDbPassword, "pass")

			cfg, err := NewConfig(false)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tc.expectedErrorContains) {
					t.Errorf("error message\ngot:  %s\nwant: %s", err.Error(), tc.expectedErrorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if cfg == nil {
				t.Fatal("cfg is nil")
			}

			if cfg.DefaultRadius != tc.expectedDefaultRadius {
				t.Errorf("DefaultRadius: got %d, want %d", cfg.DefaultRadius, tc.expectedDefaultRadius)
			}
			if cfg.MaxRadius != tc.expectedMaxRadius {
				t.Errorf("MaxRadius: got %d, want %d", cfg.MaxRadius, tc.expectedMaxRadius)
			}
			if cfg.MaxRowsInPage != tc.expectedMaxRows {
				t.Errorf("MaxRowsInPage: got %d, want %d", cfg.MaxRowsInPage, tc.expectedMaxRows)
			}
			if cfg.StatsTimeWindow != tc.expectedStatsTime {
				t.Errorf("StatsTimeWindow: got %d, want %d", cfg.StatsTimeWindow, tc.expectedStatsTime)
			}
			if cfg.loggingUserError != tc.expectedLoggingError {
				t.Errorf("loggingUserError: got %v, want %v", cfg.loggingUserError, tc.expectedLoggingError)
			}
		})
	}
}

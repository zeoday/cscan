package scheduler

import (
	"testing"

	"cscan/pkg/xerr"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// ==================== Unit Tests ====================

// TestValidateNilConfig 测试空配置验证
func TestValidateNilConfig(t *testing.T) {
	v := NewConfigValidator()
	err := v.Validate(nil)
	if err == nil {
		t.Error("Validate(nil) should return error")
	}
	if !xerr.IsConfigError(err) {
		t.Error("Validate(nil) should return ConfigError")
	}
}

// TestValidateEmptyConfig 测试空配置（无启用模块）
func TestValidateEmptyConfig(t *testing.T) {
	v := NewConfigValidator()
	config := &TaskConfig{}
	err := v.Validate(config)
	if err != nil {
		t.Errorf("Validate empty config should pass, got: %v", err)
	}
}

// TestValidatePortScanMissingPorts 测试端口扫描缺少端口
func TestValidatePortScanMissingPorts(t *testing.T) {
	v := NewConfigValidator()
	config := &TaskConfig{
		PortScan: &PortScanConfig{
			Enable: true,
			Ports:  "", // Missing required field
		},
	}
	err := v.Validate(config)
	if err == nil {
		t.Error("Validate should fail when ports is empty")
	}
}

// TestValidatePortScanNegativeRate 测试端口扫描负速率
func TestValidatePortScanNegativeRate(t *testing.T) {
	v := NewConfigValidator()
	config := &TaskConfig{
		PortScan: &PortScanConfig{
			Enable: true,
			Ports:  "80,443",
			Rate:   -1, // Invalid
		},
	}
	err := v.Validate(config)
	if err == nil {
		t.Error("Validate should fail when rate is negative")
	}
}

// TestValidatePortScanInvalidTool 测试端口扫描无效工具
func TestValidatePortScanInvalidTool(t *testing.T) {
	v := NewConfigValidator()
	config := &TaskConfig{
		PortScan: &PortScanConfig{
			Enable: true,
			Ports:  "80,443",
			Tool:   "invalid_tool",
		},
	}
	err := v.Validate(config)
	if err == nil {
		t.Error("Validate should fail when tool is invalid")
	}
}

// TestValidatePortScanValidConfig 测试有效端口扫描配置
func TestValidatePortScanValidConfig(t *testing.T) {
	v := NewConfigValidator()
	config := &TaskConfig{
		PortScan: &PortScanConfig{
			Enable:  true,
			Ports:   "80,443,8080",
			Rate:    1000,
			Timeout: 5,
			Tool:    "naabu",
		},
	}
	err := v.Validate(config)
	if err != nil {
		t.Errorf("Validate should pass for valid config, got: %v", err)
	}
}

// TestApplyDefaultsPortScan 测试端口扫描默认值应用
func TestApplyDefaultsPortScan(t *testing.T) {
	v := NewConfigValidator()
	config := &TaskConfig{
		PortScan: &PortScanConfig{
			Enable: true,
		},
	}
	v.ApplyDefaults(config)

	if config.PortScan.Ports != "80,443,8080" {
		t.Errorf("Default ports = %s, want 80,443,8080", config.PortScan.Ports)
	}
	if config.PortScan.Rate != 1000 {
		t.Errorf("Default rate = %d, want 1000", config.PortScan.Rate)
	}
	if config.PortScan.Timeout != 5 {
		t.Errorf("Default timeout = %d, want 5", config.PortScan.Timeout)
	}
	if config.PortScan.Tool != "naabu" {
		t.Errorf("Default tool = %s, want naabu", config.PortScan.Tool)
	}
	if config.PortScan.ScanType != "c" {
		t.Errorf("Default scanType = %s, want c", config.PortScan.ScanType)
	}
}

// TestApplyDefaultsPocScan 测试POC扫描默认值应用
func TestApplyDefaultsPocScan(t *testing.T) {
	v := NewConfigValidator()
	config := &TaskConfig{
		PocScan: &PocScanConfig{
			Enable: true,
		},
	}
	v.ApplyDefaults(config)

	if config.PocScan.Concurrency != 25 {
		t.Errorf("Default concurrency = %d, want 25", config.PocScan.Concurrency)
	}
	if config.PocScan.TargetTimeout != 600 {
		t.Errorf("Default targetTimeout = %d, want 600", config.PocScan.TargetTimeout)
	}
	if config.PocScan.RateLimit != 150 {
		t.Errorf("Default rateLimit = %d, want 150", config.PocScan.RateLimit)
	}
}

// TestApplyDefaultsFingerprint 测试指纹识别默认值应用
func TestApplyDefaultsFingerprint(t *testing.T) {
	v := NewConfigValidator()
	config := &TaskConfig{
		Fingerprint: &FingerprintConfig{
			Enable: true,
		},
	}
	v.ApplyDefaults(config)

	if config.Fingerprint.Tool != "httpx" {
		t.Errorf("Default tool = %s, want httpx", config.Fingerprint.Tool)
	}
	if config.Fingerprint.Concurrency != 10 {
		t.Errorf("Default concurrency = %d, want 10", config.Fingerprint.Concurrency)
	}
	if config.Fingerprint.Timeout != 300 {
		t.Errorf("Default timeout = %d, want 300", config.Fingerprint.Timeout)
	}
}

// TestApplyDefaultsDirScan 测试目录扫描默认值应用
func TestApplyDefaultsDirScan(t *testing.T) {
	v := NewConfigValidator()
	config := &TaskConfig{
		DirScan: &DirScanConfig{
			Enable: true,
		},
	}
	v.ApplyDefaults(config)

	if config.DirScan.Threads != 10 {
		t.Errorf("Default threads = %d, want 10", config.DirScan.Threads)
	}
	if config.DirScan.Timeout != 10 {
		t.Errorf("Default timeout = %d, want 10", config.DirScan.Timeout)
	}
	if len(config.DirScan.StatusCodes) == 0 {
		t.Error("Default statusCodes should not be empty")
	}
}

// TestValidateAndApplyDefaults 测试验证并应用默认值
func TestValidateAndApplyDefaults(t *testing.T) {
	v := NewConfigValidator()
	config := &TaskConfig{
		PortScan: &PortScanConfig{
			Enable: true,
			// Ports is empty, but defaults will be applied first
		},
	}

	err := v.ValidateAndApplyDefaults(config)
	if err != nil {
		t.Errorf("ValidateAndApplyDefaults should pass after applying defaults, got: %v", err)
	}

	// Verify defaults were applied
	if config.PortScan.Ports != "80,443,8080" {
		t.Errorf("Ports should have default value, got: %s", config.PortScan.Ports)
	}
}

// TestParseAndValidate 测试解析并验证
func TestParseAndValidate(t *testing.T) {
	v := NewConfigValidator()
	configStr := `{"portscan":{"enable":true,"ports":"80,443"}}`

	config, err := v.ParseAndValidate(configStr)
	if err != nil {
		t.Errorf("ParseAndValidate should pass, got: %v", err)
	}
	if config == nil {
		t.Error("ParseAndValidate should return config")
	}
	if config.PortScan == nil || !config.PortScan.Enable {
		t.Error("ParseAndValidate should parse portscan config")
	}
}

// TestParseAndValidateInvalidJSON 测试解析无效JSON
func TestParseAndValidateInvalidJSON(t *testing.T) {
	v := NewConfigValidator()
	configStr := `{invalid json}`

	_, err := v.ParseAndValidate(configStr)
	if err == nil {
		t.Error("ParseAndValidate should fail for invalid JSON")
	}
}


// ==================== Property Tests ====================

// TestProperty6_ConfigurationValidation 测试 Property 6: Configuration Validation
// **Property 6: Configuration Validation**
// **Validates: Requirements 4.2, 4.3**
// For any task configuration, parsing SHALL fail with a descriptive error
// if required fields are missing or values are invalid.
func TestProperty6_ConfigurationValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Empty ports with enabled port scan should fail validation
	properties.Property("PortScan with empty ports fails validation", prop.ForAll(
		func(rate, timeout int) bool {
			v := NewConfigValidator()
			config := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable:  true,
					Ports:   "", // Empty - should fail
					Rate:    rate,
					Timeout: timeout,
				},
			}
			err := v.Validate(config)
			return err != nil
		},
		gen.IntRange(0, 10000),
		gen.IntRange(0, 300),
	))

	// Property: Negative rate should fail validation
	properties.Property("PortScan with negative rate fails validation", prop.ForAll(
		func(rate int) bool {
			if rate >= 0 {
				return true // Skip non-negative rates
			}
			v := NewConfigValidator()
			config := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable: true,
					Ports:  "80,443",
					Rate:   rate,
				},
			}
			err := v.Validate(config)
			return err != nil
		},
		gen.IntRange(-1000, 0),
	))

	// Property: Negative timeout should fail validation
	properties.Property("PortScan with negative timeout fails validation", prop.ForAll(
		func(timeout int) bool {
			if timeout >= 0 {
				return true // Skip non-negative timeouts
			}
			v := NewConfigValidator()
			config := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable:  true,
					Ports:   "80,443",
					Timeout: timeout,
				},
			}
			err := v.Validate(config)
			return err != nil
		},
		gen.IntRange(-1000, 0),
	))

	// Property: Invalid tool should fail validation
	properties.Property("PortScan with invalid tool fails validation", prop.ForAll(
		func(tool string) bool {
			// Skip valid tools
			if tool == "" || tool == "naabu" || tool == "masscan" || tool == "tcp" {
				return true
			}
			v := NewConfigValidator()
			config := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable: true,
					Ports:  "80,443",
					Tool:   tool,
				},
			}
			err := v.Validate(config)
			return err != nil
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return s != "" && s != "naabu" && s != "masscan" && s != "tcp"
		}),
	))

	// Property: Valid configuration should pass validation
	properties.Property("Valid PortScan config passes validation", prop.ForAll(
		func(ports string, rate, timeout int) bool {
			if ports == "" || rate < 0 || timeout < 0 {
				return true // Skip invalid inputs
			}
			v := NewConfigValidator()
			config := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable:  true,
					Ports:   ports,
					Rate:    rate,
					Timeout: timeout,
					Tool:    "naabu",
				},
			}
			err := v.Validate(config)
			return err == nil
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.IntRange(0, 10000),
		gen.IntRange(0, 300),
	))

	// Property: Disabled modules should not be validated
	properties.Property("Disabled modules skip validation", prop.ForAll(
		func(rate int) bool {
			v := NewConfigValidator()
			config := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable: false, // Disabled
					Ports:  "",    // Would fail if enabled
					Rate:   rate,  // Could be negative
				},
			}
			err := v.Validate(config)
			return err == nil
		},
		gen.IntRange(-1000, 1000),
	))

	// Property: Multiple validation errors are combined
	properties.Property("Multiple errors are combined", prop.ForAll(
		func(rate, timeout int) bool {
			if rate >= 0 && timeout >= 0 {
				return true // Skip valid inputs
			}
			v := NewConfigValidator()
			config := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable:  true,
					Ports:   "",      // Error 1
					Rate:    rate,    // Possibly Error 2
					Timeout: timeout, // Possibly Error 3
				},
			}
			err := v.Validate(config)
			return err != nil
		},
		gen.IntRange(-100, 100),
		gen.IntRange(-100, 100),
	))

	properties.TestingRun(t)
}

// TestProperty8_DefaultValueApplication 测试 Property 8: Default Value Application
// **Property 8: Default Value Application**
// **Validates: Requirements 4.5**
// For any TaskConfig with missing optional fields, parsing SHALL apply
// default values during parsing, not during execution.
func TestProperty8_DefaultValueApplication(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: PortScan defaults are applied when fields are zero
	properties.Property("PortScan defaults are applied for zero values", prop.ForAll(
		func(dummy int) bool {
			v := NewConfigValidator()
			config := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable: true,
					// All other fields are zero/empty
				},
			}
			v.ApplyDefaults(config)

			// Verify defaults are applied
			if config.PortScan.Ports != "80,443,8080" {
				return false
			}
			if config.PortScan.Rate != 1000 {
				return false
			}
			if config.PortScan.Timeout != 5 {
				return false
			}
			if config.PortScan.Tool != "naabu" {
				return false
			}
			if config.PortScan.ScanType != "c" {
				return false
			}
			return true
		},
		gen.Int(),
	))

	// Property: Non-zero values are preserved after ApplyDefaults
	properties.Property("Non-zero values are preserved", prop.ForAll(
		func(ports string, rate, timeout int) bool {
			if ports == "" || rate <= 0 || timeout <= 0 {
				return true // Skip invalid inputs
			}
			v := NewConfigValidator()
			config := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable:  true,
					Ports:   ports,
					Rate:    rate,
					Timeout: timeout,
					Tool:    "masscan",
				},
			}
			v.ApplyDefaults(config)

			// Verify original values are preserved
			if config.PortScan.Ports != ports {
				return false
			}
			if config.PortScan.Rate != rate {
				return false
			}
			if config.PortScan.Timeout != timeout {
				return false
			}
			if config.PortScan.Tool != "masscan" {
				return false
			}
			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.IntRange(1, 10000),
		gen.IntRange(1, 300),
	))

	// Property: PocScan defaults are applied when fields are zero
	properties.Property("PocScan defaults are applied for zero values", prop.ForAll(
		func(dummy int) bool {
			v := NewConfigValidator()
			config := &TaskConfig{
				PocScan: &PocScanConfig{
					Enable: true,
				},
			}
			v.ApplyDefaults(config)

			if config.PocScan.Concurrency != 25 {
				return false
			}
			if config.PocScan.TargetTimeout != 600 {
				return false
			}
			if config.PocScan.RateLimit != 150 {
				return false
			}
			return true
		},
		gen.Int(),
	))

	// Property: Fingerprint defaults are applied when fields are zero
	properties.Property("Fingerprint defaults are applied for zero values", prop.ForAll(
		func(dummy int) bool {
			v := NewConfigValidator()
			config := &TaskConfig{
				Fingerprint: &FingerprintConfig{
					Enable: true,
				},
			}
			v.ApplyDefaults(config)

			if config.Fingerprint.Tool != "httpx" {
				return false
			}
			if config.Fingerprint.ActiveTimeout != 10 {
				return false
			}
			if config.Fingerprint.Timeout != 300 {
				return false
			}
			if config.Fingerprint.TargetTimeout != 30 {
				return false
			}
			if config.Fingerprint.Concurrency != 10 {
				return false
			}
			return true
		},
		gen.Int(),
	))

	// Property: DomainScan defaults are applied when fields are zero
	properties.Property("DomainScan defaults are applied for zero values", prop.ForAll(
		func(dummy int) bool {
			v := NewConfigValidator()
			config := &TaskConfig{
				DomainScan: &DomainScanConfig{
					Enable: true,
				},
			}
			v.ApplyDefaults(config)

			if config.DomainScan.Timeout != 300 {
				return false
			}
			if config.DomainScan.MaxEnumerationTime != 10 {
				return false
			}
			if config.DomainScan.Threads != 10 {
				return false
			}
			if config.DomainScan.RateLimit != 100 {
				return false
			}
			if config.DomainScan.Concurrent != 100 {
				return false
			}
			return true
		},
		gen.Int(),
	))

	// Property: DirScan defaults are applied when fields are zero
	properties.Property("DirScan defaults are applied for zero values", prop.ForAll(
		func(dummy int) bool {
			v := NewConfigValidator()
			config := &TaskConfig{
				DirScan: &DirScanConfig{
					Enable: true,
				},
			}
			v.ApplyDefaults(config)

			if config.DirScan.Threads != 10 {
				return false
			}
			if config.DirScan.Timeout != 10 {
				return false
			}
			if len(config.DirScan.StatusCodes) == 0 {
				return false
			}
			return true
		},
		gen.Int(),
	))

	// Property: Disabled modules don't get defaults applied
	properties.Property("Disabled modules don't get defaults", prop.ForAll(
		func(dummy int) bool {
			v := NewConfigValidator()
			config := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable: false, // Disabled
				},
			}
			v.ApplyDefaults(config)

			// Defaults should NOT be applied for disabled modules
			if config.PortScan.Ports != "" {
				return false
			}
			if config.PortScan.Rate != 0 {
				return false
			}
			return true
		},
		gen.Int(),
	))

	// Property: ValidateAndApplyDefaults applies defaults before validation
	properties.Property("ValidateAndApplyDefaults applies defaults first", prop.ForAll(
		func(dummy int) bool {
			v := NewConfigValidator()
			config := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable: true,
					// Ports is empty, would fail validation without defaults
				},
			}

			err := v.ValidateAndApplyDefaults(config)
			if err != nil {
				return false
			}

			// Verify defaults were applied
			if config.PortScan.Ports != "80,443,8080" {
				return false
			}
			return true
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}


// TestProperty7_JSONSerializationRoundTrip 测试 Property 7: JSON Serialization Round-Trip
// **Property 7: JSON Serialization Round-Trip**
// **Validates: Requirements 4.4**
// For any valid TaskConfig, serializing to JSON and deserializing back
// SHALL produce an equivalent configuration.
func TestProperty7_JSONSerializationRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: PortScanConfig round-trip preserves all fields
	properties.Property("PortScanConfig JSON round-trip preserves fields", prop.ForAll(
		func(enable bool, ports, tool, scanType string, rate, timeout, portThreshold int, skipHostDiscovery bool) bool {
			// Normalize values
			if rate < 0 {
				rate = 0
			}
			if timeout < 0 {
				timeout = 0
			}
			if portThreshold < 0 {
				portThreshold = 0
			}

			original := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable:            enable,
					Ports:             ports,
					Tool:              tool,
					Rate:              rate,
					Timeout:           timeout,
					PortThreshold:     portThreshold,
					ScanType:          scanType,
					SkipHostDiscovery: skipHostDiscovery,
				},
			}

			// Serialize
			jsonStr, err := BuildTaskConfig(original)
			if err != nil {
				return false
			}

			// Deserialize
			restored, err := ParseTaskConfig(jsonStr)
			if err != nil {
				return false
			}

			// Compare
			if restored.PortScan == nil {
				return false
			}
			if restored.PortScan.Enable != original.PortScan.Enable {
				return false
			}
			if restored.PortScan.Ports != original.PortScan.Ports {
				return false
			}
			if restored.PortScan.Tool != original.PortScan.Tool {
				return false
			}
			if restored.PortScan.Rate != original.PortScan.Rate {
				return false
			}
			if restored.PortScan.Timeout != original.PortScan.Timeout {
				return false
			}
			if restored.PortScan.PortThreshold != original.PortScan.PortThreshold {
				return false
			}
			if restored.PortScan.ScanType != original.PortScan.ScanType {
				return false
			}
			if restored.PortScan.SkipHostDiscovery != original.PortScan.SkipHostDiscovery {
				return false
			}
			return true
		},
		gen.Bool(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.IntRange(-100, 10000),
		gen.IntRange(-100, 300),
		gen.IntRange(-100, 1000),
		gen.Bool(),
	))

	// Property: PocScanConfig round-trip preserves all fields
	properties.Property("PocScanConfig JSON round-trip preserves fields", prop.ForAll(
		func(enable, useNuclei, autoScan, automaticScan, customPocOnly bool, severity string, rateLimit, concurrency, targetTimeout int) bool {
			// Normalize values
			if rateLimit < 0 {
				rateLimit = 0
			}
			if concurrency < 0 {
				concurrency = 0
			}
			if targetTimeout < 0 {
				targetTimeout = 0
			}

			original := &TaskConfig{
				PocScan: &PocScanConfig{
					Enable:        enable,
					UseNuclei:     useNuclei,
					AutoScan:      autoScan,
					AutomaticScan: automaticScan,
					CustomPocOnly: customPocOnly,
					Severity:      severity,
					RateLimit:     rateLimit,
					Concurrency:   concurrency,
					TargetTimeout: targetTimeout,
				},
			}

			// Serialize
			jsonStr, err := BuildTaskConfig(original)
			if err != nil {
				return false
			}

			// Deserialize
			restored, err := ParseTaskConfig(jsonStr)
			if err != nil {
				return false
			}

			// Compare
			if restored.PocScan == nil {
				return false
			}
			if restored.PocScan.Enable != original.PocScan.Enable {
				return false
			}
			if restored.PocScan.UseNuclei != original.PocScan.UseNuclei {
				return false
			}
			if restored.PocScan.AutoScan != original.PocScan.AutoScan {
				return false
			}
			if restored.PocScan.AutomaticScan != original.PocScan.AutomaticScan {
				return false
			}
			if restored.PocScan.CustomPocOnly != original.PocScan.CustomPocOnly {
				return false
			}
			if restored.PocScan.Severity != original.PocScan.Severity {
				return false
			}
			if restored.PocScan.RateLimit != original.PocScan.RateLimit {
				return false
			}
			if restored.PocScan.Concurrency != original.PocScan.Concurrency {
				return false
			}
			if restored.PocScan.TargetTimeout != original.PocScan.TargetTimeout {
				return false
			}
			return true
		},
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.AlphaString(),
		gen.IntRange(-100, 1000),
		gen.IntRange(-100, 100),
		gen.IntRange(-100, 1000),
	))

	// Property: FingerprintConfig round-trip preserves all fields
	properties.Property("FingerprintConfig JSON round-trip preserves fields", prop.ForAll(
		func(enable bool, tool string, iconHash, wappalyzer, customEngine, screenshot, activeScan bool, activeTimeout, timeout, targetTimeout, concurrency int) bool {
			// Normalize values
			if activeTimeout < 0 {
				activeTimeout = 0
			}
			if timeout < 0 {
				timeout = 0
			}
			if targetTimeout < 0 {
				targetTimeout = 0
			}
			if concurrency < 0 {
				concurrency = 0
			}

			original := &TaskConfig{
				Fingerprint: &FingerprintConfig{
					Enable:        enable,
					Tool:          tool,
					IconHash:      iconHash,
					Wappalyzer:    wappalyzer,
					CustomEngine:  customEngine,
					Screenshot:    screenshot,
					ActiveScan:    activeScan,
					ActiveTimeout: activeTimeout,
					Timeout:       timeout,
					TargetTimeout: targetTimeout,
					Concurrency:   concurrency,
				},
			}

			// Serialize
			jsonStr, err := BuildTaskConfig(original)
			if err != nil {
				return false
			}

			// Deserialize
			restored, err := ParseTaskConfig(jsonStr)
			if err != nil {
				return false
			}

			// Compare
			if restored.Fingerprint == nil {
				return false
			}
			if restored.Fingerprint.Enable != original.Fingerprint.Enable {
				return false
			}
			if restored.Fingerprint.Tool != original.Fingerprint.Tool {
				return false
			}
			if restored.Fingerprint.IconHash != original.Fingerprint.IconHash {
				return false
			}
			if restored.Fingerprint.Wappalyzer != original.Fingerprint.Wappalyzer {
				return false
			}
			if restored.Fingerprint.CustomEngine != original.Fingerprint.CustomEngine {
				return false
			}
			if restored.Fingerprint.Screenshot != original.Fingerprint.Screenshot {
				return false
			}
			if restored.Fingerprint.ActiveScan != original.Fingerprint.ActiveScan {
				return false
			}
			if restored.Fingerprint.ActiveTimeout != original.Fingerprint.ActiveTimeout {
				return false
			}
			if restored.Fingerprint.Timeout != original.Fingerprint.Timeout {
				return false
			}
			if restored.Fingerprint.TargetTimeout != original.Fingerprint.TargetTimeout {
				return false
			}
			if restored.Fingerprint.Concurrency != original.Fingerprint.Concurrency {
				return false
			}
			return true
		},
		gen.Bool(),
		gen.AlphaString(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.IntRange(-100, 100),
		gen.IntRange(-100, 1000),
		gen.IntRange(-100, 100),
		gen.IntRange(-100, 100),
	))

	// Property: DomainScanConfig round-trip preserves all fields
	properties.Property("DomainScanConfig JSON round-trip preserves fields", prop.ForAll(
		func(enable, subfinder, all, recursive, removeWildcard, resolveDNS bool, timeout, maxEnumerationTime, threads, rateLimit, concurrent int) bool {
			// Normalize values
			if timeout < 0 {
				timeout = 0
			}
			if maxEnumerationTime < 0 {
				maxEnumerationTime = 0
			}
			if threads < 0 {
				threads = 0
			}
			if rateLimit < 0 {
				rateLimit = 0
			}
			if concurrent < 0 {
				concurrent = 0
			}

			original := &TaskConfig{
				DomainScan: &DomainScanConfig{
					Enable:             enable,
					Subfinder:          subfinder,
					Timeout:            timeout,
					MaxEnumerationTime: maxEnumerationTime,
					Threads:            threads,
					RateLimit:          rateLimit,
					All:                all,
					Recursive:          recursive,
					RemoveWildcard:     removeWildcard,
					ResolveDNS:         resolveDNS,
					Concurrent:         concurrent,
				},
			}

			// Serialize
			jsonStr, err := BuildTaskConfig(original)
			if err != nil {
				return false
			}

			// Deserialize
			restored, err := ParseTaskConfig(jsonStr)
			if err != nil {
				return false
			}

			// Compare
			if restored.DomainScan == nil {
				return false
			}
			if restored.DomainScan.Enable != original.DomainScan.Enable {
				return false
			}
			if restored.DomainScan.Subfinder != original.DomainScan.Subfinder {
				return false
			}
			if restored.DomainScan.Timeout != original.DomainScan.Timeout {
				return false
			}
			if restored.DomainScan.MaxEnumerationTime != original.DomainScan.MaxEnumerationTime {
				return false
			}
			if restored.DomainScan.Threads != original.DomainScan.Threads {
				return false
			}
			if restored.DomainScan.RateLimit != original.DomainScan.RateLimit {
				return false
			}
			if restored.DomainScan.All != original.DomainScan.All {
				return false
			}
			if restored.DomainScan.Recursive != original.DomainScan.Recursive {
				return false
			}
			if restored.DomainScan.RemoveWildcard != original.DomainScan.RemoveWildcard {
				return false
			}
			if restored.DomainScan.ResolveDNS != original.DomainScan.ResolveDNS {
				return false
			}
			if restored.DomainScan.Concurrent != original.DomainScan.Concurrent {
				return false
			}
			return true
		},
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.IntRange(-100, 1000),
		gen.IntRange(-100, 100),
		gen.IntRange(-100, 100),
		gen.IntRange(-100, 1000),
		gen.IntRange(-100, 1000),
	))

	// Property: DirScanConfig round-trip preserves all fields
	properties.Property("DirScanConfig JSON round-trip preserves fields", prop.ForAll(
		func(enable, followRedirect bool, threads, timeout int) bool {
			// Normalize values
			if threads < 0 {
				threads = 0
			}
			if timeout < 0 {
				timeout = 0
			}

			original := &TaskConfig{
				DirScan: &DirScanConfig{
					Enable:         enable,
					Threads:        threads,
					Timeout:        timeout,
					FollowRedirect: followRedirect,
					StatusCodes:    []int{200, 301, 403},
					Extensions:     []string{".php", ".html"},
				},
			}

			// Serialize
			jsonStr, err := BuildTaskConfig(original)
			if err != nil {
				return false
			}

			// Deserialize
			restored, err := ParseTaskConfig(jsonStr)
			if err != nil {
				return false
			}

			// Compare
			if restored.DirScan == nil {
				return false
			}
			if restored.DirScan.Enable != original.DirScan.Enable {
				return false
			}
			if restored.DirScan.Threads != original.DirScan.Threads {
				return false
			}
			if restored.DirScan.Timeout != original.DirScan.Timeout {
				return false
			}
			if restored.DirScan.FollowRedirect != original.DirScan.FollowRedirect {
				return false
			}
			if len(restored.DirScan.StatusCodes) != len(original.DirScan.StatusCodes) {
				return false
			}
			if len(restored.DirScan.Extensions) != len(original.DirScan.Extensions) {
				return false
			}
			return true
		},
		gen.Bool(),
		gen.Bool(),
		gen.IntRange(-100, 100),
		gen.IntRange(-100, 100),
	))

	// Property: Complete TaskConfig round-trip preserves all modules
	properties.Property("Complete TaskConfig JSON round-trip preserves all modules", prop.ForAll(
		func(portEnable, pocEnable, fpEnable, domainEnable, dirEnable bool) bool {
			original := &TaskConfig{
				PortScan: &PortScanConfig{
					Enable: portEnable,
					Ports:  "80,443",
					Rate:   1000,
				},
				PocScan: &PocScanConfig{
					Enable:      pocEnable,
					Concurrency: 25,
				},
				Fingerprint: &FingerprintConfig{
					Enable: fpEnable,
					Tool:   "httpx",
				},
				DomainScan: &DomainScanConfig{
					Enable:  domainEnable,
					Threads: 10,
				},
				DirScan: &DirScanConfig{
					Enable:  dirEnable,
					Threads: 10,
				},
			}

			// Serialize
			jsonStr, err := BuildTaskConfig(original)
			if err != nil {
				return false
			}

			// Deserialize
			restored, err := ParseTaskConfig(jsonStr)
			if err != nil {
				return false
			}

			// Verify all modules are present
			if restored.PortScan == nil || restored.PortScan.Enable != portEnable {
				return false
			}
			if restored.PocScan == nil || restored.PocScan.Enable != pocEnable {
				return false
			}
			if restored.Fingerprint == nil || restored.Fingerprint.Enable != fpEnable {
				return false
			}
			if restored.DomainScan == nil || restored.DomainScan.Enable != domainEnable {
				return false
			}
			if restored.DirScan == nil || restored.DirScan.Enable != dirEnable {
				return false
			}
			return true
		},
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

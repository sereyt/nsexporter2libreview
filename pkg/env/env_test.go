package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigMapsLibreViewDeviceSettings(t *testing.T) {
	t.Setenv("LV_DEVICE_ID", "device-123")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	config := []byte(`
libreview:
  auth:
    username: user@example.com
    password: secret
  importConfig:
    apiEndpoint: https://api.libreview.ru
    culture: ru-RU
    deviceSettings:
      firmwareVersion: 2.8.2
      glucoseTargetRangeHighInMgPerDl: 144
      glucoseTargetRangeLowInMgPerDl: 90
      hardwareDescriptor: Redmi Note 8 Pro
      hardwareName: Xiaomi
      modelName: com.freestylelibre.app.ru
      osType: Android
      osVersion: "29"
      selectedCarbType: grams of carbs
      selectedLanguage: ru_RU
      selectedTimeFormat: 24hr
      uniqueIdentifier: ${LV_DEVICE_ID}
    domain: Libreview
    gatewayType: FSLibreLink.Android
    uom: mmol/L
nightscout:
  url: https://nscout2.don82.ru
`)
	if err := os.WriteFile(configPath, config, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	settings := New()
	settings.ConfigPath = configPath
	if err := settings.LoadConfig(); err != nil {
		t.Fatalf("load config: %v", err)
	}

	got := settings.Libreview().ImportConfig.DevSettings
	if got.UniqueIdentifier != "device-123" {
		t.Errorf("uniqueIdentifier = %q, want device-123", got.UniqueIdentifier)
	}
	if got.FirmwareVersion != "2.8.2" {
		t.Errorf("firmwareVersion = %q, want 2.8.2", got.FirmwareVersion)
	}
	if got.HardwareDescriptor != "Redmi Note 8 Pro" {
		t.Errorf("hardwareDescriptor = %q, want Redmi Note 8 Pro", got.HardwareDescriptor)
	}
	if got.ModelName != "com.freestylelibre.app.ru" {
		t.Errorf("modelName = %q, want com.freestylelibre.app.ru", got.ModelName)
	}
	if got.GlucoseTargetRangeLowInMgPerDl != 90 {
		t.Errorf("glucoseTargetRangeLowInMgPerDl = %d, want 90", got.GlucoseTargetRangeLowInMgPerDl)
	}
	if got.GlucoseTargetRangeHighInMgPerDl != 144 {
		t.Errorf("glucoseTargetRangeHighInMgPerDl = %d, want 144", got.GlucoseTargetRangeHighInMgPerDl)
	}
}

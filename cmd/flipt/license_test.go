package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateLicenseKey(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		licenseType string
		wantErr     string
	}{
		// Empty key
		{
			name:        "empty key monthly",
			key:         "",
			licenseType: LicenseTypeProMonthly,
			wantErr:     "license key is required",
		},
		{
			name:        "empty key annual",
			key:         "",
			licenseType: LicenseTypeProAnnual,
			wantErr:     "license key is required",
		},

		// Valid monthly keys
		{
			name:        "valid monthly key",
			key:         "AABBCC-DDEEFF-112233-445566-778899-V3",
			licenseType: LicenseTypeProMonthly,
		},
		{
			name:        "valid monthly key without version suffix",
			key:         "AAAAAA-BBBBBB-CCCCCC-DDDDDD-EEEEEE",
			licenseType: LicenseTypeProMonthly,
		},

		// Invalid monthly keys
		{
			name:        "monthly key too short",
			key:         "ABC-DEF",
			licenseType: LicenseTypeProMonthly,
			wantErr:     "license key appears too short",
		},
		{
			name:        "monthly key just under threshold",
			key:         "AAAA-BBBB-CCCC-DDDD",
			licenseType: LicenseTypeProMonthly,
			wantErr:     "license key appears too short",
		},

		// Valid annual keys
		{
			name:        "valid annual key",
			key:         "key/dGVzdC1wYXlsb2FkLWRhdGE=.dGVzdC1zaWduYXR1cmUtZGF0YQ==",
			licenseType: LicenseTypeProAnnual,
		},
		{
			name:        "valid annual key minimal",
			key:         "key/payload.signature",
			licenseType: LicenseTypeProAnnual,
		},

		// Invalid annual keys
		{
			name:        "annual key missing key/ prefix",
			key:         "dGVzdC1wYXlsb2FkLWRhdGE=.dGVzdC1zaWduYXR1cmU=",
			licenseType: LicenseTypeProAnnual,
			wantErr:     "annual license key should start with 'key/'",
		},
		{
			name:        "annual key missing dot separator",
			key:         "key/dGVzdC1wYXlsb2FkLWRhdGE=",
			licenseType: LicenseTypeProAnnual,
			wantErr:     "annual license key format appears invalid",
		},
		{
			name:        "annual key only prefix",
			key:         "key/",
			licenseType: LicenseTypeProAnnual,
			wantErr:     "annual license key format appears invalid",
		},
		{
			name:        "monthly key used as annual",
			key:         "AABBCC-DDEEFF-112233-445566-778899-V3",
			licenseType: LicenseTypeProAnnual,
			wantErr:     "annual license key should start with 'key/'",
		},

		// Cross-format: annual key used as monthly should pass length check
		{
			name:        "annual key used as monthly passes length check",
			key:         "key/dGVzdC1wYXlsb2Fk.c2lnbmF0dXJl",
			licenseType: LicenseTypeProMonthly,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLicenseKey(tt.key, tt.licenseType)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

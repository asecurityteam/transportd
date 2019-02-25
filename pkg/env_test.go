package transportd

import (
	"reflect"
	"testing"
)

func TestEnvProcessor_Process(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    string
		env     map[string]string
		wantErr bool
	}{
		{
			name:   "empty",
			source: `key: ${test}`,
			want:   `key: `,
			env:    make(map[string]string),
		},
		{
			name:   "ascii",
			source: `key: "${TEST}"`,
			want:   `key: "VALUE"`,
			env: map[string]string{
				"TEST": "VALUE",
			},
		},
		{
			name:   "ascii multiple",
			source: `key: "${TEST1}${TEST_2}_${TEST3}_STATIC"`,
			want:   `key: "VALUE1VALUE2_VALUE3_STATIC"`,
			env: map[string]string{
				"TEST1":  "VALUE1",
				"TEST_2": "VALUE2",
				"TEST3":  "VALUE3",
			},
		},
		{
			name:   "non-ascii",
			source: `key: "${☃☃☃}"`,
			want:   `key: "VALUE"`,
			env: map[string]string{
				"☃☃☃": "VALUE",
			},
		},
		{
			name:   "non-ascii multiple",
			source: `key: "${☃☃☃1}${☃☃☃_2}_${☃☃☃3}_STATIC"`,
			want:   `key: "VALUE1VALUE2_VALUE3_STATIC"`,
			env: map[string]string{
				"☃☃☃1":  "VALUE1",
				"☃☃☃_2": "VALUE2",
				"☃☃☃3":  "VALUE3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := NewEnvProcessor()
			y.osenv = func(k string) string {
				return tt.env[k]
			}
			got, err := y.Process([]byte(tt.source))
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvProcessor.Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("EnvProcessor.Process() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

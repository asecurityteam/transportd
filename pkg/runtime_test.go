package transportd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/settings"
)

func TestRTSourceFromExtension(t *testing.T) {
	rt := &runhttp.Component{}
	grp, _ := settings.GroupFromComponent(rt)
	tests := []struct {
		name    string
		s       string
		want    string
		wantErr bool
	}{
		{
			name:    "invalid source",
			s:       `not json`,
			want:    ``,
			wantErr: true,
		},
		{
			name:    "nests properties",
			s:       `{"key": "value"}`,
			want:    fmt.Sprintf(`{"%s":{"key":"value"}}`, grp.Name()),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RuntimeSourceFromExtension([]byte(tt.s))
			if (err != nil) != tt.wantErr {
				t.Errorf("rtSourceFromExtension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			wantM := make(map[string]interface{})
			_ = json.Unmarshal([]byte(tt.want), &wantM)
			want := settings.NewMapSource(wantM)
			if !tt.wantErr && !reflect.DeepEqual(got, want) {
				t.Errorf("rtSourceFromExtension() = %v, want %v", got, want)
			}
		})
	}
}

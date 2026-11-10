package rig_test

import (
	"testing"

	"github.com/jasonkwh/tensor-rig/internal/rig"
	"github.com/stretchr/testify/assert"
)

func TestParseCloud(t *testing.T) {
	tests := []struct {
		name    string
		want    rig.Cloud
		wantErr bool
	}{
		{
			name: "gcp",
			want: rig.CloudGCP,
		},
		{
			name: "aws",
			want: rig.CloudAWS,
		},
		{
			name:    "default",
			want:    rig.CloudUnknown,
			wantErr: true,
		},
		{
			name:    " ",
			want:    rig.CloudUnknown,
			wantErr: true,
		},
		{
			name:    "unknown",
			want:    rig.CloudUnknown,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)
			got, gotErr := rig.ParseCloud(tt.name)

			as.Equal(tt.want, got)
			as.Equal(tt.wantErr, gotErr != nil)
		})
	}
}

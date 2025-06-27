// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSlurmClient(t *testing.T) {
	type args struct {
		slurm_jwt string
		server    string
		cacheFreq time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				slurm_jwt: "token",
				server:    "http://localhost:6820",
				cacheFreq: time.Duration(30 * time.Second),
			},
		},
		{
			name: "no token",
			args: args{
				slurm_jwt: "",
				server:    "http://localhost:6820",
				cacheFreq: time.Duration(30 * time.Second),
			},
			wantErr: true,
		},
		{
			name: "bad server",
			args: args{
				slurm_jwt: "token",
				server:    "",
				cacheFreq: time.Duration(30 * time.Second),
			},
			wantErr: true,
		},
		{
			name: "bad cacheFreq",
			args: args{
				slurm_jwt: "token",
				server:    "http://localhost:6820",
				cacheFreq: time.Duration(0 * time.Second),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("SLURM_JWT", tt.args.slurm_jwt)
			if err != nil {
				t.Errorf("Environment could not be set. error=%v", err)
			}
			got, err := NewSlurmClient(tt.args.server, tt.args.cacheFreq)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSlurmClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.NotNil(t, got)
			}
			err = os.Unsetenv("SLURM_JWT")
			if err != nil {
				t.Errorf("Environment could not be unset. error=%v", err)
			}
		})
	}
}

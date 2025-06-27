// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"context"
	"testing"

	"github.com/SlinkyProject/slurm-client/pkg/client"
	"github.com/SlinkyProject/slurm-client/pkg/client/fake"
	"github.com/SlinkyProject/slurm-client/pkg/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func Test_getJobPendingNodeCount(t *testing.T) {
	type args struct {
		job types.V0043JobInfo
	}
	tests := []struct {
		name string
		args args
		want uint
	}{
		{
			name: "empty",
			args: args{
				job: types.V0043JobInfo{},
			},
			want: 0,
		},
		{
			name: "test job 1",
			args: args{
				job: *job1,
			},
			want: 0,
		},
		{
			name: "test job 3",
			args: args{
				job: *job3,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getJobPendingNodeCount(tt.args.job); got != tt.want {
				t.Errorf("getJobPendingNodeCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartitionCollector_getPartitionMetrics(t *testing.T) {
	type fields struct {
		slurmClient client.Client
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *PartitionMetrics
		wantErr bool
	}{
		{
			name: "empty",
			fields: fields{
				slurmClient: fake.NewFakeClient(),
			},
			args: args{
				ctx: context.TODO(),
			},
			want: &PartitionMetrics{
				NodeMetricsPer: map[string]*NodeMetrics{},
				JobMetricsPer:  map[string]*PartitionJobMetrics{},
			},
		},
		{
			name: "test data",
			fields: fields{
				slurmClient: testDataClient,
			},
			args: args{
				ctx: context.TODO(),
			},
			want: &PartitionMetrics{
				NodeMetricsPer: map[string]*NodeMetrics{
					partition1Name: {
						NodeCount: 3,
						NodeStates: NodeStates{
							Allocated: 2,
							Idle:      1,
							Drain:     1,
						},
						NodeTres: NodeTres{
							CpusTotal:       40,
							CpusEffective:   38,
							CpusAlloc:       24,
							CpusIdle:        16,
							MemoryTotal:     10240,
							MemoryEffective: 9216,
							MemoryAlloc:     5000,
							MemoryFree:      5240,
						},
					},
					partition2Name: {
						NodeCount: 3,
						NodeStates: NodeStates{
							total:      3,
							Allocated:  2,
							Mixed:      1,
							Completing: 1,
							Drain:      1,
						},
						NodeTres: NodeTres{
							total:           3,
							CpusTotal:       30,
							CpusEffective:   30,
							CpusAlloc:       28,
							CpusIdle:        2,
							MemoryTotal:     7168,
							MemoryEffective: 7168,
							MemoryAlloc:     5800,
							MemoryFree:      1368,
						},
					},
				},
				JobMetricsPer: map[string]*PartitionJobMetrics{
					partition1Name: {
						JobMetrics: JobMetrics{
							JobCount: 2,
							JobStates: JobStates{
								Pending: 1,
								Running: 1,
								Hold:    1,
							},
							JobTres: JobTres{
								CpusAlloc:   8,
								MemoryAlloc: 1024,
							},
						},
						PendingNodeCount: 0,
					},
					partition2Name: {
						JobMetrics: JobMetrics{
							JobCount: 3,
							JobStates: JobStates{
								Pending: 2,
								Running: 1,
								Hold:    1,
							},
							JobTres: JobTres{
								CpusAlloc:   12,
								MemoryAlloc: 3072,
							},
						},
						PendingNodeCount: 2,
					},
				},
			},
		},
		{
			name: "fail",
			fields: fields{
				slurmClient: testFailClient,
			},
			args: args{
				ctx: context.TODO(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &partitionCollector{
				slurmClient: tt.fields.slurmClient,
			}
			got, err := c.getPartitionMetrics(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("partitionCollector.getPartitionMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			opts := []cmp.Option{
				cmpopts.IgnoreUnexported(PartitionMetrics{}),
				cmpopts.IgnoreFields(JobStates{}, "total"),
				cmpopts.IgnoreFields(JobTres{}, "total"),
				cmpopts.IgnoreUnexported(NodeMetrics{}),
				cmpopts.IgnoreFields(NodeStates{}, "total"),
				cmpopts.IgnoreFields(NodeTres{}, "total"),
			}
			if diff := cmp.Diff(tt.want, got, opts...); diff != "" {
				t.Errorf("partitionCollector.getPartitionMetrics() = (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestPartitionCollector_Collect(t *testing.T) {
	type fields struct {
		slurmClient client.Client
	}
	type args struct {
		ch chan prometheus.Metric
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantNone bool
	}{
		{
			name: "empty",
			fields: fields{
				slurmClient: fake.NewFakeClient(),
			},
			args: args{
				ch: make(chan prometheus.Metric),
			},
		},
		{
			name: "data",
			fields: fields{
				slurmClient: testDataClient,
			},
			args: args{
				ch: make(chan prometheus.Metric),
			},
		},
		{
			name: "failure",
			fields: fields{
				slurmClient: testFailClient,
			},
			args: args{
				ch: make(chan prometheus.Metric),
			},
			wantNone: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewPartitionCollector(tt.fields.slurmClient)
			go func() {
				c.Collect(tt.args.ch)
				close(tt.args.ch)
			}()
			var got int
			for range tt.args.ch {
				got++
			}
			if !tt.wantNone {
				assert.GreaterOrEqual(t, got, 0)
			} else {
				assert.Equal(t, got, 0)
			}
		})
	}
}

func TestPartitionCollector_Describe(t *testing.T) {
	type fields struct {
		slurmClient client.Client
	}
	type args struct {
		ch chan *prometheus.Desc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test",
			fields: fields{
				slurmClient: fake.NewFakeClient(),
			},
			args: args{
				ch: make(chan *prometheus.Desc),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewPartitionCollector(tt.fields.slurmClient)
			go func() {
				c.Describe(tt.args.ch)
				close(tt.args.ch)
			}()
			var desc *prometheus.Desc
			for desc = range tt.args.ch {
				assert.NotNil(t, desc)
			}
		})
	}
}

// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"context"
	"testing"

	api "github.com/SlinkyProject/slurm-client/api/v0043"
	"github.com/SlinkyProject/slurm-client/pkg/client"
	"github.com/SlinkyProject/slurm-client/pkg/client/fake"
	"github.com/SlinkyProject/slurm-client/pkg/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func Test_calculateNodeState(t *testing.T) {
	type args struct {
		node types.V0043Node
	}
	tests := []struct {
		name string
		args args
		want *NodeStates
	}{
		{
			name: "empty",
			want: &NodeStates{},
		},
		{
			name: "allocated",
			args: args{
				node: types.V0043Node{V0043Node: api.V0043Node{
					State: ptr.To([]api.V0043NodeState{
						api.V0043NodeStateALLOCATED,
					}),
				}},
			},
			want: &NodeStates{Allocated: 1},
		},
		{
			name: "down",
			args: args{
				node: types.V0043Node{V0043Node: api.V0043Node{
					State: ptr.To([]api.V0043NodeState{
						api.V0043NodeStateDOWN,
					}),
				}},
			},
			want: &NodeStates{Down: 1},
		},
		{
			name: "error",
			args: args{
				node: types.V0043Node{V0043Node: api.V0043Node{
					State: ptr.To([]api.V0043NodeState{
						api.V0043NodeStateERROR,
					}),
				}},
			},
			want: &NodeStates{Error: 1},
		},
		{
			name: "future",
			args: args{
				node: types.V0043Node{V0043Node: api.V0043Node{
					State: ptr.To([]api.V0043NodeState{
						api.V0043NodeStateFUTURE,
					}),
				}},
			},
			want: &NodeStates{Future: 1},
		},
		{
			name: "idle",
			args: args{
				node: types.V0043Node{V0043Node: api.V0043Node{
					State: ptr.To([]api.V0043NodeState{
						api.V0043NodeStateIDLE,
					}),
				}},
			},
			want: &NodeStates{Idle: 1},
		},
		{
			name: "mixed",
			args: args{
				node: types.V0043Node{V0043Node: api.V0043Node{
					State: ptr.To([]api.V0043NodeState{
						api.V0043NodeStateMIXED,
					}),
				}},
			},
			want: &NodeStates{Mixed: 1},
		},
		{
			name: "unknown",
			args: args{
				node: types.V0043Node{V0043Node: api.V0043Node{
					State: ptr.To([]api.V0043NodeState{
						api.V0043NodeStateUNKNOWN,
					}),
				}},
			},
			want: &NodeStates{Unknown: 1},
		},
		{
			name: "all states, all flags",
			args: args{
				node: types.V0043Node{V0043Node: api.V0043Node{
					State: ptr.To([]api.V0043NodeState{
						api.V0043NodeStateALLOCATED,
						api.V0043NodeStateCLOUD,
						api.V0043NodeStateCOMPLETING,
						api.V0043NodeStateDOWN,
						api.V0043NodeStateDRAIN,
						api.V0043NodeStateDYNAMICFUTURE,
						api.V0043NodeStateDYNAMICNORM,
						api.V0043NodeStateERROR,
						api.V0043NodeStateFAIL,
						api.V0043NodeStateFUTURE,
						api.V0043NodeStateIDLE,
						api.V0043NodeStateINVALID,
						api.V0043NodeStateINVALIDREG,
						api.V0043NodeStateMAINTENANCE,
						api.V0043NodeStateMIXED,
						api.V0043NodeStateNOTRESPONDING,
						api.V0043NodeStatePLANNED,
						api.V0043NodeStatePOWERDOWN,
						api.V0043NodeStatePOWERDRAIN,
						api.V0043NodeStatePOWEREDDOWN,
						api.V0043NodeStatePOWERINGDOWN,
						api.V0043NodeStatePOWERINGUP,
						api.V0043NodeStatePOWERUP,
						api.V0043NodeStateREBOOTCANCELED,
						api.V0043NodeStateREBOOTISSUED,
						api.V0043NodeStateREBOOTREQUESTED,
						api.V0043NodeStateRESERVED,
						api.V0043NodeStateRESUME,
						api.V0043NodeStateUNDRAIN,
						api.V0043NodeStateUNKNOWN,
					}),
				}},
			},
			want: &NodeStates{
				Allocated:       1,
				Completing:      1,
				Drain:           1,
				Fail:            1,
				Maintenance:     1,
				NotResponding:   1,
				Planned:         1,
				RebootRequested: 1,
				Reserved:        1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &NodeStates{}
			calculateNodeState(metrics, tt.args.node)
			opts := []cmp.Option{
				cmpopts.IgnoreUnexported(NodeMetrics{}),
				cmpopts.IgnoreFields(NodeStates{}, "total"),
			}
			if diff := cmp.Diff(tt.want, metrics, opts...); diff != "" {
				t.Errorf("calculateNodeState() = (-want,+got):\n%s", diff)
			}
		})
	}
}

func Test_calculateNodeTres(t *testing.T) {
	type args struct {
		node types.V0043Node
	}
	tests := []struct {
		name string
		args args
		want *NodeTres
	}{
		{
			name: "empty",
			want: &NodeTres{},
		},
		{
			name: "node0",
			args: args{
				node: *node0,
			},
			want: &NodeTres{
				CpusTotal:       16,
				CpusEffective:   14,
				CpusIdle:        16,
				MemoryTotal:     4096,
				MemoryEffective: 3072,
				MemoryFree:      4096,
			},
		},
		{
			name: "node1",
			args: args{
				node: *node1,
			},
			want: &NodeTres{
				CpusTotal:       8,
				CpusEffective:   8,
				CpusAlloc:       8,
				MemoryTotal:     2048,
				MemoryEffective: 2048,
				MemoryAlloc:     2000,
				MemoryFree:      48,
			},
		},
		{
			name: "node2",
			args: args{
				node: *node2,
			},
			want: &NodeTres{
				CpusTotal:       16,
				CpusEffective:   16,
				CpusAlloc:       16,
				MemoryTotal:     4096,
				MemoryEffective: 4096,
				MemoryAlloc:     3000,
				MemoryFree:      1096,
			},
		},
		{
			name: "node3",
			args: args{
				node: *node3,
			},
			want: &NodeTres{
				CpusTotal:       6,
				CpusEffective:   6,
				CpusAlloc:       4,
				CpusIdle:        2,
				MemoryTotal:     1024,
				MemoryEffective: 1024,
				MemoryAlloc:     800,
				MemoryFree:      224,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &NodeTres{}
			calculateNodeTres(metrics, tt.args.node)
			opts := []cmp.Option{
				cmpopts.IgnoreUnexported(NodeMetrics{}),
				cmpopts.IgnoreFields(NodeTres{}, "total"),
			}
			if diff := cmp.Diff(tt.want, metrics, opts...); diff != "" {
				t.Errorf("calculateNodeTres() = (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestNodeCollector_getNodeMetrics(t *testing.T) {
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
		want    *NodeCollectorMetrics
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
			want: &NodeCollectorMetrics{
				NodeTresPer: map[string]*NodeTres{},
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
			want: &NodeCollectorMetrics{
				NodeMetrics: NodeMetrics{
					NodeCount: 4,
					NodeStates: NodeStates{
						Allocated:  2,
						Idle:       1,
						Mixed:      1,
						Completing: 1,
						Drain:      1,
					},
					NodeTres: NodeTres{
						CpusTotal:       46,
						CpusEffective:   44,
						CpusAlloc:       28,
						CpusIdle:        18,
						MemoryTotal:     11264,
						MemoryEffective: 10240,
						MemoryAlloc:     5800,
						MemoryFree:      5464,
					},
				},
				NodeTresPer: map[string]*NodeTres{
					"node0": {
						CpusTotal:       16,
						CpusEffective:   14,
						CpusIdle:        16,
						MemoryTotal:     4096,
						MemoryEffective: 3072,
						MemoryFree:      4096,
					},
					"node1": {
						CpusTotal:       8,
						CpusEffective:   8,
						CpusAlloc:       8,
						MemoryTotal:     2048,
						MemoryEffective: 2048,
						MemoryAlloc:     2000,
						MemoryFree:      48,
					},
					"node2": {
						CpusTotal:       16,
						CpusEffective:   16,
						CpusAlloc:       16,
						MemoryTotal:     4096,
						MemoryEffective: 4096,
						MemoryAlloc:     3000,
						MemoryFree:      1096,
					},
					"node3": {
						CpusTotal:       6,
						CpusEffective:   6,
						CpusAlloc:       4,
						CpusIdle:        2,
						MemoryTotal:     1024,
						MemoryEffective: 1024,
						MemoryAlloc:     800,
						MemoryFree:      224,
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
			c := &nodeCollector{
				slurmClient: tt.fields.slurmClient,
			}
			got, err := c.getNodeMetrics(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("nodeCollector.getNodeMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			opts := []cmp.Option{
				cmpopts.IgnoreUnexported(NodeMetrics{}),
				cmpopts.IgnoreFields(NodeStates{}, "total"),
				cmpopts.IgnoreFields(NodeTres{}, "total"),
			}
			if diff := cmp.Diff(tt.want, got, opts...); diff != "" {
				t.Errorf("nodeCollector.getNodeMetrics() = (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestNodeCollector_Collect(t *testing.T) {
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
			c := NewNodeCollector(tt.fields.slurmClient)
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

func TestNodeCollector_Describe(t *testing.T) {
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
			c := NewNodeCollector(tt.fields.slurmClient)
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

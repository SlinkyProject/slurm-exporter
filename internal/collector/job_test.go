// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"context"
	"sort"
	"testing"

	api "github.com/SlinkyProject/slurm-client/api/v0041"
	"github.com/SlinkyProject/slurm-client/pkg/client"
	"github.com/SlinkyProject/slurm-client/pkg/client/fake"
	"github.com/SlinkyProject/slurm-client/pkg/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/utils/ptr"
)

func Test_getJobResourceAlloc(t *testing.T) {
	type args struct {
		job types.V0041JobInfo
	}
	tests := []struct {
		name string
		args args
		want jobResources
	}{
		{
			name: "empty",
			args: args{
				job: types.V0041JobInfo{},
			},
			want: jobResources{},
		},
		{
			name: "test job 0",
			args: args{
				job: *job0,
			},
			want: jobResources{
				Cpus:   8,
				Memory: 1024,
			},
		},
		{
			name: "test job 2",
			args: args{
				job: *job2,
			},
			want: jobResources{
				Cpus:   12,
				Memory: 3072,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getJobResourceAlloc(tt.args.job); !apiequality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("getJobResourceAlloc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateJobState(t *testing.T) {
	type args struct {
		job types.V0041JobInfo
	}
	tests := []struct {
		name string
		args args
		want *JobStates
	}{
		{
			name: "empty",
			want: &JobStates{},
		},
		{
			name: "boot fail",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStateBOOTFAIL,
					}),
				}},
			},
			want: &JobStates{BootFail: 1},
		},
		{
			name: "cancelled",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStateCANCELLED,
					}),
				}},
			},
			want: &JobStates{Cancelled: 1},
		},
		{
			name: "completed",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStateCOMPLETED,
					}),
				}},
			},
			want: &JobStates{Completed: 1},
		},
		{
			name: "deadline",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStateDEADLINE,
					}),
				}},
			},
			want: &JobStates{Deadline: 1},
		},
		{
			name: "failed",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStateFAILED,
					}),
				}},
			},
			want: &JobStates{Failed: 1},
		},
		{
			name: "pending",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStatePENDING,
					}),
				}},
			},
			want: &JobStates{Pending: 1},
		},
		{
			name: "preempted",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStatePREEMPTED,
					}),
				}},
			},
			want: &JobStates{Preempted: 1},
		},
		{
			name: "running",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStateRUNNING,
					}),
				}},
			},
			want: &JobStates{Running: 1},
		},
		{
			name: "suspended",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStateSUSPENDED,
					}),
				}},
			},
			want: &JobStates{Suspended: 1},
		},
		{
			name: "timeout",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStateTIMEOUT,
					}),
				}},
			},
			want: &JobStates{Timeout: 1},
		},
		{
			name: "node fail",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStateNODEFAIL,
					}),
				}},
			},
			want: &JobStates{NodeFail: 1},
		},
		{
			name: "out of memory",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStateOUTOFMEMORY,
					}),
				}},
			},
			want: &JobStates{OutOfMemory: 1},
		},
		{
			name: "all states, all flags",
			args: args{
				job: types.V0041JobInfo{V0041JobInfo: api.V0041JobInfo{
					JobState: ptr.To([]api.V0041JobInfoJobState{
						api.V0041JobInfoJobStateBOOTFAIL,
						api.V0041JobInfoJobStateCANCELLED,
						api.V0041JobInfoJobStateCOMPLETED,
						api.V0041JobInfoJobStateCOMPLETING,
						api.V0041JobInfoJobStateCONFIGURING,
						api.V0041JobInfoJobStateDEADLINE,
						api.V0041JobInfoJobStateFAILED,
						api.V0041JobInfoJobStateLAUNCHFAILED,
						api.V0041JobInfoJobStateNODEFAIL,
						api.V0041JobInfoJobStateOUTOFMEMORY,
						api.V0041JobInfoJobStatePENDING,
						api.V0041JobInfoJobStatePOWERUPNODE,
						api.V0041JobInfoJobStatePREEMPTED,
						api.V0041JobInfoJobStateRECONFIGFAIL,
						api.V0041JobInfoJobStateREQUEUED,
						api.V0041JobInfoJobStateREQUEUEFED,
						api.V0041JobInfoJobStateREQUEUEHOLD,
						api.V0041JobInfoJobStateRESIZING,
						api.V0041JobInfoJobStateRESVDELHOLD,
						api.V0041JobInfoJobStateREVOKED,
						api.V0041JobInfoJobStateRUNNING,
						api.V0041JobInfoJobStateSIGNALING,
						api.V0041JobInfoJobStateSPECIALEXIT,
						api.V0041JobInfoJobStateSTAGEOUT,
						api.V0041JobInfoJobStateSTOPPED,
						api.V0041JobInfoJobStateSUSPENDED,
						api.V0041JobInfoJobStateTIMEOUT,
						api.V0041JobInfoJobStateUPDATEDB,
					}),
					Hold: ptr.To(true),
				}},
			},
			want: &JobStates{
				BootFail:    1,
				Completing:  1,
				Configuring: 1,
				PowerUpNode: 1,
				StageOut:    1,
				Hold:        1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &JobStates{}
			calculateJobState(metrics, tt.args.job)
			opts := []cmp.Option{
				cmpopts.IgnoreUnexported(JobStates{}),
			}
			if diff := cmp.Diff(tt.want, metrics, opts...); diff != "" {
				t.Errorf("calculateJobState() = (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestJobCollector_getJobMetrics(t *testing.T) {
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
		want    *JobMetrics
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
			want: &JobMetrics{
				JobIndividualStates: []JobIndividualStates{},
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
			want: &JobMetrics{
				JobCount:  4,
				JobStates: JobStates{Pending: 2, Running: 2, Hold: 1},
				JobTres:   JobTres{CpusAlloc: 20, MemoryAlloc: 4096},
				JobIndividualStates: []JobIndividualStates{
					{JobID: "0", JobName: "test_job_0", Nodes: []string{"node1"}, Running: 1},
					{JobID: "1", JobName: "test_job_1", Nodes: []string{""}, Pending: 1, Hold: 1},
					{JobID: "2", JobName: "test_job_2", Nodes: []string{"node2", "node3"}, Running: 1},
					{JobID: "3", JobName: "test_job_3", Nodes: []string{""}, Pending: 1},
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
			c := &jobCollector{
				slurmClient: tt.fields.slurmClient,
			}
			got, err := c.getJobMetrics(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("jobCollector.getJobMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// Sort JobIndividualStates for consistent comparison
			if got != nil {
				sort.Slice(got.JobIndividualStates, func(i, j int) bool {
					return got.JobIndividualStates[i].JobID < got.JobIndividualStates[j].JobID
				})
			}
			
			opts := []cmp.Option{
				cmpopts.IgnoreUnexported(JobMetrics{}),
				cmpopts.IgnoreFields(JobStates{}, "total"),
				cmpopts.IgnoreFields(JobTres{}, "total"),
			}
			if diff := cmp.Diff(tt.want, got, opts...); diff != "" {
				t.Errorf("jobCollector.getJobMetrics() = (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestJobCollector_Collect(t *testing.T) {
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
			c := NewJobCollector(tt.fields.slurmClient)
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

func TestJobCollector_Describe(t *testing.T) {
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
			c := NewJobCollector(tt.fields.slurmClient)
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

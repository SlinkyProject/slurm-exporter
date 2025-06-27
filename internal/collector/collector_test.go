// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"context"
	"errors"
	"net/http"
	"strings"

	api "github.com/SlinkyProject/slurm-client/api/v0043"
	"github.com/SlinkyProject/slurm-client/pkg/client"
	"github.com/SlinkyProject/slurm-client/pkg/client/fake"
	"github.com/SlinkyProject/slurm-client/pkg/client/interceptor"
	"github.com/SlinkyProject/slurm-client/pkg/object"
	"github.com/SlinkyProject/slurm-client/pkg/types"
	"k8s.io/utils/ptr"
)

const (
	partition1Name = "blue"
	partition2Name = "green"
)

var (
	partition1 = &types.V0043PartitionInfo{V0043PartitionInfo: api.V0043PartitionInfo{
		Name: ptr.To(partition1Name),
		Partition: &struct {
			State *[]api.V0043PartitionInfoPartitionState "json:\"state,omitempty\""
		}{
			State: ptr.To([]api.V0043PartitionInfoPartitionState{
				api.V0043PartitionInfoPartitionStateUP,
			}),
		},
		Cpus: &struct {
			TaskBinding *int32 "json:\"task_binding,omitempty\""
			Total       *int32 "json:\"total,omitempty\""
		}{
			Total: ptr.To(*node0.Cpus + *node1.Cpus + *node2.Cpus),
		},
		Nodes: &struct {
			AllowedAllocation *string "json:\"allowed_allocation,omitempty\""
			Configured        *string "json:\"configured,omitempty\""
			Total             *int32  "json:\"total,omitempty\""
		}{
			Total: ptr.To[int32](3),
		},
	}}
	partition2 = &types.V0043PartitionInfo{V0043PartitionInfo: api.V0043PartitionInfo{
		Name: ptr.To(partition2Name),
		Partition: &struct {
			State *[]api.V0043PartitionInfoPartitionState "json:\"state,omitempty\""
		}{
			State: ptr.To([]api.V0043PartitionInfoPartitionState{
				api.V0043PartitionInfoPartitionStateDOWN,
			}),
		},
		Cpus: &struct {
			TaskBinding *int32 "json:\"task_binding,omitempty\""
			Total       *int32 "json:\"total,omitempty\""
		}{
			Total: ptr.To(*node1.Cpus + *node2.Cpus + *node3.Cpus),
		},
		Nodes: &struct {
			AllowedAllocation *string "json:\"allowed_allocation,omitempty\""
			Configured        *string "json:\"configured,omitempty\""
			Total             *int32  "json:\"total,omitempty\""
		}{
			Total: ptr.To[int32](3),
		},
	}}
	partitionList = &types.V0043PartitionInfoList{
		Items: []types.V0043PartitionInfo{
			*partition1, *partition2,
		},
	}
)

var (
	node0 = &types.V0043Node{V0043Node: api.V0043Node{
		Name:       ptr.To("node0"),
		Partitions: ptr.To(api.V0043CsvString{partition1Name}),
		State: ptr.To([]api.V0043NodeState{
			api.V0043NodeStateIDLE,
		}),
		Cpus:              ptr.To[int32](16),
		EffectiveCpus:     ptr.To[int32](14),
		AllocCpus:         ptr.To[int32](0),
		AllocIdleCpus:     ptr.To[int32](16),
		RealMemory:        ptr.To[int64](4096),
		SpecializedMemory: ptr.To[int64](1024),
		AllocMemory:       ptr.To[int64](0),
		FreeMem: &api.V0043Uint64NoValStruct{
			Number: ptr.To[int64](4096),
			Set:    ptr.To(true),
		},
	}}
	node1 = &types.V0043Node{V0043Node: api.V0043Node{
		Name:       ptr.To("node1"),
		Partitions: ptr.To(api.V0043CsvString{partition1Name, partition2Name}),
		State: ptr.To([]api.V0043NodeState{
			api.V0043NodeStateALLOCATED,
		}),
		Cpus:          ptr.To[int32](8),
		EffectiveCpus: ptr.To[int32](8),
		AllocCpus:     ptr.To[int32](8),
		AllocIdleCpus: ptr.To[int32](0),
		RealMemory:    ptr.To[int64](2048),
		AllocMemory:   ptr.To[int64](2000),
		FreeMem: &api.V0043Uint64NoValStruct{
			Number: ptr.To[int64](48),
			Set:    ptr.To(true),
		},
	}}
	node2 = &types.V0043Node{V0043Node: api.V0043Node{
		Name:       ptr.To("node2"),
		Partitions: ptr.To(api.V0043CsvString{partition1Name, partition2Name}),
		State: ptr.To([]api.V0043NodeState{
			api.V0043NodeStateALLOCATED,
			api.V0043NodeStateDRAIN,
		}),
		Cpus:          ptr.To[int32](16),
		EffectiveCpus: ptr.To[int32](16),
		AllocCpus:     ptr.To[int32](16),
		AllocIdleCpus: ptr.To[int32](0),
		RealMemory:    ptr.To[int64](4096),
		AllocMemory:   ptr.To[int64](3000),
		FreeMem: &api.V0043Uint64NoValStruct{
			Number: ptr.To[int64](1096),
			Set:    ptr.To(true),
		},
	}}
	node3 = &types.V0043Node{V0043Node: api.V0043Node{
		Name:       ptr.To("node3"),
		Partitions: ptr.To(api.V0043CsvString{partition2Name}),
		State: ptr.To([]api.V0043NodeState{
			api.V0043NodeStateMIXED,
			api.V0043NodeStateCOMPLETING,
		}),
		Cpus:          ptr.To[int32](6),
		EffectiveCpus: ptr.To[int32](6),
		AllocCpus:     ptr.To[int32](4),
		AllocIdleCpus: ptr.To[int32](2),
		RealMemory:    ptr.To[int64](1024),
		AllocMemory:   ptr.To[int64](800),
		FreeMem: &api.V0043Uint64NoValStruct{
			Number: ptr.To[int64](224),
			Set:    ptr.To(true),
		},
	}}
	nodeList = &types.V0043NodeList{
		Items: []types.V0043Node{
			*node0, *node1, *node2, *node3,
		},
	}
)

var (
	job0 = &types.V0043JobInfo{V0043JobInfo: api.V0043JobInfo{
		JobId:     ptr.To[int32](0),
		JobState:  ptr.To([]api.V0043JobInfoJobState{api.V0043JobInfoJobStateRUNNING}),
		Partition: partition1.Name,
		JobResources: &api.V0043JobRes{
			Nodes: &struct {
				Allocation *api.V0043JobResNodes             "json:\"allocation,omitempty\""
				Count      *int32                            "json:\"count,omitempty\""
				List       *string                           "json:\"list,omitempty\""
				SelectType *[]api.V0043JobResNodesSelectType "json:\"select_type,omitempty\""
				Whole      *bool                             "json:\"whole,omitempty\""
			}{
				Allocation: &api.V0043JobResNodes{
					{
						Cpus: &struct {
							Count *int32 "json:\"count,omitempty\""
							Used  *int32 "json:\"used,omitempty\""
						}{
							Count: ptr.To[int32](8),
						},
						Memory: &struct {
							Allocated *int64 "json:\"allocated,omitempty\""
							Used      *int64 "json:\"used,omitempty\""
						}{
							Allocated: ptr.To[int64](1024),
						},
					},
				},
			},
		},
		UserId:   ptr.To[int32](0),
		UserName: ptr.To("root"),
		Account:  ptr.To("root"),
	}}
	job1 = &types.V0043JobInfo{V0043JobInfo: api.V0043JobInfo{
		JobId:     ptr.To[int32](1),
		JobState:  ptr.To([]api.V0043JobInfoJobState{api.V0043JobInfoJobStatePENDING}),
		Partition: ptr.To(strings.Join([]string{partition1Name, partition2Name}, ",")),
		Hold:      ptr.To(true),
		NodeCount: &api.V0043Uint32NoValStruct{
			Number: ptr.To[int32](3),
			Set:    ptr.To(true),
		},
		UserId:   ptr.To[int32](0),
		UserName: ptr.To("root"),
	}}
	job2 = &types.V0043JobInfo{V0043JobInfo: api.V0043JobInfo{
		JobId:     ptr.To[int32](2),
		JobState:  ptr.To([]api.V0043JobInfoJobState{api.V0043JobInfoJobStateRUNNING}),
		Partition: partition2.Name,
		JobResources: &api.V0043JobRes{
			Nodes: &struct {
				Allocation *api.V0043JobResNodes             "json:\"allocation,omitempty\""
				Count      *int32                            "json:\"count,omitempty\""
				List       *string                           "json:\"list,omitempty\""
				SelectType *[]api.V0043JobResNodesSelectType "json:\"select_type,omitempty\""
				Whole      *bool                             "json:\"whole,omitempty\""
			}{
				Allocation: &api.V0043JobResNodes{
					{
						Cpus: &struct {
							Count *int32 "json:\"count,omitempty\""
							Used  *int32 "json:\"used,omitempty\""
						}{
							Count: ptr.To[int32](8),
						},
						Memory: &struct {
							Allocated *int64 "json:\"allocated,omitempty\""
							Used      *int64 "json:\"used,omitempty\""
						}{
							Allocated: ptr.To[int64](1024),
						},
					},
					{
						Cpus: &struct {
							Count *int32 "json:\"count,omitempty\""
							Used  *int32 "json:\"used,omitempty\""
						}{
							Count: ptr.To[int32](4),
						},
						Memory: &struct {
							Allocated *int64 "json:\"allocated,omitempty\""
							Used      *int64 "json:\"used,omitempty\""
						}{
							Allocated: ptr.To[int64](2048),
						},
					},
				},
			},
		},
		UserId:  ptr.To[int32](1000),
		Account: ptr.To("root"),
	}}
	job3 = &types.V0043JobInfo{V0043JobInfo: api.V0043JobInfo{
		JobId:     ptr.To[int32](3),
		JobState:  ptr.To([]api.V0043JobInfoJobState{api.V0043JobInfoJobStatePENDING}),
		Partition: ptr.To(partition2Name),
		NodeCount: &api.V0043Uint32NoValStruct{
			Number: ptr.To[int32](2),
			Set:    ptr.To(true),
		},
		UserId: ptr.To[int32](1000),
	}}
	jobList = &types.V0043JobInfoList{
		Items: []types.V0043JobInfo{
			*job0, *job1, *job2, *job3,
		},
	}
)

var (
	stats = &types.V0043Stats{V0043StatsMsg: api.V0043StatsMsg{
		ScheduleCycleDepth:     ptr.To[int32](1),
		ScheduleCycleLast:      ptr.To[int32](1),
		ScheduleCycleMax:       ptr.To[int32](1),
		ScheduleCycleMean:      ptr.To[int64](1),
		ScheduleCycleMeanDepth: ptr.To[int64](1),
		ScheduleCyclePerMinute: ptr.To[int64](1),
		ScheduleCycleSum:       ptr.To[int32](1),
		ScheduleCycleTotal:     ptr.To[int32](1),
		ScheduleQueueLength:    ptr.To[int32](1),
		ScheduleExit: &api.V0043ScheduleExitFields{
			DefaultQueueDepth: ptr.To[int32](2),
			EndJobQueue:       ptr.To[int32](2),
			Licenses:          ptr.To[int32](2),
			MaxJobStart:       ptr.To[int32](2),
			MaxRpcCnt:         ptr.To[int32](2),
			MaxSchedTime:      ptr.To[int32](2),
		},

		BfActive:             ptr.To(true),
		BfBackfilledHetJobs:  ptr.To[int32](3),
		BfBackfilledJobs:     ptr.To[int32](3),
		BfCycleCounter:       ptr.To[int32](3),
		BfCycleLast:          ptr.To[int32](3),
		BfCycleMax:           ptr.To[int32](3),
		BfCycleMean:          ptr.To[int64](3),
		BfCycleSum:           ptr.To[int64](3),
		BfDepthMean:          ptr.To[int64](3),
		BfDepthMeanTry:       ptr.To[int64](3),
		BfDepthSum:           ptr.To[int32](3),
		BfDepthTrySum:        ptr.To[int32](3),
		BfLastBackfilledJobs: ptr.To[int32](3),
		BfLastDepth:          ptr.To[int32](3),
		BfLastDepthTry:       ptr.To[int32](3),
		BfQueueLen:           ptr.To[int32](3),
		BfQueueLenMean:       ptr.To[int64](3),
		BfQueueLenSum:        ptr.To[int32](3),
		BfTableSize:          ptr.To[int32](3),
		BfTableSizeMean:      ptr.To[int64](3),
		BfTableSizeSum:       ptr.To[int32](3),
		BfWhenLastCycle: &api.V0043Uint64NoValStruct{
			Number: ptr.To[int64](3),
			Set:    ptr.To(true),
		},

		BfExit: &api.V0043BfExitFields{
			BfMaxJobStart:   ptr.To[int32](4),
			BfMaxJobTest:    ptr.To[int32](4),
			BfMaxTime:       ptr.To[int32](4),
			BfNodeSpaceSize: ptr.To[int32](4),
			EndJobQueue:     ptr.To[int32](4),
			StateChanged:    ptr.To[int32](4),
		},

		JobStatesTs: &api.V0043Uint64NoValStruct{
			Number: ptr.To[int64](5),
			Set:    ptr.To(true),
		},
		JobsCanceled:  ptr.To[int32](5),
		JobsCompleted: ptr.To[int32](5),
		JobsFailed:    ptr.To[int32](5),
		JobsPending:   ptr.To[int32](5),
		JobsRunning:   ptr.To[int32](5),
		JobsStarted:   ptr.To[int32](5),
		JobsSubmitted: ptr.To[int32](5),

		AgentCount:       ptr.To[int32](6),
		AgentQueueSize:   ptr.To[int32](6),
		AgentThreadCount: ptr.To[int32](6),

		ServerThreadCount: ptr.To[int32](7),

		DbdAgentQueueSize: ptr.To[int32](8),
	}}
)

var testDataClient = fake.NewClientBuilder().
	WithLists(partitionList, nodeList, jobList).
	WithObjects(stats).
	Build()

var testFailClient = fake.NewClientBuilder().
	WithLists(partitionList, nodeList, jobList).
	WithInterceptorFuncs(interceptor.Funcs{
		List: func(ctx context.Context, list object.ObjectList, opts ...client.ListOption) error {
			return errors.New(http.StatusText(http.StatusInternalServerError))
		},
	}).
	Build()

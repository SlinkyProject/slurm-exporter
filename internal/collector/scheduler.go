// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/SlinkyProject/slurm-client/pkg/client"
	"github.com/SlinkyProject/slurm-client/pkg/types"
)

// Ref: https://prometheus.io/docs/practices/naming/#metric-names
func NewSchedulerCollector(slurmClient client.Client) prometheus.Collector {
	return &schedulerCollector{
		slurmClient: slurmClient,

		schedulerStats: schedulerStats{
			ScheduleCycleDepth:     prometheus.NewDesc("slurm_scheduler_cycle_depth_total", "Total number of jobs processed in scheduling cycles", nil, nil),
			ScheduleCycleLast:      prometheus.NewDesc("slurm_scheduler_cycle_last_seconds", "Time in microseconds for last scheduling cycle", nil, nil),
			ScheduleCycleMax:       prometheus.NewDesc("slurm_scheduler_cycle_max_seconds", "Max time of any scheduling cycle in microseconds since last reset", nil, nil),
			ScheduleCycleMean:      prometheus.NewDesc("slurm_scheduler_cycle_mean_seconds", "Mean time in microseconds for all scheduling cycles since last reset", nil, nil),
			ScheduleCycleMeanDepth: prometheus.NewDesc("slurm_scheduler_cycle_depth_mean_total", "Mean of the number of jobs processed in a scheduling", nil, nil),
			ScheduleCyclePerMinute: prometheus.NewDesc("slurm_scheduler_cycle_perminute_total", "Number of scheduling executions per minute", nil, nil),
			ScheduleCycleSum:       prometheus.NewDesc("slurm_scheduler_cycle_sum_seconds_total", "Total run time in microseconds for all scheduling cycles since last reset", nil, nil),
			ScheduleCycleTotal:     prometheus.NewDesc("slurm_scheduler_cycle_total", "Number of scheduling cycles since last reset", nil, nil),
			ScheduleQueueLength:    prometheus.NewDesc("slurm_scheduler_queue_total", "Number of jobs pending in queue", nil, nil),
			// Limits
			DefaultQueueDepth: prometheus.NewDesc("slurm_scheduler_defaultqueuedepth_total", "Reached number of jobs allowed to be tested", nil, nil),
			EndJobQueue:       prometheus.NewDesc("slurm_scheduler_endjobqueue_total", "Reached end of queue", nil, nil),
			Licenses:          prometheus.NewDesc("slurm_scheduler_licenses_total", "Blocked on licenses", nil, nil),
			MaxJobStart:       prometheus.NewDesc("slurm_scheduler_maxjobstart_total", "Reached number of jobs allowed to start", nil, nil),
			MaxRpcCnt:         prometheus.NewDesc("slurm_scheduler_maxrpc_total", "Reached RPC limit", nil, nil),
			MaxSchedTime:      prometheus.NewDesc("slurm_scheduler_maxschedtime_total", "Reached maximum allowed scheduler time", nil, nil),
		},
		bfSchedulerStats: bfSchedulerStats{
			BfActive:             prometheus.NewDesc("slurm_bfscheduler_active_bool", "Backfill scheduler currently running", nil, nil),
			BfBackfilledHetJobs:  prometheus.NewDesc("slurm_bfscheduler_backfilledhetjobs_total", "Number of heterogeneous job components started through backfilling since last Slurm start", nil, nil),
			BfBackfilledJobs:     prometheus.NewDesc("slurm_bfscheduler_backfilledjobs_total", "Number of jobs started through backfilling since last slurm start", nil, nil),
			BfCycleCounter:       prometheus.NewDesc("slurm_bfscheduler_cycle_total", "Number of backfill scheduling cycles since last reset", nil, nil),
			BfCycleLast:          prometheus.NewDesc("slurm_bfscheduler_cycle_seconds", "Execution time in microseconds of last backfill scheduling cycle", nil, nil),
			BfCycleMax:           prometheus.NewDesc("slurm_bfscheduler_cycle_max_seconds", "Execution time in microseconds of longest backfill scheduling cycle", nil, nil),
			BfCycleMean:          prometheus.NewDesc("slurm_bfscheduler_cycle_mean_seconds", "Mean time in microseconds of backfilling scheduling cycles since last reset", nil, nil),
			BfCycleSum:           prometheus.NewDesc("slurm_bfscheduler_cycle_sum_seconds", "Total time in microseconds of backfilling scheduling cycles since last reset", nil, nil),
			BfDepthMean:          prometheus.NewDesc("slurm_bfscheduler_depth_mean_total", "Mean number of eligible to run jobs processed during all backfilling scheduling cycles since last reset", nil, nil),
			BfDepthMeanTry:       prometheus.NewDesc("slurm_bfscheduler_depth_try_total", "The subset of Depth Mean that the backfill scheduler attempted to schedule", nil, nil),
			BfDepthSum:           prometheus.NewDesc("slurm_bfscheduler_depth_sum_total", "Total number of jobs processed during all backfilling scheduling cycles since last reset", nil, nil),
			BfDepthTrySum:        prometheus.NewDesc("slurm_bfscheduler_depth_trysum_total", "Subset of bf_depth_sum that the backfill scheduler attempted to schedule", nil, nil),
			BfLastBackfilledJobs: prometheus.NewDesc("slurm_bfscheduler_lastbackfilledjobs_total", "Number of jobs started through backfilling since last reset", nil, nil),
			BfLastDepth:          prometheus.NewDesc("slurm_bfscheduler_lastdepth_total", "Number of processed jobs during last backfilling scheduling cycle", nil, nil),
			BfLastDepthTry:       prometheus.NewDesc("slurm_bfscheduler_lastdepthtry_total", "Number of processed jobs during last backfilling scheduling cycle that had a chance to start using available resources", nil, nil),
			BfQueueLen:           prometheus.NewDesc("slurm_bfscheduler_queue_total", "Number of jobs pending to be processed by backfilling algorithm", nil, nil),
			BfQueueLenMean:       prometheus.NewDesc("slurm_bfscheduler_queue_mean_total", "Mean number of jobs pending to be processed by backfilling algorithm", nil, nil),
			BfQueueLenSum:        prometheus.NewDesc("slurm_bfscheduler_queue_sum_total", "Total number of jobs pending to be processed by backfilling algorithm since last reset", nil, nil),
			BfTableSize:          prometheus.NewDesc("slurm_bfscheduler_table_total", "Number of different time slots tested by the backfill scheduler in its last iteration", nil, nil),
			BfTableSizeMean:      prometheus.NewDesc("slurm_bfscheduler_tablemean_total", "Mean number of different time slots tested by the backfill scheduler", nil, nil),
			BfTableSizeSum:       prometheus.NewDesc("slurm_bfscheduler_tablesum_total", "Total number of different time slots tested by the backfill scheduler", nil, nil),
			BfWhenLastCycle:      prometheus.NewDesc("slurm_bfscheduler_lastcycle_timestamp", "When the last backfill scheduling cycle happened (UNIX timestamp)", nil, nil),
			// Limits
			BfMaxJobStart:   prometheus.NewDesc("slurm_bfscheduler_maxjobstart_total", "Reached number of jobs allowed to be tested", nil, nil),
			BfMaxJobTest:    prometheus.NewDesc("slurm_bfscheduler_maxjobtest_total", "Reached end of queue", nil, nil),
			BfMaxTime:       prometheus.NewDesc("slurm_bfscheduler_maxtime_total", "Blocked on licenses", nil, nil),
			BfNodeSpaceSize: prometheus.NewDesc("slurm_bfscheduler_nodespace_total", "Reached table size limit", nil, nil),
			BfEndJobQueue:   prometheus.NewDesc("slurm_bfscheduler_endjobqueue_total", "Reached RPC limit", nil, nil),
			BfStateChanged:  prometheus.NewDesc("slurm_bfscheduler_statechanged_total", "Reached maximum allowed scheduler time", nil, nil),
		},
		jobStats: jobStats{
			JobStatesTs:   prometheus.NewDesc("slurm_scheduler_jobs_stats_timestamp", "When the job state counts were gathered (UNIX timestamp)", nil, nil),
			JobsCanceled:  prometheus.NewDesc("slurm_scheduler_jobs_canceled_total", "Number of jobs canceled since the last reset", nil, nil),
			JobsCompleted: prometheus.NewDesc("slurm_scheduler_jobs_completed_total", "Number of jobs completed since last reset", nil, nil),
			JobsFailed:    prometheus.NewDesc("slurm_scheduler_jobs_failed_total", "Number of jobs failed due to slurmd or other internal issues since last reset", nil, nil),
			JobsPending:   prometheus.NewDesc("slurm_scheduler_jobs_pending_total", "Number of jobs pending at the time of listed in job_state_ts", nil, nil),
			JobsRunning:   prometheus.NewDesc("slurm_scheduler_jobs_running_total", "Number of jobs running at the time of listed in job_state_ts", nil, nil),
			JobsStarted:   prometheus.NewDesc("slurm_scheduler_jobs_started_total", "Number of jobs started since last reset", nil, nil),
			JobsSubmitted: prometheus.NewDesc("slurm_scheduler_jobs_submitted_total", "Number of jobs submitted since last reset", nil, nil),
		},
		agentStats: agentStats{
			AgentCount:       prometheus.NewDesc("slurm_scheduler_agent_total", "Number of agent threads", nil, nil),
			AgentQueueSize:   prometheus.NewDesc("slurm_scheduler_agent_queue_total", "Number of enqueued outgoing RPC requests in an internal retry list", nil, nil),
			AgentThreadCount: prometheus.NewDesc("slurm_scheduler_agent_thread_total", "Total number of active threads created by all agent threads", nil, nil),
		},
		ServerThreadCount: prometheus.NewDesc("slurm_scheduler_thread_total", "Number of current active slurmctld threads", nil, nil),
		DbdAgentQueueSize: prometheus.NewDesc("slurm_scheduler_dbdagentqueue_total", "Number of messages for SlurmDBD that are queued", nil, nil),
	}
}

type schedulerCollector struct {
	slurmClient client.Client

	schedulerStats
	bfSchedulerStats
	jobStats
	agentStats
	ServerThreadCount *prometheus.Desc
	DbdAgentQueueSize *prometheus.Desc
}

type schedulerStats struct {
	ScheduleCycleDepth     *prometheus.Desc
	ScheduleCycleLast      *prometheus.Desc
	ScheduleCycleMax       *prometheus.Desc
	ScheduleCycleMean      *prometheus.Desc
	ScheduleCycleMeanDepth *prometheus.Desc
	ScheduleCyclePerMinute *prometheus.Desc
	ScheduleCycleSum       *prometheus.Desc
	ScheduleCycleTotal     *prometheus.Desc
	ScheduleQueueLength    *prometheus.Desc
	// Limits
	DefaultQueueDepth *prometheus.Desc
	EndJobQueue       *prometheus.Desc
	Licenses          *prometheus.Desc
	MaxJobStart       *prometheus.Desc
	MaxRpcCnt         *prometheus.Desc
	MaxSchedTime      *prometheus.Desc
}

type bfSchedulerStats struct {
	BfActive             *prometheus.Desc
	BfBackfilledHetJobs  *prometheus.Desc
	BfBackfilledJobs     *prometheus.Desc
	BfCycleCounter       *prometheus.Desc
	BfCycleLast          *prometheus.Desc
	BfCycleMax           *prometheus.Desc
	BfCycleMean          *prometheus.Desc
	BfCycleSum           *prometheus.Desc
	BfDepthMean          *prometheus.Desc
	BfDepthMeanTry       *prometheus.Desc
	BfDepthSum           *prometheus.Desc
	BfDepthTrySum        *prometheus.Desc
	BfLastBackfilledJobs *prometheus.Desc
	BfLastDepth          *prometheus.Desc
	BfLastDepthTry       *prometheus.Desc
	BfQueueLen           *prometheus.Desc
	BfQueueLenMean       *prometheus.Desc
	BfQueueLenSum        *prometheus.Desc
	BfTableSize          *prometheus.Desc
	BfTableSizeMean      *prometheus.Desc
	BfTableSizeSum       *prometheus.Desc
	BfWhenLastCycle      *prometheus.Desc
	// Limits
	BfMaxJobStart   *prometheus.Desc
	BfMaxJobTest    *prometheus.Desc
	BfMaxTime       *prometheus.Desc
	BfNodeSpaceSize *prometheus.Desc
	BfEndJobQueue   *prometheus.Desc
	BfStateChanged  *prometheus.Desc
}

type jobStats struct {
	JobStatesTs   *prometheus.Desc
	JobsCanceled  *prometheus.Desc
	JobsCompleted *prometheus.Desc
	JobsFailed    *prometheus.Desc
	JobsPending   *prometheus.Desc
	JobsRunning   *prometheus.Desc
	JobsStarted   *prometheus.Desc
	JobsSubmitted *prometheus.Desc
}

type agentStats struct {
	AgentCount       *prometheus.Desc
	AgentQueueSize   *prometheus.Desc
	AgentThreadCount *prometheus.Desc
}

func (c *schedulerCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *schedulerCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.TODO()
	logger := log.FromContext(ctx).WithName("SchedulerCollector")

	logger.V(1).Info("collecting metrics")

	metrics, err := c.getSchedulerMetrics(ctx)
	if err != nil {
		logger.Error(err, "failed to collect scheduler metrics")
		return
	}

	// Scheduler
	ch <- prometheus.MustNewConstMetric(c.ScheduleCycleDepth, prometheus.GaugeValue, float64(metrics.ScheduleCycleDepth))
	ch <- prometheus.MustNewConstMetric(c.ScheduleCycleLast, prometheus.GaugeValue, float64(metrics.ScheduleCycleLast))
	ch <- prometheus.MustNewConstMetric(c.ScheduleCycleMax, prometheus.GaugeValue, float64(metrics.ScheduleCycleMax))
	ch <- prometheus.MustNewConstMetric(c.ScheduleCycleMean, prometheus.GaugeValue, float64(metrics.ScheduleCycleMean))
	ch <- prometheus.MustNewConstMetric(c.ScheduleCycleMeanDepth, prometheus.GaugeValue, float64(metrics.ScheduleCycleMeanDepth))
	ch <- prometheus.MustNewConstMetric(c.ScheduleCyclePerMinute, prometheus.GaugeValue, float64(metrics.ScheduleCyclePerMinute))
	ch <- prometheus.MustNewConstMetric(c.ScheduleCycleSum, prometheus.GaugeValue, float64(metrics.ScheduleCycleSum))
	ch <- prometheus.MustNewConstMetric(c.ScheduleCycleTotal, prometheus.GaugeValue, float64(metrics.ScheduleCycleTotal))
	ch <- prometheus.MustNewConstMetric(c.ScheduleQueueLength, prometheus.GaugeValue, float64(metrics.ScheduleQueueLength))
	// Scheduler Limits
	ch <- prometheus.MustNewConstMetric(c.DefaultQueueDepth, prometheus.GaugeValue, float64(metrics.ScheduleExit.DefaultQueueDepth))
	ch <- prometheus.MustNewConstMetric(c.EndJobQueue, prometheus.GaugeValue, float64(metrics.ScheduleExit.EndJobQueue))
	ch <- prometheus.MustNewConstMetric(c.Licenses, prometheus.GaugeValue, float64(metrics.ScheduleExit.Licenses))
	ch <- prometheus.MustNewConstMetric(c.MaxJobStart, prometheus.GaugeValue, float64(metrics.ScheduleExit.MaxJobStart))
	ch <- prometheus.MustNewConstMetric(c.MaxRpcCnt, prometheus.GaugeValue, float64(metrics.ScheduleExit.MaxRpcCnt))
	ch <- prometheus.MustNewConstMetric(c.MaxSchedTime, prometheus.GaugeValue, float64(metrics.ScheduleExit.MaxSchedTime))
	// Backfill
	if metrics.BfActive {
		ch <- prometheus.MustNewConstMetric(c.BfActive, prometheus.GaugeValue, float64(1))
	} else {
		ch <- prometheus.MustNewConstMetric(c.BfActive, prometheus.GaugeValue, float64(0))
	}
	ch <- prometheus.MustNewConstMetric(c.BfBackfilledHetJobs, prometheus.GaugeValue, float64(metrics.BfBackfilledHetJobs))
	ch <- prometheus.MustNewConstMetric(c.BfBackfilledJobs, prometheus.GaugeValue, float64(metrics.BfBackfilledJobs))
	ch <- prometheus.MustNewConstMetric(c.BfCycleCounter, prometheus.GaugeValue, float64(metrics.BfCycleCounter))
	ch <- prometheus.MustNewConstMetric(c.BfCycleLast, prometheus.GaugeValue, float64(metrics.BfCycleLast))
	ch <- prometheus.MustNewConstMetric(c.BfCycleMax, prometheus.GaugeValue, float64(metrics.BfCycleMax))
	ch <- prometheus.MustNewConstMetric(c.BfCycleMean, prometheus.GaugeValue, float64(metrics.BfCycleMean))
	ch <- prometheus.MustNewConstMetric(c.BfCycleSum, prometheus.GaugeValue, float64(metrics.BfCycleSum))
	ch <- prometheus.MustNewConstMetric(c.BfDepthMean, prometheus.GaugeValue, float64(metrics.BfDepthMean))
	ch <- prometheus.MustNewConstMetric(c.BfDepthMeanTry, prometheus.GaugeValue, float64(metrics.BfDepthMeanTry))
	ch <- prometheus.MustNewConstMetric(c.BfDepthSum, prometheus.GaugeValue, float64(metrics.BfDepthSum))
	ch <- prometheus.MustNewConstMetric(c.BfDepthTrySum, prometheus.GaugeValue, float64(metrics.BfDepthTrySum))
	ch <- prometheus.MustNewConstMetric(c.BfLastBackfilledJobs, prometheus.GaugeValue, float64(metrics.BfLastBackfilledJobs))
	ch <- prometheus.MustNewConstMetric(c.BfLastDepth, prometheus.GaugeValue, float64(metrics.BfLastDepth))
	ch <- prometheus.MustNewConstMetric(c.BfLastDepthTry, prometheus.GaugeValue, float64(metrics.BfLastDepthTry))
	ch <- prometheus.MustNewConstMetric(c.BfQueueLen, prometheus.GaugeValue, float64(metrics.BfQueueLen))
	ch <- prometheus.MustNewConstMetric(c.BfQueueLenMean, prometheus.GaugeValue, float64(metrics.BfQueueLenMean))
	ch <- prometheus.MustNewConstMetric(c.BfQueueLenSum, prometheus.GaugeValue, float64(metrics.BfQueueLenSum))
	ch <- prometheus.MustNewConstMetric(c.BfTableSize, prometheus.GaugeValue, float64(metrics.BfTableSize))
	ch <- prometheus.MustNewConstMetric(c.BfTableSizeMean, prometheus.GaugeValue, float64(metrics.BfTableSizeMean))
	ch <- prometheus.MustNewConstMetric(c.BfTableSizeSum, prometheus.GaugeValue, float64(metrics.BfTableSizeSum))
	ch <- prometheus.MustNewConstMetric(c.BfWhenLastCycle, prometheus.GaugeValue, float64(metrics.BfWhenLastCycle))
	// Backfill Limits
	ch <- prometheus.MustNewConstMetric(c.BfMaxJobStart, prometheus.GaugeValue, float64(metrics.BfExit.BfMaxJobStart))
	ch <- prometheus.MustNewConstMetric(c.BfMaxJobTest, prometheus.GaugeValue, float64(metrics.BfExit.BfMaxJobTest))
	ch <- prometheus.MustNewConstMetric(c.BfMaxTime, prometheus.GaugeValue, float64(metrics.BfExit.BfMaxTime))
	ch <- prometheus.MustNewConstMetric(c.BfNodeSpaceSize, prometheus.GaugeValue, float64(metrics.BfExit.BfNodeSpaceSize))
	ch <- prometheus.MustNewConstMetric(c.BfStateChanged, prometheus.GaugeValue, float64(metrics.BfExit.StateChanged))
	ch <- prometheus.MustNewConstMetric(c.BfEndJobQueue, prometheus.GaugeValue, float64(metrics.BfExit.EndJobQueue))
	// Jobs
	ch <- prometheus.MustNewConstMetric(c.JobStatesTs, prometheus.GaugeValue, float64(metrics.JobStatesTs))
	ch <- prometheus.MustNewConstMetric(c.JobsCanceled, prometheus.GaugeValue, float64(metrics.JobsCanceled))
	ch <- prometheus.MustNewConstMetric(c.JobsCompleted, prometheus.GaugeValue, float64(metrics.JobsCompleted))
	ch <- prometheus.MustNewConstMetric(c.JobsFailed, prometheus.GaugeValue, float64(metrics.JobsFailed))
	ch <- prometheus.MustNewConstMetric(c.JobsPending, prometheus.GaugeValue, float64(metrics.JobsPending))
	ch <- prometheus.MustNewConstMetric(c.JobsRunning, prometheus.GaugeValue, float64(metrics.JobsRunning))
	ch <- prometheus.MustNewConstMetric(c.JobsStarted, prometheus.GaugeValue, float64(metrics.JobsStarted))
	ch <- prometheus.MustNewConstMetric(c.JobsSubmitted, prometheus.GaugeValue, float64(metrics.JobsSubmitted))
	// Agent
	ch <- prometheus.MustNewConstMetric(c.AgentCount, prometheus.GaugeValue, float64(metrics.AgentCount))
	ch <- prometheus.MustNewConstMetric(c.AgentQueueSize, prometheus.GaugeValue, float64(metrics.AgentQueueSize))
	ch <- prometheus.MustNewConstMetric(c.AgentThreadCount, prometheus.GaugeValue, float64(metrics.AgentThreadCount))
	// Other
	ch <- prometheus.MustNewConstMetric(c.ServerThreadCount, prometheus.GaugeValue, float64(metrics.ServerThreadCount))
	ch <- prometheus.MustNewConstMetric(c.DbdAgentQueueSize, prometheus.GaugeValue, float64(metrics.DbdAgentQueueSize))
}

func (c *schedulerCollector) getSchedulerMetrics(ctx context.Context) (*SchedulerMetrics, error) {
	statsList := &types.V0043StatsList{}
	if err := c.slurmClient.List(ctx, statsList); err != nil {
		return nil, err
	}
	metrics := calculateSchedulerMetrics(statsList)
	return metrics, nil
}

func calculateSchedulerMetrics(statsList *types.V0043StatsList) *SchedulerMetrics {
	metrics := &SchedulerMetrics{}
	// NOTE: 0 <= len(statsList.Items) <= 1
	for _, stats := range statsList.Items {
		calculateSchedulerStats(metrics, stats)
	}
	return metrics
}

func calculateSchedulerStats(metrics *SchedulerMetrics, stats types.V0043Stats) {
	metrics.ScheduleCycleDepth = ptr.Deref(stats.ScheduleCycleDepth, 0)
	metrics.ScheduleCycleLast = ptr.Deref(stats.ScheduleCycleLast, 0)
	metrics.ScheduleCycleMax = ptr.Deref(stats.ScheduleCycleMax, 0)
	metrics.ScheduleCycleMean = ptr.Deref(stats.ScheduleCycleMean, 0)
	metrics.ScheduleCycleMeanDepth = ptr.Deref(stats.ScheduleCycleMeanDepth, 0)
	metrics.ScheduleCyclePerMinute = ptr.Deref(stats.ScheduleCyclePerMinute, 0)
	metrics.ScheduleCycleSum = ptr.Deref(stats.ScheduleCycleSum, 0)
	metrics.ScheduleCycleTotal = ptr.Deref(stats.ScheduleCycleTotal, 0)
	metrics.ScheduleQueueLength = ptr.Deref(stats.ScheduleQueueLength, 0)

	if stats.ScheduleExit != nil {
		metrics.ScheduleExit = ScheduleExitFields{
			DefaultQueueDepth: ptr.Deref(stats.ScheduleExit.DefaultQueueDepth, 0),
			EndJobQueue:       ptr.Deref(stats.ScheduleExit.EndJobQueue, 0),
			Licenses:          ptr.Deref(stats.ScheduleExit.Licenses, 0),
			MaxJobStart:       ptr.Deref(stats.ScheduleExit.MaxJobStart, 0),
			MaxRpcCnt:         ptr.Deref(stats.ScheduleExit.MaxRpcCnt, 0),
			MaxSchedTime:      ptr.Deref(stats.ScheduleExit.MaxSchedTime, 0),
		}
	}

	metrics.BfActive = ptr.Deref(stats.BfActive, false)
	metrics.BfBackfilledHetJobs = ptr.Deref(stats.BfBackfilledHetJobs, 0)
	metrics.BfBackfilledJobs = ptr.Deref(stats.BfBackfilledJobs, 0)
	metrics.BfCycleCounter = ptr.Deref(stats.BfCycleCounter, 0)
	metrics.BfCycleLast = ptr.Deref(stats.BfCycleLast, 0)
	metrics.BfCycleMax = ptr.Deref(stats.BfCycleMax, 0)
	metrics.BfCycleMean = ptr.Deref(stats.BfCycleMean, 0)
	metrics.BfCycleSum = ptr.Deref(stats.BfCycleSum, 0)
	metrics.BfDepthMean = ptr.Deref(stats.BfDepthMean, 0)
	metrics.BfDepthMeanTry = ptr.Deref(stats.BfDepthMeanTry, 0)
	metrics.BfDepthSum = ptr.Deref(stats.BfDepthSum, 0)
	metrics.BfDepthTrySum = ptr.Deref(stats.BfDepthTrySum, 0)
	metrics.BfLastBackfilledJobs = ptr.Deref(stats.BfLastBackfilledJobs, 0)
	metrics.BfLastDepth = ptr.Deref(stats.BfLastDepth, 0)
	metrics.BfLastDepthTry = ptr.Deref(stats.BfLastDepthTry, 0)
	metrics.BfQueueLen = ptr.Deref(stats.BfQueueLen, 0)
	metrics.BfQueueLenMean = ptr.Deref(stats.BfQueueLenMean, 0)
	metrics.BfQueueLenSum = ptr.Deref(stats.BfQueueLenSum, 0)
	metrics.BfTableSize = ptr.Deref(stats.BfTableSize, 0)
	metrics.BfTableSizeMean = ptr.Deref(stats.BfTableSizeMean, 0)
	metrics.BfTableSizeSum = ptr.Deref(stats.BfTableSizeSum, 0)
	metrics.BfWhenLastCycle = ParseUint64NoVal(stats.BfWhenLastCycle)

	if stats.BfExit != nil {
		metrics.BfExit = BfExitFields{
			BfMaxJobStart:   ptr.Deref(stats.BfExit.BfMaxJobStart, 0),
			BfMaxJobTest:    ptr.Deref(stats.BfExit.BfMaxJobTest, 0),
			BfMaxTime:       ptr.Deref(stats.BfExit.BfMaxTime, 0),
			BfNodeSpaceSize: ptr.Deref(stats.BfExit.BfNodeSpaceSize, 0),
			EndJobQueue:     ptr.Deref(stats.BfExit.EndJobQueue, 0),
			StateChanged:    ptr.Deref(stats.BfExit.StateChanged, 0),
		}
	}

	metrics.JobStatesTs = ParseUint64NoVal(stats.JobStatesTs)
	metrics.JobsCanceled = ptr.Deref(stats.JobsCanceled, 0)
	metrics.JobsCompleted = ptr.Deref(stats.JobsCompleted, 0)
	metrics.JobsFailed = ptr.Deref(stats.JobsFailed, 0)
	metrics.JobsPending = ptr.Deref(stats.JobsPending, 0)
	metrics.JobsRunning = ptr.Deref(stats.JobsRunning, 0)
	metrics.JobsStarted = ptr.Deref(stats.JobsStarted, 0)
	metrics.JobsSubmitted = ptr.Deref(stats.JobsSubmitted, 0)

	metrics.AgentCount = ptr.Deref(stats.AgentCount, 0)
	metrics.AgentQueueSize = ptr.Deref(stats.AgentQueueSize, 0)
	metrics.AgentThreadCount = ptr.Deref(stats.AgentThreadCount, 0)

	metrics.ServerThreadCount = ptr.Deref(stats.ServerThreadCount, 0)

	metrics.DbdAgentQueueSize = ptr.Deref(stats.DbdAgentQueueSize, 0)
}

type SchedulerMetrics struct {
	ScheduleCycleDepth     int32
	ScheduleCycleLast      int32
	ScheduleCycleMax       int32
	ScheduleCycleMean      int64
	ScheduleCycleMeanDepth int64
	ScheduleCyclePerMinute int64
	ScheduleCycleSum       int32
	ScheduleCycleTotal     int32
	ScheduleQueueLength    int32
	ScheduleExit           ScheduleExitFields

	BfActive             bool
	BfBackfilledHetJobs  int32
	BfBackfilledJobs     int32
	BfCycleCounter       int32
	BfCycleLast          int32
	BfCycleMax           int32
	BfCycleMean          int64
	BfCycleSum           int64
	BfDepthMean          int64
	BfDepthMeanTry       int64
	BfDepthSum           int32
	BfDepthTrySum        int32
	BfLastBackfilledJobs int32
	BfLastDepth          int32
	BfLastDepthTry       int32
	BfQueueLen           int32
	BfQueueLenMean       int64
	BfQueueLenSum        int32
	BfTableSize          int32
	BfTableSizeMean      int64
	BfTableSizeSum       int32
	BfWhenLastCycle      uint64
	BfExit               BfExitFields

	JobStatesTs   uint64
	JobsCanceled  int32
	JobsCompleted int32
	JobsFailed    int32
	JobsPending   int32
	JobsRunning   int32
	JobsStarted   int32
	JobsSubmitted int32

	AgentCount       int32
	AgentQueueSize   int32
	AgentThreadCount int32

	ServerThreadCount int32

	DbdAgentQueueSize int32
}

type ScheduleExitFields struct {
	DefaultQueueDepth int32
	EndJobQueue       int32
	Licenses          int32
	MaxJobStart       int32
	MaxRpcCnt         int32
	MaxSchedTime      int32
}

type BfExitFields struct {
	BfMaxJobStart   int32
	BfMaxJobTest    int32
	BfMaxTime       int32
	BfNodeSpaceSize int32
	EndJobQueue     int32
	StateChanged    int32
}

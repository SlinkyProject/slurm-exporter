// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/SlinkyProject/slurm-client/api/v0043"
	"github.com/SlinkyProject/slurm-client/pkg/client"
	"github.com/SlinkyProject/slurm-client/pkg/types"
	"github.com/SlinkyProject/slurm-exporter/internal/utils"
)

// Ref: https://prometheus.io/docs/practices/naming/#metric-names
func NewPartitionCollector(slurmClient client.Client) prometheus.Collector {
	return &partitionCollector{
		slurmClient: slurmClient,

		JobCount: prometheus.NewDesc("slurm_partition_jobs_total", "Total number of jobs in the partition", partitionLabels, nil),
		JobStates: jobStatesCollector{
			// Base State
			BootFail:    prometheus.NewDesc("slurm_partition_jobs_bootfail_total", "Number of jobs in BootFail state in the partition", partitionLabels, nil),
			Cancelled:   prometheus.NewDesc("slurm_partition_jobs_cancelled_total", "Number of jobs in Cancelled state in the partition", partitionLabels, nil),
			Completed:   prometheus.NewDesc("slurm_partition_jobs_completed_total", "Number of jobs in Completed state in the partition", partitionLabels, nil),
			Deadline:    prometheus.NewDesc("slurm_partition_jobs_deadline_total", "Number of jobs in Deadline state in the partition", partitionLabels, nil),
			Failed:      prometheus.NewDesc("slurm_partition_jobs_failed_total", "Number of jobs in Failed state in the partition", partitionLabels, nil),
			Pending:     prometheus.NewDesc("slurm_partition_jobs_pending_total", "Number of jobs in Pending state in the partition", partitionLabels, nil),
			Running:     prometheus.NewDesc("slurm_partition_jobs_running_total", "Number of jobs in Running state in the partition", partitionLabels, nil),
			Preempted:   prometheus.NewDesc("slurm_partition_jobs_preempted_total", "Number of jobs in Preempted state in the partition", partitionLabels, nil),
			Suspended:   prometheus.NewDesc("slurm_partition_jobs_suspended_total", "Number of jobs in Suspended state in the partition", partitionLabels, nil),
			Timeout:     prometheus.NewDesc("slurm_partition_jobs_timeout_total", "Number of jobs in Timeout state in the partition", partitionLabels, nil),
			NodeFail:    prometheus.NewDesc("slurm_partition_jobs_nodefail_total", "Number of jobs in NodeFail state in the partition", partitionLabels, nil),
			OutOfMemory: prometheus.NewDesc("slurm_partition_jobs_outofmemory_total", "Number of jobs in OutOfMemory state in the partition", partitionLabels, nil),
			// Flag States
			Completing:  prometheus.NewDesc("slurm_partition_jobs_completing_total", "Number of jobs with Completing flag in the partition", partitionLabels, nil),
			Configuring: prometheus.NewDesc("slurm_partition_jobs_configuring_total", "Number of jobs with Configuring flag in the partition", partitionLabels, nil),
			PowerUpNode: prometheus.NewDesc("slurm_partition_jobs_powerupnode_total", "Number of jobs with PowerUpNode flag in the partition", partitionLabels, nil),
			StageOut:    prometheus.NewDesc("slurm_partition_jobs_stageout_total", "Number of jobs with StageOut flag in the partition", partitionLabels, nil),
			// Other States
			Hold: prometheus.NewDesc("slurm_partition_jobs_hold_total", "Number of jobs with Hold flag in the partition", partitionLabels, nil),
		},
		JobTres: jobTresCollector{
			// CPUs
			CpusAlloc: prometheus.NewDesc("slurm_partition_jobs_cpus_alloc_total", "Number of Allocated CPUs among jobs in the partition", partitionLabels, nil),
			// Memory
			MemoryAlloc: prometheus.NewDesc("slurm_partition_jobs_memory_alloc_bytes", "Amount of Allocated Memory (MB) among jobs in the partition", partitionLabels, nil),
		},
		PendingNodeCount: prometheus.NewDesc("slurm_partition_jobs_pending_maxnodecount_total", "Largest number of nodes required among pending jobs in the partition", partitionLabels, nil),
		NodeCount:        prometheus.NewDesc("slurm_partition_nodes_total", "Total number of slurm nodes", partitionLabels, nil),
		NodeStates: nodeStatesCollector{
			// Base State
			Allocated: prometheus.NewDesc("slurm_partition_nodes_allocated_total", "Number of nodes in Allocated state", partitionLabels, nil),
			Down:      prometheus.NewDesc("slurm_partition_nodes_down_total", "Number of nodes in Down state", partitionLabels, nil),
			Error:     prometheus.NewDesc("slurm_partition_nodes_error_total", "Number of nodes in Error state", partitionLabels, nil),
			Future:    prometheus.NewDesc("slurm_partition_nodes_future_total", "Number of nodes in Future state", partitionLabels, nil),
			Idle:      prometheus.NewDesc("slurm_partition_nodes_idle_total", "Number of nodes in Idle state", partitionLabels, nil),
			Mixed:     prometheus.NewDesc("slurm_partition_nodes_mixed_total", "Number of nodes in Mixed state", partitionLabels, nil),
			Unknown:   prometheus.NewDesc("slurm_partition_nodes_unknown_total", "Number of nodes in Unknown state", partitionLabels, nil),
			// Flag State
			Completing:      prometheus.NewDesc("slurm_partition_nodes_completing_total", "Number of nodes with Completing flag", partitionLabels, nil),
			Drain:           prometheus.NewDesc("slurm_partition_nodes_drain_total", "Number of nodes with Drain flag", partitionLabels, nil),
			Fail:            prometheus.NewDesc("slurm_partition_nodes_fail_total", "Number of nodes with Fail flag", partitionLabels, nil),
			Maintenance:     prometheus.NewDesc("slurm_partition_nodes_maintenance_total", "Number of nodes with Maintenance flag", partitionLabels, nil),
			NotResponding:   prometheus.NewDesc("slurm_partition_nodes_notresponding_total", "Number of nodes with NotResponding flag", partitionLabels, nil),
			Planned:         prometheus.NewDesc("slurm_partition_nodes_planned_total", "Number of nodes with Planned flag", partitionLabels, nil),
			RebootRequested: prometheus.NewDesc("slurm_partition_nodes_rebootrequested_total", "Number of nodes with RebootRequested flag", partitionLabels, nil),
			Reserved:        prometheus.NewDesc("slurm_partition_nodes_reserved_total", "Number of nodes with Reserved flag", partitionLabels, nil),
		},
		NodeTres: nodeTresCollector{
			// CPUs
			CpusTotal:     prometheus.NewDesc("slurm_partition_nodes_cpus_total", "Total number of CPUs on the node", partitionLabels, nil),
			CpusEffective: prometheus.NewDesc("slurm_partition_nodes_cpus_effective_total", "Total number of effective CPUs on the node, excludes CoreSpec", partitionLabels, nil),
			CpusAlloc:     prometheus.NewDesc("slurm_partition_nodes_cpus_alloc_total", "Number of Allocated CPUs on the node", partitionLabels, nil),
			CpusIdle:      prometheus.NewDesc("slurm_partition_nodes_cpus_idle_total", "Number of Idle CPUs on the node", partitionLabels, nil),
			// Memory
			MemoryTotal:     prometheus.NewDesc("slurm_partition_nodes_memory_bytes", "Total amount of Memory (MB) on the node", partitionLabels, nil),
			MemoryEffective: prometheus.NewDesc("slurm_partition_nodes_memory_effective_bytes", "Total amount of effective Memory (MB) on the node, excludes MemSpec", partitionLabels, nil),
			MemoryAlloc:     prometheus.NewDesc("slurm_partition_nodes_memory_alloc_bytes", "Amount of Allocated Memory (MB) on the node", partitionLabels, nil),
			MemoryFree:      prometheus.NewDesc("slurm_partition_nodes_memory_free_bytes", "Amount of Free Memory (MB) on the node", partitionLabels, nil),
		},
	}
}

type partitionCollector struct {
	slurmClient client.Client

	JobCount  *prometheus.Desc
	JobStates jobStatesCollector
	JobTres   jobTresCollector

	NodeCount  *prometheus.Desc
	NodeStates nodeStatesCollector
	NodeTres   nodeTresCollector

	PendingNodeCount *prometheus.Desc
}

func (c *partitionCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *partitionCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.TODO()
	logger := log.FromContext(ctx).WithName("PartitionCollector")

	logger.V(1).Info("collecting metrics")

	metrics, err := c.getPartitionMetrics(ctx)
	if err != nil {
		logger.Error(err, "failed to collect partition metrics")
		return
	}

	for partition, data := range metrics.JobMetricsPer {
		ch <- prometheus.MustNewConstMetric(c.JobCount, prometheus.GaugeValue, float64(data.JobCount), partition)
		// States
		ch <- prometheus.MustNewConstMetric(c.JobStates.BootFail, prometheus.GaugeValue, float64(data.JobStates.BootFail), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Cancelled, prometheus.GaugeValue, float64(data.JobStates.Cancelled), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Completed, prometheus.GaugeValue, float64(data.JobStates.Completed), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Deadline, prometheus.GaugeValue, float64(data.JobStates.Deadline), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Failed, prometheus.GaugeValue, float64(data.JobStates.Failed), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Pending, prometheus.GaugeValue, float64(data.JobStates.Pending), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Preempted, prometheus.GaugeValue, float64(data.JobStates.Preempted), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Running, prometheus.GaugeValue, float64(data.JobStates.Running), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Suspended, prometheus.GaugeValue, float64(data.JobStates.Suspended), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Timeout, prometheus.GaugeValue, float64(data.JobStates.Timeout), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.NodeFail, prometheus.GaugeValue, float64(data.JobStates.NodeFail), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.OutOfMemory, prometheus.GaugeValue, float64(data.JobStates.OutOfMemory), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Completing, prometheus.GaugeValue, float64(data.JobStates.Completing), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Configuring, prometheus.GaugeValue, float64(data.JobStates.Configuring), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.PowerUpNode, prometheus.GaugeValue, float64(data.JobStates.PowerUpNode), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.StageOut, prometheus.GaugeValue, float64(data.JobStates.StageOut), partition)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Hold, prometheus.GaugeValue, float64(data.JobStates.Hold), partition)
		// Tres
		ch <- prometheus.MustNewConstMetric(c.JobTres.CpusAlloc, prometheus.GaugeValue, float64(data.JobTres.CpusAlloc), partition)
		ch <- prometheus.MustNewConstMetric(c.JobTres.MemoryAlloc, prometheus.GaugeValue, float64(data.JobTres.MemoryAlloc), partition)
		// Other
		ch <- prometheus.MustNewConstMetric(c.PendingNodeCount, prometheus.GaugeValue, float64(data.PendingNodeCount), partition)
	}

	for partition, data := range metrics.NodeMetricsPer {
		ch <- prometheus.MustNewConstMetric(c.NodeCount, prometheus.GaugeValue, float64(data.NodeCount), partition)
		// States
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Allocated, prometheus.GaugeValue, float64(data.NodeStates.Allocated), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Down, prometheus.GaugeValue, float64(data.NodeStates.Down), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Error, prometheus.GaugeValue, float64(data.NodeStates.Error), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Future, prometheus.GaugeValue, float64(data.NodeStates.Future), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Idle, prometheus.GaugeValue, float64(data.NodeStates.Idle), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Mixed, prometheus.GaugeValue, float64(data.NodeStates.Mixed), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Unknown, prometheus.GaugeValue, float64(data.NodeStates.Unknown), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Completing, prometheus.GaugeValue, float64(data.NodeStates.Completing), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Drain, prometheus.GaugeValue, float64(data.NodeStates.Drain), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Fail, prometheus.GaugeValue, float64(data.NodeStates.Fail), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Maintenance, prometheus.GaugeValue, float64(data.NodeStates.Maintenance), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.NotResponding, prometheus.GaugeValue, float64(data.NodeStates.NotResponding), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Planned, prometheus.GaugeValue, float64(data.NodeStates.Planned), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.RebootRequested, prometheus.GaugeValue, float64(data.NodeStates.RebootRequested), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeStates.Reserved, prometheus.GaugeValue, float64(data.NodeStates.Reserved), partition)
		// Tres
		ch <- prometheus.MustNewConstMetric(c.NodeTres.CpusTotal, prometheus.GaugeValue, float64(data.NodeTres.CpusTotal), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeTres.CpusEffective, prometheus.GaugeValue, float64(data.NodeTres.CpusEffective), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeTres.CpusAlloc, prometheus.GaugeValue, float64(data.NodeTres.CpusAlloc), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeTres.CpusIdle, prometheus.GaugeValue, float64(data.NodeTres.CpusIdle), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeTres.MemoryTotal, prometheus.GaugeValue, float64(data.NodeTres.MemoryTotal), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeTres.MemoryEffective, prometheus.GaugeValue, float64(data.NodeTres.MemoryEffective), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeTres.MemoryAlloc, prometheus.GaugeValue, float64(data.NodeTres.MemoryAlloc), partition)
		ch <- prometheus.MustNewConstMetric(c.NodeTres.MemoryFree, prometheus.GaugeValue, float64(data.NodeTres.MemoryFree), partition)
	}
}

func (c *partitionCollector) getPartitionMetrics(ctx context.Context) (*PartitionMetrics, error) {
	partitionList := &types.V0043PartitionInfoList{}
	if err := c.slurmClient.List(ctx, partitionList); err != nil {
		return nil, err
	}
	jobList := &types.V0043JobInfoList{}
	if err := c.slurmClient.List(ctx, jobList); err != nil {
		return nil, err
	}
	nodeList := &types.V0043NodeList{}
	if err := c.slurmClient.List(ctx, nodeList); err != nil {
		return nil, err
	}
	metrics := calculatePartitionMetrics(partitionList, nodeList, jobList)
	return metrics, nil
}

func calculatePartitionMetrics(
	partitionList *types.V0043PartitionInfoList,
	nodeList *types.V0043NodeList,
	jobList *types.V0043JobInfoList,
) *PartitionMetrics {
	metrics := &PartitionMetrics{
		NodeMetricsPer: make(map[string]*NodeMetrics, len(partitionList.Items)),
		JobMetricsPer:  make(map[string]*PartitionJobMetrics, len(partitionList.Items)),
	}

	for _, partition := range partitionList.Items {
		key := string(partition.GetKey())
		metrics.NodeMetricsPer[key] = &NodeMetrics{}
		metrics.JobMetricsPer[key] = &PartitionJobMetrics{}
	}

	for _, node := range nodeList.Items {
		for _, key := range ptr.Deref(node.Partitions, []string{}) {
			if _, ok := metrics.JobMetricsPer[key]; !ok {
				metrics.NodeMetricsPer[key] = &NodeMetrics{}
			}
			metrics.NodeMetricsPer[key].NodeCount++
			calculateNodeState(&metrics.NodeMetricsPer[key].NodeStates, node)
			calculateNodeTres(&metrics.NodeMetricsPer[key].NodeTres, node)
		}
	}

	for _, job := range jobList.Items {
		for _, key := range utils.ParseCSV(ptr.Deref(job.Partition, "")) {
			if _, ok := metrics.JobMetricsPer[key]; !ok {
				metrics.JobMetricsPer[key] = &PartitionJobMetrics{}
			}
			metrics.JobMetricsPer[key].JobCount++
			calculateJobState(&metrics.JobMetricsPer[key].JobStates, job)
			calculateJobTres(&metrics.JobMetricsPer[key].JobTres, job)
			metrics.JobMetricsPer[key].PendingNodeCount = max(metrics.JobMetricsPer[key].PendingNodeCount, getJobPendingNodeCount(job))
		}
	}

	return metrics
}

// getJobPendingNodeCount returns the requested node count if the job is
// pending, otherwise returns zero.
func getJobPendingNodeCount(job types.V0043JobInfo) uint {
	isPending := job.GetStateAsSet().Has(api.V0043JobInfoJobStatePENDING)
	isHold := ptr.Deref(job.Hold, false)
	if !isPending || isHold {
		return 0
	}
	nodeCount := ParseUint32NoVal(job.NodeCount)
	return uint(nodeCount)
}

type PartitionMetrics struct {
	NodeMetricsPer map[string]*NodeMetrics
	JobMetricsPer  map[string]*PartitionJobMetrics
}

type PartitionJobMetrics struct {
	JobMetrics
	PendingNodeCount uint
}

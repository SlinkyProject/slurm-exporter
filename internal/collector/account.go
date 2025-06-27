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
func NewAccountCollector(slurmClient client.Client) prometheus.Collector {
	return &accountCollector{
		slurmClient: slurmClient,

		JobCount: prometheus.NewDesc("slurm_account_jobs_total", "Total number of account jobs", accountLabels, nil),
		JobStates: jobStatesCollector{
			// Base States
			BootFail:    prometheus.NewDesc("slurm_account_jobs_bootfail_total", "Number of account jobs in BootFail state", accountLabels, nil),
			Cancelled:   prometheus.NewDesc("slurm_account_jobs_cancelled_total", "Number of account jobs in Cancelled state", accountLabels, nil),
			Completed:   prometheus.NewDesc("slurm_account_jobs_completed_total", "Number of account jobs in Completed state", accountLabels, nil),
			Deadline:    prometheus.NewDesc("slurm_account_jobs_deadline_total", "Number of account jobs in Deadline state", accountLabels, nil),
			Failed:      prometheus.NewDesc("slurm_account_jobs_failed_total", "Number of account jobs in Failed state", accountLabels, nil),
			Pending:     prometheus.NewDesc("slurm_account_jobs_pending_total", "Number of account jobs in Pending state", accountLabels, nil),
			Preempted:   prometheus.NewDesc("slurm_account_jobs_preempted_total", "Number of account jobs in Preempted state", accountLabels, nil),
			Running:     prometheus.NewDesc("slurm_account_jobs_running_total", "Number of account jobs in Running state", accountLabels, nil),
			Suspended:   prometheus.NewDesc("slurm_account_jobs_suspended_total", "Number of account jobs in Suspended state", accountLabels, nil),
			Timeout:     prometheus.NewDesc("slurm_account_jobs_timeout_total", "Number of account jobs in Timeout state", accountLabels, nil),
			NodeFail:    prometheus.NewDesc("slurm_account_jobs_nodefail_total", "Number of account jobs in NodeFail state", accountLabels, nil),
			OutOfMemory: prometheus.NewDesc("slurm_account_jobs_outofmemory_total", "Number of account jobs in OutOfMemory state", accountLabels, nil),
			// Flag States
			Completing:  prometheus.NewDesc("slurm_account_jobs_completing_total", "Number of account jobs with Completing flag", accountLabels, nil),
			Configuring: prometheus.NewDesc("slurm_account_jobs_configuring_total", "Number of account jobs with Configuring flag", accountLabels, nil),
			PowerUpNode: prometheus.NewDesc("slurm_account_jobs_powerupnode_total", "Number of account jobs with PowerUpNode flag", accountLabels, nil),
			StageOut:    prometheus.NewDesc("slurm_account_jobs_stageout_total", "Number of account jobs with StageOut flag", accountLabels, nil),
			// Other States
			Hold: prometheus.NewDesc("slurm_account_jobs_hold_total", "Number of account jobs with Hold flag", accountLabels, nil),
		},
		JobTres: jobTresCollector{
			// CPUs
			CpusAlloc: prometheus.NewDesc("slurm_account_jobs_cpus_alloc_total", "Number of Allocated CPUs among account jobs", accountLabels, nil),
			// Memory
			MemoryAlloc: prometheus.NewDesc("slurm_account_jobs_memory_alloc_bytes", "Amount of Allocated Memory (MB) among account jobs", accountLabels, nil),
		},
	}
}

type accountCollector struct {
	slurmClient client.Client

	JobCount  *prometheus.Desc
	JobStates jobStatesCollector
	JobTres   jobTresCollector
}

func (c *accountCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *accountCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.TODO()
	logger := log.FromContext(ctx).WithName("AccountCollector")

	logger.V(1).Info("collecting metrics")

	metrics, err := c.getAccountMetrics(ctx)
	if err != nil {
		logger.Error(err, "failed to collect account metrics")
		return
	}

	for account, data := range metrics.JobMetricsPer {
		ch <- prometheus.MustNewConstMetric(c.JobCount, prometheus.GaugeValue, float64(data.JobCount), account)
		// States
		ch <- prometheus.MustNewConstMetric(c.JobStates.BootFail, prometheus.GaugeValue, float64(data.JobStates.BootFail), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Cancelled, prometheus.GaugeValue, float64(data.JobStates.Cancelled), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Completed, prometheus.GaugeValue, float64(data.JobStates.Completed), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Deadline, prometheus.GaugeValue, float64(data.JobStates.Deadline), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Failed, prometheus.GaugeValue, float64(data.JobStates.Failed), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Pending, prometheus.GaugeValue, float64(data.JobStates.Pending), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Preempted, prometheus.GaugeValue, float64(data.JobStates.Preempted), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Running, prometheus.GaugeValue, float64(data.JobStates.Running), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Suspended, prometheus.GaugeValue, float64(data.JobStates.Suspended), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Timeout, prometheus.GaugeValue, float64(data.JobStates.Timeout), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.NodeFail, prometheus.GaugeValue, float64(data.JobStates.NodeFail), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.OutOfMemory, prometheus.GaugeValue, float64(data.JobStates.OutOfMemory), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Completing, prometheus.GaugeValue, float64(data.JobStates.Completing), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Configuring, prometheus.GaugeValue, float64(data.JobStates.Configuring), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.PowerUpNode, prometheus.GaugeValue, float64(data.JobStates.PowerUpNode), account)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Hold, prometheus.GaugeValue, float64(data.JobStates.Hold), account)
		// Tres
		ch <- prometheus.MustNewConstMetric(c.JobTres.CpusAlloc, prometheus.GaugeValue, float64(data.JobTres.CpusAlloc), account)
		ch <- prometheus.MustNewConstMetric(c.JobTres.MemoryAlloc, prometheus.GaugeValue, float64(data.JobTres.MemoryAlloc), account)
	}
}

func (c *accountCollector) getAccountMetrics(ctx context.Context) (*AccountMetrics, error) {
	jobList := &types.V0043JobInfoList{}
	if err := c.slurmClient.List(ctx, jobList); err != nil {
		return nil, err
	}
	metrics := calculateAccountMetrics(jobList)
	return metrics, nil
}

func calculateAccountMetrics(jobList *types.V0043JobInfoList) *AccountMetrics {
	metrics := &AccountMetrics{
		JobMetricsPer: make(map[string]*JobMetrics),
	}
	for _, job := range jobList.Items {
		key := ptr.Deref(job.Account, "")
		if _, ok := metrics.JobMetricsPer[key]; !ok {
			metrics.JobMetricsPer[key] = &JobMetrics{}
		}
		metrics.JobMetricsPer[key].JobCount++
		calculateJobState(&metrics.JobMetricsPer[key].JobStates, job)
		calculateJobTres(&metrics.JobMetricsPer[key].JobTres, job)
	}
	return metrics
}

type AccountMetrics struct {
	// Per Account
	JobMetricsPer map[string]*JobMetrics
}

// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"context"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/SlinkyProject/slurm-client/pkg/client"
	"github.com/SlinkyProject/slurm-client/pkg/types"
)

// Ref: https://prometheus.io/docs/practices/naming/#metric-names
func NewUserCollector(slurmClient client.Client) prometheus.Collector {
	return &userCollector{
		slurmClient: slurmClient,

		JobCount: prometheus.NewDesc("slurm_user_jobs_total", "Total number of user jobs", userLabels, nil),
		JobStates: jobStatesCollector{
			// Base States
			BootFail:    prometheus.NewDesc("slurm_user_jobs_bootfail_total", "Number of user jobs in BootFail state", userLabels, nil),
			Cancelled:   prometheus.NewDesc("slurm_user_jobs_cancelled_total", "Number of user jobs in Cancelled state", userLabels, nil),
			Completed:   prometheus.NewDesc("slurm_user_jobs_completed_total", "Number of user jobs in Completed state", userLabels, nil),
			Deadline:    prometheus.NewDesc("slurm_user_jobs_deadline_total", "Number of user jobs in Deadline state", userLabels, nil),
			Failed:      prometheus.NewDesc("slurm_user_jobs_failed_total", "Number of user jobs in Failed state", userLabels, nil),
			Pending:     prometheus.NewDesc("slurm_user_jobs_pending_total", "Number of user jobs in Pending state", userLabels, nil),
			Preempted:   prometheus.NewDesc("slurm_user_jobs_preempted_total", "Number of user jobs in Preempted state", userLabels, nil),
			Running:     prometheus.NewDesc("slurm_user_jobs_running_total", "Number of user jobs in Running state", userLabels, nil),
			Suspended:   prometheus.NewDesc("slurm_user_jobs_suspended_total", "Number of user jobs in Suspended state", userLabels, nil),
			Timeout:     prometheus.NewDesc("slurm_user_jobs_timeout_total", "Number of user jobs in Timeout state", userLabels, nil),
			NodeFail:    prometheus.NewDesc("slurm_user_jobs_nodefail_total", "Number of user jobs in NodeFail state", userLabels, nil),
			OutOfMemory: prometheus.NewDesc("slurm_user_jobs_outofmemory_total", "Number of user jobs in OutOfMemory state", userLabels, nil),
			// Flag States
			Completing:  prometheus.NewDesc("slurm_user_jobs_completing_total", "Number of user jobs with Completing flag", userLabels, nil),
			Configuring: prometheus.NewDesc("slurm_user_jobs_configuring_total", "Number of user jobs with Configuring flag", userLabels, nil),
			PowerUpNode: prometheus.NewDesc("slurm_user_jobs_powerupnode_total", "Number of user jobs with PowerUpNode flag", userLabels, nil),
			StageOut:    prometheus.NewDesc("slurm_user_jobs_stageout_total", "Number of user jobs with StageOut flag", userLabels, nil),
			// Other States
			Hold: prometheus.NewDesc("slurm_user_jobs_hold_total", "Number of user jobs with Hold flag", userLabels, nil),
		},
		JobTres: jobTresCollector{
			// CPUs
			CpusAlloc: prometheus.NewDesc("slurm_user_jobs_cpus_alloc_total", "Number of Allocated CPUs among user jobs", userLabels, nil),
			// Memory
			MemoryAlloc: prometheus.NewDesc("slurm_user_jobs_memory_alloc_bytes", "Amount of Allocated Memory (MB) among user jobs", userLabels, nil),
		},
	}
}

type userCollector struct {
	slurmClient client.Client

	JobCount  *prometheus.Desc
	JobStates jobStatesCollector
	JobTres   jobTresCollector
}

func (c *userCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *userCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.TODO()
	logger := log.FromContext(ctx).WithName("UserCollector")

	logger.V(1).Info("collecting metrics")

	metrics, err := c.getUserMetrics(ctx)
	if err != nil {
		logger.Error(err, "failed to collect user metrics")
		return
	}

	for userCtx, data := range metrics.JobMetricsPer {
		ch <- prometheus.MustNewConstMetric(c.JobCount, prometheus.GaugeValue, float64(data.JobCount), userCtx.UserId, userCtx.UserName)
		// States
		ch <- prometheus.MustNewConstMetric(c.JobStates.BootFail, prometheus.GaugeValue, float64(data.JobStates.BootFail), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Cancelled, prometheus.GaugeValue, float64(data.JobStates.Cancelled), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Completed, prometheus.GaugeValue, float64(data.JobStates.Completed), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Deadline, prometheus.GaugeValue, float64(data.JobStates.Deadline), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Failed, prometheus.GaugeValue, float64(data.JobStates.Failed), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Pending, prometheus.GaugeValue, float64(data.JobStates.Pending), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Preempted, prometheus.GaugeValue, float64(data.JobStates.Preempted), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Running, prometheus.GaugeValue, float64(data.JobStates.Running), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Suspended, prometheus.GaugeValue, float64(data.JobStates.Suspended), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Timeout, prometheus.GaugeValue, float64(data.JobStates.Timeout), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.NodeFail, prometheus.GaugeValue, float64(data.JobStates.NodeFail), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.OutOfMemory, prometheus.GaugeValue, float64(data.JobStates.OutOfMemory), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Completing, prometheus.GaugeValue, float64(data.JobStates.Completing), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Configuring, prometheus.GaugeValue, float64(data.JobStates.Configuring), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.PowerUpNode, prometheus.GaugeValue, float64(data.JobStates.PowerUpNode), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobStates.Hold, prometheus.GaugeValue, float64(data.JobStates.Hold), userCtx.UserId, userCtx.UserName)
		// Tres
		ch <- prometheus.MustNewConstMetric(c.JobTres.CpusAlloc, prometheus.GaugeValue, float64(data.JobTres.CpusAlloc), userCtx.UserId, userCtx.UserName)
		ch <- prometheus.MustNewConstMetric(c.JobTres.MemoryAlloc, prometheus.GaugeValue, float64(data.JobTres.MemoryAlloc), userCtx.UserId, userCtx.UserName)
	}
}

type UserContext struct {
	UserId   string
	UserName string
}

func (c *userCollector) getUserMetrics(ctx context.Context) (*UserMetrics, error) {
	jobList := &types.V0043JobInfoList{}
	if err := c.slurmClient.List(ctx, jobList); err != nil {
		return nil, err
	}
	metrics := calculateUserMetrics(jobList)
	return metrics, nil
}

func calculateUserMetrics(jobList *types.V0043JobInfoList) *UserMetrics {
	metrics := &UserMetrics{
		JobMetricsPer: make(map[UserContext]*JobMetrics),
	}
	for _, job := range jobList.Items {
		key := UserContext{
			UserId:   strconv.Itoa(int(ptr.Deref(job.UserId, 0))),
			UserName: ptr.Deref(job.UserName, ""),
		}
		if _, ok := metrics.JobMetricsPer[key]; !ok {
			metrics.JobMetricsPer[key] = &JobMetrics{}
		}
		metrics.JobMetricsPer[key].JobCount++
		calculateJobState(&metrics.JobMetricsPer[key].JobStates, job)
		calculateJobTres(&metrics.JobMetricsPer[key].JobTres, job)
	}
	return metrics
}

type UserMetrics struct {
	// Per User
	JobMetricsPer map[UserContext]*JobMetrics
}

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
)

// Ref: https://prometheus.io/docs/practices/naming/#metric-names
func NewJobCollector(slurmClient client.Client) prometheus.Collector {
	return &jobCollector{
		slurmClient: slurmClient,

		JobCount: prometheus.NewDesc("slurm_jobs_total", "Total number of jobs", nil, nil),
		JobStates: jobStatesCollector{
			// Base States
			BootFail:    prometheus.NewDesc("slurm_jobs_bootfail_total", "Number of jobs in BootFail state", nil, nil),
			Cancelled:   prometheus.NewDesc("slurm_jobs_cancelled_total", "Number of jobs in Cancelled state", nil, nil),
			Completed:   prometheus.NewDesc("slurm_jobs_completed_total", "Number of jobs in Completed state", nil, nil),
			Deadline:    prometheus.NewDesc("slurm_jobs_deadline_total", "Number of jobs in Deadline state", nil, nil),
			Failed:      prometheus.NewDesc("slurm_jobs_failed_total", "Number of jobs in Failed state", nil, nil),
			Pending:     prometheus.NewDesc("slurm_jobs_pending_total", "Number of jobs in Pending state", nil, nil),
			Preempted:   prometheus.NewDesc("slurm_jobs_preempted_total", "Number of jobs in Preempted state", nil, nil),
			Running:     prometheus.NewDesc("slurm_jobs_running_total", "Number of jobs in Running state", nil, nil),
			Suspended:   prometheus.NewDesc("slurm_jobs_suspended_total", "Number of jobs in Suspended state", nil, nil),
			Timeout:     prometheus.NewDesc("slurm_jobs_timeout_total", "Number of jobs in Timeout state", nil, nil),
			NodeFail:    prometheus.NewDesc("slurm_jobs_nodefail_total", "Number of jobs in NodeFail state", nil, nil),
			OutOfMemory: prometheus.NewDesc("slurm_jobs_outofmemory_total", "Number of jobs in OutOfMemory state", nil, nil),
			// Flag States
			Completing:  prometheus.NewDesc("slurm_jobs_completing_total", "Number of jobs with Completing flag", nil, nil),
			Configuring: prometheus.NewDesc("slurm_jobs_configuring_total", "Number of jobs with Configuring flag", nil, nil),
			PowerUpNode: prometheus.NewDesc("slurm_jobs_powerupnode_total", "Number of jobs with PowerUpNode flag", nil, nil),
			StageOut:    prometheus.NewDesc("slurm_jobs_stageout_total", "Number of jobs with StageOut flag", nil, nil),
			// Other States
			Hold: prometheus.NewDesc("slurm_jobs_hold_total", "Number of jobs with Hold flag", nil, nil),
		},
		JobTres: jobTresCollector{
			// CPUs
			CpusAlloc: prometheus.NewDesc("slurm_jobs_cpus_alloc_total", "Number of Allocated CPUs among jobs", nil, nil),
			// Memory
			MemoryAlloc: prometheus.NewDesc("slurm_jobs_memory_alloc_bytes", "Amount of Allocated Memory (MB) among jobs", nil, nil),
		},
	}
}

type jobCollector struct {
	slurmClient client.Client

	JobCount  *prometheus.Desc
	JobStates jobStatesCollector
	JobTres   jobTresCollector
}

type jobStatesCollector struct {
	// Base States
	BootFail    *prometheus.Desc
	Cancelled   *prometheus.Desc
	Completed   *prometheus.Desc
	Deadline    *prometheus.Desc
	Failed      *prometheus.Desc
	Pending     *prometheus.Desc
	Preempted   *prometheus.Desc
	Running     *prometheus.Desc
	Suspended   *prometheus.Desc
	Timeout     *prometheus.Desc
	NodeFail    *prometheus.Desc
	OutOfMemory *prometheus.Desc
	// Flag States
	Completing  *prometheus.Desc
	Configuring *prometheus.Desc
	PowerUpNode *prometheus.Desc
	StageOut    *prometheus.Desc
	// Other States
	Hold *prometheus.Desc
}

type jobTresCollector struct {
	// CPUs
	CpusAlloc *prometheus.Desc
	// Memory
	MemoryAlloc *prometheus.Desc
}

func (c *jobCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *jobCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.TODO()
	logger := log.FromContext(ctx).WithName("JobCollector")

	logger.V(1).Info("collecting metrics")

	metrics, err := c.getJobMetrics(ctx)
	if err != nil {
		logger.Error(err, "failed to collect job metrics")
		return
	}

	ch <- prometheus.MustNewConstMetric(c.JobCount, prometheus.GaugeValue, float64(metrics.JobCount))
	// States
	ch <- prometheus.MustNewConstMetric(c.JobStates.BootFail, prometheus.GaugeValue, float64(metrics.JobStates.BootFail))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Cancelled, prometheus.GaugeValue, float64(metrics.JobStates.Cancelled))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Completed, prometheus.GaugeValue, float64(metrics.JobStates.Completed))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Deadline, prometheus.GaugeValue, float64(metrics.JobStates.Deadline))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Failed, prometheus.GaugeValue, float64(metrics.JobStates.Failed))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Pending, prometheus.GaugeValue, float64(metrics.JobStates.Pending))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Preempted, prometheus.GaugeValue, float64(metrics.JobStates.Preempted))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Running, prometheus.GaugeValue, float64(metrics.JobStates.Running))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Suspended, prometheus.GaugeValue, float64(metrics.JobStates.Suspended))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Timeout, prometheus.GaugeValue, float64(metrics.JobStates.Timeout))
	ch <- prometheus.MustNewConstMetric(c.JobStates.NodeFail, prometheus.GaugeValue, float64(metrics.JobStates.NodeFail))
	ch <- prometheus.MustNewConstMetric(c.JobStates.OutOfMemory, prometheus.GaugeValue, float64(metrics.JobStates.OutOfMemory))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Completing, prometheus.GaugeValue, float64(metrics.JobStates.Completing))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Configuring, prometheus.GaugeValue, float64(metrics.JobStates.Configuring))
	ch <- prometheus.MustNewConstMetric(c.JobStates.PowerUpNode, prometheus.GaugeValue, float64(metrics.JobStates.PowerUpNode))
	ch <- prometheus.MustNewConstMetric(c.JobStates.StageOut, prometheus.GaugeValue, float64(metrics.JobStates.StageOut))
	ch <- prometheus.MustNewConstMetric(c.JobStates.Hold, prometheus.GaugeValue, float64(metrics.JobStates.Hold))
	// Tres
	ch <- prometheus.MustNewConstMetric(c.JobTres.CpusAlloc, prometheus.GaugeValue, float64(metrics.JobTres.CpusAlloc))
	ch <- prometheus.MustNewConstMetric(c.JobTres.MemoryAlloc, prometheus.GaugeValue, float64(metrics.JobTres.MemoryAlloc))
}

func (c *jobCollector) getJobMetrics(ctx context.Context) (*JobMetrics, error) {
	jobList := &types.V0043JobInfoList{}
	if err := c.slurmClient.List(ctx, jobList); err != nil {
		return nil, err
	}
	metrics := calculateJobMetrics(jobList)
	return metrics, nil
}

func calculateJobMetrics(jobList *types.V0043JobInfoList) *JobMetrics {
	metrics := &JobMetrics{
		JobCount: uint(len(jobList.Items)),
	}
	for _, job := range jobList.Items {
		calculateJobState(&metrics.JobStates, job)
		calculateJobTres(&metrics.JobTres, job)
	}
	return metrics
}

func calculateJobState(metrics *JobStates, job types.V0043JobInfo) {
	metrics.total++
	states := job.GetStateAsSet()
	// Base States
	switch {
	case states.Has(api.V0043JobInfoJobStateBOOTFAIL):
		metrics.BootFail++
	case states.Has(api.V0043JobInfoJobStateCANCELLED):
		metrics.Cancelled++
	case states.Has(api.V0043JobInfoJobStateCOMPLETED):
		metrics.Completed++
	case states.Has(api.V0043JobInfoJobStateDEADLINE):
		metrics.Deadline++
	case states.Has(api.V0043JobInfoJobStateFAILED):
		metrics.Failed++
	case states.Has(api.V0043JobInfoJobStatePENDING):
		metrics.Pending++
	case states.Has(api.V0043JobInfoJobStatePREEMPTED):
		metrics.Preempted++
	case states.Has(api.V0043JobInfoJobStateRUNNING):
		metrics.Running++
	case states.Has(api.V0043JobInfoJobStateSUSPENDED):
		metrics.Suspended++
	case states.Has(api.V0043JobInfoJobStateTIMEOUT):
		metrics.Timeout++
	case states.Has(api.V0043JobInfoJobStateNODEFAIL):
		metrics.NodeFail++
	case states.Has(api.V0043JobInfoJobStateOUTOFMEMORY):
		metrics.OutOfMemory++
	}
	// Flag States
	if states.Has(api.V0043JobInfoJobStateCOMPLETING) {
		metrics.Completing++
	}
	if states.Has(api.V0043JobInfoJobStateCONFIGURING) {
		metrics.Configuring++
	}
	if states.Has(api.V0043JobInfoJobStatePOWERUPNODE) {
		metrics.PowerUpNode++
	}
	if states.Has(api.V0043JobInfoJobStateSTAGEOUT) {
		metrics.StageOut++
	}
	// Other States
	if isHold := ptr.Deref(job.Hold, false); isHold {
		metrics.Hold++
	}
}

func calculateJobTres(metrics *JobTres, job types.V0043JobInfo) {
	metrics.total++
	res := getJobResourceAlloc(job)
	metrics.CpusAlloc += res.Cpus
	metrics.MemoryAlloc += res.Memory
}

type jobResources struct {
	Cpus   uint
	Memory uint
}

func getJobResourceAlloc(job types.V0043JobInfo) jobResources {
	var res jobResources
	jobRes := ptr.Deref(job.JobResources, api.V0043JobRes{})
	if jobRes.Nodes == nil {
		return res
	}
	jobResNode := ptr.Deref(jobRes.Nodes.Allocation, []api.V0043JobResNode{})
	for _, resNode := range jobResNode {
		if resNode.Cpus != nil {
			res.Cpus += uint(ptr.Deref(resNode.Cpus.Count, 0))
		}
		if resNode.Memory != nil {
			res.Memory += uint(ptr.Deref(resNode.Memory.Allocated, 0))
		}
	}
	return res
}

type JobMetrics struct {
	JobCount  uint
	JobStates JobStates
	JobTres   JobTres
}

// Ref: https://slurm.schedmd.com/job_state_codes.html#states
// Ref: https://slurm.schedmd.com/job_state_codes.html#flags
type JobStates struct {
	total uint
	// Base States
	BootFail    uint
	Cancelled   uint
	Completed   uint
	Deadline    uint
	Failed      uint
	Pending     uint
	Preempted   uint
	Running     uint
	Suspended   uint
	Timeout     uint
	NodeFail    uint
	OutOfMemory uint
	// Flag States
	Completing  uint
	Configuring uint
	PowerUpNode uint
	StageOut    uint
	// Other States
	Hold uint
}

type JobTres struct {
	total uint
	// CPUs
	CpusAlloc uint
	// Memory
	MemoryAlloc uint
}

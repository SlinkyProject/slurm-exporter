// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/SlinkyProject/slurm-client/api/v0041"
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
		// Individual Job State Metrics
		// Base States
		JobStateBootFail:    prometheus.NewDesc("slurm_job_state_bootfail", "The BootFail state of the job", jobLabels, nil),
		JobStateCancelled:   prometheus.NewDesc("slurm_job_state_cancelled", "The Cancelled state of the job", jobLabels, nil),
		JobStateCompleted:   prometheus.NewDesc("slurm_job_state_completed", "The Completed state of the job", jobLabels, nil),
		JobStateDeadline:    prometheus.NewDesc("slurm_job_state_deadline", "The Deadline state of the job", jobLabels, nil),
		JobStateFailed:      prometheus.NewDesc("slurm_job_state_failed", "The Failed state of the job", jobLabels, nil),
		JobStatePending:     prometheus.NewDesc("slurm_job_state_pending", "The Pending state of the job", jobLabels, nil),
		JobStatePreempted:   prometheus.NewDesc("slurm_job_state_preempted", "The Preempted state of the job", jobLabels, nil),
		JobStateRunning:     prometheus.NewDesc("slurm_job_state_running", "The Running state of the job", jobLabels, nil),
		JobStateSuspended:   prometheus.NewDesc("slurm_job_state_suspended", "The Suspended state of the job", jobLabels, nil),
		JobStateTimeout:     prometheus.NewDesc("slurm_job_state_timeout", "The Timeout state of the job", jobLabels, nil),
		JobStateNodeFail:    prometheus.NewDesc("slurm_job_state_nodefail", "The NodeFail state of the job", jobLabels, nil),
		JobStateOutOfMemory: prometheus.NewDesc("slurm_job_state_outofmemory", "The OutOfMemory state of the job", jobLabels, nil),
		// Flag States
		JobStateCompleting:  prometheus.NewDesc("slurm_job_state_completing", "The Completing state of the job", jobLabels, nil),
		JobStateConfiguring: prometheus.NewDesc("slurm_job_state_configuring", "The Configuring state of the job", jobLabels, nil),
		JobStatePowerUpNode: prometheus.NewDesc("slurm_job_state_powerupnode", "The PowerUpNode state of the job", jobLabels, nil),
		JobStateStageOut:    prometheus.NewDesc("slurm_job_state_stageout", "The StageOut state of the job", jobLabels, nil),
		// Other States
		JobStateHold: prometheus.NewDesc("slurm_job_state_hold", "The Hold state of the job", jobLabels, nil),
		// Tres
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
	// Individual Job State Metrics
	// Base States
	JobStateBootFail    *prometheus.Desc
	JobStateCancelled   *prometheus.Desc
	JobStateCompleted   *prometheus.Desc
	JobStateDeadline    *prometheus.Desc
	JobStateFailed      *prometheus.Desc
	JobStatePending     *prometheus.Desc
	JobStatePreempted   *prometheus.Desc
	JobStateRunning     *prometheus.Desc
	JobStateSuspended   *prometheus.Desc
	JobStateTimeout     *prometheus.Desc
	JobStateNodeFail    *prometheus.Desc
	JobStateOutOfMemory *prometheus.Desc
	// Flag States
	JobStateCompleting  *prometheus.Desc
	JobStateConfiguring *prometheus.Desc
	JobStatePowerUpNode *prometheus.Desc
	JobStateStageOut    *prometheus.Desc
	// Other States
	JobStateHold *prometheus.Desc
	// Tres
	JobTres jobTresCollector
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

	// Individual Job State Metrics - only emit when state is active (value = 1)
	for _, jobState := range metrics.JobIndividualStates {
		jobID := jobState.JobID
		jobName := jobState.JobName
		for _, node := range jobState.Nodes {
			// Base States
			if jobState.BootFail == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateBootFail, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.Cancelled == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateCancelled, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.Completed == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateCompleted, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.Deadline == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateDeadline, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.Failed == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateFailed, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.Pending == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStatePending, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.Preempted == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStatePreempted, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.Running == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateRunning, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.Suspended == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateSuspended, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.Timeout == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateTimeout, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.NodeFail == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateNodeFail, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.OutOfMemory == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateOutOfMemory, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			// Flag States
			if jobState.Completing == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateCompleting, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.Configuring == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateConfiguring, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.PowerUpNode == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStatePowerUpNode, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			if jobState.StageOut == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateStageOut, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
			// Other States
			if jobState.Hold == 1 {
				ch <- prometheus.MustNewConstMetric(c.JobStateHold, prometheus.GaugeValue, 1, jobID, jobName, node)
			}
		}
	}
}

func (c *jobCollector) getJobMetrics(ctx context.Context) (*JobMetrics, error) {
	jobList := &types.V0041JobInfoList{}
	if err := c.slurmClient.List(ctx, jobList); err != nil {
		return nil, err
	}
	metrics := calculateJobMetrics(jobList)
	return metrics, nil
}

func calculateJobMetrics(jobList *types.V0041JobInfoList) *JobMetrics {
	metrics := &JobMetrics{
		JobCount:            uint(len(jobList.Items)),
		JobIndividualStates: make([]JobIndividualStates, 0, len(jobList.Items)),
	}
	for _, job := range jobList.Items {
		calculateJobState(&metrics.JobStates, job)
		calculateJobTres(&metrics.JobTres, job)
		// Calculate individual job states
		jobStates := calculateJobIndividualStates(job)
		if jobStates != nil {
			metrics.JobIndividualStates = append(metrics.JobIndividualStates, *jobStates)
		}
	}
	return metrics
}

func calculateJobState(metrics *JobStates, job types.V0041JobInfo) {
	metrics.total++
	states := job.GetStateAsSet()
	// Base States
	switch {
	case states.Has(api.V0041JobInfoJobStateBOOTFAIL):
		metrics.BootFail++
	case states.Has(api.V0041JobInfoJobStateCANCELLED):
		metrics.Cancelled++
	case states.Has(api.V0041JobInfoJobStateCOMPLETED):
		metrics.Completed++
	case states.Has(api.V0041JobInfoJobStateDEADLINE):
		metrics.Deadline++
	case states.Has(api.V0041JobInfoJobStateFAILED):
		metrics.Failed++
	case states.Has(api.V0041JobInfoJobStatePENDING):
		metrics.Pending++
	case states.Has(api.V0041JobInfoJobStatePREEMPTED):
		metrics.Preempted++
	case states.Has(api.V0041JobInfoJobStateRUNNING):
		metrics.Running++
	case states.Has(api.V0041JobInfoJobStateSUSPENDED):
		metrics.Suspended++
	case states.Has(api.V0041JobInfoJobStateTIMEOUT):
		metrics.Timeout++
	case states.Has(api.V0041JobInfoJobStateNODEFAIL):
		metrics.NodeFail++
	case states.Has(api.V0041JobInfoJobStateOUTOFMEMORY):
		metrics.OutOfMemory++
	}
	// Flag States
	if states.Has(api.V0041JobInfoJobStateCOMPLETING) {
		metrics.Completing++
	}
	if states.Has(api.V0041JobInfoJobStateCONFIGURING) {
		metrics.Configuring++
	}
	if states.Has(api.V0041JobInfoJobStatePOWERUPNODE) {
		metrics.PowerUpNode++
	}
	if states.Has(api.V0041JobInfoJobStateSTAGEOUT) {
		metrics.StageOut++
	}
	// Other States
	if isHold := ptr.Deref(job.Hold, false); isHold {
		metrics.Hold++
	}
}

func calculateJobTres(metrics *JobTres, job types.V0041JobInfo) {
	metrics.total++
	res := getJobResourceAlloc(job)
	metrics.CpusAlloc += res.Cpus
	metrics.MemoryAlloc += res.Memory
}

type jobResources struct {
	Cpus   uint
	Memory uint
}

func getJobResourceAlloc(job types.V0041JobInfo) jobResources {
	var res jobResources
	jobRes := ptr.Deref(job.JobResources, api.V0041JobRes{})
	if jobRes.Nodes == nil {
		return res
	}
	jobResNode := ptr.Deref(jobRes.Nodes.Allocation, []api.V0041JobResNode{})
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
	JobCount            uint
	JobStates           JobStates
	JobTres             JobTres
	JobIndividualStates []JobIndividualStates
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

type JobIndividualStates struct {
	JobID   string
	JobName string
	Nodes   []string
	// Base States
	BootFail    int
	Cancelled   int
	Completed   int
	Deadline    int
	Failed      int
	Pending     int
	Preempted   int
	Running     int
	Suspended   int
	Timeout     int
	NodeFail    int
	OutOfMemory int
	// Flag States
	Completing  int
	Configuring int
	PowerUpNode int
	StageOut    int
	// Other States
	Hold int
}

func calculateJobIndividualStates(job types.V0041JobInfo) *JobIndividualStates {
	states := job.GetStateAsSet()
	jobID := fmt.Sprintf("%d", ptr.Deref(job.JobId, 0))
	jobName := ptr.Deref(job.Name, "")

	// Extract node list from job resources
	nodeList := ""
	if job.JobResources != nil && job.JobResources.Nodes != nil {
		nodeList = ptr.Deref(job.JobResources.Nodes.List, "")
	}
	nodes := parseNodeList(nodeList)

	// Even if there are no nodes, we still want to emit the metric, as it has the job_name and job_id
	if len(nodes) == 0 {
		nodes = []string{""}
	}

	jobStates := &JobIndividualStates{
		JobID:   jobID,
		JobName: jobName,
		Nodes:   nodes,
	}

	// Base States
	if states.Has(api.V0041JobInfoJobStateBOOTFAIL) {
		jobStates.BootFail = 1
	}
	if states.Has(api.V0041JobInfoJobStateCANCELLED) {
		jobStates.Cancelled = 1
	}
	if states.Has(api.V0041JobInfoJobStateCOMPLETED) {
		jobStates.Completed = 1
	}
	if states.Has(api.V0041JobInfoJobStateDEADLINE) {
		jobStates.Deadline = 1
	}
	if states.Has(api.V0041JobInfoJobStateFAILED) {
		jobStates.Failed = 1
	}
	if states.Has(api.V0041JobInfoJobStatePENDING) {
		jobStates.Pending = 1
	}
	if states.Has(api.V0041JobInfoJobStatePREEMPTED) {
		jobStates.Preempted = 1
	}
	if states.Has(api.V0041JobInfoJobStateRUNNING) {
		jobStates.Running = 1
	}
	if states.Has(api.V0041JobInfoJobStateSUSPENDED) {
		jobStates.Suspended = 1
	}
	if states.Has(api.V0041JobInfoJobStateTIMEOUT) {
		jobStates.Timeout = 1
	}
	if states.Has(api.V0041JobInfoJobStateNODEFAIL) {
		jobStates.NodeFail = 1
	}
	if states.Has(api.V0041JobInfoJobStateOUTOFMEMORY) {
		jobStates.OutOfMemory = 1
	}
	// Flag States
	if states.Has(api.V0041JobInfoJobStateCOMPLETING) {
		jobStates.Completing = 1
	}
	if states.Has(api.V0041JobInfoJobStateCONFIGURING) {
		jobStates.Configuring = 1
	}
	if states.Has(api.V0041JobInfoJobStatePOWERUPNODE) {
		jobStates.PowerUpNode = 1
	}
	if states.Has(api.V0041JobInfoJobStateSTAGEOUT) {
		jobStates.StageOut = 1
	}
	// Other States
	if isHold := ptr.Deref(job.Hold, false); isHold {
		jobStates.Hold = 1
	}

	return jobStates
}

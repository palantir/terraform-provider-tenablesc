package tenablesc

import (
	"fmt"
)

const jobEndpoint = "/job"

// Job represents the response structure for https://docs.tenable.com/tenablesc/api/Job.htm
type Job struct {
	BaseInfo
	AttemptNumber  ProbablyString `json:"attemptNumber,omitempty"`
	DependentJobID ProbablyString `json:"dependentJobID,omitempty"`
	ErrorCode      ProbablyString `json:"errorCode,omitempty"`
	ImmediateJob   FakeBool       `json:"immediateJob,omitempty"`
	Initiator      UserInfo       `json:"initiator,omitempty"`
	ObjectID       ProbablyString `json:"objectID,omitempty"`
	Organization   BaseInfo       `json:"organization,omitempty"`
	Params         string         `json:"params,omitempty"`
	Pid            ProbablyString `json:"pid,omitempty"`
	Priority       ProbablyString `json:"priority,omitempty"`
	StartTime      ProbablyString `json:"startTime,omitempty"`
	Status         string         `json:"status,omitempty"`
	TargetedTime   ProbablyString `json:"targetedTime,omitempty"`
	Type           string         `json:"type,omitempty"`
}

func (c *Client) GetAllJobs() ([]*Job, error) {

	var jobs []*Job
	_, err := c.getResource(jobEndpoint, &jobs)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func (c *Client) GetJob(id string) (*Job, error) {
	var job Job

	_, err := c.getResource(fmt.Sprintf("%s/%s", jobEndpoint, id), &job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (c *Client) KillJob(id string) error {
	_, err := c.postResource(fmt.Sprintf("%s/%s/kill", jobEndpoint, id), nil, nil)
	if err != nil {
		return err
	}

	return nil
}

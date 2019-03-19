/*
Copyright 2019 The Unity Scheduler Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"github.com/golang/glog"
	"github.infra.cloudera.com/yunikorn/scheduler-interface/lib/go/si"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func NewResource(memory int64, vcore int64) *si.Resource {
	return &si.Resource{
		Resources: map[string]*si.Quantity{
			Memory : { Value : memory},
			CPU : {Value : vcore},
		},
	}
}

func GetPodResource(pod *v1.Pod) (resource *si.Resource) {
	glog.V(4).Info("Get resource from pod spec")
	var memory, vcore = int64(0), int64(0)
	for _, c := range pod.Spec.Containers {
		resourceList := c.Resources.Requests
		m, c := ExplainResourceList(resourceList)
		memory += m
		vcore += c
	}
	glog.V(4).Infof("Parsed resource memory: %d, vcores: %d", memory, vcore)
	return NewResource(memory, vcore)
}

func GetNodeResource(nodeStatus *v1.NodeStatus) *si.Resource {
	memory, cpu := ExplainResourceList(nodeStatus.Allocatable)
	return NewResource(memory, cpu)
}


func ExplainResourceList(resourceList v1.ResourceList) (m int64, c int64) {
	var memory, vcore = int64(0), int64(0)
	for name, value := range resourceList {
		// log.Printf("parsing resource: resoueceName: %s, value: %s", name, value)
		switch name {
		case v1.ResourceMemory:
			//q, err := resource.ParseQuantity(value)
			//if err != nil {
			//	log.Printf("Unable to parse...")
			//}
			memory = value.ScaledValue(resource.Mega)
		case v1.ResourceCPU:
			// 500m -> 500
			// 1 -> 1000
			//q, err := resource.ParseQuantity(value)
			//if err != nil {
			//	log.Printf("Unable to parse %s, value %d",
			//		v1.ResourceCPU, value.Value())
			//	continue
			//}
			vcore = value.MilliValue()
		default:
			// log.Printf("ignore resource %s=%s", name, value)
			continue
		}
	}
	return memory, vcore
}

func ConvertRequest(jobId string, pod *v1.Pod) si.UpdateRequest {
	podResource := GetPodResource(pod)
	jobQueue := JobDefaultQueue
	if queueName, ok := pod.Labels[LabelQueueName]; ok {
		jobQueue = queueName
	}
	ask := si.AllocationAsk{
		AllocationKey: string(pod.UID),
		ResourceAsk:   podResource,
		QueueName:     jobQueue,
		JobId:         jobId,
		MaxAllocations: 1,
	}

	result := si.UpdateRequest{
		Asks:                 []*si.AllocationAsk {&ask},
		NewSchedulableNodes:  nil,
		UpdatedNodes:         nil,
		UtilizationReports:   nil,
		RmId: ClusterId,
	}

	return result
}

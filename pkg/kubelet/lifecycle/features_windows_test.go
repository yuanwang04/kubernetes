//go:build windows

/*
Copyright The Kubernetes Authors.

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

package lifecycle

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestPodFeaturesAdmitHandlerWindows(t *testing.T) {
	handler := NewPodFeaturesAdmitHandler()
	tests := []struct {
		name        string
		pod         *v1.Pod
		expectAdmit bool
		reason      string
	}{
		{
			name: "standard pod admitted",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: "c1"},
					},
				},
			},
			expectAdmit: true,
		},
		{
			name: "pod with pod-level resources declined",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Resources: &v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU: resource.MustParse("100m"),
						},
					},
					Containers: []v1.Container{
						{Name: "c1"},
					},
				},
			},
			expectAdmit: false,
			reason:      PodLevelResourcesNotAdmittedReason,
		},
		{
			name: "pod with container having RestartAllContainers declined",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "c1",
							RestartPolicyRules: []v1.ContainerRestartRule{
								{
									Action: v1.ContainerRestartRuleActionRestartAllContainers,
								},
							},
						},
					},
				},
			},
			expectAdmit: false,
			reason:      RestartAllContainersNotAdmittedReason,
		},
		{
			name: "pod with init container having RestartAllContainers declined",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name: "init-c1",
							RestartPolicyRules: []v1.ContainerRestartRule{
								{
									Action: v1.ContainerRestartRuleActionRestartAllContainers,
								},
							},
						},
					},
					Containers: []v1.Container{
						{Name: "c1"},
					},
				},
			},
			expectAdmit: false,
			reason:      RestartAllContainersNotAdmittedReason,
		},
		{
			name: "pod with container having Restart rule admitted",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "c1",
							RestartPolicyRules: []v1.ContainerRestartRule{
								{
									Action: v1.ContainerRestartRuleActionRestart,
								},
							},
						},
					},
				},
			},
			expectAdmit: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			attrs := &PodAdmitAttributes{Pod: tc.pod}
			res := handler.Admit(context.Background(), attrs)
			assert.Equal(t, tc.expectAdmit, res.Admit)
			if !tc.expectAdmit {
				assert.Equal(t, tc.reason, res.Reason)
			}
		})
	}
}

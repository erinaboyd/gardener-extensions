// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package general

import (
	"context"
	"fmt"
	"github.com/gardener/gardener/pkg/utils/kubernetes/health"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/gardener-extensions/pkg/controller/healthcheck"
)

// StatefulSetHealthChecker contains all the information for the StatefulSet HealthCheck
type StatefulSetHealthChecker struct {
	logger         logr.Logger
	seedClient     client.Client
	shootClient    client.Client
	deploymentName string
}

// CheckStatefulSet is a healthCheck function to check StatefulSets
func CheckStatefulSet(name string) healthcheck.HealthCheck {
	return &StatefulSetHealthChecker{
		deploymentName: name,
	}
}

// InjectSeedClient injects the seed client
func (healthChecker *StatefulSetHealthChecker) InjectSeedClient(seedClient client.Client) {
	healthChecker.seedClient = seedClient
}

// InjectShootClient injects the shoot client
func (healthChecker *StatefulSetHealthChecker) InjectShootClient(shootClient client.Client) {
	healthChecker.shootClient = shootClient
}

// SetLoggerSuffix injects the logger
func (healthChecker *StatefulSetHealthChecker) SetLoggerSuffix(provider, extension string) {
	healthChecker.logger = log.Log.WithName(fmt.Sprintf("%s-%s-healthcheck-deployment", provider, extension))
}

// Check executes the health check
func (healthChecker *StatefulSetHealthChecker) Check(ctx context.Context, request types.NamespacedName) (*healthcheck.SingleCheckResult, error) {
	statefulSet := &v1.StatefulSet{}
	if err := healthChecker.seedClient.Get(ctx, client.ObjectKey{Namespace: request.Namespace, Name: healthChecker.deploymentName}, statefulSet); err != nil {
		err := fmt.Errorf("failed to retrieve StatefulSet '%s' in namespace '%s': %v", healthChecker.deploymentName, request.Namespace, err)
		healthChecker.logger.Error(err, "Health check failed")
		return nil, err
	}
	if isHealthy, reason, err := statefulSetIsHealthy(statefulSet); !isHealthy {
		healthChecker.logger.Error(err, "Health check failed")
		return &healthcheck.SingleCheckResult{
			IsHealthy: false,
			Detail:    err.Error(),
			Reason:    *reason,
		}, nil
	}

	return &healthcheck.SingleCheckResult{
		IsHealthy: true,
	}, nil
}

func statefulSetIsHealthy(statefulSet *v1.StatefulSet) (bool, *string, error) {
	if err := health.CheckStatefulSet(statefulSet); err != nil {
		reason := "DeploymentUnhealthy"
		err := fmt.Errorf("statefulSet %s in namespace %s is unhealthy: %v", statefulSet.Name, statefulSet.Namespace, err)
		return false, &reason, err
	}
	return true, nil, nil
}

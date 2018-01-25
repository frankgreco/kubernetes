/*
Copyright 2017 The Kubernetes Authors.

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

package utils

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

// SetCRDCondition sets the status condition.  It either overwrites the existing one or
// creates a new one
func SetCRDCondition(crd *apiextensions.CustomResourceDefinition, newCondition apiextensions.CustomResourceDefinitionCondition) {
	existingCondition := FindCRDCondition(crd, newCondition.Type)
	if existingCondition == nil {
		newCondition.LastTransitionTime = metav1.NewTime(time.Now())
		crd.Status.Conditions = append(crd.Status.Conditions, newCondition)
		return
	}

	if existingCondition.Status != newCondition.Status {
		existingCondition.Status = newCondition.Status
		existingCondition.LastTransitionTime = newCondition.LastTransitionTime
	}

	existingCondition.Reason = newCondition.Reason
	existingCondition.Message = newCondition.Message
}

// RemoveCRDCondition removes the status condition.
func RemoveCRDCondition(crd *apiextensions.CustomResourceDefinition, conditionType apiextensions.CustomResourceDefinitionConditionType) {
	newConditions := []apiextensions.CustomResourceDefinitionCondition{}
	for _, condition := range crd.Status.Conditions {
		if condition.Type != conditionType {
			newConditions = append(newConditions, condition)
		}
	}
	crd.Status.Conditions = newConditions
}

// FindCRDCondition returns the condition you're looking for or nil
func FindCRDCondition(crd *apiextensions.CustomResourceDefinition, conditionType apiextensions.CustomResourceDefinitionConditionType) *apiextensions.CustomResourceDefinitionCondition {
	for i := range crd.Status.Conditions {
		if crd.Status.Conditions[i].Type == conditionType {
			return &crd.Status.Conditions[i]
		}
	}

	return nil
}

// IsCRDConditionTrue indicates if the condition is present and strictly true
func IsCRDConditionTrue(crd *apiextensions.CustomResourceDefinition, conditionType apiextensions.CustomResourceDefinitionConditionType) bool {
	return IsCRDConditionPresentAndEqual(crd, conditionType, apiextensions.ConditionTrue)
}

// IsCRDConditionFalse indicates if the condition is present and false true
func IsCRDConditionFalse(crd *apiextensions.CustomResourceDefinition, conditionType apiextensions.CustomResourceDefinitionConditionType) bool {
	return IsCRDConditionPresentAndEqual(crd, conditionType, apiextensions.ConditionFalse)
}

// IsCRDConditionPresentAndEqual indicates if the condition is present and equal to the arg
func IsCRDConditionPresentAndEqual(crd *apiextensions.CustomResourceDefinition, conditionType apiextensions.CustomResourceDefinitionConditionType, status apiextensions.ConditionStatus) bool {
	for _, condition := range crd.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == status
		}
	}
	return false
}

// IsCRDConditionEquivalent returns true if the lhs and rhs are equivalent except for times
func IsCRDConditionEquivalent(lhs, rhs *apiextensions.CustomResourceDefinitionCondition) bool {
	if lhs == nil && rhs == nil {
		return true
	}
	if lhs == nil || rhs == nil {
		return false
	}

	return lhs.Message == rhs.Message && lhs.Reason == rhs.Reason && lhs.Status == rhs.Status && lhs.Type == rhs.Type
}

// CRDHasFinalizer returns true if the finalizer is in the list
func CRDHasFinalizer(crd *apiextensions.CustomResourceDefinition, needle string) bool {
	for _, finalizer := range crd.Finalizers {
		if finalizer == needle {
			return true
		}
	}

	return false
}

// CRDRemoveFinalizer removes the finalizer if present
func CRDRemoveFinalizer(crd *apiextensions.CustomResourceDefinition, needle string) {
	newFinalizers := []string{}
	for _, finalizer := range crd.Finalizers {
		if finalizer != needle {
			newFinalizers = append(newFinalizers, finalizer)
		}
	}
	crd.Finalizers = newFinalizers
}

package utils

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
)

func init() {
	// In order to speed up tests, we're going to overwrite some packages variables
	retries, interval, factor = 1, 1*time.Millisecond, 1.0
}

var establishedCondition = apiextensionsv1beta1.CustomResourceDefinitionCondition{
	Type:    apiextensionsv1beta1.Established,
	Status:  apiextensionsv1beta1.ConditionTrue,
	Reason:  "InitialNamesAccepted",
	Message: "the initial names have been accepted",
}

var notEstablishedCondition = apiextensionsv1beta1.CustomResourceDefinitionCondition{
	Type:    apiextensionsv1beta1.Established,
	Status:  apiextensionsv1beta1.ConditionFalse,
	Reason:  "NotAccepted",
	Message: "not all names are accepted",
}

var acceptedCondition = apiextensionsv1beta1.CustomResourceDefinitionCondition{
	Type:    apiextensionsv1beta1.NamesAccepted,
	Status:  apiextensionsv1beta1.ConditionTrue,
	Reason:  "NoConflicts",
	Message: "no conflicts found",
}

var notAcceptedCondition = apiextensionsv1beta1.CustomResourceDefinitionCondition{
	Type:    apiextensionsv1beta1.NamesAccepted,
	Status:  apiextensionsv1beta1.ConditionFalse,
	Reason:  "TestConflict",
	Message: "conflicts found",
}

type crdBuilder struct {
	curr apiextensionsv1beta1.CustomResourceDefinition
}

func newCRD(name string) *crdBuilder {
	tokens := strings.SplitN(name, ".", 2)
	return &crdBuilder{
		curr: apiextensionsv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group: tokens[1],
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Plural: tokens[0],
				},
			},
		},
	}
}

func (b *crdBuilder) Condition(c apiextensionsv1beta1.CustomResourceDefinitionCondition) *crdBuilder {
	b.curr.Status.Conditions = append(b.curr.Status.Conditions, c)

	return b
}

func (b *crdBuilder) NewOrDie() *apiextensionsv1beta1.CustomResourceDefinition {
	return &b.curr
}

func assertErrors(expected, actual error) bool {
	if expected != nil {
		return actual != nil && expected.Error() == actual.Error()
	}
	return actual == nil
}

func TestEnsureCRDs(t *testing.T) {
	tests := []struct {
		name          string
		crds          []*apiextensionsv1beta1.CustomResourceDefinition
		expectedError error
	}{
		{
			name: "aggregate errors",
			crds: []*apiextensionsv1beta1.CustomResourceDefinition{newCRD("foos.bar.io").NewOrDie(), newCRD("bars.foo.io").NewOrDie()},
			expectedError: utilerrors.NewAggregate([]error{
				errors.New("CustomResourceDefinition bars.foo.io was not established within a reasonable amount of time."),
				errors.New("CustomResourceDefinition foos.bar.io was not established within a reasonable amount of time."),
			}),
		},
		{
			name:          "single established",
			crds:          []*apiextensionsv1beta1.CustomResourceDefinition{newCRD("foos.bar.io").Condition(establishedCondition).NewOrDie()},
			expectedError: nil,
		},
		{
			name:          "mulitple established",
			crds:          []*apiextensionsv1beta1.CustomResourceDefinition{newCRD("foos.bar.io").Condition(establishedCondition).NewOrDie(), newCRD("bars.foo.io").Condition(establishedCondition).NewOrDie()},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		cli := apiextensionsv1beta1client.NewSimpleClientset()
		if err := EnsureCRDs(cli.ApiextensionsV1beta1(), test.crds...); !assertErrors(test.expectedError, err) {
			t.Errorf("%v expected %v, got %v", test.name, test.expectedError, err)
		}
	}
}

func TestEnsureCRD(t *testing.T) {
	tests := []struct {
		name          string
		crd           *apiextensionsv1beta1.CustomResourceDefinition
		expectedError error
	}{
		{
			name:          "timeout",
			crd:           newCRD("foos.bar.io").NewOrDie(),
			expectedError: errors.New("CustomResourceDefinition foos.bar.io was not established within a reasonable amount of time."),
		},
		{
			name:          "established",
			crd:           newCRD("foos.bar.io").Condition(establishedCondition).NewOrDie(),
			expectedError: nil,
		},
		{
			name:          "will never be established",
			crd:           newCRD("foos.bar.io").Condition(notAcceptedCondition).Condition(notEstablishedCondition).NewOrDie(),
			expectedError: fmt.Errorf("Due to the naming conflict %s, the CustomResourceDefinition foos.bar.io will never become established", notAcceptedCondition.Reason),
		},
	}

	for _, test := range tests {
		cli := apiextensionsv1beta1client.NewSimpleClientset()
		if err := ensureCRD(cli.ApiextensionsV1beta1(), test.crd); !assertErrors(test.expectedError, err) {
			t.Errorf("%v expected %v, got %v", test.name, test.expectedError, err)
		}
	}
}

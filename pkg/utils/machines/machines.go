package machines

import (
	"context"
	"fmt"

	"github.com/golang/glog"

	mrv1 "kubevirt.io/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
	"kubevirt.io/machine-remediation-operator/pkg/consts"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	mapiv1 "sigs.k8s.io/cluster-api/pkg/apis/machine/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetMachineMachineDisruptionBudgets returns list of machine disruption budgets that suit for the machine
func GetMachineMachineDisruptionBudgets(c client.Client, machine *mapiv1.Machine) ([]*mrv1.MachineDisruptionBudget, error) {
	if len(machine.Labels) == 0 {
		return nil, fmt.Errorf("no MachineDisruptionBudgets found for machine %v because it has no labels", machine.Name)
	}

	list := &mrv1.MachineDisruptionBudgetList{}
	err := c.List(context.TODO(), list, client.InNamespace(machine.Namespace))
	if err != nil {
		return nil, err
	}

	var mdbs []*mrv1.MachineDisruptionBudget
	for i := range list.Items {
		mdb := &list.Items[i]
		selector, err := metav1.LabelSelectorAsSelector(mdb.Spec.Selector)
		if err != nil {
			glog.Warningf("invalid selector: %v", err)
			continue
		}

		// If a mdb with a nil or empty selector creeps in, it should match nothing, not everything.
		if selector.Empty() || !selector.Matches(labels.Set(machine.Labels)) {
			continue
		}
		mdbs = append(mdbs, mdb)
	}

	return mdbs, nil
}

// GetMachinesByLabelSelector returns machines that suit to the label selector
func GetMachinesByLabelSelector(c client.Client, selector *metav1.LabelSelector, namespace string) (*mapiv1.MachineList, error) {
	sel, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, err
	}
	if sel.Empty() {
		return nil, nil
	}

	machines := &mapiv1.MachineList{}
	listOptions := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: sel,
	}

	if err = c.List(context.TODO(), machines, client.UseListOptions(listOptions)); err != nil {
		return nil, err
	}
	return machines, nil
}

// GetNodeByMachine get the node object by machine object
func GetNodeByMachine(c client.Client, machine *mapiv1.Machine) (*v1.Node, error) {
	if machine.Status.NodeRef == nil {
		glog.Errorf("machine %s does not have NodeRef", machine.Name)
		return nil, fmt.Errorf("machine %s does not have NodeRef", machine.Name)
	}
	node := &v1.Node{}
	nodeKey := types.NamespacedName{
		Namespace: machine.Status.NodeRef.Namespace,
		Name:      machine.Status.NodeRef.Name,
	}
	if err := c.Get(context.TODO(), nodeKey, node); err != nil {
		return nil, err
	}
	return node, nil
}

// GetMachineByNode get the machine object by node object
func GetMachineByNode(c client.Client, node *v1.Node) (*mapiv1.Machine, error) {
	machineKey, ok := node.Annotations[consts.AnnotationMachine]
	if !ok {
		return nil, fmt.Errorf("No machine annotation for node %s", node.Name)
	}
	glog.Infof("Node %s is annotated with machine %s", node.Name, machineKey)

	machine := &mapiv1.Machine{}
	namespace, machineName, err := cache.SplitMetaNamespaceKey(machineKey)
	if err != nil {
		return nil, err
	}
	key := &types.NamespacedName{
		Namespace: namespace,
		Name:      machineName,
	}
	if err := c.Get(context.TODO(), *key, machine); err != nil {
		return nil, err
	}
	return machine, nil
}

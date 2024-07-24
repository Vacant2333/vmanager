package apis

import (
	"strconv"

	"k8s.io/apimachinery/pkg/labels"

	optimize_scheduling "vacant.sh/vmanager/pkg/definitions/optimize-scheduling"
)

type OptimizeSchedulingInfo struct {
	// Enable, target: optimize_scheduling.OptimizeSchedulingKey
	Enable bool
	// Strategy, target: optimize_scheduling.OptimizeSchedulingStrategyKey
	Strategy string
	// CustomOnDemand, target: optimize_scheduling.OptimizeSchedulingStrategyCustomOnDemandKey
	CustomOnDemand int

	TargetOnDemandNum int
	TargetOnSpotNum   int
}

func NewOptimizeSchedulingInfoFromLabels(labels labels.Set, replicaNum int) *OptimizeSchedulingInfo {
	osi := &OptimizeSchedulingInfo{}

	if labels == nil {
		return osi
	}

	// Get if enable the optimized scheduling.
	enable := labels.Get(optimize_scheduling.OptimizeSchedulingKey)
	if enable == "true" {
		osi.Enable = true
	}

	// Get the optimize scheduling strategy.
	osi.Strategy = labels.Get(optimize_scheduling.OptimizeSchedulingStrategyKey)

	// Get the custom on demand replica number.
	customOnDemandValue := labels.Get(optimize_scheduling.OptimizeSchedulingStrategyCustomOnDemandKey)
	osi.CustomOnDemand, _ = strconv.Atoi(customOnDemandValue)

	// Calculate the TargetOnDemandNum and TargetOnSpotNum by strategy.
	switch osi.Strategy {
	case optimize_scheduling.OptimizeSchedulingStrategyAllInOnDemand:
		osi.TargetOnDemandNum, osi.TargetOnSpotNum = replicaNum, 0
	case optimize_scheduling.OptimizeSchedulingStrategyAllInSpot:
		osi.TargetOnDemandNum, osi.TargetOnSpotNum = 0, replicaNum
	case optimize_scheduling.OptimizeSchedulingStrategyMajorityInOnDemand:
		osi.TargetOnDemandNum = (replicaNum / 2) + 1
		osi.TargetOnSpotNum = replicaNum - osi.TargetOnDemandNum
	case optimize_scheduling.OptimizeSchedulingStrategyCustom:
		osi.TargetOnDemandNum, osi.TargetOnSpotNum = osi.CustomOnDemand, replicaNum-osi.CustomOnDemand
	}

	return osi
}

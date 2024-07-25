package apis

import (
	"strconv"

	"k8s.io/apimachinery/pkg/labels"

	optimizescheduling "vacant.sh/vmanager/pkg/definitions/optimize-scheduling"
)

type OptimizeSchedulingSetting struct {
	// Enable, target: optimize_scheduling.OptimizeSchedulingKey
	Enable bool
	// Strategy, target: optimize_scheduling.OptimizeSchedulingStrategyKey
	Strategy string
	// CustomOnDemand, target: optimize_scheduling.OptimizeSchedulingStrategyCustomOnDemandKey
	CustomOnDemand int

	TargetOnDemandNum int
	TargetOnSpotNum   int
}

func NewOptimizeSchedulingSettingFromLabels(labels labels.Set, replicaNum int) *OptimizeSchedulingSetting {
	osi := &OptimizeSchedulingSetting{}

	if labels == nil {
		return osi
	}

	// Get if enable the optimized scheduling.
	enable := labels.Get(optimizescheduling.OptimizeSchedulingKey)
	if enable == "true" {
		osi.Enable = true
	}

	// Get the optimize scheduling strategy.
	osi.Strategy = labels.Get(optimizescheduling.OptimizeSchedulingStrategyKey)

	// Get the custom on demand replica number.
	customOnDemandValue := labels.Get(optimizescheduling.OptimizeSchedulingStrategyCustomOnDemandKey)
	osi.CustomOnDemand, _ = strconv.Atoi(customOnDemandValue)

	// Calculate the TargetOnDemandNum and TargetOnSpotNum by strategy.
	switch osi.Strategy {
	case optimizescheduling.OptimizeSchedulingStrategyAllInOnDemand:
		osi.TargetOnDemandNum, osi.TargetOnSpotNum = replicaNum, 0
	case optimizescheduling.OptimizeSchedulingStrategyAllInSpot:
		osi.TargetOnDemandNum, osi.TargetOnSpotNum = 0, replicaNum
	case optimizescheduling.OptimizeSchedulingStrategyMajorityInOnDemand:
		osi.TargetOnDemandNum = (replicaNum / 2) + 1
		osi.TargetOnSpotNum = replicaNum - osi.TargetOnDemandNum
	case optimizescheduling.OptimizeSchedulingStrategyCustom:
		osi.TargetOnDemandNum, osi.TargetOnSpotNum = osi.CustomOnDemand, replicaNum-osi.CustomOnDemand
	}

	return osi
}

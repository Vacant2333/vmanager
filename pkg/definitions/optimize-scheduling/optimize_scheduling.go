package optimize_scheduling

import "k8s.io/apimachinery/pkg/util/sets"

const (
	// OptimizeSchedulingKey defines whether optimization scheduling needs to be enabled for the load.
	// The value must be a boolean.
	OptimizeSchedulingKey = "vacant.sh/optimize-scheduling"

	// OptimizeSchedulingStrategyKey defines the strategy for optimization scheduling strategy.
	// If using OptimizeSchedulingStrategyCustom, the OptimizeSchedulingStrategyCustomOnDemandKey must be specified.
	OptimizeSchedulingStrategyKey                = "vacant.sh/optimize-scheduling-strategy"
	OptimizeSchedulingStrategyAllInOnDemand      = "all-in-on-demand"
	OptimizeSchedulingStrategyAllInSpot          = "all-in-spot"
	OptimizeSchedulingStrategyMajorityInOnDemand = "majority-in-on-demand"
	OptimizeSchedulingStrategyCustom             = "custom"

	// OptimizeSchedulingStrategyCustomOnDemandKey When `OptimizeSchedulingStrategy` is set to `custom`,
	// you can specify the minimum number of replicas that need to be on on-demand nodes.
	// This value must be greater than or equal to 0.
	OptimizeSchedulingStrategyCustomOnDemandKey = "vacant.sh/optimize-scheduling-strategy-custom-on-demand"
)

var OptimizeSchedulingStrategies = sets.NewString(
	OptimizeSchedulingStrategyAllInOnDemand,
	OptimizeSchedulingStrategyAllInSpot,
	OptimizeSchedulingStrategyMajorityInOnDemand,
	OptimizeSchedulingStrategyCustom,
)

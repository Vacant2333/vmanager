package optimize_scheduling

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateOptimizeSchedulingConfiguration checks whether the label configuration of a load meets the target requirements.
func ValidateOptimizeSchedulingConfiguration(labelSet labels.Set) field.ErrorList {
	if labelSet == nil {
		return nil
	}

	errs := field.ErrorList{}
	labelsPath := field.NewPath("labelSet")

	// Validate the optimizeScheduling value, must be a boolean.
	if labelSet.Has(OptimizeSchedulingKey) {
		optimizeSchedulingValue := labelSet.Get(OptimizeSchedulingKey)
		if optimizeSchedulingValue != "true" && optimizeSchedulingValue != "false" {
			errs = append(errs, field.Invalid(labelsPath.Key(OptimizeSchedulingKey),
				optimizeSchedulingValue, "value must be a boolean."))
		}
	}

	var optimizeSchedulingStrategyValue string
	// Validate the optimizeSchedulingStrategy value, must be in the OptimizeSchedulingStrategies.
	if labelSet.Has(OptimizeSchedulingStrategyKey) {
		optimizeSchedulingStrategyValue = labelSet.Get(OptimizeSchedulingStrategyKey)
		if !OptimizeSchedulingStrategies.Has(optimizeSchedulingStrategyValue) {
			errs = append(errs, field.Invalid(labelsPath.Key(OptimizeSchedulingStrategyKey),
				optimizeSchedulingStrategyValue, fmt.Sprintf("value must be in %v.", OptimizeSchedulingStrategies.List())))
		}
	}

	customOnDemandHas := labelSet.Has(OptimizeSchedulingStrategyCustomOnDemandKey)
	customOnDemandValue := labelSet.Get(OptimizeSchedulingStrategyCustomOnDemandKey)
	// Validate the optimizeSchedulingStrategyCustomOnDemand value must be greater than or equal to 0.
	if customOnDemandHas {
		if onDemandCount, err := strconv.Atoi(customOnDemandValue); err != nil {
			errs = append(errs, field.Invalid(labelsPath.Key(OptimizeSchedulingStrategyCustomOnDemandKey),
				customOnDemandValue, fmt.Sprintf("value must be a number.")))
		} else {
			if onDemandCount < 0 {
				errs = append(errs, field.Invalid(labelsPath.Key(OptimizeSchedulingStrategyCustomOnDemandKey),
					customOnDemandValue, fmt.Sprintf("value must be greater than or equal to 0.")))
			}
		}
	}

	// If using OptimizeSchedulingStrategyCustom, the OptimizeSchedulingStrategyCustomOnDemandKey must be specified.
	if optimizeSchedulingStrategyValue == OptimizeSchedulingStrategyCustom && !customOnDemandHas {
		errs = append(errs, field.Required(labelsPath.Key(OptimizeSchedulingStrategyCustomOnDemandKey),
			"value must not be empty when the strategy is custom."))
	}

	return errs
}

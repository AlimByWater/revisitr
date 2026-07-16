package lunch

import (
	"fmt"

	"revisitr/internal/entity"
)

type LunchPriceItem struct {
	MenuItemPrice float64
	Surcharge     float64
}

type LunchPriceInput struct {
	PriceMode string
	BasePrice float64
	Items     []LunchPriceItem // one per selected course
}

// CalculateTotal computes the order total for the given price mode:
// fixed → BasePrice; sum_of_items → Σ menu prices;
// base_plus_surcharge → BasePrice + Σ surcharges.
func CalculateTotal(in LunchPriceInput) (float64, error) {
	switch in.PriceMode {
	case entity.LunchPriceFixed:
		return in.BasePrice, nil
	case entity.LunchPriceSumOfItems:
		var total float64
		for _, item := range in.Items {
			total += item.MenuItemPrice
		}
		return total, nil
	case entity.LunchPriceBasePlusSurcharge:
		total := in.BasePrice
		for _, item := range in.Items {
			total += item.Surcharge
		}
		return total, nil
	default:
		return 0, fmt.Errorf("unknown price mode %q", in.PriceMode)
	}
}

// MinTotal returns the cheapest possible total for a format — used for
// "от N ₽" labels. Courses must be the format's courses; unavailable items
// are expected to be filtered out by the caller.
func MinTotal(format entity.LunchFormat, courses []entity.LunchCourse) float64 {
	switch format.PriceMode {
	case entity.LunchPriceFixed:
		return format.BasePrice
	case entity.LunchPriceSumOfItems:
		var total float64
		for _, course := range courses {
			total += minCourseValue(course, func(it entity.LunchCourseItem) float64 {
				if it.MenuItem == nil {
					return 0
				}
				return it.MenuItem.Price
			})
		}
		return total
	case entity.LunchPriceBasePlusSurcharge:
		total := format.BasePrice
		for _, course := range courses {
			total += minCourseValue(course, func(it entity.LunchCourseItem) float64 {
				return it.Surcharge
			})
		}
		return total
	default:
		return 0
	}
}

func minCourseValue(course entity.LunchCourse, value func(entity.LunchCourseItem) float64) float64 {
	if len(course.Items) == 0 {
		return 0
	}
	min := value(course.Items[0])
	for _, item := range course.Items[1:] {
		if v := value(item); v < min {
			min = v
		}
	}
	return min
}

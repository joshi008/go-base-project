package services

func GetPricingStrategy(strategy string) PricingStrategy {
	switch strategy {
	case "dynamic":
		return &DynamicPricingStrategy{}
	default:
		return &DefaultPricingStrategy{}
	}
}

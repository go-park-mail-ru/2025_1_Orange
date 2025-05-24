package dto

// easyjson:json
type SpecializationNamesResponse struct {
	Names []string `json:"specializations"`
}

// easyjson:json
type SpecializationSalaryRangesResponse struct {
	Specializations []SpecializationSalaryRange `json:"specializations"`
}

// easyjson:json
type SpecializationSalaryRange struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	MinSalary int    `json:"minSalary"`
	MaxSalary int    `json:"maxSalary"`
	AvgSalary int    `json:"avgSalary"`
}

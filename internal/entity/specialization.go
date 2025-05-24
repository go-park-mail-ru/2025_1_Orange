package entity

type Specialization struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SpecializationSalaryRange struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	MinSalary int    `json:"minSalary"`
	MaxSalary int    `json:"maxSalary"`
	AvgSalary int    `json:"avgSalary"`
}

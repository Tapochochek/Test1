package competitions

type Competition struct {
	ID              int64  `json:"id"`
	Title           string `json:"title"`
	Discipline      string `json:"discipline"`
	CompetitionType string `json:"competition_type"`
	AgeGroup        string `json:"age_group"`
	StartsAt        string `json:"starts_at"`
	Location        string `json:"location"`
	Status          string `json:"status"`
	Participants    int    `json:"participants"`
}

type CreateInput struct {
	Title           string `json:"title"`
	Discipline      string `json:"discipline"`
	CompetitionType string `json:"competition_type"`
	AgeGroup        string `json:"age_group"`
	StartsAt        string `json:"starts_at"`
	Location        string `json:"location"`
}

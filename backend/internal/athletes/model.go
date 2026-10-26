package athletes

type Athlete struct {
	ID           int64  `json:"id"`
	FullName     string `json:"full_name"`
	Organization string `json:"organization"`
	Discipline   string `json:"discipline"`
	AgeGroup     string `json:"age_group"`
	Points       int    `json:"points"`
	CoachName    string `json:"coach_name"`
	Status       string `json:"status"`
}

package httpserver

import "net/http"

type Routes struct {
	Health            http.HandlerFunc
	Dashboard         http.HandlerFunc
	Competitions      http.HandlerFunc
	CreateCompetition http.HandlerFunc
	Athletes          http.HandlerFunc
	Rating            http.HandlerFunc
	Team              http.HandlerFunc
	Notifications     http.HandlerFunc
	MaxStatus         http.HandlerFunc
	MaxWebhook        http.HandlerFunc
}

func NewRouter(routes Routes) http.Handler {
	mux := http.NewServeMux()
	register(mux, "GET /api/health", routes.Health)
	register(mux, "GET /api/dashboard", routes.Dashboard)
	register(mux, "GET /api/competitions", routes.Competitions)
	register(mux, "POST /api/competitions", routes.CreateCompetition)
	register(mux, "GET /api/athletes", routes.Athletes)
	register(mux, "GET /api/rating", routes.Rating)
	register(mux, "GET /api/team", routes.Team)
	register(mux, "GET /api/notifications", routes.Notifications)
	register(mux, "GET /api/integrations/max/status", routes.MaxStatus)
	register(mux, "POST /api/integrations/max/webhook", routes.MaxWebhook)
	return CORS(mux)
}

func register(mux *http.ServeMux, pattern string, handler http.HandlerFunc) {
	if handler == nil {
		return
	}
	mux.HandleFunc(pattern, handler)
}

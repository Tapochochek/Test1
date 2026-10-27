package httpserver

import (
	"net/http"

	"orenkod/backend/internal/auth"
	"orenkod/backend/internal/rules"
)

type Routes struct {
	Health            http.HandlerFunc
	Dashboard         http.HandlerFunc
	Competitions      http.HandlerFunc
	CreateCompetition http.HandlerFunc
	Athletes          http.HandlerFunc
	Rating            http.HandlerFunc
	Team              http.HandlerFunc
	Notifications     http.HandlerFunc
	TelegramStatus    http.HandlerFunc
	Auth              *auth.Service
	Platform          PlatformRoutes
}

type PlatformRoutes struct {
	Login, Register, Logout, Me, Switch, CreateCompetition                                                                                                                                                                                http.HandlerFunc
	Catalogs, PublicHome, Verify                                                                                                                                                                                                          http.HandlerFunc
	Publish, Apply, Applications, DecideApp                                                                                                                                                                                               http.HandlerFunc
	RequestLink, DecideLink, Links                                                                                                                                                                                                        http.HandlerFunc
	Judges, AssignJudge, Assignments                                                                                                                                                                                                      http.HandlerFunc
	SaveResults, Protocol, History                                                                                                                                                                                                        http.HandlerFunc
	RankApply, RankList, RankDecide                                                                                                                                                                                                       http.HandlerFunc
	CatApply, CatList, CatDecide                                                                                                                                                                                                          http.HandlerFunc
	TeamSave, TeamOrder                                                                                                                                                                                                                   http.HandlerFunc
	Enrollment, AthleteReport, RegistryXLSX, AddType, Development, Audit, Certificate, Inbox, Rating, TeamGet, ApproveAthlete                                                                                                             http.HandlerFunc
	TokenCreate, IntegrationsList, WebhookCreate, ExtCompetitions, ExtRating, ExtImportResults, OIDCDiscovery, OIDCAuthorizePage, OIDCAuthorizeInfo, OIDCApprove, OIDCToken, OIDCUserinfo, OIDCDemoPage, OIDCClientCreate, OIDCClientList http.HandlerFunc
	Premiums, PremiumStatus, PremiumScale, Norms, SaveNorm, SaveNormResult, Attestation, Promote                                                                                                                                          http.HandlerFunc
	TelegramLinkCode, TelegramLinkStatus, TelegramUnlink                                                                                                                                                                                  http.HandlerFunc
}

func NewRouter(routes Routes) http.Handler {
	mux := http.NewServeMux()
	register(mux, "GET /api/health", routes.Health)
	register(mux, "GET /api/public/home", routes.Platform.PublicHome)
	register(mux, "GET /api/public/catalogs", routes.Platform.Catalogs)
	register(mux, "GET /api/verify/{code}", routes.Platform.Verify)
	register(mux, "GET /api/oidc/.well-known/openid-configuration", routes.Platform.OIDCDiscovery)
	register(mux, "GET /api/oidc/authorize", routes.Platform.OIDCAuthorizePage)
	register(mux, "POST /api/oidc/token", routes.Platform.OIDCToken)
	register(mux, "GET /api/oidc/userinfo", routes.Platform.OIDCUserinfo)
	register(mux, "GET /api/oidc/demo", routes.Platform.OIDCDemoPage)
	register(mux, "GET /api/ext/competitions", routes.Platform.ExtCompetitions)
	register(mux, "GET /api/ext/rating", routes.Platform.ExtRating)
	register(mux, "POST /api/ext/competitions/{id}/results", routes.Platform.ExtImportResults)
	register(mux, "POST /api/auth/login", routes.Platform.Login)
	register(mux, "POST /api/auth/register", routes.Platform.Register)
	register(mux, "GET /api/integrations/telegram/status", routes.TelegramStatus)

	if routes.Auth != nil {
		anyUser := routes.Auth.Middleware()
		fed := routes.Auth.Middleware(rules.RoleFederation)
		fedCoach := routes.Auth.Middleware(rules.RoleFederation, rules.RoleCoach)
		coach := routes.Auth.Middleware(rules.RoleCoach)
		athlete := routes.Auth.Middleware(rules.RoleAthlete)
		judge := routes.Auth.Middleware(rules.RoleJudge)
		fedOrJudge := routes.Auth.Middleware(rules.RoleFederation, rules.RoleJudge)
		fedCoachAthlete := routes.Auth.Middleware(rules.RoleFederation, rules.RoleCoach, rules.RoleAthlete)
		coachAthlete := routes.Auth.Middleware(rules.RoleCoach, rules.RoleAthlete)

		register(mux, "POST /api/auth/logout", anyUser(routes.Platform.Logout))
		register(mux, "GET /api/auth/me", anyUser(routes.Platform.Me))
		register(mux, "GET /api/integrations/telegram/link-status", anyUser(routes.Platform.TelegramLinkStatus))
		register(mux, "POST /api/integrations/telegram/link-code", anyUser(routes.Platform.TelegramLinkCode))
		register(mux, "DELETE /api/integrations/telegram/link", anyUser(routes.Platform.TelegramUnlink))
		register(mux, "POST /api/auth/switch-role", anyUser(routes.Platform.Switch))
		register(mux, "GET /api/dashboard", anyUser(routes.Dashboard))
		register(mux, "GET /api/competitions", anyUser(routes.Competitions))
		register(mux, "POST /api/competitions", fed(routes.Platform.CreateCompetition))
		register(mux, "POST /api/competitions/{id}/publish", fed(routes.Platform.Publish))
		register(mux, "POST /api/competitions/{id}/apply", coachAthlete(routes.Platform.Apply))
		register(mux, "GET /api/competitions/{id}/applications", fed(routes.Platform.Applications))
		register(mux, "GET /api/applications", fed(routes.Platform.Applications))
		register(mux, "POST /api/applications/{id}/decide", fed(routes.Platform.DecideApp))
		register(mux, "GET /api/athletes", fedCoach(routes.Athletes))
		register(mux, "POST /api/athletes/{id}/approve", fed(routes.Platform.ApproveAthlete))
		register(mux, "GET /api/rating", anyUser(routes.Platform.Rating))
		register(mux, "GET /api/team", anyUser(routes.Platform.TeamGet))
		register(mux, "POST /api/team", fed(routes.Platform.TeamSave))
		register(mux, "GET /api/team/order", fed(routes.Platform.TeamOrder))
		register(mux, "GET /api/notifications", anyUser(routes.Platform.Inbox))
		register(mux, "POST /api/coach-links", coach(routes.Platform.RequestLink))
		register(mux, "GET /api/coach-links", anyUser(routes.Platform.Links))
		register(mux, "POST /api/coach-links/{id}/decide", athlete(routes.Platform.DecideLink))
		register(mux, "GET /api/judges", fed(routes.Platform.Judges))
		register(mux, "GET /api/oidc/authorize/info", anyUser(routes.Platform.OIDCAuthorizeInfo))
		register(mux, "POST /api/oidc/authorize/approve", anyUser(routes.Platform.OIDCApprove))
		register(mux, "POST /api/oidc/clients", fed(routes.Platform.OIDCClientCreate))
		register(mux, "GET /api/oidc/clients", fed(routes.Platform.OIDCClientList))
		register(mux, "GET /api/integrations", fed(routes.Platform.IntegrationsList))
		register(mux, "POST /api/integrations/tokens", fed(routes.Platform.TokenCreate))
		register(mux, "POST /api/integrations/webhooks", fed(routes.Platform.WebhookCreate))
		register(mux, "GET /api/premiums", fed(routes.Platform.Premiums))
		register(mux, "POST /api/premiums/status", fed(routes.Platform.PremiumStatus))
		register(mux, "POST /api/premiums/scale", fed(routes.Platform.PremiumScale))
		register(mux, "GET /api/norms", anyUser(routes.Platform.Norms))
		register(mux, "POST /api/norms", fed(routes.Platform.SaveNorm))
		register(mux, "POST /api/norm-results", fedCoach(routes.Platform.SaveNormResult))
		register(mux, "GET /api/attestation", anyUser(routes.Platform.Attestation))
		register(mux, "POST /api/attestation/{id}/promote", fedCoach(routes.Platform.Promote))
		register(mux, "POST /api/competitions/{id}/judges", fed(routes.Platform.AssignJudge))
		register(mux, "GET /api/assignments", fedOrJudge(routes.Platform.Assignments))
		register(mux, "POST /api/competitions/{id}/results", fedOrJudge(routes.Platform.SaveResults))
		register(mux, "GET /api/competitions/{id}/protocol", anyUser(routes.Platform.Protocol))
		register(mux, "GET /api/results", anyUser(routes.Platform.History))
		register(mux, "POST /api/ranks", coachAthlete(routes.Platform.RankApply))
		register(mux, "GET /api/ranks", fedCoachAthlete(routes.Platform.RankList))
		register(mux, "POST /api/ranks/{id}/decide", fed(routes.Platform.RankDecide))
		register(mux, "POST /api/categories", judge(routes.Platform.CatApply))
		register(mux, "GET /api/categories", routes.Auth.Middleware(rules.RoleFederation, rules.RoleJudge)(routes.Platform.CatList))
		register(mux, "POST /api/categories/{id}/decide", fed(routes.Platform.CatDecide))
		register(mux, "GET /api/documents/enrollment", fed(routes.Platform.Enrollment))
		register(mux, "GET /api/documents/registry.xlsx", fed(routes.Platform.RegistryXLSX))
		register(mux, "GET /api/documents/athletes/{id}", anyUser(routes.Platform.AthleteReport))
		register(mux, "POST /api/catalogs/types", fed(routes.Platform.AddType))
		register(mux, "GET /api/development", fed(routes.Platform.Development))
		register(mux, "GET /api/audit", fed(routes.Platform.Audit))
		register(mux, "POST /api/certificates", athlete(routes.Platform.Certificate))
	} else {
		register(mux, "GET /api/dashboard", routes.Dashboard)
		register(mux, "GET /api/competitions", routes.Competitions)
		register(mux, "POST /api/competitions", routes.CreateCompetition)
		register(mux, "GET /api/athletes", routes.Athletes)
		register(mux, "GET /api/rating", routes.Rating)
		register(mux, "GET /api/team", routes.Team)
		register(mux, "GET /api/notifications", routes.Notifications)
	}
	return CORS(mux)
}

func register(mux *http.ServeMux, pattern string, handler http.HandlerFunc) {
	if handler == nil {
		return
	}
	mux.HandleFunc(pattern, handler)
}

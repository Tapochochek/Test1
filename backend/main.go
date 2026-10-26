package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/001_init.sql
var migrations embed.FS

var db *sql.DB

const maxAPI = "https://platform-api2.max.ru"

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

type RatingEntry struct {
	Rank         int    `json:"rank"`
	Athlete      string `json:"athlete"`
	Organization string `json:"organization"`
	Discipline   string `json:"discipline"`
	AgeGroup     string `json:"age_group"`
	Points       int    `json:"points"`
}

type Notification struct {
	ID      int    `json:"id"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Message string `json:"message"`
	When    string `json:"when"`
}

func main() {
	port := env("PORT", "8080")
	if err := connectDB(); err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", healthHandler)
	mux.HandleFunc("GET /api/dashboard", dashboardHandler)
	mux.HandleFunc("GET /api/competitions", competitionsHandler)
	mux.HandleFunc("POST /api/competitions", createCompetitionHandler)
	mux.HandleFunc("GET /api/athletes", athletesHandler)
	mux.HandleFunc("GET /api/rating", ratingHandler)
	mux.HandleFunc("GET /api/team", teamHandler)
	mux.HandleFunc("GET /api/notifications", notificationsHandler)
	mux.HandleFunc("GET /api/integrations/max/status", maxStatusHandler)
	mux.HandleFunc("POST /api/integrations/max/webhook", maxWebhookHandler)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           cors(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("ОренКод API слушает :%s", port)
	log.Fatal(server.ListenAndServe())
}

func connectDB() error {
	dsn := env("DATABASE_URL", "postgres://orenkod:orenkod_local@localhost:5432/orenkod?sslmode=disable")
	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
	defer cancel()
	for {
		if err = db.PingContext(ctx); err == nil {
			break
		}
		if ctx.Err() != nil {
			return fmt.Errorf("can't connect to PostgreSQL: %w", err)
		}
		time.Sleep(2 * time.Second)
	}

	schema, err := migrations.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, string(schema)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return seed(ctx)
}

func seed(ctx context.Context) error {
	competitions := []struct {
		title, discipline, kind, age, starts, location, status string
		participants                                           int
	}{
		{"Кубок Оренбуржья по спортивному программированию", "Программирование", "Региональный", "Юниоры", "2026-11-08T10:00:00+05:00", "Оренбург · ОГУ", "Регистрация", 86},
		{"Осенний алгоритмический спринт", "Алгоритмическое программирование", "Открытый", "Все группы", "2026-11-16T11:00:00+05:00", "Онлайн", "Регистрация", 124},
		{"Отбор в сборную области", "Программирование систем", "Отборочный", "Юниоры и взрослые", "2026-11-29T09:30:00+05:00", "Оренбург · Точка кипения", "Планируется", 0},
		{"Практикум по робототехнике", "Робототехника", "Тренировочный", "Юниоры", "2026-12-06T12:00:00+05:00", "Орск · IT-центр", "Планируется", 0},
		{"Региональная олимпиада по ИБ", "Информационная безопасность", "Региональный", "Все группы", "2026-10-10T10:00:00+05:00", "Оренбург · ОГУ", "Завершено", 59},
		{"Летняя школа разработки", "Программная инженерия", "Образовательный", "Юниоры", "2026-07-20T10:00:00+05:00", "Бузулук", "Завершено", 42},
	}
	for _, c := range competitions {
		_, err := db.ExecContext(ctx, `INSERT INTO competitions(title, discipline, competition_type, age_group, starts_at, location, status, participants)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (title) DO NOTHING`, c.title, c.discipline, c.kind, c.age, c.starts, c.location, c.status, c.participants)
		if err != nil {
			return err
		}
	}

	athletes := []struct {
		name, org, discipline, age, coach, status string
		points                                    int
	}{
		{"Алексей Морозов", "Лицей № 1", "Алгоритмическое программирование", "Юниоры", "Мария Крылова", "Активен", 968},
		{"Дарья Соколова", "ОГУ", "Программирование", "Взрослые", "Илья Волков", "Активен", 914},
		{"Максим Орлов", "Гимназия № 2", "Робототехника", "Юниоры", "Мария Крылова", "Активен", 887},
		{"Полина Васильева", "ОГУ", "Информационная безопасность", "Взрослые", "Илья Волков", "Активен", 841},
		{"Артём Кузнецов", "Лицей № 1", "Программирование", "Юниоры", "Мария Крылова", "Активен", 812},
		{"София Попова", "IT-центр Орск", "Программная инженерия", "Юниоры", "Никита Егоров", "Активен", 779},
		{"Иван Лебедев", "ОГУ", "Алгоритмическое программирование", "Взрослые", "Илья Волков", "Активен", 745},
		{"Ева Фёдорова", "Гимназия № 2", "Робототехника", "Юниоры", "Мария Крылова", "Активен", 702},
		{"Кирилл Павлов", "IT-центр Орск", "Информационная безопасность", "Взрослые", "Никита Егоров", "Активен", 684},
		{"Алиса Семёнова", "Лицей № 1", "Программная инженерия", "Юниоры", "Мария Крылова", "Требует проверки", 633},
	}
	for _, a := range athletes {
		_, err := db.ExecContext(ctx, `INSERT INTO athletes(full_name, organization, discipline, age_group, points, coach_name, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (full_name, organization) DO NOTHING`, a.name, a.org, a.discipline, a.age, a.points, a.coach, a.status)
		if err != nil {
			return err
		}
	}
	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.PingContext(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "degraded", "database": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "database": true, "service": "orenkod-api"})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	var athletes, competitions, activeCompetitions, applications int
	_ = db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM athletes`).Scan(&athletes)
	_ = db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM competitions`).Scan(&competitions)
	_ = db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM competitions WHERE status IN ('Регистрация','Планируется')`).Scan(&activeCompetitions)
	_ = db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(participants),0) FROM competitions WHERE status IN ('Регистрация','Планируется')`).Scan(&applications)

	items, _ := listCompetitions(r.Context(), true)
	rating, _ := listRating(r.Context(), "", 5)
	writeJSON(w, http.StatusOK, map[string]any{
		"stats":              map[string]int{"athletes": athletes, "competitions": activeCompetitions, "applications": applications, "judges": 24},
		"competitions_total": competitions,
		"competitions":       items,
		"top_athletes":       rating,
		"updated_at":         time.Now().Format(time.RFC3339),
	})
}

func competitionsHandler(w http.ResponseWriter, r *http.Request) {
	items, err := listCompetitions(r.Context(), false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось получить соревнования")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func listCompetitions(ctx context.Context, upcomingOnly bool) ([]Competition, error) {
	query := `SELECT id,title,discipline,competition_type,age_group,starts_at,location,status,participants FROM competitions ORDER BY (starts_at < NOW()) ASC, starts_at ASC`
	if upcomingOnly {
		query = `SELECT id,title,discipline,competition_type,age_group,starts_at,location,status,participants FROM competitions WHERE starts_at >= NOW() - INTERVAL '1 day' ORDER BY starts_at ASC LIMIT 4`
	}
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Competition{}
	for rows.Next() {
		var c Competition
		var starts time.Time
		if err := rows.Scan(&c.ID, &c.Title, &c.Discipline, &c.CompetitionType, &c.AgeGroup, &starts, &c.Location, &c.Status, &c.Participants); err != nil {
			return nil, err
		}
		c.StartsAt = starts.Format(time.RFC3339)
		items = append(items, c)
	}
	return items, rows.Err()
}

func createCompetitionHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var input struct {
		Title           string `json:"title"`
		Discipline      string `json:"discipline"`
		CompetitionType string `json:"competition_type"`
		AgeGroup        string `json:"age_group"`
		StartsAt        string `json:"starts_at"`
		Location        string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Проверьте формат JSON")
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Discipline = strings.TrimSpace(input.Discipline)
	input.Location = strings.TrimSpace(input.Location)
	if input.Title == "" || input.Discipline == "" || input.StartsAt == "" {
		writeError(w, http.StatusBadRequest, "Название, дисциплина и дата обязательны")
		return
	}
	starts, err := time.Parse(time.RFC3339, input.StartsAt)
	if err != nil {
		starts, err = time.Parse("2006-01-02", input.StartsAt)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "Дата должна быть в формате ISO: YYYY-MM-DD или RFC3339")
		return
	}
	if input.CompetitionType == "" {
		input.CompetitionType = "Региональный"
	}
	if input.AgeGroup == "" {
		input.AgeGroup = "Все группы"
	}
	if input.Location == "" {
		input.Location = "Оренбург"
	}
	var c Competition
	var stored time.Time
	err = db.QueryRowContext(r.Context(), `INSERT INTO competitions(title,discipline,competition_type,age_group,starts_at,location,status,participants)
		VALUES($1,$2,$3,$4,$5,$6,'Черновик',0) RETURNING id,title,discipline,competition_type,age_group,starts_at,location,status,participants`,
		input.Title, input.Discipline, input.CompetitionType, input.AgeGroup, starts, input.Location).Scan(&c.ID, &c.Title, &c.Discipline, &c.CompetitionType, &c.AgeGroup, &stored, &c.Location, &c.Status, &c.Participants)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			writeError(w, http.StatusConflict, "Соревнование с таким названием уже существует")
			return
		}
		writeError(w, http.StatusInternalServerError, "Не удалось сохранить соревнование")
		return
	}
	c.StartsAt = stored.Format(time.RFC3339)
	if chatID := strings.TrimSpace(os.Getenv("MAX_DEFAULT_CHAT_ID")); chatID != "" {
		go func() {
			_ = sendMAXMessage("chat_id", chatID, fmt.Sprintf("Создан черновик соревнования: %s · %s", c.Title, c.StartsAt))
		}()
	}
	writeJSON(w, http.StatusCreated, c)
}

func athletesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.QueryContext(r.Context(), `SELECT id,full_name,organization,discipline,age_group,points,coach_name,status FROM athletes ORDER BY full_name`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось получить реестр")
		return
	}
	defer rows.Close()
	items := []Athlete{}
	for rows.Next() {
		var a Athlete
		if err := rows.Scan(&a.ID, &a.FullName, &a.Organization, &a.Discipline, &a.AgeGroup, &a.Points, &a.CoachName, &a.Status); err != nil {
			writeError(w, http.StatusInternalServerError, "Ошибка чтения реестра")
			return
		}
		items = append(items, a)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func listRating(ctx context.Context, discipline string, limit int) ([]RatingEntry, error) {
	query := `SELECT full_name, organization, discipline, age_group, points FROM athletes`
	args := []any{}
	if discipline != "" {
		query += ` WHERE discipline=$1`
		args = append(args, discipline)
	}
	query += ` ORDER BY points DESC, full_name ASC`
	if limit > 0 {
		query += ` LIMIT $` + strconv.Itoa(len(args)+1)
		args = append(args, limit)
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []RatingEntry{}
	for rows.Next() {
		var item RatingEntry
		if err := rows.Scan(&item.Athlete, &item.Organization, &item.Discipline, &item.AgeGroup, &item.Points); err != nil {
			return nil, err
		}
		item.Rank = len(items) + 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func ratingHandler(w http.ResponseWriter, r *http.Request) {
	items, err := listRating(r.Context(), strings.TrimSpace(r.URL.Query().Get("discipline")), 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось рассчитать рейтинг")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "calculation": "Демонстрационная сортировка по сумме очков; регламент расчёта необходимо согласовать с ФСП."})
}

func teamHandler(w http.ResponseWriter, r *http.Request) {
	items, err := listRating(r.Context(), "", 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось получить сборную")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "basis": "Топ-5 демонстрационного рейтинга. Реальный состав должен формироваться по регламенту отбора."})
}

func notificationsHandler(w http.ResponseWriter, r *http.Request) {
	items := []Notification{
		{1, "warning", "Проверить судейские категории", "У 3 судей скоро истекает срок категории.", "Сегодня"},
		{2, "info", "Новые заявки на проверке", "12 заявок ожидают модерации Федерации.", "2 ч назад"},
		{3, "success", "Протокол готов к публикации", "По последнему турниру сформирован черновик протокола.", "Вчера"},
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func maxStatusHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"configured":                strings.TrimSpace(os.Getenv("MAX_BOT_TOKEN")) != "",
		"default_chat_configured":   strings.TrimSpace(os.Getenv("MAX_DEFAULT_CHAT_ID")) != "",
		"webhook_secret_configured": strings.TrimSpace(os.Getenv("MAX_WEBHOOK_SECRET")) != "",
		"webhook_path":              "/api/integrations/max/webhook",
	})
}

type maxUpdate struct {
	UpdateType string `json:"update_type"`
	Message    struct {
		Sender struct {
			UserID int64 `json:"user_id"`
		} `json:"sender"`
		Recipient struct {
			ChatID int64 `json:"chat_id"`
		} `json:"recipient"`
		Body struct {
			Text string `json:"text"`
		} `json:"body"`
	} `json:"message"`
}

func maxWebhookHandler(w http.ResponseWriter, r *http.Request) {
	secret := strings.TrimSpace(os.Getenv("MAX_WEBHOOK_SECRET"))
	if secret != "" && r.Header.Get("X-Max-Bot-Api-Secret") != secret {
		writeError(w, http.StatusUnauthorized, "Неверный секрет webhook")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Не удалось прочитать событие")
		return
	}
	var update maxUpdate
	if err = json.Unmarshal(payload, &update); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный JSON события MAX")
		return
	}
	if update.UpdateType == "message_created" || update.UpdateType == "bot_started" {
		text := strings.TrimSpace(update.Message.Body.Text)
		var reply string
		switch strings.ToLower(text) {
		case "/start", "start":
			reply = "Привет! Я помощник цифровой платформы ФСП Оренбуржья.\nКоманды: /events — ближайшие соревнования, /rating — рейтинг, /help — помощь."
		case "/events", "events":
			reply = formatUpcomingForMAX(r.Context())
		case "/rating", "rating":
			reply = formatRatingForMAX(r.Context())
		case "/help", "help":
			reply = "Команды:\n/events — ближайшие соревнования\n/rating — топ рейтинга\n/help — список команд"
		default:
			if update.UpdateType == "bot_started" {
				reply = "Привет! Напиши /help, чтобы увидеть команды."
			}
		}
		if reply != "" && strings.TrimSpace(os.Getenv("MAX_BOT_TOKEN")) != "" {
			if update.Message.Recipient.ChatID != 0 {
				_ = sendMAXMessage("chat_id", strconv.FormatInt(update.Message.Recipient.ChatID, 10), reply)
			} else if update.Message.Sender.UserID != 0 {
				_ = sendMAXMessage("user_id", strconv.FormatInt(update.Message.Sender.UserID, 10), reply)
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func formatUpcomingForMAX(ctx context.Context) string {
	items, err := listCompetitions(ctx, true)
	if err != nil || len(items) == 0 {
		return "Сейчас нет ближайших соревнований."
	}
	var b strings.Builder
	b.WriteString("Ближайшие соревнования ФСП:\n")
	for _, c := range items {
		date, _ := time.Parse(time.RFC3339, c.StartsAt)
		b.WriteString(fmt.Sprintf("• %s — %s, %s\n", c.Title, date.Format("02.01.2006"), c.Discipline))
	}
	return strings.TrimSpace(b.String())
}

func formatRatingForMAX(ctx context.Context) string {
	items, err := listRating(ctx, "", 5)
	if err != nil || len(items) == 0 {
		return "Рейтинг пока пуст."
	}
	var b strings.Builder
	b.WriteString("Топ рейтинга (демо):\n")
	for _, a := range items {
		b.WriteString(fmt.Sprintf("%d. %s — %d очк.\n", a.Rank, a.Athlete, a.Points))
	}
	return strings.TrimSpace(b.String())
}

func sendMAXMessage(targetType, targetID, message string) error {
	token := strings.TrimSpace(os.Getenv("MAX_BOT_TOKEN"))
	if token == "" {
		return errors.New("MAX_BOT_TOKEN is not configured")
	}
	parsed, err := url.Parse(maxAPI + "/messages")
	if err != nil {
		return err
	}
	q := parsed.Query()
	q.Set(targetType, targetID)
	parsed.RawQuery = q.Encode()
	body, _ := json.Marshal(map[string]any{"text": message, "notify": true})
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, parsed.String(), strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("MAX API returned %s: %s", resp.Status, string(data))
	}
	return nil
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Max-Bot-Api-Secret")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

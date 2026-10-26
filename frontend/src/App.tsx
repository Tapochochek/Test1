import { useCallback, useEffect, useMemo, useState, type FormEvent } from 'react'
import {
  Activity, ArrowDownRight, ArrowRight, ArrowUpRight, Bell, CalendarDays,
  CheckCircle2, ChevronDown, ChevronRight, CircleHelp, ClipboardList, Code2,
  FileBadge, FileText, Filter, Flag, LayoutDashboard, Menu, Plus, Search,
  Settings2, ShieldCheck, Sparkles, Users, UserRound, Trophy, X, Zap,
} from 'lucide-react'

type PageKey = 'overview' | 'competitions' | 'participants' | 'rating' | 'team' | 'documents' | 'notifications'
type Competition = {
  id: number; title: string; discipline: string; competition_type: string; age_group: string;
  starts_at: string; location: string; status: string; participants: number
}
type Athlete = {
  id: number; full_name: string; organization: string; discipline: string; age_group: string;
  points: number; coach_name: string; status: string
}
type RatingEntry = {
  rank: number; athlete: string; organization: string; discipline: string; age_group: string; points: number
}
type Notice = { id: number; kind: string; title: string; message: string; when: string }
type MaxStatus = { configured: boolean; default_chat_configured: boolean; webhook_secret_configured: boolean; webhook_path: string }
type DashboardData = { stats: { athletes: number; competitions: number; applications: number; judges: number }; competitions_total: number; competitions: Competition[]; top_athletes: RatingEntry[] }

type Role = 'Федерация' | 'Тренер' | 'Спортсмен' | 'Судья'

const fallbackCompetitions: Competition[] = [
  { id: 1, title: 'Кубок Оренбуржья по спортивному программированию', discipline: 'Программирование', competition_type: 'Региональный', age_group: 'Юниоры', starts_at: '2026-11-08T10:00:00+05:00', location: 'Оренбург · ОГУ', status: 'Регистрация', participants: 86 },
  { id: 2, title: 'Осенний алгоритмический спринт', discipline: 'Алгоритмическое программирование', competition_type: 'Открытый', age_group: 'Все группы', starts_at: '2026-11-16T11:00:00+05:00', location: 'Онлайн', status: 'Регистрация', participants: 124 },
  { id: 3, title: 'Отбор в сборную области', discipline: 'Программирование систем', competition_type: 'Отборочный', age_group: 'Юниоры и взрослые', starts_at: '2026-11-29T09:30:00+05:00', location: 'Оренбург · Точка кипения', status: 'Планируется', participants: 0 },
]
const fallbackAthletes: Athlete[] = [
  { id: 1, full_name: 'Алексей Морозов', organization: 'Лицей № 1', discipline: 'Алгоритмическое программирование', age_group: 'Юниоры', points: 968, coach_name: 'Мария Крылова', status: 'Активен' },
  { id: 2, full_name: 'Дарья Соколова', organization: 'ОГУ', discipline: 'Программирование', age_group: 'Взрослые', points: 914, coach_name: 'Илья Волков', status: 'Активен' },
  { id: 3, full_name: 'Максим Орлов', organization: 'Гимназия № 2', discipline: 'Робототехника', age_group: 'Юниоры', points: 887, coach_name: 'Мария Крылова', status: 'Активен' },
  { id: 4, full_name: 'Полина Васильева', organization: 'ОГУ', discipline: 'Информационная безопасность', age_group: 'Взрослые', points: 841, coach_name: 'Илья Волков', status: 'Активен' },
  { id: 5, full_name: 'Артём Кузнецов', organization: 'Лицей № 1', discipline: 'Программирование', age_group: 'Юниоры', points: 812, coach_name: 'Мария Крылова', status: 'Активен' },
  { id: 6, full_name: 'София Попова', organization: 'IT-центр Орск', discipline: 'Программная инженерия', age_group: 'Юниоры', points: 779, coach_name: 'Никита Егоров', status: 'Активен' },
  { id: 7, full_name: 'Иван Лебедев', organization: 'ОГУ', discipline: 'Алгоритмическое программирование', age_group: 'Взрослые', points: 745, coach_name: 'Илья Волков', status: 'Активен' },
  { id: 8, full_name: 'Ева Фёдорова', organization: 'Гимназия № 2', discipline: 'Робототехника', age_group: 'Юниоры', points: 702, coach_name: 'Мария Крылова', status: 'Активен' },
]
const fallbackRating: RatingEntry[] = fallbackAthletes.map((a, i) => ({ rank: i + 1, athlete: a.full_name, organization: a.organization, discipline: a.discipline, age_group: a.age_group, points: a.points }))
const fallbackNotices: Notice[] = [
  { id: 1, kind: 'warning', title: 'Проверить судейские категории', message: 'У 3 судей скоро истекает срок категории.', when: 'Сегодня' },
  { id: 2, kind: 'info', title: 'Новые заявки на проверке', message: '12 заявок ожидают модерации Федерации.', when: '2 ч назад' },
  { id: 3, kind: 'success', title: 'Протокол готов к публикации', message: 'По последнему турниру сформирован черновик протокола.', when: 'Вчера' },
]

const navItems: { key: PageKey; label: string; icon: typeof LayoutDashboard; tag?: string }[] = [
  { key: 'overview', label: 'Главная', icon: LayoutDashboard },
  { key: 'competitions', label: 'Соревнования', icon: CalendarDays, tag: '12' },
  { key: 'participants', label: 'Реестр участников', icon: Users },
  { key: 'rating', label: 'Рейтинг', icon: Trophy },
  { key: 'team', label: 'Сборная области', icon: Flag },
  { key: 'documents', label: 'Документы', icon: FileText },
  { key: 'notifications', label: 'Уведомления', icon: Bell, tag: '3' },
]
const roleDescriptions: Record<Role, string> = {
  Федерация: 'Региональное управление', Тренер: 'Кабинет тренера', Спортсмен: 'Кабинет спортсмена', Судья: 'Судейская коллегия',
}
const titles: Record<PageKey, string> = {
  overview: 'Обзор системы', competitions: 'Соревнования', participants: 'Реестр участников', rating: 'Рейтинг спортсменов', team: 'Сборная области', documents: 'Документы и отчёты', notifications: 'Центр уведомлений',
}

function formatFullDate(raw: string) {
  const d = new Date(raw)
  if (Number.isNaN(d.getTime())) return raw
  return new Intl.DateTimeFormat('ru-RU', { day: '2-digit', month: 'long', year: 'numeric' }).format(d)
}
async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, { ...init, headers: { 'Content-Type': 'application/json', ...(init?.headers || {}) } })
  if (!response.ok) {
    const payload = await response.json().catch(() => ({}))
    throw new Error(payload.error || `Ошибка запроса: ${response.status}`)
  }
  return response.json() as Promise<T>
}

export default function App() {
  const [page, setPage] = useState<PageKey>('overview')
  const [role, setRole] = useState<Role>('Федерация')
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [dashboard, setDashboard] = useState<DashboardData>({
    stats: { athletes: 248, competitions: 6, applications: 312, judges: 24 }, competitions_total: 12, competitions: fallbackCompetitions, top_athletes: fallbackRating.slice(0, 5),
  })
  const [competitions, setCompetitions] = useState<Competition[]>(fallbackCompetitions)
  const [athletes, setAthletes] = useState<Athlete[]>(fallbackAthletes)
  const [rating, setRating] = useState<RatingEntry[]>(fallbackRating)
  const [team, setTeam] = useState<RatingEntry[]>(fallbackRating.slice(0, 5))
  const [notices, setNotices] = useState<Notice[]>(fallbackNotices)
  const [maxStatus, setMaxStatus] = useState<MaxStatus>({ configured: false, default_chat_configured: false, webhook_secret_configured: false, webhook_path: '/api/integrations/max/webhook' })
  const [notice, setNotice] = useState('')
  const [search, setSearch] = useState('')
  const [discipline, setDiscipline] = useState('Все дисциплины')
  const [modalOpen, setModalOpen] = useState(false)
  const [loading, setLoading] = useState(false)

  const loadData = useCallback(async () => {
    try {
      setLoading(true)
      const [d, c, a, r, t, n, nMax] = await Promise.all([
        api<DashboardData>('/api/dashboard'), api<{ items: Competition[] }>('/api/competitions'),
        api<{ items: Athlete[] }>('/api/athletes'), api<{ items: RatingEntry[] }>('/api/rating'),
        api<{ items: RatingEntry[] }>('/api/team'), api<{ items: Notice[] }>('/api/notifications'),
        api<MaxStatus>('/api/integrations/max/status'),
      ])
      setDashboard(d); setCompetitions(c.items); setAthletes(a.items); setRating(r.items); setTeam(t.items); setNotices(n.items); setMaxStatus(nMax)
    } catch {
      setNotice('API пока недоступно — показаны демонстрационные данные. Запустите проект через Docker Compose.')
    } finally { setLoading(false) }
  }, [])

  useEffect(() => { void loadData() }, [loadData])
  useEffect(() => { if (!notice) return; const timer = window.setTimeout(() => setNotice(''), 4800); return () => window.clearTimeout(timer) }, [notice])

  const visibleCompetitions = useMemo(() => competitions.filter(c => `${c.title} ${c.discipline} ${c.location} ${c.status}`.toLowerCase().includes(search.toLowerCase())), [competitions, search])
  const visibleAthletes = useMemo(() => athletes.filter(a => `${a.full_name} ${a.organization} ${a.discipline} ${a.coach_name}`.toLowerCase().includes(search.toLowerCase())), [athletes, search])
  const visibleRating = useMemo(() => rating.filter(a => (discipline === 'Все дисциплины' || a.discipline === discipline) && `${a.athlete} ${a.organization} ${a.discipline}`.toLowerCase().includes(search.toLowerCase())), [rating, discipline, search])

  function openPage(next: PageKey) { setPage(next); setSearch(''); setSidebarOpen(false) }

  return <div className="app-shell">
    <aside className={`sidebar ${sidebarOpen ? 'sidebar-open' : ''}`}>
      <div className="brand-lockup">
        <div className="brand-mark" aria-hidden="true"><span className="brand-mark-letter">Ф</span><i /></div>
        <div className="brand-text"><b>ФСП<span> 56</span></b><small>ОРЕНБУРГСКАЯ ОБЛАСТЬ</small></div>
        <button className="icon-button sidebar-close" onClick={() => setSidebarOpen(false)} aria-label="Закрыть меню"><X size={18} /></button>
      </div>
      <div className="season-chip"><span className="pulse-dot" /> СЕЗОН 2026 / 27</div>
      <p className="nav-caption">РАБОЧЕЕ ПРОСТРАНСТВО</p>
      <nav className="main-nav">
        {navItems.map(item => <button key={item.key} onClick={() => openPage(item.key)} className={`nav-item ${page === item.key ? 'active' : ''}`}>
          <item.icon size={18} strokeWidth={1.9} /><span>{item.label}</span>{item.tag && <em>{item.tag}</em>}
        </button>)}
      </nav>
      <div className="sidebar-bottom">
        <div className="sidebar-help"><div className="help-icon"><CircleHelp size={18} /></div><div><b>Нужна помощь?</b><small>База знаний платформы</small></div><ChevronRight size={15} /></div>
        <button className="nav-item settings-item" onClick={() => setNotice('Настройки появятся в следующем срезе прототипа.')}><Settings2 size={18} /><span>Настройки системы</span></button>
        <div className="profile-card"><div className="avatar avatar-violet">АП</div><div className="profile-copy"><b>Администратор</b><small>ФСП · Оренбург</small></div><span className="online-dot" /></div>
      </div>
    </aside>

    {sidebarOpen && <button className="mobile-overlay" aria-label="Закрыть меню" onClick={() => setSidebarOpen(false)} />}

    <main className="main-area">
      <header className="topbar">
        <div className="topbar-left"><button className="icon-button mobile-menu" onClick={() => setSidebarOpen(true)} aria-label="Открыть меню"><Menu size={20} /></button><div className="breadcrumb">Платформа <ChevronRight size={14} /><b>{titles[page]}</b></div></div>
        <div className="topbar-right">
          <div className="environment"><span /> DEMO-СРЕДА</div>
          <button className="top-icon" aria-label="Поиск" onClick={() => document.getElementById('global-search')?.focus()}><Search size={18} /></button>
          <button className="top-icon notification-top" aria-label="Уведомления" onClick={() => openPage('notifications')}><Bell size={18} /><i /></button>
          <div className="role-select-wrap"><ShieldCheck size={16} /><select aria-label="Демонстрационная роль" value={role} onChange={e => setRole(e.target.value as Role)}>{(['Федерация', 'Тренер', 'Спортсмен', 'Судья'] as Role[]).map(r => <option key={r}>{r}</option>)}</select><ChevronDown size={14} /></div>
        </div>
      </header>

      <div className="page-content">
        <section className="welcome-row">
          <div><div className="eyebrow"><span className="eyebrow-line" /> ЕДИНАЯ ЦИФРОВАЯ СРЕДА ФСП</div><h1>{page === 'overview' ? 'Управление спортивным программированием' : titles[page]}</h1><p>{page === 'overview' ? 'Все процессы Федерации — в одном пространстве.' : `${roleDescriptions[role]} · Оренбургская область`}</p></div>
          <div className="welcome-actions"><div className="last-updated"><span className={`refresh-dot ${loading ? 'loading' : ''}`} />{loading ? 'Обновление…' : 'Данные актуальны'}</div>{(page === 'overview' || page === 'competitions') && <button className="button button-primary" onClick={() => setModalOpen(true)}><Plus size={17} /> Создать соревнование</button>}</div>
        </section>

        {page === 'overview' && <Dashboard dashboard={dashboard} competitions={competitions} rating={rating} notices={notices} maxStatus={maxStatus} onNavigate={openPage} />}
        {page === 'competitions' && <CompetitionsView items={visibleCompetitions} search={search} setSearch={setSearch} onCreate={() => setModalOpen(true)} />}
        {page === 'participants' && <AthletesView items={visibleAthletes} search={search} setSearch={setSearch} />}
        {page === 'rating' && <RatingView items={visibleRating} discipline={discipline} setDiscipline={setDiscipline} search={search} setSearch={setSearch} />}
        {page === 'team' && <TeamView items={team} />}
        {page === 'documents' && <DocumentsView onMessage={setNotice} />}
        {page === 'notifications' && <NotificationsView items={notices} />}

        <footer className="footer"><div><span className="footer-mark"><Code2 size={14} /></span> ФСП · ОРЕНБУРГСКАЯ ОБЛАСТЬ</div><span>Прототип для ОренКод 2026 <b>·</b> v0.1</span></footer>
      </div>
    </main>

    {modalOpen && <CreateCompetitionModal onClose={() => setModalOpen(false)} onCreated={async () => { setModalOpen(false); setNotice('Соревнование сохранено в базе данных.'); await loadData() }} onError={setNotice} />}
    {notice && <div className="toast"><CheckCircle2 size={18} /><span>{notice}</span><button onClick={() => setNotice('')} aria-label="Закрыть уведомление"><X size={15} /></button></div>}
  </div>
}

function Dashboard({ dashboard, competitions, rating, notices, maxStatus, onNavigate }: { dashboard: DashboardData; competitions: Competition[]; rating: RatingEntry[]; notices: Notice[]; maxStatus: MaxStatus; onNavigate: (page: PageKey) => void }) {
  const stats = [
    { label: 'Спортсмены в реестре', value: dashboard.stats.athletes.toLocaleString('ru-RU'), change: '+8,4%', icon: Users, cls: 'stat-indigo', foot: 'к прошлому сезону', up: true },
    { label: 'Активные соревнования', value: String(dashboard.stats.competitions).padStart(2, '0'), change: '3 скоро', icon: CalendarDays, cls: 'stat-mint', foot: 'в календаре', up: true },
    { label: 'Заявки участников', value: dashboard.stats.applications.toLocaleString('ru-RU'), change: '12 новых', icon: ClipboardList, cls: 'stat-orange', foot: 'требуют модерации', up: false },
    { label: 'Судьи в системе', value: String(dashboard.stats.judges).padStart(2, '0'), change: '3 срока', icon: ShieldCheck, cls: 'stat-blue', foot: 'скоро истекают', up: false },
  ]
  return <>
    <section className="hero-banner">
      <div className="hero-grid" />
      <div className="hero-copy"><div className="hero-label"><Sparkles size={13} /> ПАНЕЛЬ ФЕДЕРАЦИИ</div><h2>Спортивное программирование.<br /><span>Единая точка управления.</span></h2><p>Соревнования, спортсмены, рейтинг и документы — синхронизированы в одном цифровом контуре.</p><button onClick={() => onNavigate('competitions')} className="hero-cta">К календарю соревнований <ArrowRight size={16} /></button></div>
      <div className="hero-visual" aria-hidden="true"><div className="orbit orbit-one" /><div className="orbit orbit-two" /><div className="orbit orbit-three" /><div className="hero-core"><Code2 size={37} /></div><div className="float-tag tag-left"><span className="mini-check"><CheckCircle2 size={12} /></span> Данные синхронизированы</div><div className="float-tag tag-right"><Zap size={13} /> НА СВЯЗИ</div><div className="hero-scanline" /></div>
      <div className="hero-index">ФСП <b>56</b><span>ОРЕНБУРГСКАЯ ОБЛАСТЬ · 2026</span></div>
    </section>

    <div className="section-heading"><div><h2>Показатели системы</h2><p>Текущая картина по региону</p></div><button className="text-button" onClick={() => onNavigate('participants')}>Открыть реестр <ArrowRight size={15} /></button></div>
    <section className="stats-grid">
      {stats.map((s, i) => <div className="stat-card" key={s.label}><div className={`stat-icon ${s.cls}`}><s.icon size={19} /></div><div className="stat-meta"><span>{s.label}</span><b>{s.value}</b><div className={`stat-foot ${s.up ? 'positive' : 'attention'}`}>{s.up ? <ArrowUpRight size={14} /> : <Activity size={13} />}<strong>{s.change}</strong><span>{s.foot}</span></div></div><div className={`stat-watermark wm-${i}`}><s.icon size={52} /></div></div>)}
    </section>

    <section className="content-grid main-dashboard-grid">
      <div className="panel competition-panel"><div className="panel-heading"><div><div className="panel-kicker">КАЛЕНДАРЬ</div><h3>Ближайшие соревнования</h3></div><button className="panel-action" onClick={() => onNavigate('competitions')}>Все события <ArrowRight size={14} /></button></div>
        <div className="event-list">{competitions.slice(0, 3).map(c => <div className="event-row" key={c.id}><div className="event-date"><b>{new Date(c.starts_at).toLocaleDateString('ru-RU', { day: '2-digit' })}</b><span>{new Date(c.starts_at).toLocaleDateString('ru-RU', { month: 'short' }).replace('.', '')}</span></div><div className="event-copy"><b>{c.title}</b><span><Code2 size={12} /> {c.discipline} <i /> {c.location}</span></div><div className={`status-pill ${c.status === 'Регистрация' ? 'status-open' : c.status === 'Завершено' ? 'status-done' : 'status-planned'}`}>{c.status}</div><button className="row-arrow" aria-label={`Открыть ${c.title}`} onClick={() => onNavigate('competitions')}><ChevronRight size={17} /></button></div>)}</div>
        <button className="panel-bottom-link" onClick={() => onNavigate('competitions')}><CalendarDays size={15} /> Перейти в календарь <ArrowRight size={14} /></button>
      </div>
      <div className="panel rating-panel"><div className="panel-heading"><div><div className="panel-kicker">ТОП РЕЙТИНГА</div><h3>Топ рейтинга региона</h3></div><button className="round-action" onClick={() => onNavigate('rating')} aria-label="Открыть рейтинг"><ArrowUpRight size={17} /></button></div><div className="rating-list">{rating.slice(0, 5).map((a, i) => <div className="rating-row" key={a.athlete}><div className={`rank-num ${i === 0 ? 'rank-gold' : i === 1 ? 'rank-silver' : i === 2 ? 'rank-bronze' : ''}`}>{String(i + 1).padStart(2, '0')}</div><div className={`athlete-avatar avatar-${i % 4}`}>{a.athlete.split(' ').map(x => x[0]).slice(0, 2).join('')}</div><div className="rating-athlete"><b>{a.athlete}</b><span>{a.organization}</span></div><div className="rating-points"><b>{a.points}</b><span>очков</span></div></div>)}</div><button className="panel-bottom-link" onClick={() => onNavigate('rating')}><Trophy size={15} /> Полный рейтинг <ArrowRight size={14} /></button></div>
    </section>

    <section className="content-grid lower-grid"><div className="panel workflow-panel"><div className="panel-heading"><div><div className="panel-kicker">ТРЕБУЮТ ВНИМАНИЯ</div><h3>Рабочие задачи</h3></div><span className="count-badge">03</span></div><div className="workflow-item"><div className="workflow-icon workflow-yellow"><ClipboardList size={17} /></div><div><b>Модерация заявок</b><span>12 заявок ждут решения Федерации</span></div><button onClick={() => onNavigate('competitions')} aria-label="Проверить заявки"><ArrowRight size={15} /></button></div><div className="workflow-item"><div className="workflow-icon workflow-red"><ShieldCheck size={17} /></div><div><b>Срок действия категорий</b><span>3 судейские категории требуют проверки</span></div><button onClick={() => onNavigate('participants')} aria-label="Проверить категории"><ArrowRight size={15} /></button></div><div className="workflow-item"><div className="workflow-icon workflow-green"><FileBadge size={17} /></div><div><b>Протокол к публикации</b><span>Один документ готов к согласованию</span></div><button onClick={() => onNavigate('documents')} aria-label="Открыть документы"><ArrowRight size={15} /></button></div></div>
      <div className="panel notices-panel"><div className="panel-heading"><div><div className="panel-kicker">СИСТЕМА</div><h3>Последние уведомления</h3></div><button className="panel-action" onClick={() => onNavigate('notifications')}>Все <ArrowRight size={14} /></button></div><div className="notice-list">{notices.slice(0, 3).map(n => <div className="notice-row" key={n.id}><div className={`notice-mark notice-${n.kind}`}>{n.kind === 'warning' ? <Activity size={15} /> : n.kind === 'success' ? <CheckCircle2 size={15} /> : <Bell size={15} />}</div><div><b>{n.title}</b><p>{n.message}</p><small>{n.when}</small></div></div>)}</div></div></section>
    <section className="max-integration-card"><div className="max-brand-block"><span className="max-logo">M</span><div><span className="panel-kicker">ВНЕШНИЕ ИНТЕГРАЦИИ</span><h3>Оповещения в MAX</h3><p>Напоминания о сроках, соревнованиях и публикации результатов — прямо в мессенджере.</p></div></div><div className="max-status-block"><span className={`max-connection ${maxStatus.configured && maxStatus.webhook_secret_configured ? 'max-connected' : ''}`}><span />{maxStatus.configured && maxStatus.webhook_secret_configured ? 'Переменные заданы' : maxStatus.configured ? 'Нужен секрет webhook' : 'Ожидает подключения'}</span><div className="max-meta"><span><CheckCircle2 size={13} /> Go webhook готов</span><span><ShieldCheck size={13} /> Проверка секрета</span></div><button className="button button-light" onClick={() => onNavigate('documents')}>Как подключить <ArrowRight size={14} /></button></div></section>
  </>
}

function PageToolbar({ search, setSearch, placeholder = 'Поиск по реестру...' }: { search: string; setSearch: (value: string) => void; placeholder?: string }) {
  return <div className="table-toolbar"><label className="search-box"><Search size={17} /><input id="global-search" value={search} onChange={e => setSearch(e.target.value)} placeholder={placeholder} /></label><button className="button button-light" onClick={() => setSearch('')}><Filter size={15} /> Сбросить</button></div>
}

function CompetitionsView({ items, search, setSearch, onCreate }: { items: Competition[]; search: string; setSearch: (value: string) => void; onCreate: () => void }) {
  return <div className="panel table-panel"><div className="table-header-copy"><div><h3>Календарь соревнований</h3><p>Создание и контроль региональных событий</p></div><span className="total-chip">{items.length} событий</span></div><PageToolbar search={search} setSearch={setSearch} placeholder="Название, дисциплина или место..." /><div className="table-scroll"><table><thead><tr><th>Соревнование</th><th>Дата</th><th>Дисциплина</th><th>Участники</th><th>Статус</th><th /></tr></thead><tbody>{items.map(c => <tr key={c.id}><td><div className="table-primary"><span className="table-icon"><CalendarDays size={16} /></span><div><b>{c.title}</b><small>{c.competition_type} · {c.age_group}</small></div></div></td><td><span className="date-cell">{formatFullDate(c.starts_at)}</span><small className="cell-sub">{c.location}</small></td><td>{c.discipline}</td><td><span className="participants-number">{c.participants || '—'}</span></td><td><span className={`status-pill ${c.status === 'Регистрация' ? 'status-open' : c.status === 'Завершено' ? 'status-done' : c.status === 'Черновик' ? 'status-draft' : 'status-planned'}`}>{c.status}</span></td><td><button className="row-arrow" aria-label="Открыть событие"><ChevronRight size={17} /></button></td></tr>)}</tbody></table>{items.length === 0 && <EmptyState label="Ничего не найдено" detail="Попробуйте изменить запрос." />}</div><div className="table-footer"><span>Показано {items.length} событий</span><button className="button button-primary" onClick={onCreate}><Plus size={16} /> Создать соревнование</button></div></div>
}

function AthletesView({ items, search, setSearch }: { items: Athlete[]; search: string; setSearch: (value: string) => void }) {
  return <div className="panel table-panel"><div className="table-header-copy"><div><h3>Реестр участников</h3><p>Спортсмены и их организации</p></div><span className="total-chip">{items.length} записей</span></div><PageToolbar search={search} setSearch={setSearch} placeholder="ФИО, организация, тренер..." /><div className="table-scroll"><table><thead><tr><th>Спортсмен</th><th>Организация</th><th>Дисциплина</th><th>Возрастная группа</th><th>Очки</th><th>Статус</th></tr></thead><tbody>{items.map((a, i) => <tr key={a.id}><td><div className="table-primary"><span className={`athlete-avatar avatar-${i % 4}`}>{a.full_name.split(' ').map(x => x[0]).slice(0, 2).join('')}</span><div><b>{a.full_name}</b><small>Тренер: {a.coach_name}</small></div></div></td><td>{a.organization}</td><td>{a.discipline}</td><td>{a.age_group}</td><td><b>{a.points}</b></td><td><span className={`status-pill ${a.status === 'Активен' ? 'status-open' : 'status-planned'}`}>{a.status}</span></td></tr>)}</tbody></table>{items.length === 0 && <EmptyState label="Участники не найдены" detail="Попробуйте другой поисковый запрос." />}</div><div className="table-footer"><span>Данные демонстрационные · персональные данные вымышлены</span><span className="table-foot-note"><ShieldCheck size={14} /> Реестр ФСП</span></div></div>
}

function RatingView({ items, discipline, setDiscipline, search, setSearch }: { items: RatingEntry[]; discipline: string; setDiscipline: (value: string) => void; search: string; setSearch: (value: string) => void }) {
  const disciplines = ['Все дисциплины', ...Array.from(new Set(fallbackAthletes.map(a => a.discipline))) ]
  return <><section className="rating-overview-grid"><div className="rating-highlight"><div className="highlight-top"><span>РЕГИОНАЛЬНЫЙ РЕЙТИНГ</span><Trophy size={20} /></div><b>{items.length.toLocaleString('ru-RU')}</b><p>спортсменов в выбранной выборке</p><div className="highlight-rule"><span /><span /><span /><span /><span /><span /><span /><span /><span /></div><small>Сортировка по сумме набранных очков</small></div><div className="rating-note"><div className="rating-note-icon"><Activity size={19} /></div><div><b>Прозрачный расчёт</b><p>В этой версии таблица сортируется по демонстрационной сумме очков. Для защиты нужно зафиксировать официальную формулу начисления с организаторами.</p></div></div></section><div className="panel table-panel rating-table-panel"><div className="table-header-copy"><div><h3>Таблица рейтинга</h3><p>Позиции обновляются на основе данных участников</p></div><label className="select-box"><span>Дисциплина</span><select value={discipline} onChange={e => setDiscipline(e.target.value)}>{disciplines.map(d => <option key={d}>{d}</option>)}</select><ChevronDown size={14} /></label></div><PageToolbar search={search} setSearch={setSearch} placeholder="Поиск спортсмена..." /><div className="table-scroll"><table><thead><tr><th>Место</th><th>Спортсмен</th><th>Дисциплина</th><th>Группа</th><th>Очки</th><th /></tr></thead><tbody>{items.map(a => <tr key={a.athlete}><td><span className={`rank-num table-rank ${a.rank <= 3 ? `rank-${a.rank === 1 ? 'gold' : a.rank === 2 ? 'silver' : 'bronze'}` : ''}`}>{String(a.rank).padStart(2, '0')}</span></td><td><div className="table-primary"><div className="rating-athlete"><b>{a.athlete}</b><small>{a.organization}</small></div></div></td><td>{a.discipline}</td><td>{a.age_group}</td><td><b>{a.points}</b></td><td><button className="row-arrow" aria-label="Карточка спортсмена"><ChevronRight size={17} /></button></td></tr>)}</tbody></table>{items.length === 0 && <EmptyState label="Нет результатов" detail="В этой выборке пока нет спортсменов." />}</div></div></>
}

function TeamView({ items }: { items: RatingEntry[] }) {
  return <><div className="team-banner"><div className="team-banner-mark"><Flag size={28} /></div><div><div className="panel-kicker">СОСТАВ СБОРНОЙ · 2026</div><h2>Сборная Оренбургской области</h2><p>Предварительный состав по результатам демонстрационного рейтинга.</p></div><span className="team-count">{items.length} <small>участников</small></span></div><div className="team-grid">{items.map((a, i) => <div className="team-card" key={a.athlete}><div className="team-card-top"><div className={`athlete-avatar avatar-${i % 4}`}>{a.athlete.split(' ').map(x => x[0]).slice(0, 2).join('')}</div><span className={`rank-num ${i === 0 ? 'rank-gold' : i === 1 ? 'rank-silver' : i === 2 ? 'rank-bronze' : ''}`}>#{i + 1}</span></div><h3>{a.athlete}</h3><p>{a.organization}</p><div className="team-card-discipline"><Code2 size={13} /> {a.discipline}</div><div className="team-score"><span>Рейтинг</span><b>{a.points} <small>очков</small></b></div><div className="team-progress"><span style={{ width: `${Math.max(14, Math.round(a.points / (items[0]?.points || 968) * 100))}%` }} /></div></div>)}</div><div className="prototype-note"><CircleHelp size={17} /><span>Состав здесь определяется по топ-5 рейтинга. В реальной версии критерии отбора, ограничения по дисциплинам и возрасту нужно согласовать с Федерацией.</span></div></>
}

function DocumentsView({ onMessage }: { onMessage: (message: string) => void }) {
  const docs = [
    { title: 'Протокол соревнования', kind: 'PDF / DOCX', desc: 'Итоговый протокол с результатами и местами участников', icon: FileBadge, color: 'doc-purple', status: 'Шаблон' },
    { title: 'Отчёт по спортсмену', kind: 'PDF', desc: 'Индивидуальная карточка, рейтинг и история результатов', icon: UserRound, color: 'doc-blue', status: 'Шаблон' },
    { title: 'Реестр участников', kind: 'XLSX', desc: 'Список спортсменов, организации и спортивные показатели', icon: Users, color: 'doc-green', status: 'Демо-данные' },
    { title: 'Отчёт по динамике', kind: 'XLSX', desc: 'Количество участников и организаций в разрезе периода', icon: Activity, color: 'doc-orange', status: 'В разработке' },
  ]
  return <><div className="doc-intro"><div className="doc-intro-icon"><FileText size={24} /></div><div><h3>Библиотека документов</h3><p>Шаблоны и отчётность Федерации в едином пространстве</p></div></div><div className="document-grid">{docs.map(d => <div className="document-card" key={d.title}><div className="document-card-head"><div className={`document-icon ${d.color}`}><d.icon size={20} /></div><span className="file-type">{d.kind}</span></div><h3>{d.title}</h3><p>{d.desc}</p><div className="document-card-foot"><span className={`doc-status ${d.status === 'В разработке' ? 'doc-pending' : ''}`}>{d.status}</span><button onClick={() => onMessage(`«${d.title}»: экспорт будет добавлен в следующем срезе прототипа.`)} aria-label={`Открыть ${d.title}`}><ArrowDownRight size={17} /></button></div></div>)}</div><div className="prototype-note"><CircleHelp size={17} /><span>Экспорт пока не подключён. План следующего этапа: XLSX для реестра и динамики, PDF/DOCX для протокола и индивидуального отчёта.</span></div></>
}

function NotificationsView({ items }: { items: Notice[] }) {
  return <div className="panel notifications-full"><div className="table-header-copy"><div><h3>Входящие уведомления</h3><p>Системные задачи и напоминания для Федерации</p></div><span className="total-chip">{items.length} сообщения</span></div><div className="notifications-full-list">{items.map(n => <div className="notifications-full-row" key={n.id}><div className={`notice-mark notice-${n.kind}`}>{n.kind === 'warning' ? <Activity size={17} /> : n.kind === 'success' ? <CheckCircle2 size={17} /> : <Bell size={17} />}</div><div className="notifications-full-copy"><div><b>{n.title}</b><span>{n.when}</span></div><p>{n.message}</p></div><span className={`notice-label label-${n.kind}`}>{n.kind === 'warning' ? 'Внимание' : n.kind === 'success' ? 'Готово' : 'Информация'}</span></div>)}</div></div>
}

function CreateCompetitionModal({ onClose, onCreated, onError }: { onClose: () => void; onCreated: () => Promise<void>; onError: (message: string) => void }) {
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState({ title: '', discipline: 'Программирование', competition_type: 'Региональный', age_group: 'Все группы', starts_at: '2026-11-21', location: 'Оренбург' })
  async function submit(e: FormEvent) {
    e.preventDefault(); setSaving(true)
    try { await api<Competition>('/api/competitions', { method: 'POST', body: JSON.stringify({ ...form, starts_at: `${form.starts_at}T10:00:00+05:00` }) }); await onCreated() }
    catch (error) { onError(error instanceof Error ? error.message : 'Не удалось создать соревнование') }
    finally { setSaving(false) }
  }
  function update(key: keyof typeof form, value: string) { setForm(current => ({ ...current, [key]: value })) }
  return <div className="modal-backdrop" onMouseDown={e => { if (e.target === e.currentTarget) onClose() }}><section className="modal-card"><div className="modal-header"><div><div className="panel-kicker">НОВОЕ СОБЫТИЕ</div><h2>Создать соревнование</h2><p>Новый черновик будет сохранён в календаре.</p></div><button className="icon-button" onClick={onClose} aria-label="Закрыть окно"><X size={20} /></button></div><form onSubmit={submit}><label className="field-label">Название соревнования<input autoFocus required value={form.title} onChange={e => update('title', e.target.value)} placeholder="Например, Кубок Оренбуржья" /></label><div className="form-row"><label className="field-label">Дисциплина<select value={form.discipline} onChange={e => update('discipline', e.target.value)}>{['Программирование', 'Алгоритмическое программирование', 'Программирование систем', 'Информационная безопасность', 'Робототехника', 'Программная инженерия'].map(x => <option key={x}>{x}</option>)}</select></label><label className="field-label">Тип соревнования<select value={form.competition_type} onChange={e => update('competition_type', e.target.value)}>{['Региональный', 'Открытый', 'Отборочный', 'Тренировочный', 'Образовательный'].map(x => <option key={x}>{x}</option>)}</select></label></div><div className="form-row"><label className="field-label">Возрастная группа<select value={form.age_group} onChange={e => update('age_group', e.target.value)}>{['Все группы', 'Юниоры', 'Взрослые', 'Юниоры и взрослые'].map(x => <option key={x}>{x}</option>)}</select></label><label className="field-label">Дата проведения<input required type="date" value={form.starts_at} onChange={e => update('starts_at', e.target.value)} /></label></div><label className="field-label">Место проведения<input required value={form.location} onChange={e => update('location', e.target.value)} placeholder="Оренбург / Онлайн" /></label><div className="modal-note"><ShieldCheck size={16} /><span>Создаётся черновик. Публикация и приём заявок будут отдельными действиями.</span></div><div className="modal-actions"><button className="button button-light" type="button" onClick={onClose}>Отмена</button><button className="button button-primary" type="submit" disabled={saving}>{saving ? 'Сохраняем…' : <><Plus size={16} /> Создать черновик</>}</button></div></form></section></div>
}

function EmptyState({ label, detail }: { label: string; detail: string }) { return <div className="empty-state"><div><Search size={20} /></div><b>{label}</b><span>{detail}</span></div> }

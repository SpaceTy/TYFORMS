package database

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
	"tyforms/internal/models"
)

const (
	performanceApps          = 5000
	performanceWordsPerField = 150
	performanceIterations    = 15
	performanceMaxAvg        = 500 * time.Millisecond
)

var performanceVocabulary = []string{
	"adventure", "alliance", "ancient", "apple", "autumn", "bamboo", "beacon",
	"border", "bridge", "castle", "cavern", "cherry", "cobble", "copper",
	"creative", "crystal", "desert", "diamond", "dragon", "emerald", "forest",
	"fortress", "frontier", "garden", "glacier", "granite", "harbor", "harvest",
	"island", "jungle", "lantern", "legend", "mansion", "meadow", "mining",
	"mountain", "mushroom", "nether", "ocean", "orchard", "outpost", "pathway",
	"pillage", "plains", "portal", "prairie", "quartz", "railway", "redstone",
	"river", "savanna", "settle", "shadow", "silver", "spawner", "spruce",
	"stronghold", "summit", "sunset", "tavern", "temple", "tundra", "village",
	"vineyard", "volcano", "waterfall", "willow", "winter", "workshop",
}

// generateWords builds a deterministic sentence of n vocabulary words.
func generateWords(rng *rand.Rand, n int) string {
	var b strings.Builder
	b.Grow(n * 10)
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(performanceVocabulary[rng.Intn(len(performanceVocabulary))])
	}
	return b.String()
}

// seedPerformanceDatabase creates the generated application dataset and
// returns the store plus the temporary database path.
func seedPerformanceDatabase(tb testing.TB) *SQLiteStore {
	tb.Helper()

	store, err := NewSQLiteStore(filepath.Join(tb.TempDir(), "performance.db"))
	if err != nil {
		tb.Fatalf("create store: %v", err)
	}

	// The dataset is regenerated on every run, so durability is not needed;
	// disabling fsync keeps the 5000 inserts reasonably fast.
	if _, err := store.db.Exec(`PRAGMA journal_mode = MEMORY`); err != nil {
		tb.Fatalf("set journal mode: %v", err)
	}
	if _, err := store.db.Exec(`PRAGMA synchronous = OFF`); err != nil {
		tb.Fatalf("set synchronous mode: %v", err)
	}

	rng := rand.New(rand.NewSource(42))
	start := time.Now()
	for i := 0; i < performanceApps; i++ {
		favorite := generateWords(rng, performanceWordsPerField)
		understanding := generateWords(rng, performanceWordsPerField)
		if i%100 == 0 {
			favorite += " needleword"
		}

		app := &models.MinecraftApplication{
			DiscordUsername:        fmt.Sprintf("discord_user_%d", i),
			MinecraftUsername:      fmt.Sprintf("Player%d", i),
			Age:                    18 + i%30,
			FavoriteAboutMinecraft: favorite,
			UnderstandingOfSMP:     understanding,
			SubmissionDate:         time.Now().UTC(),
			AcceptanceStatus:       "pending",
		}
		if err := store.CreateApplication(app); err != nil {
			store.Close()
			tb.Fatalf("seed application %d: %v", i, err)
		}
	}
	tb.Logf("seeded %d applications (~%d words each) in %s", performanceApps, performanceWordsPerField*2, time.Since(start).Round(time.Millisecond))

	return store
}

type searchTiming struct {
	total int
	rows  int
	min   time.Duration
	avg   time.Duration
	p95   time.Duration
	max   time.Duration
}

func measureSearch(tb testing.TB, store *SQLiteStore, query string, fields []string, page, pageSize int) searchTiming {
	tb.Helper()

	durations := make([]time.Duration, 0, performanceIterations)
	var total int
	var rows int
	for i := 0; i < performanceIterations; i++ {
		start := time.Now()
		apps, count, err := store.SearchApplications(query, fields, page, pageSize)
		durations = append(durations, time.Since(start))
		if err != nil {
			tb.Fatalf("search %q: %v", query, err)
		}
		total = count
		rows = len(apps)
	}

	sorted := append([]time.Duration(nil), durations...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	p95 := sorted[(len(sorted)*95)/100-1]
	if p95 == 0 {
		p95 = sorted[len(sorted)-1]
	}

	return searchTiming{
		total: total,
		rows:  rows,
		min:   sorted[0],
		avg:   sum / time.Duration(len(durations)),
		p95:   p95,
		max:   sorted[len(sorted)-1],
	}
}

func TestSearchPerformanceLargeDatabase(t *testing.T) {
	store := seedPerformanceDatabase(t)
	defer store.Close()

	cases := []struct {
		name     string
		query    string
		fields   []string
		wantSome bool
	}{
		{name: "no query, first page", query: "", wantSome: true},
		{name: "needle word (50 results)", query: "needleword", wantSome: true},
		{name: "needle substring", query: "needle", wantSome: true},
		{name: "common word", query: "redstone", wantSome: true},
		{name: "short 2-char query", query: "ne", wantSome: true},
		{name: "id prefix", query: "499", fields: []string{"id"}, wantSome: true},
		{name: "minecraft field only", query: "Player25", fields: []string{"minecraftUsername"}, wantSome: true},
		{name: "wildcard escaped scan", query: "_", wantSome: true},
		{name: "no results", query: "thisworddoesnotexist", wantSome: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			timing := measureSearch(t, store, tc.query, tc.fields, 1, 50)
			t.Logf("query=%q total=%d rows=%d min=%s avg=%s p95=%s max=%s",
				tc.query, timing.total, timing.rows,
				timing.min.Round(time.Microsecond),
				timing.avg.Round(time.Microsecond),
				timing.p95.Round(time.Microsecond),
				timing.max.Round(time.Microsecond),
			)

			if tc.wantSome && timing.total == 0 {
				t.Fatalf("query %q unexpectedly returned no results", tc.query)
			}
			if timing.avg > performanceMaxAvg {
				t.Fatalf("query %q too slow: avg=%s exceeds limit %s", tc.query, timing.avg, performanceMaxAvg)
			}
		})
	}
}

// TestPaginationPerformanceLargeDatabase measures LIMIT/OFFSET pagination cost
// across page depths and page sizes. Deep offsets are the interesting case
// because SQLite must still walk the skipped index entries.
func TestPaginationPerformanceLargeDatabase(t *testing.T) {
	store := seedPerformanceDatabase(t)
	defer store.Close()

	cases := []struct {
		name     string
		query    string
		fields   []string
		page     int
		pageSize int
		wantRows int
	}{
		{name: "page 1 size 50", page: 1, pageSize: 50, wantRows: 50},
		{name: "page 10 size 50", page: 10, pageSize: 50, wantRows: 50},
		{name: "page 50 size 50", page: 50, pageSize: 50, wantRows: 50},
		{name: "page 100 size 50 (last)", page: 100, pageSize: 50, wantRows: 50},
		{name: "page 1 size 100", page: 1, pageSize: 100, wantRows: 100},
		{name: "page 50 size 100 (last)", page: 50, pageSize: 100, wantRows: 100},
		{name: "page 25 size 20", page: 25, pageSize: 20, wantRows: 20},
		{name: "page beyond end", page: 101, pageSize: 50, wantRows: 0},
		{name: "search page 1", query: "redstone", page: 1, pageSize: 50, wantRows: 50},
		{name: "search deep page", query: "redstone", page: 90, pageSize: 50, wantRows: 50},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			timing := measureSearch(t, store, tc.query, tc.fields, tc.page, tc.pageSize)
			t.Logf("query=%q page=%d pageSize=%d total=%d rows=%d min=%s avg=%s p95=%s max=%s",
				tc.query, tc.page, tc.pageSize, timing.total, timing.rows,
				timing.min.Round(time.Microsecond),
				timing.avg.Round(time.Microsecond),
				timing.p95.Round(time.Microsecond),
				timing.max.Round(time.Microsecond),
			)

			if timing.rows != tc.wantRows {
				t.Fatalf("page %d size %d returned %d rows, want %d", tc.page, tc.pageSize, timing.rows, tc.wantRows)
			}
			if timing.avg > performanceMaxAvg {
				t.Fatalf("page %d size %d too slow: avg=%s exceeds limit %s", tc.page, tc.pageSize, timing.avg, performanceMaxAvg)
			}
		})
	}
}

// BenchmarkSearchApplicationsLargeDatabase measures steady-state query latency
// on the same generated dataset. Run with:
//
//	go test -tags sqlite_fts5 -run '^$' -bench BenchmarkSearchApplicationsLargeDatabase -benchtime 20x ./internal/database
func BenchmarkSearchApplicationsLargeDatabase(b *testing.B) {
	store := seedPerformanceDatabase(b)
	defer store.Close()

	queries := []struct {
		name   string
		query  string
		fields []string
	}{
		{name: "needle_word", query: "needleword"},
		{name: "needle_substring", query: "needle"},
		{name: "common_word", query: "redstone"},
		{name: "short_query", query: "ne"},
		{name: "id", query: "499", fields: []string{"id"}},
		{name: "paginated", query: "", fields: nil},
	}

	for _, q := range queries {
		b.Run(q.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				page := 1
				if q.name == "paginated" {
					page = performanceApps/50 + 1
				}
				if _, _, err := store.SearchApplications(q.query, q.fields, page, 50); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

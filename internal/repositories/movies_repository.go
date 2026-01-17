package repositories

import (
	"context"
	"fmt"
	"imdb-performance-optimization/internal/models"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type moviesRepository struct {
	optimizedPool   *pgxpool.Pool
	unoptimizedPool *pgxpool.Pool
}

func NewMoviesRepository(optimizedPool, unoptimizedPool *pgxpool.Pool) models.MoviesRepository {
	return &moviesRepository{
		optimizedPool:   optimizedPool,
		unoptimizedPool: unoptimizedPool,
	}
}

func (r *moviesRepository) selectPool(optimize bool) *pgxpool.Pool {
	if optimize {
		return r.optimizedPool
	}
	return r.unoptimizedPool
}

var GET_ALL_QUERY_BASE = `
SELECT tb.tconst, tt.name, tb.primary_title, tb.original_title, 
       tb.is_adult, tb.start_year, tb.end_year, tb.runtime_minutes,
       ARRAY(
           SELECT g.name FROM title_genres tg
           JOIN genres g ON g.id = tg.genre_id
           WHERE tg.tconst = tb.tconst
       ) AS genres,
       tr.average_rating rating,
       tr.num_votes votes
FROM title_basics tb
JOIN title_type tt ON tt.id = tb.title_type_id
LEFT JOIN title_ratings tr ON tr.tconst = tb.tconst`

var GET_ALL_QUERY_FILTER = `
WHERE tb.primary_title ILIKE '%' || $1 || '%'
   OR tb.original_title ILIKE '%' || $1 || '%'`

var GET_ALL_QUERY_LIMIT = `
LIMIT 20`

var GET_BY_ID_QUERY = `
	SELECT tb.tconst, tt.name, tb.primary_title, tb.original_title, 
	       tb.is_adult, tb.start_year, tb.end_year, tb.runtime_minutes,
	       COALESCE(tr.average_rating, 0) as average_rating,
	       COALESCE(tr.num_votes, 0) as num_votes,
	       ARRAY(
	           SELECT g.name FROM title_genres tg
	           JOIN genres g ON g.id = tg.genre_id
	           WHERE tg.tconst = tb.tconst
	       ) AS genres
	FROM title_basics tb
	LEFT JOIN title_type tt ON tt.id = tb.title_type_id
	LEFT JOIN title_ratings tr ON tr.tconst = tb.tconst
	WHERE tb.tconst = $1;
`

var GET_CAST_QUERY = `
	SELECT nb.primary_name, tp.characters, tp.ordering
	FROM title_principals tp
	JOIN name_basics nb ON nb.nconst = tp.nconst
	JOIN principal_categories pc ON pc.id = tp.category_id
	WHERE tp.tconst = $1 
	  AND pc.name IN ('actor', 'actress', 'self')
	ORDER BY tp.ordering;
`

var GET_CREW_QUERY = `
	SELECT nb.primary_name, pc.name, COALESCE(tp.job, ''), tp.ordering
	FROM title_principals tp
	JOIN name_basics nb ON nb.nconst = tp.nconst
	JOIN principal_categories pc ON pc.id = tp.category_id
	WHERE tp.tconst = $1 
	  AND pc.name NOT IN ('actor', 'actress', 'self')
	ORDER BY tp.ordering;
`

func buildGetAllQuery(title string) string {
	query := GET_ALL_QUERY_BASE
	if title != "" {
		query += "\n" + GET_ALL_QUERY_FILTER
	}
	query += "\n" + GET_ALL_QUERY_LIMIT
	return query
}

func (r *moviesRepository) GetAll(ctx context.Context, title string, optimize bool) (*models.Movies, error) {
	pool := r.selectPool(optimize)

	movies := &models.Movies{
		Items: []models.Movie{},
	}

	query := buildGetAllQuery(title)

	var rows interface {
		Next() bool
		Scan(dest ...interface{}) error
		Err() error
		Close()
	}
	var err error

	if title != "" {
		rows, err = pool.Query(ctx, query, title)
	} else {
		rows, err = pool.Query(ctx, query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var movie models.Movie
		err := rows.Scan(
			&movie.Tconst, &movie.TitleType, &movie.PrimaryTitle,
			&movie.OriginalTitle, &movie.IsAdult, &movie.StartYear,
			&movie.EndYear, &movie.RuntimeMinutes, &movie.Genres,
			&movie.Rating, &movie.Votes,
		)
		if err != nil {
			return nil, err
		}
		movies.Items = append(movies.Items, movie)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (r *moviesRepository) GetAllExplain(ctx context.Context, title string, optimize bool) (string, error) {
	pool := r.selectPool(optimize)
	query := buildGetAllQuery(title)

	if title != "" {
		return RunInExplain(ctx, pool, query, title)
	}
	return RunInExplain(ctx, pool, query)
}

func (r *moviesRepository) GetByID(ctx context.Context, tconst string, optimize bool) (*models.MovieDetail, error) {
	pool := r.selectPool(optimize)

	detail := &models.MovieDetail{}

	err := pool.QueryRow(ctx, GET_BY_ID_QUERY, tconst).Scan(
		&detail.Tconst, &detail.TitleType, &detail.PrimaryTitle,
		&detail.OriginalTitle, &detail.IsAdult, &detail.StartYear,
		&detail.EndYear, &detail.RuntimeMinutes, &detail.Rating,
		&detail.Votes, &detail.Genres,
	)
	if err != nil {
		return nil, err
	}

	castRows, err := pool.Query(ctx, GET_CAST_QUERY, tconst)
	if err != nil {
		return nil, err
	}
	defer castRows.Close()

	for castRows.Next() {
		var cast models.Cast
		var characters []string

		err := castRows.Scan(&cast.Name, &characters, &cast.Ordering)
		if err != nil {
			return nil, err
		}

		if len(characters) > 0 {
			cast.Character = characters[0]
		}

		detail.Cast = append(detail.Cast, cast)
	}

	crewRows, err := pool.Query(ctx, GET_CREW_QUERY, tconst)
	if err != nil {
		return nil, err
	}
	defer crewRows.Close()

	for crewRows.Next() {
		var crew models.Crew

		err := crewRows.Scan(&crew.Name, &crew.Category, &crew.Job, &crew.Ordering)
		if err != nil {
			return nil, err
		}

		detail.Crew = append(detail.Crew, crew)
	}

	return detail, nil
}

func (r *moviesRepository) GetByIDExplain(ctx context.Context, tconst string, optimize bool) (string, error) {
	pool := r.selectPool(optimize)

	var result strings.Builder
	var mainTime, castTime, crewTime float64

	result.WriteString("========== MAIN QUERY (Movie Details) ==========\n")
	mainExplain, err := RunInExplain(ctx, pool, GET_BY_ID_QUERY, tconst)
	if err != nil {
		return "", fmt.Errorf("main query explain failed: %w", err)
	}
	result.WriteString(mainExplain)
	mainTime = extractExecutionTime(mainExplain)
	result.WriteString("\n")

	result.WriteString("========== CAST QUERY ==========\n")
	castExplain, err := RunInExplain(ctx, pool, GET_CAST_QUERY, tconst)
	if err != nil {
		return "", fmt.Errorf("cast query explain failed: %w", err)
	}
	result.WriteString(castExplain)
	castTime = extractExecutionTime(castExplain)
	result.WriteString("\n")

	result.WriteString("========== CREW QUERY ==========\n")
	crewExplain, err := RunInExplain(ctx, pool, GET_CREW_QUERY, tconst)
	if err != nil {
		return "", fmt.Errorf("crew query explain failed: %w", err)
	}
	result.WriteString(crewExplain)
	crewTime = extractExecutionTime(crewExplain)
	result.WriteString("\n")

	result.WriteString("========== SUMMARY ==========\n")
	result.WriteString(fmt.Sprintf("Main Query (Movie Details): %.3f ms\n", mainTime))
	result.WriteString(fmt.Sprintf("Cast Query:                 %.3f ms\n", castTime))
	result.WriteString(fmt.Sprintf("Crew Query:                 %.3f ms\n", crewTime))
	fmt.Fprintf(&result, "─────────────────────────────────────\n")
	result.WriteString(fmt.Sprintf("TOTAL:                      %.3f ms\n", mainTime+castTime+crewTime))

	return result.String(), nil
}

func extractExecutionTime(explainOutput string) float64 {
	lines := strings.Split(explainOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Execution Time:") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "Time:" && i+1 < len(parts) {
					timeStr := strings.TrimSpace(parts[i+1])
					time, _ := strconv.ParseFloat(timeStr, 64)
					return time
				}
			}
		}
	}
	return 0.0
}

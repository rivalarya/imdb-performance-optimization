package models

import "context"

type Movie struct {
	Tconst         string   `json:"tconst"`
	TitleType      string   `json:"titleType"`
	PrimaryTitle   string   `json:"primaryTitle"`
	OriginalTitle  string   `json:"originalTitle"`
	IsAdult        bool     `json:"isAdult"`
	StartYear      *int16   `json:"startYear"`
	EndYear        *int16   `json:"endYear"`
	RuntimeMinutes *int16   `json:"runtimeMinutes"`
	Genres         []string `json:"genres"`
	Rating         *float32 `json:"rating"`
	Votes          *int     `json:"votes"`
}

type Movies struct {
	Items []Movie `json:"items"`
}

type MovieDetail struct {
	Movie
	Cast []Cast `json:"cast"`
	Crew []Crew `json:"crew"`
}

type Cast struct {
	Name      string `json:"name"`
	Character string `json:"character"`
	Ordering  int    `json:"ordering"`
}

type Crew struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Job      string `json:"job"`
	Ordering int    `json:"ordering"`
}

type MoviesRepository interface {
	GetAll(ctx context.Context, title string, optimize bool) (*Movies, error)
	GetAllExplain(ctx context.Context, title string, optimize bool) (string, error)
	GetByID(ctx context.Context, tconst string, optimize bool) (*MovieDetail, error)
	GetByIDExplain(ctx context.Context, tconst string, optimize bool) (string, error)
}

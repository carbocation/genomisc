package main

type Sample struct {
	ID           string  `db:"sample_id"`
	FileRow      int     `db:"file_row"`
	SumScore     float64 `db:"score"`
	NIncremented int     `db:"n_incremented"`
}

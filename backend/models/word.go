package models

type WordStatus string

const (
	StatusFamiliar   WordStatus = "familiar"
	StatusUnfamiliar WordStatus = "unfamiliar"
	StatusLearning   WordStatus = "learning"
)

type Word struct {
	Text        string     `json:"text"`
	Translation string     `json:"translation"`
	Status      WordStatus `json:"status"`
}

var Words map[string]Word

func init() {
	Words = make(map[string]Word)
}

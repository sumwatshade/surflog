package journal

import "github.com/sumwatshade/surflog/cmd/create"

type Journal struct {
	Entries []create.Entry `json:"entries"`
}

func (j *Journal) AddEntry(entry create.Entry) {
	j.Entries = append(j.Entries, entry)
}

func NewJournal() *Journal {
	return &Journal{
		Entries: []create.Entry{},
	}
}

package Data

import (
	"encoding/json"
	"io"
)

type Tweet struct {
	ID        string `json:"id"`
	Username  string `json:"username" validate:"required"`
	Body      string `json:"body" validate:"required"`
	CreatedOn string `json:"createdOn"`
	UpdatedOn string `json:"updatedOn"`
}

type Tweets []*Tweet

func (p *Tweets) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *Tweet) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *Tweet) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(p)
}

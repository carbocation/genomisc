package ukbb

import (
	"database/sql"
)

type DictionaryField struct {
	Path         string         `db:"Path"`
	Category     string         `db:"Category"`
	FieldID      int            `db:"FieldID"`
	Field        string         `db:"Field"`
	Participants int            `db:"Participants"`
	Items        int            `db:"Items"`
	Stability    string         `db:"Stability"`
	ValueType    string         `db:"ValueType"`
	Units        sql.NullString `db:"Units"`
	ItemType     string         `db:"ItemType"`
	Strata       string         `db:"Strata"`
	Sexed        string         `db:"Sexed"`
	Instances    int            `db:"Instances"`
	Array        int            `db:"Array"`
	Coding       sql.NullString `db:"coding_file_id"`
	Notes        sql.NullString `db:"Notes"`
	Link         string         `db:"Link"`
}

package genomisc

// Map columns in the BIM file to their positions
const (
	Chromosome int = iota
	VariantID
	Morgans
	Coordinate
	Allele1
	Allele2
)

type BIMRow struct {
	Chromosome string
	Coordinate uint32 // Labeled "position" by most applications
	VariantID  string // E.g., RSID
	Allele1    string // Can contain > 1 character
	Allele2    string // Can contain > 1 character
	// Morgans string // This is excluded intentionally
}

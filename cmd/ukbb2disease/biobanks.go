package main

const (
	BiobankSourceUKBiobank = "ukbb"
	BiobankSourceAllOfUs   = "aou"
)

func DoesBiobankUseNumericFieldID(biobank string) bool {
	switch biobank {
	case BiobankSourceUKBiobank:
		return true
	case BiobankSourceAllOfUs:
		return false
	}

	return false
}

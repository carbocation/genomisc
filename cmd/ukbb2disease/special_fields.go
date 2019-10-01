package main

var (
	// These identify which FieldIDs should be special-cased: either to the
	// HESIN table or the table with fields that have a specially-known date.
	// All other fields can be queried against the main phenotype table, with an
	// assumed date of whatever makes the most sense for your purposes
	// (generally, I use the enrollment date, but you could use birthdate, etc).

	MaterializedHesin = map[int]struct{}{
		41210: struct{}{},
		41202: struct{}{},
		41204: struct{}{},
		40001: struct{}{},
		40002: struct{}{},
		41200: struct{}{},
		41203: struct{}{},
		41205: struct{}{},
		40006: struct{}{}, // Cancer Registry ICD10
		40013: struct{}{}, // Cancer Registry ICD9
	}

	MaterializedSpecial = map[int]struct{}{
		42013: struct{}{},
		42011: struct{}{},
		42009: struct{}{},
		42007: struct{}{},
		42001: struct{}{},
		20004: struct{}{},
		20002: struct{}{},
		20001: struct{}{},
		40021: struct{}{},
		40020: struct{}{}, // Death
	}
)

var (
	// These can be helpful to make sure that the user is including all fields
	// that use the same family of codes

	ICD9 = map[int]struct{}{
		41203: struct{}{},
		41205: struct{}{},
		// 40013 Cancer Registry ICD9 is intentionally excluded, since asking
		// the user to include the cancer registry when they are not including
		// cancer codes would be counterproductive. TODO: recognize specific ICD
		// codes and make the prompt if the user includes cancer-specific codes
		// but not the cancer registry.
	}

	ICD10 = map[int]struct{}{
		41202: struct{}{},
		41204: struct{}{},
		40001: struct{}{},
		40002: struct{}{},
		// 40006 Cancer Registry ICD10 is intentionally excluded, since asking
		// the user to include the cancer registry when they are not including
		// cancer codes would be counterproductive. TODO: recognize specific ICD
		// codes and make the prompt if the user includes cancer-specific codes
		// but not the cancer registry.
	}

	OPCS = map[int]struct{}{
		41200: struct{}{},
		41210: struct{}{},
	}
)

func IsHesin(fieldID int) bool {
	_, exists := MaterializedHesin[fieldID]

	return exists
}

func IsSpecial(fieldID int) bool {
	_, exists := MaterializedSpecial[fieldID]

	return exists
}

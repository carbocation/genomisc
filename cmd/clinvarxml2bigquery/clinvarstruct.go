package main

import "encoding/xml"

// ReleaseSet was generated 2020-05-04 15:09:07 by james using zek
type ReleaseSet struct {
	XMLName                   xml.Name     `xml:"ReleaseSet"`
	Text                      string       `xml:",chardata"`
	Dated                     string       `xml:"Dated,attr"`
	Xsi                       string       `xml:"xsi,attr"`
	Type                      string       `xml:"Type,attr"`
	NoNamespaceSchemaLocation string       `xml:"noNamespaceSchemaLocation,attr"`
	ClinVarSet                []ClinVarSet `xml:"ClinVarSet"`
}

type ClinVarSet struct {
	Text                      string `xml:",chardata"`
	ID                        string `xml:"ID,attr"`
	RecordStatus              string `xml:"RecordStatus"`
	Title                     string `xml:"Title"`
	ReferenceClinVarAssertion struct {
		Text             string `xml:",chardata"`
		DateCreated      string `xml:"DateCreated,attr"`
		DateLastUpdated  string `xml:"DateLastUpdated,attr"`
		ID               string `xml:"ID,attr"`
		ClinVarAccession struct {
			Text        string `xml:",chardata"`
			Acc         string `xml:"Acc,attr"`
			Version     string `xml:"Version,attr"`
			Type        string `xml:"Type,attr"`
			DateUpdated string `xml:"DateUpdated,attr"`
		} `xml:"ClinVarAccession"`
		RecordStatus         string `xml:"RecordStatus"`
		ClinicalSignificance struct {
			Text              string `xml:",chardata"`
			DateLastEvaluated string `xml:"DateLastEvaluated,attr"`
			ReviewStatus      string `xml:"ReviewStatus"`
			Description       string `xml:"Description"`
			Explanation       struct {
				Text       string `xml:",chardata"`
				DataSource string `xml:"DataSource,attr"`
				Type       string `xml:"Type,attr"`
			} `xml:"Explanation"`
		} `xml:"ClinicalSignificance"`
		Assertion struct {
			Text string `xml:",chardata"`
			Type string `xml:"Type,attr"`
		} `xml:"Assertion"`
		ObservedIn []struct {
			Text   string `xml:",chardata"`
			Sample struct {
				Text    string `xml:",chardata"`
				Origin  string `xml:"Origin"`
				Species struct {
					Text       string `xml:",chardata"`
					TaxonomyId string `xml:"TaxonomyId,attr"`
				} `xml:"Species"`
				AffectedStatus string `xml:"AffectedStatus"`
				FamilyData     struct {
					Text        string `xml:",chardata"`
					NumFamilies string `xml:"NumFamilies,attr"`
				} `xml:"FamilyData"`
				Ethnicity     string `xml:"Ethnicity"`
				NumberTested  string `xml:"NumberTested"`
				NumberMales   string `xml:"NumberMales"`
				NumberFemales string `xml:"NumberFemales"`
			} `xml:"Sample"`
			Method []struct {
				Text         string `xml:",chardata"`
				MethodType   string `xml:"MethodType"`
				Description  string `xml:"Description"`
				NamePlatform string `xml:"NamePlatform"`
				Purpose      string `xml:"Purpose"`
				ResultType   string `xml:"ResultType"`
			} `xml:"Method"`
			ObservedData []struct {
				Text      string `xml:",chardata"`
				ID        string `xml:"ID,attr"`
				Attribute struct {
					Text         string `xml:",chardata"`
					Type         string `xml:"Type,attr"`
					IntegerValue string `xml:"integerValue,attr"`
				} `xml:"Attribute"`
				Citation []struct {
					Text   string `xml:",chardata"`
					Type   string `xml:"Type,attr"`
					Abbrev string `xml:"Abbrev,attr"`
					ID     []struct {
						Text   string `xml:",chardata"`
						Source string `xml:"Source,attr"`
					} `xml:"ID"`
					CitationText string `xml:"CitationText"`
				} `xml:"Citation"`
			} `xml:"ObservedData"`
		} `xml:"ObservedIn"`
		MeasureSet struct {
			Text    string `xml:",chardata"`
			Type    string `xml:"Type,attr"`
			ID      string `xml:"ID,attr"`
			Acc     string `xml:"Acc,attr"`
			Version string `xml:"Version,attr"`
			Measure []struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
				ID   string `xml:"ID,attr"`
				Name []struct {
					Text         string `xml:",chardata"`
					ElementValue struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"ElementValue"`
					XRef []struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"XRef"`
					Citation struct {
						Text   string `xml:",chardata"`
						Type   string `xml:"Type,attr"`
						Abbrev string `xml:"Abbrev,attr"`
						ID     []struct {
							Text   string `xml:",chardata"`
							Source string `xml:"Source,attr"`
						} `xml:"ID"`
						URL string `xml:"URL"`
					} `xml:"Citation"`
				} `xml:"Name"`
				AttributeSet []struct {
					Text      string `xml:",chardata"`
					Attribute struct {
						Text         string `xml:",chardata"`
						Type         string `xml:"Type,attr"`
						IntegerValue string `xml:"integerValue,attr"`
						Change       string `xml:"Change,attr"`
						Accession    string `xml:"Accession,attr"`
						Version      string `xml:"Version,attr"`
					} `xml:"Attribute"`
					XRef []struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"XRef"`
					Comment []struct {
						Text       string `xml:",chardata"`
						DataSource string `xml:"DataSource,attr"`
						Type       string `xml:"Type,attr"`
					} `xml:"Comment"`
					Citation []struct {
						Text   string `xml:",chardata"`
						Type   string `xml:"Type,attr"`
						Abbrev string `xml:"Abbrev,attr"`
						ID     []struct {
							Text   string `xml:",chardata"`
							Source string `xml:"Source,attr"`
						} `xml:"ID"`
					} `xml:"Citation"`
				} `xml:"AttributeSet"`
				CytogeneticLocation []string `xml:"CytogeneticLocation"`
				MeasureRelationship []struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
					Name struct {
						Text         string `xml:",chardata"`
						ElementValue struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
						} `xml:"ElementValue"`
					} `xml:"Name"`
					Symbol struct {
						Text         string `xml:",chardata"`
						ElementValue struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
						} `xml:"ElementValue"`
					} `xml:"Symbol"`
					SequenceLocation []struct {
						Text                     string `xml:",chardata"`
						Assembly                 string `xml:"Assembly,attr"`
						AssemblyAccessionVersion string `xml:"AssemblyAccessionVersion,attr"`
						AssemblyStatus           string `xml:"AssemblyStatus,attr"`
						Chr                      string `xml:"Chr,attr"`
						Accession                string `xml:"Accession,attr"`
						Start                    string `xml:"start,attr"`
						Stop                     string `xml:"stop,attr"`
						DisplayStart             string `xml:"display_start,attr"`
						DisplayStop              string `xml:"display_stop,attr"`
						Strand                   string `xml:"Strand,attr"`
						VariantLength            string `xml:"variantLength,attr"`
					} `xml:"SequenceLocation"`
					XRef []struct {
						Text string `xml:",chardata"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
						Type string `xml:"Type,attr"`
					} `xml:"XRef"`
					AttributeSet []struct {
						Text      string `xml:",chardata"`
						Attribute struct {
							Text      string `xml:",chardata"`
							DateValue string `xml:"dateValue,attr"`
							Type      string `xml:"Type,attr"`
						} `xml:"Attribute"`
						Citation struct {
							Text string `xml:",chardata"`
							URL  string `xml:"URL"`
						} `xml:"Citation"`
					} `xml:"AttributeSet"`
					Comment []struct {
						Text       string `xml:",chardata"`
						DataSource string `xml:"DataSource,attr"`
						Type       string `xml:"Type,attr"`
					} `xml:"Comment"`
				} `xml:"MeasureRelationship"`
				XRef []struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
				} `xml:"XRef"`
				Citation []struct {
					Text   string `xml:",chardata"`
					Type   string `xml:"Type,attr"`
					Abbrev string `xml:"Abbrev,attr"`
					ID     []struct {
						Text   string `xml:",chardata"`
						Source string `xml:"Source,attr"`
					} `xml:"ID"`
					URL          string `xml:"URL"`
					CitationText string `xml:"CitationText"`
				} `xml:"Citation"`
				Comment []struct {
					Text       string `xml:",chardata"`
					DataSource string `xml:"DataSource,attr"`
					Type       string `xml:"Type,attr"`
				} `xml:"Comment"`
				SequenceLocation []struct {
					Text                     string `xml:",chardata"`
					Assembly                 string `xml:"Assembly,attr"`
					AssemblyAccessionVersion string `xml:"AssemblyAccessionVersion,attr"`
					AssemblyStatus           string `xml:"AssemblyStatus,attr"`
					Chr                      string `xml:"Chr,attr"`
					Accession                string `xml:"Accession,attr"`
					Start                    string `xml:"start,attr"`
					Stop                     string `xml:"stop,attr"`
					DisplayStart             string `xml:"display_start,attr"`
					DisplayStop              string `xml:"display_stop,attr"`
					VariantLength            string `xml:"variantLength,attr"`
					ReferenceAllele          string `xml:"referenceAllele,attr"`
					AlternateAllele          string `xml:"alternateAllele,attr"`
					InnerStart               string `xml:"innerStart,attr"`
					InnerStop                string `xml:"innerStop,attr"`
					OuterStart               string `xml:"outerStart,attr"`
					OuterStop                string `xml:"outerStop,attr"`
					PositionVCF              string `xml:"positionVCF,attr"`
					ReferenceAlleleVCF       string `xml:"referenceAlleleVCF,attr"`
					AlternateAlleleVCF       string `xml:"alternateAlleleVCF,attr"`
					Strand                   string `xml:"Strand,attr"`
				} `xml:"SequenceLocation"`
				CanonicalSPDI       string `xml:"CanonicalSPDI"`
				AlleleFrequencyList struct {
					Text            string `xml:",chardata"`
					AlleleFrequency []struct {
						Text   string `xml:",chardata"`
						Value  string `xml:"Value,attr"`
						Source string `xml:"Source,attr"`
					} `xml:"AlleleFrequency"`
				} `xml:"AlleleFrequencyList"`
				GlobalMinorAlleleFrequency struct {
					Text        string `xml:",chardata"`
					Value       string `xml:"Value,attr"`
					Source      string `xml:"Source,attr"`
					MinorAllele string `xml:"MinorAllele,attr"`
				} `xml:"GlobalMinorAlleleFrequency"`
				Symbol struct {
					Text         string `xml:",chardata"`
					ElementValue struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"ElementValue"`
					Citation struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
						ID   struct {
							Text   string `xml:",chardata"`
							Source string `xml:"Source,attr"`
						} `xml:"ID"`
					} `xml:"Citation"`
					XRef struct {
						Text string `xml:",chardata"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"XRef"`
				} `xml:"Symbol"`
				Source struct {
					Text       string `xml:",chardata"`
					DataSource string `xml:"DataSource,attr"`
					ID         string `xml:"ID,attr"`
				} `xml:"Source"`
			} `xml:"Measure"`
			Name []struct {
				Text         string `xml:",chardata"`
				ElementValue struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"ElementValue"`
				XRef []struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
				} `xml:"XRef"`
				Citation struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
					ID   struct {
						Text   string `xml:",chardata"`
						Source string `xml:"Source,attr"`
					} `xml:"ID"`
				} `xml:"Citation"`
			} `xml:"Name"`
			XRef []struct {
				Text string `xml:",chardata"`
				ID   string `xml:"ID,attr"`
				DB   string `xml:"DB,attr"`
				Type string `xml:"Type,attr"`
			} `xml:"XRef"`
			AttributeSet []struct {
				Text      string `xml:",chardata"`
				Attribute struct {
					Text         string `xml:",chardata"`
					Type         string `xml:"Type,attr"`
					Change       string `xml:"Change,attr"`
					IntegerValue string `xml:"integerValue,attr"`
				} `xml:"Attribute"`
				Citation struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
					ID   struct {
						Text   string `xml:",chardata"`
						Source string `xml:"Source,attr"`
					} `xml:"ID"`
				} `xml:"Citation"`
				XRef struct {
					Text string `xml:",chardata"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
				} `xml:"XRef"`
			} `xml:"AttributeSet"`
			Symbol struct {
				Text         string `xml:",chardata"`
				ElementValue struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"ElementValue"`
			} `xml:"Symbol"`
			Citation struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
				ID   struct {
					Text   string `xml:",chardata"`
					Source string `xml:"Source,attr"`
				} `xml:"ID"`
			} `xml:"Citation"`
		} `xml:"MeasureSet"`
		TraitSet struct {
			Text  string `xml:",chardata"`
			Type  string `xml:"Type,attr"`
			ID    string `xml:"ID,attr"`
			Trait []struct {
				Text string `xml:",chardata"`
				ID   string `xml:"ID,attr"`
				Type string `xml:"Type,attr"`
				Name []struct {
					Text         string `xml:",chardata"`
					ElementValue struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"ElementValue"`
					XRef []struct {
						Text string `xml:",chardata"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
						Type string `xml:"Type,attr"`
					} `xml:"XRef"`
					Citation struct {
						Text   string `xml:",chardata"`
						Type   string `xml:"Type,attr"`
						Abbrev string `xml:"Abbrev,attr"`
						ID     struct {
							Text   string `xml:",chardata"`
							Source string `xml:"Source,attr"`
						} `xml:"ID"`
						URL string `xml:"URL"`
					} `xml:"Citation"`
				} `xml:"Name"`
				AttributeSet []struct {
					Text      string `xml:",chardata"`
					Attribute struct {
						Text         string `xml:",chardata"`
						Type         string `xml:"Type,attr"`
						IntegerValue string `xml:"integerValue,attr"`
					} `xml:"Attribute"`
					XRef []struct {
						Text string `xml:",chardata"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
						Type string `xml:"Type,attr"`
					} `xml:"XRef"`
					Citation struct {
						Text   string `xml:",chardata"`
						Type   string `xml:"Type,attr"`
						Abbrev string `xml:"Abbrev,attr"`
						ID     []struct {
							Text   string `xml:",chardata"`
							Source string `xml:"Source,attr"`
						} `xml:"ID"`
					} `xml:"Citation"`
				} `xml:"AttributeSet"`
				Citation []struct {
					Text   string `xml:",chardata"`
					Type   string `xml:"Type,attr"`
					Abbrev string `xml:"Abbrev,attr"`
					ID     []struct {
						Text   string `xml:",chardata"`
						Source string `xml:"Source,attr"`
					} `xml:"ID"`
					URL          string `xml:"URL"`
					CitationText string `xml:"CitationText"`
				} `xml:"Citation"`
				XRef []struct {
					Text string `xml:",chardata"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
					Type string `xml:"Type,attr"`
				} `xml:"XRef"`
				Symbol []struct {
					Text         string `xml:",chardata"`
					ElementValue struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"ElementValue"`
					XRef []struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"XRef"`
					Citation struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
						ID   struct {
							Text   string `xml:",chardata"`
							Source string `xml:"Source,attr"`
						} `xml:"ID"`
					} `xml:"Citation"`
				} `xml:"Symbol"`
				TraitRelationship struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
					ID   string `xml:"ID,attr"`
				} `xml:"TraitRelationship"`
				Comment struct {
					Text       string `xml:",chardata"`
					DataSource string `xml:"DataSource,attr"`
					Type       string `xml:"Type,attr"`
				} `xml:"Comment"`
			} `xml:"Trait"`
			Name struct {
				Text         string `xml:",chardata"`
				ElementValue struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"ElementValue"`
			} `xml:"Name"`
			Symbol struct {
				Text         string `xml:",chardata"`
				ElementValue struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"ElementValue"`
			} `xml:"Symbol"`
		} `xml:"TraitSet"`
		Citation []struct {
			Text   string `xml:",chardata"`
			Type   string `xml:"Type,attr"`
			Abbrev string `xml:"Abbrev,attr"`
			ID     struct {
				Text   string `xml:",chardata"`
				Source string `xml:"Source,attr"`
			} `xml:"ID"`
		} `xml:"Citation"`
		AttributeSet []struct {
			Text      string `xml:",chardata"`
			Attribute struct {
				Text         string `xml:",chardata"`
				Type         string `xml:"Type,attr"`
				IntegerValue string `xml:"integerValue,attr"`
			} `xml:"Attribute"`
			XRef []struct {
				Text string `xml:",chardata"`
				ID   string `xml:"ID,attr"`
				DB   string `xml:"DB,attr"`
			} `xml:"XRef"`
			Citation struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
				ID   struct {
					Text   string `xml:",chardata"`
					Source string `xml:"Source,attr"`
				} `xml:"ID"`
			} `xml:"Citation"`
		} `xml:"AttributeSet"`
		GenotypeSet struct {
			Text       string `xml:",chardata"`
			Type       string `xml:"Type,attr"`
			ID         string `xml:"ID,attr"`
			Acc        string `xml:"Acc,attr"`
			Version    string `xml:"Version,attr"`
			MeasureSet []struct {
				Text                string `xml:",chardata"`
				Type                string `xml:"Type,attr"`
				ID                  string `xml:"ID,attr"`
				Acc                 string `xml:"Acc,attr"`
				Version             string `xml:"Version,attr"`
				NumberOfChromosomes string `xml:"NumberOfChromosomes,attr"`
				Measure             []struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
					ID   string `xml:"ID,attr"`
					Name []struct {
						Text         string `xml:",chardata"`
						ElementValue struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
						} `xml:"ElementValue"`
						XRef struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
							ID   string `xml:"ID,attr"`
							DB   string `xml:"DB,attr"`
						} `xml:"XRef"`
						Citation struct {
							Text   string `xml:",chardata"`
							Type   string `xml:"Type,attr"`
							Abbrev string `xml:"Abbrev,attr"`
							ID     []struct {
								Text   string `xml:",chardata"`
								Source string `xml:"Source,attr"`
							} `xml:"ID"`
						} `xml:"Citation"`
					} `xml:"Name"`
					AttributeSet []struct {
						Text      string `xml:",chardata"`
						Attribute struct {
							Text         string `xml:",chardata"`
							Accession    string `xml:"Accession,attr"`
							Version      string `xml:"Version,attr"`
							Change       string `xml:"Change,attr"`
							Type         string `xml:"Type,attr"`
							IntegerValue string `xml:"integerValue,attr"`
						} `xml:"Attribute"`
						XRef []struct {
							Text string `xml:",chardata"`
							ID   string `xml:"ID,attr"`
							DB   string `xml:"DB,attr"`
						} `xml:"XRef"`
						Citation struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
							ID   struct {
								Text   string `xml:",chardata"`
								Source string `xml:"Source,attr"`
							} `xml:"ID"`
						} `xml:"Citation"`
					} `xml:"AttributeSet"`
					CytogeneticLocation string `xml:"CytogeneticLocation"`
					SequenceLocation    []struct {
						Text                     string `xml:",chardata"`
						Assembly                 string `xml:"Assembly,attr"`
						AssemblyAccessionVersion string `xml:"AssemblyAccessionVersion,attr"`
						AssemblyStatus           string `xml:"AssemblyStatus,attr"`
						Chr                      string `xml:"Chr,attr"`
						Accession                string `xml:"Accession,attr"`
						Start                    string `xml:"start,attr"`
						Stop                     string `xml:"stop,attr"`
						DisplayStart             string `xml:"display_start,attr"`
						DisplayStop              string `xml:"display_stop,attr"`
						VariantLength            string `xml:"variantLength,attr"`
						PositionVCF              string `xml:"positionVCF,attr"`
						ReferenceAlleleVCF       string `xml:"referenceAlleleVCF,attr"`
						AlternateAlleleVCF       string `xml:"alternateAlleleVCF,attr"`
						OuterStart               string `xml:"outerStart,attr"`
						InnerStart               string `xml:"innerStart,attr"`
						InnerStop                string `xml:"innerStop,attr"`
						OuterStop                string `xml:"outerStop,attr"`
					} `xml:"SequenceLocation"`
					MeasureRelationship []struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
						Name struct {
							Text         string `xml:",chardata"`
							ElementValue struct {
								Text string `xml:",chardata"`
								Type string `xml:"Type,attr"`
							} `xml:"ElementValue"`
						} `xml:"Name"`
						Symbol struct {
							Text         string `xml:",chardata"`
							ElementValue struct {
								Text string `xml:",chardata"`
								Type string `xml:"Type,attr"`
							} `xml:"ElementValue"`
						} `xml:"Symbol"`
						SequenceLocation []struct {
							Text                     string `xml:",chardata"`
							Assembly                 string `xml:"Assembly,attr"`
							AssemblyAccessionVersion string `xml:"AssemblyAccessionVersion,attr"`
							AssemblyStatus           string `xml:"AssemblyStatus,attr"`
							Chr                      string `xml:"Chr,attr"`
							Accession                string `xml:"Accession,attr"`
							Start                    string `xml:"start,attr"`
							Stop                     string `xml:"stop,attr"`
							DisplayStart             string `xml:"display_start,attr"`
							DisplayStop              string `xml:"display_stop,attr"`
							Strand                   string `xml:"Strand,attr"`
							VariantLength            string `xml:"variantLength,attr"`
						} `xml:"SequenceLocation"`
						XRef []struct {
							Text string `xml:",chardata"`
							ID   string `xml:"ID,attr"`
							DB   string `xml:"DB,attr"`
							Type string `xml:"Type,attr"`
						} `xml:"XRef"`
						AttributeSet []struct {
							Text      string `xml:",chardata"`
							Attribute struct {
								Text      string `xml:",chardata"`
								DateValue string `xml:"dateValue,attr"`
								Type      string `xml:"Type,attr"`
							} `xml:"Attribute"`
							Citation struct {
								Text string `xml:",chardata"`
								URL  string `xml:"URL"`
							} `xml:"Citation"`
						} `xml:"AttributeSet"`
						Comment []struct {
							Text       string `xml:",chardata"`
							DataSource string `xml:"DataSource,attr"`
							Type       string `xml:"Type,attr"`
						} `xml:"Comment"`
					} `xml:"MeasureRelationship"`
					XRef []struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"XRef"`
					CanonicalSPDI string `xml:"CanonicalSPDI"`
					Comment       struct {
						Text       string `xml:",chardata"`
						DataSource string `xml:"DataSource,attr"`
						Type       string `xml:"Type,attr"`
					} `xml:"Comment"`
					AlleleFrequencyList struct {
						Text            string `xml:",chardata"`
						AlleleFrequency []struct {
							Text   string `xml:",chardata"`
							Value  string `xml:"Value,attr"`
							Source string `xml:"Source,attr"`
						} `xml:"AlleleFrequency"`
					} `xml:"AlleleFrequencyList"`
					Citation []struct {
						Text   string `xml:",chardata"`
						Type   string `xml:"Type,attr"`
						Abbrev string `xml:"Abbrev,attr"`
						ID     []struct {
							Text   string `xml:",chardata"`
							Source string `xml:"Source,attr"`
						} `xml:"ID"`
						URL          string `xml:"URL"`
						CitationText string `xml:"CitationText"`
					} `xml:"Citation"`
					GlobalMinorAlleleFrequency struct {
						Text        string `xml:",chardata"`
						Value       string `xml:"Value,attr"`
						Source      string `xml:"Source,attr"`
						MinorAllele string `xml:"MinorAllele,attr"`
					} `xml:"GlobalMinorAlleleFrequency"`
					Symbol struct {
						Text         string `xml:",chardata"`
						ElementValue struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
						} `xml:"ElementValue"`
					} `xml:"Symbol"`
				} `xml:"Measure"`
				AttributeSet []struct {
					Text      string `xml:",chardata"`
					Attribute struct {
						Text         string `xml:",chardata"`
						Type         string `xml:"Type,attr"`
						Change       string `xml:"Change,attr"`
						IntegerValue string `xml:"integerValue,attr"`
					} `xml:"Attribute"`
				} `xml:"AttributeSet"`
				XRef []struct {
					Text string `xml:",chardata"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
					Type string `xml:"Type,attr"`
				} `xml:"XRef"`
				Name []struct {
					Text         string `xml:",chardata"`
					ElementValue struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"ElementValue"`
					XRef []struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"XRef"`
				} `xml:"Name"`
			} `xml:"MeasureSet"`
			Name []struct {
				Text         string `xml:",chardata"`
				ElementValue struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"ElementValue"`
				XRef []struct {
					Text string `xml:",chardata"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
				} `xml:"XRef"`
			} `xml:"Name"`
			AttributeSet []struct {
				Text      string `xml:",chardata"`
				Attribute struct {
					Text   string `xml:",chardata"`
					Type   string `xml:"Type,attr"`
					Change string `xml:"Change,attr"`
				} `xml:"Attribute"`
			} `xml:"AttributeSet"`
			XRef struct {
				Text string `xml:",chardata"`
				ID   string `xml:"ID,attr"`
				DB   string `xml:"DB,attr"`
			} `xml:"XRef"`
		} `xml:"GenotypeSet"`
	} `xml:"ReferenceClinVarAssertion"`
	ClinVarAssertion []ClinVarAssertion `xml:"ClinVarAssertion"`
	Replaces         []string           `xml:"Replaces"`
}

type ClinVarAssertion struct {
	Text                  string `xml:",chardata"`
	ID                    string `xml:"ID,attr"`
	SubmissionName        string `xml:"SubmissionName,attr"`
	FDARecognizedDatabase string `xml:"FDARecognizedDatabase,attr"`
	ClinVarSubmissionID   struct {
		Text                string `xml:",chardata"`
		LocalKey            string `xml:"localKey,attr"`
		Submitter           string `xml:"submitter,attr"`
		SubmitterDate       string `xml:"submitterDate,attr"`
		Title               string `xml:"title,attr"`
		SubmittedAssembly   string `xml:"submittedAssembly,attr"`
		LocalKeyIsSubmitted string `xml:"localKeyIsSubmitted,attr"`
	} `xml:"ClinVarSubmissionID"`
	ClinVarAccession struct {
		Text                 string `xml:",chardata"`
		Acc                  string `xml:"Acc,attr"`
		Version              string `xml:"Version,attr"`
		Type                 string `xml:"Type,attr"`
		OrgID                string `xml:"OrgID,attr"`
		OrganizationCategory string `xml:"OrganizationCategory,attr"`
		OrgType              string `xml:"OrgType,attr"`
		DateUpdated          string `xml:"DateUpdated,attr"`
	} `xml:"ClinVarAccession"`
	RecordStatus         string `xml:"RecordStatus"`
	ClinicalSignificance struct {
		Text              string `xml:",chardata"`
		DateLastEvaluated string `xml:"DateLastEvaluated,attr"`
		ReviewStatus      string `xml:"ReviewStatus"`
		Description       string `xml:"Description"`
		Comment           []struct {
			Text string `xml:",chardata"`
			Type string `xml:"Type,attr"`
		} `xml:"Comment"`
		Citation []struct {
			Text string `xml:",chardata"`
			Type string `xml:"Type,attr"`
			ID   struct {
				Text   string `xml:",chardata"`
				Source string `xml:"Source,attr"`
			} `xml:"ID"`
			URL          string `xml:"URL"`
			CitationText string `xml:"CitationText"`
		} `xml:"Citation"`
		ExplanationOfInterpretation string `xml:"ExplanationOfInterpretation"`
	} `xml:"ClinicalSignificance"`
	Assertion struct {
		Text string `xml:",chardata"`
		Type string `xml:"Type,attr"`
	} `xml:"Assertion"`
	ExternalID struct {
		Text string `xml:",chardata"`
		DB   string `xml:"DB,attr"`
		ID   string `xml:"ID,attr"`
		Type string `xml:"Type,attr"`
		URL  string `xml:"URL,attr"`
	} `xml:"ExternalID"`
	ObservedIn []struct {
		Text   string `xml:",chardata"`
		Sample struct {
			Text    string `xml:",chardata"`
			Origin  string `xml:"Origin"`
			Species struct {
				Text       string `xml:",chardata"`
				TaxonomyId string `xml:"TaxonomyId,attr"`
			} `xml:"Species"`
			AffectedStatus string `xml:"AffectedStatus"`
			FamilyData     struct {
				Text                               string `xml:",chardata"`
				NumFamilies                        string `xml:"NumFamilies,attr"`
				NumFamiliesWithVariant             string `xml:"NumFamiliesWithVariant,attr"`
				NumFamiliesWithSegregationObserved string `xml:"NumFamiliesWithSegregationObserved,attr"`
				PedigreeID                         string `xml:"PedigreeID,attr"`
				SegregationObserved                string `xml:"SegregationObserved,attr"`
				FamilyHistory                      string `xml:"FamilyHistory"`
			} `xml:"FamilyData"`
			Ethnicity string `xml:"Ethnicity"`
			Age       []struct {
				Text    string `xml:",chardata"`
				Type    string `xml:"Type,attr"`
				AgeUnit string `xml:"age_unit,attr"`
			} `xml:"Age"`
			Gender           string `xml:"Gender"`
			GeographicOrigin string `xml:"GeographicOrigin"`
			NumberTested     string `xml:"NumberTested"`
			Tissue           string `xml:"Tissue"`
			Indication       struct {
				Text  string `xml:",chardata"`
				Type  string `xml:"Type,attr"`
				Trait []struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
					Name struct {
						Text         string `xml:",chardata"`
						ElementValue struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
						} `xml:"ElementValue"`
					} `xml:"Name"`
					XRef struct {
						Text string `xml:",chardata"`
						DB   string `xml:"DB,attr"`
						ID   string `xml:"ID,attr"`
						Type string `xml:"Type,attr"`
					} `xml:"XRef"`
				} `xml:"Trait"`
			} `xml:"Indication"`
			CellLine          string `xml:"CellLine"`
			Proband           string `xml:"Proband"`
			SampleDescription struct {
				Text        string `xml:",chardata"`
				Description struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"Description"`
			} `xml:"SampleDescription"`
			Strain string `xml:"Strain"`
		} `xml:"Sample"`
		Method []struct {
			Text            string `xml:",chardata"`
			MethodType      string `xml:"MethodType"`
			TypePlatform    string `xml:"TypePlatform"`
			MethodAttribute []struct {
				Text      string `xml:",chardata"`
				Attribute struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"Attribute"`
			} `xml:"MethodAttribute"`
			XRef struct {
				Text string `xml:",chardata"`
				DB   string `xml:"DB,attr"`
				ID   string `xml:"ID,attr"`
				URL  string `xml:"URL,attr"`
			} `xml:"XRef"`
			Citation []struct {
				Text string `xml:",chardata"`
				ID   struct {
					Text   string `xml:",chardata"`
					Source string `xml:"Source,attr"`
				} `xml:"ID"`
			} `xml:"Citation"`
			Description  string `xml:"Description"`
			NamePlatform string `xml:"NamePlatform"`
			Purpose      string `xml:"Purpose"`
			Software     []struct {
				Text    string `xml:",chardata"`
				Name    string `xml:"name,attr"`
				Purpose string `xml:"purpose,attr"`
				Version string `xml:"version,attr"`
			} `xml:"Software"`
			SourceType         string `xml:"SourceType"`
			ObsMethodAttribute struct {
				Text      string `xml:",chardata"`
				Attribute struct {
					Text         string `xml:",chardata"`
					Type         string `xml:"Type,attr"`
					DateValue    string `xml:"dateValue,attr"`
					IntegerValue string `xml:"integerValue,attr"`
				} `xml:"Attribute"`
				Comment struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"Comment"`
			} `xml:"ObsMethodAttribute"`
			ResultType string `xml:"ResultType"`
		} `xml:"Method"`
		ObservedData []struct {
			Text      string `xml:",chardata"`
			Attribute struct {
				Text         string `xml:",chardata"`
				Type         string `xml:"Type,attr"`
				IntegerValue string `xml:"integerValue,attr"`
			} `xml:"Attribute"`
			Citation []struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
				ID   struct {
					Text   string `xml:",chardata"`
					Source string `xml:"Source,attr"`
				} `xml:"ID"`
				CitationText string `xml:"CitationText"`
			} `xml:"Citation"`
			XRef []struct {
				Text string `xml:",chardata"`
				DB   string `xml:"DB,attr"`
				ID   string `xml:"ID,attr"`
				Type string `xml:"Type,attr"`
			} `xml:"XRef"`
			Comment struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
			} `xml:"Comment"`
		} `xml:"ObservedData"`
		Comment []struct {
			Text string `xml:",chardata"`
			Type string `xml:"Type,attr"`
		} `xml:"Comment"`
		TraitSet struct {
			Text              string `xml:",chardata"`
			Type              string `xml:"Type,attr"`
			DateLastEvaluated string `xml:"DateLastEvaluated,attr"`
			Trait             []struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
				Name struct {
					Text         string `xml:",chardata"`
					ElementValue struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"ElementValue"`
					XRef struct {
						Text string `xml:",chardata"`
						DB   string `xml:"DB,attr"`
						ID   string `xml:"ID,attr"`
						Type string `xml:"Type,attr"`
					} `xml:"XRef"`
				} `xml:"Name"`
				XRef struct {
					Text   string `xml:",chardata"`
					DB     string `xml:"DB,attr"`
					ID     string `xml:"ID,attr"`
					Status string `xml:"Status,attr"`
				} `xml:"XRef"`
			} `xml:"Trait"`
			Comment struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
			} `xml:"Comment"`
		} `xml:"TraitSet"`
		Citation []struct {
			Text string `xml:",chardata"`
			ID   struct {
				Text   string `xml:",chardata"`
				Source string `xml:"Source,attr"`
			} `xml:"ID"`
			URL          string `xml:"URL"`
			CitationText string `xml:"CitationText"`
		} `xml:"Citation"`
		XRef struct {
			Text string `xml:",chardata"`
			DB   string `xml:"DB,attr"`
			ID   string `xml:"ID,attr"`
			Type string `xml:"Type,attr"`
		} `xml:"XRef"`
		CoOccurrenceSet []struct {
			Text          string `xml:",chardata"`
			Zygosity      string `xml:"Zygosity"`
			AlleleDescSet []struct {
				Text                 string `xml:",chardata"`
				Name                 string `xml:"Name"`
				Zygosity             string `xml:"Zygosity"`
				ClinicalSignificance struct {
					Text         string `xml:",chardata"`
					ReviewStatus string `xml:"ReviewStatus"`
					Description  string `xml:"Description"`
				} `xml:"ClinicalSignificance"`
			} `xml:"AlleleDescSet"`
			Count string `xml:"Count"`
		} `xml:"Co-occurrenceSet"`
	} `xml:"ObservedIn"`
	MeasureSet struct {
		Text    string `xml:",chardata"`
		Type    string `xml:"Type,attr"`
		Measure []struct {
			Text string `xml:",chardata"`
			Type string `xml:"Type,attr"`
			Name []struct {
				Text         string `xml:",chardata"`
				ElementValue struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"ElementValue"`
			} `xml:"Name"`
			AttributeSet []struct {
				Text      string `xml:",chardata"`
				Attribute struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"Attribute"`
				XRef struct {
					Text string `xml:",chardata"`
					DB   string `xml:"DB,attr"`
					ID   string `xml:"ID,attr"`
					URL  string `xml:"URL,attr"`
				} `xml:"XRef"`
				Comment struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"Comment"`
			} `xml:"AttributeSet"`
			MeasureRelationship []struct {
				Text   string `xml:",chardata"`
				Type   string `xml:"Type,attr"`
				Symbol struct {
					Text         string `xml:",chardata"`
					ElementValue struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"ElementValue"`
				} `xml:"Symbol"`
				Name struct {
					Text         string `xml:",chardata"`
					ElementValue struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"ElementValue"`
				} `xml:"Name"`
			} `xml:"MeasureRelationship"`
			XRef []struct {
				Text string `xml:",chardata"`
				DB   string `xml:"DB,attr"`
				ID   string `xml:"ID,attr"`
				Type string `xml:"Type,attr"`
			} `xml:"XRef"`
			CytogeneticLocation string `xml:"CytogeneticLocation"`
			SequenceLocation    []struct {
				Text            string `xml:",chardata"`
				Assembly        string `xml:"Assembly,attr"`
				Chr             string `xml:"Chr,attr"`
				AlternateAllele string `xml:"alternateAllele,attr"`
				ReferenceAllele string `xml:"referenceAllele,attr"`
				Start           string `xml:"start,attr"`
				Stop            string `xml:"stop,attr"`
				VariantLength   string `xml:"variantLength,attr"`
				InnerStart      string `xml:"innerStart,attr"`
				InnerStop       string `xml:"innerStop,attr"`
				OuterStart      string `xml:"outerStart,attr"`
				OuterStop       string `xml:"outerStop,attr"`
				Accession       string `xml:"Accession,attr"`
				Strand          string `xml:"Strand,attr"`
			} `xml:"SequenceLocation"`
			Citation []struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
				ID   struct {
					Text   string `xml:",chardata"`
					Source string `xml:"Source,attr"`
				} `xml:"ID"`
				URL          string `xml:"URL"`
				CitationText string `xml:"CitationText"`
			} `xml:"Citation"`
			Comment struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
			} `xml:"Comment"`
		} `xml:"Measure"`
		Comment struct {
			Text string `xml:",chardata"`
			Type string `xml:"Type,attr"`
		} `xml:"Comment"`
		AttributeSet []struct {
			Text      string `xml:",chardata"`
			Attribute struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
			} `xml:"Attribute"`
		} `xml:"AttributeSet"`
		Name []struct {
			Text         string `xml:",chardata"`
			ElementValue struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
			} `xml:"ElementValue"`
			XRef struct {
				Text   string `xml:",chardata"`
				DB     string `xml:"DB,attr"`
				ID     string `xml:"ID,attr"`
				Status string `xml:"Status,attr"`
				Type   string `xml:"Type,attr"`
				URL    string `xml:"URL,attr"`
			} `xml:"XRef"`
		} `xml:"Name"`
		XRef struct {
			Text string `xml:",chardata"`
			DB   string `xml:"DB,attr"`
			ID   string `xml:"ID,attr"`
			Type string `xml:"Type,attr"`
		} `xml:"XRef"`
		Citation struct {
			Text string `xml:",chardata"`
			Type string `xml:"Type,attr"`
			ID   struct {
				Text   string `xml:",chardata"`
				Source string `xml:"Source,attr"`
			} `xml:"ID"`
		} `xml:"Citation"`
	} `xml:"MeasureSet"`
	TraitSet struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"Type,attr"`
		Trait []struct {
			Text string `xml:",chardata"`
			Type string `xml:"Type,attr"`
			Name []struct {
				Text         string `xml:",chardata"`
				ElementValue struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"ElementValue"`
				XRef struct {
					Text string `xml:",chardata"`
					DB   string `xml:"DB,attr"`
					ID   string `xml:"ID,attr"`
				} `xml:"XRef"`
			} `xml:"Name"`
			XRef []struct {
				Text   string `xml:",chardata"`
				DB     string `xml:"DB,attr"`
				ID     string `xml:"ID,attr"`
				Type   string `xml:"Type,attr"`
				Status string `xml:"Status,attr"`
			} `xml:"XRef"`
			TraitRelationship []struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
				Name struct {
					Text         string `xml:",chardata"`
					ElementValue struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"ElementValue"`
				} `xml:"Name"`
				XRef struct {
					Text string `xml:",chardata"`
					DB   string `xml:"DB,attr"`
					ID   string `xml:"ID,attr"`
				} `xml:"XRef"`
			} `xml:"TraitRelationship"`
			Symbol struct {
				Text         string `xml:",chardata"`
				ElementValue struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"ElementValue"`
			} `xml:"Symbol"`
		} `xml:"Trait"`
		Comment struct {
			Text string `xml:",chardata"`
			Type string `xml:"Type,attr"`
		} `xml:"Comment"`
	} `xml:"TraitSet"`
	AdditionalSubmitters struct {
		Text                 string `xml:",chardata"`
		SubmitterDescription []struct {
			Text          string `xml:",chardata"`
			OrgID         string `xml:"OrgID,attr"`
			SubmitterName string `xml:"SubmitterName,attr"`
			Type          string `xml:"Type,attr"`
			Category      string `xml:"category,attr"`
		} `xml:"SubmitterDescription"`
	} `xml:"AdditionalSubmitters"`
	AttributeSet []struct {
		Text      string `xml:",chardata"`
		Attribute struct {
			Text         string `xml:",chardata"`
			Type         string `xml:"Type,attr"`
			DateValue    string `xml:"dateValue,attr"`
			IntegerValue string `xml:"integerValue,attr"`
		} `xml:"Attribute"`
		Citation []struct {
			Text   string `xml:",chardata"`
			Type   string `xml:"Type,attr"`
			Abbrev string `xml:"Abbrev,attr"`
			ID     struct {
				Text   string `xml:",chardata"`
				Source string `xml:"Source,attr"`
			} `xml:"ID"`
			URL          string `xml:"URL"`
			CitationText string `xml:"CitationText"`
		} `xml:"Citation"`
	} `xml:"AttributeSet"`
	Citation []struct {
		Text   string `xml:",chardata"`
		Type   string `xml:"Type,attr"`
		Abbrev string `xml:"Abbrev,attr"`
		ID     []struct {
			Text   string `xml:",chardata"`
			Source string `xml:"Source,attr"`
		} `xml:"ID"`
		URL          string `xml:"URL"`
		CitationText string `xml:"CitationText"`
	} `xml:"Citation"`
	StudyDescription string `xml:"StudyDescription"`
	ReplacedList     struct {
		Text     string `xml:",chardata"`
		Replaced []struct {
			Text        string `xml:",chardata"`
			Accession   string `xml:"Accession,attr"`
			DateChanged string `xml:"DateChanged,attr"`
			Version     string `xml:"Version,attr"`
			Comment     struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
			} `xml:"Comment"`
		} `xml:"Replaced"`
	} `xml:"ReplacedList"`
	StudyName string `xml:"StudyName"`
	Comment   []struct {
		Text       string `xml:",chardata"`
		Type       string `xml:"Type,attr"`
		DataSource string `xml:"DataSource,attr"`
	} `xml:"Comment"`
	CustomAssertionScore []struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"Type,attr"`
		Value string `xml:"Value,attr"`
	} `xml:"CustomAssertionScore"`
	GenotypeSet struct {
		Text       string `xml:",chardata"`
		Type       string `xml:"Type,attr"`
		MeasureSet []struct {
			Text                string `xml:",chardata"`
			Type                string `xml:"Type,attr"`
			NumberOfChromosomes string `xml:"NumberOfChromosomes,attr"`
			Measure             []struct {
				Text         string `xml:",chardata"`
				Type         string `xml:"Type,attr"`
				AttributeSet []struct {
					Text      string `xml:",chardata"`
					Attribute struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"Attribute"`
					XRef struct {
						Text string `xml:",chardata"`
						DB   string `xml:"DB,attr"`
						ID   string `xml:"ID,attr"`
						URL  string `xml:"URL,attr"`
					} `xml:"XRef"`
				} `xml:"AttributeSet"`
				MeasureRelationship struct {
					Text   string `xml:",chardata"`
					Type   string `xml:"Type,attr"`
					Symbol struct {
						Text         string `xml:",chardata"`
						ElementValue struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
						} `xml:"ElementValue"`
					} `xml:"Symbol"`
					Name struct {
						Text         string `xml:",chardata"`
						ElementValue struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
						} `xml:"ElementValue"`
					} `xml:"Name"`
				} `xml:"MeasureRelationship"`
				Comment string `xml:"Comment"`
				Name    struct {
					Text         string `xml:",chardata"`
					ElementValue struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"ElementValue"`
				} `xml:"Name"`
				SequenceLocation struct {
					Text            string `xml:",chardata"`
					Assembly        string `xml:"Assembly,attr"`
					Chr             string `xml:"Chr,attr"`
					Start           string `xml:"start,attr"`
					Stop            string `xml:"stop,attr"`
					VariantLength   string `xml:"variantLength,attr"`
					AlternateAllele string `xml:"alternateAllele,attr"`
					ReferenceAllele string `xml:"referenceAllele,attr"`
					InnerStart      string `xml:"innerStart,attr"`
					InnerStop       string `xml:"innerStop,attr"`
					OuterStart      string `xml:"outerStart,attr"`
				} `xml:"SequenceLocation"`
				XRef []struct {
					Text string `xml:",chardata"`
					DB   string `xml:"DB,attr"`
					ID   string `xml:"ID,attr"`
					Type string `xml:"Type,attr"`
				} `xml:"XRef"`
			} `xml:"Measure"`
			AttributeSet struct {
				Text      string `xml:",chardata"`
				Attribute struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"Attribute"`
			} `xml:"AttributeSet"`
			Name struct {
				Text         string `xml:",chardata"`
				ElementValue struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
				} `xml:"ElementValue"`
				XRef struct {
					Text   string `xml:",chardata"`
					DB     string `xml:"DB,attr"`
					ID     string `xml:"ID,attr"`
					Status string `xml:"Status,attr"`
					Type   string `xml:"Type,attr"`
					URL    string `xml:"URL,attr"`
				} `xml:"XRef"`
			} `xml:"Name"`
		} `xml:"MeasureSet"`
		AttributeSet struct {
			Text      string `xml:",chardata"`
			Attribute struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
			} `xml:"Attribute"`
		} `xml:"AttributeSet"`
		Name []struct {
			Text         string `xml:",chardata"`
			ElementValue struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
			} `xml:"ElementValue"`
		} `xml:"Name"`
	} `xml:"GenotypeSet"`
}

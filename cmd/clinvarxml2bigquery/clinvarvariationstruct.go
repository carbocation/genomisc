package main

import "encoding/xml"

// ClinVarVariationRelease was generated 2020-05-04 15:51:46 by james with zek
type ClinVarVariationRelease struct {
	XMLName                   xml.Name           `xml:"ClinVarVariationRelease"`
	Text                      string             `xml:",chardata"`
	Xsi                       string             `xml:"xsi,attr"`
	NoNamespaceSchemaLocation string             `xml:"noNamespaceSchemaLocation,attr"`
	ReleaseDate               string             `xml:"ReleaseDate,attr"`
	VariationArchive          []VariationArchive `xml:"VariationArchive"`
}

type VariationArchive struct {
	Text                string `xml:",chardata"`
	VariationID         string `xml:"VariationID,attr"`
	VariationName       string `xml:"VariationName,attr"`
	VariationType       string `xml:"VariationType,attr"`
	DateCreated         string `xml:"DateCreated,attr"`
	DateLastUpdated     string `xml:"DateLastUpdated,attr"`
	Accession           string `xml:"Accession,attr"`
	Version             string `xml:"Version,attr"`
	RecordType          string `xml:"RecordType,attr"`
	NumberOfSubmissions string `xml:"NumberOfSubmissions,attr"`
	NumberOfSubmitters  string `xml:"NumberOfSubmitters,attr"`
	RecordStatus        string `xml:"RecordStatus"`
	Species             string `xml:"Species"`
	InterpretedRecord   struct {
		Text         string `xml:",chardata"`
		SimpleAllele struct {
			Text        string `xml:",chardata"`
			AlleleID    string `xml:"AlleleID,attr"`
			VariationID string `xml:"VariationID,attr"`
			GeneList    struct {
				Text string `xml:",chardata"`
				Gene []struct {
					Text             string `xml:",chardata"`
					Symbol           string `xml:"Symbol,attr"`
					FullName         string `xml:"FullName,attr"`
					GeneID           string `xml:"GeneID,attr"`
					HGNCID           string `xml:"HGNC_ID,attr"`
					Source           string `xml:"Source,attr"`
					RelationshipType string `xml:"RelationshipType,attr"`
					Location         struct {
						Text                string   `xml:",chardata"`
						CytogeneticLocation []string `xml:"CytogeneticLocation"`
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
							Strand                   string `xml:"Strand,attr"`
						} `xml:"SequenceLocation"`
					} `xml:"Location"`
					OMIM               []string `xml:"OMIM"`
					Haploinsufficiency struct {
						Text          string `xml:",chardata"`
						LastEvaluated string `xml:"last_evaluated,attr"`
						ClinGen       string `xml:"ClinGen,attr"`
					} `xml:"Haploinsufficiency"`
					Triplosensitivity struct {
						Text          string `xml:",chardata"`
						LastEvaluated string `xml:"last_evaluated,attr"`
						ClinGen       string `xml:"ClinGen,attr"`
					} `xml:"Triplosensitivity"`
					Property string `xml:"Property"`
				} `xml:"Gene"`
			} `xml:"GeneList"`
			Name        string `xml:"Name"`
			VariantType string `xml:"VariantType"`
			Location    struct {
				Text                string   `xml:",chardata"`
				CytogeneticLocation []string `xml:"CytogeneticLocation"`
				SequenceLocation    []struct {
					Text                     string `xml:",chardata"`
					Assembly                 string `xml:"Assembly,attr"`
					AssemblyAccessionVersion string `xml:"AssemblyAccessionVersion,attr"`
					ForDisplay               string `xml:"forDisplay,attr"`
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
			} `xml:"Location"`
			OtherNameList struct {
				Text string   `xml:",chardata"`
				Name []string `xml:"Name"`
			} `xml:"OtherNameList"`
			XRefList struct {
				Text string `xml:",chardata"`
				XRef []struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
				} `xml:"XRef"`
			} `xml:"XRefList"`
			ProteinChange []string `xml:"ProteinChange"`
			Comment       []struct {
				Text       string `xml:",chardata"`
				DataSource string `xml:"DataSource,attr"`
				Type       string `xml:"Type,attr"`
			} `xml:"Comment"`
			HGVSlist struct {
				Text string `xml:",chardata"`
				HGVS []struct {
					Text                 string `xml:",chardata"`
					Assembly             string `xml:"Assembly,attr"`
					Type                 string `xml:"Type,attr"`
					NucleotideExpression struct {
						Text                     string `xml:",chardata"`
						SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
						SequenceAccession        string `xml:"sequenceAccession,attr"`
						SequenceVersion          string `xml:"sequenceVersion,attr"`
						Change                   string `xml:"change,attr"`
						Assembly                 string `xml:"Assembly,attr"`
						Expression               string `xml:"Expression"`
					} `xml:"NucleotideExpression"`
					ProteinExpression struct {
						Text                     string `xml:",chardata"`
						SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
						SequenceAccession        string `xml:"sequenceAccession,attr"`
						SequenceVersion          string `xml:"sequenceVersion,attr"`
						Change                   string `xml:"change,attr"`
						Expression               string `xml:"Expression"`
					} `xml:"ProteinExpression"`
					MolecularConsequence []struct {
						Text string `xml:",chardata"`
						ID   string `xml:"ID,attr"`
						Type string `xml:"Type,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"MolecularConsequence"`
				} `xml:"HGVS"`
			} `xml:"HGVSlist"`
			FunctionalConsequence []struct {
				Text  string `xml:",chardata"`
				Value string `xml:"Value,attr"`
				XRef  struct {
					Text string `xml:",chardata"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
				} `xml:"XRef"`
				Comment struct {
					Text       string `xml:",chardata"`
					DataSource string `xml:"DataSource,attr"`
					Type       string `xml:"Type,attr"`
				} `xml:"Comment"`
				Citation []struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
					ID   struct {
						Text   string `xml:",chardata"`
						Source string `xml:"Source,attr"`
					} `xml:"ID"`
				} `xml:"Citation"`
			} `xml:"FunctionalConsequence"`
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
		} `xml:"SimpleAllele"`
		ReviewStatus string `xml:"ReviewStatus"`
		RCVList      struct {
			Text         string `xml:",chardata"`
			RCVAccession []struct {
				Text                     string `xml:",chardata"`
				Title                    string `xml:"Title,attr"`
				DateLastEvaluated        string `xml:"DateLastEvaluated,attr"`
				ReviewStatus             string `xml:"ReviewStatus,attr"`
				Interpretation           string `xml:"Interpretation,attr"`
				SubmissionCount          string `xml:"SubmissionCount,attr"`
				Accession                string `xml:"Accession,attr"`
				Version                  string `xml:"Version,attr"`
				InterpretedConditionList struct {
					Text                 string `xml:",chardata"`
					InterpretedCondition []struct {
						Text string `xml:",chardata"`
						DB   string `xml:"DB,attr"`
						ID   string `xml:"ID,attr"`
					} `xml:"InterpretedCondition"`
				} `xml:"InterpretedConditionList"`
			} `xml:"RCVAccession"`
		} `xml:"RCVList"`
		Interpretations struct {
			Text           string `xml:",chardata"`
			Interpretation struct {
				Text                string `xml:",chardata"`
				DateLastEvaluated   string `xml:"DateLastEvaluated,attr"`
				NumberOfSubmissions string `xml:"NumberOfSubmissions,attr"`
				NumberOfSubmitters  string `xml:"NumberOfSubmitters,attr"`
				Type                string `xml:"Type,attr"`
				Description         string `xml:"Description"`
				Citation            []struct {
					Text         string `xml:",chardata"`
					Type         string `xml:"Type,attr"`
					Abbrev       string `xml:"Abbrev,attr"`
					CitationText string `xml:"CitationText"`
					ID           []struct {
						Text   string `xml:",chardata"`
						Source string `xml:"Source,attr"`
					} `xml:"ID"`
					URL string `xml:"URL"`
				} `xml:"Citation"`
				ConditionList struct {
					Text     string `xml:",chardata"`
					TraitSet []struct {
						Text  string `xml:",chardata"`
						ID    string `xml:"ID,attr"`
						Type  string `xml:"Type,attr"`
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
									Type string `xml:"Type,attr"`
									ID   string `xml:"ID,attr"`
									DB   string `xml:"DB,attr"`
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
				} `xml:"ConditionList"`
				DescriptionHistory []struct {
					Text        string `xml:",chardata"`
					Dated       string `xml:"Dated,attr"`
					Description string `xml:"Description"`
				} `xml:"DescriptionHistory"`
				Explanation struct {
					Text       string `xml:",chardata"`
					DataSource string `xml:"DataSource,attr"`
					Type       string `xml:"Type,attr"`
				} `xml:"Explanation"`
				Comment struct {
					Text       string `xml:",chardata"`
					DataSource string `xml:"DataSource,attr"`
					Type       string `xml:"Type,attr"`
				} `xml:"Comment"`
			} `xml:"Interpretation"`
		} `xml:"Interpretations"`
		ClinicalAssertionList struct {
			Text              string `xml:",chardata"`
			ClinicalAssertion []struct {
				Text                  string `xml:",chardata"`
				ID                    string `xml:"ID,attr"`
				DateCreated           string `xml:"DateCreated,attr"`
				DateLastUpdated       string `xml:"DateLastUpdated,attr"`
				SubmissionDate        string `xml:"SubmissionDate,attr"`
				FDARecognizedDatabase string `xml:"FDARecognizedDatabase,attr"`
				ClinVarSubmissionID   struct {
					Text                string `xml:",chardata"`
					LocalKey            string `xml:"localKey,attr"`
					Title               string `xml:"title,attr"`
					SubmittedAssembly   string `xml:"submittedAssembly,attr"`
					LocalKeyIsSubmitted string `xml:"localKeyIsSubmitted,attr"`
				} `xml:"ClinVarSubmissionID"`
				ClinVarAccession struct {
					Text                 string `xml:",chardata"`
					Accession            string `xml:"Accession,attr"`
					Type                 string `xml:"Type,attr"`
					Version              string `xml:"Version,attr"`
					SubmitterName        string `xml:"SubmitterName,attr"`
					OrgID                string `xml:"OrgID,attr"`
					OrganizationCategory string `xml:"OrganizationCategory,attr"`
					OrgAbbreviation      string `xml:"OrgAbbreviation,attr"`
				} `xml:"ClinVarAccession"`
				RecordStatus   string `xml:"RecordStatus"`
				ReviewStatus   string `xml:"ReviewStatus"`
				Interpretation struct {
					Text              string `xml:",chardata"`
					DateLastEvaluated string `xml:"DateLastEvaluated,attr"`
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
				} `xml:"Interpretation"`
				Assertion      string `xml:"Assertion"`
				ObservedInList struct {
					Text       string `xml:",chardata"`
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
							NumberTested     string `xml:"NumberTested"`
							GeographicOrigin string `xml:"GeographicOrigin"`
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
							Proband           string `xml:"Proband"`
							SampleDescription struct {
								Text        string `xml:",chardata"`
								Description struct {
									Text string `xml:",chardata"`
									Type string `xml:"Type,attr"`
								} `xml:"Description"`
							} `xml:"SampleDescription"`
							CellLine string `xml:"CellLine"`
							Strain   string `xml:"Strain"`
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
							Description        string `xml:"Description"`
							NamePlatform       string `xml:"NamePlatform"`
							Purpose            string `xml:"Purpose"`
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
							SourceType string `xml:"SourceType"`
							Software   []struct {
								Text    string `xml:",chardata"`
								Name    string `xml:"name,attr"`
								Purpose string `xml:"purpose,attr"`
								Version string `xml:"version,attr"`
							} `xml:"Software"`
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
				} `xml:"ObservedInList"`
				SimpleAllele struct {
					Text     string `xml:",chardata"`
					GeneList struct {
						Text string `xml:",chardata"`
						Gene []struct {
							Text   string `xml:",chardata"`
							Symbol string `xml:"Symbol,attr"`
							Name   string `xml:"Name"`
						} `xml:"Gene"`
					} `xml:"GeneList"`
					Name          string `xml:"Name"`
					VariantType   string `xml:"VariantType"`
					OtherNameList struct {
						Text string `xml:",chardata"`
						Name []struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
						} `xml:"Name"`
					} `xml:"OtherNameList"`
					XRefList struct {
						Text string `xml:",chardata"`
						XRef []struct {
							Text string `xml:",chardata"`
							DB   string `xml:"DB,attr"`
							ID   string `xml:"ID,attr"`
							Type string `xml:"Type,attr"`
						} `xml:"XRef"`
					} `xml:"XRefList"`
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
						} `xml:"XRef"`
					} `xml:"AttributeSet"`
					Location struct {
						Text                string `xml:",chardata"`
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
						GeneLocation []string `xml:"GeneLocation"`
					} `xml:"Location"`
					CitationList struct {
						Text     string `xml:",chardata"`
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
					} `xml:"CitationList"`
					FunctionalConsequence []struct {
						Text  string `xml:",chardata"`
						Value string `xml:"Value,attr"`
						XRef  struct {
							Text string `xml:",chardata"`
							DB   string `xml:"DB,attr"`
							ID   string `xml:"ID,attr"`
							URL  string `xml:"URL,attr"`
						} `xml:"XRef"`
						Comment struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
						} `xml:"Comment"`
					} `xml:"FunctionalConsequence"`
					ProteinChange            string `xml:"ProteinChange"`
					MolecularConsequenceList struct {
						Text                 string `xml:",chardata"`
						MolecularConsequence struct {
							Text     string `xml:",chardata"`
							Function string `xml:"Function,attr"`
							Comment  string `xml:"Comment"`
							XRef     struct {
								Text string `xml:",chardata"`
								DB   string `xml:"DB,attr"`
								ID   string `xml:"ID,attr"`
							} `xml:"XRef"`
						} `xml:"MolecularConsequence"`
					} `xml:"MolecularConsequenceList"`
					Comment struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
					} `xml:"Comment"`
				} `xml:"SimpleAllele"`
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
				SubmissionNameList struct {
					Text           string   `xml:",chardata"`
					SubmissionName []string `xml:"SubmissionName"`
				} `xml:"SubmissionNameList"`
				AdditionalSubmitters struct {
					Text                 string `xml:",chardata"`
					SubmitterDescription []struct {
						Text                 string `xml:",chardata"`
						OrgID                string `xml:"OrgID,attr"`
						SubmitterName        string `xml:"SubmitterName,attr"`
						Type                 string `xml:"Type,attr"`
						OrganizationCategory string `xml:"OrganizationCategory,attr"`
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
				Haplotype struct {
					Text         string `xml:",chardata"`
					SimpleAllele []struct {
						Text     string `xml:",chardata"`
						GeneList struct {
							Text string `xml:",chardata"`
							Gene struct {
								Text   string `xml:",chardata"`
								Symbol string `xml:"Symbol,attr"`
								Name   string `xml:"Name"`
							} `xml:"Gene"`
						} `xml:"GeneList"`
						VariantType string `xml:"VariantType"`
						Location    struct {
							Text             string `xml:",chardata"`
							GeneLocation     string `xml:"GeneLocation"`
							SequenceLocation struct {
								Text            string `xml:",chardata"`
								Assembly        string `xml:"Assembly,attr"`
								Chr             string `xml:"Chr,attr"`
								Start           string `xml:"start,attr"`
								Stop            string `xml:"stop,attr"`
								AlternateAllele string `xml:"alternateAllele,attr"`
								ReferenceAllele string `xml:"referenceAllele,attr"`
								VariantLength   string `xml:"variantLength,attr"`
							} `xml:"SequenceLocation"`
						} `xml:"Location"`
						ProteinChange string `xml:"ProteinChange"`
						CitationList  struct {
							Text     string `xml:",chardata"`
							Citation struct {
								Text string `xml:",chardata"`
								Type string `xml:"Type,attr"`
								ID   struct {
									Text   string `xml:",chardata"`
									Source string `xml:"Source,attr"`
								} `xml:"ID"`
								URL string `xml:"URL"`
							} `xml:"Citation"`
						} `xml:"CitationList"`
						MolecularConsequenceList struct {
							Text                 string `xml:",chardata"`
							MolecularConsequence struct {
								Text     string `xml:",chardata"`
								Function string `xml:"Function,attr"`
								Comment  string `xml:"Comment"`
							} `xml:"MolecularConsequence"`
						} `xml:"MolecularConsequenceList"`
						AttributeSet []struct {
							Text      string `xml:",chardata"`
							Attribute struct {
								Text string `xml:",chardata"`
								Type string `xml:"Type,attr"`
							} `xml:"Attribute"`
						} `xml:"AttributeSet"`
						Name          string `xml:"Name"`
						OtherNameList struct {
							Text string `xml:",chardata"`
							Name []struct {
								Text string `xml:",chardata"`
								Type string `xml:"Type,attr"`
							} `xml:"Name"`
						} `xml:"OtherNameList"`
						XRefList struct {
							Text string `xml:",chardata"`
							XRef struct {
								Text string `xml:",chardata"`
								DB   string `xml:"DB,attr"`
								ID   string `xml:"ID,attr"`
								Type string `xml:"Type,attr"`
							} `xml:"XRef"`
						} `xml:"XRefList"`
						FunctionalConsequence struct {
							Text  string `xml:",chardata"`
							Value string `xml:"Value,attr"`
						} `xml:"FunctionalConsequence"`
						Comment string `xml:"Comment"`
					} `xml:"SimpleAllele"`
					AttributeSet []struct {
						Text      string `xml:",chardata"`
						Attribute struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
						} `xml:"Attribute"`
					} `xml:"AttributeSet"`
					Name          string `xml:"Name"`
					OtherNameList struct {
						Text string   `xml:",chardata"`
						Name []string `xml:"Name"`
					} `xml:"OtherNameList"`
					XRefList struct {
						Text string `xml:",chardata"`
						XRef struct {
							Text string `xml:",chardata"`
							DB   string `xml:"DB,attr"`
							ID   string `xml:"ID,attr"`
							Type string `xml:"Type,attr"`
						} `xml:"XRef"`
					} `xml:"XRefList"`
					CitationList struct {
						Text     string `xml:",chardata"`
						Citation struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
							ID   struct {
								Text   string `xml:",chardata"`
								Source string `xml:"Source,attr"`
							} `xml:"ID"`
						} `xml:"Citation"`
					} `xml:"CitationList"`
					FunctionalConsequence struct {
						Text  string `xml:",chardata"`
						Value string `xml:"Value,attr"`
					} `xml:"FunctionalConsequence"`
				} `xml:"Haplotype"`
				CustomAssertionScore []struct {
					Text  string `xml:",chardata"`
					Type  string `xml:"Type,attr"`
					Value string `xml:"Value,attr"`
				} `xml:"CustomAssertionScore"`
				Genotype struct {
					Text         string `xml:",chardata"`
					SimpleAllele []struct {
						Text     string `xml:",chardata"`
						GeneList struct {
							Text string `xml:",chardata"`
							Gene struct {
								Text   string `xml:",chardata"`
								Symbol string `xml:"Symbol,attr"`
								Name   string `xml:"Name"`
							} `xml:"Gene"`
						} `xml:"GeneList"`
						VariantType  string `xml:"VariantType"`
						Comment      string `xml:"Comment"`
						AttributeSet []struct {
							Text      string `xml:",chardata"`
							Attribute struct {
								Text string `xml:",chardata"`
								Type string `xml:"Type,attr"`
							} `xml:"Attribute"`
						} `xml:"AttributeSet"`
						OtherNameList struct {
							Text string `xml:",chardata"`
							Name []struct {
								Text string `xml:",chardata"`
								Type string `xml:"Type,attr"`
							} `xml:"Name"`
						} `xml:"OtherNameList"`
						Location struct {
							Text             string `xml:",chardata"`
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
							GeneLocation string `xml:"GeneLocation"`
						} `xml:"Location"`
						FunctionalConsequence struct {
							Text  string `xml:",chardata"`
							Value string `xml:"Value,attr"`
							XRef  struct {
								Text string `xml:",chardata"`
								DB   string `xml:"DB,attr"`
								ID   string `xml:"ID,attr"`
								URL  string `xml:"URL,attr"`
							} `xml:"XRef"`
						} `xml:"FunctionalConsequence"`
						XRefList struct {
							Text string `xml:",chardata"`
							XRef []struct {
								Text string `xml:",chardata"`
								DB   string `xml:"DB,attr"`
								ID   string `xml:"ID,attr"`
								Type string `xml:"Type,attr"`
							} `xml:"XRef"`
						} `xml:"XRefList"`
					} `xml:"SimpleAllele"`
					VariationType         string `xml:"VariationType"`
					FunctionalConsequence struct {
						Text  string `xml:",chardata"`
						Value string `xml:"Value,attr"`
					} `xml:"FunctionalConsequence"`
					Haplotype []struct {
						Text                string `xml:",chardata"`
						NumberOfChromosomes string `xml:"NumberOfChromosomes,attr"`
						SimpleAllele        []struct {
							Text     string `xml:",chardata"`
							GeneList struct {
								Text string `xml:",chardata"`
								Gene struct {
									Text   string `xml:",chardata"`
									Symbol string `xml:"Symbol,attr"`
									Name   string `xml:"Name"`
								} `xml:"Gene"`
							} `xml:"GeneList"`
							VariantType  string `xml:"VariantType"`
							AttributeSet struct {
								Text      string `xml:",chardata"`
								Attribute struct {
									Text string `xml:",chardata"`
									Type string `xml:"Type,attr"`
								} `xml:"Attribute"`
							} `xml:"AttributeSet"`
							FunctionalConsequence struct {
								Text  string `xml:",chardata"`
								Value string `xml:"Value,attr"`
							} `xml:"FunctionalConsequence"`
							Name string `xml:"Name"`
						} `xml:"SimpleAllele"`
						AttributeSet struct {
							Text      string `xml:",chardata"`
							Attribute struct {
								Text string `xml:",chardata"`
								Type string `xml:"Type,attr"`
							} `xml:"Attribute"`
						} `xml:"AttributeSet"`
						FunctionalConsequence struct {
							Text  string `xml:",chardata"`
							Value string `xml:"Value,attr"`
						} `xml:"FunctionalConsequence"`
						Name string `xml:"Name"`
					} `xml:"Haplotype"`
					Name string `xml:"Name"`
				} `xml:"Genotype"`
			} `xml:"ClinicalAssertion"`
		} `xml:"ClinicalAssertionList"`
		TraitMappingList struct {
			Text         string `xml:",chardata"`
			TraitMapping []struct {
				Text                string `xml:",chardata"`
				ClinicalAssertionID string `xml:"ClinicalAssertionID,attr"`
				TraitType           string `xml:"TraitType,attr"`
				MappingType         string `xml:"MappingType,attr"`
				MappingValue        string `xml:"MappingValue,attr"`
				MappingRef          string `xml:"MappingRef,attr"`
				MedGen              struct {
					Text string `xml:",chardata"`
					CUI  string `xml:"CUI,attr"`
					Name string `xml:"Name,attr"`
				} `xml:"MedGen"`
			} `xml:"TraitMapping"`
		} `xml:"TraitMappingList"`
		DeletedSCVList struct {
			Text string `xml:",chardata"`
			SCV  []struct {
				Text      string `xml:",chardata"`
				Accession struct {
					Text        string `xml:",chardata"`
					Version     string `xml:"Version,attr"`
					DateDeleted string `xml:"DateDeleted,attr"`
				} `xml:"Accession"`
				Description string `xml:"Description"`
			} `xml:"SCV"`
		} `xml:"DeletedSCVList"`
		Haplotype struct {
			Text                string `xml:",chardata"`
			VariationID         string `xml:"VariationID,attr"`
			NumberOfChromosomes string `xml:"NumberOfChromosomes,attr"`
			SimpleAllele        []struct {
				Text        string `xml:",chardata"`
				AlleleID    string `xml:"AlleleID,attr"`
				VariationID string `xml:"VariationID,attr"`
				GeneList    struct {
					Text string `xml:",chardata"`
					Gene []struct {
						Text             string `xml:",chardata"`
						Symbol           string `xml:"Symbol,attr"`
						FullName         string `xml:"FullName,attr"`
						GeneID           string `xml:"GeneID,attr"`
						HGNCID           string `xml:"HGNC_ID,attr"`
						Source           string `xml:"Source,attr"`
						RelationshipType string `xml:"RelationshipType,attr"`
						Location         struct {
							Text                string `xml:",chardata"`
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
								Strand                   string `xml:"Strand,attr"`
							} `xml:"SequenceLocation"`
						} `xml:"Location"`
						OMIM               string `xml:"OMIM"`
						Haploinsufficiency struct {
							Text          string `xml:",chardata"`
							LastEvaluated string `xml:"last_evaluated,attr"`
							ClinGen       string `xml:"ClinGen,attr"`
						} `xml:"Haploinsufficiency"`
						Triplosensitivity struct {
							Text          string `xml:",chardata"`
							LastEvaluated string `xml:"last_evaluated,attr"`
							ClinGen       string `xml:"ClinGen,attr"`
						} `xml:"Triplosensitivity"`
						Property string `xml:"Property"`
					} `xml:"Gene"`
				} `xml:"GeneList"`
				Name          string `xml:"Name"`
				CanonicalSPDI string `xml:"CanonicalSPDI"`
				VariantType   string `xml:"VariantType"`
				Location      struct {
					Text                string `xml:",chardata"`
					CytogeneticLocation string `xml:"CytogeneticLocation"`
					SequenceLocation    []struct {
						Text                     string `xml:",chardata"`
						Assembly                 string `xml:"Assembly,attr"`
						AssemblyAccessionVersion string `xml:"AssemblyAccessionVersion,attr"`
						ForDisplay               string `xml:"forDisplay,attr"`
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
						Strand                   string `xml:"Strand,attr"`
					} `xml:"SequenceLocation"`
				} `xml:"Location"`
				ProteinChange []string `xml:"ProteinChange"`
				HGVSlist      struct {
					Text string `xml:",chardata"`
					HGVS []struct {
						Text                 string `xml:",chardata"`
						Assembly             string `xml:"Assembly,attr"`
						Type                 string `xml:"Type,attr"`
						NucleotideExpression struct {
							Text                     string `xml:",chardata"`
							SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
							SequenceAccession        string `xml:"sequenceAccession,attr"`
							SequenceVersion          string `xml:"sequenceVersion,attr"`
							Change                   string `xml:"change,attr"`
							Assembly                 string `xml:"Assembly,attr"`
							Expression               string `xml:"Expression"`
						} `xml:"NucleotideExpression"`
						ProteinExpression struct {
							Text                     string `xml:",chardata"`
							SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
							SequenceAccession        string `xml:"sequenceAccession,attr"`
							SequenceVersion          string `xml:"sequenceVersion,attr"`
							Change                   string `xml:"change,attr"`
							Expression               string `xml:"Expression"`
						} `xml:"ProteinExpression"`
						MolecularConsequence []struct {
							Text string `xml:",chardata"`
							ID   string `xml:"ID,attr"`
							Type string `xml:"Type,attr"`
							DB   string `xml:"DB,attr"`
						} `xml:"MolecularConsequence"`
					} `xml:"HGVS"`
				} `xml:"HGVSlist"`
				Interpretations struct {
					Text           string `xml:",chardata"`
					Interpretation struct {
						Text                string `xml:",chardata"`
						NumberOfSubmissions string `xml:"NumberOfSubmissions,attr"`
						NumberOfSubmitters  string `xml:"NumberOfSubmitters,attr"`
						Type                string `xml:"Type,attr"`
						DateLastEvaluated   string `xml:"DateLastEvaluated,attr"`
						Description         string `xml:"Description"`
						Citation            []struct {
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
						ConditionList struct {
							Text     string `xml:",chardata"`
							TraitSet []struct {
								Text  string `xml:",chardata"`
								ID    string `xml:"ID,attr"`
								Type  string `xml:"Type,attr"`
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
									} `xml:"Symbol"`
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
									Comment struct {
										Text       string `xml:",chardata"`
										DataSource string `xml:"DataSource,attr"`
										Type       string `xml:"Type,attr"`
									} `xml:"Comment"`
								} `xml:"Trait"`
							} `xml:"TraitSet"`
						} `xml:"ConditionList"`
						DescriptionHistory []struct {
							Text        string `xml:",chardata"`
							Dated       string `xml:"Dated,attr"`
							Description string `xml:"Description"`
						} `xml:"DescriptionHistory"`
						Explanation struct {
							Text       string `xml:",chardata"`
							DataSource string `xml:"DataSource,attr"`
							Type       string `xml:"Type,attr"`
						} `xml:"Explanation"`
						Comment struct {
							Text       string `xml:",chardata"`
							DataSource string `xml:"DataSource,attr"`
							Type       string `xml:"Type,attr"`
						} `xml:"Comment"`
					} `xml:"Interpretation"`
				} `xml:"Interpretations"`
				XRefList struct {
					Text string `xml:",chardata"`
					XRef []struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"XRef"`
				} `xml:"XRefList"`
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
				OtherNameList struct {
					Text string   `xml:",chardata"`
					Name []string `xml:"Name"`
				} `xml:"OtherNameList"`
				Comment struct {
					Text       string `xml:",chardata"`
					DataSource string `xml:"DataSource,attr"`
					Type       string `xml:"Type,attr"`
				} `xml:"Comment"`
				FunctionalConsequence []struct {
					Text  string `xml:",chardata"`
					Value string `xml:"Value,attr"`
					XRef  struct {
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
					Comment struct {
						Text       string `xml:",chardata"`
						DataSource string `xml:"DataSource,attr"`
						Type       string `xml:"Type,attr"`
					} `xml:"Comment"`
				} `xml:"FunctionalConsequence"`
			} `xml:"SimpleAllele"`
			Name          string `xml:"Name"`
			VariationType string `xml:"VariationType"`
			OtherNameList struct {
				Text string   `xml:",chardata"`
				Name []string `xml:"Name"`
			} `xml:"OtherNameList"`
			HGVSlist struct {
				Text string `xml:",chardata"`
				HGVS []struct {
					Text                 string `xml:",chardata"`
					Type                 string `xml:"Type,attr"`
					Assembly             string `xml:"Assembly,attr"`
					NucleotideExpression struct {
						Text       string `xml:",chardata"`
						Expression string `xml:"Expression"`
					} `xml:"NucleotideExpression"`
					ProteinExpression struct {
						Text       string `xml:",chardata"`
						Expression string `xml:"Expression"`
					} `xml:"ProteinExpression"`
				} `xml:"HGVS"`
			} `xml:"HGVSlist"`
			XRefList struct {
				Text string `xml:",chardata"`
				XRef []struct {
					Text string `xml:",chardata"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
					Type string `xml:"Type,attr"`
				} `xml:"XRef"`
			} `xml:"XRefList"`
			FunctionalConsequence struct {
				Text  string `xml:",chardata"`
				Value string `xml:"Value,attr"`
			} `xml:"FunctionalConsequence"`
		} `xml:"Haplotype"`
		Genotype struct {
			Text         string `xml:",chardata"`
			VariationID  string `xml:"VariationID,attr"`
			SimpleAllele []struct {
				Text        string `xml:",chardata"`
				AlleleID    string `xml:"AlleleID,attr"`
				VariationID string `xml:"VariationID,attr"`
				GeneList    struct {
					Text string `xml:",chardata"`
					Gene []struct {
						Text             string `xml:",chardata"`
						Symbol           string `xml:"Symbol,attr"`
						FullName         string `xml:"FullName,attr"`
						GeneID           string `xml:"GeneID,attr"`
						HGNCID           string `xml:"HGNC_ID,attr"`
						Source           string `xml:"Source,attr"`
						RelationshipType string `xml:"RelationshipType,attr"`
						Location         struct {
							Text                string `xml:",chardata"`
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
								Strand                   string `xml:"Strand,attr"`
							} `xml:"SequenceLocation"`
						} `xml:"Location"`
						OMIM               string `xml:"OMIM"`
						Haploinsufficiency struct {
							Text          string `xml:",chardata"`
							LastEvaluated string `xml:"last_evaluated,attr"`
							ClinGen       string `xml:"ClinGen,attr"`
						} `xml:"Haploinsufficiency"`
						Triplosensitivity struct {
							Text          string `xml:",chardata"`
							LastEvaluated string `xml:"last_evaluated,attr"`
							ClinGen       string `xml:"ClinGen,attr"`
						} `xml:"Triplosensitivity"`
						Property string `xml:"Property"`
					} `xml:"Gene"`
				} `xml:"GeneList"`
				Name          string `xml:"Name"`
				CanonicalSPDI string `xml:"CanonicalSPDI"`
				VariantType   string `xml:"VariantType"`
				Location      struct {
					Text                string `xml:",chardata"`
					CytogeneticLocation string `xml:"CytogeneticLocation"`
					SequenceLocation    []struct {
						Text                     string `xml:",chardata"`
						Assembly                 string `xml:"Assembly,attr"`
						AssemblyAccessionVersion string `xml:"AssemblyAccessionVersion,attr"`
						ForDisplay               string `xml:"forDisplay,attr"`
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
				} `xml:"Location"`
				ProteinChange []string `xml:"ProteinChange"`
				HGVSlist      struct {
					Text string `xml:",chardata"`
					HGVS []struct {
						Text                 string `xml:",chardata"`
						Assembly             string `xml:"Assembly,attr"`
						Type                 string `xml:"Type,attr"`
						NucleotideExpression struct {
							Text                     string `xml:",chardata"`
							SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
							SequenceAccession        string `xml:"sequenceAccession,attr"`
							SequenceVersion          string `xml:"sequenceVersion,attr"`
							Change                   string `xml:"change,attr"`
							Assembly                 string `xml:"Assembly,attr"`
							Expression               string `xml:"Expression"`
						} `xml:"NucleotideExpression"`
						ProteinExpression struct {
							Text                     string `xml:",chardata"`
							SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
							SequenceAccession        string `xml:"sequenceAccession,attr"`
							SequenceVersion          string `xml:"sequenceVersion,attr"`
							Change                   string `xml:"change,attr"`
							Expression               string `xml:"Expression"`
						} `xml:"ProteinExpression"`
						MolecularConsequence []struct {
							Text string `xml:",chardata"`
							ID   string `xml:"ID,attr"`
							Type string `xml:"Type,attr"`
							DB   string `xml:"DB,attr"`
						} `xml:"MolecularConsequence"`
					} `xml:"HGVS"`
				} `xml:"HGVSlist"`
				Interpretations struct {
					Text           string `xml:",chardata"`
					Interpretation struct {
						Text                string `xml:",chardata"`
						DateLastEvaluated   string `xml:"DateLastEvaluated,attr"`
						NumberOfSubmissions string `xml:"NumberOfSubmissions,attr"`
						NumberOfSubmitters  string `xml:"NumberOfSubmitters,attr"`
						Type                string `xml:"Type,attr"`
						Description         string `xml:"Description"`
						Citation            []struct {
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
						DescriptionHistory []struct {
							Text        string `xml:",chardata"`
							Dated       string `xml:"Dated,attr"`
							Description string `xml:"Description"`
						} `xml:"DescriptionHistory"`
						ConditionList struct {
							Text     string `xml:",chardata"`
							TraitSet []struct {
								Text  string `xml:",chardata"`
								ID    string `xml:"ID,attr"`
								Type  string `xml:"Type,attr"`
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
											Type string `xml:"Type,attr"`
											ID   string `xml:"ID,attr"`
											DB   string `xml:"DB,attr"`
										} `xml:"XRef"`
									} `xml:"Name"`
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
									} `xml:"Symbol"`
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
									Comment struct {
										Text       string `xml:",chardata"`
										DataSource string `xml:"DataSource,attr"`
										Type       string `xml:"Type,attr"`
									} `xml:"Comment"`
								} `xml:"Trait"`
							} `xml:"TraitSet"`
						} `xml:"ConditionList"`
						Explanation struct {
							Text       string `xml:",chardata"`
							DataSource string `xml:"DataSource,attr"`
							Type       string `xml:"Type,attr"`
						} `xml:"Explanation"`
						Comment struct {
							Text       string `xml:",chardata"`
							DataSource string `xml:"DataSource,attr"`
							Type       string `xml:"Type,attr"`
						} `xml:"Comment"`
					} `xml:"Interpretation"`
				} `xml:"Interpretations"`
				XRefList struct {
					Text string `xml:",chardata"`
					XRef []struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"XRef"`
				} `xml:"XRefList"`
				Comment struct {
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
				GlobalMinorAlleleFrequency struct {
					Text        string `xml:",chardata"`
					Value       string `xml:"Value,attr"`
					Source      string `xml:"Source,attr"`
					MinorAllele string `xml:"MinorAllele,attr"`
				} `xml:"GlobalMinorAlleleFrequency"`
				OtherNameList struct {
					Text string   `xml:",chardata"`
					Name []string `xml:"Name"`
				} `xml:"OtherNameList"`
				FunctionalConsequence []struct {
					Text  string `xml:",chardata"`
					Value string `xml:"Value,attr"`
					XRef  struct {
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
				} `xml:"FunctionalConsequence"`
			} `xml:"SimpleAllele"`
			Name          string `xml:"Name"`
			VariationType string `xml:"VariationType"`
			HGVSlist      struct {
				Text string `xml:",chardata"`
				HGVS []struct {
					Text                 string `xml:",chardata"`
					Type                 string `xml:"Type,attr"`
					NucleotideExpression struct {
						Text       string `xml:",chardata"`
						Expression string `xml:"Expression"`
					} `xml:"NucleotideExpression"`
				} `xml:"HGVS"`
			} `xml:"HGVSlist"`
			XRefList struct {
				Text string `xml:",chardata"`
				XRef struct {
					Text string `xml:",chardata"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
				} `xml:"XRef"`
			} `xml:"XRefList"`
			FunctionalConsequence struct {
				Text  string `xml:",chardata"`
				Value string `xml:"Value,attr"`
			} `xml:"FunctionalConsequence"`
			Haplotype []struct {
				Text                string `xml:",chardata"`
				VariationID         string `xml:"VariationID,attr"`
				NumberOfChromosomes string `xml:"NumberOfChromosomes,attr"`
				SimpleAllele        []struct {
					Text        string `xml:",chardata"`
					AlleleID    string `xml:"AlleleID,attr"`
					VariationID string `xml:"VariationID,attr"`
					GeneList    struct {
						Text string `xml:",chardata"`
						Gene []struct {
							Text             string `xml:",chardata"`
							Symbol           string `xml:"Symbol,attr"`
							FullName         string `xml:"FullName,attr"`
							GeneID           string `xml:"GeneID,attr"`
							HGNCID           string `xml:"HGNC_ID,attr"`
							Source           string `xml:"Source,attr"`
							RelationshipType string `xml:"RelationshipType,attr"`
							Location         struct {
								Text                string `xml:",chardata"`
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
									Strand                   string `xml:"Strand,attr"`
								} `xml:"SequenceLocation"`
							} `xml:"Location"`
							OMIM               string `xml:"OMIM"`
							Haploinsufficiency struct {
								Text          string `xml:",chardata"`
								LastEvaluated string `xml:"last_evaluated,attr"`
								ClinGen       string `xml:"ClinGen,attr"`
							} `xml:"Haploinsufficiency"`
							Triplosensitivity struct {
								Text          string `xml:",chardata"`
								LastEvaluated string `xml:"last_evaluated,attr"`
								ClinGen       string `xml:"ClinGen,attr"`
							} `xml:"Triplosensitivity"`
							Property string `xml:"Property"`
						} `xml:"Gene"`
					} `xml:"GeneList"`
					Name          string `xml:"Name"`
					CanonicalSPDI string `xml:"CanonicalSPDI"`
					VariantType   string `xml:"VariantType"`
					Location      struct {
						Text                string `xml:",chardata"`
						CytogeneticLocation string `xml:"CytogeneticLocation"`
						SequenceLocation    []struct {
							Text                     string `xml:",chardata"`
							Assembly                 string `xml:"Assembly,attr"`
							AssemblyAccessionVersion string `xml:"AssemblyAccessionVersion,attr"`
							ForDisplay               string `xml:"forDisplay,attr"`
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
						} `xml:"SequenceLocation"`
					} `xml:"Location"`
					ProteinChange []string `xml:"ProteinChange"`
					HGVSlist      struct {
						Text string `xml:",chardata"`
						HGVS []struct {
							Text                 string `xml:",chardata"`
							Assembly             string `xml:"Assembly,attr"`
							Type                 string `xml:"Type,attr"`
							NucleotideExpression struct {
								Text                     string `xml:",chardata"`
								SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
								SequenceAccession        string `xml:"sequenceAccession,attr"`
								SequenceVersion          string `xml:"sequenceVersion,attr"`
								Change                   string `xml:"change,attr"`
								Assembly                 string `xml:"Assembly,attr"`
								Expression               string `xml:"Expression"`
							} `xml:"NucleotideExpression"`
							ProteinExpression struct {
								Text                     string `xml:",chardata"`
								SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
								SequenceAccession        string `xml:"sequenceAccession,attr"`
								SequenceVersion          string `xml:"sequenceVersion,attr"`
								Change                   string `xml:"change,attr"`
								Expression               string `xml:"Expression"`
							} `xml:"ProteinExpression"`
							MolecularConsequence []struct {
								Text string `xml:",chardata"`
								ID   string `xml:"ID,attr"`
								Type string `xml:"Type,attr"`
								DB   string `xml:"DB,attr"`
							} `xml:"MolecularConsequence"`
						} `xml:"HGVS"`
					} `xml:"HGVSlist"`
					Interpretations struct {
						Text           string `xml:",chardata"`
						Interpretation struct {
							Text                string `xml:",chardata"`
							NumberOfSubmissions string `xml:"NumberOfSubmissions,attr"`
							NumberOfSubmitters  string `xml:"NumberOfSubmitters,attr"`
							Type                string `xml:"Type,attr"`
							DateLastEvaluated   string `xml:"DateLastEvaluated,attr"`
							Description         string `xml:"Description"`
							Explanation         struct {
								Text       string `xml:",chardata"`
								DataSource string `xml:"DataSource,attr"`
								Type       string `xml:"Type,attr"`
							} `xml:"Explanation"`
							Citation []struct {
								Text string `xml:",chardata"`
								Type string `xml:"Type,attr"`
								ID   struct {
									Text   string `xml:",chardata"`
									Source string `xml:"Source,attr"`
								} `xml:"ID"`
								URL string `xml:"URL"`
							} `xml:"Citation"`
							Comment struct {
								Text       string `xml:",chardata"`
								DataSource string `xml:"DataSource,attr"`
								Type       string `xml:"Type,attr"`
							} `xml:"Comment"`
							DescriptionHistory []struct {
								Text        string `xml:",chardata"`
								Dated       string `xml:"Dated,attr"`
								Description string `xml:"Description"`
							} `xml:"DescriptionHistory"`
							ConditionList struct {
								Text     string `xml:",chardata"`
								TraitSet []struct {
									Text  string `xml:",chardata"`
									ID    string `xml:"ID,attr"`
									Type  string `xml:"Type,attr"`
									Trait struct {
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
										} `xml:"Name"`
										AttributeSet []struct {
											Text      string `xml:",chardata"`
											Attribute struct {
												Text string `xml:",chardata"`
												Type string `xml:"Type,attr"`
											} `xml:"Attribute"`
											XRef struct {
												Text string `xml:",chardata"`
												ID   string `xml:"ID,attr"`
												DB   string `xml:"DB,attr"`
												Type string `xml:"Type,attr"`
											} `xml:"XRef"`
										} `xml:"AttributeSet"`
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
										} `xml:"Symbol"`
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
									} `xml:"Trait"`
								} `xml:"TraitSet"`
							} `xml:"ConditionList"`
						} `xml:"Interpretation"`
					} `xml:"Interpretations"`
					XRefList struct {
						Text string `xml:",chardata"`
						XRef []struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
							ID   string `xml:"ID,attr"`
							DB   string `xml:"DB,attr"`
						} `xml:"XRef"`
					} `xml:"XRefList"`
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
					FunctionalConsequence []struct {
						Text  string `xml:",chardata"`
						Value string `xml:"Value,attr"`
						XRef  struct {
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
					} `xml:"FunctionalConsequence"`
					OtherNameList struct {
						Text string   `xml:",chardata"`
						Name []string `xml:"Name"`
					} `xml:"OtherNameList"`
				} `xml:"SimpleAllele"`
				Name          string `xml:"Name"`
				VariationType string `xml:"VariationType"`
				OtherNameList struct {
					Text string `xml:",chardata"`
					Name string `xml:"Name"`
				} `xml:"OtherNameList"`
				HGVSlist struct {
					Text string `xml:",chardata"`
					HGVS []struct {
						Text                 string `xml:",chardata"`
						Type                 string `xml:"Type,attr"`
						Assembly             string `xml:"Assembly,attr"`
						NucleotideExpression struct {
							Text       string `xml:",chardata"`
							Expression string `xml:"Expression"`
						} `xml:"NucleotideExpression"`
					} `xml:"HGVS"`
				} `xml:"HGVSlist"`
				Interpretations struct {
					Text           string `xml:",chardata"`
					Interpretation struct {
						Text                string `xml:",chardata"`
						DateLastEvaluated   string `xml:"DateLastEvaluated,attr"`
						NumberOfSubmissions string `xml:"NumberOfSubmissions,attr"`
						NumberOfSubmitters  string `xml:"NumberOfSubmitters,attr"`
						Type                string `xml:"Type,attr"`
						Description         string `xml:"Description"`
						Citation            []struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
							ID   struct {
								Text   string `xml:",chardata"`
								Source string `xml:"Source,attr"`
							} `xml:"ID"`
							URL string `xml:"URL"`
						} `xml:"Citation"`
						ConditionList struct {
							Text     string `xml:",chardata"`
							TraitSet []struct {
								Text  string `xml:",chardata"`
								ID    string `xml:"ID,attr"`
								Type  string `xml:"Type,attr"`
								Trait struct {
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
											Type string `xml:"Type,attr"`
											ID   string `xml:"ID,attr"`
											DB   string `xml:"DB,attr"`
										} `xml:"XRef"`
									} `xml:"Name"`
									Symbol []struct {
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
									} `xml:"Symbol"`
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
											IntegerValue string `xml:"integerValue,attr"`
										} `xml:"Attribute"`
										XRef []struct {
											Text string `xml:",chardata"`
											ID   string `xml:"ID,attr"`
											DB   string `xml:"DB,attr"`
										} `xml:"XRef"`
									} `xml:"AttributeSet"`
									Citation []struct {
										Text   string `xml:",chardata"`
										Type   string `xml:"Type,attr"`
										Abbrev string `xml:"Abbrev,attr"`
										ID     []struct {
											Text   string `xml:",chardata"`
											Source string `xml:"Source,attr"`
										} `xml:"ID"`
									} `xml:"Citation"`
								} `xml:"Trait"`
							} `xml:"TraitSet"`
						} `xml:"ConditionList"`
						DescriptionHistory []struct {
							Text        string `xml:",chardata"`
							Dated       string `xml:"Dated,attr"`
							Description string `xml:"Description"`
						} `xml:"DescriptionHistory"`
					} `xml:"Interpretation"`
				} `xml:"Interpretations"`
				XRefList struct {
					Text string `xml:",chardata"`
					XRef []struct {
						Text string `xml:",chardata"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
						Type string `xml:"Type,attr"`
					} `xml:"XRef"`
				} `xml:"XRefList"`
				FunctionalConsequence struct {
					Text  string `xml:",chardata"`
					Value string `xml:"Value,attr"`
				} `xml:"FunctionalConsequence"`
			} `xml:"Haplotype"`
			OtherNameList struct {
				Text string   `xml:",chardata"`
				Name []string `xml:"Name"`
			} `xml:"OtherNameList"`
		} `xml:"Genotype"`
		GeneralCitations struct {
			Text     string `xml:",chardata"`
			Citation struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
				ID   struct {
					Text   string `xml:",chardata"`
					Source string `xml:"Source,attr"`
				} `xml:"ID"`
			} `xml:"Citation"`
		} `xml:"GeneralCitations"`
	} `xml:"InterpretedRecord"`
	IncludedRecord struct {
		Text         string `xml:",chardata"`
		SimpleAllele struct {
			Text        string `xml:",chardata"`
			AlleleID    string `xml:"AlleleID,attr"`
			VariationID string `xml:"VariationID,attr"`
			GeneList    struct {
				Text string `xml:",chardata"`
				Gene []struct {
					Text             string `xml:",chardata"`
					Symbol           string `xml:"Symbol,attr"`
					FullName         string `xml:"FullName,attr"`
					GeneID           string `xml:"GeneID,attr"`
					HGNCID           string `xml:"HGNC_ID,attr"`
					Source           string `xml:"Source,attr"`
					RelationshipType string `xml:"RelationshipType,attr"`
					Location         struct {
						Text                string `xml:",chardata"`
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
							Strand                   string `xml:"Strand,attr"`
						} `xml:"SequenceLocation"`
					} `xml:"Location"`
					OMIM               string `xml:"OMIM"`
					Haploinsufficiency struct {
						Text          string `xml:",chardata"`
						LastEvaluated string `xml:"last_evaluated,attr"`
						ClinGen       string `xml:"ClinGen,attr"`
					} `xml:"Haploinsufficiency"`
					Triplosensitivity struct {
						Text          string `xml:",chardata"`
						LastEvaluated string `xml:"last_evaluated,attr"`
						ClinGen       string `xml:"ClinGen,attr"`
					} `xml:"Triplosensitivity"`
					Property string `xml:"Property"`
				} `xml:"Gene"`
			} `xml:"GeneList"`
			Name          string `xml:"Name"`
			CanonicalSPDI string `xml:"CanonicalSPDI"`
			VariantType   string `xml:"VariantType"`
			Location      struct {
				Text                string `xml:",chardata"`
				CytogeneticLocation string `xml:"CytogeneticLocation"`
				SequenceLocation    []struct {
					Text                     string `xml:",chardata"`
					Assembly                 string `xml:"Assembly,attr"`
					AssemblyAccessionVersion string `xml:"AssemblyAccessionVersion,attr"`
					ForDisplay               string `xml:"forDisplay,attr"`
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
			} `xml:"Location"`
			ProteinChange []string `xml:"ProteinChange"`
			HGVSlist      struct {
				Text string `xml:",chardata"`
				HGVS []struct {
					Text                 string `xml:",chardata"`
					Assembly             string `xml:"Assembly,attr"`
					Type                 string `xml:"Type,attr"`
					NucleotideExpression struct {
						Text                     string `xml:",chardata"`
						SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
						SequenceAccession        string `xml:"sequenceAccession,attr"`
						SequenceVersion          string `xml:"sequenceVersion,attr"`
						Change                   string `xml:"change,attr"`
						Assembly                 string `xml:"Assembly,attr"`
						Expression               string `xml:"Expression"`
					} `xml:"NucleotideExpression"`
					ProteinExpression struct {
						Text                     string `xml:",chardata"`
						SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
						SequenceAccession        string `xml:"sequenceAccession,attr"`
						SequenceVersion          string `xml:"sequenceVersion,attr"`
						Change                   string `xml:"change,attr"`
						Expression               string `xml:"Expression"`
					} `xml:"ProteinExpression"`
					MolecularConsequence []struct {
						Text string `xml:",chardata"`
						ID   string `xml:"ID,attr"`
						Type string `xml:"Type,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"MolecularConsequence"`
				} `xml:"HGVS"`
			} `xml:"HGVSlist"`
			Interpretations struct {
				Text           string `xml:",chardata"`
				Interpretation struct {
					Text                string `xml:",chardata"`
					NumberOfSubmissions string `xml:"NumberOfSubmissions,attr"`
					NumberOfSubmitters  string `xml:"NumberOfSubmitters,attr"`
					Type                string `xml:"Type,attr"`
					Description         string `xml:"Description"`
					DescriptionHistory  struct {
						Text        string `xml:",chardata"`
						Dated       string `xml:"Dated,attr"`
						Description string `xml:"Description"`
					} `xml:"DescriptionHistory"`
				} `xml:"Interpretation"`
			} `xml:"Interpretations"`
			XRefList struct {
				Text string `xml:",chardata"`
				XRef []struct {
					Text string `xml:",chardata"`
					Type string `xml:"Type,attr"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
				} `xml:"XRef"`
			} `xml:"XRefList"`
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
			OtherNameList struct {
				Text string   `xml:",chardata"`
				Name []string `xml:"Name"`
			} `xml:"OtherNameList"`
			FunctionalConsequence struct {
				Text  string `xml:",chardata"`
				Value string `xml:"Value,attr"`
				XRef  struct {
					Text string `xml:",chardata"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
				} `xml:"XRef"`
			} `xml:"FunctionalConsequence"`
			Comment struct {
				Text       string `xml:",chardata"`
				DataSource string `xml:"DataSource,attr"`
				Type       string `xml:"Type,attr"`
			} `xml:"Comment"`
		} `xml:"SimpleAllele"`
		ReviewStatus    string `xml:"ReviewStatus"`
		Interpretations struct {
			Text           string `xml:",chardata"`
			Interpretation struct {
				Text                string `xml:",chardata"`
				NumberOfSubmissions string `xml:"NumberOfSubmissions,attr"`
				NumberOfSubmitters  string `xml:"NumberOfSubmitters,attr"`
				Type                string `xml:"Type,attr"`
				Description         string `xml:"Description"`
			} `xml:"Interpretation"`
		} `xml:"Interpretations"`
		SubmittedInterpretationList struct {
			Text string `xml:",chardata"`
			SCV  []struct {
				Text      string `xml:",chardata"`
				Title     string `xml:"Title,attr"`
				Accession string `xml:"Accession,attr"`
				Version   string `xml:"Version,attr"`
			} `xml:"SCV"`
		} `xml:"SubmittedInterpretationList"`
		InterpretedVariationList struct {
			Text                 string `xml:",chardata"`
			InterpretedVariation []struct {
				Text        string `xml:",chardata"`
				VariationID string `xml:"VariationID,attr"`
				Accession   string `xml:"Accession,attr"`
				Version     string `xml:"Version,attr"`
			} `xml:"InterpretedVariation"`
		} `xml:"InterpretedVariationList"`
		Haplotype struct {
			Text                string `xml:",chardata"`
			VariationID         string `xml:"VariationID,attr"`
			NumberOfChromosomes string `xml:"NumberOfChromosomes,attr"`
			SimpleAllele        []struct {
				Text        string `xml:",chardata"`
				AlleleID    string `xml:"AlleleID,attr"`
				VariationID string `xml:"VariationID,attr"`
				GeneList    struct {
					Text string `xml:",chardata"`
					Gene struct {
						Text             string `xml:",chardata"`
						Symbol           string `xml:"Symbol,attr"`
						FullName         string `xml:"FullName,attr"`
						GeneID           string `xml:"GeneID,attr"`
						HGNCID           string `xml:"HGNC_ID,attr"`
						Source           string `xml:"Source,attr"`
						RelationshipType string `xml:"RelationshipType,attr"`
						Location         struct {
							Text                string `xml:",chardata"`
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
								Strand                   string `xml:"Strand,attr"`
							} `xml:"SequenceLocation"`
						} `xml:"Location"`
						OMIM string `xml:"OMIM"`
					} `xml:"Gene"`
				} `xml:"GeneList"`
				Name          string `xml:"Name"`
				CanonicalSPDI string `xml:"CanonicalSPDI"`
				VariantType   string `xml:"VariantType"`
				Location      struct {
					Text                string `xml:",chardata"`
					CytogeneticLocation string `xml:"CytogeneticLocation"`
					SequenceLocation    []struct {
						Text                     string `xml:",chardata"`
						Assembly                 string `xml:"Assembly,attr"`
						AssemblyAccessionVersion string `xml:"AssemblyAccessionVersion,attr"`
						ForDisplay               string `xml:"forDisplay,attr"`
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
					} `xml:"SequenceLocation"`
				} `xml:"Location"`
				ProteinChange []string `xml:"ProteinChange"`
				HGVSlist      struct {
					Text string `xml:",chardata"`
					HGVS []struct {
						Text                 string `xml:",chardata"`
						Assembly             string `xml:"Assembly,attr"`
						Type                 string `xml:"Type,attr"`
						NucleotideExpression struct {
							Text                     string `xml:",chardata"`
							SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
							SequenceAccession        string `xml:"sequenceAccession,attr"`
							SequenceVersion          string `xml:"sequenceVersion,attr"`
							Change                   string `xml:"change,attr"`
							Assembly                 string `xml:"Assembly,attr"`
							Expression               string `xml:"Expression"`
						} `xml:"NucleotideExpression"`
						ProteinExpression struct {
							Text                     string `xml:",chardata"`
							SequenceAccessionVersion string `xml:"sequenceAccessionVersion,attr"`
							SequenceAccession        string `xml:"sequenceAccession,attr"`
							SequenceVersion          string `xml:"sequenceVersion,attr"`
							Change                   string `xml:"change,attr"`
							Expression               string `xml:"Expression"`
						} `xml:"ProteinExpression"`
						MolecularConsequence struct {
							Text string `xml:",chardata"`
							ID   string `xml:"ID,attr"`
							Type string `xml:"Type,attr"`
							DB   string `xml:"DB,attr"`
						} `xml:"MolecularConsequence"`
					} `xml:"HGVS"`
				} `xml:"HGVSlist"`
				Interpretations struct {
					Text           string `xml:",chardata"`
					Interpretation struct {
						Text                string `xml:",chardata"`
						DateLastEvaluated   string `xml:"DateLastEvaluated,attr"`
						NumberOfSubmissions string `xml:"NumberOfSubmissions,attr"`
						NumberOfSubmitters  string `xml:"NumberOfSubmitters,attr"`
						Type                string `xml:"Type,attr"`
						Description         string `xml:"Description"`
						Explanation         struct {
							Text       string `xml:",chardata"`
							DataSource string `xml:"DataSource,attr"`
							Type       string `xml:"Type,attr"`
						} `xml:"Explanation"`
						Citation struct {
							Text string `xml:",chardata"`
							Type string `xml:"Type,attr"`
							ID   struct {
								Text   string `xml:",chardata"`
								Source string `xml:"Source,attr"`
							} `xml:"ID"`
						} `xml:"Citation"`
						Comment struct {
							Text       string `xml:",chardata"`
							DataSource string `xml:"DataSource,attr"`
							Type       string `xml:"Type,attr"`
						} `xml:"Comment"`
						DescriptionHistory struct {
							Text        string `xml:",chardata"`
							Dated       string `xml:"Dated,attr"`
							Description string `xml:"Description"`
						} `xml:"DescriptionHistory"`
						ConditionList struct {
							Text     string `xml:",chardata"`
							TraitSet []struct {
								Text  string `xml:",chardata"`
								ID    string `xml:"ID,attr"`
								Type  string `xml:"Type,attr"`
								Trait struct {
									Text string `xml:",chardata"`
									ID   string `xml:"ID,attr"`
									Type string `xml:"Type,attr"`
									Name struct {
										Text         string `xml:",chardata"`
										ElementValue struct {
											Text string `xml:",chardata"`
											Type string `xml:"Type,attr"`
										} `xml:"ElementValue"`
										XRef struct {
											Text string `xml:",chardata"`
											ID   string `xml:"ID,attr"`
											DB   string `xml:"DB,attr"`
										} `xml:"XRef"`
									} `xml:"Name"`
									AttributeSet struct {
										Text      string `xml:",chardata"`
										Attribute struct {
											Text string `xml:",chardata"`
											Type string `xml:"Type,attr"`
										} `xml:"Attribute"`
										XRef struct {
											Text string `xml:",chardata"`
											ID   string `xml:"ID,attr"`
											DB   string `xml:"DB,attr"`
										} `xml:"XRef"`
									} `xml:"AttributeSet"`
									XRef []struct {
										Text string `xml:",chardata"`
										ID   string `xml:"ID,attr"`
										DB   string `xml:"DB,attr"`
										Type string `xml:"Type,attr"`
									} `xml:"XRef"`
									Symbol struct {
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
									} `xml:"Symbol"`
									Citation struct {
										Text   string `xml:",chardata"`
										Type   string `xml:"Type,attr"`
										Abbrev string `xml:"Abbrev,attr"`
										ID     []struct {
											Text   string `xml:",chardata"`
											Source string `xml:"Source,attr"`
										} `xml:"ID"`
									} `xml:"Citation"`
								} `xml:"Trait"`
							} `xml:"TraitSet"`
						} `xml:"ConditionList"`
					} `xml:"Interpretation"`
				} `xml:"Interpretations"`
				XRefList struct {
					Text string `xml:",chardata"`
					XRef []struct {
						Text string `xml:",chardata"`
						Type string `xml:"Type,attr"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"XRef"`
				} `xml:"XRefList"`
				FunctionalConsequence struct {
					Text  string `xml:",chardata"`
					Value string `xml:"Value,attr"`
					XRef  struct {
						Text string `xml:",chardata"`
						ID   string `xml:"ID,attr"`
						DB   string `xml:"DB,attr"`
					} `xml:"XRef"`
				} `xml:"FunctionalConsequence"`
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
			} `xml:"SimpleAllele"`
			Name          string `xml:"Name"`
			VariationType string `xml:"VariationType"`
			HGVSlist      struct {
				Text string `xml:",chardata"`
				HGVS struct {
					Text                 string `xml:",chardata"`
					Type                 string `xml:"Type,attr"`
					NucleotideExpression struct {
						Text       string `xml:",chardata"`
						Expression string `xml:"Expression"`
					} `xml:"NucleotideExpression"`
				} `xml:"HGVS"`
			} `xml:"HGVSlist"`
			Interpretations struct {
				Text           string `xml:",chardata"`
				Interpretation struct {
					Text                string `xml:",chardata"`
					NumberOfSubmissions string `xml:"NumberOfSubmissions,attr"`
					NumberOfSubmitters  string `xml:"NumberOfSubmitters,attr"`
					Type                string `xml:"Type,attr"`
					Description         string `xml:"Description"`
				} `xml:"Interpretation"`
			} `xml:"Interpretations"`
			FunctionalConsequence struct {
				Text  string `xml:",chardata"`
				Value string `xml:"Value,attr"`
			} `xml:"FunctionalConsequence"`
			XRefList struct {
				Text string `xml:",chardata"`
				XRef []struct {
					Text string `xml:",chardata"`
					ID   string `xml:"ID,attr"`
					DB   string `xml:"DB,attr"`
					Type string `xml:"Type,attr"`
				} `xml:"XRef"`
			} `xml:"XRefList"`
		} `xml:"Haplotype"`
	} `xml:"IncludedRecord"`
	ReplacedList struct {
		Text     string `xml:",chardata"`
		Replaced struct {
			Text        string `xml:",chardata"`
			DateChanged string `xml:"DateChanged,attr"`
			Accession   string `xml:"Accession,attr"`
			Version     string `xml:"Version,attr"`
			VariationID string `xml:"VariationID,attr"`
			Comment     struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
			} `xml:"Comment"`
		} `xml:"Replaced"`
	} `xml:"ReplacedList"`
}

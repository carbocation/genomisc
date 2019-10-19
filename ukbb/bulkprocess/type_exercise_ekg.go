package bulkprocess

import "encoding/xml"

// EKGExercise was generated 2019-10-19 15:07:26 by james on Skylake.
type EKGExercise struct {
	XMLName             xml.Name `xml:"CardiologyXML"`
	Text                string   `xml:",chardata"`
	ObservationType     string   `xml:"ObservationType"`
	ObservationDateTime struct {
		Text   string `xml:",chardata"`
		Hour   string `xml:"Hour"`
		Minute string `xml:"Minute"`
		Second string `xml:"Second"`
		Day    string `xml:"Day"`
		Month  string `xml:"Month"`
		Year   string `xml:"Year"`
	} `xml:"ObservationDateTime"`
	ObservationEndDateTime struct {
		Text   string `xml:",chardata"`
		Hour   string `xml:"Hour"`
		Minute string `xml:"Minute"`
		Second string `xml:"Second"`
		Day    string `xml:"Day"`
		Month  string `xml:"Month"`
		Year   string `xml:"Year"`
	} `xml:"ObservationEndDateTime"`
	ClinicalInfo struct {
		Text       string `xml:",chardata"`
		DeviceInfo struct {
			Text        string `xml:",chardata"`
			Desc        string `xml:"Desc"`
			SoftwareVer string `xml:"SoftwareVer"`
			AnalysisVer string `xml:"AnalysisVer"`
		} `xml:"DeviceInfo"`
	} `xml:"ClinicalInfo"`
	PatientVisit struct {
		Text                    string `xml:",chardata"`
		PatientClass            string `xml:"PatientClass"`
		AssignedPatientLocation struct {
			Text           string `xml:",chardata"`
			Facility       string `xml:"Facility"`
			LocationNumber string `xml:"LocationNumber"`
		} `xml:"AssignedPatientLocation"`
		AdmissionType     string `xml:"AdmissionType"`
		ServicingFacility string `xml:"ServicingFacility"`
	} `xml:"PatientVisit"`
	PatientInfo struct {
		Text string `xml:",chardata"`
		PID  string `xml:"PID"`
		Name struct {
			Text       string `xml:",chardata"`
			FamilyName string `xml:"FamilyName"`
			GivenName  string `xml:"GivenName"`
		} `xml:"Name"`
		Age struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"Age"`
		BirthDateTime struct {
			Text  string `xml:",chardata"`
			Day   string `xml:"Day"`
			Month string `xml:"Month"`
			Year  string `xml:"Year"`
		} `xml:"BirthDateTime"`
		Gender string `xml:"Gender"`
		Race   string `xml:"Race"`
		Height struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"Height"`
		Weight struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"Weight"`
		PaceMaker string `xml:"PaceMaker"`
	} `xml:"PatientInfo"`
	Medications          string `xml:"Medications"`
	ExerciseMeasurements struct {
		Text        string `xml:",chardata"`
		MaxWorkload struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"MaxWorkload"`
		ExercisePhaseTime struct {
			Text   string `xml:",chardata"`
			Minute string `xml:"Minute"`
			Second string `xml:"Second"`
		} `xml:"ExercisePhaseTime"`
		MaxHeartRate    string `xml:"MaxHeartRate"`
		MaxPredictedHR  string `xml:"MaxPredictedHR"`
		TargetHeartRate struct {
			Text     string `xml:",chardata"`
			Achieved string `xml:"achieved,attr"`
		} `xml:"TargetHeartRate"`
		PercentAchievedMaxPredicted string `xml:"PercentAchievedMaxPredicted"`
		PeakPercentMaxPredicted     string `xml:"PeakPercentMaxPredicted"`
		TargetHRFormula             string `xml:"TargetHRFormula"`
		TargetLoad                  struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"TargetLoad"`
		STHRSlope struct {
			Text          string `xml:",chardata"`
			STHRSlopeLead []struct {
				Text        string `xml:",chardata"`
				Lead        string `xml:"lead,attr"`
				Maxslope    string `xml:"maxslope,attr"`
				Slope       string `xml:"Slope"`
				Intercept   string `xml:"Intercept"`
				Correlation string `xml:"Correlation"`
				NumPoints   string `xml:"NumPoints"`
			} `xml:"STHRSlopeLead"`
		} `xml:"STHRSlope"`
		RestingStats struct {
			Text    string `xml:",chardata"`
			RestHR  string `xml:"RestHR"`
			RestVES string `xml:"RestVES"`
			RestST  struct {
				Text         string `xml:",chardata"`
				Measurements []struct {
					Text        string `xml:",chardata"`
					Lead        string `xml:"lead,attr"`
					STAmplitude struct {
						Text  string `xml:",chardata"`
						Units string `xml:"units,attr"`
					} `xml:"STAmplitude"`
					STSlope struct {
						Text  string `xml:",chardata"`
						Units string `xml:"units,attr"`
					} `xml:"STSlope"`
				} `xml:"Measurements"`
			} `xml:"RestST"`
		} `xml:"RestingStats"`
		PeakExStats struct {
			Text       string `xml:",chardata"`
			PeakExHR   string `xml:"PeakExHR"`
			PeakExMets string `xml:"PeakExMets"`
			PeakExVES  string `xml:"PeakExVES"`
			PeakExST   struct {
				Text         string `xml:",chardata"`
				Measurements []struct {
					Text        string `xml:",chardata"`
					Lead        string `xml:"lead,attr"`
					STAmplitude struct {
						Text  string `xml:",chardata"`
						Units string `xml:"units,attr"`
					} `xml:"STAmplitude"`
					STSlope struct {
						Text  string `xml:",chardata"`
						Units string `xml:"units,attr"`
					} `xml:"STSlope"`
				} `xml:"Measurements"`
			} `xml:"PeakExST"`
		} `xml:"PeakExStats"`
		MaxSTStats struct {
			Text           string `xml:",chardata"`
			MaxSTAmplitude struct {
				Text  string `xml:",chardata"`
				Lead  string `xml:"lead,attr"`
				Units string `xml:"units,attr"`
			} `xml:"MaxSTAmplitude"`
			MaxSTTime struct {
				Text   string `xml:",chardata"`
				Minute string `xml:"Minute"`
				Second string `xml:"Second"`
			} `xml:"MaxSTTime"`
			MaxSTHR   string `xml:"MaxSTHR"`
			MaxSTMets string `xml:"MaxSTMets"`
			MaxSTVES  string `xml:"MaxSTVES"`
			MaxST     struct {
				Text         string `xml:",chardata"`
				Measurements []struct {
					Text        string `xml:",chardata"`
					Lead        string `xml:"lead,attr"`
					STAmplitude struct {
						Text  string `xml:",chardata"`
						Units string `xml:"units,attr"`
					} `xml:"STAmplitude"`
					STSlope struct {
						Text  string `xml:",chardata"`
						Units string `xml:"units,attr"`
					} `xml:"STSlope"`
				} `xml:"Measurements"`
			} `xml:"MaxST"`
		} `xml:"MaxSTStats"`
	} `xml:"ExerciseMeasurements"`
	Interpretation string `xml:"Interpretation"`
	Protocol       struct {
		Text   string `xml:",chardata"`
		Device string `xml:"Device"`
		Phase  []struct {
			Text          string `xml:",chardata"`
			Ramping       string `xml:"ramping,attr"`
			ProtocolName  string `xml:"ProtocolName"`
			PhaseName     string `xml:"PhaseName"`
			PhaseDuration struct {
				Text   string `xml:",chardata"`
				Minute string `xml:"Minute"`
				Second string `xml:"Second"`
			} `xml:"PhaseDuration"`
			Stage []struct {
				Text          string `xml:",chardata"`
				Idx           string `xml:"idx,attr"`
				StageName     string `xml:"StageName"`
				StageDuration struct {
					Text   string `xml:",chardata"`
					Minute string `xml:"Minute"`
					Second string `xml:"Second"`
				} `xml:"StageDuration"`
			} `xml:"Stage"`
		} `xml:"Phase"`
	} `xml:"Protocol"`
	TrendData struct {
		Text            string `xml:",chardata"`
		NumberOfEntries string `xml:"NumberOfEntries"`
		TrendEntry      []struct {
			Text      string `xml:",chardata"`
			Idx       string `xml:"Idx,attr"`
			EntryTime struct {
				Text   string `xml:",chardata"`
				Minute string `xml:"Minute"`
				Second string `xml:"Second"`
			} `xml:"EntryTime"`
			HeartRate        string `xml:"HeartRate"`
			Mets             string `xml:"Mets"`
			VECount          string `xml:"VECount"`
			PaceCount        string `xml:"PaceCount"`
			Artifact         string `xml:"Artifact"`
			LeadMeasurements []struct {
				Text            string `xml:",chardata"`
				Lead            string `xml:"lead,attr"`
				JPointAmplitude struct {
					Text  string `xml:",chardata"`
					Units string `xml:"units,attr"`
				} `xml:"JPointAmplitude"`
				STAmplitude20ms struct {
					Text  string `xml:",chardata"`
					Units string `xml:"units,attr"`
				} `xml:"STAmplitude20ms"`
				STAmplitude struct {
					Text  string `xml:",chardata"`
					Units string `xml:"units,attr"`
				} `xml:"STAmplitude"`
				RAmplitude struct {
					Text  string `xml:",chardata"`
					Units string `xml:"units,attr"`
				} `xml:"RAmplitude"`
				R1Amplitude struct {
					Text  string `xml:",chardata"`
					Units string `xml:"units,attr"`
				} `xml:"R1Amplitude"`
				STSlope struct {
					Text  string `xml:",chardata"`
					Units string `xml:"units,attr"`
				} `xml:"STSlope"`
				STIntegral string `xml:"STIntegral"`
				STIndex    string `xml:"STIndex"`
			} `xml:"LeadMeasurements"`
			Load struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"Load"`
			Grade struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"Grade"`
			PhaseTime struct {
				Text   string `xml:",chardata"`
				Minute string `xml:"Minute"`
				Second string `xml:"Second"`
			} `xml:"PhaseTime"`
			PhaseName string `xml:"PhaseName"`
			StageTime struct {
				Text   string `xml:",chardata"`
				Minute string `xml:"Minute"`
				Second string `xml:"Second"`
			} `xml:"StageTime"`
			StageNumber string `xml:"StageNumber"`
			StageName   string `xml:"StageName"`
		} `xml:"TrendEntry"`
	} `xml:"TrendData"`
	MedianData struct {
		Text       string `xml:",chardata"`
		SampleRate struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"SampleRate"`
		Resolution struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"Resolution"`
		Median []struct {
			Text string `xml:",chardata"`
			Idx  string `xml:"Idx,attr"`
			Time struct {
				Text   string `xml:",chardata"`
				Minute string `xml:"Minute"`
				Second string `xml:"Second"`
			} `xml:"Time"`
			IPoint       string `xml:"I-Point"`
			JPoint       string `xml:"J-Point"`
			JXPoint      string `xml:"JX-Point"`
			PPoint       string `xml:"P-Point"`
			PEPoint      string `xml:"PE-Point"`
			QPoint       string `xml:"Q-Point"`
			TPoint       string `xml:"T-Point"`
			WaveformData []struct {
				Text     string `xml:",chardata"`
				Lead     string `xml:"lead,attr"`
				StartIdx string `xml:"startIdx,attr"`
			} `xml:"WaveformData"`
		} `xml:"Median"`
	} `xml:"MedianData"`
	TWACycleData struct {
		Text       string `xml:",chardata"`
		SampleRate struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"SampleRate"`
		Resolution struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"Resolution"`
	} `xml:"TWACycleData"`
	StripData struct {
		Text       string `xml:",chardata"`
		SampleRate struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"SampleRate"`
		Resolution struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"Resolution"`
		Strip struct {
			Text string `xml:",chardata"`
			Idx  string `xml:"idx,attr"`
			Time struct {
				Text   string `xml:",chardata"`
				Minute string `xml:"Minute"`
				Second string `xml:"Second"`
			} `xml:"Time"`
			WaveformData []struct {
				Text string `xml:",chardata"`
				Lead string `xml:"lead,attr"`
			} `xml:"WaveformData"`
		} `xml:"Strip"`
	} `xml:"StripData"`
	ArrhythmiaData struct {
		Text       string `xml:",chardata"`
		SampleRate struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"SampleRate"`
		Resolution struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"Resolution"`
	} `xml:"ArrhythmiaData"`
	FullDisclosure struct {
		Text      string `xml:",chardata"`
		StartTime struct {
			Text   string `xml:",chardata"`
			Minute string `xml:"Minute"`
			Second string `xml:"Second"`
		} `xml:"StartTime"`
		NumberOfChannels string `xml:"NumberOfChannels"`
		SampleRate       struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"SampleRate"`
		Resolution struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"Resolution"`
		LeadOrder          string `xml:"LeadOrder"`
		FullDisclosureData string `xml:"FullDisclosureData"`
	} `xml:"FullDisclosure"`
}

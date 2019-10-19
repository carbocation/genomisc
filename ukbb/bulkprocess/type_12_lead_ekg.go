package bulkprocess

import "encoding/xml"

// EKG12Lead was generated 2019-10-19 15:07:45 by james on Skylake.
type EKG12Lead struct {
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
	UID struct {
		Text          string `xml:",chardata"`
		DICOMStudyUID string `xml:"DICOMStudyUID"`
	} `xml:"UID"`
	ClinicalInfo struct {
		Text           string `xml:",chardata"`
		ReasonForStudy string `xml:"ReasonForStudy"`
		Technician     struct {
			Text       string `xml:",chardata"`
			FamilyName string `xml:"FamilyName"`
			GivenName  string `xml:"GivenName"`
			PersonID   string `xml:"PersonID"`
		} `xml:"Technician"`
		ObservationComment string `xml:"ObservationComment"`
		DeviceInfo         struct {
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
			LocationName   string `xml:"LocationName"`
		} `xml:"AssignedPatientLocation"`
		PatientRoom      string `xml:"PatientRoom"`
		AdmissionType    string `xml:"AdmissionType"`
		OrderingProvider struct {
			Text       string `xml:",chardata"`
			FamilyName string `xml:"FamilyName"`
			GivenName  string `xml:"GivenName"`
			PersonID   string `xml:"PersonID"`
		} `xml:"OrderingProvider"`
		AttendingDoctor struct {
			Text       string `xml:",chardata"`
			FamilyName string `xml:"FamilyName"`
			GivenName  string `xml:"GivenName"`
			PersonID   string `xml:"PersonID"`
		} `xml:"AttendingDoctor"`
		ReferringDoctor struct {
			Text       string `xml:",chardata"`
			FamilyName string `xml:"FamilyName"`
			GivenName  string `xml:"GivenName"`
			PersonID   string `xml:"PersonID"`
		} `xml:"ReferringDoctor"`
		ServicingFacility struct {
			Text    string `xml:",chardata"`
			Name    string `xml:"Name"`
			Address struct {
				Text    string `xml:",chardata"`
				Street1 string `xml:"Street1"`
				City    string `xml:"City"`
			} `xml:"Address"`
		} `xml:"ServicingFacility"`
		SysBP struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"SysBP"`
		DiaBP struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"DiaBP"`
		MedicalHistory struct {
			Text               string `xml:",chardata"`
			MedicalHistoryText string `xml:"MedicalHistoryText"`
		} `xml:"MedicalHistory"`
		OrderNumber string `xml:"OrderNumber"`
		Medications struct {
			Text   string `xml:",chardata"`
			Drug   string `xml:"Drug"`
			Dosage string `xml:"Dosage"`
		} `xml:"Medications"`
		ExtraQuestions struct {
			Text  string `xml:",chardata"`
			Label []struct {
				Text string `xml:",chardata"`
				Type string `xml:"Type,attr"`
			} `xml:"Label"`
			Content []string `xml:"Content"`
		} `xml:"ExtraQuestions"`
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
	FilterSetting struct {
		Text        string `xml:",chardata"`
		CubicSpline string `xml:"CubicSpline"`
		Filter50Hz  string `xml:"Filter50Hz"`
		Filter60Hz  string `xml:"Filter60Hz"`
		LowPass     struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"LowPass"`
		HighPass struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"HighPass"`
	} `xml:"FilterSetting"`
	DeviceType     string `xml:"Device-Type"`
	Interpretation struct {
		Text      string `xml:",chardata"`
		Diagnosis struct {
			Text          string   `xml:",chardata"`
			DiagnosisText []string `xml:"DiagnosisText"`
		} `xml:"Diagnosis"`
		Conclusion struct {
			Text           string   `xml:",chardata"`
			ConclusionText []string `xml:"ConclusionText"`
		} `xml:"Conclusion"`
	} `xml:"Interpretation"`
	RestingECGMeasurements struct {
		Text             string `xml:",chardata"`
		DiagnosisVersion string `xml:"DiagnosisVersion"`
		VentricularRate  struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"VentricularRate"`
		PQInterval struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"PQInterval"`
		PDuration struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"PDuration"`
		QRSDuration struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"QRSDuration"`
		QTInterval struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"QTInterval"`
		QTCInterval struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"QTCInterval"`
		RRInterval struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"RRInterval"`
		PPInterval struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"PPInterval"`
		SokolovLVHIndex struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"SokolovLVHIndex"`
		PAxis struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"PAxis"`
		RAxis struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"RAxis"`
		TAxis struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"TAxis"`
		QTDispersion struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"QTDispersion"`
		QTDispersionBazett struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"QTDispersionBazett"`
		QRSNum           string `xml:"QRSNum"`
		MeasurementTable struct {
			Text      string `xml:",chardata"`
			Creation  string `xml:"Creation,attr"`
			LeadOrder string `xml:"LeadOrder"`
			QDuration struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"QDuration"`
			RDuration struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"RDuration"`
			SDuration struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"SDuration"`
			RpDuration struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"RpDuration"`
			PAmplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"PAmplitude"`
			QAmplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"QAmplitude"`
			RAmplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"RAmplitude"`
			SAmplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"SAmplitude"`
			R1Amplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"R1Amplitude"`
			S1Amplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"S1Amplitude"`
			JAmplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"JAmplitude"`
			JXAmplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"JXAmplitude"`
			TAmplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"TAmplitude"`
			JXSlope struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"JXSlope"`
			R1Duration struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"R1Duration"`
			P1Amplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"P1Amplitude"`
			JXEAmplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"JXEAmplitude"`
			T1Amplitude struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"T1Amplitude"`
		} `xml:"MeasurementTable"`
		POnset struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"POnset"`
		POffset struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"POffset"`
		QOnset struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"QOnset"`
		QOffset struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"QOffset"`
		TOffset struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"TOffset"`
		MedianSamples struct {
			Text          string `xml:",chardata"`
			NumberOfLeads string `xml:"NumberOfLeads"`
			SampleRate    struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"SampleRate"`
			ChannelSampleCountTotal string `xml:"ChannelSampleCountTotal"`
			Resolution              struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"Resolution"`
			FirstValid struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"FirstValid"`
			LastValid struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"LastValid"`
			WaveformData []struct {
				Text string `xml:",chardata"`
				Lead string `xml:"lead,attr"`
			} `xml:"WaveformData"`
		} `xml:"MedianSamples"`
	} `xml:"RestingECGMeasurements"`
	VectorLoops struct {
		Text       string `xml:",chardata"`
		Creation   string `xml:"Creation,attr"`
		Resolution struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"Resolution"`
		ChannelSampleCountTotal string `xml:"ChannelSampleCountTotal"`
		POnset                  struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"POnset"`
		POffset struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"POffset"`
		QOnset struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"QOnset"`
		QOffset struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"QOffset"`
		TOffset struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"TOffset"`
		Frontal []struct {
			Text string `xml:",chardata"`
			Lead string `xml:"Lead,attr"`
		} `xml:"Frontal"`
		Horizontal []struct {
			Text string `xml:",chardata"`
			Lead string `xml:"Lead,attr"`
		} `xml:"Horizontal"`
		Sagittal []struct {
			Text string `xml:",chardata"`
			Lead string `xml:"Lead,attr"`
		} `xml:"Sagittal"`
	} `xml:"VectorLoops"`
	StripData struct {
		Text          string `xml:",chardata"`
		NumberOfLeads string `xml:"NumberOfLeads"`
		SampleRate    struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"SampleRate"`
		ChannelSampleCountTotal string `xml:"ChannelSampleCountTotal"`
		Resolution              struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"Resolution"`
		WaveformData []struct {
			Text string `xml:",chardata"`
			Lead string `xml:"lead,attr"`
		} `xml:"WaveformData"`
		ArrhythmiaResults struct {
			Text string `xml:",chardata"`
			Time []struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"Time"`
			BeatClass []string `xml:"BeatClass"`
		} `xml:"ArrhythmiaResults"`
	} `xml:"StripData"`
	FullDisclosure struct {
		Text             string `xml:",chardata"`
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
		FullDisclosureData struct {
			Text             string `xml:",chardata"`
			SampleCountTotal string `xml:"SampleCountTotal"`
		} `xml:"FullDisclosureData"`
		EventList struct {
			Text  string   `xml:",chardata"`
			Event []string `xml:"Event"`
			Time  []struct {
				Text   string `xml:",chardata"`
				Minute string `xml:"Minute"`
				Second string `xml:"Second"`
			} `xml:"Time"`
		} `xml:"EventList"`
		ArrhythmiaResults struct {
			Text string `xml:",chardata"`
			QRS  string `xml:"QRS"`
		} `xml:"ArrhythmiaResults"`
	} `xml:"FullDisclosure"`
	Export string `xml:"Export"`
	CSWeb  struct {
		Text string `xml:",chardata"`
		A    struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
		} `xml:"a"`
	} `xml:"CSWeb"`
}

package main

import "encoding/xml"

// CardiologyXML was generated 2021-02-14 12:58:11 by james on MacBook-Air.local.
type CardiologyXML struct {
	XMLName             xml.Name `xml:"CardiologyXML" json:"cardiologyxml,omitempty"`
	Text                string   `xml:",chardata" json:"text,omitempty"`
	ObservationType     string   `xml:"ObservationType"` // RestECG
	ObservationDateTime struct {
		Text   string `xml:",chardata" json:"text,omitempty"`
		Hour   string `xml:"Hour"`   // 12
		Minute string `xml:"Minute"` // 39
		Second string `xml:"Second"` // 11
		Day    string `xml:"Day"`    // 19
		Month  string `xml:"Month"`  // 10
		Year   string `xml:"Year"`   // 2019
	} `xml:"ObservationDateTime" json:"observationdatetime,omitempty"`
	UID struct {
		Text          string `xml:",chardata" json:"text,omitempty"`
		DICOMStudyUID string `xml:"DICOMStudyUID"` // 1.2.840.113619.2.235.3192...
	} `xml:"UID" json:"uid,omitempty"`
	ClinicalInfo struct {
		Text           string `xml:",chardata" json:"text,omitempty"`
		ReasonForStudy string `xml:"ReasonForStudy"`
		Technician     struct {
			Text       string `xml:",chardata" json:"text,omitempty"`
			FamilyName string `xml:"FamilyName"`
			GivenName  string `xml:"GivenName"`
			PersonID   string `xml:"PersonID"`
		} `xml:"Technician" json:"technician,omitempty"`
		ObservationComment string `xml:"ObservationComment"`
		DeviceInfo         struct {
			Text        string `xml:",chardata" json:"text,omitempty"`
			Desc        string `xml:"Desc"`        // CardioSoft
			SoftwareVer string `xml:"SoftwareVer"` // V6.73
			AnalysisVer string `xml:"AnalysisVer"` // 12SL V21
		} `xml:"DeviceInfo" json:"deviceinfo,omitempty"`
	} `xml:"ClinicalInfo" json:"clinicalinfo,omitempty"`
	PatientVisit struct {
		Text                    string `xml:",chardata" json:"text,omitempty"`
		PatientClass            string `xml:"PatientClass"` // O
		AssignedPatientLocation struct {
			Text           string `xml:",chardata" json:"text,omitempty"`
			Facility       string `xml:"Facility"`
			LocationNumber string `xml:"LocationNumber"` // 0
			LocationName   string `xml:"LocationName"`   // * 0 *
		} `xml:"AssignedPatientLocation" json:"assignedpatientlocation,omitempty"`
		PatientRoom      string `xml:"PatientRoom"`
		AdmissionType    string `xml:"AdmissionType"` // ROUT
		OrderingProvider struct {
			Text       string `xml:",chardata" json:"text,omitempty"`
			FamilyName string `xml:"FamilyName"`
			GivenName  string `xml:"GivenName"`
			PersonID   string `xml:"PersonID"`
		} `xml:"OrderingProvider" json:"orderingprovider,omitempty"`
		AttendingDoctor struct {
			Text       string `xml:",chardata" json:"text,omitempty"`
			FamilyName string `xml:"FamilyName"`
			GivenName  string `xml:"GivenName"`
			PersonID   string `xml:"PersonID"`
		} `xml:"AttendingDoctor" json:"attendingdoctor,omitempty"`
		ReferringDoctor struct {
			Text       string `xml:",chardata" json:"text,omitempty"`
			FamilyName string `xml:"FamilyName"`
			GivenName  string `xml:"GivenName"`
			PersonID   string `xml:"PersonID"`
		} `xml:"ReferringDoctor" json:"referringdoctor,omitempty"`
		ServicingFacility struct {
			Text    string `xml:",chardata" json:"text,omitempty"`
			Name    string `xml:"Name"`
			Address struct {
				Text    string `xml:",chardata" json:"text,omitempty"`
				Street1 string `xml:"Street1"`
				City    string `xml:"City"`
			} `xml:"Address" json:"address,omitempty"`
		} `xml:"ServicingFacility" json:"servicingfacility,omitempty"`
		SysBP struct {
			Text  string `xml:",chardata" json:"text,omitempty"`
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"SysBP" json:"sysbp,omitempty"`
		DiaBP struct {
			Text  string `xml:",chardata" json:"text,omitempty"`
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"DiaBP" json:"diabp,omitempty"`
		MedicalHistory struct {
			Text               string `xml:",chardata" json:"text,omitempty"`
			MedicalHistoryText string `xml:"MedicalHistoryText"`
		} `xml:"MedicalHistory" json:"medicalhistory,omitempty"`
		OrderNumber string `xml:"OrderNumber"`
		Medications struct {
			Text   string `xml:",chardata" json:"text,omitempty"`
			Drug   string `xml:"Drug"`
			Dosage string `xml:"Dosage"`
		} `xml:"Medications" json:"medications,omitempty"`
		ExtraQuestions struct {
			Text  string `xml:",chardata" json:"text,omitempty"`
			Label []struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				Type string `xml:"Type,attr" json:"type,omitempty"`
			} `xml:"Label" json:"label,omitempty"`
			Content []string `xml:"Content"`
		} `xml:"ExtraQuestions" json:"extraquestions,omitempty"`
	} `xml:"PatientVisit" json:"patientvisit,omitempty"`
	PatientInfo struct {
		Text string `xml:",chardata" json:"text,omitempty"`
		PID  string `xml:"PID"` // 1571485130
		Name struct {
			Text       string `xml:",chardata" json:"text,omitempty"`
			FamilyName string `xml:"FamilyName"` // 1571485130
			GivenName  string `xml:"GivenName"`  // ACE
		} `xml:"Name" json:"name,omitempty"`
		Age struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 53
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"Age" json:"age,omitempty"`
		BirthDateTime struct {
			Text  string `xml:",chardata" json:"text,omitempty"`
			Day   string `xml:"Day"`   // 1
			Month string `xml:"Month"` // 1
			Year  string `xml:"Year"`  // 1966
		} `xml:"BirthDateTime" json:"birthdatetime,omitempty"`
		Gender string `xml:"Gender"` // FEMALE
		Race   string `xml:"Race"`   // UNKNOWN
		Height struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 175
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"Height" json:"height,omitempty"`
		Weight struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 75.0
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"Weight" json:"weight,omitempty"`
		PaceMaker string `xml:"PaceMaker"` // no
	} `xml:"PatientInfo" json:"patientinfo,omitempty"`
	FilterSetting struct {
		Text        string `xml:",chardata" json:"text,omitempty"`
		CubicSpline string `xml:"CubicSpline"` // No
		Filter50Hz  string `xml:"Filter50Hz"`  // Yes
		Filter60Hz  string `xml:"Filter60Hz"`  // No
		LowPass     struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 100
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"LowPass" json:"lowpass,omitempty"`
		HighPass struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 0.01
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"HighPass" json:"highpass,omitempty"`
	} `xml:"FilterSetting" json:"filtersetting,omitempty"`
	DeviceType     string `xml:"Device-Type"` // 2
	Interpretation struct {
		Text      string `xml:",chardata" json:"text,omitempty"`
		Diagnosis struct {
			Text          string   `xml:",chardata" json:"text,omitempty"`
			DiagnosisText []string `xml:"DiagnosisText"` // Normal sinus rhythm, Norm...
		} `xml:"Diagnosis" json:"diagnosis,omitempty"`
		Conclusion struct {
			Text           string   `xml:",chardata" json:"text,omitempty"`
			ConclusionText []string `xml:"ConclusionText"` // Normal sinus rhythm, Norm...
		} `xml:"Conclusion" json:"conclusion,omitempty"`
	} `xml:"Interpretation" json:"interpretation,omitempty"`
	RestingECGMeasurements struct {
		Text             string `xml:",chardata" json:"text,omitempty"`
		DiagnosisVersion string `xml:"DiagnosisVersion"` // 12SL V21
		VentricularRate  struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 60
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"VentricularRate" json:"ventricularrate,omitempty"`
		PQInterval struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 140
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"PQInterval" json:"pqinterval,omitempty"`
		PDuration struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 98
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"PDuration" json:"pduration,omitempty"`
		QRSDuration struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 94
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"QRSDuration" json:"qrsduration,omitempty"`
		QTInterval struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 400
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"QTInterval" json:"qtinterval,omitempty"`
		QTCInterval struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 400
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"QTCInterval" json:"qtcinterval,omitempty"`
		RRInterval struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 996
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"RRInterval" json:"rrinterval,omitempty"`
		PPInterval struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 1000
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"PPInterval" json:"ppinterval,omitempty"`
		SokolovLVHIndex struct {
			Text  string `xml:",chardata" json:"text,omitempty"`
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"SokolovLVHIndex" json:"sokolovlvhindex,omitempty"`
		PAxis struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 75
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"PAxis" json:"paxis,omitempty"`
		RAxis struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 83
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"RAxis" json:"raxis,omitempty"`
		TAxis struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 65
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"TAxis" json:"taxis,omitempty"`
		QTDispersion struct {
			Text  string `xml:",chardata" json:"text,omitempty"`
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"QTDispersion" json:"qtdispersion,omitempty"`
		QTDispersionBazett struct {
			Text  string `xml:",chardata" json:"text,omitempty"`
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"QTDispersionBazett" json:"qtdispersionbazett,omitempty"`
		QRSNum           string `xml:"QRSNum"` // 10
		MeasurementTable struct {
			Text      string `xml:",chardata" json:"text,omitempty"`
			Creation  string `xml:"Creation,attr" json:"creation,omitempty"`
			LeadOrder string `xml:"LeadOrder"` // I, II, III, aVR, aVL, aVF...
			QDuration struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // , 13, 17, 54, , 14, , , 1...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"QDuration" json:"qduration,omitempty"`
			RDuration struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 48, 42, 41, 40, 22, 42, 3...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"RDuration" json:"rduration,omitempty"`
			SDuration struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 46, 39, 36, , 46, 38, 55,...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"SDuration" json:"sduration,omitempty"`
			RpDuration struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // , , , , , , , , , , ,
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"RpDuration" json:"rpduration,omitempty"`
			PAmplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 53, 166, 141, -102, -58, ...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"PAmplitude" json:"pamplitude,omitempty"`
			QAmplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // , 43, 53, 595, , 48, , , ...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"QAmplitude" json:"qamplitude,omitempty"`
			RAmplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 244, 1000, 869, 117, 34, ...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"RAmplitude" json:"ramplitude,omitempty"`
			SAmplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 107, 146, 102, , 390, 122...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"SAmplitude" json:"samplitude,omitempty"`
			R1Amplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // , , , , 34, , , , , , ,
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"R1Amplitude" json:"r1amplitude,omitempty"`
			S1Amplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // , , , , , , , , , , ,
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"S1Amplitude" json:"s1amplitude,omitempty"`
			JAmplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 4, 19, 14, -15, -5, 19, 0...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"JAmplitude" json:"jamplitude,omitempty"`
			JXAmplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 43, 68, 24, -59, 9, 48, 4...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"JXAmplitude" json:"jxamplitude,omitempty"`
			TAmplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 166, 405, 239, -288, -43,...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"TAmplitude" json:"tamplitude,omitempty"`
			JXSlope struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // , , , , , , , , , , ,
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"JXSlope" json:"jxslope,omitempty"`
			R1Duration struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // , , , , 26, , , , , , ,
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"R1Duration" json:"r1duration,omitempty"`
			P1Amplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // , , , , , , , , , , ,
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"P1Amplitude" json:"p1amplitude,omitempty"`
			JXEAmplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 73, 156, 83, -118, -5, 12...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"JXEAmplitude" json:"jxeamplitude,omitempty"`
			T1Amplitude struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // , , , , , , , , , , ,
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"T1Amplitude" json:"t1amplitude,omitempty"`
		} `xml:"MeasurementTable" json:"measurementtable,omitempty"`
		POnset struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 308
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"POnset" json:"ponset,omitempty"`
		POffset struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 406
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"POffset" json:"poffset,omitempty"`
		QOnset struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 448
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"QOnset" json:"qonset,omitempty"`
		QOffset struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 542
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"QOffset" json:"qoffset,omitempty"`
		TOffset struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 848
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"TOffset" json:"toffset,omitempty"`
		MedianSamples struct {
			Text          string `xml:",chardata" json:"text,omitempty"`
			NumberOfLeads string `xml:"NumberOfLeads"` // 12
			SampleRate    struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 500
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"SampleRate" json:"samplerate,omitempty"`
			ChannelSampleCountTotal string `xml:"ChannelSampleCountTotal"` // 600
			Resolution              struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 5
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"Resolution" json:"resolution,omitempty"`
			FirstValid struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 0
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"FirstValid" json:"firstvalid,omitempty"`
			LastValid struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 599
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"LastValid" json:"lastvalid,omitempty"`
			WaveformData []struct {
				Text string `xml:",chardata" json:"text,omitempty"` // 3,3,3,1,0,0,0,0,0,0,2,2,1...
				Lead string `xml:"lead,attr" json:"lead,omitempty"`
			} `xml:"WaveformData" json:"waveformdata,omitempty"`
		} `xml:"MedianSamples" json:"mediansamples,omitempty"`
	} `xml:"RestingECGMeasurements" json:"restingecgmeasurements,omitempty"`
	VectorLoops struct {
		Text       string `xml:",chardata" json:"text,omitempty"`
		Creation   string `xml:"Creation,attr" json:"creation,omitempty"`
		Resolution struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 5
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"Resolution" json:"resolution,omitempty"`
		ChannelSampleCountTotal string `xml:"ChannelSampleCountTotal"` // 599
		POnset                  struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 154
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"POnset" json:"ponset,omitempty"`
		POffset struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 203
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"POffset" json:"poffset,omitempty"`
		QOnset struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 224
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"QOnset" json:"qonset,omitempty"`
		QOffset struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 271
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"QOffset" json:"qoffset,omitempty"`
		TOffset struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 424
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"TOffset" json:"toffset,omitempty"`
		Frontal []struct {
			Text string `xml:",chardata" json:"text,omitempty"` // 7,7,7,7,7,8,8,8,7,7,6,6,6...
			Lead string `xml:"Lead,attr" json:"lead,omitempty"`
		} `xml:"Frontal" json:"frontal,omitempty"`
		Horizontal []struct {
			Text string `xml:",chardata" json:"text,omitempty"` // 7,7,7,7,7,8,8,8,7,7,6,6,6...
			Lead string `xml:"Lead,attr" json:"lead,omitempty"`
		} `xml:"Horizontal" json:"horizontal,omitempty"`
		Sagittal []struct {
			Text string `xml:",chardata" json:"text,omitempty"` // 3,2,2,2,3,3,3,3,3,3,2,2,2...
			Lead string `xml:"Lead,attr" json:"lead,omitempty"`
		} `xml:"Sagittal" json:"sagittal,omitempty"`
	} `xml:"VectorLoops" json:"vectorloops,omitempty"`
	StripData struct {
		Text          string `xml:",chardata" json:"text,omitempty"`
		NumberOfLeads string `xml:"NumberOfLeads"` // 12
		SampleRate    struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 500
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"SampleRate" json:"samplerate,omitempty"`
		ChannelSampleCountTotal string `xml:"ChannelSampleCountTotal"` // 5000
		Resolution              struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 5
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"Resolution" json:"resolution,omitempty"`
		WaveformData []struct {
			Text string `xml:",chardata" json:"text,omitempty"` // 0,0,-2,-3,0,7,9,5,6,10,11...
			Lead string `xml:"lead,attr" json:"lead,omitempty"`
		} `xml:"WaveformData" json:"waveformdata,omitempty"`
		ArrhythmiaResults struct {
			Text string `xml:",chardata" json:"text,omitempty"`
			Time []struct {
				Text  string `xml:",chardata" json:"text,omitempty"` // 855, 1840, 2835, 3830, 48...
				Units string `xml:"units,attr" json:"units,omitempty"`
			} `xml:"Time" json:"time,omitempty"`
			BeatClass []string `xml:"BeatClass"` // dominant, dominant, domin...
		} `xml:"ArrhythmiaResults" json:"arrhythmiaresults,omitempty"`
	} `xml:"StripData" json:"stripdata,omitempty"`
	FullDisclosure struct {
		Text             string `xml:",chardata" json:"text,omitempty"`
		NumberOfChannels string `xml:"NumberOfChannels"` // 12
		SampleRate       struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 500
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"SampleRate" json:"samplerate,omitempty"`
		Resolution struct {
			Text  string `xml:",chardata" json:"text,omitempty"` // 5
			Units string `xml:"units,attr" json:"units,omitempty"`
		} `xml:"Resolution" json:"resolution,omitempty"`
		LeadOrder          string `xml:"LeadOrder"` // I,II,III,aVR,aVL,aVF,V1,V...
		FullDisclosureData struct {
			Text             string `xml:",chardata" json:"text,omitempty"` // 46,47,47,48,49,49,51,49,4...
			SampleCountTotal string `xml:"SampleCountTotal"`                // 162000
		} `xml:"FullDisclosureData" json:"fulldisclosuredata,omitempty"`
		EventList struct {
			Text  string   `xml:",chardata" json:"text,omitempty"`
			Event []string `xml:"Event"` // L, L, L, L, L, L, L, L, L...
			Time  []struct {
				Text   string `xml:",chardata" json:"text,omitempty"`
				Minute string `xml:"Minute"` // 0, 0, 0, 0, 0, 0, 0, 0, 0...
				Second string `xml:"Second"` // 8.4, 9.4, 10.5, 11.5, 12....
			} `xml:"Time" json:"time,omitempty"`
		} `xml:"EventList" json:"eventlist,omitempty"`
		ArrhythmiaResults struct {
			Text string `xml:",chardata" json:"text,omitempty"`
			QRS  string `xml:"QRS"` // 15
		} `xml:"ArrhythmiaResults" json:"arrhythmiaresults,omitempty"`
	} `xml:"FullDisclosure" json:"fulldisclosure,omitempty"`
	Export string `xml:"Export"`
	CSWeb  struct {
		Text string `xml:",chardata" json:"text,omitempty"`
		A    struct {
			Text string `xml:",chardata" json:"text,omitempty"`
			Href string `xml:"href,attr" json:"href,omitempty"`
		} `xml:"a" json:"a,omitempty"`
	} `xml:"CSWeb" json:"csweb,omitempty"`
}

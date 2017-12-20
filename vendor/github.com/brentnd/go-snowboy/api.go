package snowboy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	EndpointBase    = "https://snowboy.kitt.ai/api/"
	EndpointVersion = "v1"
	EndpointTrain   = EndpointBase + EndpointVersion + "/train"
	EndpointImprove = EndpointBase + EndpointVersion + "/improve"
)

type AgeGroup string

const (
	AgeGroup0s     AgeGroup = "0_9"
	AgeGroup10s             = "10_19"
	AgeGroup20s             = "20_29"
	AgeGroup30s             = "30_39"
	AgeGroup40s             = "40_49"
	AgeGroup50s             = "50_59"
	AgeGroup60plus          = "60+"
)

type Gender string

const (
	GenderMale   Gender = "M"
	GenderFemale        = "F"
)

type Language string

const (
	LanguageArabic     Language = "ar"
	LanguageChinese             = "zh"
	LanguageDutch               = "nl"
	LanguageEnglish             = "en"
	LanguageFrench              = "fr"
	LanguageGerman              = "dt"
	LanguageHindi               = "hi"
	LanguageItalian             = "it"
	LanguageJapanese            = "jp"
	LanguageKorean              = "ko"
	LanguagePersian             = "fa"
	LanguagePolish              = "pl"
	LanguagePortuguese          = "pt"
	LanguageRussian             = "ru"
	LanguageSpanish             = "es"
	LanguageOther               = "ot"
)

type TrainRequest struct {
	VoiceSamples []VoiceSample `json:"voice_samples"`
	Token        string        `json:"token"`
	Name         string        `json:"name"`
	Language     Language      `json:"language"`
	AgeGroup     AgeGroup      `json:"age_group"`
	Gender       Gender        `json:"gender"`
	Microphone   string        `json:"microphone"`
}

type VoiceSample struct {
	Wave string `json:"wave"`
}

func (t *TrainRequest) AddWave(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	enc := base64.StdEncoding.EncodeToString(data)
	t.VoiceSamples = append(t.VoiceSamples, VoiceSample{
		Wave: enc,
	})
}

func (t *TrainRequest) Train() ([]byte, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return []byte{}, err
	}
	fmt.Println("sending", string(data), "to ", EndpointTrain)
	resp, err := http.DefaultClient.Post(EndpointTrain, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	d, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode, resp.Status, string(d))
	if resp.StatusCode != 200 {
		return []byte{}, errors.New("non-200 returned from kitt.ai")
	}
	return ioutil.ReadAll(resp.Body)
}

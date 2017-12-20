package snowboy

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/Kitt-AI/snowboy/swig/Go"
)

type snowboyResult int

const (
	snowboyResultSilence     snowboyResult = -2
	snowboyResultError                     = -1
	snowboyResultNoDetection               = 0
)

var (
	NoHandler           = errors.New("No handler installed")
	SnowboyLibraryError = errors.New("snowboy library error")
)

// Detector is holds the context and base impl for snowboy audio detection
type Detector struct {
	raw              snowboydetect.SnowboyDetect
	initialized      bool
	handlers         []handlerKeyword
	silenceHandler   *handlerKeyword
	modelStr         string
	sensitivityStr   string
	silenceThreshold time.Duration
	silenceElapsed   time.Duration
	ResourceFile     string
	AudioGain        float32
}

// Creates a standard Detector from a resources file
//
// Gives a default gain of 1.0
func NewDetector(resourceFile string) Detector {
	return Detector{
		ResourceFile: resourceFile,
		AudioGain:    1.0,
	}
}

// Fetch the format for the expected audio input
//
// Returns sample rate, number of channels, and bit depth
func (d *Detector) AudioFormat() (sampleRate, numChannels, bitsPerSample int) {
	d.initialize()
	sampleRate = d.raw.SampleRate()
	numChannels = d.raw.NumChannels()
	bitsPerSample = d.raw.BitsPerSample()
	return
}

// Close handles cleanup required by snowboy library
//
// Clients must call Close on detectors after doing any detection
// Returns error if Detector was never used
func (d *Detector) Close() error {
	if d.initialized {
		d.initialized = false
		snowboydetect.DeleteSnowboyDetect(d.raw)
		return nil
	} else {
		return errors.New("snowboy not initialize")
	}
}

// Calls previously installed handlers if hotwords are detected in data
//
// Does not chunk data because it is all available. Assumes entire
// length of data is filled and will only call one handler
func (d *Detector) Detect(data []byte) error {
	d.initialize()
	return d.route(d.runDetection(data))
}

// Install a handler for the given hotword
func (d *Detector) Handle(hotword Hotword, handler Handler) {
	if len(d.handlers) > 0 {
		d.modelStr += ","
		d.sensitivityStr += ","
	}
	d.modelStr += hotword.Model
	d.sensitivityStr += strconv.FormatFloat(float64(hotword.Sensitivity), 'f', 2, 64)
	d.handlers = append(d.handlers, handlerKeyword{
		Handler: handler,
		keyword: hotword.Name,
	})
}

// Installs a handle for the given hotword based on the func argument
// instead of the Handler interface
func (d *Detector) HandleFunc(hotword Hotword, handler func(string)) {
	d.Handle(hotword, handlerFunc(handler))
}

// Install a handler for when silence is detected
//
// threshold (time.Duration) determined how long silence can occur before callback is called
func (d *Detector) HandleSilence(threshold time.Duration, handler Handler) {
	d.silenceThreshold = threshold
	d.silenceHandler = &handlerKeyword{
		Handler: handler,
		keyword: "silence",
	}
}

// Installs a handle for when silence is detected based on the func argument
// instead of the Handler interface
//
// threshold (time.Duration) determined how long silence can occur before callback is called
func (d *Detector) HandleSilenceFunc(threshold time.Duration, handler func(string)) {
	d.HandleSilence(threshold, handlerFunc(handler))
}

// Reads from data and calls previously installed handlers when detection occurs
//
// Blocks while reading from data in chunks of 2048 bytes. Data examples include
// file, pipe from Stdout of exec, response from http call, any reader really.
//
// *Note, be careful with using byte.Buffer for data. When io.EOF is received from
// a read call on data, ReadAndDetect will exit
func (d *Detector) ReadAndDetect(data io.Reader) error {
	d.initialize()
	bytes := make([]byte, 2048)
	for {
		n, err := data.Read(bytes)
		if err != nil {
			if err == io.EOF {
				// Run detection on remaining bytes
				return d.route(d.runDetection(bytes))
			}
			return err
		}
		if n == 0 {
			// No data to read yet, but not eof so wait and try later
			time.Sleep(300 * time.Millisecond)
			continue
		}
		err = d.route(d.runDetection(bytes))
		if err != nil {
			return err
		}
	}
}

// Resets the detection. The underlying snowboy object handles voice
// activity detection (VAD) internally, but if you are using an external
// VAD, you should call Reset() whenever you see the segment end.
func (d *Detector) Reset() bool {
	d.initialize()
	return d.raw.Reset()
}

// Applies a fixed gain to the input audio. In case you have a very weak
// microphone, you can use this function to boost input audio level.
func (d *Detector) SetAudioGain(gain float32) {
	d.initialize()
	d.raw.SetAudioGain(gain)
}

// Returns the number of loaded hotwords
func (d *Detector) NumNotwords() int {
	d.initialize()
	return d.raw.NumHotwords()
}

// Applies or removes frontend audio processing
func (d *Detector) ApplyFrontend(apply bool) {
	d.initialize()
	d.raw.ApplyFrontend(apply)
}

func (d *Detector) initialize() {
	if d.initialized {
		return
	}
	if d.modelStr == "" {
		panic(errors.New("no models set for detector"))
	}
	d.raw = snowboydetect.NewSnowboyDetect(d.ResourceFile, d.modelStr)
	d.raw.SetSensitivity(d.sensitivityStr)
	d.raw.SetAudioGain(d.AudioGain)
	d.initialized = true
}

func (d *Detector) route(result snowboyResult) error {
	if result == snowboyResultError {
		return SnowboyLibraryError
	} else if result == snowboyResultSilence {
		if d.silenceElapsed >= d.silenceThreshold && d.silenceHandler != nil {
			d.silenceElapsed = 0
			d.silenceHandler.call()
		}
	} else if result != snowboyResultNoDetection {
		if len(d.handlers) >= int(result) {
			d.handlers[int(result)-1].call()
		} else {
			return NoHandler
		}
	}
	return nil
}

func (d *Detector) runDetection(data []byte) snowboyResult {
	if len(data) == 0 {
		return 0
	}
	ptr := snowboydetect.SwigcptrInt16_t(unsafe.Pointer(&data[0]))
	result := snowboyResult(d.raw.RunDetection(ptr, len(data)/2 /* len of int16 */))
	if result == snowboyResultSilence {
		sampleRate, numChannels, bitDepth := d.AudioFormat()
		dataElapseTime := len(data) * int(time.Second) / (numChannels * (bitDepth / 8) * sampleRate)
		d.silenceElapsed += time.Duration(dataElapseTime)
	} else {
		// Reset silence elapse duration because non-silence was detected
		d.silenceElapsed = 0
	}
	return result
}

// A Handler is used to handle when keywords are detected
//
// Detected will be call with the keyword string
type Handler interface {
	Detected(string)
}

type handlerKeyword struct {
	Handler
	keyword string
}

func (h handlerKeyword) call() {
	h.Handler.Detected(h.keyword)
}

type handlerFunc func(string)

func (f handlerFunc) Detected(keyword string) {
	f(keyword)
}

// A Hotword represents a model filename and sensitivity for a snowboy detectable word
//
// Model is the filename for the .umdl file
//
// Sensitivity is the sensitivity of this specific hotword
//
// Name is what will be used in calls to Handler.Detected(string)
type Hotword struct {
	Model       string
	Sensitivity float32
	Name        string
}

// Creates a hotword from model only, parsing the hotward name from the model filename
// and using a sensitivity of 0.5
func NewDefaultHotword(model string) Hotword {
	return NewHotword(model, 0.5)
}

// Creates a hotword from model and sensitivity only, parsing
// the hotward name from the model filename
func NewHotword(model string, sensitivity float32) Hotword {
	h := Hotword{
		Model:       model,
		Sensitivity: sensitivity,
	}
	name := strings.TrimRight(model, ".umdl")
	nameParts := strings.Split(name, "/")
	h.Name = nameParts[len(nameParts)-1]
	return h
}

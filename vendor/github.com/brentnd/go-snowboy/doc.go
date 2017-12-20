/*
Package go-snowboy provides a Go-style wrapper for the swig-generate Go code from Kitt-AI/snowboy

Simple audio hotword detection:

	det := snowboy.NewDetector("resources/common.res")
	defer det.Close()
	// ...
	det.Handle(snowboy.NewHotword("resources/alexa.umdl", 0.5), alexaHandler)

	det.HandleFunc(snowboy.NewHotword("resources/snowboy.umdl", 0.5), func(keyword string) {
		fmt.Println("detected 'snowboy' keyword")
	})

	var data io.Reader
	// ...
	log.Fatal(det.ReadAndDetect(data))

The Go bindings for snowboy audio detection (https://github.com/Kitt-AI/snowboy) are
generated using swig which creates a lot of extra types and uses calls with variable arguments.
This makes writing integrations in golang difficult because the types aren't explicit. go-snowboy
is intended to be a wrapper around the swig-generated Go code which will provide Go-style usage.

*/
package snowboy

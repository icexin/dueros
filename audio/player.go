package audio

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/bobertlo/go-mpg123/mpg123"
)

type Player struct {
	Writer *Writer
}

func NewPlayer() *Player {
	return &Player{}
}

func (p *Player) LoadMP3Reader(r io.Reader) (*Writer, error) {
	tmpfile, err := ioutil.TempFile("", "dueros-mp3")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()
	n, err := io.Copy(tmpfile, r)
	if err != nil {
		return nil, err
	}
	log.Printf("save %d bytes to %s", n, tmpfile.Name())
	return p.loadMP3File(tmpfile.Name())
}

func (p *Player) LoadMP3(uri string) (*Writer, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "http", "https":
		resp, err := http.Get(uri)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return p.LoadMP3Reader(resp.Body)
	case "", "file":
		return p.loadMP3File(u.Path)
	}
	return nil, errors.New("bad uri: " + uri)
}

func (p *Player) LoadAndPlay(uri string) error {
	w, err := p.LoadMP3(uri)
	if err != nil {
		return err
	}
	defer w.Close()
	return w.Play()
}

func (p *Player) loadMP3File(file string) (*Writer, error) {
	d, err := mpg123.NewDecoder("")
	if err != nil {
		return nil, err
	}
	defer d.Close()
	err = d.Open(file)
	if err != nil {
		return nil, err
	}
	rate, channels, encoding := d.GetFormat()
	log.Printf("rate:%d, channel:%d, encoding:%d", rate, channels, encoding)

	buf := new(bytes.Buffer)
	io.Copy(buf, d)

	return NewWriter(int(rate), channels, buf.Bytes())
}

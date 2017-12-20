package proto

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
)

var (
	ErrEmptyBody = errors.New("empty body")
)

type MessageHeader struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	MessageId       string `json:"messageId"`
	DialogRequestId string `json:"dialogRequestId"`
}

type Message struct {
	Header      MessageHeader `json:"header"`
	Payload     interface{}   `json:"payload"`
	PayloadJSON gjson.Result  `json:"-"`
	Attach      io.ReadCloser `json:"-"`
}

func NewMessage(name string, payload interface{}) *Message {
	i := strings.LastIndex(name, ".")
	namespace := name[:i]
	nam := name[i+1:]
	m := &Message{
		Header: MessageHeader{
			Namespace: namespace,
			Name:      nam,
			MessageId: uuid.NewV4().String(),
		},
		Payload: payload,
	}
	return m
}

func (m *Message) UnmarshalJSON(b []byte) error {
	root := gjson.ParseBytes(b)
	err := json.Unmarshal([]byte(root.Get("header").Raw), &m.Header)
	if err != nil {
		return err
	}

	m.PayloadJSON = root.Get("payload")
	if m.Payload != nil {
		err = json.Unmarshal([]byte(m.PayloadJSON.Raw), m.Payload)
		if err != nil {
			return err
		}
	}
	return nil
}

type ResponseReader struct {
	*http.Response
	r *multipart.Reader
}

func NewResponseReader(resp *http.Response) (*ResponseReader, error) {
	if resp.StatusCode == 204 {
		return nil, ErrEmptyBody
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	ctype := resp.Header.Get("Content-Type")
	ctype = strings.Replace(ctype, "/json", "-json", -1)
	mtype, params, err := mime.ParseMediaType(ctype)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(mtype, "multipart/") {
		return nil, errors.Errorf("bad content-type:%s", ctype)
	}

	boundary := params["boundary"]
	r := multipart.NewReader(resp.Body, boundary)
	return &ResponseReader{
		Response: resp,
		r:        r,
	}, nil
}

func (r *ResponseReader) ReadJSON() (*gjson.Result, error) {
	p, err := r.r.NextPart()
	if err != nil {
		return nil, err
	}
	defer p.Close()
	mtype, _, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	if mtype != "application/json" {
		return nil, errors.Errorf("application/json mime type expceted, actual:%s", mtype)
	}

	buf, err := ioutil.ReadAll(p)
	if err != nil {
		return nil, err
	}
	result := gjson.ParseBytes(buf)
	return &result, nil
}

func (r *ResponseReader) ReadDirective() (*Message, error) {
	root, err := r.ReadJSON()
	if err != nil {
		return nil, err
	}
	m := new(Message)
	err = json.Unmarshal([]byte(root.Get("directive").Raw), m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *ResponseReader) ReadAttach() (io.ReadCloser, error) {
	p, err := r.r.NextPart()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ResponseReader) Close() error {
	return r.Body.Close()
}

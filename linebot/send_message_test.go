package linebot

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestPushMessages(t *testing.T) {
	var toUserID = "U0cc15697597f61dd8b01cea8b027050e"
	type want struct {
		RequestBody []byte
		Response    *BasicResponse
		Error       error
	}
	var testCases = []struct {
		Messages     []Message
		Response     []byte
		ResponseCode int
		Want         want
	}{
		{
			// A test message
			Messages:     []Message{NewTextMessage("Hello, world")},
			ResponseCode: 200,
			Response:     []byte(`{}`),
			Want: want{
				RequestBody: []byte(`{"to":"U0cc15697597f61dd8b01cea8b027050e","messages":[{"type":"text","text":"Hello, world"}]}` + "\n"),
				Response:    &BasicResponse{},
			},
		},
		{
			// A image message
			Messages:     []Message{NewImageMessage("http://example.com/original.jpg", "http://example.com/preview.jpg")},
			ResponseCode: 200,
			Response:     []byte(`{}`),
			Want: want{
				RequestBody: []byte(`{"to":"U0cc15697597f61dd8b01cea8b027050e","messages":[{"type":"image","originalContentUrl":"http://example.com/original.jpg","previewImageUrl":"http://example.com/preview.jpg"}]}` + "\n"),
				Response:    &BasicResponse{},
			},
		},
		{
			// A location message
			Messages:     []Message{NewLocationMessage("title", "address", 35.65910807942215, 139.70372892916203)},
			ResponseCode: 200,
			Response:     []byte(`{}`),
			Want: want{
				RequestBody: []byte(`{"to":"U0cc15697597f61dd8b01cea8b027050e","messages":[{"type":"location","title":"title","address":"address","latitude":35.65910807942215,"longitude":139.70372892916203}]}` + "\n"),
				Response:    &BasicResponse{},
			},
		},
		{
			// A sticker message
			Messages:     []Message{NewStickerMessage("1", "1")},
			ResponseCode: 200,
			Response:     []byte(`{}`),
			Want: want{
				RequestBody: []byte(`{"to":"U0cc15697597f61dd8b01cea8b027050e","messages":[{"type":"sticker","packageId":"1","stickerId":"1"}]}` + "\n"),
				Response:    &BasicResponse{},
			},
		},
		{
			// Bad request
			Messages:     []Message{NewTextMessage(""), NewTextMessage("")},
			ResponseCode: 400,
			Response:     []byte(`{"message":"Request body has 2 error(s).","details":[{"message":"may not be empty","property":"messages[0].text"},{"message":"may not be empty","property":"messages[1].text"}]}`),
			Want: want{
				RequestBody: []byte(`{"to":"U0cc15697597f61dd8b01cea8b027050e","messages":[{"type":"text","text":""},{"type":"text","text":""}]}` + "\n"),
				Error: &APIError{
					Code: 400,
					Response: &ErrorResponse{
						Message: "Request body has 2 error(s).",
						Details: []errorResponseDetail{
							{
								Message:  "may not be empty",
								Property: "messages[0].text",
							},
							{
								Message:  "may not be empty",
								Property: "messages[1].text",
							},
						},
					},
				},
			},
		},
	}

	var currentTestIdx int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		tc := testCases[currentTestIdx]
		if r.Method != http.MethodPost {
			t.Errorf("Method %s; want %s", r.Method, http.MethodPost)
		}
		if r.URL.Path != APIEndpointEventsPush {
			t.Errorf("URLPath %s; want %s", r.URL.Path, APIEndpointEventsPush)
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(body, tc.Want.RequestBody) {
			t.Errorf("RequestBody %s; want %s", body, tc.Want.RequestBody)
		}
		w.WriteHeader(tc.ResponseCode)
		w.Write(tc.Response)
	}))
	defer server.Close()
	client, err := mockClient(server)
	if err != nil {
		t.Error(err)
	}
	for i, tc := range testCases {
		currentTestIdx = i
		res, err := client.Push(toUserID, tc.Messages).Do()
		if tc.Want.Error != nil {
			if !reflect.DeepEqual(err, tc.Want.Error) {
				t.Errorf("Error %d %q; want %q", i, err, tc.Want.Error)
			}
		} else {
			if err != nil {
				t.Error(err)
			}
		}
		if tc.Want.Response != nil {
			if !reflect.DeepEqual(res, tc.Want.Response) {
				t.Errorf("Response %d %q; want %q", i, res, tc.Want.Response)
			}
		}
	}
}

func TestPushMessagesWithContext(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		time.Sleep(10 * time.Millisecond)
		w.Write([]byte("{}"))
	}))
	defer server.Close()
	client, err := mockClient(server)
	if err != nil {
		t.Error(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, err = client.Push("U0cc15697597f61dd8b01cea8b027050e", []Message{NewTextMessage("Hello, world")}).WithContext(ctx).Do()
	if err != context.DeadlineExceeded {
		t.Errorf("err %v; want %v", err, context.Canceled)
	}
}

func TestReplyMessages(t *testing.T) {
	var replyToken = "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA"
	type want struct {
		RequestBody []byte
		Response    *BasicResponse
		Error       error
	}
	var testCases = []struct {
		Messages     []Message
		Response     []byte
		ResponseCode int
		Want         want
	}{
		{
			// A test message
			Messages:     []Message{NewTextMessage("Hello, world")},
			ResponseCode: 200,
			Response:     []byte(`{}`),
			Want: want{
				RequestBody: []byte(`{"replyToken":"nHuyWiB7yP5Zw52FIkcQobQuGDXCTA","messages":[{"type":"text","text":"Hello, world"}]}` + "\n"),
				Response:    &BasicResponse{},
			},
		},
		{
			// A location message
			Messages:     []Message{NewLocationMessage("title", "address", 35.65910807942215, 139.70372892916203)},
			ResponseCode: 200,
			Response:     []byte(`{}`),
			Want: want{
				RequestBody: []byte(`{"replyToken":"nHuyWiB7yP5Zw52FIkcQobQuGDXCTA","messages":[{"type":"location","title":"title","address":"address","latitude":35.65910807942215,"longitude":139.70372892916203}]}` + "\n"),
				Response:    &BasicResponse{},
			},
		},
		{
			// A image message
			Messages:     []Message{NewImageMessage("http://example.com/original.jpg", "http://example.com/preview.jpg")},
			ResponseCode: 200,
			Response:     []byte(`{}`),
			Want: want{
				RequestBody: []byte(`{"replyToken":"nHuyWiB7yP5Zw52FIkcQobQuGDXCTA","messages":[{"type":"image","originalContentUrl":"http://example.com/original.jpg","previewImageUrl":"http://example.com/preview.jpg"}]}` + "\n"),
				Response:    &BasicResponse{},
			},
		},
		{
			// A sticker message
			Messages:     []Message{NewStickerMessage("1", "1")},
			ResponseCode: 200,
			Response:     []byte(`{}`),
			Want: want{
				RequestBody: []byte(`{"replyToken":"nHuyWiB7yP5Zw52FIkcQobQuGDXCTA","messages":[{"type":"sticker","packageId":"1","stickerId":"1"}]}` + "\n"),
				Response:    &BasicResponse{},
			},
		},
		{
			// Bad request
			Messages:     []Message{NewTextMessage(""), NewTextMessage("")},
			ResponseCode: 400,
			Response:     []byte(`{"message":"Request body has 2 error(s).","details":[{"message":"may not be empty","property":"messages[0].text"},{"message":"may not be empty","property":"messages[1].text"}]}`),
			Want: want{
				RequestBody: []byte(`{"replyToken":"nHuyWiB7yP5Zw52FIkcQobQuGDXCTA","messages":[{"type":"text","text":""},{"type":"text","text":""}]}` + "\n"),
				Error: &APIError{
					Code: 400,
					Response: &ErrorResponse{
						Message: "Request body has 2 error(s).",
						Details: []errorResponseDetail{
							{
								Message:  "may not be empty",
								Property: "messages[0].text",
							},
							{
								Message:  "may not be empty",
								Property: "messages[1].text",
							},
						},
					},
				},
			},
		},
	}

	var currentTestIdx int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != http.MethodPost {
			t.Errorf("Method %s; want %s", r.Method, http.MethodPost)
		}
		if r.URL.Path != APIEndpointEventsReply {
			t.Errorf("URLPath %s; want %s", r.URL.Path, APIEndpointEventsReply)
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		tc := testCases[currentTestIdx]
		if !reflect.DeepEqual(body, tc.Want.RequestBody) {
			t.Errorf("RequestBody %s; want %s", body, tc.Want.RequestBody)
		}
		w.WriteHeader(tc.ResponseCode)
		w.Write(tc.Response)
	}))
	defer server.Close()
	client, err := mockClient(server)
	if err != nil {
		t.Error(err)
	}
	for i, tc := range testCases {
		currentTestIdx = i
		res, err := client.Reply(replyToken, tc.Messages).Do()
		if tc.Want.Error != nil {
			if !reflect.DeepEqual(err, tc.Want.Error) {
				t.Errorf("Error %d %q; want %q", i, err, tc.Want.Error)
			}
		} else {
			if err != nil {
				t.Error(err)
			}
		}
		if tc.Want.Response != nil {
			if !reflect.DeepEqual(res, tc.Want.Response) {
				t.Errorf("Response %d %q; want %q", i, res, tc.Want.Response)
			}
		}
	}
}

func TestReplyMessagesWithContext(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		time.Sleep(10 * time.Millisecond)
		w.Write([]byte("{}"))
	}))
	defer server.Close()
	client, err := mockClient(server)
	if err != nil {
		t.Error(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, err = client.Reply("nHuyWiB7yP5Zw52FIkcQobQuGDXCTA", []Message{NewTextMessage("Hello, world")}).WithContext(ctx).Do()
	if err != context.DeadlineExceeded {
		t.Errorf("err %v; want %v", err, context.Canceled)
	}
}

func BenchmarkPushMessages(b *testing.B) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Write([]byte("{}"))
	}))
	defer server.Close()
	client, err := mockClient(server)
	if err != nil {
		b.Error(err)
	}
	messages := []Message{NewTextMessage("Hello, world")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Push("U0cc15697597f61dd8b01cea8b027050e", messages).Do()
	}
}

func BenchmarkReplyMessages(b *testing.B) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Write([]byte("{}"))
	}))
	defer server.Close()
	client, err := mockClient(server)
	if err != nil {
		b.Error(err)
	}
	messages := []Message{NewTextMessage("Hello, world")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Reply("nHuyWiB7yP5Zw52FIkcQobQuGDXCTA", messages).Do()
	}
}

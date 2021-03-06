package couchdb2_go

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type DatabasesClient struct {
	*CouchDb2ConnDetails
}

func (d *DatabasesClient) GetConnection() *CouchDb2ConnDetails {
	return d.CouchDb2ConnDetails
}

func (d *DatabasesClient) Exists(string) (*bool, error) {
	panic("not implemented")
}

func (d *DatabasesClient) Meta(string) (*DbMetaResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) CreateDb(string) (*OkKoResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) DeleteDb(string) (*OkKoResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) CreateDocument(db string, doc interface{}) {
	panic("not implemented")
}

func (d *DatabasesClient) CreateDocumentExtra(db string, doc interface{}, batch bool, fullCommit bool) {
	panic("not implemented")
}

func (d *DatabasesClient) Documents(db string, req *AllDocsRequest) (*AllDocsResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) DocumentsWithIDs(db string, req *DocsWithIDsRequest) (*AllDocsResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) Bulk(db string, docs []interface{}, newEdits bool) (*BulkResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) Find(db string, req *FindRequest) (*FindResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) SetIndex(db string, req *SetIndexRequest) (*SetIndexResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) Index(db string) (*IndexResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) Delete(db string, designDoc string, name string) (*OkKoResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) Explain(db string, req *FindRequest) (*ExplainResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) Changes(db string, queryReq map[string]string) (*ChangesResponse, error) {
	panic("not implemented")
}

func buildURLParams(q map[string]string) (query string) {
	for k, v := range q {
		query = fmt.Sprintf("%s%s=%s&", query, k, v)
	}

	//remove last '&' on query
	query = strings.TrimSuffix(query, "&")

	return
}

func completeHeaders(r *http.Request) {
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/json")
}

func (d *DatabasesClient) setAuth(r *http.Request) {
	if d.Username != "" && d.Password != "" {
		r.SetBasicAuth(d.Username, d.Password)
	}
}

func handleResult(lineByt []byte, out chan *DbResult, quit chan struct{}, db string) {
	var result DbResult
	result.DbName = db

	if err := json.Unmarshal(lineByt, &result); err != nil {
		return
	}

	select {
	case <-quit:
		return
	case out <- &result:
	}
}

func handleScannerErr(err error, out chan *DbResult, db string, quit chan struct{}) {
	fmt.Printf("ERROR: scanner error: %s\n", err.Error())

	select {
	case <-quit:
		return
	case out <- &DbResult{
		DbName: db,
		ErrorResponse: &ErrorResponse{
			ErrorS: "Error scanning input",
			Reason: err.Error(),
		},
	}:
		close(out)
		close(quit)
	}
}

func dbResultHandler(httpRes *http.Response, out chan *DbResult, quit chan struct{}, db string) {
	defer httpRes.Body.Close()

	//Test
	reader := bufio.NewReader(httpRes.Body)

	ln, err := Readln(reader)
	for err == nil {
		handleResult(ln, out, quit, db)
		ln, err = Readln(reader)
	}

	handleScannerErr(err, out, db, quit)

	fmt.Println("Closing CouchDB client")

	close(out)
	close(quit)
}

func Readln(r *bufio.Reader) (ln []byte, err error) {
	var (
		isPrefix bool = true
		line     []byte
	)

	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}

	return ln, err
}

func (d *DatabasesClient) ChangesContinuousBytes(db string, queryReq map[string]string) (*http.Response, error) {
	if d.Client == nil {
		return nil, errors.New("You must set an HTTP Client to make requests. Current client is nil")
	}

	//take a map of kv and convert them into a "k=v&" string for URL params
	query := buildURLParams(queryReq)

	//build request
	fmt.Printf("Attempting connection to %s://%s/%s/_changes?%s\n", d.protocol, d.Address, db, query)
	reqHttp, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s/%s/_changes?%s", d.protocol, d.Address, db, query), nil)
	if err != nil {
		return nil, err
	}

	//set authentication
	d.setAuth(reqHttp)

	//set content-type and "accept"
	completeHeaders(reqHttp)

	//make request
	httpRes, err := d.Client.Do(reqHttp)
	if err != nil {
		return nil, err
	}

	return httpRes, err
}

func (d *DatabasesClient) ChangesContinuousRaw(db string, queryReq map[string]string, out chan *DbResult, quit chan struct{}) (chan *DbResult, chan<- struct{}, error) {
	if d.Client == nil {
		return nil, nil, errors.New("You must set an HTTP Client to make requests. Current client is nil")
	}

	//take a map of kv and convert them into a "k=v&" string for URL params
	query := buildURLParams(queryReq)

	//build request
	fmt.Printf("Attempting connection to %s://%s/%s/_changes?%s\n", d.protocol, d.Address, db, query)
	reqHttp, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s/%s/_changes?%s", d.protocol, d.Address, db, query), nil)
	if err != nil {
		return nil, nil, err
	}

	//set authentication
	d.setAuth(reqHttp)

	//set content-type and "accept"
	completeHeaders(reqHttp)

	//make request
	httpRes, err := d.Client.Do(reqHttp)
	if err != nil {
		return nil, nil, err
	}

	//create channels if necessary
	if out == nil {
		out = make(chan *DbResult, 10000)
	}
	if quit == nil {
		quit = make(chan struct{}, 1)
	}

	//Launch the listening goroutine that will close http.Body eventually
	go dbResultHandler(httpRes, out, quit, db)

	return out, quit, nil
}

func (d *DatabasesClient) Compact(db string) (*OkKoResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) CompactDesignDoc(db string, ddoc string) (*OkKoResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) EnsureFullCommit(db string) (*EnsureFullCommitResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) ViewCleanup(db string) (*OkKoResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) Security(db string) (*SecurityResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) SetSecurity(db string, req *SecurityRequest) (*OkKoResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) DoPurge(db string, req map[string]interface{}) (*DoPurgeResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) MissingKeys(db string, req map[string]interface{}) (*MissingKeysResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) RevsDiff(db string, req *RevsDiffRequest) (*RevsDiffResponse, error) {
	panic("not implemented")
}

func (d *DatabasesClient) RevsLimit(db string) (int, error) {
	panic("not implemented")
}

func (d *DatabasesClient) SetRevsLimit(db string, n int) (*OkKoResponse, error) {
	panic("not implemented")
}

func NewDatabase(timeout time.Duration, addr string, user, pass string, secure bool) (dat Database) {
	dat = &DatabasesClient{
		CouchDb2ConnDetails: NewConnection(timeout, addr, user, pass, secure),
	}

	return
}

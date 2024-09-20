package msclient

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	SharePointSiteId        = "root"
	SharePointShareDocument = "Shared Documents"
)

const (
	SmallFileMaxSize = 4 * 1024 * 1024
)

type SharePoint interface {
	Upload(ctx context.Context, dirId string, file *os.File, fileName string, fileSize int64) (*Value, error)
	List(ctx context.Context, dirId string) ([]Value, error)
	Download(ctx context.Context, fileWebUrl string) ([]byte, error)
}

func (c *MicrosoftGraph) MySharePoint(token Token) SharePoint {
	return &mySharePoint{token: token}
}

type mySharePoint struct {
	token Token
}

func (m mySharePoint) Download(ctx context.Context, fileWebUrl string) ([]byte, error) {
	return m.request(ctx, http.MethodGet, fileWebUrl, nil, nil)
}

func (m mySharePoint) request(ctx context.Context, method string, url string, payload io.Reader, extraHeader map[string][]string) ([]byte, error) {
	headers, err := m.token.HttpHeader(ctx)
	if err != nil {
		return nil, err
	}
	for k, vals := range extraHeader {
		headers[k] = vals
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header = headers
	client, err := m.token.HttpClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating http client: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	return body, nil
}

func (m mySharePoint) shareDocumentId(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/v1.0/sites/%s/lists", GraphAPIHost, SharePointSiteId)
	body, err := m.request(ctx, http.MethodGet, url, nil, nil)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}
	var items Answer
	_ = json.Unmarshal(body, &items)

	for _, item := range items.Value {
		if item.Name == SharePointShareDocument {
			return item.ID, nil
		}
	}
	return "", fmt.Errorf("no shared document found")
}

func (m mySharePoint) List(ctx context.Context, dirId string) ([]Value, error) {

	rootDirId, err := m.shareDocumentId(ctx)
	if err != nil {
		return nil, err
	}

	var data []Value

	var url = fmt.Sprintf(`%s/v1.0/sites/%s/lists/%s/items`, GraphAPIHost, SharePointSiteId, rootDirId)

	for {
		if url == "" {
			break
		}

		items, err := m.loopSharePointList(ctx, url)
		if err != nil {
			return nil, err
		}
		for _, item := range items.Value {
			if item.ParentReference.ID == dirId {
				data = append(data, item)
			}
		}
		url = items.OdataNextLink
	}

	return data, nil
}

func (m mySharePoint) loopSharePointList(ctx context.Context, url string) (Answer, error) {
	var items Answer

	body, err := m.request(ctx, http.MethodGet, url, nil, nil)
	if err != nil {
		return items, fmt.Errorf("error reading response body: %v", err)
	}
	_ = json.Unmarshal(body, &items)
	return items, nil
}

func (m mySharePoint) Upload(ctx context.Context, dirId string, file *os.File, fileName string, fileSize int64) (*Value, error) {

	ext := RegexGet(fileName, `(\.[^\.]+)$`)
	mineType, ok := Ext2Mime[ext]
	if !ok {
		return nil, fmt.Errorf("non-standard file extension")
	}

	headers, err := m.token.HttpHeader(ctx)
	if err != nil {
		return nil, err
	}

	headers.Set("Content-Type", mineType)
	headers.Set("Content-Length", fmt.Sprintf("%d", fileSize))

	if fileSize < SmallFileMaxSize {
		// PUT /sites/{site-id}/drive/items/{parent-id}:/{filename}:/content
		url := fmt.Sprintf("%s/v1.0/sites/%s/drive/items/%s:/%s:/content", GraphAPIHost, SharePointSiteId, dirId, fileName)

		return m.smallFileUpload(url, headers, file)
	} else {

		// POST /sites/{siteId}/drive/items/{itemId}/createUploadSession
		//sessionURL := fmt.Sprintf("%s/v1.0/sites/%s/drive/items/%s/createUploadSession", GraphAPIHost, SharePointSiteId, fileName)
		//sessionURL := fmt.Sprintf("%s/v1.0/sites/%s/drives/${driveId}/items/${folderId}:/${fileName}:/createUploadSession", GraphAPIHost, SharePointSiteId,root fileName)

		// POST /me/drive/items/{parentItemId}:/{fileName}:/createUploadSession
		sessionURL := fmt.Sprintf("%s/v1.0/me/drive/items/%s:/%s:/createUploadSession", GraphAPIHost, dirId, fileName)

		return m.bigFileUpload(sessionURL, headers, file, fileSize)
	}
}

/*
smallFileUpload 小文件上传
https://learn.microsoft.com/en-us/graph/api/driveitem-put-content?view=graph-rest-1.0&tabs=http#to-upload-a-new-file
*/
func (m mySharePoint) smallFileUpload(url string, headers http.Header, file io.Reader) (*Value, error) {
	ctx := context.Background()
	body, err := m.request(ctx, http.MethodPut, url, file, headers)
	if err != nil {
		return nil, err
	}
	v := &Value{}
	_ = json.Unmarshal(body, v)
	return v, nil
	//req, err := http.NewRequest(http.MethodPut, url, file)
	//if err != nil {
	//	return nil, fmt.Errorf("error creating request: %v", err)
	//}
	//
	//req.Header = headers
	//
	//client, err := m.token.HttpClient(ctx)
	//if err != nil {
	//	return nil, fmt.Errorf("error creating http client: %v", err)
	//}
	//
	//resp, err := client.Do(req)
	//if err != nil {
	//	return nil, fmt.Errorf("error sending request: %v", err)
	//}
	//defer resp.Body.Close()
	//
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return nil, fmt.Errorf("error reading response body: %v", err)
	//}
	//return body, nil
}

/*
https://learn.microsoft.com/en-us/graph/api/driveitem-createuploadsession?view=graph-rest-1.0#upload-bytes-to-the-upload-session
https://learn.microsoft.com/en-us/graph/sdks/large-file-upload
*/
func (m mySharePoint) bigFileUpload(url string, headers http.Header, file io.Reader, fileSize int64) (*Value, error) {
	ctx := context.Background()

	//fileuploader.UploadSession()
	//task := fileuploader.NewLargeFileUploadTask(m.token.HttpClient(ctx), )
	//task.Upload(func(current int64, total int64) {
	//
	//
	//})

	data, err := m.createUploadSession(ctx, url)
	if err != nil {
		return nil, err
	}

	sessionUrl := gjson.GetBytes(data, "uploadUrl")
	if !sessionUrl.Exists() {
		return nil, fmt.Errorf("the uploadUrl not found in big file upload session")
	}

	f := &bigFile{
		fileSize:     fileSize,
		currentWrite: 0,
		resp:         []byte{},
	}

	temp := make([]byte, 100*327680) //100 * 320kb
	_, err = io.CopyBuffer(f, file, temp)

	//req, err := http.NewRequest(http.MethodPut, sessionUrl.String(), file)
	//if err != nil {
	//	return nil, fmt.Errorf("error creating request: %v", err)
	//}
	//req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", f.currentWrite, f.currentWrite+int64(len(p))-1, f.fileSize))
	//
	//client, err := m.token.HttpClient(ctx)
	//if err != nil {
	//	return nil, fmt.Errorf("error creating http client: %v", err)
	//}
	//
	//if err != nil {
	//	return nil, fmt.Errorf("upload file failed: %v", err)
	//}

	v := &Value{}
	_ = json.Unmarshal(f.resp, v)
	return v, nil
}

type bigFile struct {
	fileSize     int64
	currentWrite int64
	resp         []byte
}

func (f *bigFile) Write(p []byte) (n int, err error) {
	//m := map[string]string{"Content-Range": fmt.Sprintf("bytes %d-%d/%d",
	//	f.currentWrite, f.currentWrite+int64(len(p))-1, f.fileSize)}
	//
	//fmt.Printf("%v \n", m["Content-Range"])
	//fmt.Printf("文件上传中==》%v \n", (f.currentWrite/f.fileSize)*100)
	//
	//req, err := http.NewRequest(http.MethodPut, f.sessionURL, bytes.NewReader(p))
	//if err != nil {
	//	return 0, err
	//}
	//
	//req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", f.currentWrite, f.currentWrite+int64(len(p))-1, f.fileSize))
	//client := &http.Client{}
	//resp, err := client.Do(req)
	//if err != nil {
	//	return 0, err
	//}
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return 0, err
	//}
	//
	//f.resp = body
	//f.currentWrite += int64(len(p))
	//fmt.Printf("%v \n", gjson.GetBytes(body, "@this|@pretty"))
	return len(p), err
}

func (m mySharePoint) createUploadSession(ctx context.Context, url string) ([]byte, error) {
	body, err := m.request(ctx, http.MethodPost, url, nil, nil)
	//req, err := http.NewRequest(http.MethodPost, url, nil)
	//if err != nil {
	//	return nil, fmt.Errorf("error creating request: %v", err)
	//}
	//
	//client, err := m.token.HttpClient(ctx)
	//if err != nil {
	//	return nil, fmt.Errorf("error creating http client: %v", err)
	//}
	//
	//resp, err := client.Do(req)
	//if err != nil {
	//	return nil, err
	//}
	//
	//body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tpl Answer
	_ = json.Unmarshal(body, &tpl)
	if tpl.Error.Code != "" {
		return nil, fmt.Errorf("api response error: %s", tpl.Error)
	}
	uploadURL := gjson.GetBytes(body, "uploadUrl")
	if !uploadURL.Exists() {
		return nil, fmt.Errorf("the uploadUrl not found in big file upload session")
	}
	return body, nil
}

func FileDownloadUrl(dirId string) string {
	return fmt.Sprintf("%s/v1.0/sites/%s/drive/items/%s/content", GraphAPIHost, SharePointSiteId, dirId)
}

package msclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type ErrJson struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	InnerError struct {
		RequestID string `json:"request-id"`
		Date      string `json:"date"`
	} `json:"innerError"`
}

func (e ErrJson) String() string {
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("code: %s, message: %d, Gender: %s", e.Code, e.Message)
	} else {
		return fmt.Sprintf("%s", b)
	}
}

type Folder struct {
	ChildCount int `json:"childCount"`
}

type Value struct {
	CreatedDateTime      time.Time `json:"createdDateTime"` // 创建时间
	ETag                 string    `json:"eTag"`
	ID                   string    `json:"id"`
	LastModifiedDateTime time.Time `json:"lastModifiedDateTime"`
	Name                 string    `json:"name"`
	WebURL               string    `json:"webUrl"`
	CTag                 string    `json:"cTag"`
	Size                 int64     `json:"size"`
	CreatedBy            struct {
		User struct {
			Email       string `json:"email"`
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"user"`
	} `json:"createdBy,omitempty"`
	LastModifiedBy struct {
		User struct {
			Email       string `json:"email"`
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"user"`
	} `json:"lastModifiedBy,omitempty"`
	ParentReference struct {
		DriveID   string `json:"driveId"`
		DriveType string `json:"driveType"`
		ID        string `json:"id"`
		Path      string `json:"path"`
	} `json:"parentReference"`
	FileSystemInfo struct {
		CreatedDateTime      time.Time `json:"createdDateTime"`
		LastModifiedDateTime time.Time `json:"lastModifiedDateTime"`
	} `json:"fileSystemInfo"`
	Folder        Folder `json:"folder,omitempty"`
	SpecialFolder struct {
		Name string `json:"name"`
	} `json:"specialFolder,omitempty"`
	MicrosoftGraphDownloadURL string `json:"@microsoft.graph.downloadUrl,omitempty"`
	File                      struct {
		MimeType string `json:"mimeType"`
		Hashes   struct {
			QuickXorHash string `json:"quickXorHash"`
		} `json:"hashes"`
	} `json:"file,omitempty"`
	Shared struct {
		Scope string `json:"scope"`
	} `json:"shared,omitempty"`
	Image struct {
		Height int `json:"height"`
		Width  int `json:"width"`
	} `json:"image,omitempty"`
	Deleted struct {
		State string `json:"state"`
	} `json:"deleted"`
	ContentType ContentType `json:"contentType"`
}

type ContentType struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (t ContentType) IsFolder() bool {
	return t.Name == "Folder"
}

type Answer struct {
	OdataContext   string  `json:"@odata.context"`
	OdataNextLink  string  `json:"@odata.nextLink"`
	OdataDeltaLink string  `json:"@odata.deltaLink"`
	Value          []Value `json:"value"`
	Error          ErrJson `json:"error,omitempty"`
}

// UnmarshalJSON 修改 Folder 的默认值，为 -1 时不是文件夹
func (v *Value) UnmarshalJSON(b []byte) error {
	type xvalue Value

	xf := &xvalue{Folder: Folder{ChildCount: -1}}
	if err := json.Unmarshal(b, xf); err != nil {
		return err
	}
	*v = Value(*xf)
	return nil
}

// CheckAnswerValid 判断收到的 Answer 是否正常
func CheckAnswerValid(ans Answer, relativePath string) error {
	if ans.Error.Code != "" {
		return errors.New("获取的 Answer 不正确")
	}
	return nil
}

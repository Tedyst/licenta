package nvd

import "time"

type NvdCpeAPIResult struct {
	ResultsPerPage int64                  `json:"resultsPerPage"`
	StartIndex     int64                  `json:"startIndex"`
	TotalResults   int64                  `json:"totalResults"`
	Format         string                 `json:"format"`
	Version        string                 `json:"version"`
	Timestamp      string                 `json:"timestamp"`
	Products       []NvdCpeProductElement `json:"products"`
}

type NvdCpeProductElement struct {
	Cpe NvdCpeCpe `json:"cpe"`
}

type NvdCpeCpe struct {
	Deprecated   bool          `json:"deprecated"`
	CpeName      string        `json:"cpeName"`
	CpeNameID    string        `json:"cpeNameId"`
	LastModified string        `json:"lastModified"`
	Created      string        `json:"created"`
	Titles       []NvdCpeTitle `json:"titles"`
	Refs         []NvdCpeRef   `json:"refs"`
}

func (c *NvdCpeCpe) LastModifiedDate() (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05", c.LastModified)
}

type NvdCpeRef struct {
	Ref  string     `json:"ref"`
	Type NvdCpeType `json:"type"`
}

type NvdCpeTitle struct {
	Title string     `json:"title"`
	Lang  NvdCpeLang `json:"lang"`
}

type NvdCpeType string

const (
	NvdCpeAdvisory  NvdCpeType = "Advisory"
	NvdCpeChangeLog NvdCpeType = "Change Log"
	NvdCpeProduct   NvdCpeType = "Product"
	NvdCpeVendor    NvdCpeType = "Vendor"
	NvdCpeVersion   NvdCpeType = "Version"
)

type NvdCpeLang string

const (
	NvdCpeEn NvdCpeLang = "en"
)

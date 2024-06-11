package nvd

import (
	"errors"
	"time"
)

type NvdCveAPIResult struct {
	ResultsPerPage  int64                 `json:"resultsPerPage"`
	StartIndex      int64                 `json:"startIndex"`
	TotalResults    int64                 `json:"totalResults"`
	Format          string                `json:"format"`
	Version         string                `json:"version"`
	Timestamp       string                `json:"timestamp"`
	Vulnerabilities []NvdCveVulnerability `json:"vulnerabilities"`
}

type NvdCveVulnerability struct {
	Cve NvdCveCve `json:"cve"`
}

type NvdCveCve struct {
	ID               string                `json:"id"`
	SourceIdentifier SourceIdentifierEnum  `json:"sourceIdentifier"`
	Published        string                `json:"published"`
	LastModified     string                `json:"lastModified"`
	VulnStatus       VulnStatus            `json:"vulnStatus"`
	Descriptions     []NvdCveDescription   `json:"descriptions"`
	Metrics          NvdCveMetrics         `json:"metrics"`
	Weaknesses       []NvdCveWeakness      `json:"weaknesses"`
	Configurations   []NvdCveConfiguration `json:"configurations"`
	References       []NvdCveReference     `json:"references"`
	VendorComments   []NvdCveVendorComment `json:"vendorComments"`
	EvaluatorComment *string               `json:"evaluatorComment,omitempty"`
}

func (c *NvdCveCve) PubslihedDate() (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05", c.Published)
}

func (c *NvdCveCve) LastModifiedDate() (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05", c.LastModified)
}

func (c *NvdCveCve) Score() (float64, error) {
	value := 0.0
	for _, metric := range c.Metrics.CvssMetricV31 {
		if metric.CvssData.BaseScore > value {
			value = metric.CvssData.BaseScore
		}
	}
	for _, metric := range c.Metrics.CvssMetricV30 {
		if metric.CvssData.BaseScore > value {
			value = metric.CvssData.BaseScore
		}
	}
	for _, metric := range c.Metrics.CvssMetricV2 {
		if metric.CvssData.BaseScore > value {
			value = metric.CvssData.BaseScore
		}
	}

	if value == 0.0 {
		return 0.0, errors.New("no score found")
	}

	return value, nil
}

type NvdCveConfiguration struct {
	Nodes    []NvdCveNode `json:"nodes"`
	Operator *string      `json:"operator,omitempty"`
}

type NvdCveNode struct {
	Operator NvdCveOperator   `json:"operator"`
	Negate   bool             `json:"negate"`
	CpeMatch []NvdCveCpeMatch `json:"cpeMatch"`
}

type NvdCveCpeMatch struct {
	Vulnerable            bool    `json:"vulnerable"`
	Criteria              string  `json:"criteria"`
	MatchCriteriaID       string  `json:"matchCriteriaId"`
	VersionStartIncluding *string `json:"versionStartIncluding,omitempty"`
	VersionEndExcluding   *string `json:"versionEndExcluding,omitempty"`
	VersionEndIncluding   *string `json:"versionEndIncluding,omitempty"`
}

type NvdCveDescription struct {
	Lang  NvdCveLang `json:"lang"`
	Value string     `json:"value"`
}

type NvdCveMetrics struct {
	CvssMetricV2  []NvdCveCvssMetricV2 `json:"cvssMetricV2"`
	CvssMetricV30 []NvdCveCvssMetricV3 `json:"cvssMetricV30"`
	CvssMetricV31 []NvdCveCvssMetricV3 `json:"cvssMetricV31"`
}

type NvdCveCvssMetricV2 struct {
	Source                  NvdCveCvssMetricV2Source   `json:"source"`
	Type                    Type                       `json:"type"`
	CvssData                NvdCveCvssMetricV2CvssData `json:"cvssData"`
	BaseSeverity            NvdCveIty                  `json:"baseSeverity"`
	ExploitabilityScore     float64                    `json:"exploitabilityScore"`
	ImpactScore             float64                    `json:"impactScore"`
	ACInsufInfo             bool                       `json:"acInsufInfo"`
	ObtainAllPrivilege      bool                       `json:"obtainAllPrivilege"`
	ObtainUserPrivilege     bool                       `json:"obtainUserPrivilege"`
	ObtainOtherPrivilege    bool                       `json:"obtainOtherPrivilege"`
	UserInteractionRequired *bool                      `json:"userInteractionRequired,omitempty"`
}

type NvdCveCvssMetricV2CvssData struct {
	Version               string               `json:"version"`
	VectorString          string               `json:"vectorString"`
	AccessVector          NvdCveVector         `json:"accessVector"`
	AccessComplexity      NvdCveIty            `json:"accessComplexity"`
	Authentication        NvdCveAuthentication `json:"authentication"`
	ConfidentialityImpact NvdCveItyImpact      `json:"confidentialityImpact"`
	IntegrityImpact       NvdCveItyImpact      `json:"integrityImpact"`
	AvailabilityImpact    NvdCveItyImpact      `json:"availabilityImpact"`
	BaseScore             float64              `json:"baseScore"`
}

type NvdCveCvssMetricV3 struct {
	Source              NvdCveCvssMetricV2Source    `json:"source"`
	Type                Type                        `json:"type"`
	CvssData            NvdCveCvssMetricV30CvssData `json:"cvssData"`
	ExploitabilityScore float64                     `json:"exploitabilityScore"`
	ImpactScore         float64                     `json:"impactScore"`
}

type NvdCveCvssMetricV30CvssData struct {
	Version               string             `json:"version"`
	VectorString          string             `json:"vectorString"`
	AttackVector          NvdCveVector       `json:"attackVector"`
	AttackComplexity      NvdCveIty          `json:"attackComplexity"`
	PrivilegesRequired    AvailabilityImpact `json:"privilegesRequired"`
	UserInteraction       UserInteraction    `json:"userInteraction"`
	Scope                 Scope              `json:"scope"`
	ConfidentialityImpact AvailabilityImpact `json:"confidentialityImpact"`
	IntegrityImpact       AvailabilityImpact `json:"integrityImpact"`
	AvailabilityImpact    AvailabilityImpact `json:"availabilityImpact"`
	BaseScore             float64            `json:"baseScore"`
	BaseSeverity          BaseSeverity       `json:"baseSeverity"`
}

type NvdCveReference struct {
	URL    string               `json:"url"`
	Source SourceIdentifierEnum `json:"source"`
	Tags   []Tag                `json:"tags"`
}

type NvdCveVendorComment struct {
	Organization string `json:"organization"`
	Comment      string `json:"comment"`
	LastModified string `json:"lastModified"`
}

type NvdCveWeakness struct {
	Source      NvdCveCvssMetricV2Source `json:"source"`
	Type        Type                     `json:"type"`
	Description []NvdCveDescription      `json:"description"`
}

type NvdCveOperator string

const (
	Or NvdCveOperator = "OR"
)

type NvdCveLang string

const (
	NvdCveEn NvdCveLang = "en"
	NvdCveEs NvdCveLang = "es"
)

type NvdCveIty string

const (
	NvdCveItyHIGH   NvdCveIty = "HIGH"
	NvdCveItyLOW    NvdCveIty = "LOW"
	NvdCveItyMEDIUM NvdCveIty = "MEDIUM"
)

type NvdCveVector string

const (
	NvdCveLocal   NvdCveVector = "LOCAL"
	NvdCveNetwork NvdCveVector = "NETWORK"
)

type NvdCveAuthentication string

const (
	NvdCveAuthenticationNONE NvdCveAuthentication = "NONE"
	NvdCveSingle             NvdCveAuthentication = "SINGLE"
)

type NvdCveItyImpact string

const (
	NvdCveComplete      NvdCveItyImpact = "COMPLETE"
	NvdCveItyImpactNONE NvdCveItyImpact = "NONE"
	NvdCvePartial       NvdCveItyImpact = "PARTIAL"
)

type NvdCveCvssMetricV2Source string

type Type string

const (
	Primary   Type = "Primary"
	Secondary Type = "Secondary"
)

type AvailabilityImpact string

const (
	AvailabilityImpactHIGH AvailabilityImpact = "HIGH"
	AvailabilityImpactLOW  AvailabilityImpact = "LOW"
	AvailabilityImpactNONE AvailabilityImpact = "NONE"
)

type BaseSeverity string

const (
	BaseSeverityHIGH   BaseSeverity = "HIGH"
	BaseSeverityMEDIUM BaseSeverity = "MEDIUM"
	Critical           BaseSeverity = "CRITICAL"
)

type Scope string

const (
	Changed   Scope = "CHANGED"
	Unchanged Scope = "UNCHANGED"
)

type UserInteraction string

const (
	Required            UserInteraction = "REQUIRED"
	UserInteractionNONE UserInteraction = "NONE"
)

type SourceIdentifierEnum string

type Tag string

const (
	BrokenLink          Tag = "Broken Link"
	Exploit             Tag = "Exploit"
	IssueTracking       Tag = "Issue Tracking"
	MailingList         Tag = "Mailing List"
	NotApplicable       Tag = "Not Applicable"
	Patch               Tag = "Patch"
	PermissionsRequired Tag = "Permissions Required"
	ReleaseNotes        Tag = "Release Notes"
	ThirdPartyAdvisory  Tag = "Third Party Advisory"
	URLRepurposed       Tag = "URL Repurposed"
	VDBEntry            Tag = "VDB Entry"
	VendorAdvisory      Tag = "Vendor Advisory"
)

type VulnStatus string

const (
	Analyzed VulnStatus = "Analyzed"
	Modified VulnStatus = "Modified"
)

package dto

// GSTR1Payload represents the root GSTR-1 offline utility JSON schema.
type GSTR1Payload struct {
	GSTIN    string          `json:"gstin"`
	FP       string          `json:"fp"`
	GT       float64         `json:"gt"`
	CurGT    float64         `json:"cur_gt"`
	Version  string          `json:"version"`
	Hash     string          `json:"hash"`
	B2B      []GSTR1B2B      `json:"b2b,omitempty"`
	B2CS     []B2CSRow       `json:"b2cs,omitempty"`
	CDNR     []GSTR1CDNR     `json:"cdnr,omitempty"`
	HSN      HSNWrapper      `json:"hsn,omitempty"`
	DocIssue DocIssueWrapper `json:"doc_issue,omitempty"`
}

// B2B Section
type GSTR1B2B struct {
	Ctin string       `json:"ctin"`
	Inv  []GSTR1B2BInv `json:"inv"`
}

type GSTR1B2BInv struct {
	Inum   string         `json:"inum"`
	Idt    string         `json:"idt"`
	Val    float64        `json:"val"`
	Pos    string         `json:"pos"`
	Rchrg  string         `json:"rchrg"` // "N"
	Etin   string         `json:"etin"`
	InvTyp string         `json:"inv_typ"` // "R"
	Itms   []GSTR1TaxItem `json:"itms"`
}

type GSTR1TaxItem struct {
	Num    int             `json:"num"`
	ItmDet GSTR1TaxDetails `json:"itm_det"`
}

type GSTR1TaxDetails struct {
	Rt    float64 `json:"rt"`
	TxVal float64 `json:"txval"`
	Iamt  float64 `json:"iamt"`
	Camt  float64 `json:"camt"`
	Samt  float64 `json:"samt"`
}

// CDNR Section
type GSTR1CDNR struct {
	Ctin string       `json:"ctin"`
	Nt   []GSTR1CDNRNt `json:"nt"`
}

type GSTR1CDNRNt struct {
	Ntty   string         `json:"ntty"` // "C" or "D"
	NtNum  string         `json:"nt_num"`
	NtDt   string         `json:"nt_dt"`
	Inum   string         `json:"inum"`
	Idt    string         `json:"idt"`
	Val    float64        `json:"val"`
	Pos    string         `json:"pos"`
	Itms   []GSTR1TaxItem `json:"itms"`
}

// B2CSRow represents the consolidated B2C Small row in GSTR-1.
type B2CSRow struct {
	SplyTy string  `json:"sply_ty"`
	POS    string  `json:"pos"`
	Rt     float64 `json:"rt"`
	TxVal  float64 `json:"txval"`
	Iamt   float64 `json:"iamt"`
	Camt   float64 `json:"camt"`
	Samt   float64 `json:"samt"`
	Typ    string  `json:"typ"` // "OE" (Other than E-commerce) or "E"
}

// HSNWrapper encapsulates the HSN data array.
type HSNWrapper struct {
	Data []HSNRow `json:"data"`
}

// HSNRow represents rate-wise and code-wise supply summary in Table 12.
type HSNRow struct {
	Num   int     `json:"num"`
	HsnSc string  `json:"hsn_sc"`
	Desc  string  `json:"desc"`
	Uqc   string  `json:"uqc"`
	Qty   float64 `json:"qty"`
	Val   float64 `json:"val"`
	TxVal float64 `json:"txval"`
	Iamt  float64 `json:"iamt"`
	Camt  float64 `json:"camt"`
	Samt  float64 `json:"samt"`
}

// DocIssueWrapper encapsulates the doc_det array for Table 13.
type DocIssueWrapper struct {
	DocDet []DocCategory `json:"doc_det"`
}

// DocCategory represents a category of documents issued (e.g. Invoices).
type DocCategory struct {
	DocNum int        `json:"doc_num"`
	Docs   []DocRange `json:"docs"`
}

// DocRange represents the sequence and counts of a document series.
type DocRange struct {
	Num      int    `json:"num"`
	From     string `json:"from"`
	To       string `json:"to"`
	TotNum   int    `json:"totnum"`
	Cancel   int    `json:"cancel"`
	NetIssue int    `json:"net_issue"`
}

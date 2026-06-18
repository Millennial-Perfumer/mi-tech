package dto

// GSTR1Payload represents the root GSTR-1 offline utility JSON schema.
type GSTR1Payload struct {
	GSTIN     string          `json:"gstin"`
	FP        string          `json:"fp"`
	FilingTyp string          `json:"filing_typ,omitempty"` // "M" (Monthly) or "Q" (Quarterly)
	GT        float64         `json:"gt"`
	CurGT     float64         `json:"cur_gt"`
	Version   string          `json:"version"`
	Hash      string          `json:"hash"`
	B2B       []GSTR1B2B      `json:"b2b,omitempty"`
	B2CS      []B2CSRow       `json:"b2cs,omitempty"`
	CDNR      []GSTR1CDNR     `json:"cdnr,omitempty"`
	HSN       HSNWrapper      `json:"hsn,omitempty"`
	DocIssue  DocIssueWrapper `json:"doc_issue,omitempty"`
	FilDt     string          `json:"fil_dt,omitempty"` // Filing Date (e.g. "12-05-2026")
}

// GSTR1B2B represents B2B supplies section.
type GSTR1B2B struct {
	Ctin string        `json:"ctin"`
	Cfs  string        `json:"cfs,omitempty"` // Consent for sharing: "N" or "Y"
	Inv  []GSTR1B2BInv `json:"inv"`
}

// GSTR1B2BInv represents a B2B invoice entry.
type GSTR1B2BInv struct {
	Inum   string         `json:"inum"`
	Idt    string         `json:"idt"`
	Val    float64        `json:"val"`
	Pos    string         `json:"pos"`
	Rchrg  string         `json:"rchrg"`            // "N"
	Etin   string         `json:"etin"`             // E-commerce Operator GSTIN
	InvTyp string         `json:"inv_typ"`          // "R" (Regular), "DE" (Deemed Exports), etc.
	Itms   []GSTR1TaxItem `json:"itms"`             // Tax rate-wise items
	Flag   string         `json:"flag,omitempty"`   // Utility upload state flag: "U", "N"
	Updby  string         `json:"updby,omitempty"`  // Updated by: "S" (Seller), "R" (Recipient)
	Cflag  string         `json:"cflag,omitempty"`  // Amendment flag
	Chksum string         `json:"chksum,omitempty"` // Checksum hash
}

// GSTR1TaxItem represents a rate-specific tax line item.
type GSTR1TaxItem struct {
	Num    int             `json:"num"` // rate code index (e.g. rate * 100 like 1800) or sequence
	ItmDet GSTR1TaxDetails `json:"itm_det"`
}

// GSTR1TaxDetails represents tax amounts.
type GSTR1TaxDetails struct {
	Rt    float64 `json:"rt"` // Tax rate (e.g. 18.0)
	TxVal float64 `json:"txval"`
	Iamt  float64 `json:"iamt"`
	Camt  float64 `json:"camt"`
	Samt  float64 `json:"samt"`
}

// GSTR1CDNR represents Credit/Debit Notes for registered users.
type GSTR1CDNR struct {
	Ctin string        `json:"ctin"`
	Nt   []GSTR1CDNRNt `json:"nt"`
}

// GSTR1CDNRNt represents a Credit/Debit Note entry.
type GSTR1CDNRNt struct {
	Ntty   string         `json:"ntty"` // "C" or "D"
	NtNum  string         `json:"nt_num"`
	NtDt   string         `json:"nt_dt"`
	Inum   string         `json:"inum"`
	Idt    string         `json:"idt"`
	Val    float64        `json:"val"`
	Pos    string         `json:"pos"`
	Itms   []GSTR1TaxItem `json:"itms"`
	Flag   string         `json:"flag,omitempty"`
	Updby  string         `json:"updby,omitempty"`
	Cflag  string         `json:"cflag,omitempty"`
	Chksum string         `json:"chksum,omitempty"`
}

// B2CSRow represents the consolidated B2C Small row in GSTR-1.
type B2CSRow struct {
	SplyTy string  `json:"sply_ty"` // "INTER" or "INTRA"
	POS    string  `json:"pos"`
	Rt     float64 `json:"rt"`
	TxVal  float64 `json:"txval"`
	Iamt   float64 `json:"iamt"`
	Camt   float64 `json:"camt"`
	Samt   float64 `json:"samt"`
	Typ    string  `json:"typ"`            // "OE" or "E"
	Flag   string  `json:"flag,omitempty"` // "N"
	Chksum string  `json:"chksum,omitempty"`
}

// HSNWrapper encapsulates the bifurcated HSN lists.
type HSNWrapper struct {
	Flag   string   `json:"flag"`             // "N"
	HsnB2B []HSNRow `json:"hsn_b2b"`          // Table 12 B2B supplies
	HsnB2C []HSNRow `json:"hsn_b2c"`          // Table 12 B2C supplies
	Chksum string   `json:"chksum,omitempty"` // Checksum
}

// HSNRow represents Table 12 rate-wise summary of outward supplies.
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
	IsB2B bool    `json:"-"` // Internal use, excluded from JSON serialization
}

// DocIssueWrapper encapsulates documents issued.
type DocIssueWrapper struct {
	Flag   string        `json:"flag"` // "N"
	DocDet []DocCategory `json:"doc_det"`
	Chksum string        `json:"chksum,omitempty"` // Checksum
}

// DocCategory represents a category of documents.
type DocCategory struct {
	DocNum int        `json:"doc_num"`
	Docs   []DocRange `json:"docs"`
}

// DocRange represents serial ranges for documents.
type DocRange struct {
	Num      int    `json:"num"`
	From     string `json:"from"`
	To       string `json:"to"`
	TotNum   int    `json:"totnum"`
	Cancel   int    `json:"cancel"`
	NetIssue int    `json:"net_issue"`
}

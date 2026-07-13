package models

type PetitionRequest struct {
	PetitionType    string `json:"petition_type" binding:"required"`
	ClientName      string `json:"client_name" binding:"required"`
	Opponent        string `json:"opponent"`
	Court           string `json:"court"`
	CaseFacts       string `json:"case_facts" binding:"required"`
	ReliefSought    string `json:"relief_sought"`
	AdditionalNotes string `json:"additional_notes"`
}

type AgreementRequest struct {
	AgreementType  string `json:"agreement_type" binding:"required"`
	PartyA         string `json:"party_a" binding:"required"`
	PartyB         string `json:"party_b" binding:"required"`
	Terms          string `json:"terms"`
	Duration       string `json:"duration"`
	Payment        string `json:"payment"`
	SpecialClauses string `json:"special_clauses"`
}

type LegalNoticeRequest struct {
	Sender      string `json:"sender" binding:"required"`
	Receiver    string `json:"receiver" binding:"required"`
	NoticeType  string `json:"notice_type" binding:"required"`
	Reason      string `json:"reason"`
	Facts       string `json:"facts"`
	LegalDemand string `json:"legal_demand"`
	TimeLimit   string `json:"time_limit"`
}

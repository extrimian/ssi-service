package credentials_bbs

type PackDIDCommRequest struct {
	DID       string             `json:"did,omitempty"`
	Message   PackDidCommMessage `json:"message,omitempty"`
	TargetDID string             `json:"targetDID,omitempty"`
}

type PackDidCommMessage struct {
	Id          string                 `json:"id,omitempty"`
	Type        string                 `json:"type,omitempty"`
	From        string                 `json:"from,omitempty"`
	To          []string               `json:"to,omitempty"`
	CreatedTime int                    `json:"created_time,omitempty"`
	ExpiresTime int                    `json:"expires_time,omitempty"`
	Body        PackDidCommMessageBody `json:"body,omitempty"`
}

type PackDidCommMessageBody struct {
	MessageTypeSpecificAttribute string `json:"message_type_specific_attribute,omitempty"`
	AnotherAttribute             string `json:"another_attribute,omitempty"`
}

type PackDIDCommResponse struct {
	PackedMessage PackedMessage `json:"packedMessage,omitempty"`
}

type PackedMessage struct {
	Protected  string      `json:"protected,omitempty"`
	Iv         string      `json:"iv,omitempty"`
	Ciphertext string      `json:"ciphertext,omitempty"`
	Tag        string      `json:"tag,omitempty"`
	Recipients []Recipient `json:"recipients,omitempty"`
}

type Recipient struct {
	EncriptedKey string `json:"encripted_key,omitempty"`
	Header       Header `json:"header,omitempty"`
}

type Header struct {
	Alg string `json:"alg,omitempty"`
	Iv  string `json:"iv,omitempty"`
	Tag string `json:"tag,omitempty"`
	Epk Epk    `json:"epk,omitempty"`
	Kid string `json:"kid,omitempty"`
}

type Epk struct {
	Kty string `json:"kty,omitempty"`
	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
}

type Styles struct {
	Background struct {
		Color string `json:"color"`
	} `json:"background"`
	Thumbnail struct {
		URI string `json:"uri"`
		Alt string `json:"alt"`
	} `json:"thumbnail"`
	Hero struct {
		URI string `json:"uri"`
		Alt string `json:"alt"`
	} `json:"hero"`
	Text struct {
		Color string `json:"color"`
	} `json:"text"`
}

type Display struct {
	Title       DisplayField   `json:"title"`
	Subtitle    DisplayField   `json:"subtitle"`
	Description DisplayField   `json:"description"`
	Properties  []DisplayField `json:"properties"`
	Styles      Styles         `json:"styles"`
}

type DisplayField struct {
	Path   []string `json:"path,omitempty"`
	Schema struct {
		Type string `json:"type,omitempty"`
	} `json:"schema,omitempty"`
	Fallback string `json:"fallback,omitempty"`
	Label    string `json:"label,omitempty"`
	Text     string `json:"text,omitempty"`
}

type VerifiableCredential struct {
	Context       []string          `json:"@context"`
	ID            string            `json:"id"`
	Type          []string          `json:"type"`
	Issuer        string            `json:"issuer"`
	IssuanceDate  string            `json:"issuanceDate"`
	CredentialSub CredentialSubject `json:"credentialSubject"`
}

type SignedVerifiableCredential struct {
	Context       []string          `json:"@context"`
	ID            string            `json:"id"`
	Type          []string          `json:"type"`
	Issuer        string            `json:"issuer"`
	IssuanceDate  string            `json:"issuanceDate"`
	CredentialSub CredentialSubject `json:"credentialSubject"`
	Proof         Proof             `json:"proof"`
}

type CredentialSubject struct {
	ID         string `json:"id"`
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

type OutputDescriptor struct {
	ID      string  `json:"id"`
	Schema  string  `json:"schema"`
	Display Display `json:"display"`
	Styles  Styles  `json:"styles"`
}

type Issuer struct {
	Name   string `json:"name"`
	Styles Styles `json:"styles"`
}

type CredentialRequest struct {
	DID                  string               `json:"did"`
	VerifiableCredential VerifiableCredential `json:"vc"`
	OutputDescriptor     OutputDescriptor     `json:"outputDescriptor"`
	Issuer               Issuer               `json:"issuer"`
}

type Proof struct {
	Type               string `json:"type"`
	Created            string `json:"created"`
	ProofPurpose       string `json:"proofPurpose"`
	ProofValue         string `json:"proofValue"`
	VerificationMethod string `json:"verificationMethod"`
}

type SignedCredential struct {
	VerifiableCredential SignedVerifiableCredential `json:"vc"`
	OutputDescriptor     OutputDescriptor           `json:"outputDescriptor"`
	Issuer               Issuer                     `json:"issuer"`
}

type VerifyResponse struct {
	Verified bool `json:"verified"`
}

type VerifiableCredentialArray struct {
	ID                   string                     `json:"id"`
	VerifiableCredential SignedVerifiableCredential `json:"vc"`
}

type VerifiableCredentialWithRenderArray struct {
	ID   string `json:"id"`
	Data struct {
		Styles               Styles                     `json:"styles"`
		Display              Display                    `json:"display"`
		VerifiableCredential SignedVerifiableCredential `json:"vc"`
	} `json:"data"`
}

type OOBRequest struct {
	DID                  string               `json:"did"`
	To                   string               `json:"to"`
	VerifiableCredential VerifiableCredential `json:"vc"`
	OutputDescriptor     OutputDescriptor     `json:"outputDescriptor"`
	Issuer               Issuer               `json:"issuer"`
}

type OOBResponse struct {
	InvitationID string `json:"invitationId"`
	OOBContent   string `json:"oobContentData"`
}

type DIDCommMessage struct {
	DID     string `json:"did"`
	To      string `json:"to,omitempty"`
	Message string `json:"message"`
}

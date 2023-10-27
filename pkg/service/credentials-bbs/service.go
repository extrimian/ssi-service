package credentials_bbs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/extrimian/ssi-service/pkg/service/framework"
)

func GetTBDPlusApiUrl() string {
	url := os.Getenv("TBD_PLUS_API")
	return url
}

type Service struct {
}

func (s *Service) Type() framework.Type {
	return framework.CredentialsBBS
}

func (s *Service) Status() framework.Status {
	return framework.Status{Status: framework.StatusReady}
}

func NewCredentialsBBSService() (*Service, error) {
	return &Service{}, nil
}

func (s *Service) PackDIDComm(req PackDIDCommRequest) (PackDIDCommResponse, error) {
	json_data, err := json.Marshal(req)
	if err != nil {
		fmt.Println("error marshaling didcomm pack request")
		return PackDIDCommResponse{}, err
	}
	putReq, err := http.NewRequest(http.MethodPut, GetTBDPlusApiUrl()+"/didcomm/pack", bytes.NewBuffer(json_data))
	if err != nil {
		fmt.Println("error creating put request")
		return PackDIDCommResponse{}, err
	}

	putReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(putReq)
	if err != nil {
		fmt.Println("error sending put request")
		return PackDIDCommResponse{}, err
	}

	defer resp.Body.Close()

	var response PackDIDCommResponse
	json.NewDecoder(resp.Body).Decode(&response)

	return response, nil
}

func (s *Service) CreateCredential(vc CredentialRequest) (SignedCredential, error) {
	json_data, err := json.Marshal(vc)
	if err != nil {
		fmt.Println("error marshaling credential request")
		return SignedCredential{}, err
	}

	var response SignedCredential
	resp, err := http.Post(GetTBDPlusApiUrl()+"/credentials-bbs", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		fmt.Println("error posting create credential request")
		return SignedCredential{}, err
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&response)

	return response, nil
}

func (s *Service) VerifyCredential(vc SignedCredential) (VerifyResponse, error) {
	json_data, err := json.Marshal(vc)
	if err != nil {
		fmt.Println("error marshaling credential request")
		return VerifyResponse{}, err
	}

	putReq, err := http.NewRequest(http.MethodPut, GetTBDPlusApiUrl()+"/credentials-bbs/verify", bytes.NewBuffer(json_data))
	if err != nil {
		fmt.Println("error creating put request")
		return VerifyResponse{}, err
	}

	putReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(putReq)
	if err != nil {
		fmt.Println("error sending put request")
		return VerifyResponse{}, err
	}

	var response VerifyResponse

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&response)

	return response, nil
}

func (s *Service) CreateOOBFromVC(req OOBRequest) (OOBResponse, error) {
	json_data, err := json.Marshal(req)
	if err != nil {
		fmt.Println("error marshaling oob request")
		return OOBResponse{}, err
	}

	var response OOBResponse
	resp, err := http.Post(GetTBDPlusApiUrl()+"/credentials-bbs/waci/oob", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		fmt.Println("error posting create oob request")
		return OOBResponse{}, err
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&response)

	return response, nil
}

func (s *Service) ListCredentials(did string) ([]VerifiableCredentialArray, error) {
	var response []VerifiableCredentialArray
	resp, err := http.Get(fmt.Sprintf(GetTBDPlusApiUrl()+"/credentials-bbs?did=%s", did))
	if err != nil {
		fmt.Println("error getting list of credentials request")
		return []VerifiableCredentialArray{}, err
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&response)

	return response, nil
}

func (s *Service) ListCredentialsWithRender(did string) ([]VerifiableCredentialWithRenderArray, error) {
	var response []VerifiableCredentialWithRenderArray
	resp, err := http.Get(fmt.Sprintf(GetTBDPlusApiUrl()+"/credentials-bbs/with-render-info?did=%s", did))
	if err != nil {
		fmt.Println("error getting list of credentials with render request")
		return []VerifiableCredentialWithRenderArray{}, err
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&response)

	return response, nil
}

func (s *Service) ProcessMessage(req DIDCommMessage) error {
	json_data, err := json.Marshal(req)
	if err != nil {
		fmt.Println("error marshaling message request")
		return err
	}

	putReq, err := http.NewRequest(http.MethodPut, GetTBDPlusApiUrl()+"/credentials-bbs/process-message", bytes.NewBuffer(json_data))
	if err != nil {
		fmt.Println("error creating put request")
		return err
	}

	putReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, err = client.Do(putReq)
	if err != nil {
		fmt.Println("error sending put request")
		return err
	}

	return nil
}

func (s *Service) Messaging(req DIDCommMessage) error {
	json_data, err := json.Marshal(req)
	if err != nil {
		fmt.Println("error marshaling oob request")
		return err
	}

	_, err = http.Post(GetTBDPlusApiUrl()+"/messaging", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		fmt.Println("error posting create oob request")
		return err
	}

	return nil
}

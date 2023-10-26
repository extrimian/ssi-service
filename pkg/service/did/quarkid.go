package did

import (
	"context"

	"github.com/extrimian/ssi-sdk/did"
	"github.com/extrimian/ssi-service/pkg/service/common"
)

type quarkidHandler struct {
	method did.Method
}

func NewQuarkidHandler() (MethodHandler, error) {
	return &quarkidHandler{method: did.QuarkidMethod}, nil
}

func (h *quarkidHandler) GetMethod() did.Method {
	return h.method
}

func (h *quarkidHandler) CreateDID(context.Context, CreateDIDRequest) (*CreateDIDResponse, error) {
	return nil, nil
}

func (h *quarkidHandler) GetDID(context.Context, GetDIDRequest) (*GetDIDResponse, error) {
	return nil, nil
}

func (h *quarkidHandler) ListDIDs(context.Context, *common.Page) (*ListDIDsResponse, error) {
	return nil, nil
}

func (h *quarkidHandler) ListDeletedDIDs(context.Context) (*ListDIDsResponse, error) {
	return nil, nil
}

func (h *quarkidHandler) SoftDeleteDID(context.Context, DeleteDIDRequest) error {
	return nil
}

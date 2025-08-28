package google

import (
	"context"

	"google.golang.org/api/forms/v1"
	"google.golang.org/api/option"
)

type FormsClient interface {
	GetForm(ctx context.Context, formID string) (*forms.Form, error)
}

type RealFormsClient struct{ svc *forms.Service }

func NewRealFormsClient(ctx context.Context, saJSONPath string) (*RealFormsClient, error) {
	svc, err := forms.NewService(ctx, option.WithCredentialsFile(saJSONPath))
	if err != nil {
		return nil, err
	}
	return &RealFormsClient{svc: svc}, nil
}

func (c *RealFormsClient) GetForm(ctx context.Context, formID string) (*forms.Form, error) {
	return c.svc.Forms.Get(formID).Context(ctx).Do()
}

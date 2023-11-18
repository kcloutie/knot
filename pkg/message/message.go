package message

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

type NotificationData struct {
	Data       map[string]interface{}
	Attributes map[string]string
	ID         string
}

func (n NotificationData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"data":       n.Data,
		"attributes": n.Attributes,
		"id":         n.ID,
	}
}

func GetCelDecl() cel.EnvOption {
	return cel.Declarations(
		decls.NewVar("data", decls.NewMapType(decls.String, decls.Dyn)),
		decls.NewVar("attributes", decls.NewMapType(decls.String, decls.String)),
		decls.NewVar("id", decls.String),
	)
}

package api

import (
	"github.com/getseabird/seabird/internal/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Encoder struct {
	*runtime.Scheme
}

func (s *Encoder) Encode(object client.Object) ([]byte, error) {
	codec := unstructured.NewJSONFallbackEncoder(serializer.NewCodecFactory(s.Scheme).LegacyCodec(s.PreferredVersionAllGroups()...))
	objWithoutManagedFields := object.DeepCopyObject().(client.Object)
	objWithoutManagedFields.SetManagedFields(nil)
	return runtime.Encode(codec, objWithoutManagedFields)
}

func (s *Encoder) EncodeYAML(object client.Object) ([]byte, error) {
	json, err := s.Encode(object)
	if err != nil {
		return nil, err
	}
	return util.JsonToYaml(json)
}

package traitdefinition

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1alpha2"
)

var handler ValidatingHandler
var req admission.Request
var reqResource metav1.GroupVersionResource
var decoder *admission.Decoder
var td v1alpha2.TraitDefinition
var tdRaw []byte
var scheme = runtime.NewScheme()

func TestTraitdefinition(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Traitdefinition Suite")
}

var _ = BeforeSuite(func(done Done) {
	td = v1alpha2.TraitDefinition{}
	td.SetGroupVersionKind(v1alpha2.TraitDefinitionGroupVersionKind)
	tdRaw, _ = json.Marshal(td)

	var err error
	decoder, err = admission.NewDecoder(scheme)
	Expect(err).Should(BeNil())

	close(done)
})

var _ = Describe("Test TraitDefinition validating handler", func() {
	BeforeEach(func() {
		reqResource = metav1.GroupVersionResource{
			Group:    v1alpha2.Group,
			Version:  v1alpha2.Version,
			Resource: "traitdefinitions"}
		handler = ValidatingHandler{}
		handler.InjectDecoder(decoder)
	})

	It("Test wrong resource of admission request", func() {
		wrongReqResource := metav1.GroupVersionResource{
			Group:    v1alpha2.Group,
			Version:  v1alpha2.Version,
			Resource: "foos"}
		req = admission.Request{
			AdmissionRequest: admissionv1beta1.AdmissionRequest{
				Operation: admissionv1beta1.Create,
				Resource:  wrongReqResource,
				Object:    runtime.RawExtension{Raw: []byte("")},
			},
		}
		resp := handler.Handle(context.TODO(), req)
		Expect(resp.Allowed).Should(BeFalse())
	})

	It("Test bad admission request", func() {
		req = admission.Request{
			AdmissionRequest: admissionv1beta1.AdmissionRequest{
				Operation: admissionv1beta1.Create,
				Resource:  reqResource,
				Object:    runtime.RawExtension{Raw: []byte("bad request")},
			},
		}
		resp := handler.Handle(context.TODO(), req)
		Expect(resp.Allowed).Should(BeFalse())
	})

	Context("Test create/update operation admission request", func() {
		var mockValidator TraitDefValidatorFn
		It("Test validation passed", func() {
			// mock a validator that always validates successfully
			mockValidator = func(_ context.Context, _ v1alpha2.TraitDefinition) error {
				return nil
			}
			handler.Validators = []TraitDefValidator{
				TraitDefValidatorFn(mockValidator),
			}
			req = admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Operation: admissionv1beta1.Create,
					Resource:  reqResource,
					Object:    runtime.RawExtension{Raw: tdRaw},
				},
			}
			resp := handler.Handle(context.TODO(), req)
			Expect(resp.Allowed).Should(BeTrue())
		})
		It("Test validation failed", func() {
			// mock a validator that always failed
			mockValidator = func(_ context.Context, _ v1alpha2.TraitDefinition) error {
				return errors.New("mock validator error")
			}
			handler.Validators = []TraitDefValidator{
				TraitDefValidatorFn(mockValidator),
			}
			req = admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Operation: admissionv1beta1.Create,
					Resource:  reqResource,
					Object:    runtime.RawExtension{Raw: tdRaw},
				},
			}
			resp := handler.Handle(context.TODO(), req)
			Expect(resp.Allowed).Should(BeFalse())
			Expect(resp.Result.Reason).Should(Equal(metav1.StatusReason("mock validator error")))
		})
	})
})

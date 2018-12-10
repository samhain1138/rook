package webhook

import (
	"encoding/json"
	"github.com/coreos/pkg/capnslog"
	cassandrav1alpha1 "github.com/rook/rook/pkg/apis/cassandra.rook.io/v1alpha1"
	"io/ioutil"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"net/http"
)

var logger = capnslog.NewPackageLogger("github.com/rook/rook", "pkg/webhook")

// WebhookConfig represents the configuration for
// an admission webhook
type WebhookConfig struct {
	Port        int32
	TLSCertFile string
	TLSKeyFile  string
}

type WebhookServer struct {
	Server       *http.Server
	AdmissionFns AdmissionFns
}

type AdmissionFns interface {
	Validate(*admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse
}

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	_ = cassandrav1alpha1.AddToScheme(runtimeScheme)
}

func (s *WebhookServer) Serve(w http.ResponseWriter, r *http.Request) {

	var body []byte
	var err error

	if r.Body == nil {
		http.Error(w, "body is nil", http.StatusBadRequest)
		return
	}
	if body, err = ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, "error processing request", http.StatusBadRequest)
		logger.Errorf("Error reading request body: %v", err)
		return
	}
	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// Verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logger.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *admissionv1beta1.AdmissionResponse
	ar := admissionv1beta1.AdmissionReview{}

	if _, _, err = deserializer.Decode(body, nil, &ar); err != nil {
		logger.Errorf("Can't decode body: %v", err)
		admissionResponse = &admissionv1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		logger.Infof(r.URL.Path)
		if r.URL.Path == "/validate" {
			admissionResponse = s.AdmissionFns.Validate(ar.Request)
		} else {
			return
		}
	}

	ar.Response = admissionResponse
	resp, err := json.Marshal(ar)
	if err != nil {
		logger.Errorf("Can't encode response: %v", err)
		http.Error(w, "could not encode response", http.StatusInternalServerError)
		return
	}
	if _, err = w.Write(resp); err != nil {
		logger.Errorf("Can't write response: %v", err)
		http.Error(w, "could not write response", http.StatusInternalServerError)
	}
}

package template

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

func TestConstructorWebhookTemplate(t *testing.T) {
	executer, err := NewWebhookTemplate(zap.NewNop(), "https://flipt-webhook.io/webhook", `{"type": "{{ .Type }}", "action": "{{ .Action }}"}`, nil, 15*time.Second)
	require.NoError(t, err)
	require.NotNil(t, executer)

	template, ok := executer.(*webhookTemplate)
	require.True(t, ok, "executer should be a webhookTemplate")

	assert.Equal(t, "https://flipt-webhook.io/webhook", template.url)
	assert.Nil(t, template.headers)
}

func TestExecuter_JSON_Failure(t *testing.T) {
	tmpl, err := template.New("").Parse(`this is invalid JSON {{ .Type }}, {{ .Action }}`)
	require.NoError(t, err)

	template := &webhookTemplate{
		url:          "https://flipt-webhook.io/webhook",
		bodyTemplate: tmpl,
	}

	err = template.Execute(context.TODO(), audit.Event{
		Type:   string(flipt.SubjectFlag),
		Action: string(flipt.ActionCreate),
	})

	assert.EqualError(t, err, "invalid JSON: this is invalid JSON flag, create")
}

func TestExecuter_Execute(t *testing.T) {
	tmpl, err := template.New("").Parse(`{
		"type": "{{ .Type }}",
		"action": "{{ .Action }}"
	}`)

	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	t.Cleanup(ts.Close)

	template := &webhookTemplate{
		url:          ts.URL,
		bodyTemplate: tmpl,
		httpClient:   retryablehttp.NewClient(),
	}

	err = template.Execute(context.TODO(), audit.Event{
		Type:   string(flipt.SubjectFlag),
		Action: string(flipt.ActionCreate),
	})

	require.NoError(t, err)
}

func TestExecuter_Execute_toJson_valid_Json(t *testing.T) {
	tmpl, err := template.New("").Funcs(funcMap).Parse(`{
		"type": "{{ .Type }}",
		"action": "{{ .Action }}",
		"payload": {{ toJson .Payload }}
	}`)

	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	t.Cleanup(ts.Close)

	template := &webhookTemplate{
		url:          ts.URL,
		bodyTemplate: tmpl,
		httpClient:   retryablehttp.NewClient(),
	}

	err = template.Execute(context.TODO(), audit.Event{
		Type:   string(flipt.SubjectFlag),
		Action: string(flipt.ActionCreate),
		Payload: &flipt.CreateFlagRequest{
			Key:          "foo",
			Name:         "foo",
			NamespaceKey: "default",
			Description:  "some desc",
			Enabled:      true,
		},
	})

	require.NoError(t, err)
}

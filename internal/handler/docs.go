package handler

import (
	"log/slog"
	"net/http"

	"github.com/iavianm/books-api/docs"
)

const swaggerUIPage = `<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>books-api — API docs</title>
	<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
	<script>
		window.onload = () => {
			SwaggerUIBundle({
				url: "/openapi.yaml",
				dom_id: "#swagger-ui",
			});
		};
	</script>
</body>
</html>
`

func (h *Handler) docsPage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(swaggerUIPage)); err != nil {
		slog.Error("write docs page", "err", err)
	}
}

func (h *Handler) openAPISpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	if _, err := w.Write(docs.OpenAPISpec); err != nil {
		slog.Error("write openapi spec", "err", err)
	}
}

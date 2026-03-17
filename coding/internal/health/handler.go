package health

import (
	"net/http"

	"github.com/R4x4nJ031/project-man/internal/response"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

package handlers

import (
	"net/http"

	"github.com/windevkay/flho/pkg/utils"
)

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := utils.Envelope{
		"status": "available",
	}

	utils.WriteJSON(w, http.StatusOK, env, nil)
}

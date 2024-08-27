package main

import (
	"net/http"

	"github.com/windevkay/flhoutils/helpers"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := helpers.Envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	helpers.WriteJSON(w, http.StatusOK, env, nil)
}
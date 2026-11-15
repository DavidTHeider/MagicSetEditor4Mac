package rest_test

// **Die Wege der eigenen Übersetzungen** — LOC-02 (#1688).

import "net/http"

func endpunkteSprache() []endpunkt {
	return []endpunkt{
		{name: "Sprachpakete", methode: http.MethodGet, weg: "/sprachpakete", erfolg: http.StatusOK},
		{name: "Sprachpaket", methode: http.MethodGet, weg: "/sprachpakete/da", erfolg: http.StatusOK},
		{
			name: "Sprachpaket anlegen", methode: http.MethodPost, weg: "/sprachpakete",
			rumpf: `{"kennung":"da","name":"Dansk","basis":"en"}`, erfolg: http.StatusCreated,
		},
		{
			name: "Sprachpaket ändern", methode: http.MethodPatch, weg: "/sprachpakete/da",
			rumpf: `{"stand":"geprueft","texte":{"allgemein.titel":"Sæt"}}`, erfolg: http.StatusOK,
		},
		{name: "Sprachpaket löschen", methode: http.MethodDelete, weg: "/sprachpakete/da", erfolg: http.StatusOK},
		{name: "Sprachpaket als Datei", methode: http.MethodGet, weg: "/sprachpakete/da/datei", erfolg: http.StatusOK},
		{
			name: "Sprachpaket einlesen", methode: http.MethodPost, weg: "/sprachpakete/einlesen",
			rumpf: `{"inhalt":"{}","ersetzen":true}`, erfolg: http.StatusCreated,
		},
	}
}

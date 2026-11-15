// **Die Wege der Paketpflege** — einbinden, anlegen, kopieren, übernehmen,
// entfernen (#1291, #1350, FMT-02 #1691).
//
// **Eine eigene Datei**, weil `endpunkteneben_test.go` mit der eigenen Kopie
// über die Zeilengrenze aus §2 ging. Sie hängen an keinem Set: ein Paket ist
// eine Vorlage und kein Entwurf.

package rest_test

import "net/http"

// endpunktePaketpflege sind die Wege, die Pakete auf die Platte bringen oder
// von ihr nehmen.
func endpunktePaketpflege() []endpunkt {
	return []endpunkt{
		{
			name: "Paket einbinden", methode: http.MethodPut, weg: "/pakete/eingebunden",
			rumpf: `{"quelle":"/tmp/magic-m15.mse-style"}`, erfolg: http.StatusCreated,
		},
		{
			// **`POST /pakete/neu` und nicht `POST /pakete`** — jener
			// **bezieht** aus dem Netz, dieser **schreibt** eines auf die
			// Platte (#1350).
			name: "Paket anlegen", methode: http.MethodPost, weg: "/pakete/neu",
			rumpf: `{"art":"mse-game","name":"neu"}`, erfolg: http.StatusCreated,
		},
		{
			// **Eine eigene Kopie** (FMT-02, #1691) — am Paket, das kopiert
			// wird; ohne Rumpf heißt sie wie das Original mit »-kopie«.
			name: "Paket kopieren", methode: http.MethodPost, weg: "/pakete/mse-style/probe-schlicht/kopie",
			rumpf: `{"neu":"meins"}`, erfolg: http.StatusCreated,
		},
		{
			// **Der Dateibaum** (FMT-02, #1691) — anlegen, umbenennen,
			// entfernen; der Stand beim Entfernen steht in der Abfrage.
			// **Einteilige Namen**: der Vergleicher in `vollstaendig_test.go`
			// zählt `{datei...}` als ein Glied.
			name: "Paketdatei anlegen", methode: http.MethodPost,
			weg:   "/pakete/probe.mse-style/dateien",
			rumpf: `{"name":"skripte/neu","inhalt":"# neu"}`, erfolg: http.StatusCreated,
		},
		{
			name: "Paketdatei umbenennen", methode: http.MethodPatch,
			weg:   "/pakete/probe.mse-style/dateien/neu",
			rumpf: `{"neu":"skripte/alt","stand":"s"}`, erfolg: http.StatusOK,
		},
		{
			name: "Paketdatei entfernen", methode: http.MethodDelete,
			weg: "/pakete/probe.mse-style/dateien/alt?stand=s", erfolg: http.StatusOK,
		},
		{
			name: "Pakete übernehmen", methode: http.MethodPost, weg: "/pakete/uebernommen",
			rumpf: `{"ort":"/tmp/mse2/data"}`, erfolg: http.StatusCreated,
		},
		{
			// **Als `.mse-installer`** (FMT-02.AK-04, #1691) — hinaus die
			// genannten Pakete, herein alles, was der Installer nennt.
			name: "Pakete weitergeben", methode: http.MethodPost, weg: "/pakete/weitergabe",
			rumpf: `{"pakete":["probe.mse-game"],"ziel":"/tmp/probe.mse-installer"}`, erfolg: http.StatusCreated,
		},
		{
			name: "Installer einlesen", methode: http.MethodPost, weg: "/pakete/eingelesen",
			rumpf: `{"datei":"/tmp/probe.mse-installer"}`, erfolg: http.StatusCreated,
		},
		{
			// **Der Pfad in der Abfrage ist die Bestätigung** — wie der Name
			// beim Löschen eines Sets (#1218).
			name: "Paket entfernen", methode: http.MethodDelete,
			weg: "/pakete/mse-style/probe?pfad=/pakete/probe.mse-style", erfolg: http.StatusOK,
		},
	}
}

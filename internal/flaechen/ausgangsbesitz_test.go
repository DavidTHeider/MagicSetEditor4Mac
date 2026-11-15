// **Eine Ausgabe holt nur ab, wer sie erzeugt hat** — SEC-P6
// (#2002),
// Befund P6 in `docs/PENTEST.md`.
//
// # Der Angriff
//
// Der Mensch legt über die Prozessmarke einen Bogen ohne Ziel an; er liegt
// unter `/ausgang/{name}`. Ein Automat mit einem **Token eines anderen
// Profils**, lesend oder schreibend, kennt den Namen — aus einem Protokoll,
// einer Weitergabe, einem Skript — und holt die Datei ab oder vergisst sie.
// Bis #2002 war der Name das einzige Recht.
//
// # Woran gemessen
//
// **An der Antwort und an der Datei danach**: der Fremde bekommt `404` wie
// für einen Namen, den es nicht gibt — ein `403` verriete, dass es ihn gibt
// —, und die Ausgabe liegt nach seinem Vergessen noch da. Die Gegenprobe:
// der Erzeuger holt sie ab, und danach ist sie fort.
package flaechen_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestEineAusgabeHoltNurAbWerSieErzeugtHat(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	status, _, rumpf := uebertrage(t, an, http.MethodPost, "/sets", "application/zip", gezipptesProbeset(t))

	var eingefahren struct {
		SetID string `json:"setID"`
	}

	if status != http.StatusCreated || json.Unmarshal(rumpf, &eingefahren) != nil {
		t.Fatalf("Einfahren: %d %s", status, rumpf)
	}

	status, kopf, rumpf := uebertrage(t, an, http.MethodPost,
		"/sets/"+eingefahren.SetID+"/bogen", "application/json", []byte(`{}`))
	ort := kopf.Get("Location")

	if status != http.StatusCreated || !strings.HasPrefix(ort, "/ausgang/") {
		t.Fatalf("Bogen: %d %v %s", status, kopf, rumpf)
	}

	anderes := profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", "1", "Nachtlauf", "--json"))
	token := func(umfang string) dienst {
		return an.mitNachweis(strings.TrimSpace(mitEingabe(t, umgebung,
			`{"principal":"Rechner","bezeichnung":"`+umfang+`","scope":"`+umfang+`"}`,
			"token-anlegen", anderes)))
	}

	leser, schreiber := token("lesen"), token("schreiben")

	for name, fremd := range map[string]dienst{"Lesetoken": leser, "Schreibtoken": schreiber} {
		if fremdstatus, _, fremdrumpf := uebertrage(t, fremd, http.MethodGet, ort, "", nil); fremdstatus != http.StatusNotFound {
			t.Errorf("ein fremdes %s holte die Ausgabe ab: %d — %.40q", name, fremdstatus, fremdrumpf)
		}
	}

	// **Vergessen braucht schreiben** — ein Lesetoken endet schon am
	// Tokenumfang; das Schreibtoken kommt bis zur Ausgabe und findet keine.
	if loeschstatus, _, _ := uebertrage(t, schreiber, http.MethodDelete, ort, "", nil); loeschstatus != http.StatusNotFound {
		t.Errorf("ein fremdes Schreibtoken vergaß die Ausgabe: %d", loeschstatus)
	}

	// **Die Gegenprobe**: der Erzeuger holt sie ab — sie liegt also noch da,
	// und die `404` darüber galt dem Rufer und nicht der Datei.
	status, _, rumpf = uebertrage(t, an, http.MethodGet, ort, "", nil)
	if status != http.StatusOK || !bytes.HasPrefix(rumpf, []byte("%PDF-")) {
		t.Errorf("der Erzeuger bekam %d — %.20q", status, rumpf)
	}

	if status, _, _ := uebertrage(t, an, http.MethodDelete, ort, "", nil); status != http.StatusNoContent {
		t.Errorf("der Erzeuger vergaß nicht: %d", status)
	}
}

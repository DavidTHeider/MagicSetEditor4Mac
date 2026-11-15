package flaechen_test

// **Ein Schritt nennt die Revision, die er meint — an jeder Fläche** —
// #2496. Kommandozeile mit `--gesehen=<abdruck>`, der Dienst mit dem Feld
// `gesehen`; ohne Angabe wird nicht gefragt (§16: erweiternd).

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

const derGeseheneHandel = "v-gesehen"

func TestEinSchrittMitFremdemAbdruckGehtAnKeinerFlaeche(t *testing.T) {
	umgebung := append(mitEigenerSaat(t),
		schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	ich := zeilenwert(t,
		mussKommandozeile(t, umgebung, "peer-anlegen"), "kennung")

	mitEingabe(t, umgebung, `[{"von":"`+ich+
		`","printing":"bolt-m10","menge":1,"guete":"nearMint",`+
		`"sprache":"de","finish":"nonfoil"}]`,
		"handel", "eroeffnen", derGeseheneHandel, "tausch", dasGegenueberImHandel)

	var stand struct {
		Abdruck string `json:"abdruck"`
	}

	if fehler := json.Unmarshal([]byte(mussKommandozeile(t, umgebung,
		"--json", "handel", "stand", derGeseheneHandel)), &stand); fehler != nil || stand.Abdruck == "" {
		t.Fatalf("der Stand: %v %+v", fehler, stand)
	}

	// **Die Kommandozeile**: ein fremder Abdruck ist ein Konflikt.
	aus, _, code := kommandozeile(t, umgebung, "--json", "handel", "anbieten",
		derGeseheneHandel, "--gesehen=ein-anderer")
	if code == 0 || !strings.Contains(aus, `"code":"konflikt"`) {
		t.Errorf("mit fremdem Abdruck: %d %q", code, aus)
	}

	// **Der Dienst**: dasselbe, als 409.
	weg := strings.ReplaceAll(pfadAus(derSchrittweg), "{vorgang}", derGeseheneHandel)
	if status, antwort := an.ruf(t, http.MethodPost, weg,
		`{"schritt":"anbieten","gesehen":"ein-anderer"}`); status != http.StatusConflict {
		t.Errorf("der Dienst mit fremdem Abdruck: %d %s", status, antwort)
	}

	// **Mit dem gesehenen geht es** — an der Kommandozeile das Anbieten,
	// am Dienst danach die Zustimmung der Gegenseite.
	mussKommandozeile(t, umgebung, "handel", "anbieten",
		derGeseheneHandel, "--gesehen="+stand.Abdruck)
	mussREST(t, an, http.MethodPost, weg,
		`{"schritt":"zustimmen","wer":"`+dasGegenueberImHandel+`","gesehen":"`+stand.Abdruck+`"}`)
}

// **Nach der Einigung kennt ein Schritt kein „gesehen"** — die Revision
// ändert sich dann nicht mehr; ein Abdruck dort wird abgewiesen statt still
// übergangen.
func TestEinAbschlussschrittWeistGesehenAb(t *testing.T) {
	umgebung := append(mitEigenerSaat(t),
		schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	weg := strings.ReplaceAll(pfadAus(derSchrittweg), "{vorgang}", derGeseheneHandel)
	status, antwort := an.ruf(t, http.MethodPost, weg,
		`{"schritt":"abgleichen","gesehen":"irgendeiner"}`)

	if status != http.StatusBadRequest || !strings.Contains(antwort, "gesehen") {
		t.Errorf("abgleichen mit Abdruck: %d %s", status, antwort)
	}
}

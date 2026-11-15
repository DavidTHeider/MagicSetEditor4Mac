package flaechen_test

// **Festgehalten ist festgehalten — an jeder Fläche** — §7, Regel 3
// (#1970).
//
// # Was hier gemessen wird
//
// Ein Auslöser legt hinter jede neue Zusage eine mit krummer Frist
// ([imBestand]); das Festhalten glückt, das Nachlesen der Liste scheitert.
// **Die Kommandozeile** endet mit 0 und sagt es auf dem Fehlerstrom — auch
// mit `--json`, denn ihre Ausgabe ist eine Liste. **Der Dienst** antwortet 201 —
// sein Rumpf ist eine Liste, und ein Feld darin gäbe es nur um den Preis
// einer anderen Gestalt; deshalb steht der Grund im Kopf `X-Nachladefehler`,
// so wie die Vorgangskennung in `X-Vorgang` (0033: erweiternd). MCP hält
// nichts fest und hat den Weg nicht.

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// derNachladehandel ist der Vorgang dieses Laufs.
const derNachladehandel = "v-nachladen"

func TestFestgehaltenIstFestgehaltenAnJederFlaeche(t *testing.T) {
	umgebung := append(mitEigenerSaat(t),
		schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	ich := zeilenwert(t,
		mussKommandozeile(t, umgebung, "peer-anlegen"), "kennung")
	// **In der Variante des Tauschs** (#2492) — ohne Angabe wäre es englisch.
	exemplarID := eineKennungAus(t, mussKommandozeile(t, umgebung,
		"--json", "exemplar", "bolt-m10", "3", "nearMint", "de", "nonfoil"))

	mitEingabe(t, umgebung, `[{"von":"`+ich+`","printing":"bolt-m10",`+
		`"menge":3,"guete":"nearMint","sprache":"de","finish":"nonfoil"}]`,
		"handel", "eroeffnen", derNachladehandel, "tausch", dasGegenueberImHandel)

	imBestand(t, umgebung, `
CREATE TRIGGER krumme_zusage AFTER INSERT ON reservierungen
WHEN NEW.frist <> 'krumm'
BEGIN
  INSERT INTO reservierungen
    (profil_id, vorgang, revision, exemplar_id, menge, fassung, frist)
  VALUES (NEW.profil_id, NEW.vorgang, NEW.revision, NEW.exemplar_id, 1,
          NEW.fassung, 'krumm');
END`)

	// **Die krumme Zeile geht nach jedem Ruf wieder** — das Festhalten
	// liest die Zusagen vorher, und an ihr scheiterte sonst schon der
	// nächste Schritt selbst.
	aufraeumen := func() {
		imBestand(t, umgebung, `DELETE FROM reservierungen WHERE frist = 'krumm'`)
	}

	if aus, fehlerstrom, code := kommandozeile(t, umgebung, "handel-festhalten",
		derNachladehandel, exemplarID, "1", "1"); code != 0 ||
		!strings.Contains(fehlerstrom, "geschrieben") {
		t.Fatalf("Kommandozeile: %d, %q, %q", code, aus, fehlerstrom)
	}

	aufraeumen()

	// **Mit `--json` ebenso** — die Ausgabe ist eine Liste, und sie bleibt
	// es; der Hinweis steht auf dem Fehlerstrom.
	if aus, fehlerstrom, code := kommandozeile(t, umgebung, "--json",
		"handel-festhalten", derNachladehandel, exemplarID, "1", "1"); code != 0 ||
		strings.TrimSpace(aus) != "[]" || !strings.Contains(fehlerstrom, "geschrieben") {
		t.Fatalf("Kommandozeile mit --json: %d, %q, %q", code, aus, fehlerstrom)
	}

	aufraeumen()

	status, koepfe := an.kopfruf(t, http.MethodPost,
		"/handel/"+derNachladehandel+"/zusagen",
		`{"exemplarId":`+exemplarID+`,"menge":1,"fassung":1}`)
	if status != http.StatusCreated || koepfe.Get("X-Nachladefehler") != "validierung" {
		t.Fatalf("Dienst: %d, %v", status, koepfe)
	}

	aufraeumen()
	imBestand(t, umgebung, `DROP TRIGGER krumme_zusage`)

	// **Und festgehalten ist alles drei.**
	if liste := mussKommandozeile(t, umgebung,
		"handelszusagen", derNachladehandel); strings.Count(liste, "\n") != 3 {
		t.Errorf("festgehalten sind: %q", liste)
	}
}

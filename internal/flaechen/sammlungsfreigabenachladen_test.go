package flaechen_test

// **Erteilt und widerrufen ist geschehen — an jeder Fläche** — §7, Regel 3
// (#1970).
//
// # Was hier gemessen wird
//
// Ein Auslöser legt nach dem Erteilen und nach dem Widerruf eine Freigabe
// mit einem Ablauf neben die geschriebene, der kein Zeitpunkt ist
// ([imBestand]). Geschrieben ist dann; die Liste danach scheitert. Die
// Kommandozeile endet mit 0 und sagt es, der Dienst antwortet 200 mit
// `nachladefehler`. MCP hat diese Wege nicht.

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// krummeSichtfreigabe ist der Auslöser — einer je Ereignis.
func krummeSichtfreigabe(ereignis, zeile string) string {
	return `CREATE TRIGGER krumme_freigabe_` + strings.ToLower(strings.Fields(ereignis)[1]) +
		` ` + ereignis + ` WHEN ` + zeile + `.ablauf <> 'krumm'
BEGIN
  INSERT INTO sammlungsfreigaben (profil_id, empfaenger, umfang, zweck, ablauf, erteilt)
  VALUES (` + zeile + `.profil_id, ` + zeile + `.empfaenger, ` + zeile + `.umfang,
          'teilen', 'krumm', ` + zeile + `.erteilt);
END`
}

func TestDieSichtfreigabeGiltAnJederFlaecheAuchOhneListe(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	kennung := string(fremd.Kennung())

	mussKommandozeile(t, umgebung, "kontakt-aufnehmen", kennung, "Annas Rechner")
	mussKommandozeile(t, umgebung, "kontakt-bestaetigen", kennung, fremd.Fingerabdruck())
	mussKommandozeile(t, umgebung, "zweck-erlauben", kennung, "handel")

	aufraeumen := func() {
		imBestand(t, umgebung, `DELETE FROM sammlungsfreigaben WHERE ablauf = 'krumm'`)
	}

	imBestand(t, umgebung, krummeSichtfreigabe(`AFTER INSERT ON sammlungsfreigaben`, "NEW"))

	if aus, fehlerstrom, code := kommandozeile(t, umgebung,
		"kontakt-freigeben", kennung, "handel", "sammlung"); code != 0 ||
		!strings.Contains(fehlerstrom, "geschrieben") {
		t.Fatalf("Kommandozeile: %d, %q, %q", code, aus, fehlerstrom)
	}

	aufraeumen()

	status, rumpf := an.ruf(t, http.MethodPost, "/peer/kontakte/"+kennung+"/freigaben",
		`{"zweck":"handel","umfang":"auswahl","auswahl":["bolt-m10"]}`)
	if dienst := nachgeladenAus(t, rumpf); status >= 300 ||
		dienst.Nachladefehler == nil || dienst.Nachladefehler.Code != "validierung" {
		t.Fatalf("Dienst: %d, %s", status, rumpf)
	}

	aufraeumen()
	imBestand(t, umgebung, `DROP TRIGGER krumme_freigabe_insert`)
	imBestand(t, umgebung, krummeSichtfreigabe(`AFTER DELETE ON sammlungsfreigaben`, "OLD"))

	aus, fehlerstrom, code := kommandozeile(t, umgebung, "--json",
		"kontakt-widerrufen", kennung, "handel", "sammlung")
	if zeile := nachgeladenAus(t, aus); code != 0 ||
		!strings.Contains(fehlerstrom, "geschrieben") || zeile.Nachladefehler == nil {
		t.Fatalf("Kommandozeile, Widerruf: %d, %s, %q", code, aus, fehlerstrom)
	}

	aufraeumen()

	status, rumpf = an.ruf(t, http.MethodDelete,
		"/peer/kontakte/"+kennung+"/freigaben/handel/auswahl", "")
	if dienst := nachgeladenAus(t, rumpf); status != http.StatusOK || dienst.Nachladefehler == nil {
		t.Fatalf("Dienst, Widerruf: %d, %s", status, rumpf)
	}

	aufraeumen()

	// **Und geschehen ist alles vier** — keine Freigabe an Anna steht mehr.
	if anzahl := ausDemBestand(t, umgebung,
		`SELECT CAST(count(*) AS TEXT) FROM sammlungsfreigaben`); anzahl != "0" {
		t.Errorf("es stehen %s Freigaben da", anzahl)
	}
}

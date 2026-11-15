package app_test

// **Die Freigabe für den Neuanfang lässt sich ausstellen** — G12
// (#2282).
//
// # Der Befund
//
// **Die ganze Anwendung ließ sich an keiner Fläche zurücksetzen.** Die Art
// `anwendung` verlangt einen Nachweis mit dem Zweck `ruecksetzung.anwendung`,
// und das Fenster fragte nach seiner Kennung — **ausgestellt hat ihn nichts**.
// Gelegt wurde er nur in Tests, unmittelbar in die Ablage
// (`ruecksetzungsnaht_test.go`): die Vorrichtung war großzügiger als der
// Betrieb (§3), dieselbe Lage wie #1539 bei den KI-Freigaben.
//
// # Was hier geprüft wird
//
// **Der Weg von der Freigabe bis zum beiseitegelegten Bestand** — und das
// Nein an jeder Fläche außer dem Fenster.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieFreigabeFuerDenNeuanfangLaesstSichAusstellen(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")

	db, fehler := storage.Oeffnen(probekontext(t), ort)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	ctx := probekontext(t)
	anwendung := app.RuecksetzungsAnwendungUeber(db, ort)

	// **In der Reihenfolge des Fensters**: erst nachsehen, dann freigeben,
	// dann ausführen — mit der Revision der Vorschau. Der erste Lauf am
	// nativen Fenster (2026-10-10) scheiterte genau so: die Freigabe ist eine
	// Zeile im Bestand, die Vorschau zählte Freigabenachweise mit, und ihre
	// Revision galt danach nicht mehr. Die Flächenprobe fährt die andere
	// Reihenfolge (`internal/flaechen/neuanfangsfreigabe_test.go`).
	vorschau := anwendung.Ruecksetzungsvorschau(ctx, "anwendung", "", "")
	if vorschau.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", vorschau.Fehler)
	}

	freigabe := anwendung.RuecksetzungFreigabeErteilen(ctx)
	if freigabe.Fehler != nil || freigabe.Kennung == "" {
		t.Fatalf("die Freigabe: %+v", freigabe)
	}

	if freigabe.Laeuftab == "" {
		t.Error("die Freigabe sagt nicht, bis wann sie gilt")
	}

	bilanz := anwendung.Zuruecksetzen(ctx, "anwendung", "", "", "", freigabe.Kennung,
		vorschau.Revision)
	if bilanz.Fehler != nil {
		t.Fatalf("zurücksetzen mit der ausgestellten Freigabe: %+v", bilanz.Fehler)
	}

	if _, err := os.Stat(bilanz.Beiseite); err != nil {
		t.Errorf("der alte Bestand liegt nicht bei %q: %v", bilanz.Beiseite, err)
	}
}

// **Über die Kommandozeile, MCP und den Dienst gibt es keine** — dieselbe
// Zusage wie bei Handel und KI (0078): `Nachweise.Ausstellen` weist jeden
// Modus außer `fenster` ab, und ein Modell ist kein Mensch.
func TestDieFreigabeFuerDenNeuanfangGibtEsNurAmFenster(t *testing.T) {
	anwendung := app.RuecksetzungsAnwendungUeber(frischerBestand(t),
		filepath.Join(t.TempDir(), "bestand.db"))

	for _, modus := range []domain.Dienstmodus{domain.ModusKommandozeile, domain.ModusMCP, domain.ModusDienst} {
		antwort := anwendung.RuecksetzungFreigabeErteilen(
			service.MitDienstmodus(probekontext(t), modus))

		switch {
		case antwort.Fehler == nil:
			t.Errorf("%s bekam eine Freigabe", modus)
		case antwort.Kennung != "":
			t.Errorf("%s: abgewiesen, und doch eine Kennung %q", modus, antwort.Kennung)
		case !strings.Contains(antwort.Fehler.Meldung, "Fenster"):
			t.Errorf("%s: die Absage nennt das Fenster nicht: %s", modus, antwort.Fehler.Meldung)
		}
	}
}

// **Ohne Bestand gibt es keine Freigabe** — und keine Kennung, die nach
// einer aussähe.
func TestOhneBestandGibtEsKeineFreigabeFuerDenNeuanfang(t *testing.T) {
	antwort := app.NeueApp("").RuecksetzungFreigabeErteilen(probekontext(t))
	if antwort.Fehler == nil || antwort.Kennung != "" {
		t.Errorf("ohne Bestand: %+v", antwort)
	}
}

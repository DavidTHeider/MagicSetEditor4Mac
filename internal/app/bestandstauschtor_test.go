//go:build darwin

// **Nach einem Tausch des Bestands geht das Tor wieder auf** — A02
// (#2216).
//
// # Was gemessen wurde
//
// **Am gebauten Fenster** (Befund F-01 des Audits zu 0.8.0): eine Sicherung
// zurückspielen meldet Erfolg, und danach weist jede gebundene Methode mit
// „die Anwendung endet" ab — bis zum Neustart. Der Bestandsweg schließt über
// `App.Beenden`, und das schließt **das Tor des Schlusses**; `App.Starten`
// öffnet den Bestand wieder, das Tor aber nicht.
//
// # Warum die Tests hier über die Bindungen gehen
//
// **Weil die Tests daneben die Lücke nicht sehen konnten**:
// `sicherungzurueckspielen_test.go` ruft `App.SicherungZurueckspielen` und
// danach `App.Einstellungen` — am Tor vorbei, denn das Tor steht in
// [app.Bindungen]. Beide Seiten waren grün, und die Naht dazwischen trug
// nicht (§10, Nähte).
//
// # Und warum alle drei Wege
//
// **Die Fehlerklasse ist der Tausch, nicht das Zurückspielen**: Verschlüsseln
// und Neuanfang nehmen denselben Weg — schließen, tauschen, öffnen.

package app_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestNachDemZurueckspielenGehtDasTorWiederAuf(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)
	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // nach dem Tausch offen ist eine andere Datei

	bindungen := app.NeueBindungen(anwendung)
	sicherung := filepath.Join(filepath.Dir(bestand), "sicherung.db")

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return sicherung, nil
	})

	if antwort := bindungen.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	antwort := bindungen.SicherungZurueckspielen(t.Context(), sicherung, "")
	if antwort.Fehler != nil || !antwort.Ersetzt {
		t.Fatalf("zurückspielen: %+v", antwort)
	}

	durchsTor(t, bindungen, "nach dem Zurückspielen")
}

func TestNachDemVerschluesselnGehtDasTorWiederAuf(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // siehe oben

	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.VerschluesselungUmschalten(t.Context(), true); antwort.Fehler != nil {
		t.Fatalf("einschalten: %+v", antwort.Fehler)
	}

	durchsTor(t, bindungen, "nach dem Einschalten")

	if antwort := bindungen.VerschluesselungUmschalten(t.Context(), false); antwort.Fehler != nil {
		t.Fatalf("ausschalten: %+v", antwort.Fehler)
	}

	durchsTor(t, bindungen, "nach dem Ausschalten")
}

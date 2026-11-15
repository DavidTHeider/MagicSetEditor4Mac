package app_test

// **Die eingestellte Adresse gilt wirklich** —
// #1540.
//
// # Warum es diesen Lauf **neben** `dienstadresse_intern_test.go` gibt
//
// **Weil jener die Faltung überspringt.** Er setzt den Stand als fertige
// Karte (`service.Wahl{Wirksam: adresse}`) und prüft von dort ab — genau die
// Vorrichtung, die großzügiger ist als der Betrieb (§3).
//
// **Und in der übersprungenen Hälfte lag der Fehler**: `findeRichtlinie` nahm
// jeden Schlüssel unter `dienst.` als An/Aus-Schranke, eine Adresse fand sich
// in keiner Reihung `{aus, an}`, und `Wirksam` fiel auf die leere Vorgabe
// zurück. Der Lauf daneben blieb grün, weil er nie danach fragte.
//
// **Gefunden hat es die Sichtprüfung am laufenden Fenster** (2026-09-13), und
// deshalb geht dieser Lauf durch den **echten** Bestand: Einstellung schreiben,
// Katalog lesen, Adresse vergleichen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine geschriebene Adresse steht im Katalog — und zwar sie.**
func TestEineGeschriebeneDienstadresseStehtImKatalog(t *testing.T) {
	t.Parallel()

	const gateway = "http://127.0.0.1:8099"

	db := frischerBestand(t)
	ctx := probekontext(t)

	for schluessel, wert := range map[string]string{
		domain.DienstEinstellung(domain.DienstOpenAI): domain.DienstAn,
		domain.DienstAdresse(domain.DienstOpenAI):     gateway,
	} {
		if fehler := storage.NeueEinstellungen(db).Setzen(
			ctx, schluessel, wert); fehler != nil {
			t.Fatalf("%s setzen: %v", schluessel, fehler)
		}
	}

	antwort := app.KIAnwendungUeber(db).MitKIZuhause(t.TempDir()).
		KIModelle(ctx)
	if antwort.Fehler != nil {
		t.Fatalf("KIModelle: %+v", antwort.Fehler)
	}

	gesehen := map[string]string{}
	for _, einer := range antwort.Anbieter {
		gesehen[einer.Name] = einer.Adresse
	}

	if gesehen["openai"] != gateway {
		t.Errorf("OpenAI steht unter %q — erwartet %q",
			gesehen["openai"], gateway)
	}
}

// **Und ohne Einstellung bleibt es die Vorgabe des Anbieters.**
//
// **Die Gegenprobe zur Zeile darüber** (§10): eine Faltung, die jeden Wert
// durchreichte, ließe auch eine leere Einstellung durch — und dann spräche
// niemand mehr mit dem Anbieter.
func TestOhneEinstellungBleibtDieVorgabeadresse(t *testing.T) {
	t.Parallel()

	db := frischerBestand(t)
	ctx := probekontext(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(ctx,
		domain.DienstEinstellung(domain.DienstOpenAI),
		domain.DienstAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	antwort := app.KIAnwendungUeber(db).MitKIZuhause(t.TempDir()).
		KIModelle(ctx)
	if antwort.Fehler != nil {
		t.Fatalf("KIModelle: %+v", antwort.Fehler)
	}

	for _, einer := range antwort.Anbieter {
		if einer.Name != "openai" {
			continue
		}

		if einer.Adresse == "" {
			t.Fatal("OpenAI steht ohne Adresse da — dann weiß niemand, " +
				"wohin gefragt wird")
		}

		return
	}

	t.Fatal("OpenAI steht nicht im Katalog")
}

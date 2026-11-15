package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein Sprachwechsel wirkt auf die Menüleiste.**
//
// Geprüft an der Leiste, die die Anwendung setzt — nicht an der, die ein Test
// selbst baut. Was hier nicht steht: dass macOS sie auch zeichnet. Das zeigt
// nur die gebaute Anwendung (§3).
func TestSpracheSetzenTauschtDieMenueleiste(t *testing.T) {
	anwendung := app.NeueApp("")

	var gesetzt *app.Menue

	anwendung.MitMenue(func(_ context.Context, leiste *app.Menue) { gesetzt = leiste })

	if antwort := anwendung.SpracheSetzen("en"); antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if gesetzt == nil {
		t.Fatal("es wurde keine Leiste gesetzt")
	}

	eintrag := eintragMitKuerzel(gesetzt, app.Befehlstaste(","))
	if eintrag == nil {
		t.Fatal("kein Eintrag auf ⌘, gefunden")
	}

	if eintrag.Beschriftung != app.MenuetexteFuer("en").Einstellungen {
		t.Errorf("%q — erwartet war die englische Beschriftung", eintrag.Beschriftung)
	}
}

// Die Handlungen überleben den Wechsel; sonst wäre die Leiste danach tot.
func TestDieHandlungenUeberlebenDenSprachwechsel(t *testing.T) {
	gerufen := 0
	anwendung := app.NeueApp("")

	var gesetzt *app.Menue

	anwendung.MitMenue(func(_ context.Context, leiste *app.Menue) { gesetzt = leiste })
	anwendung.MitMenuhandlungen(app.Menuhandlungen{
		EinstellungenOeffnen: func() { gerufen++ },
	})

	anwendung.SpracheSetzen("en")

	eintrag := eintragMitKuerzel(gesetzt, app.Befehlstaste(","))
	if eintrag == nil {
		t.Fatal("kein Eintrag auf ⌘, gefunden")
	}

	eintrag.Handlung()

	if gerufen != 1 {
		t.Errorf("die Handlung wurde %d-mal gerufen", gerufen)
	}
}

// Eine Sprache, die es nicht gibt, wird abgewiesen — und nichts getauscht.
func TestEineUnbekannteSpracheWirdAbgewiesen(t *testing.T) {
	anwendung := app.NeueApp("")

	getauscht := false

	anwendung.MitMenue(func(context.Context, *app.Menue) { getauscht = true })

	befund := anwendung.SpracheSetzen("kl").Fehler
	if befund == nil || befund.Feld != "sprache" {
		t.Fatalf("%+v", befund)
	}

	if getauscht {
		t.Error("die Leiste wurde trotzdem getauscht")
	}
}

// **Der Kontext aus `Bereit` ist der, mit dem getauscht wird.**
//
// Ohne ihn ginge der Tausch gegen `nil`, und die Leiste bliebe stehen — ohne
// Fehler, ohne Meldung, und sichtbar erst im gebauten Fenster.
func TestDerKontextAusBereitWirdBenutzt(t *testing.T) {
	anwendung := app.NeueApp("")
	fenster := context.WithValue(probekontext(t), probeschluessel{}, "das Fenster")

	var bekommen context.Context

	anwendung.MitMenue(func(ctx context.Context, _ *app.Menue) { bekommen = ctx })
	anwendung.Bereit(fenster)
	anwendung.SpracheSetzen("de")

	if bekommen == nil || bekommen.Value(probeschluessel{}) != "das Fenster" {
		t.Errorf("%v", bekommen)
	}
}

type probeschluessel struct{}

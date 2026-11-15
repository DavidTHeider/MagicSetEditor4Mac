package app_test

// **Die langen Läufe des Fensters stehen im Auftragsbuch** — FMT-03.AK-02
// (#1692):
// mit ihrer Art und der Datei, um die es geht, und mit dem Stand, in dem sie
// endeten. Eingetragen wird erst **nach** dem Dialog — wer ihn schließt, hat
// nichts angestoßen.

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// imBuch prüft den jüngsten Auftrag.
func imBuch(t *testing.T, anwendung *app.App, art, was, stand string) {
	t.Helper()

	buch := anwendung.Auftraege(t.Context()).Auftraege
	if len(buch) != 1 || buch[0].Art != art || buch[0].Was != was || buch[0].Stand != stand {
		t.Errorf("im Buch %+v, erwartet %s „%s“ %s", buch, art, was, stand)
	}
}

func TestDieLangenLaeufeDesFenstersStehenImBuch(t *testing.T) {
	for _, fall := range []struct {
		name, stand string
		fehler      error
	}{
		{"fertig", "fertig", nil},
		{"abgebrochen", "abgebrochen", fmt.Errorf("gelegt: %w", context.Canceled)},
		{"gescheitert", "gescheitert", errors.New("die Platte ist voll")},
	} {
		t.Run("Bogen "+fall.name, func(t *testing.T) {
			anwendung := app.NeueApp("")
			anwendung.MitBogenleger(&bogenprobe{lauf: &service.Bogenlauf{Pfad: "/tmp/b.pdf"}, fehler: fall.fehler})
			mitDialog(anwendung, "/tmp/Mein Bogen.pdf", nil)

			app.NeueBindungen(anwendung).BogenLegen(t.Context(), "7", "a4", 0, 300, true, nil)
			imBuch(t, anwendung, "bogen", "Mein Bogen.pdf", fall.stand)
		})
	}

	t.Run("Weitergabe", func(t *testing.T) {
		anwendung := app.NeueApp("")
		anwendung.MitPaketen(&weitergabeprobe{})
		mitDialog(anwendung, "/tmp/probe.mse-installer", nil)

		app.NeueBindungen(anwendung).PaketWeitergeben(t.Context(), "probe.mse-game")
		imBuch(t, anwendung, "weitergabe", "probe.mse-installer", "fertig")
	})

	t.Run("Installer", func(t *testing.T) {
		anwendung := app.NeueApp("")
		anwendung.MitPaketen(&weitergabeprobe{uebernahme: &service.Uebernahme{}})
		mitDialog(anwendung, "/tmp/geschenk.mse-installer", nil)

		app.NeueBindungen(anwendung).InstallerEinlesen(t.Context())
		imBuch(t, anwendung, "installer", "geschenk.mse-installer", "fertig")
	})

	// **Die drei übrigen scheitern hier** — für diesen Lauf genügt, dass sie
	// mit Art und Datei im Buch stehen.
	for _, fall := range []struct {
		art, datei string
		fahren     func(*app.Bindungen)
	}{
		{"mtgjson", "AllPrintings.json", func(b *app.Bindungen) { b.MTGJSONOeffnen(t.Context(), "magic") }},
		{"vorlage", "ausfuhr.txt", func(b *app.Bindungen) {
			b.VorlageAusfuehren(t.Context(), "keine Kennung", "probe.mse-export-template", "*.txt", nil, nil)
		}},
		{"set", "probe.mse-set", func(b *app.Bindungen) { b.SetAusDatei(t.Context(), "/tmp/probe.mse-set") }},
	} {
		t.Run(fall.art, func(t *testing.T) {
			anwendung := mitImporteur(&importprobe{fehler: errors.New("kaputt")})
			mitDialog(anwendung, "/tmp/"+fall.datei, nil)

			fall.fahren(app.NeueBindungen(anwendung))
			imBuch(t, anwendung, fall.art, fall.datei, "gescheitert")
		})
	}

	t.Run("ein geschlossener Dialog trägt nichts ein", func(t *testing.T) {
		anwendung := mitImporteur(&importprobe{})
		anwendung.MitBogenleger(&bogenprobe{})
		anwendung.MitPaketen(&weitergabeprobe{})
		mitDialog(anwendung, "", nil)

		bindungen := app.NeueBindungen(anwendung)
		bindungen.BogenLegen(t.Context(), "7", "a4", 0, 300, true, nil)
		bindungen.PaketWeitergeben(t.Context(), "probe.mse-game")
		bindungen.InstallerEinlesen(t.Context())
		bindungen.MTGJSONOeffnen(t.Context(), "magic")
		bindungen.VorlageAusfuehren(t.Context(), "7", "probe.mse-export-template", "*.txt", nil, nil)

		if buch := anwendung.Auftraege(t.Context()).Auftraege; len(buch) != 0 {
			t.Errorf("im Buch %+v", buch)
		}
	})
}

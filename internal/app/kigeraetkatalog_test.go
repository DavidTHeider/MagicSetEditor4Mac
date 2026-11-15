package app_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Wie der Katalog das Gerätemodell zeigt** — AI-05.AK-03/04 (#2042):
// eingeschaltet erst, wenn es gewollt **und** eingerichtet ist; sonst „aus",
// und der Satz darunter sagt, warum.

// geraetezeile gibt die Zeile des Gerätemodells im Katalog.
func geraetezeile(t *testing.T, welt kiwelt) app.KIAnbieterDTO {
	t.Helper()

	antwort := welt.anwendung.KIModelle(probekontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for _, einer := range antwort.Anbieter {
		if einer.Name == "apple" {
			return einer
		}
	}

	t.Fatalf("keine Zeile apple: %+v", antwort.Anbieter)

	return app.KIAnbieterDTO{}
}

func TestDerKatalogSagtWasDemGeraetFehlt(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	mitGeraet(t, welt, applefm.EinrichtungLizenz, fmName)

	zeile := geraetezeile(t, welt)
	if zeile.Einrichtung != app.KIEinrichtungAus || len(zeile.Befunde) != 1 ||
		!strings.Contains(zeile.Befunde[0], "sudo fm license") {
		t.Errorf("%+v", zeile)
	}

	// **Und eine ausdrückliche Wahl nennt denselben Grund.**
	antwort := welt.anwendung.KIVorschlagenMit(probekontext(t), app.KIVorschlagsbezug{
		KarteID: welt.karte, Anbieter: "apple", Alias: applefm.Modell, Auftrag: "x",
	})
	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "sudo fm license") {
		t.Errorf("%+v", antwort.Fehler)
	}
}

func TestDerKatalogZeigtDieSitzung(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	mitGeraet(t, welt, applefm.EinrichtungBereit, fmName)

	if aus := welt.anwendung.KIGeraetemodellFuerSitzung(probekontext(t), false); aus.Fehler != nil {
		t.Fatalf("%+v", aus.Fehler)
	}

	zeile := geraetezeile(t, welt)
	if !zeile.SitzungAus || zeile.Einrichtung != app.KIEinrichtungAus || len(zeile.Befunde) != 1 {
		t.Errorf("%+v", zeile)
	}
}

func TestDieEinstellungAusSchaltetDasGeraetAb(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	gestartet := mitGeraet(t, welt, applefm.EinrichtungBereit, fmName)

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		domain.KIGeraetemodell, domain.KIGeraetemodellAus); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	zeile := geraetezeile(t, welt)
	if zeile.Einrichtung != app.KIEinrichtungAus || len(zeile.Befunde) != 0 || zeile.SitzungAus {
		t.Errorf("%+v", zeile)
	}

	if antwort := vorrang(t, welt, "name"); antwort.Anbieter != "ollama" || gestartet.Load() != 0 {
		t.Errorf("%+v, gestartet %d", antwort, gestartet.Load())
	}
}

// **Das Gerätemodell nennt sein Kontextfenster — gemessen** (AI-01.AK-02,
// #1742): Apple sagt es nicht; 8.009 Eingangstokens gingen, 9.009 nicht
// (Vermerk 0166).
func TestDasGeraetNenntSeinKontextfenster(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	mitGeraet(t, welt, applefm.EinrichtungBereit, fmName)

	for _, modell := range welt.anwendung.KIModelle(probekontext(t)).Modelle {
		if modell.Anbieter == "apple" {
			if modell.Kontextfenster != 8009 || !strings.Contains(modell.KontextBeleg, "0166") {
				t.Errorf("%+v", modell)
			}

			return
		}
	}

	t.Fatal("kein Gerätemodell im Katalog")
}

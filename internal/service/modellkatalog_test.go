package service_test

// **Der Modellkatalog** — Werkstück A6 von Phase 5.3
// (#1445).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einKatalogmodell baut ein gültiges Modell mit einer Fähigkeit.
func einKatalogmodell(alias string, faehigkeit domain.KIFaehigkeit) domain.KIModell {
	modell := domain.KIModell{
		Alias: alias, Anbieter: "ollama", Laufzeit: "ollama 0.12.3",
		Ort: domain.KIOrtLokal,
		Geprueft: []domain.KIGeprueft{{
			Faehigkeit: faehigkeit, Datum: "2026-09-11",
			Beleg: "eigener Lauf gegen die Laufzeit",
		}},
	}

	if faehigkeit == domain.KIEinbettung {
		modell.Einbettung = &domain.KIEinbettungsvertrag{
			Dimension: 768, Pooling: domain.KIPoolingMittel, Zerteiler: "x",
		}
	}

	return modell
}

// **Der Katalog gibt je Fähigkeit nur, was sie belegt.**
func TestDerKatalogGibtNurBelegteFaehigkeiten(t *testing.T) {
	t.Parallel()

	katalog, fehler := service.NeuerModellkatalog(
		einKatalogmodell("bettet:3b", domain.KIEinbettung),
		einKatalogmodell("schreibt:8b", domain.KITextgenerierung),
	)
	if fehler != nil {
		t.Fatalf("der Katalog: %v", fehler)
	}

	einbetter := katalog.Fuer(domain.KIEinbettung)

	if len(einbetter) != 1 || einbetter[0].Alias != "bettet:3b" {
		t.Errorf("die Einbetter sind %+v", einbetter)
	}

	// **Eine Fähigkeit, die niemand belegt, gibt eine leere Liste** — und
	// keinen Fehler: „es kann gerade keiner" ist eine Auskunft.
	if bilder := katalog.Fuer(domain.KIBildgenerierung); len(bilder) != 0 {
		t.Errorf("die Bildmodelle sind %+v", bilder)
	}
}

// **Ein kaputtes Modell kommt gar nicht erst in den Katalog.**
//
// **Beim Bauen und nicht beim Fragen**: ein Katalog, der ungültige Einträge
// führt und sie erst beim Zugriff prüft, ist eine Liste mit einer Falle.
func TestEinKaputtesModellKommtNichtInDenKatalog(t *testing.T) {
	t.Parallel()

	kaputt := einKatalogmodell("ohne:laufzeit", domain.KITextgenerierung)
	kaputt.Laufzeit = ""

	if _, fehler := service.NeuerModellkatalog(kaputt); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Dasselbe Modell zweimal ebenso.**
func TestDasselbeModellZweimalWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	eines := einKatalogmodell("bettet:3b", domain.KIEinbettung)

	if _, fehler := service.NeuerModellkatalog(eines, eines); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ein Aliaswechsel wird sichtbar — und nur ein Wechsel.**
//
// Das Issue verlangt „Modellaliasänderung sichtbar". Verglichen werden
// **Bindungen**, nicht Namen: derselbe Name hinter einem anderen Artefakt ist
// der Fall, den man sonst nicht bemerkt.
func TestEinAliaswechselWirdSichtbar(t *testing.T) {
	t.Parallel()

	vorher := einKatalogmodell("bettet:3b", domain.KIEinbettung)
	vorher.Herkunft = eineKettenprobe("sha256:alt")

	nachher := einKatalogmodell("bettet:3b", domain.KIEinbettung)
	nachher.Herkunft = eineKettenprobe("sha256:neu")

	katalog, fehler := service.NeuerModellkatalog(nachher)
	if fehler != nil {
		t.Fatalf("der Katalog: %v", fehler)
	}

	befunde := katalog.Wechsel(map[string]string{
		vorher.Anbieter + "/" + vorher.Alias: vorher.Bindung(),
	})

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %+v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "bettet:3b") {
		t.Errorf("die Meldung ist %q", befunde[0].Meldung)
	}

	// **Und ohne Wechsel schweigt er** — ein Wächter, der immer meldet, wird
	// nicht gelesen.
	gleich, bauefehler := service.NeuerModellkatalog(vorher)
	if bauefehler != nil {
		t.Fatalf("der zweite Katalog: %v", bauefehler)
	}

	if ruhig := gleich.Wechsel(map[string]string{
		vorher.Anbieter + "/" + vorher.Alias: vorher.Bindung(),
	}); len(ruhig) != 0 {
		t.Errorf("ohne Wechsel kamen %d Befunde: %+v", len(ruhig), ruhig)
	}
}

// **Ein Modell, das vorher nicht dastand, ist kein Wechsel.**
//
// **Es ist ein Zugang**, und der Unterschied zählt: wer ein Modell aufnimmt,
// weiß es; wer eines unter der Hand getauscht bekommt, nicht.
func TestEinNeuesModellIstKeinWechsel(t *testing.T) {
	t.Parallel()

	katalog, fehler := service.NeuerModellkatalog(
		einKatalogmodell("neu:3b", domain.KITextgenerierung))
	if fehler != nil {
		t.Fatalf("der Katalog: %v", fehler)
	}

	if befunde := katalog.Wechsel(map[string]string{}); len(befunde) != 0 {
		t.Errorf("ein neues Modell gab %d Befunde: %+v", len(befunde), befunde)
	}
}

// **Die Bindungen kommen als Karte heraus** — zum Aufheben für das nächste Mal.
func TestDerKatalogGibtSeineBindungen(t *testing.T) {
	t.Parallel()

	modell := einKatalogmodell("bettet:3b", domain.KIEinbettung)

	katalog, fehler := service.NeuerModellkatalog(modell)
	if fehler != nil {
		t.Fatalf("der Katalog: %v", fehler)
	}

	bindungen := katalog.Bindungen()

	if len(bindungen) != 1 ||
		bindungen["ollama/bettet:3b"] != modell.Bindung() {
		t.Errorf("die Bindungen sind %v", bindungen)
	}
}

// eineKettenprobe baut eine Herkunft mit einer bestimmten Artefaktrevision.
func eineKettenprobe(revision string) domain.KIHerkunft {
	return domain.KIHerkunft{Glieder: []domain.KIGlied{{
		Art: domain.KIArtefakt, Kennung: "ausgefuehrt",
		Quelle: "ollama:beispiel", Revision: revision,
		Dateien: []domain.KIDateihash{{Name: "modell.gguf", Hash: revision}},
		Lizenz:  "Apache-2.0",
	}}}
}

// **Der Katalog gibt eine Kopie und nicht seine Liste.**
//
// **Sonst könnte ein Leser ihn ändern** — und der nächste Aufrufer bekäme
// einen Katalog, den niemand geprüft hat.
func TestDerKatalogGibtEineKopie(t *testing.T) {
	t.Parallel()

	katalog, fehler := service.NeuerModellkatalog(
		einKatalogmodell("bettet:3b", domain.KIEinbettung))
	if fehler != nil {
		t.Fatalf("der Katalog: %v", fehler)
	}

	alle := katalog.Alle()

	if len(alle) != 1 {
		t.Fatalf("%d Modelle, erwartet 1", len(alle))
	}

	alle[0].Alias = "erfunden"

	if katalog.Alle()[0].Alias != "bettet:3b" {
		t.Error("der Katalog ließ sich von außen ändern")
	}
}

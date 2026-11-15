package app_test

// **Die Naht zur KI** — Werkstück A6 von Phase 5.3
// (#1445).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ohne Modelle ist die Liste leer — und die Regel steht dabei.**
//
// **Das ist die Auskunft und kein Mangel** (Issue): „Unbekannt bleibt
// unbekannt; Modelllisten garantieren keine Capability."
func TestOhneModelleIstDieListeLeer(t *testing.T) {
	t.Parallel()

	antwort := app.KIAnwendungUeber(frischerBestand(t)).KIModelle(probekontext(t))

	if antwort.Fehler != nil {
		t.Fatalf("KIModelle: %+v", antwort.Fehler)
	}

	if len(antwort.Modelle) != 0 {
		t.Errorf("die Liste ist %+v", antwort.Modelle)
	}

	// **Die Vorgabe ist »aus«** — `richtlinie.nurlokal` ist der Zustand von
	// gestern, und eine Einstellung, die beim Einbau etwas abschaltet,
	// schaltet es für jeden ab, der nie danach gefragt wurde.
	if antwort.NurLokal {
		t.Error("ohne Wahl gilt »nur lokal«")
	}
}

// **Die Richtlinie kommt an.**
func TestDieRichtlinieKommtAmKatalogAn(t *testing.T) {
	t.Parallel()

	db := frischerBestand(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(probekontext(t),
		domain.RichtlinieNurLokal, domain.NurLokalAn); fehler != nil {
		t.Fatalf("die Richtlinie setzen: %v", fehler)
	}

	antwort := app.KIAnwendungUeber(db).KIModelle(probekontext(t))

	if antwort.Fehler != nil {
		t.Fatalf("KIModelle: %+v", antwort.Fehler)
	}

	if !antwort.NurLokal {
		t.Error("die gesetzte Richtlinie kam nicht an")
	}
}

// **Ein Modell mit Herkunft kommt vollständig durch.**
func TestEinModellKommtVollstaendigDurch(t *testing.T) {
	t.Parallel()

	modell := domain.KIModell{
		Alias: "beispiel:3b", Anbieter: "ollama", Ort: domain.KIOrtLokal,
		Laufzeit: "ollama 0.12.3",
		Geprueft: []domain.KIGeprueft{{
			Faehigkeit: domain.KITextgenerierung, Datum: "2026-09-11",
			Beleg: "eigener Lauf gegen die Laufzeit",
		}},
		Herkunft: domain.KIHerkunft{Glieder: []domain.KIGlied{{
			Art: domain.KIArtefakt, Kennung: "ausgefuehrt",
			Quelle: "ollama:beispiel", Revision: "sha256:dddd",
			Dateien: []domain.KIDateihash{{Name: "m.gguf", Hash: "sha256:dddd"}},
			Lizenz:  "Apache-2.0",
		}}},
	}

	antwort := app.KIAnwendungUeber(frischerBestand(t)).
		MitKIModellen(modell).KIModelle(probekontext(t))

	if antwort.Fehler != nil {
		t.Fatalf("KIModelle: %+v", antwort.Fehler)
	}

	if len(antwort.Modelle) != 1 {
		t.Fatalf("die Liste ist %+v", antwort.Modelle)
	}

	gesehen := antwort.Modelle[0]

	if gesehen.Alias != "beispiel:3b" || gesehen.Ort != "lokal" ||
		gesehen.Bindung != modell.Bindung() {
		t.Errorf("das Modell ist %+v", gesehen)
	}

	if len(gesehen.Faehigkeiten) != 1 ||
		gesehen.Faehigkeiten[0].Beleg != "eigener Lauf gegen die Laufzeit" {
		t.Errorf("die Fähigkeiten sind %+v", gesehen.Faehigkeiten)
	}

	if len(gesehen.Herkunft) != 1 || gesehen.Herkunft[0].Art != "artefakt" {
		t.Errorf("die Herkunft ist %+v", gesehen.Herkunft)
	}
}

// **Ein kaputtes Modell macht den Katalog zum Fehler.**
//
// **Beim Fragen und nicht beim Setzen**: `MitKIModellen` nimmt, was es
// bekommt; geprüft wird, wenn jemand hinsieht — und dann ist ein
// unvollständiges Modell ein Fehler und keine halbe Liste.
func TestEinKaputtesModellMachtDenKatalogZumFehler(t *testing.T) {
	t.Parallel()

	antwort := app.KIAnwendungUeber(frischerBestand(t)).
		MitKIModellen(domain.KIModell{Alias: "ohne:alles"}).
		KIModelle(probekontext(t))

	if antwort.Fehler == nil {
		t.Fatal("ein unvollständiges Modell kam durch")
	}
}

// **Ohne Bestand gibt es keinen Katalog.**
func TestOhneBestandGibtEsKeinenKatalog(t *testing.T) {
	t.Parallel()

	if antwort := app.NeueApp("").KIModelle(probekontext(t)); antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam ein Katalog")
	}
}

// **Ein Fehler der Einstellungen reist weiter.**
func TestEinFehlerDerEinstellungenHaeltDenKatalogAuf(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Einstellungen klemmen")

	anwendung := app.KIAnwendungUeber(frischerBestand(t))
	anwendung.MitEinstellungen(klemmendeEinstellungen{fehler: eigen})

	if antwort := anwendung.KIModelle(probekontext(t)); antwort.Fehler == nil {
		t.Fatal("ein Einstellungsfehler kam als Katalog zurück")
	}
}

// klemmendeEinstellungen antworten mit einem Fehler.
type klemmendeEinstellungen struct{ fehler error }

func (k klemmendeEinstellungen) Alle(context.Context) (map[string]string, error) {
	return nil, k.fehler
}

func (k klemmendeEinstellungen) Setzen(context.Context, string, string) error {
	return k.fehler
}

func (k klemmendeEinstellungen) Stand(context.Context) (map[string]service.Wahl, error) {
	return nil, k.fehler
}

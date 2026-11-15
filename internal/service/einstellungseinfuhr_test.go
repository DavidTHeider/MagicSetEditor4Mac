package service_test

// **Die Übernahme** — Werkstück O3 von Phase 5.2
// (#1427).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitRevision holt die Revision aus einer Vorschau derselben Datei.
func mitRevision(t *testing.T, bank *werkbank, roh []byte) string {
	t.Helper()

	vorschau, fehler := bank.austausch.Vorschau(t.Context(), roh)
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	return vorschau.Revision
}

// **Der geglückte Weg: was in der Vorschau stand, steht danach im Bestand.**
func TestDieUebernahmeSchreibtBeideEbenen(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	datei := alsDatei(t, map[string]map[string]string{
		string(domain.EbeneProfil): {domain.EinstellungSprache: "en"},
		string(domain.EbeneNutzer): {domain.NutzerKontrast: domain.KontrastHoch},
	})

	ergebnis, fehler := bank.austausch.Uebernehmen(
		t.Context(), datei, mitRevision(t, bank, datei), nil)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if ergebnis.Uebernommen != 2 {
		t.Errorf("es kamen %d Zeilen an", ergebnis.Uebernommen)
	}

	if bank.schreiber.profil[domain.EinstellungSprache] != "en" {
		t.Errorf("im Profil steht %v", bank.schreiber.profil)
	}

	if bank.schreiber.nutzer[domain.NutzerKontrast] != domain.KontrastHoch {
		t.Errorf("beim Menschen steht %v", bank.schreiber.nutzer)
	}

	// **Die neue Revision ist eine andere** — sonst ginge ein zweiter Aufruf
	// mit derselben Marke noch einmal durch.
	if ergebnis.Revision == mitRevision(t, bank, datei) {
		t.Error("die Revision hat sich nicht geändert")
	}
}

// **Eine überholte Revision lässt alles stehen** (§7, Regel 2).
func TestEineUeberholteRevisionAendertNichts(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	datei := alsDatei(t, map[string]map[string]string{
		string(domain.EbeneProfil): {domain.EinstellungSprache: "en"},
	})

	_, fehler := bank.austausch.Uebernehmen(t.Context(), datei, "vongestern", nil)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("%v", fehler)
	}

	if bank.schreiber.rufe != 0 {
		t.Error("es wurde trotzdem geschrieben")
	}
}

// **Eine Vorschau aus einem anderen Profil zählt nicht.**
//
// # Warum das derselbe Mechanismus ist und kein zweiter
//
// **Weil die Marke das Profil mitfasst** ([domain.Standmarke]). Ein Format,
// das die Profilkennung mitschickte, wäre der andere Weg — und eine lokale
// Kennung in einer Datei, die von A nach B reisen soll, ist genau das, was
// nicht mitreisen darf ([0079](../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md)).
//
// **Der Fall ist real**: zwischen „zeig mir den Unterschied" und „übernimm"
// liegt am Fenster ein Klick, und dazwischen kann jemand das Profil wechseln.
func TestEineVorschauAusEinemAnderenProfilZaehltNicht(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	datei := alsDatei(t, map[string]map[string]string{
		string(domain.EbeneProfil): {domain.EinstellungSprache: "en"},
	})

	revision := mitRevision(t, bank, datei)

	// **Der Wechsel** — derselbe Bestand, ein anderes Profil.
	bank.wache.akteur.ProfilID = 8

	if _, fehler := bank.austausch.Uebernehmen(
		t.Context(), datei, revision, nil); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("%v", fehler)
	}

	if bank.schreiber.rufe != 0 {
		t.Error("es wurde ins fremde Profil geschrieben")
	}
}

// **Und ohne Revision wird gar nicht erst angefangen.**
//
// Eine Übernahme ohne sie wäre ein Import ohne Vorschau — und die Vorschau ist
// der Grund, warum es dieses Werkstück gibt.
func TestOhneRevisionKeineUebernahme(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()

	_, fehler := bank.austausch.Uebernehmen(t.Context(), alsDatei(t, nil), "", nil)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("%v", fehler)
	}

	if bank.schreiber.rufe != 0 {
		t.Error("es wurde geschrieben")
	}
}

// **Ein Leser übernimmt nicht — und das Nein kommt vor jeder anderen
// Antwort.**
func TestOhneErlaubnisKeineUebernahme(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	bank.wache.nein = domain.ErrErlaubnis

	_, fehler := bank.austausch.Uebernehmen(
		t.Context(), []byte("das ist kein JSON"), "", nil)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("%v", fehler)
	}

	if len(bank.wache.gefragt) != 1 ||
		bank.wache.gefragt[0] != domain.WirkungSchreiben {
		t.Errorf("gefragt wurde nach %v", bank.wache.gefragt)
	}

	if bank.schreiber.rufe != 0 {
		t.Error("es wurde geschrieben")
	}
}

// **Die Auswahl nimmt eine Zeile und lässt die andere stehen.**
func TestDieAuswahlNimmtNurWasGenanntIst(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	datei := alsDatei(t, map[string]map[string]string{
		string(domain.EbeneProfil): {
			domain.EinstellungSprache:      "en",
			domain.EinstellungBildkonflikt: domain.BildkonfliktAltbehalten,
		},
	})

	ergebnis, fehler := bank.austausch.Uebernehmen(
		t.Context(), datei, mitRevision(t, bank, datei),
		[]string{string(domain.EbeneProfil) + ":" + domain.EinstellungSprache})
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if ergebnis.Uebernommen != 1 {
		t.Errorf("es kamen %d Zeilen an", ergebnis.Uebernommen)
	}

	if _, dabei := bank.schreiber.profil[domain.EinstellungBildkonflikt]; dabei {
		t.Errorf("die abgewählte Zeile kam mit: %v", bank.schreiber.profil)
	}
}

// **Eine Auswahl, die im Diff nicht steht, wird gemeldet.**
func TestEineVerfehlteAuswahlWirdGemeldet(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	datei := alsDatei(t, map[string]map[string]string{
		string(domain.EbeneProfil): {domain.EinstellungSprache: "en"},
	})

	ergebnis, fehler := bank.austausch.Uebernehmen(
		t.Context(), datei, mitRevision(t, bank, datei),
		[]string{"profil:gibtesnicht"})
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if ergebnis.Uebernommen != 0 {
		t.Errorf("es kamen %d Zeilen an", ergebnis.Uebernommen)
	}

	if len(ergebnis.Befunde) != 1 {
		t.Fatalf("%+v", ergebnis.Befunde)
	}

	// **Und es wurde nicht geschrieben** — eine Transaktion ohne Wirkung wäre
	// eine Sperre auf der Datei für nichts.
	if bank.schreiber.rufe != 0 {
		t.Error("es wurde geschrieben")
	}
}

// **Ein Fehler beim Schreiben ist ein Fehler und kein halber Stand.**
func TestEinFehlerBeimSchreibenKommtDurch(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	bank.schreiber.fehler = errors.New("die Platte ist voll")

	datei := alsDatei(t, map[string]map[string]string{
		string(domain.EbeneProfil): {domain.EinstellungSprache: "en"},
	})

	if _, fehler := bank.austausch.Uebernehmen(
		t.Context(), datei, mitRevision(t, bank, datei), nil); fehler == nil {
		t.Fatal("der Fehler wurde verschluckt")
	}
}

// **Was der Bestand nicht hergibt, bricht früh ab — an jedem der drei
// Lesewegen.**
func TestWennDerBestandNichtAntwortet(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*werkbank){
		"kein Akteur":       func(b *werkbank) { b.wache.akteurAus = errors.New("kein Zeiger") },
		"keine Profilwerte": func(b *werkbank) { b.ablage.fehler = errors.New("keine Tafel") },
		"keine Nutzerwerte": func(b *werkbank) { b.nutzerwerte.fehler = errors.New("kein Mensch") },
	}

	for name, brechen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bank := werkbankBauen()
			brechen(bank)

			if _, fehler := bank.austausch.Ausfuhr(t.Context(), nil); fehler == nil {
				t.Error("die Ausfuhr kam durch")
			}

			if _, fehler := bank.austausch.Vorschau(
				t.Context(), alsDatei(t, nil)); fehler == nil {
				t.Error("die Vorschau kam durch")
			}

			if _, fehler := bank.austausch.Uebernehmen(
				t.Context(), alsDatei(t, nil), "egal", nil); fehler == nil {
				t.Error("die Übernahme kam durch")
			}
		})
	}
}

// **Eine Datei, die sich nicht lesen lässt, erreicht den Bestand nicht.**
func TestEineUnlesbareDateiWirdNichtUebernommen(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()

	if _, fehler := bank.austausch.Uebernehmen(
		t.Context(), []byte("{"), "egal", nil); fehler == nil {
		t.Fatal("sie kam durch")
	}

	if bank.schreiber.rufe != 0 {
		t.Error("es wurde geschrieben")
	}
}

package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Stiloption wird für sich geschrieben, nicht mit der ganzen Liste** —
// aus dem Zweitblick über Phase L (2026-08-22).
//
// # Der Befund
//
// `service.OptionSetzen` las die **vollständige** Optionsliste, änderte einen
// Eintrag und schrieb alles zurück — samt `kopf.Game` und `kopf.Style`. Daraus
// folgen zwei Verluste, und beide sind der Normalfall im
// Zwei-Schreiber-Betrieb ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)):
//
//  1. **Zwei Optionsschreibungen verlieren einander.** Beide lesen denselben
//     Stand; die zuletzt fertige ersetzt die Änderung der ersten.
//  2. **Ein alter Optionslauf dreht einen Stilwechsel zurück.** Er trägt den
//     Stil mit, den er beim Lesen sah.
//
// §7 Regel 4 sagt „ein Vorgang je Datensatz" — und `docs/SCHREIBWEGE.md` nannte
// als Datensatz die **Option**. Solange geschrieben wurde, was gelesen worden
// war, war der Datensatz das **ganze Set-Styling**.
//
// # Was hier zugesichert wird
//
// `styling_values` trägt `PRIMARY KEY (set_id, feld)`. Eine einzelne Option
// lässt sich damit setzen und entfernen, ohne die Nachbarn und ohne `sets.stil`
// zu berühren.
func TestEineStiloptionWirdFuerSichGeschrieben(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueStilablage(db)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Style: "probe-schlicht",
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if fehler := ablage.StilSetzen(probekontext(t), setID, "probe", "probe-schlicht",
		[]domain.Value{{Field: "hintergrund", Text: "hell"}}); fehler != nil {
		t.Fatalf("Grundstand: %v", fehler)
	}

	// **Zwei Optionen nacheinander, jede für sich** — die zweite darf die
	// erste nicht wegnehmen.
	if fehler := ablage.StiloptionSetzen(probekontext(t), setID, "rahmen zeigen", "yes"); fehler != nil {
		t.Fatalf("setzen: %v", err)
	}

	if fehler := ablage.StiloptionSetzen(probekontext(t), setID, "titelfarbe", "rgb(1,2,3)"); fehler != nil {
		t.Fatalf("setzen: %v", err)
	}

	werte, lesefehler := ablage.Stilwerte(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(werte) != 3 {
		t.Fatalf("%d Werte, erwartet drei: %+v", len(werte), werte)
	}

	// **Und der Stil bleibt unberührt** — das ist die zweite Hälfte.
	koepfe, kopffehler := storage.NeuerSetLeser(db).Setkoepfe(probekontext(t))
	if kopffehler != nil || len(koepfe) != 1 {
		t.Fatalf("Köpfe: %v (%d)", kopffehler, len(koepfe))
	}

	if koepfe[0].Style != "probe-schlicht" {
		t.Errorf("der Stil steht auf %q — eine Optionsschreibung hat ihn angefasst", koepfe[0].Style)
	}
}

// **Und sie lässt sich für sich entfernen** — der Rückweg auf die Vorgabe
// (#764).
func TestEineStiloptionWirdFuerSichEntfernt(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueStilablage(db)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Style: "probe-schlicht",
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if fehler := ablage.StilSetzen(probekontext(t), setID, "probe", "probe-schlicht",
		[]domain.Value{
			{Field: "hintergrund", Text: "dunkel"},
			{Field: "titelfarbe", Text: "rgb(1,2,3)"},
		}); fehler != nil {
		t.Fatalf("Grundstand: %v", fehler)
	}

	if fehler := ablage.StiloptionEntfernen(probekontext(t), setID, "titelfarbe"); fehler != nil {
		t.Fatalf("entfernen: %v", err)
	}

	werte, lesefehler := ablage.Stilwerte(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(werte) != 1 || werte[0].Field != "hintergrund" {
		t.Errorf("übrig: %+v — erwartet nur »hintergrund«", werte)
	}
}

// **Ein Set, das es nicht gibt, ist ein `ErrNotFound`** — nicht ein
// stillschweigend geglücktes Schreiben.
//
// Seit #1815 entscheidet es die Profilklemme vor dem Schreiben; bis dahin war
// es die Zählung der betroffenen Zeilen danach.
func TestEineOptionAnEinemFehlendenSetIstNichtGefunden(t *testing.T) {
	db := frischeDatenbank(t)

	fehler := storage.NeueStilablage(db).StiloptionSetzen(probekontext(t), 99, "hintergrund", "hell")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", fehler)
	}
}

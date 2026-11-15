package service_test

// **Gespeichert wird, was gerechnet ist** — #2396
// (Issue (#2396)).
//
// MSE rechnet nach einer Eingabe das Wertskript des Felds und zieht nach, wer
// es liest (`SetScriptManager::updateValue`, `src/script/script_manager.cpp`,
// Zeilen 249–262, Fassung f9d9356 vom 2021-07-06); gesichert wird danach der
// gerechnete Wert. Bis #2396 speicherte diese Anwendung den eingegebenen.

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// wegstile gibt jeder Karte dieselbe Vorlage — oder keine.
type wegstile struct {
	vorlage    service.Vorlage
	gefunden   bool
	gefragtMit []domain.Value
}

func (w *wegstile) StilFuerKarte(
	_ context.Context, karte *domain.CardRow,
) (service.Vorlage, []domain.Befund, bool) {
	w.gefragtMit = karte.Values

	return w.vorlage, nil, w.gefunden
}

func nachzugfelder() []domain.Field {
	return []domain.Field{
		{Name: "eins", Gespeichert: true, Skript: `"[" + value + "]"`},
		{Name: "leser", Gespeichert: true, Skript: `card.eins + "!"`},
		{Name: "anzeige", Gespeichert: false, Skript: `card.eins + "?"`},
		{Name: "fremd", Gespeichert: true, Skript: `"<" + value + ">"`},
	}
}

func nachzugkarte() *domain.CardRow {
	return &domain.CardRow{ID: 7, SetID: 1, Values: []domain.Value{
		{Field: "eins", Text: "[alt]"}, {Field: "leser", Text: "[alt]!"}, {Field: "fremd", Text: "roh"},
	}}
}

func TestFeldwegSpeichertDenGerechnetenWertUndZiehtNach(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	stile := &wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: nachzugfelder()}, ablage).
		MitNachzug(stile, nil, nil)

	if _, _, fehler := weg.FeldSetzenMitVorwert(t.Context(), 7, "eins", "Funke"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	// **Das Feld zuerst, dann wer es liest — gespeichert** — `anzeige` wird
	// mitgerechnet und nicht gesichert, `fremd` liest nichts und bleibt.
	if want := []string{"eins=[Funke]", "leser=[Funke]!"}; !slices.Equal(ablage.geschrieben, want) {
		t.Errorf("geschrieben %v, erwartet %v", ablage.geschrieben, want)
	}

	// **Gerechnet wird an der Karte mit dem neuen Wert.**
	if !slices.Contains(stile.gefragtMit, domain.Value{Field: "eins", Text: "Funke"}) {
		t.Errorf("die Vorlage wurde mit %v gefragt", stile.gefragtMit)
	}
}

// **Ohne Vorlage bleibt der eingegebene Wert** — wie bisher, wenn das Spiel
// nicht installiert ist (#720).
func TestFeldwegOhneVorlageSpeichertDenEingegebenenWert(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: nachzugfelder()}, ablage).
		MitNachzug(&wegstile{}, nil, nil)

	if _, _, fehler := weg.FeldSetzenMitVorwert(t.Context(), 7, "eins", "Funke"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if want := []string{"eins=Funke"}; !slices.Equal(ablage.geschrieben, want) {
		t.Errorf("geschrieben %v, erwartet %v", ablage.geschrieben, want)
	}
}

// **Die Teile eines Verbunds rechnen jedes sein eigenes Skript** — wie in
// `combined_editor`, das je Teil `update` ruft
// (`src/script/functions/editor.cpp`, Zeilen 76–92, Fassung f9d9356 vom
// 2021-07-06).
func TestFeldwegRechnetDieTeileEinesVerbunds(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{
		{Name: "text", Gespeichert: false, Skript: `combined_editor(field1: card.rule_text, ` +
			`separator: "<line>\n</line>", field2: card.flavor_text)`},
		{Name: "rule text", Gespeichert: true, Skript: `"R:" + value`},
		{Name: "flavor text", Gespeichert: true, Skript: `"F:" + value`},
	}
	karte := &domain.CardRow{ID: 7, SetID: 1}
	ablage := &wegablage{}
	weg := service.NeuerFeldweg(wegkarten{karte: karte}, wegfelder{felder: felder}, ablage).
		MitNachzug(&wegstile{vorlage: service.Vorlage{Kartenfelder: felder}, gefunden: true}, nil, nil)

	if _, _, fehler := weg.FeldSetzenMitVorwert(t.Context(), 7, "text",
		"Fliegend<sep><line>\n</line></sep>Hoch"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if want := []string{"rule text=R:Fliegend", "flavor text=F:Hoch"}; !slices.Equal(ablage.geschrieben, want) {
		t.Errorf("geschrieben %v, erwartet %v", ablage.geschrieben, want)
	}
}

// **Auch mehrere Felder in einem Zug** — `felder-setzen` und die Übernahme
// eines KI-Vorschlags gehen über denselben Weg (`aufgeloest`).
func TestFeldwegRechnetAuchMehrereFelder(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: nachzugfelder()}, ablage).
		MitNachzug(&wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}, nil, nil)

	if _, _, fehler := weg.FelderVonAussen(t.Context(), 7, []domain.Value{
		{Field: "eins", Text: "Funke"}, {Field: "fremd", Text: "neu"},
	}, ""); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if want := []string{"eins=[Funke]", "fremd=<neu>", "leser=[Funke]!"}; !slices.Equal(ablage.geschrieben, want) {
		t.Errorf("geschrieben %v, erwartet %v", ablage.geschrieben, want)
	}
}

// **Der Nachzug schreibt auf dem Stand, aus dem er rechnete** — Gesamtreview
// über 0.9.0, Befund 2 (#2382): gerechnet wurde vor der Transaktion und ohne
// Bedingung geschrieben; ein zweiter Schreiber dazwischen, und der gerechnete
// Wert passte zu keinem Stand der Karte. **Und nachgezogen ist, was das
// Skript zog** — Befund 1: nur das unterliegt nicht der Bedingung des Fensters.
func TestFeldwegSchreibtAufDemStandSeinerRechnung(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	stile := &wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: nachzugfelder()}, ablage).
		MitNachzug(stile, nil, nil)

	if _, _, fehler := weg.FeldSetzenMitVorwert(t.Context(), 7, "eins", "Funke"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if want := []string{rechenstand(nil)}; !slices.Equal(ablage.erwartet, want) {
		t.Errorf("erwartet %q, gewollt %q", ablage.erwartet, want)
	}

	if want := []string{"leser"}; !slices.Equal(ablage.nachgezogen, want) {
		t.Errorf("nachgezogen %v, erwartet %v", ablage.nachgezogen, want)
	}
}

// **Ohne Rechnung keine Bedingung** — sonst hielte ein Schreiben, das von
// keinem anderen Feld abhängt, jede fremde Änderung der Karte für einen
// Konflikt.
func TestFeldwegOhneVorlageSchreibtOhneStand(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: nachzugfelder()}, ablage).
		MitNachzug(&wegstile{}, nil, nil)

	if _, _, fehler := weg.FeldSetzenMitVorwert(t.Context(), 7, "eins", "Funke"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if want := []string{""}; !slices.Equal(ablage.erwartet, want) || len(ablage.nachgezogen) != 0 {
		t.Errorf("erwartet %q, nachgezogen %v — ohne Vorlage rechnet nichts", ablage.erwartet, ablage.nachgezogen)
	}
}

// **Eine Gruppe ebenso — und ein Stand, den der Auftrag nennt, muss der
// gerechnete sein** (Nachlauf über das Gesamtreview, Befund 3): sonst prüfte
// die Transaktion den Stand des Auftrags, während aus einem anderen gerechnet
// wurde, und eine Rücknahme auf den alten ließe die neuen Werte durch.
func TestEineGruppeSchreibtAufDemStandIhrerRechnung(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	stile := &wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: nachzugfelder()}, ablage).
		MitNachzug(stile, nil, nil)

	if fehler := weg.GruppeSetzenMitHerkunft(t.Context(), []domain.Kartenschritt{
		{KarteID: 7, SetID: 1, Werte: []domain.Value{{Field: "eins", Text: "Glut"}}, Herkunft: "ki"},
		{KarteID: 7, SetID: 1, Werte: []domain.Value{{Field: "eins", Text: "Asche"}}, Herkunft: "ki", Erwartet: nachzugkarte().Stand()},
	}); fehler != nil {
		t.Fatalf("Gruppe: %v", fehler)
	}

	if want := []string{rechenstand(nil), rechenstand(nil)}; !slices.Equal(ablage.erwartet, want) {
		t.Errorf("erwartet %q, gewollt %q", ablage.erwartet, want)
	}

	// **Ein anderer Stand im Auftrag ist ein Konflikt**, und die Gruppe
	// schreibt nichts.
	ablage.geschrieben = nil
	fehler := weg.GruppeSetzenMitHerkunft(t.Context(), []domain.Kartenschritt{
		{KarteID: 7, SetID: 1, Werte: []domain.Value{{Field: "eins", Text: "Glut"}}, Herkunft: "ki"},
		{KarteID: 7, SetID: 1, Werte: []domain.Value{{Field: "eins", Text: "Asche"}}, Herkunft: "ki", Erwartet: "ein anderer Stand"},
	})
	if !errors.Is(fehler, domain.ErrConflict) || len(ablage.geschrieben) != 0 {
		t.Errorf("Gruppe mit anderem Stand: %v, geschrieben %v — erwartet ErrConflict und nichts", fehler, ablage.geschrieben)
	}
}

// **Und die Übernahme eines Vorschlags** — mit dem gerechneten Stand; nennt
// der Auftrag einen anderen, ist das ein Konflikt.
func TestEineUebernahmeSchreibtAufDemStandIhrerRechnung(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	stile := &wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: nachzugfelder()}, ablage).
		MitNachzug(stile, nil, nil)

	for _, erwartet := range []string{"", nachzugkarte().Stand()} {
		if _, _, fehler := weg.FelderVonAussenMitHerkunft(t.Context(), 7,
			[]domain.Value{{Field: "eins", Text: "Glut"}}, "ki", nil, nil, erwartet); fehler != nil {
			t.Fatalf("Übernahme: %v", fehler)
		}
	}

	if want := []string{rechenstand(nil), rechenstand(nil)}; !slices.Equal(ablage.erwartet, want) {
		t.Errorf("erwartet %q, gewollt %q", ablage.erwartet, want)
	}

	_, _, fehler := weg.FelderVonAussenMitHerkunft(t.Context(), 7,
		[]domain.Value{{Field: "eins", Text: "Glut"}}, "ki", nil, nil, "ein anderer Stand")
	if !errors.Is(fehler, domain.ErrConflict) || len(ablage.erwartet) != 2 {
		t.Errorf("Übernahme mit anderem Stand: %v, geschrieben mit %q — erwartet ErrConflict und nichts", fehler, ablage.erwartet)
	}
}

// rechenstand ist der Stand, aus dem der Nachzug an `nachzugkarte` rechnet —
// Karte und Setangaben (#2533).
func rechenstand(setangaben []domain.Value) string {
	return domain.Rechenstand(nachzugkarte().Stand(), setangaben)
}

// **Die Setangaben, die die Vorlage las, gehen in den Stand ein** — #2533:
// ein Wertskript liest auch `set.*`.
func TestFeldwegSchreibtAufKarteUndSetangaben(t *testing.T) {
	t.Parallel()

	setangaben := []domain.Value{{Field: "copyright", Text: "2026"}}
	ablage := &wegablage{}
	stile := &wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder(), Setangaben: setangaben}, gefunden: true}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: nachzugfelder()}, ablage).
		MitNachzug(stile, nil, nil)

	if _, _, fehler := weg.FeldSetzenMitVorwert(t.Context(), 7, "eins", "Funke"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if want := []string{rechenstand(setangaben)}; !slices.Equal(ablage.erwartet, want) {
		t.Errorf("erwartet %q, gewollt %q", ablage.erwartet, want)
	}

	if rechenstand(setangaben) == rechenstand(nil) {
		t.Error("die Setangaben ändern den Stand nicht")
	}
}

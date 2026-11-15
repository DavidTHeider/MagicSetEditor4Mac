package app_test

// **Die Turnierfläche und die Turnierwerkstatt** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Gegen einen echten Bestand** und nicht gegen ein Doppel: was hier geprüft
// wird, ist die **Verdrahtung** — dass die Fläche über denselben Ablagen
// entsteht wie das Fenster. Ein Doppel könnte das nicht zeigen.

import (
	"errors"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Der ganze Weg über die Fläche ohne Fenster.**
func TestDieTurnierflaecheFuehrtEinTurnier(t *testing.T) {
	flaeche := app.TurnierflaecheUeber(frischerBestand(t))

	angelegt := flaeche.TurnierAnlegen(probekontext(t), "Küchentisch", "magic", "", false, "")
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	turnierID := angelegt.Turnier.ID

	for _, name := range []string{"Ada", "Bert"} {
		if antwort := flaeche.TurnierAnmelden(
			probekontext(t), turnierID, name, ""); antwort.Fehler != nil {
			t.Fatalf("%s anmelden: %+v", name, antwort.Fehler)
		}
	}

	// **Erst der Vorschlag** — er schafft keine Tatsachen.
	vorschlag := flaeche.TurnierPaarungsvorschlag(probekontext(t), turnierID)
	if vorschlag.Fehler != nil || len(vorschlag.Paarungen) != 1 {
		t.Fatalf("vorschlagen: %+v", vorschlag)
	}

	runde := flaeche.TurnierRundeLegen(probekontext(t), turnierID)
	if runde.Fehler != nil {
		t.Fatalf("legen: %+v", runde.Fehler)
	}

	if antwort := flaeche.TurnierErgebnisSetzen(probekontext(t), turnierID,
		flaeche.Turnier(probekontext(t), turnierID).Turnier.Fassung,
		runde.Runde.Paarungen[0].ID,
		&app.ErgebnisDTO{SpieleA: 2}); antwort.Fehler != nil {
		t.Fatalf("Ergebnis: %+v", antwort.Fehler)
	}

	rangliste := flaeche.TurnierRangliste(probekontext(t), turnierID)
	if rangliste.Fehler != nil || len(rangliste.Rangliste) != 2 {
		t.Fatalf("Rangliste: %+v", rangliste)
	}

	if rangliste.Rangliste[0].Matchpunkte != 3 {
		t.Errorf("Platz 1 hat %d Punkte, erwartet 3",
			rangliste.Rangliste[0].Matchpunkte)
	}

	// **Abmelden — und er bleibt in der Rangliste.**
	if antwort := flaeche.TurnierAussteigen(probekontext(t), turnierID,
		angelegtsZweiter(t, flaeche, turnierID), true); antwort.Fehler != nil {
		t.Fatalf("aussteigen: %+v", antwort.Fehler)
	}

	nachher := flaeche.TurnierRangliste(probekontext(t), turnierID)
	if len(nachher.Rangliste) != 2 {
		t.Errorf("%d Ränge, erwartet 2", len(nachher.Rangliste))
	}

	// **Auflisten und löschen.**
	if liste := flaeche.Turniere(probekontext(t), "magic"); liste.Fehler != nil ||
		len(liste.Turniere) != 1 {
		t.Errorf("auflisten: %+v", liste)
	}

	if antwort := flaeche.TurnierLoeschen(
		probekontext(t), turnierID); antwort.Fehler != nil {
		t.Fatalf("löschen: %+v", antwort.Fehler)
	}

	if gelesen := flaeche.Turnier(probekontext(t), turnierID); gelesen.Fehler == nil {
		t.Error("das gelöschte Turnier ist noch da")
	}
}

// angelegtsZweiter gibt die Kennung des zweiten Teilnehmers.
func angelegtsZweiter(t *testing.T, flaeche *app.Turnierflaeche, turnierID string) string {
	t.Helper()

	antwort := flaeche.Turnier(probekontext(t), turnierID)
	if antwort.Fehler != nil {
		t.Fatalf("lesen: %+v", antwort.Fehler)
	}

	return antwort.Turnier.Teilnehmer[1].ID
}

// **Und dieselben Wege über die Werkstatt, die MCP und REST bedienen.**
//
// **Gleichlautend und trotzdem geprüft**: sie geben `(any, error)` statt
// einer Antwortstruktur, und diese Umwandlung ist die Stelle, an der ein
// Fehler verschwinden könnte.
func TestDieTurnierwerkstattReichtBeideEndenDurch(t *testing.T) {
	werkstatt := app.NeueTurnierwerkstatt(frischerBestand(t))

	angelegt, fehler := werkstatt.TurnierAnlegen(
		probekontext(t), "Küchentisch", "magic", "", false, 0)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	turnierID := kennungAusAntwort(t, angelegt)

	// **Eine Formatkennung, die es nicht gibt, ist ein Nein** — und der Weg,
	// der aus einer Zahl die Kennung macht, ist betreten (DOM-04, #1699).
	if _, err := werkstatt.TurnierAnlegen(
		probekontext(t), "Mit Format", "magic", "", false, 999); err == nil {
		t.Fatal("mit einem fehlenden Format angelegt, erwartet abgewiesen")
	}

	if _, err := werkstatt.TurnierAnmelden(
		probekontext(t), turnierID, "Ada", 0); err != nil {
		t.Fatalf("anmelden: %v", err)
	}

	// **Mit einer Deckkennung** — der Weg, der aus der Null eine leere
	// Zeichenkette macht, hat zwei Seiten, und die zweite bleibt sonst
	// unbetreten. Das Deck gibt es nicht: erwartet ist ein **Nein** und
	// keine stille Anmeldung ohne Liste.
	if _, err := werkstatt.TurnierAnmelden(
		probekontext(t), turnierID, "Bert", 999); err == nil {
		t.Fatal("mit einem fehlenden Deck angemeldet, erwartet abgewiesen")
	}

	if _, err := werkstatt.TurnierAnmelden(
		probekontext(t), turnierID, "Bert", 0); err != nil {
		t.Fatalf("anmelden: %v", err)
	}

	if _, err := werkstatt.TurnierPaarungsvorschlag(probekontext(t), turnierID); err != nil {
		t.Fatalf("vorschlagen: %v", err)
	}

	if _, err := werkstatt.TurnierRundeLegen(probekontext(t), turnierID); err != nil {
		t.Fatalf("legen: %v", err)
	}

	if _, err := werkstatt.Turnier(probekontext(t), turnierID); err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if _, err := werkstatt.Turniere(probekontext(t), ""); err != nil {
		t.Fatalf("auflisten: %v", err)
	}

	if _, err := werkstatt.TurnierRangliste(probekontext(t), turnierID); err != nil {
		t.Fatalf("Rangliste: %v", err)
	}

	// **Eine Vorlage über die Werkstatt** (DOM-04, #1699) — mit beiden Namen
	// und über denselben Bestand.
	kopie, err := werkstatt.TurnierKopieren(probekontext(t), turnierID, "Küchentisch II", true)
	if err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	if gelesen, ok := kopie.(*app.TurnierDTO); !ok || len(gelesen.Teilnehmer) != 2 {
		t.Errorf("die Kopie: %#v", kopie)
	}

	// **Ein Ergebnis über die Werkstatt** — die Zahlen kommen als `any`
	// herein, weil MCP sie so schickt.
	paarung := ersteWerkstattpaarung(t, werkstatt, turnierID)

	if err := werkstatt.TurnierErgebnisSetzen(probekontext(t), turnierID, paarung,
		map[string]any{"spieleA": 2.0, "spieleB": 0}); err != nil {
		t.Fatalf("Ergebnis: %v", err)
	}

	// **Die Folgen über die Werkstatt** (DOM-04.AK-02, #1699) — mit einem
	// Stand und ohne.
	if _, err := werkstatt.TurnierKorrekturfolgen(probekontext(t), turnierID, paarung,
		map[string]any{"spieleB": 2.0}); err != nil {
		t.Fatalf("die Folgen: %v", err)
	}

	if _, err := werkstatt.TurnierKorrekturfolgen(probekontext(t), turnierID, paarung, nil); err != nil {
		t.Fatalf("die Folgen des Zurücknehmens: %v", err)
	}

	// **Und `nil` nimmt es zurück.**
	if err := werkstatt.TurnierErgebnisSetzen(
		probekontext(t), turnierID, paarung, nil); err != nil {
		t.Fatalf("zurücknehmen: %v", err)
	}

	if err := werkstatt.TurnierAussteigen(probekontext(t), turnierID, 1, true); err != nil {
		t.Fatalf("aussteigen: %v", err)
	}

	// **Abschluss und Protokoll über die Werkstatt** (DOM-04.AK-04, #1699) —
	// ein fehlendes Turnier und ein offenes, das sich nicht wiedereröffnen
	// lässt, sind ein Nein; das Protokoll kommt.
	if err := werkstatt.TurnierAbschliessen(probekontext(t), 999); err == nil {
		t.Error("ein fehlendes Turnier abgeschlossen")
	}

	if err := werkstatt.TurnierAbschliessen(probekontext(t), turnierID); err == nil {
		t.Error("mit offenem Tisch abgeschlossen")
	}

	if err := werkstatt.TurnierWiedereroeffnen(probekontext(t), turnierID); err == nil {
		t.Error("ein offenes Turnier wiedereröffnet")
	}

	if _, err := werkstatt.TurnierProtokoll(probekontext(t), turnierID); err != nil {
		t.Fatalf("das Protokoll: %v", err)
	}

	if err := werkstatt.TurnierLoeschen(probekontext(t), turnierID); err != nil {
		t.Fatalf("löschen: %v", err)
	}
}

// **Ein Turnier, das es nicht gibt, kommt als Fehler heraus.**
func TestDieTurnierwerkstattMeldetEinFehlendesTurnier(t *testing.T) {
	werkstatt := app.NeueTurnierwerkstatt(frischerBestand(t))

	if _, err := werkstatt.Turnier(probekontext(t), 999); err == nil {
		t.Error("angenommen, erwartet abgewiesen")
	}

	if err := werkstatt.TurnierLoeschen(probekontext(t), 999); err == nil {
		t.Error("angenommen, erwartet abgewiesen")
	}
}

// kennungAusAntwort liest die Kennung aus einer Werkstattantwort.
func kennungAusAntwort(t *testing.T, antwort any) int64 {
	t.Helper()

	dto, gut := antwort.(*app.TurnierDTO)
	if !gut {
		t.Fatalf("die Antwort ist kein Turnier: %T", antwort)
	}

	kennung, err := strconv.ParseInt(dto.ID, 10, 64)
	if err != nil {
		t.Fatalf("die Kennung %q: %v", dto.ID, err)
	}

	return kennung
}

// ersteWerkstattpaarung gibt die Kennung der ersten Paarung.
func ersteWerkstattpaarung(
	t *testing.T, werkstatt *app.Turnierwerkstatt, turnierID int64,
) int64 {
	t.Helper()

	roh, err := werkstatt.Turnier(probekontext(t), turnierID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	turnier, gut := roh.(*app.TurnierDTO)
	if !gut {
		t.Fatalf("die Antwort ist kein Turnier: %T", roh)
	}

	kennung, wandelfehler := strconv.ParseInt(
		turnier.Runden[0].Paarungen[0].ID, 10, 64)
	if wandelfehler != nil {
		t.Fatalf("die Paarungskennung: %v", wandelfehler)
	}

	return kennung
}

// **Und die Werkstatt reicht auch ein fachliches Nein durch.**
func TestDieTurnierwerkstattMeldetEinUnmoeglichesErgebnis(t *testing.T) {
	werkstatt := app.NeueTurnierwerkstatt(frischerBestand(t))

	err := werkstatt.TurnierErgebnisSetzen(probekontext(t), 1, 1,
		map[string]any{"spieleA": 0, "spieleB": 0, "unentschieden": 0})
	if err == nil {
		t.Fatal("angenommen, erwartet abgewiesen")
	}

	if errors.Is(err, nil) {
		t.Error("der Fehler ist leer")
	}
}

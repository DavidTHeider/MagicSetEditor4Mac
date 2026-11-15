package service_test

// **Nur was ausgewählt ist — und mit Herkunft** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Zwei Zusagen, und sie hängen zusammen
//
// Das Issue: „**Nur ausgewählte gültige Felder werden geändert.**" Und:
// „**Provenienz bei erster Übernahme verpflichtend, nicht erst P2.**"
//
// Sie hängen zusammen, weil beide an derselben Stelle eingelöst werden: der
// Übernahme. Wer auswählt, entscheidet, was geschrieben wird — und **was
// geschrieben wird, trägt danach seine Herkunft** (§16, Satz 6).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Übernommen wird, was ausgewählt ist — und sonst nichts.**
func TestUebernommenWirdNurDasAusgewaehlte(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{revision: "stand-abc"}
	werk := einWerk(t, karten)

	auftrag := einVorschlagMit(t, werk, map[string][]string{
		"name": {"Blitzschlag"}, "text": {"3 Schaden"},
		"type": {"Instant"},
	})

	bericht, fehler := werk.Uebernehmen(t.Context(), auftrag,
		[]string{"name", "type"}, 7)
	if fehler != nil {
		t.Fatalf("Uebernehmen: %v", fehler)
	}

	if bericht.Felder != 2 {
		t.Errorf("übernommen wurden %d Felder", bericht.Felder)
	}

	// **Der Text steht nicht dabei** — er war nicht ausgewählt.
	if _, dabei := karten.geschrieben["text"]; dabei {
		t.Errorf("geschrieben wurde %v", karten.geschrieben)
	}

	if len(karten.geschrieben) != 2 {
		t.Errorf("geschrieben wurde %v", karten.geschrieben)
	}
}

// **Eine leere Auswahl schreibt nichts und sagt es.**
//
// # Warum kein »dann eben alles«
//
// **Weil das die stillste Art wäre, eine Karte zu überschreiben.** Ein
// Aufrufer, der die Auswahl vergisst — eine Fläche ohne Fenster etwa —,
// bekäme das Gegenteil dessen, was er meinte. Das Issue verlangt Auswahl,
// also ist keine Auswahl ein Fehler und keine Abkürzung.
func TestEineLeereAuswahlSchreibtNichts(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{revision: "stand-abc"}
	werk := einWerk(t, karten)

	auftrag := einVorschlagMit(t, werk,
		map[string][]string{"name": {"Blitzschlag"}})

	_, fehler := werk.Uebernehmen(t.Context(), auftrag, nil, 7)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if karten.schreibvorgaenge != 0 {
		t.Errorf("geschrieben wurde %d mal", karten.schreibvorgaenge)
	}
}

// **Ein Feld, das der Vorschlag nicht trägt, wird abgewiesen.**
//
// # Warum nicht still überspringen
//
// **Weil die Auswahl von außen kommt.** An MCP und REST tippt sie ein
// Aufrufer; ein Tippfehler dort ergäbe sonst eine Übernahme, die weniger tut
// als verlangt — und niemand sähe es. Das ist dieselbe Regel wie bei einem
// unbekannten Feld im Vorschlag selbst.
func TestEinFremdesFeldInDerAuswahlWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{revision: "stand-abc"}
	werk := einWerk(t, karten)

	auftrag := einVorschlagMit(t, werk,
		map[string][]string{"name": {"Blitzschlag"}})

	_, fehler := werk.Uebernehmen(t.Context(), auftrag,
		[]string{"name", "gibtsnicht"}, 7)
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "gibtsnicht") {
		t.Errorf("die Meldung nennt das Feld nicht: %v", fehler)
	}

	// **Und nichts davor wurde geschrieben** — erst prüfen, dann schreiben.
	if karten.schreibvorgaenge != 0 {
		t.Errorf("geschrieben wurde %d mal", karten.schreibvorgaenge)
	}
}

// **Jedes übernommene Feld bekommt seine Herkunft — im selben Zug.**
func TestJedesUebernommeneFeldBekommtSeineHerkunft(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{revision: "stand-abc"}
	werk := einWerk(t, karten)

	auftrag := einVorschlagMit(t, werk, map[string][]string{
		"name": {"Blitzschlag"}, "text": {"3 Schaden"},
	})

	if _, fehler := werk.Uebernehmen(t.Context(), auftrag,
		[]string{"name"}, 7); fehler != nil {
		t.Fatalf("Uebernehmen: %v", fehler)
	}

	if len(karten.herkunft) != 1 {
		t.Fatalf("die Herkunft ist %+v", karten.herkunft)
	}

	herkunft := karten.herkunft[0]

	if befunde := herkunft.Validate(); len(befunde) > 0 {
		t.Fatalf("die Herkunft ist unvollständig: %v", befunde)
	}

	if herkunft.Feld != "name" || herkunft.Auftrag != auftrag {
		t.Errorf("die Herkunft ist %+v", herkunft)
	}

	// **Das Modell ist das, was geantwortet hat** — nicht der Alias, unter
	// dem gefragt wurde.
	if herkunft.Modell != "schreibt:8b-q4" {
		t.Errorf("das Modell ist %q", herkunft.Modell)
	}

	if herkunft.Aufgabe != "Schreibe den Regeltext." {
		t.Errorf("die Aufgabe ist %q", herkunft.Aufgabe)
	}

	// **Der Bezug gehört zum geschriebenen Wert.**
	if herkunft.Bezug != domain.KIInhaltsbezug([]string{"Blitzschlag"}) {
		t.Errorf("der Bezug ist %q", herkunft.Bezug)
	}

	// **Und die Zeit ist die der Übernahme.**
	if !herkunft.Zeitpunkt.Equal(derLauf) {
		t.Errorf("der Zeitpunkt ist %v", herkunft.Zeitpunkt)
	}
}

// **Scheitert die Herkunft, steht kein halber Stand.**
//
// # Was dieser Lauf misst, und was nicht
//
// **Er misst die Naht, nicht die Datenbank**: die Ablage bekommt Werte und
// Herkunft in **einem** Ruf, und sie sagt, dass sie gescheitert ist. Dass
// darunter eine Transaktion liegt, ist die Zusage der Ablage — geprüft wird
// sie dort, wo sie eingelöst wird (§10).
func TestScheiterndeHerkunftLaesstKeinenHalbenStand(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{
		revision: "stand-abc", schreibfehler: errors.New("die Herkunft klemmt"),
	}
	werk := einWerk(t, karten)

	auftrag := einVorschlagMit(t, werk,
		map[string][]string{"name": {"Blitzschlag"}})

	if _, fehler := werk.Uebernehmen(t.Context(), auftrag,
		[]string{"name"}, 7); fehler == nil {
		t.Fatal("die gescheiterte Übernahme ist als Erfolg durchgegangen")
	}

	// **Und der Auftrag bleibt offen** — ein zweiter Versuch ist möglich,
	// denn geschrieben wurde nichts.
	if _, fehler := werk.Uebernehmen(t.Context(), auftrag,
		[]string{"name"}, 7); errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("der Auftrag wurde verbraucht, obwohl nichts geschrieben wurde")
	}
}

// einVorschlagMit legt einen Auftrag an und hinterlegt einen Vorschlag.
func einVorschlagMit(
	t *testing.T, werk *service.KIAuftragswerk, felder map[string][]string,
) string {
	t.Helper()

	auftrag, fehler := werk.Anlegen(t.Context(), einWunsch())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, fehler := werk.Vorschlagen(t.Context(), domain.KIVorschlag{
		Auftrag: auftrag.Kennung, Felder: felder,
		Modell: "schreibt:8b-q4", Aufgabe: "Schreibe den Regeltext.",
	}); fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	return auftrag.Kennung
}

var _ context.Context = context.Background()

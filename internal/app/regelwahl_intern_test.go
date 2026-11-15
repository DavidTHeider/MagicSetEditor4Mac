package app

// **Modus und Abwahl an der Bindung** — PLAY-02.AK-01 und AK-02
// (#1710).
//
// **Eine eigene Datei**, damit `regelfreigabe_intern_test.go` unter der
// Zeilengrenze aus §2 bleibt.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineRegelanwendung baut eine Anwendung mit Bestand und Profil.
func eineRegelanwendung(t *testing.T) *App {
	t.Helper()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)

	return anwendung
}

// **Der Baum kommt an jeder Antwort mit — und er ist vollständig.**
func TestDieBindungReichtDenBaumDurch(t *testing.T) {
	t.Parallel()

	freigabe := eineRegelanwendung(t).Regelfreigabe(
		t.Context(), "magic", "sandkasten", 2, "", nil).Freigabe

	if len(freigabe.Baum) != 2 {
		t.Fatalf("der Baum hat %d Gruppen", len(freigabe.Baum))
	}

	gezaehlt := 0
	for _, gruppe := range freigabe.Baum {
		gezaehlt += len(gruppe.Bereiche)
	}

	if gezaehlt != len(domain.Regelbereiche()) {
		t.Errorf("der Baum führt %d Bereiche, es gibt %d",
			gezaehlt, len(domain.Regelbereiche()))
	}

	// **Ohne Angabe steht der Anfängermodus in der Antwort** — wer nichts
	// sagt, soll sehen, was er bekommen hat.
	if freigabe.Modus != domain.ModusAnfaenger {
		t.Errorf("der Modus ist %q", freigabe.Modus)
	}
}

// **Eine Abwahl im freien Modus ändert die Matrix und wird erklärt.**
func TestDieBindungErklaertEinenWiderspruch(t *testing.T) {
	t.Parallel()

	freigabe := eineRegelanwendung(t).Regelfreigabe(
		t.Context(), "magic", "sandkasten", 2, domain.ModusFrei,
		[]string{string(domain.BereichPrioritaet)}).Freigabe

	if !freigabe.Erlaubt {
		t.Fatalf("die Abwahl wurde abgewiesen: %v", freigabe.Hindernisse)
	}

	if len(freigabe.Widersprueche) == 0 {
		t.Fatal("der Widerspruch fehlt")
	}

	erster := freigabe.Widersprueche[0]

	switch {
	case erster.Bereich != string(domain.BereichPrioritaet):
		t.Errorf("der Widerspruch nennt %q", erster.Bereich)
	case erster.Haengt == "":
		t.Error("der Widerspruch sagt nicht, was daran hängt")
	case !strings.HasPrefix(erster.Quelle, "CR "):
		t.Errorf("die Quelle lautet %q", erster.Quelle)
	}

	// **Und die Matrix zeigt die wirksame Lage** — nicht die des Profils.
	for _, eintrag := range freigabe.Matrix {
		if eintrag.Bereich == string(domain.BereichPrioritaet) &&
			eintrag.Grad != string(domain.UnterstuetzungManuell) {
			t.Errorf("die Priorität steht auf %q", eintrag.Grad)
		}
	}
}

// **Im Anfängermodus ist die Abwahl ein Hindernis** — und der Satz nennt
// den Ausweg.
func TestDieBindungWeistDieAbwahlImAnfaengermodusAb(t *testing.T) {
	t.Parallel()

	freigabe := eineRegelanwendung(t).Regelfreigabe(
		t.Context(), "magic", "sandkasten", 2, domain.ModusAnfaenger,
		[]string{string(domain.BereichCombat)}).Freigabe

	if freigabe.Erlaubt {
		t.Fatal("die Abwahl ging im Anfängermodus durch")
	}

	if !strings.Contains(strings.Join(freigabe.Hindernisse, " "),
		"freien Modus") {
		t.Errorf("das Hindernis nennt den Ausweg nicht: %v",
			freigabe.Hindernisse)
	}
}

// **Ein Bereich, den es nicht gibt, ist ein Fehler und keine Antwort.**
func TestDieBindungWeistEinenUnbekanntenBereichAb(t *testing.T) {
	t.Parallel()

	antwort := eineRegelanwendung(t).Regelfreigabe(
		t.Context(), "magic", "sandkasten", 2, domain.ModusFrei,
		[]string{"zonnen"})

	if antwort.Fehler == nil {
		t.Fatal("der Tippfehler ging durch")
	}

	// **Und die leere Antwort trägt ihre leeren Listen** — dieselbe Zusage
	// wie auf dem Erfolgsweg.
	if antwort.Freigabe.Baum == nil || antwort.Freigabe.Widersprueche == nil {
		t.Error("die leere Antwort trägt eine Liste als nil")
	}
}

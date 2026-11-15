package service_test

// **Die abgewählten Regeln werden mit eingefroren** — PLAY-02.AK-03
// (#1710):
// „Profil und aktive Regeln sind Bestandteil des vereinbarten Partiestands."

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Was die Runde selbst übernimmt, steht danach in der Partie** — im
// Regelwerk **und** in der wirksamen Matrix.
func TestDieAbwahlStehtInDerFertigenPartie(t *testing.T) {
	t.Parallel()

	werk := einRegelwerk()
	werk.Abgewaehlt = []domain.Regelbereich{domain.BereichZonen}

	partie, fehler := einPartieaufbau().Aufbauen(
		context.Background(), "p1", "magic", werk, "sitz-1",
		[]service.Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: 1},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
		})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if len(partie.Regelwerk.Abgewaehlt) != 1 ||
		partie.Regelwerk.Abgewaehlt[0] != domain.BereichZonen {
		t.Errorf("das Regelwerk führt %v", partie.Regelwerk.Abgewaehlt)
	}

	// **Die Matrix der Partie ist die wirksame** — wer sie zeigt, zeigt,
	// was für **diese** Runde gilt.
	if grad := partie.Profil.Matrix.Grad(
		domain.BereichZonen); grad != domain.UnterstuetzungManuell {
		t.Errorf("die Zonen stehen in der Partie auf %q", grad)
	}

	// **Und was nicht abgewählt wurde, bleibt** — eine Abwahl trifft ihren
	// Bereich und keinen zweiten (§10).
	if grad := partie.Profil.Matrix.Grad(
		domain.BereichMarken); grad != domain.UnterstuetzungAutomatisch {
		t.Errorf("die Marken stehen auf %q", grad)
	}

	// **Die Partie bleibt prüfbar** — eine Quelle neben „manuell" wiese
	// [domain.Partie.Pruefe] ab.
	if fehler := partie.Pruefe(); fehler != nil {
		t.Errorf("die Partie mit Abwahl ist ungültig: %v", fehler)
	}
}

// **Ohne Abwahl bleibt alles, wie es war.**
func TestOhneAbwahlBleibtDieMatrixDieDesProfils(t *testing.T) {
	t.Parallel()

	partie, fehler := einPartieaufbau().Aufbauen(
		context.Background(), "p1", "magic", einRegelwerk(), "sitz-1",
		[]service.Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: 1},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
		})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	vorlage := domain.Grundprofil("magic")

	for _, bereich := range domain.Regelbereiche() {
		if partie.Profil.Matrix.Grad(bereich) != vorlage.Matrix.Grad(bereich) {
			t.Errorf("%q steht auf %q statt auf %q", bereich,
				partie.Profil.Matrix.Grad(bereich),
				vorlage.Matrix.Grad(bereich))
		}
	}
}

// **Ein Bereich, den es nicht gibt, kommt gar nicht erst in eine Partie.**
func TestEineUnbekannteAbwahlBautKeinePartie(t *testing.T) {
	t.Parallel()

	werk := einRegelwerk()
	werk.Abgewaehlt = []domain.Regelbereich{"zonnen"}

	_, fehler := einPartieaufbau().Aufbauen(
		context.Background(), "p1", "magic", werk, "sitz-1",
		[]service.Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: 1},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
		})

	if fehler == nil {
		t.Fatal("der Tippfehler baute eine Partie")
	}
}

package domain_test

// **Der Vorrat gehört einem Sitz und einem Abschnitt** — Werkstück M3 der
// Regelmaschine
// (#1605).
//
// # Wogegen hier geprüft wird
//
//   - **Er leert sich am Ende jedes Abschnitts und jeder Phase** — CR 106.4,
//     CR 500.5. Das ist eine zugbasierte Handlung und geht nicht über den
//     Stapel (CR 703.4q).
//   - **Er gehört einem Sitz** und nicht der Partie.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Mana landet beim Sitz, den es angeht.**
func TestManaLandetBeimEigenenSitz(t *testing.T) {
	t.Parallel()

	partie, fehler := einePartie().ManaHinzu("sitz-1", domain.ManaRot, 2)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	switch {
	case partie.Sitze[0].Manavorrat.Zahl(domain.ManaRot) != 2:
		t.Errorf("sitz-1 hat %v", partie.Sitze[0].Manavorrat)
	case !partie.Sitze[1].Manavorrat.Leer():
		t.Errorf("sitz-2 hat %v", partie.Sitze[1].Manavorrat)
	}
}

// **Einen Sitz, den es nicht gibt, gibt es nicht.**
func TestManaFuerEinenUnbekanntenSitzWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := einePartie().ManaHinzu("gibtsnicht", domain.ManaRot, 1); fehler == nil {
		t.Error("ein unbekannter Sitz bekam Mana")
	}
}

// **Am Ende des Abschnitts ist der Vorrat leer** — CR 500.5.
//
// **Über [domain.Partie.Weiter] und nicht über einen Ruf an das Leeren
// selbst**: die Abnahme des Issues lautet „der Vorrat ist am Ende des
// Abschnitts leer", und das ist eine Aussage über den **Weg** (§3).
func TestDerVorratIstAmEndeDesAbschnittsLeer(t *testing.T) {
	t.Parallel()

	partie := einePartie().ZugBeginnen("sitz-1")

	partie, fehler := partie.ManaHinzu("sitz-1", domain.ManaGruen, 3)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	vorher := partie.Zug.Abschnitt

	danach := partie.Weiter()

	switch {
	case danach.Zug.Abschnitt == vorher:
		t.Fatalf("der Abschnitt steht noch bei %+v", danach.Zug.Abschnitt)
	case !danach.Sitze[0].Manavorrat.Leer():
		t.Errorf("der Vorrat steht noch bei %v", danach.Sitze[0].Manavorrat)
	}
}

// **Und die Gegenprobe: ohne Abschnittswechsel bleibt er stehen.**
//
// **Ohne sie sagte der Lauf darüber wenig** (§10): ein Weg, der den Vorrat
// bei **jeder** Gelegenheit leerte, käme genauso durch — und dann wäre Mana
// nie zu gebrauchen.
func TestOhneAbschnittswechselBleibtDerVorratStehen(t *testing.T) {
	t.Parallel()

	partie, fehler := einePartie().ZugBeginnen("sitz-1").
		ManaHinzu("sitz-1", domain.ManaGruen, 3)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	danach := partie.Zustandspruefung()

	if danach.Sitze[0].Manavorrat.Zahl(domain.ManaGruen) != 3 {
		t.Errorf("der Vorrat steht bei %v", danach.Sitze[0].Manavorrat)
	}
}

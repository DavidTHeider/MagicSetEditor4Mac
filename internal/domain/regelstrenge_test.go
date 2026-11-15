package domain_test

// **Welche Strenge eine Matrix trägt** — Werkstück G4 von Phase 5.5
// (#1461).
//
// **Eine eigene Datei neben `regelmatrix_test.go`**, weil jene über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht, **was**
// die Matrix ist, hier, **was sie trägt**.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Streng trägt nur, wenn jeder Bereich durchgesetzt wird.**
//
// # Die Gegenprobe, die den Satz erst zu einem macht
//
// **Ein einziger Bereich weniger kippt es.** Ohne sie sagte der Lauf nur, dass
// eine volle Matrix trägt — nicht, dass eine fast volle es nicht tut (§10).
func TestStrengTraegtNurEineVolleMatrix(t *testing.T) {
	t.Parallel()

	voll := volleMatrix()

	if !voll.TraegtStrenge(domain.StrengeStreng) {
		t.Fatalf("eine volle Matrix trägt streng nicht — Lücken: %v",
			voll.Luecken())
	}

	if len(voll.Luecken()) != 0 {
		t.Errorf("Lücken in einer vollen Matrix: %v", voll.Luecken())
	}

	// **Ein Bereich zurück auf begleitet — und streng fällt.**
	for i := range voll.Eintraege {
		fast := volleMatrix()
		fast.Eintraege[i].Grad = domain.UnterstuetzungBegleitet
		fast.Eintraege[i].Grenzen = nil

		if fast.TraegtStrenge(domain.StrengeStreng) {
			t.Errorf("streng trägt, obwohl %q nur begleitet ist",
				fast.Eintraege[i].Bereich)
		}

		if luecken := fast.Luecken(); len(luecken) != 1 {
			t.Errorf("%d Lücken, erwartet eine: %v", len(luecken), luecken)
		}
	}
}

// **Die Lücken stehen in der Reihenfolge der Bereiche.**
//
// **Und nicht in der der Einträge**: eine Liste, deren Reihenfolge von der
// Eingabe abhängt, liest sich bei jedem Aufruf anders — und ein Vergleich
// darüber flackerte.
func TestDieLueckenStehenInFesterReihenfolge(t *testing.T) {
	t.Parallel()

	verdreht := domain.Regelmatrix{Eintraege: []domain.Matrixeintrag{
		{
			Bereich: domain.BereichTypen,
			Grad:    domain.UnterstuetzungAutomatisch,
			Quelle:  eineQuelle("205"),
			Grenzen: []string{"keine eigenen Typen"},
		},
		{
			Bereich: domain.BereichZonen,
			Grad:    domain.UnterstuetzungAutomatisch,
			Quelle:  eineQuelle("400"),
			Grenzen: []string{"keine Zusatzzonen"},
		},
	}}

	luecken := verdreht.Luecken()
	if len(luecken) == 0 {
		t.Fatal("keine Lücken")
	}

	vorige := -1

	for _, lücke := range luecken {
		stelle := -1

		for i, bereich := range domain.Regelbereiche() {
			if bereich == lücke {
				stelle = i
			}
		}

		if stelle <= vorige {
			t.Errorf("%q steht an Stelle %d nach %d", lücke, stelle, vorige)
		}

		vorige = stelle
	}
}

// **Wer durchsetzt, belegt es — und nennt seine Grenzen.**
func TestEineZusageBrauchtQuelleUndGrenzen(t *testing.T) {
	t.Parallel()

	faelle := map[string]domain.Matrixeintrag{
		"automatisch ohne Quelle": {
			Bereich: domain.BereichZonen,
			Grad:    domain.UnterstuetzungAutomatisch,
			Grenzen: []string{"eine"},
		},
		"automatisch ohne Grenzen": {
			Bereich: domain.BereichZonen,
			Grad:    domain.UnterstuetzungAutomatisch,
			Quelle:  eineQuelle("400"),
		},
		"begleitet ohne Quelle": {
			Bereich: domain.BereichZonen,
			Grad:    domain.UnterstuetzungBegleitet,
		},
		"unbekannt mit Quelle": {
			Bereich: domain.BereichZonen,
			Grad:    domain.UnterstuetzungUnbekannt,
			Quelle:  eineQuelle("400"),
		},
		"fremder Bereich": {
			Bereich: "raumzeit",
			Grad:    domain.UnterstuetzungAutomatisch,
			Quelle:  eineQuelle("400"),
			Grenzen: []string{"eine"},
		},
		"fremder Grad": {
			Bereich: domain.BereichZonen,
			Grad:    "halbwegs",
		},
	}

	for name, eintrag := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			matrix := domain.Regelmatrix{Eintraege: []domain.Matrixeintrag{eintrag}}

			fehler := matrix.Pruefe()
			if fehler == nil {
				t.Fatal("kommt durch")
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler %v ist keine Validierung", fehler)
			}
		})
	}

	// **Und die Gegenprobe: `manuell` ohne Quelle kommt durch.**
	gut := domain.Regelmatrix{Eintraege: []domain.Matrixeintrag{{
		Bereich: domain.BereichCombat, Grad: domain.UnterstuetzungManuell,
	}}}

	if fehler := gut.Pruefe(); fehler != nil {
		t.Errorf("eine ehrliche Nichtzusage wird abgewiesen: %v", fehler)
	}
}

// **Ein Bereich steht nur einmal.**
//
// **Zwei Einträge für denselben Bereich wären zwei Antworten auf eine
// Frage** — und welche gälte, entschiede die Reihenfolge.
func TestEinBereichStehtNurEinmal(t *testing.T) {
	t.Parallel()

	doppelt := domain.Regelmatrix{Eintraege: []domain.Matrixeintrag{
		{Bereich: domain.BereichZonen, Grad: domain.UnterstuetzungManuell},
		{Bereich: domain.BereichZonen, Grad: domain.UnterstuetzungManuell},
	}}

	if fehler := doppelt.Pruefe(); fehler == nil {
		t.Fatal("kommt durch")
	}
}

// **Die Aufzählungen trennen.**
func TestDieAufzaehlungenDerMatrixTrennen(t *testing.T) {
	t.Parallel()

	// **Dreizehn seit `M19`** (#1620):
	// der Commanderbereich. Er steht in **jeder** Aufzählung, weil eine
	// Matrix ohne ihn sagte, die Variante sei geprüft und ohne Befund — und
	// ein Bereich ohne Eintrag ist `unbekannt` und damit ehrlich.
	if len(domain.Regelbereiche()) != 13 {
		t.Errorf("%d Bereiche", len(domain.Regelbereiche()))
	}

	for _, bereich := range domain.Regelbereiche() {
		if !domain.IstRegelbereich(bereich) {
			t.Errorf("%q gilt nicht", bereich)
		}
	}

	if domain.IstRegelbereich("raumzeit") {
		t.Error("ein erfundener Bereich gilt")
	}

	if len(domain.Unterstuetzungsgrade()) != 4 {
		t.Errorf("%d Grade", len(domain.Unterstuetzungsgrade()))
	}

	if domain.IstUnterstuetzung("halbwegs") {
		t.Error("ein erfundener Grad gilt")
	}
}

// **Ein Strengegrad, den es nicht gibt, wird nicht getragen.**
//
// **Der sichere Ausgang**: ein Tippfehler darf keine Freigabe sein.
func TestEinUnbekannterStrengegradWirdNichtGetragen(t *testing.T) {
	t.Parallel()

	if volleMatrix().TraegtStrenge("turnier") {
		t.Error("ein unbekannter Grad wird getragen")
	}
}

// **Eine vollständige Matrix kommt durch, und ihre Einträge lassen sich
// finden.**
//
// # Warum der Erfolgsweg einen eigenen Lauf bekommt
//
// **Weil eine Prüfung, die nur Abweisungen kennt, auch alles abweisen
// könnte** (§10: „zum Ergebnis gehört dabei auch das Nein" — und umgekehrt).
// Die Läufe darüber zeigen, was nicht durchkommt; dieser zeigt, dass
// überhaupt etwas durchkommt.
func TestEineVollstaendigeMatrixKommtDurch(t *testing.T) {
	t.Parallel()

	voll := volleMatrix()

	if fehler := voll.Pruefe(); fehler != nil {
		t.Fatalf("eine vollständige Matrix wird abgewiesen: %v", fehler)
	}

	eintrag, da := voll.Eintrag(domain.BereichCombat)
	if !da {
		t.Fatal("der Kampfabschnitt steht nicht darin")
	}

	switch {
	case eintrag.Bereich != domain.BereichCombat:
		t.Errorf("gefunden wurde %q", eintrag.Bereich)
	case eintrag.Grad != domain.UnterstuetzungAutomatisch:
		t.Errorf("Grad %q", eintrag.Grad)
	case len(eintrag.Grenzen) == 0:
		t.Error("ohne Grenzen")
	}

	// **Und eine leere Matrix kommt auch durch** — sie sagt nur nichts zu.
	if fehler := (domain.Regelmatrix{}).Pruefe(); fehler != nil {
		t.Errorf("eine leere Matrix wird abgewiesen: %v", fehler)
	}
}

// **Begleitet trägt, sobald über einen Bereich etwas bekannt ist.**
//
// # Die Zwischenstufe, und warum sie eine ist
//
// **Gewarnt wird über das, worüber es etwas zu sagen gibt.** Streng verlangt
// jeden Bereich; begleitet verlangt **einen**. Dazwischen gibt es nichts —
// wer über die Hälfte warnt, warnt.
//
// **`manuell` genügt nicht**, und das ist der Kern: dort hat jemand
// entschieden, es **nicht** zu bauen — das Programm schweigt. Eine Strenge
// „begleitet", unter der nichts begleitet, wäre eine Zusage ohne Einlösung.
func TestBegleitetTraegtNurWoDasProgrammEtwasSagt(t *testing.T) {
	t.Parallel()

	for name, grad := range map[string]domain.Unterstuetzung{
		"begleitet":   domain.UnterstuetzungBegleitet,
		"automatisch": domain.UnterstuetzungAutomatisch,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			matrix := domain.Regelmatrix{
				Eintraege: []domain.Matrixeintrag{{
					Bereich: domain.BereichCombat,
					Grad:    grad,
					Quelle:  eineQuelle("506"),
					Grenzen: []string{"keine Karte mit eigenem Text"},
				}},
			}

			if !matrix.TraegtStrenge(domain.StrengeBegleitet) {
				t.Error("begleitet wird nicht getragen")
			}

			// **Und streng trägt trotzdem nicht** — zwölf Bereiche fehlen.
			if matrix.TraegtStrenge(domain.StrengeStreng) {
				t.Error("streng wird getragen")
			}
		})
	}
}

// **Und `manuell` allein trägt nichts außer dem Sandkasten.**
//
// **Die Gegenprobe zum Lauf darüber** (§10): ohne sie sagte er nur, dass
// `begleitet` und `automatisch` tragen — nicht, dass `manuell` es **nicht**
// tut, und genau darum geht es.
func TestManuellTraegtNurDenSandkasten(t *testing.T) {
	t.Parallel()

	nurManuell := domain.Regelmatrix{Eintraege: []domain.Matrixeintrag{{
		Bereich: domain.BereichCombat, Grad: domain.UnterstuetzungManuell,
	}}}

	switch {
	case !nurManuell.TraegtStrenge(domain.StrengeSandkasten):
		t.Error("der Sandkasten wird nicht getragen")
	case nurManuell.TraegtStrenge(domain.StrengeBegleitet):
		t.Error("begleitet wird getragen, obwohl das Programm schweigt")
	case nurManuell.TraegtStrenge(domain.StrengeStreng):
		t.Error("streng wird getragen")
	}
}

// **Die Strengegrade stehen von schwach nach stark.**
//
// # Wer sich darauf verlässt
//
// **`service.hoechsteStrenge` geht die Liste entlang und behält die letzte,
// die trägt.** Wer die Reihenfolge umdreht, bekommt für jedes Profil den
// Sandkasten als höchste Stufe — still, denn alle drei sind gültige Werte.
func TestDieStrengegradeStehenVonSchwachNachStark(t *testing.T) {
	t.Parallel()

	erwartet := []domain.Strenge{
		domain.StrengeSandkasten, domain.StrengeBegleitet, domain.StrengeStreng,
	}

	grade := domain.Strengegrade()
	if len(grade) != len(erwartet) {
		t.Fatalf("%d Grade", len(grade))
	}

	for i, grad := range grade {
		if grad != erwartet[i] {
			t.Errorf("an Stelle %d steht %q, erwartet %q", i, grad, erwartet[i])
		}
	}

	// **Und die Ordnung ist die der Matrix**: was eine stärkere Stufe
	// trägt, trägt auch jede schwächere.
	voll := volleMatrix()

	for _, grad := range grade {
		if !voll.TraegtStrenge(grad) {
			t.Errorf("eine volle Matrix trägt %q nicht", grad)
		}
	}
}

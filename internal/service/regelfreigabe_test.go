package service_test

// **Die Regelfreigabe** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Wogegen hier geprüft wird
//
// **Dass „streng" nicht zur Wahl steht, solange acht Bereiche von Hand
// gemacht werden** — und dass die Absage sagt, wie viele es sind.
//
// **Dass die höchste Stufe auch dann dasteht, wenn die gewünschte durchgeht.**
// Wer „begleitet" bekommt, soll sehen, dass „streng" nicht geht — sonst
// erfährt er es erst beim Versuch.
//
// **Und dass eine Lücke etwas anderes ist als ein Konflikt**: die eine
// verhindert nur die Strenge, der andere den Anfang.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einFreigabewunsch gibt einen Freigabewunsch, der durchgeht.
func einFreigabewunsch() service.Freigabewunsch {
	return service.Freigabewunsch{
		Profil:  domain.Grundprofil("magic"),
		Strenge: domain.StrengeSandkasten,
		Sitze:   2,
	}
}

// **Der Sandkasten geht, streng nicht — und die Absage zählt die Lücken.**
func TestDerSandkastenGehtUndStrengNicht(t *testing.T) {
	t.Parallel()

	freigabe, fehler := service.NeueRegelfreigabe().Pruefe(einFreigabewunsch())
	if fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}

	switch {
	case !freigabe.Erlaubt:
		t.Errorf("der Sandkasten geht nicht: %v", freigabe.Hindernisse)
	case freigabe.Hoechste != domain.StrengeBegleitet:
		t.Errorf("höchste Stufe %q, erwartet begleitet", freigabe.Hoechste)
	// **Vier seit `M19`** (#1620):
	// der Commanderbereich ist der dreizehnte.
	//
	// **Drei seit `M9`** (#1611):
	// die **Wechselwirkungen** sind dazugekommen.
	//
	// **Und zwei seit dem 2026-09-19**
	// ([0123](../../docs/entscheidungen/0123-streng-zaehlt-nur-was-zur-variante-gehoert.md),
	// #1626):
	// der Commander zählt im Grundprofil nicht mehr mit — die Variante wird
	// dort nicht gespielt. Übrig sind Spielsteine und Typen, und beide sind
	// Arbeit.
	case len(freigabe.Luecken) != 2:
		t.Errorf("%d Lücken: %v", len(freigabe.Luecken), freigabe.Luecken)
	}

	// **Und streng wird abgewiesen, mit einer Zahl im Satz.**
	streng := einFreigabewunsch()
	streng.Strenge = domain.StrengeStreng

	abgewiesen, fehler := service.NeueRegelfreigabe().Pruefe(streng)
	if fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}

	if abgewiesen.Erlaubt {
		t.Fatal("streng geht")
	}

	if len(abgewiesen.Hindernisse) != 1 {
		t.Fatalf("%d Hindernisse: %v",
			len(abgewiesen.Hindernisse), abgewiesen.Hindernisse)
	}

	satz := abgewiesen.Hindernisse[0]
	if !strings.Contains(satz, "2 von 13") {
		t.Errorf("%q nennt die Zahl der Lücken nicht", satz)
	}
}

// **Die höchste Stufe steht auch bei einer glatten Freigabe da.**
//
// **Sonst erführe jemand erst beim Versuch, dass „streng" nicht geht** — und
// das Issue verlangt das Gegenteil: „Spieler sehen vor Start, welche Regeln
// automatisch durchgesetzt werden."
func TestDieHoechsteStufeStehtAuchBeiEinerFreigabeDa(t *testing.T) {
	t.Parallel()

	wunsch := einFreigabewunsch()
	wunsch.Strenge = domain.StrengeBegleitet

	freigabe, fehler := service.NeueRegelfreigabe().Pruefe(wunsch)
	if fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}

	switch {
	case !freigabe.Erlaubt:
		t.Errorf("begleitet geht nicht: %v", freigabe.Hindernisse)
	case freigabe.Hoechste != domain.StrengeBegleitet:
		t.Errorf("höchste Stufe %q", freigabe.Hoechste)
	case len(freigabe.Luecken) == 0:
		t.Error("keine Lücken bei einer Freigabe, die nur begleitet trägt")
	}
}

// **Zu viele Sitze sind ein Hindernis, und der Satz nennt die Spanne.**
func TestZuVieleSitzeSindEinHindernis(t *testing.T) {
	t.Parallel()

	wunsch := einFreigabewunsch()
	wunsch.Sitze = 5

	freigabe, fehler := service.NeueRegelfreigabe().Pruefe(wunsch)
	if fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}

	if freigabe.Erlaubt {
		t.Fatal("fünf Sitze gehen")
	}

	satz := strings.Join(freigabe.Hindernisse, " ")
	for _, teil := range []string{"1 bis 4", "gewünscht sind 5"} {
		if !strings.Contains(satz, teil) {
			t.Errorf("%q nennt %q nicht", satz, teil)
		}
	}
}

// **Eine Hausregel, die das Profil nicht erlaubt, hält den Anfang auf.**
//
// **Der Unterschied zur Lücke**: eine Lücke verhindert nur die Strenge, ein
// Konflikt den Anfang. Sonst wäre die Leiter aus `domain.Regelrang` eine
// Empfehlung.
func TestEineVerbotenHausregelHaeltDenAnfangAuf(t *testing.T) {
	t.Parallel()

	wunsch := einFreigabewunsch()
	wunsch.Profil.ErlaubteHausregeln = []domain.Hausregelart{domain.HausregelProxy}
	wunsch.Hausregeln = []domain.Hausregel{
		{Art: domain.HausregelProxy},
		{Art: domain.HausregelOffeneHaende},
	}

	freigabe, fehler := service.NeueRegelfreigabe().Pruefe(wunsch)
	if fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}

	if freigabe.Erlaubt {
		t.Fatal("die verbotene Hausregel kommt durch")
	}

	// **Nur die eine, nicht beide.**
	if len(freigabe.Hindernisse) != 1 {
		t.Fatalf("%d Hindernisse: %v",
			len(freigabe.Hindernisse), freigabe.Hindernisse)
	}

	if !strings.Contains(freigabe.Hindernisse[0], "offene-haende") {
		t.Errorf("%q nennt die Regel nicht", freigabe.Hindernisse[0])
	}
}

// **Ein widersprüchliches Profil ist ein Fehler und kein Nein.**
//
// **Der Unterschied zählt**: ein Nein zeigt das Fenster als Hinweis, ein
// Fehler als Fehler — und ein Profil, das sich selbst widerspricht, ist
// nichts, was jemand durch eine andere Wahl behebt.
func TestEinWidersprechendesProfilIstEinFehler(t *testing.T) {
	t.Parallel()

	wunsch := einFreigabewunsch()
	wunsch.Profil.Fassung = 0

	_, fehler := service.NeueRegelfreigabe().Pruefe(wunsch)
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("%v ist keine Validierung", fehler)
	}

	// **Und eine Strenge, die es nicht gibt, ebenso.**
	falsch := einFreigabewunsch()
	falsch.Strenge = "turnier"

	if _, fehler := service.NeueRegelfreigabe().Pruefe(falsch); fehler == nil {
		t.Error("eine erfundene Strenge kommt durch")
	}
}

// **Ein Profil, das nichts kann, trägt nur den Sandkasten.**
//
// **Die Gegenprobe zur höchsten Stufe**: ohne sie sagte der Lauf nur, dass
// das Grundprofil `begleitet` erreicht — nicht, dass die Rechnung überhaupt
// unterscheidet (§10).
func TestEinProfilOhneZusagenTraegtNurDenSandkasten(t *testing.T) {
	t.Parallel()

	wunsch := einFreigabewunsch()
	wunsch.Profil.Matrix = domain.Regelmatrix{}

	freigabe, fehler := service.NeueRegelfreigabe().Pruefe(wunsch)
	if fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}

	switch {
	case !freigabe.Erlaubt:
		t.Errorf("der Sandkasten geht nicht: %v", freigabe.Hindernisse)
	case freigabe.Hoechste != domain.StrengeSandkasten:
		t.Errorf("höchste Stufe %q, erwartet sandkasten", freigabe.Hoechste)
	case len(freigabe.Luecken) != 13:
		t.Errorf("%d Lücken", len(freigabe.Luecken))
	}
}

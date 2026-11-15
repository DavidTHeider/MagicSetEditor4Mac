package domain_test

// **Was das Grundprofil seit G9 durchsetzt** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Wogegen hier geprüft wird
//
// **Dass aus zwei Bereichen `manuell` ein `automatisch` geworden ist** — und
// dass beide ihre **Grenzen** nennen. Ein Grad ohne Grenzen behauptet
// Vollständigkeit; genau das ist die Gestalt aus §10, und die Matrix aus G4
// gibt es, damit sie nicht entsteht.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Zehn von dreizehn Bereichen sind automatisch — drei waren es seit G4.**
//
// **Fünf seit G9** (#1592: Zugstruktur und Priorität), **sechs seit `M5`**
// (#1607):
// die zustandsbasierten Aktionen. **Sieben seit `M3`**
// (#1605):
// der Manavorrat — und er ist zugleich der **zwölfte** Bereich. **Acht seit
// `M4`** (#1606):
// der Stapel, auf den ein Zauber kommt. **Neun seit `M6`**
// (#1608):
// der Kampf. **Zehn seit `M9`**
// (#1611):
// die **Wechselwirkungen** — eine Fähigkeit löst aus, wartet, kommt auf den
// Stapel und wird verrechnet, ohne dass jemand etwas stellt.
func TestZehnBereicheSindAutomatisch(t *testing.T) {
	t.Parallel()

	matrix := domain.Grundprofil("magic").Matrix

	automatisch := []domain.Regelbereich{}

	for _, eintrag := range matrix.Eintraege {
		if eintrag.Grad == domain.UnterstuetzungAutomatisch {
			automatisch = append(automatisch, eintrag.Bereich)
		}
	}

	if len(automatisch) != 10 {
		t.Fatalf("automatisch sind %d Bereiche: %v", len(automatisch), automatisch)
	}

	for _, bereich := range []domain.Regelbereich{
		domain.BereichZugstruktur, domain.BereichPrioritaet,
		domain.BereichWechselwirkungen,
	} {
		gefunden := false

		for _, hat := range automatisch {
			if hat == bereich {
				gefunden = true
			}
		}

		if !gefunden {
			t.Errorf("%q ist nicht automatisch", bereich)
		}
	}
}

// **Und jeder automatische Bereich nennt Quelle und Grenzen.**
//
// **Beides, und die Grenzen sind der schwierigere Teil**: eine Quelle sagt,
// woher die Regel kommt; eine Grenze sagt, was **nicht** geschieht — und nur
// die zweite verhindert, dass jemand mehr erwartet, als das Programm hält.
func TestJederAutomatischeBereichNenntQuelleUndGrenzen(t *testing.T) {
	t.Parallel()

	for _, eintrag := range domain.Grundprofil("magic").Matrix.Eintraege {
		if eintrag.Grad != domain.UnterstuetzungAutomatisch {
			continue
		}

		switch {
		case len(eintrag.Quelle.Nummern) == 0:
			t.Errorf("%q nennt keine Regelnummer", eintrag.Bereich)
		case len(eintrag.Grenzen) == 0:
			t.Errorf("%q nennt keine Grenze", eintrag.Bereich)
		}
	}
}

// **Die Fassung steigt, wenn sich die Matrix ändert.**
//
// **Sonst gäbe es zwei Profile mit derselben Nummer** — und eine Partie, die
// unter der alten angefangen hat, ließe sich nicht mehr von einer unter der
// neuen unterscheiden (G4, #1461).
func TestDieFassungDesGrundprofilsIstZweiundzwanzig(t *testing.T) {
	t.Parallel()

	if domain.Grundprofilfassung != 22 {
		t.Errorf("die Fassung ist %d", domain.Grundprofilfassung)
	}
}

// **Und keine Grenze behauptet mehr, was nicht mehr stimmt.**
//
// # Der Befund, für den dieser Wächter steht
//
// **Gesehen am 2026-09-17** in der Regelfreigabe einer laufenden Partie:
// dort stand „kein Stapel, also keine Verrechnung (CR 608)" und „keine
// Abkürzungen (CR 732)", während `M1`
// (#1603)
// beides gebaut hatte. **Eine Grenze, die nicht mehr gilt, ist eine falsche
// Zusage** — und sie steht dort, wo jemand vor dem Anfangen nachliest.
//
// **Was er nicht hält**: ob eine Grenze stimmt. Er hält die zwei Sätze fest,
// die einmal falsch waren.
func TestKeineGrenzeBehauptetWasM1GebautHat(t *testing.T) {
	t.Parallel()

	for _, eintrag := range domain.Grundprofil("magic").Matrix.Eintraege {
		for _, grenze := range eintrag.Grenzen {
			switch {
			case strings.Contains(grenze, "kein Stapel"):
				t.Errorf("%q behauptet, es gebe keinen Stapel", grenze)
			case strings.Contains(grenze, "keine Abkürzungen"):
				t.Errorf("%q behauptet, es gebe keine Abkürzungen", grenze)
			}
		}
	}
}

// **Und der Friedhof wird nicht mehr als ungeordnet geführt.**
//
// **CR 400.5** nennt drei Zonen, deren Reihenfolge sich nur ändern lässt,
// wenn eine Regel oder ein Effekt es erlaubt — Bibliothek, **Friedhof** und
// Stapel. Bis zum 2026-09-17 sagte `Zone.Geordnet` das Gegenteil, das
// gezeichnete Tischbild sagte es richtig, und die Matrix führte den
// Widerspruch als Grenze mit dem Vermerk, G9 löse ihn auf.
func TestDieMatrixFuehrtDenFriedhofNichtMehrAlsUngeordnet(t *testing.T) {
	t.Parallel()

	for _, eintrag := range domain.Grundprofil("magic").Matrix.Eintraege {
		for _, grenze := range eintrag.Grenzen {
			if strings.Contains(grenze, "Friedhof wird als ungeordnet") {
				t.Errorf("%q führt die Grenze noch: %q", eintrag.Bereich, grenze)
			}
		}
	}

	if !domain.ZoneFriedhof.Geordnet() {
		t.Error("der Friedhof gilt weiter als ungeordnet")
	}
}

// **Und keine Grenze behauptet, was M12 gebaut hat.**
//
// **Derselbe Wächter, ein Werkstück später** (#1614): „kein Spielstein wird
// erzeugt" stand als Grenze bei `spielsteine`, und „keine Legendenregel"
// bei `zustandsbasiert` — beides bis zum 2026-09-17 wahr und seitdem nicht
// mehr. Eine Grenze, die nicht mehr gilt, ist eine falsche Zusage, und sie
// steht dort, wo jemand vor dem Anfangen nachliest.
func TestKeineGrenzeBehauptetWasM12GebautHat(t *testing.T) {
	t.Parallel()

	for _, eintrag := range domain.Grundprofil("magic").Matrix.Eintraege {
		for _, grenze := range eintrag.Grenzen {
			switch {
			case strings.Contains(grenze, "kein Spielstein wird erzeugt"):
				t.Errorf("%q behauptet, es entstehe kein Spielstein", grenze)
			case strings.Contains(grenze, "keine Legendenregel"):
				t.Errorf("%q behauptet, es gebe keine Legendenregel", grenze)
			}
		}
	}
}

// **Und die Grenze nennt so viele Wirkungen, wie es gibt.**
//
// # Der Befund, für den dieser Wächter steht
//
// **Gesehen am 2026-09-18**: die Grenze sagte „neun Wirkungen", während
// [domain.Wirkungsarten] dreizehn führte. `M10` und `M11` hatten vier
// dazugelegt, und vier Stellen im Baum waren mitgezogen worden — diese
// nicht. **Eine Zahl ohne Wächter veraltet still**, und das ist dieselbe
// Gestalt wie bei der Zahl der Bindungen (§16).
//
// **Was er nicht hält**: ob die dreizehn die richtigen sind, und ob die
// Grenze im Übrigen stimmt. Er hält die Zahl.
func TestDieGrenzeNenntSoVieleWirkungenWieEsGibt(t *testing.T) {
	t.Parallel()

	anzahl := len(domain.Wirkungsarten())

	wort := map[int]string{
		8: "acht", 9: "neun", 10: "zehn", 11: "elf", 12: "zwölf",
		13: "dreizehn", 14: "vierzehn", 15: "fünfzehn", 16: "sechzehn",
		17: "siebzehn", 18: "achtzehn", 19: "neunzehn", 20: "zwanzig",
	}[anzahl]

	if wort == "" {
		t.Fatalf("für %d Wirkungen kennt dieser Wächter kein Zahlwort — "+
			"wer eine hinzufügt, trägt es hier ein", anzahl)
	}

	const anfang = "ein Spielskript kennt "

	for _, eintrag := range domain.Grundprofil("magic").Matrix.Eintraege {
		for _, grenze := range eintrag.Grenzen {
			if !strings.HasPrefix(grenze, anfang) {
				continue
			}

			if !strings.HasPrefix(grenze, anfang+wort+" Wirkungen") {
				t.Errorf("die Grenze lautet %q — es sind aber %d Wirkungen",
					grenze, anzahl)
			}

			return
		}
	}

	t.Error("keine Grenze der Matrix nennt die Zahl der Wirkungen")
}

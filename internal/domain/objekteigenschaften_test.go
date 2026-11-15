package domain_test

// **Was am Tisch gilt** — Werkstück M2 der Regelmaschine
// (#1604).
//
// # Wogegen hier geprüft wird
//
//   - CR 122.1a — +1/+1 und −1/−1 ändern Stärke und Widerstandskraft.
//   - CR 122.3 mit CR 704.5q — sie heben sich gegeneinander auf.
//   - CR 208.3 — eine Marke auf einem Land macht daraus keine 1/1.
//   - Und die Zusage aus G1/G2: **ein geheimes Objekt gibt nichts her**.
//
// **Eine eigene Datei, und der Schnitt folgt der Frage** (§2): nebenan steht,
// wie aus Feldern eine Karte wird — hier, was aus einer Karte auf dem Tisch
// wird.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineKartenpartie gibt eine Partie mit **einer** echten Karte im Abzug.
func eineKartenpartie() domain.Partie {
	partie := einePartie()
	partie.ID = "p1"
	partie.Sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{{
			Kennung: "k1", Name: "Grizzlybär", Menge: 1,
			Seiten: []domain.Abzugsseite{{
				Name: "Grizzlybär",
				Felder: map[string]string{
					"name": "Grizzlybär", "casting cost": "{1}{G}",
					"super type": "Creature", "sub type": "Bear",
					"power": "2", "toughness": "3",
				},
			}},
		}},
	}
	partie.Objekte = []domain.Spielobjekt{{
		Kennung: "sitz-1/k1/0", Art: domain.ObjektKarte, Karte: "k1",
		Beherrscher: "sitz-1", Eigner: "sitz-1", Zone: domain.ZoneTisch,
	}}

	return partie
}

// **Die zweite Abnahme des Issues**: eine Kreatur 2/3 mit zwei
// +1/+1-Marken ist 4/5 — gerechnet, nicht getippt (CR 122.1a).
func TestMarkenAendernStaerkeUndWiderstandskraft(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte[0].Zaehler = map[string]int{domain.MarkePlusEins: 2}

	staerke, widerstand := partie.Werte(partie.Objekte[0])

	if staerke.Zahl != 4 || widerstand.Zahl != 5 {
		t.Fatalf("%d/%d statt 4/5", staerke.Zahl, widerstand.Zahl)
	}
}

// **Minusmarken ziehen ab** — CR 122.1a, zweiter Satz.
func TestMinusmarkenZiehenAb(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte[0].Zaehler = map[string]int{domain.MarkeMinusEins: 1}

	staerke, widerstand := partie.Werte(partie.Objekte[0])

	if staerke.Zahl != 1 || widerstand.Zahl != 2 {
		t.Fatalf("%d/%d statt 1/2", staerke.Zahl, widerstand.Zahl)
	}
}

// **Plus und Minus heben sich auf** — CR 122.3 mit CR 704.5q: N von jeder
// Art werden entfernt, wo N die kleinere Zahl ist.
func TestPlusUndMinusMarkenHebenSichAuf(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte[0].Zaehler = map[string]int{
		domain.MarkePlusEins: 3, domain.MarkeMinusEins: 1,
	}

	danach := partie.Zustandspruefung()

	zaehler := danach.Objekte[0].Zaehler
	if zaehler[domain.MarkePlusEins] != 2 ||
		zaehler[domain.MarkeMinusEins] != 0 {
		t.Fatalf("Zähler = %+v", zaehler)
	}
}

// **Und verrechnet wird nur auf dem Tisch** — CR 704.5q spricht vom
// Permanent, und ein Permanent liegt auf dem Schlachtfeld (CR 110.1).
func TestMarkenAusserhalbDesTischesWerdenNichtVerrechnet(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte[0].Zone = domain.ZoneFriedhof
	partie.Objekte[0].Zaehler = map[string]int{
		domain.MarkePlusEins: 3, domain.MarkeMinusEins: 1,
	}

	danach := partie.Zustandspruefung()

	if danach.Objekte[0].Zaehler[domain.MarkePlusEins] != 3 {
		t.Fatalf("Zähler = %+v", danach.Objekte[0].Zaehler)
	}
}

// **Eine Marke auf einem Land macht daraus keine 1/1** — CR 208.3 mit
// CR 613.4c: die Marke liegt über den gedruckten Werten, und wo keine
// stehen, liegt sie über nichts.
func TestEineMarkeAufEinemLandGibtKeineWerte(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Sitze[0].Deck.Karten[0].Seiten[0].Felder = map[string]string{
		"name": "Wald", "super type": "Basic Land", "sub type": "Forest",
	}
	partie.Objekte[0].Zaehler = map[string]int{domain.MarkePlusEins: 1}

	staerke, widerstand := partie.Werte(partie.Objekte[0])

	if staerke.Steht || widerstand.Steht {
		t.Fatalf("Werte = %+v / %+v", staerke, widerstand)
	}
}

// **Ein Objekt ohne Karte im Abzug hat keine Eigenschaften.**
//
// Ein Spielstein ist der gewöhnliche Fall (`Karte` ist leer); ein Verweis ins
// Leere kommt aus einem von Hand geänderten Bestand. Beide brechen nichts.
func TestEinObjektOhneKarteHatKeineEigenschaften(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte[0].Karte = "gibtesnicht"

	if _, da := partie.Eigenschaften(partie.Objekte[0]); da {
		t.Fatal("ein Verweis ins Leere hat Eigenschaften geliefert")
	}
}

// **Ein Objekt eines Sitzes, den es nicht gibt, hat keine Eigenschaften.**
//
// Er entsteht, wenn ein Sitz die Partie verlassen hat (CR 800.4a) und ein
// Objekt zurückbleibt, das ihm gehörte — oder aus einem von Hand geänderten
// Bestand.
func TestEinObjektOhneSitzHatKeineEigenschaften(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte[0].Eigner = "geist"

	if _, da := partie.Eigenschaften(partie.Objekte[0]); da {
		t.Fatal("ein Objekt ohne Sitz hat Eigenschaften geliefert")
	}
}

// **Eine Abzugskarte ohne Seiten hat keine Eigenschaften.**
//
// Sie entsteht bei einem **echten Druck**: dort gibt es keine eigene Karte in
// diesem Bestand, und die Seitenliste bleibt leer.
func TestEineAbzugskarteOhneSeitenHatKeineEigenschaften(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Sitze[0].Deck.Karten[0].Seiten = nil

	if _, da := partie.Eigenschaften(partie.Objekte[0]); da {
		t.Fatal("eine Karte ohne Seiten hat Eigenschaften geliefert")
	}
}

// **Ein Objekt ohne Eigenschaften hat auch keine Werte.**
func TestEinObjektOhneEigenschaftenHatKeineWerte(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte[0].Karte = ""

	staerke, widerstand := partie.Werte(partie.Objekte[0])

	if staerke.Steht || widerstand.Steht {
		t.Fatalf("Werte = %+v / %+v", staerke, widerstand)
	}
}

// **Die Sicht trägt die Eigenschaften — und ein geheimes Objekt nicht.**
//
// Dieselbe Regel wie für Name, Karte und Kartenkennung (G1/G2): wer die
// Stärke eines Objekts auf der gegnerischen Hand läse, wüsste, welche Karte
// es ist, und die Projektion daneben sähe weiter richtig aus.
func TestDieSichtTraegtDieEigenschaften(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte[0].Zaehler = map[string]int{domain.MarkePlusEins: 2}

	sicht := partie.Sicht("sitz-1")

	eigen := sicht.Objekte[0].Eigenschaften
	if eigen == nil {
		t.Fatal("die Eigenschaften fehlen")
	}

	switch {
	case len(eigen.Farben) != 1 || eigen.Farben[0] != domain.FarbeGruen:
		t.Errorf("Farben = %v", eigen.Farben)
	case eigen.Manawert != 2:
		t.Errorf("Manawert = %d", eigen.Manawert)
	case eigen.Staerke.Zahl != 4 || eigen.Widerstandskraft.Zahl != 5:
		t.Errorf("%+v/%+v — die Marken fehlen",
			eigen.Staerke, eigen.Widerstandskraft)
	}
}

func TestEinGeheimesObjektTraegtKeineEigenschaften(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte[0].Zone = domain.ZoneHand

	sicht := partie.Sicht("sitz-2")

	if sicht.Objekte[0].Eigenschaften != nil {
		t.Fatalf("ein geheimes Objekt trägt %+v",
			sicht.Objekte[0].Eigenschaften)
	}
}

// **Ein unbekannter Kartentyp erreicht die Sicht** — §8, und die dritte
// Abnahme von M2: kein stiller Ausfall.
//
// **Über die Sicht und nicht über [domain.TypzeileAus]**: dass die Funktion
// einen Befund **gibt**, prüft `kartenmerkmale_test.go`. Hier wird geprüft,
// dass er **ankommt** — denn ein Befund, den niemand liest, ist keiner.
func TestEinUnbekannterTypErreichtDieSicht(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Sitze[0].Deck.Karten[0].Seiten[0].Felder = map[string]string{
		"name": "Zauberding", "super type": "Zauberding Creature",
	}

	sicht := partie.Sicht("sitz-1")

	eigen := sicht.Objekte[0].Eigenschaften
	if eigen == nil {
		t.Fatal("die Eigenschaften fehlen")
	}

	if len(eigen.UnbekannteTypen) != 1 ||
		eigen.UnbekannteTypen[0] != "Zauberding" {
		t.Fatalf("UnbekannteTypen = %v", eigen.UnbekannteTypen)
	}
}

// **Der Name der Karte kommt durch die Sicht** — sonst steht am Tisch die
// Objektkennung.
//
// # Der Befund, für den dieser Lauf steht
//
// **Gesehen am 2026-09-17** bei der ersten Sichtprüfung mit einem echten
// Deck: jede offene Karte trug ihre Objektkennung als Beschriftung
// (`sitz-1/p1-12/0`) statt ihres Namens. [Spielobjekt.Name] ist bei einer
// Karte **immer leer** — er steht dort, wo *keine* Karte dahintersteht —,
// und der Name der Karte lag bis dahin nur im Abzug.
//
// **Die Testvorrichtung des Fensters setzte ihn**, der Betrieb nicht. Das ist
// §3 in seiner teuersten Gestalt: eine Vorrichtung, die großzügiger ist als
// der Betrieb.
func TestDerKartennameKommtDurchDieSicht(t *testing.T) {
	t.Parallel()

	sicht := eineKartenpartie().Sicht("sitz-1")

	eigen := sicht.Objekte[0].Eigenschaften
	if eigen == nil {
		t.Fatal("die Eigenschaften fehlen")
	}

	if eigen.Name != "Grizzlybär" {
		t.Fatalf("Name = %q", eigen.Name)
	}
}

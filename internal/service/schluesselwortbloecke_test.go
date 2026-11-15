package service

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// zerteiltZuSchluesselwoertern zerteilt einen Setinhalt für diese Läufe.
func zerteiltZuSchluesselwoertern(t *testing.T, inhalt string) *msepaket.Baum {
	t.Helper()

	baum, err := msepaket.Zerteile("set", []byte(inhalt))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	return baum
}

// **Die Schlüsselwörter gehen zurück in die Datei** —
// #1062.
//
// Derselbe Weg wie bei den Karten (`kartenbloecke.go`): der Baum kommt aus der
// **Quelldatei**, und was der Bestand führt, wird hineingelegt. So bleibt alles
// stehen, was niemand angefasst hat — Kommentare, Abstände, Zeilenenden.
func TestSchluesselwoerterGehenZurueckInDieDatei(t *testing.T) {
	faelle := []struct {
		name     string
		quelle   string
		woerter  []*domain.Keyword
		erwartet []string
		fehlt    []string
	}{
		{
			name: "ein geänderter Erinnerungstext",
			quelle: "game: probe\nkeyword:\n\tkeyword: Flug\n\tmatch: Flug\n" +
				"\treminder: alt\n",
			woerter:  []*domain.Keyword{{Word: "Flug", Match: "Flug", Reminder: "neu"}},
			erwartet: []string{"keyword: Flug", "reminder: neu"},
			fehlt:    []string{"reminder: alt"},
		},
		{
			// **Ein Block ohne Schlüsselwort ist ein gelöschtes** — und er geht
			// mit. Ihn stehen zu lassen schriebe eines zurück, das jemand
			// entfernt hat.
			name:     "ein gelöschtes verschwindet",
			quelle:   "game: probe\nkeyword:\n\tkeyword: Flug\nkeyword:\n\tkeyword: Beben\n",
			woerter:  []*domain.Keyword{{Word: "Beben"}},
			erwartet: []string{"keyword: Beben"},
			fehlt:    []string{"keyword: Flug"},
		},
		{
			// **Ein neues bekommt einen Block.** Er steht hinter dem letzten
			// vorhandenen — und wo es keinen gibt, am Ende: so schreibt MSE
			// selbst (`IMPLEMENT_REFLECTION(Set)`, `src/data/set.cpp`, Zeilen
			// 184–204, Fassung f9d9356 vom 2021-07-06 — `keywords` steht hinter
			// den Karten).
			name:     "ein neues kommt hinzu",
			quelle:   "game: probe\ncard:\n\tname: Probe\n",
			woerter:  []*domain.Keyword{{Word: "Beben", Match: "Beben", Mode: "custom"}},
			erwartet: []string{"keyword:", "\tkeyword: Beben", "\tmatch: Beben", "\tmode: custom"},
		},
		{
			name:     "ohne Schlüsselwörter bleibt die Datei, wie sie war",
			quelle:   "game: probe\ncard:\n\tname: Probe\n",
			woerter:  nil,
			erwartet: []string{"game: probe", "name: Probe"},
			fehlt:    []string{"keyword"},
		},
		{
			// **Ein leeres Feld bekommt keine Zeile.** MSE legt ein neues
			// Schlüsselwort in jedem Feld leer an; fünf leere Zeilen wären
			// fünf Zeilen, die nichts sagen.
			name:     "leere Felder bleiben aus",
			quelle:   "game: probe\n",
			woerter:  []*domain.Keyword{{Word: "Nur ein Name"}},
			erwartet: []string{"\tkeyword: Nur ein Name"},
			fehlt:    []string{"match:", "reminder:", "rules:", "mode:"},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			baum := zerteiltZuSchluesselwoertern(t, fall.quelle)

			schluesselwoerterAngleichen(baum, fall.woerter)

			heraus := string(msepaket.Schreibe(baum))

			for _, erwartet := range fall.erwartet {
				if !strings.Contains(heraus, erwartet) {
					t.Errorf("%q fehlt in\n%s", erwartet, heraus)
				}
			}

			for _, fehlt := range fall.fehlt {
				if strings.Contains(heraus, fehlt) {
					t.Errorf("%q steht noch in\n%s", fehlt, heraus)
				}
			}
		})
	}
}

// **Ein unangetastetes Set kommt Byte für Byte heraus** (0002).
//
// Das ist die Zusage, an der die ganze Schreibseite hängt: wer nichts ändert,
// bekommt dieselbe Datei zurück.
func TestEinUnangetastetesSetBleibtByteGleich(t *testing.T) {
	quelle := "\ufeffmse version: 2.0.0\r\ngame: probe\r\n\r\n# ein Kommentar\r\n" +
		"keyword:\r\n\tkeyword:  Flug\r\n\tmatch: Flug\r\ncard:\r\n\tname: Probe"

	baum := zerteiltZuSchluesselwoertern(t, quelle)

	schluesselwoerterAngleichen(baum, []*domain.Keyword{{Word: "Flug", Match: "Flug"}})

	if heraus := string(msepaket.Schreibe(baum)); heraus != quelle {
		t.Errorf("heraus kam\n%q\nerwartet\n%q", heraus, quelle)
	}
}

// **Was nicht zum Schlüsselwort gehört, bleibt stehen.**
//
// Ein Kommentar unter einem Block, eine Zeile aus einer neueren MSE-Fassung:
// beides gehört nicht uns, und beides kommt wieder heraus (0002).
func TestFremdeZeilenImSchluesselwortblockBleiben(t *testing.T) {
	baum := zerteiltZuSchluesselwoertern(t,
		"keyword:\n\tkeyword: Flug\n\t# ein Kommentar\n\tetwas neues: ja\n")

	schluesselwoerterAngleichen(baum, []*domain.Keyword{{Word: "Flug", Match: "Flug"}})

	heraus := string(msepaket.Schreibe(baum))

	for _, erwartet := range []string{"# ein Kommentar", "etwas neues: ja", "match: Flug"} {
		if !strings.Contains(heraus, erwartet) {
			t.Errorf("%q fehlt in\n%s", erwartet, heraus)
		}
	}
}

// **Ein geleertes Feld verliert seine Zeile.**
//
// Sie stünde sonst als Rest von etwas da, das jemand gelöscht hat — dieselbe
// Regel wie bei einem Kartenfeld (#191).
func TestEinGeleertesSchluesselwortfeldVerliertSeineZeile(t *testing.T) {
	baum := zerteiltZuSchluesselwoertern(t,
		"keyword:\n\tkeyword: Flug\n\treminder: weg damit\n")

	schluesselwoerterAngleichen(baum, []*domain.Keyword{{Word: "Flug"}})

	if heraus := string(msepaket.Schreibe(baum)); strings.Contains(heraus, "reminder") {
		t.Errorf("die Zeile steht noch in\n%s", heraus)
	}
}

// **Ein mehrzeiliger Erinnerungstext wird ersetzt und nicht verdoppelt.**
//
// Bei einem mehrzeiligen Wert steht hinter dem Doppelpunkt nichts; der Wert
// liegt in den eingerückten Kindern. Wer sie stehen ließe, hätte beide — das
// war #704 an
// der Karte, und hier wäre es dasselbe.
func TestEinMehrzeiligerErinnerungstextWirdErsetzt(t *testing.T) {
	baum := zerteiltZuSchluesselwoertern(t,
		"keyword:\n\tkeyword: Flug\n\treminder:\n\t\tErste Zeile\n\t\tZweite Zeile\n")

	schluesselwoerterAngleichen(baum,
		[]*domain.Keyword{{Word: "Flug", Reminder: "Nur noch eine."}})

	heraus := string(msepaket.Schreibe(baum))

	if !strings.Contains(heraus, "reminder: Nur noch eine.") {
		t.Errorf("der neue Text fehlt in\n%s", heraus)
	}

	if strings.Contains(heraus, "Erste Zeile") {
		t.Errorf("der alte Text steht noch in\n%s", heraus)
	}
}

// TestNachEinerLoeschungBleibtJedeFremdeZeileAnIhremWort ist Befund 2 des
// Zweitblicks über Phase W, und er ist der schwerere von beiden.
//
// **Das Abzählen schrieb in den falschen Block.** Blieben zwei Blöcke A und B
// stehen und war A gelöscht, bekam **As Block** die Felder von B — samt aller
// Zeilen, die A trug und die dieses Programm nicht kennt. Bs Block fiel weg,
// und mit ihm Bs eigene fremde Zeilen.
//
// Der Kartenweg hat denselben Fehler schon einmal gehabt und trägt seither
// eine Stellenzuordnung (`kartenbloecke.go`, Zeilen 33–40); die
// Schlüsselwörter zählten weiter mit.
func TestNachEinerLoeschungBleibtJedeFremdeZeileAnIhremWort(t *testing.T) {
	t.Parallel()

	baum := zerteiltZuSchluesselwoertern(t, ""+
		"keyword:\n\tkeyword: Alpha\n\tmatch: Alpha\n\tmode: core\n\tzukuenftiges feld: A\n"+
		"keyword:\n\tkeyword: Beta\n\tmatch: Beta\n\tmode: core\n\tzukuenftiges feld: B\n")

	// Alpha stand an Stelle 0 und ist gelöscht; Beta stand an Stelle 1.
	schluesselwoerterAngleichen(baum, []*domain.Keyword{
		{Word: "Beta", Match: "Beta", Mode: "core", Reihenfolge: 1},
	})

	hat := string(msepaket.Schreibe(baum))

	if strings.Contains(hat, "zukuenftiges feld: A") {
		t.Errorf("Alphas fremde Zeile hängt an Beta:\n%s", hat)
	}

	if !strings.Contains(hat, "zukuenftiges feld: B") {
		t.Errorf("Betas eigene fremde Zeile ist weg:\n%s", hat)
	}

	if strings.Count(hat, "keyword:\n") != 1 {
		t.Errorf("nicht genau ein Block:\n%s", hat)
	}
}

// TestEinNeuesWortNachEinerLoeschungKommtAn hält den Fehler fest, den die
// Stellenzuordnung selbst hätte einführen können.
//
// **Der Nachtrag zählte die Liste ab.** Blieb von zwei Blöcken einer übrig und
// kam ein drittes Wort hinzu, standen zwei Wörter in der Liste und zwei Blöcke
// in der Datei — der Nachtrag fand nichts mehr und schrieb das neue Wort nie.
func TestEinNeuesWortNachEinerLoeschungKommtAn(t *testing.T) {
	t.Parallel()

	baum := zerteiltZuSchluesselwoertern(t, ""+
		"keyword:\n\tkeyword: Alpha\n\tmatch: Alpha\n\tmode: core\n"+
		"keyword:\n\tkeyword: Beta\n\tmatch: Beta\n\tmode: core\n")

	schluesselwoerterAngleichen(baum, []*domain.Keyword{
		{Word: "Beta", Match: "Beta", Mode: "core", Reihenfolge: 1},
		{Word: "Gamma", Match: "Gamma", Mode: "core", Reihenfolge: 2},
	})

	hat := string(msepaket.Schreibe(baum))

	for _, will := range []string{"keyword: Beta", "keyword: Gamma"} {
		if !strings.Contains(hat, will) {
			t.Errorf("%q fehlt in:\n%s", will, hat)
		}
	}

	if strings.Contains(hat, "keyword: Alpha") {
		t.Errorf("das gelöschte Alpha steht noch da:\n%s", hat)
	}
}

// **Und der zweite Speichervorgang schreibt dasselbe wie der erste.**
//
// Der Nachlauf des Zweitblicks über Phase W hat es angestrichen: die
// Stellenzuordnung bleibt nur gültig, solange die **Quelle** dieselbe bleibt.
// Sie tut es — `Setschreiber` liest `set_quellen.inhalt` und schreibt es nie
// zurück —, aber geprüft war es nicht.
//
// **Der Lauf fährt es zweimal über denselben Baum-Ausgangsstand**, so wie zwei
// Speichervorgänge es täten: derselbe rohe Inhalt, dieselbe Wortliste.
func TestZweiSpeichervorgaengeSchreibenDasselbe(t *testing.T) {
	t.Parallel()

	const inhalt = "keyword:\n\tkeyword: Alpha\n\tmatch: Alpha\n\tmode: core\n\tfremd: A\n" +
		"keyword:\n\tkeyword: Beta\n\tmatch: Beta\n\tmode: core\n\tfremd: B\n"

	// Alpha (Stelle 0) ist gelöscht, Beta bleibt, Gamma kommt hinzu.
	woerter := func() []*domain.Keyword {
		return []*domain.Keyword{
			{Word: "Beta", Match: "Beta", Mode: "core", Reihenfolge: 1},
			{Word: "Gamma", Match: "Gamma", Mode: "core", Reihenfolge: 2},
		}
	}

	erster := zerteiltZuSchluesselwoertern(t, inhalt)
	schluesselwoerterAngleichen(erster, woerter())

	zweiter := zerteiltZuSchluesselwoertern(t, inhalt)
	schluesselwoerterAngleichen(zweiter, woerter())

	if hat, will := string(msepaket.Schreibe(zweiter)), string(msepaket.Schreibe(erster)); hat != will {
		t.Errorf("der zweite Vorgang schreibt anderes:\n%s\nder erste:\n%s", hat, will)
	}

	// **Und Betas fremde Zeile hängt weiter an Beta** — die Zusage, um die es
	// überhaupt geht.
	geschrieben := string(msepaket.Schreibe(erster))
	if !strings.Contains(geschrieben, "fremd: B") || strings.Contains(geschrieben, "fremd: A") {
		t.Errorf("die fremden Zeilen stimmen nicht:\n%s", geschrieben)
	}
}

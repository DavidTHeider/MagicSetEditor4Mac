//go:build fremdpakete

// Die Zerlegung des Abstands nach Ebenenart, über den ganzen Kanonlauf
// (#521).
//
// # Warum sie hier steht und nicht in `render`
//
// `render.ZerlegeAbstand` rechnet **eine** Karte; die Frage dieser Phase geht
// über 510. Das Zusammenzählen gehört dorthin, wo der Lauf steht — und der
// steht hinter dem Build-Tag `fremdpakete`, weil er die kanonischen Sets und
// MSE2s Ausgabe braucht.
package app

import (
	"image"
	"sort"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// messstand ist alles, was ein Lauf über die Karten mitzählt.
//
// **Ein Verbund und keine drei Parameter**: die Diagnose wächst — Zerlegung
// (#521), Versatzprobe und Vorlagen (#522) —, und jede neue Größe hätte sonst
// die Signatur von `vergleicheMitMSE2` und `messeSatzGegenMSE2` geändert.
type messstand struct {
	teile    *abstandsteile
	versatz  *versatzstand
	vorlagen vorlagenstand

	// textversatz ist die Versatzprobe **allein über die Textpunkte**
	// (#1029).
	//
	// **Warum getrennt von `versatz`**: der über alles findet nichts, und die
	// Bilder zeigen den Regeltext trotzdem doppelt. Ein Versatz **einer** Ebene
	// verschwindet in einer Probe über alle — der Rahmen stellt 93 % der Fläche
	// und zieht jeden Kandidaten außer (0,0) ins Minus.
	textversatz *versatzstand

	// setstil ist der Stil des Sets — der Rückfall für Karten ohne eigenen.
	setstil string

	// karten ist der Abstand jeder gezogenen Karte, für den Vergleich mit
	// einem früheren Lauf (F02, `kanonjekarte_intern_test.go`).
	karten []kartenwert
}

func neuerMessstand() *messstand {
	return &messstand{
		teile:       neueAbstandsteile(),
		versatz:     neuerVersatzstand(),
		vorlagen:    vorlagenstand{},
		textversatz: neuerVersatzstand(),
	}
}

func (s *messstand) uebernimm(anderer *messstand) {
	s.teile.uebernimm(anderer.teile)
	s.versatz.uebernimm(anderer.versatz)
	s.textversatz.uebernimm(anderer.textversatz)
	s.karten = append(s.karten, anderer.karten...)

	for name, abstaende := range anderer.vorlagen {
		s.vorlagen[name] = append(s.vorlagen[name], abstaende...)
	}
}

func (s *messstand) melde(t *testing.T, was string) {
	t.Helper()

	s.teile.melde(t, was)
	s.versatz.melde(t, was)
	s.textversatz.melde(t, was+" — nur Text")
}

// abstandsteile zählt Zerlegungen über viele Karten zusammen.
//
// **Quadratsummen und keine Abstände.** Ein Abstand ist eine Wurzel und
// addiert sich nicht; wer Mittelwerte mittelt, bekommt bei ungleich großen
// Flächen eine Zahl, die zu nichts gehört (§10).
type abstandsteile struct {
	teile  map[render.Ebenenart]render.Teilabstand
	gesamt render.Teilabstand

	// karten ist die Zahl der Karten, über die zusammengezählt wurde — die
	// Reichweite der Aussage.
	karten int

	// ohneZerlegung zählt die Karten, für die kein Plan zu bekommen war. Ohne
	// sie sähe eine Tabelle über 400 Karten aus wie eine über 510.
	ohneZerlegung int
}

func neueAbstandsteile() *abstandsteile {
	return &abstandsteile{teile: map[render.Ebenenart]render.Teilabstand{}}
}

// nimm zählt eine Zerlegung hinzu.
func (s *abstandsteile) nimm(zerlegung render.Zerlegung) {
	s.karten++

	for art, teil := range zerlegung.Teile {
		zusammen := s.teile[art]
		zusammen.Punkte += teil.Punkte
		zusammen.Quadratsumme += teil.Quadratsumme
		s.teile[art] = zusammen
	}

	s.gesamt.Punkte += zerlegung.Gesamt.Punkte
	s.gesamt.Quadratsumme += zerlegung.Gesamt.Quadratsumme
}

// melde schreibt die Tabelle in den Testlauf.
//
// **Nach Anteil geordnet und nicht nach Namen**: die Frage lautet „wo sitzt
// der Abstand", und die Antwort steht dann oben.
func (s *abstandsteile) melde(t *testing.T, was string) {
	t.Helper()

	if s.karten == 0 {
		return
	}

	arten := make([]render.Ebenenart, 0, len(s.teile))
	for art := range s.teile {
		arten = append(arten, art)
	}

	sort.Slice(arten, func(i, j int) bool {
		return s.teile[arten[i]].Quadratsumme > s.teile[arten[j]].Quadratsumme
	})

	t.Logf("  %s — Zerlegung über %d Karten, %d Bildpunkte:",
		was, s.karten, s.gesamt.Punkte)

	for _, art := range arten {
		teil := s.teile[art]
		t.Logf("      %-12s Anteil %6.2f %%   Abstand %7.3f   %6.2f %% der Fläche",
			art, 100*s.anteil(art), teil.Abstand(),
			100*float64(teil.Punkte)/float64(s.gesamt.Punkte))
	}

	if s.ohneZerlegung > 0 {
		t.Logf("      %d Karten ohne Plan — nicht zerlegt", s.ohneZerlegung)
	}
}

// anteil ist der Bruchteil des Gesamtquadrats, den eine Art trägt.
func (s *abstandsteile) anteil(art render.Ebenenart) float64 {
	if s.gesamt.Quadratsumme == 0 {
		return 0
	}

	return s.teile[art].Quadratsumme / s.gesamt.Quadratsumme
}

// herkunftZurKarte sagt, welche Ebene welchen Bildpunkt zuletzt gemalt hat.
//
// # Warum ein zweiter Zeichenlauf
//
// Gemessen wird das Bild aus [App.Vorschau] — **die Naht, um die es geht**
// (Befund 1 des Zweitblicks über Phase A). Die Herkunft kommt daneben aus
// demselben Plan und demselben Rasterer; dass beide Wege dasselbe Bild geben,
// hält `render.TestZeichneMitHerkunftGibtDasselbeBild` fest.
//
// Der zweite Lauf kostet — er ist der Preis dafür, dass die gemessene Karte
// die der Anwendung bleibt und nicht eine aus einer eigenen Kette.
func herkunftZurKarte(
	t *testing.T, anwendung *App, karteID string,
) (*render.Herkunft, bool) {
	plan := planZurKarte(t, anwendung, karteID)

	// **Der Rasterer der Anwendung**, nicht ein neuer: ein eigener hätte weder
	// die Bildquelle der Pakete noch die Schriftsuche, und die Herkunft träfe
	// auf ein Bild, das so nie entsteht.
	rasterer, ok := anwendung.rasterer.(*render.Rasterer)
	if !ok {
		return nil, false
	}

	_, herkunft, _, err := rasterer.ZeichneMitHerkunft(probekontext(t), plan, 1)
	if err != nil {
		return nil, false
	}

	return herkunft, true
}

// uebernimm zählt einen zweiten Sammler hinzu — je Satz einer, alle in einen.
func (s *abstandsteile) uebernimm(anderer *abstandsteile) {
	s.karten += anderer.karten
	s.ohneZerlegung += anderer.ohneZerlegung

	for art, teil := range anderer.teile {
		zusammen := s.teile[art]
		zusammen.Punkte += teil.Punkte
		zusammen.Quadratsumme += teil.Quadratsumme
		s.teile[art] = zusammen
	}

	s.gesamt.Punkte += anderer.gesamt.Punkte
	s.gesamt.Quadratsumme += anderer.gesamt.Quadratsumme
}

// zerlegeUndMisst gibt den Abstand einer Karte und zählt ihre Zerlegung mit.
//
// **Eine Messung und nicht zwei.** Der Abstand kommt aus derselben Rechnung
// wie die Zerlegung — sonst stünden im Lauf zwei Zahlen, die dasselbe heißen
// und es nicht sind.
//
// **Ohne Plan wird trotzdem gemessen.** Die Zerlegung fällt dann aus und wird
// gezählt; die Karte aus dem Abstand zu nehmen, machte die Zahl unvergleichbar
// mit der von gestern (§10).
func zerlegeUndMisst(
	t *testing.T, anwendung *App, karte KarteDTO, name string,
	unsere, referenz image.Image, stand *messstand,
) (float64, bool) {
	t.Helper()

	abstand, weiter := gemessen(t, anwendung, karte, name, unsere, referenz, stand)
	if !weiter {
		return 0, false
	}

	vorlage := vorlageDerKarte(karte, stand.setstil)
	stand.vorlagen.nimm(vorlage, abstand)

	// **Der Name trägt seine Vorlage** (#1029). „48 Karten bei (0,−1)" sagt
	// nicht, ob es eine Klasse ist; „48 Karten, alle über `m15-altered`" sagt
	// es. Die Vorlage ist die Einheit, in der sich Satzfehler zeigen — das
	// hat die Aufschlüsselung nach Vorlage schon gezeigt.
	benannt := name + " [" + vorlage + "]"

	probe, err := render.ProbeVersatz(unsere, referenz, Versatzreichweite)
	if err != nil {
		t.Errorf("%s: Versatzprobe: %v", name, err)

		return abstand, true
	}

	stand.versatz.nimm(probe, benannt)

	// **Und dieselbe Frage allein über die Textpunkte** (#1029, Ω.2). Ohne
	// Herkunft geht sie nicht — dann bleibt es bei der Probe über alles.
	if herkunft, ok := herkunftZurKarte(t, anwendung, karte.ID); ok {
		textprobe, fehler := render.ProbeVersatzJeEbene(unsere, referenz,
			herkunft, render.ArtText, Versatzreichweite)
		if fehler != nil {
			t.Errorf("%s: Textversatzprobe: %v", name, fehler)

			return abstand, true
		}

		// **Karten ohne Textpunkte zählen nicht mit.** Eine Karte, die keinen
		// Text malt, hätte sonst eine Stimme für „kein Versatz" — und die
		// Mehrheit sagte dann etwas über das Schweigen.
		if textprobe.Punkte > 0 {
			stand.textversatz.nimm(textprobe, benannt)
		}
	}

	return abstand, true
}

// gemessen gibt den Abstand einer Karte und zählt ihre Zerlegung mit.
func gemessen(
	t *testing.T, anwendung *App, karte KarteDTO, name string,
	unsere, referenz image.Image, stand *messstand,
) (float64, bool) {
	t.Helper()

	teile := stand.teile

	herkunft, ok := herkunftZurKarte(t, anwendung, karte.ID)
	if !ok {
		teile.ohneZerlegung++

		ganz, err := render.Ganzabstand(unsere, referenz)
		if err != nil {
			t.Errorf("%s: %v", name, err)

			return 0, false
		}

		return ganz.Abstand(), true
	}

	zerlegung, err := render.ZerlegeAbstand(herkunft, unsere, referenz)
	if err != nil {
		t.Errorf("%s: %v", name, err)

		return 0, false
	}

	teile.nimm(zerlegung)

	return zerlegung.Gesamt.Abstand(), true
}

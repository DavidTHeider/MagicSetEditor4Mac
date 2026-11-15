//go:build fremdpakete

// **Was der Fehler der Referenz am Abstand kostet**
// (#458).
//
// # Der Fehler
//
// MSE2 schneidet das Innere eines Setsymbols über `dc.Blit(…, wxAND_INVERT)`
// aus. Unter **wxOSX-Cocoa 3.3.3 gibt dieser Blit `false` zurück und tut
// nichts** — gemessen. Damit bekommt das ganze Symbol die **Randfarbe** der
// Fassung statt der Füllfarbe: über die 3.761 Karten von `White` **100 %
// Randfarbe, 0 % Füllfarbe**, und `common` erscheint 1.526-mal reinweiß,
// obwohl seine Füllfarbe schwarz ist.
//
// **Das ist kein Abstand zur Parität.** Wer `symbolfilter.go` daran anpasste,
// portierte eine Einschränkung von wxOSX.
//
// # Was dieser Lauf beiträgt
//
// Bis hierher war das eine **Aussage ohne Preisschild**: „dort ist nichts zu
// holen" sagt nicht, wie viel „dort" ist. Dieser Lauf beziffert es — er
// rechnet den Abstand einmal ganz und einmal ohne das Feld `rarity` und
// meldet beide.
//
// # Gerechnet wird, wie der Maßstab rechnet
//
// **Je Karte eine Wurzel, dann das Mittel** — nicht die Wurzel über alle
// Bildpunkte zusammen. Die beiden sind verschieden: gepoolt ergibt derselbe
// Lauf **36,383**, gemittelt **20,692**, und wer sie nebeneinanderlegt, sieht
// einen Unterschied, den es nicht gibt (§10).
//
// # Die Zahl, und was sie nicht ist
//
// Am 2026-08-20 über 573 Karten aus vier Sätzen: das Feld `rarity` hat einen
// eigenen Abstand von **87,031** bei **0,452 %** der Fläche; ohne es fiele der
// Gesamtabstand von **20,692 auf 19,774**. Der Fehler der Referenz kostet also
// **0,918** — 4,4 % des Ganzen.
//
// **Das ist eine Obergrenze und keine Zuordnung.** Was von diesen 0,918 auf
// unser Konto geht, trennt dieser Lauf nicht: er nimmt das ganze Feld heraus,
// nicht nur den Anteil, den `wxAND_INVERT` verdirbt. Dass der Löwenanteil der
// Referenz gehört, steht in #458 an der Auszählung — 100 % Randfarbe, 0 %
// Füllfarbe —, nicht in dieser Zahl.
//
// # Was er nicht tut
//
// **Er ändert den Maßstab nicht.** Die Zahl in `docs/ROADMAP.md` bleibt die
// über die ganze Karte; hier steht daneben, welcher Teil davon der Referenz
// gehört und nicht uns.
package app

import (
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// symbolfeld ist das Feld, das die Seltenheit zeichnet.
const symbolfeld = "rarity"

// kanaele sind die Farbkanäle, über die [render.Teilabstand.Abstand] mittelt.
const kanaele = 3

// symbolstand sammelt die Abstände je Karte, ganz und ohne das Symbolfeld.
type symbolstand struct {
	ganz                   []float64
	ohne                   []float64
	uebrige                []float64
	feld                   []float64
	flaeche, gesamtflaeche int
}

// nimm trägt eine Karte ein.
//
// **`teil` darf leer sein** — nicht jede Vorlage führt ein Symbolfeld; dann
// sind `ganz` und `ohne` für diese Karte gleich.
func (s *symbolstand) nimm(voll, teil render.Teilabstand) {
	s.ganz = append(s.ganz, voll.Abstand())
	s.gesamtflaeche += voll.Punkte

	rest := voll.Quadratsumme - teil.Quadratsumme
	if rest < 0 {
		rest = 0
	}

	// **Durch `Punkte × Kanäle`**, wie [render.Teilabstand.Abstand] es tut —
	// sonst käme das Ergebnis um √3 zu groß heraus und stünde über dem
	// vollen Abstand, was beim ersten Anlauf auch passierte.
	s.ohne = append(s.ohne, math.Sqrt(rest/float64(voll.Punkte*kanaele)))

	// **Die zweite Lesart daneben**: der Abstand der übrigen Fläche, also
	// ohne die Bildpunkte des Feldes im Nenner. Sie beantwortet „wie weit ist
	// der Rest der Karte", nicht „was kostet der Fehler am Gesamtmaß".
	if uebrig := voll.Punkte - teil.Punkte; uebrig > 0 {
		s.uebrige = append(s.uebrige, math.Sqrt(rest/float64(uebrig*kanaele)))
	}

	if teil.Punkte > 0 {
		s.feld = append(s.feld, teil.Abstand())
		s.flaeche += teil.Punkte
	}
}

// **Was der Fehler der Referenz am Setsymbol kostet.**
func TestWasDerReferenzfehlerAmSymbolKostet(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	anwendung := gestartetMit(t, func() string { return wurzel })

	saetze := saetzeMitReferenz(t)
	jeSatz := stichprobengroesse(t, len(saetze))

	stand := &symbolstand{}
	verluste := ausfaelle{}

	gemessen := 0

	for _, satz := range saetze {
		if _, err := os.Stat(mse2ausgabe(satz)); err != nil {
			// **Auch ein fehlender ganzer Satz wird gezählt.** Sonst spräche
			// die Zahl über weniger als die vier Sätze, die daneben stehen,
			// und niemand sähe es — angestrichen vom Nachlauf zum Zweitblick
			// über Phase K.
			verluste.nimm("keine MSE2-Ausgabe für den Satz " + satz)

			continue
		}

		gemessen += sammleSymbolanteil(t, anwendung, satz, jeSatz, stand, verluste)
	}

	if gemessen == 0 {
		t.Skip("keine MSE2-Ausgabe — scripts/mse2-referenz.sh")
	}

	ganz, ohne := mittel(stand.ganz), mittel(stand.ohne)

	t.Logf("%d Karten, Abstand im Mittel je Karte: ganz %.3f", gemessen, ganz)
	t.Logf("  Feld %q: eigener Abstand %.3f über %d Karten, %.3f %% der Fläche",
		symbolfeld, mittel(stand.feld), len(stand.feld),
		100*float64(stand.flaeche)/float64(stand.gesamtflaeche))
	// **„Ohne dieses Feld" heißt: als träfe es die Referenz genau** — die
	// Bildpunkte des Feldes zählen weiter zur Karte, sie tragen nur nichts
	// mehr bei. Der Nenner bleibt deshalb die **ganze** Karte.
	//
	// **Die andere Lesart wäre der Abstand der übrigen Fläche**, also mit
	// `voll.Punkte − teil.Punkte` im Nenner. Sie beantwortet eine andere
	// Frage; gefragt ist hier, was der Fehler der Referenz am **Gesamtmaß**
	// kostet. Angestrichen vom Zweitblick über Phase K, der die Zweideutigkeit
	// zu Recht benannt hat — beide stehen jetzt da.
	t.Logf("  ohne dieses Feld, ganze Karte im Nenner: %.3f — "+
		"der Fehler der Referenz kostet %.3f", ohne, ganz-ohne)
	t.Logf("  zum Vergleich, nur die übrige Fläche im Nenner: %.3f",
		mittel(stand.uebrige))

	verluste.melde(t, gemessen)
}

// sammleSymbolanteil misst die Stichprobe eines Satzes und gibt ihre Größe.
func sammleSymbolanteil(
	t *testing.T, anwendung *App, satz string, jeSatz int, stand *symbolstand,
	verluste ausfaelle,
) int {
	t.Helper()

	geoeffnet := anwendung.SetOeffnen(t.Context(),
		filepath.Join(kanonverzeichnis(), satz+".mse-set"))
	if geoeffnet.Fehler != nil {
		t.Errorf("%s öffnen: %+v", satz, geoeffnet.Fehler)

		return 0
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil {
		t.Errorf("%s: Karten: %+v", satz, karten.Fehler)

		return 0
	}

	gezaehlt := 0

	for _, karte := range verteilteStichprobe(karten.Karten, jeSatz) {
		name := kartenname(karte)
		if name == "" {
			verluste.nimm("die Karte hat keinen Namen")

			continue
		}

		referenz, err := ladeBild(filepath.Join(referenzort(t, satz), referenzdatei(name)))
		if err != nil {
			verluste.nimm("kein Referenzbild", name)

			continue
		}

		vorschau := anwendung.Vorschau(t.Context(), karte.ID, 1)
		if vorschau.Fehler != nil {
			verluste.nimm("die Vorschau scheiterte", name)

			continue
		}

		unser := bildAus(t, anwendung, vorschau.Vorschau.Bild)

		voll, err := render.Ganzabstand(unser, referenz)
		if err != nil {
			verluste.nimm("der Abstand ließ sich nicht rechnen", name)

			continue
		}

		var teil render.Teilabstand

		// **Ein Fehler hier fiele sonst als Null durch.** `teil` bliebe leer,
		// die Karte zählte als gemessen, und der Anteil, den dieser Lauf
		// beziffern soll, verschwände still — angestrichen vom Nachlauf zum
		// Zweitblick über Phase K.
		if rahmen, da := feldrahmen(vorschau.Vorschau.Flaechen, symbolfeld); da {
			gemessen, err := render.Ganzabstand(
				ausschnitt(t, unser, rahmen), ausschnitt(t, referenz, rahmen))
			if err != nil {
				verluste.nimm("der Abstand des Symbolfeldes ließ sich nicht rechnen", name)

				continue
			}

			teil = gemessen
		}

		gezaehlt++

		stand.nimm(voll, teil)
	}

	return gezaehlt
}

// feldrahmen sucht das Rechteck eines Feldes.
func feldrahmen(flaechen []FeldflaecheDTO, feld string) (image.Rectangle, bool) {
	for _, flaeche := range flaechen {
		if flaeche.Feld != feld {
			continue
		}

		return image.Rect(
			int(flaeche.Links), int(flaeche.Oben),
			int(flaeche.Links+flaeche.Breite), int(flaeche.Oben+flaeche.Hoehe)), true
	}

	return image.Rectangle{}, false
}

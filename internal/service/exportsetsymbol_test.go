package service_test

import (
	"image"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Das Setsymbol als Bild** —
// #660.
//
// Zwei der neunzehn Vorlagen brauchen es, und beide scheitern heute an
// `symbol_variation`: `magic-spoiler` schreibt je Seltenheit ein
// `set-symbol-<var>.png`, `magic-symbol-extractor` alle sieben Fassungen auf
// einmal.
//
//	write_image_file(file: "set-symbol.png", width: 200,
//	                 symbol_variation(symbol: set.symbol, variation: "rare"))
//
// **Zwei Dinge stecken darin**, und keines davon gab es hier:
//
//  1. `symbol_variation` steht nicht in der Umgebung der Ausfuhr — beim
//     Zeichnen wird es von `stilumgebung.go` gesetzt, und dort gibt es einen
//     Stil.
//  2. `write_image_file` nahm eine **Karte**. MSE nimmt beides:
//
//     ScriptObject<CardP>* card = dynamic_cast<ScriptObject<CardP>*>(input.get());
//     if (card) { … } else { image = input->toImage()->generateConform(options); }
//
//     `src/script/functions/export.cpp`, Zeilen 414–421, Fassung f9d9356 vom
//     2021-07-06.

// symbolgestaltprobe antwortet **nur auf einen auflösbaren Verweis**.
//
// **Sie tat es bis zum 2026-08-21 für jeden**, und das hat eine tote Naht
// verdeckt: `msepaket.Bildquelle.Symbolgestalt` konnte den Setpaketverweis
// nicht lesen, und die Zeichnung war im Betrieb **immer** leer. Angestrichen
// vom Zweitblick durch Codex.
//
// **Eine Vorrichtung, die großzügiger ist als der Betrieb, prüft sich selbst**
// (§3). Sie verlangt jetzt dieselbe Form wie die echte Quelle: einen
// Setpaketverweis oder einen Paketpfad mit führendem `/`.
type symbolgestaltprobe struct {
	gefragt    []string
	unlesbares []string
}

func (s *symbolgestaltprobe) Symbolgestalt(verweis string) msepaket.Symbolzeichnung {
	s.gefragt = append(s.gefragt, verweis)

	_, _, ist := msepaket.ZerlegeSetpaketverweis(verweis)
	if !ist && !strings.HasPrefix(verweis, "/") {
		s.unlesbares = append(s.unlesbares, verweis)

		return msepaket.Symbolzeichnung{}
	}

	return msepaket.Standardsymbol()
}

// baumrastererprobe hält fest, welcher Baum in welcher Größe verlangt wurde.
type baumrastererprobe struct {
	symbolrastererprobe

	baeume  []render.Bildbaum
	groesse [][2]int

	scheitert string
}

func (r *baumrastererprobe) BildAusBaum(
	baum render.Bildbaum, breite, hoehe int,
) (bild *image.NRGBA, befund string) {
	if r.scheitert != "" {
		return nil, r.scheitert
	}

	r.baeume = append(r.baeume, baum)
	r.groesse = append(r.groesse, [2]int{breite, hoehe})

	return image.NewNRGBA(image.Rect(0, 0, 7, 9)), ""
}

func mitSetsymbol(
	mitschrift *beigabenmitschrift, rasterer *baumrastererprobe, gestalt *symbolgestaltprobe,
) ausfuhrprobe {
	ausfuehrer := service.NeuerVorlagenausfuehrer()
	ausfuehrer.MitSymbolen(manaschrift(), rasterer)

	return ausfuhrprobe{
		ausfuehrer: ausfuehrer,
		mittel: service.Ausfuhrmittel{
			Beigaben:          mitschrift,
			Symbolzeichnungen: gestalt,
			Setpaket:          "/Orte/Probe.mse-set",
			Verzeichnis:       "ausfuhr-files",
		},
	}
}

// vorlageMitFassung gibt eine Vorlage samt den `variation:`-Blöcken des Spiels.
func vorlageMitFassung(skript string) msepaket.Exportvorlage {
	vorlage := vorlageMit(skript)
	vorlage.Symbolfassungen = []msepaket.Symbolfassung{
		{Name: "rare", Fuellart: "solid", Fuellfarbe: "rgb(200,160,0)", Randfarbe: "black"},
	}

	vorlage.Setfelder = []domain.Field{{Name: "symbol"}}

	return vorlage
}

func TestDasSetsymbolWirdGeschrieben(t *testing.T) {
	t.Parallel()

	mitschrift := &beigabenmitschrift{}
	rasterer := &baumrastererprobe{}
	gestalt := &symbolgestaltprobe{}

	ausgabe, _, err := mitSetsymbol(mitschrift, rasterer, gestalt).
		Fuehre(t.Context(),
			vorlageMitFassung(`write_image_file(file: "set-symbol.png", width: 200,`+
				` symbol_variation(symbol: set.symbol, variation: "rare"))`),
			eineKarte(), []domain.Value{{Field: "symbol", Text: "symbol.mse-symbol"}}, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	// **Zurück kommt der Dateiname** — die Vorlage hängt ihn in ein `<img>`.
	if ausgabe != "set-symbol.png" {
		t.Errorf("Ausgabe %q", ausgabe)
	}

	if _, da := mitschrift.gelegt["set-symbol.png"]; !da {
		t.Errorf("abgelegt wurde %v", mitschrift.gelegt)
	}

	// **Die Zeichnung wird im Setpaket gesucht**, nicht im Vorlagenpaket:
	// `set.symbol` nennt eine Datei des Sets.
	if len(gestalt.gefragt) != 1 ||
		!strings.Contains(gestalt.gefragt[0], "/Orte/Probe.mse-set") ||
		!strings.Contains(gestalt.gefragt[0], "symbol.mse-symbol") {
		t.Errorf("gefragt wurde %q", gestalt.gefragt)
	}

	// **Die Breite kommt aus dem Aufruf, die Höhe bleibt offen** — MSE
	// bestimmt sie dann aus dem Seitenverhältnis (`conform_image`).
	if len(rasterer.groesse) != 1 || rasterer.groesse[0] != [2]int{200, 0} {
		t.Errorf("verlangt wurde %v", rasterer.groesse)
	}

	if len(rasterer.baeume) != 1 || rasterer.baeume[0].Knotenart != "symbol_variation" {
		t.Errorf("gerastert wurde %+v", rasterer.baeume)
	}

	// **Und der Verweis war auflösbar.** Die Attrappe hält fest, was sie nicht
	// lesen könnte; ohne diese Zusage verdeckte sie die Naht ein zweites Mal.
	if len(gestalt.unlesbares) > 0 {
		t.Errorf("unlesbare Verweise: %q", gestalt.unlesbares)
	}
}

// **Ein zweiter Aufruf mit demselben Dateinamen schreibt nicht noch einmal.**
//
//	if (ei.exported_images.find(file) != ei.exported_images.end()) {
//	  SCRIPT_RETURN(file); // already written an image with this name
//	}
//
// `export.cpp`, Zeilen 406–409, Fassung f9d9356 vom 2021-07-06 — und es ist
// **dieselbe** Merkliste wie die von `symbols_to_html`.
//
// **Daran hängt mehr als ein gesparter Schreibvorgang.** `magic-spoiler` ruft
// `write_image_file(file: "set-symbol-{var}.png", …)` **je Karte**; über ein Set
// mit 300 Karten wäre dasselbe Symbol sonst dreihundertmal gerastert und
// geschrieben.
func TestEinZweiterAufrufMitDemselbenNamenSchreibtNichtNochmal(t *testing.T) {
	t.Parallel()

	mitschrift := &beigabenmitschrift{}
	rasterer := &baumrastererprobe{}

	_, _, err := mitSetsymbol(mitschrift, rasterer, &symbolgestaltprobe{}).
		Fuehre(t.Context(),
			vorlageMitFassung(
				`x := write_image_file(file: "s.png", width: 20,`+
					` symbol_variation(symbol: set.symbol, variation: "rare"))`+
					"\n"+
					`write_image_file(file: "s.png", width: 20,`+
					` symbol_variation(symbol: set.symbol, variation: "rare"))`),
			eineKarte(), []domain.Value{{Field: "symbol", Text: "s.mse-symbol"}}, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if len(rasterer.baeume) != 1 {
		t.Errorf("gerastert wurde %d-mal, erwartet einmal", len(rasterer.baeume))
	}
}

// **Und die Merkliste ist mit `symbols_to_html` geteilt.**
func TestDieMerklisteGiltFuerBeideFunktionen(t *testing.T) {
	t.Parallel()

	mitschrift := &beigabenmitschrift{}
	rasterer := &baumrastererprobe{}

	_, _, err := mitSetsymbol(mitschrift, rasterer, &symbolgestaltprobe{}).
		Fuehre(t.Context(),
			vorlageMitFassung(
				`x := write_image_file(file: "mana-W.png", width: 20,`+
					` symbol_variation(symbol: set.symbol, variation: "rare"))`+
					"\n"+
					`symbols_to_html("W", symbol_font: "mana")`),
			eineKarte(), []domain.Value{{Field: "symbol", Text: "s.mse-symbol"}}, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	// `symbols_to_html` hätte `mana-W.png` geschrieben — der Name steht schon
	// in der Liste, also rastert es nicht.
	if len(rasterer.gerastert) != 0 {
		t.Errorf("das Symbol wurde gerastert: %v", rasterer.gerastert)
	}
}

// **Eine Fassung, die das Spiel nicht erklärt, ist ein Fehler.**
//
//	throw ScriptError(_("Variation of symbol not found"));
//
// `src/script/functions/image.cpp`, Zeile 169. Eine erfundene Fassung zeichnete
// das Symbol in falschen Farben, und niemand sähe warum (§8).
func TestEineUnbekannteFassungIstEinFehler(t *testing.T) {
	t.Parallel()

	_, _, err := mitSetsymbol(&beigabenmitschrift{}, &baumrastererprobe{}, &symbolgestaltprobe{}).
		Fuehre(t.Context(),
			vorlageMitFassung(`symbol_variation(symbol: set.symbol, variation: "gibtsnicht")`),
			eineKarte(), []domain.Value{{Field: "symbol", Text: "s.mse-symbol"}}, nil)
	if err == nil || !strings.Contains(err.Error(), "gibtsnicht") {
		t.Fatalf("Fehler = %v", err)
	}
}

// **Ohne die Fähigkeit gibt es den Namen nicht** — dieselbe Regel wie überall
// sonst in der Ausfuhr (#216, #657, #658).
func TestOhneSetsymbolGibtEsDenNamenNicht(t *testing.T) {
	t.Parallel()

	_, _, err := probelauf(service.Ausfuhrmittel{
		Beigaben: &beigabenmitschrift{}, Verzeichnis: "ausfuhr-files",
	}).
		Fuehre(t.Context(),
			vorlageMitFassung(`symbol_variation(symbol: set.symbol, variation: "rare")`),
			eineKarte(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "nicht bekannt") {
		t.Fatalf("Fehler = %v", err)
	}
}

// **Ein Bild, das sich nicht rastern lässt, wird benannt.**
//
// Dieselbe Regel wie beim Symbolbild (#658) und aus demselben Grund: die Datei
// entstünde leer, die Vorlage verwiese darauf, und im Browser stünde ein leerer
// Rahmen. Eine halbe Ausfuhr ist eine, die jemand für vollständig hält (§8).
func TestEinUnrasterbaresSetsymbolWirdBenannt(t *testing.T) {
	t.Parallel()

	mitschrift := &beigabenmitschrift{}
	rasterer := &baumrastererprobe{scheitert: "die Zeichnung liegt nicht vor"}

	_, _, err := mitSetsymbol(mitschrift, rasterer, &symbolgestaltprobe{}).
		Fuehre(t.Context(),
			vorlageMitFassung(`write_image_file(file: "s.png", width: 20,`+
				` symbol_variation(symbol: set.symbol, variation: "rare"))`),
			eineKarte(), []domain.Value{{Field: "symbol", Text: "s.mse-symbol"}}, nil)
	if err == nil || !strings.Contains(err.Error(), "liegt nicht vor") {
		t.Fatalf("Fehler = %v", err)
	}

	if len(mitschrift.gelegt) != 0 {
		t.Errorf("trotz Fehler wurde abgelegt: %v", mitschrift.gelegt)
	}
}

// **Und ein Rasterer, der gar keine Bäume kann, sagt das.**
//
// Der Ausführer bekommt einen [service.Symbolrasterer]; ob der auch Bäume
// ausführt, entscheidet sein Typ — dieselbe Bauform wie bei der Symbolquelle
// des Planbauers. Wer es nicht kann, bekommt eine Meldung und keinen Absturz.
func TestEinRastererOhneBaeumeSagtEs(t *testing.T) {
	t.Parallel()

	ausfuehrer := service.NeuerVorlagenausfuehrer()
	ausfuehrer.MitSymbolen(manaschrift(), &symbolrastererprobe{})

	_, _, err := ausfuhrprobe{ausfuehrer: ausfuehrer, mittel: service.Ausfuhrmittel{
		Beigaben:          &beigabenmitschrift{},
		Symbolzeichnungen: &symbolgestaltprobe{},
		Setpaket:          "/Orte/Probe.mse-set",
		Verzeichnis:       "ausfuhr-files",
	}}.Fuehre(t.Context(),
		vorlageMitFassung(`write_image_file(file: "s.png", width: 20,`+
			` symbol_variation(symbol: set.symbol, variation: "rare"))`),
		eineKarte(), []domain.Value{{Field: "symbol", Text: "s.mse-symbol"}}, nil)
	if err == nil || !strings.Contains(err.Error(), "keine Bildbäume") {
		t.Fatalf("Fehler = %v", err)
	}
}

// **`write_image_file` ohne Kartenschreiber und mit einer Karte.**
//
// Die Gegenrichtung: wer nur Bildwerte schreiben darf, bekommt für eine Karte
// einen benannten Fehler.
func TestOhneKartenschreiberIstEineKarteEinFehler(t *testing.T) {
	t.Parallel()

	_, _, err := mitSetsymbol(&beigabenmitschrift{}, &baumrastererprobe{}, &symbolgestaltprobe{}).
		Fuehre(t.Context(),
			vorlageMitFassung(`write_image_file(cards[0], file: "k.png")`),
			eineKarte(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "keine Karten schreiben") {
		t.Fatalf("Fehler = %v", err)
	}
}

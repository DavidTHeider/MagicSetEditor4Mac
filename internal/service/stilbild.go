package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// bildebene baut die Ebene eines Felds, das ein Bild zeichnet.
//
// # Was `image:` von `default:` unterscheidet
//
// `image:` rechnet, `default:` fängt auf. Ergibt `image:` nichts, ist das im
// Normalfall **keine Lücke**, sondern die Antwort: `{if styling.overlay == ""
// then nil else …}` steht 99-mal in den Stilen von `M15-Magic-Pack` und meint
// bei leerer Wahl „keine Überlagerung". Erst wenn danach auch `default:` nichts
// hergibt, bleibt die Ebene weg — ohne Befund, weil es keiner ist.
//
// # Und was `default:` über das Feld verrät
//
// Es steht über alle 190 Stile des Packs an **genau zwei** Feldnamen — `image`
// und `image 2` — und **nie** neben einem `image:`. Es ist damit kein zweites
// Bildskript, sondern das, was gilt, **wenn** die Karte nichts mitbringt.
//
// # Was `setbild` sagt
//
// Ob dieses Feld eines ist, an dem die **Karte** ihr Bild mitbringt — ein
// Kartenfeld der Art `image`. Trägt sie eines, wird **es** gezeichnet: die
// Datei liegt im Set-Paket, und dorthin führt seit
// #380 ein Weg
// ([msepaket.SetpaketVerweis]).
//
// **Die Art des Feldes und nicht sein `default:`.** Bis dahin stand
// `feld.Standardbild != ""` als Marker; das traf die 190 Stile des Packs, war
// aber der falsche Grund — ein Kunstfenster ohne Ersatzbild ist immer noch
// eines. Wo die Art unbekannt ist, weil kein Spielpaket dalag, fällt
// `istBildfeld` auf den alten Marker zurück.
func bildebene(
	ctx context.Context, feld *msepaket.Feldstil, eigenes, paket string,
	rahmen render.Rechteck, umgebung *msescript.Umgebung, symbole Symbolquelle,
	passung *render.Bildpassung, setbild bool, setpaket, ablage string,
) (*render.Ebene, []domain.Befund) {
	pfad, befunde := bildangabe(ctx, feld.Feld, "image", feld.Bild, umgebung)

	// **Die Wahlbilder vor dem Ersatzbild.** Ein `choice images:`-Block ist die
	// Angabe des Stils für genau diesen Feldwert; `default:` ist das, was gilt,
	// wenn es keine gibt.
	//
	// **Und sie gelten nur, wenn kein `image:` dasteht** — nicht, wenn dessen
	// Ergebnis leer ist:
	//
	//	void ChoiceStyle::initImage() {
	//	  if (image.isSet() || choice_images.empty()) return;
	//
	// `src/data/field/choice.cpp`, Zeile 172, Fassung f9d9356 vom 2021-07-06.
	// Steht ein `image:` da, werden die Wahlbilder gar nicht erst in das
	// Skript übersetzt; ergibt es nichts, wird nichts gezeichnet
	// (`draw_choice_viewer` prüft `img.isReady()`, `render/value/choice.cpp`,
	// Zeilen 61–76).
	//
	// **Woran es hing:** `card stamp` führt beides, und sein Skript gibt für
	// den Stempel „none" die leere Zeichenkette (`magic.mse-game/script`,
	// Zeile 4565). Wir nahmen dann die Wahlliste, dort steht
	// `stamps/none.png` — und diese Datei ist nicht leer. Auf jeder Karte der
	// M15-Vorlagen lag ein dunkler Fleck bei (166, 473)
	// (#454).
	//
	// Über `Full-Magic-Pack` führen **6** Feldstile beides, 1.558 nur
	// Wahlbilder und 2.920 nur `image:`; gezählt am Feldstil (§10). Einer der
	// sechs ist der Stempel.
	if strings.TrimSpace(feld.Bild) == "" && len(feld.Wahlbilder) > 0 {
		gewaehlt, weitere := wahlbild(ctx, feld, eigenes, umgebung)
		befunde = append(befunde, weitere...)
		pfad = gewaehlt
	}

	// **Das Bild der Karte vor dem Ersatzbild**, und beides vor gar nichts:
	//
	//	if (!value().filename.empty()) { … }
	//	// nice placeholder
	//	if (!image.Ok() && style().default_image.isReady()) { … }
	//
	// `ImageValueViewer::draw`, `src/render/value/image.cpp`, Zeilen 30–46,
	// Fassung f9d9356 vom 2021-07-06.
	if pfad.Knotenart == "" && eigenes != "" && setbild {
		gefunden, weitere := kartenbild(feld.Feld, eigenes, setpaket, ablage)
		befunde = append(befunde, weitere...)
		pfad = gefunden
	}

	if pfad.Knotenart == "" && feld.Standardbild != "" {
		ersatz, weitere := bildangabe(ctx, feld.Feld, "default", feld.Standardbild, umgebung)
		befunde = append(befunde, weitere...)
		pfad = ersatz
	}

	// **Ohne Bild zeigt ein Bildfeld das Schachbrett** — ab 41 Punkten Breite,
	// und auch beim Export:
	//
	//	// checkerboard placeholder
	//	if (!image.Ok() && !bitmap.Ok() && style().width > 40) {
	//	  bitmap = imagePlaceholder(dc, w, h, wxNullImage, what & DRAW_EDITING);
	//
	// `ImageValueViewer::draw`, `src/render/value/image.cpp`, Zeilen 54–61,
	// Fassung f9d9356 vom 2021-07-06. Nur ein Bildfeld der Karte, keine Wahl:
	// die schweigt ohne Bild (#506). Bis F02
	// (#2244)
	// blieb das Fenster leer, und bei einem Rahmen aus JPG stand dort Schwarz.
	if pfad.Knotenart == "" && (!setbild || rahmen.Breite <= 40) {
		return nil, befunde
	}

	maske, weitere := bildangabe(ctx, feld.Feld, "mask", feld.Maske, umgebung)
	befunde = append(befunde, weitere...)

	baum := render.Bildbaum{Knotenart: render.Platzhalterknoten}
	if pfad.Knotenart != "" {
		baum = alsZeichenbaum(pfad, paket, symbole)
	}

	return &render.Ebene{
		Art:    render.ArtBild,
		Rahmen: rahmen,
		Bild:   baum,
		Maske:  alsZeichenbaum(maske, paket, symbole),

		// **Die Mischart steht am Baum und gehört an die Ebene.** MSE fragt
		// sie am erzeugten Bild und gibt sie dem Zeichner mit
		// (`src/render/value/choice.cpp`, Zeilen 78–90, Fassung f9d9356 vom
		// 2021-07-06); hier steht der Plan dazwischen, und er muss sie tragen
		// (#470).
		Mischart: baum.Mischart(),

		Winkel:  winkelAus(ctx, feld, umgebung),
		Deckung: 1,
		Passung: passung,
	}, befunde
}

// kartenbild macht aus dem Dateinamen der Karte einen Verweis ins Set-Paket.
//
// **Ohne Ort kein Verweis.** Ein Set, dessen Herkunft der Bestand nicht führt,
// hat kein Paket, in dem die Datei läge; das Ersatzbild danach ist dann die
// richtige Antwort — aber eine, die gesagt wird (§8).
//
// # Die Ablage geht vor
//
// Ein Bild, das jemand **in dieser Anwendung** eingesetzt hat, liegt neben dem
// Set ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md)) —
// und nicht im Herkunftspaket. Läge das Paket vorn, zeigte die Karte nach dem
// Einsetzen weiter das alte Bild.
//
// **Gefragt wird die Platte, nicht eine Liste.** Ob eine Datei in der Ablage
// liegt, weiß nur sie; ein Verzeichnis im Speicher müsste beim Anlegen,
// Löschen und Speichern nachgezogen werden, und die dritte Stelle vergisst es.
func kartenbild(feld, datei, setpaket, ablage string) (msescript.Bildbaum, []domain.Befund) {
	if ort := abgelegt(ablage, datei); ort != "" {
		return msescript.Bildbaum{
			Knotenart: "datei", Pfad: msepaket.SetpaketVerweis(ort, datei),
		}, nil
	}

	if strings.TrimSpace(setpaket) == "" {
		return msescript.Bildbaum{}, []domain.Befund{{
			Ort: feld,
			Meldung: "die Karte trägt für dieses Feld ein eigenes Bild (" + datei +
				"); der Ort des Set-Pakets ist nicht bekannt — gezeichnet wird das Ersatzbild",
		}}
	}

	return msescript.Bildbaum{
		Knotenart: "datei", Pfad: msepaket.SetpaketVerweis(setpaket, datei),
	}, nil
}

// bildangabe wertet eine Angabe aus, die ein Dateiname **oder** ein Skript ist.
//
// Beide Formen stehen nebeneinander: von den 352 `mask:`-Angaben des Packs sind
// 174 Skripte und 173 einfache Namen. Ein Leser, der nur eine Form kennt, deutet
// die Hälfte nicht.
//
// **Und eine dritte**: mit `{` in der Mitte ist die Angabe eine Zeichenkette
// mit Ausdruck — `border-mask{pt()}.png` in `magic-crystal`
// (`parse(*this, true)` in `Reader::handle(ScriptableImage&)`,
// `src/script/image.cpp` Zeile 73, Fassung f9d9356 vom 2021-07-06; #1934).
func bildangabe(
	ctx context.Context, feld, angabe, quelle string, umgebung *msescript.Umgebung,
) (msescript.Bildbaum, []domain.Befund) {
	baum, befunde := gelesenerBildbaum(ctx, feld, angabe, quelle, umgebung)

	if erfundenerSetort(baum) {
		return msescript.Bildbaum{}, append(befunde, domain.Befund{
			Ort: feld + " → " + angabe,
			Meldung: "die Angabe nennt einen Ort auf der Platte („setpaket:“) — eine Vorlage " +
				"liest nur aus ihren Paketen; gezeichnet wird an dieser Stelle nichts",
		})
	}

	return baum, befunde
}

// gelesenerBildbaum liest eine Angabe — wörtlich oder gerechnet.
func gelesenerBildbaum(
	ctx context.Context, feld, angabe, quelle string, umgebung *msescript.Umgebung,
) (msescript.Bildbaum, []domain.Befund) {
	roh := strings.TrimSpace(quelle)
	if roh == "" {
		return msescript.Bildbaum{}, nil
	}

	var (
		wert msescript.Wert
		err  error
	)

	switch rumpf, geklammert := skriptrumpf(roh); {
	case geklammert:
		wert, err = msescript.AuswertenQuelle(ctx, rumpf, umgebung)

	case !bildGerechnet(roh):
		return msescript.Bildbaum{Knotenart: "datei", Pfad: roh}, nil

	default:
		if befund := textskriptbefund(roh); befund != "" {
			return msescript.Bildbaum{}, []domain.Befund{{Ort: feld + " → " + angabe, Meldung: befund}}
		}

		wert, err = msescript.AuswertenTextskript(ctx, roh, umgebung)
	}

	if err != nil {
		return msescript.Bildbaum{}, []domain.Befund{{
			Ort:     feld + " → " + angabe,
			Meldung: "die Angabe ließ sich nicht ausrechnen: " + err.Error(),
		}}
	}

	baum, istBild := wert.AlsBild()
	if !istBild {
		return msescript.Bildbaum{}, nil
	}

	return baum, nil
}

// bildGerechnet sagt, ob ein Bildwert gerechnet wird — MSE2s `isScripted()`
// (`src/script/image.hpp` Zeile 33, Fassung f9d9356 vom 2021-07-06): wahr bei
// jeder `{`, ob ganz geklammert oder in der Mitte. Ein `script:`-Unterblock
// kommt aus `msepaket` schon geklammert an.
func bildGerechnet(roh string) bool {
	return strings.Contains(roh, "{")
}

// textskriptbefund sagt, warum eine Angabe mit Ausdruck sich nicht lesen
// lässt — oder nichts.
//
// **Eine offene Klammer ist kein Dateiname.** Wer sie als solchen
// durchreichte, baute einen Pfad wie `/paket.mse-style/{ if …` — die Meldung
// käme dann aus der Bildquelle und sagte „Datei nicht gefunden", wo „Angabe
// kaputt" richtig wäre. **Streng geprüft** (`msescript.TextskriptFehler`),
// denn `AuswertenTextskript` ließe einen kaputten Ausdruck als Text stehen.
func textskriptbefund(roh string) string {
	err := msescript.TextskriptFehler(roh)
	if err == nil {
		return ""
	}

	if strings.HasPrefix(roh, "{") {
		return "die Angabe beginnt mit { und endet nicht darauf: " + roh
	}

	return "die Angabe ließ sich nicht lesen: " + msescript.Befund(err).Meldung
}

// skriptrumpf streift die geschweifte Klammer ab.
//
// **Sie ist in MSE-Script ein Funktionswert** (`docs/FORMATE.md`): wer sie
// auswertet, bekommt die Funktion und nicht ihr Ergebnis. Ein Feldstil meint
// aber den Wert — was dazwischen steht, ist der Ausdruck.
func skriptrumpf(roh string) (string, bool) {
	innen, auf := strings.CutPrefix(roh, "{")
	if !auf {
		return roh, false
	}

	rumpf, zu := strings.CutSuffix(strings.TrimSpace(innen), "}")
	if !zu {
		return roh, false
	}

	return strings.TrimSpace(rumpf), true
}

// ungedeutet sagt, **warum** ein Feld nicht gezeichnet wird.
//
// # Warum ein Satz je Lage und kein Sammelsatz
//
// **Ein Sammelsatz sagt bei verschiedenen Lagen dasselbe und bei keiner, was
// fehlt** — und sobald etwas davon gedeutet ist, zählt er Gedeutetes mit und
// ist falsch.
func ungedeutet(feld *msepaket.Feldstil) string {
	if feld.Symbolschrift.Name != "" {
		return "dieses Feld wird über die Symbolschrift " + feld.Symbolschrift.Name +
			" gezeichnet; Symbolschriften sind noch nicht gedeutet"
	}

	return "dieses Feld nennt weder eine Schrift noch ein Bild; " +
		"es ist noch nicht gedeutet"
}

// abgelegt sagt, ob eine Datei in der Bildablage liegt — und gibt dann deren
// Ort zurück.
//
// **`filepath.Base` gegen den Pfaddurchgriff.** Der Dateiname kommt aus einem
// Feldwert, und ein Set ist fremde Datei (§8): `../../../etc/passwd` als
// Bildname wäre sonst ein Weg aus dem Verzeichnis heraus.
func abgelegt(ablage, datei string) string {
	if strings.TrimSpace(ablage) == "" || strings.TrimSpace(datei) == "" {
		return ""
	}

	if _, err := os.Stat(filepath.Join(ablage, filepath.Base(datei))); err != nil {
		return ""
	}

	return ablage
}

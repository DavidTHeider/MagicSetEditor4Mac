package service

import (
	"path"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// alsZeichenbaum macht aus dem Bildbaum eines Skripts den des Zeichenplans.
//
// # Warum zwei Bäume und nicht einer
//
// **`msescript` darf `render` nicht kennen und umgekehrt** (§2). Der eine Baum
// ist ein Wert einer Sprache, der andere eine Zeichenanweisung; dass sie
// dieselbe Gestalt haben, ist kein Grund, sie zu einem zu machen. Diese Stelle
// ist die Naht, und sie gehört nach `service` — dorthin, wo die drei
// zusammenlaufen.
//
// # Die Paketpfade entstehen hier
//
// Ein Name ohne `/` meint das **eigene** Paket, und das weiß nur der Aufrufer.
// Deshalb wandert die Wurzel mit hinein statt hinterher darübergelegt zu
// werden: bei einem Baum gäbe es sonst kein „hinterher", sondern viele Blätter.
func alsZeichenbaum(baum msescript.Bildbaum, paket string, symbole Symbolquelle) render.Bildbaum {
	// **Kein Zweig für den leeren Baum.** Der Aufrufer prüft ihn, bevor er hier
	// hereinkommt, und ein Kind ist nie leer — [msescript.Wert.AlsBild] gibt
	// entweder einen gültigen Knoten oder gar keinen. Ein Weg, den kein Aufruf
	// betreten kann, ist ein Zweig zu viel (§3).
	if baum.Knotenart == "datei" {
		return render.Blatt(paketpfad(paket, baum.Pfad))
	}

	umgesetzt := render.Bildbaum{Knotenart: baum.Knotenart}

	for _, kind := range baum.Kinder {
		umgesetzt.Kinder = append(umgesetzt.Kinder, render.Bildkind{
			Name: kind.Name, Baum: alsZeichenbaum(kind.Baum, paket, symbole),
		})
	}

	for _, eintrag := range baum.Werte {
		uebernimm(&umgesetzt, eintrag)
	}

	// **Die Gestalt einer `.mse-symbol` reist mit dem Plan.** Der Rasterer
	// öffnet keine Pakete (§2, Vermerk 0022); gelesen wird hier, gezeichnet
	// dort (#323).
	if umgesetzt.Knotenart == "symbol_variation" {
		umgesetzt.Symbol = alsZeichenteile(symbolgestalt(symbole, umgesetzt.Texte["symbol"]).Teile)
	}

	return umgesetzt
}

// symbolgestalt liest die Gestalt — und gibt ohne Datei MSEs Standardsymbol.
//
//	if (filename.empty()) {
//	  the_symbol = default_symbol();
//	} else {
//	  the_symbol = package->readFile<SymbolP>(filename);
//	}
//
// `SymbolToImage::generate`, `src/gfx/generated_image.cpp`, Zeilen 457–462,
// Fassung f9d9356 vom 2021-07-06.
//
// **Ein Set ohne eigenes Symbol ist der Normalfall.** `set.symbol` ist dann
// leer, und MSE malt an der Typzeile ein Quadrat; dieses Programm meldete bis
// hierher „die Symbolzeichnung liegt nicht vor" und zeichnete nichts
// (#270).
//
// **Hier und nicht an der Bildquelle**: MSE entscheidet es beim Zeichnen, und
// so gilt es für jede [Symbolquelle] — auch für die eines Tests.
func symbolgestalt(symbole Symbolquelle, verweis string) msepaket.Symbolzeichnung {
	if strings.TrimSpace(verweis) == "" {
		return msepaket.Standardsymbol()
	}

	// **Eine Quelle steht immer davor.** Seit [SymbolquelleAusAblage] im
	// Planbauer davor hängt (#890), kommt hier nie ein blanker `nil` an, und ein
	// `if symbole == nil` wäre ein Zweig, den kein Aufruf betritt — eine
	// Behauptung (§3).
	//
	// **Die leere Gestalt gibt es weiterhin**, nur eine Ebene höher: eine
	// Ablage ohne Ort und ohne Quelle dahinter antwortet leer, und der Rasterer
	// benennt es. Das Standardsymbol braucht kein Paket — eine benannte Datei
	// schon.
	return symbole.Symbolgestalt(verweis)
}

// Symbolquelle liest die Gestalt einer `.mse-symbol` aus einem Paket.
//
// **Eine eigene Schnittstelle neben [render.Bildquelle]**, weil die Frage eine
// andere ist: der Rasterer will Bilder, `symbol_variation` will Formen. Wer nur
// Bilder liefert, antwortet hier nie — und dann bleibt die Gestalt leer, was
// der Rasterer benennt.
type Symbolquelle interface {
	Symbolgestalt(verweis string) msepaket.Symbolzeichnung
}

// uebernimm sortiert ein benanntes Argument nach seiner Art ein.
//
// **Nach Art getrennt und nicht als Text**, denn eine Farbe als Zeichenkette
// wieder zu zerlegen hieße, sie zweimal zu deuten — und die zweite Deutung
// könnte anders ausfallen als die erste.
//
// **Eine Zeichenkette ist trotzdem eine Farbe, wenn sie eine benennt.**
// `ScriptValue::toColor` (`src/script/value.cpp`, Zeilen 354–360, Fassung
// f9d9356 vom 2021-07-06) schickt **jede** Zeichenkette durch `parse_color`
// (`src/gfx/color.cpp`, Zeile 42); eine Farbe ist bei MSE-Script kein eigener
// Typ, den man führen oder verfehlen kann, sondern eine Lesart.
//
// `planechase.mse-game` schreibt `symbol_font_small_color := { "black" }`, und
// damit kam bei `recolor_image(color: …)` eine Zeichenkette an
// (#422). Sie
// fiel unter die Texte, und der Rasterer sagte, ihm fehle `color:`.
//
// **Der Satz oben bleibt richtig**: gedeutet wird hier **einmal**, und wer
// schon eine Farbe schickt, wird nicht noch einmal gelesen.
func uebernimm(ziel *render.Bildbaum, eintrag msescript.Eintrag) {
	farbe, istFarbe := gedeuteteWertfarbe(eintrag.Wert)
	if istFarbe {
		if ziel.Farben == nil {
			ziel.Farben = map[string]render.Farbe{}
		}

		ziel.Farben[eintrag.Name] = farbe

		return
	}

	if eintrag.Wert.Art == msescript.ArtZahlwert {
		if ziel.Zahlen == nil {
			ziel.Zahlen = map[string]float64{}
		}

		ziel.Zahlen[eintrag.Name] = eintrag.Wert.Zahl

		return
	}

	if ziel.Texte == nil {
		ziel.Texte = map[string]string{}
	}

	ziel.Texte[eintrag.Name] = eintrag.Wert.AlsText()
}

// gedeuteteWertfarbe liest eine Farbe aus einem Skriptwert.
//
// Erst als Farbwert, dann als Zeichenkette — der Grund steht bei [uebernimm].
func gedeuteteWertfarbe(wert msescript.Wert) (render.Farbe, bool) {
	if farbe, istFarbe := wert.AlsFarbe(); istFarbe {
		return render.Farbe{R: farbe.R, G: farbe.G, B: farbe.B, A: farbe.A}, true
	}

	if wert.Art != msescript.ArtTextwert {
		return render.Farbe{}, false
	}

	return gedeuteteFarbe(wert.Text)
}

// paketpfad macht aus einer Stilangabe den Paketpfad, den der Plan trägt.
//
// **Ein Name ohne `/` meint das eigene Paket.** `mask: border_mask.png` und
// `mask: /magic-genevensis-00-main.mse-style/pride_mask.png` stehen beide im
// Pack; wer die erste Form roh weiterreichte, suchte sie in der Paketwurzel und
// fände sie nie.
func paketpfad(paket, pfad string) string {
	if pfad == "" || strings.HasPrefix(pfad, "/") {
		return pfad
	}

	// **Ein Verweis ins Set-Paket trägt seinen Ort schon.** Ihm das eigene
	// Paket voranzustellen machte aus `setpaket:/Orte/S.mse-set!image1` den
	// Pfad `/probe.mse-style/setpaket:…` — eine Datei, die es nirgends gibt
	// (#380).
	if _, _, ist := msepaket.ZerlegeSetpaketverweis(pfad); ist {
		return pfad
	}

	return path.Join("/", paket, pfad)
}

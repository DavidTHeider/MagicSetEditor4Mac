// **Der Symboleditor über MCP**
// (#1391).
//
// # Warum das hierher gehört
//
// §2 sagt, wofür es dieses Paket gibt: „Werkzeuge, mit denen ein Sprachmodell
// Karten liest und ändert." Ein Setsymbol aus einer Beschreibung zu bauen ist
// genau dieser Fall — und dafür braucht es keine Maus, sondern die
// Rechnungen, die das Fenster benutzt.
//
// # Warum fast alle `lesen.` heißen
//
// **Weil sie nichts ändern.** Sie nehmen Teile entgegen und geben Teile
// zurück; im Bestand bleibt alles, wie es war. Nur `schreiben.symbol` legt
// etwas neben das Set. Das Präfix trägt die **Art** eines Werkzeugs, und die
// ist hier lesend — auch wenn „einen Handgriff ziehen" nach einer Änderung
// klingt.
//
// # Was ein Aufrufer mitbringen muss
//
// **Das ganze Symbol.** Es gibt keinen Sitzungszustand: `lesen.symbol` holt
// es, die Rechnungen geben je ein neues zurück, `schreiben.symbol` legt es
// ab. Warum kein „gerade offenes Symbol", steht in
// `internal/app/symbolwerkstattnaht.go`.

package mcp

import (
	"context"
	"encoding/json"
)

// Symbolwerkstatt ist, was die Symbolwerkzeuge brauchen.
//
// **Die Rechnungen stehen nicht darin.** Sie sind reine Funktionen aus
// `internal/app` und brauchen keine Naht; hier stehen nur die beiden Enden,
// die den Bestand anfassen.
type Symbolwerkstatt interface {
	// **Mit dem Kontext des Rufs** (#1815): er trägt das Profil, und nur
	// darin gibt es das Set. Ohne ihn fasste die Werkstatt jedes Set an.
	SymbolLesen(ctx context.Context, setID int64, name string) (teile any, fehler error)
	SymbolAblegen(ctx context.Context, setID int64, teile json.RawMessage, unter string) (name string, fehler error)

	// Rechnen führt eine der Rechnungen aus — benannt durch `was`.
	//
	// **Eine Methode und nicht zwölf.** `internal/mcp` ist ein Blatt (§2) und
	// darf `internal/app` nicht kennen; jede Rechnung einzeln durchzureichen
	// hieße, ihre zwölf Argumentgestalten hier ein zweites Mal zu
	// beschreiben. Was die Werkzeuge unterscheidet, ist ihr **Schema** — und
	// das steht hier; was sie rechnen, steht drüben.
	SymbolRechnen(was string, argumente json.RawMessage) (ergebnis any, fehler error)

	// Palette gibt, was ein Kartenfeld zum Einfügen anbietet — **ohne die
	// gerasterten Bildchen**. Warum, steht in
	// `internal/app/symbolpalettenaht.go`.
	Palette(ctx context.Context, karteID int64, feld string) (eintraege any, fehler error)
}

// symbolrechnung beschreibt ein Werkzeug, das Teile in Teile rechnet.
type symbolrechnung struct {
	name       string
	erklaerung string
	felder     map[string]any
	pflicht    []string
}

// symbolrechnungen sind die neun Fähigkeiten des Editors, kopflos.
//
// **Neun und nicht siebzehn.** Das Fenster hat für den Zuschnitt zwei
// Bindungen — ziehen und begrenzen —, und `docs/FAEHIGKEITEN.md` nennt sie
// selbst „zwei Schritte **einer**". Ohne Zeigegerät gibt es kein Ziehen; was
// bleibt, ist das Rechteck. Dieselbe Überlegung trennt „einen Punkt einfügen"
// nicht von „einen Punkt löschen": es ist eine Rechnung mit einer Weise.
func symbolrechnungen() []symbolrechnung {
	teilefeld := map[string]any{
		"teile": rohliste("Die Teile des Symbols — aus »lesen.symbol« oder aus einer " +
			"vorigen Rechnung. Ein Teil hat »art« (Form, Symmetrie, Gruppe), »name«, " +
			"»kombination«, »punkte« und bei einer Symmetrie »spiegelart«, »kopien«, " +
			"»mitteX«, »mitteY«, »griffX«, »griffY«; eine Gruppe trägt »teile«."),
	}

	mit := func(felder map[string]any) map[string]any {
		zusammen := map[string]any{}
		for name, wert := range teilefeld {
			zusammen[name] = wert
		}

		for name, wert := range felder {
			zusammen[name] = wert
		}

		return zusammen
	}

	return []symbolrechnung{
		{
			name: "lesen.symbol.grundform",
			erklaerung: "Gibt eine der vier Grundformen als Symbolteil — mittig und in " +
				"voller Größe. ÄNDERT NICHTS. Der Anfang jeder Zeichnung: das Ergebnis " +
				"geht in »teile« der übrigen Werkzeuge.",
			felder: map[string]any{
				"art":   text("Die Grundform: »Kreis«, »Rechteck«, »Vieleck« oder »Stern«."),
				"ecken": zahl("Die Zahl der Ecken — nur bei Vieleck und Stern."),
				"eng":   jaNein("Gleichseitig: Kreis statt Ellipse, Quadrat statt Rechteck."),
			},
			pflicht: []string{"art"},
		},
		{
			name: "lesen.symbol.griff",
			erklaerung: "Setzt einen Bézierhandgriff eines Punktes auf eine Stelle. " +
				"ÄNDERT NICHTS im Bestand — gibt die geänderten Teile zurück. Die " +
				"Stelle ist die SPITZE des Griffs, nicht seine Länge.",
			felder: mit(map[string]any{
				"teil":  zahl("Die Stelle des Teils in »teile«, ab null."),
				"punkt": zahl("Die Stelle des Punktes in diesem Teil, ab null."),
				"davor": jaNein("Der Griff VOR dem Punkt statt dahinter."),
				"x":     bruch("Die Stelle der Griffspitze, waagerecht, 0 bis 1."),
				"y":     bruch("Die Stelle der Griffspitze, senkrecht, 0 bis 1."),
			}),
			pflicht: []string{"teile", "teil", "punkt", "x", "y"},
		},
		{
			name: "lesen.symbol.sperre",
			erklaerung: "Gibt einem Punkt seine Art: »free« (frei), »direction« (glatt) " +
				"oder »size« (symmetrisch) — und mittelt dabei seine beiden Griffe wie MSE. " +
				"ÄNDERT NICHTS im Bestand — gibt die geänderten Teile zurück.",
			felder: mit(map[string]any{
				"teil":   zahl("Die Stelle des Teils in »teile«, ab null."),
				"punkt":  zahl("Die Stelle des Punktes in diesem Teil, ab null."),
				"sperre": text("»free«, »direction« oder »size«."),
			}),
			pflicht: []string{"teile", "teil", "punkt", "sperre"},
		},
		{
			name: "lesen.symbol.punkt",
			erklaerung: "Fügt einen Punkt ein oder löscht einen. ÄNDERT NICHTS im " +
				"Bestand. Eingefügt wird in die MITTE der Strecke hinter dem genannten " +
				"Punkt; gelöscht wird der genannte.",
			felder: mit(map[string]any{
				"teil":  zahl("Die Stelle des Teils in »teile«, ab null."),
				"punkt": zahl("Die Stelle des Punktes in diesem Teil, ab null."),
				"weise": text("»einfuegen« oder »loeschen«."),
			}),
			pflicht: []string{"teile", "teil", "punkt", "weise"},
		},
		{
			name: "lesen.symbol.gruppe",
			erklaerung: "Fasst Teile zu einer Gruppe zusammen oder löst eine auf. " +
				"ÄNDERT NICHTS im Bestand. Eine Gruppe wird als eines gezeichnet und " +
				"als eines gespiegelt.",
			felder: mit(map[string]any{
				"stellen": zahlenliste("Die Stellen der Teile in »teile«, ab null.",
					"Eine Stelle."),
				"weise": text("»gruppieren« oder »aufloesen«."),
			}),
			pflicht: []string{"teile", "stellen", "weise"},
		},
		{
			name: "lesen.symbol.reihenfolge",
			erklaerung: "Verschiebt ein Teil in der Zeichenreihenfolge. ÄNDERT NICHTS " +
				"im Bestand. Die Reihenfolge IST die Zeichenreihenfolge: was später " +
				"kommt, liegt oben.",
			felder: mit(map[string]any{
				"von":  zahl("Die Stelle des Teils heute, ab null."),
				"nach": zahl("Die Stelle, an die es soll, ab null."),
			}),
			pflicht: []string{"teile", "von", "nach"},
		},
		{
			name: "lesen.symbol.symmetrie",
			erklaerung: "Legt eine Symmetrie an, ändert sie oder löst sie auf. ÄNDERT " +
				"NICHTS im Bestand. Eine Symmetrie zeichnet ihr Teil mehrfach — " +
				"gespiegelt oder gedreht um eine Mitte.",
			felder: mit(map[string]any{
				"stelle":     zahl("Die Stelle des Teils in »teile«, ab null."),
				"weise":      text("»anlegen«, »setzen« oder »aufloesen«."),
				"spiegelart": text("Bei »setzen«: »rotation« oder »reflection«."),
				"kopien":     zahl("Bei »setzen«: die Zahl der Kopien, mindestens zwei."),
				"mitteX":     bruch("Bei »setzen«: die Mitte, waagerecht, 0 bis 1."),
				"mitteY":     bruch("Bei »setzen«: die Mitte, senkrecht, 0 bis 1."),
				"griffX":     bruch("Bei »setzen«: der Griff, waagerecht, 0 bis 1."),
				"griffY":     bruch("Bei »setzen«: der Griff, senkrecht, 0 bis 1."),
			}),
			pflicht: []string{"teile", "stelle", "weise"},
		},
		{
			name: "lesen.symbol.symmetriekopien",
			erklaerung: "Gibt die gespiegelten Züge einer Symmetrie — was sie zeichnet, " +
				"ohne dass es in den Teilen steht. ÄNDERT NICHTS.",
			felder: mit(map[string]any{
				"stelle": zahl("Die Stelle der Symmetrie in »teile«, ab null."),
			}),
			pflicht: []string{"teile", "stelle"},
		},
		{
			name: "lesen.symbol.zuschnitt",
			erklaerung: "Hält eine Bildauswahl in ihren Grenzen und gibt die gültige " +
				"zurück. ÄNDERT NICHTS. Das Fenster hat dafür zwei Bindungen — ziehen " +
				"und begrenzen; ohne Zeigegerät bleibt das Rechteck.",
			felder: map[string]any{
				"quelleBreite": zahl("Die Breite des Bildes in Bildpunkten."),
				"quelleHoehe":  zahl("Die Höhe des Bildes in Bildpunkten."),
				"zielBreite":   zahl("Die Breite des Feldes in Bildpunkten."),
				"zielHoehe":    zahl("Die Höhe des Feldes in Bildpunkten."),
				"auswahlX":     zahl("Die linke Kante der Auswahl in Bildpunkten."),
				"auswahlY":     zahl("Die obere Kante der Auswahl in Bildpunkten."),
				"breite":       zahl("Die Breite der Auswahl in Bildpunkten."),
				"hoehe":        zahl("Die Höhe der Auswahl in Bildpunkten."),
				"frei":         jaNein("Ohne festes Seitenverhältnis."),
				"bevorzugt":    text("Welche Kante nachgibt: »breite« oder »hoehe«."),
			},
			pflicht: []string{
				"quelleBreite", "quelleHoehe", "zielBreite", "zielHoehe",
				"auswahlX", "auswahlY", "breite", "hoehe",
			},
		},
	}
}

// Der Weg einer Exportvorlage bis vor das Schreiben.
//
// # Ein Weg, zwei Ausgänge
//
// **Der echte Lauf und der Probelauf gehen beide hier hindurch**
// (#2015,
// UXR-F7): dieselben Karten, dieselbe Auswahl, dieselben Setangaben, dieselbe
// Auswertung der Vorlage und damit dieselben Fehler. Was sie trennt, ist, wohin
// die Schreiber der Vorlage ihren Inhalt legen — auf die Platte oder in eine
// [service.Probeablage] — und dass der echte Lauf danach die Hauptdatei
// schreibt.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// vorlagenrechnung ist, was die Auswertung einer Vorlage ergibt.
type vorlagenrechnung struct {
	// ausgabe ist der Text, den die Vorlage zurückgibt — die Hauptdatei.
	ausgabe string

	befunde []domain.Befund

	// karten sind die Karten, über die die Vorlage lief — nach der Auswahl.
	karten int

	// verzeichnis ist der relative Name des Beigabenverzeichnisses; leer, wenn
	// die Vorlage keines will.
	verzeichnis string

	// gelaufen sagt, dass die Vorlage ausgewertet wurde. Ein Fehler davor
	// (Karten lesen, Auswahl, Ort) trägt keine Befunde der Vorlage.
	gelaufen bool
}

// orte nennt, wohin die Beigaben gingen: im echten Lauf wird das Verzeichnis
// angelegt, im Probelauf (`ablage` gesetzt) nur sein Name gebildet.
func orte(ziel string, gewollt bool, ablage *service.Probeablage) (ort, name string, fehler error) {
	if ablage != nil {
		ort, name = beigabenname(ziel, gewollt)

		return ort, name, nil
	}

	return beigabenort(ziel, gewollt)
}

// rechneVorlage wertet eine Vorlage über den Karten eines Sets aus.
//
// `ablage` ist `nil` im echten Lauf und die Ablage des Probelaufs sonst.
func (a *App) rechneVorlage(
	ctx context.Context, setID int64, vorlage msepaket.Exportvorlage,
	ziel string, wahl map[string]string, auswahl domain.Kartenauswahl,
	ablage *service.Probeablage,
) (vorlagenrechnung, error) {
	alle, err := a.leser().Karten(ctx, setID)
	if err != nil {
		return vorlagenrechnung{}, err
	}

	// **Die Auswahl wird hier angewandt, nicht im Dienst** — und das ist eine
	// Ausnahme mit Grund: die Karten werden an dieser Stelle gelesen, und der
	// Vorlagenausführer bekommt sie als Liste. Entschieden wird nichts:
	// [domain.Kartenauswahl] trägt die Regel, hier steht der Aufruf.
	if fehler := auswahl.Pruefe(alle); fehler != nil {
		return vorlagenrechnung{}, fehler
	}

	karten := auswahl.Anwenden(alle)

	// **Die Setangaben gehören dazu** (#609). Neun der achtzehn Fehlschläge
	// über `M15-Magic-Pack` waren `set.title` und `set.set_code`.
	//
	// **Ein Fehler dabei ist kein Abbruch**: die Vorlage läuft dann mit lauter
	// leeren Setfeldern, und das ist mehr als gar keine Ausfuhr. Was fehlt,
	// meldet sich am Glied.
	setwerte, _ := a.setwerte.Setwerte(ctx, setID) //nolint:errcheck // der Grund steht darüber

	// **Die Fähigkeiten, Dateien zu schreiben** (#216, #657).
	//
	// Sie werden **je Ausfuhr** gesetzt, denn beide hängen am Ziel: das
	// Verzeichnis steht erst fest, wenn jemand eines gewählt hat. Ohne sie gibt
	// es die Namen nicht — und genau das war der Zustand: `MitBildschreiber`
	// stand seit #216 da und wurde von keiner Zeile Produktivcode gerufen.
	// **Die Beigaben gehen in ein Unterverzeichnis, nicht daneben** (#732).
	//
	//	wxFileName fn(outname);
	//	info.directory_relative = fn.GetName() + _("-files");
	//	fn.SetFullName(info.directory_relative);
	//	if (!wxDirExists(info.directory_absolute)) wxMkdir(info.directory_absolute);
	//
	// `src/gui/html_export_window.cpp`, Zeilen 63–71, Fassung f9d9356 vom
	// 2021-07-06. Die Vorlage baut ihre URLs mit `{directory}/`; ohne das
	// Verzeichnis zeigte jedes `src` auf eine Datei, die woanders liegt.
	neben, unterName, fehler := orte(ziel, vorlage.Verzeichnis, ablage)
	if fehler != nil {
		return vorlagenrechnung{}, fehler
	}

	// **Ein Wert je Lauf und kein geteilter Zustand** —
	// #736.
	//
	// Es gibt **einen** `Vorlagenausfuehrer` je Anwendung
	// (`service/vorlagenliste.go`, Zeile 235). Bis zum 2026-08-21 setzte diese
	// Stelle seine Felder je Aufruf; `SetVorlagenausfuhr` sperrt aber über
	// `useVorgang` **je Set**, und zwei Ausfuhren verschiedener Sets laufen
	// damit ausdrücklich nebeneinander. Der zweite Lauf überschrieb dem ersten
	// sein Ziel — `go test -race` zeigte drei Rennzustände.
	//
	// **Erst stand hier eine Sperre.** Sie nahm der Sache die Wirkung, nicht
	// die Gestalt; der Wert nimmt ihr beides, und zwei Ausfuhren laufen wieder
	// wirklich nebeneinander.
	//
	// **`set.symbol` nennt eine `.mse-symbol` im Set** (#660); ohne dessen Ort
	// suchte die Quelle sie in der Paketwurzel. Ein Set ohne bekannten Ort
	// bekommt den Namen trotzdem — ein Set aus der Saat hat keine
	// Herkunftsdatei, und `symbol_variation` gibt dann MSEs Standardsymbol
	// (`generated_image.cpp`, Zeilen 457–462, Fassung f9d9356 vom 2021-07-06).
	mittel := service.Ausfuhrmittel{
		Bilder:   bildnaht{aus: a.bildausfuhr, ablage: ablage},
		Bildziel: neben,
		Beigaben: beigaben{
			ziel:          neben,
			vorlagenpaket: vorlage.Ort,
			setID:         setID,
			setschreibt:   a.speicherer,
			ablage:        ablage,
		},
		Symbolzeichnungen: a.symbolzeichnungen,
		Setpaket:          a.setort(ctx, setID),
		Stile:             a.ausfuhrstile(ctx, setID),
		Verzeichnis:       unterName,
	}

	ausgabe, befunde, err := a.vorlagen.Fuehre(ctx, mittel, vorlage, karten, setwerte, wahl)
	rechnung := vorlagenrechnung{
		ausgabe: ausgabe, befunde: befunde, karten: len(karten),
		verzeichnis: unterName, gelaufen: true,
	}

	return rechnung, err
}

// ausfuhrstile gibt der Ausfuhr die Stilpakete des Sets — `set.stylesheet`
// und `card.stylesheet` (QA-04-F3, #2039). **Jede Fläche hat eine Stilwahl**:
// das Fenster bekommt sie beim Starten, der Vorlagenweg von Kommandozeile,
// REST und MCP seit VK-01b (#2182). Hier stand ein Zweig „ohne Stilwahl keine
// Quelle" — er war der Weg dieser drei Flächen, und `set.stylesheet` war dort
// `nil`.
func (a *App) ausfuhrstile(ctx context.Context, setID int64) service.Ausfuhrstile {
	return func(paket string) (msepaket.Stil, bool) {
		stil, err := a.stile.Stilpaket(ctx, setID, paket)

		return stil, err == nil
	}
}

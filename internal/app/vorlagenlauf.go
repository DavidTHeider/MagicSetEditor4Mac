// Das Ausführen einer Exportvorlage und das Schreiben ihres Ergebnisses.
//
// **Eine eigene Datei**, seit `vorlagenausfuhr.go` an der Zeilengrenze aus §2
// stand. Der Schnitt folgt der Frage: dort steht, **welche** Vorlagen und
// Optionen es gibt, hier, **was** beim Ausführen einer geschieht: die
// Kartenauswahl, der Lauf selbst und die Datei am Ziel.

package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// VorlageAusfuehren führt eine Vorlage aus und schreibt ihr Ergebnis.
//
// **Der Pfad kommt vom Aufrufer**, nicht aus einem Dialog: `internal/app`
// trifft keine fachlichen Entscheidungen, und ein Dialog gehört an die
// Bindung (§2).
// **Und eine Teilmenge geht auch** — Werkstück 4 von Phase 3.6
// (#1167).
// `auswahl` leer heißt: das ganze Set. MSE2 stellt die Frage vor der
// HTML-Ausgabe genauso (`SetWindow::onFileExportHTML`,
// `src/gui/set/window.cpp`, Zeile 677, Fassung f9d9356 vom 2021-07-06).
func (a *App) VorlageAusfuehren(
	ctx context.Context, setID, paket, ziel string, wahl map[string]string,
	auswahl []string,
) VorlagenausfuhrAntwort {
	ctx = a.imProfil(ctx)

	leer := VorlagenausfuhrAntwort{Befunde: []string{}}

	if ziel == "" {
		leer.Abgebrochen = true

		return leer
	}

	kennung, befund := kennungAus(setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	// **Und vor allem anderen die Frage, ob dieser Aufruf schreiben darf**
	// (0078) — die Vorlage schreibt Dateien, und `query_only` sieht das
	// nicht.
	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	vorlage, gewaehlt, fehler := a.vorlageWaehlen(ctx, kennung, paket, auswahl)
	if fehler != nil {
		leer.Fehler = fehler

		return leer
	}

	return a.schreibeVorlage(ctx, kennung, vorlage, ziel, wahl, gewaehlt)
}

// vorlageWaehlen prüft die Auswahl und sucht die Vorlage — der Anfang beider
// Wege, des echten Laufs und des Probelaufs
// (#2015).
func (a *App) vorlageWaehlen(
	ctx context.Context, kennung int64, paket string, auswahl []string,
) (msepaket.Exportvorlage, domain.Kartenauswahl, *Fehler) {
	// **Die Auswahl wird vor dem Paket geprüft.** Eine unbrauchbare Kennung
	// ist eine Eingabe; sie erst nach dem Suchen der Vorlage abzuweisen hieße,
	// für einen Tippfehler das Paketverzeichnis zu durchsuchen — und die
	// Meldung wäre die des Pakets, wenn es das auch nicht gibt.
	gewaehlt, auswahlbefund := kartenauswahlAus(auswahl)
	if auswahlbefund != nil {
		return msepaket.Exportvorlage{}, domain.Kartenauswahl{}, auswahlbefund
	}

	vorlage, gefunden, err := a.vorlageMit(ctx, kennung, paket)
	if err != nil {
		return msepaket.Exportvorlage{}, domain.Kartenauswahl{}, Uebersetze(ctx, err)
	}

	if !gefunden {
		return msepaket.Exportvorlage{}, domain.Kartenauswahl{}, Uebersetze(ctx, fmt.Errorf(
			"%w: die Exportvorlage %q ist nicht zu finden", domain.ErrNotFound, paket))
	}

	return vorlage, gewaehlt, nil
}

func (a *App) schreibeVorlage(
	ctx context.Context, setID int64, vorlage msepaket.Exportvorlage,
	ziel string, wahl map[string]string, auswahl domain.Kartenauswahl,
) VorlagenausfuhrAntwort {
	antwort := VorlagenausfuhrAntwort{Befunde: []string{}}

	rechnung, err := a.rechneVorlage(ctx, setID, vorlage, ziel, wahl, auswahl, nil)
	if err != nil {
		antwort.Fehler = Uebersetze(ctx, err)

		// **Der Satz des Skriptfehlers reist als Befund mit.**
		//
		// `Uebersetze` ersetzt rohe Meldungen durch den Katalogsatz, und das ist
		// richtig: ein „SQL logic error" sagt Nutzenden nichts und verrät den
		// inneren Aufbau (§8). Hier liegt der Fall anders — „der Name
		// `write_image_file` ist hier nicht bekannt" ist eine Aussage über
		// **ihre** Vorlage, und ohne sie stünde nur „Die Eingabe ist nicht
		// gültig". #209
		// verlangt ausdrücklich, dass eine Vorlage sagt, **warum** sie nicht
		// durchläuft: sie ist der Normalfall, nicht die Ausnahme.
		if rechnung.gelaufen {
			antwort.Befunde = append(meldungen(rechnung.befunde), err.Error())
		}

		return antwort
	}

	ausgabe := rechnung.ausgabe

	if err := service.SchreibeAtomar(ziel, []byte(ausgabe)); err != nil {
		antwort.Fehler = Uebersetze(ctx, err)

		return antwort
	}

	antwort.Pfad = ziel
	antwort.Befunde = meldungen(rechnung.befunde)

	// **Eine leere Ausgabe wird geschrieben — und gesagt** (§15, §8).
	//
	// # Was MSE2 tut, nachgelesen
	//
	// `src/gui/html_export_window.cpp:81–87` (Fassung `f9d9356` vom
	// 2021-07-06) schreibt `result->toString()` **ohne jede Prüfung**, sobald
	// ein Dateiname dasteht:
	//
	//	if (!outname.empty()) {
	//	  wxFileOutputStream file(outname);
	//	  wxTextOutputStream stream(file);
	//	  stream.WriteString(result->toString());
	//	}
	//
	// Eine Vorlage, die nichts schreibt, erzeugt dort also eine Null-Byte-
	// Datei und meldet Erfolg. **§15 sagt: Semantik konservativ** — also wird
	// sie hier auch geschrieben.
	//
	// # Und warum trotzdem etwas dabeisteht
	//
	// Weil MSE2 dazu **schweigt**, und das ist der Teil, der nicht zu
	// portieren ist. Der Druckbogen weist seit dem 2026-08-30 ein leeres Set
	// ab — dort ist eine Datei ohne Seiten **kein Dokument**. Hier ist eine
	// leere Datei ein gültiges Ergebnis (eine CSV-Vorlage mit nur einer
	// Kopfzeile ist derselbe Fall), aber sie sieht wie ein Fehlschlag aus.
	//
	// **Ein Befund sagt es, ohne die Semantik zu ändern**
	// (#1347).
	if ausgabe == "" {
		antwort.Befunde = append(antwort.Befunde, befundLeereAusgabe)
	}

	// **Auch eine Exportvorlage ist eine Ausgabe** — wie der Bogen (FMT-03).
	a.arbeitVermerken(ctx, domain.ArbeitsartSet, strconv.FormatInt(setID, 10), domain.VorgangAusgegeben)

	return antwort
}

// befundLeereAusgabe sagt, dass eine Vorlage nichts ergeben hat — im echten Lauf
// wie im Probelauf derselbe Satz.
const befundLeereAusgabe = "diese Vorlage hat nichts geschrieben — die Datei ist leer"

// vorlageMit sucht eine Vorlage nach ihrem Paketnamen.
func (a *App) vorlageMit(ctx context.Context, setID int64, paket string) (
	msepaket.Exportvorlage, bool, error,
) {
	vorlagen, _, err := a.vorlagen.Vorlagen(ctx, setID)
	if err != nil {
		return msepaket.Exportvorlage{}, false, err
	}

	for i := range vorlagen {
		if vorlagen[i].Paket == paket {
			return vorlagen[i], true, nil
		}
	}

	return msepaket.Exportvorlage{}, false, nil
}

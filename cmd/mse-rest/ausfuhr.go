// Die Ausfuhr des Dienstes — einfahren, sichern, zeichnen
// (#997).
//
// # Warum die Wache hier steht und nicht in `internal/rest`
//
// Derselbe Grund wie bei `cmd/mse-mcp`: „nichts überschreiben" ist eine
// Aussage über die **Platte**, und `internal/rest` ist ein Blatt — es kennt
// weder Dateien noch die Anwendung.
//
// # Und warum der Dienst sie überhaupt hat
//
// [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)
// misst nicht die Menge einer Änderung, sondern ihre **Umkehrbarkeit**, und
// der Maßstab dafür ist die **Aufsicht**: wer tippt, sieht seinen Aufruf; wer
// über HTTP ruft, tut es, während niemand hinsieht. Der Dienst steht damit auf
// derselben Seite wie MCP und nicht auf der der Kommandozeile.

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ausfuhr erfüllt `rest.Ausfuhr` über den Nähten von `internal/app`.
type ausfuhr struct{ db *sql.DB }

// Einfahren liest ein Setpaket in den Bestand.
//
// **Ohne Wache**: die Regel „eine Quelle, ein Set" steht im Fachkern und macht
// den zweiten Ruf zu einem Lesevorgang — er gibt das Set zurück, das schon da
// ist. Es wird nichts überschrieben.
func (a ausfuhr) Einfahren(
	ctx context.Context, pfad string,
) (setID int64, karten int, befunde []string, fehler error) {
	ergebnis, err := app.Einfuhrnaht(a.db).Set(ctx, pfad)
	if err != nil {
		return 0, 0, nil, err
	}

	gefunden := make([]string, 0, len(ergebnis.Befunde))
	for _, befund := range ergebnis.Befunde {
		gefunden = append(gefunden, befund.String())
	}

	return ergebnis.SetID, ergebnis.Karten, gefunden, nil
}

// Sichern schreibt ein Set — nur dorthin, wo nichts liegt.
func (a ausfuhr) Sichern(ctx context.Context, setID int64, ziel string) ([]string, error) {
	aufraeumen, err := reserviereVerzeichnis(ziel)
	if err != nil {
		return nil, err
	}

	// **Die Reservierung fällt, wenn nichts daraus wird** (#1343):
	// eine Null-Byte-Datei sähe aus wie ein geglücktes Schreiben.
	geschrieben := false

	defer func() {
		if !geschrieben {
			aufraeumen()
		}
	}()

	schreiber, err := app.Setschreibnaht(a.db)
	if err != nil {
		return nil, err
	}

	ergebnis, fehler := schreiber.Schreibe(ctx, setID, ziel)
	if fehler != nil {
		return nil, fehler
	}

	geschrieben = true

	befunde := make([]string, 0, len(ergebnis.Befunde))
	for _, befund := range ergebnis.Befunde {
		befunde = append(befunde, befund.String())
	}

	return befunde, nil
}

// Zeichnen rastert eine Karte — nur dorthin, wo nichts liegt.
func (a ausfuhr) Zeichnen(
	ctx context.Context, karteID int64, ziel, guete string,
) (breite, hoehe int, befunde []string, fehler error) {
	gewaehlt, guetefehler := app.GueteAusName(guete)
	if guetefehler != nil {
		return 0, 0, nil, guetefehler
	}

	aufraeumen, err := reserviereOrt(ziel)
	if err != nil {
		return 0, 0, nil, err
	}

	// **Die Reservierung fällt, wenn nichts daraus wird** (#1343):
	// eine Null-Byte-Datei sähe aus wie ein geglücktes Schreiben.
	geschrieben := false

	defer func() {
		if !geschrieben {
			aufraeumen()
		}
	}()

	werk, err := app.Zeichennaht(a.db)
	if err != nil {
		return 0, 0, nil, err
	}

	gebild, schreibfehler := werk.SchreibeBild(ctx, karteID, ziel, gewaehlt)
	if schreibfehler != nil {
		return 0, 0, nil, schreibfehler
	}

	geschrieben = true

	return gebild.Breite, gebild.Hoehe, gebild.Befunde, nil
}

// Bogen legt die Karten eines Sets — nur dorthin, wo nichts liegt.
//
// **Dieselbe Wache wie über MCP** (#1291): der Dienst wird über die Leitung
// angesprochen, und wer ruft, sieht seinen Aufruf nicht.
func (a ausfuhr) Bogen(ctx context.Context, setID int64, ziel string) (
	seiten, karten int, befunde []string, fehler error,
) {
	aufraeumen, err := reserviereOrt(ziel)
	if err != nil {
		return 0, 0, nil, err
	}

	// **Die Reservierung fällt, wenn nichts daraus wird** (#1343):
	// eine Null-Byte-Datei sähe aus wie ein geglücktes Schreiben.
	geschrieben := false

	defer func() {
		if !geschrieben {
			aufraeumen()
		}
	}()

	werk, err := app.Zeichennaht(a.db)
	if err != nil {
		return 0, 0, nil, err
	}

	lauf, legefehler := werk.LegeBogen(ctx, setID, ziel)
	if legefehler != nil {
		return 0, 0, nil, legefehler
	}

	geschrieben = true

	return lauf.Seiten, lauf.Karten, lauf.Befunde, nil
}

// MTGJSONLesen fährt eine MTG-JSON-Quelle ein (#1291).
//
// **Ohne Wache**, wie das Einfahren eines Setpakets: „eine Quelle, ein Set"
// steht im Fachkern und macht den zweiten Ruf zu einem Lesevorgang.
func (a ausfuhr) MTGJSONLesen(ctx context.Context, pfad, spiel string) (
	angelegt []int64, befunde []string, fehler error,
) {
	eingefahren, err := app.MTGJSONLesen(ctx, a.db, pfad, spiel)
	if eingefahren == nil {
		return nil, nil, err
	}

	return eingefahren.Angelegt, eingefahren.Befunde, err
}

// MTGJSONSchreiben schreibt ein Set — nur dorthin, wo nichts liegt.
func (a ausfuhr) MTGJSONSchreiben(ctx context.Context, setID int64, spiel, ziel string) (
	karten int, befunde []string, fehler error,
) {
	aufraeumen, err := reserviereOrt(ziel)
	if err != nil {
		return 0, nil, err
	}

	// **Die Reservierung fällt, wenn nichts daraus wird** (#1343):
	// eine Null-Byte-Datei sähe aus wie ein geglücktes Schreiben.
	geschrieben := false

	defer func() {
		if !geschrieben {
			aufraeumen()
		}
	}()

	gejson, err := app.MTGJSONSchreiben(ctx, a.db, setID, spiel, ziel)
	if err != nil {
		return 0, nil, err
	}

	geschrieben = true

	return gejson.Karten, gejson.Befunde, nil
}

// MTGJSONSpiele nennt die Spiele, für die es eine Zuordnung gibt.
func (a ausfuhr) MTGJSONSpiele() []string { return app.MTGJSONSpiele() }

// freierOrt reserviert ein Ziel — und weist ab, wenn schon etwas dort liegt.
//
// # Warum reserviert und nicht nur nachgesehen
//
// Hier stand `os.Stat`, und der Kommentar räumte den Wettlauf ein: „was die
// Wache verhindert, ist das Wahrscheinliche — ein Modell, das nach einer
// Zeitüberschreitung denselben Aufruf wiederholt."
//
// **Der Befund traf trotzdem** (#1343):
// die Begründung stand im Quelltext, die **Zusage** in der Werkzeug-
// beschreibung — und dort ohne Einschränkung. Zwei gleichzeitig arbeitende
// Aufrufer sind bei einem Sprachmodell nicht ausgeschlossen; beide bestanden
// die Prüfung, und der spätere ersetzte den früheren.
//
// **`O_CREATE|O_EXCL` entscheidet es im Kern**: genau einer legt die Datei an,
// der andere bekommt `EEXIST`. Was danach schreibt, schreibt in **seine**
// Reservierung — `os.WriteFile` überschreibt sie, und `schreibeAtomar`
// benennt darüber.
//
// # Was sie zurückgibt
//
// Ein `aufraeumen`, das die leere Reservierung wieder wegnimmt. **Wer
// scheitert, hinterlässt sonst eine Null-Byte-Datei** — und die sähe aus wie
// ein geglücktes Schreiben (§7, Regel 3 andersherum: was **nicht**
// geschrieben ist, soll auch nicht dastehen).
//
// **`ErrConflict` und nicht `ErrValidation`**: der Aufruf war richtig, der Ort
// ist belegt (§8).
//
// # Was danach schreibt, verträgt die Reservierung
//
// Nachgelesen am 2026-08-31, je Weg bis zum letzten Plattenzugriff — die
// Reichweite dieser Aussage ist genau diese fünf:
//
//   - `Zeichnen` → `service.Bildausfuhr.Schreibe` mit der Regel
//     `ueberschreiben`
//   - `Bogen` → `service.schreibeDatei` → `schreibeAtomar` (schreiben,
//     umbenennen)
//   - `MTGJSONSchreiben` → `service.Exporteur.MTGJSON` → `schreibeAtomar`
//   - `Vorlagen.Ausfuehren` → `os.WriteFile`
//   - `Sichern` → ein **Verzeichnis**, siehe [reserviereVerzeichnis]
//
// Keiner der ersten vier befragt sein Ziel vorher; die leere Reservierung wird
// ersetzt. **Der fünfte tat es**, und deshalb steht er woanders.
func reserviereOrt(ziel string) (aufraeumen func(), fehler error) {
	datei, err := os.OpenFile(ziel, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- gewähltes Ziel
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("%w: an %s liegt schon etwas — dieses "+
				"Werkzeug überschreibt nichts, wähle einen anderen Namen",
				domain.ErrConflict, ziel)
		}

		return nil, fmt.Errorf("%w: %s lässt sich nicht anlegen: %w",
			domain.ErrValidation, ziel, err)
	}

	// **Sofort schließen.** Was danach schreibt, öffnet die Datei selbst; ein
	// offener Griff daneben wäre auf Windows eine Sperre.
	//
	// **Und der Schließfehler wird gemeldet.** Auf einer eben angelegten,
	// leeren Datei sagt er etwas über das Dateisystem — ihn zu verschlucken
	// hieße, eine Reservierung zurückzugeben, die es vielleicht nicht gibt.
	if err := datei.Close(); err != nil {
		//nolint:errcheck // der Ort ist ohnehin nicht zu gebrauchen
		_ = os.Remove(ziel)

		return nil, fmt.Errorf("%w: %s lässt sich nicht anlegen: %w",
			domain.ErrValidation, ziel, err)
	}

	//nolint:errcheck // ein Aufräumer hat keinen Weg nach draußen; bleibt die
	// leere Reservierung liegen, weist der nächste Ruf sie als belegt ab
	return func() { _ = os.Remove(ziel) }, nil
}

// reserviereVerzeichnis reserviert ein Ziel, das ein **Verzeichnis** wird.
//
// # Warum es zwei Reservierungen gibt
//
// `Sichern` schreibt ein `.mse-set` als **Verzeichnis**
// ([0018](../../docs/entscheidungen/0018-ein-archiv-wird-nicht-zurueckgeschrieben.md)).
// Eine Reservierung als leere **Datei** hatte dort eine Folge, die kein Test
// der beiden Kommandos zeigte und die der Ablauf über alle Flächen sofort
// fand: `service.keinArchiv` sieht eine reguläre Datei am Ziel und weist mit
// „ein Archiv wird nicht zurückgeschrieben" ab — die Reservierung selbst
// verhinderte das Schreiben.
//
// `os.Mkdir` gibt dieselbe Zusage wie `O_EXCL`: es legt an oder scheitert mit
// `EEXIST`, und der Kern entscheidet es. Was danach schreibt, findet sein
// Verzeichnis vor; `os.MkdirAll` nimmt ein vorhandenes hin.
//
// **Aufgeräumt wird mit `os.RemoveAll`** und nicht mit `os.Remove`: ein
// gescheitertes Schreiben lässt halbe Dateien darin liegen, und `os.Remove`
// bekäme das Verzeichnis nicht weg. Rekursiv zu löschen ist hier vertretbar,
// weil das Verzeichnis in **diesem** Aufruf entstanden ist — lag vorher etwas
// da, hat die Reservierung schon abgewiesen.
func reserviereVerzeichnis(ziel string) (aufraeumen func(), fehler error) {
	if err := os.Mkdir(ziel, 0o700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("%w: an %s liegt schon etwas — dieses "+
				"Werkzeug überschreibt nichts, wähle einen anderen Namen",
				domain.ErrConflict, ziel)
		}

		return nil, fmt.Errorf("%w: %s lässt sich nicht anlegen: %w",
			domain.ErrValidation, ziel, err)
	}

	//nolint:errcheck // wie bei [reserviereOrt]: ein Aufräumer hat keinen Weg
	// nach draußen
	return func() { _ = os.RemoveAll(ziel) }, nil
}

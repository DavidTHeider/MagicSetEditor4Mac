// **Der Tausch der Bestandsdatei** (#940).
//
// Eine eigene Datei, weil er nichts von der Anwendung weiß: er bekommt zwei
// Pfade und gibt einen Rückweg. Die Frage „wann darf getauscht werden?"
// beantwortet [App.SicherungZurueckspielen], die Frage „wie?" steht hier.

package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Rueckspielsicherung ist der Name, unter dem der ersetzte Bestand liegen
// bleibt.
//
// **Er bleibt liegen, und er wird überschrieben.** Beides ist eine
// Entscheidung: wer zurückspielt und sich vertut, hat den vorigen Stand noch;
// wer zweimal hintereinander zurückspielt, hat den davor nicht mehr. Eine
// Kette von Ständen anzulegen hieße, das Datenverzeichnis mit Dateien zu
// füllen, die niemand je wieder ansieht — und die eigentliche Sicherung ist
// die, die jemand selbst schreibt.
const Rueckspielsicherung = "bestand-vor-rueckspielung.db"

// Eingehend ist die Endung, unter der eine Sicherung neben dem Bestand liegt,
// bevor sie ihn ersetzt.
//
// # Warum es diesen Zwischenschritt gibt
//
// Bis zum 2026-08-25 wurde die fremde Datei über ihren **Pfad** geprüft und
// danach über **denselben Pfad** ein zweites Mal geöffnet, um sie zu kopieren.
// Zwischen beiden Öffnungen kann ein anderer Prozess sie ersetzen — ein
// Abgleichdienst, ein Skript, ein Mensch. Die Nutzenden bestätigen Bestand A,
// übernommen wird Bestand B, und ist B ein **anderer gültiger** Bestand, wird
// er ohne erneute Rückfrage angenommen.
//
// **Der Kommentar an der Stelle nannte genau diese Bedrohung** — als Grund für
// die zweite Prüfung — und band Prüfung und Übernahme trotzdem nicht an
// dasselbe Objekt. Angestrichen vom Zweitblick über Phase S, Befund 7
// (#964).
//
// **Jetzt wird zuerst kopiert, dann die Kopie geprüft, dann die Kopie
// eingesetzt.** Was ankommt, ist damit dasselbe Objekt, das geprüft wurde.
//
// # Was es kostet
//
// Kurzzeitig liegt der Bestand doppelt da — die eingehende Kopie neben dem
// laufenden. Auf einem vollen Datenträger scheitert das Zurückspielen
// deshalb **früher** als bisher: beim Kopieren statt beim Tauschen. Das ist
// die bessere Stelle, denn dort ist noch nichts bewegt.
//
// **Und das Aufräumen bekommt einen Weg mehr.** Er läuft über `defer` und
// meldet sich nicht: eine liegengebliebene Kopie ist ein Ärgernis, kein
// Datenverlust, und sie wird beim nächsten Zurückspielen überschrieben.
const Eingehend = ".eingehend"

// nebendateien sind die Journaldateien, die SQLite neben den Bestand legt.
//
// **Sie gehören zum alten Stand.** Bleiben sie liegen, öffnet SQLite den
// zurückgespielten Bestand mit dem Journal des ersetzten — das ist der Fall,
// vor dem `docs/RELEASE.md` bei der Handarbeit warnt.
var nebendateien = []string{"-wal", "-shm"}

// bestandErsetzen legt die Sicherung an die Stelle des Bestands.
//
// Der zurückgegebene Rückweg stellt den vorigen Zustand her — **einschließlich
// des Falls, dass es gar keinen Bestand gab**. Er wird gerufen, wenn sich die
// zurückgespielte Datei nicht öffnen lässt; danach ist der Stand von vorher da,
// und zwar ganz.
//
// # Was er nicht tut
//
// **Er prüft die Sicherung nicht.** Das ist [storage.PruefeSicherung], und es
// geschieht davor — hier wäre es zu spät. Und er schließt keine Datenbank: wer
// tauscht, hat sie geschlossen, sonst schreibt eine offene Verbindung in eine
// Datei, die nicht mehr da ist, wo sie war.
// bestandBeiseite legt den Bestand weg, **ohne einen neuen einzusetzen** —
// Werkstück O4 von Phase 5.2
// (#1428).
//
// **Die erste Hälfte von [bestandErsetzen] und keine zweite Fassung davon**:
// dieselben zwei Schritte — beiseitelegen und das Journal wegräumen —, nur
// ohne das Umbenennen der eingehenden Datei. Wer sie zusammenlegte, bekäme
// einen Parameter, der manchmal leer ist, und einen Zweig darum.
func bestandBeiseite(bestand string) (func() error, error) {
	beiseite := filepath.Join(filepath.Dir(bestand), Rueckspielsicherung)

	gabEsEinen, err := verschieben(bestand, beiseite)
	if err != nil {
		return nil, err
	}

	zurueck := func() error { return zuruecklegen(bestand, beiseite, gabEsEinen) }

	if err := aufraeumen(bestand); err != nil {
		return nil, errors.Join(err, rueckweg(zurueck))
	}

	return zurueck, nil
}

func bestandErsetzen(eingehend, bestand string) (func() error, error) {
	// **Die eingehende Kopie zuerst ansehen, bevor irgendetwas bewegt wird.**
	// Ein fehlender Quellpfad darf den Bestand nicht kosten.
	if _, err := os.Stat(eingehend); err != nil {
		return nil, fmt.Errorf("die eingehende Sicherung %q lesen: %w", eingehend, err)
	}

	// **Die erste Hälfte ist [bestandBeiseite]** — beiseitelegen und das
	// Journal wegräumen, mit dem Rückweg in derselben Zeile wie der einzige
	// unumkehrbare Schritt.
	//
	// Hier stand bis zum Zweitblick über Phase S ein `beiseiteLegen`, das
	// beides tat. Scheiterte das Wegräumen, kam ein Fehler zurück und **kein
	// Rückweg**: die Hauptdatei lag unter dem Namen der Rückspielsicherung,
	// und `Starten` legte danach eine neue leere Datenbank an.
	zurueck, err := bestandBeiseite(bestand)
	if err != nil {
		return nil, err
	}

	// **Umbenannt und nicht kopiert.** Die eingehende Datei liegt im selben
	// Verzeichnis; das Umbenennen ist ein Schritt und braucht keinen Platz.
	// Ein zweites Kopieren wäre derselbe Bestand ein drittes Mal.
	if err := os.Rename(eingehend, bestand); err != nil {
		return nil, errors.Join(
			fmt.Errorf("die Sicherung nach %q einsetzen: %w", bestand, err), rueckweg(zurueck))
	}

	// **Und der neue Name wird festgeschrieben** (BK-06b): erst dann
	// übersteht das Einsetzen einen Absturz. Scheitert es, gilt das Einsetzen
	// als gescheitert, und der Rückweg legt den alten Bestand zurück.
	if err := verzeichnisFestschreiben(filepath.Dir(bestand), fuerDenWiederanlauf); err != nil {
		return nil, errors.Join(
			fmt.Errorf("die Sicherung nach %q festschreiben: %w", bestand, err), rueckweg(zurueck))
	}

	return zurueck, nil
}

// verschieben legt den vorhandenen Bestand beiseite — **und sonst nichts**.
//
// Der Rückgabewert sagt, ob es einen gab; nur dann hat der Rückweg etwas
// zurückzulegen.
func verschieben(bestand, beiseite string) (bool, error) {
	if _, err := os.Stat(bestand); errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	// **Umbenennen und nicht kopieren.** Ein Bestand kann groß sein, und eine
	// Kopie verdoppelte den Platzbedarf im ungünstigsten Augenblick. Beide
	// Pfade liegen im selben Verzeichnis, also im selben Dateisystem.
	if err := os.Rename(bestand, beiseite); err != nil {
		return false, fmt.Errorf("den bisherigen Bestand beiseitelegen: %w", err)
	}

	return true, nil
}

// zuruecklegen macht [verschieben] und die Kopie rückgängig.
//
// # Warum es die Journaldateien in Ruhe lässt
//
// Der Weg hin räumt sie weg, weil sie zum **ersetzten** Stand gehören. Der Weg
// zurück stellt genau diesen Stand wieder her — was dann noch danebenliegt,
// gehört zu ihm. Der erste Anlauf räumte auch hier auf, und das kostete die
// Zusage: liegt am Ort des Journals etwas, das sich nicht entfernen lässt,
// scheiterte das Aufräumen **erneut**, und die Hauptdatei kam nie zurück.
// Angestrichen vom Zweitblick über Phase S.
//
// **Die Reihenfolge ist die Zusage**: erst die zurückgespielte Datei weg, dann
// der alte Stand an seinen Platz. Was dazwischen scheitert, meldet sich.
func zuruecklegen(bestand, beiseite string, gabEsEinen bool) error {
	if err := os.Remove(bestand); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("die zurückgespielte Datei entfernen: %w", err)
	}

	if !gabEsEinen {
		return nil
	}

	if err := os.Rename(beiseite, bestand); err != nil {
		return fmt.Errorf("den bisherigen Bestand zurücklegen: %w", err)
	}

	return nil
}

// aufraeumen entfernt die Journaldateien neben einem Bestand.
func aufraeumen(bestand string) error {
	for _, endung := range nebendateien {
		if err := os.Remove(bestand + endung); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s entfernen: %w", bestand+endung, err)
		}
	}

	return nil
}

// eingehendeKopie legt eine fremde Sicherung neben den Bestand.
//
// **Sie ist der Gegenstand, der geprüft und eingesetzt wird** ([Eingehend]).
// Der Rückgabewert ist ihr Pfad; wer sie nicht mehr braucht, entfernt sie über
// [eingehendeWegraeumen].
func eingehendeKopie(sicherung, bestand string) (string, error) {
	// #nosec G304 -- der Pfad kommt aus dem Dateidialog der Nutzenden; genau
	// diese Datei sollen sie wählen dürfen.
	quelle, err := os.Open(sicherung)
	if err != nil {
		return "", fmt.Errorf("die Sicherung %q öffnen: %w", sicherung, err)
	}

	defer func() { _ = quelle.Close() }() //nolint:errcheck // gelesen ist gelesen

	ziel := bestand + Eingehend
	if fehler := kopieren(quelle, ziel); fehler != nil {
		return "", fehler
	}

	return ziel, nil
}

// eingehendeWegraeumen entfernt die Kopie, falls sie noch daliegt.
//
// **Ohne Fehlerweg.** Ist sie eingesetzt worden, gibt es sie nicht mehr; ist
// sie liegengeblieben, ist das ein Ärgernis und kein Datenverlust, und der
// nächste Lauf überschreibt sie.
func eingehendeWegraeumen(eingehend string) {
	_ = os.Remove(eingehend) //nolint:errcheck // siehe darüber: kein Datenverlust
}

// kopieren schreibt die Sicherung an den Ort des Bestands.
//
// **Über den einen Weg der Nachbardatei, und festgeschrieben bis zum
// Datenträger** (BK-06b, Vermerk 0174). Hier stand „es wird nicht `Sync`
// gerufen — eine Zusage, die diese Anwendung an keiner anderen Schreibstelle
// gibt". Seit BK-06b gibt sie sie nach der Rolle, und ein eingesetzter
// Bestand ist das, was der nächste Start liest: `F_FULLFSYNC`, Datei und
// Verzeichnis, Modus 0600.
func kopieren(quelle io.Reader, ziel string) error {
	if err := schreibeNach(ziel, func(nach io.Writer) error {
		_, err := io.Copy(nach, quelle)

		return err
	}, fuerDenWiederanlauf, true); err != nil {
		return fmt.Errorf("die Sicherung nach %q übertragen: %w", ziel, err)
	}

	return nil
}

package service

// **Was ein gescheitertes Einfügen zurücknimmt** —
// #1278,
// Befund 2.
//
// **Eine eigene Datei**, weil `kartenablage.go` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht, **was ein Einfügen tut**,
// hier, **was es hinterlässt, wenn es nicht durchkommt**.

import (
	"errors"
	"fmt"
)

// bilderZuruecknehmen entfernt die Dateien, die dieser Lauf umsonst geschrieben
// hat.
//
// **Umsonst heißt: keine angelegte Karte zeigt darauf.** Ein Teilerfolg lässt
// Karten stehen, und die Bilder dieser Karten müssen mit ihnen stehen bleiben —
// sonst nähme die Rücknahme dem Geglückten sein Bild weg (#1282, Befund 2).
//
// **Ein Fehlschlag wird gemeldet, nicht verschluckt.** Der erste Anlauf warf
// ihn weg, mit dem Argument, eine zweite Meldung verdecke die erste. Das war
// verkehrt herum: der Aufrufer verbindet beide mit `errors.Join`, und was
// bleibt, wenn niemand es sagt, ist genau der halbe Stand, gegen den es diese
// Rücknahme gibt (Befund 4).
func (k *Kartenablage) bilderZuruecknehmen(
	setID int64, umbenannt map[string]string, benutzt map[string]bool,
) error {
	if k.bilder == nil {
		return nil
	}

	return zuruecknehmen(umbenannt, benutzt, func(name string) error {
		return k.bilder.Wegnehmen(setID, name)
	})
}

// zuruecknehmen entfernt jeden unbenutzten Namen und sammelt die Fehlschläge.
//
// **Als eigene Funktion mit der Wegnahme als Parameter**, und das hat einen
// gemessenen Grund: über [Kartenablage.bilderZuruecknehmen] kann
// `Bildablage.Wegnehmen` gar nicht scheitern. Die Namen kommen aus
// `Bildablage.Bytes` und sind damit einfache Dateinamen; ein fehlendes Bild
// ist dort ausdrücklich kein Fehler, und ein Verzeichnis, in das eben noch
// geschrieben wurde, lässt auch löschen.
//
// Der Fehlerzweig wäre damit einer, den kein Aufruf betritt — und das ist eine
// Behauptung darüber, was im Ernstfall geschieht (§3). Hier ist er prüfbar.
func zuruecknehmen(
	umbenannt map[string]string, benutzt map[string]bool, weg func(string) error,
) error {
	var fehler []error

	for _, neu := range umbenannt {
		if benutzt[neu] {
			continue
		}

		if schief := weg(neu); schief != nil {
			fehler = append(fehler, fmt.Errorf(
				"das Bild %q ließ sich nicht zurücknehmen: %w", neu, schief))
		}
	}

	return errors.Join(fehler...)
}

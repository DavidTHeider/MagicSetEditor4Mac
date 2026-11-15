package storage

// **Vorblick, Sicherung und Wanderung unter einer Sperre** — M02
// (#2346),
// der erste Befund aus dem Abnahmelauf von M01.
//
// Ob gesichert wird, entschied ein Blick in die Buchführung ohne Sperre, und
// gesichert wurde vor der Schreibsperre von SQLite. Zwei Starter sicherten
// deshalb beide, und der zweite legte den eben gewanderten Bestand als zweite
// „Sicherung vor" ab — das Platzräumen hielte sie für die jüngste, und die
// echte von vorher würde räumbar.
//
// # Eine eigene Nachbardatei
//
// `bestand.db.wanderung`, **nicht** `bestand.db.sperre`: wer öffnet, gibt
// jene nach [Sperrfrist] auf und öffnet ohne sie — eine lange Sicherung wäre
// danach wieder ungeschützt. Diese wartet [Wanderungsfrist], und wer sie
// nicht bekommt, startet nicht, sondern hört „belegt" und versucht es gleich
// noch einmal (F05). Sie wird nie umbenannt; ein Rest nach einem Absturz ist
// eine Datei ohne Sperre.

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// migriereAusMitVorbereitung ist migriereAus mit einem Rückruf davor.
//
// **`basis`** ist die Grundlinie — `nil` heißt: ohne sie. Nur die
// eingebettete Quelle wandert über ihr; eine erfundene der Fehlerpfade hat mit
// ihr nichts zu tun.
func migriereAusMitVorbereitung(
	ctx context.Context, db *sql.DB, quelle fs.FS,
	pruefen Bestandspruefung, vorbereiten Vorbereitung, basis fs.FS,
) error {
	// **Alles unter der Sperre über die Wanderung** (M02,
	// `wanderungssperre.go`): ein zweiter Starter wartet, bis der erste
	// gesichert und gewandert hat, und sieht dann nichts mehr ausstehen.
	freigeben, sperrfehler := wanderungHalten(ctx, db)
	if sperrfehler != nil {
		return sperrfehler
	}

	defer freigeben()

	return wandernImGriff(ctx, db, quelle, pruefen, vorbereiten, basis)
}

// Wanderungsendung hängt an den Bestandspfad.
const Wanderungsendung = ".wanderung"

// Wanderungsfrist ist, wie lange ein Starter auf einen anderen wartet, der
// gerade sichert und wandert.
var Wanderungsfrist = 60 * time.Second

// wanderungHalten nimmt die Sperre über die Wanderung des Bestands hinter
// `db` **allein** — oder nichts, wo er keinen Ort hat: im Arbeitsspeicher.
//
// **Ohne Nachbardatei wandert niemand** — Zweitblick über Phase 2, Befund 4.
// Hier stand „wie beim Öffnen" ([beimOeffnen]), und ein Verzeichnis unter dem
// Namen der Sperre setzte die Koordination außer Kraft. Beim Öffnen schützt
// die Sperre nur das Lesen; eine Wanderung schreibt ohnehin, und wer neben
// dem Bestand nichts anlegen kann, kann ihn auch nicht sichern.
//
// # Mit Vorrang vor neuen Rufen
//
// **Erst die Schranke, dann die Sperre** (#2406): eine wartende Wanderung
// nimmt [Eintrittsendung] allein, und ein neuer Ruf kommt an ihr nicht
// vorbei ([HalteSchemafreigabe]). Laufende Rufe enden mit ihrer Antwort, und
// die Wanderung bekommt die Sperre, statt unter überlappenden Rufen ohne
// Lücke nach [Wanderungsfrist] mit „belegt" zu enden. **Die Frist gilt für
// beides zusammen.**
func wanderungHalten(ctx context.Context, db *sql.DB) (freigeben func(), err error) {
	griff, schranke, ortlos, err := wanderungsgriffe(ctx, db)
	if err != nil || ortlos {
		return func() {}, err
	}

	defer func() { _ = schranke.Schliessen() }() //nolint:errcheck // die Schranke hält nur bis zur Sperre

	ende := time.Now().Add(Wanderungsfrist)

	fehler := schranke.AlleinBis(ctx, Wanderungsfrist, Sperrtakt)
	if fehler == nil {
		fehler = griff.AlleinBis(ctx, time.Until(ende), Sperrtakt)
	}

	if fehler != nil {
		_ = griff.Schliessen() //nolint:errcheck // die Sperre kam nicht zustande

		return nil, domain.MitGrund(domain.ErrBelegt, fehler,
			"ein anderes Programm sichert und wandert den Bestand gerade — versuche es gleich noch einmal")
	}

	return func() { _ = griff.Schliessen() }, nil //nolint:errcheck // geschlossen ist geschlossen
}

// HalteSchemafreigabe fragt, ob dieser Prozess den Bestand noch kennt — und
// hält die Antwort, bis `frei` gerufen wird.
//
// # Warum eine Sperre und nicht nur die Frage
//
// **Zwischen Frage und Schreiben konnte eine Wanderung festschreiben** —
// Zweitblick über Phase 2, Befund 2: ein Dienst fragte vor dem Ruf, und eine
// neuere Fassung wanderte, bevor er schrieb. Die Freigabe hält deshalb die
// Sperre über die Wanderung **geteilt**; eine Wanderung braucht sie allein und
// wartet, bis der Ruf fertig ist — und der nächste sieht die neue Fassung.
//
// # Und wer den Stand nicht erfährt, bekommt keine
//
// **Ein Lesefehler ist kein Ja** (Befund 3): ein Prozess, der nicht weiß, ob
// er den Bestand kennt, schreibt nicht hinein. Im Arbeitsspeicher gibt es
// nichts zu sperren, gefragt wird trotzdem.
func HalteSchemafreigabe(ctx context.Context, db *sql.DB) (frei func(), err error) {
	griff, schranke, ortlos, err := wanderungsgriffe(ctx, db)
	if err != nil {
		return nil, err
	}

	freigeben := func() {}

	if !ortlos {
		// **Durch die Schranke und gleich wieder hinaus** (#2406): wartet
		// eine Wanderung, kommt kein neuer Ruf an ihr vorbei.
		ende := time.Now().Add(Wanderungsfrist)

		fehler := schranke.TeilenBis(ctx, Wanderungsfrist, Sperrtakt)
		if fehler == nil {
			fehler = griff.TeilenBis(ctx, time.Until(ende), Sperrtakt)
		}

		_ = schranke.Schliessen() //nolint:errcheck // die Schranke hält nur bis zur Sperre

		if fehler != nil {
			_ = griff.Schliessen() //nolint:errcheck // die Sperre kam nicht zustande

			return nil, domain.MitGrund(domain.ErrBelegt, fehler,
				"der Bestand wird gerade gesichert und gewandert — versuche es gleich noch einmal")
		}

		freigeben = func() { _ = griff.Schliessen() } //nolint:errcheck // geschlossen ist geschlossen
	}

	if fehler := PruefeWanderungsstand(ctx, db); fehler != nil {
		freigeben()

		return nil, fehler
	}

	return freigeben, nil
}

// Eintrittsendung ist die Schranke vor der Sperre über die Wanderung (#2406).
const Eintrittsendung = ".eintritt"

// wanderungsgriffe öffnet die Sperre über die Wanderung und ihre Schranke —
// `ortlos`, wo der Bestand keinen Ort hat.
func wanderungsgriffe(ctx context.Context, db *sql.DB) (griff, schranke *dateisperre.Griff, ortlos bool, err error) {
	datei, ortfehler := dateiVon(ctx, db)
	if ortfehler != nil {
		return nil, nil, false, ortfehler
	}

	if wanderungspfad(datei) == "" {
		return nil, nil, true, nil
	}

	griff, err = neben(wanderungspfad(datei))
	if err != nil {
		return nil, nil, false, err
	}

	if schranke, err = neben(datei + Eintrittsendung); err != nil {
		_ = griff.Schliessen() //nolint:errcheck // ohne Schranke keine Sperre

		return nil, nil, false, err
	}

	return griff, schranke, false, nil
}

// neben öffnet eine Nachbardatei der Wanderung.
func neben(pfad string) (*dateisperre.Griff, error) {
	griff, fehler := dateisperre.Oeffnen(pfad)
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrGesperrt, fehler,
			"neben dem Bestand lässt sich die Sperre über die Wanderung nicht anlegen — gib den Ort frei")
	}

	return griff, nil
}

// wanderungspfad gibt die Nachbardatei — leer, wo es keine geben kann, aus
// demselben Grund wie bei [sperrpfad].
func wanderungspfad(bestandspfad string) string {
	if sperrpfad(bestandspfad) == "" {
		return ""
	}

	return bestandspfad + Wanderungsendung
}

// Bestandsdatei gibt die Datei des Bestands hinter `db` — leer im
// Arbeitsspeicher. Für den Wärter vor der Wanderung (`internal/app`), der
// fragt, ob der eingerichtete Dienst **diese** Datei hält.
func Bestandsdatei(ctx context.Context, db *sql.DB) (string, error) { return dateiVon(ctx, db) }

// dateiVon gibt die Datei der Hauptdatenbank — leer im Arbeitsspeicher, und
// ein Fehler, wenn die Verbindung es nicht mehr sagt.
func dateiVon(ctx context.Context, db *sql.DB) (string, error) {
	var (
		nummer      int
		name, datei string
	)

	if fehler := db.QueryRowContext(ctx, "SELECT seq, name, file FROM pragma_database_list WHERE name = 'main'").
		Scan(&nummer, &name, &datei); fehler != nil {
		return "", fmt.Errorf("den Ort des Bestands erfragen: %w", einordnen(fehler))
	}

	return datei, nil
}

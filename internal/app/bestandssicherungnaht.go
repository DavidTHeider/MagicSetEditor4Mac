// **Sichern und Zurückspielen für die Flächen ohne Fenster** —
// #1169,
// Werkstück 2.
//
// # Warum es das erst jetzt gibt
//
// Bis zum 2026-08-28 lag der ganze Weg in `internal/app` und hing am Fenster:
// er schloss `App`s Verbindung, tauschte und öffnete sie wieder. Eine Fläche
// ohne Fenster hatte kein `App`, also keinen Weg hinein — der Grund im
// Katalog war ✋, „mindestens eine Bindung öffnet einen nativen Dateidialog".
//
// **Der Dialog war nie der Kern.** Er beantwortet die Frage *welche Datei*,
// und eine Kommandozeile beantwortet sie im Argument. Was wirklich fehlte, war
// die Trennung von Weg und Lebenszyklus — und die hat #1175 gebracht:
// `service.Bestandsweg` bekommt Orte und Lebenszyklus **hineingereicht**.
// Hier wird beides für einen Prozess ohne Fenster gebaut.
//
// # Was hier **nicht** steht
//
// **Kein MCP-Werkzeug zum Zurückspielen.** Der Maßstab ist die
// **Umkehrbarkeit** ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)):
// „Ein Werkzeug schreibt nur, was es nicht überschreibt." Ein Zurückspielen
// überschreibt den ganzen Bestand. Warum **nicht** §16 Regel 6, steht in
// `internal/mcp/bestandssicherung.go`.
//
// **Sichern schon.** Es liest und legt daneben; wer es ruft, verliert nichts.

package app

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Sicherungskennwortumgebung ist die Umgebungsvariable, aus der die
// Kommandozeile das Kennwort einer Sicherung liest — OPS-06.AK-02
// (#1723).
//
// **Nicht als Argument**: ein Argument steht in der Prozessliste und im
// Verlauf der Shell. Dieselbe Form wie bei `restic` und `borg`.
const Sicherungskennwortumgebung = "MSE_SICHERUNGSKENNWORT"

// MitSicherungswunsch legt Fläche, Kennwort und Klartextwunsch in den
// Kontext — für die Flächen, die `internal/service` nicht kennen. **Die
// Fläche ist Pflicht** (OPS01, #2347): sie steht im Vermerk der Sicherung,
// und ein Vermerk ohne sie besteht seine Prüfung nicht.
func MitSicherungswunsch(ctx context.Context, flaeche domain.Flaeche, kennwort string, klartext bool) context.Context {
	return service.MitSicherungswunsch(ctx, service.Sicherungswunsch{Flaeche: flaeche, Kennwort: kennwort, Klartext: klartext})
}

// Bestandsnaht ist der Sicherungsweg, wie eine Fläche ohne Fenster ihn sieht.
type Bestandsnaht struct {
	weg *service.Bestandsweg

	// bestandsort ist der Pfad der laufenden Bestandsdatei — er wird beim
	// Zurückspielen gebraucht, um sie neu zu öffnen.
	bestandsort string

	// offen hält die Verbindung, die der Aufrufer mitgebracht hat.
	offen **sql.DB
}

// BestandsnahtUeber baut sie über einer offenen Datenbank.
//
// # Der Lebenszyklus ist hier ein anderer als im Fenster
//
// Das Fenster hält eine Verbindung über seine ganze Laufzeit; ein Befehl hält
// sie für einen Aufruf. **Geschlossen und wieder geöffnet wird trotzdem** —
// nicht aus Symmetrie, sondern weil SQLite eine Datei nicht unter einer
// offenen Verbindung getauscht bekommen darf.
//
// **Der Zeiger auf den Zeiger ist kein Kunststück, sondern die Naht**: nach
// dem Tausch ist die alte Verbindung tot, und der Aufrufer muss die neue
// bekommen. Wer ihm die alte ließe, gäbe ihm einen Griff auf eine Datei, die
// es nicht mehr gibt.
func BestandsnahtUeber(db **sql.DB, bestandsort string) *Bestandsnaht {
	naht := &Bestandsnaht{bestandsort: bestandsort, offen: db}

	// **Dieselben Nähte wie im Fenster** und keine zweiten: `sicherungsnaht`
	// und `alleinnaht` stehen im Composition Root nebenan, und zwei Fassungen
	// davon liefen an dem Tag auseinander, an dem eine von beiden eine Frage
	// dazubekommt.
	naht.weg = service.NeuerBestandsweg(
		service.Sicherungsorte{Datenbank: bestandsort, Bilder: Bilderverzeichnis},
		sicherungsnaht,
		alleinnaht,
		bestandsnaht{db: *db},
		naht.schliessen,
		naht.oeffnen,
		Fassung,
	)

	// **Und die Verschlüsselung** (Werkstück 14 von Phase 4.12): dieselbe
	// Fähigkeit an jeder Fläche (§16). Die Verbindung wird **geholt**, weil
	// ein Tausch sie wechselt — der Grund steht bei `chiffrenaht`.
	naht.weg.MitBestandschiffre(chiffrenaht{holen: func() *sql.DB { return *db }})

	// **Und das Buch der Sicherungen** (OPS01, #2347) — über der Verbindung,
	// die gerade offen ist, aus demselben Grund wie die Chiffre.
	naht.weg.MitSicherungsbuch(vermerkbuch{holen: func() *sql.DB { return *db }})

	// **Und die eigenen Pakete** (OPS-06.AK-01) — derselbe Katalog wie im
	// Fenster, über der Verbindung des Aufrufers.
	naht.weg.MitPaketen(storage.NeuerPaketkatalog(*db), func() (string, error) { return paketablage(false) })

	// **Und die drei Nähte des Rücksprungvergleichs** (O7, #1431) —
	// dieselben wie im Fenster und keine zweiten. Die Verbindung wird
	// **geholt**, weil ein Tausch sie wechselt: die Epoche gehört in den
	// zurückgespielten Bestand, nicht in den ersetzten.
	holen := func() *sql.DB { return *db }
	naht.weg.MitRuecksprungschau(
		bestandsschaunaht,
		eigenstandnaht{holen: holen}.stand,
		rueckkehrnaht{holen: holen},
	)

	return naht
}

// Verschluesselungsstand sagt, ob der Bestand verschlüsselt auf der Platte liegt.
func (n *Bestandsnaht) Verschluesselungsstand() (bool, error) {
	return n.weg.Verschluesselungsstand()
}

// VerschluesselungUmschalten schaltet sie an oder aus und gibt den neuen Stand.
func (n *Bestandsnaht) VerschluesselungUmschalten(
	ctx context.Context, an bool,
) (bool, error) {
	return umschalten(ctx, n.weg, an)
}

// schliessen gibt die Verbindung des Aufrufers zurück.
//
// **Der Zeiger bleibt stehen und wird nicht auf `nil` gesetzt.**
//
// Der erste Anlauf tat es, und das war ein Fehler mit Folgen: `mitBestand` in
// `cmd/mse` gibt die Verbindung im `defer` zurück, und `(*sql.DB)(nil).Close()`
// stürzt ab. Ein Zurückspielen, dessen Wiederöffnen scheitert, hätte den
// Befehl also nicht mit einer Meldung beendet, sondern mit einem Absturz —
// genau auf dem Weg, der die Meldung am nötigsten hat.
//
// **Ein zweites `Close` auf dieselbe Verbindung ist harmlos**: `database/sql`
// gibt dann `nil`. Was bleibt, ist ein geschlossener Griff, und jeder Aufruf
// darauf meldet es.
func (n *Bestandsnaht) schliessen() error { return (*n.offen).Close() }

// oeffnen holt eine neue Verbindung auf dieselbe Datei.
//
// **Migriert wird dabei**, wie an jedem anderen Root auch: eine
// zurückgespielte Sicherung kann aus einer älteren Fassung stammen, und ihr
// Schema muss vor dem ersten Lesen stimmen.
func (n *Bestandsnaht) oeffnen(ctx context.Context) error {
	db, fehler := storage.Oeffnen(ctx, n.bestandsort)
	if fehler != nil {
		return fmt.Errorf("den Bestand wieder öffnen: %w", fehler)
	}

	if fehler := migriere(ctx, db,
		SicherungVorMigration(db, filepath.Dir(n.bestandsort), Fassung)); fehler != nil {
		_ = db.Close() //nolint:errcheck // der Fehler darunter ist der, der zählt

		return fmt.Errorf("den zurückgespielten Bestand migrieren: %w", fehler)
	}

	*n.offen = db

	return nil
}

// Sichern schreibt Bestand **und** Bilder als eine `.msesicherung`.
func (n *Bestandsnaht) Sichern(ctx context.Context, ziel string) error {
	return n.weg.Sichern(ctx, ziel)
}

// Beschauen sagt, was in einer Sicherung liegt — ohne sie zu übernehmen.
func (n *Bestandsnaht) Beschauen(ctx context.Context, pfad string) (
	service.Sicherungsschau, error,
) {
	return n.weg.Beschauen(ctx, pfad)
}

// Weg gibt den Bestandsweg für die Fassade — Werkstück O7 von Phase 5.2
// (#1431).
//
// **Damit es keine dritte Wurzel braucht.** Der Rücksprung fragt dasselbe
// Objekt wie das Sichern; eine zweite Stelle, die einen `Bestandsweg` baut,
// liefe an dem Tag auseinander, an dem eine von beiden eine Naht dazubekommt —
// und genau das hält `nahtverdrahtung_test.go`.
func (n *Bestandsnaht) Weg() *service.Bestandsweg { return n.weg }

// Zurueckspielen ersetzt den Bestand und gibt, wohin der alte kam.
func (n *Bestandsnaht) Zurueckspielen(ctx context.Context, pfad string) (string, error) {
	return n.weg.Zurueckspielen(ctx, pfad)
}

// sichernUeber ist `storage.Sichern` über `db`, und ein Fehler des
// Datenträgers heißt „voll", wenn der Platz für die Sicherung nicht reicht
// (`sicherungsplatz.go`, M02, #2346) — für die Sicherung vor der Wanderung und
// die auf Wunsch.
func sichernUeber(db *sql.DB) func(context.Context, string) error {
	return mitPlatzfrage(
		func(ctx context.Context, ziel string) error { return storage.Sichern(ctx, db, ziel) },
		func(ctx context.Context) int64 { return storage.Sicherungsbedarf(ctx, db) },
		freierPlatz)
}

// **Das Buch der Sicherungen über der offenen Verbindung** — OPS01
// (#2347).
// **Geholt und nicht gehalten**: ein Zurückspielen tauscht die Verbindung,
// und ein Buch über der alten schriebe in eine Datei, die es nicht mehr gibt
// — derselbe Grund wie bei `chiffrenaht`. Es steht hier, weil diese Datei
// `storage` kennen darf (Composition Root, §2).
//
// vermerkbuch erfüllt `service.Sicherungsbuch` über der Verbindung, die
// `holen` gerade gibt.
type vermerkbuch struct{ holen func() *sql.DB }

func (v vermerkbuch) Vermerken(ctx context.Context, vermerk domain.Sicherungsvermerk) error {
	return storage.NeueSicherungsvermerke(v.holen()).Vermerken(ctx, vermerk)
}

func (v vermerkbuch) Letzter(ctx context.Context) (domain.Sicherungsvermerk, bool, error) {
	return storage.NeueSicherungsvermerke(v.holen()).Letzter(ctx)
}

func (v vermerkbuch) Vermerke(ctx context.Context, anlass domain.Sicherungsanlass) ([]domain.Sicherungsvermerk, error) {
	return storage.NeueSicherungsvermerke(v.holen()).Vermerke(ctx, anlass)
}

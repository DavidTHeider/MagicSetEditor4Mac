package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Migrationen reisen in der Binärdatei mit.
//
// Eingebettet und nicht aus dem Dateisystem gelesen: eine ausgelieferte
// Anwendung hat kein `internal/storage/migrationen/` neben sich, und eine
// Migration, die zur Laufzeit fehlt, zeigt sich erst beim ersten Start bei
// Nutzenden.
//
//go:embed migrationen
var migrationen embed.FS

// pruefsumme bildet den Inhalt einer Migration ab.
//
// SHA-256 und nicht etwa die Länge: eine Änderung, die die Länge behält — ein
// vertauschtes Wort, eine geänderte Zahl —, ist genau die, die sonst unbemerkt
// bliebe.
func pruefsumme(inhalt []byte) string {
	summe := sha256.Sum256(inhalt)

	return hex.EncodeToString(summe[:])
}

// migrationsname verlangt vier Ziffern, Unterstrich, Bezeichnung, `.sql`.
//
// Die Reihenfolge entsteht aus der lexikographischen Sortierung der Namen. Das
// trägt genau so lange, wie alle Namen dieselbe Form haben: „10_x.sql" käme vor
// „9_x.sql", und niemand sähe es, weil beide Namen für sich plausibel sind.
var migrationsname = regexp.MustCompile(`^\d{4}_[a-z0-9]+(?:[_-][a-z0-9]+)*\.sql$`)

// ErrMigrationen meldet einen Fehler im Bestand der Migrationen selbst.
var ErrMigrationen = errors.New("der Bestand der Migrationen ist nicht in Ordnung")

// Migriere wendet alle noch nicht angewendeten Migrationen an.
//
// Idempotent, und das ist keine Nebensache: der Normalfall ist nicht die
// frische Datenbank, sondern die vorhandene. Jeder Start migriert.
func Migriere(ctx context.Context, db *sql.DB, vorbereiten Vorbereitung) error {
	return migriereAusMitVorbereitung(ctx, db, migrationen, PruefeBestand, vorbereiten, grundlinie)
}

// Vorbereitung läuft **einmal**, bevor eine ausstehende Migration beginnt.
//
// # Wofür
//
// Die Sicherung vor der Migration
// (#939).
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)
// nennt sie „notwendig" und macht sie zur Bedingung, unter der der Satz
// „rückwärts, nie" vertretbar ist.
//
// # Warum ein Rückruf und keine Sicherung an dieser Stelle
//
// `storage` weiß, **ob** etwas aussteht. **Wohin** eine Sicherung gehört und
// wie sie heißt, weiß die Anwendung — sie kennt ihre Fassung und ihr
// Verzeichnis. Eine Sicherung hier hieße, `storage` eine Ablageordnung
// beizubringen, die es sonst nirgends braucht.
//
// # Wann sie **nicht** läuft
//
// Wenn nichts aussteht. Ein Start, der nur öffnet, schreibt keine Kopie —
// sonst läge nach einer Woche für jeden Tag eine da.
//
// **`nil` heißt: nichts vorbereiten.** Die Fehlerpfadtests brauchen sie nicht,
// und ein Rückruf, den sie erfinden müssten, prüfte die Vorrichtung.
type Vorbereitung func(ctx context.Context) error

// migriereAus ist Migriere mit austauschbarer Quelle.
//
// Die Trennung dient der Prüfbarkeit: die Fehlerpfade — eine unlesbare Quelle,
// eine Migration mit fehlerhaftem SQL — lassen sich mit der eingebetteten
// Quelle nicht herbeiführen, weil sie richtig ist. Ein Fehlerpfad, den kein
// Test je betritt, ist eine Behauptung darüber, was im Ernstfall geschieht.
func migriereAus(ctx context.Context, db *sql.DB, quelle fs.FS, pruefen Bestandspruefung) error {
	return migriereAusMitVorbereitung(ctx, db, quelle, pruefen, nil, nil)
}

// wandernImGriff ist [migriereAusMitVorbereitung] unter der Sperre — und der
// Weg einer Fassung vor M02, die sie nicht kennt.
func wandernImGriff(
	ctx context.Context, db *sql.DB, quelle fs.FS,
	pruefen Bestandspruefung, vorbereiten Vorbereitung, basis fs.FS,
) error {
	namen, err := migrationsnamen(quelle)
	if err != nil {
		return err
	}

	// **Mit der Grundlinie sind ihre Wanderungen Teil der Folge** (Vermerk
	// 0189): fort als Dateien, da mit Name und Prüfsumme. `alle` ist die
	// Folge, gegen die gezählt und abgewiesen wird; `namen` sind die
	// Wanderungen danach, die noch als Datei laufen.
	var (
		umfang []string
		summen map[string]string
	)

	if basis != nil {
		if umfang, summen, err = grundlinienumfang(basis); err != nil {
			return err
		}
	}

	alle := append(slices.Clone(umfang), namen...)

	if fehler := luecken(alle); fehler != nil {
		return fehler
	}

	// **Die Vorbereitung läuft vor *jeder* Schemaänderung** und nur bei Bedarf.
	//
	// # Warum vor `buchfuehrungAnlegen` und nicht danach
	//
	// `buchfuehrungAnlegen` ist selbst eine Schemaänderung — bei einer
	// Datenbank aus der Zeit vor der Prüfsumme trägt es per `ALTER TABLE` eine
	// Spalte nach. Danach gesichert, hielte die Sicherung den Stand **nach**
	// diesem Zug, während `README.md` und `docs/ADMINISTRATION.md` zusagen, sie
	// halte den Stand davor. Es ginge um Buchführungsmetadaten und nicht um
	// Nutzdaten — aber eine Zusage, die zu 99 % gilt, ist eine, die niemand
	// nachprüfen kann.
	//
	// # Woran „bei Bedarf" jetzt hängt
	//
	// An einem **lesenden** Blick in die Buchführung: `verbuchteNamen` gibt
	// eine leere Menge, wenn es die Tabelle nicht gibt, und rührt nichts an.
	// Ein frischer Bestand sichert damit weiterhin nicht — `gibtEsWasZuSichern`
	// sagt bei leerer Buchführung nein.
	//
	// **`unbekannteAbweisen` läuft erst danach.** Ein Bestand aus einer
	// neueren Fassung wird also gesichert, bevor er abgewiesen wird; das ist
	// eine Kopie, die niemand braucht, und der umgekehrte Fehler wäre der
	// teure.
	verbucht, leer, lesefehler := vorblick(ctx, db)
	if lesefehler != nil {
		return lesefehler
	}

	// **Ein Bestand vor der Grundlinie wird abgewiesen, bevor irgendetwas
	// geschieht** — auch vor `buchfuehrungAnlegen`, das einer Buchführung aus
	// der Zeit vor der Prüfsumme die Spalte nachtrüge, und ebenso Tabellen
	// ohne Buchführung (Zweitblick über die Folgen vor 0.8.0, Befund 1). Er
	// bleibt, wie er ist, und 0.7.0 lässt ihn aufsteigen (Vermerk 0189). Die
	// Prüfsummen vergleicht der zweite Blick unter der Sperre.
	if basis != nil {
		if fehler := grundlinieVorDerSperre(umfang, summen, verbucht, leer); fehler != nil {
			return fehler
		}
	}

	if vorbereiten != nil && gibtEsWasZuSichern(namen, verbucht) {
		if fehler := vorbereiten(ctx); fehler != nil {
			return fehler
		}
	}

	if fehler := buchfuehrungAnlegen(ctx, db); fehler != nil {
		return fehler
	}

	// **Geprüft wird unter der Sperre**, in [anwendenAlle]: was dort steht,
	// ist der Stand, auf dem gewandert wird.
	return anwendenAlle(ctx, db, quelle, namen, pruefen, basis, umfang, summen)
}

// Bestandspruefung sagt, ob der migrierte Bestand brauchbar ist.
//
// **Ein Parameter und keine feste Zuweisung** (#938): sie prüft eine Eigenschaft
// des **echten** Schemas — `bestand.angelegt_am`, angelegt von der ersten
// Migration. Die Fehlerpfadtests fahren erfundene Quellen, die kein solches
// Schema erzeugen; eine fest verdrahtete Prüfung machte sie alle rot, ohne über
// den Betrieb etwas auszusagen.
//
// **`nil` heißt: nicht prüfen**, und das nennt den Unterschied, statt ihn zu
// erraten.
type Bestandspruefung func(ctx context.Context, quelle Zeilenquelle) error

// anwendenAlle führt die ausstehenden Migrationen in **einer** Transaktion.
//
// # Warum die Folge und nicht jede für sich
//
// Hier stand bis zum 2026-08-23 eine Schleife über `anwenden`, und jede
// Migration hatte ihre eigene Transaktion. Scheiterte die dritte von fünf,
// waren die ersten beiden festgeschrieben — der Bestand stand auf einem Stand,
// den es als Fassung **nie gab**
// ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md),
// #938).
//
// **Nachgesehen, dass es geht** — damals an zwölf Migrationen: keine enthielt
// `PRAGMA` oder `VACUUM`, und beides wäre in einer Transaktion wirkungslos
// oder verboten. SQLite führt DDL transaktional. Das `PRAGMA foreign_keys` in
// `0056_scrydexkennung.sql` ist seither ein solcher Fall: wirkungslos, wie
// `0070_ausfuehrung.sql` für sich selbst festhält.
//
// # Was es kostet
//
// **Den Fortschritt.** Vorher blieb erhalten, was schon lief, und der nächste
// Start setzte dort auf; jetzt beginnt er von vorn. Das ist der bessere Handel:
// „von vorn" ist ein Zustand, den es gibt, „zur Hälfte" nicht.
//
// # Was sie nicht abdeckt
//
// **Den Weg zurück auf eine ältere Fassung.** Dessen Bestand ist vollständig
// migriert und wird von der alten Fassung abgewiesen; dagegen hilft keine
// Transaktion, sondern nur die Sicherung
// (#939).
//
// # Die Buchführung wird unter der Sperre gelesen
//
// **Und nicht davor.** Legen zwei Prozesse denselben neuen Bestand an — das
// Fenster und der Dienst beim ersten Start —, lesen beide die leere
// Buchführung, bevor einer die Schreibsperre hat (`_txlock=immediate`). Der
// zweite wartet, bis der erste fertig ist; arbeitete er dann mit dem Stand
// von vorher, wandte er bis 0.7.0 `0001` noch einmal an und scheiterte an
// einer Tabelle, die es schon gab — und hielte seit der Grundlinie den
// frischen Bestand für einen fremden (`grundlinienwettlauf_intern_test.go`).
func anwendenAlle(
	ctx context.Context, db *sql.DB, quelle fs.FS, namen []string,
	pruefen Bestandspruefung, basis fs.FS, umfang []string, summen map[string]string,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "Migration beginnen")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach einem Commit ist sie wirkungslos

	bereits, err := verbucht(ctx, tx, `SELECT name, pruefsumme FROM migrationen ORDER BY name`)
	if err != nil {
		return err
	}

	// **Und geprüft, was dort steht** (Zweitblick über die Folgen vor 0.8.0,
	// Befund 3): hat eine neuere Fassung inzwischen gewandert, wird abgewiesen
	// — sonst kehrte diese still zurück, als stünde der Bestand auf ihrem
	// Stand. Was verbucht ist, muss es noch geben, und was die Grundlinie
	// enthält, muss er mit ihrer Prüfsumme führen.
	if fehler := unbekannteAbweisen(append(slices.Clone(umfang), namen...), bereits); fehler != nil {
		// **Als Abweisung des Bestands, der hier geöffnet wird** — sie
		// unterscheidet sich so von der einer Sicherung (Nachlauf über
		// Phase 2 von 0.9.0, Befund 2).
		return &domain.BestandNeuer{Grund: fehler}
	}

	if basis != nil && len(bereits) > 0 {
		if fehler := grundlinieGefuehrt(umfang, summen, bereits); fehler != nil {
			return fehler
		}
	}

	if basis != nil && len(bereits) == 0 {
		gebucht, fehler := grundlinieWennLeer(ctx, tx, basis, umfang, summen)
		if fehler != nil {
			return fehler
		}

		bereits = gebucht
	}

	for _, name := range namen {
		summe, schonGelaufen := bereits[name]
		if schonGelaufen {
			if fehler := summeVergleichen(quelle, name, summe); fehler != nil {
				return fehler
			}

			continue
		}

		if fehler := anwenden(ctx, tx, quelle, name); fehler != nil {
			return fehler
		}
	}

	// **Und die Gegenprobe innerhalb der Transaktion** (#938). Sie lief bis
	// zum 2026-08-23 **danach**: wies sie ab, war die Migration bereits
	// festgeschrieben, und der Bestand ließ sich weder benutzen noch mit der
	// alten Fassung öffnen. Hier ist die Abweisung eine Rücknahme.
	if pruefen != nil {
		if fehler := pruefen(ctx, tx); fehler != nil {
			return fehler
		}
	}

	if vorDemFestschreiben != nil {
		vorDemFestschreiben()
	}

	if fehler := tx.Commit(); fehler != nil {
		return uebersetze(fehler, "Migrationen abschließen")
	}

	return nil
}

// AngewendeteMigrationen gibt die verbuchten Migrationen in ihrer Reihenfolge.
//
// Setzt auf `angewendeteMitSummen` auf statt eine zweite Abfrage zu führen: zwei
// Abfragen auf dieselbe Tabelle sind zwei Fehlerpfade, und der zweite wäre nur
// mit einer Attrappe zu erreichen, die nachstellt, was SQLite ohnehin tut.
func AngewendeteMigrationen(ctx context.Context, db *sql.DB) ([]string, error) {
	summen, err := angewendeteMitSummen(ctx, db)
	if err != nil {
		return nil, err
	}

	namen := make([]string, 0, len(summen))
	for name := range summen {
		namen = append(namen, name)
	}

	sort.Strings(namen)

	return namen, nil
}

// migrationsnamen gibt die eingebetteten Migrationen in aufsteigender Ordnung.
func migrationsnamen(quelle fs.FS) ([]string, error) {
	eintraege, err := fs.ReadDir(quelle, "migrationen")
	if err != nil {
		return nil, uebersetze(err, "Migrationen auflisten")
	}

	namen := make([]string, 0, len(eintraege))
	for _, eintrag := range eintraege {
		name := eintrag.Name()

		// **Nur Wanderungen** — neben ihnen liegt `LIESMICH.md`, seit die bis
		// zur Grundlinie fort sind (Vermerk 0189).
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		if !migrationsname.MatchString(name) {
			return nil, fmt.Errorf("%w: %q entspricht nicht der Form NNNN_bezeichnung.sql",
				ErrMigrationen, name)
		}

		namen = append(namen, name)
	}

	sort.Strings(namen)

	return namen, nil
}

// gibtEsWasZuSichern sagt, ob eine Migration aussteht **und** schon ein Stand
// da ist.
//
// # Warum die zweite Bedingung
//
// Beim allerersten Start stehen **alle** Migrationen aus, und es gibt nichts zu
// sichern: die Datenbank ist eben erst entstanden. Eine Kopie davon wäre eine
// Datei, die niemand je braucht — und sie trüge einen Namen, der einen Rückweg
// verspricht, den es nicht gibt
// (#939).
//
// **Der Stand ist die Buchführung**, nicht die Datei: eine leere Datenbank mit
// angelegter Buchführung gibt es nur zwischen zwei Anweisungen desselben Laufs.
func gibtEsWasZuSichern(namen []string, bereits map[string]string) bool {
	if len(bereits) == 0 {
		return false
	}

	for _, name := range namen {
		if _, schonGelaufen := bereits[name]; !schonGelaufen {
			return true
		}
	}

	return false
}

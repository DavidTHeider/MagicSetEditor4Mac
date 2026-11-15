// Package storage hält die SQLite-Repositories und die Migrationen
// (AGENTS.md §2).
//
// Es kennt `internal/domain` und sonst nichts aus dem Projekt. Die
// Repository-**Schnittstellen** werden in `internal/service` definiert; dass
// dieses Paket sie erfüllt, ohne `service` zu importieren, liegt an Gos
// struktureller Typisierung und ist Absicht — `internal/schichten` kann diese
// Zuordnung deshalb nicht prüfen, und das steht dort auch so.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ImArbeitsspeicher ist die Datenquelle für eine Datenbank ohne Datei.
//
// `cache=shared` ist nötig, weil `database/sql` einen Pool führt: ohne sie
// bekäme jede Verbindung ihre **eigene** leere Datenbank, und ein Test würde
// die Tabellen nicht finden, die er gerade angelegt hat.
const ImArbeitsspeicher = "file::memory:?cache=shared"

// Zeitform ist die eine Schreibweise für Zeitstempel in dieser Datenbank.
//
// Drei Nachkommastellen, nicht neun. Der Grund steht in der ersten Migration:
// SQLite schreibt mit `strftime('%f')` drei, Go mit RFC3339Nano bis zu neun —
// und zwei Schreibwege, die verschiedene Formen erzeugen, laufen so lange
// nebeneinander her, bis eine Prüfung eine davon verlangt.
const Zeitform = "2006-01-02T15:04:05.000Z"

// Zeitstempel bringt einen Zeitpunkt in die Form dieser Datenbank.
func Zeitstempel(zeitpunkt time.Time) string {
	return zeitpunkt.UTC().Format(Zeitform)
}

// AlsURI bringt einen Dateipfad in die Form, die SQLite als URI liest.
//
// # Warum das nötig ist
//
// `"file:" + pfad` ist **keine** gültige URI, sobald der Pfad `?`, `#` oder
// `%` enthält — alle drei sind auf macOS erlaubte Zeichen in einem
// Dateinamen. SQLite liest den Teil vor `?` als Pfad und den Rest als
// Abfrage, den Teil nach `#` als Fragment: aus `vorher?eins.db` wird die
// Datei `vorher`, und angehängte Parameter wie `mode=ro` gehen verloren oder
// treffen die falsche Datei. Statt zu scheitern **legt SQLite sie an**.
//
// **Beleg:** „Note that … the '?' character in the filename must be converted
// into %3f and the '#' character into %23" —
// https://www.sqlite.org/uri.html#the_uri_path, nachgelesen am 2026-08-23.
//
// # Die Reihenfolge zählt
//
// `%` zuerst: wer erst `?` in `%3f` verwandelte und danach jedes `%`
// verdoppelte, machte daraus `%253f`.
//
// # Was sie nicht tut
//
// **Sie hängt keine Parameter an.** Das tut der Aufrufer — und er weiß, ob er
// ein `?` oder ein `&` als Trenner braucht.
func AlsURI(pfad string) string {
	ersetzt := strings.ReplaceAll(pfad, "%", "%25")
	ersetzt = strings.ReplaceAll(ersetzt, "?", "%3f")
	ersetzt = strings.ReplaceAll(ersetzt, "#", "%23")

	return "file:" + ersetzt
}

// Oeffnen öffnet eine Datenbank und stellt die Pragmas ein.
//
// **`quelle` ist ein Dateipfad**, keine URI — außer [ImArbeitsspeicher] und
// dem, was schon mit `file:` beginnt. Die Umwandlung steht in [AlsURI] und
// geschieht hier, an **einer** Stelle: ein Pfad mit `?` oder `#` erreichte
// SQLite sonst zerschnitten, und `mitPragmas` wählte obendrein den falschen
// Trenner. Angestrichen vom Zweitblick über Phase S.
func Oeffnen(ctx context.Context, quelle string) (*sql.DB, error) {
	// **Die geteilte Sperre wird beim Öffnen genommen** (#1163, Befund Z6-02).
	//
	// **Hier und nicht bei den Aufrufern**: es sind vier — Fenster,
	// Kommandozeile, MCP-Server, Dienst —, und eine Sperre, die einer von
	// ihnen vergisst, ist keine. Was sie zusagt und was nicht, steht in
	// `sperre.go`.
	//
	// **Gehalten, solange geöffnet wird — nicht länger.** Warum das genügt
	// und warum es nicht länger sein darf, steht in `sperre.go`.
	return oeffnenMit(ctx, quelle, nachgeschlagen(ctx, quelle), schreibend)
}

// OeffnenMitSchluessel macht einen Bestand mit einem ausdrücklich genannten
// Schlüssel auf.
//
// **Für die Umschaltung und für Läufe** — im Betrieb weiß [Oeffnen] selbst,
// wo der Schlüssel liegt, und das ist die Zusage: vier Flächen öffnen den
// Bestand, und ein Schlüssel, den eine davon vergisst, ist keiner.
//
// Ein leerer Schlüssel heißt „ohne Hülle" — dann ist es dasselbe wie [Oeffnen]
// auf einem Klartextbestand.
func OeffnenMitSchluessel(
	ctx context.Context, quelle, schluessel string,
) (*sql.DB, error) {
	return oeffnenMit(ctx, quelle, func() (string, error) { return schluessel, nil }, schreibend)
}

// OeffnenNurLesend macht den Bestand auf, ohne dass jemand hineinschreiben
// kann — Werkstück 16 von Phase 4.12
// (#1351).
//
// # Der Schutz sitzt im Treiber, und das ist der ganze Punkt
//
// **`query_only` gilt für die Verbindung, nicht für eine Abfrage.** Damit gibt
// es keine Stelle, an der jemand ihn vergessen kann — und genau deshalb wird
// er hier genommen und nicht an den Bindungen geprüft. Vermerk
// [0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md)
// sagt es: eine Zusage, die an 116 Bindungen einzeln hängt — **am 2026-09-06
// gezählt** —, bricht beim ersten vergessenen Ruf still.
//
// **Gemessen am 2026-09-06** mit der Produktivadresse: `INSERT`, `DELETE`,
// `CREATE TABLE` und schon `BeginTx` mit `_txlock=immediate` geben
// `SQLITE_READONLY`; Lesen glückt. Die Gegenprobe steht daneben — dieselbe
// Datei gewöhnlich geöffnet nimmt den Schreibvorgang.
//
// # Was er nicht ist
//
// **Kein Ersatz für die Freigabeprüfung.** Wer gar nicht hinein darf, wird
// vorher abgewiesen; das hier gilt für den, der hinein darf und nichts ändern
// soll.
//
// **Und keine Aussage über die Datei.** Sie bleibt beschreibbar; es ist
// **diese Verbindung**, die nichts annimmt. Ein zweiter Prozess mit einer
// gewöhnlichen Verbindung schreibt weiter — das ist richtig so, denn die
// Rolle gilt einem Menschen und nicht einer Datei.
func OeffnenNurLesend(ctx context.Context, quelle string) (*sql.DB, error) {
	return oeffnenMit(ctx, quelle, nachgeschlagen(ctx, quelle), nurLesend)
}

// nachgeschlagen sucht den Schlüssel eines Bestands, wenn er gebraucht wird.
func nachgeschlagen(ctx context.Context, quelle string) func() (string, error) {
	return func() (string, error) { return schluesselFuer(ctx, quelle) }
}

// leseart sagt, ob eine Verbindung schreiben darf.
type leseart bool

const (
	schreibend leseart = false
	nurLesend  leseart = true
)

// oeffnenMit ist der gemeinsame Rumpf von [Oeffnen] und [OeffnenMitSchluessel].
func oeffnenMit(
	ctx context.Context, quelle string, schluesselfrage func() (string, error), art leseart,
) (*sql.DB, error) {
	defer beimOeffnen(ctx, quelle)()

	// **Der Schlüssel unter der Sperre.** Ob eine Datei verschlüsselt ist,
	// sagt ihr Kopf — und hält ein anderer die Sperre, weil er sie gerade
	// anlegt, steht dort womöglich noch nichts. Ein Kopf aus Nullen ist nicht
	// `SQLite format 3`, und eine frische Datei galt als verschlüsselt
	// (`grundlinienvorblick_intern_test.go`). Hier stand bis dahin „der
	// Schlüssel vor der Sperre", damit keine Sperre wieder fallen muss; sie
	// fällt mit dem `defer` darüber.
	schluessel, err := schluesselfrage()
	if err != nil {
		return nil, domain.Zugriffsmangel(err)
	}

	adresse := mitHuelle(mitPragmas(alsQuelle(quelle)), schluessel)
	if art == nurLesend {
		adresse += "&_pragma=query_only(1)"
	} else {
		// **Ein neuer Bestand gehört seinem Eigner** (#2508).
		adresse = eigenerBestand(quelle, adresse)
	}

	db, err := oeffner(adresse, schluessel)
	if err != nil {
		return nil, fmt.Errorf("Datenbank %q öffnen: %w", quelle, err)
	}

	// **Eine Verbindung je Datenbank — für jede, nicht nur für die im
	// Arbeitsspeicher.**
	//
	// Dort war sie immer die Bedingung: nur die erste Verbindung hält den
	// Inhalt, und schließt der Vorrat sie, ist alles fort. Für eine Datei
	// stand hier bis zum 2026-09-09 ein Vorrat — und der hat gekostet, was
	// #1485
	// beschreibt: `PRAGMA data_version` zählt je **Verbindung**, und jedes
	// eigene Schreiben des Fensters lief auf einer anderen als die Frage
	// danach. Das Fenster hielt sich selbst für einen fremden Prozess und lud
	// nach jedem Schreiben die Seite neu — drei Rasterläufe je Feld.
	//
	// **Was die eine Verbindung kostet**: Lesen wartet auf Schreiben, statt
	// daneben zu laufen. SQLite hat ohnehin nur einen Schreiber; die Rufe des
	// Fensters kommen nacheinander über die Brücke, und der MCP-Server und die
	// Zeile arbeiten seit je so (`alleinzugriff.go`). Was sie **verlangt**:
	// kein Weg hält eine Verbindung und fragt daneben nach einer zweiten —
	// jeder Test dieses Pakets läuft gegen den Arbeitsspeicher mit derselben
	// Zahl, und einer, der das täte, stünde dort still.
	db.SetMaxOpenConns(1)

	// **Die Pragmas gelten je Verbindung, nicht je Datenbank.**
	//
	// Der erste Anlauf setzte sie über `db.ExecContext` — und traf damit
	// **eine** Verbindung aus dem Vorrat. Jede weitere, die `database/sql`
	// später öffnet, kam ohne sie: ohne Fremdschlüssel, ohne Wartezeit. Gesehen
	// hat das erst der Test mit zwei Schreibern auf einer Datei (#143): 17 von
	// 20 Schreibvorgängen scheiterten mit `SQLITE_BUSY`, obwohl fünf Sekunden
	// Wartezeit eingestellt schienen.
	//
	// In der Datenquelle stehen sie für **jede** Verbindung.
	if _, fehler := db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); fehler != nil {
		// Der Schließfehler wird angehängt, nicht verworfen. Der eigentliche
		// Befund ist der des Pragmas — aber eine Datenbank, die sich auch nicht
		// schließen lässt, ist die zweite Hälfte der Nachricht, und wer sie
		// wegwirft, sucht später an der falschen Stelle.
		//
		// `journal_mode` steht hier und nicht in der Datenquelle: es ist eine
		// Eigenschaft der **Datei** und bleibt, einmal gesetzt.
		return nil, errors.Join(nurLesbar(quelle, uebersetze(fehler, "journal_mode")), db.Close())
	}

	return db, nil
}

// alsQuelle macht aus einem Dateipfad eine Datenquelle für SQLite.
//
// **Was schon eine ist, bleibt.** [ImArbeitsspeicher] und jede Angabe, die mit
// `file:` beginnt, hat ihre Form bereits — sie ein zweites Mal zu kodieren
// machte aus `%25` ein `%2525`.
func alsQuelle(quelle string) string {
	if strings.HasPrefix(quelle, "file:") {
		return quelle
	}

	return AlsURI(quelle)
}

// mitPragmas hängt die Pragmas an eine Datenquelle.
//
// **Je Verbindung**, deshalb in der Datenquelle: der Treiber führt `_pragma=`
// aus, sobald er eine Verbindung öffnet. **Nachgemessen am 2026-09-06** für
// `ncruces/go-sqlite3`.
//
// `busy_timeout` lässt einen zweiten Schreiber warten, statt ihn abzuweisen;
// `foreign_keys` ist die Bedingung dafür, dass ein Schreiben auf eine Karte,
// die es nicht gibt, als `ErrNotFound` ankommt und nicht als stiller Erfolg.
//
// # Und `synchronous` — nachgetragen am 2026-09-07
//
// **Es fehlte** (Werkstück 7 von Phase 4.13,
// #1352), und
// wie es sich zeigte, hing davon ab, wen man fragte. Eine einzelne Abfrage
// unmittelbar nach `Oeffnen` gab **2** — die Verbindung, die eben `PRAGMA
// journal_mode = WAL` ausgeführt hatte, kam aus dem Vorrat zurück. Vier
// **gleichzeitig** offene Verbindungen gaben **1**.
//
// **Das ist die Reichweite einer Messung** (§10): die erste Zahl war richtig
// und sagte nichts über den Vorrat. Was wirklich galt, war NORMAL — und ob ein
// Schreibvorgang FULL bekam, war Zufall des Vorrats.
//
// Der Unterschied ist nicht theoretisch: NORMAL im WAL-Modus überlebt einen
// **Prozessabsturz**, aber keinen **Stromausfall** — die letzten quittierten
// Änderungen können fehlen.
//
// **Was FULL kostet, ist gemessen** (2026-09-07, APFS auf SSD): 200 einzelne
// Schreibvorgänge in 8,6 ms statt 2,4 ms — 43 µs statt 12 µs je Vorgang. Für
// eine Anwendung, die ein Kartenfeld je Tastendruck schreibt, ist das nicht zu
// bemerken; die Zusage ist es.
//
// **Gesetzt wird es trotzdem nicht hier, sondern im Verbinder** ([oeffner]):
// der Treiber führt `_pragma=` **vor** dem Rückruf aus, also vor dem
// Schlüssel, und auf einem verschlüsselten Bestand scheitert es dann. Es steht
// hier beschrieben, weil es hier hingehört; es wirkt einen Schritt weiter.
func mitPragmas(quelle string) string {
	trenner := "?"
	if strings.Contains(quelle, "?") {
		trenner = "&"
	}

	// **`_txlock=immediate` ist keine Feinheit, sondern die Bedingung.**
	//
	// Eine Transaktion beginnt sonst **aufschiebend**: sie nimmt die
	// Leseschranke und will beim ersten Schreiben auf die Schreibschranke
	// wechseln. Im WAL-Modus geht das nicht, wenn inzwischen jemand anders
	// geschrieben hat — SQLite meldet `SQLITE_BUSY_SNAPSHOT`, und dagegen hilft
	// `busy_timeout` **nicht**: es wartet auf eine freie Schranke, nicht auf
	// eine Momentaufnahme, die es nicht mehr gibt.
	//
	// Gefunden hat das der Test mit zwei Schreibern auf einer Datei, sobald er
	// ohne `-race` lief: `database is locked (517)`. Mit `-race` blieb er grün,
	// weil der Wettlaufprüfer die Läufe so verlangsamt, dass sie einander nicht
	// mehr treffen — ein grüner Lauf, der seinen Fall nicht ausführt.
	//
	// Mit `immediate` nimmt jede Transaktion die Schreibschranke gleich am
	// Anfang; wer sie nicht bekommt, wartet die fünf Sekunden ab.
	return quelle + trenner + "_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_txlock=immediate"
}

// FuerLaeufe macht eine Datenbank auf demselben Weg auf wie [Oeffnen] — ohne
// Pragmas und ohne Sperre.
//
// # Wofür es das gibt
//
// **Weil ein Lauf, der `sql.Open` selbst ruft, den Verbindungshaken verliert.**
// Seit Werkstück 14 von Phase 4.12 meldet dieser Treiber die Volltextsuche je
// Verbindung an; wer daran vorbei öffnet, bekommt „no such module: fts5",
// sobald die dritte Wanderung läuft.
//
// **Gemessen am 2026-09-06**: neun Läufe der Speicherschicht und fünf in
// `internal/app` öffneten selbst. Ein Helfer ist hier billiger als vierzehn
// Kopien desselben Hakens — und er kann nicht auseinanderlaufen.
//
// **Exportiert, weil `internal/app` ihn auch braucht**: dort stehen fünf
// Läufe, die eine Datei anlegen, um zu sehen, was die Anwendung damit tut.
func FuerLaeufe(quelle string) (*sql.DB, error) { return oeffner(quelle, "") }

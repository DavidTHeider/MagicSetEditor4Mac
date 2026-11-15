// Das Protokoll der Anwendung — in eine Datei und ins Unified Log.
//
// # Was heute protokolliert wurde
//
// **Eine Stelle**: ein `log.Printf`, wenn sich der Bestand nicht schließen
// ließ. `log/slog` kam im ganzen Projekt nicht vor
// (#848).
//
// Das ist eine Schuld und keine Zugabe: die Fehlerwege aus den Phasen M und N
// sind genau die, die man später **nur** im Protokoll sieht.
//
// # Warum zwei Senken
//
// Die **Datei** liegt neben dem Bestand und ist das, was jemand mitschicken
// kann. Das **Unified Log** ist das, worin man ohne die Datei sucht — mit
// `log show`, gegen den Zeitpunkt eines Absturzes.

package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Protokollname ist der Dateiname neben dem Bestand.
const Protokollname = "mse.log"

// Vorprotokollname ist die Fassung vom letzten Umlauf.
//
// **Genau eine, und keine Kette.** `.1`, `.2`, `.3` wüchse weiter — nur
// langsamer; zwei Dateien sind eine Grenze.
const Vorprotokollname = Protokollname + ".1"

// Protokollgrenze ist die Größe, ab der eine neue Datei beginnt.
//
// # Woher die Zahl kommt
//
// **Gemessen und nicht gewählt** (Werkstück 6 von Phase 4.13,
// #1352): am
// 2026-09-07 trug `mse.log` auf dem Rechner des Auftraggebers 1.347.580 Byte
// in 9.220 Zeilen, ältester Eintrag vom 2026-08-22 — sechzehn Tage, rund 84 kB
// am Tag.
//
// Vier Megabyte sind danach **rund sieben Wochen** je Datei und mit der
// Vorfassung gut drei Monate. Das ist reichlich für „was war letzte Woche?" und
// begrenzt den Platz auf acht Megabyte statt auf gar nichts.
const Protokollgrenze int64 = 4 << 20

// Protokollpfad gibt den Ort der Protokolldatei.
func Protokollpfad() (string, error) {
	// **Neben dem Bestand, der wirklich benutzt wird** — der Befund vom
	// 2026-09-14. Hier stand [Datenbankpfad], und damit schrieb ein Lauf mit
	// gesetztem `MSE_BESTAND` sein Protokoll neben den **echten** Bestand:
	// die Sichtprüfung suchte den Grund eines Fehlers an einem Ort, an dem
	// nichts stand. Dieselbe Klasse wie in `main.go`, und sie fiel im selben
	// Zug auf.
	bestand, err := BestandsortFuerFlaechen()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(bestand), Protokollname), nil
}

// Serverprotokoll ist der Schreiber von Dienst und Werkzeugserver —
// geschwärzt und mit Vorgangskennung, auf die Standardfehlerausgabe.
//
// **Befund 7 des Zweitblicks über Phase 6.4**: beide schrieben über den
// Standardschreiber von Go, ohne Schwärzung und ohne Kennung, während die
// Dokumentation sagte, jede Zeile eines Rufs trage sie. Die Einstellungen
// des Protokolls (OPS-04) gelten für das Fenster; ein Server schreibt dorthin,
// wo der ihn Startende seine Fehlerausgabe hinlegt.
func Serverprotokoll(ziel io.Writer) *slog.Logger {
	return slog.New(&geschwaerzt{innen: slog.NewTextHandler(ziel, nil)})
}

// Protokoll legt einen Schreiber an, der in die Datei **und** ins Unified Log
// schreibt — über dieselbe Weiche wie das Fenster.
//
// **Der Rückgabewert schließt beides.** Ein Protokoll, dessen Datei offen
// bleibt, verliert seine letzten Zeilen — und das sind die interessanten.
//
// # Warum über die Weiche
//
// **Nachlauf über Phase 6.4, N12**: hier stand ein eigener Weg mit einem
// `io.MultiWriter`, den der Zweitblick an der Weiche schon behoben hatte —
// ein Kanal, der scheitert, nahm den anderen mit. Kein Weg im Produkt rief
// ihn noch, und acht Testdateien prüften ihn. Jetzt prüfen sie die Weiche.
func Protokoll(pfad string) (*slog.Logger, func() error, error) {
	if filepath.Base(pfad) != Protokollname {
		return nil, nil, fmt.Errorf("ein Protokoll heißt %s und nicht %s", Protokollname, filepath.Base(pfad))
	}

	// **Gedreht wird beim Start** (Werkstück 6 von Phase 4.13) — mit der
	// Vorgabe, weil hier keine Einstellung gilt — **und seit PERF03 auch
	// während der Sitzung** (`protokolldrehung.go`).
	weiche := NeueProtokollweiche(filepath.Dir(pfad))
	weiche.Einrichten(domain.ProtokollregelAus(nil))

	stand := weiche.Stand()
	if stand.Datei != filepath.Clean(pfad) {
		grund := ""

		for _, kanal := range stand.Kanaele {
			if kanal.Name == KanalDatei {
				grund = kanal.Grund
			}
		}

		return nil, nil, errors.Join(fmt.Errorf("Protokolldatei öffnen: %s", grund), weiche.Schliessen())
	}

	return weiche.Logger(), weiche.Schliessen, nil
}

// Protokollsperre ist die Nachbardatei, über die sich zwei Prozesse
// verständigen, wer das Protokoll drehen darf.
//
// **Wie `bestand.db.sperre` und aus demselben Grund** ([0.3.0, #1163]): eine
// Datei, die **nie** umbenannt wird, ist der einzige verlässliche Anker, wenn
// die andere gerade verschwindet.
const Protokollsperre = Protokollname + ".sperre"

// drehenMit legt ein zu großes Protokoll als Vorfassung beiseite — **wenn der
// Rufer allein ist**.
//
// # Warum eine Sperre
//
// **Befund 5 des Zweitblicks über Phase 4.13** (2026-09-07): jeder Prozess
// öffnet denselben Pfad, und gedreht wurde ohne Absprache. Startet B, während
// A in ein volles `mse.log` schreibt, wird **As offene Datei** zu `.1` — und
// ein späteres Wegräumen nimmt ihr den letzten Namen. POSIX sagt zu, dass der
// Inhalt bis zum Schließen bestehen bleibt; nur findet ihn niemand mehr.
//
// **Die Klasse war**: „laufend" wurde aus einem Pfadnamen abgeleitet statt
// aus der Frage, ob jemand anderes schreibt. Die Frage lässt sich stellen —
// der Alleinanspruch auf die Nachbardatei ist die Antwort.
//
// # Der Griff bleibt liegen, und das ist Absicht
//
// Er wird **geteilt** gehalten, solange das Protokoll offen ist: daran
// scheitert der Alleinanspruch des nächsten Starts, und genau das ist die
// Auskunft. Ein Leck ist es nicht — es ist **eine** Datei je Prozess, und sie
// geht mit dem Schließen zurück (anders als bei `bestand.db.sperre`, wo ein
// Griff **je Bestand** entstünde).
//
// # Jeder Fehlschlag heißt „nicht gedreht"
//
// Ein Protokoll, das sich nicht drehen lässt, hält die Anwendung nicht auf —
// dasselbe gilt schon dafür, dass es sich gar nicht anlegen lässt (`main.go`).
// Was bliebe, wäre eine Meldung über das Protokoll **im** Protokoll, das es
// nicht gibt.
//
// # Mit Grenze und Zahl der Vorfassungen
//
// OPS-04 (#1721):
// `mse.log.1` ist die jüngste, `mse.log.<aufbewahrung>` die älteste, und die
// älteste wird beim nächsten Drehen überschrieben. **Was über einer
// gesenkten Aufbewahrung schon liegt, bleibt** — gedreht wird nur bis zur
// eingestellten Zahl; die übrigen räumt die Platzübersicht (Nachlauf über
// Phase 6.4, N11 — hier stand, sie fielen weg).
func drehenMit(pfad string, grenze int64, aufbewahrung int) *dateisperre.Griff {
	griff, fehler := dateisperre.Oeffnen(
		filepath.Join(filepath.Dir(pfad), Protokollsperre))
	if fehler != nil {
		return nil
	}

	// **Hier einmal gefragt und nicht abgewartet** (#2401): neben einem
	// laufenden Programm hält jeder Start die Sperre geteilt (unten), und ein
	// Abwarten des Erben kostete dann jeden Start die volle Erbfrist. Fällt
	// das Drehen hier einem Erben zum Opfer, holt die Drehung beim Schreiben
	// es nach (`protokolldatei.drehen`).
	angabe, statfehler := os.Stat(pfad)
	if statfehler == nil && angabe.Size() >= grenze &&
		griff.Allein() == nil {
		vorfassungenRuecken(pfad, aufbewahrung)
	}

	// **Und danach geteilt** — der Alleinanspruch von eben war nur die Frage;
	// gehalten wird die Antwort für den nächsten Start.
	//
	// # Wartend und nicht auf einen Versuch
	//
	// **Hier stand `griff.Teilen()` mit der Begründung, `LOCK_SH` könne auf
	// einem gültigen Deskriptor nicht scheitern. Das ist falsch**, und der
	// Nachlauf zum Zweitblick über Phase 4.13 hat es angestrichen: `flock`
	// steht hier unter `LOCK_NB`, und wenn ein **gleichzeitig** startender
	// Prozess gerade den Alleinanspruch hält, scheitert die geteilte Sperre
	// mit `EWOULDBLOCK`.
	//
	// Danach schriebe dieser Prozess **ohne Sperre** — und der nächste Start
	// dürfte seine offene Datei drehen. Genau der Fall, den Befund 5
	// abstellen sollte, eine Ebene tiefer.
	//
	// **Die Frist ist kurz**: der fremde Alleinanspruch dauert ein `Stat` und
	// ein `Rename`. Wer sie trotzdem verstreichen lässt, schreibt ohne Sperre
	// weiter — ein Protokoll, das auf eine Sperre wartet, ist schlimmer als
	// eines ohne.
	_ = griff.TeilenBis(context.Background(), //nolint:errcheck // siehe oben
		sperrfrist, sperrtakt)

	return griff
}

// Wie lange auf die geteilte Protokollsperre gewartet wird.
//
// **Zwei Sekunden in Schritten von zwanzig Millisekunden** — der fremde
// Alleinanspruch, auf den gewartet wird, dauert ein `Stat` und ein `Rename`.
const (
	sperrfrist = 2 * time.Second
	sperrtakt  = 20 * time.Millisecond
)

// schliesseGriff gibt eine Sperre zurück, wenn es eine gibt.
func schliesseGriff(griff *dateisperre.Griff) {
	if griff != nil {
		_ = griff.Schliessen() //nolint:errcheck // beim Aufräumen ist nichts zu retten
	}
}

// geschwaerzt nimmt aus jeder Zeile, was nicht hineingehört.
//
// # Drei Regeln
//
// **Der Pfad des Benutzerverzeichnisses wird zur Tilde.** Ein Protokoll wird
// mitgeschickt, und `/Users/vorname` sagt jedem, der es liest, wie jemand
// heißt. Der **Dateiname** bleibt — er ist die Auskunft, die den Fehler
// verständlich macht; ohne ihn wäre die Meldung nutzlos.
//
// **Ein Merkmal, das niemand angemeldet hat, verliert seinen Wert**
// ([erlaubteMerkmale], #857). Damit kostet eine vergessene Anmeldung eine
// unlesbare Zeile und nicht einen ausgeplauderten Kartentext.
//
// **Die Gestalt eines Nachweises wird geschwärzt** ([redigiere]) — SEC-P5
// (#2001).
// Bis dahin tat das nur das Diagnosepaket, und `mse.log`, das Unified Log
// und die Fehlerausgabe der Programme trugen einen Zugang aus der Adresse
// eines privaten Repositoriums, sobald ein Bezug scheiterte. **Hier und nur
// hier**, weil jede Zeile jedes Kanals hier hindurchgeht.
//
// # Warum als Handler und nicht als Regel am Aufrufer
//
// **Weil ein Aufrufer sie vergisst.** Ein Wächter, den jede Schreibstelle
// selbst rufen muss, hält genau so lange, bis jemand eine neue Stelle
// schreibt. Hier geht jede Zeile hindurch.
//
// # Was er nicht hält
//
// **Ob der Wert zu seinem Namen passt.** Ein `ort`, unter dem ein Kartenname
// steht, kommt durch — die Liste kennt Namen, nicht Bedeutungen, und keine
// Maschine unterscheidet einen getippten Kartennamen von einem getippten
// Dateinamen. Das ist eine Frage ans Review, und sie steht hier, damit
// niemand die Liste für mehr hält, als sie ist (§10).
//
// **Und den Meldungstext.** `schreiber.Info("Feuerball gespeichert")` ginge
// durch. Ein Wächter über die Meldungen selbst prüft das
// (`internal/werkzeuge/protokollmerkmale_test.go`): dort müssen sie
// Zeichenkettenliterale sein, damit sie nichts Gerechnetes tragen können.
type geschwaerzt struct{ innen slog.Handler }

func (g *geschwaerzt) Enabled(ctx context.Context, stufe slog.Level) bool {
	return g.innen.Enabled(ctx, stufe)
}

func (g *geschwaerzt) Handle(ctx context.Context, satz slog.Record) error {
	sauber := slog.NewRecord(satz.Time, satz.Level, redigiere(satz.Message), satz.PC)

	satz.Attrs(func(merkmal slog.Attr) bool {
		sauber.AddAttrs(sauberesMerkmal(merkmal))

		return true
	})

	// **Die Vorgangskennung setzt der Handler und nicht der Aufrufer** —
	// Werkstück O9 von Phase 5.2 (#1433), und aus demselben Grund, aus dem
	// das Schwärzen hier steht: **weil ein Aufrufer sie vergisst**. Hier geht
	// jede Zeile hindurch, und wer `…Context(ctx, …)` ruft, bekommt sie ohne
	// eine Zeile Code.
	//
	// **Wer sie nicht bekommt**, ist eine Zeile ohne Kontext —
	// `slog.Default().Error(…)` statt `ErrorContext(ctx, …)`. Ein Wächter
	// zählt die Stellen nach, an denen ein Kontext in Reichweite war
	// (`internal/werkzeuge/protokollkontext_test.go`).
	if kennung := service.VorgangAus(ctx); kennung != "" {
		sauber.AddAttrs(slog.String("kennung", string(kennung)))
	}

	return g.innen.Handle(ctx, sauber)
}

// WithAttrs schwärzt auch die **gebundenen** Merkmale.
//
// **Der erste Anlauf reichte sie ungeschwärzt durch**, und der Kommentar
// darüber behauptete das Gegenteil. Gefangen hat es der Test, der `With(…)`
// benutzt: `ort=/Users/david/geheim` stand im Protokoll, während das Merkmal
// derselben Zeile schon geschwärzt war. Ein Wächter, der die halbe Zusage
// hält, macht die andere unsichtbar (§10).
func (g *geschwaerzt) WithAttrs(merkmale []slog.Attr) slog.Handler {
	sauber := make([]slog.Attr, 0, len(merkmale))
	for _, merkmal := range merkmale {
		sauber = append(sauber, sauberesMerkmal(merkmal))
	}

	return &geschwaerzt{innen: g.innen.WithAttrs(sauber)}
}

// sauberesMerkmal gibt ein Merkmal mit geschwärztem Wert.
//
// **Der Name bleibt auch dann stehen, wenn der Wert fällt.** Ohne ihn sähe
// niemand, dass etwas fehlt — und eine stille Lücke ist schlimmer als eine
// sichtbare.
//
// **Eine Gruppe bleibt eine Gruppe** (Nachlauf über Phase 5 von 0.9.0,
// Befund 1, #2399): bis dahin wurde sie hier zu ihrer Zeichenkette,
// `[panik=Kennwort hunter]`, und darin erkannte der Filter des Pakets nur
// `panik=Kennwort`. Jetzt wird jedes Merkmal darin für sich geschwärzt, unter
// seinem eigenen Namen, und die Weiche schreibt `grund.panik="…"` oder
// `"grund":{"panik":"…"}` — die Gestalt, die der Filter kennt.
func sauberesMerkmal(merkmal slog.Attr) slog.Attr {
	if !merkmalErlaubt(merkmal.Key) {
		return slog.String(merkmal.Key, Nichterklaert)
	}

	wert := merkmal.Value.Resolve()
	if wert.Kind() == slog.KindGroup {
		innen := wert.Group()
		sauber := make([]slog.Attr, 0, len(innen))

		for _, eines := range innen {
			sauber = append(sauber, sauberesMerkmal(eines))
		}

		return slog.Attr{Key: merkmal.Key, Value: slog.GroupValue(sauber...)}
	}

	return slog.Attr{
		Key:   merkmal.Key,
		Value: slog.StringValue(redigiere(wert.String())),
	}
}

func (g *geschwaerzt) WithGroup(name string) slog.Handler {
	return &geschwaerzt{innen: g.innen.WithGroup(name)}
}

// schwaerze ersetzt das Benutzerverzeichnis durch eine Tilde.
//
// **Wie eine Muschel es schreibt** — `~/Dokumente/…`. Wer das Protokoll liest,
// versteht den Ort, ohne den Namen zu erfahren.
func schwaerze(text string) string {
	zuhause, err := os.UserHomeDir()
	if err != nil || zuhause == "" {
		return text
	}

	return strings.ReplaceAll(text, zuhause, "~")
}

// vorfassung gibt den Namen der n-ten Vorfassung neben einer Protokolldatei.
func vorfassung(pfad string, stelle int) string {
	return pfad + "." + strconv.Itoa(stelle)
}

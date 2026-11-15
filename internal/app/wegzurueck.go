package app

// **Der Weg zurück aus einer neueren Fassung** — Werkstück 5 von Phase 4.13
// (#1352).
//
// # Was hier fehlte
//
// Die Abweisung steht seit Werkstück 3 und rät: „Hole dir die neuere Fassung,
// oder spiele eine Sicherung von vorher zurück." **Welche**, sagt sie nicht —
// und neben dem Bestand liegen bis zu dreißig Dateien mit Fassungsnummern im
// Namen, in einem Ordner, den niemand von sich aus öffnet (Werkstück 4).
//
// # Warum nicht einfach die jüngste
//
// **Weil sie im Ernstfall die falsche ist.** `bestand-vor-<F>.db` hält den
// Stand **vor** der Migration auf F. Wer von 0.3.0 über 0.4.0 auf 0.5.0
// gegangen ist, hat neben dem Bestand:
//
//	bestand-vor-0.4.0.db   Stand 0.3.0 — 0.3.0 liest ihn
//	bestand-vor-0.5.0.db   Stand 0.4.0 — jünger, und für 0.3.0 **kein** Weg
//
// Ein Rat auf die jüngste führte in dieselbe Abweisung, aus der er
// herausführen soll. **Gefragt wird deshalb die Datei selbst** — über
// dieselbe Prüfung, die auch das Zurückspielen benutzt
// ([Sicherungspruefer]). Eine zweite wäre eine zweite Fassung derselben
// Frage.

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// RueckwegDTO nennt die Sicherung, mit der diese Fassung wieder arbeiten kann.
//
// **Der Name und nicht der Pfad.** Ein Pfad aus dem Inneren ist genau das, was
// die Abnahme dieser Phase verbietet; der Name ist das, was jemand im Finder
// sieht und im Dateidialog wiedererkennt.
//
// **Und die Zahlen dazu**, weil „eine Sicherung" nichts sagt und „vom Stand
// 0035, mit 12 Sets und 840 Karten" die Entscheidung trägt.
type RueckwegDTO struct {
	Datei       string `json:"datei"`
	Schemastand string `json:"schemastand"`
	Sets        int    `json:"sets"`
	Karten      int    `json:"karten"`
}

// wegZurueck sucht neben dem Bestand die jüngste Sicherung, die aufgeht.
//
// **Gibt es keine, gibt es `nil`** — und der Befund nennt dann keine. Ein
// Hinweis auf eine Datei, die niemand findet, ist schlimmer als keiner (§10).
func (a *App) wegZurueck(ctx context.Context) *RueckwegDTO {
	pfad, befund := a.juengsteLesbare(ctx)
	if pfad == "" {
		return nil
	}

	return &RueckwegDTO{
		Datei:       filepath.Base(pfad),
		Schemastand: befund.Schemastand,
		Sets:        befund.Sets,
		Karten:      befund.Karten,
	}
}

// juengsteLesbare gibt die jüngste Vorsicherung, die diese Fassung öffnet.
//
// **Ein Lesefehler am Ordner heißt „kein Weg" und nicht „Fehler".** Gerufen
// wird das hier, während der Start ohnehin gescheitert ist; eine zweite
// Meldung darüber, dass auch das Nachsehen nicht ging, verdeckte die erste.
func (a *App) juengsteLesbare(ctx context.Context) (pfad string, befund Sicherungsbefund) {
	dateien, fehler := vorsicherungen(a.datenbankpfad)
	if fehler != nil {
		return "", Sicherungsbefund{}
	}

	sicherungen := ohneBegleiter(dateien)
	nachAlterAbsteigend(sicherungen)

	for _, datei := range sicherungen {
		if gelesen, untauglich := a.sicherungspruefer(ctx, datei.pfad); untauglich == nil {
			return datei.pfad, gelesen
		}
	}

	return "", Sicherungsbefund{}
}

// begleiterendungen sind die Dateien, die SQLite **neben** eine Datenbank legt.
//
// **Sie tragen denselben Namensanfang und sind keine Sicherung.** `-wal` und
// `-shm` gehören zu der Datei, neben der sie liegen; für sich genommen sind
// sie wertlos.
var begleiterendungen = []string{"-wal", "-shm", "-journal"}

// ohneBegleiter gibt die eigentlichen Sicherungen.
//
// # Der Befund, an der laufenden Anwendung gesehen
//
// **Am 2026-09-07 in der Sichtprüfung**: nach dem Wegräumen standen
// `bestand-vor-0.4.0.db` und `bestand-vor-0.4.0.db-shm` da — zwei Dateien, aber
// **ein** Rückweg. Der Platz, den [App.bleibendeVorsicherungen] für die
// jüngste Datei überhaupt freihält, war von einem Begleiter belegt.
//
// **Und der entsteht beim Prüfen selbst.** Gemessen am selben Tag: eine
// `PruefeSicherung` auf einer Datenbank im WAL-Modus legt `-shm` und `-wal`
// daneben an, obwohl sie mit `mode=ro` öffnet — SQLite braucht den geteilten
// Speicher, um den WAL zu lesen. Der Begleiter ist damit **immer** der
// jüngste: er entsteht in dem Augenblick, in dem gefragt wird.
//
// **Ein Test hätte das nicht gezeigt.** Er sieht die Zahl, und zwei Dateien
// sind zwei Dateien.
func ohneBegleiter(dateien []platzdatei) []platzdatei {
	eigentliche := make([]platzdatei, 0, len(dateien))

	for _, datei := range dateien {
		if !einBegleiter(datei.pfad) {
			eigentliche = append(eigentliche, datei)
		}
	}

	return eigentliche
}

// einBegleiter sagt, ob ein Pfad zu einer Datenbank daneben gehört.
func einBegleiter(pfad string) bool {
	for _, endung := range begleiterendungen {
		if strings.HasSuffix(pfad, endung) {
			return true
		}
	}

	return false
}

// bleibendeVorsicherungen sagt, welche Dateien das Wegräumen stehen lässt.
//
// **Zwei, und die Gründe sind verschieden:**
//
//  1. **die jüngste überhaupt** — der Weg zurück für die Fassung, die sie
//     geschrieben hat. Vielleicht installiert die jemand als Nächstes wieder;
//     sie wegzuräumen nähme ihm diesen Weg.
//  2. **die jüngste lesbare** — der Weg zurück für **diese** Fassung. Sie ist
//     im Normalfall dieselbe Datei; im Ernstfall, in dem jemand aufräumt,
//     weil der Bestand aus der Zukunft nicht aufgeht, ist sie es nicht.
//
// **Beim ersten Aufräumen bleibt damit oft nur eine übrig**, und das ist
// richtig so: dann sind beide Gründe dieselbe Datei.
func (a *App) bleibendeVorsicherungen(ctx context.Context, dateien []platzdatei) map[string]bool {
	bleiben := map[string]bool{}

	// **Begleiter sind keine Kandidaten** ([ohneBegleiter]) — sonst belegte
	// ein `-shm`, das die Prüfung eben erst angelegt hat, den Platz der
	// jüngsten Datei.
	sicherungen := ohneBegleiter(dateien)
	nachAlterAbsteigend(sicherungen)

	if len(sicherungen) > 0 {
		samtBegleitern(bleiben, sicherungen[0].pfad, dateien)
	}

	if lesbar, _ := a.juengsteLesbare(ctx); lesbar != "" {
		samtBegleitern(bleiben, lesbar, dateien)
	}

	return bleiben
}

// samtBegleitern merkt eine Sicherung vor — und was zu ihr gehört.
//
// **Eine Sicherung ohne ihren `-wal` ist der Stand von vorgestern.** Wer sie
// behielte und den Begleiter wegräumte, ließe eine Datei stehen, die weniger
// enthält, als sie enthielt.
func samtBegleitern(bleiben map[string]bool, sicherung string, alle []platzdatei) {
	bleiben[sicherung] = true

	for _, datei := range alle {
		for _, endung := range begleiterendungen {
			if datei.pfad == sicherung+endung {
				bleiben[datei.pfad] = true
			}
		}
	}
}

// nachAlterAbsteigend sortiert die jüngste nach vorn.
func nachAlterAbsteigend(dateien []platzdatei) {
	sort.Slice(dateien, func(eine, andere int) bool {
		return dateien[andere].alter.Before(dateien[eine].alter)
	})
}

// WegZurueck nennt die Sicherung neben einem Bestand, die diese Fassung öffnet.
//
// **Für die Flächen ohne Fenster** (§16): das Fenster zeigt den Weg zurück auf
// der Wiederherstellungsseite, und wer über Kommandozeile, Dienst oder MCP auf
// denselben Bestand trifft, soll dieselbe Auskunft bekommen.
//
// **Über den Ort und nicht über eine Verbindung** — wie [PlatzflaecheUeber],
// und aus demselben Grund: gefragt wird genau dann, wenn der Bestand **nicht**
// aufgeht.
func WegZurueck(ctx context.Context, bestandsort string) *RueckwegDTO {
	return NeueApp(bestandsort).wegZurueck(ctx)
}

// MitWegZurueck hängt an einen gescheiterten Start den Weg zurück.
//
// **Angehängt und nicht ersetzt.** Der Grund bleibt das, was jemand
// weitergibt; der Weg zurück ist das, was er als Nächstes tut.
//
// **Liegt keine brauchbare Sicherung da, bleibt der Fehler, wie er war** —
// ein Hinweis auf eine Datei, die es nicht gibt, ist schlimmer als keiner.
func MitWegZurueck(ctx context.Context, bestandsort string, fehler error) error {
	return mitWegZurueck(ctx, bestandsort, fehler, freierPlatz)
}

// mitWegZurueck ist [MitWegZurueck] mit der Frage nach dem freien Platz.
//
// **Ein Fehler des Datenträgers bei knappem Platz ist ein voller** — wie
// beim Start des Fensters ([vollStattEA]); gefunden an der Kommandozeile mit
// `wanderung09` (M02, #2346).
func mitWegZurueck(ctx context.Context, bestandsort string, fehler error, frei func(string) (int64, error)) error {
	fehler = vollStattEA(fehler, filepath.Dir(bestandsort), frei)
	weg := WegZurueck(ctx, bestandsort)
	if weg == nil {
		return fehler
	}

	return fmt.Errorf(
		"%w\nZurück kommst du über %q neben dem Bestand — Schemastand %s, "+
			"%d Set(s) mit %d Karten",
		fehler, weg.Datei, weg.Schemastand, weg.Sets, weg.Karten)
}

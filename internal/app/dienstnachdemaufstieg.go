package app

// **Wer wandert, hält den eingerichteten Dienst vorher an — nachweislich** —
// M02 (#2346),
// Befund A-11; Zweitblick über Phase 2 von 0.9.0, Befunde 5 und 6 (#2323).
//
// Ein Dienst unter launchd überlebt den Tausch des Bündels und läuft mit dem
// Code der Fassung davor weiter. Ab 0.9.0 holt er vor jedem Ruf eine Freigabe
// (`fassungswache.go`); **ein Dienst aus 0.8.0 kennt sie nicht**. Ihn hält
// jede Wurzel an, die wandern wird — Fenster, `mse`, `mse-mcp` und `mse-rest`
// über [SicherungVorMigration] —, mit `launchctl kickstart -k`: launchd
// beendet ihn und startet das Programm neu, auf das die plist zeigt. Das
// geschieht **in der Vorbereitung**, also unter der Sperre über die
// Wanderung und **vor** der Sicherung; der neu gestartete Dienst wartet an
// der Sperre, bis gewandert ist.
//
// # Nachweislich
//
// **Gewandert wird erst, wenn der alte Prozess fort ist.** Bis zum Zweitblick
// reichte ein Ruf, dessen Scheitern nur im Protokoll stand — und ein alter
// Dienst, der weiterlief, schrieb in den neuen Stand. Bleibt er, ist der
// Bestand belegt, und es wird weder gesichert noch gewandert.
//
// # Wann nicht
//
// Ohne laufenden Dienst, wenn dieser Prozess selbst der Dienst ist, und wenn
// der Dienst einen anderen Bestand hält. **Ein Testlauf und jeder Prozess
// mit [Dienstsperre] ruft `launchctl` nie verändernd** (`befehlAusfuehren`):
// das Ziel ist `gui/<UID>/<Etikett>`, und ein umgelenktes Zuhause lenkt das
// nicht um.
//
// **Was das nicht kann**: eine MCP-Sitzung aus 0.8.0, die ein Sprachmodell
// offen hält, gehört dessen Programm; und zeigt die plist auf ein Programm,
// das selbst alt ist, startet launchd das alte — es weist den neueren Bestand
// beim Start ab und endet so, dass launchd es nicht wieder anwirft
// (`cmd/mse-rest/wiederanlauf.go`).

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dienstprozesse nennt Etikett und Prozessnummer jedes laufenden Dienstes —
// eine Naht für den Lauf, der einen Prozess an seine Stelle setzt.
var dienstprozesse = laufendeDienste

// halterFragen nennt die Prozesse, die den Bestand offen halten — eine Naht
// für die Läufe des Wartens, die sonst davon abhingen, wie schnell `lsof`
// unter Last antwortet.
var halterFragen = lsofFragen

// dienstanhaltefrist und dienstanhaltetakt sind, wie lange und wie oft auf das
// Ende des alten Prozesses gesehen wird.
var (
	dienstanhaltefrist = 10 * time.Second
	dienstanhaltetakt  = 50 * time.Millisecond
)

// dienstwaerter hält den eingerichteten Dienst an.
type dienstwaerter struct {
	befehl befehlsausfuehrung
}

// dienstwaerter ist der Wärter des Fensters — über dessen Ausführer.
func (a *App) dienstwaerter() dienstwaerter {
	return dienstwaerter{befehl: a.schnittstellen.ausfuehren}
}

// mitDienstanhalt gibt eine Vorbereitung, die vor `sichern` den Dienst anhält.
func mitDienstanhalt(sichern, anhalten func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		if fehler := anhalten(ctx); fehler != nil {
			return fehler
		}

		return sichern(ctx)
	}
}

// anhalten hält den Dienst an, wenn er den Bestand unter `pfad` hält, und
// wartet, bis der alte Prozess fort ist — **alles in einer Frist**
// ([dienstanhaltefrist]), auch der Ruf an `launchctl` (Nachlauf über Phase 2,
// Befund 5).
//
// **Nur ein Dienst, der diesen Bestand hält** (Befunde 1, 5 und 6): gefragt
// bei `lsof`, nicht an der plist auf der Platte — ein geladener Dienst kann
// ohne sie weiterlaufen, und ein Dienst, der einen anderen Bestand bedient,
// schreibt nicht in diesen. Ein Lauf mit eigenem Bestand trifft so nie den
// Dienst dieses Rechners. **Ein Nachweis, der scheitert, ist kein Nein.**
func (w dienstwaerter) anhalten(ctx context.Context, pfad string) error {
	if pfad == "" {
		return nil
	}

	ctx, ende := context.WithTimeout(ctx, dienstanhaltefrist)
	defer ende()

	laeufe, fehler := dienstprozesse(ctx)
	if fehler != nil {
		return dienstNichtAngehalten(fmt.Errorf("den Dienst befragen: %w", fehler))
	}

	// **Jeder, der läuft** (Zweitblick über Phase 3, Befund 8) — der neue
	// kann einen anderen Bestand bedienen und der alte diesen halten.
	for _, lauf := range laeufe {
		if lauf.pid == os.Getpid() {
			continue
		}

		if fehler := w.einenAnhalten(ctx, pfad, lauf); fehler != nil {
			return fehler
		}
	}

	return nil
}

// einenAnhalten hält einen Dienst an, wenn er den Bestand hält, und wartet,
// bis er ihn nicht mehr hält.
func (w dienstwaerter) einenAnhalten(ctx context.Context, pfad string, lauf dienstprozess) error {
	haelt, fehler := haeltDatei(ctx, pfad, lauf.pid)
	if fehler != nil {
		return dienstNichtAngehalten(fehler)
	}

	if !haelt {
		return nil
	}

	befehl := []string{"launchctl", "kickstart", "-k", "gui/" + strconv.Itoa(os.Getuid()) + "/" + lauf.etikett}
	if fehler := w.befehl(ctx, befehl); fehler != nil {
		return dienstNichtAngehalten(fehler)
	}

	// **Gewartet wird, bis er den Bestand nicht mehr hält** — nicht, bis
	// seine Nummer verschwunden ist (Nachlauf über Phase 2, Befund 5): ein
	// Prozess, der noch nicht abgeholt ist, hält keine Datei mehr, und eine
	// wiederverwendete Nummer müsste ausgerechnet diesen Bestand halten.
	for {
		haelt, fehler := haeltDatei(ctx, pfad, lauf.pid)
		if fehler == nil && !haelt {
			break
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return dienstNichtAngehalten(fmt.Errorf("der Prozess %d hält den Bestand nach %s noch", lauf.pid, dienstanhaltefrist))
			}

			return ctx.Err()
		case <-time.After(dienstanhaltetakt):
		}
	}

	// **Und das Protokoll sagt es** — ein Neustart, den niemand sieht, erklärt
	// keine abgebrochene Anfrage (Abnahme 2 von #2346).
	slog.InfoContext(ctx, "der Dienst wurde vor der Wanderung neu gestartet", "dienst", lauf.etikett)

	return nil
}

// haeltDatei sagt, ob der Prozess `pid` die Datei `pfad` offen hält —
// gefragt bei `lsof`.
func haeltDatei(ctx context.Context, pfad string, pid int) (bool, error) {
	halter, fehler := halterFragen(ctx, pfad)
	if fehler != nil {
		return false, fmt.Errorf("fragen, wer den Bestand hält: %w", fehler)
	}

	return slices.Contains(splitZeilen(halter), "p"+strconv.Itoa(pid)), nil
}

// dienstNichtAngehalten ist der Fehler, wenn der alte Dienst bleibt — der
// Bestand ist dann belegt, und es wird weder gesichert noch gewandert.
func dienstNichtAngehalten(grund error) error {
	slog.Warn("der Dienst ließ sich vor der Wanderung nicht anhalten", "grund", grund.Error())

	return domain.MitGrund(domain.ErrBelegt, grund,
		"der Dienst für Skripte ließ sich vor der Aktualisierung nicht anhalten — versuche es gleich noch einmal")
}

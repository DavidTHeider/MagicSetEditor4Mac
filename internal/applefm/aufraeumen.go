package applefm

// **Was ein abgestürzter Lauf zurückließ** — AI-05.AK-07, seit dem
// Zweitblick über Phase 6.10 (Z4) an der genauen Befehlszeile statt an einem
// Muster.

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// aufraeumen beendet, was ein abgestürzter Lauf zurückließ — AI-05.AK-07.
//
// **Der Socket trägt die Kennung des Prozesses, der ihn bestellt hat**
// (`fm-<pid>.sock`). Lebt der nicht mehr, ist das `fm serve` dahinter ein
// Waise: es hängt an launchd und lauschte weiter. Es wird beendet und sein
// Socket gelöscht; der eines lebenden Prozesses bleibt.
func (d *Dienst) aufraeumen(ctx context.Context, ordner string) {
	// **Ein unlesbarer Ordner hat nichts zum Aufräumen** — der Start danach
	// sagt, ob er brauchbar ist.
	eintraege, _ := os.ReadDir(ordner) //nolint:errcheck // s. o.

	for _, eintrag := range eintraege {
		var kennung int
		if _, err := fmt.Sscanf(eintrag.Name(), "fm-%d.sock", &kennung); err != nil ||
			kennung == os.Getpid() || d.lebt(kennung) {
			continue
		}

		// **Scheitert das Beenden, bleibt der Socket** (Z4 des Zweitblicks
		// über Phase 6.10): ohne ihn fände kein späterer Start den Waisen
		// wieder. Der eigene Start nimmt ohnehin einen eigenen Namen.
		pfad := filepath.Join(ordner, eintrag.Name())
		if d.verwaisteBeenden(ctx, pfad) != nil {
			continue
		}

		// **Ohne Treffer heißt nicht: beendet** (N5 des Nachlaufs): weicht
		// die Darstellung der Befehlszeile ab, fand die Liste den Waisen
		// nicht. Lauscht noch jemand am Socket, bleibt er stehen — der
		// nächste Start findet ihn wieder.
		if lauscht(ctx, pfad) {
			continue
		}

		_ = os.Remove(pfad) //nolint:errcheck // ein Rest, der nicht geht, stört den Start nicht
	}
}

// verwaistenBeenden beendet das `fm serve`, das an diesem Socket hing —
// **an seiner genauen Befehlszeile**, nicht an einem Muster. Bis zum
// Zweitblick über Phase 6.10 (Z4) ging der Pfad an `pkill -f`: dort ist er
// ein regulärer Ausdruck, ein `.` trifft jedes Zeichen und ein `[` macht das
// Muster ungültig.
func verwaistenBeenden(ctx context.Context, pfad string) error {
	return genauBeenden(ctx, Programm+" serve --socket "+pfad)
}

// genauBeenden beendet jeden Prozess, dessen Befehlszeile genau diese ist.
func genauBeenden(ctx context.Context, befehl string) error {
	liste, fehler := exec.CommandContext(ctx, "/bin/ps", "-axww", "-o", "pid=,command=").Output()
	if fehler != nil {
		return fmt.Errorf("die Prozessliste lesen: %w", fehler)
	}

	return beendeKennungen(ctx, befehl, kennungenMitBefehl(string(liste), befehl), toete)
}

// beendeKennungen beendet diese Prozesse — **wenn sie unmittelbar vor dem
// Signal noch denselben Befehl tragen** (N4 des Nachlaufs über Phase 6.10):
// zwischen der Liste und dem Signal kann ein Waise enden und seine Kennung
// an einen fremden Prozess gehen. Ein Fenster von Mikrosekunden bleibt —
// macOS kennt keinen Griff, der eine Prozessidentität bis zum Signal bindet.
func beendeKennungen(ctx context.Context, befehl string, kennungen []int, signal func(int) error) error {
	for _, kennung := range kennungen {
		if befehlVon(ctx, kennung) != befehl {
			continue
		}

		if fehler := signal(kennung); fehler != nil {
			return fmt.Errorf("den verwaisten Prozess %d beenden: %w", kennung, fehler)
		}
	}

	return nil
}

// befehlVon liest die Befehlszeile einer Kennung — leer, wenn es sie nicht
// mehr gibt.
func befehlVon(ctx context.Context, kennung int) string {
	// #nosec G204 -- fester Pfad, die Kennung ist eine Zahl
	zeile, fehler := exec.CommandContext(ctx, "/bin/ps", "-ww", "-p", strconv.Itoa(kennung), "-o", "command=").Output()
	if fehler != nil {
		return ""
	}

	return strings.TrimSpace(string(zeile))
}

// toete schickt einer Kennung das Ende.
func toete(kennung int) error {
	prozess, fehler := os.FindProcess(kennung)
	if fehler == nil {
		fehler = prozess.Kill()
	}

	return fehler
}

// kennungenMitBefehl liest die Ausgabe von `ps -o pid=,command=` und nennt
// die Kennungen, deren Befehlszeile **genau** dieser Befehl ist — kein
// Anfang, kein Teil, kein Muster.
func kennungenMitBefehl(liste, befehl string) []int {
	var kennungen []int

	for _, zeile := range strings.Split(liste, "\n") {
		kennung, rest, gefunden := strings.Cut(strings.TrimSpace(zeile), " ")
		if !gefunden || strings.TrimSpace(rest) != befehl {
			continue
		}

		if zahl, fehler := strconv.Atoi(kennung); fehler == nil {
			kennungen = append(kennungen, zahl)
		}
	}

	return kennungen
}

// lauscht sagt, ob an einem Socket noch ein `fm serve` antwortet — über
// denselben Klienten wie jede Frage und nicht über `net`: das Netz kennt in
// dieser Anwendung nur `internal/netz` (#54).
func lauscht(ctx context.Context, pfad string) bool {
	return NeuerKlient(netz.NeuerKlientUeberSocket(netz.UeblicheFristen(), pfad), Socketadresse).
		Gesund(ctx) == nil
}

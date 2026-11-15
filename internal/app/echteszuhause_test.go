// **Kein Lauf dieses Pakets arbeitet im echten Zuhause** — §3, „ein Test, der
// nach draußen greift, ohne dass jemand es wollte".
//
// # Der Befund
//
// Der Nachlauf über Phase 3.2 hat angestrichen, dass die Läufe dieses Pakets
// drei Nachbardateien in `~/Library/Application Support/MagicSetEditor4Mac/`
// angelegt haben — und dabei **grün** blieben
// (#1163).
//
// **Die Pipeline war nicht betroffen**: `scripts/ci.sh` setzt ein leeres
// Zuhause. Betroffen war, wer `go test ./internal/app/` von Hand fährt — und
// das tut hier jeder.
//
// # Die Antwort: das Paket setzt sich selbst eines
//
// **Nicht ein Wächter, der es meldet.** Ein Wächter sagte „du hast dein
// Zuhause verschmutzt", **nachdem** es geschehen ist; die Dateien lägen
// trotzdem da. Ein Zuhause je Testlauf macht die ganze Klasse unmöglich.
//
// **Ein Lauf, der ein eigenes braucht, setzt es weiterhin selbst** — `t.Setenv`
// gilt für den Lauf und wird danach zurückgenommen. Was hier steht, ist der
// **Boden** darunter.
package app_test

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// eigenesZuhauseFuerDenLauf gibt dem ganzen Testlauf ein Zuhause für sich.
//
// Gibt die Aufräumfunktion zurück; sie läuft nach `m.Run()`.
func eigenesZuhauseFuerDenLauf() (aufraeumen func(), fehler error) {
	zuhause, err := os.MkdirTemp("", "mse-testzuhause-")
	if err != nil {
		return nil, fmt.Errorf("ein Testzuhause anlegen: %w", err)
	}

	// **Und mit einem Schlüsselbund, den es nicht gibt.**
	//
	// **Damit kein Lauf den Anmeldebund dieses Rechners anfasst** (§3).
	// `internal/schluesselbund` weist ihn einem Testbinärprogramm ohnehin ab;
	// ohne diese Zeile käme die Abweisung aber dort an, wo bisher ein
	// `ErrNotFound` ankam — und ein Löschweg, der „nichts da" verschluckt,
	// meldete plötzlich einen Fehler.
	//
	// **Gemessen am 2026-09-14**: 84 Zugriffe auf den Anmeldebund aus acht
	// Paketen, darunter zwei Schreibvorgänge von `peeridentitaet`, und
	// dreimal ein Dialog mitten im Lauf.
	if fehler := os.Setenv(schluesselbund.Bundort,
		filepath.Join(zuhause, "keiner.keychain")); fehler != nil {
		return nil, fmt.Errorf("den Probebundort setzen: %w", fehler)
	}

	// **Mit einem Downloadordner darin.** Ein echtes Zuhause hat einen, und
	// der Aktualisierungsweg schreibt dorthin.
	if fehler := os.MkdirAll(filepath.Join(zuhause, "Downloads"), 0o700); fehler != nil {
		return nil, fmt.Errorf("den Downloadordner anlegen: %w", fehler)
	}

	// **Und mit den Schriften dieses Rechners** — als Verweis, nur zu lesen
	// (#1170).
	//
	// # Warum das die Trennung nicht aufhebt
	//
	// Getrennt wird der **Zwischenspeicher**: kein Test soll den Stand dieses
	// Rechners sehen. Die Schriften sind das Gegenteil davon — sie sind
	// **Gegenstand** der Messung. `internal/schriften` sucht unter
	// `~/Library/Fonts`, und genau dort liegen Beleren und MPlantin, die jede
	// Magic-Vorlage nennt.
	//
	// # Was ohne sie geschah
	//
	// **Jede Karte wurde ohne Text gezeichnet, und der Kanonlauf maß das.**
	// Der Abstand zu MSE2 stieg von 18,853 auf 28,362, und die Zerlegung sagte
	// es deutlich: 0 % `Text` statt 56 %. Das sah aus wie ein Rückschritt an
	// der Rasterung und war einer an der Vorrichtung — dieselbe Klasse, die
	// `scryfallabstand_netzlauf_test.go` seit dem 2026-08-23 an seiner Stelle
	// benennt und dort mit [pruefeTextGemalt] hält.
	if fehler := schriftenVerweisen(zuhause); fehler != nil {
		return nil, fehler
	}

	if fehler := os.Setenv("HOME", zuhause); fehler != nil {
		return nil, fmt.Errorf("das Testzuhause setzen: %w", fehler)
	}

	return func() { _ = os.RemoveAll(zuhause) }, nil //nolint:errcheck // ein Verzeichnis unter /tmp
}

// dasEchteZuhauseIstUnberuehrt sagt, ob das Zuhause des Rechners noch gilt.
//
// **Über `os/user` und nicht über `$HOME`**: die Umgebungsvariable ist gerade
// das, was hier ausgetauscht wurde. Gefragt ist, ob der Austausch gewirkt hat.
func dasEchteZuhauseIstUnberuehrt() error {
	wer, err := user.Current()
	if err != nil || wer.HomeDir == "" {
		// **Ohne Auskunft keine Zusicherung.** Sie hier zu erfinden wäre eine
		// Behauptung; wo es kein Benutzerverzeichnis gibt, ist auch nichts zu
		// verschmutzen.
		return nil //nolint:nilerr // siehe darüber
	}

	if os.Getenv("HOME") == wer.HomeDir {
		return fmt.Errorf("die Läufe arbeiten im echten Zuhause %s — "+
			"eigenesZuhauseFuerDenLauf hat nicht gegriffen", wer.HomeDir)
	}

	return nil
}

// schriftenVerweisen legt einen Verweis auf die Schriften dieses Rechners.
//
// **Ein Verweis und keine Kopie**: die Schriften wiegen zusammen mehr als
// alles andere in diesem Zuhause, und gelesen wird ohnehin nur.
//
// **Ein fehlendes `~/Library/Fonts` ist kein Fehler.** Nicht jeder Rechner hat
// eines; was dann fehlt, sind die Vorlagenschriften — und das sagt der
// Kanonlauf selbst, seit er [pruefeTextGemalt] fragt.
func schriftenVerweisen(zuhause string) error {
	echt, err := user.Current()
	if err != nil {
		return fmt.Errorf("das echte Zuhause finden: %w", err)
	}

	quelle := filepath.Join(echt.HomeDir, "Library", "Fonts")
	if !vorhanden(quelle) {
		return nil
	}

	if fehler := os.MkdirAll(filepath.Join(zuhause, "Library"), 0o700); fehler != nil {
		return fmt.Errorf("das Schriftverzeichnis anlegen: %w", fehler)
	}

	if fehler := os.Symlink(quelle, filepath.Join(zuhause, "Library", "Fonts")); fehler != nil {
		return fmt.Errorf("die Schriften verweisen: %w", fehler)
	}

	return nil
}

// vorhanden sagt, ob ein Pfad da ist.
//
// **Eine eigene Funktion, damit „nicht da" kein Fehler wird.** Ein `os.Stat`
// unmittelbar davor gäbe einen Fehler zurück, den der Aufrufer verwerfen
// müsste — und der Linter läse darin einen verschluckten Fehler.
func vorhanden(pfad string) bool {
	_, fehler := os.Stat(pfad)

	return fehler == nil
}

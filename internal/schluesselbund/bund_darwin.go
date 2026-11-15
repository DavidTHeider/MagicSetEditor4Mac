//go:build darwin

package schluesselbund

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Weg geht über `/usr/bin/security`, nicht über cgo.**
//
// # Warum kein cgo
//
// Dieses Projekt hat **keine** einzige cgo-Datei — der SQLite-Treiber ist
// bewusst reines Go. Ein `import "C"` hier zöge einen C-Werkzeugkasten in
// jeden Bau und in jeden Lauf mit `-race`, und zwar für **eine** Aufgabe.
//
// # Was das kostet, gehört dazu
//
// `security` bekommt seinen Befehl über die **Standardeingabe** (`-i`), nicht
// über die Argumentliste. Ein Geheimnis als Argument stünde in der
// Prozessliste — **gemessen am 2026-09-01**: `ps -ww -o args= -A` zeigt die
// Argumente jedes Prozesses desselben Nutzers (nachgestellt an einem `sleep`).
// Wie **lange** das Fenster bei einem `security`-Aufruf offen steht, ist
// **nicht** gemessen; ein Späher hat es in 250 Durchgängen nicht gefangen.
// Der Weg über die Standardeingabe macht die Frage gegenstandslos, und das
// ist der Grund, ihn zu nehmen — nicht ein gemessener Angriff.
//
// # Und warum das Geheimnis als Hex reist
//
// Weil `security` seine Befehlszeile selbst zerteilt — siehe [alsHex].

// befehl baut den Aufruf an `security`.
//
// **Das Programm steht als Literal da**, weil `internal/schichten` es
// verlangt: ein Programmname aus einer Veränderlichen ließe den Zustand des
// Programms entscheiden, was startet.
//
// **Und er bekommt eine Frist** (siehe [UeblicheFrist]): `security` fragt bei
// manchen Fehlschlägen nach einem Passwort und wartet dann. Der Rückgabewert
// gibt sie frei — wer ihn vergisst, hält einen Wecker offen.
func (b *Bund) befehl(ctx context.Context, zeile string) (
	gestellt *exec.Cmd, freigeben func(),
) {
	befristet, fertig := context.WithTimeout(ctx, b.frist)

	gestellt = exec.CommandContext(befristet, "/usr/bin/security", "-i")
	gestellt.Stdin = strings.NewReader(zeile + "\n")

	return gestellt, fertig
}

// ortsteil gibt den Bundpfad als Anhang einer Befehlszeile — oder nichts.
func (b *Bund) ortsteil() string {
	if b.ort == "" {
		return ""
	}

	return " " + b.ort
}

// pruefeOrt sagt, ob der benannte Bund überhaupt da ist.
//
// # Warum das vor jedem Aufruf steht
//
// **Weil `security` sonst den Menschen fragt.** Gemeldet am 2026-09-02 vom
// Nutzer dieses Rechners: beim Testlauf erschien ein Fenster von macOS —
//
//	Schlüsselbund nicht gefunden
//	Es wurde kein Schlüsselbund gefunden, um „konto" zu sichern.
//	[Abbrechen] [Auf Standard zurücksetzen]
//
// **Und „Auf Standard zurücksetzen" schreibt in den Anmeldebund.** Genau das
// ist geschehen: drei Einträge aus Testläufen lagen danach im echten Bund
// dieses Rechners. Ein Testlauf, der den Rechner verändert, ist keiner (§3) —
// und ein Programm, das im Hintergrund einen Systemdialog auslöst, hat die
// Kontrolle abgegeben.
//
// # Der Dialog ist das sichtbare Symptom, nicht das schlimmste
//
// **Bei der Gegenprobe am 2026-09-02 kam gar kein Dialog mehr** — `security`
// schrieb das Geheimnis **stillschweigend in den Anmeldebund** und meldete
// Erfolg, in dreißig Millisekunden. Nachgeprüft und wieder entfernt.
//
// Das ist der eigentliche Befund: ein Geheimnis, das für einen bestimmten
// Bund gedacht war, landet ohne ein Wort in einem anderen. Wer nur den Dialog
// abgestellt hätte, hätte den lauten Fall behoben und den leisen stehen
// lassen — §10: nicht der Fall, die Klasse.
//
// **Der Anmeldebund wird nicht geprüft** (`ort == ""`). Seinen Pfad kennt nur
// macOS; ihn zu erfragen wäre ein zweiter Unterprozess für einen Fall, den es
// auf einem eingerichteten Rechner nicht gibt. **Was hier nicht gehalten
// wird**, steht damit ausgesprochen da: fehlt der Anmeldebund selbst, kann
// dieser Weg den Dialog nicht verhindern.
func (b *Bund) pruefeOrt() error {
	if fehler := b.pruefeLauf(); fehler != nil {
		return fehler
	}

	return b.OrtFehlt()
}

// pruefeLauf weist den Anmeldebund ab, wenn ein Lauf ihn anfassen will.
//
// **Getrennt von [Bund.OrtFehlt], und das ist der Punkt.** `loesche`
// verschluckt einen fehlenden Bund mit Grund — wer löscht, will, dass nichts
// da ist, und bei einem Bund, den es nicht gibt, ist das schon der Fall. Für
// eine **Abweisung** gilt das nicht: sie heißt „du darfst nicht", nicht
// „schon erledigt". Ein Löschweg, der sie verschluckte, meldete Erfolg über
// etwas, das gar nicht stattgefunden hat.
func (b *Bund) pruefeLauf() error {
	if b.ort == "" && imLauf() {
		return fmt.Errorf(
			"%w: ein Lauf fasst den Anmeldebund dieses Rechners nicht an — "+
				"setze %s auf einen eigenen Bund", domain.ErrValidation, Bundort)
	}

	return nil
}

// imLauf sagt, ob dieses Programm ein Testbinärprogramm ist.
//
// # Warum das hier steht und nicht in einer Leitlinie
//
// **Weil die Leitlinie nicht getragen hat.** §3 verlangt seit jeher, dass ein
// Lauf weder Netz noch echtes Zuhause noch Schlüsselbund anfasst, und
// `MSE_SCHLUESSELBUND` steht dafür bereit. Am 2026-09-14 gemessen: **84**
// Zugriffe auf den Anmeldebund dieses Rechners aus acht Paketen, darunter
// **zwei Schreibvorgänge** von `peeridentitaet` — und macOS fragte den
// Menschen davor dreimal um Erlaubnis, mitten in einem Lauf.
//
// Eine Vorgabe, die sicher ist, schlägt eine Regel, an die sich jeder
// erinnern muss (§10). Wer den Anmeldebund in einem Lauf wirklich braucht,
// bekommt ihn nicht — er bekommt einen eigenen, und das ist die Absicht.
//
// # Warum über den Programmnamen
//
// **Weil `testing` nicht in den Produktivcode gehört.** Ein Import davon zöge
// Testflaggen in jedes Programm, das dieses Paket einbindet. `go test` baut
// sein Binärprogramm als `<paket>.test`; das ist die Auskunft, und sie kostet
// nichts.
//
// **Was es nicht fängt**: ein Kindprozess, den ein Lauf startet — `mse`
// heißt `mse` und nicht `mse.test`. Dort hilft nur die Umgebung, und
// `internal/flaechen` setzt sie seit demselben Tag von sich aus.
func imLauf() bool { return strings.HasSuffix(os.Args[0], ".test") }

// OrtFehlt ist [Bund.pruefeOrt] nach außen — für den Lauf, der belegt, dass
// der Anmeldebund durchgelassen wird, ohne ihn anzufassen.
func (b *Bund) OrtFehlt() error {
	if b.ort == "" {
		return nil
	}

	if _, err := os.Stat(b.ort); err != nil {
		return fmt.Errorf(
			"%w: den Schlüsselbund %q gibt es nicht", domain.ErrNotFound, b.ort)
	}

	return nil
}

func (b *Bund) schreibe(ctx context.Context, dienst, konto, geheimnis string) error {
	if err := b.pruefeOrt(); err != nil {
		return err
	}

	// **`-U` ersetzt einen vorhandenen Eintrag.** Ohne es läge ein zweiter
	// mit demselben Namen im Bund, und welcher gelesen wird, entschiede die
	// Reihenfolge.
	zeile := fmt.Sprintf("add-generic-password -a %s -s %s -w %s -U%s",
		konto, dienst, alsHex(geheimnis), b.ortsteil())

	var laerm bytes.Buffer

	gestellt, fertig := b.befehl(ctx, zeile)
	defer fertig()

	gestellt.Stderr = &laerm

	if err := gestellt.Run(); err != nil {
		return fmt.Errorf("das Geheimnis ablegen: %w — %s",
			err, strings.TrimSpace(laerm.String()))
	}

	return nil
}

func (b *Bund) lies(ctx context.Context, dienst, konto string) (string, error) {
	if err := b.pruefeOrt(); err != nil {
		return "", err
	}

	zeile := fmt.Sprintf("find-generic-password -a %s -s %s -w%s",
		konto, dienst, b.ortsteil())

	var (
		fund  bytes.Buffer
		laerm bytes.Buffer
	)

	gestellt, fertig := b.befehl(ctx, zeile)
	defer fertig()

	gestellt.Stdout = &fund
	gestellt.Stderr = &laerm

	if err := gestellt.Run(); err != nil {
		// **„gibt es nicht" ist eine fachliche Auskunft und kein Ausfall**
		// (§8). Wer sie nicht unterscheidet, meldet ein fehlendes Geheimnis
		// als kaputten Schlüsselbund.
		if strings.Contains(laerm.String(), "could not be found") {
			return "", fmt.Errorf(
				"%w: für %s liegt kein Geheimnis im Schlüsselbund",
				domain.ErrNotFound, dienst)
		}

		return "", fmt.Errorf("das Geheimnis lesen: %w — %s",
			err, strings.TrimSpace(laerm.String()))
	}

	return ausHex(fund.String())
}

func (b *Bund) loesche(ctx context.Context, dienst, konto string) error {
	// **Eine Abweisung ist kein erledigtes Löschen.**
	if fehler := b.pruefeLauf(); fehler != nil {
		return fehler
	}

	// **Ein Bund, den es nicht gibt, trägt nichts** — und wer löscht, will,
	// dass nichts da ist. Das ist dann schon der Fall.
	if b.OrtFehlt() != nil {
		return nil
	}

	zeile := fmt.Sprintf("delete-generic-password -a %s -s %s%s",
		konto, dienst, b.ortsteil())

	var laerm bytes.Buffer

	gestellt, fertig := b.befehl(ctx, zeile)
	defer fertig()

	gestellt.Stdout = &bytes.Buffer{}
	gestellt.Stderr = &laerm

	if err := gestellt.Run(); err != nil {
		if strings.Contains(laerm.String(), "could not be found") {
			return nil
		}

		return fmt.Errorf("das Geheimnis herausnehmen: %w — %s",
			err, strings.TrimSpace(laerm.String()))
	}

	return nil
}

// alsHex macht aus einem Geheimnis eine Zeichenkette aus [0-9a-f].
//
// # Warum überhaupt umkodiert wird
//
// **Weil `security` seine Befehlszeile selbst zerteilt.** Gemessen am
// 2026-09-01: ein Geheimnis mit **Leerzeichen** kam als **leer** zurück, eines
// mit **Gegenschrägstrich** verstümmelt („a\b" → „ab"). Hex hat keines der
// Zeichen, an denen ein Zerteiler entscheidet.
//
// **Und der Rückweg ist damit eindeutig.** Der naheliegende Weg — das
// Geheimnis als Hex **ablegen** (`-X`) — bricht anders: `find-generic-password
// -w` gibt dann bei Nicht-ASCII wieder Hex zurück und bei ASCII den Klartext,
// und was von beidem kam, ist nicht zu unterscheiden. Wer **immer** Hex
// ablegt, bekommt **immer** Hex zurück.
func alsHex(geheimnis string) string { return hex.EncodeToString([]byte(geheimnis)) }

// ausHex macht die Umkodierung rückgängig.
func ausHex(roh string) (string, error) {
	entpackt, err := hex.DecodeString(strings.TrimSpace(roh))
	if err != nil {
		return "", fmt.Errorf(
			"der Schlüsselbund gab etwas zurück, das kein Hex ist: %w", err)
	}

	return string(entpackt), nil
}

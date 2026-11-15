// **Eine Sicherung, die alles trägt** —
// #1162.
//
// # Was hier steht und was nicht
//
// **Ein Blatt** (§2): dieses Paket kennt `domain` und sonst nichts. **Was** in
// eine Sicherung gehört, entscheidet der Aufrufer — hier steht, wie ein Archiv
// aussieht, wie es entsteht und wie man ihm ansieht, ob es noch das ist, das
// geschrieben wurde.
//
// # Warum ein Archiv und nicht eine Datenbank
//
// Bis 0.2.0 schrieb „Bestand sichern" genau eine `.db`. Die eingesetzten
// Kartenbilder liegen daneben unter `beiwerk/`, und
// [0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md) sagt
// **selbst**, ein Set sei ohne dieses Verzeichnis unvollständig.
//
// **Auf demselben Rechner fällt das nie auf** — das alte `beiwerk/` liegt ja
// noch da. Sichtbar wird es erst nach einem Rechnerwechsel, also genau dann,
// wenn die Sicherung das Einzige ist, was bleibt. Angestrichen vom Audit für
// 0.3.0 (Z6-01).
//
// # Warum ZIP
//
// Weil das Format schon im Haus ist: `.mse-set`, `.mse-game` und jedes
// bezogene Paket sind ZIPs, und `internal/msepaket` liest sie. Ein eigenes
// Format wäre eine zweite Sache, die jemand verstehen muss.
//
// # Warum ein Manifest
//
// **Ein Archiv sagt, was drin ist. Eine Sicherung sagt, ob es noch stimmt.**
// Das Manifest führt je Teil Größe und SHA-256; ohne es fiele eine
// beschädigte Sicherung erst auf, wenn jemand sie braucht.

// Package sicherung schreibt und liest das Sicherungsarchiv.
package sicherung

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Unfertigvorsatz und Unfertigendung bilden den Namen, unter dem ein Archiv
// entsteht: `.sicherung-<zufall>.unfertig`. **Die Restepflege erkennt es an
// denselben Zeichen** (`service.RestePflegen`, Zweitblick über Phase 4,
// Befund 12) — ein Name an zwei Stellen liefe auseinander.
const (
	Unfertigvorsatz = ".sicherung-"
	Unfertigendung  = ".unfertig"
)

// Manifestname ist der Eintrag, der eine Sicherung zur Sicherung macht.
const Manifestname = "sicherung.json"

// Endung ist die Endung einer Sicherung.
//
// **Nicht `.mse-…`**: dieser Namensraum gehört Magic Set Editor, und eine
// Sicherung ist kein MSE-Paket. Nicht `.zip`: das verschweigt, was es ist.
const Endung = ".msesicherung"

// Teil ist eine Datei, die in das Bündel soll.
type Teil struct {
	// Name ist der Weg **im Archiv**, mit Schrägstrichen.
	Name string

	// Quelle ist die Datei auf der Platte.
	Quelle string

	// Inhalt ist der Text, wenn es keine Datei gibt — Werkstück O10 von
	// Phase 5.2 (#1434).
	//
	// **Genau einer von beiden.** Eine Sicherung packt Dateien, ein
	// Diagnosepaket packt, was gerade zusammengestellt wurde; ein Umweg über
	// eine Nachbardatei schriebe den Inhalt zweimal auf die Platte und ließe
	// beim Abbruch die Hälfte davon liegen. Ist [Teil.Quelle] leer, gilt
	// dieses Feld — auch dann, wenn es selbst leer ist: ein leerer Teil ist
	// ein Teil ohne Inhalt und kein Fehler.
	Inhalt []byte
}

// Angaben sind die Angaben über die Sicherung selbst.
type Angaben struct {
	// Fassung ist die Fassung der Anwendung, die sie geschrieben hat.
	//
	// **Sie steht hier, damit ein Rückspielweg entscheiden kann**, ob er eine
	// Sicherung aus einer **neueren** Fassung annimmt — und
	// `app.fassungPruefen` tut das.
	//
	// # Der Migrationsstand steht bewusst nicht daneben
	//
	// **Und das ist eine Abweichung von
	// #1162**,
	// die dort „Manifest mit Fassung, **Schemastand**, Zeitpunkt und
	// Prüfsummen je Teil" verlangt. Angestrichen vom Nachlauf des
	// Zweitblicks, Befund 8; hier steht der Grund (§15, Satz 7).
	//
	// Der Stand steht in der Datenbank **im Archiv**, und dort ist er die
	// Wahrheit: `storage.PruefeSicherung` liest ihn von dort, und die
	// Rückfrage zeigt ihn. Ihn zusätzlich ins Manifest zu schreiben schüfe
	// eine zweite Angabe über dieselbe Sache — mit der Frage, was gilt, wenn
	// sie auseinanderlaufen. Bei einer veränderten Sicherung tun sie das, und
	// dann wäre der Eintrag im Manifest genau die Angabe, der man nicht
	// glauben darf.
	//
	// **Die Fassung hat dieses Problem nicht**: sie beschreibt das **Archiv**
	// und steht nirgendwo sonst.
	Fassung string `json:"fassung"`

	// Erstellt ist der Zeitpunkt.
	Erstellt time.Time `json:"erstellt"`

	// Pakete nennt die Vorlagenpakete des Bestands — OPS-06.AK-01
	// (#1723):
	// die eigenen **liegen darin**, die bezogenen **stehen nur hier**, mit
	// Herkunft und Prüfsumme, damit ein Zurückspielen sagen kann, was fehlt.
	Pakete []Paketverweis `json:"pakete,omitempty"`
}

// Paketverweis nennt ein Paket des gesicherten Bestands.
type Paketverweis struct {
	// Name ist der Name auf der Platte, etwa `magic-m15.mse-style`.
	Name       string `json:"name"`
	Herkunft   string `json:"herkunft,omitempty"`
	Pruefsumme string `json:"pruefsumme,omitempty"`

	// Enthalten sagt, ob es in der Sicherung liegt.
	Enthalten bool `json:"enthalten"`
}

// Beschreibung ist ein gelesenes Manifest.
type Beschreibung struct {
	Angaben

	Teile []Teilangabe `json:"teile"`
}

// Teilangabe beschreibt einen Teil der Sicherung.
type Teilangabe struct {
	Name    string `json:"name"`
	Groesse int64  `json:"groesse"`
	Summe   string `json:"sha256"`
}

// IstArchiv sagt, ob eine Datei eine Sicherung dieses Formats sein kann.
//
// **Über die ersten vier Bytes.** Ein ZIP fängt mit `PK\x03\x04` an, eine
// SQLite-Datei mit `SQLite format 3`. Die Unterscheidung ist nötig, damit eine
// Sicherung aus 0.1.0 oder 0.2.0 — eine **nackte Datenbank** — weiterhin
// zurückgespielt werden kann. Sie wertlos zu machen wäre eine schlechtere
// Behebung als der Fehler.
//
// **Sie sagt nicht, ob es eine gültige Sicherung ist** — nur, welchen Weg man
// versuchen muss. Ob ein Manifest darin steht, sagt [Entpacken].
func IstArchiv(pfad string) bool {
	datei, err := os.Open(pfad) // #nosec G304 -- der Pfad ist die Sicherung, die der Aufrufer benennt
	if err != nil {
		return false
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	kopf := make([]byte, 4)
	if _, fehler := io.ReadFull(datei, kopf); fehler != nil {
		return false
	}

	return string(kopf) == "PK\x03\x04"
}

// Packen schreibt die Teile samt Manifest nach `ziel`.
//
// **Das Ziel wird reserviert, nicht überschrieben** — dieselbe Überlegung wie
// bei `storage.Sichern`: eine Sicherung, die eine andere wegnimmt, ist die
// Falle, gegen die es sie gibt.
func Packen(ziel string, teile []Teil, angaben Angaben) error {
	return PackenAls(ziel, Manifestname, teile, angaben)
}

// PackenAls schreibt ein Bündel mit **eigenem** Manifestnamen — Werkstück O10
// von Phase 5.2 (#1434).
//
// # Warum dieses Paket zwei Bündelarten schreibt
//
// **Weil die Mechanik dieselbe ist und die Zusage auch**: daneben schreiben,
// die Summe je Teil rechnen, ein Manifest anhängen, unteilbar veröffentlichen
// — und **kein bestehendes Bündel wegnehmen**. Was sich unterscheidet, ist der
// Name des Manifests und damit die Erkennbarkeit des Formats: eine Sicherung
// spielt man zurück, ein Diagnosepaket verschickt man, und ein `sicherung.json`
// in einem Diagnosepaket wäre die Einladung, das eine für das andere zu halten.
//
// **Ein zweiter Schreiber daneben wäre die schlechtere Antwort** (§15): zwei
// Stellen, an denen jemand die Prüfsumme oder das Veröffentlichen ändern
// müsste, und eine davon würde vergessen.
func PackenAls(ziel, manifestname string, teile []Teil, angaben Angaben) error {
	_, fehler := packenIn(ziel, manifestname, teile, angaben, offeneHuelle, offenGelesen)

	return fehler
}

// offeneHuelle schreibt das Archiv, wie es ist.
func offeneHuelle(nach io.Writer, schreiben func(io.Writer) error) error {
	return schreiben(nach)
}

// packenIn schreibt das Bündel durch eine Hülle — offen oder verschlüsselt
// ([PackenVerschluesselt]). **Eine Mechanik für beide**, aus demselben Grund
// wie bei [PackenAls]: eine zweite Stelle, die veröffentlicht, würde
// vergessen.
func packenIn(ziel, manifestname string, teile []Teil, angaben Angaben,
	huelle func(io.Writer, func(io.Writer) error) error, lesen lesart,
) (Abdruck, error) {
	// **Daneben geschrieben, dann verknüpft.**
	//
	// Bis zum Zweitblick über
	// #1162
	// entstand die Sicherung unmittelbar unter ihrem endgültigen Namen. Ein
	// Absturz, ein erzwungenes Beenden, ein Stromausfall — und eine halbe
	// Datei blieb dort liegen, unter der richtigen Endung. Beim nächsten Mal
	// fiel es doppelt auf: sie war unbrauchbar, **und** ein neuer Versuch
	// unter demselben Namen scheiterte am belegten Ziel.
	//
	// **`os.Link` und nicht `os.Rename`**: Umbenennen ist ebenso unteilbar,
	// **überschreibt** aber. Die Zusage „eine Sicherung nimmt keine andere
	// weg" hinge dann an einer Prüfung davor — und zwischen Prüfung und
	// Umbenennen liegt ein Fenster. `Link` scheitert selbst mit `EEXIST`.
	unterwegs, err := os.CreateTemp(filepath.Dir(ziel), Unfertigvorsatz+"*"+Unfertigendung)
	if err != nil {
		return Abdruck{}, fmt.Errorf("die Sicherung anlegen: %w", err)
	}

	defer func() { _ = os.Remove(unterwegs.Name()) }() //nolint:errcheck // Nachbardatei

	// **Schreiben und Schließen auf einem Weg**, wie in `storage.Sichern` und
	// aus demselben Grund: ein Griff auf eine eben angelegte Datei scheitert
	// beim Schließen nicht auf eine Art, die ein Test herbeiführen könnte —
	// und ein Zweig, den kein Test betritt, ist eine Behauptung darüber, was
	// im Ernstfall geschieht (§3). Beide enden ohnehin gleich: melden.
	fehler := huelle(unterwegs, func(nach io.Writer) error {
		return hineinschreiben(nach, manifestname, teile, angaben)
	})

	// **Auf die Platte, bevor verknüpft wird.** `Close` sagt, dass der Puffer
	// des Prozesses leer ist — nicht, dass etwas dauerhaft liegt. Was `Sync`
	// nicht kann, steht bei [Packen].
	if syncfehler := unterwegs.Sync(); fehler == nil {
		fehler = syncfehler
	}

	// **Und zurückgelesen, bevor sie ihren Namen bekommt** (OPS01, #2347) —
	// erst dann gibt es einen Abdruck, an dem ein Vermerk sie wiedererkennt.
	var abdruck Abdruck

	if fehler == nil {
		var prueffehler error
		if abdruck, prueffehler = zuruecklesen(unterwegs, manifestname, lesen); prueffehler != nil {
			_ = unterwegs.Close() //nolint:errcheck // verworfen ist verworfen

			return Abdruck{}, domain.MitGrund(domain.ErrSicherungVerworfen, prueffehler,
				"versuche es noch einmal, oder sichere an einen anderen Ort")
		}
	}

	if schliessfehler := unterwegs.Close(); fehler == nil {
		fehler = schliessfehler
	}

	if fehler != nil {
		return Abdruck{}, fmt.Errorf("die Sicherung schreiben: %w", fehler)
	}

	if fehler := veroeffentlichen(unterwegs.Name(), ziel, os.Link); fehler != nil {
		return Abdruck{}, fehler
	}

	return abdruck, nil
}

// veroeffentlichen gibt der fertigen Nachbardatei ihren endgültigen Namen.
//
// # Zwei Wege, und der zweite ist der schlechtere
//
// **`os.Link`** ist unteilbar und scheitert selbst mit `EEXIST` — beides
// zusammen gibt es sonst nicht. Auf APFS und HFS+ ist er der Weg.
//
// **FAT und exFAT kennen keine harten Verweise**, und macOS richtet externe
// Datenträger ausdrücklich so ein, wenn sie auch an Windows gehen sollen. Ein
// USB-Stick ist der wahrscheinlichste Ort einer Sicherung überhaupt; dort
// hätte `Link` allein bedeutet, dass alles geschrieben wird und der letzte
// Schritt scheitert (Befund 4 des Nachlaufs über #1162, nachgelesen am
// 2026-08-26 bei Apple und Microsoft).
//
// Der Rückfall legt die Datei mit `O_EXCL` an und kopiert. **Er ist nicht
// unteilbar**: stirbt der Prozess dabei, bleibt eine halbe Sicherung unter dem
// endgültigen Namen stehen — genau das, was der erste Weg verhindert. Ein
// Dateisystem ohne harte Verweise hat aber auch keine andere Möglichkeit, und
// eine Sicherung, die auf einem USB-Stick gar nicht erst entsteht, ist
// schlechter als eine, deren seltener Abbruch sichtbar bleibt.
// **Der Verknüpfer kommt als Wert und nicht aus einem Paketfeld.** Ein Feld
// wäre eine Naht, die alle Läufe teilen — und Läufe mit `t.Parallel()` nähmen
// sie einander weg, wie es der Riegel in `internal/app` schon einmal getan
// hat. Ein Aufrufer reicht `os.Link`, ein Test reicht etwas anderes.
func veroeffentlichen(unterwegs, ziel string, verknuepfen func(alt, neu string) error) error {
	fehler := verknuepfen(unterwegs, ziel)
	if fehler == nil {
		return nil
	}

	if errors.Is(fehler, os.ErrExist) {
		return fmt.Errorf("%w: %s liegt schon da — eine Sicherung "+
			"überschreibt nichts, wähle einen anderen Namen",
			domain.ErrConflict, filepath.Base(ziel))
	}

	if !errors.Is(fehler, syscall.ENOTSUP) && !errors.Is(fehler, syscall.EPERM) &&
		!errors.Is(fehler, syscall.EXDEV) && !errors.Is(fehler, syscall.EOPNOTSUPP) {
		return fmt.Errorf("die Sicherung an ihren Ort legen: %w", fehler)
	}

	return kopieren(unterwegs, ziel)
}

// kopieren ist der Rückfall für Dateisysteme ohne harte Verweise.
func kopieren(unterwegs, ziel string) error {
	// #nosec G304 -- der Pfad ist die Sicherung, die der Aufrufer benennt
	quelle, err := os.Open(unterwegs)
	if err != nil {
		return fmt.Errorf("die Sicherung an ihren Ort legen: %w", err)
	}

	defer func() { _ = quelle.Close() }() //nolint:errcheck // nur gelesen

	// #nosec G304 -- derselbe Pfad wie oben
	datei, err := os.OpenFile(ziel, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%w: %s liegt schon da — eine Sicherung "+
				"überschreibt nichts, wähle einen anderen Namen",
				domain.ErrConflict, filepath.Base(ziel))
		}

		return fmt.Errorf("die Sicherung an ihren Ort legen: %w", err)
	}

	// **Kopieren, Durchschreiben und Schließen auf einem Weg** — sie scheitern
	// an derselben Ursache, und drei Zweige daraus wären zwei, die kein Test
	// betritt (§3).
	_, err = io.Copy(datei, quelle)
	if syncfehler := datei.Sync(); err == nil {
		err = syncfehler
	}

	if schliessfehler := datei.Close(); err == nil {
		err = schliessfehler
	}

	if err != nil {
		return errors.Join(
			fmt.Errorf("die Sicherung an ihren Ort legen: %w", err),
			os.Remove(ziel))
	}

	return nil
}

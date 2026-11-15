package sicherung

// **Eine Sicherung wird zurückgelesen, bevor sie ihren Namen bekommt** —
// OPS01 (#2347),
// dieselbe Regel, die F05 (#2344) für die Datenbank darin eingeführt hat.
//
// # Warum hier und nicht im Dienst
//
// **Weil nur hier die Nachbardatei noch keinen Namen hat.** Eine Prüfung
// danach fände eine verdorbene Sicherung schon unter dem gewählten Namen, und
// sie wegzunehmen hieße, eine Datei zu löschen, die jemand gesehen haben kann.
// Davor entscheidet sie, ob es den Namen überhaupt gibt.
//
// # Was zurückgelesen heißt
//
// **Die Datei, wie sie nach `fsync` liegt**: einmal ganz für Größe und
// SHA-256 — das ist der [Abdruck], an dem ein Vermerk sie wiedererkennt —,
// und einmal als Archiv, jeder Teil gegen Größe und Summe aus seinem
// Manifest, eine verschlüsselte durch age hindurch. **Nicht zurückgelesen
// wird am Plattencache vorbei**: macOS kennt dafür keinen Weg, der nicht das
// Laufwerk selbst anfasst. Was das Zurücklesen findet, ist ein Schreib- oder
// Programmfehler — eine Platte, die später verdirbt, findet es nicht.

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"filippo.io/age"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Abdruck ist, woran eine geschriebene Sicherung wiederzuerkennen ist —
// Größe und SHA-256 über die Datei, wie sie auf der Platte liegt.
type Abdruck struct {
	Groesse int64
	Summe   string
}

// lesart öffnet die geschriebene Datei als Archiv — offen oder durch die
// Verschlüsselung hindurch.
type lesart func(datei *os.File, groesse int64) (io.ReaderAt, int64, error)

// offenGelesen liest ein offenes Archiv, wie es liegt.
func offenGelesen(datei *os.File, groesse int64) (io.ReaderAt, int64, error) {
	return datei, groesse, nil
}

// chiffreGelesen entschlüsselt mit der Identität des Kennworts, mit dem eben
// verschlüsselt wurde.
func chiffreGelesen(identitaet age.Identity) lesart {
	return func(datei *os.File, groesse int64) (io.ReaderAt, int64, error) {
		return age.DecryptReaderAt(datei, groesse, identitaet)
	}
}

// PackenGeprueft schreibt eine Sicherung — mit Kennwort, wenn eines dasteht
// — und gibt den Abdruck der zurückgelesenen Datei.
func PackenGeprueft(ziel string, teile []Teil, angaben Angaben, kennwort string) (Abdruck, error) {
	if kennwort == "" {
		return packenIn(ziel, Manifestname, teile, angaben, offeneHuelle, offenGelesen)
	}

	return packenVerschluesselt(ziel, teile, angaben, kennwort)
}

// zuruecklesen liest die geschriebene Datei ganz und prüft jeden Teil gegen
// sein Manifest.
//
// **Gelesen wird über den Griff, mit dem geschrieben wurde** — nach `fsync`
// liefert ein neu geöffneter Griff dieselben Seiten aus demselben Cache, und
// ein Zweig für „lässt sich nicht wieder öffnen" prüfte nichts.
func zuruecklesen(datei *os.File, manifestname string, lesen lesart) (Abdruck, error) {
	rechner := sha256.New()

	groesse, lesefehler := io.Copy(rechner, io.NewSectionReader(datei, 0, HoechstensGesamt+1))

	klar, laenge, oeffnungsfehler := lesen(datei, groesse)
	if fehler := errors.Join(lesefehler, oeffnungsfehler); fehler != nil {
		return Abdruck{}, fmt.Errorf("die Sicherung zurücklesen: %w", fehler)
	}

	archiv, fehler := zip.NewReader(klar, laenge)
	if fehler != nil {
		return Abdruck{}, fmt.Errorf("die Sicherung als Archiv lesen: %w", fehler)
	}

	_, gewollt, nachName, fehler := gewollteTeile(archiv, manifestname, HoechstensManifest, nil)
	if fehler != nil {
		return Abdruck{}, fehler
	}

	for _, angabe := range gewollt {
		if fehler := teilPruefen(nachName[angabe.Name], angabe); fehler != nil {
			return Abdruck{}, fehler
		}
	}

	return Abdruck{Groesse: groesse, Summe: hex.EncodeToString(rechner.Sum(nil))}, nil
}

// teilPruefen liest einen Teil und vergleicht ihn mit seiner Angabe —
// **höchstens ein Byte mehr** als angekündigt, wie beim Auspacken.
//
// **Öffnen, Lesen und Vergleichen enden gleich**: in „stimmt nicht". Ein
// Teil, dessen Kompression `zip` nicht kennt, und einer mit falscher Summe
// sind für eine eben geschriebene Sicherung dasselbe — sie ist keine.
func teilPruefen(eintrag *zip.File, angabe Teilangabe) error {
	rechner := sha256.New()

	var gelesen int64

	offen, fehler := eintrag.Open()
	if fehler == nil {
		gelesen, fehler = io.Copy(rechner, io.LimitReader(offen, angabe.Groesse+1))
		fehler = errors.Join(fehler, offen.Close())
	}

	if fehler != nil || gelesen != angabe.Groesse || hex.EncodeToString(rechner.Sum(nil)) != angabe.Summe {
		return errors.Join(fmt.Errorf("%w: %s stimmt zurückgelesen nicht mit dem Manifest überein",
			domain.ErrConflict, angabe.Name), fehler)
	}

	return nil
}

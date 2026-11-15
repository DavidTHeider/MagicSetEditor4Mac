package sicherung

// **Eine Sicherung mit Kennwort** — OPS-06.AK-02
// (#1723).
//
// # Warum age und kein eigenes Format
//
// **Weil „portabel" zwei Dinge heißt**: die Sicherung hängt an keinem
// Schlüssel dieses Rechners — anders als der verschlüsselte Bestand, dessen
// Schlüssel im Schlüsselbund liegt ([0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md)) —,
// **und** sie lässt sich ohne diese Anwendung öffnen. `age -d` mit demselben
// Kennwort gibt das ZIP heraus, auf jedem Rechner, für den es age gibt. Ein
// eigenes Format aus AES und einer Schlüsselableitung wäre die erste
// Eigenschaft ohne die zweite, und eine zweite Stelle, an der sich
// Kryptographie falsch zusammensetzen ließe.
//
// Das Kennwort wird mit **scrypt** zum Schlüssel (age, `scrypt.go`,
// Arbeitsfaktor 2^18, rund eine Sekunde); der Inhalt ist mit
// ChaCha20-Poly1305 in Blöcken zu 64 KiB **authentifiziert** verschlüsselt.
// Eine veränderte Sicherung fällt deshalb beim Entschlüsseln auf, und nicht
// erst an der Prüfsumme eines Teils.
//
// # Kein Klartext auf dem Weg
//
// **Das ZIP entsteht im Strom**, durch die Hülle hindurch, und wird beim
// Zurückspielen über [age.DecryptReaderAt] gelesen — es liegt zu keinem
// Zeitpunkt entschlüsselt auf der Platte. Was herauskommt, sind die Teile, die
// der Aufrufer ohnehin auspackt.

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"filippo.io/age"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Chiffrekopf ist der Anfang jeder verschlüsselten Sicherung.
const Chiffrekopf = "age-encryption.org/v1\n"

// vorgabefaktor ist der Arbeitsfaktor von age selbst, 2^18.
const vorgabefaktor = 18

var arbeitsfaktor atomic.Int32

func init() { arbeitsfaktor.Store(vorgabefaktor) }

// ArbeitsfaktorSetzen setzt den Aufwand der Schlüsselableitung — **für
// Läufe**, die nicht je Sicherung eine Sekunde warten sollen. Die Rückgabe
// stellt die Vorgabe wieder her.
func ArbeitsfaktorSetzen(logN int) func() {
	arbeitsfaktor.Store(int32(logN)) // #nosec G115 -- ein Faktor zwischen 1 und 30

	return func() { arbeitsfaktor.Store(vorgabefaktor) }
}

// IstVerschluesselt sagt, ob eine Datei eine Sicherung mit Kennwort sein
// kann — über ihren Kopf, wie [IstArchiv].
func IstVerschluesselt(pfad string) bool {
	datei, err := os.Open(pfad) // #nosec G304 -- die Sicherung, die der Aufrufer benennt
	if err != nil {
		return false
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	kopf := make([]byte, len(Chiffrekopf))
	_, err = io.ReadFull(datei, kopf)

	return err == nil && string(kopf) == Chiffrekopf
}

// PackenVerschluesselt schreibt eine Sicherung mit Kennwort — reserviert und
// nicht überschreibend wie [Packen].
func PackenVerschluesselt(ziel string, teile []Teil, angaben Angaben, kennwort string) error {
	_, fehler := packenVerschluesselt(ziel, teile, angaben, kennwort)

	return fehler
}

// packenVerschluesselt ist [PackenVerschluesselt] mit dem Abdruck.
func packenVerschluesselt(ziel string, teile []Teil, angaben Angaben, kennwort string) (Abdruck, error) {
	empfaenger, identitaet, fehler := chiffrepaar(kennwort)
	if fehler != nil {
		return Abdruck{}, fehler
	}

	return packenIn(ziel, Manifestname, teile, angaben, chiffrehuelle(empfaenger), chiffreGelesen(identitaet))
}

// chiffrepaar gibt Empfänger und Identität eines Kennworts — den einen zum
// Verschlüsseln mit dem Arbeitsfaktor dieses Laufs, die andere zum
// Zurücklesen (OPS01, #2347). **Beide scheitern an demselben**, einem leeren
// Kennwort.
func chiffrepaar(kennwort string) (*age.ScryptRecipient, *age.ScryptIdentity, error) {
	empfaenger, err := age.NewScryptRecipient(kennwort)
	identitaet, identitaetsfehler := age.NewScryptIdentity(kennwort)

	if errors.Join(err, identitaetsfehler) != nil {
		return nil, nil, domain.NewFieldError("kennwort", domain.SchluesselSicherungskennwortKurz,
			"Ohne Kennwort lässt sich nicht verschlüsseln.")
	}

	empfaenger.SetWorkFactor(int(arbeitsfaktor.Load()))
	identitaet.SetMaxWorkFactor(max(vorgabefaktor, int(arbeitsfaktor.Load())))

	return empfaenger, identitaet, nil
}

// chiffrehuelle schreibt das Archiv durch age hindurch.
func chiffrehuelle(empfaenger age.Recipient) func(io.Writer, func(io.Writer) error) error {
	return func(nach io.Writer, schreiben func(io.Writer) error) error {
		huelle, fehler := age.Encrypt(nach, empfaenger)
		if fehler != nil {
			return fmt.Errorf("die Verschlüsselung beginnen: %w", fehler)
		}

		// **Erst das Schließen schreibt den letzten Block** — ohne es wäre
		// die Sicherung abgeschnitten und nicht zu öffnen.
		return errors.Join(schreiben(huelle), huelle.Close())
	}
}

// EntpackenVerschluesselt legt die Teile einer Sicherung mit Kennwort heraus
// — wie [EntpackenNur], nach dem Entschlüsseln.
//
// **Ein falsches Kennwort ist ein Befund am Feld** und keine beschädigte
// Datei: wer es sieht, soll wissen, dass er es noch einmal eingeben kann.
func EntpackenVerschluesselt(pfad, verzeichnis, kennwort string, nur ...string) (Beschreibung, error) {
	identitaet, err := age.NewScryptIdentity(kennwort)
	if err != nil {
		return Beschreibung{}, domain.NewFieldError("kennwort", domain.SchluesselSicherungskennwortFehlt,
			"Diese Sicherung ist verschlüsselt — sie braucht ihr Kennwort.")
	}

	// **Der Kopf einer fremden Sicherung bestimmt nicht, wie lange gerechnet
	// wird** (QA-06c, #1976). Ohne Deckel nimmt age bis 2^22 an — „15s on a
	// modern machine" und mit r = 8 vier GiB, **bevor** das Kennwort geprüft
	// ist (`scrypt.go`, Zeile 129, age v1.3.2). Diese Anwendung und `age -p`
	// schreiben 2^18; mehr stammt von keinem der beiden.
	identitaet.SetMaxWorkFactor(max(vorgabefaktor, int(arbeitsfaktor.Load())))

	datei, err := os.Open(pfad) // #nosec G304 -- die Sicherung, die der Aufrufer benennt
	if err != nil {
		return Beschreibung{}, domain.MitGrund(domain.ErrValidation, err,
			"die Sicherung lässt sich nicht öffnen — wähle eine andere Datei")
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	// **Ein Fehler aus `Stat` nach einem geglückten `Open` ist hier dasselbe
	// wie eine leere Datei**: age weist beide als unlesbar ab, und ein
	// eigener Zweig dafür wäre einer, den kein Test herstellen kann (§3).
	angaben, _ := datei.Stat() //nolint:errcheck // siehe oben

	var groesse int64
	if angaben != nil {
		groesse = angaben.Size()
	}

	klar, laenge, err := age.DecryptReaderAt(datei, groesse, identitaet)

	var falsch *age.NoIdentityMatchError
	if errors.As(err, &falsch) {
		return Beschreibung{}, domain.NewFieldError("kennwort", domain.SchluesselSicherungskennwortFalsch,
			"Das Kennwort passt nicht zu dieser Sicherung.")
	}

	if err != nil {
		return Beschreibung{}, domain.MitGrund(domain.ErrValidation, err,
			"die Sicherung lässt sich nicht entschlüsseln — sie ist beschädigt")
	}

	archiv, err := zip.NewReader(klar, laenge)
	if err != nil {
		return Beschreibung{}, domain.MitGrund(domain.ErrValidation, err,
			"in der Sicherung liegt kein lesbares Archiv — sie ist beschädigt")
	}

	return entpackenAus(archiv, verzeichnis, Manifestname, nur...)
}

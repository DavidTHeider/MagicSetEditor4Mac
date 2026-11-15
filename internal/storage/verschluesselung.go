// **Der Bestand liegt verschlüsselt auf der Platte** — Werkstück 14 von Phase
// 4.12 (#1351),
// entschieden in
// [0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md).
//
// # Was hier steht und was nicht
//
// Hier steht das **Handwerk**: würfeln, umschreiben, wiedererkennen, öffnen.
// **Ob** verschlüsselt wird, wann und was dabei mit der laufenden Verbindung
// geschieht, entscheidet `service.Bestandsweg` — das ist ein Lebenszyklus und
// keine Speicherfrage (§2).

package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	// **Die Chiffre meldet sich in ihrem `init()` an.** Ohne diesen Import
	// gäbe es sie nicht — SQLite antwortet dann „no such vfs: adiantum", und
	// zwar erst beim Schreiben, nicht beim Bauen.
	_ "github.com/ncruces/go-sqlite3/vfs/adiantum"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Chiffrehuelle ist das VFS, das die Verschlüsselung leistet.
//
// **Adiantum und nicht `xts`** — die Begründung und was beide *nicht* zusagen,
// steht im Vermerk. Der Name kommt aus `vfs/adiantum`, das ihn in seinem
// `init()` anmeldet.
const Chiffrehuelle = "adiantum"

// Klarhuelle ist das VFS ohne Chiffre — der gewöhnliche Dateizugriff.
//
// # Warum es beim Entschlüsseln ausdrücklich dastehen muss
//
// **Weil `VACUUM INTO` ohne `vfs=` das VFS der Hauptdatenbank erbt.** Beim
// Entschlüsseln ist das genau die Chiffre, aus der herausgeschrieben werden
// soll — das Ziel entstünde also wieder verschlüsselt, und zwar mit einem
// Schlüssel, den niemand kennt.
//
// **Gefunden hat es der Flächenlauf am 2026-09-06**: das Einschalten glückte,
// das Ausschalten endete mit „unable to open database file". Ein Ausschalten,
// das *scheinbar* glückt und eine unlesbare Datei hinterlässt, wäre schlimmer
// gewesen.
//
// Der Name kommt aus dem Treiber: `vfs.Find` gibt für `""` **und** für `"os"`
// den gewöhnlichen Dateizugriff (`vfs/registry.go`, Zeile 17 in v0.35.4), und
// beide Namen sind für eine eigene Anmeldung gesperrt. In einer Adresse muss
// er ausgeschrieben stehen, denn ein leeres `vfs=` wäre kein Name.
const Klarhuelle = "os"

// Klartextmarke sind die Bytes, mit denen eine **un**verschlüsselte
// SQLite-Datei beginnt.
//
// **Beleg:** „Every valid SQLite database file begins with the following 16
// bytes (in hex): 53 51 4c 69 74 65 20 66 6f 72 6d 61 74 20 33 00" —
// https://www.sqlite.org/fileformat.html#the_database_header, nachgelesen am
// 2026-09-06.
const Klartextmarke = "SQLite format 3\x00"

// Schluessellaenge ist die Länge eines Bestandsschlüssels in Bytes.
//
// 32 Bytes, weil die Voreinstellung von Adiantum einen 32-Byte-Schlüssel
// nimmt; als Hex sind das 64 Zeichen.
const Schluessellaenge = 32

// NeuerSchluessel würfelt einen Bestandsschlüssel.
//
// **Gewürfelt und nicht abgeleitet.** Die Hülle nähme auch einen `textkey` und
// liefe ihn durch Argon2id — bequemer zu merken, und genau deshalb schwächer
// als 32 Zufallsbytes. Gemerkt werden muss hier nichts: der Schlüssel liegt im
// Schlüsselbund, den macOS hütet.
//
// # Ohne Fehler, und das ist nachgelesen
//
// Der erste Anlauf gab einen zurück und prüfte ihn — ein Zweig, den kein Test
// betreten kann, ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
// **`crypto/rand.Read` sagt es selbst**: „It never returns an error, and
// always fills b entirely" (nachgelesen am 2026-09-06 mit `go doc`). Es stürzt
// ab, statt zu scheitern.
func NeuerSchluessel() string {
	roh := make([]byte, Schluessellaenge)

	// Der Rückgabewert wird nicht geprüft — siehe oben.
	_, _ = rand.Read(roh) //nolint:errcheck // kann nach der Doku nicht scheitern

	return hex.EncodeToString(roh)
}

// IstVerschluesselt sagt, ob eine Bestandsdatei verschlüsselt ist.
//
// # Warum am Dateikopf und nicht an einer Merkdatei
//
// **Weil eine Merkdatei verlorengeht, sobald jemand nur die `.db` kopiert** —
// und dann stünde die Anwendung vor einer Datei, von der sie das Falsche
// glaubt. Der Kopf reist mit dem Inhalt.
//
// # Was sie nicht sagt
//
// **Sie unterscheidet nicht zwischen „verschlüsselt" und „kaputt".** Beides
// sieht gleich aus, und das ist keine Nachlässigkeit, sondern die Lage: eine
// Chiffre, der man ansieht, dass sie eine ist, verrät etwas. Wer den Schlüssel
// hat, merkt den Unterschied beim Öffnen.
//
// **Eine leere Datei gilt als unverschlüsselt** — so entsteht ein neuer
// Bestand, und SQLite schreibt den Kopf erst mit der ersten Seite.
//
// **Eine Datei, die es nicht gibt, ebenso.** Das ist kein Fehler: `Oeffnen`
// legt sie an.
func IstVerschluesselt(pfad string) (bool, error) {
	// #nosec G304 -- der Pfad ist die Bestandsdatei, die der Aufrufer benennt;
	// genau sie anzusehen ist der Zweck dieser Funktion. Gelesen werden
	// sechzehn Bytes, geschrieben wird nichts.
	datei, err := os.Open(pfad)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("den Bestand %q ansehen: %w", pfad, err)
	}

	defer datei.Close() //nolint:errcheck // nur gelesen

	kopf := make([]byte, len(Klartextmarke))

	gelesen, err := io.ReadFull(datei, kopf)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		// Kürzer als der Kopf: eine leere oder halb geschriebene Datei. Sie
		// trägt keine Chiffre, denn eine wiegt mindestens eine Seite.
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("den Bestand %q lesen: %w", pfad, err)
	}

	return string(kopf[:gelesen]) != Klartextmarke, nil
}

// Verschluesseln schreibt den geöffneten Bestand verschlüsselt an einen Ort.
//
// # Ein `VACUUM INTO` und keine Kopie
//
// Es schreibt eine **vollständige, aufgeräumte** Datenbank — kein
// Seitenabbild. Damit trägt das Ziel keine Reste gelöschter Zeilen, und genau
// das ist hier der Punkt: eine byteweise Kopie des Klartextbestands nähme
// mit, was in freien Seiten noch steht.
//
// # Der Schlüssel steht ausnahmsweise in der Adresse
//
// Sonst gilt das Gegenteil (siehe [OeffnenMitSchluessel]). Hier gibt es keine
// Verbindung, an der ein `PRAGMA` stehen könnte: das Ziel ist kein geöffneter
// Bestand, sondern die Ausgabe **einer** Anweisung. Sie lebt so lange wie der
// Aufruf und wird nirgends abgelegt.
func Verschluesseln(
	ctx context.Context, db *sql.DB, ziel, schluessel string,
) error {
	if schluessel == "" {
		return fmt.Errorf("%w: ohne Schlüssel lässt sich nichts verschlüsseln",
			domain.ErrValidation)
	}

	return vacuumNach(ctx, db,
		alsQuelle(ziel)+"?vfs="+Chiffrehuelle+"&hexkey="+schluessel)
}

// Entschluesseln schreibt ihn im Klartext an einen Ort.
//
// **Es gibt ihn, weil eine Entscheidung, die sich nicht zurücknehmen lässt,
// keine ist.** Wer verschlüsselt hat und den Bestand weitergeben will, käme
// sonst nicht mehr heran.
func Entschluesseln(ctx context.Context, db *sql.DB, ziel string) error {
	return vacuumNach(ctx, db, alsQuelle(ziel)+"?vfs="+Klarhuelle)
}

// vacuumNach schreibt die geöffnete Datenbank an eine Adresse.
//
// # Der Ort ist ein Platzhalter und keine Zeichenkette in der Anweisung
//
// **`VACUUM INTO ?` geht**: der Ausdruck hinter `INTO` wird zur **Laufzeit**
// ausgewertet, und `Sichern` nebenan bindet den Ort seit jeher als Wert. Eine
// Verkettung meldet gosec als G202.
func vacuumNach(ctx context.Context, db *sql.DB, adresse string) error {
	if err := mustern(ctx, db, "VACUUM INTO ?", adresse); err != nil {
		return uebersetze(err, "den Bestand nach %q schreiben", adresse)
	}

	return nil
}

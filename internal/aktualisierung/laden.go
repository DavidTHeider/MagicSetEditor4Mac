// **Stufe 2: holen**
// (#1100).
//
// > Sie lädt sie ins **Standard-Downloadverzeichnis** und bittet um die
// > Installation.
//
// # Die Reihenfolge ist der Punkt
//
// Erst die **Prüfsummen** — sie sind ein halbes Kilobyte —, dann das Abbild.
// 188 MB zu holen, um danach zu sagen „ich kann es nicht prüfen", wäre die
// falsche Reihenfolge; die Prüfsummenzeile fehlt entweder von Anfang an oder
// gar nicht.
//
// # Und die Datei entsteht erst, wenn sie stimmt
//
// Geschrieben wird in eine **Nachbardatei**, und erst nach der Prüfung wird
// umbenannt. Ein abgebrochener Ladevorgang lässt damit keine halbe Datei
// zurück, und eine Datei mit falscher Prüfsumme steht nie dort, wo jemand sie
// anklickt. Dasselbe Vorgehen wie bei `service.schreibeAtomar`.

package aktualisierung

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// HoechstensPruefsummen ist die Obergrenze für `pruefsummen.txt`.
//
// **Ein halbes Mebibyte für eine Datei von 412 Byte.** Die echte hat vier
// Zeilen; die Grenze ist die Absicherung gegen eine Gegenstelle, die etwas
// anderes schickt, nicht eine Vorhersage.
const HoechstensPruefsummen = 512 << 10

// Fortschritt sagt, wie weit ein Ladevorgang ist.
//
// **`ganz` ist −1, wenn die Gegenstelle keine Länge nennt.** Dann zeigt eine
// Oberfläche, was gelesen wurde, statt einen Anteil zu erfinden.
type Fortschritt func(gelesen, ganz int64)

// Laden holt das Abbild einer Freigabe und prüft seine Prüfsumme.
//
// Gibt den Pfad der fertigen Datei zurück.
func (k *Klient) Laden(
	ctx context.Context, freigabe Freigabe, verzeichnis string, melden Fortschritt,
) (string, error) {
	if freigabe.Abbild.Adresse == "" {
		return "", fmt.Errorf(
			"%w: diese Freigabe trägt kein Abbild", domain.ErrNotFound)
	}

	// **Mehrere Abbilder: keines** — welches gälte, ist nicht zu entscheiden.
	if len(freigabe.WeitereAbbilder) > 0 {
		return "", fmt.Errorf(
			"%w: diese Freigabe trägt mehrere Abbilder (%s, %s) — welches gilt, "+
				"ist nicht zu entscheiden; geladen wird keines", domain.ErrConflict,
			freigabe.Abbild.Name, strings.Join(freigabe.WeitereAbbilder, ", "))
	}

	if freigabe.Pruefsummen.Adresse == "" {
		return "", fmt.Errorf(
			"%w: diese Freigabe trägt keine Prüfsummen — ohne sie wird nichts "+
				"geladen", domain.ErrNotFound)
	}

	// **Und ohne Signatur ebenso wenig.** Die Prüfsumme sagt, dass zwei
	// Dateien zueinander passen; die Signatur sagt, von wem sie kommen. Wer
	// nur die erste hat, hat eine Prüfung gegen den, gegen den er sich schützt.
	if freigabe.Signatur.Adresse == "" {
		return "", fmt.Errorf(
			"%w: diese Freigabe trägt keine Signatur über den Prüfsummen "+
				"(`pruefsummen.txt.asc`) — ohne sie wird nichts geladen",
			domain.ErrNotFound)
	}

	// **Das Abbild trägt die Fassung, die der Tag behauptet** — SEC03
	// (#2356), Vermerk 0194. Die Signatur sagt, dass die Liste vom
	// Herausgeber kommt, nicht zu welcher Fassung; die Fassung kommt aus dem
	// Tag. Ohne diesen Vergleich lud die Anwendung unter `v9.9.9` die echten
	// Prüfsummen, die echte Signatur und das echte Abbild von 0.7.0. Der Name
	// steht in der signierten Liste — eine alte Liste nennt den neuen nicht.
	if freigabe.Abbild.Name != Abbildname(freigabe.Fassung) {
		return "", fmt.Errorf(
			"%w: das Abbild heißt %q, eine Freigabe von %s heißt %q — es wird "+
				"nichts geladen, was eine andere Fassung sein könnte", domain.ErrConflict,
			freigabe.Abbild.Name, freigabe.Fassung, Abbildname(freigabe.Fassung))
	}

	erwartet, err := k.pruefsumme(ctx, freigabe)
	if err != nil {
		return "", err
	}

	return k.abbildHolen(ctx, freigabe, verzeichnis, erwartet, melden)
}

// Abbildname ist der Name, unter dem `scripts/release.sh` das Abbild einer
// Fassung ablegt (`$NAME-$VERSION.dmg`).
func Abbildname(fassung string) string { return "MagicSetEditor4Mac-" + fassung + ".dmg" }

// pruefsumme holt `pruefsummen.txt` und sucht die Zeile des Abbilds.
func (k *Klient) pruefsumme(ctx context.Context, freigabe Freigabe) (string, error) {
	roh, err := k.Pruefsummen(ctx, freigabe)
	if err != nil {
		return "", err
	}

	summe, da := SummeAus(string(roh), freigabe.Abbild.Name)
	if !da {
		return "", fmt.Errorf(
			"%w: die Prüfsummen nennen %q nicht — es gibt nichts zu vergleichen",
			domain.ErrNotFound, freigabe.Abbild.Name)
	}

	return summe, nil
}

// Pruefsummen holt die Liste **und prüft ihre Signatur**.
//
// **Erst die Signatur, dann die Zeile.** Wer die Liste deutete und danach
// prüfte, hätte einen Weg, auf dem eine ungeprüfte Zeile schon etwas
// entschieden hat.
//
// **Herausgelöst für Werkstück O6** (#1430): das Manifest hängt an derselben
// signierten Liste wie das Abbild, und ein zweiter Weg dorthin wäre ein
// zweiter Ort, an dem jemand die Signaturprüfung vergessen kann.
func (k *Klient) Pruefsummen(ctx context.Context, freigabe Freigabe) ([]byte, error) {
	antwort, err := netz.Hole(ctx, k.unten, k.drossel, k.beigabenkenner(),
		freigabe.Pruefsummen.Adresse, HoechstensPruefsummen)
	if err != nil {
		return nil, fmt.Errorf("die Prüfsummen holen: %w", err)
	}

	if antwort.Status != netz.StatusGut {
		return nil, fmt.Errorf("%w: die Prüfsummen kamen mit %d",
			domain.ErrConflict, antwort.Status)
	}

	if err := k.signaturHolenUndPruefen(ctx, freigabe, antwort.Rumpf); err != nil {
		return nil, err
	}

	return antwort.Rumpf, nil
}

// abbildHolen lädt das Abbild in eine Nachbardatei und benennt sie um.
func (k *Klient) abbildHolen(
	ctx context.Context, freigabe Freigabe, verzeichnis, erwartet string,
	melden Fortschritt,
) (string, error) {
	antwort, err := netz.Strom(ctx, k.unten, k.drossel, k.beigabenkenner(),
		freigabe.Abbild.Adresse)
	if err != nil {
		return "", fmt.Errorf("das Abbild holen: %w", err)
	}

	defer func() { _ = antwort.Rumpf.Close() }() //nolint:errcheck // ein gelesener Rumpf ist nicht mehr zu retten

	if antwort.Status != netz.StatusGut {
		return "", fmt.Errorf("%w: das Abbild kam mit %d",
			domain.ErrConflict, antwort.Status)
	}

	ziel := filepath.Join(verzeichnis, filepath.Base(freigabe.Abbild.Name))

	nachbar, err := os.CreateTemp(verzeichnis, "."+filepath.Base(ziel)+".*")
	if err != nil {
		return "", fmt.Errorf("die Nachbardatei anlegen: %w", err)
	}

	// **Der Nachbar wird immer weggeräumt.** Nach dem Umbenennen gibt es ihn
	// nicht mehr, und das Entfernen läuft ins Leere; solange etwas
	// schiefgeht, bleibt keine halbe Datei liegen.
	defer func() {
		_ = nachbar.Close()           //nolint:errcheck // nach dem Umbenennen ohne Wirkung
		_ = os.Remove(nachbar.Name()) //nolint:errcheck // dito
	}()

	gerechnet, err := stromSchreiben(nachbar, antwort, freigabe.Abbild.Groesse, melden)
	if err != nil {
		return "", err
	}

	if gerechnet != erwartet {
		return "", fmt.Errorf(
			"%w: die Prüfsumme stimmt nicht — erwartet %s, gerechnet %s. "+
				"Die Datei wurde nicht abgelegt", domain.ErrConflict, erwartet, gerechnet)
	}

	// **Umbenannt wird mit offener Datei**, und das ist Absicht.
	// [os.File.Write] geht ohne Zwischenspeicher in den Kern; nach einem
	// geglückten Schreiben steht der Inhalt, und ein Close danach kann ihn
	// nicht mehr verlieren. Ein geprüftes Close wäre hier ein Zweig, den kein
	// Test betreten kann — geschlossen wird die Datei ohnehin, oben im defer.
	// **Eine vorhandene Datei wird ersetzt**, und das ist Absicht: der Name
	// trägt die Fassung, eine Namensgleichheit heißt also „dieselbe Freigabe
	// schon einmal geladen" — und was hier ankommt, ist gegen die Prüfsumme
	// gehalten. Eine zweite Datei mit `(1)` im Namen wäre die schlechtere
	// Antwort: dann stünden zwei da, und niemand wüsste, welche gilt.
	if err := os.Rename(nachbar.Name(), ziel); err != nil {
		return "", fmt.Errorf("die Datei an ihren Ort legen: %w", err)
	}

	return ziel, nil
}

// stromSchreiben schreibt den Strom weg und rechnet dabei die Prüfsumme.
//
// **In einem Zug und nicht zweimal.** Wer die Datei erst schriebe und danach
// noch einmal läse, läse 188 MB ein zweites Mal — für eine Zahl, die beim
// Schreiben ohnehin durch die Hand geht.
func stromSchreiben(
	nachbar io.Writer, antwort netz.Stromantwort, angekuendigt int64,
	melden Fortschritt,
) (string, error) {
	rechner := sha256.New()

	zaehler := &mitzaehler{
		unten:  io.MultiWriter(nachbar, rechner),
		ganz:   antwort.Laenge,
		melden: melden,
	}

	// **Gelesen wird höchstens, was angekündigt war.**
	//
	// Ohne Deckel liest `io.Copy` bis zum Dateiende — und eine Gegenstelle,
	// die stetig weitersendet, lässt die Nachbardatei wachsen, bis die Platte
	// voll ist. Die Frist darunter hilft nicht: sie ist eine **Stillstands**-
	// frist und wird vor jedem Lesen erneuert. Angestrichen vom Gesamtreview
	// vor 0.2.0.
	//
	// **Ein Byte mehr als angekündigt**, damit ein zu langer Strom als Fehler
	// auffällt statt als stiller Abschnitt: die Prüfsumme wäre dann zwar
	// ohnehin falsch, aber „zu lang" ist die Auskunft, die weiterhilft.
	if _, err := io.Copy(zaehler, io.LimitReader(antwort.Rumpf, Deckel(angekuendigt)+1)); err != nil {
		return "", fmt.Errorf("das Abbild schreiben: %w", err)
	}

	if zaehler.gelesen > Deckel(angekuendigt) {
		return "", fmt.Errorf(
			"%w: das Abbild ist länger als die angekündigten %d Byte — abgebrochen",
			domain.ErrConflict, Deckel(angekuendigt))
	}

	return hex.EncodeToString(rechner.Sum(nil)), nil
}

// HoechstensAbbild ist die Grenze, wenn die Freigabe keine Größe nennt.
//
// **Ein Gigabyte.** Das ausgelieferte Abbild wiegt 224 MB; das Vielfache
// davon ist großzügig und trotzdem eine Grenze. Ohne sie hinge der
// Plattenbedarf an dem, was ein fremder Rechner schickt.
const HoechstensAbbild = 1 << 30

// Deckel gibt die Obergrenze für einen Ladevorgang.
//
// **Die angekündigte Größe, sonst [HoechstensAbbild].** GitHub nennt sie in
// jeder Freigabeantwort; ein Nullwert heißt „nicht genannt" und nicht „leer".
func Deckel(angekuendigt int64) int64 {
	if angekuendigt > 0 && angekuendigt < HoechstensAbbild {
		return angekuendigt
	}

	return HoechstensAbbild
}

// mitzaehler schreibt weiter und meldet dabei den Fortschritt.
type mitzaehler struct {
	unten   io.Writer
	gelesen int64
	ganz    int64
	melden  Fortschritt
}

func (m *mitzaehler) Write(teil []byte) (int, error) {
	geschrieben, err := m.unten.Write(teil)

	m.gelesen += int64(geschrieben)

	if m.melden != nil {
		m.melden(m.gelesen, m.ganz)
	}

	return geschrieben, err
}

// SummeAus liest die Prüfsumme einer Datei aus dem Text von `shasum -a 256`.
//
// # Die Gestalt
//
//	e3b0c442…  ./MagicSetEditor4Mac-0.2.0.dmg
//
// Zwei Leerzeichen zwischen Summe und Name, und der Name trägt das `./`, weil
// `release.sh` im Ausgabeverzeichnis `shasum -a 256 ./*.dmg` fährt. Verglichen
// wird deshalb der **Grundname** — wer auf die Zeichenkette prüfte, prüfte den
// Aufruf, mit dem die Liste entstanden ist.
func SummeAus(text, name string) (string, bool) {
	gesucht := filepath.Base(name)

	for _, zeile := range strings.Split(text, "\n") {
		felder := strings.Fields(zeile)
		if len(felder) != 2 {
			continue
		}

		if filepath.Base(felder[1]) == gesucht {
			return felder[0], true
		}
	}

	return "", false
}

// signaturHolenUndPruefen holt `pruefsummen.txt.asc` und prüft sie.
func (k *Klient) signaturHolenUndPruefen(
	ctx context.Context, freigabe Freigabe, pruefsummen []byte,
) error {
	antwort, err := netz.Hole(ctx, k.unten, k.drossel, k.beigabenkenner(),
		freigabe.Signatur.Adresse, HoechstensSignatur)
	if err != nil {
		return fmt.Errorf("die Signatur holen: %w", err)
	}

	if antwort.Status != netz.StatusGut {
		return fmt.Errorf("%w: die Signatur kam mit %d",
			domain.ErrConflict, antwort.Status)
	}

	return SignaturPruefen(pruefsummen, antwort.Rumpf)
}

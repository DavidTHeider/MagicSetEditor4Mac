package msepaket

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// ErrZuGross meldet eine Paketdatei, die größer ist, als ein Paket sie tragen
// darf — FMT-06.AK-01 (#1695).
var ErrZuGross = errors.New("die Datei ist größer, als eine Paketdatei sein darf")

// GrenzePaketdatei ist, wie groß eine einzelne Datei eines Pakets ausgepackt
// höchstens sein darf.
//
// **Gemessen, nicht geschätzt** (2026-09-23): die größte Datei im vollen Pack
// ist ein Wörterbuch von 10,7 MB, das größte Bild 7,5 MB, die größte
// Setdatei der zehn kanonischen Sätze 3,9 MB. 256 MiB lassen das
// Vierundzwanzigfache — Raum für Kunst in Druckauflösung — und halten eine
// Datei auf, die beim Auspacken zu Gigabytes wird.
const GrenzePaketdatei int64 = 256 << 20

// leseBegrenzt liest eine Datei aus einem Paketbaum — höchstens
// [GrenzePaketdatei] Byte.
//
// **Ohne der Größe zu glauben, die der Baum nennt.** `fs.ReadFile` teilt so
// viel Speicher zu, wie `Stat` sagt, bevor ein Byte gelesen ist, und für einen
// Zip-Eintrag ist das die Zahl aus seinem Kopf. Hier weist eine Behauptung
// über der Grenze ab, eine darunter teilt nichts zu: gelesen wird, was
// wirklich kommt, und nach der Grenze ist Schluss — auch wenn der Kopf
// weniger behauptet hat, als ausgepackt wird.
func leseBegrenzt(baum fs.FS, name string) ([]byte, error) {
	return leseBisZu(baum, name, GrenzePaketdatei)
}

// LeseDateiBegrenzt liest eine einzelne Datei von der Platte — höchstens
// [GrenzePaketdatei] Byte.
//
// **Für jeden Weg, der eine Datei aus einem Paket liest, ohne das Paket zu
// öffnen** — Befund 4 des Zweitblicks über Phase 6.1: der Schreibweg, die
// Symbolzeichnung und die `game`-Datei beim Anlegen eines Stils lasen mit
// `os.ReadFile` ganz, und eine ausgetauschte Datei belegte beliebig viel
// Speicher, bevor der Konflikt kam.
func LeseDateiBegrenzt(pfad string) ([]byte, error) {
	return leseBegrenzt(os.DirFS(filepath.Dir(pfad)), filepath.Base(pfad))
}

// leseBisZu ist [leseBegrenzt] mit einer genannten Grenze.
func leseBisZu(baum fs.FS, name string, grenze int64) ([]byte, error) {
	datei, err := baum.Open(name)
	if err != nil {
		return nil, err
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen ist gelesen

	if angaben, fehler := datei.Stat(); fehler == nil && angaben.Size() > grenze {
		return nil, fmt.Errorf("%w: %s nennt %d Byte, höchstens %d", ErrZuGross, name,
			angaben.Size(), grenze)
	}

	inhalt, err := io.ReadAll(io.LimitReader(datei, grenze+1))
	if err != nil {
		return nil, err
	}

	if int64(len(inhalt)) > grenze {
		return nil, fmt.Errorf("%w: %s wird ausgepackt größer als %d Byte", ErrZuGross, name, grenze)
	}

	return inhalt, nil
}

// ErrBildZuGross meldet ein Bild, das zu groß ist, um es auszupacken — über
// der Grenze je Seite, der Bildpunkte oder des Speichers ([GrenzeAuspacken]).
//
// **Nicht mehr „als ein Paketbild sein darf"** (#1974): dieselbe Grenze gilt
// für ein Bild, das jemand vom Schreibtisch wählt, und der Satz kam dort an.
var ErrBildZuGross = errors.New("das Bild ist zu groß, um es auszupacken")

// GrenzeBildseite und GrenzeBildpunkte begrenzen, was ausgepackt wird.
//
// **Gemessen** (2026-09-23) über 30.261 PNG im vollen Pack: das größte ist
// 2976 × 4157 — eine Karte bei 1200 dpi, 12,4 Millionen Bildpunkte. Hundert
// Millionen lassen das Achtfache und kosten ausgepackt 400 MB — **bei 8 Bit
// je Kanal**.
//
// **Die Grenze zählt Bildpunkte und nicht Byte**, und hier stand bis QA-06b
// (#1967) nur der eine Fall. `FuzzDekodiereBild` fand ein PNG mit 16 Bit je
// Kanal, das je Bildpunkt 8 Byte kostet; ein progressives JPEG legt je
// Farbauszug 4 Byte Koeffizienten dazu (`image/jpeg`, `scan.go` Zeile 167,
// Go 1.27.1), mit vier Auszügen samt Bild und Schwarzauszug bis zu 24 Byte.
// Unter der Grenze waren damit bis zu 2,4 GB erlaubt. **Seit FMT-06b**
// (#1974)
// steht daneben eine Grenze in Byte, [GrenzeAuspacken], und die Schätzung
// je Format und Farbmodell ist gemessen (`auspackkosten.go`).
const (
	GrenzeBildseite  = 32_768
	GrenzeBildpunkte = 100_000_000
)

// GrenzeEinbindungen und GrenzeEinbindetiefe begrenzen, wie weit
// `include file:` trägt — je Lauf des [Einbinder]s.
//
// **Gemessen** (2026-09-23) über alle 526 Pakete des vollen Packs: die meisten
// Einbindungen hat `magic.mse-game` mit 87, die tiefste Kette
// `magic-m15-meld-3in1.mse-style` mit fünf Ebenen. Tausend lassen das
// Elffache, sechzehn das Dreifache — und halten eine Raute auf, die sonst
// Millionen Dateien läse.
const (
	GrenzeEinbindungen  = 1_000
	GrenzeEinbindetiefe = 16
)

package service

// **Eine eigene Kopie eines Pakets** — FMT-02.AK-04 und S10.NAT-01
// (#1691).
//
// Mitgelieferte und fremdverwaltete Pakete werden über eigene Kopien
// angepasst. Die Saat liegt im Bündel und lässt sich gar nicht schreiben; ein
// bezogenes Paket ließe es, verlöre dabei aber seinen Stand — beim nächsten
// Bezug stünde die eigene Änderung gegen die fremde.
//
// # Portierungsmarke
//
// **new** — gesucht in `src/util/io/package.cpp` und
// `src/util/io/package_manager.cpp` (Fassung f9d9356 vom 2021-07-06): MSE2
// kennt `Package::saveCopy`, aber nur, um ein offenes Paket unter einem
// anderen Dateinamen zu sichern; eine Kopie mit eigenem Namen im
// Paketverzeichnis bietet es nirgends an (`docs/FAEHIGKEITEN.md`, Kopieren).

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// kopiezusatz hängt an den Namen einer Kopie, für die niemand einen gewählt
// hat — ein Schlüssel und kein Menschentext, denn er wird zum Dateinamen.
const kopiezusatz = "-kopie"

// Kopieren legt von einem Paket eine eigene Kopie im verwalteten Verzeichnis
// an und gibt sie zurück.
//
// **Ohne gewählten Namen** heißt sie wie das Original mit `-kopie`, und gibt
// es den schon, mit `-kopie-2` und so fort. **Ein Stil heißt nach seinem
// Spiel** (`<spiel>-<name>`), sonst fände ihn kein Set über `stylesheet:` —
// ein gewählter Name bekommt das Spiel vorangestellt wie beim Anlegen.
//
// **Die Kopie bekommt einen eigenen Kurznamen**, abgeleitet aus ihrem Namen:
// sonst stünde derselbe in jeder Wahl zweimal, und niemand sähe, welcher die
// Kopie ist. Alles andere bleibt Byte für Byte.
func (v *Paketverwalter) Kopieren(ctx context.Context, art, name, neu string) (domain.Paket, error) {
	original, pakete, err := v.gefunden(art, name)
	if err != nil {
		return domain.Paket{}, err
	}

	return v.kopieVon(ctx, original, pakete, neu)
}

// kopieVon legt die Kopie eines gefundenen Pakets an.
//
// **Das Öffnen kann hier noch scheitern**, obwohl `gefunden` das Paket eben
// gelesen hat: zwischen beidem liegt die Platte, und jemand anderes kann es
// entfernt haben.
func (v *Paketverwalter) kopieVon(
	ctx context.Context, original domain.Paket, pakete []domain.Paket, neu string,
) (domain.Paket, error) {
	art := original.Art

	quelle, err := msepaket.Oeffne(original.Pfad)
	if err != nil {
		return domain.Paket{}, err
	}

	defer func() { _ = quelle.Schliessen() }() //nolint:errcheck // nur gelesen

	spiel := ""
	if art == string(msepaket.ArtStil) {
		stil, _ := quelle.Stilmodell()
		spiel = stil.Spiel
	}

	ziel, err := v.kopiename(pakete, original, spiel, neu)
	if err != nil {
		return domain.Paket{}, err
	}

	bau, ort, err := wegwerfpaket(v.bauort(), "mse4-kopie-", ziel+"."+art)
	if err != nil {
		return domain.Paket{}, fmt.Errorf("die Kopie schreiben: %w", err)
	}

	defer func() { _ = os.RemoveAll(bau) }() //nolint:errcheck // Wegwerfverzeichnis

	kurz := strings.TrimPrefix(ziel, spiel+"-")
	if err := kopiereDateien(quelle, ort, msepaket.Art(art).Hauptdatei(), kurz, v.ablegen); err != nil {
		return domain.Paket{}, err
	}

	return v.einbinden(ctx, ort, false)
}

// gefunden sucht ein Paket nach Art und Namen — und gibt die Pakete mit,
// gegen die der Name der Kopie geprüft wird.
//
// **Von der Platte und nicht aus dem Katalog.** Die Flächen ohne Fenster
// frischen ihn nicht vor jedem Ruf auf; ein Paket, das seit dem letzten
// Abgleich hinzukam, gäbe es sonst nicht zu kopieren, und ein vergebener Name
// wäre frei.
func (v *Paketverwalter) gefunden(art, name string) (original domain.Paket, pakete []domain.Paket, fehler error) {
	pakete, _ = v.lesen(nil)

	for _, paket := range pakete {
		if paket.Art == art && paket.Name == name {
			return paket, pakete, nil
		}
	}

	return domain.Paket{}, nil, fmt.Errorf("%w: es gibt kein Paket %s.%s", domain.ErrNotFound, name, art)
}

// kopiename bestimmt den Namen der Kopie — gewählt oder der nächste freie.
func (v *Paketverwalter) kopiename(pakete []domain.Paket, original domain.Paket, spiel, neu string) (string, error) {
	if neu != "" {
		if spiel != "" {
			neu = stilverzeichnis(spiel, neu)
		}

		if !istPaketname(neu) {
			return "", domain.NewFieldError("name", domain.SchluesselPaketnameUngueltig,
				"Ein Paketname besteht aus Buchstaben, Ziffern, Bindestrichen und "+
					"Unterstrichen; %q geht nicht.", neu)
		}

		return neu, nil
	}

	vergeben := map[string]bool{}
	for _, paket := range pakete {
		vergeben[paket.Dateiname()] = true
	}

	basis := original.Name + kopiezusatz

	for nummer := 1; ; nummer++ {
		kandidat := basis
		if nummer > 1 {
			kandidat = basis + "-" + strconv.Itoa(nummer)
		}

		if !vergeben[kandidat+"."+original.Art] && !belegt(filepath.Join(v.verwaltet, kandidat+"."+original.Art)) {
			return kandidat, nil
		}
	}
}

// kopiereDateien schreibt jede Datei des Pakets nach `ort` — die Hauptdatei
// mit dem neuen Kurznamen.
//
// **Aus dem geöffneten Paket und nicht vom Verzeichnis**: ein Original kann
// ein Archiv sein, und installiert wird nur ein Verzeichnis (`einbinden`).
func kopiereDateien(
	quelle *msepaket.Paket, ort, hauptdatei, kurz string, ablegen func(string, []byte) error,
) error {
	dateien, err := quelle.Dateien()
	if err != nil {
		return err
	}

	// **Die Namen sind Paketpfade** aus `fs.WalkDir` — gültig im Sinn von
	// `fs.ValidPath` und damit ohne `..` und ohne führenden Schrägstrich.
	for _, name := range dateien {
		inhalt, fehler := quelle.Datei(name)
		if fehler != nil {
			return fehler
		}

		if name == hauptdatei {
			inhalt = mitKurzname(inhalt, kurz)
		}

		pfad := filepath.Join(ort, filepath.FromSlash(name))

		fehler = os.MkdirAll(filepath.Dir(pfad), 0o750)
		if fehler == nil {
			fehler = ablegen(pfad, inhalt)
		}

		if fehler != nil {
			return fmt.Errorf("die Kopie schreiben: %w", fehler)
		}
	}

	return nil
}

// mitKurzname setzt `short name:` auf der obersten Ebene — oder fügt die Zeile
// nach der ersten ein, wenn es keine gibt.
//
// **Zeilen und nicht der Baum**: ein Zurückschreiben über den Zerteiler
// verlöre Kommentare, Einrückung und die Reihenfolge — und die Kopie soll
// außer diesem einen Wert gleich bleiben.
func mitKurzname(inhalt []byte, kurz string) []byte {
	zeilen := strings.Split(string(inhalt), "\n")
	neu := "short name: " + kurz

	for nummer, zeile := range zeilen {
		ohneMarke := strings.TrimPrefix(zeile, "\ufeff")
		if !strings.HasPrefix(ohneMarke, "short name:") {
			continue
		}

		if strings.HasSuffix(zeile, "\r") {
			neu += "\r"
		}

		zeilen[nummer] = zeile[:len(zeile)-len(ohneMarke)] + neu

		return []byte(strings.Join(zeilen, "\n"))
	}

	zeilen = append(zeilen[:1], append([]string{neu}, zeilen[1:]...)...)

	return []byte(strings.Join(zeilen, "\n"))
}

// belegt sagt, ob unter einem Pfad schon etwas liegt — auch ein Paket, das
// der Katalog noch nicht kennt.
func belegt(pfad string) bool {
	_, err := os.Stat(pfad)

	return err == nil
}

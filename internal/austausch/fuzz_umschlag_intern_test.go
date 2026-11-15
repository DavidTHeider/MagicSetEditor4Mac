package austausch

// **Ein Umschlag von einem bekannten Absender** — QA-06b
// (#1967),
// Flächen 27 und 28 aus `docs/FUZZZIELE.md`: Behälter und Manifest.
//
// # Warum unterschrieben
//
// **Weil die Quarantäne gerade dann halten muss.** Ein Umschlag mit fremder
// Unterschrift endet an `peer.Pruefe`, und dahinter sähe ein Fuzzer nichts.
// Der Kopf von `quarantaene.go` sagt es: ein bekannter Absender kann ein
// übergroßes Bild schicken, und ein übernommenes Konto bleibt ein bekannter
// Absender. Das Ziel unterschreibt deshalb **jedes** Manifest selbst — der
// Fuzzer schreibt dessen Bytes, die Namen und die Inhalte, und `@ich`,
// `@summe0`, `@groesse0` … setzt es vor dem Unterschreiben ein, damit die
// Saat bis hinter die Prüfsummen kommt.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **etwas außerhalb des Verzeichnisses**, das der Aufrufer nennt
//   - zweimal derselbe Umschlag, zwei Urteile
//   - abgelehnt ohne Klasse — jede Ablehnung ist [domain.ErrValidation] oder
//     [domain.ErrConflict]
//   - angenommen, und **ein Objekt hält nicht, was das Manifest sagt**: ein
//     Name, den [filepath.IsLocal] nicht für einen Weg im Verzeichnis hält,
//     eine Datei, deren Prüfsumme oder **Größe** eine andere ist, eine Art,
//     die es nicht gibt, ein `braucht`, das nicht im Umschlag liegt
//   - angenommen über einer Grenze der Quarantäne — gemessen an den
//     **ausgepackten** Bytes und nicht an der Angabe
//   - angenommen mit einem anderen Urheber als dem, der unterschrieben hat
//   - **ergänzt nach dem Saatlauf**: das Verzeichnis der Sendung ist ein
//     Verzeichnis, und jedes Objekt liegt **darin**. Die Saat mit dem Namen
//     `.` ging durch alle Zusagen oben — [filepath.IsLocal] hält `.` für
//     lokal —, und ausgepackt wurde sie an die Stelle des Verzeichnisses
//
// # Was der Lauf gefunden hat
//
// **Die Größe im Manifest hielt niemand gegen die Bytes** (beim Festlegen
// der Zusage oben gefunden, nicht vom Lauf: er kam in dreißig Sekunden nur
// auf 241 Ausführungen) und **ein Name `.` oder `a/../b`** stand nicht
// kanonisch da. Beide behoben; die Saat trägt beide Fälle.
//
//	go test ./internal/austausch -run Fuzz -fuzz FuzzUmschlag -fuzztime 2m

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// umschlagsaat ist ein Manifest, wie [Schreibe] es schreibt — mit den
// Platzhaltern aus dem Kopf.
const umschlagsaat = `{"fassung":1,"urheber":"@ich","erstellt":"2026-10-01T00:00:00Z","objekte":[` +
	`{"art":"karte","name":"karten/eine.json","bezeichnung":"Eine","sha256":"@summe0","groesse":@groesse0,"braucht":null},` +
	`{"art":"vorlage","name":"v/probe.mse-style","bezeichnung":"","sha256":"@summe1","groesse":@groesse1,"braucht":["karten/eine.json"]}]}`

// **Was angenommen ist, liegt im Verzeichnis und ist, was das Manifest sagt.**
func FuzzUmschlag(f *testing.F) {
	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		f.Fatalf("Erzeugen: %v", fehler)
	}

	for _, saat := range [][3]string{
		{umschlagsaat, "karten/eine.json\nv/probe.mse-style", `{"name":"Mausbraten"}`},
		{strings.ReplaceAll(umschlagsaat, `"karten/eine.json"`, `"a/../karten/eine.json"`), "a/../karten/eine.json\nv/probe.mse-style", "x"},
		{strings.ReplaceAll(umschlagsaat, `"art":"karte"`, `"art":"zauber"`), "karten/eine.json\nv/probe.mse-style", "x"},
		{strings.ReplaceAll(umschlagsaat, `"fassung":1`, `"fassung":2`), "karten/eine.json\nv/probe.mse-style", "x"},
		{`{"fassung":1,"urheber":"@ich","objekte":[{"art":"karte","name":".","sha256":"@summe0","groesse":@groesse0}]}`, ".", "x"},
		{strings.ReplaceAll(umschlagsaat, `"groesse":@groesse0`, `"groesse":1`), "karten/eine.json\nv/probe.mse-style", "zu lang"},
		{`{"fassung":1,"urheber":"@ich","objekte":[]}`, "", ""},
		{`null`, "a", "b"},
	} {
		f.Add([]byte(saat[0]), saat[1], []byte(saat[2]))
	}

	f.Fuzz(func(t *testing.T, manifest []byte, namen string, inhalt []byte) {
		teile := umschlagteile(namen, inhalt)
		roh := platzhalterEinsetzen(manifest, ich, teile)

		rahmen := t.TempDir()
		ort := filepath.Join(rahmen, "probe"+Endung)

		if sicherung.PackenAls(ort, aeusseresManifest, append(teile,
			sicherung.Teil{Name: Manifestname, Inhalt: roh},
			sicherung.Teil{Name: Unterschriftsname, Inhalt: ich.Signiere(roh)}), sicherung.Angaben{}) != nil {
			return
		}

		sendung, lesefehler := Lies(ort, filepath.Join(rahmen, "eins"))
		nochmal, nochmalFehler := Lies(ort, filepath.Join(rahmen, "zwei"))

		if (lesefehler == nil) != (nochmalFehler == nil) || !reflect.DeepEqual(sendung.Manifest, nochmal.Manifest) {
			t.Fatalf("derselbe Umschlag, zwei Urteile: %v / %v", lesefehler, nochmalFehler)
		}

		nurImRahmen(t, rahmen)

		if lesefehler != nil {
			if !errors.Is(lesefehler, domain.ErrValidation) && !errors.Is(lesefehler, domain.ErrConflict) {
				t.Errorf("abgelehnt ohne Klasse: %v", lesefehler)
			}

			return
		}

		if sendung.Urheber != ich.Kennung() || sendung.Manifest.Fassung != Fassung {
			t.Errorf("angenommen von %q in Fassung %d", sendung.Urheber, sendung.Manifest.Fassung)
		}

		pruefeObjekte(t, sendung)
	})
}

// umschlagteile macht aus jeder Zeile von `namen` einen Teil — höchstens acht,
// jeder mit anderen Bytes.
func umschlagteile(namen string, inhalt []byte) []sicherung.Teil {
	if namen == "" {
		return nil
	}

	var teile []sicherung.Teil

	for stelle, name := range strings.SplitN(namen, "\n", 8) {
		teile = append(teile, sicherung.Teil{Name: name, Inhalt: append([]byte{byte(stelle)}, inhalt...)})
	}

	return teile
}

// platzhalterEinsetzen setzt Kennung, Summen und Größen in das Manifest.
func platzhalterEinsetzen(manifest []byte, ich peer.Identitaet, teile []sicherung.Teil) []byte {
	text := strings.ReplaceAll(string(manifest), "@ich", string(ich.Kennung()))

	for stelle := len(teile) - 1; stelle >= 0; stelle-- {
		summe := sha256.Sum256(teile[stelle].Inhalt)
		nummer := strconv.Itoa(stelle)
		text = strings.ReplaceAll(text, "@summe"+nummer, hex.EncodeToString(summe[:]))
		text = strings.ReplaceAll(text, "@groesse"+nummer, strconv.Itoa(len(teile[stelle].Inhalt)))
	}

	return []byte(text)
}

// nurImRahmen sieht nach, dass neben den zwei Zielen nur der Umschlag liegt.
func nurImRahmen(t *testing.T, rahmen string) {
	t.Helper()

	eintraege, fehler := os.ReadDir(rahmen)
	if fehler != nil {
		t.Fatalf("den Rahmen lesen: %v", fehler)
	}

	for _, eintrag := range eintraege {
		if name := eintrag.Name(); name != "probe"+Endung && name != "eins" && name != "zwei" {
			t.Errorf("außerhalb des Verzeichnisses liegt %q", name)
		}
	}
}

// pruefeObjekte hält jedes angenommene Objekt gegen die Platte.
func pruefeObjekte(t *testing.T, sendung Sendung) {
	t.Helper()

	if angaben, fehler := os.Stat(sendung.Verzeichnis); fehler != nil || !angaben.IsDir() {
		t.Fatalf("das Verzeichnis der Sendung ist keines: %v", fehler)
	}

	namen := map[string]bool{}

	var gesamt int64

	for _, objekt := range sendung.Manifest.Objekte {
		namen[objekt.Name] = true

		if !filepath.IsLocal(objekt.Name) || !objekt.Art.gilt() {
			t.Errorf("angenommen: %q als %q", objekt.Name, objekt.Art)

			continue
		}

		liegt, fehler := os.ReadFile(filepath.Join(sendung.Verzeichnis, filepath.FromSlash(objekt.Name)))
		if fehler != nil {
			t.Errorf("%q ist angenommen und nicht zu lesen: %v", objekt.Name, fehler)

			continue
		}

		summe := sha256.Sum256(liegt)
		if hex.EncodeToString(summe[:]) != objekt.Summe || int64(len(liegt)) != objekt.Groesse {
			t.Errorf("%q: %d Byte im Verzeichnis, das Manifest sagt %d", objekt.Name, len(liegt), objekt.Groesse)
		}

		if int64(len(liegt)) > HoechstensObjekt {
			t.Errorf("%q: %d Byte ausgepackt, erlaubt sind %d", objekt.Name, len(liegt), HoechstensObjekt)
		}

		gesamt += int64(len(liegt))
	}

	if gesamt > HoechstensGesamt || len(sendung.Manifest.Objekte) > HoechstensObjekte {
		t.Errorf("%d Objekte mit %d Byte angenommen", len(sendung.Manifest.Objekte), gesamt)
	}

	for _, objekt := range sendung.Manifest.Objekte {
		for _, gebraucht := range objekt.Braucht {
			if !namen[gebraucht] {
				t.Errorf("%q braucht %q, und das liegt nicht im Umschlag", objekt.Name, gebraucht)
			}
		}
	}
}

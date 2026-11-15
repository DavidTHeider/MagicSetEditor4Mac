package sicherung_test

// **Eine verschlüsselte Sicherung aus fremder Hand** — QA-06c
// (#1976),
// Fläche 23 aus `docs/FUZZZIELE.md`.
//
// # Warum ein eigenes Ziel neben `FuzzEntpacken`
//
// **Vor dem ZIP liegt age**, und dessen Kopf bestimmt der Absender: die
// Empfänger, das Salz, den Arbeitsfaktor der Schlüsselableitung, die
// Blöcke. `FuzzEntpacken` erreicht diesen Weg nie — dort gibt es kein
// Kennwort.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **eine Datei außerhalb des genannten Verzeichnisses**
//   - zweimal dieselbe Datei, zwei Ergebnisse
//   - eine Beschreibung **und** ein Fehler, oder keins von beidem
//   - **ein Arbeitsfaktor über dem Deckel wird gerechnet** — erkennbar daran,
//     dass danach das Kennwort verworfen wird: das setzt die Ableitung voraus
//   - **mit falschem Kennwort kommt etwas heraus**
//
//	go test ./internal/sicherung -run Fuzz -fuzz FuzzEntpackenVerschluesselt -fuzztime 2m -fuzzminimizetime 5s

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// faktorgrenze ist der Deckel, den das Entpacken setzt — der Faktor, mit dem
// diese Anwendung und `age -p` schreiben.
const faktorgrenze = 18

// verschluesseltesaat baut eine gültige Sicherung und ihre Abwandlungen.
func verschluesseltesaat(f *testing.F) [][]byte {
	f.Helper()
	f.Cleanup(sicherung.ArbeitsfaktorSetzen(10))

	ort := f.TempDir()

	quelle := filepath.Join(ort, "bestand.db")
	if fehler := os.WriteFile(quelle, []byte("ein Bestand"), 0o600); fehler != nil {
		f.Fatal(fehler)
	}

	ziel := filepath.Join(ort, "s"+sicherung.Endung)
	if fehler := sicherung.PackenVerschluesselt(ziel,
		[]sicherung.Teil{{Name: "bestand.db", Quelle: quelle}},
		sicherung.Angaben{Fassung: "0.6.0"}, probekennwort); fehler != nil {
		f.Fatalf("packen: %v", fehler)
	}

	gueltig, fehler := os.ReadFile(ziel) //nolint:gosec // die Datei dieses Laufs
	if fehler != nil {
		f.Fatal(fehler)
	}

	faktoren := []string{"19", "30", "0", "-1", "99999999999"}

	saat := make([][]byte, 0, 4+len(faktoren))
	saat = append(saat, gueltig, gueltig[:len(gueltig)/2], []byte(sicherung.Chiffrekopf), []byte{})

	for _, faktor := range faktoren {
		saat = append(saat, scryptzeile.ReplaceAll(gueltig, []byte("-> scrypt $1 "+faktor)))
	}

	return saat
}

// verlangterFaktor gibt den Arbeitsfaktor, den der Kopf nennt — und ob er
// einen nennt.
func verlangterFaktor(roh []byte) (int, bool) {
	treffer := scryptzeile.FindSubmatch(roh)
	if treffer == nil {
		return 0, false
	}

	faktor, fehler := strconv.Atoi(string(treffer[2]))

	return faktor, fehler == nil
}

func FuzzEntpackenVerschluesselt(f *testing.F) {
	for _, saat := range verschluesseltesaat(f) {
		f.Add(saat, true)
	}

	f.Fuzz(func(t *testing.T, roh []byte, richtig bool) {
		heim := t.TempDir()
		ziel := filepath.Join(heim, "hinein")
		daneben := filepath.Join(heim, "daneben")

		for _, ordner := range []string{ziel, daneben} {
			if fehler := os.MkdirAll(ordner, 0o750); fehler != nil {
				t.Fatalf("%s anlegen: %v", ordner, fehler)
			}
		}

		pfad := filepath.Join(heim, "sicherung.zip")
		if fehler := os.WriteFile(pfad, roh, 0o600); fehler != nil {
			t.Fatalf("die Sicherung schreiben: %v", fehler)
		}

		kennwort := probekennwort
		if !richtig {
			kennwort = "ein anderes kennwort"
		}

		beschreibung, fehler := sicherung.EntpackenVerschluesselt(pfad, ziel, kennwort)

		nochmal, zweiterFehler := sicherung.EntpackenVerschluesselt(pfad, ziel, kennwort)
		if (fehler == nil) != (zweiterFehler == nil) {
			t.Fatalf("dieselbe Sicherung einmal so und einmal anders: %v gegen %v", fehler, zweiterFehler)
		}

		if fehler == nil && !reflect.DeepEqual(beschreibung, nochmal) {
			t.Errorf("dieselbe Sicherung gab zwei Beschreibungen")
		}

		if fehler == nil && !richtig {
			t.Errorf("mit falschem Kennwort kam eine Sicherung heraus: %+v", beschreibung)
		}

		var feld *domain.FieldError
		if faktor, genannt := verlangterFaktor(roh); genannt && faktor > faktorgrenze &&
			errors.As(fehler, &feld) && feld.Field == "kennwort" {
			t.Errorf("Arbeitsfaktor %d gerechnet und dann das Kennwort verworfen", faktor)
		}

		nichtsDaneben(t, heim, daneben)
	})
}

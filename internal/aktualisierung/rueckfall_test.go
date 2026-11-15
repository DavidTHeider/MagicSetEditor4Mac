// **Eine alte Freigabe unter einem neuen Tag wird nicht geladen** — SEC03
// (#2356),
// die Zeile „Rollback" der Signaturmatrix.
//
// # Was die Signatur allein nicht leistet
//
// Sie sagt, dass `pruefsummen.txt` vom Herausgeber kommt — **nicht, zu
// welcher Fassung**. Die Fassung kommt aus dem Tag, und den nennt die
// Gegenstelle. Wer den Freigabeweg beherrscht, legt unter `v9.9.9` die
// Prüfsummen, die Signatur und das Abbild von 0.7.0 ab: alle drei echt, alle
// drei passen zueinander — und die Anwendung hielte eine alte Fassung für
// die neueste.
//
// **Die Fassung steht im Namen des Abbilds, und der Name steht in der
// signierten Liste** (`release.sh`: `$NAME-$VERSION.dmg`). Verlangt die
// Anwendung den Namen der Fassung, die der Tag behauptet, findet sie ihn in
// einer alten Liste nicht.
package aktualisierung_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Nicht parallel**: [probeschluessel] setzt den Bund für das ganze Paket.
// Parallel tauschten die Fälle ihn sich gegenseitig aus, und ein Fall war
// grün, weil die Signatur aus dem falschen Grund scheiterte — gesehen am
// ersten roten Lauf, an dem einer der drei grün blieb.
//
//nolint:paralleltest,tparallel // der Bund ist eine Paketvariable
func TestEineAlteFreigabeUnterNeuemTagWirdNichtGeladen(t *testing.T) {
	faelle := map[string]struct {
		fassung, abbild string
	}{
		"das alte Abbild unter seinem Namen":    {"9.9.9", "MagicSetEditor4Mac-0.7.0.dmg"},
		"ein Abbild ohne Fassung im Namen":      {"9.9.9", "MagicSetEditor4Mac.dmg"},
		"die Fassung nur als Teil eines Namens": {"0.7", "MagicSetEditor4Mac-0.7.0.dmg"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			inhalt := abbild()
			gegenstelle := gegenstelleMit(t, inhalt, summe(inhalt)+"  ./"+fall.abbild+"\n")
			ziel := t.TempDir()

			_, err := probeklient(t, gegenstelle.URL).Laden(t.Context(),
				aktualisierung.Freigabe{
					Fassung:     fall.fassung,
					Abbild:      beigabe(fall.abbild, gegenstelle.URL+"/abbild"),
					Pruefsummen: beigabe("pruefsummen.txt", gegenstelle.URL+"/pruefsummen"),
					Signatur:    beigabe("pruefsummen.txt.asc", gegenstelle.URL+"/signatur"),
				}, ziel, nil)
			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("Laden: %v — erwartet ein Konflikt, weil das Abbild nicht die Fassung %s trägt", err, fall.fassung)
			}

			eintraege, lesefehler := os.ReadDir(ziel)
			if lesefehler != nil {
				t.Fatal(lesefehler)
			}

			if len(eintraege) != 0 {
				t.Errorf("im Ziel liegt %d Datei(en) — ein abgewiesenes Abbild bleibt nicht liegen", len(eintraege))
			}
		})
	}
}

func TestDerNameDerFassungIstDerDesFreigabewegs(t *testing.T) {
	t.Parallel()

	if name := aktualisierung.Abbildname("0.9.0"); name != "MagicSetEditor4Mac-0.9.0.dmg" {
		t.Errorf("Abbildname: %q", name)
	}
}

// **Und das Manifest nennt dieselbe Fassung wie der Tag.** Seine Summe steht
// in der signierten Liste; eine alte Liste trägt das alte Manifest, und das
// Urteil über die Verträglichkeit fiele über eine andere Fassung als die, die
// angeboten wird.
func TestEinManifestEinerAnderenFassungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	dienst := manifestdienst(t, echtesManifest)
	klient := aktualisierung.NeuerKlient(dienst.URL,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), "")

	_, err := klient.Manifest(t.Context(), aktualisierung.Freigabe{
		Fassung:  "9.9.9",
		Manifest: aktualisierung.Beigabe{Name: aktualisierung.Manifestname, Adresse: dienst.URL},
	}, []byte(summenliste()))
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Manifest: %v — erwartet ein Konflikt, weil es 0.5.0 nennt und 9.9.9 angeboten ist", err)
	}
}

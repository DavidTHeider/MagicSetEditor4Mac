package aktualisierung

// **Das Manifest einer Freigabe** — Werkstück O6 von Phase 5.2
// (#1430).
//
// # Woran es hängt, und warum das genügt
//
// **Am vorhandenen Vertrauensanker.** `pruefsummen.txt` wird mit OpenPGP
// signiert (`signatur.go`), und `release.sh` nimmt `manifest.json` in dieselbe
// Liste auf. Wer das Manifest lädt, prüft seine SHA-256 gegen die **signierte**
// Liste — damit ist es genauso authentisch wie das Abbild daneben, ohne einen
// zweiten Schlüssel, ein zweites Format und eine zweite Prüfung.
//
// **Ein manipuliertes Manifest fällt damit auf.** Das ist die erste der drei
// Abnahmen aus dem Issue, und sie kostet keine Zeile eigener Kryptografie.
//
// # Und mit Apple-Notarisierung hat das nichts zu tun
//
// Das Issue sagt es ausdrücklich, und
// [0003](../../docs/entscheidungen/0003-keine-notarisierung.md) sagt warum:
// jene entscheidet, ob Gatekeeper das Bündel startet, diese, ob die Fassung
// diesen Bestand öffnen kann. Zwei Fragen, zwei Antworten, kein gemeinsamer
// Weg — und keine Aussage über die andere.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Manifestname ist die Beigabe, die das Manifest trägt.
const Manifestname = "manifest.json"

// HoechstensManifest ist die Grenze für seinen Umfang.
//
// **Zweihundertsechsundfünfzig Kilobyte.** Es trägt eine Handvoll Zeilen und
// je Wanderung eine — bei den heute zwei Dutzend Migrationen sind das wenige
// Kilobyte. Die Grenze ist keine Schätzung des Bedarfs, sondern die Zusage,
// dass eine fremde Antwort den Speicher nicht füllt (§16, Regel 3).
const HoechstensManifest = 256 << 10

// Manifest holt das Manifest einer Freigabe und prüft es gegen die
// **signierte** Prüfsummenliste.
//
// **Ohne Beigabe gibt es keins**, und das ist kein Fehler: eine Freigabe von
// vor diesem Werkstück trägt keines. Der Nullwert kommt zurück, und
// `domain.Vertraeglichkeit` macht daraus **ungeklärt**.
func (k *Klient) Manifest(
	ctx context.Context, freigabe Freigabe, pruefsummen []byte,
) (domain.Fassungsmanifest, error) {
	if freigabe.Manifest.Adresse == "" {
		return domain.Fassungsmanifest{}, nil
	}

	erwartet, gefunden := SummeAus(string(pruefsummen), Manifestname)
	if !gefunden {
		return domain.Fassungsmanifest{}, fmt.Errorf(
			"%w: die signierte Prüfsummenliste nennt %q nicht — das Manifest "+
				"gehört nicht zu dieser Freigabe", domain.ErrConflict, Manifestname)
	}

	antwort, err := netz.Hole(ctx, k.unten, k.drossel, k.beigabenkenner(),
		freigabe.Manifest.Adresse, HoechstensManifest)
	if err != nil {
		return domain.Fassungsmanifest{}, fmt.Errorf("das Manifest holen: %w", err)
	}

	if antwort.Status != netz.StatusGut {
		return domain.Fassungsmanifest{}, fmt.Errorf(
			"%w: das Manifest kam mit %d", domain.ErrConflict, antwort.Status)
	}

	if summe := sha256.Sum256(antwort.Rumpf); hex.EncodeToString(summe[:]) != erwartet {
		return domain.Fassungsmanifest{}, fmt.Errorf(
			"%w: das Manifest passt nicht zu seiner signierten Prüfsumme",
			domain.ErrConflict)
	}

	manifest, lesefehler := ManifestAus(antwort.Rumpf)
	if lesefehler != nil {
		return domain.Fassungsmanifest{}, lesefehler
	}

	// **Das Manifest muss dasselbe Abbild meinen wie die signierte Liste** —
	// Befund 6 des Zweitblicks über Phase 5.2. Es trug seinen Artefakthash,
	// und niemand verglich ihn: ein Manifest, das zu einer **anderen**
	// Freigabe gehört, sähe damit aus wie das dieser. Die Abnahme von O6 sagt
	// es ausdrücklich — „manipuliertes Manifest/Artefakt … werden vor
	// Bestandstausch abgewiesen".
	//
	// **Verglichen wird gegen die signierte Liste** und nicht gegen die
	// geladene Datei: die gibt es zu diesem Zeitpunkt noch nicht, und die
	// Liste ist die Aussage, hinter der eine Unterschrift steht.
	if fehler := abbildBinden(manifest, freigabe, pruefsummen); fehler != nil {
		return domain.Fassungsmanifest{}, fehler
	}

	// **Und dieselbe Fassung wie der Tag** — SEC03 (#2356), Vermerk 0194: eine
	// alte, echt signierte Liste trägt das alte Manifest, und das Urteil über
	// die Verträglichkeit fiele über eine andere Fassung als die angebotene.
	if manifest.Fassung != freigabe.Fassung {
		return domain.Fassungsmanifest{}, fmt.Errorf(
			"%w: das Manifest nennt die Fassung %s, angeboten ist %s",
			domain.ErrConflict, manifest.Fassung, freigabe.Fassung)
	}

	return manifest, nil
}

// ManifestAus liest ein Manifest aus seinem JSON.
//
// **Pflichtfelder streng, Zusatzfelder tolerant** (§16): ohne Fassung und
// Plattform ist es keine Aussage, und was künftige Fassungen zusätzlich
// hineinschreiben, wird übergangen.
func ManifestAus(roh []byte) (domain.Fassungsmanifest, error) {
	var manifest domain.Fassungsmanifest

	if err := json.Unmarshal(roh, &manifest); err != nil {
		// **Der Satz für Menschen, die Ursache ins Protokoll** (§8,
		// [0075](../../docs/entscheidungen/0075-der-satz-und-sein-grund.md)).
		// Der erste Anlauf setzte den JSON-Fehler in die Meldung; der
		// Linter wollte dafür ein zweites `%w`, und der Wächter über die
		// Meldungen wollte beides nicht. Recht hat der Wächter: eine
		// Deutungsmeldung von `encoding/json` sagt Nutzenden nichts und
		// verrät den inneren Aufbau.
		return domain.Fassungsmanifest{}, domain.MitGrund(domain.ErrValidation, err,
			"Das Manifest dieser Freigabe ließ sich nicht lesen.")
	}

	if manifest.Fassung == "" || manifest.Plattform == "" {
		return domain.Fassungsmanifest{}, fmt.Errorf(
			"%w: das Manifest nennt keine Fassung oder keine Plattform",
			domain.ErrValidation)
	}

	return manifest, nil
}

// abbildBinden hält das Manifest an das Abbild dieser Freigabe.
//
// **Der Vergleich ist Pflicht, sobald es ein Abbild gibt** — Befund 4 des
// Nachlaufs zum Zweitblick über Phase 5.2. Die erste Fassung verglich nur,
// wenn **beide** Angaben dastanden; ein Manifest ohne `artefakt` oder eine
// Liste ohne Abbildzeile kam damit ohne Vergleich durch, und der zugesagte
// Schutz war eine Wahl des Gegenübers.
//
// **Ohne Abbild in der Freigabe gibt es nichts zu binden.** Das ist der Weg
// der Vorprüfung, die nur das Manifest holt — sie lädt kein Abbild und nennt
// keines.
func abbildBinden(
	manifest domain.Fassungsmanifest, freigabe Freigabe, pruefsummen []byte,
) error {
	if freigabe.Abbild.Name == "" {
		return nil
	}

	if manifest.Artefakt == "" {
		return fmt.Errorf(
			"%w: das Manifest dieser Freigabe nennt kein Abbild — es lässt "+
				"sich nicht an das binden, was geladen wird", domain.ErrValidation)
	}

	abbild, genannt := SummeAus(string(pruefsummen), freigabe.Abbild.Name)
	if !genannt {
		return fmt.Errorf(
			"%w: die signierte Prüfsummenliste nennt %q nicht — das Manifest "+
				"lässt sich nicht binden", domain.ErrConflict, freigabe.Abbild.Name)
	}

	if manifest.Artefakt != abbild {
		return fmt.Errorf(
			"%w: das Manifest nennt ein anderes Abbild als die signierte "+
				"Prüfsummenliste", domain.ErrConflict)
	}

	return nil
}

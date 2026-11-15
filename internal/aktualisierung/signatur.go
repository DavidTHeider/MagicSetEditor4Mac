// **Die Signatur über den Prüfsummen** —
// #892,
// angestrichen vom Gesamtreview vor 0.2.0.
//
// # Was die Prüfsumme allein nicht leistet
//
// Sie sagt, dass Abbild und Prüfsummenliste **zueinander** passen. Woher
// beide kommen, sagt sie nicht: eine Gegenstelle, die beide Antworten
// beherrscht, liefert ein eigenes Abbild samt eigener Prüfsumme, und der
// Vergleich geht auf.
//
// **Die Signatur schließt genau diese Lücke**, und sie lag jeder Freigabe
// ohnehin bei: `scripts/release.sh` schreibt `pruefsummen.txt.asc` seit dem
// 2026-08-25 und prüft sie im selben Lauf. `docs/RELEASE.md` und `README.md`
// beschreiben sie als das, was den gemeinsamen Austausch verhindert — nur sah
// diese Anwendung sie nie an.
//
// # Der Schlüssel liegt im Programm, nicht im Schlüsselbund
//
// **Er ist eine Zusage über den Herausgeber, keine Einstellung.** Ein
// Schlüssel, den die Anwendung beim ersten Lauf von irgendwo holte, könnte
// derselbe Angreifer setzen, gegen den er schützen soll.
//
// **`go:embed` und nicht als Zeichenkette im Quelltext**: eine
// ASCII-gepanzerte Datei bleibt lesbar und vergleichbar mit dem, was
// `gpg --armor --export` gibt.

package aktualisierung

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/ProtonMail/go-crypto/openpgp"
)

// Signaturschluessel ist der Schlüssel, mit dem freigegeben wird.
//
// **Seit 1.0.0 ein eigener, getrennt vom Commitschlüssel** — wie Vermerk 0194
// es vorsah; bis 0.9.0 war es der Commitschlüssel (Vermerk 0205). Er steht
// hier als kurze Kennung, weil `gpg --local-user` sie erwartet; der volle
// Fingerabdruck kommt aus dem eingebetteten Schlüssel selbst
// ([Schluesselfingerabdruck]) und wird nicht danebengeschrieben — zwei
// Angaben über dieselbe Sache laufen auseinander.
const Signaturschluessel = "44C59428DE1F766E"

// freigabeschluessel ist der öffentliche Schlüssel des Herausgebers.
//
//go:embed freigabeschluessel.asc
var freigabeschluessel string

// freigabeanker sind die Fingerabdrücke der Schlüssel, deren Signatur gilt —
// SEC03 (#2356), Vermerk 0194.
//
// **Ein Anker heute, zwei im Übergang.** Wechselt der Freigabeschlüssel,
// trägt eine Fassung den alten **und** den neuen: die installierten nehmen
// sie an, weil der alte sie signiert, und sie selbst nimmt die nächste an,
// die der neue signiert. Danach steht nur der neue hier.
//
// **Benannt und nicht bloß eingebettet**, damit ein Schlüssel, den jemand in
// die Datei legt, nicht still mitgilt: der Bund muss genau diese Anker
// führen. Ein widerrufener Schlüssel darf daneben stehen — er trägt seinen
// Widerruf, und seine Signatur gilt nicht.
//
// **Eine Veränderliche nur für die Matrix** (`anker_intern_test.go`).
var freigabeanker = []string{"97721629ABDC00BA1711488244C59428DE1F766E"}

// HoechstensSignatur ist die Obergrenze für `pruefsummen.txt.asc`.
//
// **Eine abgetrennte EdDSA-Signatur ist rund 230 Byte.** 64 KiB sind das
// Hundertfache und trotzdem eine Grenze.
const HoechstensSignatur = 64 << 10

// pruefbund ist der Schlüsselbund, gegen den geprüft wird.
//
// **Eine Veränderliche, damit ein Test einen eigenen Schlüssel einsetzen
// kann** — dieselbe Naht wie [MitFreigabeadresse], und aus demselben Grund:
// sonst bräuchte jeder Ladeversuch ein `gpg` auf dem Rechner und den privaten
// Schlüssel des Herausgebers.
//
// **Nicht, damit die Prüfung nachgiebiger wird.** Ein Test setzt einen
// **echten** Schlüssel ein und signiert damit wirklich; was er nicht tut, ist
// die Prüfung abschalten. Eine Vorrichtung, die großzügiger ist als der
// Betrieb, prüft den falschen Weg (§3).
//
// Im Betrieb steht hier `nil`, und dann gilt der eingebettete Schlüssel.
var pruefbund openpgp.EntityList

// MitFreigabebund setzt den Schlüsselbund — **nur für Tests**.
//
// Sie gibt zurück, was vorher dastand, damit ein Aufrufer ihn zurückstellen
// kann.
func MitFreigabebund(neuer openpgp.EntityList) (vorher openpgp.EntityList) {
	vorher, pruefbund = pruefbund, neuer

	return vorher
}

// bund gibt den Schlüsselbund — den eingesetzten oder den eingebetteten.
func bund() (openpgp.EntityList, error) {
	if pruefbund != nil {
		return pruefbund, nil
	}

	return eingebetteterBund()
}

// eingebetteterBund liest den eingebetteten Schlüssel.
func eingebetteterBund() (openpgp.EntityList, error) {
	gelesen, err := openpgp.ReadArmoredKeyRing(strings.NewReader(freigabeschluessel))
	if err != nil {
		return nil, fmt.Errorf("den eingebetteten Freigabeschlüssel lesen: %w", err)
	}

	return ankerPruefen(gelesen, freigabeanker, time.Now())
}

// ankerPruefen hält den Bund gegen die benannten Anker: **jeder Anker steht
// darin und ist nicht widerrufen, und jeder andere Schlüssel ist
// widerrufen.**
//
// Hier stand „genau ein Schlüssel" (#892) — aus demselben Grund, aus dem
// die Anker jetzt benannt sind: ein Bund, der wächst, weil jemand einen
// zweiten Schlüssel dazulegte, ist die Änderung, die niemand bemerkt. Für
// einen Wechsel des Schlüssels braucht es zwei, und deshalb steht die Zahl
// jetzt nicht mehr im Prüfer, sondern als Liste daneben.
func ankerPruefen(bund openpgp.EntityList, anker []string, jetzt time.Time) (openpgp.EntityList, error) {
	if len(anker) == 0 {
		return nil, fmt.Errorf("%w: es ist kein Freigabeanker benannt", domain.ErrValidation)
	}

	gefunden := map[string]bool{}

	for _, wer := range bund {
		abdruck := strings.ToUpper(fmt.Sprintf("%X", wer.PrimaryKey.Fingerprint))
		benannt := slices.Contains(anker, abdruck)

		switch {
		case wer.Revoked(jetzt) && benannt:
			return nil, fmt.Errorf(
				"%w: der Freigabeanker %s ist widerrufen", domain.ErrValidation, abdruck)
		case wer.Revoked(jetzt):
			continue
		case !benannt:
			return nil, fmt.Errorf(
				"%w: der eingebettete Schlüsselbund führt %s, und das ist kein "+
					"benannter Freigabeanker", domain.ErrValidation, abdruck)
		}

		gefunden[abdruck] = true
	}

	for _, abdruck := range anker {
		if !gefunden[abdruck] {
			return nil, fmt.Errorf(
				"%w: der Freigabeanker %s fehlt im eingebetteten Schlüsselbund",
				domain.ErrValidation, abdruck)
		}
	}

	return bund, nil
}

// Schluesselfingerabdruck gibt den Fingerabdruck des eingebetteten Schlüssels.
//
// **Für den Wächter, der ihn gegen AGENTS.md hält.** Ein Programm, das eine
// Signatur gegen einen Schlüssel prüft, den niemand kennt, sieht aus wie
// Sicherheit und ist keine.
func Schluesselfingerabdruck() (string, error) {
	gelesen, err := eingebetteterBund()
	if err != nil {
		return "", err
	}

	// **Der Anker, mit dem heute signiert wird** — im Übergang stehen zwei
	// im Bund, und dieser ist der, auf den [Signaturschluessel] endet.
	for _, wer := range gelesen {
		abdruck := strings.ToUpper(fmt.Sprintf("%X", wer.PrimaryKey.Fingerprint))
		if strings.HasSuffix(abdruck, Signaturschluessel) && !wer.Revoked(time.Now()) {
			return abdruck, nil
		}
	}

	return "", fmt.Errorf("%w: kein Freigabeanker endet auf %s",
		domain.ErrValidation, Signaturschluessel)
}

// SignaturPruefen prüft die abgetrennte Signatur über den Prüfsummen.
//
// **Öffentlich, damit ein Test die Naht zum Freigabeweg prüfen kann**:
// `scripts/release.sh` schreibt die Signatur mit `gpg`, dieses Paket liest sie
// — zwei Programme in zwei Sprachen, und zwischen ihnen eine Naht, die sonst
// niemand ansieht (§10).
func SignaturPruefen(pruefsummen, signatur []byte) error {
	gelesen, err := bund()
	if err != nil {
		return err
	}

	_, err = openpgp.CheckArmoredDetachedSignature(gelesen,
		strings.NewReader(string(pruefsummen)),
		strings.NewReader(string(signatur)), nil)
	if err != nil {
		return domain.MitGrund(domain.ErrConflict, err,
			"die Signatur über den Prüfsummen passt nicht zum Freigabeschlüssel "+
				"%s — es wird nichts geladen", Signaturschluessel)
	}

	return nil
}

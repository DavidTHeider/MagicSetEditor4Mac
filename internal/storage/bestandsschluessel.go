// **Wo der Schlüssel des Bestands liegt** — Werkstück 14 von Phase 4.12
// (#1351),
// nach [0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md).

package storage

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// Schluesselbunddienst ist der Dienstname, unter dem der Schlüssel liegt.
//
// Dieselbe Form wie `MagicSetEditor4Mac-Anmeldung` in `internal/app` — ein
// eigener Name je Bereich, damit ein Blick in den Schlüsselbund sagt, wozu ein
// Eintrag gehört.
const Schluesselbunddienst = "MagicSetEditor4Mac-Bestand"

// Schluesselkonto ist der Kontoname des Bestandsschlüssels.
//
// **Einer je Schlüsselbund und nicht einer je Datei.** Ein Kontoname aus dem
// Pfad wäre naheliegend und falsch: wer seinen Bestand verschiebt, verlöre den
// Schlüssel und damit die Daten. Wer **zwei** Bestände führt, gibt ihnen zwei
// Bünde — dieselbe Form, die `MSE_SCHLUESSELBUND` schon anbietet.
const Schluesselkonto = "schluessel"

// schluesselleser holt den Schlüssel aus dem Schlüsselbund.
//
// **Eine Naht und kein direkter Ruf**, damit ein Lauf nicht den Schlüsselbund
// dieses Rechners anfasst (§3). Wer sie nicht ersetzt, bekommt den echten —
// und das ist richtig so: der Betrieb soll nichts einrichten müssen.
var schluesselleser = func(ctx context.Context) (string, error) {
	return schluesselbund.Anmeldebund().Lies(ctx, Schluesselbunddienst, Schluesselkonto)
}

// SchluesselSchreiben legt den Bestandsschlüssel im Schlüsselbund ab.
func SchluesselSchreiben(ctx context.Context, schluessel string) error {
	return schluesselbund.Anmeldebund().
		Schreibe(ctx, Schluesselbunddienst, Schluesselkonto, schluessel)
}

// SchluesselLoeschen nimmt ihn wieder heraus.
func SchluesselLoeschen(ctx context.Context) error {
	return schluesselbund.Anmeldebund().
		Loesche(ctx, Schluesselbunddienst, Schluesselkonto)
}

// schluesselFuer sagt, mit welchem Schlüssel ein Bestand aufzumachen ist.
//
// # Der Dateikopf entscheidet, ob überhaupt gefragt wird
//
// **Ein unverschlüsselter Bestand fasst den Schlüsselbund nicht an.** Das ist
// nicht nur schnell, es ist die Bedingung dafür, dass Läufe und ein frischer
// Start ohne Einrichtung funktionieren: `security` fragt bei manchen
// Fehlschlägen nach einem Kennwort und **wartet dann** (siehe
// `schluesselbund.UeblicheFrist`).
//
// # Und wo es keine Datei gibt, gibt es keinen Schlüssel
//
// Eine Datenbank im Arbeitsspeicher und eine `file:`-URI mit Parametern haben
// keinen Ort — dieselbe Unterscheidung, die `sperrpfad` trifft, und aus
// demselben Grund.
func schluesselFuer(ctx context.Context, bestandspfad string) (string, error) {
	if bestandspfad == "" || bestandspfad == ImArbeitsspeicher ||
		strings.HasPrefix(bestandspfad, "file:") {
		return "", nil
	}

	verschluesselt, err := IstVerschluesselt(bestandspfad)
	if err != nil {
		return "", err
	}

	if !verschluesselt {
		return "", nil
	}

	schluessel, err := schluesselleser(ctx)
	if err != nil {
		// **Die Auskunft, die ein Mensch braucht** (§8). Ohne sie käme
		// `SQLITE_NOTADB` an, und das heißt in dieser Anwendung „die Datei ist
		// kein Bestand" — woraufhin jemand sie ersetzt. Sie ist in Ordnung;
		// es fehlt der Schlüssel.
		return "", domain.MitGrund(domain.ErrSchluesselFehlt, err,
			"%q ist verschlüsselt, und der Schlüssel liegt nicht im "+
				"Schlüsselbund dieses Rechners — er wurde beim Einschalten "+
				"der Verschlüsselung dort abgelegt und ist nicht "+
				"wiederherstellbar", filepath.Base(bestandspfad))
	}

	return schluessel, nil
}

// mitHuelle hängt das Chiffre-VFS an eine Datenquelle, wenn es einen Schlüssel
// gibt.
func mitHuelle(quelle, schluessel string) string {
	if schluessel == "" {
		return quelle
	}

	trenner := "?"
	if strings.Contains(quelle, "?") {
		trenner = "&"
	}

	return quelle + trenner + "vfs=" + Chiffrehuelle
}

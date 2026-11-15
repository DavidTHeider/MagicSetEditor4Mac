package app

// **Eine Leihe im Umschlag** — Werkstück G6 von Phase 5.5
// (#1463).
//
// **Der Weg hinaus und herein**, getrennt von den Bindungen daneben: jene
// beantworten „was darf ein Fenster rufen", diese „wie reist eine Leihe"
// (§2, der Schnitt folgt der Frage).

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die beiden Namen, unter denen eine Leihe im Umschlag liegt.
//
// **Zwei Objekte und nicht eines.** Das Dokument ist unterschrieben und darf
// sich nicht ändern; der Abzug ist der Inhalt, den es meint. Beide in einer
// Datei hießen, dass man das eine nicht ohne das andere prüfen kann.
const (
	Leihdokumentname = "leihe.json"
	Leihabzugname    = "abzug.json"
)

// leiheAblegen stellt die Leihe aus und schreibt den Umschlag.
func (a *App) leiheAblegen(
	ctx context.Context, wunsch LeihwunschDTO, ziel string,
) (string, error) {
	deckID, fehler := strconv.ParseInt(wunsch.DeckID, 10, 64)
	if fehler != nil {
		return "", domain.MitGrund(
			domain.ErrValidation, fehler, "die Deckkennung deuten")
	}

	ich, fehler := a.peerwerk().Ausweis(ctx)
	if fehler != nil {
		return "", fehler
	}

	leihwunsch, fehler := wunschAus(wunsch, deckID)
	if fehler != nil {
		return "", fehler
	}

	leihe, abzug, fehler := a.leihstelle().Anbieten(ctx, leihwunsch)
	if fehler != nil {
		return "", fehler
	}

	inhalte := leihinhalte(leihe, abzug)

	if _, fehler := austausch.Schreibe(ziel, inhalte, ich); fehler != nil {
		return "", fehler
	}

	return ziel, nil
}

// leiheAnnehmen liest einen Umschlag und legt seine Leihe ab.
func (a *App) leiheAnnehmen(ctx context.Context, pfad string) error {
	unterwegs, fehler := os.MkdirTemp("", "mse-leihe-")
	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"ein Verzeichnis für die Leihe anlegen")
	}

	// **Der Wegwerfort wird geräumt, gleich wie es ausgeht** — das Issue
	// verlangt es: „Transferabbruch entfernt temporäre Reste."
	defer func() { _ = os.RemoveAll(unterwegs) }() //nolint:errcheck // Wegwerfort

	sendung, fehler := austausch.Lies(pfad, unterwegs)
	if fehler != nil {
		return fehler
	}

	leihe, abzug, fehler := leiheAusSendung(sendung.Verzeichnis)
	if fehler != nil {
		return fehler
	}

	return a.leihstelle().Annehmen(ctx, leihe, abzug)
}

// leiheAusSendung liest die beiden Objekte aus dem geprüften Verzeichnis.
func leiheAusSendung(verzeichnis string) (
	domain.Deckleihe, domain.Partiedeck, error,
) {
	var leihe domain.Deckleihe

	var abzug domain.Partiedeck

	for name, ziel := range map[string]any{
		Leihdokumentname: &leihe, Leihabzugname: &abzug,
	} {
		roh, fehler := os.ReadFile(filepath.Join(verzeichnis, name)) // #nosec G304 -- geprüftes Quarantäneverzeichnis
		if fehler != nil {
			return domain.Deckleihe{}, domain.Partiedeck{}, domain.MitGrund(
				domain.ErrValidation, fehler,
				"dieser Umschlag trägt keine Leihe")
		}

		if fehler := json.Unmarshal(roh, ziel); fehler != nil {
			return domain.Deckleihe{}, domain.Partiedeck{}, domain.MitGrund(
				domain.ErrValidation, fehler,
				"die Leihe in diesem Umschlag ließ sich nicht lesen")
		}
	}

	return leihe, abzug, nil
}

// leihinhalte macht aus Dokument und Abzug die beiden Umschlagobjekte.
//
// **Ohne Fehlerweg**, und der Grund steht bei `storage.kodiereLeihe`: was
// hineingeht, sind Zeichenketten, Zahlen, Zeitpunkte und Abbildungen davon;
// `encoding/json` kann daran nicht scheitern. Ein Zweig dafür stünde
// unbetreten da und sähe aus wie eine geprüfte Zusage (§10).
func leihinhalte(
	leihe domain.Deckleihe, abzug domain.Partiedeck,
) []austausch.Inhalt {
	rohesDokument, _ := json.Marshal(leihe) //nolint:errcheck,errchkjson // siehe oben
	roherAbzug, _ := json.Marshal(abzug)    //nolint:errcheck,errchkjson // siehe oben

	return []austausch.Inhalt{
		{
			Art: austausch.ArtLeihe, Name: Leihdokumentname,
			Bezeichnung: "Deckleihe", Bytes: rohesDokument,

			// **Das Dokument braucht den Abzug.** Ohne ihn nennt es einen
			// Abdruck, zu dem es nichts gibt — und der Umschlag sagt das,
			// statt dass es beim Annehmen auffällt.
			Braucht: []string{Leihabzugname},
		},
		{
			Art: austausch.ArtLeihe, Name: Leihabzugname,
			Bezeichnung: "Eingefrorener Abzug", Bytes: roherAbzug,
		},
	}
}

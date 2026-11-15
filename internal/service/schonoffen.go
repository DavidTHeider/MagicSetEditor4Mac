// **Eine Quelle, ein Set** — die Regel und ihr Pfadvergleich.
//
// # Warum sie hier steht und nicht in einer Fassade
//
// In der Fassade **des Fensters** ginge die Kommandozeile
// (#886) an
// ihr vorbei, riefe den Importeur unmittelbar und legte dasselbe Set ein
// zweites Mal an — der REST-Dienst ebenso.
//
// **Eine fachliche Regel in einer Fassade ist eine Regel je Fassade.**
//
// **Eine eigene Datei**, weil `import.go` mit ihr über 400 Zeilen ginge (§2).

package service

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kopfliste nennt die Sets, die schon im Bestand liegen.
//
// **Eine eigene und nicht [Kopfquelle]**: jene verlangt zusätzlich die
// Setwerte, und der Importeur braucht sie nicht. Eine Schnittstelle sagt, was
// der Aufrufer **braucht** (§2) — wer mehr verlangt, verengt den Kreis derer,
// die sie erfüllen können.
type Kopfliste interface {
	Setkoepfe(ctx context.Context) ([]domain.SetHead, error)
}

// MitKoepfen setzt die Kopfquelle — und damit die Regel „eine Quelle, ein Set".
//
// **Nil heißt: die Regel gilt nicht.** Ein Test, der nur den Weg durch das
// Paket prüft, braucht keinen Bestand; und ein Aufrufer, der ein zweites Set
// aus derselben Datei will, sagt das, statt die Regel zu erben.
func (i *Importeur) MitKoepfen(koepfe Kopfliste) *Importeur {
	i.koepfe = koepfe

	return i
}

// schonOffen sagt, ob diese Quelle bereits im Bestand liegt.
//
// **Der Pfad wird aufgelöst, bevor er verglichen wird.** Ein Vergleich als
// Zeichenkette hielte `/tmp/a.mse-set` und einen Alias darauf für zwei Dateien
// — und legte dieselben Karten zweimal an. Angestrichen vom Nachlauf des
// Zweitblicks über Phase 9.
//
// **Ein Fehler beim Auflösen ist keiner**: dann gilt der Pfad, wie er kam. Wer
// eine Datei öffnet, die es nicht gibt, erfährt das gleich darauf beim Öffnen
// — mit einer Meldung, die von der Datei handelt und nicht von symbolischen
// Verweisen.
//
// **Ein Fehler beim Nachsehen ist einer**: wer ihn überginge, legte ein Set
// zum zweiten Mal an, weil der Bestand gerade nicht lesbar war.
func (i *Importeur) schonOffen(ctx context.Context, pfad string) (*Ergebnis, error) {
	if i.koepfe == nil {
		return nil, nil //nolint:nilnil // „keine Regel, kein Fehler" — der Aufrufer prüft auf nil
	}

	echterPfad := pfad
	if aufgeloest, err := filepath.EvalSymlinks(pfad); err == nil {
		echterPfad = aufgeloest
	}

	koepfe, err := i.koepfe.Setkoepfe(ctx)
	if err != nil {
		return nil, fmt.Errorf("den Bestand nachsehen: %w", err)
	}

	for _, kopf := range koepfe {
		if gleicheQuelle(kopf.Origin, echterPfad) {
			return &Ergebnis{
				SetID:      kopf.ID,
				Karten:     kopf.Cards,
				Befunde:    []domain.Befund{},
				Angelegt:   []int64{},
				SchonOffen: true,
			}, nil
		}
	}

	return nil, nil //nolint:nilnil // siehe oben
}

// gleicheQuelle sagt, ob zwei Pfade auf dieselbe Datei zeigen.
//
// **Beide Richtungen.** Der Bestand kann den Verweis kennen und die echte
// Datei gemeint sein — oder umgekehrt. Wer nur den gesuchten Pfad auflöste,
// sähe im ersten Fall zwei Dateien.
func gleicheQuelle(gespeichert, gesucht string) bool {
	if gespeichert == "" {
		return false
	}

	if gespeichert == gesucht {
		return true
	}

	aufgeloest, err := filepath.EvalSymlinks(gespeichert)

	return err == nil && aufgeloest == gesucht
}

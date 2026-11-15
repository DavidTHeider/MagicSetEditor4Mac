package service

// **Die Auskunft über Stände und das Aufheben** — G04
// (#2251).
//
// [Paketbezieher] erfüllt [Quellauskunft]: er weiß, welche Gestalt eine
// Quelle hat, und `internal/bezug` weiß, wie man fragt.

import (
	"context"
	"net/url"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// githubauskunft ist die Wurzel von GitHubs Auskunft über Repositorien —
// derselbe Dienst, bei dem [markenadresse] eine Freigabe nachschlägt.
const githubauskunft = "https://api.github.com/repos/"

// Spitze fragt, auf welchen Commit der Hauptzweig einer Quelle gerade zeigt.
func (p *Paketbezieher) Spitze(ctx context.Context, adresse string) (bezug.Spitzenstand, error) {
	wurzel, err := p.auskunftsadresse(adresse)
	if err != nil {
		return bezug.Spitzenstand{}, err
	}

	return bezug.Spitze(ctx, wurzel+"/commits/HEAD", p.kenner)
}

// Vergleich fragt, was zwischen zwei Ständen einer Quelle liegt.
func (p *Paketbezieher) Vergleich(ctx context.Context, adresse, von, bis string) (bezug.Vergleich, error) {
	wurzel, err := p.auskunftsadresse(adresse)
	if err != nil {
		return bezug.Vergleich{}, err
	}

	vergleich, err := bezug.Vergleiche(ctx, wurzel+"/compare/"+url.PathEscape(von)+"..."+url.PathEscape(bis), p.kenner)
	if err != nil || !vergleich.Gekuerzt {
		return vergleich, err
	}

	return p.ausDenBaeumen(ctx, wurzel, von, bis, vergleich), nil
}

// ausDenBaeumen ergänzt einen gekürzten Vergleich aus den Bäumen beider
// Stände (`bezug.Baum`).
//
// **Gefunden am nativen Fenster** (`g04bezuege`): ein Commit brachte 435
// Dateien, der Vergleich nannte 300, und die Vorschau nannte nur einen Teil
// der Pakete — und damit nur einen Teil der Sets, die sie nennen.
//
// **Scheitert eine der beiden Fragen oder ist ein Baum selbst gekürzt, bleibt
// der Vergleich, wie er war** — gekürzt und als solcher markiert, und die
// Vorschau verlangt dann eine Bestätigung. Ein halber Baum sähe vollständig
// aus; ein Fehler hier nähme der Vorschau, was sie schon weiß.
func (p *Paketbezieher) ausDenBaeumen(ctx context.Context, wurzel, von, bis string, vergleich bezug.Vergleich) bezug.Vergleich {
	baeume := make([]bezug.Baumstand, 0, 2)

	for _, stand := range []string{von, bis} {
		baum, err := bezug.Baum(ctx, wurzel+"/git/trees/"+url.PathEscape(stand)+"?recursive=1", p.kenner)
		if err != nil || baum.Gekuerzt {
			return vergleich
		}

		baeume = append(baeume, baum)
	}

	vergleich.Dateien, vergleich.Gekuerzt = bezug.Baumvergleich(baeume[0], baeume[1]), false

	return vergleich
}

// auskunftsadresse macht aus einer Repositoriumsadresse die der Auskunft.
//
// **Nur GitHub nennt einen neueren Stand** — die Vorlagen liegen dort, und
// eine andere Gegenstelle hat eine andere Auskunft. Eine solche Quelle ist
// nicht still „aktuell", sondern sagt, dass hier nicht gefragt werden kann.
func (p *Paketbezieher) auskunftsadresse(adresse string) (string, error) {
	kern := strings.TrimSuffix(strings.TrimSuffix(adresse, "/"), ".git")

	rest, ist := strings.CutPrefix(kern, "https://github.com/")
	if !ist || strings.Count(rest, "/") != 1 || strings.HasPrefix(rest, "/") || strings.HasSuffix(rest, "/") {
		return "", domain.NewFieldError("adresse", domain.SchluesselValidierung,
			"Bei %s lässt sich kein neuerer Stand erfragen: das kann hier nur GitHub.", adresse)
	}

	wurzel := p.api
	if wurzel == "" {
		wurzel = githubauskunft
	}

	return wurzel + rest, nil
}

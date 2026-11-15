package service

// **Die gemeinsame Recherche** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Vier Fälle über einen Index
//
// **A16 baut keinen zweiten Retrievalkern** — das Issue sagt es ausdrücklich:
// „A16 baut keine Indizes/Encoder/Fusion erneut." Was hier steht, ist die
// **Fachnaht**: welche Korpora ein Fall durchsucht, was als Treffer zählt,
// und was eine leere Antwort bedeutet.
//
// # Ohne Generierung
//
// **Reine Recherche fragt kein Sprachmodell.** Das Issue nennt es im Ziel
// und noch einmal in der Abnahme: die Kartenfrage „liefert im eingefrorenen
// Referenzbestand Quellenkarten **ohne Textgenerierungsaufruf**". Ein
// Umschreiben der Frage durch ein Modell wäre genau der unsichtbare Schritt,
// den das Issue verbietet.

import (
	"context"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Recherchequelle ist, was die Recherche vom Retrieval braucht.
//
// **Zwei Methoden**, und beide gibt es schon: [Retrievalverwaltung] erfüllt
// sie. Eine eigene Schnittstelle steht hier, damit ein Lauf die Recherche
// ohne Index prüfen kann.
type Recherchequelle interface {
	Stand(ctx context.Context) Indexstand
	Suchen(ctx context.Context, wunsch Suchwunsch) (Suchergebnis, error)
}

// Profilquelle sagt, für wen gesucht wird.
//
// # Warum der Dienst das wissen muss
//
// **Weil der Fortsetzungszeiger es bindet.** Das Issue: „Begrenzte
// Pagination/Cursor binden Anfrage, **Profil**, Rechte, Generation und
// Ablauf." Wer das Profil wechselt, bekommt eine Absage statt der zweiten
// Seite einer Liste, die für jemand anderen entstand.
//
// **Sie kommt aus dem Kontext und steht trotzdem hier**: die Kennung liest
// `storage`, und `service` importiert `storage` nicht (§2).
type Profilquelle interface {
	Profil(ctx context.Context) (int64, error)
}

// Ausgangsquelle liefert den Text einer Karte, zu der Ähnliches gesucht wird.
//
// **Mit ihrem Stand**, und der wird geprüft: eine Karte, die sich seit der
// Anfrage geändert hat, ist eine andere Frage.
type Ausgangsquelle interface {
	// Ausgangstext gibt den Text, den **wirklichen Stand** und die
	// Kennungen der Abschnitte, die zu dieser Karte gehören.
	//
	// **Der Stand kommt zurück, auch wenn er hineinging**: der Aufrufer
	// nennt ihn wahlweise, und was gesucht wurde, sagt erst die Antwort.
	Ausgangstext(ctx context.Context, karteID int64, stand string) (
		text, jetziger string, kennungen []string, fehler error)
}

// Recherchefenster ist der Faktor, um den weiter gefragt wird als geantwortet.
//
// **Weil gebündelt und ausgeschlossen wird.** Zwei Seiten derselben Karte
// werden ein Treffer, und die Ausgangskarte fällt heraus; ohne Zuschlag
// käme die letzte Seite zu kurz. **Was er nicht leistet**, steht bei
// [Kandidatenfenster]: er verschiebt die Grenze, er hebt sie nicht auf.
const Recherchefenster = 3

// Recherchedienst beantwortet die vier Fälle.
type Recherchedienst struct {
	quelle  Recherchequelle
	ausgang Ausgangsquelle
	profil  Profilquelle

	// jetzt ist die Uhr — eingesetzt, damit ein Lauf den Ablauf eines
	// Zeigers prüfen kann, ohne zu warten.
	jetzt func() time.Time
}

// NeuerRecherchedienst baut ihn.
func NeuerRecherchedienst(
	quelle Recherchequelle, ausgang Ausgangsquelle, profil Profilquelle,
) *Recherchedienst {
	return &Recherchedienst{
		quelle: quelle, ausgang: ausgang, profil: profil, jetzt: time.Now,
	}
}

// MitUhr setzt die Uhr.
func (d *Recherchedienst) MitUhr(jetzt func() time.Time) *Recherchedienst {
	d.jetzt = jetzt

	return d
}

// Recherchieren beantwortet eine Anfrage.
func (d *Recherchedienst) Recherchieren(
	ctx context.Context, anfrage domain.Rechercheanfrage, fortsetzung string,
) (domain.Rechercheergebnis, error) {
	if fehler := anfrage.Pruefe(); fehler != nil {
		return domain.Rechercheergebnis{}, fehler
	}

	stand := d.quelle.Stand(ctx)

	if !stand.Aktiv {
		return ohneIndex(stand), nil
	}

	if fehler := pruefeKorpusstand(anfrage, stand); fehler != nil {
		return domain.Rechercheergebnis{}, fehler
	}

	frage, ausgangsstand, ausgeschlossen, fehler := d.frageFuer(ctx, anfrage)
	if fehler != nil {
		return domain.Rechercheergebnis{}, fehler
	}

	// **Der gerechnete Stand geht in die Anfrage, bevor der Zeiger entsteht.**
	// Sonst bände er den leeren — und die zweite Seite gehörte zu einer
	// Karte, die es so nicht mehr gibt (Issue: „keine still gemischten
	// Seiten").
	anfrage.Ausgangsstand = ausgangsstand

	abdruck, fehler := d.abdruckFuer(ctx, anfrage, stand)
	if fehler != nil {
		return domain.Rechercheergebnis{}, fehler
	}

	stelle, fehler := stelleAus(fortsetzung, abdruck, d.jetzt())
	if fehler != nil {
		return domain.Rechercheergebnis{}, fehler
	}

	ergebnis, fehler := d.quelle.Suchen(ctx, Suchwunsch{
		Frage:  frage,
		Grenze: (stelle + anfrage.Grenze) * Recherchefenster,
		Filter: Retrievalfilter{
			// **Die Suchquellen und nicht die des Falls** (A13.4, #1452):
			// eine private Quelle ohne ausdrückliche Freigabe bleibt außen,
			// und die Antwort sagt es
			// ([domain.Rechercheanfrage.AusgelasseneQuellen]).
			Quellen:  anfrage.Suchquellen(),
			Sprachen: sprachenAus(anfrage.Sprache),
		},
	})
	if fehler != nil {
		return domain.Rechercheergebnis{}, fehler
	}

	return d.antwortAus(ergebnis, anfrage, stand, ausgeschlossen, stelle,
		abdruck), nil
}

// frageFuer gibt den Text, nach dem gesucht wird, und was auszuschließen ist.
func (d *Recherchedienst) frageFuer(
	ctx context.Context, anfrage domain.Rechercheanfrage,
) (frage, ausgangsstand string, ausgeschlossen []string, fehler error) {
	if anfrage.Fall != domain.FallAehnlich {
		return anfrage.Frage, "", nil, nil
	}

	if d.ausgang == nil {
		return "", "", nil, domain.MitGrund(domain.ErrValidation, nil,
			"zu dieser Fassung lässt sich keine Ausgangskarte lesen")
	}

	// **Der Stand wird an die Quelle gereicht und dort geprüft.** Wer ihn
	// hier prüfte, prüfte eine Zahl gegen eine Zahl; wer ihn dort prüft,
	// prüft ihn gegen die Karte.
	text, jetziger, kennungen, fehler := d.ausgang.Ausgangstext(ctx,
		anfrage.AusgangskarteID, anfrage.Ausgangsstand)
	if fehler != nil {
		return "", "", nil, fehler
	}

	return text, jetziger, kennungen, nil
}

// stelleAus liest den Fortsetzungszeiger — oder gibt null.
func stelleAus(fortsetzung, abdruck string, jetzt time.Time) (int, error) {
	if fortsetzung == "" {
		return 0, nil
	}

	zeiger, fehler := domain.ZeigerAus(fortsetzung, abdruck, jetzt)
	if fehler != nil {
		return 0, fehler
	}

	return zeiger.Stelle, nil
}

// abdruckFuer bindet Anfrage, Profil und Indexstand.
func (d *Recherchedienst) abdruckFuer(
	ctx context.Context, anfrage domain.Rechercheanfrage, stand Indexstand,
) (string, error) {
	profil, fehler := d.profil.Profil(ctx)
	if fehler != nil {
		return "", fehler
	}

	return domain.Zeigerabdruck(anfrage, profil, indexabdruck(stand)), nil
}

// indexabdruck bindet Verfahren **und** Korpusstand.
//
// # Warum beides
//
// **Weil der Fingerabdruck des Manifests den Korpus absichtlich auslässt**
// ([domain.Indexmanifest.Fingerabdruck]): ein Dokument mehr ändert nicht, wie
// verglichen wird. Für eine **Seite** gilt das Gegenteil — sie ist eine
// Stelle in einer Liste, und die Liste ändert sich mit dem Korpus.
//
// **Gesehen im ersten Anlauf**: ein Zeiger überlebte einen Neuaufbau mit
// anderem Korpusstand, und die zweite Seite gehörte zu einer Liste, die es
// nicht mehr gab. Das ist die „still gemischte Seite", die das Issue
// ausschließt.
func indexabdruck(stand Indexstand) string {
	teile := make([]string, 0, len(stand.Manifest.Korpusstaende)+1)
	teile = append(teile, stand.Manifest.Fingerabdruck())

	for _, korpus := range stand.Manifest.Korpusstaende {
		teile = append(teile, string(korpus.Quelle)+"="+korpus.Revision)
	}

	return strings.Join(teile, "|")
}

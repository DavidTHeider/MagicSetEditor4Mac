package app

// **Die Recherche am Fenster** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Zwei Rufe und kein dritter
//
// **Recherchieren und Zusammenstellen.** Das Issue trennt sie scharf: „Zwei
// getrennte Wege: Suche → Treffer ohne Generierung; Kontext → ausdrücklich
// ausgelöster A6-Auftrag → validierter Vorschlag."
//
// **Der dritte Weg steht in A6 und A11** — generieren und übernehmen. Hier
// wird nichts gerufen, was Geld kostet, und nichts geschrieben.
//
// # „Kein Modell" heißt „kein generatives Modell"
//
// **Das Einbettungsmodell rechnet sehr wohl mit** — es muss: die Frage gehört
// in denselben Vektorraum wie der Index, sonst gäbe es keine semantische
// Hälfte. Es läuft auf diesem Rechner, es kostet nichts, und es schreibt
// keinen Text.
//
// **Der Unterschied ist nicht Wortklauberei.** Ohne Einbettungsmodell fällt
// die Suche sichtbar auf den Volltext zurück (`leere: ohne-modell`); ohne
// generatives Modell fehlt hier gar nichts.
//
// # Der Modus steht in jeder Antwort
//
// **Und der Grund für eine leere ebenso.** „Nichts gefunden" heißt sechs
// verschiedene Dinge, und fünf davon sagen jemandem, was zu tun ist.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Recherchefuehrer ist die Naht zum Dienst.
type Recherchefuehrer interface {
	Recherchieren(ctx context.Context, anfrage domain.Rechercheanfrage,
		fortsetzung string) (domain.Rechercheergebnis, error)
	Zusammenstellen(ctx context.Context, wunsch service.Kontextwunsch) (
		service.Kontextpaket, error)
}

// RecherchefilterDTO ist die geprüfte Einschränkung einer Recherche.
type RecherchefilterDTO struct {
	// Sprache schränkt auf eine ein — leer heißt: jede.
	Sprache string `json:"sprache"`

	// Korpusstand verlangt einen bestimmten Stand — leer heißt: der aktive.
	Korpusstand string `json:"korpusstand"`

	// Freigegeben nennt die **privaten** Korpora, die ein Mensch ausdrücklich
	// ausgewählt hat — Werkstück A13.4 von Phase 5.3
	// (#1452).
	//
	// # Warum sie hier steht, obwohl sie kein Filter ist
	//
	// **Weil sie an derselben Stelle gebraucht wird und dieselbe Reise
	// macht.** Ein Filter engt ein, eine Freigabe weitet aus — der
	// Unterschied ist echt, und deshalb steht er hier. Was sie **nicht** darf,
	// ist stillschweigend fehlen: leer heißt „keine", der eigene Bestand
	// bleibt außen, und die Antwort sagt es unter `einschraenkungen`.
	//
	// Eine Freigabe, die nichts bewirkt, wird **abgewiesen** und nicht
	// übergangen ([domain.Rechercheanfrage.Pruefe]).
	Freigegeben []string `json:"freigegeben"`
}

// freigegebeneQuellen übersetzt die Auswahl in Quellen.
//
// **Ohne zu prüfen, ob es sie gibt** — das tut `domain`, und zwar mit einer
// Meldung, die den Namen nennt. Wer hier stillschweigend verwürfe, was er
// nicht kennt, ließe einen Tippfehler wie eine bewusste Nichtfreigabe
// aussehen.
func freigegebeneQuellen(namen []string) []domain.Retrievalquelle {
	quellen := make([]domain.Retrievalquelle, 0, len(namen))
	for _, name := range namen {
		quellen = append(quellen, domain.Retrievalquelle(name))
	}

	return quellen
}

// RecherchetrefferDTO ist ein Fund samt auflösbarer Herkunft.
type RecherchetrefferDTO struct {
	Rang    int    `json:"rang"`
	Kennung string `json:"kennung"`

	Quelle string `json:"quelle"`

	// Kennungssystem, Fachkennung und Semantik sind die Identität nach A2 —
	// **kein Name als Identität**.
	Kennungssystem string `json:"kennungssystem"`
	Fachkennung    string `json:"fachkennung"`
	Semantik       string `json:"semantik"`

	// EigeneKarteID ist gesetzt, wo der Treffer eine Karte **aus diesem
	// Bestand** ist — dann führt der Weg dorthin und nicht nach außen.
	EigeneKarteID int64 `json:"eigeneKarteID"`

	Titel      string `json:"titel"`
	Ausschnitt string `json:"ausschnitt"`
	Text       string `json:"text"`

	Sprache     string `json:"sprache"`
	Regelnummer string `json:"regelnummer"`
	Urheber     string `json:"urheber"`

	Abschnitt int `json:"abschnitt"`
	Von       int `json:"von"`
	Bis       int `json:"bis"`

	Bezugsadresse string `json:"bezugsadresse"`

	Quellrevision string `json:"quellrevision"`
	Quelldatum    string `json:"quelldatum"`
	Abgerufen     string `json:"abgerufen"`

	// Weitere nennt die Abschnitte desselben Konzepts, die gebündelt wurden.
	Weitere []string `json:"weitere"`
}

// RechercheAntwort ist das Ergebnis einer Recherche.
type RechercheAntwort struct {
	Treffer []RecherchetrefferDTO `json:"treffer"`

	// Modus ist `hybrid`, `lexikalisch` oder `ohne-index`.
	Modus string `json:"modus"`

	// Leere sagt, **warum** nichts kam — leer, wo etwas kam.
	Leere string `json:"leere"`
	Grund string `json:"grund"`

	Modell            string `json:"modell"`
	Modellrevision    string `json:"modellrevision"`
	Suchkonfiguration string `json:"suchkonfiguration"`

	// Indexfingerabdruck bindet die Liste, Raumabdruck den Vektorraum —
	// **zwei Fragen, zwei Abdrücke**.
	Indexfingerabdruck string `json:"indexfingerabdruck"`
	Raumabdruck        string `json:"raumabdruck"`

	// Ausgangsstand ist der Stand der Karte, zu der Ähnliches gesucht wurde
	// — **die Antwort nennt ihn**, weil ihn vorher niemand kennt. Wer die
	// nächste Seite holt, gibt ihn zurück.
	Ausgangsstand string `json:"ausgangsstand"`

	Korpusstaende []KorpusstandDTO `json:"korpusstaende"`

	// Einschraenkungen nennt, was **nicht** getan wurde.
	Einschraenkungen []string `json:"einschraenkungen"`

	// Fortsetzung ist der Zeiger auf die nächste Seite.
	Fortsetzung string `json:"fortsetzung"`

	Fehler *Fehler `json:"fehler"`
}

// MitRecherche hängt die Recherche ans Fenster.
func (a *App) MitRecherche(fuehrer Recherchefuehrer) { a.recherche = fuehrer }

// Recherchieren beantwortet eine Frage gegen den Wissensbestand.
//
// **Ohne Generierung** — gerufen wird höchstens das **Einbettungsmodell**
// (lokal, für die Vektorhälfte), und geschrieben wird nichts.
func (a *App) Recherchieren(
	ctx context.Context, fall, frage string, grenze int,
	filter RecherchefilterDTO, fortsetzung string,
) RechercheAntwort {
	ctx = a.imProfil(ctx)

	return a.rechercheAus(ctx, domain.Rechercheanfrage{
		Fall: domain.Recherchefall(fall), Frage: frage, Grenze: grenze,
		Sprache: filter.Sprache, Korpusstand: filter.Korpusstand,
		Freigegeben: freigegebeneQuellen(filter.Freigegeben),
	}, fortsetzung)
}

// AehnlicheKarten sucht Karten, die einer anderen ähneln.
//
// **Mit dem Stand der Ausgangskarte** — eine Karte, die sich seit der Anfrage
// geändert hat, ist eine andere Frage.
func (a *App) AehnlicheKarten(
	ctx context.Context, karteID int64, stand string, grenze int,
	filter RecherchefilterDTO, fortsetzung string,
) RechercheAntwort {
	ctx = a.imProfil(ctx)

	return a.rechercheAus(ctx, domain.Rechercheanfrage{
		Fall: domain.FallAehnlich, AusgangskarteID: karteID,
		Ausgangsstand: stand, Grenze: grenze,
		Sprache: filter.Sprache, Korpusstand: filter.Korpusstand,
		Freigegeben: freigegebeneQuellen(filter.Freigegeben),
	}, fortsetzung)
}

// rechercheAus fährt eine Anfrage und übersetzt die Antwort.
//
// **Das Profil steht schon im Kontext**, und zwar in beiden Bindungen — jede
// setzt es als erste Zeile, so wie es der Wächter über die Profilnaht
// verlangt. Ein zweites [App.imProfil] hier gäbe eine zweite
// Vorgangskennung für denselben Ruf.
func (a *App) rechercheAus(
	ctx context.Context, anfrage domain.Rechercheanfrage, fortsetzung string,
) RechercheAntwort {
	antwort := RechercheAntwort{
		Treffer: []RecherchetrefferDTO{}, Korpusstaende: []KorpusstandDTO{},
		Einschraenkungen: []string{},
	}

	if a.recherche == nil {
		antwort.Fehler = Uebersetze(ctx, fehlenderBestand())

		return antwort
	}

	ergebnis, fehler := a.recherche.Recherchieren(ctx, anfrage, fortsetzung)
	if fehler != nil {
		antwort.Fehler = Uebersetze(ctx, fehler)

		return antwort
	}

	return rechercheAntwort(ergebnis)
}

// rechercheAntwort übersetzt das Ergebnis.
func rechercheAntwort(ergebnis domain.Rechercheergebnis) RechercheAntwort {
	treffer := make([]RecherchetrefferDTO, 0, len(ergebnis.Treffer))
	for _, einzelner := range ergebnis.Treffer {
		treffer = append(treffer, alsRecherchetrefferDTO(einzelner))
	}

	staende := make([]KorpusstandDTO, 0, len(ergebnis.Korpusstaende))
	for _, stand := range ergebnis.Korpusstaende {
		staende = append(staende, KorpusstandDTO{
			Quelle: string(stand.Quelle), Revision: stand.Revision,
			Datum: stand.Datum,
		})
	}

	einschraenkungen := ergebnis.Einschraenkungen
	if einschraenkungen == nil {
		einschraenkungen = []string{}
	}

	return RechercheAntwort{
		Treffer: treffer, Modus: ergebnis.Modus,
		Leere: string(ergebnis.Leere), Grund: ergebnis.Grund,
		Modell:             ergebnis.Modell,
		Modellrevision:     ergebnis.Modellrevision,
		Suchkonfiguration:  ergebnis.Suchkonfiguration,
		Indexfingerabdruck: ergebnis.Indexfingerabdruck,
		Raumabdruck:        ergebnis.Raumabdruck,
		Ausgangsstand:      ergebnis.Ausgangsstand,
		Korpusstaende:      staende,
		Einschraenkungen:   einschraenkungen,
		Fortsetzung:        ergebnis.Fortsetzung,
	}
}

// alsRecherchetrefferDTO übersetzt einen Treffer.
func alsRecherchetrefferDTO(treffer domain.Recherchetreffer) RecherchetrefferDTO {
	weitere := treffer.Weitere
	if weitere == nil {
		weitere = []string{}
	}

	return RecherchetrefferDTO{
		Rang: treffer.Rang, Kennung: treffer.Kennung,
		Quelle:         string(treffer.Quelle),
		Kennungssystem: string(treffer.Kennungssystem),
		Fachkennung:    treffer.Fachkennung,
		Semantik:       string(treffer.Semantik),
		EigeneKarteID:  treffer.EigeneKarteID,
		Titel:          treffer.Titel, Ausschnitt: treffer.Ausschnitt,
		Text: treffer.Text, Sprache: treffer.Sprache,
		Regelnummer: treffer.Regelnummer,
		Urheber:     string(treffer.Urheber),
		Abschnitt:   treffer.Abschnitt, Von: treffer.Von, Bis: treffer.Bis,
		Bezugsadresse: treffer.Bezugsadresse,
		Quellrevision: treffer.Quellrevision,
		Quelldatum:    treffer.Quelldatum,
		Abgerufen:     treffer.Abgerufen,
		Weitere:       weitere,
	}
}

// Recherchieren ist die Fensterbindung für die Suche.
func (b *Bindungen) Recherchieren(
	ctx context.Context,
	fall, frage string, grenze int, filter RecherchefilterDTO,
	fortsetzung string,
) RechercheAntwort {
	return b.app(ctx).Recherchieren(b.ruf(ctx), fall, frage, grenze, filter,
		fortsetzung)
}

// AehnlicheKarten ist die Fensterbindung für die Ähnlichkeitsfrage.
func (b *Bindungen) AehnlicheKarten(
	ctx context.Context,
	karteID int64, stand string, grenze int, filter RecherchefilterDTO,
	fortsetzung string,
) RechercheAntwort {
	return b.app(ctx).AehnlicheKarten(b.ruf(ctx), karteID, stand, grenze,
		filter, fortsetzung)
}

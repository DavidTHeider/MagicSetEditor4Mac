package app

// **Der Retrievalindex am Fenster** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Drei Rufe und kein vierter
//
// **Stand, Aufbau mit Abbruch, Suche.** Das Issue nennt genau diese: „Ein
// begrenzter tatsächlicher Diagnose-/Verwaltungsaufruf für Indexstatus,
// Aufbau/Abbruch und Suche" — und es sagt dazu, was das Nichtvorhandensein
// kostet: „Ein Paket mit Tests ohne Aufrufer schließt A15 nicht."
//
// # Der Modus steht in jeder Antwort
//
// **Weil man ihn an den Treffern nicht sieht.** Eine lexikalische Antwort
// ist eine andere Auskunft als eine hybride; wer das verschwiege, verkaufte
// den Rückfall als die Sache selbst — und genau das verbietet das Issue in
// seinem letzten Absatz.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Retrievalfuehrer ist die Naht zum Dienst.
type Retrievalfuehrer interface {
	Stand(ctx context.Context) service.Indexstand
	Suchen(ctx context.Context, wunsch service.Suchwunsch) (service.Suchergebnis, error)
	Aufbauen(ctx context.Context, auftrag service.Aufbauauftrag) (service.Aufbaubericht, error)
	Abbrechen() bool
}

// IndexmanifestDTO nennt, gegen welchen Stand gesucht wird.
type IndexmanifestDTO struct {
	Modell         string `json:"modell"`
	Modellrevision string `json:"modellrevision"`
	Laufzeit       string `json:"laufzeit"`
	Abstandsmass   string `json:"abstandsmass"`

	// Dimension ist die Länge der Vektoren.
	Dimension int `json:"dimension"`

	Zerteilfassung    string `json:"zerteilfassung"`
	Suchkonfiguration string `json:"suchkonfiguration"`

	// Korpusstaende nennt je Quelle ihren Stand — **einzeln**, weil eine
	// gemeinsame Zahl verschiedene Regelstände mischte.
	Korpusstaende []KorpusstandDTO `json:"korpusstaende"`
}

// KorpusstandDTO ist der Stand einer Quelle.
type KorpusstandDTO struct {
	Quelle   string `json:"quelle"`
	Revision string `json:"revision"`

	// Datum ist das **Gültigkeitsdatum** — leer, wo die Quelle keines nennt.
	Datum string `json:"datum"`
}

// IndexaufbaustandDTO ist der Fortschritt des letzten oder laufenden Aufbaus.
type IndexaufbaustandDTO struct {
	Laeuft      bool   `json:"laeuft"`
	Generation  int64  `json:"generation"`
	Geschrieben int    `json:"geschrieben"`
	Fehler      string `json:"fehler"`
}

// IndexstandAntwort ist die Auskunft über den Index.
type IndexstandAntwort struct {
	Aktiv    bool             `json:"aktiv"`
	Manifest IndexmanifestDTO `json:"manifest"`

	Aufbau IndexaufbaustandDTO `json:"aufbau"`

	// Suchbar sagt, ob **semantisch** gesucht werden kann; Grund sagt, warum
	// nicht.
	Suchbar bool   `json:"suchbar"`
	Grund   string `json:"grund"`

	Fehler *Fehler `json:"fehler"`
}

// IndexsuchfilterDTO ist die geprüfte Filterstruktur.
//
// **Und keine Suchsprache.** Das Issue: „Querytexte ersetzen keine geprüfte
// Filterstruktur." Wer nach roten Spontanzaubern filtert, sagt es in Feldern.
type IndexsuchfilterDTO struct {
	Quellen  []string `json:"quellen"`
	Sprachen []string `json:"sprachen"`
	Sets     []string `json:"sets"`

	// Farben ist die Farbe der Karte, Farbidentitaet ihre Identität nach
	// Regel 903.4 — **zwei Felder, weil es zwei Dinge sind**.
	Farben         []string `json:"farben"`
	Farbidentitaet []string `json:"farbidentitaet"`

	Typen []string `json:"typen"`

	// ManawertVon und ManawertBis sind einschließlich; `nil` heißt offen.
	ManawertVon *float64 `json:"manawertVon"`
	ManawertBis *float64 `json:"manawertBis"`

	// Format, Formatlage und Formatstand gehören zusammen: ohne Datenstand
	// gibt es gar keine Legalitätsauskunft.
	Format      string `json:"format"`
	Formatlage  string `json:"formatlage"`
	Formatstand string `json:"formatstand"`
}

// IndextrefferDTO ist ein Fund samt auflösbarer Herkunft.
type IndextrefferDTO struct {
	Rang    int    `json:"rang"`
	Kennung string `json:"kennung"`

	Quelle        string `json:"quelle"`
	Bezugsadresse string `json:"bezugsadresse"`

	Titel string `json:"titel"`
	Text  string `json:"text"`

	// Ausschnitt ist die Textstelle — leer, wo nur der Vektor passte.
	Ausschnitt string `json:"ausschnitt"`

	Sprache     string `json:"sprache"`
	Regelnummer string `json:"regelnummer"`
	Urheber     string `json:"urheber"`

	// Quellrevision, Quelldatum und Abgerufen sind die Herkunft — **drei
	// Angaben**, weil das Abrufdatum kein Gültigkeitsdatum ist.
	Quellrevision string `json:"quellrevision"`
	Quelldatum    string `json:"quelldatum"`
	Abgerufen     string `json:"abgerufen"`
}

// IndexsucheAntwort ist das Ergebnis einer Suche.
type IndexsucheAntwort struct {
	Treffer []IndextrefferDTO `json:"treffer"`

	// Modus ist `hybrid`, `lexikalisch` oder `ohne-index`.
	Modus string `json:"modus"`
	Grund string `json:"grund"`

	Fehler *Fehler `json:"fehler"`
}

// IndexaufbauAntwort ist, was aus einem Aufbau wurde.
type IndexaufbauAntwort struct {
	Generation  int64 `json:"generation"`
	Geschrieben int   `json:"geschrieben"`

	// Dauer ist die Bauzeit in Millisekunden.
	Dauer int64 `json:"dauer"`

	// Fortgeschrieben sagt, ob auf der aktiven Generation aufgebaut wurde.
	Fortgeschrieben bool `json:"fortgeschrieben"`

	Manifest IndexmanifestDTO `json:"manifest"`

	Fehler *Fehler `json:"fehler"`
}

// MitRetrieval hängt den Index ans Fenster.
func (a *App) MitRetrieval(fuehrer Retrievalfuehrer) { a.retrieval = fuehrer }

// Indexstand sagt, was dasteht — **und ändert nichts**.
func (a *App) Indexstand(ctx context.Context) IndexstandAntwort {
	ctx = a.imProfil(ctx)

	if a.retrieval == nil {
		return IndexstandAntwort{
			Manifest: manifestDTO(domain.Indexmanifest{}),
			Fehler:   Uebersetze(ctx, fehlenderBestand()),
		}
	}

	return standAntwort(a.retrieval.Stand(ctx))
}

// Indexsuche beantwortet eine Frage gegen den aktiven Index.
func (a *App) Indexsuche(
	ctx context.Context, frage string, grenze int, filter IndexsuchfilterDTO,
) IndexsucheAntwort {
	ctx = a.imProfil(ctx)

	antwort := IndexsucheAntwort{Treffer: []IndextrefferDTO{}}

	if a.retrieval == nil {
		antwort.Fehler = Uebersetze(ctx, fehlenderBestand())

		return antwort
	}

	ergebnis, fehler := a.retrieval.Suchen(ctx, service.Suchwunsch{
		Frage: frage, Grenze: grenze, Filter: alsRetrievalfilter(filter),
	})
	if fehler != nil {
		antwort.Fehler = Uebersetze(ctx, fehler)

		return antwort
	}

	return sucheAntwort(ergebnis)
}

// IndexAufbauen baut eine neue Generation und schaltet auf sie um.
//
// **Der Ruf kehrt erst am Ende zurück.** Ein Aufbau über Hunderttausende
// Abschnitte dauert; wer ihn anhalten will, ruft [App.IndexAbbrechen] —
// deshalb steht der Fortschritt im [App.Indexstand] und nicht in dieser
// Antwort.
//
// **`fortschreiben` baut auf der aktiven Generation auf**: die Quelle liefert
// dann nur, was sich geändert hat, und der Rest wird **kopiert** statt neu
// eingebettet. Wer dabei ein anderes Modell nimmt, wird abgewiesen — zwei
// Vektorräume in einer Generation sind Abstände ohne Bedeutung.
func (a *App) IndexAufbauen(
	ctx context.Context, modell, quelle, revision, datum string,
	fortschreiben bool,
) IndexaufbauAntwort {
	ctx = a.imProfil(ctx)

	// **Die Marke vor dem Start** (Nachlauf über Phase 6.12, N2).
	marke := a.laufmarke(domain.LaufIndexaufbau)

	if a.retrieval == nil {
		return leererAufbau(Uebersetze(ctx, fehlenderBestand()))
	}

	// **Ein eigener Abbruch, bevor der Aufbau beginnt** (Zweitblick über
	// Phase 6.12, Z2): `ki` aus erreicht ihn über [App.laufende], auch wenn
	// der Führer seinen noch nicht eingetragen hat.
	ctx, abbrechen := context.WithCancel(ctx)
	defer abbrechen()

	a.indexabbruch.setzen(abbrechen)
	defer a.indexabbruch.setzen(nil)

	a.nachDemStart(ctx, domain.LaufIndexaufbau, marke, func(context.Context) { abbrechen() })

	bericht, fehler := a.retrieval.Aufbauen(ctx, service.Aufbauauftrag{
		Modell: modell, Quelle: quelle, Revision: revision, Datum: datum,
		// **Jetzt, weil die Datei jetzt gelesen wird.** Wer sie früher geholt
		// hat, sagt es über die Revision; eine Abrufzeit, die vor dem Lesen
		// liegt, kennt dieses Fenster nicht.
		Abgerufen: time.Now(), Fortschreiben: fortschreiben,
	})
	if fehler != nil {
		return leererAufbau(Uebersetze(ctx, fehler))
	}

	return IndexaufbauAntwort{
		Generation: bericht.Generation, Geschrieben: bericht.Geschrieben,
		Dauer: bericht.Dauer.Milliseconds(), Fortgeschrieben: fortschreiben,
		Manifest: manifestDTO(bericht.Manifest),
	}
}

// IndexAbbrechen hält einen laufenden Aufbau an.
//
// **Es gibt `false` zurück, wenn keiner lief** — und keinen Fehler: „es läuft
// nichts" ist eine Auskunft, kein Fehlschlag.
func (a *App) IndexAbbrechen() bool {
	if a.retrieval == nil {
		return false
	}

	return a.retrieval.Abbrechen()
}

// leererAufbau ist die Antwort, wenn nichts gebaut wurde.
//
// **Mit einem Manifest aus leeren Listen und nicht aus Nullwerten**: ein
// `null` statt `[]` kommt im Fenster als „gibt es nicht" an, und die Seite
// bliebe an einer Stelle leer, an der ein Fehler stehen sollte (§3, die Naht
// zwischen den Prüfebenen).
func leererAufbau(fehler *Fehler) IndexaufbauAntwort {
	return IndexaufbauAntwort{
		Manifest: manifestDTO(domain.Indexmanifest{}), Fehler: fehler,
	}
}

// alsRetrievalfilter übersetzt die Filterstruktur.
func alsRetrievalfilter(filter IndexsuchfilterDTO) service.Retrievalfilter {
	quellen := make([]domain.Retrievalquelle, 0, len(filter.Quellen))
	for _, quelle := range filter.Quellen {
		quellen = append(quellen, domain.Retrievalquelle(quelle))
	}

	return service.Retrievalfilter{
		Quellen: quellen, Sprachen: filter.Sprachen, Sets: filter.Sets,
		Farben: filter.Farben, Farbidentitaet: filter.Farbidentitaet,
		Typen:       filter.Typen,
		ManawertVon: filter.ManawertVon, ManawertBis: filter.ManawertBis,
		Format:      filter.Format,
		Formatlage:  filter.Formatlage,
		Formatstand: filter.Formatstand,
	}
}

// standAntwort baut die Auskunft.
func standAntwort(stand service.Indexstand) IndexstandAntwort {
	return IndexstandAntwort{
		Aktiv:    stand.Aktiv,
		Manifest: manifestDTO(stand.Manifest),
		Aufbau: IndexaufbaustandDTO{
			Laeuft:      stand.Aufbau.Laeuft,
			Generation:  stand.Aufbau.Generation,
			Geschrieben: stand.Aufbau.Geschrieben,
			Fehler:      stand.Aufbau.Fehler,
		},
		Suchbar: stand.Suchbar, Grund: stand.Grund,
	}
}

// manifestDTO übersetzt das Manifest.
func manifestDTO(manifest domain.Indexmanifest) IndexmanifestDTO {
	staende := make([]KorpusstandDTO, 0, len(manifest.Korpusstaende))
	for _, stand := range manifest.Korpusstaende {
		staende = append(staende, KorpusstandDTO{
			Quelle: string(stand.Quelle), Revision: stand.Revision,
			Datum: stand.Datum,
		})
	}

	return IndexmanifestDTO{
		Modell: manifest.Modell, Modellrevision: manifest.Modellrevision,
		Laufzeit:          manifest.Laufzeit,
		Abstandsmass:      string(manifest.Abstandsmass),
		Dimension:         manifest.Vertrag.Dimension,
		Zerteilfassung:    manifest.Zerteilfassung,
		Suchkonfiguration: manifest.Suchkonfiguration,
		Korpusstaende:     staende,
	}
}

// sucheAntwort baut das Ergebnis.
func sucheAntwort(ergebnis service.Suchergebnis) IndexsucheAntwort {
	treffer := make([]IndextrefferDTO, 0, len(ergebnis.Treffer))

	for _, einzelner := range ergebnis.Treffer {
		dokument := einzelner.Dokument

		treffer = append(treffer, IndextrefferDTO{
			Rang: einzelner.Rang, Kennung: dokument.Kennung,
			Quelle:        string(dokument.Quelle),
			Bezugsadresse: dokument.Bezugsadresse,
			Titel:         dokument.Titel, Text: dokument.Text,
			Ausschnitt: einzelner.Ausschnitt, Sprache: dokument.Sprache,
			Regelnummer:   dokument.Regelnummer,
			Urheber:       string(dokument.Urheber),
			Quellrevision: dokument.Quellrevision,
			Quelldatum:    dokument.Quelldatum,
			Abgerufen:     dokument.Abgerufen.UTC().Format(time.RFC3339),
		})
	}

	return IndexsucheAntwort{
		Treffer: treffer, Modus: string(ergebnis.Modus), Grund: ergebnis.Grund,
	}
}

// Indexstand ist die Fensterbindung für die Auskunft.
func (b *Bindungen) Indexstand(ctx context.Context) IndexstandAntwort {
	return b.app(ctx).Indexstand(b.ruf(ctx))
}

// Indexsuche ist die Fensterbindung für die Suche.
func (b *Bindungen) Indexsuche(
	ctx context.Context,
	frage string, grenze int, filter IndexsuchfilterDTO,
) IndexsucheAntwort {
	return b.app(ctx).Indexsuche(b.ruf(ctx), frage, grenze, filter)
}

// IndexAufbauen ist die Fensterbindung für den Aufbau.
func (b *Bindungen) IndexAufbauen(
	ctx context.Context,
	modell, quelle, revision, datum string, fortschreiben bool,
) IndexaufbauAntwort {
	return b.app(ctx).IndexAufbauen(b.ruf(ctx), modell, quelle, revision,
		datum, fortschreiben)
}

// IndexAbbrechen ist die Fensterbindung für den Abbruch.
func (b *Bindungen) IndexAbbrechen(ctx context.Context) bool { return b.app(ctx).IndexAbbrechen() }

package app

// **Der Dienstkatalog** — Werkstück A1 von Phase 5.3
// (#1440),
// entschieden in [0082](../../docs/entscheidungen/0082-der-dienstkatalog.md).
//
// # Was er ist
//
// Eine **Auskunft** über jede Fähigkeit eines fremden Dienstes: der Vertrag
// aus `domain.Dienstvertraege` und dazu, ob sie **jetzt** bereit ist — und
// wenn nicht, was zuerst dagegen steht. Die Tatsachen sammelt diese Datei:
// den Schalter aus den Einstellungen, den Zugang aus dem Schlüsselbund, das
// Nein des Riegels, die Wahl je Frage. **Welche gewinnt, sagt `domain`**
// ([domain.Dienstlage]), nicht ein `if` hier (§2).
//
// # Was er nicht tut
//
// **Er schickt nichts.** Kein Klient wird gebaut, keine Adresse gerufen —
// der Katalog sagt, ob eine Anfrage ginge, und versucht es nicht. Ein Dienst,
// dessen Bereitschaft man nur durch eine Anfrage feststellen könnte, hätte
// keinen Vertrag.
//
// # Ein Hindernis bleibt bei seinem Dienst
//
// Ein Schlüsselbund, der für Scrydex nicht antwortet, macht Scrydex
// »gestört« — und sonst nichts. Was den **ganzen** Katalog scheitern lässt,
// ist nur, was für alle gilt: die Einstellungen sind nicht lesbar. Dieselbe
// Trennung wie in [Suchquellen], und aus demselben Grund (#1440: kein
// einzelner Anbieterfehler blockiert die anderen).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// DienstvertragDTO ist der Vertrag, wie das Fenster ihn liest.
type DienstvertragDTO struct {
	Basisadresse string   `json:"basisadresse"`
	Endpunkte    []string `json:"endpunkte"`
	Zugang       string   `json:"zugang"`
	Tarif        string   `json:"tarif"`
	Grenze       string   `json:"grenze"`
	Frist        string   `json:"frist"`
	Bedingungen  string   `json:"bedingungen"`
	Abgerufen    string   `json:"abgerufen"`
	Beleg        string   `json:"beleg"`
}

// NachweisDTO ist der letzte Lauf gegen den echten Dienst.
type NachweisDTO struct {
	Lauf      string   `json:"lauf"`
	Datum     string   `json:"datum"`
	Fassung   string   `json:"fassung"`
	Endpunkte []string `json:"endpunkte"`

	// GiltFuerDieseFassung sagt, ob der Lauf mit **dieser** Fassung lief.
	//
	// **Gerechnet, nicht behauptet**: ein Nachweis aus 0.3.0 ist keine
	// Prüfung von 0.4.0, und die Oberfläche soll das sagen, ohne die
	// Fassungen selbst vergleichen zu müssen (#1440).
	GiltFuerDieseFassung bool `json:"giltFuerDieseFassung"`
}

// DienstfaehigkeitDTO ist eine Fähigkeit eines Dienstes samt Status.
type DienstfaehigkeitDTO struct {
	Dienst string `json:"dienst"`
	Frage  string `json:"frage"`

	// Status ist einer der Zustände aus `domain.Dienstlage`.
	Status string `json:"status"`
	// Grund sagt, was dem Status zugrunde liegt — leer bei »bereit«.
	Grund string `json:"grund"`
	// Angabe ist bei »zugangFehlt« die Zugangsangabe, die fehlt — sonst leer.
	//
	// **Neben dem Grund, nicht statt seiner**: der Grund ist ein Satz für
	// das Protokoll, die Angabe ein Schlüssel, den das Fenster in seiner
	// Sprache benennt (§9).
	Angabe string `json:"angabe"`
	// HatWahl sagt, ob die Frage mehr als ein Dienst beantworten kann.
	HatWahl bool `json:"hatWahl"`
	// Gewaehlt sagt, ob dieser Dienst die Frage beantwortet.
	Gewaehlt bool `json:"gewaehlt"`
	// Wirkung ist die Klasse aus `domain.Wirkung` — die Externwirkung neben
	// der Lesend-/Schreibend-Einordnung der Bindungen (#1440).
	Wirkung string `json:"wirkung"`

	Vertrag  DienstvertragDTO `json:"vertrag"`
	Nachweis *NachweisDTO     `json:"nachweis"`
}

// DienstkatalogAntwort ist der Umschlag für das Fenster.
type DienstkatalogAntwort struct {
	Eintraege []DienstfaehigkeitDTO `json:"eintraege"`
	Fehler    *Fehler               `json:"fehler"`
}

// Dienstkatalog nennt jede Fähigkeit eines fremden Dienstes samt Status.
func (a *App) Dienstkatalog(ctx context.Context) DienstkatalogAntwort {
	ctx = a.imProfil(ctx)

	eintraege, err := Dienstkatalogeintraege(ctx, a.wahl(), a.bundOderAnmeldebund())
	if err != nil {
		return DienstkatalogAntwort{Eintraege: []DienstfaehigkeitDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return DienstkatalogAntwort{Eintraege: eintraege}
}

// Dienstkatalogweg ist der Katalog für die Flächen ohne Fenster.
type Dienstkatalogweg struct {
	wahl Dienstwahl
	bund Geheimnisleser
}

// DienstkatalognahtUeber baut den Weg über einem Bestand — mit derselben
// Wahl samt Riegel wie die Suche.
func DienstkatalognahtUeber(db *sql.DB) *Dienstkatalogweg {
	return &Dienstkatalogweg{wahl: dienstwahlUeber(db), bund: schluesselbund.Anmeldebund()}
}

// MitBund setzt den Schlüsselbund — für Tests, die den echten nicht anfassen dürfen.
func (w *Dienstkatalogweg) MitBund(bund Geheimnisleser) *Dienstkatalogweg {
	w.bund = bund

	return w
}

// Katalog gibt die Einträge.
func (w *Dienstkatalogweg) Katalog(ctx context.Context) ([]DienstfaehigkeitDTO, error) {
	return Dienstkatalogeintraege(ctx, w.wahl, w.bund)
}

// Dienstkatalogeintraege sammelt die Tatsachen und lässt `domain` urteilen.
//
// **Die Wahl je Frage kommt aus derselben Quelle wie der Schalter**, wenn
// sie eine [Fragewahl] ist — sonst gilt die Vorgabe, der erste Dienst mit
// Vertrag. Ein Schalter, der keine Fragen kennt, ist einer aus einem Test.
func Dienstkatalogeintraege(
	ctx context.Context, wahl Dienstwahl, bund Geheimnisleser,
) ([]DienstfaehigkeitDTO, error) {
	an, err := angeschaltet(ctx, wahl)
	if err != nil {
		return nil, err
	}

	gewaehlt, err := gewaehlteDienste(ctx, wahl)
	if err != nil {
		return nil, err
	}

	var sperre error
	if traeger, hatRiegel := wahl.(Riegeltraeger); hatRiegel {
		sperre = traeger.NachAussen(ctx)
	}

	zugaenge := map[domain.Dienst]zugangslage{}

	vertraege := domain.Dienstvertraege()
	eintraege := make([]DienstfaehigkeitDTO, 0, len(vertraege))

	for _, vertrag := range vertraege {
		lage := domain.Dienstlage{An: an[vertrag.Dienst]}

		if lage.An {
			zugang, gesehen := zugaenge[vertrag.Dienst]
			if !gesehen {
				zugang = zugangslageVon(ctx, bund, vertrag.Dienst)
				zugaenge[vertrag.Dienst] = zugang
			}

			lage.FehlendeAngabe, lage.Stoerung, lage.Sperre = zugang.fehlt, zugang.stoerung, sperre
		}

		lage.HatWahl = len(domain.DiensteFuer(vertrag.Frage)) > 1
		lage.Gewaehlt = gewaehlt[vertrag.Frage] == vertrag.Dienst

		eintraege = append(eintraege, alsFaehigkeit(vertrag, lage))
	}

	return eintraege, nil
}

// zugangslage ist, was der Schlüsselbund über einen Dienst sagt.
type zugangslage struct {
	fehlt    string
	stoerung error
}

// zugangslageVon liest jede Zugangsangabe eines Dienstes — bis zur ersten,
// die fehlt oder nicht lesbar ist.
func zugangslageVon(ctx context.Context, bund Geheimnisleser, dienst domain.Dienst) zugangslage {
	for _, angabe := range dienst.Zugangsangaben() {
		wert, err := geheimnis(ctx, bund, domain.ZugangKonto(dienst, angabe))
		if err != nil {
			return zugangslage{stoerung: err}
		}

		if wert == "" {
			return zugangslage{fehlt: angabe}
		}
	}

	return zugangslage{}
}

// gewaehlteDienste sagt je Frage mit Wahl, welcher Dienst sie beantwortet.
func gewaehlteDienste(ctx context.Context, wahl Dienstwahl) (map[domain.Frage]domain.Dienst, error) {
	fragewahl, kenntFragen := wahl.(Fragewahl)
	gewaehlt := map[domain.Frage]domain.Dienst{}

	for _, frage := range domain.Fragen() {
		dienst := domain.DiensteFuer(frage)[0]

		if kenntFragen {
			gefragt, err := fragewahl.FuerFrage(ctx, frage)
			if err != nil {
				return nil, err
			}

			dienst = gefragt
		}

		gewaehlt[frage] = dienst
	}

	return gewaehlt, nil
}

// alsFaehigkeit bildet Vertrag und Lage auf den DTO ab.
func alsFaehigkeit(vertrag domain.Dienstvertrag, lage domain.Dienstlage) DienstfaehigkeitDTO {
	eintrag := DienstfaehigkeitDTO{
		Dienst: string(vertrag.Dienst), Frage: string(vertrag.Frage),
		Status: lage.Status(), Grund: lage.Grund(), Angabe: lage.FehlendeAngabe,
		HatWahl: lage.HatWahl, Gewaehlt: !lage.HatWahl || lage.Gewaehlt,
		Wirkung: string(vertrag.Wirkung),
		Vertrag: DienstvertragDTO{
			Basisadresse: vertrag.Basisadresse, Endpunkte: vertrag.Endpunkte,
			Zugang: vertrag.Zugang, Tarif: vertrag.Tarif, Grenze: vertrag.Grenze,
			Frist: vertrag.Frist, Bedingungen: vertrag.Bedingungen,
			Abgerufen: vertrag.Abgerufen, Beleg: vertrag.Beleg,
		},
	}

	if vertrag.Nachweis != nil {
		eintrag.Nachweis = &NachweisDTO{
			Lauf: vertrag.Nachweis.Lauf, Datum: vertrag.Nachweis.Datum,
			Fassung: vertrag.Nachweis.Fassung, Endpunkte: vertrag.Nachweis.Endpunkte,
			GiltFuerDieseFassung: vertrag.Nachweis.Fassung == Fassung,
		}
	}

	return eintrag
}

// Dienstkatalog nennt jede Fähigkeit eines fremden Dienstes samt Status —
// die Fensterbindung, mit dem Kontext der Anwendung.
func (b *Bindungen) Dienstkatalog(ctx context.Context) DienstkatalogAntwort {
	return b.app(ctx).Dienstkatalog(b.ruf(ctx))
}

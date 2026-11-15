package app

// **Tauschkandidaten und ihr Vorschlag am Fenster** — Werkstück `T4` von
// Phase 5.6
// (#1469).
//
// # Warum der fremde Stand hereingereicht wird
//
// **Weil er eine Freigabe ist und keine Abfrage.** Was ein Gegenüber zeigt,
// entscheidet seine Freigabe (T3, [#1468]); wie es hierherkommt, entscheidet
// der Weg — heute von Hand, mit `T5` über eine Leitung. Ein Abgleich, der
// selbst nachschlüge, ginge an beidem vorbei.
//
// **Und er hat dieselbe Gestalt, die T3 hergibt.** Was `Freigabesicht`
// ausgibt, ist genau das, was hier hereingeht: eine Zeile mit sechs Feldern.
// Zwei Gestalten für dieselbe Sache liefen auseinander, und die erste, die es
// merkte, wäre keine.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// FremdstandDTO ist, was von einem Gegenüber angekommen ist.
type FremdstandDTO struct {
	// Angebot ist, was es hergeben kann — seine Dubletten.
	Angebot []SichtkarteDTO `json:"angebot"`

	// Wuensche ist, was es sucht.
	Wuensche []SichtkarteDTO `json:"wuensche"`
}

// TauschkandidatDTO ist ein Treffer samt seiner Erklärung.
type TauschkandidatDTO struct {
	// Wunsch ist die Zeile, die jemand wollte.
	Wunsch string `json:"wunsch"`

	// Angebot ist die Zeile, die sie bedienen kann.
	Angebot string `json:"angebot"`

	// AngebotDruck ist ihr Druck — aus ihm entsteht eine Handelszeile, wenn
	// jemand den Vorschlag übernimmt (P11, #2379). **Erweiternd**.
	AngebotDruck string `json:"angebotDruck"`

	// AngebotZustand, AngebotSprache und AngebotFinish sind seine Variante —
	// ohne sie entstand aus „Vorschlag übernehmen" eine Zeile, die drüben
	// nicht verbucht werden konnte (Befund 2 des Zweitblicks über Phase 8).
	// **Erweiternd**.
	AngebotZustand string `json:"angebotZustand"`
	AngebotSprache string `json:"angebotSprache"`
	AngebotFinish  string `json:"angebotFinish"`

	// Menge ist, wie viele Stücke dieser Treffer deckt.
	Menge int `json:"menge"`

	// Fehlend ist, wie viele dem Wunsch danach noch fehlen.
	Fehlend int `json:"fehlend"`

	// Rang ist die Dringlichkeit des Wunsches.
	Rang string `json:"rang"`

	// Art ist `genau` oder `jederdruck` — **die Erklärung, warum es ein
	// Treffer ist**.
	Art string `json:"art"`
}

// TauschseiteDTO ist, was eine Seite gibt.
type TauschseiteDTO struct {
	Zeilen []TauschkandidatDTO `json:"zeilen"`

	Stueck  int `json:"stueck"`
	Gedeckt int `json:"gedeckt"`
	Offen   int `json:"offen"`

	// Wert und Waehrung gelten nur mit Bepreist.
	Wert     int64  `json:"wert"`
	Waehrung string `json:"waehrung"`

	// Bepreist sagt, ob **jede** Zeile einen Preis in derselben Währung trug.
	Bepreist bool `json:"bepreist"`
}

// TauschabgleichAntwort ist das Ergebnis samt seinem Fehler.
type TauschabgleichAntwort struct {
	Kennung string `json:"kennung"`

	// Gebe sind meine Dubletten für seine Wünsche.
	Gebe []TauschkandidatDTO `json:"gebe"`

	// Bekomme sind seine Zeilen für meine Wünsche.
	Bekomme []TauschkandidatDTO `json:"bekomme"`

	// Vorschlag gilt nur mit HatVorschlag.
	Vorschlag TauschvorschlagDTO `json:"vorschlag"`

	// HatVorschlag sagt, ob es einen gibt — **ohne beide Seiten gibt es
	// keinen**.
	HatVorschlag bool `json:"hatVorschlag"`

	// Befunde sind die Stellen, an denen eine Kennung gefehlt hat.
	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler"`
}

// TauschvorschlagDTO ist der Vorschlag für beide Seiten.
type TauschvorschlagDTO struct {
	Gebe    TauschseiteDTO `json:"gebe"`
	Bekomme TauschseiteDTO `json:"bekomme"`
}

// Tauschabgleich sagt, wer was von wem bekommen könnte.
func (a *App) Tauschabgleich(
	ctx context.Context, kennung string, stand FremdstandDTO,
	behalten, hoechstzahl int,
) TauschabgleichAntwort {
	ctx = a.imProfil(ctx)

	abgleich, fehler := a.tauschwerk().Abgleichen(ctx, service.Fremdstand{
		Kennung:  kennung,
		Angebot:  ausSichtkarten(stand.Angebot),
		Wuensche: ausSichtkarten(stand.Wuensche),
	}, behalten, domain.Tauschumfang{Hoechstzahl: hoechstzahl})
	if fehler != nil {
		return TauschabgleichAntwort{
			Kennung: kennung,
			Gebe:    []TauschkandidatDTO{}, Bekomme: []TauschkandidatDTO{},
			Befunde: []string{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return TauschabgleichAntwort{
		Kennung: abgleich.Kennung,
		Gebe:    alsKandidaten(abgleich.Gebe),
		Bekomme: alsKandidaten(abgleich.Bekomme),
		Vorschlag: TauschvorschlagDTO{
			Gebe:    alsTauschseite(abgleich.Vorschlag.Gebe),
			Bekomme: alsTauschseite(abgleich.Vorschlag.Bekomme),
		},
		HatVorschlag: abgleich.HatVorschlag,
		Befunde:      alsBefundtexte(abgleich.Befunde),
	}
}

// ausSichtkarten macht aus den DTOs die Zeilen der Domäne.
func ausSichtkarten(dtos []SichtkarteDTO) []domain.Sichtkarte {
	karten := make([]domain.Sichtkarte, 0, len(dtos))

	for _, dto := range dtos {
		karten = append(karten, domain.Sichtkarte{
			Referenz: dto.Referenz, Printing: dto.Printing, Menge: dto.Menge,
			Zustand: domain.Zustand(dto.Zustand), Sprache: dto.Sprache,
			Finish: domain.Finish(dto.Finish), Bezeichnung: dto.Bezeichnung,
			Wunsch: dto.Wunsch,
		})
	}

	return karten
}

// alsKandidaten macht aus den Treffern ihre DTOs.
func alsKandidaten(treffer []domain.Tauschkandidat) []TauschkandidatDTO {
	dtos := make([]TauschkandidatDTO, 0, len(treffer))

	for _, kandidat := range treffer {
		dtos = append(dtos, alsKandidat(kandidat))
	}

	return dtos
}

// alsKandidat macht aus einem Treffer sein DTO.
func alsKandidat(kandidat domain.Tauschkandidat) TauschkandidatDTO {
	return TauschkandidatDTO{
		Wunsch:         kandidat.Wunsch.Bezeichnung,
		Angebot:        kandidat.Angebot.Bezeichnung,
		AngebotDruck:   kandidat.Angebot.Druck,
		AngebotZustand: string(kandidat.Angebot.Zustand),
		AngebotSprache: kandidat.Angebot.Sprache,
		AngebotFinish:  string(kandidat.Angebot.Finish),
		Menge:          kandidat.Menge,
		Fehlend:        kandidat.Fehlend,
		Rang:           string(kandidat.Wunsch.Rang),
		Art:            string(kandidat.Wunsch.Art),
	}
}

// alsTauschseite macht aus einer Seite ihr DTO.
func alsTauschseite(seite domain.Tauschseite) TauschseiteDTO {
	return TauschseiteDTO{
		Zeilen:   alsKandidaten(seite.Zeilen),
		Stueck:   seite.Stueck,
		Gedeckt:  seite.Gedeckt,
		Offen:    seite.Offen,
		Wert:     seite.Wert,
		Waehrung: seite.Waehrung,
		Bepreist: seite.Bepreist,
	}
}

// alsBefundtexte macht aus den Befunden ihre Sätze.
//
// **Als Text und nicht als Verbund**: ein Befund dieses Abgleichs hat keine
// Datei und keine Zeilennummer, und `Befund.String()` setzt Ort und Meldung
// schon richtig zusammen.
func alsBefundtexte(befunde []domain.Befund) []string {
	texte := make([]string, 0, len(befunde))
	for _, befund := range befunde {
		texte = append(texte, befund.String())
	}

	return texte
}

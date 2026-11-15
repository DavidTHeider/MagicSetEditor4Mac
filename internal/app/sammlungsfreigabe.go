package app

// **Freigegebene Sammlungssichten am Fenster** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// # Vier Bindungen und nicht eine
//
// **Weil sie vier verschiedene Entscheidungen sind.** Lesen sagt, was gilt;
// Erteilen gibt etwas frei; Widerrufen nimmt es zurück; und die **Vorschau**
// zeigt, was ein Gegenüber davon wirklich sähe. Ein Ruf, der eine Freigabe
// „setzt" und dabei alle Umfänge entgegennähme, machte aus einem vergessenen
// Feld eine stille Freigabe — derselbe Grund wie bei den Kontakten.
//
// # Warum es die Vorschau gibt
//
// **Weil niemand einer Zusage glauben muss.** „Deine Preise gehen nicht
// hinaus" ist ein Satz; die Vorschau ist der Blick darauf. Sie geht durch
// **denselben** Weg wie die Antwort an ein Gegenüber — eine zweite Rechnung
// daneben zeigte, was sie selbst ausrechnet, und nicht, was hinausgeht.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// SammlungsfreigabeDTO ist eine erteilte Sicht.
type SammlungsfreigabeDTO struct {
	// Empfaenger ist die Peerkennung des Gegenübers.
	Empfaenger string `json:"empfaenger"`

	// Umfang ist `auswahl`, `dubletten`, `wuensche`, `sammlung` oder
	// `digitales`.
	Umfang string `json:"umfang"`

	// Zweck ist `teilen`, `handel` oder `partie` — **nie `paarung`**.
	Zweck string `json:"zweck"`

	// Auswahl sind die gewählten Druckkennungen bei `auswahl`.
	//
	// **Ohne `omitempty`**: eine leere Liste muss als `[]` ankommen und nicht
	// als `null` — die Oberfläche zählt sie (§3).
	Auswahl []string `json:"auswahl"`

	// Behalten ist die Behaltemenge bei `dubletten`.
	Behalten int `json:"behalten"`

	// Ablauf ist RFC 3339 — **leer heißt unbefristet**.
	Ablauf string `json:"ablauf"`

	// Erteilt ist der Zeitpunkt der Erteilung, RFC 3339.
	Erteilt string `json:"erteilt"`
}

// SammlungsfreigabenAntwort ist die Liste samt ihrem Fehler.
type SammlungsfreigabenAntwort struct {
	Freigaben []SammlungsfreigabeDTO `json:"freigaben"`

	// Umfaenge ist die Aufzählung, aus der eine Fläche wählt.
	//
	// **Sie kommt mit und steht nicht in der Oberfläche**: zwei Aufzählungen
	// laufen auseinander, und die erste, die es merkt, ist keine.
	Umfaenge []string `json:"umfaenge"`

	// Nachladefehler sagt: erteilt oder widerrufen **ist**, nur die Liste
	// danach war nicht zu lesen (§7, Regel 3; #1970) — dann ist
	// `freigaben` leer und `fehler` auch.
	Nachladefehler *Fehler `json:"nachladefehler,omitempty"`

	Fehler *Fehler `json:"fehler"`
}

// SichtkarteDTO ist eine Zeile, wie ein Gegenüber sie sähe.
//
// **Dieselben sechs Felder wie [domain.Sichtkarte]** — und keines mehr. Wer
// hier eines ergänzte, ergänzte es an der Vorschau und nicht am Draht: die
// Vorschau löge dann über das, was wirklich hinausgeht.
type SichtkarteDTO struct {
	Referenz    string `json:"referenz"`
	Printing    string `json:"printing"`
	Menge       int    `json:"menge"`
	Zustand     string `json:"zustand"`
	Sprache     string `json:"sprache"`
	Finish      string `json:"finish"`
	Bezeichnung string `json:"bezeichnung"`

	// Wunsch heißt: gesucht und nicht angeboten (P11, #2379) — nur bei einem
	// Wunsch im Draht und in der Datei, wie an [domain.Sichtkarte].
	Wunsch bool `json:"wunsch,omitempty"`
}

// FreigabesichtAntwort ist die Vorschau samt ihrem Fehler.
type FreigabesichtAntwort struct {
	Zeilen []SichtkarteDTO `json:"zeilen"`
	Fehler *Fehler         `json:"fehler"`
}

// Sammlungsfreigaben gibt, was einem Gegenüber erteilt ist.
func (a *App) Sammlungsfreigaben(
	ctx context.Context, kennung string,
) SammlungsfreigabenAntwort {
	ctx = a.imProfil(ctx)

	erteilt, fehler := a.freigabepflege().Freigaben(ctx, kennung)
	if fehler != nil {
		return SammlungsfreigabenAntwort{
			Freigaben: []SammlungsfreigabeDTO{}, Umfaenge: alleUmfaenge(),
			Fehler: Uebersetze(ctx, fehler),
		}
	}

	return SammlungsfreigabenAntwort{
		Freigaben: alsFreigabeDTOs(erteilt), Umfaenge: alleUmfaenge(),
	}
}

// SammlungFreigeben erteilt einem Gegenüber eine Sicht.
func (a *App) SammlungFreigeben(
	ctx context.Context, dto SammlungsfreigabeDTO,
) SammlungsfreigabenAntwort {
	ctx = a.imProfil(ctx)

	freigabe, fehler := ausFreigabeDTO(dto)
	if fehler != nil {
		return a.freigabenMitFehler(ctx, dto.Empfaenger, fehler)
	}

	if fehler := a.freigabepflege().Erteilen(ctx, freigabe); fehler != nil {
		return a.freigabenMitFehler(ctx, dto.Empfaenger, fehler)
	}

	return a.sichtfreigabenNachDemSchreiben(ctx, dto.Empfaenger)
}

// SammlungsfreigabeWiderrufen nimmt eine Sicht zurück.
//
// **Künftige Antworten hören damit auf** — mehr kann dieses Programm nicht
// zusagen. Was ein Gegenüber schon kopiert hat, ist bei ihm.
func (a *App) SammlungsfreigabeWiderrufen(
	ctx context.Context, kennung, zweck, umfang string,
) SammlungsfreigabenAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.freigabepflege().Widerrufen(ctx, kennung,
		domain.Zweck(zweck), domain.Freigabeumfang(umfang)); fehler != nil {
		return a.freigabenMitFehler(ctx, kennung, fehler)
	}

	return a.sichtfreigabenNachDemSchreiben(ctx, kennung)
}

// sichtfreigabenNachDemSchreiben gibt die Liste nach einem **geglückten**
// Schritt.
//
// **Geschrieben ist geschrieben** (§7, Regel 3;
// #1970):
// scheitert das Lesen danach, steht sein Fehler in
// [SammlungsfreigabenAntwort.Nachladefehler] und nicht in `fehler` — wer
// `fehler` läse, widerriefe danach ins Leere.
func (a *App) sichtfreigabenNachDemSchreiben(
	ctx context.Context, kennung string,
) SammlungsfreigabenAntwort {
	antwort := a.Sammlungsfreigaben(ctx, kennung)
	antwort.Nachladefehler, antwort.Fehler = antwort.Fehler, nil

	return antwort
}

// Freigabesicht zeigt, was ein Gegenüber wirklich sähe.
func (a *App) Freigabesicht(
	ctx context.Context, kennung, zweck string,
) FreigabesichtAntwort {
	ctx = a.imProfil(ctx)

	// **Eine Vorschau ist eine eigene Sitzung.** Sie darf keine Namen zeigen,
	// die in einem laufenden Gespräch etwas bedeuten — sonst wäre die
	// Vorschau selbst der Weg, über den sich zwei Sichten aneinanderlegen
	// lassen.
	sicht, fehler := a.freigabepflege().Sicht(ctx, kennung,
		"vorschau-"+a.uhr().UTC().Format(time.RFC3339Nano), domain.Zweck(zweck))
	if fehler != nil {
		return FreigabesichtAntwort{
			Zeilen: []SichtkarteDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return FreigabesichtAntwort{Zeilen: alsSichtkarten(sicht)}
}

// freigabenMitFehler gibt die Liste samt dem Fehler, der dazu führte.
func (a *App) freigabenMitFehler(
	ctx context.Context, kennung string, fehler error,
) SammlungsfreigabenAntwort {
	antwort := a.Sammlungsfreigaben(ctx, kennung)
	antwort.Fehler = Uebersetze(ctx, fehler)

	return antwort
}

// alsFreigabeDTOs macht aus den Entitäten ihre DTOs.
func alsFreigabeDTOs(
	erteilt []domain.Sammlungsfreigabe,
) []SammlungsfreigabeDTO {
	dtos := make([]SammlungsfreigabeDTO, 0, len(erteilt))

	for _, freigabe := range erteilt {
		auswahl := freigabe.Auswahl
		if auswahl == nil {
			auswahl = []string{}
		}

		dtos = append(dtos, SammlungsfreigabeDTO{
			Empfaenger: freigabe.Empfaenger,
			Umfang:     string(freigabe.Umfang),
			Zweck:      string(freigabe.Zweck),
			Auswahl:    auswahl,
			Behalten:   freigabe.Behalten,
			Ablauf:     alsZeitpunkt(freigabe.Ablauf),
			Erteilt:    alsZeitpunkt(freigabe.Erteilt),
		})
	}

	return dtos
}

// ausFreigabeDTO macht aus dem DTO die Entität.
//
// **Ein unlesbarer Ablauf ist ein Feldbefund**, kein stiller Nullwert: „leer"
// heißt unbefristet, und wer sich vertippt hat, bekäme sonst genau das.
func ausFreigabeDTO(dto SammlungsfreigabeDTO) (domain.Sammlungsfreigabe, error) {
	ablauf, fehler := zeitpunktAus("ablauf", dto.Ablauf)
	if fehler != nil {
		return domain.Sammlungsfreigabe{}, fehler
	}

	return domain.Sammlungsfreigabe{
		Empfaenger: dto.Empfaenger,
		Umfang:     domain.Freigabeumfang(dto.Umfang),
		Zweck:      domain.Zweck(dto.Zweck),
		Auswahl:    dto.Auswahl,
		Behalten:   dto.Behalten,
		Ablauf:     ablauf,
	}, nil
}

// zeitpunktAus liest einen Zeitpunkt — **leer bleibt der Nullwert**.
func zeitpunktAus(feld, text string) (time.Time, error) {
	if text == "" {
		return time.Time{}, nil
	}

	gelesen, fehler := time.Parse(time.RFC3339, text)
	if fehler != nil {
		return time.Time{}, domain.MitGrund(
			domain.NewFieldError(feld, domain.SchluesselValidierung,
				"%q ist kein Zeitpunkt.", text), fehler, "einen Zeitpunkt lesen")
	}

	return gelesen, nil
}

// alsSichtkarten macht aus den Sichtkarten ihre DTOs.
func alsSichtkarten(sicht []domain.Sichtkarte) []SichtkarteDTO {
	dtos := make([]SichtkarteDTO, 0, len(sicht))

	for _, karte := range sicht {
		dtos = append(dtos, SichtkarteDTO{
			Referenz:    karte.Referenz,
			Printing:    karte.Printing,
			Menge:       karte.Menge,
			Zustand:     string(karte.Zustand),
			Sprache:     karte.Sprache,
			Finish:      string(karte.Finish),
			Bezeichnung: karte.Bezeichnung,
			Wunsch:      karte.Wunsch,
		})
	}

	return dtos
}

// alleUmfaenge gibt die Aufzählung aus `domain` als Text.
func alleUmfaenge() []string {
	umfaenge := make([]string, 0, len(domain.DieFreigabeumfaenge))
	for _, umfang := range domain.DieFreigabeumfaenge {
		umfaenge = append(umfaenge, string(umfang))
	}

	return umfaenge
}

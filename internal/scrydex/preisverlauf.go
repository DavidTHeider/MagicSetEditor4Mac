package scrydex

// **Der Preisverlauf** — Werkstück A5 von Phase 5.3
// (#1444).
//
// # Zwei Preise je Zeile, und sie sind zwei Aussagen
//
// **Nachgelesen am 2026-09-11** unter
// <https://scrydex.com/docs/magicthegathering/price-history>: je Tag und
// Ausführung stehen `low` **und** `market` — das niedrigste Angebot und der
// Marktpreis. Beides sind Forderungen, und wer sie in **eine** Reihe legte,
// mischte zwei Fragen.
//
// Getrennt werden sie über den Quellentrenner: `scrydex:low` und
// `scrydex:market`, dieselbe Gestalt wie `mtgjson:cardkingdom`
// ([0093](../../docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)).

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Familie ist der Name, unter dem alles von Scrydex läuft.
const Familie domain.Preisquelle = "scrydex"

// Die beiden Preise, die Scrydex je Zeile nennt.
const (
	// QuelleNiedrigst ist das niedrigste Angebot.
	QuelleNiedrigst = Familie + domain.Quellentrenner + "low"

	// QuelleMarkt ist der Marktpreis.
	QuelleMarkt = Familie + domain.Quellentrenner + "market"
)

// Verlaufsweg ist, was hinter der Kartenkennung steht.
const Verlaufsweg = "/price_history"

// Verlaufswunsch engt ein, wonach gefragt wird.
//
// **Nur, was gesetzt ist, geht hinaus**: ein Nullwert als Filter nähme
// stillschweigend etwas weg, wonach niemand gefragt hat.
type Verlaufswunsch struct {
	// Tage ist die Zahl der Tage zurück — 0 heißt: keine Angabe.
	Tage int

	// Von und Bis sind Tage in der Form `YYYY-MM-DD` — leer heißt: keine.
	Von string
	Bis string

	// Ausfuehrung und Zustand engen ein — leer heißt: alle.
	Ausfuehrung string
	Zustand     string
}

// Preispunkt ist ein Preis an einem Tag.
type Preispunkt struct {
	// Tag ist der Tag, den die Quelle nennt — `YYYY-MM-DD`.
	Tag string

	Ausfuehrung string

	// Zustand ist die Erhaltung **in der Ausdrucksweise von Scrydex**.
	//
	// **Sie wird nicht übersetzt** — dieselbe Regel wie bei CardTrader
	// ([0095](../../docs/entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md)):
	// eine Gleichsetzung zweier Skalen, die niemand nachprüfen kann, wäre
	// nicht mehr zu entfernen.
	Zustand string

	// Art ist `raw` oder `graded`.
	Art string

	// Niedrigst und Markt sind Beträge in der kleinsten Einheit.
	Niedrigst int64
	Markt     int64
	Waehrung  string

	Makellos  bool
	Signiert  bool
	Fehldruck bool
}

// verlaufDTO ist die Gestalt, in der Scrydex den Verlauf schickt.
type verlaufDTO struct {
	Daten []struct {
		Tag    string `json:"date"`
		Preise []struct {
			Ausfuehrung string      `json:"variant"`
			Zustand     string      `json:"condition"`
			Art         string      `json:"type"`
			Niedrigst   json.Number `json:"low"`
			Markt       json.Number `json:"market"`
			Waehrung    string      `json:"currency"`
			Makellos    bool        `json:"is_perfect"`
			Signiert    bool        `json:"is_signed"`
			Fehldruck   bool        `json:"is_error"`
		} `json:"prices"`
	} `json:"data"`
}

// Preisverlauf gibt die Preise einer Karte über die Zeit.
func (k *Klient) Preisverlauf(ctx context.Context, kennung string, wunsch Verlaufswunsch) (
	[]Preispunkt, []domain.Befund, error,
) {
	sauber := strings.TrimSpace(kennung)
	if sauber == "" {
		return nil, nil, fmt.Errorf(
			"%w: ohne Kartenkennung gibt Scrydex keinen Verlauf",
			domain.ErrValidation)
	}

	// **Die Kennung wird entschärft** — ein Schrägstrich darin machte sonst
	// einen neuen Weg auf, und der zeigte irgendwohin.
	adresse := k.grundadresse + Kartenweg + "/" + url.PathEscape(sauber) + Verlaufsweg

	if abfrage := wunsch.abfrage(); len(abfrage) > 0 {
		adresse += "?" + abfrage.Encode()
	}

	antwort, fehler := k.frage(ctx, adresse)
	if fehler != nil {
		return nil, nil, fehler
	}

	var dto verlaufDTO

	leser := json.NewDecoder(bytes.NewReader(antwort.Rumpf))
	leser.UseNumber()

	if lesefehler := leser.Decode(&dto); lesefehler != nil {
		return nil, nil, domain.MitGrund(domain.ErrValidation, lesefehler,
			"den Preisverlauf von Scrydex lesen")
	}

	return punkteAus(dto, sauber)
}

// abfrage baut die Abfrageparameter — und nur, was gesetzt ist.
func (w Verlaufswunsch) abfrage() url.Values {
	abfrage := url.Values{}

	for name, wert := range map[string]string{
		"start_date": w.Von, "end_date": w.Bis,
		"variant": w.Ausfuehrung, "condition": w.Zustand,
	} {
		if wert != "" {
			abfrage.Set(name, wert)
		}
	}

	if w.Tage > 0 {
		abfrage.Set("days", strconv.Itoa(w.Tage))
	}

	return abfrage
}

// punkteAus macht aus dem DTO die neutrale Gestalt.
func punkteAus(dto verlaufDTO, kennung string) ([]Preispunkt, []domain.Befund, error) {
	var (
		punkte  []Preispunkt
		befunde []domain.Befund
	)

	for _, tag := range dto.Daten {
		// **Der Tag wird geprüft, bevor seine Preise zählen** — ein Verlauf
		// mit einem Nullzeitpunkt sortierte sich vor jeden echten Tag (§8).
		if _, fehler := time.Parse(time.DateOnly, tag.Tag); fehler != nil {
			befunde = append(befunde, domain.Befund{
				Datei: kennung, Ort: tag.Tag, Art: domain.ArtUmgebung,
				Meldung: fmt.Sprintf(
					"Scrydex nennt den Tag %q, den diese Anwendung nicht lesen "+
						"kann. Seine Preise werden übergangen.", tag.Tag),
			})

			continue
		}

		for _, preis := range tag.Preise {
			if preis.Waehrung == "" {
				befunde = append(befunde, domain.Befund{
					Datei: kennung, Ort: tag.Tag, Art: domain.ArtUmgebung,
					Meldung: fmt.Sprintf(
						"Scrydex nennt am %s einen Preis ohne Währung. "+
							"Ein Preis ohne sie ist keine Aussage und wird "+
							"übergangen.", tag.Tag),
				})

				continue
			}

			niedrigst, marktwert, umrechenfehler := betraege(
				preis.Niedrigst, preis.Markt, preis.Waehrung)
			if umrechenfehler != nil {
				befunde = append(befunde, domain.Befund{
					Datei: kennung, Ort: tag.Tag, Art: domain.ArtUmgebung,
					Meldung: fmt.Sprintf(
						"Scrydex nennt am %s einen Preis, der keine Zahl ist: %v",
						tag.Tag, umrechenfehler),
				})

				continue
			}

			punkte = append(punkte, Preispunkt{
				Tag: tag.Tag, Ausfuehrung: preis.Ausfuehrung,
				Zustand: preis.Zustand, Art: preis.Art,
				Niedrigst: niedrigst, Markt: marktwert, Waehrung: preis.Waehrung,
				Makellos: preis.Makellos, Signiert: preis.Signiert,
				Fehldruck: preis.Fehldruck,
			})
		}
	}

	return punkte, befunde, nil
}

// betraege macht aus zwei Zahlen zwei Beträge in der kleinsten Einheit.
//
// **Die Währung sagt, wie viele Stellen das sind** — LOC-03.AK-02 (#1689).
// Scrydex führt sechs Spiele und nennt seine Währung je Preis.
func betraege(niedrigst, markt json.Number, waehrung string) (
	untenWert, obenWert int64, fehler error,
) {
	untenWert, fehler = domain.MarktpreisAus(niedrigst.String(), waehrung)
	if fehler != nil {
		return 0, 0, fehler
	}

	obenWert, fehler = domain.MarktpreisAus(markt.String(), waehrung)
	if fehler != nil {
		return 0, 0, fehler
	}

	return untenWert, obenWert, nil
}

// AlsStaende macht aus einem Punkt seine **zwei** Preisbeobachtungen.
//
// **Zwei und nicht eine** (0096): `low` und `market` beantworten verschiedene
// Fragen, und eine gemeinsame Reihe hieße, dass mal die eine und mal die
// andere darin stünde.
func (p Preispunkt) AlsStaende(druck string) []domain.Preisstand {
	// **Der Tag ist geprüft** — [punkteAus] lässt keinen anderen durch.
	beobachtet, _ := time.Parse(time.DateOnly, p.Tag) //nolint:errcheck // siehe oben

	grund := domain.Preisstand{
		Printing: druck,
		Markt:    domain.MarktPapier,
		Art:      domain.PreisartVerkauf,
		Waehrung: p.Waehrung,
		Finish:   domain.Finish(p.Ausfuehrung),
		Zustand:  p.Zustand,
		Lage:     domain.PreisGenannt,
		// **Der Beobachtungstag kommt von der Quelle** (0093): wann gefragt
		// wurde, weiß der Dienst; welchen Tag die Quelle meint, nur sie.
		Beobachtet: beobachtet,
	}

	niedrigst, marktwert := grund, grund
	niedrigst.Quelle, niedrigst.Wert = QuelleNiedrigst, p.Niedrigst
	marktwert.Quelle, marktwert.Wert = QuelleMarkt, p.Markt

	return []domain.Preisstand{niedrigst, marktwert}
}

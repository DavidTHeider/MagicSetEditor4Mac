package scrydex

// **Der Kontostand** — Werkstück A5 von Phase 5.3
// (#1444).
//
// # Wofür es das braucht
//
// **Scrydex rechnet in Krediten, nicht in Anfragen.** Wie viele eine Anfrage
// kostet, hängt davon ab, was sie holt; Vision kostet fünf. Eine Grenze in
// „Anfragen je Tag" ginge daneben, und wer über den Tarif hinausgeht, zahlt
// Gebühren.
//
// # Und warum dieser Stand die Grenze **nicht** ersetzt
//
// **Nachgelesen am 2026-09-11**: „Usage info is updated every 20-30 minutes."
// Eine Zahl, die eine halbe Stunde alt sein kann, ist keine Sperre. Sie steht
// **neben** der lokalen Zählung und sagt, wie alt sie ist — die Grenze, die
// gilt, ist die lokale
// ([0096](../../docs/entscheidungen/0096-welche-optionalen-datenquellen-zugaenglich-sind.md)).

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Kontoweg ist der Weg zum Verbrauch eines Kontos.
//
// **Nachgelesen am 2026-09-11** unter
// <https://scrydex.com/docs/getting-started/rate-limits>.
const Kontoweg = "/account/v1/usage"

// Aktualisierungsfrist ist, wie alt der Stand beim Anbieter sein kann.
//
// **Die obere der beiden genannten Zahlen**: „updated every 20-30 minutes".
// Wer die untere nähme, verspräche eine Frische, die der Anbieter nicht
// zusagt.
const Aktualisierungsfrist = 30 * time.Minute

// Kontostand ist, was ein Konto bei Scrydex noch hat.
type Kontostand struct {
	// Gesamt, Rest und Verbraucht sind Kredite — ganze Zahlen.
	Gesamt     int64
	Rest       int64
	Verbraucht int64

	// Ueberzugssatz ist, was ein zusätzlicher Kredit kostet — **als Text**.
	//
	// Es ist ein Geldbetrag, und `0.0004` ist in `float64` nicht `0.0004`.
	// Dieses Paket rechnet damit nicht; es reicht ihn weiter, wie er kam
	// (0093).
	Ueberzugssatz string
}

// kontostandDTO ist die Gestalt, in der Scrydex ihn schickt.
type kontostandDTO struct {
	Gesamt        int64       `json:"total_credits"`
	Rest          int64       `json:"remaining_credits"`
	Verbraucht    int64       `json:"used_credits"`
	Ueberzugssatz json.Number `json:"overage_credit_rate"`
}

// Kontostand fragt, wie viele Kredite das Konto noch hat.
func (k *Klient) Kontostand(ctx context.Context) (Kontostand, error) {
	antwort, fehler := k.frage(ctx, k.grundadresse+Kontoweg)
	if fehler != nil {
		return Kontostand{}, fehler
	}

	var dto kontostandDTO

	// **Über einen Decoder mit `UseNumber`** und nicht über `json.Unmarshal`:
	// sonst läse Go den Satz als `float64`, und der Text `0.0004` käme als
	// `0.00039999999999999996` zurück.
	leser := json.NewDecoder(bytes.NewReader(antwort.Rumpf))
	leser.UseNumber()

	if lesefehler := leser.Decode(&dto); lesefehler != nil {
		return Kontostand{}, domain.MitGrund(domain.ErrValidation, lesefehler,
			"den Kontostand von Scrydex lesen")
	}

	return Kontostand{
		Gesamt: dto.Gesamt, Rest: dto.Rest, Verbraucht: dto.Verbraucht,
		Ueberzugssatz: dto.Ueberzugssatz.String(),
	}, nil
}

// frage stellt eine Anfrage mit Zugang und gibt die geglückte Antwort.
//
// **Die gemeinsame Tür** von Kontostand und Preisverlauf: der Zugang wird an
// **einer** Stelle geprüft und die Fehlermeldung an einer gedeutet. Eine
// Prüfung je Weg wäre so oft dieselbe Zusage und einmal die vergessene (§10).
func (k *Klient) frage(ctx context.Context, adresse string) (netz.Antwort, error) {
	if fehlt := k.zugang.Fehlt(); fehlt != "" {
		return netz.Antwort{}, fmt.Errorf(
			"%w: für Scrydex fehlt %s — beides gehört in den Schlüsselbund",
			domain.ErrValidation, fehlt)
	}

	antwort, holfehler := netz.Hole(ctx, k.klient, k.drossel, k.kenner(), adresse, hoechstens)
	if holfehler != nil {
		return netz.Antwort{}, fmt.Errorf("Scrydex fragen: %w", holfehler)
	}

	if antwort.Status != netz.StatusGut {
		return netz.Antwort{}, fmt.Errorf("%w: Scrydex antwortet mit %d — %s",
			domain.ErrValidation, antwort.Status, meldungAus(antwort.Rumpf))
	}

	return antwort, nil
}

// kenner ist, womit sich dieses Programm bei Scrydex vorstellt.
func (k *Klient) kenner() netz.Kenner {
	return netz.Kenner{
		Name:    "MagicSetEditor4Mac",
		Annahme: "application/json",
		Zusatz: map[string]string{
			"X-Api-Key": k.zugang.Schluessel,
			"X-Team-ID": k.zugang.Team,
		},
	}
}

// meldungAus zieht den Satz des Anbieters aus einem Fehlerrumpf.
//
// **Er sagt, was zu tun ist** — „Invalid credentials" nennt den Grund, „401"
// nennt ihn nicht (§8). Steht keiner da, bleibt es bei der Zahl.
func meldungAus(rumpf []byte) string {
	var umschlag antwortDTO

	if json.Unmarshal(rumpf, &umschlag) != nil ||
		umschlag.Fehler == nil || umschlag.Fehler.Meldung == "" {
		return "ohne Angabe eines Grundes"
	}

	return umschlag.Fehler.Meldung
}

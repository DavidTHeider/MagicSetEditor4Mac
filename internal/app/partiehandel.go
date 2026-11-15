package app

// **Nach der Partie mit demselben Gegenüber handeln** — Werkstück `T7` von
// Phase 5.6
// (#1472).
//
// # Warum über den Fingerabdruck und nicht über eine neue Paarung
//
// **Weil beide sich schon kennen.** Ein Gast an einem freigegebenen Tisch hat
// seine Peeridentität mitgebracht; sie über eine zweite Paarung noch einmal
// zu bestätigen wäre eine Frage, die schon beantwortet ist. Das Issue
// verlangt es wörtlich: „ohne neues Konto oder erneutes Pairing".
//
// # Und warum trotzdem nichts sichtbar wird
//
// **Weil eine Partie keine Freigabe ist.** Wer miteinander gespielt hat, darf
// deshalb nicht in die Sammlung des anderen sehen — das Issue sagt auch das:
// „fragt aber neue Sammlungsfreigabe ab". Was hier zurückkommt, ist **wer**
// am Tisch saß, und sonst nichts; was er sehen darf, entscheidet `T3`.
//
// # Warum der Sitz und nicht der Gast gefragt wird
//
// **Weil eine Sitzung einen Rechner belegt und keinen Sitz** (G5). Die
// Freigabe vergibt Sitze an Fingerabdrücke; der Sitz ist, was am Tisch steht,
// und der Fingerabdruck ist, wer dahintersteht.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// HandelspartnerAntwort ist das Gegenüber zu einem Sitz.
type HandelspartnerAntwort struct {
	Sitz string `json:"sitz"`

	// Kontakt gilt nur mit Bekannt.
	Kontakt Peerkontakt `json:"kontakt"`

	// Bekannt sagt, ob zu diesem Sitz ein bekanntes Gegenüber gehört.
	//
	// **Eine eigene Angabe und keine leere Kennung**: „niemand saß da" und
	// „da saß jemand, den ich nicht kenne" sind zwei Auskünfte, und die
	// zweite ist die interessantere.
	Bekannt bool `json:"bekannt"`

	Fehler *Fehler `json:"fehler"`
}

// Handelspartner sagt, welcher bekannte Kontakt auf einem Sitz saß.
//
// **Es öffnet nichts und fragt nichts an.** Es beantwortet die eine Frage,
// die nach einer Partie offen ist: mit wem habe ich gerade gespielt? Was
// daraus wird, entscheidet der Mensch — und die Freigabe des anderen.
func (a *App) Handelspartner(
	ctx context.Context, sitz string,
) HandelspartnerAntwort {
	ctx = a.imProfil(ctx)

	stand := a.PartieFreigabeStand(ctx)
	if stand.Fehler != nil {
		return HandelspartnerAntwort{
			Sitz: sitz, Kontakt: leererKontakt(), Fehler: stand.Fehler,
		}
	}

	return a.kontaktZumAbdruck(ctx, sitz, abdruckAmSitz(stand.Vergeben, sitz))
}

// abdruckAmSitz sucht den Fingerabdruck, der auf einem Sitz sitzt.
//
// **Eine eigene Funktion, damit sie zu fahren ist.** Eine laufende Freigabe
// mit besetztem Sitz herzustellen braucht zwei Rechner; die Frage „wer sitzt
// auf a?" braucht eine Liste.
func abdruckAmSitz(vergeben []SitzvergabeDTO, sitz string) string {
	for _, vergabe := range vergeben {
		if vergabe.Sitz == sitz {
			return vergabe.Fingerabdruck
		}
	}

	return ""
}

// kontaktZumAbdruck sucht den bekannten Kontakt zu einem Fingerabdruck.
//
// **Über den Fingerabdruck und nicht über die Kennung**, weil die Freigabe
// nur ihn führt — und weil er aus der Kennung entsteht: wer denselben hat,
// hat dieselbe Identität. Ein zweiter Nachweis wäre keiner.
func (a *App) kontaktZumAbdruck(
	ctx context.Context, sitz, abdruck string,
) HandelspartnerAntwort {
	// **Ein leerer Sitz kommt hier an und nicht davor.** Zwei Ausgänge in
	// zwei Verfahren ließen den einen ungeprüft, sobald der andere seinen
	// Weg ändert — und „da sitzt niemand" ist dieselbe Frage wie „wer sitzt
	// da".
	if abdruck == "" {
		return HandelspartnerAntwort{
			Sitz: sitz, Kontakt: leererKontakt(),
			Fehler: Uebersetze(ctx,
				domain.MitGrund(domain.ErrNotFound, nil,
					"auf dem Sitz %s sitzt niemand von außen", sitz)),
		}
	}

	kontakte := a.Peerkontakte(ctx)
	if kontakte.Fehler != nil {
		return HandelspartnerAntwort{
			Sitz: sitz, Kontakt: leererKontakt(), Fehler: kontakte.Fehler,
		}
	}

	for _, kontakt := range kontakte.Kontakte {
		if kontakt.Fingerabdruck == abdruck {
			return HandelspartnerAntwort{
				Sitz: sitz, Kontakt: kontakt, Bekannt: true,
			}
		}
	}

	// **„Kenne ich nicht" ist kein Fehler.** Wer als Gast an einem Tisch
	// saß, muss deshalb nicht in der Kontaktliste stehen; er kann
	// aufgenommen werden, und das ist die Antwort auf diese Auskunft.
	return HandelspartnerAntwort{Sitz: sitz, Kontakt: leererKontakt()}
}

// leererKontakt gibt einen Kontakt, dessen Listen leer und nicht `nil` sind.
//
// **Eine leere Liste und kein `null`** — die Zusage aller Antworten dieses
// Pakets (§3): das Fenster zählt sie, und `null` zählt nicht.
func leererKontakt() Peerkontakt {
	return Peerkontakt{Rechte: []string{}}
}

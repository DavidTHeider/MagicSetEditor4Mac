package app

// **Den Ausgang zu einem Gegenüber tragen** —
// #1634.
//
// # Warum die Adresse nicht im Ruf steht
//
// **Weil sie am Kontakt steht.** Wer handelt, hat sein Gegenüber bestätigt
// und seine Adresse eingetragen; eine zweite im Ruf wäre die Gelegenheit,
// eine Kennung an eine fremde Adresse zu schicken. Die Sitzung schlägt sie
// nach (`adresswaehler`), und was dort sitzt, sagt der Handschlag.
//
// # Und warum der Ausgang danach voll bleibt
//
// **Weil „abgeschickt" nicht „angekommen" heißt** (`T5`). Geräumt wird er
// von der Bestätigung des Gegenübers.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Zustellungsbericht sagt, was hinausgegangen ist.
type Zustellungsbericht struct {
	// Gegenueber ist die Peerkennung, zu der getragen wurde.
	Gegenueber string `json:"gegenueber"`

	// Getragen ist die Zahl der Nachrichten, die über die Leitung gingen.
	//
	// **Getragen und nicht zugestellt**: ob sie angekommen sind, sagt die
	// Bestätigung von drüben und nicht diese Zahl.
	Getragen int `json:"getragen"`

	Fehler *Fehler `json:"fehler"`
}

// Handelszustellung baut eine Leitung zu einem Gegenüber und trägt den
// Ausgang hinüber.
//
// **Ohne Fenster** — das ist der Punkt dieses Weges
// ([0121](../../docs/entscheidungen/0121-eine-leitung-ohne-fenster.md)). Er
// endet, wenn nichts mehr da ist; er bleibt nicht liegen.
func (a *App) Handelszustellung(
	ctx context.Context, kennung string,
) Zustellungsbericht {
	ctx = a.imProfil(ctx)

	bericht := Zustellungsbericht{Gegenueber: kennung}

	// **`externSenden` und nicht `lokalSchreiben`** — der Unterschied zu
	// `HandelSchicken` eine Zeile weiter: jenes legt in den **eigenen**
	// Ausgang, dieses baut eine Verbindung zu einem fremden Rechner und
	// schickt. Was hinausgeht, kommt nicht zurück (O1,
	// [docs/AUTORISIERUNG.md](../../docs/AUTORISIERUNG.md)).
	if fehler := a.darf(ctx, domain.WirkungExternSenden); fehler != nil {
		bericht.Fehler = Uebersetze(ctx, fehler)

		return bericht
	}

	if peer.Fingerabdruck(peer.Kennung(kennung)) == "" {
		bericht.Fehler = Uebersetze(ctx, domain.MitGrund(
			domain.ErrValidation, nil,
			"%q ist keine Peerkennung", kennung))

		return bericht
	}

	griff, fehler := a.sitzungsverwaltung().Oeffne(
		ctx, peer.Kennung(kennung), domain.ZweckHandel)
	if fehler != nil {
		bericht.Fehler = Uebersetze(ctx, fehler)

		return bericht
	}

	defer schliesseHandelssitzung(ctx, griff)

	getragen, fehler := a.handelspost().Nachliefern(
		ctx, handelsstrecke{griff: griff}, kennung)

	bericht.Getragen = getragen

	if fehler != nil {
		bericht.Fehler = Uebersetze(ctx, fehler)
	}

	return bericht
}

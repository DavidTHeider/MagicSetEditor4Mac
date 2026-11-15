package service

// **Eine Handelsnachricht über eine Leitung** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Was hier geschieht und was nicht
//
// **Hier wird verpackt und ausgepackt.** Was eine Nachricht bedeutet, weiß
// `domain`; wie sie über den Draht kommt, weiß `internal/leitung`. Diese
// Stelle ist dazwischen — dieselbe Bauform wie [Partiesitzung] in Phase 5.5.
//
// # Warum die Strecke eine Schnittstelle ist
//
// **Damit hier kein Transporttyp steht.** Das Issue verlangt ausdrücklich
// „keine libp2p-/Pion-Typen im Service"; was dieser Dienst braucht, sind zwei
// Wege — senden und empfangen —, und wer sie erfüllt, entscheidet die Naht.
// Über dieselbe Schnittstelle laufen die Schleife im selben Prozess, TCP im
// eigenen Netz und tailcat über fremde.

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Handelskanal ist der Name, unter dem gehandelt wird.
//
// **Eine Fassung im Namen**, damit ein Gegenüber mit einer anderen gar nicht
// erst zustande kommt — dieselbe Vorsicht wie bei der Protokollfassung im
// Umschlag, nur eine Ebene tiefer.
const Handelskanal = "mse4mac/trade/1"

// Handelsstrecke ist, was die Post vom Transport braucht.
type Handelsstrecke interface {
	Sende(ctx context.Context, nachricht []byte) error
	Empfange(ctx context.Context) ([]byte, error)
}

// SendeUeber schickt eine Nachricht über eine Strecke — **und legt sie
// vorher in den Ausgang**.
//
// **In dieser Reihenfolge.** Wer erst schickt und dann einreiht, verliert die
// Nachricht, wenn der Prozess dazwischen endet: drüben ist sie angekommen,
// hier weiß niemand davon, und die Wiederaufnahme liefert sie nicht nach.
func (p *Handelspost) SendeUeber(
	ctx context.Context, strecke Handelsstrecke, empfaenger string,
	nachricht Handelsnachricht,
) error {
	if fehler := p.Senden(ctx, empfaenger, nachricht); fehler != nil {
		return fehler
	}

	// **Ohne Fehlerzweig, und das ist keine Nachlässigkeit**: eine
	// [Handelsnachricht] trägt Zeichenketten, Zahlen, einen Wahrheitswert,
	// einen Zeitpunkt und Scheiben davon — `encoding/json` kann daran nicht
	// scheitern. Ein Zweig, den kein Lauf betreten kann, ist eine Zusage,
	// die niemand hält (§10). Dieselbe Stelle wie in
	// `storage.Handelsablage.Schreiben`; wer dem Verbund ein `any` gibt,
	// ändert sie mit.
	roh, _ := json.Marshal(nachricht) //nolint:errcheck // der Grund steht darüber

	return strecke.Sende(ctx, roh)
}

// EmpfangeUeber holt **eine** Nachricht von der Strecke und behandelt sie.
//
// **Eine und keine Schleife.** Wer hier schliefe, bis etwas kommt, nähme dem
// Aufrufer die Entscheidung, wann er aufhört — und ein Abbruch ginge nur noch
// über den Kontext. Die Schleife gehört dorthin, wo jemand sie beenden kann.
func (p *Handelspost) EmpfangeUeber(
	ctx context.Context, strecke Handelsstrecke, sender string,
) (Zustellungsantwort, error) {
	roh, fehler := strecke.Empfange(ctx)
	if fehler != nil {
		return Zustellungsantwort{}, fehler
	}

	var nachricht Handelsnachricht

	// **Ein unlesbarer Rumpf ist ein Befund und kein Absturz** (§8). Was von
	// drüben kommt, ist der Normalfall für fremde Daten: eine Nachricht, die
	// sich nicht deuten lässt, endet hier mit ihrer Stelle.
	if fehler := json.Unmarshal(roh, &nachricht); fehler != nil {
		return Zustellungsantwort{}, domain.MitGrund(domain.ErrValidation, fehler,
			"Diese Nachricht ließ sich nicht lesen.")
	}

	// **Eine Sichtanfrage bekommt ihre Antwort auf derselben Strecke** (P11)
	// — sie ist kein Zustand eines Handels und geht an Briefkasten und
	// Ausgang vorbei.
	if nachricht.Umschlag.Art == domain.NachrichtSichtanfrage {
		return p.sichtBeantworten(ctx, strecke, sender, nachricht)
	}

	return p.Empfangen(ctx, sender, nachricht)
}

// Nachliefern trägt über eine Strecke hinaus, was im Ausgang liegt.
//
// # Warum das neben [Handelspost.SendeUeber] steht
//
// **Weil jener einreiht und dann schickt.** Was schon dasteht, noch einmal
// einzureihen wäre bestenfalls wirkungslos — der Ausgang ist idempotent —
// und schlimmstenfalls eine zweite Zeile unter neuer Kennung. Hier wird
// **getragen**, was dasteht.
//
// # Und warum der Ausgang danach voll bleibt
//
// **Weil „abgeschickt" nicht „angekommen" heißt** (`T5`). Geräumt wird er
// von der **Bestätigung**, und die kommt von drüben; wer hier räumte,
// verlöre jede Nachricht, die zwischen Senden und Ankommen verloren geht.
//
// **Zurück kommt die Zahl** — sie sagt einem Menschen, ob etwas unterwegs
// war, und einem Skript, ob es etwas zu warten gibt
// (#1634).
func (p *Handelspost) Nachliefern(
	ctx context.Context, strecke Handelsstrecke, empfaenger string,
) (int, error) {
	if p.ausgang == nil {
		return 0, domain.MitGrund(domain.ErrValidation, nil,
			"Diese Post kann nur empfangen und nichts senden.")
	}

	offen, fehler := p.ausgang.Offene(ctx, empfaenger)
	if fehler != nil {
		return 0, fehler
	}

	getragen := 0

	for _, zeile := range offen {
		// **Ohne Fehlerzweig am Kodierer** — derselbe Grund wie bei
		// [Handelspost.SendeUeber] eine Ebene höher.
		roh, _ := json.Marshal(zeile.Nachricht) //nolint:errcheck // der Grund steht dort

		// **Ein Abriss mitten in der Reihe bricht ab und nennt die Zahl.**
		// Weiterzumachen hieße, gegen eine Leitung zu schreiben, die es
		// nicht mehr gibt; die übrigen bleiben liegen und gehen beim
		// nächsten Ruf.
		if fehler := strecke.Sende(ctx, roh); fehler != nil {
			return getragen, fehler
		}

		getragen++
	}

	return getragen, nil
}

package app

// **Ein Mensch gibt einen kostenpflichtigen Auftrag frei** —
// #1539.
//
// # Der Befund
//
// **Es gab keinen Weg dazu.** `kibudget.Reservieren` weist ab, solange kein
// Nachweis vorliegt; gefüllt wurde er nur von [App.MitKIFreigabe], und die
// rief **kein Produktivcode**. Am laufenden Fenster scheiterte damit jeder
// kostenpflichtige KI-Auftrag mit „braucht die ausdrückliche Freigabe eines
// Menschen" — und niemand konnte sie geben.
//
// **Der Mechanismus war fertig gebaut**: `service.Nachweise.Ausstellen`,
// `Einloesen`, `domain.Nachweisbindung` mit `Kosten` und `Waehrung`. Er war
// nur nicht angeschlossen.
//
// # Warum die Freigabe einem Auftrag gilt und nicht einer Sitzung
//
// **Weil sie sonst eine Vollmacht wäre.** Ein Nachweis trägt eine Grenze und
// eine Frist; wer ihn einlöst, verbraucht ihn. Ein zweiter Auftrag braucht
// einen zweiten — das hält `Nachweise.Einloesen`, und die Zusage steht hier,
// damit sie jemand liest.
//
// # Und warum es sie über MCP nicht gibt
//
// **Weil ein Modell kein Mensch ist.** `Nachweise.Ausstellen` weist jeden
// Modus außer `fenster` ab; eine Bindung daneben, die es doch täte, hebelte
// genau die Zusage aus, für die es Nachweise gibt
// ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)).

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// KIFreigabefrist ist, wie lange eine erteilte Freigabe gilt.
//
// **Fünf Minuten**, und die Zahl folgt dem Zweck: sie deckt den Weg vom Knopf
// bis zum Ruf, und nicht den Nachmittag. Wer länger braucht, gibt sie neu.
const KIFreigabefrist = 5 * time.Minute

// NachweisAntwort ist die erteilte Freigabe.
type NachweisAntwort struct {
	// Kennung ist, was der nächste Auftrag vorzeigt — leer bei einem Fehler.
	Kennung string `json:"kennung"`

	// Laeuftab sagt, bis wann sie gilt.
	Laeuftab string `json:"laeuftab"`

	Fehler *Fehler `json:"fehler"`
}

// KIFreigabeErteilen stellt den Nachweis für **einen** Auftrag aus.
//
// **Der Höchstbetrag kommt vom Aufrufer und nicht von hier**: das Fenster
// zeigt ihn, bevor es fragt, und der Mensch gibt genau ihn frei.
func (a *App) KIFreigabeErteilen(
	ctx context.Context, konto string, betrag int64, waehrung string,
) NachweisAntwort {
	ctx = a.imProfil(ctx)

	if a.autorisierung == nil {
		return NachweisAntwort{Fehler: Uebersetze(ctx, fehlenderBestand())}
	}

	nachweis, fehler := a.autorisierung.Ausstellen(ctx, service.Nachweisantrag{
		// **Derselbe Zweck, den die Kostenwache einlöst.** Ein anderer wäre
		// ein Nachweis, der zu nichts passt — `Einloesen` hält beide
		// gegeneinander.
		Zweck: service.KostenjournalZweck,
		Ziel:  konto,
		Grenze: domain.Kostengrenze{
			Betrag: betrag, Waehrung: waehrung,
		},
		Frist: KIFreigabefrist,
	})
	if fehler != nil {
		return NachweisAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return NachweisAntwort{
		Kennung:  nachweis.Kennung,
		Laeuftab: nachweis.Ablauf.UTC().Format(time.RFC3339),
	}
}

// KIFreigabeErteilen ist die Fensterbindung für die Freigabe.
func (b *Bindungen) KIFreigabeErteilen(
	ctx context.Context,
	konto string, betrag int64, waehrung string,
) NachweisAntwort {
	return b.app(ctx).KIFreigabeErteilen(b.ruf(ctx), konto, betrag, waehrung)
}

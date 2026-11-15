package app

// **Was mit einer hereinkommenden Handelssitzung geschieht** —
// #1634.
//
// # Warum es das gibt
//
// **Weil eine Handelsnachricht bisher über keine Leitung kam.**
// `service.Handelspost.EmpfangeUeber` stand seit `T5` gebaut da, und der
// einzige Aufrufer war ein Lauf über die Schleife im selben Prozess. Der
// Horchposten nahm an — und reichte jede Sitzung an die **Partie**; eine mit
// dem Zweck `handel` fiel still auf den Boden.
//
// # Warum die Schleife hier steht und nicht im Dienst
//
// **Weil `EmpfangeUeber` genau eine Nachricht holt**, und zwar mit Absicht:
// „wer hier schliefe, bis etwas kommt, nähme dem Aufrufer die Entscheidung,
// wann er aufhört". Diese Stelle ist der Aufrufer, und ihr Abbruch ist der
// Kontext des Horchpostens.

import (
	"context"
	"errors"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// errKeinZweckDafuer ist die Absage an eine Sitzung, für die es hier keinen
// Weg gibt.
var errKeinZweckDafuer = errors.New(
	"für diesen Zweck gibt es auf diesem Rechner keinen Weg")

// sitzungAufnehmen reicht eine angenommene Sitzung an den Weg ihres Zwecks.
//
// # Warum ein `switch` und nicht zwei Aufnahmen nebeneinander
//
// **Weil der Horchposten eine Aufnahme kennt und nicht zwei.** Zwei
// Registrierungen hießen zwei Stellen, an denen jemand einen Zweck vergisst —
// hier steht jeder einmal, und was keiner ist, wird **abgewiesen** statt
// still fallen gelassen (§10).
func (a *App) sitzungAufnehmen(
	ctx context.Context, griff service.Angenommene,
) {
	switch griff.Stand().Zweck {
	case domain.ZweckPartie:
		a.partiesitzungAufnehmen(ctx, griff)

	case domain.ZweckHandel:
		a.handelssitzungAufnehmen(ctx, griff)

	// **Die Paarung ist das Kennenlernen** — und seit NET-01b heißt das:
	// jemand löst eine Einladung ein.
	case domain.ZweckPaarung:
		a.einloesungAufnehmen(ctx, griff)

	default:
		a.sitzungAbweisen(ctx, griff, errKeinZweckDafuer)
	}
}

// handelssitzungAufnehmen nimmt entgegen, was über diese Leitung kommt.
//
// **Eine Nachricht nach der anderen, bis die Leitung endet.** Was ankommt,
// geht durch dieselbe Post wie ein Umschlag aus einer Datei: derselbe
// Idempotenzvermerk, dieselbe Wirkung in derselben Transaktion (`T5`).
//
// **Der Absender ist, wer wirklich spricht** — die Sitzung hat ihn beim
// Handschlag geprüft. Was im Umschlag steht, ist eine Behauptung, und die
// Post hält beides gegeneinander.
func (a *App) handelssitzungAufnehmen(
	ctx context.Context, griff service.Angenommene,
) {
	// **Ohne [App.imProfil]** (Zweitblick über Phase 6.7, Z2) — dieselbe
	// Verklemmung wie bei der Einlösung: der Kontext trägt das Profil des
	// Horchpostens, und die Umschaltsperre wartet auf ihn.
	ctx = MitNeuemVorgang(ctx)
	stand := griff.Stand()

	defer schliesseHandelssitzung(ctx, griff)

	for {
		// **Je Nachricht ein Werk** — dieselbe Regel wie an jeder anderen
		// Naht dieser Schicht: Bestandsverbindung und Kontaktliste sollen
		// die des Augenblicks sein. Eine Sitzung bleibt, ein Profil kann
		// darunter wechseln, und dann ist die Rechteprüfung von vorhin die
		// falsche.
		antwort, fehler := a.handelspost().EmpfangeUeber(
			ctx, handelsstrecke{griff: griff}, string(stand.Gegenueber))
		if fehler != nil {
			// **Ein Abriss ist kein Befund über dieses Programm.** Wer
			// auflegt, legt auf; was hier gemeldet wird, ist der Grund —
			// und die Schleife endet.
			slog.InfoContext(ctx, "eine Handelssitzung endete",
				"gegenueber", stand.Fingerabdruck, "grund", fehler)

			return
		}

		// **Das Fenster soll es sehen**, ohne dass jemand die Seite neu
		// öffnet — siehe [standnaht].
		a.nebenherGeschrieben()

		slog.InfoContext(ctx, "eine Handelsnachricht kam an",
			"gegenueber", stand.Fingerabdruck,
			"art", string(antwort.Art), "ergebnis", antwort.Ergebnis)
	}
}

// leitungsende ist, was jede Sitzung kann: sich niederlegen.
type leitungsende interface{ Schliesse() error }

// schliesseHandelssitzung legt eine Leitung nieder und meldet, wenn das
// misslingt.
//
// # Warum beide Wege dieselbe Stelle nehmen
//
// **Weil es dieselbe Aussage ist**, und sie stand zweimal da: eine Leitung,
// die sich nicht schließen lässt, ändert **nichts am Ergebnis** (§7, Regel
// 3) — angekommen ist angekommen, getragen ist getragen. Zwei Abschriften
// desselben Satzes laufen auseinander, und die zweite steht an einer Stelle,
// die kein Lauf erreicht.
func schliesseHandelssitzung(ctx context.Context, sitzung leitungsende) {
	if fehler := sitzung.Schliesse(); fehler != nil {
		slog.WarnContext(ctx,
			"eine Handelssitzung ließ sich nicht schließen", "grund", fehler)
	}
}

// handelsstrecke macht aus einer Peersitzung eine [service.Handelsstrecke].
//
// # Warum das Leeren dazugehört
//
// **Weil `Sende` am Griff einreiht und nicht schickt.** Er sammelt, damit
// eine volle Warteschlange den Absender erreicht statt still zu verwerfen;
// hinaus geht es mit `Leere`. Eine Strecke, die nur einreihte, sähe aus wie
// eine, die sendet — und die Nachricht läge noch da.
type handelsstrecke struct{ griff handelssitzung }

// handelssitzung ist der Teil einer Sitzung, den eine Strecke braucht.
//
// **Eine Schnittstelle und nicht `service.Angenommene`**: jene kann mehr —
// Stand, Schließen, die Zustellungsprüfung —, und was hier gebraucht wird,
// sind zwei Wege und das Leeren dazwischen (§2).
type handelssitzung interface {
	Sende(ctx context.Context, nachricht []byte, art service.Zustellung) error
	Leere(ctx context.Context) error
	Empfange(ctx context.Context) ([]byte, error)
}

// Sende reicht eine Nachricht ein und schickt sie sofort hinaus.
func (h handelsstrecke) Sende(ctx context.Context, nachricht []byte) error {
	if fehler := h.griff.Sende(
		ctx, nachricht, service.Zuverlaessig); fehler != nil {
		return fehler
	}

	return h.griff.Leere(ctx)
}

// Empfange holt die nächste Nachricht — blockierend bis zum Kontext.
func (h handelsstrecke) Empfange(ctx context.Context) ([]byte, error) {
	return h.griff.Empfange(ctx)
}

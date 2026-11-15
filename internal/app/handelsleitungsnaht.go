package app

// **Die Naht der Handelsleitung** —
// #1637.
//
// **Weil nur hier `internal/storage` stehen darf** (§2) — derselbe Grund wie
// bei [handelsausgang] und [verbuchungsstelle]. Nebenan steht, **was** ein
// Handelsschritt ist; hier steht, **woraus** er gebaut wird.
//
// # Und hier steht der Adapter, der `T5` gefehlt hat
//
// **`storage.Ausgangsablage` erfüllt [service.Ausgang] nicht.** Die Gestalten
// gehen auseinander: der Bestand hält die Nachricht als **JSON** und kennt
// ihre Folge nicht, der Dienst hält sie als [service.Handelsnachricht] und
// braucht die Folge. Eine Naht ohne Erfüllung ist eine Zusage ohne Einlösung
// (§10) — und genau das war sie, bis zu diesem Issue.

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// ausgangsadapter macht aus der Ablage einen [service.Ausgang].
//
// # Warum die Folge aus der Nachricht kommt und nicht aus einer Spalte
//
// **Weil sie darin steht.** Der Bestand legt Umschlag und Angebot als ganze
// Nachricht ab; die Folge daneben noch einmal zu führen hieße, dieselbe
// Angabe an zwei Stellen zu halten — und die zweite läuft irgendwann von der
// ersten weg. **Was er daraus nicht lesen kann, ist die nächste Folge**; die
// kommt von [storage.Folgeablage], und der Grund steht dort.
type ausgangsadapter struct{ ablage *storage.Ausgangsablage }

// Einreihen legt eine Nachricht in den Ausgang.
//
// **Ohne Fehlerzweig am Kodierer, und das ist keine Nachlässigkeit**: eine
// [service.Handelsnachricht] trägt Zeichenketten, Zahlen, Wahrheitswerte,
// eine Zeit und Scheiben davon — `encoding/json` kann daran nicht scheitern.
// Ein Zweig, den kein Lauf betreten kann, ist eine Zusage, die niemand hält
// (§10). Derselbe Satz steht an `storage.revisionSchreiben`; wer dem Verbund
// ein `any` gibt, ändert diese Stelle mit.
func (a ausgangsadapter) Einreihen(
	ctx context.Context, zeile service.Ausgangsnachricht,
) error {
	roh, _ := json.Marshal(zeile.Nachricht) //nolint:errcheck // der Grund steht darüber

	return a.ablage.Einreihen(ctx, storage.Ausgangszeile{
		NachrichtID: zeile.NachrichtID, Vorgang: zeile.Vorgang,
		Empfaenger: zeile.Empfaenger, Nachricht: string(roh),
	})
}

// Offene gibt, was für dieses Gegenüber noch hinausmuss.
//
// **Eine Zeile, die sich nicht lesen lässt, hält den ganzen Ausgang auf** —
// und das ist Absicht. Sie stillschweigend zu überspringen hieße, eine
// Nachricht zu verlieren, von der die Gegenseite noch nichts weiß.
func (a ausgangsadapter) Offene(
	ctx context.Context, empfaenger string,
) ([]service.Ausgangsnachricht, error) {
	zeilen, fehler := a.ablage.Offene(ctx, empfaenger)
	if fehler != nil {
		return nil, fehler
	}

	offen := make([]service.Ausgangsnachricht, 0, len(zeilen))

	for _, zeile := range zeilen {
		var nachricht service.Handelsnachricht

		if fehler := json.Unmarshal(
			[]byte(zeile.Nachricht), &nachricht); fehler != nil {
			// **Der Satz für Menschen, der Grund fürs Protokoll**
			// ([0075](../../docs/entscheidungen/0075-der-satz-und-sein-grund.md)).
			return nil, domain.MitGrund(domain.ErrValidation, fehler,
				"Die Nachricht %q im Ausgang lässt sich nicht lesen.",
				zeile.NachrichtID)
		}

		offen = append(offen, service.Ausgangsnachricht{
			NachrichtID: zeile.NachrichtID, Vorgang: zeile.Vorgang,
			Empfaenger: empfaenger, Folge: nachricht.Umschlag.Folge,
			Nachricht: nachricht,
		})
	}

	return offen, nil
}

// Bestaetigen räumt eine Nachricht weg, die angekommen ist.
func (a ausgangsadapter) Bestaetigen(
	ctx context.Context, nachrichtID string,
) error {
	return a.ablage.Bestaetigen(ctx, nachrichtID)
}

// handelsleitung baut die Leitung für diesen Ruf.
//
// **Über [App.handelsablage] und damit ohne eigenen Griff auf `storage`** —
// dieselbe Ablage, die auch die Verbuchung liest. Zwei Wege zu denselben
// Revisionen wären zwei Stände.
func (a *App) handelsleitung() *service.Handelsleitung {
	return service.NeueHandelsleitung(
		a.handelsablage(), storage.NeueReservierungsablage(
			a.laufendeVerbindung())).MitHandelsschranke(einstellungsschranke{a: a})
}

// handelsverzeichnis baut Liste und Entwürfe für diesen Ruf (TRADE-01,
// #1715) — über dieselbe Ablage wie die Leitung.
func (a *App) handelsverzeichnis() *service.Handelsverzeichnis {
	db := a.laufendeVerbindung()

	return service.NeuesHandelsverzeichnis(a.handelsablage(), storage.NeueKontaktablage(db),
		service.NeueObjektmarken(storage.NeueObjektmarkenablage(db))).
		MitHandelsschranke(einstellungsschranke{a: a})
}

// zusagenwerk baut das Werk für diesen Ruf, das Exemplare festhält.
//
// **Je Ruf eines**, wie jede Naht dieser Schicht — die Bestandsverbindung
// soll die des Augenblicks sein.
func (a *App) zusagenwerk() *service.Zusagenwerk {
	return service.NeuesZusagenwerk(
		storage.NeueReservierungsablage(a.laufendeVerbindung()),
		a.handelsleitung()).MitHandelsschranke(einstellungsschranke{a: a})
}

// handelspost baut die Post samt ihrem Ausgang.
func (a *App) handelspost() *service.Handelspost {
	db := a.laufendeVerbindung()

	return service.NeueHandelspost(
		storage.NeueBriefkastenablage(db), a.handelsleitung(),
		storage.NeueKontaktablage(db),
	).MitHandelsschranke(einstellungsschranke{a: a}).MitAusgang(ausgangsadapter{ablage: storage.NeueAusgangsablage(db)}).
		// **Und mit den Gegenbelegen**
		// (#1650):
		// ein angekommener Beleg wird geprüft und abgelegt, und erst das
		// macht `beidseitig` wahr.
		MitGegenbelegen(storage.NeueAbschlussablage(db)).
		// **Und mit der Klärung** (TRADE-03, #1717): ein Konflikt in der
		// Übergabe führt in sie hinein.
		MitKlaerung(a.handelsleitung()).
		// **Und mit dem Ausweis** (Befund Z3 des Zweitblicks über Phase
		// 6.8): ein erster Vorgang entsteht nur, wo dieses Profil Partei ist.
		MitAusweis(a.peerwerk()).
		// **Und mit der Freigabe als Sichtgeber** (P11, #2379): eine
		// Sichtanfrage bekommt, was dieses Profil dem Fragenden zeigt.
		MitSichtgeber(a.freigabepflege())
}

// handelsversand baut den Absender für diesen Ruf.
func (a *App) handelsversand() *service.Handelsversand {
	db := a.laufendeVerbindung()

	// **Mit den Belegen**, damit ein Abschluss seine Urkunde mitnimmt
	// (#1650).
	// Ohne sie ginge die Nachricht `beleg` hinaus und sagte den Zustand —
	// und drüben bliebe `beidseitig` falsch, weil nichts sie deckt.
	return service.NeuerHandelsversand(a.handelspost(), a.handelsleitung(),
		storage.NeueFolgeablage(db), a.peerwerk()).
		MitBelegen(storage.NeueAbschlussablage(db))
}

// preisleser gibt dem Preisanhalt den Preisbestand (TRADE-02b, #1951).
func (a *App) preisleser() service.Preislese {
	return storage.NeuerPreisspeicher(a.laufendeVerbindung())
}

// handelsdruckleser liest die Drucke dieses Bestands — für die Karte einer
// Handelszeile (P10, #2378).
func (a *App) handelsdruckleser() service.Handelsdruckleser {
	return storage.NeueDrucke(a.laufendeVerbindung())
}

// einstellungswerte liest die Einstellungen dieses Profils — für Wege, die
// an jeder Fläche gelten und deshalb nicht an `a.einstellungen` hängen
// können: der Preisanhalt (`vorgabe.waehrung`, TRADE-02b) und die
// Schnittstelle der Suche (`netz.schnittstelle`, NET-02).
func (a *App) einstellungswerte(ctx context.Context) (map[string]string, error) {
	return storage.NeueEinstellungen(a.laufendeVerbindung()).Alle(ctx)
}

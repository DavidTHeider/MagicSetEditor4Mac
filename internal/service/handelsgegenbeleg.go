package service

// **Der Beleg der Gegenseite wird geprüft, bevor er einer ist** —
// #1650.
//
// # Die Reihenfolge ist die Aussage
//
// **Zuerst die Unterschrift, dann der Inhalt.** Was nicht echt ist, wird
// nicht gelesen — und eine Meldung über ein Feld eines gefälschten Belegs
// sagte etwas über den Fälscher aus. Dieselbe Reihenfolge wie bei
// [PruefeLeihe].
//
// # Was geprüft wird, und was nicht
//
// **Die Unterschrift** gegen die Peerkennung des Absenders — und die kommt
// aus der **Leitung**, nicht aus der Nutzlast: wer spricht, sagt die Sitzung.
// **Der Inhalt** gegen den eigenen Beleg: derselbe Vorgang, dieselbe
// Revision, dieselben Parteien, und die Zeilen **gespiegelt**.
//
// **Was nicht geprüft wird, ist die Zeit.** Ein Beleg trägt eine Uhr des
// Absenders, und sie ist ein Metadatum — sie steht nicht in der Unterschrift
// und schützt vor nichts (derselbe Satz wie am [domain.Handelsumschlag]).
//
// # Was mit einem Beleg geschieht, der nicht passt
//
// **Er wird festgehalten, und die Zustellung geht weiter** (#2500). Bis
// dahin endete sie mit einem Fehler und die Sitzung mit ihr: beim Absender
// stand weiter „hinausgegangen", beim Empfänger nichts. **Der Absender
// erfährt es auf seiner Seite**: seine Prüfung des hiesigen Belegs ist
// dieselbe, gespiegelt, und hält denselben Mangel fest — ohne eine neue Art
// auf dem Draht (§16: verträglich).

import (
	"context"
	"encoding/hex"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Gegenbelegstelle nimmt geprüfte Belege der Gegenseite entgegen.
//
// **Dieselben zwei Rufe wie die [Belegquelle] plus das Schreiben**: gelesen
// wird der eigene Beleg samt seinen Zeilen, geschrieben der fremde.
type Gegenbelegstelle interface {
	Belegquelle

	Gegenzeichnen(ctx context.Context, fremd domain.Fremdbeleg) error

	// Abweisen hält einen Gegenbeleg fest, der nicht passt (#2500).
	Abweisen(ctx context.Context, befund domain.Gegenbelegbefund) error
}

// MitGegenbelegen sagt der Post, wohin ein geprüfter Fremdbeleg gehört.
//
// **Wahlfrei, und ohne sie geschieht nichts.** Eine Post, die nur Stände
// zustellt, ist vollständig; ein Beleg ohne Ablage ginge verloren, und das
// ist besser als eine halbe Ablage.
func (p *Handelspost) MitGegenbelegen(stelle Gegenbelegstelle) *Handelspost {
	p.gegenbelege = stelle

	return p
}

// gegenbeleg prüft eine angekommene Urkunde und legt sie ab.
//
// **Eine Urkunde, die nicht passt, wird nicht gegengezeichnet**: sie
// behauptet, dieser Handel sei drüben anders abgeschlossen worden als hier.
// Wer sie durchließe, hätte einen Vorgang, den beide Seiten für erledigt
// halten — mit zwei verschiedenen Inhalten. **Sie wird festgehalten** und
// hält die Zustellung nicht mehr an (#2500).
//
// **Ein Fehler bleibt, was sich wiederholen lässt**: ohne eigenen Beleg ist
// hier noch nicht verbucht, und der nächste Anlauf prüft noch einmal
// (TRADE-03); ein Befund, der sich nicht festhalten lässt, ebenso.
func (p *Handelspost) gegenbeleg(
	ctx context.Context, sender string, nachricht Handelsnachricht,
	stand domain.Handelsangebot,
) error {
	if p.gegenbelege == nil || nachricht.Beleg == nil {
		return nil
	}

	urkunde := *nachricht.Beleg

	// **Zuerst die Unterschrift, dann der Inhalt** — festgehalten wird ein
	// Mangel aber erst neben dem eigenen Beleg: ohne ihn gäbe es nichts,
	// woran er hängt.
	unterschrift, unecht := unterschriftVon(sender, urkunde)

	eigener, fehler := p.eigenerBeleg(
		ctx, nachricht.Umschlag.Vorgang, stand.Parteien)
	if fehler != nil {
		return fehler
	}

	if unecht != "" {
		return p.abweisen(ctx, sender, eigener.Vorgang, urkunde,
			domain.MangelUnterschrift, unecht)
	}

	if abweichung := urkunde.Abweichung(eigener); abweichung != nil {
		return p.abweisen(ctx, sender, eigener.Vorgang, urkunde,
			abweichung.Mangel, abweichung.Satz)
	}

	// **Abgelegt wird die Unterschrift, die geprüft ist** — in kanonischer
	// Schreibung. Das Journal vergleicht sie als Zeichenkette, und dieselbe
	// Unterschrift in Großschreibung stünde sonst als anderer Beleg da
	// (Nebenbefund zu QA-FZ01, #2091).
	return p.gegenbelege.Gegenzeichnen(ctx, domain.Fremdbeleg{
		Vorgang: eigener.Vorgang, Von: sender,
		Abdruck:      urkunde.Beleg.Abdruck,
		Unterschrift: hex.EncodeToString(unterschrift),
	})
}

// eigenerBeleg baut den hiesigen Beleg, gegen den geprüft wird.
//
// **Er muss dastehen.** Ein Gegenbeleg ohne eigenen heißt: drüben ist
// abgeschlossen und hier nicht — dann gibt es nichts zu vergleichen, und eine
// Gegenzeichnung wäre eine Unterschrift unter etwas Ungeschriebenes.
func (p *Handelspost) eigenerBeleg(
	ctx context.Context, vorgang string, parteien []string,
) (domain.Handelsbeleg, error) {
	beleg, fehler := p.gegenbelege.Beleg(ctx, vorgang)
	if fehler != nil {
		return domain.Handelsbeleg{}, fehler
	}

	if beleg == nil {
		return domain.Handelsbeleg{}, domain.MitGrund(
			domain.ErrConflict, nil,
			"Zu diesem Handel ist hier noch nichts verbucht — ein Beleg "+
				"der Gegenseite lässt sich deshalb nicht prüfen.")
	}

	zeilen, fehler := p.gegenbelege.Journal(ctx, vorgang)
	if fehler != nil {
		return domain.Handelsbeleg{}, fehler
	}

	// **Die Parteien kommen von oben und nicht aus einem zweiten Lesen.**
	// [Handelspost.ausfuehren] hat den Stand schon in der Hand; ihn hier
	// noch einmal zu holen gäbe einen Fehlerzweig, den kein Lauf betreten
	// kann — scheiterte das Lesen, wäre die Zustellung längst gescheitert
	// (§10).
	return domain.Handelsbeleg{
		Vorgang: vorgang, Abdruck: beleg.Abdruck,
		Parteien: parteien, Zeilen: zeilen,
	}, nil
}

// unterschriftVon dekodiert und prüft die Unterschrift einer Urkunde — und
// gibt, wenn sie nicht trägt, den Satz dazu.
func unterschriftVon(
	sender string, urkunde domain.Belegurkunde,
) (unterschrift []byte, unecht string) {
	unterschrift, fehler := hex.DecodeString(urkunde.Unterschrift)
	if fehler != nil {
		return nil, "Die Unterschrift des Gegenbelegs ist kein Hex."
	}

	if falsch := peer.Pruefe(peer.Kennung(sender),
		urkunde.Beleg.Unterschriftsform(), unterschrift); falsch != nil {
		return nil, falsch.Error()
	}

	return unterschrift, ""
}

// abweisen hält den Befund fest und meldet ihn — **ohne Fehler**, damit die
// Zustellung weitergeht.
func (p *Handelspost) abweisen(
	ctx context.Context, sender, vorgang string, urkunde domain.Belegurkunde,
	mangel domain.Gegenbelegmangel, grund string,
) error {
	if fehler := p.gegenbelege.Abweisen(ctx, domain.Gegenbelegbefund{
		Vorgang: vorgang, Von: sender, Abdruck: urkunde.Beleg.Abdruck,
		Mangel: mangel, Grund: grund,
	}); fehler != nil {
		return fehler
	}

	// **Der Fingerabdruck und nicht die Kennung**, und **ohne den Satz** —
	// der nennt bei abweichenden Parteien deren Kennungen. Er steht am
	// Beleg; hier steht, dass es geschah (`app/protokollmerkmale.go`).
	slog.WarnContext(ctx, "ein Gegenbeleg passte nicht",
		"gegenueber", peer.Fingerabdruck(peer.Kennung(sender)),
		"ergebnis", vorgang, "art", string(mangel))

	return nil
}

package app

import "context"

// **Die Bindungen zu Scryfall** — abgetrennt am 2026-08-26, weil
// `bindungenkarten.go` über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Sie sind die einzigen Bindungen, die einen **fremden Dienst** anfassen
// (§16). Was sie von den anderen unterscheidet, ist nicht ihre Gestalt — auch
// sie reichen durch —, sondern woran sie hängen: an einer Gegenstelle, deren
// Bedingungen, Grenzen und Ausfälle in `DRITTANBIETER.md` stehen.

// ScryfallVorschlag zeigt, was eine Übernahme von Scryfall ändern würde
// (#1005).
//
// **Sie schreibt nichts** (§16, Regel 6).
func (b *Bindungen) ScryfallVorschlag(ctx context.Context, setID, karteID, name string) ScryfallvorschlagAntwort {
	return b.app(ctx).ScryfallVorschlag(b.ruf(ctx), setID, karteID, name)
}

// ScryfallUebernehmen schreibt die genannten Felder eines Vorschlags.
//
// **`stand` bindet ihn an die Kartenwerte, gegen die verglichen wurde**
// (#1241) —
// leer heißt: ohne Prüfung.
func (b *Bindungen) ScryfallUebernehmen(
	ctx context.Context,
	setID, karteID, name string, felder []string, stand string,
) SchreibAntwort {
	return b.app(ctx).ScryfallUebernehmen(b.ruf(ctx), setID, karteID, name, felder, stand)
}

// ScryfallBild holt das Bild einer Karte von Scryfall und bereitet den
// Zuschnitt vor
// (#1005).
//
// **Es schreibt nichts** — wie `BildAusDatei`. Erst [Bindungen.ScryfallBildEinsetzen]
// legt ab.
func (b *Bindungen) ScryfallBild(ctx context.Context, karteID, feld, name, art string) ScryfallBildAntwort {
	return b.app(ctx).ScryfallBild(b.ruf(ctx), karteID, feld, name, art)
}

// ScryfallBildEinsetzen legt das geholte Bild ab **und schreibt den Künstler**
// (#1005).
func (b *Bindungen) ScryfallBildEinsetzen(
	ctx context.Context,
	setID, karteID, feld, pfad, kuenstler, herkunft string, auswahl AuswahlDTO,
) SchreibAntwort {
	return b.app(ctx).ScryfallBildEinsetzen(
		b.ruf(ctx), setID, karteID, feld, pfad, kuenstler, herkunft, auswahl)
}

// ScryfallKarteAnlegen legt eine Karte aus einem Scryfall-Namen an
// (#1032).
//
// **Das Bild ist wahlweise dabei**: ohne `art` wird keines geholt. Mit `art`
// muss `bildfeld` dastehen — welches Feld ein Bild aufnimmt, weiß nur die
// Vorlage.
//
// **Das Spiel steht am Set** und wird nicht mitgegeben: eine Fläche, die es
// mitgeben müsste, könnte es falsch mitgeben.
func (b *Bindungen) ScryfallKarteAnlegen(
	ctx context.Context,
	setID, name, bildfeld, art string,
) ScryfallKarteAntwort {
	return b.app(ctx).ScryfallKarteAnlegen(b.ruf(ctx), setID, name, bildfeld, art)
}

// ScryfallSatzEinlesen liest einen ganzen Satz von Scryfall als Set ein
// (#1033).
//
// **Ohne Bilder**: ein Satz mit 300 Karten wären 300 Abrufe und 300 Nennungen;
// der Bildweg je Karte steht daneben.
func (b *Bindungen) ScryfallSatzEinlesen(ctx context.Context, code, spiel string) OeffnenAntwort {
	return b.app(ctx).ScryfallSatzEinlesen(b.ruf(ctx), code, spiel)
}

// ScryfallRegeln holt die Regelhinweise zu einer Karte
// (#888).
//
// **Es schreibt nichts** — eine Auskunft ist eine Auskunft.
func (b *Bindungen) ScryfallRegeln(ctx context.Context, name string) RegelauskunftAntwort {
	return b.app(ctx).ScryfallRegeln(b.ruf(ctx), name)
}

// ScryfallManakosten prüft einen Manakosten gegen die Symbologie von Scryfall
// (#888).
func (b *Bindungen) ScryfallManakosten(ctx context.Context, kosten string) ManakostenAntwort {
	return b.app(ctx).ScryfallManakosten(b.ruf(ctx), kosten)
}

// ScryfallSammeldaten sagt, ob die Sammeldatei bereitliegt
// (#888).
//
// **Sie liest nur.** Geholt wird über [Bindungen.ScryfallSammeldatenHolen].
func (b *Bindungen) ScryfallSammeldaten(ctx context.Context) SammeldatenAntwort {
	return b.app(ctx).ScryfallSammeldaten(b.ruf(ctx))
}

// ScryfallSammeldatenHolen lädt die Sammeldatei
// (#888).
func (b *Bindungen) ScryfallSammeldatenHolen(ctx context.Context) SammeldatenAntwort {
	return b.app(ctx).ScryfallSammeldatenHolen(b.ruf(ctx))
}

// ScryfallSuche sucht Karten bei Scryfall
// (#1167).
//
// **Sie schreibt nichts.** Wer eine gefundene Karte übernehmen will, geht über
// den Abgleich — und dort steht vorher, was sich ändern würde (§16, Regel 6).
func (b *Bindungen) ScryfallSuche(ctx context.Context, frage string) ScryfallsuchAntwort {
	return b.app(ctx).ScryfallSuche(b.ruf(ctx), frage)
}

// **Der Verweis auf einen echten Druck** — Werkstück 2 von Phase 3.6
// (#1167).
//
// Sie stehen hier und nicht bei den Kartenbindungen, weil sie an derselben
// Gegenstelle hängen: **gesetzt** wird ein Verweis erst, wenn Scryfall den
// Druck bestätigt hat.

// KarteVerweis nennt den Druck, auf den eine Karte verweist.
func (b *Bindungen) KarteVerweis(ctx context.Context, karteID string) VerweisAntwort {
	return b.app(ctx).KarteVerweis(b.ruf(ctx), karteID)
}

// KarteVerweisSetzen bezieht eine Karte auf einen Druck.
func (b *Bindungen) KarteVerweisSetzen(ctx context.Context, karteID, printing string) VerweisAntwort {
	return b.app(ctx).KarteVerweisSetzen(b.ruf(ctx), karteID, printing)
}

// KarteVerweisLoesen nimmt den Verweis einer Karte zurück.
func (b *Bindungen) KarteVerweisLoesen(ctx context.Context, karteID string) SchreibAntwort {
	return b.app(ctx).KarteVerweisLoesen(b.ruf(ctx), karteID)
}

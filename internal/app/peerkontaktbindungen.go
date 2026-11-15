package app

import "context"

// **Die Kontakte am Fenster** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Acht Bindungen und nicht eine
//
// **Weil sie sieben verschiedene Entscheidungen sind.** Aufnehmen ist ein
// Eintrag, Bestätigen ein Vergleich, Sperren ein Ausschluss, Entsperren eine
// neue Prüfung, Erlauben und Entziehen betreffen einen **einzelnen** Zweck,
// und die Adresse sagt, **wo** — nicht, **was**. Ein Ruf, der einen Kontakt
// „setzt", nähme jede dieser Entscheidungen als Feld entgegen — und eine
// vergessene Angabe wäre eine stille Freigabe.
//
// **Die achte kam mit `T3`** (#1468): eine **Notiz**, die nichts erlaubt
// und nichts zeigt — und genau deshalb nicht in einen Ruf gehört, der
// etwas erlaubt.

// Peerkontakte gibt die bekannten Gegenüber — **rein lesend**.
func (b *Bindungen) Peerkontakte(ctx context.Context) PeerkontakteAntwort {
	return b.app(ctx).Peerkontakte(b.ruf(ctx))
}

// PeerkontaktAufnehmen trägt ein Gegenüber ein, dessen Kennung jemand kennt.
//
// **Es erlaubt nichts** — was entsteht, ist unbestätigt.
func (b *Bindungen) PeerkontaktAufnehmen(
	ctx context.Context,
	kennung, name string,
) PeerkontakteAntwort {
	return b.app(ctx).PeerkontaktAufnehmen(b.ruf(ctx), kennung, name)
}

// PeerkontaktBestaetigen hält den verglichenen Abdruck gegen die Kennung.
//
// **Der Abdruck ist ein Argument und kein Häkchen**: wer bestätigt, sagt,
// **was** er verglichen hat.
func (b *Bindungen) PeerkontaktBestaetigen(
	ctx context.Context,
	kennung, verglichen string,
) PeerkontakteAntwort {
	return b.app(ctx).PeerkontaktBestaetigen(b.ruf(ctx), kennung, verglichen)
}

// PeerkontaktSperren schließt ein Gegenüber aus — samt allem, was es durfte.
func (b *Bindungen) PeerkontaktSperren(ctx context.Context, kennung string) PeerkontakteAntwort {
	return b.app(ctx).PeerkontaktSperren(b.ruf(ctx), kennung)
}

// PeerkontaktEntsperren nimmt die Sperre zurück — auf unbestätigt.
func (b *Bindungen) PeerkontaktEntsperren(ctx context.Context, kennung string) PeerkontakteAntwort {
	return b.app(ctx).PeerkontaktEntsperren(b.ruf(ctx), kennung)
}

// PeerzweckErlauben gibt einem bestätigten Gegenüber einen Zweck frei.
func (b *Bindungen) PeerzweckErlauben(
	ctx context.Context,
	kennung, zweck string,
) PeerkontakteAntwort {
	return b.app(ctx).PeerzweckErlauben(b.ruf(ctx), kennung, zweck)
}

// PeerzweckEntziehen nimmt eine Freigabe zurück.
func (b *Bindungen) PeerzweckEntziehen(
	ctx context.Context,
	kennung, zweck string,
) PeerkontakteAntwort {
	return b.app(ctx).PeerzweckEntziehen(b.ruf(ctx), kennung, zweck)
}

// PeerkontaktNotiz schreibt, was dieses Profil sich zu einem Gegenüber
// notiert — **lokal und nur lokal** (T3, #1468).
func (b *Bindungen) PeerkontaktNotiz(
	ctx context.Context,
	kennung, notiz string,
) PeerkontakteAntwort {
	return b.app(ctx).PeerkontaktNotiz(b.ruf(ctx), kennung, notiz)
}

// PeerkontaktAdresse trägt ein, wo ein Gegenüber erreichbar ist.
//
// **Eine leere löscht die eingetragene** — „ich kenne keine mehr" ist eine
// Aussage (N2, #1455).
func (b *Bindungen) PeerkontaktAdresse(
	ctx context.Context,
	kennung, adresse string,
) PeerkontakteAntwort {
	return b.app(ctx).PeerkontaktAdresse(b.ruf(ctx), kennung, adresse)
}

// KontakteinladungErstellen stellt eine Einladung dieses Profils aus
// (NET-01, #1705).
func (b *Bindungen) KontakteinladungErstellen(
	ctx context.Context,
	name, zweck string,
) KontakteinladungAntwort {
	return b.app(ctx).KontakteinladungErstellen(b.ruf(ctx), name, zweck)
}

// KontakteinladungPruefen sagt, was eine Einladung ist — rein lesend.
func (b *Bindungen) KontakteinladungPruefen(ctx context.Context, code string) EinladungsvorschauAntwort {
	return b.app(ctx).KontakteinladungPruefen(b.ruf(ctx), code)
}

// KontakteinladungAnnehmen nimmt eine Einladung an — ganz oder gar nicht.
func (b *Bindungen) KontakteinladungAnnehmen(
	ctx context.Context,
	code, eigenerName, verglichen string, freigeben bool, meinName string,
) PeerkontakteAntwort {
	return b.app(ctx).KontakteinladungAnnehmen(
		b.ruf(ctx), code, eigenerName, verglichen, freigeben, meinName)
}

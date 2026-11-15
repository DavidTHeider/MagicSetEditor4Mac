package app

// **Die Kontakte an der Naht** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum die Bestätigung den verglichenen Abdruck mitbringt
//
// **Weil sie sonst ein Klick wäre.** Das Issue verlangt: „Gegenüber per
// Fingerprint/Bestätigung prüfen." Was ein Mensch verglichen hat, ist die
// einzige Angabe, an der ein Irrtum sichtbar wird — und sie gehört deshalb in
// den Ruf und nicht in ein Häkchen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Peerkontakt ist ein bekanntes Gegenüber an der Naht.
type Peerkontakt struct {
	// Kennung ist der Nachweis — sie **ist** der öffentliche Schlüssel.
	Kennung string `json:"kennung"`

	// Fingerabdruck ist die Form, die zwei Menschen vergleichen.
	Fingerabdruck string `json:"fingerabdruck"`

	// SelbstgewaehlterName ist, wie das Gegenüber sich nennt.
	//
	// **Er belegt nichts**, und das Feld heißt so, damit keine Fläche ihn
	// für einen Nachweis hält.
	SelbstgewaehlterName string `json:"selbstgewaehlterName"`

	// EigenerName ist, wie dieses Profil ihn nennt.
	EigenerName string `json:"eigenerName"`

	// Stand ist `unbestaetigt`, `bestaetigt` oder `gesperrt`.
	Stand string `json:"stand"`

	// Gesehen ist der Zeitpunkt der letzten Begegnung, RFC 3339.
	Gesehen string `json:"gesehen"`

	// Rechte sind die freigegebenen Zwecke.
	Rechte []string `json:"rechte"`

	// Adresse ist, wo dieses Gegenüber erreichbar ist — leer heißt: nirgends
	// bekannt (N2, #1455).
	Adresse string `json:"adresse"`

	// Erstmals ist die **erste** Begegnung, RFC 3339 — leer heißt: nicht
	// bekannt (T3, #1468). Für einen Kontakt aus der Zeit davor weiß dieses
	// Programm sie nicht, und ein geratenes Datum sähe aus wie ein gemessenes.
	Erstmals string `json:"erstmals"`

	// Notiz ist, was dieses Profil sich notiert hat — **lokal und nur lokal**.
	Notiz string `json:"notiz"`

	// Lage ist einer der fünf Zustände aus NET-01.AK-03 — ohne `unbekannt`,
	// denn wer in der Liste steht, ist bekannt.
	Lage string `json:"lage"`

	// Vergleichscode ist, was beide Seiten laut vergleichen — **nur, solange
	// es etwas zu vergleichen gibt**: an einem unbestätigten Kontakt, und
	// nur, wenn dieses Profil eine Identität hat (NET-01, #1705).
	Vergleichscode string `json:"vergleichscode"`
}

// PeerkontakteAntwort ist die Liste samt ihrem Fehler.
type PeerkontakteAntwort struct {
	Kontakte []Peerkontakt `json:"kontakte"`

	// Zwecke sind die, die es überhaupt gibt.
	//
	// **Sie reisen mit, damit keine Fläche sie nachbaut.** Eine zweite
	// Aufzählung liefe beim fünften Zweck auseinander, und die erste, die es
	// merkt, ist keine (§10).
	Zwecke []string `json:"zwecke"`

	// Gemeldet sagt, ob sich dieses Profil nach einer Annahme beim
	// Einladenden melden konnte — nur dort gesetzt (NET-01b).
	Gemeldet *Einloesemeldung `json:"gemeldet,omitempty"`

	// Nachladefehler sagt, dass die Änderung **geschrieben** ist und nur die
	// Liste danach nicht zu lesen war (§7, Regel 3; #1960) — dann sind
	// `kontakte` leer und `fehler` ist es auch. Nur dort gesetzt.
	//
	// **Neben `fehler` und nicht an seiner Stelle**: `fehler` heißt „die
	// Änderung ist nicht geschehen", und wer das liest, versucht es noch
	// einmal — gegen einen Stand, der schon „angenommen" oder „bestätigt"
	// sagt.
	Nachladefehler *Fehler `json:"nachladefehler,omitempty"`

	Fehler *Fehler `json:"fehler"`
}

// Peerkontakte gibt die Kontakte dieses Profils — **rein lesend**.
func (a *App) Peerkontakte(ctx context.Context) PeerkontakteAntwort {
	ctx = a.imProfil(ctx)

	liste, fehler := a.kontaktwerk().Liste(ctx)
	if fehler != nil {
		return kontaktfehler(ctx, fehler)
	}

	eigene := a.eigeneKennung(ctx)

	kontakte := make([]Peerkontakt, 0, len(liste))
	for _, kontakt := range liste {
		dto := alsKontakt(kontakt)
		if eigene != "" && kontakt.Stand == domain.KontaktUnbestaetigt {
			dto.Vergleichscode = domain.Vergleichscode(eigene, kontakt.Kennung)
		}

		kontakte = append(kontakte, dto)
	}

	return PeerkontakteAntwort{Kontakte: kontakte, Zwecke: alleZwecke()}
}

// PeerkontaktAufnehmen trägt ein Gegenüber ein, dessen Kennung jemand kennt.
//
// # Warum es diesen Weg gibt, obwohl Kontakte aus Begegnungen entstehen
//
// **Weil eine Begegnung eine Verbindung voraussetzt und eine Verbindung einen
// Kontakt.** Der Austausch der Kennung von Hand ist der Anfang, und er ist
// kein Notbehelf, sondern genau das, wofür der Zweck `paarung` da ist.
//
// **Er erlaubt nichts.** Was hier entsteht, ist ein **unbestätigter** Kontakt:
// er darf sich vorstellen und sonst nichts, bis ein Mensch den Fingerabdruck
// verglichen und danach einen Zweck freigegeben hat.
func (a *App) PeerkontaktAufnehmen(
	ctx context.Context, kennung, name string,
) PeerkontakteAntwort {
	ctx = a.imProfil(ctx)

	return a.amKontakt(ctx, func(ctx context.Context) error {
		return a.kontaktwerk().Begegnet(ctx, peer.Kennung(kennung), name)
	})
}

// PeerkontaktBestaetigen hält den verglichenen Abdruck gegen die Kennung.
func (a *App) PeerkontaktBestaetigen(
	ctx context.Context, kennung, verglichen string,
) PeerkontakteAntwort {
	ctx = a.imProfil(ctx)

	return a.amKontakt(ctx, func(ctx context.Context) error {
		return a.kontaktwerk().MitAusweis(a.peerwerk()).Bestaetigen(
			ctx, peer.Kennung(kennung), verglichen)
	})
}

// PeerkontaktSperren schließt ein Gegenüber aus — samt allem, was es durfte.
func (a *App) PeerkontaktSperren(
	ctx context.Context, kennung string,
) PeerkontakteAntwort {
	ctx = a.imProfil(ctx)

	return a.amKontakt(ctx, func(ctx context.Context) error {
		return a.kontaktwerk().Sperren(ctx, peer.Kennung(kennung))
	})
}

// PeerkontaktEntsperren nimmt die Sperre zurück — **auf unbestätigt**.
func (a *App) PeerkontaktEntsperren(
	ctx context.Context, kennung string,
) PeerkontakteAntwort {
	ctx = a.imProfil(ctx)

	return a.amKontakt(ctx, func(ctx context.Context) error {
		return a.kontaktwerk().Entsperren(ctx, peer.Kennung(kennung))
	})
}

// PeerzweckErlauben gibt einem bestätigten Gegenüber einen Zweck frei.
func (a *App) PeerzweckErlauben(
	ctx context.Context, kennung, zweck string,
) PeerkontakteAntwort {
	ctx = a.imProfil(ctx)

	return a.amKontakt(ctx, func(ctx context.Context) error {
		return a.kontaktwerk().Erlauben(
			ctx, peer.Kennung(kennung), domain.Zweck(zweck))
	})
}

// PeerzweckEntziehen nimmt eine Freigabe zurück.
func (a *App) PeerzweckEntziehen(
	ctx context.Context, kennung, zweck string,
) PeerkontakteAntwort {
	ctx = a.imProfil(ctx)

	return a.amKontakt(ctx, func(ctx context.Context) error {
		return a.kontaktwerk().Entziehen(
			ctx, peer.Kennung(kennung), domain.Zweck(zweck))
	})
}

// PeerkontaktAdresse trägt ein, wo ein Gegenüber erreichbar ist.
//
// # Warum das keine Bestätigung voraussetzt
//
// **Weil eine Adresse nichts erlaubt.** Sie sagt, wohin gewählt wird, und
// nicht, was danach gilt; wer sie an die Bestätigung bände, könnte ein frisch
// eingetragenes Gegenüber nicht einmal anrufen, um seinen Abdruck zu
// vergleichen.
//
// **Und eine leere löscht die alte** — „ich kenne keine mehr" ist eine
// Aussage, und der Grund steht bei [storage.Kontaktablage.AdresseSetzen].
func (a *App) PeerkontaktAdresse(
	ctx context.Context, kennung, adresse string,
) PeerkontakteAntwort {
	ctx = a.imProfil(ctx)

	return a.amKontakt(ctx, func(ctx context.Context) error {
		return a.kontaktwerk().AdresseSetzen(
			ctx, peer.Kennung(kennung), adresse)
	})
}

// PeerkontaktNotiz schreibt, was dieses Profil sich zu einem Gegenüber
// notiert (T3, #1468).
//
// # Warum eine eigene Bindung und kein Feld an einer anderen
//
// **Weil sie nichts erlaubt und nichts zeigt.** Eine Notiz ist Freitext für
// den, der sie schreibt; sie geht über keine Leitung. Ein Ruf, der sie
// zusammen mit dem Stand oder den Rechten setzte, machte aus einer harmlosen
// Änderung eine, die man prüfen müsste — und eine vergessene Angabe darin
// wäre eine stille Freigabe.
//
// **Eine leere löscht die vorhandene**, dieselbe Form wie bei der Adresse.
func (a *App) PeerkontaktNotiz(
	ctx context.Context, kennung, notiz string,
) PeerkontakteAntwort {
	ctx = a.imProfil(ctx)

	return a.amKontakt(ctx, func(ctx context.Context) error {
		return a.kontaktwerk().NotizSetzen(ctx, peer.Kennung(kennung), notiz)
	})
}

// amKontakt führt eine Änderung und gibt danach die ganze Liste.
//
// **Die Liste danach und nicht der geänderte Kontakt.** Eine Sperre räumt
// Rechte, eine Rotation räumt alles — wer nur die eine Zeile zurückgäbe, ließe
// eine Fläche mit einem Stand stehen, den es nicht mehr gibt (§7, Regel 3).
//
// **Ohne [App.imProfil]**: das tut jeder Ruf selbst, in seiner ersten Zeile. Ein Wächter zählt genau dort nach — und eine Bindung, die die
// Bindung an ein Profil einer Hilfsfunktion überließe, sähe für ihn aus wie
// eine, die sie vergisst.
//
// # Ein Nachladefehler ist kein gescheiterter Schritt
//
// **Geschrieben ist geschrieben** (§7, Regel 3;
// #1960).
// Scheitert das Lesen der Liste **nach** der Änderung, steht sein Fehler in
// [PeerkontakteAntwort.Nachladefehler] und nicht in `fehler`: wer `fehler`
// liest, hält die Änderung für ungeschehen und versucht sie noch einmal —
// gegen einen Stand, der dann „schon angenommen" sagt. Dieselbe Klasse hat
// der Zweitblick über Phase 6.7 (Z9) im Dienst behoben; hier gilt sie für
// alle neun Wege, die durch diese Funktion gehen.
func (a *App) amKontakt(
	ctx context.Context, aendern func(context.Context) error,
) PeerkontakteAntwort {
	if fehler := aendern(ctx); fehler != nil {
		return kontaktfehler(ctx, fehler)
	}

	antwort := a.Peerkontakte(ctx)
	antwort.Nachladefehler, antwort.Fehler = antwort.Fehler, nil

	return antwort
}

// alsKontakt macht aus der Entität ihr DTO.
func alsKontakt(kontakt domain.Peerkontakt) Peerkontakt {
	rechte := make([]string, 0, len(kontakt.Rechte))
	for _, zweck := range kontakt.Rechte {
		rechte = append(rechte, string(zweck))
	}

	return Peerkontakt{
		Kennung:              kontakt.Kennung,
		Fingerabdruck:        kontakt.Fingerabdruck,
		SelbstgewaehlterName: kontakt.SelbstgewaehlterName,
		EigenerName:          kontakt.EigenerName,
		Stand:                string(kontakt.Stand),
		Gesehen:              alsZeitpunkt(kontakt.Gesehen),
		Rechte:               rechte,
		Adresse:              kontakt.Adresse,
		Erstmals:             alsZeitpunkt(kontakt.Erstmals),
		Notiz:                kontakt.Notiz,
		Lage:                 string(kontakt.Lage()),
	}
}

// eigeneKennung gibt die Kennung dieses Profils — leer, wenn es keine
// bereite Identität hat.
//
// **Leer und kein Fehler**: die Liste ist auch ohne Identität eine Liste, nur
// ohne Vergleichscode — dann bleibt der Abdruck.
func (a *App) eigeneKennung(ctx context.Context) string {
	lage, fehler := a.peerwerk().Lage(ctx)
	if fehler != nil || lage.Stand != domain.PeerstandBereit {
		return ""
	}

	return string(lage.Kennung)
}

// alleZwecke gibt die Aufzählung aus `domain` als Text.
func alleZwecke() []string {
	zwecke := make([]string, 0, len(domain.DieZwecke))
	for _, zweck := range domain.DieZwecke {
		zwecke = append(zwecke, string(zweck))
	}

	return zwecke
}

// kontaktfehler ist die Antwort, wenn ein Weg scheitert.
//
// **Mit leeren Listen und nicht mit `null`** — der Wächter
// `leerelistenlaufzeit_test.go` hält das für jede Bindung.
func kontaktfehler(ctx context.Context, fehler error) PeerkontakteAntwort {
	return PeerkontakteAntwort{
		Kontakte: []Peerkontakt{}, Zwecke: alleZwecke(),
		Fehler: Uebersetze(ctx, fehler),
	}
}

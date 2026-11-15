package service

// **Wen dieses Profil kennt und was es ihm erlaubt** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// # Warum die Bestätigung hier liegt und nicht in der Naht
//
// **Weil sie eine fachliche Aussage ist.** „Ich habe den Abdruck verglichen"
// ist eine Behauptung eines Menschen über die Welt; wer sie in einer Naht
// entgegennähme, hätte die Regel dort, wo sie niemand sucht (§2) — und die
// Fläche entschiede, wie streng sie ist.
//
// # Was dieses Werk nicht tut
//
// **Es verbindet nichts und prüft keine Nachricht.** Ob ein Handschlag
// stimmt, entscheidet `internal/peer`; ob sein Gegenüber **willkommen** ist,
// entscheidet dieses Werk. Die Trennung steht auch am anderen Ende — in
// [peer.Auspacke] —, und sie ist der Grund, warum eine Rechteliste nicht in
// einem Blatt liegt.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Kontaktvermerke ist, was das Werk vom Bestand braucht.
type Kontaktvermerke interface {
	Liste(ctx context.Context) ([]domain.Peerkontakt, error)
	Lesen(ctx context.Context, kennung string) (domain.Peerkontakt, error)
	Begegnet(
		ctx context.Context, kennung, fingerabdruck, name string,
		jetzt time.Time,
	) error
	Stand(
		ctx context.Context, kennung string, stand domain.Kontaktstand,
	) error
	Erlauben(
		ctx context.Context, kennung string, zweck domain.Zweck,
		jetzt time.Time,
	) error
	Entziehen(
		ctx context.Context, kennung string, zweck domain.Zweck,
	) error
	AdresseSetzen(ctx context.Context, kennung, adresse string) error
	NotizSetzen(ctx context.Context, kennung, notiz string) error
	AllesZurueck(ctx context.Context) error
}

// Kontaktwerk führt die Kontakte eines Profils.
type Kontaktwerk struct {
	vermerke Kontaktvermerke
	jetzt    func() time.Time
	ausweis  Ausweisstelle
}

// NeuesKontaktwerk baut es über einem Bestand und einer Uhr.
func NeuesKontaktwerk(
	vermerke Kontaktvermerke, jetzt func() time.Time,
) *Kontaktwerk {
	return &Kontaktwerk{vermerke: vermerke, jetzt: jetzt}
}

// MitAusweis lässt es den Vergleichscode kennen (NET-01, #1705).
//
// **Ohne Ausweis gibt es kein Paar** und damit keinen Code — dann bestätigt
// nur der Abdruck, wie vor NET-01.
func (w *Kontaktwerk) MitAusweis(ausweis Ausweisstelle) *Kontaktwerk {
	w.ausweis = ausweis

	return w
}

// Liste gibt die Kontakte dieses Profils.
func (w *Kontaktwerk) Liste(
	ctx context.Context,
) ([]domain.Peerkontakt, error) {
	return w.vermerke.Liste(ctx)
}

// Begegnet vermerkt, dass jemand sich gemeldet hat.
//
// **Der Abdruck wird gerechnet und nicht entgegengenommen.** Er entsteht aus
// der Kennung ([peer.Fingerabdruck]); käme er aus der Nachricht, könnte ein
// Gegenüber den Abdruck eines anderen behaupten — und ein Mensch vergliche ihn
// gegen genau das, was der Fälscher geschickt hat.
func (w *Kontaktwerk) Begegnet(
	ctx context.Context, kennung peer.Kennung, name string,
) error {
	abdruck := peer.Fingerabdruck(kennung)

	// **Eine Kennung, aus der kein Schlüssel wird, ist keine.**
	//
	// Ohne diese Zeile stünde sie mit **leerem** Abdruck in der Liste — und
	// ein leerer Abdruck ist genau das, was [Kontaktwerk.Bestaetigen]
	// später abweist. Ein Eintrag, den niemand je bestätigen kann, sieht
	// aus wie ein Kontakt und ist keiner.
	if abdruck == "" {
		return fmt.Errorf(
			"%w: %q ist keine Peerkennung", domain.ErrValidation, kennung)
	}

	return w.vermerke.Begegnet(ctx, string(kennung), abdruck, name, w.jetzt())
}

// Bestaetigen hält den verglichenen Abdruck gegen die Kennung.
//
// # Warum der Abdruck ein Argument ist
//
// **Weil die Bestätigung sonst ein Klick wäre.** Wer bestätigt, hat etwas
// verglichen — und **was** er verglichen hat, ist die einzige Angabe, die ein
// Irrtum sichtbar macht. Ein `Bestaetigen(kennung)` ohne sie nähme jedes „ja"
// entgegen, auch das aus dem falschen Fenster.
func (w *Kontaktwerk) Bestaetigen(
	ctx context.Context, kennung peer.Kennung, verglichen string,
) error {
	passt, fehler := w.verglichen(ctx, kennung, verglichen)
	if fehler != nil {
		return fehler
	}

	if !passt {
		return fmt.Errorf(
			"%w: der verglichene Abdruck gehört nicht zu diesem Gegenüber",
			domain.ErrValidation)
	}

	return w.vermerke.Stand(
		ctx, string(kennung), domain.KontaktBestaetigt)
}

// verglichen sagt, ob das Verglichene der Abdruck **oder** der Code ist.
//
// **Der Abdruck zuerst**, weil er keinen Ausweis braucht: wer ihn verglichen
// hat, soll nicht an einem Schlüsselbund scheitern, den er nicht braucht.
func (w *Kontaktwerk) verglichen(
	ctx context.Context, kennung peer.Kennung, verglichen string,
) (bool, error) {
	abdruck := peer.Fingerabdruck(kennung)
	if abdruck != "" && abdruck == verglichen {
		return true, nil
	}

	if abdruck == "" || w.ausweis == nil {
		return false, nil
	}

	ich, fehler := w.ausweis.Ausweis(ctx)
	if fehler != nil {
		return false, fehler
	}

	return domain.VergleichscodePasst(
		string(ich.Kennung()), string(kennung), verglichen), nil
}

// Sperren schließt ein Gegenüber aus — samt allem, was es durfte.
func (w *Kontaktwerk) Sperren(
	ctx context.Context, kennung peer.Kennung,
) error {
	return w.vermerke.Stand(ctx, string(kennung), domain.KontaktGesperrt)
}

// Entsperren nimmt die Sperre zurück — **auf unbestätigt**.
//
// **Nicht auf den Stand von vorher.** Wer gesperrt war, wird neu geprüft; ein
// Gegenüber, das nach dem Entsperren wieder alles dürfte, machte die Sperre
// zu einer Pause.
func (w *Kontaktwerk) Entsperren(
	ctx context.Context, kennung peer.Kennung,
) error {
	return w.vermerke.Stand(ctx, string(kennung), domain.KontaktUnbestaetigt)
}

// Erlauben gibt einem bestätigten Gegenüber einen Zweck frei.
func (w *Kontaktwerk) Erlauben(
	ctx context.Context, kennung peer.Kennung, zweck domain.Zweck,
) error {
	return w.vermerke.Erlauben(ctx, string(kennung), zweck, w.jetzt())
}

// Entziehen nimmt eine Freigabe zurück.
func (w *Kontaktwerk) Entziehen(
	ctx context.Context, kennung peer.Kennung, zweck domain.Zweck,
) error {
	return w.vermerke.Entziehen(ctx, string(kennung), zweck)
}

// AdresseSetzen trägt ein, wo ein Gegenüber erreichbar ist.
//
// # Warum das keine Bestätigung voraussetzt
//
// **Weil eine Adresse nichts erlaubt.** Sie sagt, wohin gewählt wird, und
// nicht, was danach gilt — das entscheidet [Peerkontakt.Darf] nach dem
// Handschlag. Wer sie an die Bestätigung bände, könnte einen frisch
// eingetragenen Kontakt nicht einmal anrufen, um seinen Abdruck zu
// vergleichen.
//
// **Und warum sie hier trotzdem durch die Fachschicht geht**: die Form ist
// eine Regel aus `domain`, und der Bestand prüft sie ein zweites Mal. Zwei
// Wächter für eine Sache, und beide sind gewollt — die Naht gibt eine
// Zeichenkette weiter, die ein Mensch getippt hat.
func (w *Kontaktwerk) AdresseSetzen(
	ctx context.Context, kennung peer.Kennung, adresse string,
) error {
	if fehler := domain.PruefeAdresse(adresse); fehler != nil {
		return fehler
	}

	return w.vermerke.AdresseSetzen(ctx, string(kennung), adresse)
}

// NotizSetzen schreibt, was dieses Profil sich zu einem Gegenüber notiert
// (T3, #1468).
//
// **Ohne Prüfung, und das ist der Punkt**: eine Notiz ist Freitext für den,
// der sie schreibt. Sie geht über keine Leitung — eine Sichtkarte trägt sie
// nicht —, und deshalb gibt es nichts, wogegen sie zu halten wäre.
func (w *Kontaktwerk) NotizSetzen(
	ctx context.Context, kennung peer.Kennung, notiz string,
) error {
	return w.vermerke.NotizSetzen(ctx, string(kennung), notiz)
}

// Lesen gibt einen Kontakt dieses Profils.
//
// **Ein `ErrNotFound` heißt hier „kenne ich nicht"**, und der Aufrufer
// entscheidet, was das bedeutet — anders als bei [Kontaktwerk.Darf], das für
// einen Unbekannten ein klares Nein gibt. Der Unterschied ist der Zweck: dort
// wird eine Erlaubnis gefragt, hier ein Stand gelesen.
func (w *Kontaktwerk) Lesen(
	ctx context.Context, kennung peer.Kennung,
) (domain.Peerkontakt, error) {
	return w.vermerke.Lesen(ctx, string(kennung))
}

// Adresse gibt, wo ein Gegenüber erreichbar ist.
//
// # Warum ein fehlender Eintrag hier ein Fehler ist und bei [Kontaktwerk.Darf]
// nicht
//
// **Weil die Frage eine andere ist.** „Darf er handeln?" hat für einen
// Unbekannten ein klares Nein; „wo erreiche ich ihn?" hat keine Antwort, und
// eine leere Zeichenkette wäre eine erfundene. Wer wählt, muss den Unterschied
// sehen — sonst wählte er nach `""` und bekäme einen Fehler, der von der
// Leitung kommt statt von der Sache.
func (w *Kontaktwerk) Adresse(
	ctx context.Context, kennung peer.Kennung,
) (string, error) {
	kontakt, fehler := w.vermerke.Lesen(ctx, string(kennung))
	if fehler != nil {
		return "", fehler
	}

	// **Mit eigenem Ausgang** (#2491): der Rat ist ein Feld am Kontakt und
	// nicht „lade die Liste neu".
	if kontakt.Adresse == "" {
		return "", fmt.Errorf(
			"%w (%s) — ohne sie findet dieses Programm es nur im selben Netz",
			domain.ErrPeerOhneAdresse, peer.Fingerabdruck(kennung))
	}

	return kontakt.Adresse, nil
}

// Darf beantwortet die Frage, um die es am Ende geht.
//
// # Warum ein Unbekannter kein Fehler ist
//
// **Weil die Frage beantwortet ist.** „Darf dieses Gegenüber handeln?" hat für
// einen Unbekannten ein klares Nein; ein `ErrNotFound` zwänge jeden Aufrufer,
// ihn wegzufangen — bis einer es vergisst und eine Verbindung mit einer
// Meldung endet statt mit einem Nein.
func (w *Kontaktwerk) Darf(
	ctx context.Context, kennung peer.Kennung, zweck domain.Zweck,
) (bool, error) {
	kontakt, fehler := w.vermerke.Lesen(ctx, string(kennung))

	switch {
	case errors.Is(fehler, domain.ErrNotFound):
		return domain.Peerkontakt{
			Kennung: string(kennung), Stand: domain.KontaktUnbestaetigt,
		}.Darf(zweck), nil
	case fehler != nil:
		return false, fehler
	}

	return kontakt.Darf(zweck), nil
}

// NachRotation verlangt von jedem Kontakt eine neue Prüfung.
//
// **Die Abnahme des Issues verlangt es wörtlich**: „Rotation erzwingt
// Neuprüfung bekannter Kontakte." Wer rotiert, hat für jeden Kontakt einen
// anderen Fingerabdruck als den, den er einmal vorgelesen hat — jede
// Bestätigung von vorher gilt einer Identität, die es nicht mehr gibt.
func (w *Kontaktwerk) NachRotation(ctx context.Context) error {
	return w.vermerke.AllesZurueck(ctx)
}

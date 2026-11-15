package service

// **Eine Einladung genügt** — NET-01b
// (#1954),
// entschieden in
// [0156](../../docs/entscheidungen/0156-eine-einladung-genuegt.md).
//
// # Der Ablauf, und wer was weiß
//
//  1. Anna stellt eine Einladung mit Adresse aus; sie ist **einlösbar**, und
//     Anna merkt sich die Nonce.
//  2. Bert nimmt sie an — wie bisher, ganz oder gar nicht — und kennt damit
//     Annas Schlüssel.
//  3. Bert **meldet sich** ([Einladungswerk.Melden]): eine Paarungssitzung
//     zu der Adresse aus der Einladung, festgenagelt auf Annas Kennung, und
//     darüber die Nonce.
//  4. Anna **beantwortet** ([Einladungswerk.EinloesungBeantworten]): die
//     Kennung, die der Handschlag geprüft hat, wird ein **unbestätigter**
//     Kontakt. Bestätigt wird weiterhin nur über den Vergleichscode (AK-01).
//
// **Scheitert Schritt 3, bleibt Schritt 2** (§7, Regel 3): Bert hat Anna
// aufgenommen, und das bleibt so. Was fehlt, ist Annas Seite — und dafür
// gibt es den Weg mit zwei Einladungen, den der Satz nennt.

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Einloesefrist ist, wie lange eine Seite höchstens auf die andere wartet.
//
// **Wählen, Handschlag und Antwort zusammen.** Über ein fremdes Netz kostet
// schon der Aufbau Sekunden; mehr als fünfzehn heißt: da ist niemand.
const Einloesefrist = 15 * time.Second

// Einloeseleitung ist, was der Annehmende zum Melden braucht.
type Einloeseleitung interface {
	Sende(ctx context.Context, nachricht []byte, art Zustellung) error
	Leere(ctx context.Context) error
	Empfange(ctx context.Context) ([]byte, error)
	Schliesse() error
}

// Einloesebote öffnet eine Paarungssitzung zum Einladenden.
type Einloesebote func(ctx context.Context, kennung peer.Kennung) (Einloeseleitung, error)

// Einloesezugang ist, was der Einladende zum Beantworten braucht — eine
// Sitzung, die schon steht und deren Gegenüber der Handschlag geprüft hat.
type Einloesezugang interface {
	Stand() Sitzungsstand
	Sende(ctx context.Context, nachricht []byte, art Zustellung) error
	Leere(ctx context.Context) error
	Empfange(ctx context.Context) ([]byte, error)
}

// MitGrenze setzt die Ratengrenze — **eine je Anwendung**, nicht je Werk:
// ein Werk entsteht je Ruf, und eine Grenze, die mit ihm stürbe, zählte
// nichts.
func (w *Einladungswerk) MitGrenze(grenze *Einloesegrenze) *Einladungswerk {
	w.grenze = grenze

	return w
}

// MitRiegel setzt den Prüfer, den eine Meldung fragt — SEC-P4
// (#2000).
func (w *Einladungswerk) MitRiegel(riegel Pruefer) *Einladungswerk {
	w.riegel = riegel

	return w
}

// darfMelden fragt, ob sich eine Annahme hinaus melden darf.
//
// **Eine einlösbare Einladung wählt die Adresse, die in ihr steht** — und
// die schreibt der Einladende selbst, denn unterschreiben kann jeder, der
// einen Schlüssel hat. Die Meldung schickt dorthin den eigenen Namen und die
// eigene Adresse: das ist `externSenden`, und ein Token reicht nie dorthin
// ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)). Bis #2000
// fragte die Annahme nur `lokalSchreiben`, und ein Token `schreiben` löste
// damit eine Verbindung zu beliebigem `host:port` aus.
//
// **Eine Einladung ohne Ort fragt nicht**: sie wählt niemanden an.
//
// **Keine Bereichssperre auf die Adresse** — das eigene Netz ist hier der
// Zweck, und die Verbindung ist auf die Kennung des Einladenden
// festgenagelt (`internal/leitung`).
func (w *Einladungswerk) darfMelden(ctx context.Context, einladung domain.Kontakteinladung) error {
	if w.riegel == nil || !einladung.Einloesbar {
		return nil
	}

	return w.riegel.Pruefe(ctx, domain.WirkungExternSenden)
}

// Einloesen nimmt auf, wer eine Einladung dieses Profils einlöst.
//
// **`gegenueber` kommt aus dem Handschlag** und nicht aus der Nachricht: wer
// spricht, hat dort bewiesen, dass er den Schlüssel hält. Der Name ist eine
// Behauptung und wird nur übernommen, wenn ein Mensch ihn lesen kann.
func (w *Einladungswerk) Einloesen(
	ctx context.Context, gegenueber peer.Kennung, einloesung domain.Einloesung,
) error {
	if w.grenze == nil {
		return fmt.Errorf("%w: ohne Ratengrenze wird nichts eingelöst",
			domain.ErrEinloesungGescheitert)
	}

	if fehler := w.grenze.Versuch(string(gegenueber), einloesung.Nonce); fehler != nil {
		return fehler
	}

	return w.ablage.Einloesen(ctx, einloesung.Nonce, domain.Peerkontakt{
		Kennung:              string(gegenueber),
		Fingerabdruck:        peer.Fingerabdruck(gegenueber),
		SelbstgewaehlterName: domain.BrauchbarerName(einloesung.Name),
		Adresse:              einloesung.Adresse,
		Stand:                domain.KontaktUnbestaetigt,
	}, w.jetzt())
}

// EinloesungBeantworten liest eine Einlösung, entscheidet und antwortet.
//
// **Jedes Nein bekommt eine Antwort** (AK-03): wer drüben wartet, soll
// „verbraucht" lesen und nicht in eine Frist laufen. Was keinen Namen hat,
// heißt „gescheitert"; der Grund bleibt hier und wird zurückgegeben.
func (w *Einladungswerk) EinloesungBeantworten(
	ctx context.Context, zugang Einloesezugang,
) error {
	ctx, abbrechen := context.WithTimeout(ctx, Einloesefrist)
	defer abbrechen()

	roh, fehler := zugang.Empfange(ctx)
	if fehler != nil {
		return fehler
	}

	befund := w.einloesenAus(ctx, zugang.Stand(), roh)

	if fehler := zugang.Sende(ctx, domain.AntwortAuf(befund).Kodiere(), Zuverlaessig); fehler != nil {
		return fehler
	}

	if fehler := zugang.Leere(ctx); fehler != nil {
		return fehler
	}

	return befund
}

func (w *Einladungswerk) einloesenAus(
	ctx context.Context, stand Sitzungsstand, roh []byte,
) error {
	einloesung, fehler := domain.ZerlegeEinloesung(roh)
	if fehler != nil {
		return fehler
	}

	return w.Einloesen(ctx, stand.Gegenueber, einloesung)
}

// Melden meldet dieses Profil beim Einladenden — nach der Annahme.
//
// **`ich` trägt Adresse und Namen**; die Nonce setzt Melden aus der
// Einladung.
//
// **Eine Einladung, die nicht einlösbar ist, wählt niemanden an**: sie
// stammt aus einer Fassung, die das Einlösen nicht kennt, oder nennt keine
// Adresse. Der Fehler sagt dann, was bleibt — die eigene Einladung zurück.
func (w *Einladungswerk) Melden(
	ctx context.Context, bote Einloesebote, code string, ich domain.Einloesung,
) error {
	vorschau, fehler := w.Pruefen(ctx, code)
	if fehler != nil {
		return fehler
	}

	einladung := vorschau.Einladung
	if !einladung.Einloesbar {
		return domain.ErrEinladungNichtEinloesbar
	}

	// **Auch hier und nicht nur bei der Annahme** — der Riegel steht an der
	// Netzgrenze, und wer `Melden` ruft, ohne angenommen zu haben, käme
	// sonst an ihm vorbei.
	if verwehrt := w.darfMelden(ctx, einladung); verwehrt != nil {
		return verwehrt
	}

	ich.Nonce = einladung.Nonce

	nachricht, fehler := ich.Kodiere()
	if fehler != nil {
		return fehler
	}

	ctx, abbrechen := context.WithTimeout(ctx, Einloesefrist)
	defer abbrechen()

	leitung, fehler := bote(ctx, peer.Kennung(einladung.Kennung))
	if fehler != nil {
		return unbeantwortet(fehler)
	}

	defer func() { _ = leitung.Schliesse() }() //nolint:errcheck // der Weg endet hier ohnehin

	roh, fehler := hinUndZurueck(ctx, leitung, nachricht)
	if fehler != nil {
		return unbeantwortet(fehler)
	}

	antwort, fehler := domain.ZerlegeEinloeseantwort(roh)
	if fehler != nil {
		return fehler
	}

	return antwort.Fehler()
}

func hinUndZurueck(ctx context.Context, leitung Einloeseleitung, nachricht []byte) ([]byte, error) {
	if fehler := leitung.Sende(ctx, nachricht, Zuverlaessig); fehler != nil {
		return nil, fehler
	}

	if fehler := leitung.Leere(ctx); fehler != nil {
		return nil, fehler
	}

	return leitung.Empfange(ctx)
}

// unbeantwortet behält den Grund fürs Protokoll und zeigt den Rückfall.
func unbeantwortet(grund error) error {
	return domain.MitGrund(domain.ErrEinloesungUnbeantwortet, grund,
		"%s", domain.ErrEinloesungUnbeantwortet.Error())
}

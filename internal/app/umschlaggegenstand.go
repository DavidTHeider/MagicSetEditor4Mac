package app

// **Was ein Umschlag außer einem Set tragen kann** — Werkstück `T7` von
// Phase 5.6
// (#1472).
//
// # Warum kein zweites Format entsteht
//
// **Weil es eines gibt.** `N4`
// (#1457)
// kennt fünf Arten — Karte, Set, Vorlage, Deck und Sammlungssicht —, und bis
// heute bediente der Umschlag genau eine davon. Das Issue verlangt
// ausdrücklich „kein zweites `.msetrade`-Format, kein zweiter Serializer";
// deshalb wird hier **nichts** neu serialisiert: jede Art geht in der Gestalt
// hinaus, in der dieses Programm sie ohnehin über seine Flächen ausgibt.
//
// # Und warum eine Sammlungssicht nicht „übernommen" wird
//
// **Weil sie kein Besitz ist.** Sie sagt, was ein Gegenüber **hat** — und ein
// Empfang ist keine Abbuchung und keine Buchung. Sie wird angesehen und geht
// in den Abgleich (`T4`); in den eigenen Bestand geht sie nie. Das Issue sagt
// es in einem Satz: „Empfang/Import ist keine physische Abbuchung."

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Ordner, unter denen die Arten im Umschlag liegen.
//
// **Je Art einer.** Ein gemeinsamer Ordner ließe beim Lesen offen, was eine
// Datei ist; und ein Umschlag mit zwei Decks braucht zwei Namen.
const (
	// Vorlagenordner trägt die Dateien eines Vorlagenpakets.
	Vorlagenordner = "vorlage"

	// Kartendatei ist die eine Karte.
	Kartendatei = "karte.json"

	// Deckdatei ist das eine Deck samt seinen Plätzen.
	Deckdatei = "deck.json"

	// Sichtdatei ist die freigegebene Sammlungssicht.
	Sichtdatei = "sicht.json"
)

// Umschlagarten sind die Arten, die geschnürt werden können.
//
// **Eine Aufzählung und kein freier Text**: was hier steht, muss die
// empfangende Seite einordnen können (§16), und die Liste ist dieselbe wie in
// `internal/austausch`.
func Umschlagarten() []string {
	return []string{
		string(austausch.ArtSet), string(austausch.ArtKarte),
		string(austausch.ArtVorlage), string(austausch.ArtDeck),
		string(austausch.ArtSammlung),
	}
}

// umschlagart deutet die Angabe — **leer heißt Set**.
//
// **Der Rückfall ist keine Bequemlichkeit, sondern die Verträglichkeit**
// (0033, Regel 6): bis heute nahm jede Fläche eine Setkennung und sonst
// nichts. Wer weiter nur eine schickt, bekommt weiter ein Set.
func umschlagart(roh string) (austausch.Objektart, error) {
	if roh == "" {
		return austausch.ArtSet, nil
	}

	for _, art := range Umschlagarten() {
		if roh == art {
			return austausch.Objektart(roh), nil
		}
	}

	return "", fmt.Errorf("%w: %q ist keine Art, die ein Umschlag trägt — "+
		"es gibt %v", domain.ErrValidation, roh, Umschlagarten())
}

// inhalteFuer packt den Gegenstand, der gemeint ist.
func (a *App) inhalteFuer(
	ctx context.Context, art austausch.Objektart, kennung string,
) ([]austausch.Inhalt, error) {
	switch art {
	case austausch.ArtKarte:
		karte := a.Karte(ctx, kennung)
		if karte.Fehler != nil {
			return nil, karte.Fehler.AlsFehler()
		}

		return alsEineDatei(art, Kartendatei, karte.Karte), nil

	case austausch.ArtDeck:
		// **Beides in einer Frage.** Ein Deck, das es nicht gibt, kommt als
		// Fehler; ein `nil` ohne Fehler kann diese Naht nicht hergeben, und
		// ein eigener Zweig dafür wäre eine Zusage, die kein Lauf betritt
		// (§10).
		deck := a.Deck(ctx, kennung)
		if deck.Fehler != nil || deck.Deck == nil {
			return nil, fmt.Errorf("%w: das Deck %s gibt es nicht",
				domain.ErrNotFound, kennung)
		}

		// **Was privat ist, bleibt hier** (FMT-05.AK-04, #1694): die Notiz
		// ist eine Aufzeichnung für sich, und die Herkunft nennt Quellen und
		// Profilnamen anderer. Ein Umschlag geht an andere — beides reiste bis
		// dahin ungefragt mit. Kennungen des eigenen Bestands (Deck, Fassung,
		// Format) sagen drüben nichts und gehen ebenso wenig mit.
		gepackt := *deck.Deck
		gepackt.ID, gepackt.Fassung, gepackt.FormatID, gepackt.Notiz = "", "", "", ""
		gepackt.Herkunft = HerkunftDTO{}

		return alsEineDatei(art, Deckdatei, gepackt), nil

	case austausch.ArtVorlage:
		return a.vorlageEntfalten(ctx, kennung)

	case austausch.ArtSammlung:
		return a.sichtEntfalten(ctx, kennung)
	}

	return nil, fmt.Errorf(
		"%w: %q wird hier nicht gepackt", domain.ErrValidation, art)
}

// alsEineDatei macht aus einer Gestalt den einen Inhalt.
//
// **Über `json.Marshal` derselben DTO, die auch REST und die Kommandozeile
// ausgeben.** Eine eigene Gestalt daneben wäre der zweite Serializer, den
// das Issue ausschließt — und sie liefe mit der ersten auseinander.
func alsEineDatei[Gestalt any](
	art austausch.Objektart, name string, wert Gestalt,
) []austausch.Inhalt {
	// **Ohne Fehlerzweig, und das ist keine Nachlässigkeit**: die beiden
	// Gestalten, die hier hereinkommen, tragen Zeichenketten, Zahlen,
	// Wahrheitswerte und Scheiben davon — `encoding/json` kann daran nicht
	// scheitern. Ein Zweig, den kein Lauf betreten kann, ist eine Zusage,
	// die niemand hält (§10); wer einer DTO ein `any`, eine Funktion oder
	// einen Kanal gibt, ändert diese Stelle mit.
	roh, _ := json.Marshal(wert) //nolint:errcheck // der Grund steht darüber

	return []austausch.Inhalt{{
		Art: art, Name: name, Bezeichnung: name, Bytes: roh,
	}}
}

// vorlageEntfalten packt die Dateien eines Vorlagenpakets.
func (a *App) vorlageEntfalten(
	ctx context.Context, paket string,
) ([]austausch.Inhalt, error) {
	verzeichnis := a.Paketdateien(ctx, paket)
	if verzeichnis.Fehler != nil {
		return nil, verzeichnis.Fehler.AlsFehler()
	}

	if len(verzeichnis.Dateien) == 0 {
		return nil, fmt.Errorf("%w: das Paket %s führt keine Datei",
			domain.ErrNotFound, paket)
	}

	inhalte := make([]austausch.Inhalt, 0, len(verzeichnis.Dateien))

	for _, name := range verzeichnis.Dateien {
		inhalt := a.PaketdateiLesen(ctx, paket, name)
		if inhalt.Fehler != nil {
			return nil, inhalt.Fehler.AlsFehler()
		}

		inhalte = append(inhalte, austausch.Inhalt{
			Art:         austausch.ArtVorlage,
			Name:        path.Join(Vorlagenordner, name),
			Bezeichnung: name,
			Bytes:       []byte(inhalt.Inhalt),
		})
	}

	return inhalte, nil
}

// sichtEntfalten packt, was ein Gegenüber sehen darf.
//
// **Die Kennung ist die des Empfängers, und der Zweck steht dahinter** — eine
// Sicht ohne Empfänger gibt es nicht: was hinausgeht, entscheidet **seine**
// Freigabe (T3).
func (a *App) sichtEntfalten(
	ctx context.Context, kennung string,
) ([]austausch.Inhalt, error) {
	wer, zweck, geteilt := zerlegeSichtkennung(kennung)
	if !geteilt {
		return nil, fmt.Errorf("%w: eine Sammlungssicht braucht Empfänger und "+
			"Zweck, getrennt durch einen Schrägstrich — etwa "+
			"%q", domain.ErrValidation, "bgegenueber/handel")
	}

	sicht := a.Freigabesicht(ctx, wer, zweck)
	if sicht.Fehler != nil {
		return nil, sicht.Fehler.AlsFehler()
	}

	// **Ohne Fehlerzweig** — derselbe Grund wie bei [alsEineDatei].
	roh, _ := json.Marshal(sicht.Zeilen) //nolint:errcheck // der Grund steht dort

	return []austausch.Inhalt{{
		Art: austausch.ArtSammlung, Name: Sichtdatei,
		Bezeichnung: wer + "/" + zweck, Bytes: roh,
	}}, nil
}

// zerlegeSichtkennung trennt Empfänger und Zweck.
func zerlegeSichtkennung(kennung string) (wer, zweck string, geteilt bool) {
	for stelle := range len(kennung) {
		if kennung[stelle] != '/' {
			continue
		}

		wer, zweck = kennung[:stelle], kennung[stelle+1:]

		return wer, zweck, wer != "" && zweck != ""
	}

	return "", "", false
}

// alsKennung deutet eine Kennung, die eine Zahl sein muss.
func alsKennung(roh, was string) (int64, error) {
	kennung, fehler := strconv.ParseInt(roh, 10, 64)
	if fehler != nil {
		return 0, domain.MitGrund(domain.ErrValidation, fehler,
			"die %s deuten", was)
	}

	return kennung, nil
}

package service

// **Was hinausgeht** — der Aufrufer für [Handelspost.Senden] aus `T5`
// (#1470),
// nachgetragen mit
// #1637.
//
// # Warum das eine eigene Stelle ist und kein Verfahren an der Post
//
// **Weil die Post nichts vom Ausweis weiß und nichts vom Stand.** Sie nimmt
// eine fertige Nachricht und legt sie in den Ausgang; **hier** entsteht sie —
// aus dem heutigen Stand, mit einer Folge, die hochzählt, und einer
// Unterschrift, die gegen die eigene Kennung prüft.
//
// # Und warum der Absender den Stand liest, statt ihn zu bekommen
//
// **Weil eine mitgeschickte Revision eine Behauptung wäre.** Wer schickt,
// schickt **das, was hier steht** — sonst unterschriebe er einen Abdruck, den
// er von irgendwoher hat, und die Gegenseite bekäme eine Unterschrift über
// etwas, das daneben nicht liegt.

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Folgezaehler vergibt die Folge je Vorgang.
//
// **Eine Schnittstelle und kein Zähler im Arbeitsspeicher**: sie muss einen
// Neustart überleben, und die Gegenseite vergleicht sie mit dem, was sie
// schon hat ([domain.Nachzuliefern]).
type Folgezaehler interface {
	NaechsteFolge(ctx context.Context, vorgang string) (int64, error)
}

// Belegquelle gibt, was von einem Abschluss hier liegt.
//
// **Zwei Rufe und nicht einer**: der Beleg trägt Abdruck und Unterschrift,
// die Zeilen stehen im Journal. Ein dritter Verbund, der beides zusammen
// hielte, wäre dieselbe Angabe zweimal — und die zweite liefe auseinander.
type Belegquelle interface {
	Beleg(ctx context.Context, vorgang string) (*domain.Belegstand, error)
	Journal(ctx context.Context, vorgang string) ([]domain.Journalzeile, error)
}

// Handelsversand schickt den Stand eines Vorgangs an ein Gegenüber.
type Handelsversand struct {
	post    *Handelspost
	stand   Handelsstand
	folgen  Folgezaehler
	ausweis Ausweisstelle

	// belege ist, woher die Urkunde kommt, wenn ein Beleg hinausgeht
	// (#1650).
	//
	// **Wahlfrei**: ein Versand ohne sie schickt jede andere Art wie bisher
	// und einen Beleg **ohne** Urkunde — und drüben bleibt `beidseitig`
	// falsch. Das ist die ehrliche Antwort auf „ich weiß es nicht".
	belege Belegquelle

	uhr     func() time.Time
	kennung func() string
}

// MitBelegen hängt die Quelle der Urkunden an.
func (v *Handelsversand) MitBelegen(belege Belegquelle) *Handelsversand {
	v.belege = belege

	return v
}

// NeuerHandelsversand stellt ihn auf.
func NeuerHandelsversand(
	post *Handelspost, stand Handelsstand, folgen Folgezaehler,
	ausweis Ausweisstelle,
) *Handelsversand {
	return &Handelsversand{
		post: post, stand: stand, folgen: folgen, ausweis: ausweis,
		uhr: time.Now, kennung: nachweiskennung,
	}
}

// MitUhr setzt die Uhr — für Läufe, die einen Zeitstempel vergleichen.
func (v *Handelsversand) MitUhr(uhr func() time.Time) *Handelsversand {
	v.uhr = uhr

	return v
}

// MitKennungen setzt die Quelle der Nachrichtenkennungen.
func (v *Handelsversand) MitKennungen(kennung func() string) *Handelsversand {
	v.kennung = kennung

	return v
}

// Schicken legt den heutigen Stand eines Vorgangs in den Ausgang.
//
// **Es schickt nicht über die Leitung.** Der Ausgang ist die Zusage aus `T5`:
// „abgeschickt" heißt nicht „angekommen", und was hier hineingeht, bleibt
// liegen, bis die Gegenseite es bestätigt.
//
// **Zurück kommt die Nachrichtenkennung** — an ihr räumt
// [Handelspost.Bestaetigt] die Zeile wieder weg.
func (v *Handelsversand) Schicken(
	ctx context.Context, vorgang, empfaenger string,
) (string, error) {
	angebot, fehler := v.stand.Stand(ctx, vorgang)
	if fehler != nil {
		return "", fehler
	}

	art, hinaus := domain.NachrichtenartFuer(angebot.Zustand)
	if !hinaus {
		return "", domain.MitGrund(domain.ErrValidation, nil,
			"aus dem Stand %q geht keine Nachricht hinaus", angebot.Zustand)
	}

	// **Der Ausweis zuletzt vor dem Bauen** — er kommt aus dem
	// Schlüsselbund, und der fragt zurück; wer ihn für einen Vorgang holte,
	// aus dem nichts hinausgeht, ließe jemanden einen Dialog wegklicken, um
	// danach eine Absage zu lesen (derselbe Satz wie bei
	// [Verbuchungsstelle.Verbuchen]).
	ich, fehler := v.ausweis.Ausweis(ctx)
	if fehler != nil {
		return "", fehler
	}

	// **Die Folge wird vor dem Einreihen vergeben, und ein Fehlschlag
	// danach verbrennt sie.**
	//
	// **Das ist die richtige Reihenfolge und nicht die bequeme.** Wer erst
	// einreihte und dann zählte, gäbe bei einem Abbruch dieselbe Nummer ein
	// zweites Mal — zwei Nachrichten mit derselben Folge, und drüben wäre
	// eine davon unsichtbar. Eine **Lücke** kostet dagegen nichts:
	// [domain.Nachzuliefern] gibt eine Untergrenze und zählt nicht nach.
	folge, fehler := v.folgen.NaechsteFolge(ctx, vorgang)
	if fehler != nil {
		return "", fehler
	}

	umschlag := domain.Handelsumschlag{
		Fassung: domain.Handelsprotokollfassung, Art: art,
		NachrichtID: v.kennung(), Vorgang: vorgang,
		Basisrevision: angebot.Revision, Abdruck: angebot.Abdruck(),
		Folge: folge, Absender: string(ich.Kennung()), Gesendet: v.uhr(),
	}

	umschlag.Unterschrift = hex.EncodeToString(
		ich.Signiere(umschlag.Unterschriftsform()))

	nachricht := Handelsnachricht{Umschlag: umschlag, Angebot: angebot.ZumVersand()}

	// **Ein Beleg geht mit seiner Urkunde hinaus** (#1650). Ohne sie bliebe
	// drüben `beidseitig` falsch — und das ist richtig so: was nicht belegt
	// ist, wird nicht behauptet.
	if art == domain.NachrichtBeleg {
		urkunde, fehler := v.urkunde(ctx, vorgang, angebot)
		if fehler != nil {
			return "", fehler
		}

		nachricht.Beleg = urkunde
	}

	if fehler := v.post.Senden(ctx, empfaenger, nachricht); fehler != nil {
		return "", fehler
	}

	return umschlag.NachrichtID, nil
}

// urkunde baut den unterschriebenen Abschluss, wie er hinausgeht.
//
// **Ohne Quelle und ohne Beleg gibt es keine** — und das ist kein Fehler:
// eine Nachricht der Art `beleg` sagt den **Zustand**, und der steht auch
// ohne Urkunde. Was fehlt, ist die Deckung, und die täuscht niemand vor.
func (v *Handelsversand) urkunde(
	ctx context.Context, vorgang string, angebot domain.Handelsangebot,
) (*domain.Belegurkunde, error) {
	if v.belege == nil {
		return nil, nil //nolint:nilnil // „keine Urkunde" ist kein Fehler
	}

	stand, fehler := v.belege.Beleg(ctx, vorgang)
	if fehler != nil {
		return nil, fehler
	}

	if stand == nil {
		return nil, nil //nolint:nilnil // der Grund steht im Kopf
	}

	zeilen, fehler := v.belege.Journal(ctx, vorgang)
	if fehler != nil {
		return nil, fehler
	}

	// **Der Stand kommt von oben** — [Handelsversand.Schicken] hat ihn
	// schon gelesen. Ein zweites Lesen gäbe einen Fehlerzweig, den kein
	// Lauf betreten kann (§10).
	beleg := domain.Handelsbeleg{
		Vorgang: vorgang, Abdruck: stand.Abdruck,
		Parteien: angebot.Parteien, Zeilen: zeilen,
		Gebucht: stand.Gebucht,
	}

	// **Das Lokale bleibt hier.** Bestandskennung, Lagerort und Notiz gehen
	// einen Fremden nichts an — und die Unterschrift passt trotzdem, weil
	// [domain.Handelsbeleg.Unterschriftsform] ohnehin ohne sie rechnet.
	return &domain.Belegurkunde{
		Beleg: beleg.Oeffentlich(), Unterschrift: stand.Unterschrift,
	}, nil
}

// Wiederaufnehmen sagt, was nach einem Abriss noch einmal hinausmuss.
//
// **Der eigene Fortschritt wird gelesen und nicht genannt.** Ein Aufrufer,
// der ihn mitschickte, nennte einen Stand, den er von irgendwoher hat — und
// eine Klärung, die daran hinge, wäre keine Aussage über diese beiden
// Rechner.
//
// **Die mittlere Rückgabe ist kein Fehler**: „es klemmt" heißt, dass beide
// Seiten ordentlich gearbeitet haben und trotzdem zwei Verläufe dastehen
// ([domain.Nachzuliefern]).
func (v *Handelsversand) Wiederaufnehmen(
	ctx context.Context, vorgang, empfaenger string,
	seiner domain.Handelsfortschritt,
) ([]Ausgangsnachricht, bool, error) {
	angebot, fehler := v.stand.Stand(ctx, vorgang)
	if fehler != nil {
		return nil, false, fehler
	}

	meiner := domain.Handelsfortschritt{
		Vorgang: vorgang, Revision: angebot.Revision,
		Abdruck: angebot.Abdruck(),
	}

	return v.post.Wiederaufnehmen(ctx, empfaenger, meiner, seiner)
}

// Bestaetigt räumt eine Nachricht weg, die angekommen ist.
//
// **Hier und nicht nur an der Post**: wer schickt, räumt auch weg — ein
// Aufrufer, der für das eine diese Stelle nähme und für das andere die Post,
// hielte zwei Wege auf denselben Ausgang.
func (v *Handelsversand) Bestaetigt(
	ctx context.Context, nachrichtID string,
) error {
	if nachrichtID == "" {
		return fmt.Errorf(
			"%w: eine Bestätigung nennt die Nachricht", domain.ErrValidation)
	}

	return v.post.Bestaetigt(ctx, nachrichtID)
}

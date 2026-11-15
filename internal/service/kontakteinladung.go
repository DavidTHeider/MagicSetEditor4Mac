package service

// **Person einladen, Code vergleichen, Zweck freigeben** — NET-01
// (#1705).
//
// # Der Ablauf, und wer was weiß
//
// **Beide tauschen ihre Einladung aus**, so wie es das Artboard *Gegenüber*
// zeichnet („tauscht ihr einen Einladungscode aus"): wer die des anderen
// annimmt, kennt dessen Schlüssel — und damit beide. Erst dann gibt es einen
// Vergleichscode, und er ist auf beiden Seiten derselbe.
//
// **Annehmen ist ein Schritt, nicht zwei**: ohne Vergleich entsteht ein
// unbestätigter Kontakt, mit passendem Vergleich ein bestätigter, mit
// falschem **keiner**. Einen „halb bestätigten" gibt es damit nicht (AK-04):
// die Vorschau schreibt nichts, und die Annahme schreibt ganz oder gar nicht.
//
// # Was nicht über MCP geht
//
// **Keiner der drei Wege.** Annehmen und Bestätigen sind
// Vertrauensentscheidungen ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md));
// Einladen gibt heraus, wie dieser Rechner zu erreichen ist, und wem, sagt
// ein Mensch.

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Einladungsablage hält Kontakte, angenommene und ausgestellte Einladungen.
type Einladungsablage interface {
	Lesen(ctx context.Context, kennung string) (domain.Peerkontakt, error)
	Angenommen(ctx context.Context, nonce string) (bool, error)
	Aufnehmen(ctx context.Context, aufnahme domain.Kontaktaufnahme, jetzt time.Time) error

	// Ausgestellt und Einloesen gehören dem Einladenden (NET-01b).
	Ausgestellt(ctx context.Context, nonce string, zweck domain.Zweck, ablauf, jetzt time.Time) error
	Einloesen(ctx context.Context, nonce string, kontakt domain.Peerkontakt, jetzt time.Time) error
}

// Rotationsstelle sagt, welche Kennung vor der letzten Rotation galt.
type Rotationsstelle interface {
	Lage(ctx context.Context) (Peerlage, error)
}

// Einladungswerk stellt Einladungen aus und nimmt sie an.
type Einladungswerk struct {
	ausweis  Ausweisstelle
	rotation Rotationsstelle
	ablage   Einladungsablage
	jetzt    func() time.Time
	zufall   io.Reader
	grenze   *Einloesegrenze

	// riegel fragt vor einer Meldung nach `externSenden` (#2000) — `nil`
	// heißt: niemand fragt.
	riegel Pruefer
}

// NeuesEinladungswerk baut es.
func NeuesEinladungswerk(
	ausweis Ausweisstelle, rotation Rotationsstelle, ablage Einladungsablage,
	jetzt func() time.Time,
) *Einladungswerk {
	return &Einladungswerk{
		ausweis: ausweis, rotation: rotation, ablage: ablage, jetzt: jetzt,
	}
}

// MitZufall setzt die Quelle der Nonce — für Tests, die sie festhalten.
func (w *Einladungswerk) MitZufall(zufall io.Reader) *Einladungswerk {
	w.zufall = zufall

	return w
}

// Einladungsbitte ist, was eine neue Einladung sagen soll.
type Einladungsbitte struct {
	Name    string
	Zweck   domain.Zweck
	Adresse string
}

// Ausgestellte ist eine neue Einladung.
type Ausgestellte struct {
	Code          string
	Ablauf        time.Time
	Fingerabdruck string
}

// Erstellen stellt eine Einladung aus, unterschrieben mit diesem Profil.
func (w *Einladungswerk) Erstellen(
	ctx context.Context, bitte Einladungsbitte,
) (Ausgestellte, error) {
	ich, fehler := w.ausweis.Ausweis(ctx)
	if fehler != nil {
		return Ausgestellte{}, fehler
	}

	lage, fehler := w.rotation.Lage(ctx)
	if fehler != nil {
		return Ausgestellte{}, fehler
	}

	nonce := make([]byte, 16)
	if _, lesefehler := io.ReadFull(w.zufallsquelle(), nonce); lesefehler != nil {
		return Ausgestellte{}, fmt.Errorf("eine Nonce ziehen: %w", lesefehler)
	}

	einladung := domain.Kontakteinladung{
		Kennung: string(ich.Kennung()), Name: strings.TrimSpace(bitte.Name),
		Zweck: bitte.Zweck, Adresse: bitte.Adresse,
		Ablauf: w.jetzt().UTC().Truncate(time.Second).
			Add(domain.KontakteinladungGueltig),
		Nonce:  base64.RawURLEncoding.EncodeToString(nonce),
		Vorher: string(lage.Vorher),

		// **Einlösbar ist, was sagt, wo** (NET-01b): ohne Adresse gibt es
		// keinen Ort, an dem sich jemand melden könnte.
		Einloesbar: bitte.Adresse != "",
	}

	code, fehler := einladung.Kodiere(ich.Signiere)
	if fehler != nil {
		return Ausgestellte{}, fehler
	}

	// **Erst vermerkt, dann ausgegeben**: eine Einladung, die der Einladende
	// nicht kennt, hieße beim Einlösen „fremd".
	if einladung.Einloesbar {
		if fehler := w.ablage.Ausgestellt(ctx, einladung.Nonce, einladung.Zweck,
			einladung.Ablauf, w.jetzt()); fehler != nil {
			return Ausgestellte{}, fehler
		}
	}

	return Ausgestellte{
		Code: code, Ablauf: einladung.Ablauf, Fingerabdruck: ich.Fingerabdruck(),
	}, nil
}

// Einladungsvorschau ist, was eine Einladung sagt — ohne dass etwas
// geschrieben ist.
type Einladungsvorschau struct {
	Einladung      domain.Kontakteinladung
	Fingerabdruck  string
	Vergleichscode string

	// Lage ist die eines vorhandenen Kontakts mit dieser Kennung.
	Lage domain.Kontaktlage

	// Vorgaenger ist der Kontakt, dessen neuer Schlüssel diese Einladung
	// sein will — nur, wenn es ihn hier gibt.
	Vorgaenger *domain.Peerkontakt

	Angenommen bool
}

// Pruefen liest eine Einladung und sagt, was sie ist.
//
// **Die Unterschrift zuerst**, und über die Bytes, die ankamen: erst danach
// sagt irgendein Feld etwas.
func (w *Einladungswerk) Pruefen(
	ctx context.Context, code string,
) (Einladungsvorschau, error) {
	einladung, unterschrieben, fehler := domain.ZerlegeKontakteinladung(code)
	if fehler != nil {
		return Einladungsvorschau{}, fehler
	}

	kennung := peer.Kennung(einladung.Kennung)
	if peer.Pruefe(kennung, unterschrieben.Daten, unterschrieben.Unterschrift) != nil {
		return Einladungsvorschau{}, fmt.Errorf(
			"%w: die Unterschrift passt nicht — die Einladung wurde verändert "+
				"oder stammt nicht von dem, den sie nennt", domain.ErrValidation)
	}

	if befund := einladung.Pruefe(w.jetzt()); befund != nil {
		return Einladungsvorschau{}, befund
	}

	ich, fehler := w.ausweis.Ausweis(ctx)
	if fehler != nil {
		return Einladungsvorschau{}, fehler
	}

	if ich.Kennung() == kennung {
		return Einladungsvorschau{}, fmt.Errorf(
			"%w: das ist deine eigene Einladung — sie gehört zu deinem "+
				"Gegenüber", domain.ErrValidation)
	}

	return w.vorschau(ctx, ich, einladung)
}

// vorschau schlägt nach, was dieses Profil über den Einladenden weiß.
func (w *Einladungswerk) vorschau(
	ctx context.Context, ich peer.Identitaet, einladung domain.Kontakteinladung,
) (Einladungsvorschau, error) {
	vorschau := Einladungsvorschau{
		Einladung:     einladung,
		Fingerabdruck: peer.Fingerabdruck(peer.Kennung(einladung.Kennung)),
		Vergleichscode: domain.Vergleichscode(
			string(ich.Kennung()), einladung.Kennung),
		Lage: domain.LageUnbekannt,
	}

	vorhanden, fehler := w.nachschlagen(ctx, einladung.Kennung)
	if fehler != nil {
		return Einladungsvorschau{}, fehler
	}

	if vorhanden != nil {
		vorschau.Lage = vorhanden.Lage()
	}

	if einladung.Vorher != "" {
		if vorschau.Vorgaenger, fehler = w.nachschlagen(
			ctx, einladung.Vorher); fehler != nil {
			return Einladungsvorschau{}, fehler
		}
	}

	if vorschau.Angenommen, fehler = w.ablage.Angenommen(
		ctx, einladung.Nonce); fehler != nil {
		return Einladungsvorschau{}, fehler
	}

	return vorschau, nil
}

// nachschlagen gibt einen Kontakt — oder nichts, wenn es ihn nicht gibt.
func (w *Einladungswerk) nachschlagen(
	ctx context.Context, kennung string,
) (*domain.Peerkontakt, error) {
	kontakt, fehler := w.ablage.Lesen(ctx, kennung)

	switch {
	case errors.Is(fehler, domain.ErrNotFound):
		return nil, nil //nolint:nilnil // „keiner" ist eine Antwort
	case fehler != nil:
		return nil, fehler
	}

	return &kontakt, nil
}

// Einladungsannahme ist, was ein Mensch zu einer Einladung entschieden hat.
type Einladungsannahme struct {
	Code string

	// EigenerName ist, wie dieses Profil das Gegenüber nennen will.
	EigenerName string

	// Verglichen ist der Code, den beide verglichen haben — leer heißt:
	// noch nicht.
	Verglichen string

	// Freigeben gibt den Zweck der Einladung frei — nur mit Vergleich.
	Freigeben bool

	// MeinName ist, wie dieses Profil sich beim Einladenden nennt — er geht
	// hinaus (NET-01b). Leer heißt: ohne Namen.
	MeinName string
}

// Annehmen nimmt eine Einladung an — ganz oder gar nicht.
func (w *Einladungswerk) Annehmen(
	ctx context.Context, annahme Einladungsannahme,
) (domain.Peerkontakt, error) {
	vorschau, fehler := w.Pruefen(ctx, annahme.Code)
	if fehler != nil {
		return domain.Peerkontakt{}, fehler
	}

	if verwehrt := w.darfMelden(ctx, vorschau.Einladung); verwehrt != nil {
		return domain.Peerkontakt{}, verwehrt
	}

	kontakt, fehler := aufnahmeAus(vorschau, annahme)
	if fehler != nil {
		return domain.Peerkontakt{}, fehler
	}

	if fehler := w.ablage.Aufnehmen(ctx, domain.Kontaktaufnahme{
		Kontakt: kontakt, Nonce: vorschau.Einladung.Nonce,
	}, w.jetzt()); fehler != nil {
		return domain.Peerkontakt{}, fehler
	}

	// **Geschrieben ist geschrieben** (§7, Regel 3; Zweitblick über Phase
	// 6.7, Z9): das Nachlesen gibt den gespeicherten Stand — einer, der schon
	// bestätigt war, bleibt es —, aber klemmt es, ist die Annahme trotzdem
	// geschehen, und die Antwort ist das, was geschrieben wurde.
	if gespeichert, fehler := w.ablage.Lesen(ctx, kontakt.Kennung); fehler == nil {
		return gespeichert, nil
	}

	return kontakt, nil
}

// aufnahmeAus entscheidet, was geschrieben wird.
func aufnahmeAus(
	vorschau Einladungsvorschau, annahme Einladungsannahme,
) (domain.Peerkontakt, error) {
	einladung := vorschau.Einladung

	switch {
	case vorschau.Angenommen:
		return domain.Peerkontakt{}, fmt.Errorf(
			"%w: diese Einladung ist schon angenommen", domain.ErrConflict)
	case vorschau.Lage == domain.LageWiderrufen:
		return domain.Peerkontakt{}, fmt.Errorf(
			"%w: dieses Gegenüber ist gesperrt — erst entsperren",
			domain.ErrConflict)
	case utf8.RuneCountInString(annahme.EigenerName) > 64:
		return domain.Peerkontakt{}, fmt.Errorf(
			"%w: ein eigener Name hat höchstens 64 Zeichen", domain.ErrValidation)
	case strings.TrimSpace(annahme.MeinName) != "" &&
		domain.BrauchbarerName(annahme.MeinName) == "":
		return domain.Peerkontakt{}, fmt.Errorf(
			"%w: der Name, unter dem du dich meldest, hat ein bis 64 lesbare "+
				"Zeichen", domain.ErrValidation)
	}

	kontakt := domain.Peerkontakt{
		Kennung: einladung.Kennung, Fingerabdruck: vorschau.Fingerabdruck,
		SelbstgewaehlterName: einladung.Name,
		EigenerName:          strings.TrimSpace(annahme.EigenerName),
		Adresse:              einladung.Adresse, Stand: domain.KontaktUnbestaetigt,
	}

	if vorschau.Vorgaenger != nil {
		kontakt.Vorgaenger = einladung.Vorher
	}

	if annahme.Verglichen != "" {
		if !domain.VergleichscodeGleich(
			vorschau.Vergleichscode, annahme.Verglichen) {
			return domain.Peerkontakt{}, fmt.Errorf(
				"%w: der verglichene Code stimmt nicht — aufgenommen wurde "+
					"nichts", domain.ErrValidation)
		}

		kontakt.Stand = domain.KontaktBestaetigt
	}

	if annahme.Freigeben {
		if kontakt.Stand != domain.KontaktBestaetigt {
			return domain.Peerkontakt{}, fmt.Errorf(
				"%w: freigegeben wird erst nach dem Vergleich",
				domain.ErrValidation)
		}

		if einladung.Zweck != domain.ZweckPaarung {
			kontakt.Rechte = []domain.Zweck{einladung.Zweck}
		}
	}

	return kontakt, nil
}

// zufallsquelle gibt die eingesetzte Quelle oder die des Systems.
func (w *Einladungswerk) zufallsquelle() io.Reader {
	if w.zufall != nil {
		return w.zufall
	}

	return rand.Reader
}

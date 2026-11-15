package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Zusatzfeldwert ist ein Zusatzfeld des Stils einer Karte samt dem Wert, den
// sie dafür trägt
// (#2298).
type Zusatzfeldwert struct {
	Feld domain.Field

	// Wert ist, was die Karte unter `extra_data:` für ihre Stilvorlage trägt.
	Wert string

	// Gespeichert sagt, ob sie dort überhaupt etwas trägt — ein leerer Wert
	// und keiner sind zwei Aussagen (0010): ohne gilt die Vorgabe des Stils.
	Gespeichert bool
}

// Kartenstilquelle gibt den Stil **einer Karte**.
//
// **Schmaler als [Stilquelle]**, und das ist §2: die Pflege fragt nie nach dem
// Stil eines Sets.
type Kartenstilquelle interface {
	StilFuerKarte(ctx context.Context, karte *domain.CardRow) (Vorlage, []domain.Befund, bool)
}

// Kartenfeldablage schreibt einen Feldwert an eine Karte — **nur, wenn sie
// noch auf dem gelesenen Stand steht** (Befund 1 des Nachlaufs über E/F),
// geprüft in derselben Transaktion, sonst `ErrConflict`. Erfüllt vom
// Bestand (`storage.Kartenschreiber`).
//
// **Ohne Set**: ob die Karte zum Set gehört, prüft die Pflege beim Lesen,
// und ändert sich die Karte danach, scheitert der Stand.
type Kartenfeldablage interface {
	FeldSetzenAufStand(ctx context.Context, karteID int64, feld, wert, erwartet string) (
		vorher string, hatte bool, fehler error)
}

// Zusatzvorwert ist, was vor dem Setzen galt — der Block `extra data` und
// darin das Feld. **Aus demselben Stand, auf den geschrieben wurde**: mit ihm
// geht die Rücknahme.
type Zusatzvorwert struct {
	// Block ist der ganze Wert von `extra data`; BlockDa sagt, ob die Karte
	// ihn trug.
	Block   string
	BlockDa bool

	// Wert ist der gespeicherte Wert des Felds; Gespeichert sagt, ob es einen
	// gab — sonst galt die Vorgabe des Stils.
	Wert        string
	Gespeichert bool
}

// zusatzversuche ist, wie oft das Setzen neu liest, wenn die Karte sich
// zwischen Lesen und Schreiben geändert hat. Danach meldet es den Widerspruch.
const zusatzversuche = 3

// Zusatzwertpflege liest und setzt die Werte der Zusatzfelder einer Karte.
//
// # Wie MSE2 sie bearbeitet
//
// Auf der Karte selbst: der Editor zeigt ihre Felder **und** die Zusatzfelder
// ihres Stils (`setData(card->data, &card->extraDataFor(*stylesheet))`,
// `src/render/card/viewer.cpp`, Zeile 139, Fassung f9d9356 vom 2021-07-06),
// und bearbeitbar ist, was `editable` ist. Hier steht die Karte nicht als
// Eingabefläche da; die Felder stehen deshalb im Formular neben ihr.
//
// # Gespeichert wird über den gewöhnlichen Weg
//
// **Als ein Feldwert `extra data`** — derselbe Weg wie jede andere Eingabe,
// mit Verlauf, Rücknahme und der Prüfung, ob die Karte zum Set gehört. Ein
// eigener Speicherweg für einen Teil eines Feldwerts wäre eine zweite
// Antwort auf dieselbe Frage.
type Zusatzwertpflege struct {
	karten Kartensicht
	stile  Kartenstilquelle
	ablage Kartenfeldablage
}

// NeueZusatzwertpflege baut die Pflege über ihren drei Nähten.
func NeueZusatzwertpflege(
	karten Kartensicht, stile Kartenstilquelle, ablage Kartenfeldablage,
) *Zusatzwertpflege {
	return &Zusatzwertpflege{karten: karten, stile: stile, ablage: ablage}
}

// Felder gibt die Zusatzfelder, die sich an dieser Karte von Hand setzen
// lassen, samt ihrem Wert.
//
// **Nur, was gespeichert wird und bearbeitbar ist** — `save value: false`
// rechnet ein Skript bei jedem Zeichnen neu, und `editable: false` lässt
// auch MSE2 nicht anfassen.
func (p *Zusatzwertpflege) Felder(ctx context.Context, setID, karteID int64) ([]Zusatzfeldwert, error) {
	karte, vorlage, err := p.karteMitStil(ctx, setID, karteID)
	if err != nil {
		return nil, err
	}

	vorher, _ := karte.Value(msepaket.Entkanonisch(zusatzdatenfeld))

	gespeichert, befunde := zusatzwerteAus(vorher, vorlage.Stil.Paket, vorlage.Stil.Zusatzfelder)
	if len(befunde) > 0 {
		return nil, unlesbareZusatzdaten(befunde)
	}

	felder := make([]Zusatzfeldwert, 0, len(vorlage.Stil.Zusatzfelder))

	for i := range vorlage.Stil.Zusatzfelder {
		feld := vorlage.Stil.Zusatzfelder[i]
		if !vonHand(feld) {
			continue
		}

		wert, da := gespeichert[msepaket.Kanonisch(feld.Name)]
		felder = append(felder, Zusatzfeldwert{Feld: feld, Wert: wert, Gespeichert: da})
	}

	return felder, nil
}

// Setzen schreibt den Wert eines Zusatzfelds in den Block der Stilvorlage der
// Karte und gibt, was vorher galt.
//
// **Auf den gelesenen Stand** (Befund 1 des Nachlaufs über E/F): der Block
// wird aus der gelesenen Karte gebaut und ganz geschrieben. Setzte eine andere
// Fläche dazwischen ein anderes Zusatzfeld derselben Karte, schrieb dieser
// Aufruf deren Änderung weg, und beide meldeten Erfolg. Jetzt gilt der Stand
// der gelesenen Karte als Bedingung; hat sie sich geändert, wird neu gelesen
// und neu gebaut — höchstens [zusatzversuche] Mal.
func (p *Zusatzwertpflege) Setzen(
	ctx context.Context, setID, karteID int64, feld, wert string,
) (Zusatzvorwert, error) {
	var fehler error

	for range zusatzversuche {
		var vorwert Zusatzvorwert

		vorwert, fehler = p.einmalSetzen(ctx, setID, karteID, feld, wert)
		if !errors.Is(fehler, domain.ErrConflict) {
			return vorwert, fehler
		}
	}

	return Zusatzvorwert{}, fehler
}

// einmalSetzen liest, baut und schreibt einmal.
func (p *Zusatzwertpflege) einmalSetzen(
	ctx context.Context, setID, karteID int64, feld, wert string,
) (Zusatzvorwert, error) {
	karte, vorlage, err := p.karteMitStil(ctx, setID, karteID)
	if err != nil {
		return Zusatzvorwert{}, err
	}

	if !vonHandSetzbar(vorlage.Stil.Zusatzfelder, feld) {
		return Zusatzvorwert{}, domain.NewFieldError("feld", domain.SchluesselValidierung,
			"%q ist kein Zusatzfeld dieses Stils, das sich von Hand setzen lässt.", feld)
	}

	block, blockDa := karte.Value(msepaket.Entkanonisch(zusatzdatenfeld))
	// Ein unlesbarer Block meldet [mitZusatzwert] gleich darunter.
	gespeichert, _ := zusatzwerteAus(block, vorlage.Stil.Paket, vorlage.Stil.Zusatzfelder)
	alt, gesetzt := gespeichert[msepaket.Kanonisch(feld)]

	neu, err := mitZusatzwert(block, vorlage.Stil, msepaket.Kanonisch(feld), wert)
	if err != nil {
		return Zusatzvorwert{}, err
	}

	_, _, err = p.ablage.FeldSetzenAufStand(ctx, karteID,
		msepaket.Entkanonisch(zusatzdatenfeld), neu, karte.Stand())
	if err != nil {
		return Zusatzvorwert{}, err
	}

	return Zusatzvorwert{Block: block, BlockDa: blockDa, Wert: alt, Gespeichert: gesetzt}, nil
}

// karteMitStil liest die Karte, prüft, dass sie zum Set gehört, und gibt
// ihren Stil.
//
// **Ohne Stil gibt es keine Zusatzfelder** — und auch nichts zu setzen: wo
// der Stil sich nicht lesen ließ, wäre jeder Name geraten.
//
// **Und das wird gesagt, statt eine leere Liste zu geben** (J05, #2269): bis
// zum 2026-10-11 kam ein unlesbarer Stil als „keine Zusatzfelder" an, die
// Oberfläche blendete den Abschnitt aus, und ein Setzen meldete, das Feld sei
// keins. Die Befunde des Ladens gehen als Grund mit — ins Protokoll (0075).
func (p *Zusatzwertpflege) karteMitStil(
	ctx context.Context, setID, karteID int64,
) (*domain.CardRow, Vorlage, error) {
	karte, err := p.karten.KarteLesen(ctx, karteID)
	if err != nil {
		return nil, Vorlage{}, fmt.Errorf("die Karte lesen: %w", err)
	}

	if karte.SetID != setID {
		return nil, Vorlage{}, fmt.Errorf("die Karte %d gehört nicht zum Set %d: %w",
			karteID, setID, domain.ErrNotFound)
	}

	vorlage, befunde, da := p.stile.StilFuerKarte(ctx, karte)
	if !da {
		return nil, Vorlage{}, domain.MitGrund(domain.ErrConflict, befundgrund(befunde),
			"Der Stil dieser Karte ließ sich nicht laden — ohne ihn gibt es keine Zusatzfelder.")
	}

	return karte, vorlage, nil
}

// vonHand sagt, ob ein Zusatzfeld einen Wert trägt, den jemand setzt.
func vonHand(feld domain.Field) bool {
	return feld.Editierbar && feld.Gespeichert
}

// vonHandSetzbar sagt, ob der Stil ein Zusatzfeld dieses Namens führt, das sich von
// Hand setzen lässt.
func vonHandSetzbar(felder []domain.Field, name string) bool {
	for i := range felder {
		if msepaket.Kanonisch(felder[i].Name) == msepaket.Kanonisch(name) {
			return vonHand(felder[i])
		}
	}

	return false
}

// unlesbareZusatzdaten macht aus dem Befund des Lesens einen Fehler am Feld.
//
// **Eine Beanstandung und kein stilles Überschreiben**: wer an Daten setzt,
// die sich nicht lesen lassen, verlöre alles, was darin steht.
func unlesbareZusatzdaten(befunde []domain.Befund) error {
	return domain.NewFieldError(msepaket.Entkanonisch(zusatzdatenfeld), domain.SchluesselValidierung,
		"%s", befunde[0].Meldung)
}

// befundgrund fasst die Befunde des Ladens zu einem Grund zusammen.
func befundgrund(befunde []domain.Befund) error {
	saetze := make([]string, 0, len(befunde))
	for _, befund := range befunde {
		saetze = append(saetze, befund.Meldung)
	}

	if len(saetze) == 0 {
		return errors.New("der Stil meldet keinen Befund")
	}

	return errors.New(strings.Join(saetze, "; "))
}

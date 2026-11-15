package service

// **Was ein Gegenüber zum Tausch freigibt, über die Leitung** — P11
// (#2379),
// Befunde C-15 und Z-02.
//
// # Eine Frage und eine Antwort auf derselben Sitzung
//
// **Die Handelspost ist sonst einseitig**: was hinausgeht, liegt im Ausgang,
// bis drüben bestätigt ist. Eine Sicht ist kein Zustand eines Handels und
// wird nirgends gespeichert — sie ist eine Auskunft, und die kommt zurück,
// solange die Leitung steht, oder gar nicht. Deshalb kein Ausgang und kein
// Briefkasten, sondern `sichtanfrage` hin und `sicht` zurück.
//
// # Was der Gefragte prüft und was der Fragende
//
// **Der Gefragte** antwortet nur dem, der spricht, und nur mit dem, was
// [Freigabepflege.Sicht] diesem Gegenüber zeigt — dieselbe Rechnung wie die
// Vorschau „was sähe Bo" und der Umschlag aus T7. Nach einem Widerruf ist das
// eine Absage. **Der Fragende** nimmt nur eine Antwort in seiner Fassung, auf
// seine Anfrage, vom Gefragten, und wartet nicht länger als seine Frist.

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sichtgeber ist, was die Post für eine Antwort braucht — erfüllt von
// [Freigabepflege].
type Sichtgeber interface {
	Sicht(ctx context.Context, empfaenger, sitzung string, zweck domain.Zweck) ([]domain.Sichtkarte, error)
}

// MitSichtgeber lässt diese Post Sichtanfragen beantworten.
func (p *Handelspost) MitSichtgeber(geber Sichtgeber) *Handelspost {
	p.sichtgeber = geber

	return p
}

// Die zwei Absagen, als Wort und nicht als Satz: ein Satz von drüben ist eine
// Behauptung in einer fremden Sprache, und den Satz am Fenster schreibt der
// Fragende selbst.
const (
	absageNichtFreigegeben = "nichtFreigegeben"
	absageNichtLesbar      = "nichtLesbar"
)

// sichtBeantworten antwortet auf eine Sichtanfrage über dieselbe Strecke.
func (p *Handelspost) sichtBeantworten(
	ctx context.Context, strecke Handelsstrecke, sender string, anfrage Handelsnachricht,
) (Zustellungsantwort, error) {
	if p.sichtgeber == nil || p.ausweis == nil {
		return Zustellungsantwort{}, domain.MitGrund(domain.ErrValidation, nil,
			"Diese Post beantwortet keine Sichtanfrage.")
	}

	if sender == "" || sender != anfrage.Umschlag.Absender {
		return Zustellungsantwort{}, domain.MitGrund(domain.ErrErlaubnis, nil,
			"Diese Nachricht nennt einen anderen Absender, als wer spricht.")
	}

	if befunde := anfrage.Umschlag.Validate(); len(befunde) > 0 {
		return Zustellungsantwort{}, befunde[0]
	}

	ich, fehler := p.ausweis.Ausweis(ctx)
	if fehler != nil {
		return Zustellungsantwort{}, fehler
	}

	antwort := Handelsnachricht{Umschlag: domain.Handelsumschlag{
		Fassung: domain.Handelsprotokollfassung, Art: domain.NachrichtSicht,
		NachrichtID: nachweiskennung(), Vorgang: anfrage.Umschlag.Vorgang, Folge: 1,
		Absender: string(ich.Kennung()), Gesendet: time.Now(),
	}}
	antwort.Umschlag.Unterschrift = hex.EncodeToString(ich.Signiere(antwort.Umschlag.Unterschriftsform()))

	// **Die Sitzung ist die Antwort**: innerhalb einer Antwort zeigt derselbe
	// Name auf dieselbe Zeile, über zwei Anfragen hinweg nicht. **Und ihr
	// Schlüssel kommt von hier**, nicht aus der Anfrage (Befund 5 des
	// Zweitblicks über Phase 8): die Vorgangskennung schreibt der Fragende,
	// und mit ihr rechnete er die Zeilenkennungen hinter den Referenzen nach.
	// Zurückgerechnet wird eine Referenz der Sicht nirgends.
	sicht, fehler := p.sichtgeber.Sicht(ctx, sender, "sicht/"+nachweiskennung(), domain.ZweckHandel)

	ergebnis := "sicht"

	switch {
	case errors.Is(fehler, domain.ErrErlaubnis):
		antwort.Absage, ergebnis = absageNichtFreigegeben, "abgesagt"
	case fehler != nil:
		antwort.Absage, ergebnis = absageNichtLesbar, "abgesagt"
	default:
		antwort.Sicht = sicht
	}

	// **Ohne Fehlerzweig am Kodierer** — derselbe Grund wie bei
	// [Handelspost.SendeUeber].
	roh, _ := json.Marshal(antwort) //nolint:errcheck // der Grund steht darüber

	if fehler := strecke.Sende(ctx, roh); fehler != nil {
		return Zustellungsantwort{}, fehler
	}

	return Zustellungsantwort{Art: domain.ZustellungNeu, Ergebnis: ergebnis}, nil
}

// Sichtfrager fragt ein Gegenüber nach seiner freigegebenen Sicht.
type Sichtfrager struct {
	ausweis Ausweisstelle
	frist   time.Duration
}

// Sichtfrist ist, wie lange der Fragende auf die Antwort wartet.
//
// **Zehn Sekunden**, wie beim Bild einer Karte am fremden Tisch (P9): eine
// Sicht ist eine Liste aus einem lokalen Bestand und kein Weg durchs Netz.
const Sichtfrist = 10 * time.Second

// NeuerSichtfrager stellt ihn auf.
func NeuerSichtfrager(ausweis Ausweisstelle) *Sichtfrager {
	return &Sichtfrager{ausweis: ausweis, frist: Sichtfrist}
}

// MitFrist setzt die Frist — für Läufe, die nicht zehn Sekunden warten.
func (f *Sichtfrager) MitFrist(frist time.Duration) *Sichtfrager {
	f.frist = frist

	return f
}

// Fragen schickt eine Sichtanfrage über die Strecke und gibt die Antwort.
func (f *Sichtfrager) Fragen(
	ctx context.Context, strecke Handelsstrecke, gegenueber string,
) ([]domain.Sichtkarte, error) {
	ich, fehler := f.ausweis.Ausweis(ctx)
	if fehler != nil {
		return nil, fehler
	}

	kennung := nachweiskennung()
	anfrage := domain.Handelsumschlag{
		Fassung: domain.Handelsprotokollfassung, Art: domain.NachrichtSichtanfrage,
		NachrichtID: kennung, Vorgang: "sicht/" + kennung, Folge: 1,
		Absender: string(ich.Kennung()), Gesendet: time.Now(),
	}
	anfrage.Unterschrift = hex.EncodeToString(ich.Signiere(anfrage.Unterschriftsform()))

	roh, _ := json.Marshal(Handelsnachricht{Umschlag: anfrage}) //nolint:errcheck // wie bei [Handelspost.SendeUeber]

	if gesendet := strecke.Sende(ctx, roh); gesendet != nil {
		return nil, gesendet
	}

	warten, ende := context.WithTimeout(ctx, f.frist)
	defer ende()

	geantwortet, fehler := strecke.Empfange(warten)
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"Das Gegenüber hat nicht geantwortet — vielleicht kennt seine Fassung keine Sichtanfrage.")
	}

	var antwort Handelsnachricht
	if unlesbar := json.Unmarshal(geantwortet, &antwort); unlesbar != nil {
		return nil, domain.MitGrund(domain.ErrValidation, unlesbar,
			"Die Antwort des Gegenübers ließ sich nicht lesen.")
	}

	return sichtAus(antwort, anfrage.Vorgang, gegenueber)
}

// sichtAus prüft eine Antwort und gibt ihre Zeilen — jede Referenz einmal.
func sichtAus(antwort Handelsnachricht, vorgang, gegenueber string) ([]domain.Sichtkarte, error) {
	umschlag := antwort.Umschlag

	switch {
	case umschlag.Fassung != domain.Handelsprotokollfassung:
		return nil, domain.MitGrund(domain.ErrValidation, nil,
			"Das Gegenüber spricht eine andere Fassung des Handelsprotokolls.")
	case umschlag.Art != domain.NachrichtSicht || len(umschlag.Validate()) > 0:
		return nil, domain.MitGrund(domain.ErrValidation, nil,
			"Die Antwort des Gegenübers ist keine Sicht.")
	case umschlag.Absender != gegenueber:
		return nil, domain.MitGrund(domain.ErrErlaubnis, nil,
			"Diese Nachricht nennt einen anderen Absender, als wer spricht.")
	case umschlag.Vorgang != vorgang:
		return nil, domain.MitGrund(domain.ErrConflict, nil,
			"Die Antwort gehört zu einer anderen Anfrage.")
	case antwort.Absage == absageNichtFreigegeben:
		return nil, domain.MitGrund(domain.ErrNichtFreigegeben, nil,
			"Das Gegenüber gibt dir zum Handeln nichts frei.")
	case antwort.Absage != "":
		return nil, domain.MitGrund(domain.ErrValidation, nil,
			"Das Gegenüber konnte seine Sicht gerade nicht lesen.")
	}

	gesehen := map[string]bool{}
	zeilen := make([]domain.Sichtkarte, 0, len(antwort.Sicht))

	for _, zeile := range antwort.Sicht {
		if gesehen[zeile.Referenz] {
			continue
		}

		gesehen[zeile.Referenz] = true
		zeilen = append(zeilen, zeile)
	}

	return zeilen, nil
}

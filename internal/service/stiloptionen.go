package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Die Stiloptionen eines Sets (#41).
//
// **Eine eigene Datei, und der Schnitt folgt der Frage** (§2): `stilwahl.go`
// beantwortet „welchen Stil", hier steht „welche Optionen hat er".

// Stiloption ist eine Option samt ihrem gegenwärtigen Wert.
type Stiloption struct {
	domain.StyleField

	// Wert ist der Stand des Sets — oder [domain.StyleField.Initial], solange
	// niemand etwas gesetzt hat.
	Wert string

	// Gesetzt sagt, ob [Stiloption.Wert] aus dem Bestand kommt oder aus der
	// Vorgabe ([stiloptionsvorgabe]).
	//
	// **Zwei Angaben und nicht eine.** Bei MSE ist ein Wert ein `Defaultable`;
	// die Markierung überlebt sogar das Vorgabeskript, weil
	// `change(Defaultable<T>&, …)` `assignDontChangeDefault` ruft
	// (`src/script/scriptable.hpp`, Zeile 40, Fassung f9d9356 vom
	// 2021-07-06). Ohne die Unterscheidung gibt es weder den Vorgabeeintrag im
	// Farbeditor noch die Handlung *zurück auf die Vorgabe*
	// (#764).
	Gesetzt bool

	// HatVorgabe sagt, ob die Vorlage überhaupt ein `default:` nennt.
	//
	// **Nicht dasselbe wie ein nichtleerer [Stiloption.Vorgabe]** — und der
	// Zweitblick über Phase L hat genau daran den Fall gefunden: `default:`
	// darf ein Skript sein, das **scheitert**. `magic-m15-bigtext` schreibt
	// `default: black` ohne Anführungszeichen; das ist ein Fehler, und der Wert
	// fällt auf `initial:` zurück.
	//
	// **MSE2 fragt nach dem Skript, nicht nach seinem Ergebnis**:
	// `hasDefault() { return field().default_script; }`
	// (`src/gui/value/color.cpp`, Zeile 38, Fassung f9d9356 vom 2021-07-06).
	// Der Eintrag steht dort also auch dann, wenn nichts herauskommt.
	HatVorgabe bool

	// Vorgabe ist der **gerechnete** Wert aus `default:` — leer, wenn die
	// Vorlage keinen nennt **oder das Skript nicht aufgeht**.
	//
	// **Nicht dasselbe wie [Stiloption.Wert].** Der trägt den Rückfall und ist
	// fast nie leer; hinge der Vorgabeeintrag daran, stünde er bei fast jeder
	// Option. MSE zeigt ihn nur bei `hasDefault()`, und das ist genau das
	// Vorhandensein des Skripts (`src/gui/value/color.cpp`, Zeile 38, Fassung
	// f9d9356 vom 2021-07-06).
	Vorgabe string

	// Moeglichkeiten sind die Wahlen der Option — Wert und Beschriftung.
	//
	// **Für jede Art, die eine Auswahl ist**: [domain.KindChoice],
	// [domain.KindBoolean], [domain.KindPackageChoice] und
	// [domain.KindColor]. Bei den ersten beiden sind Wert und Beschriftung
	// dasselbe; bei den anderen nicht (#632, #634).
	//
	// **Sie stehen hier und nicht in [domain.StyleField]**, weil eine
	// Paketwahl nicht im Stilpaket steht: sie entsteht aus dem Muster **und**
	// dem, was auf dieser Platte installiert ist. Dieselbe Vorlage gibt auf
	// zwei Rechnern zwei Listen.
	Moeglichkeiten []domain.Wahlmoeglichkeit

	// HatWahlbilder sagt, ob diese Option zu ihren Wahlen Bilder erklärt
	// (#1225).
	//
	// **Der Grund, warum die Oberfläche nicht alles rastert.** `magic-m15`
	// erklärt dreizehn Optionen und gibt vieren ein Bild; über das ganze Pack
	// wären es 1.097 Blöcke. Wer die Bilder ohne diese Auskunft holte, fragte
	// für jede Option nach und bekäme neunmal von dreizehn nichts.
	HatWahlbilder bool
}

// hatWahlbilder sagt, ob eine Stiloption Bilder zu ihren Wahlen hat.
//
// **Zwei Wege, wie in [Planbauer.Optionsbilder]**: ein `choice images:`-Block
// oder ein **gerechnetes** `image:` (`src/gui/value/choice.cpp`, Zeilen
// 147–160, Fassung f9d9356 vom 2021-07-06). Wer nur den ersten prüfte, ließe
// die Reihe genau dort leer, wo der Stil sie rechnet.
func hatWahlbilder(stil msepaket.Stil, name string) bool {
	sicht, da := stil.Optionsstil(name)
	if !da {
		return false
	}

	if len(sicht.Wahlbilder) > 0 {
		return true
	}

	return bildGerechnet(sicht.Bild)
}

// Optionen gibt die Stiloptionen des Sets samt Werten.
func (w *Stilwaehler) Optionen(ctx context.Context, setID int64) ([]Stiloption, error) {
	kopf, err := w.kopf(ctx, setID)
	if err != nil {
		return nil, err
	}

	if kopf.Style == "" {
		return []Stiloption{}, nil
	}

	stil, err := w.stil(kopf.Style, kopf.Game)
	if err != nil {
		return nil, err
	}

	werte, err := w.ablage.Stilwerte(ctx, setID)
	if err != nil {
		return nil, err
	}

	// **Dazwischen der Block der Datei** — dieselbe Reihenfolge wie bei der
	// Rasterung (`stilwirkung.go`): sonst zeichnete die Karte »dunkel«, und
	// die Liste nannte »hell«.
	angaben, err := w.koepfe.Setwerte(ctx, setID)
	if err != nil {
		return nil, err
	}

	ausDatei := setstilwerte(angaben, stilname(stil))

	optionen := make([]Stiloption, 0, len(stil.Stiloptionen))

	// **Erst die Anfangswerte, dann die gespeicherten, dann die Vorgaben** —
	// MSEs zwei Schritte (#789) und der dritte aus #860. Ohne den ersten sähe
	// ein `default:`, das eine Nachbaroption liest, gar nichts; ohne den
	// zweiten sähe es immer den Anfangswert und nie den, den jemand
	// eingestellt hat.
	styling := mitGespeicherten(mitGespeicherten(anfangsstyling(stil.Stiloptionen), alsStilwerte(ausDatei)), werte)

	for _, feld := range stil.Stiloptionen {
		option := stiloptionAus(ctx, feld, styling)

		if wert, genannt := ausDatei[strings.ReplaceAll(feld.Name, " ", "_")]; genannt {
			option.Wert = wert
		}

		for _, wert := range werte {
			if wert.Field == feld.Name {
				option.Wert = wert.Text
				option.Gesetzt = true
			}
		}

		// **Auch der gelesene Wert steht in Normalform** (#637). MSE ruft
		// `normalForm` in `MultipleChoiceValue::update`, also bei jedem
		// Durchgang und nicht nur beim Schreiben
		// (`src/data/field/multiple_choice.cpp`, Zeilen 54–60, Fassung
		// f9d9356 vom 2021-07-06). Ohne das käme ein `initial:` in der
		// Reihenfolge der Datei an, und die Kästchen stünden anders als der
		// gespeicherte Wert.
		option.Wert = gespeicherteForm(feld, option.Wert)

		wahlen, fehler := moeglichkeitenVon(ctx, w.katalog, feld)
		if fehler != nil {
			return nil, fehler
		}

		option.Moeglichkeiten = wahlen
		option.HatWahlbilder = hatWahlbilder(stil, feld.Name)

		optionen = append(optionen, option)
	}

	return optionen, nil
}

// OptionSetzen schreibt den Wert einer Stiloption.
//
// **Geprüft gegen das Paket, nicht gegen eine Liste bei uns**: eine Option, die
// der Stil nicht führt, wird abgewiesen — sonst stünde im Bestand ein Wert, den
// kein Skript liest.
func (w *Stilwaehler) OptionSetzen(ctx context.Context, setID int64, feld, wert string) error {
	kopf, stil, err := w.kopfUndStil(ctx, setID)
	if err != nil {
		return err
	}

	beschreibung := optionsbeschreibung(stil.Stiloptionen, feld)
	if beschreibung == nil {
		return domain.NewFieldError("feld", domain.SchluesselValidierung,
			"Der Stil %q kennt die Option %q nicht.", kopf.Style, feld)
	}

	wert = gespeicherteForm(*beschreibung, wert)

	// **Eine Anweisung, kein Lesen davor** — aus dem Zweitblick über Phase L.
	//
	// Hier stand ein Lesen der vollständigen Liste, eine Änderung an einem
	// Eintrag und ein Zurückschreiben von allem, samt `kopf.Game` und
	// `kopf.Style`. Zwei Optionsschreibungen lasen denselben Stand und
	// verloren einander; ein alter Lauf trug den Stil mit, den er beim Lesen
	// sah, und drehte einen Stilwechsel zurück.
	//
	// **`kopf` wird trotzdem geholt** — er beantwortet, ob es das Set gibt und
	// welcher Stil die Option kennt. Er wird nur nicht mehr zurückgeschrieben.
	return w.ablage.StiloptionSetzen(ctx, setID, feld, wert)
}

func gespeicherteForm(feld domain.StyleField, wert string) string {
	if feld.Kind != domain.KindMultipleChoice {
		return wert
	}

	return normalform(feld.Choices, wert, feld.Leerwahl)
}

// optionsbeschreibung sucht eine Stiloption im Paket.
//
// **Sie gibt die Beschreibung und nicht nur ein Ja**: wer schreibt, muss
// wissen, welcher **Art** die Option ist — eine Mehrfachwahl wird anders
// gespeichert, als sie ankommt (#637).
func optionsbeschreibung(optionen []domain.StyleField, feld string) *domain.StyleField {
	for i := range optionen {
		if optionen[i].Name == feld {
			return &optionen[i]
		}
	}

	return nil
}

func (w *Stilwaehler) OptionZuruecksetzen(ctx context.Context, setID int64, feld string) error {
	kopf, stil, err := w.kopfUndStil(ctx, setID)
	if err != nil {
		return err
	}

	if optionsbeschreibung(stil.Stiloptionen, feld) == nil {
		return domain.NewFieldError("feld", domain.SchluesselValidierung,
			"Der Stil %q kennt die Option %q nicht.", kopf.Style, feld)
	}

	// **Eine Anweisung, kein Lesen davor** — derselbe Grund wie bei
	// [Stilwaehler.OptionSetzen]. Eine Option, die gar nicht dasteht, ist kein
	// Fehler: es ist bereits der Zustand, den der Aufrufer wollte.
	return w.ablage.StiloptionEntfernen(ctx, setID, feld)
}

// kopfUndStil holt Setkopf und Stilmodell — der gemeinsame Vorlauf von
// [Stilwaehler.OptionSetzen] und [Stilwaehler.OptionZuruecksetzen].
func (w *Stilwaehler) kopfUndStil(ctx context.Context, setID int64) (domain.SetHead, msepaket.Stil, error) {
	kopf, err := w.kopf(ctx, setID)
	if err != nil {
		return domain.SetHead{}, msepaket.Stil{}, err
	}

	stil, err := w.stil(kopf.Style, kopf.Game)
	if err != nil {
		return domain.SetHead{}, msepaket.Stil{}, err
	}

	return kopf, stil, nil
}

// stiloptionAus baut eine Option samt Vorgabezustand.
//
// **Zwei Angaben, nicht eine**: [Stiloption.HatVorgabe] sagt, ob die Vorlage
// ein `default:` nennt, [Stiloption.Vorgabe], was dabei herauskommt. Ein
// Skript, das scheitert, ist vorhanden und rechnet nichts — und MSE zeigt
// seinen Eintrag trotzdem.
func stiloptionAus(
	ctx context.Context, feld domain.StyleField, styling msescript.Wert,
) Stiloption {
	vorgabe, _ := vorgabeAusAusdruck(ctx, feld.Default, styling)

	return Stiloption{
		StyleField: feld,
		Wert:       stiloptionsvorgabeMit(ctx, feld, styling),
		HatVorgabe: strings.TrimSpace(feld.Default) != "",
		Vorgabe:    vorgabe,
	}
}

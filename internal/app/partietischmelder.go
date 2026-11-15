package app

// **Was am fremden Tisch geschieht, meldet sich** — Werkstück G5 von Phase
// 5.5 (#1462).
//
// # Warum eine Meldung und keine Antwort
//
// **Weil der Gast nicht weiß, wann etwas geschieht.** Was vom Wirt kommt,
// erreicht ihn über dieselbe Naht, über die die Menüleiste meldet
// (`fensternaht.go`). Ein Fenster, das fragte, fragte die meiste Zeit umsonst
// und wäre trotzdem zu spät.
//
// # Was hier ankommt
//
// **Ein anderer Sitz zieht, und er erfährt es** — seit dem 2026-09-19:
// [service.Tischfunk] trägt die Tatsachen des Wirts an jeden Sitz
// (#1627),
// und die Zusage steht als Lauf in
// `internal/service/partieabnahmeleitung_test.go`.
//
// # Und warum drei Ereignisse und nicht eines
//
// **Weil sie Verschiedenes sind.** Das Bild sagt, wie der Tisch **aussieht**;
// der Strom, was **geschehen** ist; Chat und Präsenz gehören zu keinem von
// beiden. Wer alles durch eine Tür schickte, zwänge das Fenster, bei jeder
// Meldung erst zu unterscheiden — und irgendwann unterschiede es falsch.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// tischmelder gibt weiter, was vom Wirt kommt.
type tischmelder struct {
	app    *App
	ctx    context.Context //nolint:containedctx // der Grund steht bei [App.tischmelder]
	partie string

	// eigen heißt: der Tisch dieser Anwendung, gemeldet an das Fenster des
	// Wirts — unter [EreignisEigenesTischbild] und [EreignisEigenerZug].
	eigen bool
}

// tischmelder baut ihn für **einen** Tisch.
//
// # Warum der Kontext im Verbund steckt
//
// **Weil die Schnittstellen der Fachschicht ihn schon führen** und der
// **Meldeweg** einen anderen führt: gemeldet wird mit dem Kontext des
// Fensters, nicht mit dem des Rufs. Beide zugleich gehen nicht
// durch ein Argument; der eine kommt hier hinein, der andere durch die
// Methode. `containedctx` wird deshalb an **einer** Stelle abgeschaltet und
// nicht überall.
func (a *App) tischmelder(ctx context.Context, partie string) *tischmelder {
	return &tischmelder{app: a, ctx: ctx, partie: partie}
}

// eigenerTischmelder ist derselbe für den **eigenen** Tisch — was ein Gast
// dort tut, unter eigenem Namen (Zweitblick über Phase 6.11, zu Z1).
func (a *App) eigenerTischmelder(ctx context.Context, partie string) *tischmelder {
	return &tischmelder{app: a, ctx: ctx, partie: partie, eigen: true}
}

// Ereignis meldet, was geschehen ist.
func (m *tischmelder) Ereignis(
	_ context.Context, ereignis domain.Ereignis,
) error {
	// **Durchgereicht und nicht noch einmal gekürzt.** Was hier ankommt,
	// hat der Wirt für **diesen** Sitz projiziert (#1627); ein zweites
	// `Fuer("")` warf die eigene Ergänzung weg — ein Gast sah nie, was er
	// selbst gezogen hatte.
	m.melde(m.name(EreignisPartiezug, EreignisEigenerZug), ereignisDTO(ereignis))

	return nil
}

// Tischbild meldet, wie es jetzt aussieht und was gerade geht.
//
// **Als Partieantwort und nicht als zwei Meldungen**: das Fenster bekommt
// dieselbe Gestalt wie bei einer eigenen Partie — und zeichnet sie mit
// demselben Bauteil (§16).
func (m *tischmelder) Tischbild(
	_ context.Context, bild domain.Tischbild,
) error {
	erlaubt := bild.Erlaubt
	if erlaubt == nil {
		erlaubt = []string{}
	}

	antwort := PartieAntwort{
		Partie: partieAus(bild.Sicht), Erlaubt: erlaubt,
		Handlungen: handlungenAus(bild.Handlungen),
	}

	if bild.Absage != nil {
		antwort.Absage = &WunschabsageDTO{
			Befehl: bild.Absage.Befehl, Art: bild.Absage.Art, Grund: bild.Absage.Grund,
		}
	}

	m.melde(m.name(EreignisPartietischbild, EreignisEigenesTischbild), antwort)

	return nil
}

// WunschabsageDTO ist ein Wunsch, den der Kern des Wirts abwies (#2478).
type WunschabsageDTO struct {
	Befehl string `json:"befehl"`
	Art    string `json:"art"`
	Grund  string `json:"grund"`
}

// Nebenbei meldet Chat und Präsenz.
func (m *tischmelder) Nebenbei(
	_ context.Context, befund service.Nebenbefund,
) error {
	m.melde(EreignisPartienebenbei, nebenbeiAus(befund))

	return nil
}

// name wählt den Namen eines Ereignisses — den des Gastes oder den eigenen.
func (m *tischmelder) name(gast, eigen string) string {
	if m.eigen {
		return eigen
	}

	return gast
}

// melde schickt, wenn jemand zuhört.
//
// **Ohne Fenster geschieht nichts, und das ist kein Fehler**: an der
// Kommandozeile gibt es keines, und ein Tisch, den niemand zeichnet, ist
// trotzdem ein Tisch.
func (m *tischmelder) melde(ereignis string, daten any) {
	if m.app.melden == nil {
		return
	}

	// **Der eigene Tisch trägt die Epoche des Bestands** (#2307): nach dem
	// Zurückspielen einer älteren Sicherung derselben Partie sagt erst sie
	// dem Fenster, dass ein verspätetes Bild aus dem alten Bestand stammt.
	// Ein fremder Tisch liegt im Bestand des Wirts; ein Tausch hier ändert
	// an ihm nichts.
	if m.eigen {
		m.app.melden(m.app.fenster, ereignis, m.partie, daten, m.app.lauf.epoche.Load())

		return
	}

	m.app.melden(m.app.fenster, ereignis, m.partie, daten)
}

// NebenbeiDTO ist Chat oder Präsenz, wie das Fenster sie bekommt.
//
// # Warum flach und nicht mit offenem Inhalt
//
// **Weil das Fenster sonst raten müsste.** Der erste Anlauf reichte eine
// offene Abbildung durch; damit stand die Bedeutung jedes Feldes in keinem
// Typ, und der **Sprecher** kam obendrein aus dem falschen Feld — aus dem
// Umschlag statt aus der Nutzlast. Ein Satz von Anna, den der Wirt an Bert
// weiterreichte, stand bei Bert als Satz des Wirts. Gefunden hat es der Lauf
// über beide Enden und kein Einzelteil (§16).
//
// # Und warum alle Felder nebeneinander stehen
//
// **Weil `Art` sagt, welche etwas bedeuten.** Zwei Gestalten mit einem
// Unterscheider sind noch keine Sammlung leerer Felder; erst ab der dritten
// lohnte ein eigener Umschlag ([domain.Kontrolle]).
type NebenbeiDTO struct {
	// Art ist „chat" oder „praesenz" und sagt, was gilt.
	Art string `json:"art"`

	// Von ist der Sitz, der geredet oder gezeigt hat.
	//
	// **Nicht der, der die Nachricht geschickt hat.** Der Wirt reicht weiter,
	// und er bleibt dabei der Absender des Sprungs — wer geredet hat, setzt
	// er aus der Sitzung ([domain.Chat]).
	Von string `json:"von"`

	// Text gilt bei „chat".
	Text string `json:"text"`

	// Zeigerart, Objekt, X und Y gelten bei „praesenz".
	//
	// **X und Y sind Weltkoordinaten in Millimetern** und keine Bildpunkte —
	// zwei Fenster sind verschieden groß ([domain.Praesenz]).
	Zeigerart string  `json:"zeigerart"`
	Objekt    string  `json:"objekt"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
}

// nebenbeiAus macht aus einem Befund, was das Fenster bekommt.
//
// **Der Stand steht ausdrücklich nicht darin.** Er ist eine Zusage gegen das
// andere Ende der Leitung und in `internal/service` schon eingelöst: was hier
// ankommt, ist neu. Ein Feld, das niemand liest, wäre eine Behauptung ohne
// Verbraucher (§10).
func nebenbeiAus(befund service.Nebenbefund) NebenbeiDTO {
	if befund.Nachricht.Art == domain.ArtChat {
		return NebenbeiDTO{
			Art:  string(domain.ArtChat),
			Von:  befund.Chat.Von,
			Text: befund.Chat.Text,
		}
	}

	return NebenbeiDTO{
		Art:       string(domain.ArtPraesenz),
		Von:       befund.Praesenz.Von,
		Zeigerart: string(befund.Praesenz.Art),
		Objekt:    befund.Praesenz.Objekt,
		X:         befund.Praesenz.X,
		Y:         befund.Praesenz.Y,
	}
}

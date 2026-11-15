package service

// **Die Identität dieses Profils gegenüber anderen** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// # Zwei Quellen, eine Identität
//
// **Der private Schlüssel liegt im Schlüsselbund, die Kennung im Bestand.**
// Warum getrennt, steht bei `internal/schluesselbund` und in der Wanderung
// `0061_peeridentitaet.sql`; was hier steht, ist die Folge daraus: was gilt,
// sagt erst der **Vergleich** der beiden.
//
// # Warum das eine fachliche Entscheidung ist und keine Naht
//
// **Weil aus „der Bund hat nichts" drei verschiedene Dinge werden können** —
// je nachdem, was im Bestand steht. Das Issue verlangt ausdrücklich, dass
// Schlüsselverlust, Rotation und Profilkopie „nie **still** eine fremde
// Identität übernehmen"; wer das in einer Naht entschiede, hätte die Regel
// dort, wo sie niemand sucht (§2).
//
// # Was dieses Werk nicht tut
//
// **Es verbindet nichts.** Wer eine Verbindung aufbaut, ist N2 und N3; hier
// entsteht nur, womit sich dieses Programm ausweist.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Peerdienst ist der Name, unter dem das Geheimnis im Bund liegt.
//
// **Ein eigener Dienst und kein Konto neben den Zugangsdaten**: was hier
// liegt, ist kein fremder Zugang, sondern die eigene Identität — und wer im
// Schlüsselbund nachsieht, soll den Unterschied sehen.
//
// **Mit Punkt und ohne Leerzeichen**: `schluesselbund.pruefeNamen` lässt
// Buchstaben, Ziffern, Punkt, Strich und Unterstrich zu, und der Grund steht
// dort — Dienst und Konto reisen auf derselben Befehlszeile wie das
// Geheimnis. Der erste Anlauf hieß „MagicSetEditor4Mac Peer" und wurde
// abgewiesen; **gefunden hat es der Flächenlauf**, nicht die Abdeckung.
const Peerdienst = "MagicSetEditor4Mac.Peer"

// Peervermerke ist, was das Werk vom Bestand braucht.
type Peervermerke interface {
	Lesen(ctx context.Context) (domain.Peervermerk, error)
	Merken(ctx context.Context, kennung string, jetzt time.Time) error
	Rotieren(ctx context.Context, kennung string, jetzt time.Time) error
}

// Peergeheimnisse ist, was es vom Schlüsselbund braucht.
//
// **`Lies` gibt `ErrNotFound`, wenn nichts da ist** — nicht eine leere
// Zeichenkette: „gibt es nicht" ist eine Auskunft und kein leeres Geheimnis
// (§8).
type Peergeheimnisse interface {
	Lies(ctx context.Context, konto string) (string, error)
	Schreibe(ctx context.Context, konto, geheimnis string) error
}

// Peerlage ist, was über die Identität feststeht.
type Peerlage struct {
	Stand   domain.Peerstand
	Kennung peer.Kennung

	// Fingerabdruck steht dabei und wird nicht auf Nachfrage gerechnet.
	//
	// **Er ist das, was zwei Menschen vergleichen** — eine Fläche, die ihn
	// selbst rechnen müsste, kann ihn vergessen.
	Fingerabdruck string

	Seit time.Time

	// Vorher ist die Kennung vor der letzten Rotation — leer, wenn es keine
	// gab.
	Vorher peer.Kennung

	// VorherAbdruck ist ihr Fingerabdruck.
	//
	// **Er steht hier und wird nicht von der Naht gerechnet**: sonst kennte
	// `app` das Blatt `internal/peer`, und die Kante bestünde für eine
	// einzige Zeile. Dieselbe Überlegung wie beim Abdruck darüber.
	VorherAbdruck string

	// Befunde sind die Sätze, die erklären, was geschehen ist.
	Befunde []string
}

// Peerwerk beantwortet, wer dieses Profil gegenüber anderen ist.
type Peerwerk struct {
	vermerke Peervermerke
	bund     Peergeheimnisse
	jetzt    func() time.Time

	// zufall ist die Quelle für neue Schlüssel — `nil` heißt: die des
	// Betriebssystems.
	//
	// **Ein Feld und kein Argument**: neue Schlüssel entstehen an zwei
	// Stellen, und ein Argument an beiden wäre an einer irgendwann ein
	// anderes. Was ein Lauf damit erreicht, ist der Weg „der Zufall versiegt"
	// — ohne ihn bliebe er unbetreten und seine Meldung unbelegt.
	zufall io.Reader

	kontakte Kontaktrueckstellung
}

// MitZufall setzt die Quelle für neue Schlüssel.
func (w *Peerwerk) MitZufall(zufall io.Reader) *Peerwerk {
	w.zufall = zufall

	return w
}

// Kontaktrueckstellung ist, was nach einer Rotation zu tun ist.
//
// **Eine Schnittstelle und kein [Kontaktwerk]**: was hier gebraucht wird, ist
// ein Satz und keine Fläche — und ein Werk, das ein anderes ganz kennt, ist
// die Naht, an der später beide hängen.
type Kontaktrueckstellung interface {
	NachRotation(ctx context.Context) error
}

// NeuesPeerwerk baut es über Bestand, Bund und die Kontakte.
//
// **Die Kontakte sind kein Zusatz, sondern Teil der Zusage.** Eine Rotation
// ohne sie ließe jede Bestätigung stehen, die einer Identität gilt, die es
// nicht mehr gibt — und ein Argument, das man weglassen kann, lässt irgendwer
// weg.
func NeuesPeerwerk(
	vermerke Peervermerke, bund Peergeheimnisse,
	kontakte Kontaktrueckstellung, jetzt func() time.Time,
) *Peerwerk {
	return &Peerwerk{
		vermerke: vermerke, bund: bund, kontakte: kontakte, jetzt: jetzt,
	}
}

// Lage sagt, wie es steht — **und legt nichts an**.
//
// **Lesen und Anlegen sind zwei Rufe**, weil sie zwei Entscheidungen sind:
// „wie steht es" darf jede Fläche fragen, „lege eine an" ist eine Handlung
// (0044).
func (w *Peerwerk) Lage(ctx context.Context) (Peerlage, error) {
	vermerk, fehler := w.vermerke.Lesen(ctx)
	if fehler != nil {
		return Peerlage{}, fehler
	}

	geheimnis, bundfehler := w.bund.Lies(ctx, peerkonto)

	switch {
	case bundfehler != nil && !errors.Is(bundfehler, domain.ErrNotFound):
		return Peerlage{}, bundfehler

	case bundfehler != nil && !vermerk.Gibts():
		return Peerlage{Stand: domain.PeerstandOhne}, nil

	case bundfehler != nil:
		// **Der Vermerk steht, der Schlüssel fehlt.** Die verlorene Kennung
		// geht mit: ohne sie könnte niemand einem Kontakt sagen, welche
		// Identität fort ist.
		//
		// **`nil` als Fehler, obwohl einer kam** — und das ist die Aussage
		// dieses Zweigs: „im Bund liegt nichts" ist eine **Auskunft** und
		// kein Ausfall (§8). Sie steht im Stand, nicht im Fehler; wer sie
		// durchreichte, machte aus einem Verlust einen Absturz.
		//nolint:nilerr // der Fehler ist der Stand, und er steht in der Antwort
		return Peerlage{
			Stand: domain.PeerstandVerloren, Kennung: peer.Kennung(vermerk.Kennung),
			Fingerabdruck: peer.Fingerabdruck(peer.Kennung(vermerk.Kennung)),
			Seit:          vermerk.Seit, Vorher: peer.Kennung(vermerk.Vorher),
			VorherAbdruck: peer.Fingerabdruck(peer.Kennung(vermerk.Vorher)),
			Befunde: []string{
				"Der Schlüssel zu dieser Peerkennung liegt nicht im " +
					"Schlüsselbund dieses Rechners. Eine kopierte " +
					"Bestandsdatei bringt die Kennung mit, den Schlüssel " +
					"nicht — wer hier weitermachen will, legt mit einer " +
					"Rotation ausdrücklich eine neue Identität an.",
			},
		}, nil
	}

	ich, fehler := peer.AusGeheimnis(geheimnis)
	if fehler != nil {
		return Peerlage{}, fehler
	}

	return w.gegenDenVermerk(ctx, ich, vermerk)
}

// gegenDenVermerk hält den vorhandenen Schlüssel gegen den Bestand.
func (w *Peerwerk) gegenDenVermerk(
	ctx context.Context, ich peer.Identitaet, vermerk domain.Peervermerk,
) (Peerlage, error) {
	lage := Peerlage{
		Stand: domain.PeerstandBereit, Kennung: ich.Kennung(),
		Fingerabdruck: ich.Fingerabdruck(), Seit: vermerk.Seit,
		Vorher:        peer.Kennung(vermerk.Vorher),
		VorherAbdruck: peer.Fingerabdruck(peer.Kennung(vermerk.Vorher)),
	}

	switch {
	case !vermerk.Gibts():
		// **Der Schlüssel steht, der Vermerk fehlt** — der Bestand kann
		// zurückgesetzt worden sein. Vermerkt wird nachträglich, und **es
		// wird gesagt**: sonst merkte niemand, dass der Bestand jünger ist
		// als die Identität.
		if fehler := w.vermerke.Merken(
			ctx, string(ich.Kennung()), w.jetzt()); fehler != nil {
			return Peerlage{}, fehler
		}

		lage.Seit = w.jetzt()
		lage.Befunde = []string{
			"Der Schlüssel dieses Profils lag im Schlüsselbund, im Bestand " +
				"stand keine Kennung dazu — sie ist jetzt vermerkt. Der " +
				"Bestand ist damit jünger als die Identität.",
		}

	case vermerk.Kennung != string(ich.Kennung()):
		lage.Stand = domain.PeerstandFremd
		lage.Kennung = peer.Kennung(vermerk.Kennung)
		lage.Fingerabdruck = peer.Fingerabdruck(lage.Kennung)
		lage.Befunde = []string{
			"Im Schlüsselbund liegt ein anderer Schlüssel als der, den der " +
				"Bestand vermerkt (" + ich.Fingerabdruck() + " gegen " +
				lage.Fingerabdruck + "). Welcher gilt, entscheidet kein " +
				"Programm — wer weitermachen will, legt mit einer Rotation " +
				"ausdrücklich eine neue Identität an.",
		}
	}

	return lage, nil
}

// Sicherstellen legt eine Identität an, **wenn es noch keine gibt**.
//
// **Und nur dann.** Bei Verlust und Widerspruch ist es ein Konflikt: eine
// neue Identität dort anzulegen wäre genau das stille Übernehmen, das das
// Issue ausschließt.
func (w *Peerwerk) Sicherstellen(ctx context.Context) (Peerlage, error) {
	lage, fehler := w.Lage(ctx)
	if fehler != nil {
		return Peerlage{}, fehler
	}

	switch lage.Stand {
	case domain.PeerstandBereit:
		return lage, nil

	case domain.PeerstandVerloren, domain.PeerstandFremd:
		return Peerlage{}, fmt.Errorf("%w: %s", domain.ErrConflict, lage.Befunde[0])
	}

	return w.anlegen(ctx, domain.Peervermerk{}, false)
}

// Rotieren legt **ausdrücklich** eine neue Identität an.
//
// **Auch aus einem Verlust oder einem Widerspruch heraus** — das ist der Weg
// hinaus, und er ist eine Handlung und keine Reparatur: danach kennt jeder
// Kontakt eine Kennung, die nicht mehr gilt.
func (w *Peerwerk) Rotieren(ctx context.Context) (Peerlage, error) {
	vermerk, fehler := w.vermerke.Lesen(ctx)
	if fehler != nil {
		return Peerlage{}, fehler
	}

	// **Die Kontakte zuerst, der Schlüssel danach.**
	//
	// Scheitert das Zurückstellen, ist nichts geschehen. Scheiterte
	// stattdessen die Rotation **nach** dem Zurückstellen, stünde eine
	// unveränderte Identität mit lauter unbestätigten Kontakten da — lästig,
	// und jederzeit von Hand zu heilen. Andersherum stünde eine **neue**
	// Identität mit lauter Bestätigungen da, die einer alten gelten; das
	// heilt niemand, weil es niemand sieht.
	if fehler := w.kontakte.NachRotation(ctx); fehler != nil {
		return Peerlage{}, fehler
	}

	return w.anlegen(ctx, vermerk, true)
}

// anlegen erzeugt ein Paar, legt es ab und vermerkt es.
//
// **Erst der Bund, dann der Vermerk.** Scheitert der Vermerk, liegt ein
// Schlüssel ohne Kennung — und den findet [Peerwerk.Lage] wieder und trägt
// ihn nach. Andersherum stünde eine Kennung ohne Schlüssel da, und das ist
// ein Verlust, den niemand verursacht hat.
func (w *Peerwerk) anlegen(
	ctx context.Context, vorher domain.Peervermerk, rotiert bool,
) (Peerlage, error) {
	ich, fehler := peer.Erzeugen(w.zufall)
	if fehler != nil {
		return Peerlage{}, fehler
	}

	// **Ein eigener Satz für ein gescheitertes Ablegen** (S13.NAT-03,
	// #1706) — „der Schlüssel fehlt" wäre hier falsch: er war da und ließ
	// sich nicht speichern.
	if fehler := w.bund.Schreibe(ctx, peerkonto, ich.Geheimnis()); fehler != nil {
		return Peerlage{}, domain.MitGrund(domain.ErrSchluesselbundAblage, fehler,
			"den Schlüssel dieses Profils ablegen")
	}

	jetzt := w.jetzt()

	merken := w.vermerke.Merken
	if rotiert {
		merken = w.vermerke.Rotieren
	}

	if fehler := merken(ctx, string(ich.Kennung()), jetzt); fehler != nil {
		return Peerlage{}, fehler
	}

	lage := Peerlage{
		Stand: domain.PeerstandBereit, Kennung: ich.Kennung(),
		Fingerabdruck: ich.Fingerabdruck(), Seit: jetzt,
	}

	if rotiert && vorher.Gibts() {
		lage.Vorher = peer.Kennung(vorher.Kennung)
		lage.VorherAbdruck = peer.Fingerabdruck(lage.Vorher)
		lage.Befunde = []string{
			"Diese Identität ist neu. Wer " +
				peer.Fingerabdruck(peer.Kennung(vorher.Kennung)) +
				" bestätigt hat, muss den Fingerabdruck erneut prüfen — " +
				"eine alte Bestätigung gilt für den neuen Schlüssel nicht.",
		}
	}

	return lage, nil
}

// peerkonto ist das Konto im Schlüsselbund.
//
// **Ein fester Name und keine Profilkennung darin**: der Bund liegt je
// Betriebssystemnutzer, und das Profil steht schon im Bestand. Eine Kennung
// im Kontonamen hieße, dass eine kopierte Bestandsdatei mit einer anderen
// Profilnummer denselben Schlüssel **nicht** fände — und das ist genau der
// Fall, den [domain.PeerstandVerloren] benennen soll.
const peerkonto = "peeridentitaet"

// Ausweis gibt die Identität, mit der dieses Profil unterschreibt.
//
// # Warum er nicht in [Peerlage] steht
//
// **Weil die Lage an eine Fläche geht und der Ausweis nicht.** `Peerlage`
// trägt Kennung und Fingerabdruck und reist bis ins Fenster; der **private**
// Schlüssel bleibt in dieser Schicht. Ein Feld daneben wäre die Stelle, an der
// er einmal in einer Antwort landet — und die Abnahme von N1 verlangt
// ausdrücklich, dass das nie geschieht.
//
// # Und warum „gibt es nicht" hier ein Fehler ist
//
// **Weil der Aufrufer unterschreiben will.** Für [Peerwerk.Lage] ist „es gibt
// noch keine" eine Auskunft; wer einen Ausweis verlangt, kann mit ihr nichts
// anfangen — und ein leerer Ausweis unterschriebe mit einem Nullschlüssel.
//
// **Das gilt auch für `verloren` und `fremd`**: dort steht eine Kennung, zu
// der der Schlüssel fehlt oder nicht passt. Daraus einen Ausweis zu machen
// hieße, still eine andere Identität anzunehmen.
func (w *Peerwerk) Ausweis(ctx context.Context) (peer.Identitaet, error) {
	lage, fehler := w.Lage(ctx)
	if fehler != nil {
		return peer.Identitaet{}, fehler
	}

	if lage.Stand != domain.PeerstandBereit {
		return peer.Identitaet{}, fmt.Errorf(
			"%w (%s)", domain.ErrOhnePeeridentitaet, lage.Stand)
	}

	geheimnis, fehler := w.bund.Lies(ctx, peerkonto)
	if fehler != nil {
		return peer.Identitaet{}, fehler
	}

	return peer.AusGeheimnis(geheimnis)
}

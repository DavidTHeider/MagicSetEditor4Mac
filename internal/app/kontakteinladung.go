package app

// **Einladen und annehmen an der Naht** — NET-01
// (#1705).
//
// # Woher die Adresse in der Einladung kommt
//
// **Aus dem, worauf dieses Programm gerade horcht** — und in dieser
// Reihenfolge: eine offene Einladung über fremde Netze zuerst, weil sie
// überall trägt, dann die erste Adresse im eigenen Netz. Horcht nichts, bleibt
// sie leer; die Einladung gilt trotzdem, denn sie sagt vor allem, **wer** ich
// bin, und erst dann, wo.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Kontakteinladung ist eine ausgestellte Einladung an der Naht.
type Kontakteinladung struct {
	Code string `json:"code"`

	// Ablauf ist RFC 3339.
	Ablauf string `json:"ablauf"`

	// Fingerabdruck ist der eigene — für die technische Angabe.
	Fingerabdruck string `json:"fingerabdruck"`
}

// KontakteinladungAntwort ist die Einladung samt ihrem Fehler.
type KontakteinladungAntwort struct {
	Einladung Kontakteinladung `json:"einladung"`
	Fehler    *Fehler          `json:"fehler"`
}

// Einladungsvorschau ist, was eine Einladung sagt, bevor jemand annimmt.
type Einladungsvorschau struct {
	Kennung        string `json:"kennung"`
	Name           string `json:"name"`
	Zweck          string `json:"zweck"`
	Adresse        string `json:"adresse"`
	Ablauf         string `json:"ablauf"`
	Fingerabdruck  string `json:"fingerabdruck"`
	Vergleichscode string `json:"vergleichscode"`

	// Lage ist die eines vorhandenen Kontakts — `unbekannt`, wenn es keinen
	// gibt.
	Lage string `json:"lage"`

	// Vorgaenger ist der Name des Kontakts, dessen neuer Schlüssel diese
	// Einladung sein will — leer, wenn es ihn hier nicht gibt.
	Vorgaenger string `json:"vorgaenger"`

	Angenommen bool `json:"angenommen"`

	// Einloesbar sagt, dass sich dieses Profil nach der Annahme beim
	// Einladenden melden wird (NET-01b) — dann fragt das Fenster nach dem
	// Namen, der hinausgeht.
	Einloesbar bool `json:"einloesbar"`
}

// EinladungsvorschauAntwort ist die Vorschau samt ihrem Fehler.
type EinladungsvorschauAntwort struct {
	Vorschau Einladungsvorschau `json:"vorschau"`
	Fehler   *Fehler            `json:"fehler"`
}

// KontakteinladungErstellen stellt eine Einladung dieses Profils aus.
func (a *App) KontakteinladungErstellen(
	ctx context.Context, name, zweck string,
) KontakteinladungAntwort {
	ctx = a.imProfil(ctx)

	return a.KontakteinladungErstellenUnter(ctx, name, zweck, "")
}

// KontakteinladungErstellenUnter stellt eine Einladung mit einer genannten
// Adresse aus — leer heißt: die, unter der diese Anwendung horcht.
//
// **Für Kommandozeile und Dienst** (NET-01b.AK-04): dort horcht ein
// **anderer** Prozess (`mse horchen`), und wer einlösbar einladen will,
// nennt, wo er steht. Beantwortet wird die Einlösung dort; der Vermerk der
// Nonce liegt im selben Bestand.
func (a *App) KontakteinladungErstellenUnter(
	ctx context.Context, name, zweck, adresse string,
) KontakteinladungAntwort {
	ctx = a.imProfil(ctx)

	if adresse == "" {
		adresse = a.einladungsadresse()
	}

	ausgestellt, fehler := a.einladungswerk().Erstellen(ctx, service.Einladungsbitte{
		Name: name, Zweck: domain.Zweck(zweck), Adresse: adresse,
	})
	if fehler != nil {
		return KontakteinladungAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return KontakteinladungAntwort{Einladung: Kontakteinladung{
		Code: ausgestellt.Code, Ablauf: alsZeitpunkt(ausgestellt.Ablauf),
		Fingerabdruck: ausgestellt.Fingerabdruck,
	}}
}

// KontakteinladungPruefen sagt, was eine Einladung ist — **rein lesend**.
func (a *App) KontakteinladungPruefen(
	ctx context.Context, code string,
) EinladungsvorschauAntwort {
	ctx = a.imProfil(ctx)

	vorschau, fehler := a.einladungswerk().Pruefen(ctx, code)
	if fehler != nil {
		return EinladungsvorschauAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return EinladungsvorschauAntwort{Vorschau: alsVorschau(vorschau)}
}

// KontakteinladungAnnehmen nimmt eine Einladung an — ganz oder gar nicht.
//
// **Die Liste danach**, wie bei jedem Schritt an einem Kontakt
// ([App.amKontakt]). **Und danach meldet es sich beim Einladenden**
// (NET-01b), wenn die Einladung das zulässt — unter `meinName`, den ein
// Mensch gewählt hat; was dabei herauskommt, steht in
// [PeerkontakteAntwort.Gemeldet].
func (a *App) KontakteinladungAnnehmen(
	ctx context.Context, code, eigenerName, verglichen string, freigeben bool,
	meinName string,
) PeerkontakteAntwort {
	ctx = a.imProfil(ctx)

	var gemeldet *Einloesemeldung

	antwort := a.amKontakt(ctx, func(ctx context.Context) error {
		werk := a.einladungswerk()

		if _, fehler := werk.Annehmen(ctx, service.Einladungsannahme{
			Code: code, EigenerName: eigenerName, Verglichen: verglichen,
			Freigeben: freigeben, MeinName: meinName,
		}); fehler != nil {
			return fehler
		}

		gemeldet = alsMeldung(ctx, werk.Melden(ctx, a.einloesebote(), code,
			domain.Einloesung{
				Adresse: a.einladungsadresse(), Name: domain.BrauchbarerName(meinName),
			}))

		return nil
	})
	antwort.Gemeldet = gemeldet

	return antwort
}

// einladungswerk baut das Werk für diesen Ruf.
//
// **Mit dem Riegel vor der Meldung** (#2000): eine einlösbare Einladung
// wählt die Adresse, die ihr Aussteller hineingeschrieben hat.
func (a *App) einladungswerk() *service.Einladungswerk {
	werk := a.peerwerk()

	return service.NeuesEinladungswerk(
		werk, werk, kontaktnaht{holen: a.laufendeVerbindung}, a.uhr).
		MitGrenze(a.peerlage.einloesegrenze).
		MitRiegel(a.senderiegel())
}

// senderiegel gibt den Prüfer, den eine Meldung fragt — eine Autorisierung
// über dem Bestand dieses Rufs.
//
// **Über dem Bestand und nicht die des Fensters**, weil Kommandozeile und
// Dienst keine eigene setzen: sie bauen ihre Anwendung über
// [PeerflaecheUeber], und dort steht nur der Bestand. Dieselbe Verdrahtung
// für jede Wurzel (§16). **Ohne Bestand fragt niemand** — dann gibt es auch
// nichts anzunehmen: das Werk scheitert, bevor es fragen könnte.
func (a *App) senderiegel() service.Pruefer {
	db := a.laufendeVerbindung()
	if db == nil {
		return nil
	}

	return AutorisierungAus(db)
}

// einladungsadresse gibt, wo dieses Programm gerade erreichbar ist.
func (a *App) einladungsadresse() string {
	if fern := a.fernstand(); fern.Einladung != "" {
		return fern.Einladung
	}

	if horchen := a.horchstand(); len(horchen.ErreichbarAdressen) > 0 {
		return horchen.ErreichbarAdressen[0]
	}

	return ""
}

// alsVorschau macht aus der Vorschau des Dienstes ihr DTO.
func alsVorschau(vorschau service.Einladungsvorschau) Einladungsvorschau {
	einladung := vorschau.Einladung

	dto := Einladungsvorschau{
		Kennung: einladung.Kennung, Name: einladung.Name,
		Zweck: string(einladung.Zweck), Adresse: einladung.Adresse,
		Ablauf:         alsZeitpunkt(einladung.Ablauf),
		Fingerabdruck:  vorschau.Fingerabdruck,
		Vergleichscode: vorschau.Vergleichscode,
		Lage:           string(vorschau.Lage), Angenommen: vorschau.Angenommen,
		Einloesbar: einladung.Einloesbar,
	}

	if vorgaenger := vorschau.Vorgaenger; vorgaenger != nil {
		dto.Vorgaenger = vorgaenger.EigenerName
		if dto.Vorgaenger == "" {
			dto.Vorgaenger = vorgaenger.SelbstgewaehlterName
		}
	}

	return dto
}

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Grenzen dieses Prozesses** — Werkstück O5 von Phase 5.2
// (#1429).
//
// Was ein Ruf kostet, wie viel ein Maschinenzugang je Minute darf, wie viele
// Arbeiten einer Klasse zugleich laufen, wie groß Rumpf und Antwort sein
// dürfen: **eine** Auskunft, dieselbe am Fenster, über MCP, auf der
// Kommandozeile und im Dienst (§16). Gerechnet wird sie aus dem
// Maschinenhaushalt, nicht aus einer zweiten Tabelle — und seit QA-08.AK-03
// mit dem Schalter aus dem Bestand (`grenzen.kontingente`).

// KontingentDTO ist ein Eimer, wie das Fenster ihn nennt.
type KontingentDTO struct {
	Burst           int `json:"burst"`
	Nachfuellen     int `json:"nachfuellen"`
	FensterSekunden int `json:"fensterSekunden"`
}

// GrenzenDTO ist die Auskunft.
type GrenzenDTO struct {
	Parallel     map[string]int `json:"parallel"`
	AntwortBytes int64          `json:"antwortBytes"`
	Principal    KontingentDTO  `json:"principal"`
	Global       KontingentDTO  `json:"global"`
	Reserve      int            `json:"reserve"`
	Kosten       map[string]int `json:"kosten"`

	// Kontingente ist `an` oder `aus` — ob die Guthaben gelten
	// (QA-08.AK-03, [0161](../../docs/entscheidungen/0161-die-kontingente-lassen-sich-abschalten.md)).
	Kontingente string `json:"kontingente"`

	// Verbindungen und Auftraege sind die Grenzen der Dienste über HTTP —
	// bis #2003 stand keine davon in einer Auskunft.
	Verbindungen int                `json:"verbindungen"`
	Auftraege    AuftragsgrenzenDTO `json:"auftraege"`

	// Kostentafel nennt je Fähigkeit, die nicht kostet, was ihre Art kostet,
	// ihre Klasse und ihre Wege (#2003, `domain.Kostentafel`) — auch die
	// lesenden `POST`-Wege, die Lesen kosten (#2020).
	Kostentafel []domain.Kostenauskunft `json:"kostentafel"`
}

// AuftragsgrenzenDTO sind die Grenzen des Auftragsbuchs im Dienst.
type AuftragsgrenzenDTO struct {
	Laufend              int `json:"laufend"`
	Behalten             int `json:"behalten"`
	AufbewahrungSekunden int `json:"aufbewahrungSekunden"`
}

// GrenzenAntwort ist der Umschlag fürs Fenster.
type GrenzenAntwort struct {
	Grenzen GrenzenDTO `json:"grenzen"`
	Fehler  *Fehler    `json:"fehler"`
}

// GrenzenauskunftVon rechnet Grenzen in die Auskunft um — dieselbe an jeder
// Fläche.
func GrenzenauskunftVon(grenzen domain.Haushaltsgrenzen) GrenzenDTO {
	parallel := make(map[string]int, len(grenzen.Parallel))
	for klasse, anzahl := range grenzen.Parallel {
		parallel[string(klasse)] = anzahl
	}

	kosten := make(map[string]int, len(domain.Kostenklassen()))
	for _, klasse := range domain.Kostenklassen() {
		kosten[string(klasse)] = klasse.Kosten()
	}

	return GrenzenDTO{
		Parallel:     parallel,
		AntwortBytes: grenzen.Antwort,
		Principal:    kontingentDTO(grenzen.Principal),
		Global:       kontingentDTO(grenzen.Global),
		Reserve:      grenzen.Reserve,
		Kosten:       kosten,
		Kontingente:  grenzen.Kontingentstand(),
		Verbindungen: grenzen.Verbindungen,
		Auftraege: AuftragsgrenzenDTO{
			Laufend: grenzen.Auftraege.Laufend, Behalten: grenzen.Auftraege.Behalten,
			AufbewahrungSekunden: int(grenzen.Auftraege.Aufbewahrung.Seconds()),
		},
		Kostentafel: domain.KostentafelAlsAuskunft(),
	}
}

func kontingentDTO(kontingent domain.Kontingent) KontingentDTO {
	return KontingentDTO{
		Burst:           kontingent.Burst,
		Nachfuellen:     kontingent.Nachfuellen,
		FensterSekunden: int(kontingent.Fenster.Seconds()),
	}
}

// Grenzen nennt die Grenzen, die ein jetzt gestarteter Dienst über diesem
// Bestand anwendete — das Fenster selbst bucht kein Guthaben. **Ohne Bestand
// die Vorgabe**; ein Lesefehler steht im Umschlag, neben der Vorgabe.
func (a *App) Grenzen(ctx context.Context) GrenzenAntwort {
	ctx = a.imProfil(ctx)

	grenzen := service.Maschinenhaushalt().Grenzen()
	if a.einstellungen == nil {
		return GrenzenAntwort{Grenzen: GrenzenauskunftVon(grenzen)}
	}

	werte, fehler := a.einstellungen.Alle(ctx)
	if fehler != nil {
		return GrenzenAntwort{Grenzen: GrenzenauskunftVon(grenzen), Fehler: Uebersetze(ctx, fehler)}
	}

	return GrenzenAntwort{Grenzen: GrenzenauskunftVon(grenzen.NachEinstellungen(werte))}
}

// Grenzen ist die Fensterbindung.
func (b *Bindungen) Grenzen(ctx context.Context) GrenzenAntwort {
	return b.app(ctx).Grenzen(b.ruf(ctx))
}

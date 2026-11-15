package app

// **Wählen nach einer Adresse von Hand** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum dieser Wähler hier steht und nicht in `service`
//
// **Weil er zwei Welten zusammenhängt.** [service.Waehler] ist eine
// Schnittstelle, `internal/leitung` ein Blatt — und dass ein `*leitung.Leitung`
// als [service.Verbindung] durchgeht, darf keine der beiden wissen. Das
// Zusammenhängen ist die Aufgabe des Composition Root (§2).
//
// # Was er nicht tut
//
// **Er entscheidet nichts.** Ob dieses Gegenüber überhaupt gewählt werden
// darf, hat [service.Peersitzungen.Oeffne] vorher gefragt; wer am anderen Ende
// sitzt, sagt danach der Handschlag. Dieser Wähler schlägt eine Adresse nach
// und baut eine Leitung.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// adresswaehler wählt die Adresse, die ein Mensch für dieses Gegenüber
// eingetragen hat.
type adresswaehler struct{ app *App }

// Waehle baut eine gesicherte Leitung zu einem Kontakt.
//
// # Warum das Protokoll hier nicht vorkommt
//
// **Weil dieser Transport eine Verbindung je Zweck führt.** Das Protokoll
// gehört in die Hülle des Handschlags, und dorthin trägt es
// `internal/peer` selbst; ein libp2p-Adapter hätte es zur Aushandlung
// gebraucht, und tailcat könnte daraus einen Port machen. Die Naht behält es
// deshalb — **dieser** Wähler benutzt es nicht, und das steht hier, damit
// niemand es für tragend hält.
func (w adresswaehler) Waehle(
	ctx context.Context, kennung peer.Kennung, _ string,
) (service.Verbindung, error) {
	adresse, fehler := w.app.kontaktwerk().Adresse(ctx, kennung)
	if fehler != nil {
		return nil, fehler
	}

	// **Der Ausweis vor der Leitung**, aus demselben Grund wie in
	// [service.Peersitzungen.Oeffne]: ohne ihn ließe sich nichts
	// unterschreiben, und der Griff in den Schlüsselbund kann nachfragen.
	ich, fehler := w.app.peerwerk().Ausweis(ctx)
	if fehler != nil {
		return nil, fehler
	}

	// **Zwei Wege, ein Feld** — N3
	// (#1456).
	//
	// Im Adressfeld eines Kontakts steht entweder eine Adresse im eigenen
	// Netz oder eine **Einladung** über ein fremdes. Beides beantwortet
	// dieselbe Frage — *wo erreiche ich dieses Gegenüber?* —, und zwei
	// Felder daneben wären zwei Antworten auf eine.
	//
	// **Die erwartete Kennung ist beide Male dieselbe.** Eine Adresse sagt
	// nicht, wer an ihr sitzt, und eine Einladung auch nicht: wer eine
	// abfängt, bekommt eine Leitung und keine Identität. Das entscheidet
	// der Handschlag darüber.
	if domain.IstEinladung(adresse) {
		return w.app.fernwaehler()(
			ctx, domain.Einladung(adresse), ich, kennung)
	}

	return leitung.Waehle(ctx, adresse, ich, kennung)
}

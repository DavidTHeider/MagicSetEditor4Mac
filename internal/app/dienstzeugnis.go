package app

// **Welches Zeugnis ein Dienst vorzeigt** — NET-03
// (#1707),
// entschieden in
// [0154](../../docs/entscheidungen/0154-das-dienstzeugnis-ist-nicht-die-identitaet.md).
//
// **An einer Stelle für beide Dienste.** Bis hierhin baute `mse-rest` sein
// Zeugnis in `cmd/mse-rest/zeugnis.go` und `mse-mcp` dasselbe noch einmal in
// `ueberhttp.go` — zwei Fassungen einer Wahl, die jetzt drei Wege hat.

import (
	"context"
	"crypto/tls"
	"flag"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

// Dienstzeugnisangabe beschreibt das Zeugnis, das ein Dienst vorzeigt.
type Dienstzeugnisangabe struct {
	// Fingerabdruck ist der des vorgezeigten Zeugnisses — beim eigenen in
	// der Form von `openssl x509 -fingerprint -sha256`.
	Fingerabdruck string

	// GueltigBis, Aussteller und Namen kennt nur ein eigenes Zeugnis; beim
	// Zeugnis auf der Peeridentität bleiben sie leer.
	GueltigBis time.Time
	Aussteller string
	Namen      []string
}

// Dienstzeugnis wählt das Zeugnis eines Dienstes jenseits der Schleife.
//
// **Eine Flagge schlägt die Einstellung, Feld für Feld.** Nennen beide
// zusammen Zertifikat und Schlüssel, gilt das eigene Zeugnis; nennen sie
// nichts, das auf der Peeridentität. **Alles dazwischen startet nichts** —
// ein stiller Rückfall gäbe dem, der eine PKI eingetragen hat, ein Zeugnis,
// das seine Klienten abweisen.
//
// **melde** bekommt die Befunde eines misslungenen Nachladens.
func (a *App) Dienstzeugnis(
	ctx context.Context, adressen []string, flaggen leitung.Zeugnisdateien,
	melde func(error),
) (*tls.Config, Dienstzeugnisangabe, error) {
	ctx = a.imProfil(ctx)

	stand, fehler := a.einstellungen.Stand(ctx)
	if fehler != nil {
		return nil, Dienstzeugnisangabe{}, fehler
	}

	dateien := leitung.Zeugnisdateien{
		Zertifikat: ersteVon(flaggen.Zertifikat, stand[domain.DienstTLSZertifikat].Wirksam),
		Schluessel: ersteVon(flaggen.Schluessel, stand[domain.DienstTLSSchluessel].Wirksam),
		Kette:      ersteVon(flaggen.Kette, stand[domain.DienstTLSKette].Wirksam),
		Anker:      ersteVon(flaggen.Anker, stand[domain.DienstTLSAnker].Wirksam),
	}

	if dateien.Eigene() {
		// **Jede Adresse muss es tragen** (G08, #2272), beim Start wie bei
		// jedem Austausch — sonst läge auf einer ein Zeugnis, das die
		// Gegenseite abweist. Das Werk prüft sie alle (Befund 2 des
		// Zweitblicks über Phase G).
		werk, werkfehler := leitung.NeuesZeugniswerk(dateien, adressen, time.Now, melde)
		if werkfehler != nil {
			return nil, Dienstzeugnisangabe{}, werkfehler
		}

		befund := werk.Befund()

		return werk.Einstellung(), Dienstzeugnisangabe{
			Fingerabdruck: befund.Fingerabdruck, GueltigBis: befund.GueltigBis,
			Aussteller: befund.Aussteller, Namen: befund.Namen,
		}, nil
	}

	if dateien != (leitung.Zeugnisdateien{}) {
		return nil, Dienstzeugnisangabe{}, domain.MitGrund(domain.ErrValidation, nil,
			"ein eigenes Zeugnis braucht Zertifikat und Schlüssel — "+
				"trag beide ein oder leere auch Kette und Anker")
	}

	ich, fehler := a.Ausweis(ctx)
	if fehler != nil {
		return nil, Dienstzeugnisangabe{}, fehler
	}

	bauen := a.peerlage.serverzeugnis
	if bauen == nil {
		bauen = leitung.Serverzeugnis
	}

	zeugnis, fehler := bauen(ich)
	if fehler != nil {
		return nil, Dienstzeugnisangabe{}, fehler
	}

	return zeugnis, Dienstzeugnisangabe{Fingerabdruck: ich.Fingerabdruck()}, nil
}

// ersteVon gibt den ersten nicht leeren Wert.
func ersteVon(werte ...string) string {
	for _, wert := range werte {
		if wert != "" {
			return wert
		}
	}

	return ""
}

// Zeugnisflaggen hängt die vier Pfade an die Flaggen eines Dienstes und gibt
// ihren Leser — **dieselben vier für `mse-rest` und `mse-mcp`**.
//
// **Eine Flagge schlägt die Einstellung, Feld für Feld** ([App.Dienstzeugnis]):
// wer von Hand startet, will nicht erst die Einstellungen ändern.
func Zeugnisflaggen(fahne *flag.FlagSet) func() leitung.Zeugnisdateien {
	zertifikat := fahne.String("zeugnis", "", "das eigene Zertifikat (PEM); sonst "+domain.DienstTLSZertifikat)
	schluessel := fahne.String("schluessel", "", "sein Schlüssel (PEM, 0600); sonst "+domain.DienstTLSSchluessel)
	kette := fahne.String("kette", "", "Zwischenzertifikate (PEM); sonst "+domain.DienstTLSKette)
	anker := fahne.String("anker", "", "Vertrauensanker der Kette (PEM); sonst "+domain.DienstTLSAnker)

	return func() leitung.Zeugnisdateien {
		return leitung.Zeugnisdateien{
			Zertifikat: *zertifikat, Schluessel: *schluessel, Kette: *kette, Anker: *anker,
		}
	}
}

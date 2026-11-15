// **Wo ein Handel steht, über MCP** —
// #1637.
//
// # Ein lesendes Werkzeug und kein zweites
//
// **Ein Modell sieht nach und handelt nicht.** Eröffnen, anbieten, zustimmen,
// vorbereiten und übergeben binden — sie sind Zusagen gegenüber einem anderen
// Menschen, und die gibt keiner in fremdem Namen. Verbuchen und Bestätigen
// sind darüber hinaus über die **Reichweite eines Tokens** ausgeschlossen:
// sie verlangen [domain.WirkungVerbindlich], und ein Token reicht bis
// [domain.WirkungSchreiben]
// ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md), T10).
//
// **Das ist eine Entscheidung und keine Lücke** (§16): dieselbe Fähigkeit
// überall, oder es steht begründet, warum nicht. Hier steht es.
//
// # Und warum Lesen trotzdem angeboten wird
//
// **Weil ein Modell dabei helfen soll, einen Handel zu verstehen.** Was
// angeboten ist, gegen welchen Abdruck jemand zugestimmt hat und wo der
// Vorgang steht — das ist Auskunft über den **eigenen** Bestand, und sie
// verlässt den Rechner nicht.

package mcp

import (
	"context"
	"encoding/json"
)

// Handelsauskunft ist die Naht zum Stand eines Handels.
type Handelsauskunft interface {
	Handelsstand(ctx context.Context, vorgang string) (Handelsstand, error)

	// Handelszusagen sagt, welche Besitzzeilen für einen Vorgang
	// festgehalten sind (#1643).
	Handelszusagen(ctx context.Context, vorgang string) ([]Handelszusage, error)

	// Handelsvorgaenge nennt die Handel mit einem Gegenüber, je Vorgang die
	// letzte Revision (UX-08.AK-04, #1680).
	Handelsvorgaenge(ctx context.Context, gegenueber string) ([]Handelsstand, error)

	// Handelsuebersicht nennt die Vorgänge dieses Profils, gefiltert
	// (TRADE-01, #1715).
	Handelsuebersicht(ctx context.Context, filter Handelsfilter) ([]Handelsuebersichtszeile, error)

	// Handelsprotokoll nennt die Zustandswechsel eines Vorgangs (TRADE-03,
	// #1717).
	Handelsprotokoll(ctx context.Context, vorgang string) ([]Handelsprotokollzeile, error)

	// Handelsanhalt gibt, was die Seiten ungefähr wert sind — in der
	// Gestalt der App-Naht (TRADE-02b,
	// #1951).
	Handelsanhalt(ctx context.Context, vorgang string) (any, error)

	// Handelsbeleg gibt Stand und Beleg — in der Gestalt der App-Naht
	// (Hüllenprobe `nat04handel`, #1672).
	Handelsbeleg(ctx context.Context, vorgang string) (any, error)
}

// Handelszusage ist eine festgehaltene Besitzzeile.
type Handelszusage struct {
	ExemplarID int64  `json:"exemplarId"`
	Menge      int    `json:"menge"`
	Fassung    int64  `json:"fassung"`
	Revision   int    `json:"revision"`
	Frist      string `json:"frist,omitempty"`
	Uebergeben bool   `json:"uebergeben"`
}

// Handelszeile ist eine Position eines Angebots.
type Handelszeile struct {
	Von      string `json:"von"`
	Printing string `json:"printing"`
	Menge    int    `json:"menge"`
	Guete    string `json:"guete"`
	Sprache  string `json:"sprache"`
	Finish   string `json:"finish"`
}

// Handelsstand ist **eine** Revision eines Vorgangs.
type Handelsstand struct {
	Vorgang  string `json:"vorgang"`
	Revision int    `json:"revision"`

	// Abdruck ist die Kennung dieser Revision — eine Zustimmung nennt ihn,
	// statt „ja" zu sagen.
	Abdruck string `json:"abdruck"`

	Art       string   `json:"art"`
	Initiator string   `json:"initiator"`
	Parteien  []string `json:"parteien"`

	// Ich ist die Partei dieses Profils — die, deren Stimme Naechster nennt
	// (#2026). Leer, wo es keine Identität gibt.
	Ich     string `json:"ich"`
	Zustand string `json:"zustand"`

	Positionen   []Handelszeile `json:"positionen"`
	Zustimmungen []string       `json:"zustimmungen"`

	// Ende sagt, ob aus diesem Zustand nichts mehr folgt.
	Ende bool `json:"ende"`

	// Beschluss (TRADE-03) und wo der Vorgang für dieses Profil steht
	// (TRADE-02, #1716). **Moegliche ist eine Auskunft und keine
	// Einladung**: über diese Fläche geht keiner der Schritte.
	Beschluss string   `json:"beschluss"`
	Lage      string   `json:"lage"`
	Naechster string   `json:"naechster"`
	Moegliche []string `json:"moegliche"`
	Phase     string   `json:"phase"`
}

// handelsstandwerkzeug gibt das eine lesende — oder keines.
//
// **Ohne Naht kein Werkzeug**: eines, das in der Liste steht und beim Ruf
// abstürzt, wäre schlimmer als keines (eigener Befund aus `T8`).
func handelsstandwerkzeug(auskunft Handelsauskunft) []Werkzeug {
	if auskunft == nil {
		return nil
	}

	return append(handelsstandwerkzeuge(auskunft), handelsuebersichtwerkzeug(auskunft),
		handelsprotokollwerkzeug(auskunft), handelsanhaltwerkzeug(auskunft),
		handelsbelegwerkzeug(auskunft))
}

// handelsstandwerkzeuge sind die drei ersten lesenden.
func handelsstandwerkzeuge(auskunft Handelsauskunft) []Werkzeug {
	return []Werkzeug{{
		Name: "lesen.handel", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt, wo ein Handelsvorgang steht: seine Art, seine " +
			"Parteien, die Positionen der letzten Revision, ihren Abdruck " +
			"und wer ihr zugestimmt hat. LIEST NUR — über diese Fläche wird " +
			"kein Handel eröffnet, keiner angeboten, keinem zugestimmt und " +
			"keiner verbucht. Das ist Absicht: eine Zusage gegenüber einem " +
			"anderen Menschen gibt kein Modell in fremdem Namen, und die " +
			"verbindlichen Schritte sind über die Reichweite eines Tokens " +
			"ausgeschlossen.",
		Schema: objekt(map[string]any{
			"vorgang": text("Die Kennung des Handelsvorgangs."),
		}, "vorgang"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Vorgang string `json:"vorgang"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			stand, err := auskunft.Handelsstand(ctx, argumente.Vorgang)
			if err != nil {
				return "", err
			}

			return alsJSON(stand)
		},
	}, {
		Name: "lesen.handelsvorgaenge", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Handel mit einem Gegenüber — je Vorgang die " +
			"letzte Revision, zuletzt geänderte zuerst, mit Kennung, Art, " +
			"Zustand und Positionen. LIEST NUR, aus demselben Grund wie " +
			"lesen.handel: einen Handel bewegt über diese Fläche niemand.",
		Schema: objekt(map[string]any{
			"gegenueber": text("Die Peerkennung des Gegenübers."),
		}, "gegenueber"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Gegenueber string `json:"gegenueber"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			vorgaenge, err := auskunft.Handelsvorgaenge(ctx, argumente.Gegenueber)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"vorgaenge": vorgaenge})
		},
	}, {
		Name: "lesen.handelszusagen", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt, welche Besitzzeilen für einen Handelsvorgang " +
			"festgehalten sind — Besitzzeile, Stückzahl, Angebotsrevision " +
			"und ob sie schon über den Tisch gegangen sind. LIEST NUR: " +
			"**welches** Exemplar weggeht, wählt ein Mensch, und zwar am " +
			"Fenster, an der Kommandozeile oder über den Dienst. Ein Modell, " +
			"das es wählte, legte die falsche Karte auf den Tisch — und der " +
			"Besitzer sähe es erst, wenn sie weg ist.",
		Schema: objekt(map[string]any{
			"vorgang": text("Die Kennung des Handelsvorgangs."),
		}, "vorgang"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Vorgang string `json:"vorgang"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			zusagen, err := auskunft.Handelszusagen(ctx, argumente.Vorgang)
			if err != nil {
				return "", err
			}

			return alsJSON(zusagen)
		},
	}}
}

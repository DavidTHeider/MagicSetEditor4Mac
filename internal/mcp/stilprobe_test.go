// **Die Doubles für Stile und Bestand** (#1107).
//
// **Sie merken sich, was gesetzt wird.** Die Idempotenzprobe ruft zweimal und
// vergleicht; ein Double, das nichts merkt, machte jedes schreibende Werkzeug
// idempotent (§10, „eine Gegenprobe, die nichts ändert, ist keine").
package mcp_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// stilprobe antwortet auf die sechs Stilwerkzeuge.
type stilprobe struct {
	gewaehlt string
	optionen map[string]string
	fehler   error
}

func neueStilprobe() *stilprobe {
	return &stilprobe{gewaehlt: "probe-schlicht", optionen: map[string]string{}}
}

func (s *stilprobe) Stile(context.Context, int64) ([]mcp.Stil, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	return []mcp.Stil{
		{Paket: "probe-schlicht", Spiel: "probe", Gewaehlt: s.gewaehlt == "probe-schlicht"},
		{Paket: "probe-einzeiler", Spiel: "probe", Gewaehlt: s.gewaehlt == "probe-einzeiler"},
	}, nil
}

func (s *stilprobe) Folgen(context.Context, int64, string) (mcp.Wechselfolgen, error) {
	if s.fehler != nil {
		return mcp.Wechselfolgen{}, s.fehler
	}

	return mcp.Wechselfolgen{UngedeuteteFelder: []string{"rarity"}}, nil
}

func (s *stilprobe) Wechseln(_ context.Context, _ int64, paket string) error {
	if s.fehler != nil {
		return s.fehler
	}

	s.gewaehlt = paket

	return nil
}

func (s *stilprobe) Optionen(context.Context, int64) ([]mcp.Stiloption, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	wert, gesetzt := s.optionen["rand"]

	return []mcp.Stiloption{
		{Name: "rand", Art: "choice", Wert: wert, Gesetzt: gesetzt},
	}, nil
}

func (s *stilprobe) OptionSetzen(_ context.Context, _ int64, feld, wert string) error {
	if s.fehler != nil {
		return s.fehler
	}

	s.optionen[feld] = wert

	return nil
}

func (s *stilprobe) OptionZuruecksetzen(_ context.Context, _ int64, feld string) error {
	if s.fehler != nil {
		return s.fehler
	}

	delete(s.optionen, feld)

	return nil
}

// bestandsprobeWeit antwortet auf die acht Bestandswerkzeuge.
type bestandsprobeWeit struct {
	einstellungen map[string]string
	schritte      int
	fehler        error
}

func neueBestandsprobeWeit() *bestandsprobeWeit {
	return &bestandsprobeWeit{
		einstellungen: map[string]string{"sprache": "de"},
		schritte:      2,
	}
}

func (b *bestandsprobeWeit) Pakete(context.Context) ([]mcp.Paketzeile, []string, error) {
	if b.fehler != nil {
		return nil, nil, b.fehler
	}

	return []mcp.Paketzeile{{Art: "game", Name: "probe", Version: "1"}}, []string{}, nil
}

func (b *bestandsprobeWeit) PaketBeziehen(
	context.Context, string, string, string, string,
) (mcp.Paketbezug, error) {
	if b.fehler != nil {
		return mcp.Paketbezug{}, b.fehler
	}

	return mcp.Paketbezug{Ort: "/pakete/probe", Stand: "abc1234"}, nil
}

func (b *bestandsprobeWeit) Statistik(context.Context, int64) (mcp.Setzahlen, error) {
	if b.fehler != nil {
		return mcp.Setzahlen{}, b.fehler
	}

	return mcp.Setzahlen{
		Karten: 1, Felder: 1,
		Belegung: []mcp.Zahl{{Name: "name", Karten: 1}},
	}, nil
}

func (b *bestandsprobeWeit) Verteilung(context.Context, int64, string) ([]mcp.Wertzahl, error) {
	if b.fehler != nil {
		return nil, b.fehler
	}

	return []mcp.Wertzahl{{Name: "Feuerball", Karten: 1, KartenIDs: []string{"4"}}}, nil
}

// Zuruecknehmen zählt herunter — beim zweiten Ruf ist der Verlauf am Ende.
//
// **Deshalb sagt `schreiben.zuruecknehmen` keine Idempotenz zu**, und dieser
// Zähler ist der Beleg dafür.
func (b *bestandsprobeWeit) Zuruecknehmen(context.Context, int64, string) (mcp.Bewegung, error) {
	if b.fehler != nil {
		return mcp.Bewegung{}, b.fehler
	}

	if b.schritte == 0 {
		return mcp.Bewegung{}, nil
	}

	b.schritte--

	return mcp.Bewegung{
		Bewegt: true, KarteID: "2", Feld: "name", Umfang: 1,
		Urheber: "server",
	}, nil
}

// Verlauf gibt zwei Schritte, solange etwas zu bewegen ist: eine
// Massenänderung über zwei Karten (der jüngste) und eine Einzeländerung
// (U6, #1417).
func (b *bestandsprobeWeit) Verlauf(context.Context, int64) ([]mcp.Verlaufsschritt, error) {
	if b.fehler != nil {
		return nil, b.fehler
	}

	if b.schritte == 0 {
		return []mcp.Verlaufsschritt{}, nil
	}

	return []mcp.Verlaufsschritt{
		{
			ID: "3", Vorgang: "g1", Umfang: 2, KarteID: "8", Handlung: "feld", Feld: "typ",
			Vorher: "Zauber", GabVorher: true, Nachher: "Land", GabNachher: true,
			Urheber: "server", Zeitpunkt: "2026-09-09T02:01:00Z",
		},
		{
			ID: "1", Umfang: 1, KarteID: "7", Handlung: "feld", Feld: "name",
			Vorher: "Alt", GabVorher: true, Nachher: "Neu", GabNachher: true,
			Urheber: "fenster", Zeitpunkt: "2026-09-09T02:00:00Z",
		},
	}, nil
}

func (b *bestandsprobeWeit) Wiederherstellen(context.Context, int64, string) (mcp.Bewegung, error) {
	if b.fehler != nil {
		return mcp.Bewegung{}, b.fehler
	}

	b.schritte++

	return mcp.Bewegung{
		Bewegt: true, KarteID: "2", Feld: "name", Umfang: 1,
		Urheber: "server",
	}, nil
}

func (b *bestandsprobeWeit) Einstellungen(
	context.Context,
) ([]mcp.Einstellungszeile, error) {
	if b.fehler != nil {
		return nil, b.fehler
	}

	return []mcp.Einstellungszeile{{
		Schluessel: "sprache", Wert: b.einstellungen["sprache"],
		Werte: []string{"de", "en"}, Bereich: "anwendung",
	}}, nil
}

func (b *bestandsprobeWeit) EinstellungSetzen(
	_ context.Context, schluessel, wert string,
) error {
	if b.fehler != nil {
		return b.fehler
	}

	b.einstellungen[schluessel] = wert

	return nil
}

// **Und der Austausch** — Werkstück O3 von Phase 5.2
// (#1427).
//
// **Sie antworten knapp und richtig**: was der Austausch **kann**, prüft
// `internal/service` und der Lauf über alle drei Flächen; hier geht es um die
// Werkzeugliste, und dafür reicht eine Antwort, die die Gestalt hat.

func (b *bestandsprobeWeit) EinstellungenAusfuehren(
	context.Context, []string,
) (string, error) {
	if b.fehler != nil {
		return "", b.fehler
	}

	return `{"fassung":1,"bereiche":{}}`, nil
}

func (b *bestandsprobeWeit) EinstellungenVorschau(context.Context, []byte) (
	mcp.Einstellungssicht, error,
) {
	if b.fehler != nil {
		return mcp.Einstellungssicht{}, b.fehler
	}

	return mcp.Einstellungssicht{
		Fassung: 1, Revision: "abc123",
		Zeilen: []mcp.Einstellungsdiffzeile{{
			Ebene: "profil", Schluessel: "sprache",
			Vorher: "de", Nachher: "en", Herkunft: "profil",
		}},
		Befunde: []mcp.Einstellungsbefund{},
	}, nil
}

func (b *bestandsprobeWeit) EinstellungenUebernehmen(
	context.Context, []byte, string, []string,
) (mcp.Einstellungsbericht, error) {
	if b.fehler != nil {
		return mcp.Einstellungsbericht{}, b.fehler
	}

	return mcp.Einstellungsbericht{
		Uebernommen: 1, Revision: "def456",
		Zeilen:  []mcp.Einstellungsdiffzeile{},
		Befunde: []mcp.Einstellungsbefund{},
	}, nil
}

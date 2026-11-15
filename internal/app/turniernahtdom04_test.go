package app_test

// **Das Turnierdoppel für alles, was DOM-04 dazubrachte** — Folgen, Kopie,
// Abschluss, Protokoll, Datei und Bericht
// (#1699).
//
// **Eine eigene Datei**, damit `turniernaht_test.go` unter der Zeilengrenze
// aus §2 bleibt. Der Typ bleibt einer — Methoden dürfen daneben liegen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func (d *turniernahtdoppel) Abschliessen(_ context.Context, turnier, fassung int64) error {
	d.abschluss.turnier, d.abschluss.fassung, d.abschluss.zu = turnier, fassung, true

	return d.fehler
}

func (d *turniernahtdoppel) Wiedereroeffnen(_ context.Context, turnier, fassung int64) error {
	d.abschluss.turnier, d.abschluss.fassung, d.abschluss.zu = turnier, fassung, false

	return d.fehler
}

func (d *turniernahtdoppel) Protokoll(_ context.Context, _ int64) (
	[]domain.Protokolleintrag, error,
) {
	return []domain.Protokolleintrag{
		{Zeit: "2026-09-27T12:00:00Z", Vorgang: domain.TurnierErgebnis, Angaben: "2:0:0", PaarungID: 12},
	}, d.fehler
}

// Ausgeben gibt eine Datei, die ihre Kennung nennt.
func (d *turniernahtdoppel) Ausgeben(context.Context, int64) ([]byte, error) {
	if d.fehler != nil {
		return nil, d.fehler
	}

	return []byte("ausgabe von 7\n"), nil
}

// Einlesen merkt sich die Datei und antwortet mit Turnier 8 und einem Befund.
func (d *turniernahtdoppel) Einlesen(_ context.Context, roh []byte) (
	domain.Turnier, []domain.Einfuhrbefund, error,
) {
	d.eingelesen = roh
	if d.fehler != nil {
		return domain.Turnier{}, nil, d.fehler
	}

	return domain.Turnier{ID: 8, Name: "Freitag"},
		[]domain.Einfuhrbefund{{Schluessel: "format", Grund: "ohne Format"}}, nil
}

// Bericht merkt sich die Wörter und nennt die Überschrift der Rangliste.
func (d *turniernahtdoppel) Bericht(_ context.Context, _ int64, texte domain.Berichttexte) (string, error) {
	d.worte = texte

	return "# Bericht mit " + texte.Rangliste + "\n", d.fehler
}

func (d *turniernahtdoppel) Korrekturfolgen(
	_ context.Context, _, _ int64, stand *domain.Ergebnis,
) (domain.Korrekturfolgen, error) {
	d.gefragt = stand

	return domain.Korrekturfolgen{
		Runde: 1, SpaetereRunden: []int{2},
		Verschoben: []domain.Platzwechsel{{
			TeilnehmerID: 3, Name: "Bert", PlatzVorher: 2, PlatzNachher: 1,
			PunkteVorher: 3, PunkteNachher: 6,
		}},
	}, d.fehler
}

func (d *turniernahtdoppel) Kopieren(
	ctx context.Context, quelle int64, name string, mitTeilnehmenden bool,
) (domain.Turnier, error) {
	d.kopiert.quelle, d.kopiert.name, d.kopiert.mitTeilnehmenden = quelle, name, mitTeilnehmenden

	return d.Anlegen(ctx, domain.Turnier{Name: name, Spiel: "magic", FormatID: 4})
}

package rest_test

// **Die Doppel und die Liste für die Aufträge** — FMT-03 (#1692). Die
// Wirkung der Aufträge prüft `auftraege_test.go` gegen das echte Buch; hier
// stehen nur die vier Klassen.

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func endpunkteAuftraege() []endpunkt {
	return []endpunkt{
		{name: "die Aufträge zeigen", methode: http.MethodGet, weg: "/auftraege", erfolg: http.StatusOK},
		{name: "einen Auftrag zeigen", methode: http.MethodGet, weg: "/auftraege/a", erfolg: http.StatusOK},
		{
			name: "das Ergebnis eines Auftrags", methode: http.MethodGet,
			weg: "/auftraege/a/ergebnis", erfolg: http.StatusOK,
		},
		{name: "einen Auftrag abbrechen", methode: http.MethodDelete, weg: "/auftraege/a", erfolg: http.StatusAccepted},
		{
			name: "einen Auftrag wiederholen", methode: http.MethodPost,
			weg: "/auftraege/a/wiederholung", erfolg: http.StatusAccepted,
		},
		{
			// **Ein Set vorab ansehen** (FMT-03.AK-01) — `POST`, weil der
			// Pfad im Rumpf steht, wie bei jedem Weg, der eine Datei liest.
			name: "ein Set vorab ansehen", methode: http.MethodPost,
			weg: "/sets/vorschau", rumpf: `{"pfad":"/tmp/probe.mse-set"}`, erfolg: http.StatusOK,
		},
		// **Der Ausgang** (FMT-03.AK-03) — abholen und vergessen, und nur
		// durch den, der die Ausgabe anstieß (#2002).
		{
			name: "eine Ausgabe abholen", methode: http.MethodGet, weg: "/ausgang/a.pdf",
			erfolg: http.StatusOK, vorher: rest.AusgabeDerMarke("a.pdf"),
		},
		{
			name: "eine Ausgabe vergessen", methode: http.MethodDelete, weg: "/ausgang/a.pdf",
			erfolg: http.StatusNoContent, vorher: rest.AusgabeDerMarke("a.pdf"),
		},
	}
}

func (williger) Annehmen(context.Context, rest.Auftragseingabe, rest.Auftragslauf) (rest.AuftragDTO, error) {
	return rest.AuftragDTO{ID: "a", Stand: "laeuft"}, nil
}

func (williger) Wiederholen(context.Context, string, rest.Auftragslauf) (rest.AuftragDTO, error) {
	return rest.AuftragDTO{ID: "b", Stand: "laeuft", WiederholungVon: "a"}, nil
}

func (williger) Auftraege(context.Context) ([]rest.AuftragDTO, error) {
	return []rest.AuftragDTO{{ID: "a", Stand: "fertig"}}, nil
}

func (williger) Auftrag(context.Context, string) (rest.AuftragDTO, error) {
	return rest.AuftragDTO{ID: "a", Stand: "fertig"}, nil
}

func (williger) Ergebnis(context.Context, string) (rest.Auftragsergebnis, error) {
	return rest.Auftragsergebnis{Status: http.StatusOK, Art: "application/json", Rumpf: []byte("{}")}, nil
}

func (williger) Abbrechen(context.Context, string) (bool, error) { return true, nil }

func (z *zaehlender) Annehmen(ctx context.Context, eingabe rest.Auftragseingabe, lauf rest.Auftragslauf) (rest.AuftragDTO, error) {
	z.zaehle()

	return z.williger.Annehmen(ctx, eingabe, lauf)
}

func (z *zaehlender) Wiederholen(ctx context.Context, id string, lauf rest.Auftragslauf) (rest.AuftragDTO, error) {
	z.zaehle()

	return z.williger.Wiederholen(ctx, id, lauf)
}

func (z *zaehlender) Auftraege(ctx context.Context) ([]rest.AuftragDTO, error) {
	z.zaehle()

	return z.williger.Auftraege(ctx)
}

func (z *zaehlender) Auftrag(ctx context.Context, id string) (rest.AuftragDTO, error) {
	z.zaehle()

	return z.williger.Auftrag(ctx, id)
}

func (z *zaehlender) Ergebnis(ctx context.Context, id string) (rest.Auftragsergebnis, error) {
	z.zaehle()

	return z.williger.Ergebnis(ctx, id)
}

func (z *zaehlender) Abbrechen(ctx context.Context, id string) (bool, error) {
	z.zaehle()

	return z.williger.Abbrechen(ctx, id)
}

func (u unwilliger) Annehmen(context.Context, rest.Auftragseingabe, rest.Auftragslauf) (rest.AuftragDTO, error) {
	return rest.AuftragDTO{}, u.fehler
}

func (u unwilliger) Wiederholen(context.Context, string, rest.Auftragslauf) (rest.AuftragDTO, error) {
	return rest.AuftragDTO{}, u.fehler
}

func (u unwilliger) Auftraege(context.Context) ([]rest.AuftragDTO, error) { return nil, u.fehler }

func (u unwilliger) Auftrag(context.Context, string) (rest.AuftragDTO, error) {
	return rest.AuftragDTO{}, u.fehler
}

func (u unwilliger) Ergebnis(context.Context, string) (rest.Auftragsergebnis, error) {
	return rest.Auftragsergebnis{}, u.fehler
}

func (u unwilliger) Abbrechen(context.Context, string) (bool, error) { return false, u.fehler }

func (williger) SetVorschau(_ context.Context, pfad string) (any, error) {
	return map[string]any{"pfad": pfad, "name": "Mein Probeset", "karten": 3}, nil
}

func (z *zaehlender) SetVorschau(ctx context.Context, pfad string) (any, error) {
	z.zaehle()

	return z.williger.SetVorschau(ctx, pfad)
}

func (u unwilliger) SetVorschau(context.Context, string) (any, error) { return nil, u.fehler }

func (williger) Ablegen(context.Context, string, io.Reader) (string, error) {
	return "/eingang/probe.mse-set", nil
}

func (williger) Zwischenlegen(context.Context, string, io.Reader) (pfad string, weg func(), err error) {
	return "/vorschau/probe.mse-set", func() {}, nil
}

func (williger) Ausgangsort(_ context.Context, endung string) (name, pfad string, err error) {
	return "a" + endung, "/ausgang/a" + endung, nil
}

func (williger) Abholen(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("%PDF-")), nil
}

func (williger) Vergessen(context.Context, string) error { return nil }

func (z *zaehlender) Abholen(ctx context.Context, name string) (io.ReadCloser, error) {
	z.zaehle()

	return z.williger.Abholen(ctx, name)
}

func (z *zaehlender) Vergessen(ctx context.Context, name string) error {
	z.zaehle()

	return z.williger.Vergessen(ctx, name)
}

func (u unwilliger) Ablegen(context.Context, string, io.Reader) (string, error) { return "", u.fehler }

func (u unwilliger) Zwischenlegen(context.Context, string, io.Reader) (pfad string, weg func(), err error) {
	return "", nil, u.fehler
}

func (u unwilliger) Ausgangsort(context.Context, string) (name, pfad string, err error) {
	return "", "", u.fehler
}

func (u unwilliger) Abholen(context.Context, string) (io.ReadCloser, error) {
	return nil, u.fehler
}

func (u unwilliger) Vergessen(context.Context, string) error { return u.fehler }

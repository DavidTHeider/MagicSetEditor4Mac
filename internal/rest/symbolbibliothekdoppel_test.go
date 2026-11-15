// **Die Doppel für die Symbolbibliothek** — DOM-02
// (#1697).
//
// **Alle drei haben jedes Verfahren**, weil `rest.Naehte` sie als **eine**
// Schnittstelle nimmt — dieselbe Lehre wie bei den Turnieren.

package rest_test

import (
	"context"
	"encoding/json"
)

func (w williger) Symbolbibliothek(context.Context, string) (any, error) {
	return []any{map[string]any{"id": "1", "name": "Krone"}}, nil
}

func (w williger) BibliothekssymbolText(context.Context, int64) (any, error) {
	return map[string]any{"inhalt": "mse_version: 0.3.5"}, nil
}

func (w williger) BibliothekssymbolAufnehmen(context.Context, int64, string) (any, error) {
	return map[string]any{"id": "1"}, nil
}

func (w williger) BibliothekssymbolKopieren(context.Context, int64, string) (any, error) {
	return map[string]any{"id": "2"}, nil
}

func (w williger) BibliothekssymbolUmbenennen(context.Context, int64, string) (any, error) {
	return map[string]any{"id": "1"}, nil
}

func (w williger) BibliothekssymbolLoeschen(context.Context, int64, bool) (any, error) {
	return map[string]any{"sets": []string{"Kronenset"}}, nil
}

func (w williger) BibliothekssymbolEinsetzen(context.Context, int64, int64) (any, error) {
	return map[string]any{"datei": "symbol2.mse-symbol"}, nil
}

func (w williger) BibliothekssymbolAusText(context.Context, string, string) (any, error) {
	return map[string]any{"id": "3"}, nil
}

func (u unwilliger) Symbolbibliothek(context.Context, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) BibliothekssymbolText(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) BibliothekssymbolAufnehmen(context.Context, int64, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) BibliothekssymbolKopieren(context.Context, int64, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) BibliothekssymbolUmbenennen(context.Context, int64, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) BibliothekssymbolLoeschen(context.Context, int64, bool) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) BibliothekssymbolEinsetzen(context.Context, int64, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) BibliothekssymbolAusText(context.Context, string, string) (any, error) {
	return nil, u.fehler
}

func (z *zaehlender) Symbolbibliothek(ctx context.Context, suche string) (any, error) {
	z.zaehle()

	return z.williger.Symbolbibliothek(ctx, suche)
}

func (z *zaehlender) BibliothekssymbolText(ctx context.Context, id int64) (any, error) {
	z.zaehle()

	return z.williger.BibliothekssymbolText(ctx, id)
}

func (z *zaehlender) BibliothekssymbolAufnehmen(ctx context.Context, setID int64, name string) (any, error) {
	z.zaehle()

	return z.williger.BibliothekssymbolAufnehmen(ctx, setID, name)
}

func (z *zaehlender) BibliothekssymbolKopieren(ctx context.Context, id int64, name string) (any, error) {
	z.zaehle()

	return z.williger.BibliothekssymbolKopieren(ctx, id, name)
}

func (z *zaehlender) BibliothekssymbolUmbenennen(ctx context.Context, id int64, name string) (any, error) {
	z.zaehle()

	return z.williger.BibliothekssymbolUmbenennen(ctx, id, name)
}

func (z *zaehlender) BibliothekssymbolLoeschen(ctx context.Context, id int64, bestaetigt bool) (any, error) {
	z.zaehle()

	return z.williger.BibliothekssymbolLoeschen(ctx, id, bestaetigt)
}

func (z *zaehlender) BibliothekssymbolEinsetzen(ctx context.Context, id, setID int64) (any, error) {
	z.zaehle()

	return z.williger.BibliothekssymbolEinsetzen(ctx, id, setID)
}

func (z *zaehlender) BibliothekssymbolAusText(ctx context.Context, inhalt, name string) (any, error) {
	z.zaehle()

	return z.williger.BibliothekssymbolAusText(ctx, inhalt, name)
}

func (w williger) BibliothekssymbolTeile(context.Context, int64) (any, error) {
	return map[string]any{"teile": []any{}}, nil
}

func (w williger) BibliothekssymbolSichern(context.Context, int64, json.RawMessage) (any, error) {
	return map[string]any{"veraltet": []string{}}, nil
}

func (w williger) BibliothekssymbolZuruecknehmen(context.Context, int64) (any, error) {
	return map[string]any{"veraltet": []string{}}, nil
}

func (w williger) BibliothekssymbolNachziehen(context.Context, int64) (any, error) {
	return map[string]any{"sets": []string{}}, nil
}

func (u unwilliger) BibliothekssymbolTeile(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) BibliothekssymbolSichern(context.Context, int64, json.RawMessage) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) BibliothekssymbolZuruecknehmen(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) BibliothekssymbolNachziehen(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (z *zaehlender) BibliothekssymbolTeile(ctx context.Context, id int64) (any, error) {
	z.zaehle()

	return z.williger.BibliothekssymbolTeile(ctx, id)
}

func (z *zaehlender) BibliothekssymbolSichern(ctx context.Context, id int64, teile json.RawMessage) (any, error) {
	z.zaehle()

	return z.williger.BibliothekssymbolSichern(ctx, id, teile)
}

func (z *zaehlender) BibliothekssymbolZuruecknehmen(ctx context.Context, id int64) (any, error) {
	z.zaehle()

	return z.williger.BibliothekssymbolZuruecknehmen(ctx, id)
}

func (z *zaehlender) BibliothekssymbolNachziehen(ctx context.Context, id int64) (any, error) {
	z.zaehle()

	return z.williger.BibliothekssymbolNachziehen(ctx, id)
}

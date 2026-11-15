// **Der Dateibaum an den Doubles** — FMT-02
// (#1691).
// Eine eigene Datei, weil `willigerdoppel_test.go` an der Zeilengrenze steht.

package rest_test

import "context"

func (williger) DateiAnlegen(context.Context, string, string, string) (string, error) {
	return "neu", nil
}

func (williger) DateiUmbenennen(context.Context, string, string, string, string) error { return nil }

func (williger) DateiEntfernen(context.Context, string, string, string) error { return nil }

func (u unwilliger) DateiAnlegen(context.Context, string, string, string) (string, error) {
	return "", u.fehler
}

func (u unwilliger) DateiUmbenennen(context.Context, string, string, string, string) error {
	return u.fehler
}

func (u unwilliger) DateiEntfernen(context.Context, string, string, string) error { return u.fehler }

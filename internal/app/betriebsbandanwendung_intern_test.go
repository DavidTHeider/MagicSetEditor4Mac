package app

// **Das Betriebsband an der Anwendung** — OPS02 (#2348): die Quellen
// des Betriebs, über einem eigenen Zuhause und einem Wegwerfbestand.

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"testing"
)

// bandAnwendung startet eine Anwendung über einem eigenen Zuhause.
func bandAnwendung(t *testing.T, heim func() (string, error)) *App {
	t.Helper()

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitSchnittstellenorten(heim, func() (string, error) { return t.TempDir(), nil }, func(context.Context) string { return "" })

	bilder := t.TempDir()
	anwendung.MitBilderort(func() (string, error) { return bilder, nil })

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // nur zum Aufräumen

	return anwendung
}

// **Ein neu geladener Webprozess meldet sich, mit seiner Zahl** — und ein
// frischer Bestand ohne Dienst und ohne Sicherung gibt sonst nichts.
func TestDasFensterZaehltNeuGeladeneWebprozesse(t *testing.T) {
	t.Parallel()

	heim := t.TempDir()
	anwendung := bandAnwendung(t, func() (string, error) { return heim, nil })
	bindungen := &Bindungen{anwendung: anwendung}

	if vorher := bindungen.Betriebsband(t.Context()); len(vorher.Hinweise) != 0 || slices.Contains(vorher.Ungemessen, "sicherung") || slices.Contains(vorher.Ungemessen, "dienst") {
		t.Fatalf("vorher %+v", vorher)
	}

	anwendung.WebprozessNeuGestartet(t.Context())
	anwendung.WebprozessNeuGestartet(t.Context())

	danach := anwendung.Betriebsband(t.Context())
	if len(danach.Hinweise) != 1 || danach.Hinweise[0].Art != "webprozessNeu" || danach.Hinweise[0].Werte["anzahl"] != "2" {
		t.Errorf("danach %+v", danach)
	}
}

// **Ohne Zuhause und ohne Bestand** steht beides als ungemessen da.
func TestOhneZuhauseUndBestandIstDasBandUngemessen(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitSchnittstellenorten(func() (string, error) { return "", errors.New("kein Zuhause") },
		func() (string, error) { return "", nil }, func(context.Context) string { return "" })

	antwort := anwendung.Betriebsband(t.Context())
	if !slices.Contains(antwort.Ungemessen, "sicherung") || !slices.Contains(antwort.Ungemessen, "dienst") || len(antwort.Hinweise) != 0 {
		t.Errorf("Antwort %+v", antwort)
	}
}

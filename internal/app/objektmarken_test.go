package app_test

// **Marken an jedem Fachobjekt, am Fenster** — UX-12
// (#1684).

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDasFensterSetztFiltertUndPflegtMarken(t *testing.T) {
	t.Parallel()

	db, fehler := storage.Oeffnen(t.Context(), storage.ImArbeitsspeicher)
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	if fehler := storage.Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	for _, anweisung := range []string{
		`INSERT INTO decks (id, name, spiel, profil_id) VALUES (11, 'Rot', 'magic', 1)`,
		`INSERT INTO decks (id, name, spiel, profil_id) VALUES (12, 'Blau', 'magic', 1)`,
	} {
		if _, fehler := db.ExecContext(t.Context(), anweisung); fehler != nil {
			t.Fatal(fehler)
		}
	}

	anwendung := app.NeueApp("")
	anwendung.MitObjektmarken(app.ObjektmarkenUeber(db))

	ctx := storage.MitProfil(t.Context(), 1)

	if antwort := anwendung.ObjektmarkeSetzen(ctx, "deck", "11", "Tausch"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.MarkeVergeben(ctx, "deck", []string{"11", "12"}, "turnier"); antwort.Fehler != nil {
		t.Fatalf("vergeben: %+v", antwort.Fehler)
	}

	je := anwendung.MarkenJeObjekt(ctx, "deck")
	if je.Fehler != nil || !slices.Equal(je.Marken["11"], []string{"tausch", "turnier"}) {
		t.Errorf("je Objekt: %+v", je)
	}

	gefiltert := anwendung.MarkenFiltern(ctx, "deck", []string{"tausch", "turnier"}, "und")
	if gefiltert.Fehler != nil || !slices.Equal(gefiltert.Kennungen, []string{"11"}) {
		t.Errorf("filtern: %+v", gefiltert)
	}

	katalog := anwendung.Markenkatalog(ctx)
	if katalog.Fehler != nil || len(katalog.Marken) != 2 || katalog.Marken[1].Anzahl != 2 {
		t.Errorf("Katalog: %+v", katalog)
	}

	if antwort := anwendung.MarkeUmbenennen(ctx, "tausch", "turnier", false); antwort.Fehler == nil ||
		antwort.Fehler.Code != "konflikt" {
		t.Errorf("Kollision: %+v", antwort.Fehler)
	}

	if antwort := anwendung.MarkeUmbenennen(ctx, "tausch", "turnier", true); antwort.Fehler != nil || antwort.Anzahl != 1 {
		t.Errorf("zusammenführen: %+v", antwort)
	}

	// **Die Naht der Flächen ohne Fenster liest dieselben Marken.**
	if an, fehler := app.ObjektmarkenUeber(db).ObjektmarkenAn(ctx, "deck", "12"); fehler != nil ||
		!slices.Equal(an, []string{"turnier"}) {
		t.Errorf("an: %v %v", an, fehler)
	}

	if antwort := anwendung.ObjektmarkeEntfernen(ctx, "deck", "12", "turnier"); antwort.Fehler != nil {
		t.Errorf("entfernen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.MarkeUeberallEntfernen(ctx, "turnier"); antwort.Fehler != nil || antwort.Anzahl != 1 {
		t.Errorf("überall entfernen: %+v", antwort)
	}

	// **Fehler kommen als Umschlag, Listen nie als `null`.**
	for name, antwort := range map[string]struct {
		fehler *app.Fehler
		leer   bool
	}{
		"je Objekt":       {anwendung.MarkenJeObjekt(ctx, "schrift").Fehler, anwendung.MarkenJeObjekt(ctx, "schrift").Marken == nil},
		"filtern":         {anwendung.MarkenFiltern(ctx, "deck", nil, "und").Fehler, anwendung.MarkenFiltern(ctx, "deck", nil, "und").Kennungen == nil},
		"setzen":          {anwendung.ObjektmarkeSetzen(ctx, "deck", "99", "x").Fehler, false},
		"umbenennen":      {anwendung.MarkeUmbenennen(ctx, "", "x", false).Fehler, false},
		"Katalog fremd":   {anwendung.Markenkatalog(t.Context()).Fehler, anwendung.Markenkatalog(t.Context()).Marken == nil},
		"überall leer":    {anwendung.MarkeUeberallEntfernen(ctx, "").Fehler, false},
		"vergeben leer":   {anwendung.MarkeVergeben(ctx, "deck", nil, "x").Fehler, false},
		"entfernen fremd": {anwendung.ObjektmarkeEntfernen(ctx, "deck", "99", "x").Fehler, false},
	} {
		if antwort.fehler == nil || antwort.leer {
			t.Errorf("%s: Fehler %+v, leer %v", name, antwort.fehler, antwort.leer)
		}
	}
}

// **Jede Bindung erreicht ihren Weg** — über einem Bestand ohne Tafeln antworten
// alle mit einem Fehler, und keine mit `null`.
func TestDieMarkenbindungenJederArtErreichenIhreAnwendung(t *testing.T) {
	t.Parallel()

	db, fehler := storage.Oeffnen(t.Context(), storage.ImArbeitsspeicher)
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	anwendung := app.NeueApp("")
	anwendung.MitObjektmarken(app.ObjektmarkenUeber(db))
	bindungen := app.NeueBindungen(anwendung)

	fehler8 := []*app.Fehler{
		bindungen.ObjektmarkeSetzen(t.Context(), "deck", "1", "x").Fehler,
		bindungen.ObjektmarkeEntfernen(t.Context(), "deck", "1", "x").Fehler,
		bindungen.MarkeVergeben(t.Context(), "deck", []string{"1"}, "x").Fehler,
		bindungen.MarkenJeObjekt(t.Context(), "deck").Fehler,
		bindungen.MarkenFiltern(t.Context(), "deck", []string{"x"}, "und").Fehler,
		bindungen.Markenkatalog(t.Context()).Fehler,
		bindungen.MarkeUmbenennen(t.Context(), "x", "y", false).Fehler,
		bindungen.MarkeUeberallEntfernen(t.Context(), "x").Fehler,
	}

	for stelle, gemeldet := range fehler8 {
		if gemeldet == nil {
			t.Errorf("Bindung %d kam ohne Bestand durch", stelle)
		}
	}
}

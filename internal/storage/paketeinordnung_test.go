package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Einordnung überlebt den Neustart** (#244).
//
// `Auffrischen` liest die Platte, `Pakete` liest den Bestand — und dazwischen
// liegt die Tabelle. Eine Gruppe, die nur im Arbeitsspeicher steht, wäre beim
// zweiten Öffnen der Seite fort, und die Verwaltung fiele auf die flache Liste
// zurück, die #244 gerade abschafft.
func TestDieEinordnungStehtImBestand(t *testing.T) {
	katalog := storage.NeuerPaketkatalog(frischeDatenbank(t))

	pakete := []domain.Paket{{
		Art: "mse-style", Name: "magic-m15", Version: "2026-08-15",
		Pfad: "/p/magic-m15.mse-style", Gruppe: "magic/m15 style/normal cards", Rang: 95,
	}}

	if err := katalog.Abgleichen(probekontext(t), pakete); err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	gelesen, err := katalog.Pakete(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen) != 1 {
		t.Fatalf("%d Pakete", len(gelesen))
	}

	if gelesen[0].Gruppe != "magic/m15 style/normal cards" {
		t.Errorf("Gruppe = %q", gelesen[0].Gruppe)
	}

	if gelesen[0].Rang != 95 {
		t.Errorf("Rang = %d", gelesen[0].Rang)
	}
}

// **Ein Paket, das seine Gruppe verliert, verliert sie auch im Bestand.**
//
// Anders als bei `herkunft` und `pruefsumme`, die beim Abgleich **erhalten**
// bleiben, wenn nichts Neues kommt: die beiden stehen für etwas, das einmal
// geschehen ist und niemand nachträglich weiß. Die Gruppe steht in der
// Paketdatei — wer sie dort löscht, meint es, und ein Katalog, der sie
// festhielte, zeigte eine Ordnung, die es nicht mehr gibt.
func TestEineGeloeschteGruppeVerschwindet(t *testing.T) {
	katalog := storage.NeuerPaketkatalog(frischeDatenbank(t))

	mit := []domain.Paket{{
		Art: "mse-style", Name: "magic-m15", Pfad: "/p/magic-m15.mse-style",
		Gruppe: "magic/m15 style", Rang: 95,
	}}

	if err := katalog.Abgleichen(probekontext(t), mit); err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	ohne := []domain.Paket{{
		Art: "mse-style", Name: "magic-m15", Pfad: "/p/magic-m15.mse-style",
	}}

	if err := katalog.Abgleichen(probekontext(t), ohne); err != nil {
		t.Fatalf("zweiter abgleich: %v", err)
	}

	gelesen, err := katalog.Pakete(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen[0].Gruppe != "" {
		t.Errorf("Gruppe = %q", gelesen[0].Gruppe)
	}
}

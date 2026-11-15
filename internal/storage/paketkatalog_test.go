package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Tabelle `packages` steht seit Phase 2 und war bis hierher leer.**
//
// Der Zweitblick über Phase 9 hat das angestrichen: eine Tabelle, in die nur
// hineingeschrieben wird — hier gar nicht —, ist eine Zusage ohne Einlösung.
// #54 löst sie
// ein: installierte Pakete stehen mit Art, Name, Version **und Herkunft** im
// Bestand.
func TestDerKatalogFuehrtArtNameVersionUndHerkunft(t *testing.T) {
	katalog := storage.NeuerPaketkatalog(frischeDatenbank(t))

	pakete := []domain.Paket{
		{Art: "mse-game", Name: "probe", Version: "2026-08-15", Pfad: "/p/probe.mse-game"},
		{
			Art: "mse-style", Name: "probe-schlicht", Version: "2026-08-15",
			Pfad: "/p/probe-schlicht.mse-style", Herkunft: "https://example.invalid/vorlagen",
			Braucht: []domain.Bedarf{{Paket: "probe.mse-game", Version: "2026-08-15"}},
		},
	}

	if err := katalog.Abgleichen(probekontext(t), pakete); err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	gelesen, err := katalog.Pakete(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen) != 2 {
		t.Fatalf("%d Pakete, erwartet 2", len(gelesen))
	}

	// Nach Art und Name geordnet — eine Liste ohne Ordnung ist bei jedem
	// Aufruf eine andere, und die Oberfläche zeigte sie springend.
	if gelesen[0].Name != "probe" || gelesen[1].Name != "probe-schlicht" {
		t.Fatalf("Reihenfolge: %q, %q", gelesen[0].Name, gelesen[1].Name)
	}

	stil := gelesen[1]
	if stil.Version != "2026-08-15" || stil.Herkunft != "https://example.invalid/vorlagen" {
		t.Errorf("Version %q, Herkunft %q", stil.Version, stil.Herkunft)
	}

	if len(stil.Braucht) != 1 || stil.Braucht[0].Paket != "probe.mse-game" {
		t.Errorf("Abhängigkeiten: %+v", stil.Braucht)
	}
}

// **Was nicht mehr dasteht, fällt aus dem Katalog.**
//
// Ein Abgleich ist keine Ergänzung: wer ein Paket aus dem Verzeichnis nimmt,
// hätte sonst einen Eintrag, der auf nichts zeigt — und die Stilauswahl böte
// einen Stil an, den es nicht gibt.
func TestEinAbgleichRaeumtWasVerschwundenIst(t *testing.T) {
	katalog := storage.NeuerPaketkatalog(frischeDatenbank(t))

	erst := []domain.Paket{
		{Art: "mse-game", Name: "probe", Pfad: "/p/probe.mse-game"},
		{Art: "mse-style", Name: "alt", Pfad: "/p/alt.mse-style"},
	}
	if err := katalog.Abgleichen(probekontext(t), erst); err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	if err := katalog.Abgleichen(probekontext(t), erst[:1]); err != nil {
		t.Fatalf("zweiter Abgleich: %v", err)
	}

	gelesen, err := katalog.Pakete(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen) != 1 || gelesen[0].Name != "probe" {
		t.Errorf("übrig: %+v", gelesen)
	}
}

// **Eine bekannte Herkunft überlebt den Abgleich.**
//
// Der Abgleich liest die Pakete von der Platte, und die Platte weiß nicht,
// woher sie kommen. Wer die Spalte dabei überschriebe, verlöre bei der ersten
// Auffrischung genau die Angabe, wegen der sie geführt wird (§11).
func TestDerAbgleichBehaeltEineBekannteHerkunft(t *testing.T) {
	katalog := storage.NeuerPaketkatalog(frischeDatenbank(t))

	// **Der Abdruck reist wie die Herkunft** (DOM-01.AK-01, #1696): er
	// entsteht beim Einbinden, und ein Abgleich ohne ihn behält ihn.
	mitHerkunft := []domain.Paket{{
		Art: "mse-style", Name: "bezogen", Pfad: "/p/bezogen.mse-style",
		Herkunft: "https://example.invalid/pack", Abdruck: "abdruck:ab",
	}}
	if err := katalog.Abgleichen(probekontext(t), mitHerkunft); err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	ohneHerkunft := []domain.Paket{{
		Art: "mse-style", Name: "bezogen", Pfad: "/p/bezogen.mse-style", Version: "2026-08-16",
	}}
	if err := katalog.Abgleichen(probekontext(t), ohneHerkunft); err != nil {
		t.Fatalf("zweiter Abgleich: %v", err)
	}

	gelesen, err := katalog.Pakete(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen) != 1 {
		t.Fatalf("%d Pakete", len(gelesen))
	}

	if gelesen[0].Herkunft != "https://example.invalid/pack" {
		t.Errorf("Herkunft %q — der Abgleich hat sie überschrieben", gelesen[0].Herkunft)
	}

	if gelesen[0].Abdruck != "abdruck:ab" {
		t.Errorf("Abdruck %q — der Abgleich hat ihn überschrieben", gelesen[0].Abdruck)
	}

	if gelesen[0].Version != "2026-08-16" {
		t.Errorf("Version %q — die neue Angabe kam nicht an", gelesen[0].Version)
	}
}

// **Ein Eintrag, der die Domänenregel nicht besteht, kommt nicht in den
// Bestand** — und der ganze Abgleich nicht mit.
func TestDerKatalogWeistEinKaputtesPaketAb(t *testing.T) {
	katalog := storage.NeuerPaketkatalog(frischeDatenbank(t))

	err := katalog.Abgleichen(probekontext(t), []domain.Paket{
		{Art: "mse-game", Name: "gut", Pfad: "/p/gut.mse-game"},
		{Art: "", Name: "schlecht", Pfad: "/p/schlecht"},
	})
	if err == nil {
		t.Fatal("kein Fehler")
	}

	gelesen, fehler := katalog.Pakete(probekontext(t))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gelesen) != 0 {
		t.Errorf("halb geschrieben: %+v", gelesen)
	}
}

// **Entfernen nimmt den Eintrag und seine Abhängigkeiten mit.**
func TestEntfernenNimmtDenEintragAusDemKatalog(t *testing.T) {
	db := frischeDatenbank(t)
	katalog := storage.NeuerPaketkatalog(db)

	pakete := []domain.Paket{{
		Art: "mse-style", Name: "weg", Pfad: "/p/weg.mse-style",
		Braucht: []domain.Bedarf{{Paket: "probe.mse-game"}},
	}}
	if err := katalog.Abgleichen(probekontext(t), pakete); err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	if err := katalog.PaketEntfernen(probekontext(t), "mse-style", "weg"); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	gelesen, err := katalog.Pakete(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen) != 0 {
		t.Errorf("noch da: %+v", gelesen)
	}

	var abhaengigkeiten int
	if err := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM package_deps`).Scan(&abhaengigkeiten); err != nil {
		t.Fatalf("zählen: %v", err)
	}

	if abhaengigkeiten != 0 {
		t.Errorf("%d verwaiste Abhängigkeiten", abhaengigkeiten)
	}
}

// **Ein Paket, das der Katalog nicht führt, ist hier kein Fehler.**
//
// Das Urteil fällt in `service.Paketverwalter.Entfernen`, das den Pfad ohnehin
// suchen muss, um die Dateien zu räumen. Zwei Urteile über dieselbe Frage
// liefen auseinander — der Grund steht bei `Paketkatalog.PaketEntfernen`.
func TestEntfernenSchweigtUeberEinUnbekanntesPaket(t *testing.T) {
	katalog := storage.NeuerPaketkatalog(frischeDatenbank(t))

	if err := katalog.PaketEntfernen(probekontext(t), "mse-style", "niegesehen"); err != nil {
		t.Errorf("Fehler: %v", err)
	}
}

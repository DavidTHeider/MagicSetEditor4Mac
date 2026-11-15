package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Anwendung findet die Exportvorlagen ihres Spiels.**
//
// # Die Naht, die #209 gefehlt hat
//
// `Vorlagenausfuehrer` stand seit #51 und wurde außerhalb von Tests nirgends
// gerufen. Zwischen ihm und dem Fenster fehlte die Frage, **welche** Vorlagen
// es gibt.
func TestDieVorlagenlisteFindetDieVorlagenIhresSpiels(t *testing.T) {
	wurzel := vorlagenwurzel(t)

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	vorlagen, befunde, err := service.NeueVorlagenliste(quelle, wurzel).
		Vorlagen(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(vorlagen) != 1 {
		t.Fatalf("%d Vorlagen, erwartet 1; Befunde: %v", len(vorlagen), befunde)
	}

	if vorlagen[0].KurzerName != "Probeausfuhr" {
		t.Errorf("%+v", vorlagen[0])
	}
}

// **Eine Vorlage eines anderen Spiels wird nicht angeboten.**
//
// Sie ergäbe Skriptfehler über Feldnamen, die es dort nicht gibt — und der
// Anwendende bekäme eine Fehlerliste statt der Auskunft, dass sie nicht passt.
func TestEineVorlageEinesAnderenSpielsWirdNichtAngeboten(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "einanderes"}}}

	vorlagen, _, err := service.NeueVorlagenliste(quelle, vorlagenwurzel(t)).
		Vorlagen(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(vorlagen) != 0 {
		t.Errorf("%d Vorlagen — keine gehört zu diesem Spiel", len(vorlagen))
	}
}

// **Eine Vorlage ohne Skript wird benannt, nicht angeboten.**
//
// Zwei Pakete des bezogenen Packs tragen die Endung und kein `script:`; einen
// Knopf zu zeigen, der nur scheitern kann, ist schlechter als keiner.
func TestEineVorlageOhneSkriptWirdBenannt(t *testing.T) {
	wurzel := vorlagenwurzel(t)
	legeVerzeichnis(t, filepath.Join(wurzel, "leer.mse-export-template"))

	if err := os.WriteFile(
		filepath.Join(wurzel, "leer.mse-export-template", "export-template"),
		[]byte("mse version: 2.0.0\ngame: probe\nshort name: Leer\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	vorlagen, befunde, err := service.NeueVorlagenliste(quelle, wurzel).
		Vorlagen(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(vorlagen) != 1 || len(befunde) != 1 {
		t.Errorf("%d Vorlagen, %d Befunde: %v", len(vorlagen), len(befunde), befunde)
	}
}

// **Ein Set, das der Bestand nicht führt, ist ein Befund.**
func TestDieVorlagenlisteMeldetEinUnbekanntesSet(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "probe"}}}

	_, befunde, err := service.NeueVorlagenliste(quelle, vorlagenwurzel(t)).
		Vorlagen(t.Context(), 99)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(befunde) == 0 {
		t.Error("kein Befund")
	}
}

// **Ein Lesefehler des Bestands kommt als Fehler heraus.**
func TestDieVorlagenlisteReichtEinenLesefehlerDurch(t *testing.T) {
	quelle := &koepfequelle{fehler: os.ErrPermission}

	if _, _, err := service.NeueVorlagenliste(quelle, vorlagenwurzel(t)).
		Vorlagen(t.Context(), 7); err == nil {
		t.Error("kein Fehler")
	}
}

// **Eine Wurzel, die es nicht gibt, hält den Lauf nicht auf.**
func TestEineFehlendeWurzelHaeltDieListeNichtAuf(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	vorlagen, _, err := service.NeueVorlagenliste(quelle, "/gibtesnicht").
		Vorlagen(t.Context(), 7)
	if err != nil || len(vorlagen) != 0 {
		t.Errorf("%d Vorlagen, Fehler %v", len(vorlagen), err)
	}
}

// **Ein Paket, das sich nicht öffnen lässt, wird benannt.**
func TestEinUnlesbaresVorlagenpaketWirdBenannt(t *testing.T) {
	wurzel := t.TempDir()
	legeVerzeichnis(t, filepath.Join(wurzel, "kaputt.mse-export-template"))

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	_, befunde, err := service.NeueVorlagenliste(quelle, wurzel).Vorlagen(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(befunde) == 0 {
		t.Error("kein Befund für ein Paket ohne Hauptdatei")
	}
}

// vorlagenwurzel legt ein Verzeichnis mit einer Probevorlage an.
func vorlagenwurzel(t *testing.T) string {
	t.Helper()

	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "probe-ausfuhr.mse-export-template")
	legeVerzeichnis(t, paket)

	inhalt := "mse version: 2.0.0\ngame: probe\nshort name: Probeausfuhr\n" +
		"full name: Ausfuhr der Probe\nfile type: *.txt|*.txt\n" +
		"option field:\n\ttype: boolean\n\tname: mit namen\n\tinitial: yes\n" +
		"script:\n\tto_string(\"fertig\")\n"

	if err := os.WriteFile(filepath.Join(paket, "export-template"),
		[]byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return wurzel
}

// legeVerzeichnis legt ein Verzeichnis samt Eltern an.
func legeVerzeichnis(t *testing.T, pfad string) {
	t.Helper()

	if err := os.MkdirAll(pfad, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}
}

// **Dieselbe Vorlage aus zwei Wurzeln zählt einmal.**
//
// Wer die Saat und ein bezogenes Pack hat, hätte sonst jede doppelt in der
// Auswahl — und die zweite wäre eine, die niemand gewählt hat.
func TestEineVorlageAusZweiWurzelnZaehltEinmal(t *testing.T) {
	erste := vorlagenwurzel(t)
	zweite := vorlagenwurzel(t)

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	vorlagen, _, err := service.NeueVorlagenliste(quelle, erste, zweite).
		Vorlagen(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(vorlagen) != 1 {
		t.Errorf("%d Vorlagen — dieselbe aus zwei Wurzeln", len(vorlagen))
	}
}

// **Was keine Exportvorlage ist, wird übergangen.**
func TestEinFremdesPaketWirdUebergangen(t *testing.T) {
	wurzel := vorlagenwurzel(t)
	legeVerzeichnis(t, filepath.Join(wurzel, "probe.mse-game"))

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	vorlagen, befunde, err := service.NeueVorlagenliste(quelle, wurzel).
		Vorlagen(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(vorlagen) != 1 || len(befunde) != 0 {
		t.Errorf("%d Vorlagen, Befunde %v", len(vorlagen), befunde)
	}
}

// **Der Dienst führt Liste und Ausführer zusammen.**
//
// Er ist das, was der Composition Root verdrahtet; ohne diesen Fall wäre die
// Naht, die #209 gebaut hat, an genau der Stelle ungeprüft, an der sie benutzt
// wird.
func TestDerVorlagendienstKannBeides(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}
	dienst := service.NeuerVorlagendienst(quelle, vorlagenwurzel(t))

	vorlagen, _, err := dienst.Vorlagen(t.Context(), 7)
	if err != nil || len(vorlagen) != 1 {
		t.Fatalf("%d Vorlagen, Fehler %v", len(vorlagen), err)
	}

	ausgabe, _, err := dienst.Fuehre(t.Context(), service.Ausfuhrmittel{}, vorlagen[0], nil, nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if ausgabe != "fertig" {
		t.Errorf("%q", ausgabe)
	}
}

// **Die Vorlagen kommen nach Namen geordnet.**
//
// Ohne feste Ordnung hinge die Reihenfolge der Auswahlliste daran, wie das
// Dateisystem seine Einträge zurückgibt — und die wechselte zwischen zwei
// Rechnern, ohne dass jemand etwas geändert hat.
func TestDieVorlagenKommenGeordnet(t *testing.T) {
	wurzel := vorlagenwurzel(t)
	legeVorlage(t, wurzel, "a-zweite.mse-export-template", "Zweite Ausfuhr")

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	vorlagen, _, err := service.NeueVorlagenliste(quelle, wurzel).Vorlagen(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(vorlagen) != 2 {
		t.Fatalf("%d Vorlagen", len(vorlagen))
	}

	// „Ausfuhr der Probe" vor „Zweite Ausfuhr" — nach dem **Namen**, nicht nach
	// dem Dateinamen: `a-zweite` stünde sonst vorn.
	if vorlagen[0].VollerName != "Ausfuhr der Probe" {
		t.Errorf("%q vorn", vorlagen[0].VollerName)
	}
}

// legeVorlage schreibt eine Exportvorlage in eine Wurzel.
func legeVorlage(t *testing.T, wurzel, paketname, name string) {
	t.Helper()

	paket := filepath.Join(wurzel, paketname)
	legeVerzeichnis(t, paket)

	inhalt := "mse version: 2.0.0\ngame: probe\nshort name: Kurz\n" +
		"full name: " + name + "\nscript:\n\tto_string(\"fertig\")\n"

	if err := os.WriteFile(filepath.Join(paket, "export-template"),
		[]byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

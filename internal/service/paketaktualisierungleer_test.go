package service_test

// **Ein Baum ohne Pakete ist trotzdem ein Bezug** — G04
// (#2251).
//
// Gefunden am nativen Fenster (`scripts/huellenprobe/g04bezuege.mjs`): der
// erste Commit von `Magic-Mini-Packs` trägt nur seine README. Der Bezug
// glückte, der Katalog bekam kein Paket — und der Baum stand **nirgends**,
// auch nicht unter den bezogenen Ständen. Aktualisieren ließ er sich damit
// auch nicht, obwohl der nächste Commit die Pakete bringt.

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const (
	miniAdresse = "https://github.com/MagicSetEditorPacks/Magic-Mini-Packs"
	miniStand   = "97bb53a0e3f8f093cab19f18044c45792625e616"
	miniSpitze  = "c3d9b16ecd539a939e448a9fb8c6e5df2a619395"
)

// leererBaum legt einen bezogenen Baum ohne Paket an — nur mit seinem Vermerk.
func leererBaum(t *testing.T, verwaltet, name, vermerk string) {
	t.Helper()

	ort := filepath.Join(verwaltet, name)
	if fehler := os.MkdirAll(ort, 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(filepath.Join(ort, "README.md"), []byte("# Mini\n"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	if vermerk == "" {
		return
	}

	if fehler := os.WriteFile(filepath.Join(ort, ".mse4mac-bezug.json"), []byte(vermerk), 0o600); fehler != nil {
		t.Fatal(fehler)
	}
}

func TestEinBaumOhnePaketeStehtUnterDenBezuegen(t *testing.T) {
	aktualisierung, _, verwaltet := aktualisierungsplatz(t)
	leererBaum(t, verwaltet, "Magic-Mini-Packs",
		`{"adresse":"`+miniAdresse+`","stand":"`+miniStand+`","unterverzeichnis":"","pruefsumme":"sha256:mini"}`)
	// **Ohne Vermerk ist ein Verzeichnis kein Bezug** — es kann alles sein.
	leererBaum(t, verwaltet, "Irgendwas", "")
	// **Ein unlesbarer Vermerk nennt keine Quelle**, also auch keinen Bezug.
	leererBaum(t, verwaltet, "Kaputt", "kein json")
	// **Ein Bezug, der gerade entsteht, ist noch keiner**: er liegt verborgen
	// neben dem Ziel (`Paketbezieher.beziehen`), und sein Vermerk steht schon.
	leererBaum(t, verwaltet, ".Magic-Mini-Packs-123",
		`{"adresse":"`+miniAdresse+`","stand":"`+miniSpitze+`"}`)

	staende, fehler := aktualisierung.Staende(t.Context(), false)
	if fehler != nil {
		t.Fatalf("Staende: %v", fehler)
	}

	namen := make([]string, 0, len(staende))
	for _, stand := range staende {
		namen = append(namen, stand.Name)
	}

	if !slices.Equal(namen, []string{"M15-Magic-Pack", "Magic-Mini-Packs"}) {
		t.Fatalf("Bäume %v, erwartet M15-Magic-Pack und Magic-Mini-Packs", namen)
	}

	mini := staende[1]
	if mini.Ort != filepath.Join(verwaltet, "Magic-Mini-Packs") || mini.Adresse != miniAdresse ||
		mini.Stand != miniStand || mini.Pakete != 0 || mini.Lage != "ungefragt" || mini.Geprueft != "" {
		t.Errorf("leerer Baum %+v", mini)
	}
}

func TestEinBaumOhnePaketeLaesstSichAktualisieren(t *testing.T) {
	aktualisierung, auskunft, verwaltet := aktualisierungsplatz(t)
	leererBaum(t, verwaltet, "Magic-Mini-Packs",
		`{"adresse":"`+miniAdresse+`","stand":"`+miniStand+`","unterverzeichnis":"","pruefsumme":"sha256:mini"}`)
	auskunft.vergleich = bezug.Vergleich{Lage: "ahead", Vorsprung: 1, Dateien: []bezug.Aenderung{
		{Datei: "Deckbox/magic-deckbox.mse-style/style", Art: "added"},
		{Datei: "Deckbox/magic-deckbox.mse-style/frame.png", Art: "added"},
	}}

	vorschau, fehler := aktualisierung.Vorschau(t.Context(), "Magic-Mini-Packs", miniSpitze)
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if vorschau.Von != miniStand || vorschau.Zu != miniSpitze || len(vorschau.Pakete) != 1 ||
		vorschau.Pakete[0].Paket != "magic-deckbox.mse-style" || vorschau.Pakete[0].Neu != 2 || vorschau.BestaetigungNoetig {
		t.Errorf("Vorschau %+v", vorschau)
	}

	if !slices.Contains(auskunft.gefragt, "vergleich "+miniAdresse+" "+miniStand+"..."+miniSpitze) {
		t.Errorf("gefragt %v", auskunft.gefragt)
	}
}

func TestDerPaketordnerFehltOderIstKeiner(t *testing.T) {
	ablageMit := func(verwaltet string) *service.Paketaktualisierung {
		verwalter := service.NeuerPaketverwalter(&ablage{}, verwaltet)

		return service.NeuePaketaktualisierung(verwalter, service.NeuerPaketbezieher(verwaltet, "probe"), &auskunftDouble{})
	}

	t.Run("ein fehlender Ordner heißt: nichts bezogen", func(t *testing.T) {
		staende, fehler := ablageMit(filepath.Join(t.TempDir(), "gibt-es-nicht")).Staende(t.Context(), false)
		if fehler != nil || len(staende) != 0 {
			t.Errorf("Stände %+v, Fehler %v", staende, fehler)
		}
	})

	t.Run("eine Datei an seiner Stelle ist ein Fehler", func(t *testing.T) {
		datei := filepath.Join(t.TempDir(), "pakete")
		if fehler := os.WriteFile(datei, nil, 0o600); fehler != nil {
			t.Fatal(fehler)
		}

		if _, fehler := ablageMit(datei).Staende(t.Context(), false); fehler == nil {
			t.Error("eine Datei als Paketordner ging durch")
		}
	})
}

// **Eine Datei, die in ein anderes Paket wandert, fehlt dem alten** —
// Befund 3 des Zweitblicks über Phase G, zweiter Teil: die Vorschau zählte
// sie als Änderung des neuen Pakets, und ein Set auf dem alten verlor sie,
// ohne dass eine Bestätigung verlangt wurde.
func TestEineUmbenennungUeberPaketeIstEinVerlust(t *testing.T) {
	aktualisierung, auskunft, _ := aktualisierungsplatz(t)
	auskunft.vergleich = bezug.Vergleich{Lage: "ahead", Vorsprung: 1, Dateien: []bezug.Aenderung{
		{Datei: "data/anders.mse-style/style", Art: "renamed", Vorher: "data/magic-m15.mse-style/style"},
		{Datei: "data/magic-m15.mse-style/neu.png", Art: "renamed", Vorher: "data/magic-m15.mse-style/alt.png"},
		// **Und über die Grenze eines Pakets hinaus**: hinein zählt als neu,
		// hinaus als Verlust — und nichts doppelt unter den übrigen.
		{Datei: "data/rein.mse-style/style", Art: "renamed", Vorher: "LIESMICH.txt"},
		{Datei: "alt/raus.txt", Art: "renamed", Vorher: "data/raus.mse-style/style"},
	}}

	vorschau, fehler := aktualisierung.Vorschau(t.Context(), "M15-Magic-Pack", m15Spitze)
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	pakete := map[string]service.Paketaenderung{}
	for _, paket := range vorschau.Pakete {
		pakete[paket.Paket] = paket
	}

	if m15 := pakete["magic-m15.mse-style"]; m15.Entfernt != 1 || m15.Geaendert != 1 {
		t.Errorf("magic-m15: %+v", m15)
	}

	if anders := pakete["anders.mse-style"]; anders.Neu != 1 {
		t.Errorf("anders: %+v", anders)
	}

	if pakete["rein.mse-style"].Neu != 1 || pakete["raus.mse-style"].Entfernt != 1 || vorschau.Uebrige != 0 {
		t.Errorf("rein %+v, raus %+v, übrige %d", pakete["rein.mse-style"], pakete["raus.mse-style"], vorschau.Uebrige)
	}

	if !vorschau.BestaetigungNoetig || !slices.Contains(vorschau.Gruende, service.GrundSetVerliert) {
		t.Errorf("Gründe %v", vorschau.Gruende)
	}
}

// **Der Vermerk auf der Platte gilt vor dem Katalog** — Befund 4 des
// Zweitblicks über Phase G, zweiter Teil: nach einem geglückten Tausch und
// einem gescheiterten Vermerken nannte der Katalog den alten Stand, und die
// nächste Vorschau verglich gegen ihn.
func TestDerVermerkGiltVorDemKatalog(t *testing.T) {
	aktualisierung, _, verwaltet := aktualisierungsplatz(t)
	leererBaum(t, verwaltet, "M15-Magic-Pack",
		`{"adresse":"`+m15Adresse+`","stand":"`+m15Spitze+`","unterverzeichnis":"","pruefsumme":"sha256:neu"}`)

	staende, fehler := aktualisierung.Staende(t.Context(), false)
	if fehler != nil || len(staende) != 1 {
		t.Fatalf("Stände %+v, Fehler %v", staende, fehler)
	}

	if staende[0].Stand != m15Spitze || staende[0].Pakete != 2 {
		t.Errorf("Stand %s, Pakete %d — erwartet der Vermerk und die Pakete des Katalogs", staende[0].Stand, staende[0].Pakete)
	}
}

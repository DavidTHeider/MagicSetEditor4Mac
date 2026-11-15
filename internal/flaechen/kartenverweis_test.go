// **Der Verweis auf einen echten Druck über jede Fläche** — Werkstück 2 von
// Phase 3.6
// (#1167).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. **Die Erklärung, was
// diese Läufe halten und was nicht, steht in `scryfallsatz_test.go`.**
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

// verweisdruck ist die Kennung, auf die in diesen Läufen verwiesen wird.
//
// **Eine UUID und kein Wort.** Der Klient glättet sie und vergleicht sie mit
// der Antwort; ein Kürzel ginge durch, sagte aber nichts über den Weg, den
// eine echte Kennung nimmt.
const verweisdruck = "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1"

// **Setzen, Lesen und Lösen — über jede Fläche.**
//
// Der Ablauf ist an jeder derselbe: Set anlegen, Karte darin nehmen, sie auf
// einen Druck beziehen, den Bezug lesen, ihn lösen.
func TestEinKartenverweisUeberJedeFlaeche(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)

		gesetzt := mussKommandozeile(t, umgebung, "verweis", karteID, verweisdruck, "--json")
		if !strings.Contains(gesetzt, verweisdruck) {
			t.Fatalf("setzen: %s", gesetzt)
		}

		gelesen := mussKommandozeile(t, umgebung, "verweis", karteID, "--json")
		if !strings.Contains(gelesen, "Erste Probe") {
			t.Fatalf("lesen: %s", gelesen)
		}

		// **Die Setliste sieht ihn auch** — sie ist der Weg, den eine
		// Kartenliste ginge.
		jeSet := mussKommandozeile(t, umgebung, "verweise", setID, "--json")
		if !strings.Contains(jeSet, verweisdruck) {
			t.Fatalf("je Set: %s", jeSet)
		}

		mussKommandozeile(t, umgebung, "verweis", karteID, "--loesen", "--json")

		danach := mussKommandozeile(t, umgebung, "verweise", setID, "--json")
		if strings.Contains(danach, verweisdruck) {
			t.Fatalf("nach dem Lösen: %s", danach)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		_, karteID := geleerteKarte(t, umgebung)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.verweis",
			map[string]any{"karteID": zahl(t, karteID), "printing": verweisdruck})
		if istFehler {
			t.Fatalf("schreiben.verweis: %s", text)
		}

		gelesen, istFehler := mcpRuf(t, umgebung, "lesen.verweis",
			map[string]any{"karteID": zahl(t, karteID)})
		if istFehler || !strings.Contains(gelesen, verweisdruck) {
			t.Fatalf("lesen.verweis: %s", gelesen)
		}

		if _, istFehler := mcpRuf(t, umgebung, "schreiben.verweisloesen",
			map[string]any{"karteID": zahl(t, karteID)}); istFehler {
			t.Fatal("schreiben.verweisloesen scheiterte")
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPut, "/karten/"+karteID+"/verweis",
			`{"printing":"`+verweisdruck+`"}`)
		if status != http.StatusOK || !strings.Contains(rumpf, verweisdruck) {
			t.Fatalf("PUT verweis: %d (%s)", status, rumpf)
		}

		status, rumpf = an.ruf(t, http.MethodGet, "/sets/"+setID+"/verweise", "")
		if status != http.StatusOK || !strings.Contains(rumpf, "Erste Probe") {
			t.Fatalf("GET verweise: %d (%s)", status, rumpf)
		}

		status, rumpf = an.ruf(t, http.MethodDelete, "/karten/"+karteID+"/verweis", "")
		if status != http.StatusOK {
			t.Fatalf("DELETE verweis: %d (%s)", status, rumpf)
		}
	})
}

// **Ein Druck, den es nicht gibt, wird überall abgewiesen** — und hinterlässt
// keinen Verweis.
//
// # Warum gerade dieser Fall
//
// Er ist die **Grenze** des Verweises: er wird aufgelöst, bevor er entsteht.
// Stünde die Regel nur an einer Fläche, legte eine andere eine UUID ab, die
// niemand lesen kann — und der Fehler zeigte sich erst beim nächsten Abgleich.
func TestEinUnbekannterDruckWirdUeberallAbgewiesen(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		_, karteID := geleerteKarte(t, umgebung)

		aus, _, code := kommandozeile(t, umgebung, "verweis", karteID, "gibtsnicht")
		if code == 0 {
			t.Fatalf("der unbekannte Druck kam durch: %s", aus)
		}

		gelesen := mussKommandozeile(t, umgebung, "verweis", karteID, "--json")
		if strings.Contains(gelesen, "gibtsnicht") {
			t.Fatalf("es blieb ein Verweis zurück: %s", gelesen)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		_, karteID := geleerteKarte(t, umgebung)

		if _, istFehler := mcpRuf(t, umgebung, "schreiben.verweis",
			map[string]any{
				"karteID": zahl(t, karteID), "printing": "gibtsnicht",
			}); !istFehler {
			t.Fatal("der unbekannte Druck kam durch")
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		_, karteID := geleerteKarte(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPut, "/karten/"+karteID+"/verweis",
			`{"printing":"gibtsnicht"}`)
		if status == http.StatusOK {
			t.Fatalf("der unbekannte Druck kam durch: %s", rumpf)
		}
	})
}

// **Und eine Karte ohne Verweis lässt sich nicht lösen** — überall gleich.
//
// Ein stiller Erfolg ließe niemanden merken, dass er die falsche Karte
// erwischt hat.
func TestEinLoesenOhneVerweisWirdUeberallAbgewiesen(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		_, karteID := geleerteKarte(t, umgebung)

		aus, _, code := kommandozeile(t, umgebung, "verweis", karteID, "--loesen")
		if code == 0 {
			t.Fatalf("das Lösen ohne Verweis kam durch: %s", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		_, karteID := geleerteKarte(t, umgebung)

		if _, istFehler := mcpRuf(t, umgebung, "schreiben.verweisloesen",
			map[string]any{"karteID": zahl(t, karteID)}); !istFehler {
			t.Fatal("das Lösen ohne Verweis kam durch")
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		_, karteID := geleerteKarte(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodDelete, "/karten/"+karteID+"/verweis", "")
		if status == http.StatusOK {
			t.Fatalf("das Lösen ohne Verweis kam durch: %s", rumpf)
		}
	})
}

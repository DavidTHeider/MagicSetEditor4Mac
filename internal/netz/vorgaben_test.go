package netz_test

// **Die Einstellung schlägt die Umgebung** — Werkstück N5 von Phase 5.4
// (#1571).
//
// # Was hier gemessen wird
//
// **Je Feld und nicht als Ganzes.** Wer nur `netz.proxy.https` setzt, behält
// `HTTP_PROXY` aus der Umgebung — sonst schaltete eine einzelne Einstellung
// die andere still ab, und niemand sähe es.
//
// **Und dass die Umgebung bei jedem Ruf neu gelesen wird.** Gos
// `http.ProxyFromEnvironment` speichert sie beim ersten Ruf zwischen
// (gemessen am 2026-09-12); ein Lauf, der die Variable setzt, prüfte damit
// einen anderen Wert als den gesetzten und wäre **grün**.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestDieEinstellungSchlaegtDieUmgebung(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://umgebung.http:3128")
	t.Setenv("HTTPS_PROXY", "http://umgebung.https:3128")
	t.Setenv("NO_PROXY", "aus.der.umgebung")

	// **Ohne Vorgabe gilt die Umgebung.**
	ohne := netz.StandMit(t.Context(), nil)
	if ohne.HTTPS.Wert != "http://umgebung.https:3128" ||
		ohne.HTTPS.Herkunft != netz.AusDerUmgebung {
		t.Errorf("ohne Einstellung steht %+v", ohne.HTTPS)
	}

	// **Mit Vorgabe je Feld** — und die anderen bleiben, wo sie waren.
	mit := netz.StandMit(t.Context(), func(context.Context) netz.Proxyvorgabe {
		return netz.Proxyvorgabe{HTTPS: "http://einstellung.https:3128"}
	})

	if mit.HTTPS.Wert != "http://einstellung.https:3128" ||
		mit.HTTPS.Herkunft != netz.AusDerEinstellung {
		t.Errorf("mit Einstellung steht %+v", mit.HTTPS)
	}

	// **Das ist die Zusage, die leicht durchfällt**: eine gesetzte
	// https-Adresse darf die http-Adresse aus der Umgebung nicht abräumen.
	if mit.HTTP.Wert != "http://umgebung.http:3128" ||
		mit.HTTP.Herkunft != netz.AusDerUmgebung {
		t.Errorf("die Umgebung für http ist fort: %+v", mit.HTTP)
	}

	if mit.Aus.Wert != "aus.der.umgebung" {
		t.Errorf("die Ausnahmeliste ist fort: %+v", mit.Aus)
	}

	if !mit.Benutzt() {
		t.Error("ein gesetzter Proxy gilt als nicht benutzt")
	}
}

// **Ohne beides steht nichts da — und das ist eine Aussage.**
func TestOhneProxyStehtOhneWert(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("NO_PROXY", "")

	stand := netz.StandMit(t.Context(), func(context.Context) netz.Proxyvorgabe { return netz.Proxyvorgabe{} })

	for name, feld := range map[string]netz.Proxyfeld{
		"http": stand.HTTP, "https": stand.HTTPS, "aus": stand.Aus,
	} {
		if feld.Wert != "" || feld.Herkunft != netz.OhneWert {
			t.Errorf("%s steht als %+v da", name, feld)
		}
	}

	if stand.Benutzt() {
		t.Error("ohne Proxy gilt einer als benutzt")
	}
}

// **Der Proxy wird je Anfrage gelesen und nicht beim Bauen.**
//
// # Warum das die Zusage ist
//
// **Weil ein Klient beim Verdrahten entsteht.** Das Fenster baut seine Nähte
// beim Start; wer danach einen Proxy einträgt, erwartet nicht, die Anwendung
// neu starten zu müssen. `http.Transport` ruft `Proxy` je Anfrage — und
// deshalb steht dort eine Funktion und kein Wert.
//
// # Warum das Ziel `.invalid` heißt und nicht die Schleife
//
// **Weil `httpproxy` die Schleifenadresse ausnahmslos am Proxy vorbeischickt**
// — das steht als gemessene Eigenschaft in `docs/NETZEIGENSCHAFTEN.md`, und
// ein Ziel auf `127.0.0.1` prüfte hier deshalb gar nichts. Der **Proxy**
// darf auf der Schleife liegen; er ist nicht das Ziel.
func TestEinGeaenderterProxyGiltAbDerNaechstenAnfrage(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("NO_PROXY", "")

	gefragt := make(chan string, 4)

	proxy := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gefragt <- anfrage.URL.String()

			schreiben.Header().Set("Content-Type", "application/json")
			if _, fehler := schreiben.Write([]byte(`{}`)); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	defer proxy.Close()

	var vorgabe netz.Proxyvorgabe

	klient := netz.NeuerKlientMit(netz.UeblicheFristen(),
		func(context.Context) netz.Proxyvorgabe { return vorgabe })

	kenner := netz.Kenner{Name: "probe", Annahme: "*/*"}

	// **Ohne Proxy geht die Anfrage direkt** — und scheitert, weil
	// `.invalid` nirgends aufgelöst wird (RFC 2606). Der Proxy sieht nichts.
	if _, fehler := netz.Hole(t.Context(), klient, nil, kenner,
		"http://beispiel.invalid/", 1024); fehler == nil {
		t.Fatal("eine Anfrage an .invalid glückte ohne Proxy")
	}

	select {
	case weg := <-gefragt:
		t.Fatalf("der Proxy wurde ohne Einstellung gefragt: %s", weg)
	default:
	}

	// **Jetzt kommt er dazu — an demselben Klienten.**
	vorgabe = netz.Proxyvorgabe{HTTP: proxy.URL}

	if _, fehler := netz.Hole(t.Context(), klient, nil, kenner,
		"http://beispiel.invalid/", 1024); fehler != nil {
		t.Fatalf("mit Proxy: %v", fehler)
	}

	select {
	case weg := <-gefragt:
		if weg != "http://beispiel.invalid/" {
			t.Errorf("der Proxy bekam %q", weg)
		}
	default:
		t.Fatal("der Proxy wurde nicht gefragt — die Einstellung wirkte nicht")
	}
}

// **Und ein Fehlschlag hinter einem Proxy nennt ihn** (§8).
//
// **Der offene Punkt aus A18** (#1533):
// „connection refused" ohne den Zusatz, dass die Verbindung gar nicht zur
// Gegenstelle ging, schickt jemanden in die falsche Richtung.
func TestEinFehlschlagHinterEinemProxyNenntIhn(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("NO_PROXY", "")

	klient := netz.NeuerKlientMit(netz.UeblicheFristen(),
		func(context.Context) netz.Proxyvorgabe {
			return netz.Proxyvorgabe{HTTP: "http://127.0.0.1:1"}
		})

	_, fehler := netz.Hole(t.Context(), klient, nil,
		netz.Kenner{Name: "probe", Annahme: "*/*"},
		"http://beispiel.invalid/", 1024)
	if fehler == nil {
		t.Fatal("die Anfrage über einen toten Proxy glückte")
	}

	if !enthaelt(fehler.Error(), "127.0.0.1:1") ||
		!enthaelt(fehler.Error(), string(netz.AusDerEinstellung)) {
		t.Errorf("die Meldung nennt den Proxy nicht: %v", fehler)
	}
}

// enthaelt sagt, ob ein Text ein Stück enthält.
func enthaelt(text, teil string) bool {
	for i := 0; i+len(teil) <= len(text); i++ {
		if text[i:i+len(teil)] == teil {
			return true
		}
	}

	return false
}

// **Ohne Proxy bleibt eine Meldung, wie sie war.**
//
// **Die Gegenprobe zum Lauf darüber**: sonst sagte er nur, dass irgendwo eine
// Adresse im Text steht.
func TestOhneProxyBleibtDieMeldungWieSieWar(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("NO_PROXY", "")

	klient := netz.NeuerKlient(netz.UeblicheFristen())

	// **Ein Port, auf dem nichts horcht.**
	_, fehler := netz.Hole(t.Context(), klient, nil,
		netz.Kenner{Name: "probe", Annahme: "*/*"}, "http://127.0.0.1:1", 1024)
	if fehler == nil {
		t.Fatal("die Anfrage an einen toten Port glückte")
	}

	if enthaelt(fehler.Error(), "über den Proxy") {
		t.Errorf("die Meldung nennt einen Proxy, den es nicht gibt: %v", fehler)
	}
}

// **Nur die eine Richtung bekommt den Hinweis.**
//
// **Wer `netz.proxy.http` setzt und `https://` ruft, hat für diesen Ruf
// keinen** — und dann stand auch keiner dazwischen. Ein Hinweis darauf wäre
// ein Satz über etwas, das nicht geschehen ist (§10).
func TestDerProxyhinweisGiltNurDerRichtungMitProxy(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("NO_PROXY", "")

	klient := netz.NeuerKlientMit(netz.UeblicheFristen(),
		func(context.Context) netz.Proxyvorgabe {
			return netz.Proxyvorgabe{HTTP: "http://127.0.0.1:1"}
		})

	// **Über `https://`** — dafür ist keiner gesetzt, also geht die Anfrage
	// direkt und scheitert am Namen.
	_, fehler := netz.Hole(t.Context(), klient, nil,
		netz.Kenner{Name: "probe", Annahme: "*/*"},
		"https://beispiel.invalid/", 1024)
	if fehler == nil {
		t.Fatal("die Anfrage an .invalid glückte")
	}

	if enthaelt(fehler.Error(), "über den Proxy") {
		t.Errorf("die Meldung nennt einen Proxy für die falsche Richtung: %v", fehler)
	}
}

// **Eine geglückte Anfrage bekommt keinen Hinweis.**
//
// **Die Gegenprobe zum Hinweis**: er hängt an einem Transportfehler und nicht
// an jeder Antwort — ein 404 ist eine Antwort und kein Verbindungsproblem.
func TestEineGeglueckteAnfrageBekommtKeinenHinweis(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("NO_PROXY", "")

	ziel := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.Header().Set("Content-Type", "application/json")
			if _, fehler := schreiben.Write([]byte(`{}`)); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	defer ziel.Close()

	klient := netz.NeuerKlientMit(netz.UeblicheFristen(),
		func(context.Context) netz.Proxyvorgabe {
			return netz.Proxyvorgabe{HTTP: "http://127.0.0.1:1"}
		})

	// **Die Schleife geht am Proxy vorbei** — gemessen und in
	// `docs/NETZEIGENSCHAFTEN.md` festgehalten; hier ist es der Weg zum
	// Erfolg trotz gesetztem Proxy.
	if _, fehler := netz.Hole(t.Context(), klient, nil,
		netz.Kenner{Name: "probe", Annahme: "*/*"}, ziel.URL, 1024); fehler != nil {
		t.Errorf("die Anfrage an die Schleife scheiterte: %v", fehler)
	}
}

// **Die Auswahl für einen fremden Transport ist dieselbe.**
//
// **Für `internal/anmeldung`** (N5, #1571): es ist ein Blatt und kennt dieses
// Paket nicht, bekommt aber die Gestalt, die `http.Transport.Proxy` erwartet.
// Ohne diesen Lauf wäre der einzige Beleg dafür, dass beide dasselbe tun, ein
// Satz in `docs/NETZEIGENSCHAFTEN.md`.
func TestDieProxyauswahlFuerFremdeTransporte(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("NO_PROXY", "")

	auswahl := netz.Proxyauswahl(func(context.Context) netz.Proxyvorgabe {
		return netz.Proxyvorgabe{HTTP: "http://gesetzt:3128"}
	})

	anfrage, fehler := http.NewRequestWithContext(t.Context(),
		http.MethodGet, "http://beispiel.invalid/", http.NoBody)
	if fehler != nil {
		t.Fatalf("die Anfrage bauen: %v", fehler)
	}

	durch, fehler := auswahl(anfrage)
	if fehler != nil {
		t.Fatalf("auswählen: %v", fehler)
	}

	if durch == nil || durch.Host != "gesetzt:3128" {
		t.Errorf("gewählt wurde %v", durch)
	}
}

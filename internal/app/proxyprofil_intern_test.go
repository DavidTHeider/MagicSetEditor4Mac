package app

// **Der Proxy einer Einstellungswahl gilt für die Anfrage, die ihn braucht** —
// gefunden von D06
// (#2237)
// am nativen Fenster.
//
// # Der Befund
//
// **Die Suche bei Scryfall ging am eingestellten Proxy vorbei.** Das
// Drehbuch `d06` setzte `netz.proxy.http`, und das Fenster meldete, der Name
// des Dienstes lasse sich nicht auflösen: es fragte ihn selbst. Der Klient
// fragte seine Proxyquelle mit `context.Background()`, der Speicher klemmt
// jede Abfrage aber auf das Profil des Kontexts (0066) — ohne eines gab es
// einen Lesefehler, und ein Lesefehler heißt hier „wie ohne Einstellung".
// So ging jeder Klient über [proxyquelleAus] direkt hinaus: Scryfall, die
// anderen Kartendienste und der Preisbezug.
//
// **Kein Test sah es**, weil sie den Proxy über eine Probe lasen, die nach
// keinem Profil fragt — eine Vorrichtung, großzügiger als der Betrieb (§3).
// Dieser liest über den echten Speicher.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDerProxyEinerEinstellungswahlGiltFuerDieAnfrage(t *testing.T) {
	ctx := t.Context()

	db, fehler := storage.Oeffnen(ctx, filepath.Join(t.TempDir(), "bestand.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migriere(ctx, db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	imProfil, profilfehler := storage.MitGezeigtemProfil(ctx, db)
	if profilfehler != nil {
		t.Fatalf("das Profil: %v", profilfehler)
	}

	einstellungen := storage.NeueEinstellungen(db)
	if err := einstellungen.Setzen(imProfil, domain.NetzProxyHTTP, "http://127.0.0.1:3128"); err != nil {
		t.Fatalf("den Proxy setzen: %v", err)
	}

	// **Die Anfrage trägt das Profil**, wie jede, die eine Bindung oder eine
	// Fläche ohne Fenster stellt.
	anfrage := httptest.NewRequestWithContext(imProfil, http.MethodGet, "http://scryfall.test/cards/search", nil)

	gewaehlt, err := netz.Proxyauswahl(proxyquelleAus(NeueEinstellungswahl(einstellungen)))(anfrage)
	if err != nil {
		t.Fatalf("die Wahl: %v", err)
	}

	if gewaehlt == nil || gewaehlt.Host != "127.0.0.1:3128" {
		t.Errorf("gewählt ist %v — erwartet der eingestellte Proxy 127.0.0.1:3128", gewaehlt)
	}
}

// **Der Fernweg der Kommandozeile nimmt den eingestellten Proxy** — Befund 1
// des Zweitblicks über Phase D.
//
// `mse fernhorchen` baut seine Anwendung über [PeerflaecheUeber], und die
// hatte keinen Einstellungsverwalter: die Vorgabe blieb leer. Dazu fragt
// tailcat seine Proxywahl mit einem eigenen Kontext, **ohne Profil** — die
// Quelle muss es sich selbst holen.
func TestDerFernwegDerKommandozeileNimmtDenEingestelltenProxy(t *testing.T) {
	ctx := t.Context()

	db, fehler := storage.Oeffnen(ctx, filepath.Join(t.TempDir(), "bestand.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migriere(ctx, db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	imProfil, profilfehler := storage.MitGezeigtemProfil(ctx, db)
	if profilfehler != nil {
		t.Fatalf("das Profil: %v", profilfehler)
	}

	if err := storage.NeueEinstellungen(db).Setzen(imProfil, domain.NetzProxyHTTPS, "http://127.0.0.1:3129"); err != nil {
		t.Fatalf("den Proxy setzen: %v", err)
	}

	// **Ein Kontext ohne Profil**, wie ihn tailcat für DERP baut.
	anfrage := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "https://derp.example/derp", nil)

	gewaehlt, err := netz.Proxyauswahl(PeerflaecheUeber(db).Proxyquelle())(anfrage)
	if err != nil {
		t.Fatalf("die Wahl: %v", err)
	}

	if gewaehlt == nil || gewaehlt.Host != "127.0.0.1:3129" {
		t.Errorf("gewählt ist %v — erwartet der eingestellte Proxy 127.0.0.1:3129", gewaehlt)
	}
}

// **Ohne lesbares Profil bleibt es bei der Umgebung** — ein Bestand, der
// noch keines führt, hält die Anfrage nicht auf (dieselbe Haltung wie
// [App.netzvorgaben]: ein Lesefehler heißt „wie ohne Einstellung").
func TestOhneLesbaresProfilGiltDieUmgebung(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("https_proxy", "")

	db, fehler := storage.Oeffnen(t.Context(), filepath.Join(t.TempDir(), "bestand.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if vorgabe := PeerflaecheUeber(db).Proxyquelle()(context.Background()); vorgabe != (netz.Proxyvorgabe{}) {
		t.Errorf("ohne Profil kam die Vorgabe %+v", vorgabe)
	}
}

// **Eine verschlossene Sitzung gibt kein Profil her, auch nicht dem Proxy** —
// neuer Befund des Nachlaufs über Phase D. Der Rückfall auf das gezeigte
// Profil las dessen Einstellungen, obwohl `imProfilOhneSperre` bei
// verschlossener Sitzung gerade keines anhängt (ID-02). Dann gilt die
// Umgebung.
func TestEineVerschlosseneSitzungLiestDenProxyNicht(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("http_proxy", "")

	anwendung, _ := gestartetMitWeiche(t, nil)

	imProfil, profilfehler := storage.MitGezeigtemProfil(t.Context(), anwendung.laufendeVerbindung())
	if profilfehler != nil {
		t.Fatalf("das Profil: %v", profilfehler)
	}

	if err := storage.NeueEinstellungen(anwendung.laufendeVerbindung()).Setzen(imProfil,
		domain.NetzProxyHTTP, "http://127.0.0.1:3128"); err != nil {
		t.Fatalf("den Proxy setzen: %v", err)
	}

	// **Erst der Beleg, dass jemand zuhört** (§3): offen liest die Quelle
	// den eingestellten Proxy.
	if vorgabe := anwendung.Proxyquelle()(context.Background()); vorgabe.HTTP == "" {
		t.Fatalf("offen kam keine Vorgabe: %+v", vorgabe)
	}

	if antwort := anwendung.KennwortWechseln(t.Context(), "", "ein-langes-kennwort"); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if antwort := anwendung.Sperren(t.Context()); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if vorgabe := anwendung.Proxyquelle()(context.Background()); vorgabe != (netz.Proxyvorgabe{}) {
		t.Errorf("verschlossen kam die Vorgabe %+v", vorgabe)
	}
}

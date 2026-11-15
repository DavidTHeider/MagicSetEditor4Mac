// Jeder Endpunkt prüft seine Eingaben.
//
// **Eine eigene Datei**, weil `durchreichen_test.go` mit der wachsenden
// Fallliste über die Zeilengrenze ginge (§2). Der Schnitt folgt der Frage:
// dort **kommt der Fehler durch**, hier **wird abgewiesen**.
package rest_test

import (
	"net/http"
	"testing"
)

// die Fälle entstehen daraus.
func TestJederEndpunktPrueftSeineEingaben(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		methode string
		weg     string
		rumpf   string
	}{
		{"feld: unbrauchbares Set", http.MethodPut, "/sets/x/karten/2/felder/name", `{"wert":"a"}`},
		{"feld: unbrauchbare Karte", http.MethodPut, "/sets/1/karten/x/felder/name", `{"wert":"a"}`},
		{"feld: Rumpf ist kein JSON", http.MethodPut, "/sets/1/karten/2/felder/name", `{`},
		{"sichern: unbrauchbares Set", http.MethodPost, "/sets/x/sicherung", `{"ziel":"/tmp/a"}`},
		{"sichern: Rumpf ist kein JSON", http.MethodPost, "/sets/1/sicherung", `{`},
		{"zeichnen: unbrauchbare Karte", http.MethodPost, "/karten/x/bild", `{"ziel":"/tmp/a"}`},
		{"zeichnen: Rumpf ist kein JSON", http.MethodPost, "/karten/2/bild", `{`},
		{"bogen: unbrauchbares Set", http.MethodPost, "/sets/x/bogen", `{"ziel":"/tmp/a"}`},
		{"bogen: Rumpf ist kein JSON", http.MethodPost, "/sets/1/bogen", `{`},
		{"mtgjson lesen: Rumpf ist kein JSON", http.MethodPost, "/mtgjson", `{`},
		{
			"mtgjson schreiben: unbrauchbares Set", http.MethodPost,
			"/sets/x/mtgjson", `{"spiel":"magic","ziel":"/tmp/a"}`,
		},
		{"mtgjson schreiben: Rumpf ist kein JSON", http.MethodPost, "/sets/1/mtgjson", `{`},
		{"liste deuten: Rumpf ist kein JSON", http.MethodPost, "/sammlung/liste/deuten", `{`},
		{
			// **Die Partieeinfuhr** — PLAY-03.AK-05 (#1711).
			"partie einlesen: Rumpf ist kein JSON", http.MethodPost,
			"/partien/einfuhr", `{`,
		},
		{
			"liste einlesen: Rumpf ist kein JSON", http.MethodPost,
			"/sammlung/liste/einlesen", `{`,
		},
		{
			"sammlung einlesen: Rumpf ist kein JSON", http.MethodPost,
			"/sammlung/einlesen", `{`,
		},
		{
			"liste vorschauen: Rumpf ist kein JSON", http.MethodPost,
			"/sammlung/liste/vorschau", `{`,
		},
		{
			"sammlung vorschauen: Rumpf ist kein JSON", http.MethodPost,
			"/sammlung/vorschau", `{`,
		},
		{
			"dubletten: unbrauchbare Zahl", http.MethodGet,
			"/sammlung/dubletten?behalten=viele", "",
		},
		{"masse ändern: Rumpf ist kein JSON", http.MethodPost, "/sammlung/masse", `{`},
		{
			"masse löschen: Rumpf ist kein JSON", http.MethodPost,
			"/sammlung/masse/loeschen", `{`,
		},
		{
			"masse Preise: Rumpf ist kein JSON", http.MethodPost,
			"/sammlung/masse/preise", `{`,
		},
		{"vorlagen: unbrauchbares Set", http.MethodGet, "/sets/x/vorlagen", ""},
		{"paketbedarf: unbrauchbares Set", http.MethodGet, "/sets/x/paketbedarf", ""},
		{
			"vorlage ausführen: unbrauchbares Set", http.MethodPost, "/sets/x/vorlagen",
			`{"paket":"p","ziel":"/tmp/a"}`,
		},
		{"vorlage ausführen: Rumpf ist kein JSON", http.MethodPost, "/sets/1/vorlagen", `{`},
		{
			"vorlagenoption: Rumpf ist kein JSON", http.MethodPut,
			"/vorlagen/p/optionen/f", `{`,
		},
		{
			"kartenbild: unbrauchbares Set", http.MethodPut,
			"/sets/x/karten/2/felder/kunst/bild", `{"pfad":"/tmp/a.png"}`,
		},
		{
			"kartenbild: unbrauchbare Karte", http.MethodPut,
			"/sets/1/karten/x/felder/kunst/bild", `{"pfad":"/tmp/a.png"}`,
		},
		{
			"kartenbild: Rumpf ist kein JSON", http.MethodPut,
			"/sets/1/karten/2/felder/kunst/bild", `{`,
		},
		{"paket einbinden: Rumpf ist kein JSON", http.MethodPut, "/pakete/eingebunden", `{`},
		{"paket ersetzen: Rumpf ist kein JSON", http.MethodPut, "/pakete/ersetzt", `{`},
		{"bezug aktualisieren: Rumpf ist kein JSON", http.MethodPost, "/paketbezuege/M15-Magic-Pack/aktualisierung", `{`},
		{"paket anlegen: Rumpf ist kein JSON", http.MethodPost, "/pakete/neu", `{`},
		{"paket kopieren: Rumpf ist kein JSON", http.MethodPost, "/pakete/mse-game/probe/kopie", `{`},
		{"pakete weitergeben: Rumpf ist kein JSON", http.MethodPost, "/pakete/weitergabe", `{`},
		{"installer einlesen: Rumpf ist kein JSON", http.MethodPost, "/pakete/eingelesen", `{`},
		{"nutzer anlegen: Rumpf ist kein JSON", http.MethodPost, "/nutzer", `{`},
		{
			"Profilbild: Rumpf ist kein JSON", http.MethodPut,
			"/nutzer/1/bild", `{`,
		},
		{
			"Nutzereinstellung: Rumpf ist kein JSON", http.MethodPut,
			"/nutzer/1/einstellungen/barrierefreiheit.kontrast", `{`,
		},
		{
			"Profilbadge: Rumpf ist kein JSON", http.MethodPut,
			"/profile/1/badge", `{`,
		},
		{
			"profil anlegen: Rumpf ist kein JSON", http.MethodPost,
			"/nutzer/1/profile", `{`,
		},
		{"pakete übernehmen: Rumpf ist kein JSON", http.MethodPost, "/pakete/uebernommen", `{`},
		{"skript: unbrauchbare Karte", http.MethodPost, "/karten/x/skript", `{"ausdruck":"1"}`},
		{"skript: Rumpf ist kein JSON", http.MethodPost, "/karten/2/skript", `{`},
		{"skriptnamen: unbrauchbare Karte", http.MethodGet, "/karten/x/skriptnamen", ""},
		{
			"paketdatei schreiben: Rumpf ist kein JSON", http.MethodPut,
			"/pakete/probe.mse-style/dateien/style", `{`,
		},
		{"paketdatei anlegen: Rumpf ist kein JSON", http.MethodPost, "/pakete/probe.mse-style/dateien", `{`},
		{"paketdatei umbenennen: Rumpf ist kein JSON", http.MethodPatch, "/pakete/probe.mse-style/dateien/neu", `{`},

		// **Die Schlüsselwörter** (#1062) — zwei Kennungen und ein Rumpf.
		{
			"schluesselwoerter: unbrauchbares Set", http.MethodGet,
			"/sets/x/schluesselwoerter", "",
		},

		// **Die Marken** (#1208) — zwei Kennungen und ein Wegabschnitt.
		{"marken: unbrauchbares Set", http.MethodGet, "/sets/x/marken", ""},
		{
			"kartenmarken: unbrauchbares Set", http.MethodGet,
			"/sets/x/karten/2/marken", "",
		},
		{
			"kartenmarken: unbrauchbare Karte", http.MethodGet,
			"/sets/1/karten/x/marken", "",
		},
		{
			"marke setzen: unbrauchbares Set", http.MethodPut,
			"/sets/x/karten/2/marken/fertig", "",
		},
		{
			"marke setzen: unbrauchbare Karte", http.MethodPut,
			"/sets/1/karten/x/marken/fertig", "",
		},
		{
			"marke setzen: leere Marke", http.MethodPut,
			"/sets/1/karten/2/marken/%20", "",
		},
		{
			"marke entfernen: unbrauchbares Set", http.MethodDelete,
			"/sets/x/karten/2/marken/fertig", "",
		},
		{
			"marke entfernen: unbrauchbare Karte", http.MethodDelete,
			"/sets/1/karten/x/marken/fertig", "",
		},
		{
			"schluesselwort anlegen: unbrauchbares Set", http.MethodPost,
			"/sets/x/schluesselwoerter", `{}`,
		},
		{
			"schluesselwort setzen: unbrauchbares Set", http.MethodPut,
			"/sets/x/schluesselwoerter/2/felder/muster", `{"wert":"B"}`,
		},
		{
			"schluesselwort setzen: unbrauchbares Wort", http.MethodPut,
			"/sets/1/schluesselwoerter/x/felder/muster", `{"wert":"B"}`,
		},
		{
			"schluesselwort setzen: Rumpf ist kein JSON", http.MethodPut,
			"/sets/1/schluesselwoerter/2/felder/muster", `{`,
		},
		{
			"schluesselwort loeschen: unbrauchbares Wort", http.MethodDelete,
			"/sets/1/schluesselwoerter/x", "",
		},

		// **Die Wortlisten** (#1074) — drei Kennungen und ein Rumpf.
		{
			"wortlisten: unbrauchbare Karte", http.MethodGet,
			"/karten/x/felder/type/wortlisten", "",
		},
		{
			"wortwahl: unbrauchbares Set", http.MethodPost,
			"/sets/x/karten/2/felder/type/wortwahl", `{"bereich":0,"weg":[0]}`,
		},
		{
			"wortwahl: unbrauchbare Karte", http.MethodPost,
			"/sets/1/karten/x/felder/type/wortwahl", `{"bereich":0,"weg":[0]}`,
		},
		{
			"wortwahl: Rumpf ist kein JSON", http.MethodPost,
			"/sets/1/karten/2/felder/type/wortwahl", `{`,
		},

		// **Die Zwischenablage** (#1168). Sie hat zwei eigene Wachen: eine
		// leere Kartenliste ist keine Auswahl, und ein leerer Inhalt ist keine
		// Ablage — beides käme sonst als geglückter Aufruf zurück, der nichts
		// getan hat.
		{
			"ablage kopieren: unbrauchbares Set", http.MethodPost,
			"/sets/x/ablage/kopieren", `{"karten":[2]}`,
		},
		{
			"ablage kopieren: keine Karte genannt", http.MethodPost,
			"/sets/1/ablage/kopieren", `{"karten":[]}`,
		},
		{
			"ablage ausschneiden: keine Karte genannt", http.MethodPost,
			"/sets/1/ablage/ausschneiden", `{}`,
		},
		{
			"ablage kopieren: Rumpf ist kein JSON", http.MethodPost,
			"/sets/1/ablage/kopieren", `{`,
		},
		{
			"ablage einfuegen: unbrauchbares Set", http.MethodPost,
			"/sets/x/ablage/einfuegen", `{"inhalt":"game_name: probe\n"}`,
		},
		{
			"ablage einfuegen: kein Inhalt", http.MethodPost,
			"/sets/1/ablage/einfuegen", `{"inhalt":"   "}`,
		},
		{
			"ablage einfuegen: Rumpf ist kein JSON", http.MethodPost,
			"/sets/1/ablage/einfuegen", `{`,
		},

		// **Das Ersetzen** (#1168). Eine Frage ohne Suchtext träfe jede
		// Stelle jedes Feldes — sie wird abgewiesen, bevor gelesen wird.
		{
			"ersetzung: unbrauchbares Set", http.MethodPost,
			"/sets/x/ersetzung", `{"text":"a","ersatz":"b"}`,
		},
		{
			"ersetzung: Rumpf ist kein JSON", http.MethodPost,
			"/sets/1/ersetzung", `{`,
		},
		{
			"ersetzung: kein Suchtext", http.MethodPost,
			"/sets/1/ersetzung", `{"text":"  ","ersatz":"b"}`,
		},
		{
			"ersetzung vorschau: kein Suchtext", http.MethodPost,
			"/sets/1/ersetzung/vorschau", `{"ersatz":"b"}`,
		},

		// **Die Gestalt der Kartenliste** (#1168).
		{
			"kartenliste: unbrauchbares Set", http.MethodGet,
			"/sets/x/kartenliste", "",
		},
		{
			"listenspalten: unbrauchbares Set", http.MethodPut,
			"/sets/x/kartenliste/spalten", `{"spalten":[]}`,
		},
		{
			"listenspalten: Rumpf ist kein JSON", http.MethodPut,
			"/sets/1/kartenliste/spalten", `{`,
		},
		{
			"listensortierung: unbrauchbares Set", http.MethodPut,
			"/sets/x/kartenliste/sortierung", `{"feld":"name"}`,
		},
		{
			"listensortierung: Rumpf ist kein JSON", http.MethodPut,
			"/sets/1/kartenliste/sortierung", `{`,
		},

		// **Die Kartenskripte** (#1337) — eine Kennung an beiden Wegen, und
		// ein Rumpf am zweiten.
		{
			"kartenskripte: unbrauchbares Set", http.MethodGet,
			"/sets/x/kartenskripte", "",
		},
		{
			"kartenskript ausführen: unbrauchbares Set", http.MethodPost,
			"/sets/x/kartenskripte", `{"name":"Zwei Farben"}`,
		},
		{
			"kartenskript ausführen: Rumpf ist kein JSON", http.MethodPost,
			"/sets/1/kartenskripte", `{`,
		},

		// **Die Auszählung** (#1338) — eine Kennung an beiden Wegen.
		{
			"auszähldimensionen: unbrauchbares Set", http.MethodGet,
			"/sets/x/auszaehlung", "",
		},
		{
			"auszählung: unbrauchbares Set", http.MethodGet,
			"/sets/x/auszaehlung/farbe", "",
		},

		// **Die Parameterhilfen** (#1342) — zwei Kennungen.
		{
			"parameterhilfen: unbrauchbares Set", http.MethodGet,
			"/sets/x/schluesselwoerter/2/parameter", "",
		},
		{
			"parameterhilfen: unbrauchbares Wort", http.MethodGet,
			"/sets/1/schluesselwoerter/x/parameter", "",
		},

		// **Das Nachzeichnen** (#1340) — eine Kennung und ein Rumpf.
		{
			"symbol nachzeichnen: unbrauchbares Set", http.MethodPost,
			"/sets/x/symbol-nachzeichnen", `{"bildpfad":"/tmp/k.png"}`,
		},
		{
			"symbol nachzeichnen: Rumpf ist kein JSON", http.MethodPost,
			"/sets/1/symbol-nachzeichnen", `{`,
		},

		// **Die Autoersetzungen** (#1341) — eine Kennung an fünf Wegen, dazu
		// zwei Rümpfe.
		{
			"autoersetzungen: unbrauchbares Set", http.MethodGet,
			"/sets/x/autoersetzungen", "",
		},
		{
			"autoersetzung setzen: unbrauchbares Set", http.MethodPut,
			"/sets/x/autoersetzungen/(W)", `{"ersatz":"y"}`,
		},
		{
			"autoersetzung setzen: Rumpf ist kein JSON", http.MethodPut,
			"/sets/1/autoersetzungen/(W)", `{`,
		},
		{
			"autoersetzung löschen: unbrauchbares Set", http.MethodDelete,
			"/sets/x/autoersetzungen/(W)", "",
		},
		{
			"autoersetzungen zurücksetzen: unbrauchbares Set", http.MethodDelete,
			"/sets/x/autoersetzungen", "",
		},
		{
			"autoersetzung schalten: unbrauchbares Set", http.MethodPut,
			"/sets/x/autoersetzungen", `{"an":true}`,
		},
		{
			"autoersetzung schalten: Rumpf ist kein JSON", http.MethodPut,
			"/sets/1/autoersetzungen", `{`,
		},

		// **Die Regelfreigabe** (G4, #1461) — ihre Angaben stehen in der
		// Abfrage, und `sitze` ist eine Zahl. Ohne diese Prüfung fragte sie
		// mit einer Null und bekäme eine Absage, deren Grund mit der Frage
		// nichts zu tun hat.
		{
			"regelfreigabe: sitze ist keine Zahl", http.MethodGet,
			"/regelfreigabe?spiel=magic&strenge=sandkasten&sitze=zwei", "",
		},
		{
			"regelfreigabe: sitze fehlt", http.MethodGet,
			"/regelfreigabe?spiel=magic&strenge=sandkasten", "",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t),
				endpunkt{methode: fall.methode, weg: fall.weg, rumpf: fall.rumpf}, gueltig)
			pruefeAbweisung(t, antwort, http.StatusBadRequest, "validierung")
		})
	}
}

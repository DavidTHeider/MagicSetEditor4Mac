// Eine Gegenstelle, die Scryfall nachstellt
// (#1028).
//
// # Warum sie hier steht und nicht im Klienten
//
// Die Flächenläufe starten `mse`, `mse-mcp` und `mse-rest` als **eigene
// Prozesse**. Eine Veränderliche im selben Programm erreichen sie nicht — die
// Adresse kommt über `MSE_SCRYFALL_ADRESSE` (`app.Scryfallumgebung`).
//
// # Und warum kein Aufruf hinausgeht
//
// **Zwei Dinge gehören an jeden solchen Lauf** (§3, seit #1021): eine
// umgelenkte Adresse **und** ein eigenes Zuhause. Das eine ohne das andere hat
// in dieser Phase zweimal wirklich 25 MB von Scryfall geladen. Hier setzt
// [scryfallumgebung] beides — die Adresse zeigt auf diese Gegenstelle, und
// `MSE_BESTAND` liegt in einem `t.TempDir()`, aus dem auch der
// Zwischenspeicher entsteht (`app.Scryfallverzeichnis`).
package flaechen_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// genauerName und setangabe erkennen die beiden Teile einer Zeilenfrage.
//
// **Über die kodierte Gestalt**: die Frage steht in `RawQuery`, also `%21%22`
// für `!"` und `e%3A` für `e:`. Ein Vergleich gegen die rohen Zeichen ginge
// vorbei — geprüft am 2026-08-30 an der Anfrage, die `frageFuer` baut.
var (
	genauerName = regexp.MustCompile(`%21%22|!"`)
	setangabe   = regexp.MustCompile(`(^|\+|%20)e(%3A|:)`)
)

// probekarte ist eine Karte, wie Scryfall sie schickt — selbst geschrieben
// (§11), nicht abgeschrieben.
const probekarte = `{
  "object": "card",
  "id": "00000000-0000-4000-8000-000000000001",
  "name": "%s",
  "lang": "en",
  "layout": "normal",
  "mana_cost": "{2}{W}",
  "type_line": "Creature — Testling",
  "oracle_text": "{T}: Add one mana of any color.",
  "power": "2",
  "toughness": "3",
  "rarity": "rare",
  "set": "prf",
  "set_name": "Prüfsatz",
  "collector_number": "%d",
  "artist": "Selbst Gemalt",
  "image_uris": {"normal": "%s/bild.png", "art_crop": "%s/bild.png"},
  "scryfall_uri": "https://beispiel.invalid/karte/%d"
}`

// sammelvorlage ist derselbe Kartensatz unter **erfragtem Set** (#1348).
//
// **Der echte Dienst gibt die Karte, nach der gefragt wurde** — ihr `set` und
// `collector_number` sind die der Kennung. Eine Gegenstelle, die eine Karte
// aus einem anderen Set zurückgäbe, ließe die Zuordnung über die Kennung
// scheitern und sähe aus wie ein Fehler des Programms.
const sammelvorlage = `{
  "object": "card",
  "id": "00000000-0000-4000-8000-000000000001",
  "name": "Erste Probe",
  "lang": "en",
  "layout": "normal",
  "mana_cost": "{2}{W}",
  "type_line": "Creature — Testling",
  "oracle_text": "{T}: Add one mana of any color.",
  "power": "2",
  "toughness": "3",
  "rarity": "rare",
  "set": "%s",
  "set_name": "Prüfsatz",
  "collector_number": "%s",
  "artist": "Selbst Gemalt",
  "image_uris": {"normal": "%s/bild.png", "art_crop": "%s/bild.png"},
  "scryfall_uri": "https://beispiel.invalid/karte/1"
}`

// druckvorlage ist derselbe Kartensatz unter **erfragter** Kennung (#1167).
//
// **Eine eigene Vorlage und keine zweite Karte**: was sie von [probekarte]
// unterscheidet, ist genau die Kennung — und das ist die Angabe, um die es
// beim Verweis geht.
const druckvorlage = `{
  "object": "card",
  "id": "%s",
  "name": "Erste Probe",
  "lang": "en",
  "layout": "normal",
  "mana_cost": "{2}{W}",
  "type_line": "Creature — Testling",
  "oracle_text": "{T}: Add one mana of any color.",
  "power": "2",
  "toughness": "3",
  "rarity": "rare",
  "set": "prf",
  "set_name": "Prüfsatz",
  "collector_number": "1",
  "artist": "Selbst Gemalt",
  "image_uris": {"normal": "%s/bild.png", "art_crop": "%s/bild.png"},
  "scryfall_uri": "https://beispiel.invalid/karte/1",
  "prices": {"eur": "1.50", "eur_foil": "4.25", "usd": "1.80",
             "usd_foil": "5.10", "tix": "0.03"}
}`

// scryfallgegenstelle beantwortet die Wege, die die schreibenden Fähigkeiten
// gehen — und **nur** die.
//
// **Was sie nicht ist**: eine Nachbildung von Scryfall. Sie antwortet auf
// `/cards/named`, `/cards/:id`, `/cards/:id/rulings`, `/cards/search`,
// `/sets/{code}`, `/symbology`, `/bulk-data`, die Sammeldatei und ein Bild.
// Alles andere endet mit 404, und das ist Absicht: ein Weg, der etwas anderes
// fragt, soll auffallen.
//
// **Die Preise hängen an [druckvorlage]** (#1169): sie kommen aus
// `/cards/:id`, nicht aus einem eigenen Weg — so hat Scryfall es, und so
// liest es `scryfall.PreiseFuerDruck`.
func scryfallgegenstelle(t *testing.T) *httptest.Server {
	t.Helper()

	return gefragteGegenstelle(t, nil)
}

// gefragteGegenstelle ist dieselbe und zählt mit, wie oft sie gefragt wurde.
//
// # Wofür der Zähler da ist
//
// **Er belegt, dass die Umlenkung wirkt.** Ohne ihn ließe sich „kein Aufruf
// ging ans echte Netz" nur behaupten: ein Lauf, dessen Adresse nicht ankommt,
// fragte Scryfall und wäre trotzdem grün oder rot — und man sähe nicht, welches
// von beidem er geprüft hat. Die Gegenprobe dazu ginge ins echte Netz und wird
// deshalb **nicht** gefahren; der Zähler ist ihr Ersatz.
func gefragteGegenstelle(t *testing.T, zaehler *atomic.Int64) *httptest.Server {
	t.Helper()

	// **Unstarted und erst danach gestartet.** Die Antworten nennen die
	// eigene Adresse — das Bild, die Sammeldatei —, und die steht erst fest,
	// wenn der Lauscher existiert. Ein `httptest.NewServer(nil)` mit einem
	// später gesetzten `Config.Handler` täte es auch, **aber**: er bedient
	// dann schon, und der Renner las den Zeiger, während dieser Test ihn
	// schrieb. Gesehen unter `go test -race`, nicht überlegt.
	server := httptest.NewUnstartedServer(nil)
	t.Cleanup(server.Close)

	grund := "http://" + server.Listener.Addr().String()

	weg := http.NewServeMux()
	weg.HandleFunc("GET /cards/named", func(schreiben http.ResponseWriter, anfrage *http.Request) {
		name := anfrage.URL.Query().Get("exact")
		if name == "" || strings.EqualFold(name, "Gibtsnicht") {
			schreiben.WriteHeader(http.StatusNotFound)
			schreibeText(t, schreiben, `{"object":"error","code":"not_found",`+
				`"status":404,"details":"No card found."}`)

			return
		}

		schreibeText(t, schreiben, fmt.Sprintf(probekarte, name, 1, grund, grund, 1))
	})

	weg.HandleFunc("GET /cards/search", func(schreiben http.ResponseWriter, anfrage *http.Request) {
		if strings.Contains(anfrage.URL.RawQuery, "gibtsnicht") {
			schreiben.WriteHeader(http.StatusNotFound)
			schreibeText(t, schreiben, `{"object":"error","code":"not_found",`+
				`"status":404,"details":"No cards found."}`)

			return
		}

		// **Ein genauer Name ohne `unique=prints` gibt genau einen** (#1292).
		//
		// Nachgemessen am 2026-08-30: `!"Counterspell"` gibt `total_cards` 1,
		// dasselbe mit `&unique=prints` gibt 72. Scryfalls Vorgabe ist
		// `unique=cards` — eine Zeile je **Karte**.
		//
		// **Diese Zeile ist der Grund, warum die Gegenstelle den Befund
		// verdeckt hat.** Sie gab auf jede Suche zwei Treffer, gleich wie
		// gefragt wurde; die Listeneinfuhr sah damit eine Mehrdeutigkeit, die
		// es beim echten Dienst nicht gab. Gefunden hat es die laufende
		// Brücke — zwei grüne Seiten einer Grenze sind kein Nachweis für die
		// Grenze.
		//
		// **Und der genaue Name gehört in die Bedingung.** Der erste Anlauf
		// prüfte nur `unique=prints` und nahm damit der **Kartensuche** ihren
		// zweiten Treffer: `t:creature` gibt zwei verschiedene Karten, und
		// `unique=cards` ist dort genau richtig. Zwei Fragen, zwei Antworten —
		// die Bedingung muss sie unterscheiden.
		if genauerName.MatchString(anfrage.URL.RawQuery) &&
			!strings.Contains(anfrage.URL.RawQuery, "unique=prints") {
			schreibeText(t, schreiben, fmt.Sprintf(
				`{"object":"list","has_more":false,"total_cards":1,"data":[%s]}`,
				fmt.Sprintf(probekarte, "Erste Probe", 1, grund, grund, 1)))

			return
		}

		// **Ein genauer Name samt Setangabe gibt genau einen** (#1292).
		//
		// So verhält sich der echte Dienst, nachgemessen am 2026-08-30:
		// `!"Lightning Bolt" e:m11 cn:149` gibt `total_cards` 1,
		// `!"Counterspell"` gibt 72. Eine Gegenstelle, die **immer** zwei
		// gäbe, ließe den eindeutigen Fall der Listeneinfuhr unbetreten — und
		// der ist der einzige, der ohne Rückfrage schreibt.
		//
		// **Beides zusammen und nicht `e:` allein.** Der erste Anlauf prüfte
		// nur die Setangabe — und nahm damit `TestEinenSatzEinlesenUeberJede
		// Flaeche` die zweite Karte: `Setkarten` fragt `e:<code>` **ohne**
		// Namen und meint den ganzen Satz. Das ist derselbe Unterschied wie im
		// echten Dienst, und er gehört in die Bedingung.
		if genauerName.MatchString(anfrage.URL.RawQuery) &&
			setangabe.MatchString(anfrage.URL.RawQuery) {
			schreibeText(t, schreiben, fmt.Sprintf(
				`{"object":"list","has_more":false,"total_cards":1,"data":[%s]}`,
				fmt.Sprintf(probekarte, "Erste Probe", 1, grund, grund, 1)))

			return
		}

		schreibeText(t, schreiben, fmt.Sprintf(
			`{"object":"list","has_more":false,"total_cards":2,"data":[%s,%s]}`,
			fmt.Sprintf(probekarte, "Erste Probe", 1, grund, grund, 1),
			fmt.Sprintf(probekarte, "Zweite Probe", 2, grund, grund, 2)))
	})

	// **Die Sammelanfrage** (#1348): `POST /cards/collection`, bis zu 75
	// Kennungen in einem Rumpf.
	//
	// # Was diese Gegenstelle vom echten Dienst nachbildet
	//
	// Sie ordnet über die **Kennung** zu und nicht über die Stelle — genau
	// wie es Scryfall verlangt: „cards that aren't found will throw off the
	// mapping of request identifiers to results." Eine Gegenstelle, die die
	// Antwort in der Reihenfolge der Anfrage zurückgäbe und Fehlendes
	// ausließe, ließe den Fehler unbetreten, gegen den die Zuordnung steht.
	//
	// **Die Bedingung ist dieselbe wie bei der Suche daneben**: eine Kennung
	// mit Setangabe trifft genau einen Druck. Das ist der Fall, den die
	// Listeneinfuhr ohne Rückfrage schreibt, und eine Gegenstelle, die hier
	// großzügiger wäre als der echte Dienst, ließe ihn unbetreten (§3).
	//
	// **`gibtsnicht` findet nichts** — so wie `/cards/search` daneben mit 404
	// antwortet. Für eine Sammelanfrage ist das kein Status, sondern ein
	// Eintrag in `not_found`.
	weg.HandleFunc("POST /cards/collection", func(schreiben http.ResponseWriter, anfrage *http.Request) {
		var rumpf struct {
			Kennungen []struct {
				Name    string `json:"name"`
				Setcode string `json:"set"`
				Nummer  string `json:"collector_number"`
			} `json:"identifiers"`
		}

		if err := json.NewDecoder(anfrage.Body).Decode(&rumpf); err != nil {
			schreiben.WriteHeader(http.StatusBadRequest)

			return
		}

		var (
			karten  []string
			fehlend []string
		)

		for _, kennung := range rumpf.Kennungen {
			passt := kennung.Setcode != "" &&
				!strings.Contains(strings.ToLower(kennung.Name), "gibtsnicht")
			if !passt {
				fehlend = append(fehlend, fmt.Sprintf(
					`{"name":%q,"set":%q,"collector_number":%q}`,
					kennung.Name, kennung.Setcode, kennung.Nummer))

				continue
			}

			nummer := kennung.Nummer
			if nummer == "" {
				nummer = "1"
			}

			karten = append(karten, fmt.Sprintf(
				sammelvorlage, kennung.Setcode, nummer, grund, grund))
		}

		schreibeText(t, schreiben, fmt.Sprintf(
			`{"object":"list","data":[%s],"not_found":[%s]}`,
			strings.Join(karten, ","), strings.Join(fehlend, ",")))
	})

	// **Ein einzelner Druck über `/cards/:id`** (#1167). Der Verweis holt ihn
	// darüber; ohne diesen Weg endete jeder Verweislauf mit 404, und das sähe
	// aus wie ein Fehler der Anwendung.
	//
	// **Die Antwort nennt die erfragte Kennung.** Der Klient vergleicht sie —
	// eine feste Kennung ließe den Vergleich immer scheitern, und der Lauf
	// prüfte dann die Wache statt den Weg.
	weg.HandleFunc("GET /cards/{id}", func(schreiben http.ResponseWriter, anfrage *http.Request) {
		kennung := anfrage.PathValue("id")
		if kennung == "gibtsnicht" {
			schreiben.WriteHeader(http.StatusNotFound)
			schreibeText(t, schreiben, `{"object":"error","code":"not_found",`+
				`"status":404,"details":"No card found."}`)

			return
		}

		schreibeText(t, schreiben, fmt.Sprintf(druckvorlage, kennung, grund, grund))
	})

	weg.HandleFunc("GET /sets/{code}", func(schreiben http.ResponseWriter, anfrage *http.Request) {
		schreibeText(t, schreiben, fmt.Sprintf(
			`{"object":"set","id":"00000000-0000-4000-8000-000000000010",`+
				`"code":%q,"name":"Prüfsatz","set_type":"expansion",`+
				`"released_at":"2026-08-24","card_count":2,"digital":false}`,
			anfrage.PathValue("code")))
	})

	weg.HandleFunc("GET /bulk-data", func(schreiben http.ResponseWriter, _ *http.Request) {
		schreibeText(t, schreiben, fmt.Sprintf(
			`{"object":"list","has_more":false,"data":[{"object":"bulk_data",`+
				`"type":"oracle_cards","name":"Oracle Cards","description":"Eine Probe",`+
				`"jsonl_download_uri":%q,"compressed_size":64,`+
				`"updated_at":"2026-08-24T09:00:00.000+00:00"}]}`,
			grund+"/oracle.jsonl.gz"))
	})

	weg.HandleFunc("GET /oracle.jsonl.gz", func(schreiben http.ResponseWriter, _ *http.Request) {
		schreibeBytes(t, schreiben, gepackteSammeldatei(t, grund))
	})

	weg.HandleFunc("GET /bild.png", func(schreiben http.ResponseWriter, _ *http.Request) {
		schreibeBytes(t, schreiben, winzigesPNG(t))
	})

	// **Die Regelhinweise zu einem Druck** (#1169). Ohne diesen Weg endete
	// jeder Regellauf mit 404, und das sähe aus wie ein Fehler der Anwendung
	// statt wie eine fehlende Vorrichtung.
	weg.HandleFunc("GET /cards/{id}/rulings", func(schreiben http.ResponseWriter, _ *http.Request) {
		schreibeText(t, schreiben, `{"object":"list","has_more":false,"data":[`+
			`{"object":"ruling","oracle_id":"probe","source":"wotc",`+
			`"published_at":"2020-01-01","comment":"Ein Hinweis aus der Probe."}]}`)
	})

	// **Die Symbolliste** — dagegen prüft `manakosten`, ob ein Symbol es
	// wirklich gibt. Zwei Symbole reichen: eines, das vorkommt, und keines
	// mehr, damit ein drittes im Kosten als Befund auffällt.
	weg.HandleFunc("GET /symbology", func(schreiben http.ResponseWriter, _ *http.Request) {
		schreibeText(t, schreiben, `{"object":"list","has_more":false,"data":[`+
			`{"object":"card_symbol","symbol":"{R}","english":"one red mana",`+
			`"represents_mana":true,"cmc":1.0},`+
			`{"object":"card_symbol","symbol":"{2}","english":"two generic mana",`+
			`"represents_mana":true,"cmc":2.0}]}`)
	})

	server.Config.Handler = gezaehlt(weg, zaehler)
	server.Start()

	return server
}

// gezaehlt legt einen Zähler vor einen Weg — oder gibt ihn unverändert.
func gezaehlt(weg http.Handler, zaehler *atomic.Int64) http.Handler {
	if zaehler == nil {
		return weg
	}

	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		zaehler.Add(1)
		weg.ServeHTTP(schreiben, anfrage)
	})
}

// scryfallumgebung gibt die Umgebung für einen Lauf: eigener Bestand, eigenes
// Zuhause, umgelenkte Adresse.
func scryfallumgebung(t *testing.T, server *httptest.Server) []string {
	t.Helper()

	return append(eigenerBestand(t),
		app.Scryfallumgebung+"="+server.URL,
		"HOME="+t.TempDir())
}

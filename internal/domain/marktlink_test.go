package domain_test

// **Wohin eine Karte außerhalb dieses Programms führt** — Werkstück `T8` von
// Phase 5.6
// (#1473).
//
// # Was hier gemessen wird
//
// **Dass die Reihenfolge gilt**: ein Produktlink schlägt eine Suche, eine
// Suche schlägt die Startseite — und je Anbieter gibt es genau einen Weg.
//
// **Dass ein Produktlink nur aus einer belegten Kennung entsteht.** Ein
// Konzept nennt mehrere Drucke, ein Kunstwerk noch mehr; wer damit baute,
// führte auf irgendeinen davon.
//
// **Und dass keine Adresse hinausgeht, die nicht hinausgehen darf** — das
// ist die Hälfte dieses Werkstücks: Lookalike, Unterdomäne, Benutzerangabe,
// fremder Port, fremdes Protokoll, Geheimnis in der Abfrage.

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// linkVon sucht den Link eines Anbieters.
func linkVon(links []domain.Marktlink, anbieter domain.Linkanbieter) domain.Marktlink {
	for _, link := range links {
		if link.Anbieter == anbieter {
			return link
		}
	}

	return domain.Marktlink{}
}

func TestJederAnbieterBekommtGenauEinenWeg(t *testing.T) {
	t.Parallel()

	links := domain.Marktlinks(nil, "Lightning Bolt")

	if len(links) != len(domain.Linkanbieterliste()) {
		t.Fatalf("es kamen %d Links für %d Anbieter",
			len(links), len(domain.Linkanbieterliste()))
	}

	gesehen := map[domain.Linkanbieter]int{}
	for _, link := range links {
		gesehen[link.Anbieter]++

		// **Jede Adresse hält die Prüfung** — sonst wäre die Prüfung ein
		// Wächter über etwas anderes als das, was gebaut wird.
		if fehler := domain.PruefeMarktlink(link); fehler != nil {
			t.Errorf("%s: %v", link.Anbieter, fehler)
		}
	}

	for anbieter, zahl := range gesehen {
		if zahl != 1 {
			t.Errorf("%s kam %dmal", anbieter, zahl)
		}
	}
}

// **Ohne Kennung und ohne Namen bleibt die Startseite.**
func TestOhneAngabeBleibtDieStartseite(t *testing.T) {
	t.Parallel()

	for _, link := range domain.Marktlinks(nil, "   ") {
		if link.Art != domain.LinkStartseite {
			t.Errorf("%s gab %q", link.Anbieter, link.Art)
		}

		if link.Herkunft != domain.HerkunftKeine {
			t.Errorf("%s gab die Herkunft %q", link.Anbieter, link.Herkunft)
		}
	}
}

// **Ein Name gibt eine Suche — überall außer bei Gatherer.**
//
// **Gatherer ist kein Marktplatz.** Es kennt eine Karte an ihrer
// Multiverse-ID und an nichts sonst; ohne sie ist es nicht verfügbar, und
// eine Suche nach dem Namen wäre die Erfindung, die das Issue ausschließt.
func TestEinNameGibtEineSucheAusserBeiGatherer(t *testing.T) {
	t.Parallel()

	links := domain.Marktlinks(nil, "Lightning Bolt")

	for _, link := range links {
		if link.Anbieter == domain.LinkGatherer {
			if link.Art != domain.LinkStartseite {
				t.Errorf("Gatherer sucht: %q", link.Adresse)
			}

			continue
		}

		if link.Art != domain.LinkSuche {
			t.Errorf("%s gab %q statt einer Suche", link.Anbieter, link.Art)
		}

		// **Der Name steht in der Abfrage** — und zwar kodiert.
		gelesen, fehler := url.Parse(link.Adresse)
		if fehler != nil {
			t.Fatalf("%s: %v", link.Anbieter, fehler)
		}

		if !strings.Contains(gelesen.RawQuery, "Lightning") {
			t.Errorf("%s sucht nicht nach dem Namen: %q",
				link.Anbieter, link.Adresse)
		}
	}
}

// **Eine belegte Kennung schlägt die Suche.**
func TestEineBelegteKennungSchlaegtDieSuche(t *testing.T) {
	t.Parallel()

	kennungen := []domain.Druckkennung{
		{System: domain.SystemTCGplayer, Semantik: domain.SemantikProdukt, Wert: "33632"},
		{System: domain.SystemCardTrader, Semantik: domain.SemantikProdukt, Wert: "2076"},
		{System: domain.SystemGatherer, Semantik: domain.SemantikDruck, Wert: "191089"},
		{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "241866"},
	}

	links := domain.Marktlinks(kennungen, "Lightning Bolt")

	for _, anbieter := range []domain.Linkanbieter{
		domain.LinkTCGplayer, domain.LinkCardTrader, domain.LinkGatherer,
	} {
		link := linkVon(links, anbieter)
		if link.Art != domain.LinkProdukt {
			t.Errorf("%s gab %q statt eines Produkts", anbieter, link.Art)
		}

		if link.Herkunft != domain.HerkunftKennung {
			t.Errorf("%s gab die Herkunft %q", anbieter, link.Herkunft)
		}
	}

	// **Cardmarket bekommt trotz Kennung die Suche.** Gemessen am
	// 2026-09-19: `idProduct` landet auf der Sammelseite und nicht auf der
	// Karte — ein Produktlink, der keiner ist, wäre schlimmer als keiner.
	if link := linkVon(links, domain.LinkCardmarket); link.Art != domain.LinkSuche {
		t.Errorf("Cardmarket gab %q: %q", link.Art, link.Adresse)
	}

	// **Und Card Kingdom und MTGGoldfish ebenso** — sie führen keine
	// Kennung, die dieses Programm hätte.
	for _, anbieter := range []domain.Linkanbieter{
		domain.LinkCardKingdom, domain.LinkMTGGoldfish,
	} {
		if link := linkVon(links, anbieter); link.Art != domain.LinkSuche {
			t.Errorf("%s gab %q", anbieter, link.Art)
		}
	}
}

// **Eine Kennung mit der falschen Semantik baut keinen Produktlink.**
func TestEineKennungMitFalscherSemantikBautKeinProdukt(t *testing.T) {
	t.Parallel()

	for _, semantik := range []domain.Semantik{
		domain.SemantikKonzept, domain.SemantikKunstwerk,
	} {
		links := domain.Marktlinks([]domain.Druckkennung{
			{System: domain.SystemTCGplayer, Semantik: semantik, Wert: "33632"},
		}, "Lightning Bolt")

		if link := linkVon(links, domain.LinkTCGplayer); link.Art == domain.LinkProdukt {
			t.Errorf("%s baute ein Produkt: %q", semantik, link.Adresse)
		}
	}

	// **Und eine leere ebenso wenig.**
	links := domain.Marktlinks([]domain.Druckkennung{
		{System: domain.SystemTCGplayer, Semantik: domain.SemantikProdukt},
	}, "Lightning Bolt")

	if link := linkVon(links, domain.LinkTCGplayer); link.Art == domain.LinkProdukt {
		t.Errorf("die leere Kennung baute ein Produkt: %q", link.Adresse)
	}
}

// **Was nicht hinausgehen darf, geht nicht hinaus.**
func TestKeineUnerlaubteAdresseKommtDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]domain.Marktlink{
		"ein fremder Anbieter": {
			Anbieter: "kartenkiosk", Adresse: "https://www.cardmarket.com/",
		},
		"kein https": {
			Anbieter: domain.LinkCardmarket,
			Adresse:  "http://www.cardmarket.com/",
		},
		"ein anderes Protokoll": {
			Anbieter: domain.LinkCardmarket, Adresse: "file:///etc/passwd",
		},
		// **Der Lookalike** — er liest sich wie der echte Wirt.
		"ein ähnlicher Wirt": {
			Anbieter: domain.LinkCardmarket,
			Adresse:  "https://www.cardmarket.com.beispiel.tld/",
		},
		"eine Unterdomäne": {
			Anbieter: domain.LinkCardmarket,
			Adresse:  "https://boese.www.cardmarket.com/",
		},
		"ein Bindestrichnachbar": {
			Anbieter: domain.LinkCardmarket,
			Adresse:  "https://www-cardmarket.com/",
		},
		// **Die Benutzerangabe** — die Adresszeile zeigt den Teil davor.
		//
		// **Zwei Gestalten, und sie fallen an zwei Stellen.** Beim ersten
		// ist `www.cardmarket.com` der **Benutzer** und `boese.example` der
		// Wirt; der Wirt fängt ihn. Beim zweiten stimmt der Wirt, und
		// gefangen wird die Angabe selbst.
		"ein Wirt hinter dem Klammeraffen": {
			Anbieter: domain.LinkCardmarket,
			Adresse:  "https://www.cardmarket.com@boese.example/",
		},
		"eine Benutzerangabe am richtigen Wirt": {
			Anbieter: domain.LinkCardmarket,
			Adresse:  "https://opfer@www.cardmarket.com/",
		},
		"ein eigener Port": {
			Anbieter: domain.LinkCardmarket,
			Adresse:  "https://www.cardmarket.com:8443/",
		},
		"ein Geheimnis in der Abfrage": {
			Anbieter: domain.LinkCardmarket,
			Adresse:  "https://www.cardmarket.com/?api_key=geheim",
		},
		"eine Marke in der Abfrage": {
			Anbieter: domain.LinkCardTrader,
			Adresse:  "https://www.cardtrader.com/?Token=abc",
		},
		"eine Adresse, die keine ist": {
			Anbieter: domain.LinkCardmarket, Adresse: "https://%zz",
		},
	}

	for name, link := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := domain.PruefeMarktlink(link); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Errorf("die Adresse kam mit %v durch", fehler)
			}
		})
	}
}

// **Und eine kodierte Eingabe bleibt in der Abfrage.**
//
// **Der Fund, für den dieser Lauf steht**: ein Name mit `&` oder `?` könnte
// eine zweite Angabe anhängen, wenn er ungeprüft in die Adresse ginge. Er
// geht durch `url.Values.Encode`, und der Wirt bleibt derselbe.
func TestEinNameMitSonderzeichenBrichtDieAdresseNicht(t *testing.T) {
	t.Parallel()

	boese := "Bolt&api_key=geheim#/../../etc/passwd?x=1"

	for _, link := range domain.Marktlinks(nil, boese) {
		if fehler := domain.PruefeMarktlink(link); fehler != nil {
			t.Errorf("%s: %v", link.Anbieter, fehler)
		}

		gelesen, fehler := url.Parse(link.Adresse)
		if fehler != nil {
			t.Fatalf("%s: %v", link.Anbieter, fehler)
		}

		if gelesen.Host != domain.LinkhostFuer(link.Anbieter) {
			t.Errorf("%s zeigt auf %q", link.Anbieter, gelesen.Host)
		}

		if gelesen.Query().Get("api_key") != "" {
			t.Errorf("%s trägt einen Schlüssel: %q", link.Anbieter, link.Adresse)
		}
	}
}

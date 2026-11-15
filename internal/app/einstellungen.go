package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Einstellungsfuehrer führt die Anwendungseinstellungen — die Naht zu `service`.
type Einstellungsfuehrer interface {
	Alle(ctx context.Context) (map[string]string, error)
	Setzen(ctx context.Context, schluessel, wert string) error

	// Stand gibt je Schlüssel die Wahl **und** das, was gilt — Werkstück 9
	// (#1351), Punkt 9.
	//
	// **Beides, weil die Seite beides zeigt**: die Auswahlliste steht auf
	// der Wahl (dort vielleicht „wie der Nutzer"), und die Anwendung
	// richtet sich nach dem Wirksamen.
	Stand(ctx context.Context) (map[string]service.Wahl, error)
}

// EinstellungDTO ist eine Einstellung samt ihrer möglichen Werte.
//
// **Die Werte kommen mit** (#245). Ohne sie baute das Fenster die Auswahlliste
// selbst — und dann stünden die erlaubten Werte an zwei Stellen, von denen die
// im Fenster die ungeprüfte wäre. Ein `grau` käme dort durch und im Bestand
// nicht an; sichtbar wäre nur, dass nichts geschieht.
type EinstellungDTO struct {
	Schluessel string   `json:"schluessel"`
	Wert       string   `json:"wert"`
	Werte      []string `json:"werte"`

	// Wirksam ist, was **gilt** — Werkstück 9 (#1351), Punkt 9.
	//
	// **Er weicht von `Wert` genau dann ab, wenn dort `erben` steht**: dann
	// gilt, was beim Menschen gewählt ist. Die Auswahlliste zeigt `Wert`,
	// die Anwendung richtet sich nach `Wirksam` — wer nur eines von beiden
	// hätte, zeigte entweder eine Wahl, die niemand getroffen hat, oder
	// eine Anwendung, die anders aussieht, als die Liste sagt.
	Wirksam string `json:"wirksam"`

	// Bereich sagt, **wo** die Einstellung steht — `anwendung` oder
	// `bildausfuhr` (#491).
	//
	// **Nicht jede Einstellung gehört auf die Einstellungsseite.** Die Wahl
	// beim Namenskonflikt gehört zu dem Ausgabelauf, den man gerade startet;
	// MSE hält sie ebenso im Ausgabefenster.
	//
	// **Die Zuordnung kommt aus der Domäne**, nicht aus dem Fenster: sonst
	// entschiede die Oberfläche, was eine Anwendungseinstellung ist, und die
	// Antwort stünde an zwei Stellen.
	Bereich string `json:"bereich"`

	// Ebene sagt, **wem** die Einstellung gehört — `geraet`, `nutzer`,
	// `profil` oder `arbeitsplatz`. Werkstück O3 von Phase 5.2
	// (#1427).
	//
	// **Nicht dasselbe wie `Bereich`**: jener sagt, **wo im Fenster** die
	// Einstellung steht, diese, **wie weit sie reist**. Zwei Fragen, die man
	// leicht verwechselt — deshalb stehen sie nebeneinander.
	Ebene string `json:"ebene"`

	// Herkunft sagt, wessen Entscheidung `Wirksam` war — `profil`, `nutzer`
	// oder `vorgabe`.
	//
	// **Ohne sie wäre `Wirksam` eine Zahl ohne Einheit**: „dunkel" sagt
	// nicht, ob es hier gewählt wurde oder überall gilt, und damit nicht,
	// was ein Zurücksetzen bewirkte.
	Herkunft string `json:"herkunft"`

	// Wirkung sagt, **wann** ein geänderter Wert gilt — [domain.WirkungSofort]
	// oder [domain.WirkungNeustart] (UX-04.AK-02,
	// #1676).
	//
	// **Auf allen vier Flächen dieselbe Antwort** (§16): ob der Proxy sofort
	// wirkt, hängt daran, wie `internal/netz` ihn liest — nicht daran, ob
	// jemand ihn im Fenster, an der Kommandozeile, über REST oder über MCP
	// umlegt.
	Wirkung string `json:"wirkung"`

	// Vorgabe ist der Wert, mit dem die Anwendung anfängt — Werkstück U11 von
	// Phase 5.1 (#1422). **Damit ein Zurücksetzen weiß, wohin**: das Fenster
	// vergleicht `Wert` damit und setzt darauf zurück, statt die Vorgabe ein
	// zweites Mal zu führen.
	Vorgabe string `json:"vorgabe"`
}

// EinstellungenAntwort trägt die Einstellungen.
type EinstellungenAntwort struct {
	Einstellungen []EinstellungDTO `json:"einstellungen"`
	Fehler        *Fehler          `json:"fehler"`
}

// Einstellungen gibt die Einstellungen samt ihrer möglichen Werte.
func (a *App) Einstellungen(ctx context.Context) EinstellungenAntwort {
	ctx = a.imProfil(ctx)

	antwort := EinstellungenAntwort{Einstellungen: []EinstellungDTO{}}

	stand, err := a.einstellungen.Stand(ctx)
	if err != nil {
		antwort.Fehler = Uebersetze(ctx, err)

		return antwort
	}

	// **Die Reihe kommt aus der Domäne**, nicht von hier: sonst entschiede die
	// Fassade, welche Einstellungen es gibt, und die Antwort stünde an zwei
	// Stellen (Zweitblick über Phase G).
	for _, schluessel := range domain.Einstellungsreihe() {
		antwort.Einstellungen = append(antwort.Einstellungen, EinstellungDTO{
			Schluessel: schluessel,
			// **Ohne Proxykennwort** — S13.NAT-02 (#1706).
			Wert:     domain.Einstellungsanzeige(schluessel, stand[schluessel].Gewaehlt),
			Wirksam:  domain.Einstellungsanzeige(schluessel, stand[schluessel].Wirksam),
			Werte:    domain.ErlaubteEinstellungswerte(schluessel),
			Bereich:  domain.EinstellungsBereich(schluessel),
			Ebene:    ebeneVon(schluessel),
			Herkunft: stand[schluessel].Herkunft,
			Wirkung:  domain.EinstellungsWirkung(schluessel),
			Vorgabe:  domain.EinstellungsVorgabe(schluessel),
		})
	}

	return antwort
}

// ebeneVon gibt die Ebene eines Schlüssels als Text.
//
// **Ein leerer Text ist unmöglich** — der Wächter in `internal/domain` hält,
// dass jeder Schlüssel beider Reihen in genau einer Ebene steht. Er steht
// hier trotzdem, weil eine Naht nichts behauptet, was sie nicht selbst weiß.
func ebeneVon(schluessel string) string {
	ebene, _ := domain.EbeneVon(schluessel)

	return string(ebene)
}

// EinstellungSetzen schreibt eine Einstellung fort.
func (a *App) EinstellungSetzen(ctx context.Context, schluessel, wert string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.einstellungen.Setzen(ctx, schluessel, wert); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	a.schalterWirkt(ctx, schluessel, wert)

	return SchreibAntwort{}
}

// MitEinstellungen setzt die Naht zum Einstellungsverwalter.
func (a *App) MitEinstellungen(fuehrer Einstellungsfuehrer) { a.einstellungen = fuehrer }

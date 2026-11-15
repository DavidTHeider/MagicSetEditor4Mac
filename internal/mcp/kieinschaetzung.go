package mcp

// **Der belegte Regelhinweis über MCP** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// # Warum das Werkzeug »schreiben« heißt, obwohl es nichts schreibt
//
// **Weil es Geld kosten kann.** An den Karten ändert es nichts — das ist die
// Zusage des Issues —, aber es bucht einen Kostenposten und fragt ein Modell.
// Ein Werkzeug der Art `Lesend` verspricht beides nicht zu tun; dieselbe
// Einordnung wie bei `schreiben.ki.batch`.

import (
	"context"
	"encoding/json"
)

// KIEinschaetzungsbezug ist, was für eine Einschätzung gebraucht wird.
type KIEinschaetzungsbezug struct {
	KarteID int64  `json:"karteId"`
	Art     string `json:"art"`

	Anbieter string `json:"anbieter"`
	Alias    string `json:"alias"`
	Frage    string `json:"frage"`

	// Format ist, worauf sich ein Vergleich bezieht (A13.2b, #1452) — leer
	// heißt: keines, und dann sagt die Antwort es.
	Format string `json:"format"`

	// Freigegeben nennt die **privaten** Korpora, die mitgesucht werden
	// dürfen — A13.4 (#1452). Leer heißt: keine.
	Freigegeben []string `json:"freigegeben"`
}

// KIBelegbefundsicht ist ein Mangel am Quellenanspruch einer Aussage.
type KIBelegbefundsicht struct {
	Art     string `json:"art"`
	Stelle  string `json:"stelle"`
	Meldung string `json:"meldung"`
}

// KIBelegsichtQuelle ist ein Ausschnitt, auf den sich eine Aussage stützt.
type KIBelegsichtQuelle struct {
	Kennung string `json:"kennung"`
	Quelle  string `json:"quelle"`
	Titel   string `json:"titel"`

	// Ausschnitt ist **der Text, der wirklich mitgegeben wurde**.
	Ausschnitt string `json:"ausschnitt"`

	Regelnummer string `json:"regelnummer"`
	Urheber     string `json:"urheber"`

	Quellrevision string `json:"quellrevision"`
	Quelldatum    string `json:"quelldatum"`
	Abgerufen     string `json:"abgerufen"`
	Veraltet      bool   `json:"veraltet"`
}

// KIKorpusstandsicht sagt, wie alt eine Quelle ist.
type KIKorpusstandsicht struct {
	Quelle   string `json:"quelle"`
	Revision string `json:"revision"`
	Datum    string `json:"datum"`
}

// KIVergleichsicht ist eine Karte, gegen die verglichen wurde.
type KIVergleichsicht struct {
	Beleg string `json:"beleg"`
	Name  string `json:"name"`

	// Unterschiede sind die konkreten — leer gibt es nicht: ein Vergleich
	// ohne Unterschied sagt „ähnlich", und das sagt der Rang schon.
	Unterschiede []string `json:"unterschiede"`
}

// KIBalancesicht trägt, was nur eine Balanceaussage hat.
//
// **Ohne Zahlenfeld**: „kein objektiver Balancewert aus Ähnlichkeitsscore"
// steht hier nicht als Prüfung, sondern als fehlendes Feld.
type KIBalancesicht struct {
	// Strategie sagt, **wie** die Vergleichskarten gefunden wurden — vom
	// Programm und nicht vom Modell.
	Strategie string `json:"strategie"`

	Format  string               `json:"format"`
	Staende []KIKorpusstandsicht `json:"staende"`

	Vergleiche      []KIVergleichsicht `json:"vergleiche"`
	NichtBetrachtet []string           `json:"nichtBetrachtet"`
}

// KIEinschaetzungssicht ist, was dabei herauskommt.
type KIEinschaetzungssicht struct {
	Art       string `json:"art"`
	Satz      string `json:"satz"`
	Belegt    bool   `json:"belegt"`
	Vorbehalt string `json:"vorbehalt"`

	Befunde          []KIBelegbefundsicht `json:"befunde"`
	Einschraenkungen []string             `json:"einschraenkungen"`
	Belege           []KIBelegsichtQuelle `json:"belege"`
	Korpusstaende    []KIKorpusstandsicht `json:"korpusstaende"`

	Anbieter string `json:"anbieter"`
	Modell   string `json:"modell"`

	// Balance **fehlt** bei `regel` — der Schlüssel steht gar nicht da.
	//
	// **Und das ist die dritte Möglichkeit neben Liste und `null`.** Eine
	// Regeleinschätzung hat keine Vergleichskarten; ein leerer Verbund sähe
	// aus, als hätte die Suche keine gefunden, und `null` ist genau die
	// Verwechslung, gegen die der Lauf daneben steht — „keine Belege" heißt
	// etwas anderes als „die Frage ging schief". Was hier gilt, ist ein
	// drittes: **diesen Abschnitt gibt es für diese Art nicht.**
	Balance *KIBalancesicht `json:"balance,omitempty"`
}

// kieinschaetzungswerkzeug gibt das Werkzeug für eine Einschätzung.
func kieinschaetzungswerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.ki.einschaetzung", Art: Schreibend,
		Erklaerung: "Lässt ein Modell eine Karte BEURTEILEN — »regel«, wie " +
			"sie nach den Regeln wirkt, oder »balance«, wie sie im " +
			"Vergleich dasteht. An der KARTE, der SAMMLUNG und am " +
			"SPIELSTAND ändert es NICHTS. ÄNDERT den Bestand trotzdem: es " +
			"bucht einen Kostenposten und schließt ihn wieder. " +
			"Die Belege kommen aus dem lokalen Suchindex und werden vom " +
			"Dienst vergeben; »belegt« sagt, ob die Antwort nur vergebene " +
			"Kennungen nennt. »befunde« zählt auf, was sie sonst nennt — " +
			"eine erfundene Regelnummer oder Adresse steht DORT und nicht " +
			"als Quelle. " +
			"KEINE AMTLICHE REGELAUSKUNFT: das ist die Einschätzung eines " +
			"Modells, und die deterministische Formatprüfung bleibt ein " +
			"eigener Weg. " +
			"Bei »balance« steht unter »balance« zusätzlich, WOMIT " +
			"verglichen wurde: je Vergleichskarte ihr Beleg und mindestens " +
			"ein KONKRETER Unterschied. Ein Vergleich ohne Unterschied ist " +
			"ein BEFUND — »ähnlich« sagt schon der Rang, der die Karte " +
			"gefunden hat, und der bedeutet nichts. »strategie« sagt, wie " +
			"gesucht wurde; sie kommt vom Programm und nicht vom Modell. " +
			"EINE NOTE GIBT ES NICHT: es gibt kein Feld dafür. " + vorbehalt,
		Schema: objekt(map[string]any{
			"karteId": zahl("Die Kennung der Karte, um die es geht."),
			"art":     text("»regel« oder »balance«."),
			"format": text("Wahlfrei und nur bei »balance«: worauf sich der " +
				"Vergleich bezieht. Ohne Angabe gilt er für kein Format " +
				"besonders, und die Antwort sagt es."),
			"alias": text("Das Modell, wie »lesen.ki.modelle« es nennt — es " +
				"muss die strukturierte Ausgabe BELEGT haben."),
			"frage": text("Was beurteilt werden soll. Sie geht AUCH an den " +
				"Suchindex — je genauer sie ist, desto einschlägiger sind " +
				"die Belege."),
		}, "karteId", "art", "alias", "frage"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var bezug KIEinschaetzungsbezug

			if err := lies(roh, &bezug); err != nil {
				return "", err
			}

			sicht, fehler := werkstatt.KIEinschaetzung(ctx, bezug)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(mitLeerenEinschaetzungslisten(sicht))
		},
	}
}

// mitLeerenEinschaetzungslisten macht aus `null` eine leere Liste.
//
// **Für ein Sprachmodell sieht `null` aus wie „konnte nicht beantwortet
// werden"** — und bei den Belegen wäre das die gefährlichste Verwechslung:
// „keine Belege" heißt etwas anderes als „die Frage ging schief".
func mitLeerenEinschaetzungslisten(sicht KIEinschaetzungssicht) KIEinschaetzungssicht {
	if sicht.Befunde == nil {
		sicht.Befunde = []KIBelegbefundsicht{}
	}

	if sicht.Einschraenkungen == nil {
		sicht.Einschraenkungen = []string{}
	}

	if sicht.Belege == nil {
		sicht.Belege = []KIBelegsichtQuelle{}
	}

	if sicht.Korpusstaende == nil {
		sicht.Korpusstaende = []KIKorpusstandsicht{}
	}

	if sicht.Balance != nil {
		sicht.Balance = mitLeerenBalancelisten(sicht.Balance)
	}

	return sicht
}

// mitLeerenBalancelisten macht auch dort aus `null` eine leere Liste.
func mitLeerenBalancelisten(balance *KIBalancesicht) *KIBalancesicht {
	if balance.Vergleiche == nil {
		balance.Vergleiche = []KIVergleichsicht{}
	}

	if balance.NichtBetrachtet == nil {
		balance.NichtBetrachtet = []string{}
	}

	if balance.Staende == nil {
		balance.Staende = []KIKorpusstandsicht{}
	}

	return balance
}

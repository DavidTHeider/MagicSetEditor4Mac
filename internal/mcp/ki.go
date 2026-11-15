package mcp

// **Die KI als Werkzeug** — Werkstück A6 von Phase 5.3
// (#1445).

import (
	"context"
	"encoding/json"
)

// KIKatalogsicht ist der Modellkatalog samt der Regel, die für ihn gilt.
type KIKatalogsicht struct {
	// NurLokal sagt, ob etwas diesen Rechner verlassen darf.
	NurLokal bool `json:"nurLokal"`

	// Einrichtung ist der Stand, der am weitesten kommt — einer der fünf
	// aus `app/kilaufzeit.go` (A7, #1446).
	//
	// **Der beste über alle Anbieter** (A8, #1447): welcher woran hängt,
	// steht je Zeile unter `anbieter`.
	Einrichtung string `json:"einrichtung"`

	// Anbieter sind die Quellen, jede mit ihrem eigenen Stand.
	Anbieter []KIAnbietersicht `json:"anbieter"`

	// Cloud sagt, was über die Auslagerung der **lokalen** Laufzeit
	// feststeht — und was nicht.
	Cloud KICloudsicht `json:"cloud"`

	// Befunde sind die Sätze, die über alle Anbieter hinweg gelten.
	Befunde []string `json:"befunde"`

	Modelle []KIModellsicht `json:"modelle"`

	// Vorrang nennt, wen ein Vorschlag mit Vorrang fragt (#2054).
	Vorrang []string `json:"vorrang"`
}

// KIAnbietersicht ist ein Anbieter samt dem, was über ihn feststeht.
type KIAnbietersicht struct {
	Name        string `json:"name"`
	Einrichtung string `json:"einrichtung"`
	Laufzeit    string `json:"laufzeit"`
	Adresse     string `json:"adresse"`

	// Ort ist `lokal` oder `cloud` — **wo dieser Anbieter rechnet** (#1539).
	// Eine Messung an einem Cloudanbieter kostet Geld, und wer das erst an
	// der Absage merkt, hat schon gerufen.
	Ort string `json:"ort"`

	Befunde []string `json:"befunde"`

	// Messbar sind die Modelle, die er **nennt**, ohne etwas über sie zu
	// sagen (A9, #1448).
	Messbar []string `json:"messbar"`

	// Gemessen sind die Modelle mit einer Messung im Bestand.
	Gemessen []string `json:"gemessen"`

	// SitzungAus sagt, dass er für diese Sitzung abgeschaltet ist — nur
	// Apples Gerätemodell kennt das (AI-05, #2042).
	SitzungAus bool `json:"sitzungAus"`
}

// KICloudsicht ist der Cloudstand samt Fundstelle.
//
// **Drei Stände und nicht zwei** (A7, #1446): „keine Datei" heißt ungewiss,
// denn die zweite dokumentierte Abschaltung steht in der Umgebung eines
// fremden Prozesses.
type KICloudsicht struct {
	Stand string `json:"stand"`
	Beleg string `json:"beleg"`
}

// KIProbensicht ist, was eine Messung ergeben hat (A9, #1448).
type KIProbensicht struct {
	Proben []KIProbesicht `json:"proben"`
}

// KIProbesicht ist eine gemessene Fähigkeit.
type KIProbesicht struct {
	Anbieter   string `json:"anbieter"`
	Alias      string `json:"alias"`
	Faehigkeit string `json:"faehigkeit"`
	Datum      string `json:"datum"`
	Beleg      string `json:"beleg"`

	// Modell ist, was wirklich geantwortet hat — nicht der Alias.
	Modell string `json:"modell"`
}

// KIProbenbezug nennt, was gemessen werden soll.
type KIProbenbezug struct {
	Anbieter string `json:"anbieter"`
	Alias    string `json:"alias"`
}

// KIVorschlagssicht ist ein Kartenvorschlag.
type KIVorschlagssicht struct {
	Kennung string `json:"kennung"`
	Modell  string `json:"modell"`

	Felder []KIFeldwertsicht `json:"felder"`

	// Befunde sind die Felder, die das Schema **nicht** durchgelassen hat.
	Befunde []string `json:"befunde"`

	Ablauf string `json:"ablauf"`

	// Uebernommen ist die Zahl der geschriebenen Felder — null, wenn nur
	// vorgeschlagen wurde.
	Uebernommen int `json:"uebernommen"`

	// Anbieter und Weg sagen, wer geantwortet hat und wie er drankam —
	// `gewaehlt`, `vorrang` oder `rueckfall` (AI-05, #2042).
	Anbieter string `json:"anbieter"`
	Weg      string `json:"weg"`

	// Ausgang ist leer, oder `abgelehnt`: das Modell hat verweigert, und
	// Rueckfaelle nennt, wer es versuchen darf.
	Ausgang     string   `json:"ausgang,omitempty"`
	Rueckfaelle []string `json:"rueckfaelle,omitempty"`
}

// KIFeldwertsicht ist ein vorgeschlagener Feldwert.
type KIFeldwertsicht struct {
	Feld  string   `json:"feld"`
	Werte []string `json:"werte"`

	// Bisher ist, was **heute** im Feld steht — leer heißt: nichts.
	//
	// **Damit sichtbar ist, was sich ändert** (A11, #1450). Ein Vorschlag,
	// der nur den neuen Wert zeigt, verlangt von jemandem, dass er den alten
	// im Kopf hat — auch an einer Fläche ohne Fenster.
	Bisher string `json:"bisher"`
}

// KIVorschlagsbezug ist, was für einen Vorschlag gebraucht wird.
type KIVorschlagsbezug struct {
	KarteID int64 `json:"karteId"`

	// Anbieter ist die Quelle — `ollama` oder `anthropic` (A8, #1447).
	Anbieter string `json:"anbieter"`

	Alias   string `json:"alias"`
	Auftrag string `json:"auftrag"`

	// Uebernehmen sagt, ob der Vorschlag geschrieben werden soll.
	//
	// **Ohne ihn wird nichts geschrieben** — er ist die ausdrückliche
	// Entscheidung, die am Fenster der zweite Klick ist.
	Uebernehmen bool `json:"uebernehmen"`

	// Felder nennt, **was** übernommen wird — leer heißt: alles, was das
	// Schema durchgelassen hat (A11, #1450).
	//
	// # Warum hier leer »alles« heißt und am Fenster nicht
	//
	// **Weil hier niemand den Vorschlag gesehen hat.** Am Fenster steht er
	// vor jemandem, der je Feld entscheidet; hier ist die Entscheidung
	// `uebernehmen` selbst — und sie ist ausdrücklich. Wer enger wählen
	// will, nennt die Felder.
	Felder []string `json:"felder,omitempty"`

	// Belegfrage geht an den Suchindex — Werkstück A13.4 von Phase 5.3
	// (#1452).
	//
	// **Leer heißt: gar nicht fragen.** Wo sie steht, bekommt das Modell
	// dieselben Belege wie eine Einschätzung — aus demselben Index, mit
	// demselben Manifest —, und die Antwort führt sie unter `belege`.
	Belegfrage string `json:"belegfrage,omitempty"`

	// Freigegeben nennt die **privaten** Korpora, die dabei mitgesucht
	// werden dürfen. Leer heißt: keine.
	Freigegeben []string `json:"freigegeben,omitempty"`

	// Nur nennt, **wonach** gefragt wird — leer heißt: alle Felder. Name
	// und Flavortext machen eine kleine Aufgabe (AI-05, #2042).
	Nur []string `json:"nur,omitempty"`
}

// KIModellsicht ist ein Modell.
type KIModellsicht struct {
	Anbieter string `json:"anbieter"`
	Alias    string `json:"alias"`
	Laufzeit string `json:"laufzeit"`
	Ort      string `json:"ort"`

	Faehigkeiten []KIBelegsicht `json:"faehigkeiten"`
	Herkunft     []KIGliedsicht `json:"herkunft"`
	Bindung      string         `json:"bindung"`

	// Kontextfenster und KontextBeleg wie am Fenster (AI-01.AK-02, #1742):
	// 0 heißt nicht angegeben.
	Kontextfenster int    `json:"kontextfenster"`
	KontextBeleg   string `json:"kontextbeleg"`
}

// KIBelegsicht ist eine belegte Fähigkeit.
type KIBelegsicht struct {
	Faehigkeit string `json:"faehigkeit"`
	Datum      string `json:"datum"`
	Beleg      string `json:"beleg"`
}

// KIGliedsicht ist ein Glied der Herkunftskette.
type KIGliedsicht struct {
	Art      string `json:"art"`
	Kennung  string `json:"kennung"`
	Aus      string `json:"aus"`
	Quelle   string `json:"quelle"`
	Revision string `json:"revision"`
	Lizenz   string `json:"lizenz"`
}

// kimodellewerkzeug gibt das Werkzeug für den Modellkatalog.
//
// # Er fragt im Netz
//
// **Seit A7 fragt dieser Ruf jeden eingeschalteten Anbieter** — die lokale
// Laufzeit und jeden Cloudanbieter, dessen `dienst.`-Schalter auf `an` steht
// ([app.App.frageAnbieter]). Er muss es: „kein Katalogeintrag ohne
// Anbieterantwort" ([0097](../../docs/entscheidungen/0097-was-ueber-ein-ki-modell-feststeht.md)),
// und eine belegte Fähigkeit entsteht nirgends sonst.
//
// **Der Lauf daneben** (`internal/flaechen/kimodelle_test.go`) misst den
// **Vorgabestand** — dort ist nichts eingeschaltet, und dann fragt der Ruf
// tatsächlich niemanden. Die Reichweite einer Messung ist die Reichweite ihrer
// Aussage (§10, Satz 6).
func kimodellewerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "lesen.ki.modelle", Art: Lesend,
		Erklaerung: "Nennt, was diese Anwendung über KI-Modelle WEISS: je " +
			"Modell die BELEGTEN Fähigkeiten mit Datum und Beleg, die " +
			"Herkunft bis zum ausgeführten Artefakt und den Ort, an dem es " +
			"rechnet. ÄNDERT NICHTS. Es FRAGT aber jeden EINGESCHALTETEN " +
			"Anbieter — ohne dessen Antwort gibt es keinen Beleg; ist keiner " +
			"eingeschaltet, geht auch keine Anfrage hinaus. " +
			"Eine leere Liste heißt: es steht KEIN Modell fest — eine " +
			"Modellliste eines Anbieters ist keine belegte Fähigkeit. " +
			"»nurLokal« sagt, ob überhaupt etwas diesen Rechner verlassen " +
			"darf, und gilt VOR jeder Anfrage. " + vorbehalt,
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			sicht, fehler := werkstatt.KIModelle(ctx)
			if fehler != nil {
				return "", fehler
			}

			// **Leere Listen bleiben Listen** — ein `null` sähe für ein
			// Sprachmodell aus wie „konnte nicht beantwortet werden".
			if sicht.Modelle == nil {
				sicht.Modelle = []KIModellsicht{}
			}

			return alsJSON(sicht)
		},
	}
}

// kivorschlagwerkzeug gibt das Werkzeug für einen Kartenvorschlag.
func kivorschlagwerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.ki.vorschlag", Art: Schreibend,
		Erklaerung: "Lässt ein LOKALES Modell Felder einer Karte " +
			"vorschlagen. FRAGT die eingerichtete KI-Laufzeit; ohne " +
			"eingeschaltete Laufzeit scheitert der Ruf. " +
			"ÄNDERT NICHTS, solange »uebernehmen« nicht gesetzt ist — dann " +
			"kommen die Felder nur zurück. Die Form kommt aus den " +
			"TATSÄCHLICHEN Feldern des Pakets; was das Schema nicht " +
			"durchlässt, steht unter »befunde« und wird NICHT geschrieben. " +
			"Eine abgeschnittene Antwort wird abgewiesen. " +
			"Das Schema prüft Form und Felder, NIEMALS Wahrheitsgehalt " +
			"oder Spielbalance. " + vorbehalt,
		Schema: objekt(map[string]any{
			"karteId": zahl("Die Kennung der Zielkarte."),
			"anbieter": text("Wahlfrei: der Anbieter, wie »lesen.ki.modelle« ihn nennt. " +
				"»vorrang« lässt die Regel wählen: Name und Flavortext (»nur«) " +
				"zuerst Apples Gerätemodell, dann lokale Anbieter, NIE die Cloud; " +
				"»alias« bleibt dann leer."),
			"alias": text("Das Modell, wie »lesen.ki.modelle« es nennt — es " +
				"muss die strukturierte Ausgabe BELEGT haben."),
			"nur": textliste("Wahlfrei: nach WELCHEN Feldern gefragt wird. Leer heißt: alle.",
				"Ein Feldname des Pakets, etwa »name« oder »flavor text«."),
			"auftrag": text("Was das Modell tun soll."),
			"uebernehmen": jaNein("Wahlfrei: die Felder wirklich " +
				"schreiben. Ohne ihn wird nichts geändert."),
			"felder": textliste(
				"Wahlfrei: WELCHE Felder übernommen werden. Leer heißt: "+
					"alle, die das Schema durchgelassen hat. Ein Name, den "+
					"der Vorschlag nicht trägt, wird abgewiesen — "+
					"geschrieben wird dann NICHTS.",
				"Ein Feldname aus dem Vorschlag."),
		}, "karteId", "alias", "auftrag"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var bezug KIVorschlagsbezug

			if err := lies(roh, &bezug); err != nil {
				return "", err
			}

			sicht, fehler := werkstatt.KIVorschlag(ctx, bezug)
			if fehler != nil {
				return "", fehler
			}

			// **Leere Listen bleiben Listen** — ein `null` sähe für ein
			// Sprachmodell aus wie „konnte nicht beantwortet werden".
			if sicht.Felder == nil {
				sicht.Felder = []KIFeldwertsicht{}
			}

			if sicht.Befunde == nil {
				sicht.Befunde = []string{}
			}

			return alsJSON(sicht)
		},
	}
}

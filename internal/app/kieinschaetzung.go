package app

// **Der belegte Regelhinweis** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// # Der erste Aufrufer von [domain.PruefeBelege]
//
// **A16 hat die Prüfung gebaut und niemand hat sie gerufen.** Sie stand mit
// Tests und voller Abdeckung da, und es gab keinen Weg dorthin — die erste
// Dimension aus §10. Dieses Werkstück ist der Weg.
//
// # Keine zweite Pipeline
//
// **Das Issue verbietet sie wörtlich**: „A11 erhält dieselben Manifest-/
// Belegdaten wie A16; keine zweite Prompt-/Retrievalpipeline." Deshalb steht
// hier **kein** Retrieval: das Kontextpaket kommt aus [service.Kontextdienst],
// die Kostenprüfung aus [service.KIAuftragswerk], die Belegprüfung aus
// `domain`. Was diese Datei tut, ist die drei zu verbinden.
//
// # Und was daneben steht
//
// **Die Anfrageseite** — Bezug, Prüfung der Art, Belege, Recherchefall —
// liegt seit A13.2b in `kieinschaetzungsbezug.go`; die Balanceanalyse mit
// ihrem eigenen Schema in `kibalance.go`. Der Schnitt folgt der Frage (§2):
// hier steht, **was zurückkommt** und wie es gelesen wird.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// KIEinschaetzungsanweisung steht vor jeder Frage nach einer Einschätzung.
//
// # Warum sie nicht die Kartenanweisung ist
//
// **Weil sie von etwas anderem spricht.** „Du füllst Felder einer
// Sammelkarte" wäre die erste Zeile, die das Modell vor einer Regelfrage
// liest — und die erste Zeile ist die, an der es sich ausrichtet.
//
// # Und warum sie die Fußnoten verlangt
//
// **Weil die Prüfung sonst nichts zu prüfen hätte.** `domain.PruefeBelege`
// hält genannte Kennungen gegen die vergebenen; ein Modell, das keine nennt,
// bekommt keinen Befund für eine falsche Nummer — es bekommt einen dafür,
// dass es gar nichts belegt.
const KIEinschaetzungsanweisung = "Du beurteilst eine Sammelkarte anhand der " +
	"mitgegebenen Belege. Stütze jede Aussage auf einen davon und nenne ihn " +
	"in eckigen Klammern, etwa [B1]. Trage nichts nach, was nicht in den " +
	"Belegen steht — sag stattdessen, dass es dort nicht steht. Du triffst " +
	"keine Regelentscheidung; du sagst, was die Belege hergeben."

// Das Budget, mit dem eine Einschätzung ihr Kontextpaket schnürt.
//
// # Warum hier Zahlen stehen und nicht eine Messung
//
// **Weil dieses Programm nirgends weiß, wie groß das Fenster eines Modells
// ist.** Der Katalog führt belegte **Fähigkeiten**, keine Fenstergrößen — ein
// Modell nennt sie nicht, und A6 verbietet, sie aus einem Namen abzuleiten
// ([0097](../../docs/entscheidungen/0097-was-ueber-ein-ki-modell-feststeht.md)).
//
// # Und warum die Untergrenze die sichere Richtung ist
//
// **Weil die beiden Fehler nicht gleich viel kosten.** Wer zu groß packt,
// bekommt eine **abgeschnittene** Antwort — und die wird ganz verworfen
// ([einschaetzungAus]). Wer zu klein packt, bekommt eine mit **weniger
// Belegen**; sie ist enger und bleibt brauchbar. Gewählt ist deshalb die
// zweite.
//
// **Das ist eine Annahme und keine Messung**, und sie steht hier, damit sie
// jemand ändern kann, wenn der Katalog eines Tages Fenstergrößen führt.
const (
	// KIEinschaetzungsfenster ist das angenommene Kontextfenster.
	KIEinschaetzungsfenster = 8192

	// KIEinschaetzungsreserve bleibt für die Antwort frei.
	KIEinschaetzungsreserve = 1024

	// KIEinschaetzungstreffer ist, wie weit die Recherche sucht.
	//
	// **Weiter als [KIEinschaetzungsbelege]**, und das ist Absicht: der
	// Kontextdienst wählt aus, entdoppelt und kürzt (A16). Wer genau so
	// viele suchte, wie mitgehen sollen, ließe ihm nichts zu wählen.
	KIEinschaetzungstreffer = 20

	// KIEinschaetzungsbelege begrenzt, wie viele Ausschnitte mitgehen.
	//
	// **Das Issue verlangt die Grenze**: „Kandidatenzahl/Kontextbudget
	// begrenzen." Acht Stellen sind genug, um eine Regelfrage zu stützen,
	// und wenige genug, dass ein Mensch sie nachliest — was er nicht
	// nachliest, ist keine Quelle, sondern eine Beteuerung.
	KIEinschaetzungsbelege = 8
)

// EinschaetzungsbefundDTO ist ein Mangel am Quellenanspruch einer Aussage.
type EinschaetzungsbefundDTO struct {
	Art     string `json:"art"`
	Stelle  string `json:"stelle"`
	Meldung string `json:"meldung"`
}

// KIEinschaetzungAntwort ist, was eine Einschätzung hergibt.
type KIEinschaetzungAntwort struct {
	Art string `json:"art"`

	// Satz ist die Einschätzung — mit den Fußnoten darin, wie das Modell sie
	// gesetzt hat.
	Satz string `json:"satz"`

	// Belegt sagt, ob sie als quellengestützt gelten darf.
	Belegt bool `json:"belegt"`

	// Vorbehalt steht an jeder — [domain.EinschaetzungsVorbehalt].
	Vorbehalt string `json:"vorbehalt"`

	Befunde []EinschaetzungsbefundDTO `json:"befunde"`

	// Einschraenkungen sagt, was das Modell **nicht** betrachtet hat.
	Einschraenkungen []string `json:"einschraenkungen"`

	// Belege sind die Ausschnitte, die mitgegeben wurden — alle, nicht nur
	// die genannten. Wer nur die genannten zeigte, verschwiege, was das
	// Modell hatte und nicht benutzt hat.
	Belege []BelegDTO `json:"belege"`

	// Korpusstaende sagen, **wie alt** die Quellen sind.
	Korpusstaende []KorpusstandDTO `json:"korpusstaende"`

	Anbieter string `json:"anbieter"`
	Modell   string `json:"modell"`

	// Balance trägt, was nur eine Balanceaussage hat — `nil` bei `regel`
	// (A13.2b, #1452).
	//
	// **Am Fenster steht dann `null`, an den drei anderen Flächen fehlt der
	// Schlüssel.** Das ist kein Versehen: hier liest TypeScript, und der Typ
	// sagt `KIBalanceDTO | null`; dort liest ein Sprachmodell, und für das
	// ist `null` von „leer" nicht zu unterscheiden — deshalb trägt
	// [mcp.KIEinschaetzungssicht] ein `omitempty`.
	Balance *BalanceDTO `json:"balance"`

	Fehler *Fehler `json:"fehler"`
}

// einschaetzungsantwort ist die Form, in der das Modell antwortet.
type einschaetzungsantwort struct {
	Einschaetzung    string   `json:"einschaetzung"`
	Einschraenkungen []string `json:"einschraenkungen"`
}

// einschaetzungsschema zwingt die Antwort in die Form.
func einschaetzungsschema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"einschaetzung": map[string]any{"type": "string"},
			"einschraenkungen": map[string]any{
				"type": "array", "items": map[string]any{"type": "string"},
			},
		},
		"required":             []string{"einschaetzung"},
		"additionalProperties": false,
	}
}

// leereEinschaetzung ist die Antwort, wenn nichts entstand.
//
// **Mit leeren Listen und nicht mit Nullwerten**: ein `null` statt `[]` kommt
// im Fenster als „keine Liste" an, und ein Sprachmodell liest es als „konnte
// nicht beantwortet werden". Ein Wächter zählt es nach
// (`leerelistenlaufzeit_test.go`) — und hat genau das hier gefunden.
func leereEinschaetzung() KIEinschaetzungAntwort {
	return KIEinschaetzungAntwort{
		Befunde: []EinschaetzungsbefundDTO{}, Einschraenkungen: []string{},
		Belege: []BelegDTO{}, Korpusstaende: []KorpusstandDTO{},
	}
}

// mitFehler setzt den Fehler an eine leere Antwort.
func mitFehler(ctx context.Context, fehler error) KIEinschaetzungAntwort {
	antwort := leereEinschaetzung()
	antwort.Fehler = Uebersetze(ctx, fehler)

	return antwort
}

// einschaetzungMit fährt den Lauf mit einem fertigen Werkzeugsatz.
func (a *App) einschaetzungMit(
	ctx context.Context, lauf einschaetzungslauf, satz kiwerkzeugsatz,
) KIEinschaetzungAntwort {
	angelegt, fehler := satz.werk.Anlegen(ctx, service.KIWunsch{
		ProfilID: lauf.profil, Faehigkeit: domain.KIStrukturierteGenerierung,
		Anbieter: lauf.bezug.Anbieter, Alias: lauf.bezug.Alias,
		Zielkarte: lauf.bezug.KarteID,
	})
	if fehler != nil {
		return mitFehler(ctx, fehler)
	}

	// **Der Posten wird geschlossen, gleich wie es ausgeht** (#1542).
	defer a.kiPostenSchliessen(ctx, angelegt.Kostenposten)

	if lauf.art == domain.EinschaetzungBalance {
		return a.balanceMit(ctx, lauf, satz)
	}

	roh, fehler := satz.quelle.Vorschlagen(ctx, KIEinschaetzungsanweisung,
		lauf.bezug.Alias, auftragMitBelegen(lauf.paket, lauf.bezug.Frage),
		einschaetzungsschema())
	if fehler != nil {
		return mitFehler(ctx, fehler)
	}

	gelesen, fehler := einschaetzungAus(roh)
	if fehler != nil {
		return mitFehler(ctx, fehler)
	}

	belege := lauf.paket.Manifest.Belege
	geprueft := domain.PruefeEinschaetzung(lauf.art, gelesen.Einschaetzung, belege)

	return lauf.antwortAus(geprueft, alsListe(gelesen.Einschraenkungen),
		roh.Modell)
}

// antwortAus baut die gemeinsame Antwort — ohne die Balanceangaben.
func (lauf einschaetzungslauf) antwortAus(
	geprueft domain.KIEinschaetzung, einschraenkungen []string, modell string,
) KIEinschaetzungAntwort {
	belege := lauf.paket.Manifest.Belege

	return KIEinschaetzungAntwort{
		Art: string(geprueft.Art), Satz: geprueft.Satz,
		Belegt: geprueft.Belegt(), Vorbehalt: geprueft.Vorbehalt,
		Befunde:          alsBefundDTOs(geprueft.Befunde),
		Einschraenkungen: einschraenkungen,
		Belege:           alsBelegDTOs(belege),
		Korpusstaende:    alsKorpusstandDTOs(lauf.paket.Manifest.Korpusstaende),
		Anbieter:         lauf.bezug.Anbieter, Modell: modell,
	}
}

// auftragMitBelegen setzt die Belege vor die Frage.
//
// **Die Belege zuerst und die Frage zuletzt.** Ein Modell richtet sich am
// Ende der Eingabe aus; wer die Frage nach oben stellte, bekäme eine Antwort
// auf die letzte Regelstelle statt auf die Frage.
func auftragMitBelegen(paket service.Kontextpaket, frage string) string {
	return "Belege:\n" + paket.Kontext + "\n\nFrage: " + frage
}

// einschaetzungAus liest die Antwort des Modells.
func einschaetzungAus(roh KIRohantwort) (einschaetzungsantwort, error) {
	// **Eine abgeschnittene Antwort wird ganz verworfen** — dieselbe Regel
	// wie beim Vorschlag (A7): ein halber Satz mit einer halben Fußnote
	// sähe aus wie eine Aussage.
	if roh.Abgeschnitten {
		return einschaetzungsantwort{}, fmt.Errorf(
			"%w: die Antwort des Modells ist abgeschnitten", domain.ErrValidation)
	}

	var gelesen einschaetzungsantwort
	if fehler := json.Unmarshal([]byte(roh.Inhalt), &gelesen); fehler != nil {
		return einschaetzungsantwort{}, fmt.Errorf(
			"%w: die Antwort des Modells ist kein JSON: %s",
			domain.ErrValidation, fehler.Error())
	}

	if strings.TrimSpace(gelesen.Einschaetzung) == "" {
		return einschaetzungsantwort{}, fmt.Errorf(
			"%w: das Modell hat nichts gesagt", domain.ErrValidation)
	}

	return gelesen, nil
}

// alsBefundDTOs bringt die Belegbefunde in die Gestalt des Fensters.
func alsBefundDTOs(befunde []domain.Belegbefund) []EinschaetzungsbefundDTO {
	zeilen := make([]EinschaetzungsbefundDTO, 0, len(befunde))
	for _, befund := range befunde {
		zeilen = append(zeilen, EinschaetzungsbefundDTO{
			Art: befund.Art, Stelle: befund.Stelle, Meldung: befund.Meldung,
		})
	}

	return zeilen
}

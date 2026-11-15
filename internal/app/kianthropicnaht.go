package app

// **Die Naht zu Anthropic** — Werkstück A8 von Phase 5.3
// (#1447).
//
// # Was hier steht und was im Blatt
//
// **Hier die Verdrahtung, dort die Sache.** `internal/anthropic` weiß, wie
// die Messages-API aussieht; diese Datei weiß, woher der Schlüssel kommt,
// wann gefragt werden darf und wie die Aufschrift dieser Anwendung lautet.
//
// # Der Schlüssel liegt im Schlüsselbund
//
// **Das Issue verlangt es**: „API-Key bleibt Schlüsselbund." Er steht nicht
// im Bestand — den gibt man weiter (§2, `internal/schluesselbund`) — und er
// steht in keiner Meldung.

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Die Fristen für Anthropic.
//
// # Warum sie kürzer sind als bei der lokalen Laufzeit
//
// **Weil dort ein Modell geladen wird und hier eines läuft.** Ollama legt
// Gigabyte in den Speicher, bevor das erste Zeichen entsteht; die
// Messages-API antwortet oder nicht. Die Dokumentation rät bei langen Rufen
// zum Strom — und dieses Programm fragt kurz.
const (
	kiverbindungsfristAnthropic = 10 * time.Second
	kikopffristAnthropic        = 2 * time.Minute
	kistillstandsfristAnthropic = 60 * time.Second
)

// anthropicnaht baut den Klienten samt Riegel und Drossel.
//
// **Die Drossel kommt aus demselben [drosselwerk] wie die der
// Kartendienste**: sie gehört zur Gegenstelle und nicht zur Verbindung, und
// zwei Nähte auf dieselbe Adresse teilen sie sich damit (§16, Regel 4).
// **Die Adresse steht im Aufruf und nicht im Quelltext**, obwohl Anthropic
// nur eine hat: sonst gäbe es keinen Weg, diese Naht zu prüfen, ohne den
// echten Dienst zu rufen — und ein Lauf gegen ihn kostete Geld (§3).
// `kiVerdrahten` reicht die leere Zeichenkette und bekommt damit die
// Vorgabeadresse.
func (a *App) anthropicnaht(adresse string) KILaufzeitquelle {
	return anthropicquelle{Klient: anthropic.NeuerKlient(adresse,
		a.netzklientMit(netz.Fristen{
			Verbindung: kiverbindungsfristAnthropic,
			Kopf:       kikopffristAnthropic,
			Stillstand: kistillstandsfristAnthropic,
		}).MitErlaubnis(a.pruefeAnthropic),
		anthropic.Zugang{Schluessel: a.anthropicschluessel(adresse)}).
		MitDrossel(drosselwerk.Fuer(anthropic.Vorgabeadresse, anthropic.Abstand))}
}

// pruefeAnthropic fragt **vor jeder Anfrage**, ob sie hinausgehen darf.
func (a *App) pruefeAnthropic(ctx context.Context) error {
	return a.pruefeCloudanbieter(ctx, domain.DienstAnthropic, "Anthropic")
}

// pruefeCloudanbieter ist der Riegel, den sich alle Cloudanbieter teilen.
//
// **Seit A9 gemeinsam** (#1448): die beiden Nähte unterschieden sich in
// nichts als dem Dienst und dem Namen, und zwei gleichlautende Riegel laufen
// beim ersten Umbau auseinander — genau die Gestalt, vor der §10 warnt.
//
// # Zwei Gründe, und beide halten
//
//   - **»Nur lokal«.** Ein Cloudanbieter rechnet nirgends auf diesem
//     Rechner — unter der Regel geht **gar nichts** hinaus, auch nicht die
//     Modellliste. Schon sie wäre eine Auskunft nach draußen.
//   - **Der Schalter.** `dienst.<anbieter>` fängt aus an; wer ihn umlegt,
//     während ein Vorschlag läuft, hält den nächsten Ruf auf.
//
// # Warum die Regel zuerst geprüft wird
//
// **Weil sie den Schalter mit abschaltet.** `richtlinie.nurlokal` setzt den
// **wirksamen** Wert jedes `dienst.*` auf »aus« — und wer nur den Schalter
// läse, meldete „Anthropic ist abgeschaltet". Das schickte jemanden an den
// falschen Schalter: er legte ihn um, und es ginge trotzdem nicht (§8).
// Angestrichen hat es der Lauf über »nur lokal« in
// `kianthropicquelle_intern_test.go`.
func (a *App) pruefeCloudanbieter(
	ctx context.Context, dienst domain.Dienst, name string,
) error {
	if a.einstellungen == nil {
		return fehlenderBestand()
	}

	regel, fehler := a.ortsregel(ctx)
	if fehler != nil {
		return fehler
	}

	if regel.NurLokal {
		return fmt.Errorf(
			"%w: hier gilt »nur lokal«, und %s rechnet nicht auf diesem "+
				"Rechner", domain.ErrErlaubnis, name)
	}

	stand, fehler := a.einstellungen.Stand(ctx)
	if fehler != nil {
		return fehler
	}

	if stand[domain.DienstEinstellung(dienst)].Wirksam != domain.DienstAn {
		return fmt.Errorf("%w: %s ist abgeschaltet", domain.ErrErlaubnis, name)
	}

	return nil
}

// anthropicschluessel holt ihn aus dem Schlüsselbund.
//
// **Leer heißt: es gibt keinen** — und dann weist das Blatt die Anfrage ab,
// bevor sie eine wird ([anthropic.Zugang.Fehlt]).
func (a *App) anthropicschluessel(adresse string) string {
	return a.cloudschluesselFuer(domain.DienstAnthropic, adresse)
}

// cloudschluessel holt den Schlüssel eines Dienstes aus dem Schlüsselbund.
//
// **Auch dieser Weg ist seit A9 gemeinsam**: er stand zweimal da, Zeichen für
// Zeichen gleich bis auf den Dienst.
//
// # Dort, wo das Fenster ablegt — und nicht daneben
//
// **Bis zum 2026-09-12 suchte er woanders**
// (#1537):
// `Lies(string(dienst), ZugangSchluessel)`, also `openai / schluessel`, während
// `App.ZugangSetzen` nach `MagicSetEditor4Mac / openai.schluessel` schreibt.
// Zwei Orte, die sich nicht treffen — **der Klient bekam eine leere
// Zeichenkette**, und Anthropic, OpenAI und xAI waren unbenutzbar.
//
// **Am Fenster sah alles richtig aus**, weil `Zugaenge()` die richtige Gestalt
// liest und „hinterlegt" meldete.
//
// **Kein Lauf sah es**, weil jeder beide Hälften selbst setzte — die Naht
// zwischen zwei Prüfebenen (§3). Heute fährt
// `zugangsweg_intern_test.go` den Weg von der Eingabe bis zum Klienten, für
// **jeden** Dienst.
// cloudschluesselFuer holt den Schlüssel, der **zu dieser Adresse** gehört.
//
// # Warum der Schlüssel an der Adresse hängt
//
// **Ein Schlüssel für ein Gateway ist kein Schlüssel für den Anbieter**
// (A18, #1533). Wer `dienst.openai.adresse` auf ein Firmengateway stellt,
// dort seinen Schlüssel einträgt und die Adresse später leert, schickte
// denselben Schlüssel an `api.openai.com` — eine Weitergabe an jemanden, für
// den er nie ausgestellt wurde.
//
// **Die Vorgabe behält ihren alten Ort** ([domain.ZugangKontoFuer]), damit
// ein eingetragener Schlüssel den Umbau überlebt.
func (a *App) cloudschluesselFuer(dienst domain.Dienst, adresse string) string {
	schluessel, fehler := a.bundFuerDenMarkt().Lies(context.Background(),
		Schluesselbunddienst,
		domain.ZugangKontoFuer(dienst, domain.ZugangSchluessel, adresse))
	if fehler != nil {
		return ""
	}

	return schluessel
}

// dienstadresse liest die eingestellte Adresse eines Dienstes.
//
// **Leer heißt: die Vorgabe des Anbieters**, und sie steht im Blatt (§2).
//
// # Wer sie ruft — und wer nicht
//
// **Nicht die Naht.** Die bekommt die Adresse als Parameter durchgereicht
// ([App.cloudnaehte]), und ein zweites Lesen dort hätte einen zusätzlichen
// Griff in die Einstellungen bedeutet — gemessen am 2026-09-12: ein Doppel,
// das beim zweiten Ruf scheitert, kippte dadurch an einer **anderen** Stelle,
// und ein Fehlerzweig wurde unerreichbar, ohne dass jemand ihn angefasst
// hätte.
//
// **Sondern die Freigabe** ([App.ortFuer]): sie steht außerhalb der
// Nahtverdrahtung und kennt nur den Anbieternamen.
//
// **Mit dem Kontext des Aufrufers** (Befund 3 des Zweitblicks über Phase D):
// er trägt das Profil, und ohne es scheiterte das Lesen still — ein Gateway
// auf der Schleife galt dann als Wolke.
//
// **Und ein Lesefehler ist keine fehlende Adresse** (Rest desselben Befunds,
// angestrichen im Nachlauf): beides gab „", und die Freigabe nannte dann die
// Wolke des Anbieters statt des eingestellten Gateways.
func (a *App) dienstadresse(ctx context.Context, dienst domain.Dienst) (string, error) {
	if a.einstellungen == nil {
		return "", nil
	}

	stand, fehler := a.einstellungen.Stand(ctx)
	if fehler != nil {
		return "", fehler
	}

	return strings.TrimSpace(stand[domain.DienstAdresse(dienst)].Wirksam), nil
}

// anthropicquelle bringt den Klienten auf die Gestalt, die diese Naht braucht.
//
// **Eine Hülle und keine zweite Fachlogik**: `Modelle` kommt unverändert aus
// dem eingebetteten Klienten, `Fassung` ist eine Konstante, und
// `Vorschlagen` übersetzt nur die Aufschrift.
type anthropicquelle struct{ *anthropic.Klient }

// Fassung nennt die Fassung der Schnittstelle.
//
// **Keine Anfrage.** Anders als bei Ollama gibt es hier kein Programm, dessen
// Fassung man lesen könnte — was feststeht, ist die Fassung, unter der
// gefragt wird, und die steht im Quelltext.
func (anthropicquelle) Fassung(context.Context) (string, error) {
	return "anthropic-version " + anthropic.Fassung, nil
}

// Vorschlagen fragt das Modell mit fester Form.
func (a anthropicquelle) Vorschlagen(
	ctx context.Context, anweisung, alias, auftrag string, schema map[string]any,
) (KIRohantwort, error) {
	return a.frage(ctx, anthropic.Frage{
		Alias: alias, System: anweisung, Auftrag: auftrag, Schema: schema,
	})
}

// Sehen legt ein Bild dazu — Werkstück A13.3 (#1452).
//
// **Derselbe Weg mit einem Block mehr**: `anthropic.Frage` trägt `Bilder`
// seit A8, und bis A13.3 hat niemand sie gefüllt. Die Fehlerwege sind
// dieselben, und das ist der Grund, warum hier keine zweite Fassung steht.
func (a anthropicquelle) Sehen(
	ctx context.Context, wunsch KISichtwunsch,
) (KIRohantwort, error) {
	return a.frage(ctx, anthropic.Frage{
		Alias: wunsch.Alias, System: wunsch.Anweisung,
		Auftrag: wunsch.Auftrag, Schema: wunsch.Schema,
		Bilder: []anthropic.Bild{
			{Daten: wunsch.Bild, Medientyp: wunsch.Medientyp},
		},
	})
}

// frage fährt eine Frage und übersetzt ihren Ausgang.
//
// **Einmal und nicht zweimal**: Vorschlag und Sicht unterscheiden sich in
// **einem** Block, und zwei Fassungen der Ausgangsprüfung liefen beim ersten
// neuen `stop_reason` auseinander.
func (a anthropicquelle) frage(
	ctx context.Context, gestellt anthropic.Frage,
) (KIRohantwort, error) {
	antwort, fehler := a.Rede(ctx, gestellt)
	if fehler != nil {
		return KIRohantwort{}, fehler
	}

	// **Eine Ablehnung ist ein Fehler dieses Rufs und keine leere Antwort.**
	//
	// Das Issue verlangt sie getrennt: wer sie als Inhalt durchreichte,
	// bekäme ein leeres JSON und eine Schemaprüfung, die von fehlenden
	// Feldern spräche — statt davon, dass das Modell nicht wollte.
	if antwort.Ausgang == anthropic.AusgangAbgelehnt {
		return KIRohantwort{}, fmt.Errorf(
			"%w: das Modell hat die Antwort abgelehnt (%s)",
			domain.ErrValidation, antwort.Naeheres)
	}

	if antwort.Ausgang == anthropic.AusgangUnbekannt {
		return KIRohantwort{}, fmt.Errorf(
			"%w: die Antwort endete mit %q — dieses Programm kennt den Grund "+
				"nicht und liest sie deshalb nicht als fertig",
			domain.ErrValidation, antwort.Grund)
	}

	return KIRohantwort{
		Inhalt:        antwort.Inhalt,
		Abgeschnitten: antwort.Ausgang == anthropic.AusgangUnvollstaendig,
		Modell:        antwort.Modell,
	}, nil
}

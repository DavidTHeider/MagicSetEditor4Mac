package app

// **Die Naht zu OpenAI** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Warum sie neben der von Anthropic steht und nicht in ihr
//
// **Weil zwei Anbieter derselben Fähigkeit zwei Blätter sind** (§2). Was sie
// gemeinsam haben, steht in `kianbieter.go` — die Liste, der Stand, die
// Frage, ob gefragt werden darf. Was sie unterscheidet, steht je Naht: hier
// die Aufschrift `Authorization: Bearer`, dort `x-api-key`; hier ein Ausgang
// aus drei Feldern, dort einer aus einem.

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
)

// Die Fristen für OpenAI.
//
// **Länger im Stillstand als bei Anthropic.** Ein Denkmodell schweigt,
// während es denkt — `reasoning_tokens` entstehen, ohne dass ein Zeichen
// ankommt. Zwei Minuten sind großzügig und trotzdem eine Grenze.
const (
	kiverbindungsfristOpenAI = 10 * time.Second
	kikopffristOpenAI        = 3 * time.Minute
	kistillstandsfristOpenAI = 2 * time.Minute
)

// openainaht baut den Klienten samt Riegel und Drossel.
//
// **Die Adresse steht im Aufruf und nicht im Quelltext**, obwohl OpenAI nur
// eine hat: sonst gäbe es keinen Weg, diese Naht zu prüfen, ohne den echten
// Dienst zu rufen — und ein Lauf gegen ihn kostete Geld (§3).
func (a *App) openainaht(adresse string) KILaufzeitquelle {
	return openaiquelle{Klient: openai.NeuerKlient(adresse,
		a.netzklientMit(netz.Fristen{
			Verbindung: kiverbindungsfristOpenAI,
			Kopf:       kikopffristOpenAI,
			Stillstand: kistillstandsfristOpenAI,
		}).MitErlaubnis(a.pruefeOpenAI),
		openai.Zugang{Schluessel: a.openaischluessel(adresse)}).
		MitDrossel(drosselwerk.Fuer(openai.Vorgabeadresse, openai.Abstand))}
}

// pruefeOpenAI fragt **vor jeder Anfrage**, ob sie hinausgehen darf.
//
// **Dieselbe Reihenfolge wie bei Anthropic**, und aus demselben Grund: die
// Regel schaltet den Schalter mit ab, und wer nur ihn läse, schickte jemanden
// an den falschen (§8).
func (a *App) pruefeOpenAI(ctx context.Context) error {
	return a.pruefeCloudanbieter(ctx, domain.DienstOpenAI, "OpenAI")
}

// openaischluessel holt ihn aus dem Schlüsselbund.
func (a *App) openaischluessel(adresse string) string {
	return a.cloudschluesselFuer(domain.DienstOpenAI, adresse)
}

// openaiquelle bringt den Klienten auf die Gestalt, die diese Naht braucht.
type openaiquelle struct{ *openai.Klient }

// Fassung nennt die Fassung der Schnittstelle.
//
// **Keine Anfrage, und anders als bei Anthropic auch keine Zahl.** Die
// Responses-API trägt keine Fassungsangabe — es gibt kein
// `anthropic-version`. Was feststeht, ist der Weg.
func (openaiquelle) Fassung(context.Context) (string, error) {
	return "Responses-API " + openai.Antwortweg, nil
}

// Vorschlagen fragt das Modell mit fester Form.
func (o openaiquelle) Vorschlagen(
	ctx context.Context, anweisung, alias, auftrag string, schema map[string]any,
) (KIRohantwort, error) {
	antwort, fehler := o.Rede(ctx, openai.Frage{
		Alias: alias, System: anweisung, Auftrag: auftrag, Schema: schema,
	})
	if fehler != nil {
		return KIRohantwort{}, fehler
	}

	// **Eine Ablehnung ist ein Fehler dieses Rufs und keine leere Antwort.**
	if antwort.Ausgang == openai.AusgangAbgelehnt {
		return KIRohantwort{}, fmt.Errorf(
			"%w: das Modell hat die Antwort abgelehnt (%s)",
			domain.ErrValidation, ablehnungsgrund(antwort))
	}

	if antwort.Ausgang == openai.AusgangUnbekannt {
		return KIRohantwort{}, fmt.Errorf(
			"%w: die Antwort endete mit %q — dieses Programm kennt den Grund "+
				"nicht und liest sie deshalb nicht als fertig",
			domain.ErrValidation, antwort.Grund)
	}

	return KIRohantwort{
		Inhalt:        antwort.Inhalt,
		Abgeschnitten: antwort.Ausgang == openai.AusgangUnvollstaendig,
		Modell:        antwort.Modell,
	}, nil
}

// ablehnungsgrund nennt, was das Modell gesagt hat — oder den Stand.
//
// **Zwei Wege führen zu einer Ablehnung**: das Modell sagt einen Satz, oder
// ein Filter hält den Lauf an und sagt nichts. Wer nur den Satz zeigte,
// zeigte im zweiten Fall nichts.
func ablehnungsgrund(antwort openai.Antwort) string {
	if antwort.Naeheres != "" {
		return antwort.Naeheres
	}

	return antwort.Grund
}

// Bild erzeugt eines oder bearbeitet eines — Werkstück A12 (#1451).
//
// **Nur diese Naht hat die Methode**, und damit erfüllt nur sie
// [KIBildquelle]. Das ist die ganze Prüfung: wer sie nicht hat, wird gar
// nicht erst gefragt (siehe `kibildquelle.go`).
func (o openaiquelle) Bild(ctx context.Context, wunsch KIBildwunsch) (KIRohbild, error) {
	roh, fehler := o.Klient.Bild(ctx, openai.Bildauftrag{
		Alias: wunsch.Alias, Auftrag: wunsch.Auftrag,
		Groesse: wunsch.Groesse, Guete: wunsch.Guete, Quelle: wunsch.Quelle,
	})
	if fehler != nil {
		return KIRohbild{}, fehler
	}

	return KIRohbild{
		Inhalt: roh.Inhalt, Medientyp: roh.Medientyp,
		Modell: roh.Modell, Umschrift: roh.Umschrift,
	}, nil
}

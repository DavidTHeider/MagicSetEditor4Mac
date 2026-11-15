package app

// **Die Naht zu xAI** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Warum sie neben den beiden anderen steht und nicht in ihnen
//
// **Weil drei Anbieter derselben Fähigkeit drei Blätter sind** (§2). Was sie
// gemeinsam haben, steht in `kianbieter.go` — die Liste, der Stand, die
// Frage, ob gefragt werden darf. Was sie unterscheidet, steht je Naht.
//
// **Und hier ist der Unterschied nicht nur die Aufschrift**: xAI kennt einen
// vierten Ausgang (`end_turn`), nennt den Preis jedes Rufs in Ticks und
// unterscheidet zwischen zwei eigenen Endpunkten mit **zwei verschiedenen**
// Bildformen. Wer diese Naht aus der von OpenAI abgeleitet hätte, hätte drei
// Fehler eingebaut, von denen zwei erst beim Anbieter aufgefallen wären.

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/xai"
)

// Die Fristen für xAI.
//
// **Dieselben wie bei OpenAI und aus demselben Grund**: die Grok-Modelle
// denken, bevor sie antworten, und während sie denken, kommt kein Zeichen an.
const (
	kiverbindungsfristXAI = 10 * time.Second
	kikopffristXAI        = 3 * time.Minute
	kistillstandsfristXAI = 2 * time.Minute
)

// xainaht baut den Klienten samt Riegel und Drossel.
//
// **Die Adresse steht im Aufruf und nicht im Quelltext**, obwohl xAI nur eine
// hat: sonst gäbe es keinen Weg, diese Naht zu prüfen, ohne den echten Dienst
// zu rufen — und ein Lauf gegen ihn kostete Geld (§3).
func (a *App) xainaht(adresse string) KILaufzeitquelle {
	return xaiquelle{Klient: xai.NeuerKlient(adresse,
		a.netzklientMit(netz.Fristen{
			Verbindung: kiverbindungsfristXAI,
			Kopf:       kikopffristXAI,
			Stillstand: kistillstandsfristXAI,
		}).MitErlaubnis(a.pruefeXAI),
		xai.Zugang{Schluessel: a.xaischluessel(adresse)}).
		MitDrossel(drosselwerk.Fuer(xai.Vorgabeadresse, xai.Abstand))}
}

// pruefeXAI fragt **vor jeder Anfrage**, ob sie hinausgehen darf.
func (a *App) pruefeXAI(ctx context.Context) error {
	return a.pruefeCloudanbieter(ctx, domain.DienstXAI, "xAI")
}

// xaischluessel holt ihn aus dem Schlüsselbund.
func (a *App) xaischluessel(adresse string) string {
	return a.cloudschluesselFuer(domain.DienstXAI, adresse)
}

// xaiquelle bringt den Klienten auf die Gestalt, die diese Naht braucht.
type xaiquelle struct{ *xai.Klient }

// Fassung nennt die Fassung der Schnittstelle.
//
// **Keine Anfrage und keine Zahl**, wie bei OpenAI: es gibt kein
// `anthropic-version`. Was feststeht, ist der Weg — und der ist hier die
// Auskunft, denn xAI hat zwei, und nur einer legt nichts ab.
func (xaiquelle) Fassung(context.Context) (string, error) {
	return "Chat-Completions " + xai.Unterhaltungsweg, nil
}

// Vorschlagen fragt das Modell mit fester Form.
func (x xaiquelle) Vorschlagen(
	ctx context.Context, anweisung, alias, auftrag string, schema map[string]any,
) (KIRohantwort, error) {
	antwort, fehler := x.Rede(ctx, xai.Frage{
		Alias: alias, System: anweisung, Auftrag: auftrag, Schema: schema,
	})
	if fehler != nil {
		return KIRohantwort{}, fehler
	}

	// **Eine Ablehnung ist ein Fehler dieses Rufs und keine leere Antwort.**
	//
	// **Und sie trägt hier immer ihren Satz.** Bei OpenAI gibt es zwei Wege
	// dorthin — das Modell sagt etwas, oder ein Filter hält an und sagt
	// nichts —, und die Naht dort fängt beide ab. **Bei xAI gibt es nur den
	// ersten**: eine Ablehnung entsteht ausschließlich aus einem gefüllten
	// `refusal` ([xai.Ausgang]). Ein Ausweichtext daneben wäre ein Zweig, den
	// nichts erreicht — abgeschrieben statt nachgelesen (§15).
	if antwort.Ausgang == xai.AusgangAbgelehnt {
		return KIRohantwort{}, fmt.Errorf(
			"%w: das Modell hat die Antwort abgelehnt (%s)",
			domain.ErrValidation, antwort.Naeheres)
	}

	if antwort.Ausgang == xai.AusgangUnbekannt {
		return KIRohantwort{}, fmt.Errorf(
			"%w: die Antwort endete mit %q — dieses Programm kennt den Grund "+
				"nicht und liest sie deshalb nicht als fertig",
			domain.ErrValidation, antwort.Grund)
	}

	return KIRohantwort{
		Inhalt:        antwort.Inhalt,
		Abgeschnitten: antwort.Ausgang == xai.AusgangUnvollstaendig,
		Modell:        antwort.Modell,
	}, nil
}

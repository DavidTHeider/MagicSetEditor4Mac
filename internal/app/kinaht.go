package app

// **Die Naht zur KI** — Werkstück A6 von Phase 5.3
// (#1445).
//
// # Was sie heute zeigt
//
// **Was diese Anwendung über KI-Modelle weiß — und das ist heute nichts.**
// Der Katalog ist leer, weil kein Anbieter angebunden ist; A7
// (#1446)
// bringt den ersten.
//
// **Und das ist keine Verlegenheit, sondern die Auskunft.** Das Issue sagt:
// „Unbekannt bleibt unbekannt; Modelllisten garantieren keine Capability."
// Eine Seite, die zwölf Modellnamen zeigt, von denen keiner geprüft ist, wäre
// schlechter als eine, die sagt: es steht keines fest.

import (
	"context"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// fehlenderBestand ist die Meldung, wenn keiner offen ist.
func fehlenderBestand() error {
	return fmt.Errorf("%w: es ist kein Bestand offen", domain.ErrValidation)
}

// KIModellDTO ist ein Modell, wie das Fenster es sieht.
type KIModellDTO struct {
	Anbieter string `json:"anbieter"`
	Alias    string `json:"alias"`
	Laufzeit string `json:"laufzeit"`

	// Ort ist `lokal` oder `cloud` — **wie der Anbieter es erklärt**.
	Ort string `json:"ort"`

	// Faehigkeiten sind die **belegten** — jede mit Datum und Beleg.
	Faehigkeiten []KIBelegDTO `json:"faehigkeiten"`

	// Herkunft ist die Kette bis zum ausgeführten Artefakt — leer heißt
	// **unbekannt**, und das steht dann auch da.
	Herkunft []KIGliedDTO `json:"herkunft"`

	// Bindung ist Alias und Artefakt zusammen; sie ändert sich, wenn hinter
	// dem Namen etwas anderes liegt.
	Bindung string `json:"bindung"`

	// Kontextfenster ist, wie viele Tokens das Modell auf einmal liest — 0
	// heißt nicht angegeben; kontextbeleg sagt, woher (AI-01.AK-02).
	Kontextfenster int    `json:"kontextfenster"`
	KontextBeleg   string `json:"kontextbeleg"`
}

// KIBelegDTO ist eine belegte Fähigkeit.
type KIBelegDTO struct {
	Faehigkeit string `json:"faehigkeit"`
	Datum      string `json:"datum"`
	Beleg      string `json:"beleg"`
}

// KIGliedDTO ist ein Glied der Herkunftskette.
type KIGliedDTO struct {
	Art      string `json:"art"`
	Kennung  string `json:"kennung"`
	Aus      string `json:"aus"`
	Quelle   string `json:"quelle"`
	Revision string `json:"revision"`
	Lizenz   string `json:"lizenz"`
}

// KIKatalogDTO ist der Katalog samt der Regel, die für ihn gilt.
type KIKatalogDTO struct {
	// NurLokal sagt, ob etwas diesen Rechner verlassen darf.
	//
	// **Sie steht beim Katalog und nicht in einer Fußnote**: welche Modelle
	// es gibt, ist die halbe Auskunft; ob sie gefragt werden dürfen, die
	// andere.
	NurLokal bool `json:"nurLokal"`

	// Einrichtung ist der Stand, der am weitesten kommt — einer der fünf aus
	// `kilaufzeit.go`.
	//
	// **Der beste über alle Anbieter** (A8, #1447): ein eingerichteter macht
	// die Anwendung brauchbar, auch wenn ein zweiter aus ist. Welcher woran
	// hängt, steht je Zeile unter [KIKatalogDTO.Anbieter].
	Einrichtung string `json:"einrichtung"`

	// Anbieter sind die Quellen, jede mit ihrem eigenen Stand.
	//
	// **Eine Liste seit A8** (#1447). A7 trug Einrichtung, Fassung und
	// Adresse **flach**; mit dem zweiten Anbieter sprächen sie von einem von
	// zweien und sähen aus, als sprächen sie von beiden.
	Anbieter []KIAnbieterDTO `json:"anbieter"`

	// Cloud sagt, was über die Auslagerung der **lokalen** Laufzeit
	// feststeht — und was nicht.
	Cloud KICloudDTO `json:"cloud"`

	// Befunde sind die Sätze, die über alle Anbieter hinweg gelten.
	Befunde []string `json:"befunde"`

	Modelle []KIModellDTO `json:"modelle"`

	// Vorrang nennt, wen ein Vorschlag mit Vorrang für Name oder
	// Flavortext fragt, als `anbieter/modell` und in dieser Reihenfolge —
	// #2054: es steht am Formular, bevor jemand fragt.
	Vorrang []string `json:"vorrang"`

	// Vorrangbefunde sind die Sätze, die **die Reihe** betreffen — ein
	// gewähltes Modell, das fehlt oder nicht fest antworten kann. Sie stehen
	// auch unter [KIKatalogDTO.Befunde]; eigens hier, damit das Formular den
	// seinen findet (Z4 aus dem Zweitblick über Phase 6.11).
	Vorrangbefunde []string `json:"vorrangbefunde"`
}

// KIAnbieterDTO ist ein Anbieter samt dem, was über ihn feststeht.
type KIAnbieterDTO struct {
	Name string `json:"name"`

	// Einrichtung ist einer der fünf Stände aus `kilaufzeit.go`.
	Einrichtung string `json:"einrichtung"`

	// Laufzeit ist die Fassung — bei Anthropic die der Schnittstelle, bei
	// Ollama die des Programms. Leer heißt: es hat nichts geantwortet.
	Laufzeit string `json:"laufzeit"`

	// Adresse ist, wohin gefragt wurde — leer bei einem Anbieter mit fester.
	Adresse string `json:"adresse"`

	// Ort ist `lokal` oder `cloud` — **wo dieser Anbieter rechnet**.
	//
	// **Neu am 2026-09-12** (#1539):
	// die Fläche muss wissen, ob eine Messung Geld kostet, **bevor** sie
	// hinausgeht. Ohne diese Angabe stand am Fenster die Wahl zwischen einer
	// Freigabe, die niemand braucht, und einem Ruf, der an einer fehlenden
	// scheitert — beides ohne eine Möglichkeit, es zu unterscheiden.
	//
	// **Am Anbieter und nicht nur am Modell**: gemessen wird ein Modell, über
	// das noch **nichts** feststeht — es steht nicht im Katalog, und sein
	// [KIModellDTO.Ort] gibt es deshalb nicht.
	Ort string `json:"ort"`

	// Befunde sind die Sätze **dieses** Anbieters.
	Befunde []string `json:"befunde"`

	// Messbar sind die Modelle, die er **nennt**, ohne etwas über sie zu
	// sagen (A9, #1448) — sie stehen nicht im Katalog, und ohne diese Liste
	// wären sie unerreichbar.
	Messbar []string `json:"messbar"`

	// Gemessen sind die Modelle, für die eine Messung im Bestand liegt —
	// sie lässt sich zurücknehmen.
	Gemessen []string `json:"gemessen"`

	// SitzungAus sagt, dass dieser Anbieter **für diese Sitzung**
	// abgeschaltet ist — nur das Gerätemodell kennt das; nach einem
	// Neustart gilt wieder die Einstellung (AI-05.AK-04, #2042).
	SitzungAus bool `json:"sitzungAus"`
}

// KICloudDTO ist der Cloudstand samt Fundstelle.
type KICloudDTO struct {
	// Stand ist `aus`, `an` oder `ungewiss`.
	//
	// **Drei und nicht zwei**: „keine Datei" heißt ungewiss, denn die zweite
	// dokumentierte Abschaltung steht in der Umgebung eines fremden
	// Prozesses und ist von hier nicht zu lesen.
	Stand string `json:"stand"`

	Beleg string `json:"beleg"`
}

// KIKatalogantwort trägt ihn samt Fehler.
type KIKatalogantwort struct {
	KIKatalogDTO

	Fehler *Fehler `json:"fehler"`
}

// KIModelle nennt, was über KI-Modelle feststeht.
//
// **Seit A7 (#1446) fragt sie die Laufzeit** — und nur, wenn eine
// eingeschaltet ist. Was zurückkommt, ist nicht „was es gibt", sondern was
// **belegt** ist: je Modell die geprüften Fähigkeiten, die Herkunft bis zum
// ausgeführten Artefakt und der Ort, an dem es rechnet.
//
// **Eine leere Liste bleibt eine Auskunft**, und `einrichtung` sagt, welche:
// niemand hat eine Laufzeit eingeschaltet, es antwortet keine, es ist keines
// installiert — oder es ist eines da und kann nichts Belegtes.
func (a *App) KIModelle(ctx context.Context) KIKatalogantwort {
	ctx = a.imProfil(ctx)

	regel, fehler := a.ortsregel(ctx)
	if fehler != nil {
		return KIKatalogantwort{Fehler: Uebersetze(ctx, fehler)}
	}

	staende, anbieter, fehler := a.frageAnbieter(ctx)
	if fehler != nil {
		return KIKatalogantwort{Fehler: Uebersetze(ctx, fehler)}
	}

	gefunden := slices.Concat(a.kimodelle, alleModelle(staende))

	katalog, fehler := service.NeuerModellkatalog(gefunden...)
	if fehler != nil {
		return KIKatalogantwort{Fehler: Uebersetze(ctx, fehler)}
	}

	reihe, vorrangbefunde := vorrangreihe(staende, anbieter, a.vorrangmodellwahl(ctx))

	return KIKatalogantwort{KIKatalogDTO: KIKatalogDTO{
		NurLokal:    regel.NurLokal,
		Einrichtung: besteEinrichtung(staende),
		Anbieter:    zuAnbieterDTOs(staende),
		Cloud:       a.cloudstand(),
		Befunde:     alsSaetze(append(vonDerRichtlinieAbgeschaltet(anbieter), vorrangbefunde...)),
		Modelle:     zuKIModellDTOs(katalog.Alle()),
		Vorrang:     reihe,

		Vorrangbefunde: alsSaetze(vorrangbefunde),
	}}
}

// zuAnbieterDTOs macht aus den Ständen ihre DTOs.
func zuAnbieterDTOs(staende []anbieterstand) []KIAnbieterDTO {
	dtos := make([]KIAnbieterDTO, 0, len(staende))

	for _, stand := range staende {
		dtos = append(dtos, KIAnbieterDTO{
			Name: stand.name, Einrichtung: stand.einrichtung,
			Laufzeit: stand.fassung, Adresse: stand.adresse,
			Ort:     string(ortVon(stand.name, stand.adresse)),
			Befunde: alsSaetze(stand.befunde),
			Messbar: alsListe(stand.messbar), Gemessen: alsListe(stand.gemessen),
			SitzungAus: stand.sitzungAus,
		})
	}

	return dtos
}

// alsListe macht aus einem `nil`-Schnitt eine leere Liste.
//
// # Warum das an der Naht steht und nicht im Fenster
//
// **Weil `nil` zu `null` wird und `null.length` die Seite beendet.** Ein
// Anbieter, der aus ist, trägt keine messbaren Modelle — und die Antwort
// muss trotzdem eine Liste zeigen.
//
// **Gefunden am 2026-09-12 in der Sichtprüfung**, nicht von einem Test: die
// Frontendtests gehen über `test/api.ts` und laufen nie durch die
// Serialisierung; die Go-Tests sehen die Struct-Marken nicht (§3). Seither
// hält `TestKeineListeDerKIAntwortIstNull` es.
func alsListe(werte []string) []string {
	if werte == nil {
		return []string{}
	}

	return werte
}

// alsSaetze macht aus Befunden ihre Meldungen.
//
// **Eine leere Liste bleibt eine Liste** — ein `null` sähe im Fenster aus wie
// „konnte nicht beantwortet werden".
func alsSaetze(befunde []domain.Befund) []string {
	saetze := make([]string, 0, len(befunde))
	for _, befund := range befunde {
		saetze = append(saetze, befund.Meldung)
	}

	return saetze
}

// ortsregel liest die **wirksame** Richtlinie.
//
// **Über den Stand und nicht über den rohen Wert** ([service.Wahl]): die
// Faltung über Gerät, Nutzer und Profil macht `domain`, und was hier ankommt,
// ist der Wert, der **gilt**. „Restriktivste Regel gilt" ist damit
// beantwortet, bevor diese Naht läuft — das Issue verlangt genau dieses
// Konsumieren und keine zweite Rechnung daneben.
func (a *App) ortsregel(ctx context.Context) (service.KIOrtsregel, error) {
	if a.einstellungen == nil {
		return service.KIOrtsregel{}, fehlenderBestand()
	}

	stand, fehler := a.einstellungen.Stand(ctx)
	if fehler != nil {
		return service.KIOrtsregel{}, fehler
	}

	return service.KIOrtsregelAus(stand[domain.RichtlinieNurLokal].Wirksam), nil
}

// zuKIModellDTOs macht aus den Modellen ihre DTOs.
func zuKIModellDTOs(modelle []domain.KIModell) []KIModellDTO {
	dtos := make([]KIModellDTO, 0, len(modelle))

	for _, modell := range modelle {
		dtos = append(dtos, KIModellDTO{
			Anbieter: modell.Anbieter, Alias: modell.Alias,
			Laufzeit: modell.Laufzeit, Ort: string(modell.Ort),
			Faehigkeiten: zuKIBelegen(modell.Geprueft),
			Herkunft:     zuKIGliedern(modell.Herkunft.Glieder),
			Bindung:      modell.Bindung(),

			Kontextfenster: modell.Kontextfenster, KontextBeleg: modell.KontextBeleg,
		})
	}

	return dtos
}

// zuKIBelegen macht aus den geprüften Fähigkeiten ihre DTOs.
func zuKIBelegen(geprueft []domain.KIGeprueft) []KIBelegDTO {
	belege := make([]KIBelegDTO, 0, len(geprueft))

	for _, eine := range geprueft {
		belege = append(belege, KIBelegDTO{
			Faehigkeit: string(eine.Faehigkeit),
			Datum:      eine.Datum, Beleg: eine.Beleg,
		})
	}

	return belege
}

// zuKIGliedern macht aus der Kette ihre DTOs.
func zuKIGliedern(glieder []domain.KIGlied) []KIGliedDTO {
	dtos := make([]KIGliedDTO, 0, len(glieder))

	for _, glied := range glieder {
		dtos = append(dtos, KIGliedDTO{
			Art: string(glied.Art), Kennung: glied.Kennung, Aus: glied.Aus,
			Quelle: glied.Quelle, Revision: glied.Revision, Lizenz: glied.Lizenz,
		})
	}

	return dtos
}

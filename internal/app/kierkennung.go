package app

// **Ein Modell liest ein Kartenbild** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Was hier geschieht und was nicht
//
// **Hier wird gefragt und umgerechnet, nicht entschieden.** Das Modell
// bekommt das Bild und eine Anweisung, die ausdrücklich nach **gelesenen
// Stellen** fragt; was daraus wird, entscheidet
// [service.Bilderkenner.Deuten] mit derselben Suche, die eine getippte
// Kartenliste benutzt.
//
// **Und geschrieben wird nichts.** Das Issue nennt es als eigene Abnahme:
// „Bild einer mehrdeutigen/falschen/keiner Karte erzeugt **keine** falsche
// Printing-Bindung oder Sammlungsmengen."

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// KIErkennungsanweisung steht vor jeder Frage nach einem Kartenbild.
//
// # Warum sie nach Stellen fragt und nicht nach der Karte
//
// **Weil die Karte nicht die Frage ist.** Ein Modell, das „das ist Lightning
// Bolt aus Alpha" antwortet, hat geraten — welcher Druck es ist, entscheidet
// die Suche aus den gelesenen Stellen. Wer nach der Antwort fragt, bekommt
// eine Antwort; wer nach dem Gelesenen fragt, bekommt etwas Prüfbares.
//
// # Und warum „nicht lesbar" ausdrücklich erlaubt ist
//
// **Weil ein Modell sonst rät.** Eine abgeschnittene Fußzeile hat keine
// Sammlernummer, und ein Modell, dem nur „nenne die Nummer" gesagt wurde,
// erfindet eine. `lesbar: false` ist der Ausweg, den es braucht — und der
// Befund [domain.BefundWertOhneLesbarkeit] hält, dass es ihn nicht mit einem
// Wert kombiniert.
const KIErkennungsanweisung = "Du liest eine Sammelkarte auf einem Foto. " +
	"Nenne NUR, was du WIRKLICH LESEN oder SEHEN kannst — je Stelle einen " +
	"Eintrag. Rate NICHT, welche Karte es ist: das entscheidet eine Suche " +
	"aus dem, was du liest. " +
	"Ist eine Stelle abgeschnitten, verdeckt oder unscharf, setze " +
	"»lesbar« auf false und lass »wert« LEER — ein geratener Wert ist " +
	"schlimmer als keiner. " +
	"Gib den Namen genau so wieder, wie er dasteht. " +
	"Für »kunstwerk« beschreibe, was abgebildet ist; das ist die einzige " +
	"Stelle, an der eine Beschreibung statt eines Zitats gehört."

// KIErkennungsbezug ist, was für eine Erkennung gebraucht wird.
type KIErkennungsbezug struct {
	// Pfad ist die Bilddatei auf diesem Rechner.
	//
	// **Ein Pfad und keine Adresse**: was dieses Programm liest, liegt hier
	// ([0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md)).
	// Eine Adresse zu holen, um sie einem Modell zu geben, wäre ein Holweg,
	// den niemand ausgelöst hat.
	Pfad string

	Anbieter string
	Alias    string

	// Nachweis ist die Freigabe eines Menschen — leer heißt: keine.
	Nachweis string
}

// KIBeobachtungDTO ist eine gelesene Stelle.
type KIBeobachtungDTO struct {
	Art    string `json:"art"`
	Wert   string `json:"wert"`
	Lesbar bool   `json:"lesbar"`
}

// KIBildkandidatDTO ist ein Druck, der zu den Beobachtungen passt.
type KIBildkandidatDTO struct {
	Kennung    string `json:"kennung"`
	Name       string `json:"name"`
	Setcode    string `json:"setcode"`
	Setname    string `json:"setname"`
	Nummer     string `json:"nummer"`
	Seltenheit string `json:"seltenheit"`
	Bild       string `json:"bild"`

	// Gestuetzt sind die Evidenzarten, die **diesen** Kandidaten stützen.
	Gestuetzt []string `json:"gestuetzt"`
}

// KIErkennungAntwort ist, was aus dem Bild wurde.
type KIErkennungAntwort struct {
	Beobachtungen []KIBeobachtungDTO `json:"beobachtungen"`

	// NichtGelesen sind die Stellen, zu denen nichts zu sagen war.
	NichtGelesen []string `json:"nichtGelesen"`

	Befunde []EinschaetzungsbefundDTO `json:"befunde"`

	Kandidaten []KIBildkandidatDTO `json:"kandidaten"`

	// Drucke ist die Zahl der passenden Drucke — **die Unsicherheit**.
	Drucke int `json:"drucke"`

	// Weitere sagt, dass es hinter der Seite weitergeht.
	Weitere bool `json:"weitere"`

	// Eindeutig heißt: genau ein Druck passt, und dahinter liegt nichts.
	//
	// **Es heißt nicht „übernimm ihn".** Angelegt wird nichts; die Auswahl
	// trifft ein Mensch.
	Eindeutig bool `json:"eindeutig"`

	Alter    string `json:"alter"`
	Anbieter string `json:"anbieter"`
	Modell   string `json:"modell"`

	// Abgebrochen sagt, dass niemand eine Datei gewählt hat.
	//
	// **Ein eigener Zustand neben der leeren Antwort** — dieselbe
	// Unterscheidung wie bei [OrdnerAntwort]: „ich habe nichts gelesen" und
	// „es wurde gar nicht gefragt" sind zwei Auskünfte, und nur die erste
	// sagt etwas über das Bild. Ohne ihn zeigte das Fenster nach einem
	// geschlossenen Dialog „kein Druck passt".
	//
	// **Nur am Fenster.** An MCP, REST und der Kommandozeile steht der Pfad
	// im Aufruf; dort gibt es keinen Dialog und damit keinen Abbruch.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// leereErkennung gibt eine Antwort mit Listen statt `null`.
func leereErkennung() KIErkennungAntwort {
	return KIErkennungAntwort{
		Beobachtungen: []KIBeobachtungDTO{},
		NichtGelesen:  []string{},
		Befunde:       []EinschaetzungsbefundDTO{},
		Kandidaten:    []KIBildkandidatDTO{},
	}
}

// erkennungsfehler setzt den Fehler an eine leere Antwort.
func erkennungsfehler(ctx context.Context, fehler error) KIErkennungAntwort {
	antwort := leereErkennung()
	antwort.Fehler = Uebersetze(ctx, fehler)

	return antwort
}

// KIKarteErkennen liest ein Kartenbild und nennt die passenden Drucke.
//
// **Es schreibt nichts** — weder eine Karte noch eine Sammlungsmenge.
func (a *App) KIKarteErkennen(
	ctx context.Context, bezug KIErkennungsbezug,
) KIErkennungAntwort {
	ctx = a.imProfil(ctx)

	// **Die Rolle vor dem Modell** — dieselbe Frage wie bei
	// [App.KIVorschlagen] (ID-03, #1702).
	if verweigert := a.darf(ctx, domain.WirkungSchreiben); verweigert != nil {
		return erkennungsfehler(ctx, verweigert)
	}

	if a.bilderkenner == nil {
		return erkennungsfehler(ctx, fehlenderBestand())
	}

	bild, medientyp, fehler := sichtbildLesen(bezug.Pfad)
	if fehler != nil {
		return erkennungsfehler(ctx, fehler)
	}

	// **Das Profil vor dem Buchen** — ohne eines lässt sich nicht sagen,
	// wessen Geld gemeint ist.
	profil, fehler := a.profilDiesesRufs(ctx)
	if fehler != nil {
		return erkennungsfehler(ctx, fehler)
	}

	// **Gebucht wird vor dem Werkzeugsatz** — dieselbe Reihenfolge wie beim
	// Batch ([App.KIBatchVorschlagen]), und aus demselben Grund: [kiwerkzeug]
	// baut das Budget selbst. Wer danach noch einmal buchte, fragte ein
	// zweites Mal dasselbe — und der Fehlerweg der zweiten Frage wäre ein
	// Zweig, den kein Aufruf betreten kann (§3).
	posten, buchfehler := a.erkennungBuchen(ctx, bezug)
	if buchfehler != nil {
		return erkennungsfehler(ctx, buchfehler)
	}

	werk, quelle, fehler := a.kiwerkzeug(ctx, bezug.Anbieter,
		kideckung{posten: posten})
	if fehler != nil {
		// **Freigeben und nicht als unklar merken** (#1542): gebucht ist,
		// **hinausgegangen** ist nichts — das Werk steht ja nicht.
		a.kiPostenFreigeben(ctx, posten)

		return erkennungsfehler(ctx, fehler)
	}

	sicht, fehler := sichtquelleAus(quelle, bezug.Anbieter)
	if fehler != nil {
		a.kiPostenFreigeben(ctx, posten)

		return erkennungsfehler(ctx, fehler)
	}

	defer a.kiPostenSchliessen(ctx, posten)

	return a.erkennenMit(ctx, erkennungslauf{
		profil: profil, bezug: bezug, bild: bild, medientyp: medientyp,
	}, werk, sicht)
}

// alsErkennungsantwort übersetzt die Deutung für die Flächen.
func alsErkennungsantwort(
	deutung service.Bilddeutung, anbieter, modell string,
) KIErkennungAntwort {
	antwort := leereErkennung()

	for _, beobachtung := range deutung.Erkennung.Beobachtungen {
		antwort.Beobachtungen = append(antwort.Beobachtungen, KIBeobachtungDTO{
			Art: string(beobachtung.Art), Wert: beobachtung.Wert,
			Lesbar: beobachtung.Lesbar,
		})
	}

	for _, art := range deutung.Erkennung.NichtGelesen {
		antwort.NichtGelesen = append(antwort.NichtGelesen, string(art))
	}

	antwort.Befunde = alsBefundDTOs(deutung.Befunde)

	for _, kandidat := range deutung.Kandidaten {
		gestuetzt := make([]string, 0, len(kandidat.Gestuetzt))
		for _, art := range kandidat.Gestuetzt {
			gestuetzt = append(gestuetzt, string(art))
		}

		antwort.Kandidaten = append(antwort.Kandidaten, KIBildkandidatDTO{
			Kennung: kandidat.Treffer.Kennung, Name: kandidat.Treffer.Name,
			Setcode: kandidat.Treffer.Setcode, Setname: kandidat.Treffer.Setname,
			Nummer:     kandidat.Treffer.Nummer,
			Seltenheit: kandidat.Treffer.Seltenheit,
			Bild:       kandidat.Treffer.Bild,
			Gestuetzt:  gestuetzt,
		})
	}

	antwort.Drucke = deutung.Drucke
	antwort.Weitere = deutung.Weitere
	antwort.Eindeutig = deutung.Eindeutig()
	antwort.Alter = deutung.Alter
	antwort.Anbieter = anbieter
	antwort.Modell = modell

	return antwort
}

// KIKarteErkennen ist die Fensterbindung.
//
// # Warum der Dialog hier steht und nicht in [App.KIKarteErkennen]
//
// **Weil es ihn ohne Fenster nicht gibt.** §16 verlangt dieselbe Fähigkeit an
// jeder Fläche und nennt den Dateiwahldialog als die benannte Ausnahme; an
// MCP, REST und der Kommandozeile steht der Pfad im Aufruf. Stünde der Dialog
// in `App`, öffnete ein MCP-Werkzeug mit leerem Pfad ein Fenster auf einem
// Rechner, an dem niemand sitzt.
//
// **Dieselbe Gestalt wie [Bindungen.Bildordner]** — die Bindung ist die
// Fensterseite, und ein Dialog ist Fensterseite.
func (b *Bindungen) KIKarteErkennen(ctx context.Context, bezug KIErkennungsbezug) KIErkennungAntwort {
	ctx = b.ruf(ctx)

	if strings.TrimSpace(bezug.Pfad) == "" {
		texte := DialogtexteFuer(b.app(ctx).sprache)

		pfad, fehler := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
			Titel:  texte.BildOeffnen,
			Filter: texte.BildFilter,
			Muster: "*.png;*.jpg;*.jpeg;*.gif",
		})
		if fehler != nil {
			return erkennungsfehler(ctx, fehler)
		}

		// **Ein Abbruch ist kein Fehler.** Wer den Dialog schließt, bekommt
		// eine leere Antwort — und keine rote Meldung über etwas, das er
		// selbst entschieden hat.
		if pfad == "" {
			antwort := leereErkennung()
			antwort.Abgebrochen = true

			return antwort
		}

		bezug.Pfad = pfad
	}

	return b.app(ctx).KIKarteErkennen(ctx, bezug)
}

// Bilddeuter macht aus gelesenen Stellen Kandidaten.
//
// **Eine Schnittstelle und kein fester Typ** — dieselbe Richtung wie überall
// in `internal/app`: hier steht, **was** gebraucht wird, und `internal/service`
// erfüllt es.
type Bilddeuter interface {
	Deuten(ctx context.Context, erkennung domain.Bilderkennung) (
		service.Bilddeutung, error)
}

// MitBilderkenner hängt den Deuter an die Anwendung.
func (a *App) MitBilderkenner(deuter Bilddeuter) *App {
	a.bilderkenner = deuter

	return a
}

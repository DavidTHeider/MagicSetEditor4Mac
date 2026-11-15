package app

// **Der Bildvorschlag am Fenster** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Die Reihenfolge, und warum sie so ist
//
// Das Issue verlangt: „Karte->editierbarer Bildprompt->separat freigegebener
// Bildauftrag; Kosten-/Datenfreigabe für Text- und Bildschritt getrennt."
//
// **Getrennt ist es, weil es zwei Rufe sind.** Ein Nachweis gilt **einem**
// Auftrag und ist danach eingelöst ([domain.Nachweis]); wer den Text
// freigegeben hat, hat damit kein Bild freigegeben. Das ist keine neue Regel
// dieses Werkstücks, sondern die aus A8 — hier wird sie nur ein zweites Mal
// gebraucht.
//
// # Und das Bild landet nicht in der Karte
//
// **Ein Vorschlag ist noch keine Änderung.** Was hier entsteht, liegt im
// [service.Bildvorrat]: mit Frist, mit Profilbindung, mit einer Kennung, die
// niemand errät. In die Karte kommt es erst über [App.KIBildEinsetzen], und
// der geht durch **denselben** Zuschnitt wie eine Datei von der Platte.

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// KIBildvorschlagDTO ist ein wartender Bildvorschlag.
type KIBildvorschlagDTO struct {
	// Kennung ist die gewürfelte Kennung im Vorrat.
	Kennung string `json:"kennung"`

	Anbieter string `json:"anbieter"`
	Modell   string `json:"modell"`

	// Umschrift ist der Prompt, wie der Dienst ihn umgeschrieben hat — leer,
	// wenn er nichts sagt.
	Umschrift string `json:"umschrift"`

	Breite int `json:"breite"`
	Hoehe  int `json:"hoehe"`
	Bytes  int `json:"bytes"`

	// Vorschau ist das Bild als Daten-URI.
	//
	// # Warum das ganze Bild und keine Verkleinerung
	//
	// **Weil der Vorrat es schon begrenzt.** Was hier ankommt, hat die
	// Bytegrenze und die Pixelgrenze passiert
	// ([service.UeblicheBildgrenzen]); eine zweite Verkleinerung wäre eine
	// zweite Antwort auf dieselbe Frage — und die Vorschau zeigte dann etwas
	// anderes als das, was übernommen wird.
	Vorschau string `json:"vorschau"`

	// Ablauf sagt, bis wann der Vorschlag gilt (RFC 3339).
	Ablauf string `json:"ablauf"`

	// Angaben sind die Parameter, die **wirklich feststehen**. Was der
	// Dienst nicht nennt, steht hier nicht — erfunden wird nichts.
	Angaben []KIBildangabeDTO `json:"angaben"`
}

// KIBildangabeDTO ist eine Angabe über die Entstehung.
type KIBildangabeDTO struct {
	Name string `json:"name"`
	Wert string `json:"wert"`
}

// KIBildvorschlagAntwort ist die Antwort auf einen Bildauftrag.
type KIBildvorschlagAntwort struct {
	KIBildvorschlagDTO

	Fehler *Fehler `json:"fehler,omitempty"`
}

// KIBildherkunftAntwort sind die Angaben zu einem übernommenen Bild.
type KIBildherkunftAntwort struct {
	Angaben []KIBildangabeDTO `json:"angaben"`

	Fehler *Fehler `json:"fehler,omitempty"`
}

// MitBildvorrat setzt den Vorrat für Bildvorschläge — **für Läufe**.
//
// **Im Betrieb entsteht er in der Composition Root** (`ablagennaht.go`), und
// zwar über dem Zwischenspeicher des Benutzers. Ein Lauf hat keinen; ohne
// diese Naht bräuchte er ein echtes Zuhause (§3).
func (a *App) MitBildvorrat(vorrat *service.Bildvorrat) *App {
	a.bildvorrat = vorrat

	return a
}

// KIBildVorschlagen lässt ein Bild erzeugen und legt es in den Vorrat.
//
// **Es schreibt nichts an der Karte.** Der Rückgabewert ist ein Vorschlag mit
// Frist; wer ihn will, ruft [App.KIBildEinsetzen], wer ihn nicht will,
// [App.KIBildVerwerfen] — oder nichts, und dann läuft er ab.
func (a *App) KIBildVorschlagen(
	ctx context.Context, karteID int64, feld, anbieter, alias, aufgabe,
	groesse, guete string, bearbeiten bool, nachweis string,
) KIBildvorschlagAntwort {
	ctx = a.imProfil(ctx)

	// **Die Rolle vor dem Modell** — dieselbe Frage wie bei
	// [App.KIVorschlagen] (ID-03, #1702).
	if verweigert := a.darf(ctx, domain.WirkungSchreiben); verweigert != nil {
		return KIBildvorschlagAntwort{Fehler: Uebersetze(ctx, verweigert)}
	}

	// **Ab hier ist der Auftrag abbrechbar** (A17, #1532). Angemeldet wird
	// **vor** der ersten Prüfung: ein Auftrag, der erst nach der Buchung
	// abbrechbar würde, ließe genau das Fenster offen, in dem Geld fließt.
	abbrechbar, fertig, laufend := a.bildlaufAnmelden(ctx, karteID)
	if laufend != nil {
		// **Der alte Kontext und nicht der neue**: bei einem Konflikt gibt
		// es keinen neuen — `anmelden` gibt dann `nil` zurück, und wer ihn
		// trotzdem zuwiese, übersetzte die Meldung gegen nichts.
		return KIBildvorschlagAntwort{Fehler: Uebersetze(ctx, laufend)}
	}

	ctx = abbrechbar

	defer fertig()

	profil, fehler := storage.ProfilAus(ctx)
	if fehler != nil {
		return KIBildvorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	auftrag := domain.KIBildauftrag{
		Aufgabe: aufgabe, Feld: feld, Groesse: groesse, Guete: guete,
		Bearbeiten: bearbeiten,
	}

	// **Der Auftrag wird geprüft, bevor irgendetwas Geld kostet.**
	if unvollstaendig := auftrag.Pruefe(); unvollstaendig != nil {
		return KIBildvorschlagAntwort{Fehler: Uebersetze(ctx, unvollstaendig)}
	}

	if a.bildvorrat == nil {
		return KIBildvorschlagAntwort{Fehler: Uebersetze(ctx, fehlenderVorrat())}
	}

	werk, quelle, fehler := a.kiwerkzeug(ctx, anbieter,
		kideckung{nachweis: nachweis})
	if fehler != nil {
		return KIBildvorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	bildquelle, fehler := bildquelleAus(quelle, anbieter)
	if fehler != nil {
		return KIBildvorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	vorlage, fehler := a.bildvorlage(ctx, karteID, auftrag)
	if fehler != nil {
		return KIBildvorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Erst jetzt wird gebucht.** `Anlegen` reserviert im Kostenjournal und
	// verlangt den Nachweis eines Menschen — vor dem Ruf und nicht danach
	// (A8, #1447).
	angelegt, fehler := werk.Anlegen(ctx, service.KIWunsch{
		ProfilID: profil, Faehigkeit: bildfaehigkeit(auftrag),
		Anbieter: anbieter, Alias: alias, Zielkarte: karteID,
	})
	if fehler != nil {
		return KIBildvorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Der Posten wird geschlossen, gleich wie es ausgeht** (#1542) —
	// dieselbe Zusage wie beim Textvorschlag, und aus demselben Grund.
	defer a.kiPostenSchliessen(ctx, angelegt.Kostenposten)

	roh, fehler := bildquelle.Bild(ctx, KIBildwunsch{
		Alias: alias, Auftrag: aufgabe, Groesse: groesse, Guete: guete,
		Quelle: vorlage,
	})
	if fehler != nil {
		return KIBildvorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return a.inDenVorrat(ctx, profil, karteID, angelegt, auftrag, anbieter, roh)
}

// inDenVorrat legt das fertige Bild ab und baut die Antwort.
func (a *App) inDenVorrat(
	ctx context.Context, profil, karteID int64, angelegt domain.KIAuftrag,
	auftrag domain.KIBildauftrag, anbieter string, roh KIRohbild,
) KIBildvorschlagAntwort {
	herkunft := domain.KIBildherkunft{
		Anbieter: anbieter, Modell: roh.Modell, Aufgabe: auftrag.Aufgabe,
		Umschrift: roh.Umschrift, Auftrag: angelegt.Kennung,
		Zeitpunkt: a.jetzt(), Parameter: bildangabenAus(auftrag, roh),
	}

	eintrag, fehler := a.bildvorrat.LegenMitHerkunft(service.Bildbesitz{
		ProfilID: profil, Karte: karteID, Feld: auftrag.Feld,
		Auftrag: angelegt.Kennung,

		// **Der eingefrorene Stand geht mit in den Vorrat** (Befund 2 des
		// Nachlaufs): `Anlegen` hat ihn gerade gelesen, und eingesetzt wird
		// er Minuten später.
		Erwartet: angelegt.Zielrevision,
	}, herkunft, roh.Inhalt)
	if fehler != nil {
		return KIBildvorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return KIBildvorschlagAntwort{KIBildvorschlagDTO: KIBildvorschlagDTO{
		Kennung: eintrag.Kennung, Anbieter: anbieter, Modell: roh.Modell,
		Umschrift: roh.Umschrift, Breite: eintrag.Breite, Hoehe: eintrag.Hoehe,
		Bytes:    int(eintrag.Bytes),
		Vorschau: alsDatenURI(eintrag.Medientyp, roh.Inhalt),
		Ablauf:   eintrag.Verfaellt.Format(time.RFC3339),
		Angaben:  alsAngabenDTO(herkunft.Parameter),
	}}
}

// KIBildVerwerfen nimmt einen wartenden Vorschlag weg.
func (a *App) KIBildVerwerfen(ctx context.Context, kennung string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	profil, fehler := storage.ProfilAus(ctx)
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if a.bildvorrat == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehlenderVorrat())}
	}

	if fehler := a.bildvorrat.Verwerfen(kennung, profil); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SchreibAntwort{}
}

// KIBildHerkunft gibt die Angaben zu einem übernommenen Bild.
func (a *App) KIBildHerkunft(
	ctx context.Context, karteID int64, feld string,
) KIBildherkunftAntwort {
	ctx = a.imProfil(ctx)

	leser := a.herkunftsleser()
	if leser == nil {
		return KIBildherkunftAntwort{
			Angaben: []KIBildangabeDTO{},
			Fehler:  Uebersetze(ctx, fehlenderBestand()),
		}
	}

	angaben, fehler := leser.Bildangaben(ctx, karteID, feld)
	if fehler != nil {
		return KIBildherkunftAntwort{
			Angaben: []KIBildangabeDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return KIBildherkunftAntwort{Angaben: alsAngabenDTO(angaben)}
}

// bildfaehigkeit sagt, welche Fähigkeit dieser Auftrag verlangt.
func bildfaehigkeit(auftrag domain.KIBildauftrag) domain.KIFaehigkeit {
	if auftrag.Bearbeiten {
		return domain.KIBildbearbeitung
	}

	return domain.KIBildgenerierung
}

// bildangabenAus sammelt, was **wirklich feststeht**.
//
// **Was nicht gesagt wurde, steht nicht da.** Die Ergänzung des Issues:
// „Fehlende Angaben als unbekannt führen." Eine Angabe `guete: ""` behauptete,
// es gäbe eine und sie sei leer.
func bildangabenAus(auftrag domain.KIBildauftrag, roh KIRohbild) []domain.KIBildparameter {
	angaben := make([]domain.KIBildparameter, 0, 4)

	for _, paar := range []domain.KIBildparameter{
		{Name: "groesse", Wert: auftrag.Groesse},
		{Name: "guete", Wert: auftrag.Guete},
		{Name: "umschrift", Wert: roh.Umschrift},
		{Name: "medientyp", Wert: roh.Medientyp},
	} {
		if paar.Wert != "" {
			angaben = append(angaben, paar)
		}
	}

	return angaben
}

// alsAngabenDTO bringt die Angaben auf die Gestalt des Fensters.
func alsAngabenDTO(angaben []domain.KIBildparameter) []KIBildangabeDTO {
	heraus := make([]KIBildangabeDTO, 0, len(angaben))
	for _, angabe := range angaben {
		heraus = append(heraus, KIBildangabeDTO{Name: angabe.Name, Wert: angabe.Wert})
	}

	return heraus
}

// alsDatenURI packt Bytes so, dass ein `<img>` sie zeigen kann.
func alsDatenURI(medientyp string, inhalt []byte) string {
	return "data:" + medientyp + ";base64," +
		base64.StdEncoding.EncodeToString(inhalt)
}

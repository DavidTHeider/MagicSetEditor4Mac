package service

// **Die Verwaltung des Retrievalindex** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Wofür es sie gibt
//
// **Damit vier Flächen dieselbe Antwort geben.** Das Issue verlangt einen
// „begrenzten tatsächlichen Diagnose-/Verwaltungsaufruf für Indexstatus,
// Aufbau/Abbruch und Suche" — und §16 verlangt, dass er an allen Flächen
// derselbe ist. Vier Stellen, die sich den Einbetter selbst zusammensuchen,
// wären vier Gelegenheiten, ihn anders zusammenzusuchen.
//
// # Gesucht wird mit dem Modell, mit dem gebaut wurde
//
// **Und nicht mit dem, das gerade eingerichtet ist.** Der aktive Index nennt
// sein Modell im Manifest; wer mit einem anderen suchte, suchte in einem
// fremden Raum. Ist das Modell nicht mehr da, bleibt die Suche lexikalisch
// und sagt es — genau der Fall, den das Issue „Modellwechsel" nennt.

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Einbetterquelle liefert einen Einbetter zu einem Modellnamen.
//
// **Sie steht hier und wird in `app` erfüllt** (§2): welche Laufzeit gemeint
// ist, entscheidet die Einrichtung und nicht der Fachkern — das Issue sagt
// „keine harte Ollama-Abhängigkeit im Retrievalvertrag".
type Einbetterquelle interface {
	Einbetter(ctx context.Context, modell string) (Einbetter, []domain.Befund, error)
}

// Korpusoeffner öffnet eine Korpusquelle zu einer Angabe.
//
// **Was die Angabe bedeutet, weiß der Öffner** — ein Pfad, ein Bezeichner,
// ein Verzeichnis. Der Fachkern liest keine Dateien.
type Korpusoeffner interface {
	Oeffnen(ctx context.Context, auftrag Aufbauauftrag) (
		quelle Korpusquelle, schliessen func(), fehler error)
}

// Aufbauauftrag ist, was ein Aufbau an Angaben braucht.
type Aufbauauftrag struct {
	// Modell ist der Name des Einbettungsmodells.
	Modell string

	// Quelle nennt, woraus gebaut wird — für den [Korpusoeffner].
	Quelle string

	// Revision und Datum sind der Stand der Quelle; Abgerufen ihre Abrufzeit.
	//
	// **Drei Angaben und nicht eine** (Issue): „Abrufdatum ist nicht
	// Gültigkeitsdatum."
	Revision  string
	Datum     string
	Abgerufen time.Time

	// Fortschreiben baut auf der aktiven Generation auf.
	//
	// **Dann liefert die Quelle nur, was sich geändert hat.** Was sie liefert
	// und danach nicht geschrieben wird, ist entfernt; alles übrige wird
	// **kopiert** statt neu eingebettet.
	Fortschreiben bool
}

// Indexstand ist die Auskunft über den Retrievalindex.
type Indexstand struct {
	// Aktiv sagt, ob überhaupt ein Index dasteht.
	Aktiv bool

	Manifest domain.Indexmanifest

	// Aufbau ist der Stand des letzten oder laufenden Aufbaus.
	Aufbau Aufbaustand

	// Suchbar sagt, ob semantisch gesucht werden kann, Grund warum nicht.
	Suchbar bool
	Grund   string
}

// Retrievalverwaltung führt Index, Aufbau und Suche zusammen.
type Retrievalverwaltung struct {
	ablage    Retrievalablage
	aufbau    *Indexaufbau
	einbetter Einbetterquelle
	korpus    Korpusoeffner
	haushalt  *Haushalt

	mu     sync.Mutex
	raeume map[string]Einbetter
}

// NeueRetrievalverwaltung baut sie.
func NeueRetrievalverwaltung(
	ablage Retrievalablage, schreiber Retrievalschreiber,
	einbetter Einbetterquelle, korpus Korpusoeffner, haushalt *Haushalt,
) *Retrievalverwaltung {
	return &Retrievalverwaltung{
		ablage:    ablage,
		aufbau:    NeuerIndexaufbau(schreiber, haushalt),
		einbetter: einbetter,
		korpus:    korpus,
		haushalt:  haushalt,
		raeume:    map[string]Einbetter{},
	}
}

// Stand gibt die Auskunft über den Index.
func (v *Retrievalverwaltung) Stand(ctx context.Context) Indexstand {
	stand := Indexstand{Aufbau: v.aufbau.Stand()}

	manifest, fehler := v.ablage.Manifest(ctx)
	if fehler != nil {
		stand.Grund = "Es ist kein Suchindex aktiv."
		if !errors.Is(fehler, domain.ErrNotFound) {
			stand.Grund = "Der Indexstand ist nicht lesbar. " +
				domain.SatzVon(fehler)
		}

		return stand
	}

	stand.Aktiv = true
	stand.Manifest = manifest

	if _, grund := v.einbetterFuer(ctx, manifest.Modell); grund != "" {
		stand.Grund = grund

		return stand
	}

	stand.Suchbar = true

	return stand
}

// Suchen beantwortet eine Frage gegen den aktiven Index.
//
// **Der Platz kommt aus der Klasse `suche`** und nicht aus `indexbau`: eine
// Suche darf nicht warten, weil jemand neu baut (Issue, Werkstück 4).
func (v *Retrievalverwaltung) Suchen(ctx context.Context, wunsch Suchwunsch) (
	Suchergebnis, error,
) {
	freigeben, fehler := v.haushalt.Belegen(ctx, domain.RessourceSuche)
	if fehler != nil {
		return Suchergebnis{}, fehler
	}

	defer freigeben()

	manifest, fehler := v.ablage.Manifest(ctx)
	if fehler != nil && !errors.Is(fehler, domain.ErrNotFound) {
		return Suchergebnis{}, fehler
	}

	einbetter, grund := v.einbetterFuer(ctx, manifest.Modell)

	// **Der genauere Grund kommt von hier.** Der Dienst sieht nur, dass er
	// keinen Einbetter hat; **warum** nicht, weiß die Verwaltung — und
	// „das Modell dieses Index ist weg" sagt jemandem, was zu tun ist.
	return NeuerRetrievaldienst(v.ablage, einbetter).
		MitAusfallgrund(grund).Suchen(ctx, wunsch)
}

// Aufbauen baut eine neue Generation.
func (v *Retrievalverwaltung) Aufbauen(
	ctx context.Context, auftrag Aufbauauftrag,
) (Aufbaubericht, error) {
	return v.aufbau.Aufbauen(ctx, Aufbauwunsch{
		Schemafassung:  domain.RetrievalschemaFassung,
		Zerteilfassung: Zerteilfassung,
		Indexfassung:   domain.RetrievalindexFassung,
		Fortschreiben:  auftrag.Fortschreiben,
	}, v.beschaffen(auftrag))
}

// beschaffen gibt den Bauplan für Einbetter und Korpus.
//
// **Er wird erst hinter dem Riegel ausgeführt** ([Aufbaustoff]): ein Modell zu
// messen und eine Sammeldatei zu öffnen kostet Zeit, und ein zweiter Prozess
// soll sie nicht aufwenden, um danach abgewiesen zu werden.
func (v *Retrievalverwaltung) beschaffen(auftrag Aufbauauftrag) Aufbaustoff {
	return func(ctx context.Context) (Einbetter, Korpusquelle, func(), error) {
		nichts := func() {}

		einbetter, befunde, fehler := v.einbetter.Einbetter(ctx, auftrag.Modell)
		if fehler != nil {
			return nil, nil, nichts, fehler
		}

		if len(befunde) > 0 {
			return nil, nil, nichts, domain.MitGrund(domain.ErrValidation, nil,
				"%q taugt hier nicht als Einbettungsmodell: %s",
				auftrag.Modell, befunde[0].Meldung)
		}

		if unpassend := v.pruefeFortschreibung(ctx, einbetter, auftrag); unpassend != nil {
			return nil, nil, nichts, unpassend
		}

		quelle, schliessen, fehler := v.korpus.Oeffnen(ctx, auftrag)
		if fehler != nil {
			return nil, nil, nichts, fehler
		}

		v.merken(auftrag.Modell, einbetter)

		return einbetter, quelle, schliessen, nil
	}
}

// pruefeFortschreibung weist eine Fortschreibung in einen fremden Raum ab.
//
// # Der Fall, für den es diese Zeilen gibt
//
// **Ein Deltaimport kopiert Vektoren.** Sie stammen aus dem Modell, das die
// aktive Generation gebaut hat; wer mit einem **anderen** Modell fortschreibt,
// legt zwei Vektorräume in eine Generation — und dann sind die Abstände
// zwischen alten und neuen Abschnitten Zahlen ohne Bedeutung.
//
// **Das ist derselbe Satz wie beim Suchen**, nur teurer: dort bleibt die
// Suche lexikalisch, hier entstünde ein Index, der aussieht wie einer.
func (v *Retrievalverwaltung) pruefeFortschreibung(
	ctx context.Context, einbetter Einbetter, auftrag Aufbauauftrag,
) error {
	if !auftrag.Fortschreiben {
		return nil
	}

	manifest, fehler := v.ablage.Manifest(ctx)
	if fehler != nil {
		if errors.Is(fehler, domain.ErrNotFound) {
			return domain.MitGrund(domain.ErrConflict, nil,
				"es gibt keinen aktiven Index, auf dem sich fortschreiben "+
					"ließe — der erste Aufbau ist ein vollständiger")
		}

		return fehler
	}

	if !einbetter.Raum().PasstZu(manifest.Raum()) {
		return domain.MitGrund(domain.ErrConflict, nil,
			"%q rechnet in einem anderen Vektorraum als der Index, den %q "+
				"gebaut hat — fortgeschrieben würden dabei Vektoren aus zwei "+
				"Modellen in eine Generation. Ein vollständiger Aufbau geht",
			auftrag.Modell, manifest.Modell)
	}

	return nil
}

// Abbrechen hält einen laufenden Aufbau an.
func (v *Retrievalverwaltung) Abbrechen() bool { return v.aufbau.Abbrechen() }

// MitProzesssperre reicht die Absprache über Prozessgrenzen durch.
func (v *Retrievalverwaltung) MitProzesssperre(
	sperre Prozesssperre,
) *Retrievalverwaltung {
	v.aufbau.MitProzesssperre(sperre)

	return v
}

// einbetterFuer gibt den Einbetter zu einem Modell — oder den Grund dagegen.
//
// **Gemerkt wird er, weil ihn zu bauen ein Lauf ist.** Ein Einbetter entsteht
// aus einer Messung am echten Modell; wer sie bei jeder Suche wiederholte,
// bezahlte je Frage eine Einbettung, die niemand liest.
func (v *Retrievalverwaltung) einbetterFuer(ctx context.Context, modell string) (
	einbetter Einbetter, grund string,
) {
	if modell == "" {
		return nil, "Es ist kein Suchindex aktiv."
	}

	v.mu.Lock()
	gemerkt, bekannt := v.raeume[modell]
	v.mu.Unlock()

	if bekannt {
		return gemerkt, ""
	}

	gebaut, befunde, fehler := v.einbetter.Einbetter(ctx, modell)
	if fehler != nil {
		return nil, "Das Einbettungsmodell " + modell +
			" ist nicht erreichbar. Gesucht wird im Volltext."
	}

	if len(befunde) > 0 {
		return nil, "Das Modell " + modell + ", mit dem dieser Index gebaut " +
			"wurde, taugt hier nicht mehr: " + befunde[0].Meldung
	}

	v.merken(modell, gebaut)

	return gebaut, ""
}

// merken legt einen Einbetter ab.
func (v *Retrievalverwaltung) merken(modell string, einbetter Einbetter) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.raeume[modell] = einbetter
}

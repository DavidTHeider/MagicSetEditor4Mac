package app

// **Kann die neueste Fassung diesen Bestand öffnen?** — Werkstück O6 von
// Phase 5.2 (#1430).
//
// # Die Reihenfolge ist die Aussage
//
//  1. Die neueste Freigabe holen.
//  2. Die **signierte** Prüfsummenliste holen und ihre Signatur prüfen.
//  3. Das Manifest holen und gegen diese Liste halten.
//  4. Es mit dem **echten** Ledger vergleichen und den Augenblick prüfen.
//
// **Kein Schritt darf vorgezogen werden.** Wer das Manifest vor der Signatur
// läse, hätte einen Weg, auf dem eine ungeprüfte Angabe schon etwas
// entschieden hat — derselbe Satz steht bei `Klient.Pruefsummen`.

import (
	"context"
	"runtime"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// HindernisDTO ist ein Grund, warum ein Wechsel nicht vorbereitet wird.
type HindernisDTO struct {
	Art string `json:"art"`

	// Stelle benennt es genauer — der Name einer Wanderung, die fremde
	// Plattform. Leer, wo es nichts zu benennen gibt.
	Stelle string `json:"stelle"`
}

// VertraeglichkeitAntwort sagt, ob und warum.
type VertraeglichkeitAntwort struct {
	// Fassung ist die der Freigabe — leer, wenn es keine neuere gibt.
	Fassung string `json:"fassung"`

	// Urteil ist `vertraeglich`, `unvertraeglich` oder `ungeklaert`.
	Urteil string `json:"urteil"`

	// Bereit sagt, ob der Wechsel vorbereitet werden darf.
	Bereit bool `json:"bereit"`

	Hindernisse []HindernisDTO `json:"hindernisse"`

	Fehler *Fehler `json:"fehler"`
}

// MitVorpruefer hängt die Prüfung ein.
func (a *App) MitVorpruefer(pruefer *service.Vorpruefer) { a.vorpruefer = pruefer }

// Vertraeglichkeit sagt, ob die neueste Freigabe diesen Bestand **öffnen
// kann** — und sonst nichts.
//
// **Nicht, ob der Wechsel jetzt vorbereitet werden darf**: das hängt an
// Platz, laufender Arbeit und einer Sicherung, die jemand macht und
// zurückliest, und dafür gibt es [App.WechselVorbereiten]. Die Trennung ist
// §16: diese Frage hat an jeder Fläche dieselbe Antwort, jene nicht.
//
// **Ohne Prüfer bleibt es ungeklärt**, und das ist keine Ausrede: eine
// Anwendung, die den Ledger nicht lesen kann, weiß es nicht — und „weiß es
// nicht" ist die Antwort, nicht „geht nicht" (Abnahme aus dem Issue).
func (a *App) Vertraeglichkeit(ctx context.Context) VertraeglichkeitAntwort {
	ctx = a.imProfil(ctx)

	if a.fassungen == nil || a.vorpruefer == nil {
		return ungeklaert("")
	}

	// **Ohne eingeschaltete Aktualisierung wird nicht gefragt** (§16, Regel 6
	// und die Vorgabe „aus"): dieselbe Bedingung wie bei
	// [App.NachNeuerFassungSehen], und aus demselben Grund — dieser Weg geht
	// **ins Netz**.
	//
	// **Gefunden hat es der Flächenlauf**, nicht das Lesen: er fragte gegen
	// GitHub, wo ein normaler Test gar nicht hindarf (§3).
	an, err := a.aktualisierungAn(ctx)
	if err != nil {
		return VertraeglichkeitAntwort{
			Hindernisse: []HindernisDTO{}, Fehler: Uebersetze(ctx, err),
		}
	}

	if !an {
		return VertraeglichkeitAntwort{Hindernisse: []HindernisDTO{}}
	}

	freigabe, err := a.fassungen.NeuesteFreigabe(ctx)
	if err != nil {
		return VertraeglichkeitAntwort{
			Hindernisse: []HindernisDTO{}, Fehler: Uebersetze(ctx, err),
		}
	}

	// **Keine neuere Fassung ist keine Frage.** Wer schon die neueste hat,
	// braucht kein Urteil über einen Wechsel, den es nicht gibt.
	//
	// **Eine unlesbare Marke ist ein Fehler und kein Nein**: sie sagt nicht
	// „nichts Neues", sondern „diese Freigabe ist nicht einzuordnen".
	neuer, err := domain.FassungIstNeuer(freigabe.Fassung, Fassung)
	if err != nil {
		return VertraeglichkeitAntwort{
			Fassung: freigabe.Fassung, Hindernisse: []HindernisDTO{},
			Fehler: Uebersetze(ctx, err),
		}
	}

	if !neuer {
		return VertraeglichkeitAntwort{
			Fassung: freigabe.Fassung, Hindernisse: []HindernisDTO{},
		}
	}

	summen, err := a.fassungen.Pruefsummen(ctx, freigabe)
	if err != nil {
		return VertraeglichkeitAntwort{
			Fassung: freigabe.Fassung, Hindernisse: []HindernisDTO{},
			Fehler: Uebersetze(ctx, err),
		}
	}

	manifest, err := a.fassungen.Manifest(ctx, freigabe, summen)
	if err != nil {
		return VertraeglichkeitAntwort{
			Fassung: freigabe.Fassung, Hindernisse: []HindernisDTO{},
			Fehler: Uebersetze(ctx, err),
		}
	}

	befund, err := a.vorpruefer.Urteil(ctx, manifest)
	if err != nil {
		return VertraeglichkeitAntwort{
			Fassung: freigabe.Fassung, Hindernisse: []HindernisDTO{},
			Fehler: Uebersetze(ctx, err),
		}
	}

	return alsAntwort(freigabe.Fassung, befund)
}

// aktualisierungAn sagt, ob nach Freigaben gefragt werden darf.
//
// **Leer heißt Vorgabe heißt aus** — derselbe Satz wie bei
// [App.NachNeuerFassungSehen]. Wer nie gewählt hat, hat „aus" gewählt, und
// ohne Einstellungsspeicher gilt dasselbe: eine Anwendung vor dem Start
// fragt nicht ins Netz.
func (a *App) aktualisierungAn(ctx context.Context) (bool, error) {
	if a.einstellungen == nil {
		return false, nil
	}

	werte, err := a.einstellungen.Alle(ctx)
	if err != nil {
		return false, err
	}

	wahl := werte[domain.EinstellungAktualisierung]

	return wahl != "" && wahl != domain.AktualisierungAus, nil
}

// ungeklaert gibt die Antwort, wenn niemand fragen kann.
func ungeklaert(fassung string) VertraeglichkeitAntwort {
	return VertraeglichkeitAntwort{
		Fassung: fassung,
		Urteil:  string(domain.UrteilUngeklaert),
		Hindernisse: []HindernisDTO{
			{Art: string(domain.HindernisKeinManifest)},
		},
	}
}

// alsAntwort übersetzt den Befund.
func alsAntwort(fassung string, befund domain.Aktualisierungsbefund) VertraeglichkeitAntwort {
	antwort := VertraeglichkeitAntwort{
		Fassung: fassung, Urteil: string(befund.Urteil), Bereit: befund.Bereit,
		Hindernisse: make([]HindernisDTO, 0, len(befund.Hindernisse)),
	}

	for _, hindernis := range befund.Hindernisse {
		antwort.Hindernisse = append(antwort.Hindernisse, HindernisDTO{
			Art: string(hindernis.Art), Stelle: hindernis.Stelle,
		})
	}

	return antwort
}

// VertraeglichkeitsprueferUeber baut den Prüfer für eine Fläche.
//
// **Die drei Fragen des Augenblicks kommen mit** — Platz, laufende Arbeit und
// die Sicherungsprobe. Ohne sie stünde je ein Hindernis da, und das wäre
// richtig, aber unbrauchbar.
// **Ohne Fehlerrückgabe, und das ist nachgelesen**: der einzige Weg, auf dem
// sie scheitern könnte, wäre ein `nil`-Bestand — und gerufen wird sie aus
// `scryfallStarten`, das erst läuft, wenn er offen ist. Ein Fehlerzweig
// daneben wäre einer, den kein Lauf betritt (§3).
func VertraeglichkeitsprueferUeber(
	anwendung *App, bestandspfad string, weg *service.Bestandsweg,
) *service.Vorpruefer {
	return service.NeuerVorpruefer(LedgerUeber(anwendung.db),
		runtime.GOOS+"/"+runtime.GOARCH).
		MitPlatz(PlatzNeben(bestandspfad)).
		MitVorgaengen(vorgangsnaht{anwendung: anwendung}).
		MitHaltern(bestandshalterAn(bestandspfad)).
		MitSicherungsprobe(SicherungsprobeUeber(weg))
}

// WechselVorbereiten prüft zusätzlich den Augenblick — Platz, laufende
// Arbeit, und eine Sicherung, die **zurückgelesen** wurde.
//
// # Warum nur am Fenster
//
// **Weil es eine Sicherung macht und den Bestand dafür anhält.** Eine
// Kommandozeile, die das auf Zuruf täte, nähme ihn jemandem weg, der gerade
// arbeitet. Was jede Fläche sagen kann, sagt [App.Vertraeglichkeit].
//
// **Und „bereit" heißt hier wirklich bereit**: die Abnahme aus dem Issue
// verlangt eine Sicherung, die *nachweislich* wiederherstellbar ist — nicht
// eine, die geschrieben wurde.
func (a *App) WechselVorbereiten(ctx context.Context) VertraeglichkeitAntwort {
	ctx = a.imProfil(ctx)

	if a.fassungen == nil || a.vorpruefer == nil {
		return ungeklaert("")
	}

	an, err := a.aktualisierungAn(ctx)
	if err != nil {
		return VertraeglichkeitAntwort{
			Hindernisse: []HindernisDTO{}, Fehler: Uebersetze(ctx, err),
		}
	}

	if !an {
		return VertraeglichkeitAntwort{Hindernisse: []HindernisDTO{}}
	}

	freigabe, err := a.fassungen.NeuesteFreigabe(ctx)
	if err != nil {
		return VertraeglichkeitAntwort{
			Hindernisse: []HindernisDTO{}, Fehler: Uebersetze(ctx, err),
		}
	}

	summen, err := a.fassungen.Pruefsummen(ctx, freigabe)
	if err != nil {
		return VertraeglichkeitAntwort{
			Fassung: freigabe.Fassung, Hindernisse: []HindernisDTO{},
			Fehler: Uebersetze(ctx, err),
		}
	}

	manifest, err := a.fassungen.Manifest(ctx, freigabe, summen)
	if err != nil {
		return VertraeglichkeitAntwort{
			Fassung: freigabe.Fassung, Hindernisse: []HindernisDTO{},
			Fehler: Uebersetze(ctx, err),
		}
	}

	befund, err := a.vorpruefer.Pruefe(ctx, manifest, freigabe.Abbild.Groesse)
	if err != nil {
		return VertraeglichkeitAntwort{
			Fassung: freigabe.Fassung, Hindernisse: []HindernisDTO{},
			Fehler: Uebersetze(ctx, err),
		}
	}

	return alsAntwort(freigabe.Fassung, befund)
}

// Vertraeglichkeit ist die Fensterbindung.
func (b *Bindungen) Vertraeglichkeit(ctx context.Context) VertraeglichkeitAntwort {
	return b.app(ctx).Vertraeglichkeit(b.ruf(ctx))
}

// WechselVorbereiten ist die Fensterbindung für die volle Vorprüfung.
func (b *Bindungen) WechselVorbereiten(ctx context.Context) VertraeglichkeitAntwort {
	return b.app(ctx).WechselVorbereiten(b.ruf(ctx))
}

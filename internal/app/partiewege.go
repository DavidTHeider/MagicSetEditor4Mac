package app

// **Die fünf Wege der Partie** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Die fünfte Abnahme des Werkstücks
//
// „Fähigkeiten Session erstellen/lesen/Command/pausieren/beenden sind über
// die vier bestehenden Flächen **fachlich gleich**, mit derselben Profil-
// und Rollenprüfung."
//
// **Deshalb stehen sie hier und nicht in den Bindungen**: was das Fenster
// ruft, ruft auch `mse`, `mse-mcp` und `mse-rest` — über dieselbe Anwendung
// und damit über dieselbe Prüfung (§16).
//
// # Welche Wirkung, und warum nicht `externEmpfangen`
//
// **`lokalSchreiben`, denn eine Partie bleibt auf diesem Rechner.** Sie
// schreibt in den Bestand dieses Profils, und es geht kein Byte hinaus. Wer
// **mit anderen** spielt, öffnet einen Port — und das ist der Horchposten aus
// N2 mit seiner eigenen Wirkung und seiner eigenen Zustimmung.

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// PartieAnlegen legt eine Partie an und gibt die Sicht des eigenen Sitzes.
func (a *App) PartieAnlegen(
	ctx context.Context, partie domain.Partie,
) PartieAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leerePartie(ctx, fehler)
	}

	// **Das gezeigte Profil kommt aus dem Kontext und nicht aus [App.profil]**
	// (G8, #1465)
	// — warum, steht bei [gezeigtesProfil].
	profilID, fehler := gezeigtesProfil(ctx)
	if fehler != nil {
		return leerePartie(ctx, fehler)
	}

	angelegt, fehler := a.partienaht().Anlegen(ctx, partie)
	if fehler != nil {
		return leerePartie(ctx, fehler)
	}

	// **Die Antwort ist eine Sitzsicht und nicht der Zustand.** Wer anlegt,
	// sitzt am Tisch — und sieht so viel wie jeder andere Sitz.
	//
	// **Welcher Sitz, entscheidet das gezeigte Profil**: wer einen Sitz an
	// seines gebunden hat, sitzt dort — sonst beim Wirt. Die Regel selbst
	// steht in [domain.Partie.SichtFuerProfil] und damit an **einer** Stelle
	// für alle vier Flächen (§2).
	return a.partieNachDemSchreiben(ctx, angelegt.ID, angelegt.SichtFuerProfil(profilID))
}

// Partie gibt eine Partie, wie `fuer` sie sehen darf.
func (a *App) Partie(ctx context.Context, kennung, fuer string) PartieAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return leerePartie(ctx, fehler)
	}

	// **Aus einem Stand** (#2481): Sicht, Erlaubnis und Handlungen derselben
	// Fassung.
	bild, fehler := a.partienaht().Tischbild(ctx, kennung, fuer)
	if fehler != nil {
		return leerePartie(ctx, fehler)
	}

	sicht, erlaubt, handlungen := bild.Sicht, bild.Erlaubt, bild.Handlungen
	if erlaubt == nil {
		erlaubt = []string{}
	}

	dto := partieAus(sicht)

	// **Zwei Quellen, eine Liste** — die Befunde des Tisches kommen aus der
	// Domäne (M14, CR 702: eine Schlüsselwortfähigkeit, die niemand
	// durchsetzt), die der Profile aus dem Bestand (G8). Beide sind
	// Auffälligkeiten und kein Fehler (§8), und beide gehören an dieselbe
	// Stelle im Fenster.
	return PartieAntwort{
		Partie: dto, Erlaubt: erlaubt, Handlungen: handlungenAus(handlungen),
		Befunde: append(a.profilbefunde(ctx, dto.Sitze), sicht.Befunde...),
	}
}

// PartieBefehl stellt einen Spielbefehl.
func (a *App) PartieBefehl(
	ctx context.Context, befehl domain.Spielbefehl,
) BefehlAntwort {
	ctx = a.imProfil(ctx)

	leer := BefehlAntwort{Ereignisse: []EreignisDTO{}}

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	// **Und der Rest des Tisches erfährt davon**
	// (#1627)
	// — die anderen fragen nicht, sie warten. **In der Reihe** (#2481): kein
	// Zug eines Gastes, der danach angenommen wird, geht vor diesem hinaus.
	//
	// **Ein Fehler dabei nimmt den Zug nicht zurück** (§7, Regel 3): der
	// Befehl hat gewirkt, und eine Zustellung, die nicht glückt, ist ein
	// Zustellproblem. Er wird gemeldet und nicht durchgereicht.
	funk := service.NeuerTischfunk(befehl.PartieID, a.partienaht(), &a.partie.runde)

	ausgang, fehler := a.partienaht().BefehlVerkuendet(ctx, befehl,
		func(ausgang service.Befehlsausgang) {
			if fehler := funk.Verkuende(ctx, befehl.Absender, ausgang); fehler != nil {
				slog.WarnContext(ctx, "ein Zug erreichte nicht jeden am Tisch",
					"partie", befehl.PartieID, "grund", fehler)
			}
		})
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	return BefehlAntwort{
		Fassung: strconv.FormatInt(ausgang.Fassung, 10),
		// **Der Absender bekommt sie als Antwort.** Die anderen bekommen
		// sie über den Funk darüber — jeder seine eigene Projektion.
		Ereignisse: ereignisseAus(ausgang.Ereignisse, befehl.Absender),
	}
}

// PartiePausieren hält eine Partie an.
func (a *App) PartiePausieren(
	ctx context.Context, kennung, fuer string,
) PartieAntwort {
	ctx = a.imProfil(ctx)

	return a.standwechsel(ctx, kennung, fuer, a.partienaht().Pausieren)
}

// PartieFortsetzen nimmt sie wieder auf.
func (a *App) PartieFortsetzen(
	ctx context.Context, kennung, fuer string,
) PartieAntwort {
	ctx = a.imProfil(ctx)

	return a.standwechsel(ctx, kennung, fuer, a.partienaht().Fortsetzen)
}

// PartieBeenden schließt sie ab.
func (a *App) PartieBeenden(
	ctx context.Context, kennung, fuer string,
) PartieAntwort {
	ctx = a.imProfil(ctx)

	return a.standwechsel(ctx, kennung, fuer, a.partienaht().Beenden)
}

// standwechsel ist der gemeinsame Weg der drei Standwege.
//
// **Einer und nicht drei**: sie unterscheiden sich in einer Zeile, und drei
// Kopien liefen bei der ersten Änderung auseinander.
//
// **Das Profil setzt der Rufer und nicht dieser Weg.** Er ist ein Helfer und
// keine Bindung; stünde die Zeile hier, hätten drei Bindungen sie nicht —
// und `internal/werkzeuge/profilnaht_test.go` prüft **jede** Bindung an
// ihrer ersten Zeile, weil ein Weg über einen Helfer sich nicht mechanisch
// nachverfolgen lässt.
func (a *App) standwechsel(
	ctx context.Context, kennung, fuer string,
	wechseln func(context.Context, string) error,
) PartieAntwort {
	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leerePartie(ctx, fehler)
	}

	if fehler := wechseln(ctx, kennung); fehler != nil {
		return leerePartie(ctx, fehler)
	}

	// **Und die Gäste erfahren es** (#2506) — ein Tischbild ohne Ereignis, in
	// der Reihe wie jede Verkündung (#2481). Ein Fehler dabei nimmt den
	// Wechsel nicht zurück (§7, Regel 3).
	funk := service.NeuerTischfunk(kennung, a.partienaht(), &a.partie.runde)
	if fehler := a.partienaht().InReihe(func() error {
		return funk.Verkuende(ctx, fuer, service.Befehlsausgang{})
	}); fehler != nil {
		slog.WarnContext(ctx, "ein Standwechsel erreichte nicht jeden am Tisch",
			"partie", kennung, "grund", fehler)
	}

	return a.partieNachDemSchreiben(ctx, kennung, fuer)
}

// partieNachDemSchreiben gibt die Sicht nach einem **geglückten** Schritt.
//
// **Geschrieben ist geschrieben** (§7, Regel 3;
// #1970):
// scheitert das Lesen danach, steht sein Fehler in
// [PartieAntwort.Nachladefehler] und nicht in `fehler`.
//
// **Kennung und Sicht gehen auch ohne Nachlesen mit**: wer den
// Nachladefehler liest, soll die Partie wiederfinden — nach dem Anlegen ist
// das der einzige Weg zu ihr, und wer ihn nicht hat, legt eine zweite an.
func (a *App) partieNachDemSchreiben(
	ctx context.Context, kennung, fuer string,
) PartieAntwort {
	antwort := a.Partie(ctx, kennung, fuer)
	if antwort.Fehler != nil {
		antwort.Nachladefehler, antwort.Fehler = antwort.Fehler, nil
		antwort.Partie.ID, antwort.Partie.Fuer = kennung, fuer
	}

	return antwort
}

// leerePartie gibt eine Antwort mit leeren Listen und dem übersetzten Fehler.
//
// **Leere Listen und kein `nil`** — der Fehlerweg trägt dieselbe Zusage wie
// der Erfolgsweg. Genau hier ist sie in Phase 5.4 zweimal gebrochen (N1, N4),
// und beide Male galt sie nur für den Erfolgsweg.
func leerePartie(ctx context.Context, fehler error) PartieAntwort {
	return PartieAntwort{
		Partie: PartieDTO{
			// **Auch das Regelwerk trägt seine leere Liste** (PLAY-02,
			// #1710): `abgewaehlt` ist eine Liste, und eine Antwort, die
			// dort `null` schickt, ist für das Fenster keine leere Liste,
			// sondern gar keine.
			Regelwerk: RegelwerkDTO{Abgewaehlt: []string{}},

			Sitze:          []SitzDTO{},
			Zuschauer:      []ZuschauerDTO{},
			Objekte:        []ObjektDTO{},
			Mischpflichten: []MischpflichtDTO{},
			Legendenwahlen: []LegendenwahlDTO{},
			Ausgeloest:     []AusloeserDTO{},
			Vorrat:         []string{},

			// **Auch der Zugstand trägt seine leere Liste** (G9, #1592) —
			// derselbe Satz wie die fünf darüber, und dieselbe Stelle, an
			// der er in Phase 5.4 zweimal gebrochen ist.
			Zug: ZugDTO{
				Gepasst:  []string{},
				Angriffe: []KampfpaarDTO{},
				Bloecke:  []KampfpaarDTO{},
			},
			Ausgang: AusgangDTO{Ausgeschieden: []AusgeschiedenDTO{}},
		},
		Erlaubt:    []string{},
		Handlungen: []ObjekthandlungDTO{},
		Befunde:    []string{},
		Fehler:     Uebersetze(ctx, fehler),
	}
}

// Partieaufbau ist, was diese Naht zum Aufbauen braucht.
type Partieaufbau interface {
	Aufbauen(ctx context.Context, kennung, spiel string,
		werk domain.Regelwerk, wirt string, wuensche []service.Sitzwunsch) (
		domain.Partie, error)

	// Mitbringen gibt ein eigenes Deck so, wie ein Gast es an einen fremden
	// Tisch mitbringt (P9b, #2377).
	Mitbringen(ctx context.Context, deckID int64) (domain.Mitgebracht, error)
}

// MitPartieaufbau setzt ihn — für Läufe ohne Bestand (§3).
func (a *App) MitPartieaufbau(aufbau Partieaufbau) *App {
	a.partie.aufbau = aufbau

	return a
}

// partieaufbau gibt den eingesetzten Aufbau.
func (a *App) partieaufbau() Partieaufbau { return a.partie.aufbau }

// MitPartienaht setzt die Leitung — für Läufe ohne Bestand (§3).
func (a *App) MitPartienaht(leitung Partieleitung) *App {
	a.partie.leitung = leitung

	return a
}

// partienaht gibt die eingesetzte Leitung.
func (a *App) partienaht() Partieleitung { return a.partie.leitung }

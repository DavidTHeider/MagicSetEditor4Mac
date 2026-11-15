package app

import "context"

// **Die Bindungen des Suchens, Zählens und der Ablage.**
//
// Eine dritte Datei neben `bindungen.go` und `bindungenkarten.go`, weil jene
// mit ihnen über 400 Zeilen ginge (§2). Der Schnitt folgt der Frage: dort
// stehen die Bindungen **einer Karte und ihres Pakets**, hier die, die über
// ein **ganzes Set** gehen — suchen, auszählen, ausschneiden, einfügen.

func (b *Bindungen) Suchen(ctx context.Context, text, feld, setID string) SuchAntwort {
	return b.app(ctx).Suchen(b.ruf(ctx), text, feld, setID)
}

func (b *Bindungen) Suchfelder(ctx context.Context, setID string) FelderAntwort {
	return b.app(ctx).Suchfelder(b.ruf(ctx), setID)
}

func (b *Bindungen) Statistik(ctx context.Context, setID string) StatistikAntwort {
	return b.app(ctx).Statistik(b.ruf(ctx), setID)
}

func (b *Bindungen) Verteilung(ctx context.Context, setID, feld string) VerteilungAntwort {
	return b.app(ctx).Verteilung(b.ruf(ctx), setID, feld)
}

func (b *Bindungen) SpracheSetzen(_ context.Context, sprache string) SchreibAntwort {
	return b.frei().SpracheSetzen(sprache)
}

// DarstellungsstandMelden sagt der Menüleiste, wie Seitenleiste, Navigator und
// Inspector gerade dastehen — UXR-F14
// (#2029).
//
// **Mit Tor, anders als [Bindungen.SpracheSetzen]**: die Sperranzeige braucht
// keine Seitenleiste, und gesperrt gibt es weder sie noch einen Arbeitsplatz —
// die Leiste graut sie dann selbst ([App.menueAufbauen]).
func (b *Bindungen) DarstellungsstandMelden(ctx context.Context, stand Darstellungsstand) SchreibAntwort {
	return b.app(ctx).DarstellungsstandMelden(stand)
}

// ObjekthandlungenMelden sagt der Menüleiste, welche Handlungen das gewählte
// Objekt hat — GESTE
// (#2358).
// **Mit Tor**, wie [Bindungen.DarstellungsstandMelden]: gesperrt ist nichts
// gewählt.
func (b *Bindungen) ObjekthandlungenMelden(ctx context.Context, meldung Objekthandlungen) SchreibAntwort {
	return b.app(ctx).ObjekthandlungenMelden(meldung)
}

// FensterfehlerMelden schreibt einen Fehler beim Zeichnen ins Protokoll — F01
// (#2340).
//
// **Ohne Tor**, wie [Bindungen.SpracheSetzen]: auch der Sperrbildschirm kann
// beim Zeichnen werfen, und gerade dann soll die Kennung im Protokoll stehen.
func (b *Bindungen) FensterfehlerMelden(ctx context.Context, fehler Fensterfehler) SchreibAntwort {
	return b.frei().FensterfehlerMelden(b.ruf(ctx), fehler)
}

// Wiederanlauf nennt die früheren Sitzungen, die nicht sauber endeten — F04
// (#2343).
// **Ohne Tor**, aus demselben Grund: auch hinter der Sperre soll der Hinweis
// stehen.
func (b *Bindungen) Wiederanlauf(ctx context.Context) WiederanlaufAntwort {
	return b.frei().Wiederanlauf(b.ruf(ctx))
}

// **Die Zwischenablage** — Werkstück 1 von Phase 3.7
// (#1168).
//
// MSE2 kennt `ID_EDIT_COPY`, `ID_EDIT_CUT` und `ID_EDIT_PASTE` über **Karten**
// (`src/gui/set/window.cpp`, Fassung f9d9356 vom 2021-07-06); hier duplizierte
// `KarteKopieren` bis dahin **innerhalb** eines Sets, und zwischen zwei Sets
// ging nichts.

// AblageKopieren legt die genannten Karten in die Zwischenablage.
func (b *Bindungen) AblageKopieren(ctx context.Context, setID string, karteIDs []string) SchreibAntwort {
	return b.app(ctx).AblageKopieren(b.ruf(ctx), setID, karteIDs)
}

// AblageAusschneiden legt sie hinein und entfernt sie aus dem Set.
func (b *Bindungen) AblageAusschneiden(ctx context.Context, setID string, karteIDs []string) SchreibAntwort {
	return b.app(ctx).AblageAusschneiden(b.ruf(ctx), setID, karteIDs)
}

// AblageEinfuegen legt die Karten aus der Zwischenablage im Set an.
func (b *Bindungen) AblageEinfuegen(ctx context.Context, setID string) EinfuegeAntwort {
	return b.app(ctx).AblageEinfuegen(b.ruf(ctx), setID)
}

package app

import "context"

// **Die Bindungen, die durchreichen und sonst nichts tun.**
//
// Eine zweite Datei, weil `bindungen.go` mit ihnen über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort stehen die Bindungen, die **etwas
// entscheiden** — einen Dialog öffnen, ein Muster wählen, ein Ziel bestimmen;
// hier die, die den Kontext des Fensters anhängen und weitergeben.

func (b *Bindungen) Sets(ctx context.Context) SetsAntwort {
	return b.app(ctx).Sets(b.ruf(ctx))
}

// Datenstand sagt, ob ein **anderer** Prozess geschrieben hat (#156).
func (b *Bindungen) Datenstand(ctx context.Context) StandAntwort {
	return b.app(ctx).Datenstand(b.ruf(ctx))
}

// Kartenzeilen gibt die Karten eines Sets als **Liste** — ohne ihre Feldwerte
// (#891).
//
// **Der Weg, den das Fenster beim Öffnen geht.** `Karten` daneben gibt alles;
// bei 3.788 Karten sind das 4.761.182 Bytes JSON je Aufruf.
func (b *Bindungen) Kartenzeilen(ctx context.Context, setID string) KartenzeilenAntwort {
	return b.app(ctx).Kartenzeilen(b.ruf(ctx), setID)
}

// Karte gibt **eine** Karte samt ihren Feldwerten (#891).
//
// **Das Gegenstück zu [Bindungen.Kartenzeilen].** Wer eine Karte aufschlägt,
// braucht alles von ihr — und von keiner anderen.
func (b *Bindungen) Karte(ctx context.Context, karteID string) EinzelkarteAntwort {
	return b.app(ctx).Karte(b.ruf(ctx), karteID)
}

// Kartenfeldwerte gibt **ein** Feld über alle Karten eines Sets (#891).
//
// **Für die Massenänderung**, die neben jeder angehakten Karte den Wert zeigt,
// um den es geht.
func (b *Bindungen) Kartenfeldwerte(ctx context.Context, setID, feld string) KartenfeldwerteAntwort {
	return b.app(ctx).Kartenfeldwerte(b.ruf(ctx), setID, feld)
}

func (b *Bindungen) Karten(ctx context.Context, setID string) KartenAntwort {
	return b.app(ctx).Karten(b.ruf(ctx), setID)
}

// Kartenbilder gibt die kleinen Bilder für die Kartenliste
// (#1166).
//
// **Ein Stapel je Sichtfenster**, nicht ein Ruf je Zeile — der Grund steht in
// `kartenbilder.go`.
func (b *Bindungen) Kartenbilder(ctx context.Context, setID string, karteIDs []string) Kartenbildantwort {
	return b.app(ctx).Kartenbilder(b.ruf(ctx), setID, karteIDs)
}

// Setangaben gibt die Angabenfelder eines Sets samt ihren Werten
// (#567).
func (b *Bindungen) Setangaben(ctx context.Context, setID string) SetangabenAntwort {
	return b.app(ctx).Setangaben(b.ruf(ctx), setID)
}

// SetwertSetzen schreibt eine Setangabe — **bedingt**, wenn das Fenster die
// Werte mitschickt, die es kennt ([App.SetwertSetzenBedingt], U02b).
func (b *Bindungen) SetwertSetzen(ctx context.Context, setID, feld, wert string, bekannt []string) SchreibAntwort {
	return b.app(ctx).SetwertSetzenBedingt(b.ruf(ctx), setID, feld, wert, bekannt)
}

func (b *Bindungen) Vorbehalte(ctx context.Context, setID string) VorbehalteAntwort {
	return b.app(ctx).Vorbehalte(b.ruf(ctx), setID)
}

// FeldSetzen setzt ein Feld einer Karte.
//
// **Jede Kartenänderung nennt das Set** — seit
// #808.
//
// Nicht, weil das Fenster sich vertun könnte, sondern weil es das
// nachweislich tut: §7 Regel 1 steht in der Leitlinie, **weil** eine
// stehengebliebene Kennung nach einem Setwechsel vorgekommen ist. Genau die
// zeigt auf eine Karte des vorigen Sets.
//
// **Und bedingt, wenn das Fenster die Werte mitschickt, die es kennt** —
// [App.FeldSetzenBedingt], U02b (#2179).
func (b *Bindungen) FeldSetzen(ctx context.Context, setID, karteID, feld, wert string, bekannt []string) FeldAntwort {
	return b.app(ctx).FeldSetzenBedingt(b.ruf(ctx), setID, karteID, feld, wert, bekannt)
}

// AutoErsetzen wendet die `auto replace:`-Regeln des Spiels an der
// Schreibmarke an — [App.AutoErsetzen].
//
// **Die Stelle ist eine UTF-16-Einheit**, so wie `selectionStart` sie zählt.
func (b *Bindungen) AutoErsetzen(ctx context.Context, setID, text string, stelle int) AutoersetzungAntwort {
	return b.app(ctx).AutoErsetzen(b.ruf(ctx), setID, text, stelle)
}

func (b *Bindungen) FelderVorschau(
	ctx context.Context,
	setID string, karteIDs []string, feld, wert string,
) MassenvorschauAntwort {
	return b.app(ctx).FelderVorschau(b.ruf(ctx), setID, karteIDs, feld, wert)
}

func (b *Bindungen) FelderSetzenMitErwartung(
	ctx context.Context,
	setID string, erwartet []FeldvorwertDTO, feld, wert string,
) MassenAntwort {
	return b.app(ctx).FelderSetzenMitErwartung(b.ruf(ctx), setID, erwartet, feld, wert)
}

func (b *Bindungen) Verlauf(ctx context.Context, setID string) VerlaufAntwort {
	return b.app(ctx).Verlauf(b.ruf(ctx), setID)
}

func (b *Bindungen) Zuruecknehmen(ctx context.Context, setID, vorgang string) BewegungAntwort {
	return b.app(ctx).Zuruecknehmen(b.ruf(ctx), setID, vorgang)
}

func (b *Bindungen) Wiederherstellen(ctx context.Context, setID, vorgang string) BewegungAntwort {
	return b.app(ctx).Wiederherstellen(b.ruf(ctx), setID, vorgang)
}

func (b *Bindungen) Vorschau(ctx context.Context, karteID string, vergroesserung float64) VorschauAntwort {
	return b.app(ctx).Vorschau(b.ruf(ctx), karteID, vergroesserung)
}

func (b *Bindungen) Probevorschau(
	ctx context.Context,
	karteID, feld, wert string, vergroesserung float64,
) VorschauAntwort {
	return b.app(ctx).Probevorschau(b.ruf(ctx), karteID, feld, wert, vergroesserung)
}

func (b *Bindungen) Stilprobevorschau(
	ctx context.Context,
	karteID, name, wert string, vergroesserung float64,
) VorschauAntwort {
	return b.app(ctx).Stilprobevorschau(b.ruf(ctx), karteID, name, wert, vergroesserung)
}

func (b *Bindungen) KarteAnlegen(ctx context.Context, setID string) AnlegeAntwort {
	return b.app(ctx).KarteAnlegen(b.ruf(ctx), setID)
}

// KarteKopieren legt eine Kopie der Karte an ihr Set
// (#625).
func (b *Bindungen) KarteKopieren(ctx context.Context, setID, karteID string) AnlegeAntwort {
	return b.app(ctx).KarteKopieren(b.ruf(ctx), setID, karteID)
}

// KartenTauschen vertauscht die Stellen zweier Karten
// (#628).
func (b *Bindungen) KartenTauschen(ctx context.Context, setID, eineID, andereID string) SchreibAntwort {
	return b.app(ctx).KartenTauschen(b.ruf(ctx), setID, eineID, andereID)
}

// KarteLoeschen entfernt eine Karte
// (#621).
//
// **Ohne Rückfrage hier** — die gehört ins Fenster, wo jemand sie sieht. Diese
// Stelle führt aus, was schon entschieden ist.
//
// **`trotzdem` ist die eine Rückfrage, die von unten kommt**
// (#1369):
// liegt die Karte in einem Deck, weist der Dienst sie ab und nennt die Decks.
// Das Fenster zeigt das und ruft danach mit `true`.
func (b *Bindungen) KarteLoeschen(ctx context.Context, setID, karteID string, trotzdem bool) SchreibAntwort {
	return b.app(ctx).KarteLoeschen(b.ruf(ctx), setID, karteID, trotzdem)
}

func (b *Bindungen) Kartenfelder(ctx context.Context, setID string) FeldbeschreibungenAntwort {
	return b.app(ctx).Kartenfelder(b.ruf(ctx), setID)
}

// BildSetzen fragt nach einer Bilddatei und bereitet ihren Zuschnitt vor
// (#246,
// #576).
//
// **Der Dialog gehört hierher und nicht ins Frontend** — derselbe Grund wie
// bei [Bindungen.SetOeffnen]: ein Fenster bekäme einen Namen und keinen Pfad.
func (b *Bindungen) BildSetzen(ctx context.Context, karteID, feld string) ZuschnittAntwort {
	return b.app(ctx).BildSetzen(b.ruf(ctx), karteID, feld)
}

// BildAusDatei bereitet den Zuschnitt einer Datei vor, deren Pfad feststeht
// (#246).
//
// **Für das Ziehen und Ablegen.** Die Pfade kommen aus `runtime.OnFileDrop`;
// das Fenster weiß, **über welchem Feld** losgelassen wurde, und Go weiß, wo
// die Datei liegt — der Browser sieht nur einen Namen.
func (b *Bindungen) BildAusDatei(ctx context.Context, karteID, feld, pfad string) ZuschnittAntwort {
	return b.app(ctx).BildAusDatei(b.ruf(ctx), karteID, feld, pfad)
}

// BildEinsetzen schneidet auf die gewählte Fläche zu und legt das Bild ab
// (#576).
func (b *Bindungen) BildEinsetzen(
	ctx context.Context,
	setID, karteID, feld, pfad string, auswahl AuswahlDTO,
) SchreibAntwort {
	return b.app(ctx).BildEinsetzen(b.ruf(ctx), setID, karteID, feld, pfad, auswahl)
}

// Zuschnittgrenze gibt die Auswahl, die MSEs Regeln aus einer gezogenen macht
// (#576).
//
// **Ohne Kontext**: sie rechnet und liest nichts.
func (b *Bindungen) Zuschnittgrenze(
	ctx context.Context,
	quelleBreite, quelleHoehe, zielBreite, zielHoehe int,
	auswahl AuswahlDTO, frei bool, bevorzugt string,
) AuswahlDTO {
	return b.app(ctx).Zuschnittgrenze(
		quelleBreite, quelleHoehe, zielBreite, zielHoehe, auswahl, frei, bevorzugt)
}

// Zuschnittziehen gibt die Auswahl, nachdem der Rahmen gezogen wurde
// (#576).
//
// **Ohne Kontext**: sie rechnet und liest nichts.
func (b *Bindungen) Zuschnittziehen(
	ctx context.Context,
	quelleBreite, quelleHoehe, zielBreite, zielHoehe int,
	auswahl AuswahlDTO, frei bool, x, y float64, kanteX, kanteY int,
) AuswahlDTO {
	return b.app(ctx).Zuschnittziehen(quelleBreite, quelleHoehe, zielBreite, zielHoehe,
		auswahl, frei, x, y, kanteX, kanteY)
}

// Symbolpalette gibt, was ein Feld zum Einfügen anbietet
// (#564).
//
// **An der Karte und nicht am Set**: welche Symbolschrift ein Feld benutzt,
// darf ein Skript sein, und die Bedingungen der Symbolfassungen lesen die
// Karte.
func (b *Bindungen) Symbolpalette(ctx context.Context, karteID, feld string) SymbolpaletteAntwort {
	return b.app(ctx).Symbolpalette(b.ruf(ctx), karteID, feld)
}

// Wortlisten gibt, was ein Feld zur **Auswahl** anbietet
// (#1074).
//
// **An der Karte und nicht am Set** — aus demselben Grund wie
// [Bindungen.Symbolpalette]: welche Wörter in einer Liste stehen, darf ein
// Skript entscheiden, und das liest die Karte.
func (b *Bindungen) Wortlisten(ctx context.Context, karteID, feld string) WortlistenAntwort {
	return b.app(ctx).Wortlisten(b.ruf(ctx), karteID, feld)
}

// WortWaehlen schreibt den Wert, der nach einem Klick ins Auswahlmenü dasteht
// (#1074).
//
// **Beide Kennungen**, wie bei [Bindungen.FeldSetzen] und aus demselben Grund
// (#808): eine stehengebliebene Kartenkennung zeigt nach einem Setwechsel auf
// eine Karte des vorigen Sets.
//
// **Und ein Aufruf, nicht zwei.** Rechnen und Schreiben gehören zusammen (§7,
// Regel 4); dazwischen dürfte niemand tippen.
//
// **`bekannt` kommt vom Entwurf des Feldes**, wie bei [Bindungen.FeldSetzen]
// (U02b, #2179): steht im Bestand ein Wert, den er nicht kennt, schreibt die
// Wahl nicht ([App.WortWaehlenBedingt]).
func (b *Bindungen) WortWaehlen(
	ctx context.Context,
	setID, karteID, feld string, bereich int, weg []int, bekannt []string,
) SchreibAntwort {
	return b.app(ctx).WortWaehlenBedingt(b.ruf(ctx), setID, karteID, feld, bereich, weg, bekannt)
}

// Schluesselwoerter gibt die des Sets **und** die des Spiels
// (#1062).
//
// **Beide in einer Liste**, wie MSE sie zeigt: wer nur die eigenen sähe,
// wüsste nicht, welche das Spiel schon mitbringt — und schriebe ein zweites
// »Fliegend«.
func (b *Bindungen) Schluesselwoerter(ctx context.Context, setID string) SchluesselwoerterAntwort {
	return b.app(ctx).Schluesselwoerter(b.ruf(ctx), setID)
}

// SchluesselwortAnlegen hängt ein leeres ans Ende und gibt seine Kennung.
func (b *Bindungen) SchluesselwortAnlegen(ctx context.Context, setID string) SchluesselwortAntwort {
	return b.app(ctx).SchluesselwortAnlegen(b.ruf(ctx), setID)
}

// SchluesselwortSetzen schreibt ein Feld eines Schlüsselworts des **Sets**.
//
// Die des Spiels tragen keine Kennung; der Weg dorthin ist damit gar nicht
// erst gangbar. **Bedingt**, wenn das Fenster die Werte mitschickt, die es
// kennt (U02b).
func (b *Bindungen) SchluesselwortSetzen(
	ctx context.Context, setID, wortID, feld, wert string, bekannt []string,
) SchreibAntwort {
	return b.app(ctx).SchluesselwortSetzenBedingt(b.ruf(ctx), setID, wortID, feld, wert, bekannt)
}

// SchluesselwortLoeschen nimmt eines weg.
func (b *Bindungen) SchluesselwortLoeschen(ctx context.Context, setID, wortID string) SchreibAntwort {
	return b.app(ctx).SchluesselwortLoeschen(b.ruf(ctx), setID, wortID)
}

// Skript wertet einen MSE-Script-Ausdruck im Kontext einer Karte aus
// (#1070).
//
// **Es ändert nichts.** Die Sprache kann in diesem Programm nicht schreiben,
// und diese Bindung legt nichts dazu.
func (b *Bindungen) Skript(ctx context.Context, karteID, ausdruck string) KonsolenAntwort {
	return b.app(ctx).Skript(b.ruf(ctx), karteID, ausdruck)
}

// Skriptnamen gibt, was ein Ausdruck an dieser Karte nennen darf
// (#1164).
//
// **Für die Vervollständigung und die Liste in der Konsole.** Für jemanden,
// der MSE-Script nicht kennt, ist ein leeres Feld keine Einladung.
func (b *Bindungen) Skriptnamen(ctx context.Context, karteID string) Namensantwort {
	return b.app(ctx).Skriptnamen(b.ruf(ctx), karteID)
}

// Skriptfarben zerlegt einen Ausdruck in seine Merkmale
// (#1164).
//
// **Ohne Kontext der Anwendung**: die Sprache ist die Sprache, gleich an
// welcher Karte. Deshalb nimmt sie keinen `kontext()`.
func (b *Bindungen) Skriptfarben(ctx context.Context, quelltext string) Farbantwort {
	return b.app(ctx).Skriptfarben(quelltext)
}

// Paketfarben zerlegt eine Paketdatei in ihre Merkmale
// (#1181).
func (b *Bindungen) Paketfarben(ctx context.Context, quelltext string) Farbantwort {
	return b.app(ctx).Paketfarben(quelltext)
}

// PaketdateiPruefen gibt den ersten Syntaxfehler in den Skripten einer
// Paketdatei — Werkstück B von Phase 4.9
// (#1324).
func (b *Bindungen) PaketdateiPruefen(ctx context.Context, inhalt string) Skriptbefundantwort {
	return b.app(ctx).PaketdateiPruefen(inhalt)
}

// Vorlagenvorschau zeichnet eine Karte mit einer **ungespeicherten** Vorlage
// (#1164).
// `karteID` nennt die Karte oder bleibt leer für die Probekarte des Pakets
// (#1418).
func (b *Bindungen) Vorlagenvorschau(
	ctx context.Context,
	paket, datei, inhalt string, vergroesserung float64, karteID string,
) VorschauAntwort {
	return b.app(ctx).Vorlagenvorschau(b.ruf(ctx), paket, datei, inhalt, vergroesserung, karteID)
}

// Probekarten gibt die Testkarten zu einem Paket — die Karten des ersten
// Sets, das seinen Stil benutzt (Werkstück U7 von Phase 5.1,
// #1418).
func (b *Bindungen) Probekarten(ctx context.Context, paket string) ProbekartenAntwort {
	return b.app(ctx).Probekarten(b.ruf(ctx), paket)
}

// Paketsymbole gibt jede Definition in den Dateien eines Pakets — für
// Vervollständigung, Hover und den Sprung zur Definition (U7, #1418).
func (b *Bindungen) Paketsymbole(ctx context.Context, paket string) PaketsymboleAntwort {
	return b.app(ctx).Paketsymbole(b.ruf(ctx), paket)
}

// Symbolverwendungen gibt jede Zeile eines Pakets, die einen Namen nennt —
// die Referenzen (U7, #1418).
func (b *Bindungen) Symbolverwendungen(ctx context.Context, paket, name string) SymbolverwendungenAntwort {
	return b.app(ctx).Symbolverwendungen(b.ruf(ctx), paket, name)
}

// Technikangaben nennen den Ort der Werkzeuge und fertige Befehlszeilen
// (#1164).
//
// **Zum Kopieren und nicht zum Ausführen** — das Issue sagt es ausdrücklich:
// keine beliebige Shell.
func (b *Bindungen) Technikangaben(ctx context.Context, setID string) TechnikAntwort {
	return b.app(ctx).Technikangaben(b.ruf(ctx), setID)
}

// Paketdateien nennt die bearbeitbaren Dateien eines Pakets
// (#1164).
func (b *Bindungen) Paketdateien(ctx context.Context, paket string) Paketdateiantwort {
	return b.app(ctx).Paketdateien(b.ruf(ctx), paket)
}

// PaketdateiLesen gibt eine Paketdatei samt ihrem Stand.
func (b *Bindungen) PaketdateiLesen(ctx context.Context, paket, datei string) Paketdateiinhalt {
	return b.app(ctx).PaketdateiLesen(b.ruf(ctx), paket, datei)
}

// PaketdateiSchreiben setzt eine Paketdatei zurück, wenn der Stand noch stimmt.
func (b *Bindungen) PaketdateiSchreiben(ctx context.Context, paket, datei, inhalt, stand string) Paketdateistand {
	return b.app(ctx).PaketdateiSchreiben(b.ruf(ctx), paket, datei, inhalt, stand)
}

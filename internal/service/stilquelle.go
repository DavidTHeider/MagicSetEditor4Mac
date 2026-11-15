package service

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kopfquelle gibt die Köpfe aller Sets im Bestand.
//
// **Alle und nicht einen.** Der Bestand führt keinen Einzelleser, und einen
// dafür anzulegen hieße, eine Abfrage, eine Zeilengestalt im Stolpertreiber und
// einen Fehlerweg zu bauen, um dieselbe Auskunft zu bekommen. Der Preis steht
// bei [Stilspeicher]: gefragt wird **je Ruf**, denn erst der Kopf sagt, ob das
// Gemerkte noch zum Set gehört (#2005).
type Kopfquelle interface {
	Setkoepfe(ctx context.Context) ([]domain.SetHead, error)

	// Setwerte gibt die Angaben eines Sets — die Felder seines
	// `set info:`-Blocks.
	//
	// **Hier und nicht in einer zweiten Schnittstelle**: dieselbe Auskunft über
	// dasselbe Set aus derselben Tabelle, zum selben Zeitpunkt gefragt. Zwei
	// Wege zu einem Ergebnis wären zwei Wege, die auseinanderlaufen können.
	Setwerte(ctx context.Context, setID int64) ([]domain.Value, error)
}

// Stilspeicher findet das Stilpaket eines Sets und behält es.
//
// # Die Naht aus #206
//
// Der Stilplan stand seit #120, und niemand fand das Paket dazu: `Vorschau`,
// `Bildausfuhr` und `Druckbogen` bekamen eine Kartenkennung und riefen den
// Behelfsplan. Hier steht der Weg — Setkopf, Stilname, Paket, Module, Modell.
//
// # Warum gemerkt wird
//
// Eine Vorschau entsteht bei jedem Tastendruck. Ein Stilpaket zu öffnen heißt,
// eine Textdatei zu zerteilen und ihre `include file:`-Module nachzuladen; das
// je Anschlag zu tun ist die Rechnung, die #187 an der Bildquelle aufgemacht
// hat.
//
// **Der Fehlschlag wird mitgemerkt**, denn er ist der häufige Fall: wer kein
// Paket bezogen hat, hat für **jedes** Set keines.
//
// # Wann das Gemerkte verfällt
//
// **Wenn der Setkopf einen anderen Stil oder ein anderes Spiel nennt** —
// geprüft je Ruf ([Stilspeicher.gemerktenStil]) — und **wenn sich eine Datei
// der Pakete ändert**, aus denen die Vorlage entstand: geprüft höchstens
// einmal je Sekunde (`stilstempel.go`). Wer daneben Bilder vorhält, hört davon
// über [Stilspeicher.MitPaketwechsel].
type Stilspeicher struct {
	koepfe  Kopfquelle
	wurzeln []string

	// sperre schützt `gemerkt`. **Eine volle Sperre und keine RW**: gelesen wird
	// hier nur, um gleich darauf zu schreiben, und ein Stil wird je Set genau
	// einmal geholt.
	// bilder sagt, wo die Bilder eines Sets liegen, die jemand **in dieser
	// Anwendung** eingesetzt hat ([Bildablage]). `nil`, solange keine Ablage
	// eingerichtet ist — dann bleibt [Vorlage.Bilderverzeichnis] leer, und
	// gesucht wird wie bisher nur im Herkunftspaket.
	bilder *Bildablage

	// stilwerte gibt die hier gesetzten Stiloptionen ([Stilspeicher.MitStilwerten]).
	stilwerte Stilwertquelle

	sperre  sync.Mutex
	gemerkt map[int64]gemerkterStil

	// paketwechsel hört, wenn ein gemerkter Stil an einem geänderten Paket
	// verfällt — siehe [Stilspeicher.MitPaketwechsel]. `nil`: niemand.
	paketwechsel func()

	// ersatz ersetzt beim Öffnen **eine** Datei eines **einen** Pakets
	// (#1164).
	//
	// **Für die Vorschau einer ungespeicherten Vorlage.** `nil` im Normalfall;
	// gesetzt nur an einem eigens dafür gebauten Speicher, dessen Ablagen leer
	// bleiben — ein Stand, den niemand geschrieben hat, gehört nicht in einen
	// Zwischenspeicher, den andere lesen.
	ersatz *Paketersatz

	// gemerkteStile hält die Stile, die **Karten** nennen — nach Namen.
	//
	// **Eine zweite Ablage und keine gemeinsame**: die erste beantwortet „was
	// hat Set 7", diese „was ist `archenemy`". Ein Set hat wenige Stile und
	// viele Karten (#395).
	gemerkteStile map[string]gemerkterStil

	// kartenstile nennt die eigenen Stile der Karten eines Sets — für den
	// Abdruck einer Prüfung (#2423, [Stilspeicher.MitKartenstilen]). `nil`:
	// der Abdruck kennt nur den Stil des Sets.
	kartenstile Kartenstilnamen

	// schriftpakete hält, aus welchen Paketen eine Symbolschrift gelesen
	// wird — sie selbst und was sie einbindet (#2423).
	schriftpakete map[string]gemerkteSchrift
}

// gemerkterStil ist, was der Speicher zu einem Set behält.
type gemerkterStil struct {
	vorlage Vorlage
	befunde []domain.Befund
	da      bool

	// pakete, stempel und geprueft sagen, woraus die Vorlage entstand und ob
	// es noch so dasteht (`stilstempel.go`).
	pakete   []string
	stempel  string
	geprueft time.Time

	// stil und spiel sind der Kopf des Sets, zu dem die Vorlage gelesen
	// wurde — nennt der Bestand andere, gilt sie nicht mehr (#2005). **Und
	// die Herkunft** (#2423): an ihr hängt, wo die Bilder des Sets liegen
	// ([Vorlage.Setpaket]).
	stil, spiel, herkunft string
}

// NeuerStilspeicher baut den Speicher über einer Reihe von Paketwurzeln.
//
// **Die Wurzeln liegen unmittelbar über den Paketen** — derselbe Vertrag wie
// bei [msepaket.NeueBildquelle] und [msepaket.NeuerPaketleser]. Wer ein
// bezogenes Repositorium hat, lässt [msepaket.Paketwurzeln] **einmal** darüber
// laufen und reicht das Ergebnis an alle drei.
//
// **Nicht hier drinnen**, und das hat einen Grund: der erste Anlauf suchte auch
// hier noch einmal, und damit blieb der Composition Root ungeprüft — die
// Gegenprobe, die `Paketwurzeln` aus `app.go` entfernt, blieb grün, weil dieser
// Speicher die Arbeit stillschweigend nachholte. Die Bildquelle daneben hätte
// trotzdem nichts gefunden.
func NeuerStilspeicher(koepfe Kopfquelle, wurzeln ...string) *Stilspeicher {
	return &Stilspeicher{
		koepfe:        koepfe,
		wurzeln:       wurzeln,
		gemerkt:       make(map[int64]gemerkterStil),
		gemerkteStile: make(map[string]gemerkterStil),
		schriftpakete: make(map[string]gemerkteSchrift),
	}
}

// StilFuerSet gibt den Stil eines Sets; der dritte Wert sagt, ob es einen gibt.
//
// **Kein Fehler, sondern Befunde.** Ein fehlendes Stilpaket ist der Normalfall
// auf einem frischen Rechner; wer daraus einen Fehler machte, ließe die
// Vorschau leer, statt den Behelfsplan zu zeigen und den Grund zu nennen (§8).
func (s *Stilspeicher) StilFuerSet(ctx context.Context, setID int64) (
	Vorlage, []domain.Befund, bool,
) {
	s.sperre.Lock()
	defer s.sperre.Unlock()

	vorlage, befunde, da := s.gemerktenStil(ctx, setID)

	// **Der teure Teil wird gemerkt, der billige jedes Mal gelesen.** Die
	// Setangaben ändern sich im selben Fenster; wer sie mitmerkte, zeigte ein
	// gewechseltes Symbol bis zum Neustart nicht an.
	//
	// **Erst der Stil, dann die Angaben.** Ein Set ohne Symbol ist eine Karte
	// ohne Seltenheitszeichen; ein Set ohne Stil ist keine Karte. Der eine
	// Fehler hält den anderen deshalb nicht auf (§8).
	if da {
		werte, err := s.koepfe.Setwerte(ctx, setID)
		if err != nil {
			befunde = append(befunde, befundliste("die Setangaben lesen: %v", err)...)
		}

		vorlage.Setangaben = werte
		befunde = s.gesetzteStilwerte(ctx, setID, &vorlage, befunde)
	}

	return vorlage, befunde, da
}

// Vergiss räumt, was zu einem Set gemerkt ist
// (#693).
//
// # Warum es das braucht
//
// Der Speicher merkt sich die **Vorlage** je Set. Über der Merkstelle stand:
// „ein `.mse-style` ändert sich nicht, während die Anwendung läuft". Das ist
// wahr für die **Datei** — und nicht vollständig: **welche** Datei ein Set
// benutzt, ändert sich sehr wohl im selben Fenster, und genau dafür gibt es
// die Stilseite.
//
// Bis dahin nahm der Bestand einen Stilwechsel an, und die Vorschau zeigte
// bis zum Neustart dieselbe Karte — Byte für Byte.
//
// # Wer sie ruft, und wer nicht
//
// **Der Stilwechsel.** Er ändert, welche Datei ein Set benutzt.
//
// **Nicht das Setzen einer Stiloption**: die Optionen stehen in den
// Setangaben, und die werden bei jedem Aufruf frisch gelesen — der Grund steht
// bei [Stilspeicher.StilFuerSet].
//
// **Und nicht die Kartenstile.** `gemerkteStile` hängt am **Namen** des
// Stils, nicht am Set; ein Wechsel gibt einen anderen Namen und damit einen
// anderen Schlüssel. Was dort liegt, ist die Datei — und die ändert sich
// wirklich nicht, während die Anwendung läuft.
func (s *Stilspeicher) Vergiss(setID int64) {
	s.sperre.Lock()
	defer s.sperre.Unlock()

	delete(s.gemerkt, setID)
}

// gemerktenStil gibt den Stil eines Sets aus dem Gedächtnis oder holt ihn.
//
// **Ohne die Setangaben** — sie kommen bei jedem Aufruf frisch dazu, siehe
// [Stilspeicher.StilFuerSet].
//
// # Der Kopf je Ruf
//
// **Gemerkt gilt nur, solange der Bestand dasselbe Set beschreibt** —
// #2005.
// Seit die Flächen ohne Fenster ihr Zeichenwerk behalten, sieht ein Speicher
// Stilwechsel aus **anderen** Prozessen — dem Fenster, der Kommandozeile —
// und Rufe aus **anderen** Profilen: `mse-rest` bedient jeden Principal über
// denselben Bestand. Beides beantwortet nur der Kopf, und er ist der billige
// Teil, wie die Setangaben. Der Wechsel im selben Prozess räumt zusätzlich
// über [Stilspeicher.Vergiss].
//
// **Ein Lesefehler lässt das Gemerkte gelten**, solange sein Stempel gilt — er
// sagt nichts über den Stil.
func (s *Stilspeicher) gemerktenStil(ctx context.Context, setID int64) (
	Vorlage, []domain.Befund, bool,
) {
	koepfe, err := s.koepfe.Setkoepfe(ctx)
	gemerkt, gibts := s.gemerkt[setID]

	if err != nil && !gibts {
		// **Nicht gemerkt.** Ein Bestand, der gerade nicht lesbar ist, sagt
		// nichts über das Stilpaket; das festzuhalten machte einen
		// vorübergehenden Zustand dauerhaft.
		return Vorlage{}, befundliste("den Setkopf lesen: %v", err), false
	}

	kopf, da := kopfMit(koepfe, setID)
	if err == nil && !da {
		// **Auch nicht gemerkt**: ein Set, das dieses Profil nicht führt, kann
		// ein anderes führen — und dessen nächster Ruf bekäme sonst das Fehlen.
		return Vorlage{}, befundliste("das Set %d führt der Bestand nicht", setID), false
	}

	gleicherKopf := err != nil ||
		(kopf.Style == gemerkt.stil && kopf.Game == gemerkt.spiel && kopf.Origin == gemerkt.herkunft)

	if gibts && gleicherKopf && s.stempelGilt(&gemerkt) {
		s.gemerkt[setID] = gemerkt

		// **Eine Kopie der Befundliste.** [Stilspeicher.StilFuerSet] hängt den
		// Befund über die Setangaben an; ohne Kopie schriebe es dabei in den
		// freien Platz des gemerkten Feldes. Die gemerkte Länge bliebe, wie sie
		// war — der Speicherplatz aber ist geteilt, und der zweite Leser
		// überschriebe die Meldung, die der erste noch in der Hand hält.
		return gemerkt.vorlage, slices.Clone(gemerkt.befunde), gemerkt.da
	}

	if err != nil {
		return Vorlage{}, befundliste("den Setkopf lesen: %v", err), false
	}

	vorlage, befunde, gefunden, pakete := s.holen(kopf)

	neu := s.gestempelt(vorlage, befunde, gefunden, pakete)
	neu.stil, neu.spiel, neu.herkunft = kopf.Style, kopf.Game, kopf.Origin
	s.gemerkt[setID] = neu

	return vorlage, befunde, gefunden
}

// setkopf gibt den Kopf eines Sets samt Befund, wenn er fehlt.
//
// **Nicht gemerkt** — derselbe Grund wie bei [Stilspeicher.gemerktenStil]: ein
// Bestand, der gerade nicht lesbar ist, sagt nichts über den Stil.
func (s *Stilspeicher) setkopf(ctx context.Context, setID int64) (
	domain.SetHead, []domain.Befund, bool,
) {
	koepfe, err := s.koepfe.Setkoepfe(ctx)
	if err != nil {
		return domain.SetHead{}, befundliste("den Setkopf lesen: %v", err), false
	}

	kopf, da := kopfMit(koepfe, setID)
	if !da {
		return domain.SetHead{}, befundliste("das Set %d führt der Bestand nicht", setID), false
	}

	return kopf, nil, true
}

// holen sucht den Stil zu einem Setkopf.
//
// **Jedes Ergebnis taugt zu merken**: der Kopf ist gelesen, und was ein
// Lesefehler oder ein fehlendes Set wäre, hat [Stilspeicher.gemerktenStil]
// schon abgefangen.
func (s *Stilspeicher) holen(kopf domain.SetHead) (
	vorlage Vorlage, befunde []domain.Befund, gefunden bool, pakete []string,
) {
	if kopf.Style == "" {
		return Vorlage{}, befundliste("das Set nennt keinen Stil"), false, nil
	}

	if len(s.wurzeln) == 0 {
		return Vorlage{},
			befundliste("für den Stil %q ist kein Paketverzeichnis eingerichtet", kopf.Style),
			false, nil
	}

	vorlage, befunde, gefunden, pakete = s.ausPaket(kopf.Style, kopf.Game)

	// **Der Ort des Sets reist mit der Vorlage**, weil das Bild einer Karte
	// dort liegt und nicht in einer Vorlage
	// (#380).
	//
	// **Auch wenn das Stilpaket nicht aufging.** Die Vorlage ist dann leer, der
	// Ort stimmt trotzdem. Weiter oben steht er nicht: wer den Setkopf gar nicht
	// lesen konnte, hat keinen — und die Vorschau nimmt dann den Behelfsplan.
	s.setort(&vorlage, kopf)

	return vorlage, befunde, gefunden, pakete
}

// MitPaketwechsel gibt dem Speicher jemanden, der hört, wenn ein gemerkter
// Stil an einem **geänderten Paket** verfällt
// (#2005).
//
// # Wozu
//
// Der Stempel (`stilstempel.go`) liest die Vorlage neu — die **Bilder** aber
// liegen daneben, in der Bildquelle und im Ebenenspeicher, entschlüsselt und
// nach ihrem Pfad gemerkt. Ohne die Meldung zeichnete eine Karte den neuen
// Stil mit dem alten Rahmen, bis der Prozess endet — und `mse-mcp` oder
// `mse-rest` leben so lange wie ihr Klient.
//
// **Beim Bau zu setzen, nicht danach**: ein Speicher, der schon liest, hätte
// sonst ein Feld, das zwei Läufe zugleich lesen und schreiben.
func (s *Stilspeicher) MitPaketwechsel(hoerer func()) *Stilspeicher {
	s.paketwechsel = hoerer

	return s
}

// Paketersatz ist eine Datei, die beim Öffnen eines Pakets aus dem Speicher
// kommt statt von der Platte.
type Paketersatz struct {
	// Paket ist der Pfad des Pakets, auf das der Ersatz zielt.
	Paket string

	// Datei ist der Paketpfad der Datei darin.
	Datei string

	Inhalt []byte
}

// MitErsatz gibt einen **eigenen** Speicher, der eine Datei ersetzt.
//
// # Warum ein eigener und keine Einstellung am vorhandenen
//
// Weil ein Stand, den niemand geschrieben hat, nicht in einen
// Zwischenspeicher gehört, aus dem alle anderen lesen. Wer hier den
// gemeinsamen Speicher umstellte, zeichnete jede Karte nach einer Vorlage, die
// es auf der Platte nicht gibt — und niemand sähe, woher das kommt.
//
// **Die Ablagen bleiben leer**, und das ist der Preis: jede Vorschau öffnet
// das Stilpaket neu. Eine Vorschau wird ohnehin abgebrochen und nicht
// abgewartet (§7); was sie kostet, kostet sie einmal je Tastendruckfolge.
func (s *Stilspeicher) MitErsatz(ersatz Paketersatz) *Stilspeicher {
	return &Stilspeicher{
		koepfe:        s.koepfe,
		wurzeln:       s.wurzeln,
		bilder:        s.bilder,
		ersatz:        &ersatz,
		gemerkt:       make(map[int64]gemerkterStil),
		gemerkteStile: make(map[string]gemerkterStil),
	}
}

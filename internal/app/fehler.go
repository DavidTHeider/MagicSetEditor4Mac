package app

import (
	"context"
	"errors"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fehler ist die Form, in der ein Fehler das Frontend erreicht.
//
// Drei Felder, drei Aufgaben: `Code` steuert das Verhalten, `Schluessel` die
// Übersetzung, `Meldung` ist der Rückfall, wenn der Katalog den Schlüssel nicht
// kennt (AGENTS.md §9). `Feld` sagt, wo die Meldung erscheint — leer, wenn sie
// zu keinem einzelnen Feld gehört.
type Fehler struct {
	Code       Fehlercode        `json:"code"`
	Schluessel domain.Schluessel `json:"schluessel"`
	Meldung    string            `json:"meldung"`
	Feld       string            `json:"feld,omitempty"`

	// Werte füllt die Platzhalter des Katalogsatzes.
	//
	// **Die Lücke, die #940 aufgedeckt hat.** Der Katalog kann `{name}`
	// einsetzen — für einen Fehler aus Go konnte er es nicht, und jede
	// Meldung mit einem konkreten Namen stand vor der Wahl, ihn zu verlieren
	// oder den Katalog zu umgehen. Beides ist einmal geschehen.
	//
	// **Was hier hineingehört, sind Bezeichner**, keine Sätze: Dateinamen,
	// Feldnamen, Paketnamen. Sie werden nicht übersetzt (§9), der Rahmen um
	// sie herum schon.
	//
	// **Ohne `omitempty`**, und das ist keine Nachlässigkeit: es ist genau die
	// Falle aus §3 — eine leere Tabelle verschwände aus der Antwort, und im
	// Fenster stünde `undefined` statt einer leeren Tabelle. Der Wächter
	// `leerelisten_test.go` hält es fest.
	//
	// **Und trotzdem kam sie als `null` an**, bis zum 2026-08-29: `omitempty`
	// fehlte, aber die meisten Fehler setzen dieses Feld gar nicht, und eine
	// nicht gesetzte Tabelle ist in Go `nil`. Der erzeugte Typ sagt
	// `werte: Record<string, string>`, **nicht** optional. Gemessen an der
	// laufenden Brücke, und danach an 110 von 142 Bindungen
	// (#1288).
	// [Fehler.MarshalJSON] löst es an **einer** Stelle ein.
	Werte map[string]string `json:"werte"`

	// Vorgang ist die Kennung des Rufs, in dem dieser Fehler entstand —
	// dieselbe, die in seinen Protokollzeilen steht (Werkstück O9 von Phase
	// 5.2, #1433).
	//
	// **Mit `omitempty`**, anders als [Fehler.Werte]: das Fenster zeigt sie
	// nur, wenn sie da ist, und ein Test ohne Grenze hat keine. Sie ist eine
	// Zeichenkette und keine Liste — der Grund aus §3, der `Werte` das
	// `omitempty` verbietet, greift hier nicht.
	Vorgang string `json:"vorgang,omitempty"`
}

// rueckfallmeldung ist der Satz für alles, was wir nicht einordnen können.
//
// Er steht hier und nicht im Katalog, weil er auch dann tragen muss, wenn die
// Übersetzung selbst nicht funktioniert.
const rueckfallmeldung = "Es ist ein unerwarteter Fehler aufgetreten."

// uebersetze bringt einen Fehler in die Form, die das Frontend kennt.
//
// **Rohe Meldungen aus dem Unterbau erreichen die Oberfläche nie** (AGENTS.md
// §8). Ein „SQL logic error: no such table: cards" sagt Nutzenden nichts und
// verrät zugleich den inneren Aufbau; er wird durch den Rückfallsatz ersetzt.
// Die einzige Ausnahme ist der Feldbefund: sein Satz ist für Nutzende
// geschrieben.
//
// **Nicht exportiert** — [Uebersetze] ist die Tür, und sie stempelt die
// Vorgangskennung auf (Werkstück O9, #1433).
func uebersetze(ctx context.Context, fehler error) *Fehler {
	if fehler == nil {
		return nil
	}

	// **Ein voller Datenträger ist einer, gleich wer schrieb** (#1812) — eine
	// Datei meldet ihn als `ENOSPC`, und hier wird er eingeordnet, bevor der
	// Grund ins Protokoll geht.
	fehler = domain.Zugriffsmangel(domain.Platzmangel(fehler))

	// **Ein verdeckter Grund geht ins Protokoll, bevor die Meldung hinausgeht**
	// — Werkstück 1 von Phase 4.13
	// (#1352).
	//
	// **Hier und nicht im `default`-Zweig**: wer [domain.MitGrund] benutzt,
	// hat die Ursache absichtlich aus dem Satz genommen — sie ist dann an
	// keiner anderen Stelle mehr zu finden. Der Zweig unten fängt nur die,
	// die gar nicht eingeordnet sind.
	//
	// **Und es widerspricht dem Absatz dort nicht.** „Erwartete Ausgänge
	// werden nicht mitgeschrieben" gilt weiter: eine Validierung **ohne**
	// Grund erzeugt keine Zeile. Mitgeschrieben wird der Grund, nicht der
	// Ausgang.
	if grund := domain.GrundVon(fehler); grund != nil {
		slog.Default().ErrorContext(ctx, "der verdeckte Grund einer Meldung",
			"vorgang", fehler.Error(), "grund", grund.Error())
	}

	var befund *domain.FieldError
	if errors.As(fehler, &befund) {
		// **Die Grenze normalisiert, nicht der Konstruktor allein.**
		//
		// `NewFieldError` weist einen nicht angemeldeten Schlüssel bereits ab —
		// aber `FieldError` ist ein exportierter Verbund, und
		// `&domain.FieldError{Key: "gibtsNicht"}` umgeht den Konstruktor. Der
		// Nachlauf des Zweitblicks hat genau das gefunden: geschlossen war der
		// Konstruktorweg, nicht die Zusage „kein unbekannter Schlüssel erreicht
		// das Frontend". Hier ist die Stelle, an der sie eingelöst wird, denn
		// hier führt kein Weg vorbei.
		schluessel := befund.Key
		if !domain.IstAngemeldet(schluessel) {
			schluessel = domain.SchluesselUnbekannt
		}

		// Der Code folgt der Fehlerart, nicht dem Umstand, dass es ein
		// Feldbefund ist: ein Konflikt braucht im Frontend ein Nachladen, eine
		// Validierung nicht.
		code := CodeValidierung
		if errors.Is(fehler, domain.ErrConflict) {
			code = CodeKonflikt
		}

		return &Fehler{
			Code:       code,
			Schluessel: schluessel,
			Meldung:    befund.Message,
			Feld:       befund.Field,
			// **Die Werte reisen mit** (#1369): der Katalog des Fensters
			// setzt sie in seinen eigenen Satz ein, und eine englische
			// Sitzung bekäme sie sonst nicht.
			Werte: befund.Werte,
		}
	}

	switch {
	// **Die zwei Lagen des Tors vor allem anderen** (A02, #2216): beide
	// umschließen einen allgemeinen Fehler — den Abbruch, die Belegung —,
	// und ihr eigener Satz sagt mehr.
	case errors.Is(fehler, errAnwendungEndet):
		return mitSatz(CodeAbgebrochen, domain.SchluesselAnwendungEndet, fehler)
	case errors.Is(fehler, errBestandWirdGetauscht):
		return mitSatz(CodeBelegt, domain.SchluesselBestandWirdGetauscht, fehler)

	// **Ein Ablauf vor dem Abbruch** (BK-07d, #2118): er umschließt
	// `context.DeadlineExceeded`, braucht aber einen anderen Rat.
	case errors.Is(fehler, domain.ErrAnmeldungAbgelaufen):
		return einfach(CodeAbgebrochen, domain.SchluesselAnmeldungAbgelaufen)

	// **Eine Sicherung ohne Vermerk vor allem anderen** (OPS01, #2347): sie
	// umschließt den Grund, aus dem der Vermerk scheiterte — belegt, voll,
	// abgebrochen —, und jeder dieser Sätze sagte, es sei nichts gesichert.
	case errors.Is(fehler, domain.ErrSicherungUnvermerkt):
		return mitSatz(CodeKonflikt, domain.SchluesselSicherungUnvermerkt, fehler)

	// **Der Abbruch zuerst.** Er umschließt oft einen anderen Fehler — der
	// Rasterer bricht ab und gibt „die Karte zeichnen: context canceled" —,
	// und wer ihn hinter die Fachfehler stellte, träfe ihn nie.
	case errors.Is(fehler, context.Canceled),
		errors.Is(fehler, context.DeadlineExceeded):
		return einfach(CodeAbgebrochen, domain.SchluesselVorgangAbgebrochen)

	// **Vor den Fachfehlern und aus demselben Grund wie der Abbruch.** Ein
	// belegter Bestand umschließt oft die Meldung des Vorgangs, der ihn
	// antraf; hinter `ErrNotFound` gestellt träfe er dieselbe Meldung nicht.
	case prozessVeraltet(fehler) != nil:
		return prozessVeraltet(fehler)
	case errors.Is(fehler, domain.ErrBelegt):
		return einfach(CodeBelegt, domain.SchluesselBelegt)

	case errors.Is(fehler, domain.ErrVoll):
		return mitSatz(CodeVoll, domain.SchluesselVoll, fehler)

	// **Ein Fehler des Datenträgers bei knappem Platz nennt beides** (#2407)
	// — unter dem Code des vollen Datenträgers, denn Platz zu schaffen ist
	// der erste Schritt; der Satz sagt, dass der Datenträger zu prüfen ist.
	case errors.Is(fehler, domain.ErrKaumPlatz):
		return mitSatz(CodeVoll, domain.SchluesselKaumPlatz, fehler)
	case abschaltung(fehler) != nil:
		return abschaltung(fehler)
	case dienstAbgeschaltet(fehler) != nil:
		return dienstAbgeschaltet(fehler)
	case gegenstelle(fehler) != nil:
		return gegenstelle(fehler)
	case kontingent(fehler) != nil:
		return kontingent(fehler)

	// **Ein gesperrter Ort ist eine Eingabe, die nicht passt** (FIX-F02):
	// der Rat ist ein anderer Ort, kein Nachladen.
	case errors.Is(fehler, domain.ErrGesperrt):
		return mitSatz(CodeValidierung, domain.SchluesselGesperrt, fehler)

	// **Der eigene Satz reist mit** —
	// #1055.
	//
	// Hier stand `einfach(…)`, und das ersetzte **jeden** Satz durch den
	// Rückfall. Ein `fmt.Errorf("%w: der Zugang zu Scryfall ist nicht
	// eingerichtet", domain.ErrValidation)` kam damit an jeder Fläche als „Es
	// ist ein unerwarteter Fehler aufgetreten" an — dreizehn Stellen schrieben
	// einen Satz, den niemand je gelesen hat.
	//
	// **Für das Fenster ändert sich nichts**: es übersetzt über den Schlüssel,
	// und `unbekannt` steht im Katalog. Der Satz erreicht nur die, die keinen
	// Katalog haben — Kommandozeile, MCP und REST.
	//
	// **Nur die drei fachlichen Ausgänge.** Der Zweig darunter bleibt beim
	// Rückfall: was dort ankommt, kann eine SQLite-Meldung sein, und die
	// erreicht die Oberfläche nie (§8).
	// **Vor dem allgemeinen `ErrNotFound`**, das er umschließt — dahinter
	// gestellt käme er nie an.
	case errors.Is(fehler, domain.ErrKeinSetMitStil):
		return mitSatz(CodeNichtGefunden, domain.SchluesselKeinSetMitStil, fehler)

	// **Und aus demselben Grund davor** (N2, #1455): „lade die Liste neu"
	// ist der falsche Rat für ein Profil, das noch keine Identität hat.
	case errors.Is(fehler, domain.ErrOhnePeeridentitaet):
		return mitSatz(
			CodeNichtGefunden, domain.SchluesselOhnePeeridentitaet, fehler)
	// **Ein Gegenüber, das nicht horcht** (#2491) — nichts ist gelöscht.
	case errors.Is(fehler, domain.ErrPeerNichtErreichbar):
		return mitSatz(
			CodeNichtGefunden, domain.SchluesselPeerNichtErreichbar, fehler)
	case errors.Is(fehler, domain.ErrPeerOhneAdresse):
		return mitSatz(
			CodeNichtGefunden, domain.SchluesselPeerOhneAdresse, fehler)
	case errors.Is(fehler, domain.ErrNotFound):
		return mitSatz(CodeNichtGefunden, domain.SchluesselNichtGefunden, fehler)
	// **Die drei Ausgänge eines Umschlags stehen in `fehlerumschlag.go`**
	// (N4, #1457) — sie beantworten dieselbe Frage, und diese Datei ginge
	// mit ihnen über die Zeilengrenze aus §2.
	case umschlagausgang(fehler) != nil:
		return umschlagausgang(fehler)

	// **Vor `ErrConflict`, weil er einer ist** (N3, #1456): ein offener
	// Fernhorchposten braucht einen eigenen Rat — nichts zu ändern, sondern
	// erst zu beenden.
	case errors.Is(fehler, domain.ErrFernhorchenLaeuft):
		return mitSatz(
			CodeKonflikt, domain.SchluesselFernhorchenLaeuft, fehler)

	// **Nur ein überholter Stand heißt „lade neu"** (ENG-05.AK-03, F08);
	// jeder andere Konflikt nennt seinen Grund — `konfliktgrund.go`.
	case errors.Is(fehler, domain.ErrUeberholt):
		return mitSatz(CodeKonflikt, domain.SchluesselKonflikt, fehler)

	case dienstHaelt(fehler) != nil:
		return dienstHaelt(fehler)
	case errors.Is(fehler, domain.ErrConflict):
		return konfliktgrund(fehler)

	// **Der Skriptfehler steht VOR ErrValidation** — er ist damit umschlossen
	// und landete sonst dort
	// (#1164).
	//
	// Genau derselbe Griff wie bei `fremdeArchivfassung` eine Zeile weiter
	// unten, und aus demselben Grund: der Schlüssel entscheidet, was im
	// Fenster steht. Am laufenden Programm nachgesehen — über dem
	// angestrichenen `* 2` stand „Es ist ein unerwarteter Fehler
	// aufgetreten", weil `unbekannt` im Katalog steht und das Fenster den
	// Katalog bevorzugt (§3, „zwischen den Prüfebenen").
	case istSkriptfehler(fehler):
		return skriptfehler(fehler)

	// **Und der Nur-Lese-Bestand ebenso weit vorn** — er wird auf dem Weg
	// nach oben in den Satz des Vorgangs eingewickelt, den er antraf.
	case errors.Is(fehler, domain.ErrNurLesend):
		return mitSatz(CodeNurLesend, domain.SchluesselNurLesend, fehler)

	// **Der fehlende Schlüssel steht weit vorn**, weil er auf dem Weg nach
	// oben in „Datenbank öffnen: …" eingewickelt wird und die Zweige darunter
	// ihn sonst nicht mehr erkennen.
	case errors.Is(fehler, domain.ErrSchluesselFehlt):
		return mitSatz(CodeSchluesselFehlt, domain.SchluesselSchluesselFehlt, fehler)

	// **Die zu knappe Erlaubnis steht VOR ErrValidation** — sie ist keine, und
	// wäre sie eine, stünde das Formular noch einmal da (Werkstück 11, #1351).
	case errors.Is(fehler, domain.ErrFremderNutzer):
		// **Vor der allgemeinen Erlaubnis, weil er eine ist** (ID-03, #1702):
		// ihr Satz schickte zu einem Verwalter, und hier hilft nur der
		// Wechsel zu dem Menschen selbst.
		return mitSatz(CodeErlaubnis, domain.SchluesselFremderNutzer, fehler)

	// **Und davor die Absage eines Gegenübers** (P11, #2379): sie ist eine
	// Erlaubnisfrage, aber keine nach einer Rolle in diesem Profil.
	case errors.Is(fehler, domain.ErrNichtFreigegeben):
		return mitSatz(CodeErlaubnis, domain.SchluesselNichtFreigegeben, fehler)

	case errors.Is(fehler, domain.ErrErlaubnis):
		return mitSatz(CodeErlaubnis, domain.SchluesselErlaubnis, fehler)

	// **Und davor, weil er auch einer ist** (N3, #1456): eine Einladung aus
	// einer neueren Fassung ist kein Tippfehler.
	case errors.Is(fehler, domain.ErrEinladungZuNeu):
		return mitSatz(
			CodeValidierung, domain.SchluesselEinladungZuNeu, fehler)

	// **Und vor `ErrValidation`, weil er einer ist** (N2, #1455): ins
	// lokale Netz ging nichts hinaus — weder eine Suche noch, seit K02
	// (#2188), eine Ankündigung —, und der nächste Schritt liegt in den
	// Systemeinstellungen und nicht im Formular.
	case errors.Is(fehler, domain.ErrOhneNetzerlaubnis):
		return mitSatz(
			CodeValidierung, domain.SchluesselOhneNetzerlaubnis, fehler)

	case errors.Is(fehler, domain.ErrOhneNetzroute):
		return mitSatz(CodeValidierung, domain.SchluesselOhneNetzroute, fehler)

	case errors.Is(fehler, domain.ErrSchluesselbundAblage):
		return mitSatz(CodeValidierung, domain.SchluesselSchluesselbundAblage, fehler)

	case errors.Is(fehler, domain.ErrValidation):
		return validierungsgrund(fehler)

	// **Die zwei Ausgänge der Sicherungsprüfung** (#940). Sie stehen **vor**
	// dem Unbekannten, weil sie sonst dort landeten — mit einem Protokolleintrag
	// über einen Fehler, der keiner ist, und einem Satz, der nichts sagt.
	case istNeuereFassung(fehler):
		return neuereFassung(fehler)

	case istAeltereFassung(fehler):
		return aeltereFassung(fehler)

	case istFremdeArchivfassung(fehler):
		return fremdeArchivfassung(fehler)

	case errors.Is(fehler, domain.ErrSicherung):
		return einfach(CodeValidierung, domain.SchluesselSicherungUnbrauchbar)
	case errors.Is(fehler, domain.ErrSicherungVerworfen):
		return mitSatz(CodeValidierung, domain.SchluesselSicherungVerworfen, fehler)

	// **Der unbekannte Fehler geht ins Protokoll, bevor er verschwindet**
	// (#857).
	//
	// Die Zeile darüber ersetzt ihn durch den Rückfallsatz — das ist richtig
	// (§8) und macht ihn zugleich unsichtbar. Genau das ist die Sorte Fehler,
	// von der man später **nur** im Protokoll erfährt.
	//
	// **Nur dieser Zweig.** Eine Validierung, ein fehlender Datensatz, ein
	// Konflikt und ein Abbruch sind erwartete Ausgänge; ein Protokoll, das sie
	// mitschreibt, verdeckt den einen Eintrag, um dessentwillen es da ist.
	//
	// **Über `slog.Default()` und nicht über ein durchgereichtes Protokoll**:
	// `Uebersetze` wird von jeder Bindung gerufen und hat keinen Empfänger.
	// `main.go` setzt den Standard auf den schwärzenden Schreiber; in Tests
	// bleibt es der von Go, und der schreibt nach stderr.
	default:
		slog.Default().ErrorContext(ctx, "ein unbekannter Fehler erreichte die Oberfläche",
			"grund", fehler.Error(), "code", string(CodeUnbekannt))

		return einfach(CodeUnbekannt, domain.SchluesselUnbekannt)
	}
}

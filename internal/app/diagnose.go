// **Die Diagnose** — Werkstück 4 von Phase 3.5
// (#1166).
//
// # Warum es sie gibt
//
// `docs/ADMINISTRATION.md` nennt für einen Fehlerbericht
// `tail -n 200 …/mse.log`. **Die Zielgruppe dieser Anwendung kann das nicht**
// — sie mag Magic und sammelt Karten, und ein Terminal ist für sie kein
// Werkzeug, sondern eine Hürde (Z9-01 im Audit).
//
// # Was sie enthält, und warum genau das
//
// Was einen Fehlerbericht brauchbar macht und **niemandem gehört**:
//
//   - **Fassung und Architektur** — die erste Frage bei jedem Bericht.
//   - **Der Ort des Bestands**, mit `~` statt des Benutzerverzeichnisses.
//   - **Der Schemastand** — welche Migrationen verbucht sind. Ein Fehler nach
//     einer Aktualisierung ist ohne ihn nicht einzuordnen.
//   - **Die Paketstände** — welche Vorlagenpakete in welcher Fassung
//     dastehen. Eine Karte, die falsch aussieht, hängt fast immer daran.
//   - **Das Ende des Protokolls** — begrenzt auf [Protokollzeilen]. Die
//     letzten Zeilen sind die interessanten; die ersten sind bei einem Monat
//     Betrieb Megabyte, die niemand liest.
//
// # Was sie nicht enthält
//
// **Keine Kartentexte, keine Feldwerte, keine Setinhalte.** Sie entstehen hier
// nicht, und das Protokoll trägt sie nicht: `geschwaerzt` lässt nur angemeldete
// Merkmale durch, und `protokollmerkmale_test.go` verlangt, dass jede Meldung
// ein Zeichenkettenliteral ist.
//
// **Und keine Marke.** Die Marke des Dienstes steht in einer Umgebungsvariablen
// und in keiner Zeile, die hier zusammengesetzt wird — dieser Weg liest den
// Bestand nicht und die Umgebung nicht.
//
// # Und sie wird gezeigt, bevor sie irgendwohin geht
//
// Was hier entsteht, ist eine **Zeichenkette**. Sie zu kopieren oder zu sichern
// ist eine Handlung des Fensters, und davor steht sie da. Eine Anwendung, die
// von selbst etwas hinausschickte, wäre etwas anderes.

package app

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Schemastand sagt, welche Migrationen im Bestand verbucht sind.
//
// **Eine Naht und kein Griff nach `internal/storage`.** Diese Datei ist kein
// Composition Root (§2); wer den Bestand kennt, ist `start.go`.
type Schemastand interface {
	AngewendeteMigrationen(ctx context.Context) ([]string, error)

	// Waisen zählt die Zeilen, die zu nichts mehr gehören
	// (#1220).
	//
	// **Sie stehen in der Diagnose, weil `0022_waisen.sql` sie wegräumt und
	// niemand merkte, wenn wieder welche entstünden.** Der Bestand ist nach
	// der Migration sauber; diese Zeile ist der Wächter darüber, dass er es
	// bleibt.
	Waisen(ctx context.Context) (string, error)
}

// MitSchemastand setzt die Naht — für Tests und für den Composition Root.
func (a *App) MitSchemastand(stand Schemastand) { a.schemaquelle = stand }

// Protokollzeilen ist, wie viel vom Protokollende die Diagnose trägt.
//
// **200, wie in der Administration.** Die Zahl stand dort schon als
// `tail -n 200`; sie hier anders zu wählen hieße, zwei Antworten auf dieselbe
// Frage zu haben.
const Protokollzeilen = 200

// Paketzeilen ist, wie viele Pakete die Diagnose einzeln nennt.
//
// **Fünfzig.** Ein Bestand mit dem vollen Pack hat 331; sie alle aufzuzählen
// wären 331 Zeilen vor dem Protokoll, und das Protokoll ist der Teil, den
// jemand liest. Die **Zahl** steht immer davor — sie ist die Auskunft, die
// zählt.
const Paketzeilen = 50

// DiagnoseAntwort trägt den fertigen Text.
type DiagnoseAntwort struct {
	Text   string  `json:"text"`
	Fehler *Fehler `json:"fehler"`
}

// Diagnose setzt den Text zusammen, den jemand einem Bericht beilegen kann.
//
// **Und merkt ihn sich.** [App.DiagnoseSichern] schreibt genau **diesen** Text
// und erzeugt keinen zweiten: zwischen Anzeigen und Sichern wächst das
// Protokoll weiter, und die Datei trüge sonst Zeilen, die niemand gesehen hat.
// Die Zustimmung gälte einem anderen Text als dem gespeicherten — angestrichen
// vom Zweitblick über Phase 3.5 als Befund 5.
func (a *App) Diagnose(ctx context.Context) DiagnoseAntwort {
	ctx = a.imProfil(ctx)

	// **Die Aufschriften kommen aus dem Sprachkatalog dieses Pakets**
	// (#1215, Befund 4). Der Text geht durch keinen der fünf Frontendkataloge
	// — `keintextimjsx` sieht ihn nicht —, und die englische Oberfläche zeigte
	// deutsche Bezeichner mitten in dem Bereich, der für wenig technische
	// Menschen gebaut ist.
	texte := DialogtexteFuer(a.sprache)

	var text strings.Builder

	text.WriteString("MagicSetEditor4Mac " + Fassung + "\n")
	fmt.Fprintf(&text, "%s/%s, Go %s\n",
		runtime.GOOS, runtime.GOARCH, runtime.Version())
	text.WriteString(bauzeile(DiesenBau(), texte) + "\n")
	text.WriteString(texte.DiagnoseBindungsvertrag + ": " + Vertragsabdruck() + "\n")

	// **Der Ort des Bestands, auch wenn er nicht zu ermitteln ist.**
	//
	// Genau dann ist die Diagnose am nötigsten: eine Anwendung, die ihren
	// Bestand nicht findet, tut gar nichts, und der Grund gehört in den
	// Bericht. Ein Weg, der hier abbräche, ließe den Fall unberichtet, für den
	// es ihn gibt (§8).
	//
	// **Und der, der wirklich offen ist** (QA-10.NAT-01, #1741): hier stand
	// der Vorgabeort, und mit eigenem Bestand nannte die Diagnose eine Datei,
	// in der nichts stand.
	bestand, err := BestandsortFuerFlaechen()
	if err != nil {
		bestand = texte.DiagnoseNichtLesbar + ": " + err.Error()
	}

	text.WriteString(texte.DiagnoseBestand + ": " + bestand + "\n")

	text.WriteString(texte.DiagnoseSchema + ": " + a.schemastand(ctx, texte) + "\n")
	text.WriteString(texte.DiagnoseWaisen + ": " + a.waisenstand(ctx, texte) + "\n")

	pfad, protokollfehler := a.protokolldatei()
	if protokollfehler == nil {
		text.WriteString(texte.DiagnoseProtokollort + ": " + pfad + "\n")
	}

	text.WriteString(a.protokollzeilen(texte))

	// **Der Kopf bleibt lesbar, Pakete und Protokoll werden redigiert** —
	// SEC-P5 (#2001),
	// dieselben zwei Teile, die das Diagnosepaket redigiert. Darüber stehen Fassung, Bau, Orte und Stände —
	// die Auskunft über die Installation, und ein Ort, den jemand selbst
	// gewählt hat, soll lesbar bleiben. Darunter steht freier Text, und das
	// Protokollende kann Zeilen tragen, die eine ältere Fassung ungeschwärzt
	// schrieb.
	kopf := text.String()
	text.Reset()

	text.WriteString("\n" + a.paketstaende(ctx, texte))

	// **Und das Protokoll nur, wenn die Rolle bis dorthin reicht** —
	// Werkstück O9 von Phase 5.2 (#1433). Es gehört dem **Prozess** und nicht
	// dem Profil: wer als Gast in einem fremden Profil arbeitet, läse darin
	// die Spuren seines Gastgebers. Der Rest der Diagnose — Fassung,
	// Rechnerart, Ort, Schemastand, Paketstände — bleibt stehen; er ist eine
	// Auskunft über die Installation und über niemanden.
	if a.darfDasProtokollSehen(ctx) {
		text.WriteString("\n" + protokollende(pfad, protokollfehler, texte))
	} else {
		text.WriteString("\n" + texte.DiagnoseProtokollVerdeckt + "\n")
	}

	// **Geschwärzt wird der ganze Text und nicht jede Zeile einzeln.** Die
	// Kopfzeilen entstehen hier und gehen durch keinen Protokollschreiber; das
	// Protokollende ist beim Schreiben schon geschwärzt, und eine zweite
	// Schwärzung derselben Zeichenkette ändert nichts.
	fertig := schwaerze(kopf) + redigiere(text.String())

	a.letzteDiagnose = fertig

	return DiagnoseAntwort{Text: fertig}
}

// schemastand sagt, welche Migrationen verbucht sind.
//
// **Die Zahl und die jüngste.** Eine vollständige Liste wären zwanzig Zeilen,
// die niemand liest; die Zahl sagt, ob eine fehlt, und der Name sagt, wo man
// steht.
//
// **Ein Lesefehler ist eine Auskunft.** Wer eine Diagnose braucht, hat oft
// genau hier ein Problem — ein Weg, der abbräche, ließe den Fall unberichtet,
// für den es ihn gibt (§8).
func (a *App) schemastand(ctx context.Context, texte Dialogtexte) string {
	if a.schemaquelle == nil {
		return texte.DiagnoseKeinBestand
	}

	angewendet, err := a.schemaquelle.AngewendeteMigrationen(ctx)
	if err != nil {
		return texte.DiagnoseNichtLesbar + ": " + err.Error()
	}

	if len(angewendet) == 0 {
		return texte.DiagnoseKeineSchema
	}

	return fmt.Sprintf(texte.DiagnoseMigrationen,
		len(angewendet), angewendet[len(angewendet)-1])
}

// paketstaende zählt die Vorlagenpakete samt ihren Fassungen auf.
//
// **Name und Fassung, sonst nichts.** Pfade stehen schon oben, und eine Zeile
// je Paket mit Herkunft und Prüfsumme wären bei 331 Paketen tausend Zeilen.
//
// **Und begrenzt.** Wer 331 Pakete hat, bekommt sie alle — das ist die
// Auskunft. Wer dreitausend hätte, bekäme eine Datei, die niemand öffnet;
// deshalb sagt der Kopf, wie viele es sind, und die Liste bleibt vollständig
// nur bis [Paketzeilen].
func (a *App) paketstaende(ctx context.Context, texte Dialogtexte) string {
	if a.pakete == nil {
		return texte.DiagnosePakete + ": " + texte.DiagnoseKeineVerwaltung + "\n"
	}

	antwort := a.Pakete(ctx)
	if antwort.Fehler != nil {
		return texte.DiagnosePakete + ": " + texte.DiagnoseNichtLesbar +
			" — " + antwort.Fehler.Meldung + "\n"
	}

	var gebaut strings.Builder

	fmt.Fprintf(&gebaut, "%s: %d\n", texte.DiagnosePakete, len(antwort.Pakete))

	for i, paket := range antwort.Pakete {
		if i >= Paketzeilen {
			fmt.Fprintf(&gebaut, "  "+texte.DiagnoseWeitere+"\n",
				len(antwort.Pakete)-Paketzeilen)

			break
		}

		fmt.Fprintf(&gebaut, "  %s %s\n", paket.Name, paket.Version)
	}

	return gebaut.String()
}

// DiagnoseSichern schreibt die Diagnose in eine Datei, die jemand wählt.
//
// # Warum ein Dialog und kein fester Ort
//
// **Weil die Datei dem gehört, der sie weitergibt.** Sie an einen von der
// Anwendung gewählten Ort zu legen wäre bequemer und hieße, dass jemand sie
// dort erst suchen muss — und dass eine Anwendung ohne Zutun eine Datei mit
// Protokollzeilen anlegt.
//
// **Ein leerer Pfad heißt: abgebrochen.** Dieselbe Gestalt wie bei jeder
// anderen Dateiwahl dieses Pakets; ein Abbruch ist kein Fehler.
//
// **Kein Wächter auf einen Fehler der Diagnose.** [App.Diagnose] setzt keinen
// — sie kann keinen setzen: ein unauffindbarer Bestand und ein fehlendes
// Protokoll stehen als **Auskunft** im Text, und genau dann ist sie am
// nötigsten. Ein Zweig, den kein Aufruf betritt, ist eine Behauptung darüber,
// was geschieht (§3).
func (a *App) DiagnoseSichern(ctx context.Context) SchreibAntwort {
	ctx = a.imProfil(ctx)

	// **Gesichert wird, was gezeigt wurde** — und nichts anderes.
	//
	// Der erste Anlauf rief [App.Diagnose] ein zweites Mal. Zwischen Anzeigen
	// und Sichern wächst das Protokoll weiter; die Datei trug damit Zeilen,
	// die niemand gesehen hat, und die Zustimmung galt einem anderen Text als
	// dem gespeicherten. Angestrichen vom Zweitblick über Phase 3.5 als
	// Befund 5.
	//
	// **Ohne Vorschau gibt es nichts zu sichern.** Das ist kein Sonderfall des
	// Fensters — dort erscheint der Knopf erst danach —, sondern die Zusage
	// selbst: „was du siehst, gibst du weiter" hat ohne ein Gesehenes keinen
	// Inhalt.
	if a.letzteDiagnose == "" {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Diagnose ist noch nicht angezeigt worden",
			domain.ErrValidation))}
	}

	texte := DialogtexteFuer(a.sprache)

	ziel, fehler := a.waehleDatei(ctx, Dateiauswahl{
		Speichern: true,
		Ablage:    AblageDiagnose,
		Titel:     texte.DiagnoseSpeichern,
		Vorschlag: texte.DiagnoseVorschlag,
		Filter:    texte.DiagnoseFilter,
		Muster:    "*.txt",
	})
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if ziel == "" {
		return SchreibAntwort{}
	}

	// **Modus 0600, auch über einer vorhandenen Datei** (SEC-20, #2133): die
	// Diagnose nennt Pfade und Einstellungen dieses Menschen; weitergeben
	// wird er sie ausdrücklich, nicht über die Rechte der Datei davor.
	// **Über eine Nachbardatei** (Befund 2 des Zweitblicks über Phase 4.13):
	// jemand hat im Dialog auf diese Stelle gezeigt, und dort lag vielleicht
	// schon etwas.
	if err := service.SchreibeAtomarVertraulich(ziel, []byte(a.letzteDiagnose)); err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, domain.MitGrund(
			domain.ErrValidation, err,
			"die Diagnose ließ sich nicht schreiben — wähle einen anderen Ort"))}
	}

	return SchreibAntwort{}
}

// waisenstand gibt die Zeile über verwaiste Zeilen — oder sagt, warum nicht.
//
// **Ein Fehler hält die Diagnose nicht auf** (§8). Sie ist der Bericht für
// jemanden, dessen Anwendung klemmt; ein Weg, der daran abbräche, ließe genau
// den Fall unberichtet, für den es ihn gibt.
func (a *App) waisenstand(ctx context.Context, texte Dialogtexte) string {
	if a.schemaquelle == nil {
		return texte.DiagnoseNichtLesbar
	}

	stand, err := a.schemaquelle.Waisen(ctx)
	if err != nil {
		return texte.DiagnoseNichtLesbar + ": " + err.Error()
	}

	return stand
}

// darfDasProtokollSehen sagt, ob die Rolle bis zu den Protokollzeilen reicht.
//
// **Nur der Verwalter.** Er ist der Eigner des gezeigten Profils oder wurde
// ausdrücklich dazu gemacht (0070); ein Leser oder Mitarbeiter ist zu Gast.
//
// **Ohne Autorisierung ist die Antwort ja**, und das ist keine offene Tür: die
// Gefahr, gegen die dieser Riegel steht, ist ein Mensch in einem **fremden**
// Profil — und die gibt es erst, wenn es Profile und Freigaben gibt. Der Fall
// ohne Autorisierung ist die Anwendung vor dem Start.
//
// **Ein Fehler beim Fragen verdeckt.** Wer nicht sagen kann, wer da ist, weiß
// auch nicht, ob es der Eigner ist.
func (a *App) darfDasProtokollSehen(ctx context.Context) bool {
	if a.autorisierung == nil {
		return true
	}

	akteur, fehler := a.autorisierung.Akteur(ctx)
	if fehler != nil {
		return false
	}

	return akteur.Rolle == domain.RolleVerwalter
}

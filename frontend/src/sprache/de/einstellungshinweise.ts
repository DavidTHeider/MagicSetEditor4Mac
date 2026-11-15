// **Die Hinweise der Einstellungsseite** — je Einstellung der Satz, der sagt,
// was sie bewirkt.
//
// **Aus `seiteneinstellungen.ts` herausgelöst, als die Datei über 400 Zeilen
// wuchs** (§2) — dieselbe Gestalt wie `seitenkarten.ts` aus `seiten.ts`. Der
// Schnitt folgt der Frage: dort steht, **wie** eine Einstellung heißt und
// welche Werte sie hat, hier steht, **was sie bewirkt**.
//
// **Ein Katalog bleibt es trotzdem** — er wird drüben eingesetzt und kommt als
// **einer** an; an der Zahl der Kataloge hängt der Vertragstest gegen die
// Go-Seite (§9).

import { handelshilfe } from './einstellungenhandel'
import { gestaltungshinweise } from './einstellungengestaltung'
import { netzhilfe } from './einstellungennetz'

import { protokollhinweise } from './einstellungenprotokoll'

export const einstellungshinweise = {
  ...protokollhinweise,
  'sprache.eigene':
    'Welche eigene Übersetzung das Fenster spricht — leer heißt keine. Gibt es sie nicht mehr, gilt wieder die gewählte Sprache.',
  sprache:
    'Gilt für das ganze Fenster und für die Menüleiste. Was aus einem MSE-Paket kommt — Feldnamen, Spielnamen —, bleibt, wie es in der Datei steht. Deutsch ist die Ausgangssprache; alle anderen Sprachen sind maschinell übersetzt und noch nicht redaktionell geprüft.',
  darstellung:
    'Wie Zahlen, Beträge und Daten aussehen — etwa **de-CH** oder **en-GB**. Leer heißt: wie das System. Die Sprache der Texte bleibt davon unberührt.',
  zeitzone:
    'In welcher Zone Zeitpunkte erscheinen — etwa **Europe/Berlin**. „Wie das System“ ist die Zone des Rechners. Gespeichert wird sie je Profil; bei einem Austausch reist sie nicht mit.',
  'filter.gespeichert':
    'Die gespeicherten Filter aller Sichten als Liste in JSON — Name, Sicht, Marken und ob **und** oder **oder** gilt. Angelegt werden sie im Navigator ihrer Sicht; sie gehören dem Profil und reisen mit der Ausfuhr der Einstellungen.',
  'kuerzel.ziffern':
    'Welche Seite eine Ziffer mit ⌘ öffnet. Leer heißt: die Bereiche der Seitenleiste. Eingestellt wird sie an der Kürzeltafel unter „Allgemein“.',
  erscheinungsbild:
    'Hell, dunkel, oder was der Rechner sagt. Die Wahl gilt nur für diese Anwendung.',
  schriftgroesse:
    'Skaliert die ganze Oberfläche, nicht nur den Text — Abstände und Ränder wachsen mit. Die Kartenvorschau bleibt, wie sie ist: sie zeigt, was gedruckt würde.',
  aktualisierung:
    'Vorgabe ist **aus**: dann geht nichts ins Netz und es wird nichts gefragt. Bei **melden** sieht die Anwendung beim Start nach, ob eine neuere Fassung vorliegt, und sagt es — mehr nicht. Bei **herunterladen** bietet sie zusätzlich einen Knopf an, der die Fassung ins Downloadverzeichnis holt und ihre Prüfsumme vergleicht; installiert wird nichts von selbst.',
  ersteschritte:
    'Die drei Schritte zur ersten Karte, als Fenster beim Start. Derselbe Schalter steht im Fenster selbst — wer es dort abschaltet, findet es hier wieder.',
  seitenleiste:
    'Die Leiste links mit den Flächen der Anwendung. Eingeklappt gehört ihre Breite den Arbeitsplätzen; ⌥⌘S, der Knopf am Rand und der Menüeintrag unter **Darstellung** klappen sie ein und aus. Gespeichert wird sie je Profil; bei einem Austausch reist sie nicht mit.',
  leistenlage:
    'Links steht sie senkrecht neben dem Inhalt, oben als Zeile darüber. Dieselben Einträge, Kürzel und Pfeiltasten — nur quer. Gespeichert wird sie je Profil; bei einem Austausch reist sie nicht mit.',
  fensterstart:
    'Vorgabe ist **Bildschirm füllen** — das Fenster nimmt den Platz, den es hat, und **kein Vollbild**: das nähme die Menüleiste mit. **Normalgröße** öffnet in der eingebauten Größe, **Wie zuletzt** so groß, wie du es zuletzt gelassen hast. Gespeichert wird sie je Profil; bei einem Austausch reist sie nicht mit.',
  fensterbreite:
    'Was beim letzten Schließen gemessen wurde — **nur für „Wie zuletzt“**. Leer heißt: es gab noch kein Zuletzt, und dann öffnet das Fenster in seiner eingebauten Größe.',
  fensterhoehe:
    'Dasselbe für die Höhe. Beide werden beim Schließen geschrieben; ein Maß außerhalb von 700 bis 8.192 Punkten wird nicht gemerkt — ein Fenster, das auf den nächsten Bildschirm nicht passt, wäre nicht mehr zu fassen.',
  feldtiefe:
    'Wie viel das Kartenformular zeigt. **Normal** sind die Felder, die du ausfüllst; **Erweitert** nimmt dazu, was ein Skript des Spielpakets rechnet; **Experte** dazu, was das Spielpaket gar nicht erklärt — `stylesheet` und seinesgleichen. Die drei liegen ineinander: tiefer heißt mehr, nie anderes. Gilt je Profil.',
  schriftart:
    'Die Schrift der Überschriften und Setnamen. Der Fließtext bleibt die Systemschrift — eine Zierschrift über einem Absatz liest sich schwer. Alle vier liegen auf jedem Mac; es wird nichts geladen.',
  'barrierefreiheit.bewegung':
    'Die Übersteuerung geht **nur in eine Richtung**: sie schaltet Bewegung ab, nie zu. Wer im System **weniger Bewegung** eingestellt hat, bekommt sie hier ohnehin nicht — eine Anwendung, die das aufhöbe, arbeitete gegen ihn.',
  'barrierefreiheit.kontrast':
    'Dieselbe Richtung: **hoch** nimmt die Faser des Untergrunds und hebt die Farben, bis jeder Text 7 : 1 hält. Wer im System mehr Kontrast verlangt, bekommt dasselbe.',
  'sicherheit.faktorbeimwechsel':
    'Die **einzige** Übersteuerung, die lockert — und nur beim **Wechsel** des Profils. Am Anmelden ändert sie nichts: dort steht der zweite Faktor für das ein, wofür es ihn gibt. Auf einem Rechner mit einem Menschen ist ein Einmalkennwort bei jedem Wechsel Reibung ohne Gegenwert.',
  'vorgabe.zustand':
    'Womit eine neue Zeile in der Sammlung anfängt. **Keine Vorgabe** heißt: das Feld bleibt leer, wie es immer war.',
  'vorgabe.waehrung':
    'Womit das Währungsfeld einer neuen Zeile anfängt, etwa **EUR**. Umgerechnet wird nichts.',
  'vorgabe.einzelspiel':
    'Ob ein neues Turnier Matches aus **einem** Spiel führt. Es entscheidet die Anzeige, nicht die Rechnung — Spielquote und Gegnerspielquote werden trotzdem geführt.',
  'vorgabe.hochladepfad':
    'Der Ordner, in dem ein Dialog zum Öffnen aufgeht — für Bilder, MTG JSON und alles andere, was von der Platte kommt. Ein vollständiger Ort oder einer mit **~/** vorn; leer heißt: dort, wo das Betriebssystem ihn öffnet.',
  'vorgabe.ausfuhrpfad':
    'Wo ein Sicherndialog aufgeht, wenn die Handlung keinen eigenen Ort hat. Fehlt der Ordner beim Sichern — etwa ein abgezogenes Laufwerk —, sagt die Anwendung es, statt woanders zu speichern.',
  'vorgabe.sicherungspfad':
    'Wo der Dialog für eine Sicherung des Bestands aufgeht. Ein fehlender Ordner hält die Sicherung mit einer Meldung an.',
  'vorgabe.diagnosepfad':
    'Wo Diagnosetext und Diagnosepaket gesichert werden. Ein fehlender Ordner hält mit einer Meldung an.',
  'bildausfuhr.ordner':
    'Wohin die Bildausgabe zuletzt geschrieben hat. Die Seite „Bilder“ schlägt ihn vor und trägt jeden neu gewählten Ordner hier ein — einer für alle Sets, wie in MSE. Leer heißt: der Ordnerdialog geht beim Ordner fürs Hinausgeben auf.',
  'ausgabe.ziele':
    'Welches Ziel „Ausgeben“ je Set zuletzt zeigte. Eine Liste in JSON, die jüngste Wahl hinten, höchstens 200; gesetzt wird sie bei der Wahl unter „Wohin?“. Leer heißt: nichts ist vorgewählt.',
  'dienst.anthropic.adresse':
    'Wohin gefragt wird. Leer heißt: die Schnittstelle von Anthropic selbst. Eine andere Adresse ist ein **Gateway** — ein Dienst, der die Anfrage weiterreicht; sein Schlüssel wird getrennt hinterlegt.',
  'dienst.openai.adresse':
    'Dasselbe für OpenAI. Leer heißt: `https://api.openai.com`. Wer hier etwas einträgt, spricht **nicht** mehr mit OpenAI, sondern mit dem, was unter dieser Adresse antwortet.',
  'dienst.xai.adresse': 'Dasselbe für xAI. Leer heißt: `https://api.x.ai`.',
  ...netzhilfe,
  ...handelshilfe,
  ...gestaltungshinweise,
  'konto.schau':
    'Ob dieses Programm dein **eigenes** Konto bei einem Marktplatz lesen darf — deine Angebote und deine Bestellungen. **Aus** heißt: nicht einmal nachsehen. Der Marktplatz einzuschalten reicht dafür nicht: jener beantwortet, was **fremde** Verkäufer verlangen, dies hier, was **bei dir** liegt. Geschrieben wird dort in keinem Fall — kein Angebot, kein Warenkorb, kein Kauf.',
  'ki.laufzeit':
    'Ob überhaupt eine KI-Laufzeit gefragt wird. **Aus** heißt: kein einziger Aufruf geht hinaus — auch nicht auf diesen Rechner. Installiert oder gestartet wird von hier nichts.',
  'ki.geraetemodell':
    'Ob das Gerätemodell von macOS für kleine Textaufgaben gefragt wird — Kartenname, Flavortext. **Automatisch** heißt: wenn es auf diesem Mac aktiviert ist (Apple Intelligence an, Lizenz von `fm` angenommen), ist es eingebunden und wird zuerst gefragt; die Daten verlassen den Rechner nicht. **Aus** gilt, bis du es zurückstellst; für eine Sitzung abschalten geht unter **KI**.',
  'ki.vorrangmodell':
    'Welches Ollama-Modell ein Vorschlag mit **Vorrang** fragt — nach dem Gerätemodell. **Leer** heißt: das erste, das in fester Form antworten kann, und Ollama nennt das zuletzt gezogene zuerst. Ein gewähltes Modell, das Ollama nicht mehr nennt, wird gemeldet und nicht still durch ein anderes ersetzt.',
  'ki.adresse':
    'Wo die Laufzeit lauscht. Leer heißt: die Vorgabeadresse auf diesem Rechner. Eine Adresse, die **nicht** auf diesem Rechner liegt, stuft jedes Modell dahinter als nicht-lokal ein — unter „nur lokal“ wird dann gar nicht gefragt.',
  'anbieter.github.kennung':
    'Die `client_id` der OAuth-Anwendung, die du bei GitHub angelegt hast. Leer heißt: über GitHub wird nicht angeboten, sich anzumelden.',
  'anbieter.google.kennung':
    'Die `client_id` der OAuth-Anwendung, die du bei Google angelegt hast. Sie endet meist auf `apps.googleusercontent.com`.',
  'anbieter.microsoft.kennung':
    'Die Anwendungskennung aus dem Microsoft-Entra-Portal — eine UUID. Microsoft verlangt **kein** Geheimnis für diesen Weg.',
  'anbieter.x.kennung':
    'Die `client_id` der Anwendung, die du im Entwicklerportal von X angelegt hast.',
}

import { dienstkatalog } from './einstellungenkatalog'
import { einstellungsbelegung } from './einstellungsbelegung'
import { einstellungsgruppennamen } from './einstellungsgruppennamen'
import {
  schalternamen,
  schalterwerte,
  seitenschalter,
} from './einstellungenschalter'

// **Die Texte der Einstellungsseite** — ausgelagert für Phase 4.8
// (#1321).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als
// einer ankommt** (§9) — dieselbe Gestalt wie `seitenkarten.ts`. Mit den
// Dienstschaltern ging `seiten.ts` über die Zeilengrenze aus §2.
import { ki } from './seitenki'
import { netzbereich, netznamen } from './einstellungennetz'
import { einstellungshinweise } from './einstellungshinweise'
import { handelsnamen, handelswerte } from './einstellungenhandel'
import {
  gestaltungsbeiwerk,
  gestaltungsnamen,
  gestaltungswerte,
} from './einstellungengestaltung'
import { zugaenge } from './einstellungenzugaenge'
import { ansichtsnamen, ansichtswerte } from './einstellungenansicht'
import { vorgabennamen } from './einstellungenvorgaben'
import { protokollnamen, protokollwerte } from './einstellungenprotokoll'
import { austausch } from './einstellungenaustausch'
import { eigenuebersetzung } from './eigenuebersetzung'

export const einstellungen = {
  ...einstellungsgruppennamen,
  ...seitenschalter,

  titel: 'Einstellungen',

  // Die Namen der Einstellungen. Die Schlüssel kommen aus
  // `internal/domain/einstellung.go`; ein Wächter hält beide Seiten zusammen.
  namen: {
    ...schalternamen,
    sprache: 'Sprache',
    'sprache.eigene': 'Eigene Übersetzung',
    darstellung: 'Regionale Darstellung',
    zeitzone: 'Zeitzone',
    'filter.gespeichert': 'Gespeicherte Filter',
    'kuerzel.ziffern': 'Belegung von ⌘0 bis ⌘9',
    erscheinungsbild: 'Thema',
    schriftgroesse: 'Schriftgröße',
    aktualisierung: 'Nach neuen Fassungen sehen',
    ersteschritte: 'Einführung beim Start',
    seitenleiste: 'Seitenleiste',
    leistenlage: 'Lage der Leiste',
    fensterstart: 'Fenster beim Start',
    fensterbreite: 'Zuletzt gemerkte Breite',
    fensterhoehe: 'Zuletzt gemerkte Höhe',
    schriftart: 'Anzeigeschrift',

    // **Wie viel vom Kartenformular dasteht** — UX-05.AK-01
    // (#1677).
    feldtiefe: 'Kartenformular',
    // **Die Anordnung des Tisches** — PLAY-04.AK-03
    // (#1712).
    'tisch.autoanordnung': 'Karten auf dem Tisch selbst ordnen',

    // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351).
    'barrierefreiheit.bewegung': 'Bewegung',
    'barrierefreiheit.kontrast': 'Kontrast',
    'sicherheit.faktorbeimwechsel': 'Zweiter Faktor beim Profilwechsel',

    ...vorgabennamen,
    ...protokollnamen,
    // **Die Adresse eines Cloudanbieters** — Werkstück 3 von A18 (#1533),
    // beschriftet mit #1540.
    // Hier stand der rohe Schlüssel im Fenster.
    'dienst.anthropic.adresse': 'Adresse von Anthropic',
    'dienst.openai.adresse': 'Adresse von OpenAI',
    'dienst.xai.adresse': 'Adresse von xAI',
    // **Die lokale KI-Laufzeit** — A7, #1446.
    ...netznamen,
    ...handelsnamen,
    ...gestaltungsnamen,
    ...ansichtsnamen,
    'konto.schau': 'Eigenes Marktkonto lesen',
    'ki.laufzeit': 'Lokale KI-Laufzeit',
    'ki.geraetemodell': 'Apples Gerätemodell',
    'ki.vorrangmodell': 'Ollama-Modell für den Vorrang',
    'ki.adresse': 'Adresse der KI-Laufzeit',
    // **Die Kennung je Anbieter** — Nacharbeit zu Werkstück 4 von Phase 4.12
    // (#1351). **Der Anbietername ausgeschrieben**, damit die Zeile ohne den
    // Schlüssel lesbar ist.
    'anbieter.github.kennung': 'Kennung bei GitHub',
    'anbieter.google.kennung': 'Kennung bei Google',
    'anbieter.microsoft.kennung': 'Kennung bei Microsoft',
    'anbieter.x.kennung': 'Kennung bei X',

    // **Der Arbeitsplatz und die zwei Schranken** — Werkstück O3 von
    // Phase 5.2 (#1427).
    'arbeitsplatz.navigator': 'Linke Spalte',
    'arbeitsplatz.inspektor': 'Rechte Spalte',
    'arbeitsplatz.navigatorbreite': 'Breite der linken Spalte',
    'arbeitsplatz.inspektorbreite': 'Breite der rechten Spalte',
    'richtlinie.nurlokal': 'Nur lokal arbeiten',
    'richtlinie.kostengrenze':
      'Kostenobergrenze in der kleinsten Einheit (bei EUR: Cent)',
  },

  // **Herkunft, Vorgabe und Suche** — Werkstück U11 von Phase 5.1
  // (#1422).
  herkunftMarke: 'von {herkunft}',
  wirksamKurz: 'Es gilt „{wert}“.',
  wirksamOffenKurz: 'Es gilt, was der Rechner vorgibt.',
  wirksamGehalten: 'Es gilt „{wert}“ — eine Richtlinie hält es.',
  herkunftZeile: 'Gesetzt von {herkunft}, Ebene {ebene}.',
  herkunftNamen: {
    profil: 'diesem Profil',
    nutzer: 'dir',
    geraet: 'diesem Gerät',
    vorgabe: 'niemandem (Vorgabe)',
  },
  ebenenNamen: {
    profil: 'Profil',
    nutzer: 'Nutzer',
    geraet: 'Gerät',
    arbeitsplatz: 'Arbeitsplatz',
  },
  // **Die Ebenen erklärt, an der Suche** — UX-04.AK-03 (#1676). Die Regel
  // steht in `internal/domain/geltung.go`, die Ebenen in
  // `einstellungsebene.go`; **Verwalterrechte verlangt keine Einstellung**,
  // nachgelesen am 2026-09-24 am Schreibweg (`App.EinstellungSetzen`).
  ebenenErklaerung:
    'Jede Einstellung gilt auf einer Ebene: Gerät, Nutzer, Profil oder Arbeitsplatz. **Gerät** beschreibt diesen Rechner — einen Pfad, ein Fenstermaß — und reist bei einem Austausch nicht mit. **Nutzer** gilt für dich in jedem deiner Profile. **Profil** gilt nur hier und übersteuert, was du als Nutzer gesetzt hast. **Arbeitsplatz** ist, wie die Spalten dieses Fensters stehen. Wirksam ist, was im Profil steht, sonst was du als Nutzer gesetzt hast, sonst die Vorgabe — bei einer Richtlinie der strengste Wert. Verwalterrechte verlangt keine Einstellung.',
  // **Was nur Fachleute brauchen, eingeklappt** — UX-04.AK-04, nach dem
  // Artboard *Einstellungen*.
  fachleute: 'Was Fachleute hier sonst noch einstellen',
  fachleuteSatz:
    'Fristen, Erreichbarkeit über fremde Netze, der Weg hinaus, Dienste für andere Programme. Aufklappen ändert nichts — es zeigt nur mehr.',
  zuruecksetzen: 'Zurücksetzen auf „{vorgabe}“',
  erbeMitWert: '{erbe} ({wert})',
  zuruecksetzenVon: '{name} auf die Vorgabe zurücksetzen',
  suche: 'Einstellung suchen',
  suchePlatzhalter: 'Name, Schlüssel oder Wert',
  sucheTreffer: '{zahl} Treffer',
  sucheKeine: 'Keine Einstellung passt zu „{frage}“.',
  zumGebiet: 'Im Gebiet „{gebiet}“ zeigen',
  hinweise: einstellungshinweise,

  // **Unter den Einstellungen der Gestaltung** — UX-09 (#1681).
  gestaltung: gestaltungsbeiwerk,

  // Die Werte, je Einstellung. Der Schlüssel ist `<einstellung>.<wert>`.
  werte: {
    ...schalterwerte,
    ...handelswerte,
    ...protokollwerte,
    ...gestaltungswerte,
    ...ansichtswerte,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **Die Kontoschau** — T11, #1476. Dieselben Worte wie bei der
    // KI-Laufzeit darüber: zwei Schreibweisen für denselben Zustand
    // lernt niemand gern.
    'konto.schau.aus': 'Aus',
    'konto.schau.an': 'An',
    'ki.laufzeit.aus': 'Aus',
    'ki.laufzeit.an': 'An',
    'ki.geraetemodell.automatisch': 'Automatisch',
    'ki.geraetemodell.aus': 'Aus',
    // **Was ein Profil vom Menschen erbt** — Werkstück 9 von Phase 4.12
    // (#1351), Punkt 9. Es ist die Vorgabe: ein frisches Profil behauptet
    // nichts über den Menschen.
    'sprache.erben': 'Wie beim Nutzer',
    'erscheinungsbild.erben': 'Wie beim Nutzer',
    'schriftgroesse.erben': 'Wie beim Nutzer',
    'schriftart.erben': 'Wie beim Nutzer',
    'sprache.de': 'Deutsch',
    'sprache.en': 'English',
    'sprache.es': 'Español',
    'sprache.fr': 'Français',
    'sprache.it': 'Italiano',
    'sprache.ja': '日本語',
    'sprache.ru': 'Русский',
    'sprache.zh': '简体中文',
    'erscheinungsbild.system': 'Wie das System',
    'erscheinungsbild.hell': 'Hell',
    'erscheinungsbild.dunkel': 'Dunkel',
    'schriftgroesse.klein': 'Klein',
    'schriftgroesse.normal': 'Normal',
    'schriftgroesse.gross': 'Groß',
    'schriftgroesse.sehrgross': 'Sehr groß',
    'aktualisierung.aus': 'Aus',
    'aktualisierung.melden': 'Nur melden',
    'aktualisierung.holen': 'Melden und herunterladen',
    'ersteschritte.an': 'An',
    'ersteschritte.aus': 'Aus',
    'seitenleiste.sichtbar': 'Sichtbar',
    'seitenleiste.verborgen': 'Verborgen',
    'leistenlage.senkrecht': 'Links',
    'leistenlage.waagerecht': 'Oben',
    'fensterstart.maximiert': 'Bildschirm füllen',
    'fensterstart.normal': 'Normalgröße',
    'fensterstart.zuletzt': 'Wie zuletzt',

    'feldtiefe.normal': 'Normal',
    'feldtiefe.erweitert': 'Erweitert',
    'feldtiefe.experte': 'Experte',
    'tisch.autoanordnung.an': 'Ja — der Tisch legt sie',
    'tisch.autoanordnung.aus': 'Nein — was du hinlegst, bleibt liegen',

    // **Wie der Tisch beschriftet ist** — seit dem 2026-09-24. „Aus“ nimmt
    // Namen und Köpfe weg; die Anzahlen bleiben, denn sie stehen sonst
    // nirgends.
    'tisch.beschriftung': 'Beschriftung auf dem Tisch',
    'tisch.beschriftung.aus': 'Aus — nur die Anzahlen',
    'tisch.beschriftung.klein': 'Klein',
    'tisch.beschriftung.normal': 'Mittel',
    'tisch.beschriftung.gross': 'Groß',
    'schriftart.pergament': 'Pergament — ruhig und lesbar',
    'schriftart.fantasy': 'Fantasy — verspielt',
    'schriftart.antike': 'Antike — römische Kapitalis',
    'schriftart.kalligrafie': 'Kalligrafie — von Hand geschrieben',
    'barrierefreiheit.bewegung.system': 'Wie das System',
    'barrierefreiheit.bewegung.reduziert': 'Immer wenig',
    'barrierefreiheit.kontrast.system': 'Wie das System',
    'barrierefreiheit.kontrast.hoch': 'Immer hoch',
    'sicherheit.faktorbeimwechsel.verlangen': 'Auch beim Wechseln',
    'sicherheit.faktorbeimwechsel.nurAnmelden': 'Nur beim Anmelden',
    'vorgabe.zustand.ohne': 'Keine Vorgabe',
    'vorgabe.zustand.mint': 'Mint',
    'vorgabe.zustand.nearMint': 'Near Mint',
    'vorgabe.zustand.excellent': 'Excellent',
    'vorgabe.zustand.good': 'Good',
    'vorgabe.zustand.lightPlayed': 'Light Played',
    'vorgabe.zustand.played': 'Played',
    'vorgabe.zustand.poor': 'Poor',
    'vorgabe.einzelspiel.aus': 'Aus',
    'vorgabe.einzelspiel.an': 'An',

    // **Der Arbeitsplatz und die Schranke** — Werkstück O3 von Phase 5.2
    // (#1427). Die zwei Breiten und die Kostengrenze sind Zahlen und haben
    // keine Werteliste; sie brauchen hier nichts.
    'arbeitsplatz.navigator.erben': 'Wie beim Nutzer',
    'arbeitsplatz.navigatorbreite.erben': 'Wie beim Nutzer',
    'arbeitsplatz.inspektorbreite.erben': 'Wie beim Nutzer',
    'arbeitsplatz.navigator.sichtbar': 'Sichtbar',
    'arbeitsplatz.navigator.verborgen': 'Verborgen',
    'arbeitsplatz.inspektor.erben': 'Wie beim Nutzer',
    'arbeitsplatz.inspektor.sichtbar': 'Sichtbar',
    'arbeitsplatz.inspektor.verborgen': 'Verborgen',
    'richtlinie.nurlokal.aus': 'Aus',
    'richtlinie.nurlokal.an': 'An',
  },

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351).
  // **Womit ein neuer Datensatz anfängt** — Werkstück 9 von Phase 4.12
  // (#1351), Punkte 2, 3 und 8.
  // **Wo eine Anmeldung überhaupt herkommt** — Nacharbeit zu Werkstück 4 von
  // Phase 4.12 (#1351).
  //
  // **Der Hinweis sagt, was BEIM ANBIETER zu tun ist**, nicht nur, was hier
  // einzutragen ist: ohne den ersten Schritt gibt es nichts einzutragen — und
  // die Nutzerseite verwies bisher auf eine Stelle, an der gar nichts stand.
  anmeldung: 'Anmeldung über einen Anbieter',
  anmeldungHinweis:
    'Diese Anwendung liefert **keine Zugangsdaten mit**. Wer sich über GitHub, Google, Microsoft oder X anmelden will, legt beim Anbieter selbst eine **OAuth-Anwendung** an und trägt deren `client_id` hier ein. Als Rückrufadresse gibt man dort die Adresse an, die beim Anmelden im Browser steht — sie zeigt auf diesen Rechner und gilt nur, solange die Anmeldung läuft.',

  // **Und wohin das Geheimnis gehört** — Befund 3 des Zweitblicks über Phase
  // 4.12 (#1351). Hier stand „kommt in den Schlüsselbund“, ohne zu sagen,
  // **welcher Eintrag** — und GitHub verlangt eines, also war die Anmeldung
  // darüber nicht einzurichten. Das Feld dafür baut #1401.
  anmeldungGeheimnis:
    'GitHub verlangt zusätzlich ein `client_secret` — Google stellt es frei, Microsoft verbietet es für Programme wie dieses. Es kommt **in den Schlüsselbund** und nicht hierher: den Bestand gibt man weiter, den Schlüsselbund nicht. Anzulegen in der **Schlüsselbundverwaltung** als **Passwort für Programme**, unter dem Namen `MagicSetEditor4Mac-Anmeldung` und dem Account `github`. Auf der Kommandozeile: `security add-generic-password -s MagicSetEditor4Mac-Anmeldung -a github -w`.',

  // **Das Geheimnis eines Anmeldeanbieters** — #1401, Befund 3 des
  // Zweitblicks über Phase 4.12.
  geheimnis: 'Geheimnis',
  geheimnisPflicht: 'wird verlangt und fehlt',
  geheimnisVerboten:
    'verlangt kein Geheimnis und weist eines ab — hier genügt die Kennung.',

  // **Die Anbieternamen, wie sie sich selbst schreiben.** Sie sind
  // Eigennamen und werden nicht übersetzt (§9).
  anbietername: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  },

  // **Wann eine Einstellung wirkt** — UX-04.AK-02
  // (#1676).
  // Die Auskunft kommt aus der Domäne; hier stehen nur die Sätze dazu.
  wirktNachNeustart: 'Wirkt erst nach einem Neustart.',

  // **Die Trefferanzeige** — UX-04.AK-03
  // (#1676).
  // Der Entwurf setzt die Zahl als Marke neben den Gruppennamen und
  // markiert das getroffene Feld an seinem Platz.
  trefferInGruppe: '{zahl} Treffer für „{frage}“',
  gesuchtMarke: 'gesucht',
  sucheSprung: 'Zum ersten Treffer springen',

  vorgaben: 'Vorgaben',

  untertitel: {
    allgemein: 'Sprache, Start, Kürzel',
    bereiche: 'Arbeitsbereiche, Funktionen, Reife',
    erscheinungsbild: 'Thema, Schrift, Fenster',
    nutzer: 'Anmeldung, Entsperren',
    fach: 'Karten, Sammlung, Spiel',
    ki: 'Anbieter, Modelle, Kosten',
    verbindungen: 'Netz, Gegenüber, CLI, MCP, REST',
    dienste: 'Kartendienste, Preise, Markt',
    daten: 'Bestand, Pfade, Sicherung',
    diagnose: 'Protokolle, Messung, Technik',
  },

  // **Der Verweis auf die Hilfe** — Werkstück 13 von Phase 4.12 (#1351).
  //
  // **Er steht hier und der Text dort**, weil er lang ist und hierher nicht
  // passt: diese Seite ist zum **Tun** da, die Hilfe zum Nachlesen. Aber wer
  // hier steht, soll den Weg finden — das Issue verlangt die Angabe in
  // **beidem**.

  // **Die fremden Kartendienste** — Phase 4.8
  // (#1321).
  dienste: 'Fremde Kartendienste',
  diensteHinweis:
    'Aus heißt: an diesen Dienst geht keine einzige Anfrage. Die Anwendung arbeitet ohne jeden Dienst weiter — nur was von draußen kommt, kommt dann nicht.',

  // **Die Überschrift über den Sprachmodellen** — UX-04.AK-04
  // (#1676).
  // Dieselben Schalter, ein anderer Satz darüber: „Fremde Kartendienste“
  // nennt nicht, was darunter steht.
  kiAnbieter: 'Sprachmodelle',
  kiAnbieterHinweis:
    'Wohin gefragt werden darf. Jeder Anbieter rechnet auf seinen Rechnern und sieht, was du schickst — deshalb ist jeder einzeln zu schalten, und keiner ist es von selbst.',
  // **Die beiden Werte eines Schalters.** Sie stehen hier und nicht unter
  // `werte`, weil sie für **jeden** Dienst dieselben sind — dort bräuchte
  // jeder Dienst zwei eigene Zeilen, die alle dasselbe sagen.
  dienstAus: 'Aus',
  dienstAn: 'An',
  dienstUngeprueft: 'Nie gegen den echten Dienst gelaufen',
  dienstUngeprueftHinweis:
    'Für diesen Dienst gibt es hier kein Konto. Geprüft ist unsere Seite der Grenze — die Anfrage, die Kopfzeilen, die Zuordnung. Ob der Dienst so antwortet, wie wir es annehmen, ist nicht nachgemessen.',
  // **Wer welche Frage beantwortet** — Werkstück 8 von Phase 4.8.
  fragen: 'Wer beantwortet was',
  fragenHinweis:
    'Können mehrere Dienste dieselbe Frage beantworten, entscheidet hier, welcher gefragt wird. **Es wird nicht ausgewichen**: schweigt der gewählte, bekommst du das gesagt — statt still eine Antwort aus einer anderen Quelle zu sehen.',
  'frage.suche': 'Kartensuche',

  ...dienstkatalog,

  dienstSchluesselFehlt: 'Kein Zugang hinterlegt',
  dienstSchluesselHinweis:
    'Dieser Dienst braucht Zugangsdaten. Sie liegen im Schlüsselbund von macOS und nicht in der Bestandsdatei — die gibt man weiter.',
  zugaenge: 'Zugangsdaten',
  zugaengeHinweis:
    'Sie liegen im Schlüsselbund von macOS und **nicht** in der Bestandsdatei — die gibt man weiter. Was hier steht, ist nur, **ob** eine Angabe hinterlegt ist; der Wert selbst kommt nie zurück.',
  // **Die Namen der Zugangsangaben.** Der Schlüssel ist `angabe.<name>`; die
  // Namen kommen aus `internal/domain/dienste.go`.
  //
  // **Sie stehen hier, weil sie Oberflächentext sind.** Bis zum 2026-09-02
  // setzte `Zugaenge.tsx` den rohen Bezeichner ins Fenster — neben „MTGJSON“
  // stand `marke: fehlt`, und in der englischen Fassung genauso.
  //
  // **Was sie benennen, ist nachgelesen** (0056, Tabelle der drei Dienste):
  // MTGJSON nimmt `authorization: Bearer <Token>` und gibt ihn nur an
  // Patreon-Unterstützer; Scrydex verlangt `X-Api-Key` **und** `X-Team-ID`.

  kuerzelUeberschrift: 'Tastenkürzel',
  ...einstellungsbelegung,
  kuerzelHinweis:
    'Sie gelten im ganzen Fenster. In einem Textfeld gehören ⌘Z und ⌘F dem Feld.',
  kuerzelZweiteReihe:
    'Die acht Kürzel mit ⌥ liegen auf den Tasten der obersten Reihe, von links. ' +
    'Auf einer Tastatur mit anderer Belegung — etwa AZERTY — ist dort ein anderes ' +
    'Zeichen aufgedruckt; die Taste bleibt dieselbe.',
  kuerzelTaste: 'Taste',
  kuerzelHandlung: 'Handlung',

  austausch,
  eigenuebersetzung,
  // **Das Kontingent an einem Anbieterkonto** — A5, #1444.
  kontingentTitel: 'Kontingent beim Anbieter',
  kontingentErklaerung:
    'Wie viele Kredite der Tarif hat und wie viele davon noch übrig sind. Die Abfrage geht ins Netz und kostet selbst einen Kredit.',
  kontingentHolen: 'Kontingent abfragen',
  kontingentHolenBei: 'Kontingent bei {dienst} abfragen',
  kontingentLaeuft: 'Wird abgefragt …',
  kontingentZahlen: '{rest} von {gesamt} Krediten sind übrig.',
  kontingentFrist:
    'Diese Zahl kann beim Anbieter bis zu {minuten} Minuten alt sein — sie ist keine Sperre. Was hier wirklich begrenzt, ist das Budget dieses Bestands.',
  kontingentUeberzug: 'Ein Kredit über den Tarif hinaus kostet {satz}.',
  ...ki,

  // **Die Zugangsdaten** (#1534) — eigene Datei, ein Katalog (§9).
  ...zugaenge,

  ...netzbereich,
}

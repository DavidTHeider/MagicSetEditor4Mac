// Die Texte des Bereichs **Spielen** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Ein eigener Baustein**, wie `seitenturniere.ts` und aus demselben Grund:
// `seiten.ts` steht an der Zeilengrenze aus §2.
//
// **Und die Partienliste steht noch einmal daneben**, in
// `spielenpartien.ts`: dieselbe Grenze, einen Baustein später. Es bleibt
// **ein** Katalog — er kommt unten als einer an, und der Vertragstest gegen
// die Go-Seite sieht die Naht nicht (§9).
import { abzugworte } from './spielenabzug'
import { zugworte } from './spielenzug'
import { menueworte } from './spielenmenue'
import { vorhangworte } from './spielenvorhang'
import { partienwoerter } from './spielenpartien'
import { regelwoerter } from './spielenregeln'
import { anfangworte } from './spielenanfang'
import { beginnworte } from './spielenbeginn'
import { gastworte } from './spielengast'
import { netzworte } from './spielennetz'
import { leihewoerter } from './spielenleihe'
import { leistenworte } from './spielenleiste'

export const spielen = {
  ...abzugworte,
  ...zugworte,
  ...menueworte,
  ...vorhangworte,
  ...partienwoerter,
  ...regelwoerter,
  ...anfangworte,
  ...beginnworte,
  ...gastworte,
  ...netzworte,
  ...leihewoerter,
  ...leistenworte,

  titel: 'Spielen',
  erklaerung:
    'Eine Partie ist ein Tisch mit eingefrorenen Decks. Was hier geschieht, steht als Strom von Ereignissen fest — jedes hängt am vorigen, und keines wird überschrieben. Wer eine Karte umlegt, sieht sofort, was die anderen davon sehen.',

  // Der Anfang.
  anfangTitel: 'Eine Partie anfangen',
  anfangErklaerung:
    'Jeder Sitz bringt ein Deck mit, und es wird eingefroren: wer sein Deck danach ändert, ändert diese Partie nicht. Das Spiel kommt aus dem ersten gewählten Deck.',
  leerDecksTitel: 'Noch kein Deck',
  leerDecks: 'Zum Spielen braucht es ein Deck. Lege eines unter Decks an.',
  sitz: 'Sitz',
  sitzName: 'Name am Tisch',
  sitzDeck: 'Deck',
  // **Eine angenommene Leihe am Sitz** — P1b (#2370), C-12.
  sitzLeihe: 'Geliehen von …{von} · {karten} Karten',
  deckWaehlen: 'Deck wählen',

  // **Ein Deck, das aus der Deckverwaltung mitgereist ist** — UX-03.AK-01
  // (#1675).
  anschlussWartet:
    '„{name}“ ist mitgereist und sitzt auf dem ersten Platz. Der zweite ist frei.',
  sitzHinzu: 'Sitz hinzufügen',
  sitzWeg: 'Diesen Sitz wegnehmen',
  sitzGruppe: 'Sitz {nummer}',
  spiel: 'Spiel',
  anfangen: 'Partie anfangen',

  // Der Tisch.
  tischTitel: 'Am Tisch',

  // Wie die Partie über dem Tisch heißt — die Brotkrume, UX-03.AK-02.
  partieGegen: 'Partie gegen {namen}',
  partieZwischen: 'Partie zwischen {namen}',
  partieVon: 'Partie von {namen}',
  fassung: 'Fassung',
  wirt: 'Wirt',
  wirtHinweis:
    'Der Wirt hält den Vollzustand dieser Partie. Er sieht dadurch mehr, als seine eigene Sicht zeigt — das steht hier, weil es vor dem Mitspielen zu wissen ist und nicht danach.',
  // **Was eine Zahl an einer Karte zählt** — PLAY-03.AK-04
  // (#1711).
  //
  // **Drei Zahlen, drei Wörter.** Haufenzahl, Marken und die Art eines
  // Objekts standen alle als bloße Zahl oder gar nicht da; wer hört
  // statt sieht, konnte sie nicht trennen.
  tischAufHaufen: '{zahl} auf einem Haufen',
  // **Zwei Formen, denn „1 Marken“ ist kein Deutsch.** Gefunden vom
  // Lauf über die Tischfläche, der „1 Marken aufladung“ las.
  //
  // **Und die Stellung ist nicht dieselbe wie im Englischen**: dort steht
  // die Art **vor** dem Wort („3 +1/+1 counters“). Deshalb zwei Muster je
  // Sprache statt `zaehlwort` mit einem Anhang.
  tischMarkeEine: '{zahl} Marke {art}',
  tischMarkeViele: '{zahl} Marken {art}',

  // **Die vier Arten aus `domain.Objektart`** — eine Karte sagt nicht,
  // dass sie eine ist, denn das ist der Normalfall.
  tischArt: {
    spielstein: 'Spielstein',
    emblem: 'Emblem',
    faehigkeit: 'Fähigkeit auf dem Stapel',
  },

  sicht: 'Sicht für',
  // **Die Sicht ohne Sitz** — PLAY-05.AK-02
  // (#1713).
  // Sie zeigt, was **jeder** sehen darf: offene Zonen und sonst nichts.
  sichtZuschauer: 'Zuschauer (ohne Sitz)',
  sichtWechseln: 'Sitz wechseln',
  sichtLaedt: 'Die gewählte Sicht wird geladen …',
  verlassen: 'Partie schließen',

  // Die Zonen.
  zonen: 'Was wo liegt',
  zoneBibliothek: 'Bibliothek',
  zoneHand: 'Hand',
  zoneTisch: 'Schlachtfeld',
  zoneFriedhof: 'Friedhof',
  zoneExil: 'Exil',
  zoneStapel: 'Stapel',
  zoneKommando: 'Kommandozone',

  // **Die Kurzformen der Ablagen** — über einem Stapel auf dem Tisch, wo das
  // ganze Wort nicht passt (seit dem 2026-09-24).
  zonekurzBibliothek: 'Bibl.',
  zonekurzFriedhof: 'Friedh.',
  zonekurzExil: 'Exil',
  zonekurzKommando: 'Komm.',
  zoneLeer: 'leer',

  // **Ein Befund ist kein Fehler** (§8): die Karte steht da, nur dieses
  // eine Wort kennt das Regelwerk nicht.
  unbekannterTyp: 'Unbekannter Typ: {woerter}',
  geheim: 'verdeckt',
  getappt: 'getappt',

  // Die Befehle.
  befehle: 'Handlungen',
  regelnTitel: 'Was durchgesetzt wird',
  regelwerk: 'Regelwerk',
  wirksamAb: 'wirksam ab',
  hoechsteStrenge: 'Höchste Strenge',
  strenge: 'Strenge',
  andereSeite: 'Andere Seite',
  grossAnsehen: 'Groß ansehen',
  ersteSeite: 'Erste Seite',
  stelle: 'An die Stelle',
  stelleOben: 'oben',
  stelleUnten: 'unten',
  stelleZufaellig: 'zufällig einstecken',

  // Blick und Mischung.
  mischpflichtTitel: 'Offene Mischpflichten',

  // **Die Legendenregel** — CR 704.5j (M12, #1614). Der Satz nennt die
  // Regelnummer, weil er eine Partie anhält: wer ihn liest, soll nachlesen
  // können, warum nichts weitergeht.
  legendeTitel: 'Offene Wahl: Legendenregel',
  legendeSatz: '{wer} muss wählen, welche Legende „{name}“ bleibt (CR 704.5j).',
  legendeBehalten: 'Diese behalten',

  // **Die Reihung der Auslöser** — CR 603.3b (M9, #1611). Auch hier steht
  // die Regelnummer, und aus demselben Grund: solange geordnet wird, hat
  // niemand Priorität, und wer das liest, soll nachlesen können, warum.
  reihungTitel: 'Offene Wahl: Reihenfolge der Fähigkeiten',
  reihungSatz:
    '{wer} beherrscht mehrere Fähigkeiten, die zugleich ausgelöst haben, und sagt, in welcher Reihenfolge sie auf den Stapel kommen (CR 603.3b). Was zuerst gelegt wird, liegt unten und wird zuletzt verrechnet.',
  reihungLegen: 'Als Nächste legen:',

  // Der Stand.
  stand: 'Stand',
  standLaeuft: 'läuft',
  standPausiert: 'pausiert',
  standBeendet: 'beendet',
  pausieren: 'Anhalten',
  fortsetzen: 'Fortsetzen',
  beenden: 'Beenden',
  beendenFrage:
    'Diese Partie beenden? Ein Beenden ist endgültig — eine beendete Partie lässt sich nicht fortsetzen.',
  beendenJa: 'Beenden',
  beendenNein: 'Abbrechen',

  // Was zuletzt geschah.
  ereignisse: 'Zuletzt geschehen',
  ereignisseLeer: 'Noch nichts.',
  ereignisseLuecke:
    'Hier fehlen womöglich Züge: während die Seite verborgen war, kamen mehr Meldungen, als nachgeholt werden.',
  ereignisPrivat: 'Nur für dich',
  ereignisSatz: '{sitz}: {handlung} — {objekt}',
  ereignisSatzOhneKarte: '{sitz}: {handlung}',

  // **Die Rückschau** — Werkstück G5 von Phase 5.5
  // (#1462).
  //
  // **Die Überschrift nennt den Sitz**, für den gerechnet wurde: es gibt
  // keinen Verlauf der Partie, sondern nur den eines Sitzes.
  verlauf: 'Was bisher geschah',
  verlaufAnsehen: 'Verlauf ansehen',
  verlaufNachladen: 'Neu laden',
  verlaufLaeuft: 'Wird geholt …',
  verlaufUngeholt:
    'Der Verlauf ist noch nicht geholt. Er zeigt jeden Zug dieser Partie, so wie dieser Sitz ihn sehen darf.',
  verlaufLeer: 'In dieser Partie ist noch nichts geschehen.',
  verlaufGrenze:
    'Was ein anderer Sitz nur für sich gesehen hat, steht hier nicht — auch nicht nachträglich.',

  // **Was von einer Partie übrig bleibt** — Werkstück G5 von Phase 5.5
  // (#1462).
  protokoll: 'Was von dieser Partie bleibt',
  protokollErklaerung:
    'Jeder Zug dieser Partie steht auf diesem Rechner, und jeder Sitz sieht davon seinen Teil. Wer will, gibt seinen Teil für später frei — für einen einzelnen Mitspieler oder für jeden.',
  protokollAnsehen: 'Freigaben nachsehen',
  protokollNiemand: 'Niemand hat etwas freigegeben.',
  protokollFuer: 'Freigegeben für',
  protokollAlle: 'jeden',
  protokollSitz: 'Sitz',
  protokollEmpfaenger: 'Empfänger, mit Komma getrennt — oder * für jeden',
  protokollFreigeben: 'Freigeben',
  protokollZuruecknehmen: 'Zurücknehmen',
  protokollNurSpaeter:
    'Eine Freigabe wirkt erst nach der Partie. Am Tisch sieht weiter jeder nur, was ihm gehört.',
  protokollKeineRueckholung:
    'Eine Rücknahme schließt die Tür — sie holt nicht zurück, was jemand schon gelesen hat.',
  protokollAufbewahrung:
    'Ein beendetes Protokoll steht nach 90 Tagen in der Datenpflege zur Entscheidung. Weggeräumt wird es erst, wenn du es sagst.',
  protokollLoeschen: 'Protokoll löschen',
  protokollLoeschenErklaerung:
    'Damit sind Zustand, Verlauf und Freigaben dieser Partie auf diesem Rechner weg. Über Kopien anderswo sagt das nichts.',
  protokollGeloescht: 'Das Protokoll ist gelöscht.',
  // **Gefragt wird davor** — UX-10.AK-04 (#1682): was weg ist, kommt auf
  // diesem Rechner nicht wieder.
  protokollLoeschenFrage: 'Das Protokoll dieser Partie löschen?',
  protokollLoeschenJa: 'Protokoll löschen',

  // **Die Diagnose der Leitung** (G5, #1462).
  diagnose: 'Wie die Leitung steht',
  diagnoseErklaerung:
    'Hier steht, worüber dieser Tisch spricht und wie viel gerade wartet. Nichts davon sagt etwas über Karten — weder deine noch fremde.',
  diagnoseAnsehen: 'Leitung prüfen',
  diagnoseStill: 'Gerade läuft kein Tisch.',
  diagnoseAlsWirt: 'Dieser Rechner führt den Tisch.',
  diagnoseAlsGast: 'Dieser Rechner sitzt an einem fremden Tisch.',
  diagnosePausiert: 'Die Leitung ist abgerissen — der Platz bleibt.',
  diagnoseWeg: 'Weg',
  diagnoseGegenueber: 'Gegenüber',
  diagnoseSeit: 'Steht seit',
  diagnoseWartend: 'Wartet',
  diagnoseVerworfen: 'Verworfen',
  diagnoseStand: 'Im Strom bei',
  diagnoseRunde: 'Letzte Runde',
  diagnoseBildzeit: 'Bildabstand in diesem Fenster',
  diagnoseUnbekannt: 'unbekannt',
  diagnoseOhneLeitung: 'Es steht keine Leitung.',

  // **Die Freigabe für Gäste** — Werkstück G5 von Phase 5.5
  // (#1462).
  freigabe: 'Mitspieler an diesen Tisch lassen',
  freigabeErklaerung:
    'Erreichbar sein und mitspielen lassen sind zwei Entscheidungen. Wer hier freigibt, erlaubt einem Gegenüber, an dieser Partie zu ziehen — die Sitze vergibst du, nicht er.',
  freigabeOffen: 'Dieser Tisch steht offen.',
  freigabeZu: 'Dieser Tisch steht nicht offen.',
  freigabeStarten: 'Tisch öffnen',
  freigabeBeenden: 'Tisch schließen',
  freigabeNachsehen: 'Nachsehen',
  freigabeSitze: 'Frei für',
  freigabeNiemand: 'Noch sitzt niemand.',
  freigabePartner: 'Das ist',
  freigabePartnerFremd: 'Dieses Gegenüber steht nicht in deinen Kontakten.',
  freigabeBesetzt: 'Es sitzen:',
  freigabeOhneHorchposten:
    'Damit ein Gegenüber ankommt, muss dieser Rechner außerdem erreichbar sein — das steht unter Einstellungen.',

  // **Der gezeichnete Tisch** — Werkstück G3 von Phase 5.5
  // (#1460).
  tischOhneWebgl:
    'Dieser Rechner stellt keine beschleunigte Grafik bereit. Der Tisch steht darunter als Liste, und dort geht alles, was auch am gezeichneten Tisch geht.',
  tischOhneRenderer:
    'Der Tisch ließ sich nicht zeichnen. Die Liste darunter führt jede Karte, und dort geht alles Weitere.',
  tischVerloren:
    'Die Grafikverbindung ist abgerissen. Der Tisch erscheint wieder, sobald du ihn verlässt und erneut aufrufst; bis dahin führt die Liste darunter jede Karte.',
  tischHeimatlos:
    'Diese Objekte liegen in einer Zone, die dieses Programm nicht kennt, und stehen deshalb nicht am Tisch:',
  tischVerdeckt: 'verdeckt',
  tischGetappt: 'getappt',
  tischHaengtAn: 'hängt an',
  tischbild: 'Hintergrund',
  tischbildHolz: 'Holz',
  tischbildDunklesHolz: 'Dunkles Holz',
  tischbildPicknick: 'Picknick',
  tischbildDungeon: 'Verlies',
  tischbildTaverne: 'Taverne',
  tischbildTurniertisch: 'Turniertisch',
  tischEigenes: 'Eigenes Bild …',
  tischEigenesWeg: 'Eigenes Bild entfernen',
  gespraech: 'Gespräch am Tisch',
  gespraechVerlauf: 'Was am Tisch gesagt wurde',
  gespraechEingabe: 'Etwas sagen',
  gespraechSenden: 'Sagen',
  gespraechLeer: 'Noch hat niemand etwas gesagt.',
  gespraechNiemandZeigt: 'Gerade zeigt niemand auf etwas.',
  gespraechZeigt: '{sitz}: {tut}',
  gespraechZeigtAuf: '{sitz}: {tut} {objekt}',
  zeigerZeiger: 'zeigt',
  zeigerGreift: 'zieht',
  zeigerSchwebt: 'schaut',
  zeigerPing: 'ruft',
  spielart: 'Spielart',
  spielartGoldfish: 'Allein üben',
  spielartSelfplay: 'Mehrere Sitze',
  format: 'Format',
  fehltDeck: 'Wähle für jeden Sitz ein Deck — aus dem ersten kommt das Spiel.',
  fehltEigenesDeck:
    'Ein Sitz braucht ein eigenes Deck — eine Leihe bringt kein Spiel mit, und aus dem ersten eigenen Deck kommt es.',
  fehltName:
    'Jeder Sitz braucht einen Namen; am Tisch unterscheidet er die Spielenden.',

  // **Die digitale Deckleihe** — Werkstück G6 von Phase 5.5
  // (#1463).
  leihe: 'Ein Deck verleihen oder leihen',
  // **Die Leihe in ihrer Schublade** — UXR-13 (#1921): offen ist das
  // Verleihen, das Annehmen eingeklappt, die Leihen darunter.
  leiheOeffnen: 'Verleihen oder leihen …',
  leiheVerleihenTitel: 'Ein Deck verleihen',
  leiheAnnehmenTitel: 'Eine Leihe annehmen',
  leiheListeTitel: 'Deine Leihen',
  leiheErklaerung:
    'Wer verleiht, erlaubt einem Gegenüber, mit einem eingefrorenen Stand seines Decks anzutreten. Das Deck bleibt hier, und an deinem Bestand ändert sich keine Karte.',
  leiheKeinEigentum:
    'Das ist eine Nutzungsfreigabe und keine Übertragung. Es wechselt nichts den Besitzer, weder digital noch physisch.',
  leiheKeinKopierschutz:
    'Und es ist kein Kopierschutz: Wer die Karten einmal gesehen hat, hat sie gesehen. Was die Leihe regelt, ist, womit dieses Programm eine Partie startet.',
  leiheAnsehen: 'Leihen ansehen',
  leiheKeine: 'Es gibt keine Leihen.',
  leiheEntleiher: 'An wen verleihst du?',
  leiheEntleiherLeer:
    'Noch kein bestätigtes Gegenüber. Nimm im Kasten „Gegenüber aufnehmen oder bestätigen“ darunter eines auf.',
  leiheDeck: 'Welches Deck',
  leiheDeckWaehlen: 'Ein Deck wählen …',
  leiheKeinDeck: 'Noch kein Deck — eines entsteht unter Decks.',
  leiheFachleute: 'Nur für eine bestimmte Partie verleihen',
  leiheFachleuteErklaerung:
    'Ohne Angabe gilt die Leihe für die nächste Partie des Gegenübers. Wer sie an genau eine binden will, trägt deren Kennung ein — das Gegenüber findet sie in seiner Partienliste.',
  leiheWoran: 'Kennung der Partie — leer heißt: die nächste',
  leiheStunden: 'Gilt für wie viele Stunden',
  leiheZiel: 'Wohin der Umschlag gelegt wird',
  leiheAnbieten: 'Verleihen und ablegen',
  leiheAbgelegt: 'Der Umschlag liegt hier:',
  leiheUmschlag: 'Umschlag einer Leihe',
  leiheAnnehmen: 'Leihe annehmen',
  leiheWiderrufen: 'Leihe widerrufen',
  leiheZurueckgeben: 'Zurückgeben',
  leiheVerliehen: 'verliehen an',
  leiheGeliehen: 'geliehen von',
  leiheBis: 'gilt bis',
  leiheKarten: 'Karten',
  leiheVerbraucht: 'eingelöst an',
  leiheOffen: 'noch nicht eingelöst',
  leiheZurueckgenommen: 'zurückgenommen',
  leiheModusZuende: 'Eine laufende Partie läuft aus.',
  leiheModusAbbruch: 'Eine laufende Partie wird abgebrochen.',
  leiheRuecknahmeGrenze:
    'Eine Rücknahme verhindert neue Starts. Sie entfernt nichts aus einer laufenden Partie und löscht keine Kopie beim Gegenüber.',

  // **Wie die Partie ausgegangen ist** (M17, #1618).
  ausgangTitel: 'Das Ergebnis',
  ausgangSieger: 'Gewonnen hat',
  ausgangUnentschieden:
    'Unentschieden — alle Verbliebenen haben zugleich verloren.',
  grundAufgabe: 'hat aufgegeben',
  standLeiste: 'Stand der Spieler',
  standDu: 'Du',
  standLeben: 'Leben',
  standGift: 'Gift',
  standMana: 'Mana',

  // **Die fünf Zahlen aus PLAY-03** — CR 728.1, CR 903.3, CR 903.8,
  // CR 903.10a und CR 704.5b
  // (#1711).
  standStrahlung: 'Strahlung',
  standSteuer: 'Steuer',
  standCommanderschaden: '{name}: {schaden}',

  // **Der Satz, den der Entwurf wörtlich verlangt** (Artboard
  // `TischCommander`). Er sagt, was gilt — nicht nur, dass eine Zahl hoch
  // ist.
  standCommanderTod: 'ab 21 ist die Partie vorbei',
  standLeergezogen: 'zog aus einer leeren Bibliothek',
  grundLeben: 'stand bei null Leben oder darunter',
  grundBibliothek: 'sollte aus einer leeren Bibliothek ziehen',
  grundGift: 'hatte zehn Vergiftungsmarken',

  // **Ein Sitz an einem Profil** — Werkstück G8 von Phase 5.5
  // (#1465).
  sitzprofil: 'Profil',
  sitzprofilKeines: 'kein Profil',
  sitzprofilErklaerung:
    'Wer zwei Profile führt, kann jedem Sitz eines geben. Die Partie beginnt dann in der Sicht des Sitzes, der zu deinem gezeigten Profil gehört; an die anderen kommst du über die Sichtwahl am Tisch. Was der anderen Sicht gehörte, steht danach nicht mehr da. Wählst du für keinen Sitz ein Profil, sitzt dein gezeigtes Profil am ersten Sitz.',
  sitzprofilBefunde: 'Dazu gibt es einen Hinweis:',
}

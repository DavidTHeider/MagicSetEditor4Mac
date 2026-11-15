// **Welche Meldung im Band über dem Inhalt steht** —
// #180.
//
// **Herausgelöst aus `Anwendung.tsx`**, weil sie über die Zeilengrenze wuchs
// (§2). Der Schnitt folgt der Frage: dort steht, **was** das Fenster zeigt,
// hier, **welche von fünf Meldungen** dabei den Vortritt hat — eine reine
// Funktion, einzeln prüfbar.
import { istSetseite, type Seitenschluessel } from './navigation'
import { type Bandmeldung } from './rahmen/Meldungsband'
import { einsetzen } from './sprache/einsetzen'
import { kataloge, type Sprache } from './sprache/kataloge'

// Meldungslage ist alles, was über die Reihenfolge entscheidet.
export interface Meldungslage {
  sprache: Sprache

  // rueckspielrest sagt, dass ein abgebrochenes Zurückspielen beim Start nicht
  // abgeschlossen werden konnte — sonst leer.
  rueckspielrest: string

  // dateimeldung ist die Antwort auf einen Doppelklick im Finder.
  dateimeldung: string | undefined

  verlaufsfehler: string | undefined

  // darstellungsfehler ist, woran »Darstellung zurücksetzen« aus der
  // Menüleiste scheiterte — Befund G5 des Zweitblicks über Phase 6.3.
  darstellungsfehler: string | undefined

  bestandsfehler: string | undefined
  bestandspfad: string | undefined

  // setID ist das offene Set — oder nichts, wenn keines offen ist.
  setID: string | undefined

  // setname ist der Name des offenen Sets, sobald er geladen ist.
  setname: string | undefined

  // seite ist, was gerade dasteht — S08.NAT-02 (#1680): die Meldung eines
  // Sets gehört auf die Seiten eines Sets.
  seite: Seitenschluessel

  // speicherungFuer ist das Set, zu dem die Speicherantwort gehört.
  speicherungFuer: string | undefined
  speicherfehler: string | undefined
  speicherpfad: string | undefined
}

// bandmeldung wählt aus, was dasteht.
//
// **Die Reihenfolge ist die Aussage** — ein Fehler schlägt eine
// Erfolgsmeldung, denn er verlangt etwas.
export function bandmeldung(lage: Meldungslage): Bandmeldung | undefined {
  const texte = kataloge[lage.sprache]

  // **Ein liegengebliebenes Zurückspielen steht ganz vorn** (#1170, Befund 2
  // des zweiten Freigabeblicks).
  //
  // Es ist die einzige Meldung hier, die von einem Zustand handelt, der schon
  // vor dem Fenster bestand: die Anwendung läuft mit der **neuen** Datenbank
  // und den **alten** Bildern. Alles andere in dieser Reihe ist die Antwort
  // auf eine Handlung von eben — und eine Handlung von eben kann man
  // wiederholen, diesen Zustand nicht.
  if (lage.rueckspielrest) {
    return { text: lage.rueckspielrest, art: 'schlecht' }
  }

  // **Die Datei aus dem Finder steht vorn.** Sie ist die Antwort auf eine
  // Handlung außerhalb des Fensters; wer sie nicht sieht, doppelklickt noch
  // einmal.
  if (lage.dateimeldung) {
    return { text: lage.dateimeldung, art: 'schlecht' }
  }

  // **Jeder Fehler nennt seinen Vorgang** — S08.NAT-02 (#1680). Ein
  // „Eintrag nicht gefunden." allein sah über jeder Seite aus wie ihr Fehler.
  const band = texte.allgemein.meldungsband

  if (lage.verlaufsfehler) {
    return {
      text: einsetzen(band.verlaufGescheitert, { grund: lage.verlaufsfehler }),
      art: 'schlecht',
    }
  }

  // **Hinter dem Verlauf, vor dem Bestand**: beides ist die Antwort auf einen
  // Menüeintrag von eben, und wer die Darstellung zurücksetzen wollte, steht
  // gerade nicht auf der Einstellungsseite, die es auch sagt.
  if (lage.darstellungsfehler) {
    return {
      text: einsetzen(band.darstellungGescheitert, {
        grund: lage.darstellungsfehler,
      }),
      art: 'schlecht',
    }
  }

  // **Die Sicherung des Bestands steht vor der des Sets** (#799): sie ist die
  // seltenere Handlung, und wer sie gerade ausgelöst hat, wartet auf ihre
  // Antwort. Eine Setmeldung von vorhin darf sie nicht verdecken.
  if (lage.bestandsfehler) {
    return {
      text: einsetzen(band.bestandNichtGesichert, {
        grund: lage.bestandsfehler,
      }),
      art: 'schlecht',
    }
  }

  if (lage.bestandspfad !== undefined) {
    return {
      text: `${texte.allgemein.bestandGesichert} ${lage.bestandspfad}`,
      art: 'gut',
    }
  }

  return setmeldung(lage, texte)
}

// setmeldung ist die Antwort auf ⌘S — sofern sie zum offenen Set gehört.
//
// **Nur die Meldung des offenen Sets** —
// #749.
//
// Wer ⌘S drückt, das Set wechselt und dann die Antwort bekommt, las sie bis
// dahin unter dem neuen. Bei der Erfolgsmeldung war das harmlos — sie nennt
// den Pfad; bei der **Fehlermeldung** nicht, denn sie trägt keinen Bezug (§7,
// Regeln 1 und 2).
//
// **Unterdrückt wird nur, wenn ein *anderes* Set offen ist.** Aus der Setliste
// heraus gesichert ist gar keines offen, und dort gehört die Meldung ins Band
// — der erste Anlauf hat sie auch dann verschluckt, und
// `englischesitzung.test.tsx` hat es gefangen.
function setmeldung(
  lage: Meldungslage,
  texte: (typeof kataloge)[Sprache],
): Bandmeldung | undefined {
  if (lage.setID !== undefined && lage.speicherungFuer !== lage.setID) {
    return undefined
  }

  // **Und nur auf den Seiten eines Sets** — S08.NAT-02 (#1680), Befund F10:
  // über Turnieren, Einstellungen und Hilfe stand das gescheiterte Sichern
  // wie deren eigener Fehler. **Verschwunden ist es damit nicht**: wer
  // zurückkommt, findet es wieder, bis er es schließt.
  if (lage.seite !== 'sets' && !istSetseite(lage.seite)) {
    return undefined
  }

  if (lage.speicherfehler) {
    const band = texte.allgemein.meldungsband

    return {
      text:
        lage.setname === undefined
          ? einsetzen(band.setNichtGesichertOhneName, {
              grund: lage.speicherfehler,
            })
          : einsetzen(band.setNichtGesichert, {
              name: lage.setname,
              grund: lage.speicherfehler,
            }),
      art: 'schlecht',
    }
  }

  if (lage.speicherpfad !== undefined) {
    return {
      text: `${texte.seiten.sets.gespeichert} ${lage.speicherpfad}`,
      art: 'gut',
    }
  }

  return undefined
}

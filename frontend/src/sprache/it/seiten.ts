// The texts of the individual pages.
import { bestandspflege } from './seitenbestand'
import { datenpflege } from './seitenpflege'
import { kosten } from './seitenkosten'
import { ruecksetzung } from './seitenruecksetzung'
import { ruecksprung } from './seitenruecksprung'
import { einstellungen } from './seiteneinstellungen'
import { nutzer } from './seitennutzer'
import { karten } from './seitenkarten'
import { sprachstaende } from './seitensprachstaende'
import { handeln } from './seitenhandeln'
import { technik } from './seitentechnik'
import { recherche } from './seitenrecherche'
import { decks } from './seitendecks'
import { turniere } from './seitenturniere'
import { handlungen } from './seitenhandlungen'
import { spielen } from './seitenspielen'
import { sammlung } from './seitensammlung'
import { adressbuch } from './seitenadressbuch'
import { vorlagen } from './seitenvorlagen'
import { sets } from './seitensets'
import { ausgeben } from './seitenausgeben'
import { werkzeugseiten } from './seitenwerkzeuge'
import { start } from './seitenstart'
import { paketplan } from './seitenpaketplan'
import { bezuege } from './seitenbezuege'

export const seiten = {
  start,
  sets,
  setangaben: {
    kopf: 'Dati del file',
    kopfHinweis:
      'Ciò che Magic Set Editor ha annotato su questo set al momento del salvataggio. Vengono conservati e non modificati.',
    mseFassung: 'Scritto con MSE',
    spielfassung: 'Versione del pacchetto di gioco',
    stilfassung: 'Versione del pacchetto di stile',
    versionskontrolle: 'Controllo versione',
    apprenticecode: 'Codice Apprentice',
    titel: 'Dati del set',
    leerTitel: 'Nessun campo del set',
    leer: 'Il pacchetto di gioco di questo set non dichiara blocchi `set field:`. Scegli un gioco che ne includa.',
  },
  karten,
  sprachstaende,
  decks,
  turniere,
  spielen,
  handlungen,
  sammlung,
  adressbuch,
  vorlagen,
  ...werkzeugseiten,
  bilder: {
    titel: 'Immagini',
    erklaerung:
      'Ogni carta viene disegnata e scritta singolarmente. L’esecuzione avviene qui nella finestra: mostra a che punto è e si può interrompere in qualsiasi momento.',
    ordnerWaehlen: 'Scegli una cartella …',
    keinOrdner: 'Nessuna cartella scelta finora.',
    gemerkt: 'Scelta l’ultima volta — la stessa cartella per tutti i set.',
    nichtGemerkt:
      'La cartella vale per questa uscita, ma non è stata memorizzata:',
    vorlage: 'Modello del nome',
    konflikt: 'Se un file esiste già',
    konfliktHilfe:
      'Due carte con lo stesso nome producono lo stesso nome di file. “Rinumera” è l’impostazione predefinita di Magic Set Editor 2: conta in avanti all’interno di un’esecuzione e sovrascrive ciò che un’esecuzione precedente ha lasciato.',

    // Die Schlüssel kommen aus `internal/domain/einstellung.go`.
    konflikte: {
      altbehalten: 'Mantieni quello esistente',
      ueberschreiben: 'Sovrascrivi',
      nummerieren: 'Numera',
      nummerierenneu: 'Rinumera',
    },
    vorlageHilfe:
      '{nummer} diventa il numero della carta a tre cifre, {name} il nome della carta. Ai nomi uguali viene aggiunto un numero.',
    format: 'Formato',
    formatPNG: 'PNG',
    formatJPG: 'JPG',
    formatBMP: 'BMP',
    formatTIFF: 'TIFF',
    formatSVG: 'SVG',
    guete: 'Qualità',
    gueteFast: 'Veloce — per una rapida occhiata',
    gueteMedium: 'Secondo la regola di Magic Set Editor 2',
    gueteHigh: 'Alta — senza sfarfallio sulle linee sottili',
    gueteUltra: 'Massima — per la stampa a piena risoluzione',
    breite: 'Larghezza in pixel',
    breiteHilfe:
      'Nessun valore in dpi: nessun modello indica quanto sia larga una carta in millimetri. 750 pixel su una carta di 63 mm sono 300 dpi.',
    ausgeben: 'Scrivi le immagini',
    abbrechen: 'Interrompi',
    laeuft: 'scritte su',
    fertig: 'immagini scritte in',
    abgebrochen: 'Interrotto.',
    uebersprungen:
      '{anzahl} carte saltate — il file esisteva già ed è selezionato “Mantieni quello esistente”:',
    befunde: 'Disegno incompleto. Le immagini ci sono; mancano queste parti:',
  },
  druckbogen: {
    titel: 'Foglio di stampa',
    erklaerung:
      'Le carte mantengono le loro dimensioni: 63 × 88 mm, esattamente come vengono tagliate. I segni di taglio si trovano sul bordo della carta da stampa, non sulla carta.',
    papier: 'Carta',
    papierA4: 'A4',
    papierLetter: 'Letter',
    anschnitt: 'Abbondanza in millimetri',
    anschnittHilfe:
      'L’abbondanza è la parte di immagine che supera la linea di taglio — evita che una carta tagliata mostri un bordo bianco. Le carte vengono disposte con il doppio dell’abbondanza tra l’una e l’altra.',
    aufloesung: 'Risoluzione',
    dpi150: '150 dpi',
    dpi300: '300 dpi',
    dpi600: '600 dpi',
    schnittmarken: 'Segni di taglio',
    legen: 'Componi il foglio di stampa …',
    fertig: 'su',
    seiten: 'pagine:',
    befunde:
      'Non tutto è entrato. Il foglio c’è; questi punti vanno controllati:',
  },
  // The **technical area** (#1164) — commands to copy.
  handeln,
  technik,

  // **Die Recherche** (A16, #1515) — sie steht im Sucharbeitsplatz und
  // im Karteneditor, nicht auf einer eigenen Seite.
  recherche,
  pakete: {
    plan: paketplan,
    bezuege,
    sammeldaten: {
      ueberschrift: 'Dati in blocco da Scryfall',
      erklaerung:
        'Un file con tutti i dati delle carte. Una volta presente, le funzioni di Scryfall lavorano anche senza rete — con dati vecchi al massimo di un giorno.',
      fehlt: 'Non ancora scaricati.',
      vorhanden: 'Presenti, {mb} MB.',
      angeboten: 'Scryfall offre {mb} MB, aggiornati al {stand}.',
      geholt: 'Scaricato il {zeit}.',
      holen: 'Scarica i dati in blocco',
      laedt: 'Download in corso …',
    },

    titel: 'Pacchetti',
    bezugUeberschrift: 'Scarica un modello',
    bezugErklaerung:
      'I modelli MSE si trovano in repository. Si scarica uno stato — una release o un commit — mai un branch: un branch cambia, e la stessa indicazione darebbe domani un pacchetto diverso.',
    bezugAdresse: 'Indirizzo del repository',
    bezugStand: 'Stato (release o commit)',
    bezugUnterverzeichnis: 'Sottocartella (radice del pacchetto)',
    bezugName: 'Nome nella cartella dei pacchetti',
    bezugPruefen: 'Mostra l’origine …',
    bezugStarten: 'Scarica',
    bezugAbbrechen: 'Non scaricare',
    bezugGeholt: 'Scaricato in',
    bezugPruefsumme: 'Checksum',
    bezugUebergangen: 'Non estratti — i file eseguibili non vengono inclusi:',
    bezugUnvollstaendig:
      'Mancano indirizzo e stato. Senza entrambi non c’è nulla da scaricare.',
    bezugHerkunft: 'Verrà scaricato {quelle}, stato {stand}.',
    bezugLizenz:
      'Questi contenuti non provengono da MagicSetEditor4Mac e non sono soggetti alla nostra licenza. Cornici delle carte, simboli e caratteri possono essere protetti; li usi sotto la tua responsabilità. I file eseguibili non vengono estratti.',
    mse2Ueberschrift: 'Importa da un’installazione di MSE2',
    mse2Erklaerung:
      'Chi usa MSE2 ha già i suoi modelli sul computer. Qui sotto trovi i percorsi abituali — nessuno di questi viene letto prima che tu lo scelga.',
    mse2Uebernehmen: 'Importa da qui',
    mse2VonHier: 'Importa da qui — {ort}',
    mse2Waehlen: 'Scegli una cartella …',
    mse2KeinVorschlag:
      'Nessuna installazione di MSE2 trovata nei percorsi abituali. Scegli tu la cartella con i modelli — in MSE2 si chiama “data”.',
    installerErklaerung:
      'Oppure un file .mse-installer — è così che MSE2 condivide i modelli. Ciò che è già installato resta com’è.',
    installerWaehlen: 'Leggi l’installer …',
    mse2Ergebnis: '{anzahl} modelli importati.',
    mse2Uebergangen:
      'Già presenti e quindi saltati — nulla è stato sovrascritto:',
    abhaengigUeberschrift: 'Dipendenze',
    abhaengigAllesGut: 'Tutte le dipendenze sono soddisfatte.',
    abhaengigFehlt:
      '{paket} richiede {braucht} — non installato. Scaricalo oppure importalo da un’installazione di MSE2.',
    abhaengigZuAlt:
      '{paket} richiede {braucht} nella versione {verlangt}; è installata la {da}. Senza quella più recente il modello disegna in silenzio qualcosa di diverso.',
    abhaengigZyklus:
      '{paket} e {braucht} dipendono l’uno dall’altro in un ciclo.',
    abhaengigUnlesbar:
      '{paket} richiede {braucht}: le versioni non si possono confrontare — richiesta {verlangt}, installata {da}.',
    quellenErklaerung:
      'Quattro raccolte a portata di clic — indirizzo e stato vengono compilati; non si scarica nulla finché non dai il tuo consenso. I caratteri dei modelli vengono inclusi; senza di essi nessuna carta viene disegnata correttamente. Nessuna delle quattro ha una licenza: i contenuti appartengono ai rispettivi titolari dei diritti, non a noi.',
    quellenGroesse: 'circa {megabyte} MB',

    // Die Schlüssel kommen aus `internal/bezug/quellen.go`. Ein Wächter
    // in `internal/werkzeuge` hält beide Seiten zusammen: eine Quelle
    // ohne Satz hier macht die Pipeline rot.
    quellen: {
      basicM15:
        'Le solite cornici M15. La via più rapida per ottenere una carta disegnata.',
      fullNonMagic: 'Modelli per giochi diversi da Magic.',
      m15: 'I modelli M15 pubblici — 324 pacchetti. Per la maggior parte delle persone è il punto di partenza giusto.',
      fullMagic:
        'Quasi tutti i modelli Magic pubblici — 526 pacchetti. Questa raccolta è il metro di paragone di questo progetto.',
    },
    katalogUeberschrift: 'Pacchetti installati',
    katalogLeer:
      'Nessun pacchetto installato. Senza gioco e stile l’anteprima ripiega su un sostituto.',
    katalogName: 'Nome',
    katalogArt: 'Tipo',
    katalogVersion: 'Versione',
    katalogHandlung: 'Azione',
    katalogBefunde: 'Durante la lettura delle cartelle dei pacchetti:',
    anlegen: 'Crea un pacchetto',
    anlegenArt: 'Tipo del nuovo pacchetto',
    anlegenArtSpiel: 'Gioco (.mse-game)',
    anlegenArtStil: 'Stile (.mse-style)',
    anlegenArtInclude: 'Include (.mse-include)',
    anlegenArtExport: 'Modello di esportazione (.mse-export-template)',
    anlegenName: 'Nome breve',
    anlegenNameHinweis:
      'Lettere, cifre, trattini e trattini bassi — diventa il nome della cartella',
    anlegenVollname: 'Nome completo',
    anlegenVollnameHinweis: 'Vuoto significa: uguale al nome breve',
    anlegenSpiel: 'Gioco del nuovo pacchetto',
    anlegenSpielHinweis:
      'Uno stile dispone i campi del suo gioco, un modello di esportazione li scrive',
    anlegenSpielLeer: 'Scegli',
    anlegenKnopf: 'Crea',
    anlegenGelungen: 'Creato: {name}',
    loeschen: 'Elimina',
    loeschenAn: 'Elimina — {name}',
    handlungenAn: 'Azioni — {name}',
    handlungenKurz: '…',
    katalogSuche: 'Cerca',
    auswahlLoesen: 'Mostra tutti i pacchetti',
    leiste: {
      bezeichnung: 'Cifre dei pacchetti',
      installiert: 'Installati',
      befund: 'Con rilievi',
      nichtGelesen: 'non letto',
      auswahl: '{pakete} con rilievi',
    },
    katalogSucheHinweis: 'Nome, tipo, origine o gruppo',
    katalogArtAlle: 'Tutti i tipi',
    katalogOrdnung: 'Ordina per',
    katalogOrdnungGruppen: 'Gruppi',
    katalogOrdnungName: 'Nome',
    katalogOrdnungVersion: 'Versione',
    katalogOrdnungHerkunft: 'Origine',
    katalogGefunden: '{gefunden} di {gesamt} pacchetti',
    katalogNichtsGefunden:
      'Nessun pacchetto corrisponde a questa ricerca. Prova un’altra grafia oppure azzera il filtro per tipo.',
    katalogLeerTitel: 'Ancora nessun pacchetto',
    katalogNichtsTitel: 'Nessun risultato',
    katalogOhneGruppe: 'Senza gruppo',
    katalogGruppeAn: 'Gruppo {name}',
    katalogGruppe: 'Gruppo',
    katalogOhneVersion: 'nessuna versione',
    katalogZustandFehlt: 'Dipendenza mancante',
    katalogZustandZuAlt: 'Dipendenza troppo vecchia',
    katalogZustandUnlesbar: 'Versioni non confrontabili',
    katalogZustandZyklus: 'Ciclo di dipendenze',
    katalogEinVorbehalt: 'un rilievo',
    katalogVorbehalte: '{anzahl} rilievi',
    umschlagTitel: 'Un modello da una busta',
    umschlagErklaerung:
      'Una busta contiene i **file** di un modello; il nome con l’estensione lo scegli tu, ad esempio `mygame.mse-game` o `mystyle.mse-style`. Con quel nome il modello si trova poi da te. Se il pacchetto esiste già, nulla viene sovrascritto: ricevi un rifiuto e scegli un altro nome.',
    umschlagName: 'Come deve chiamarsi il pacchetto',
    umschlagOhneNamen: 'Indica prima come deve chiamarsi il pacchetto.',
    umschlagAnnehmen: 'Accetta il modello',
    umschlagAngenommen:
      'Il modello si trova ora nella tua cartella dei pacchetti.',
  },
  // **Suchen und Ersetzen** (#1168).
  ersetzen: {
    titel: 'Sostituisci',
    erklaerung:
      'Sostituisce il testo nei campi di testo di questo set. Prima l’anteprima, poi l’applicazione — e un solo annulla ripristina l’intera operazione.',
    suchtext: 'Trova',
    ersatz: 'Sostituisci con',
    grossKlein: 'Distingui maiuscole e minuscole',
    ganzesWort: 'Solo parole intere',
    vorschauen: 'Mostra cosa cambia',
    ersetzen: 'Sostituisci ora ({treffer})',
    nichtsGefunden: 'Nessun risultato. Prova un’altra parola?',
    zusammenfassung: 'Corrispondenze: {treffer} · carte interessate: {karten}',
    getan: 'Sostituito — corrispondenze: {treffer} · carte: {karten}',
    ruecknahme: 'Premi ⌘Z per annullare l’intera operazione.',
    stellen: 'Punti interessati',
    kartenfeld: 'Carta {nummer} · {feld}',
    weitere: 'Gli altri punti non sono elencati — vengono comunque sostituiti.',
  },

  statistik: {
    worueber: 'Contare per',
    belegungWahl: 'Copertura per campo',
    gruppeSpielpaket: 'Secondo il pacchetto di gioco',
    gruppeFelder: 'Per valore di campo',
    verteilungVon:
      'Distribuzione di «{feld}» — quota delle carte che lo contengono',
    nachDimension: 'Statistiche per «{dimension}»',
    titel: 'Statistiche',
    karten: 'carte',
    felder: 'campi distinti',
    leere: 'carte senza alcun valore',
    belegung: 'Copertura per campo — quota di tutte le carte del set',
    belegungLeer: 'Questo set non contiene ancora alcun valore di campo.',
    verteilungLeer: 'Questo campo non ha alcun valore in questo set.',
    anteil: 'Quota',
    dimensionenErklaerung:
      'Un pacchetto di gioco stabilisce da sé cosa si può contare — categoria ' +
      'di colore, valore di mana, rarità. Ogni dimensione calcola il suo ' +
      'valore per ogni carta e richiede quindi un istante.',
    dimensionLeer: 'Nessuna carta ha un valore lungo questa dimensione.',
    gruppeHinweis:
      'Scegli una barra per vedere le sue carte nell’elenco delle carte.',
    gruppeGewaehlt: 'Scelto: «{gruppe}» · carte: {karten}',
    gruppeZeigen: 'Mostra le carte',
    gruppeAufheben: 'Annulla la scelta',
    gruppeLeer: 'Nessuna carta ha questo valore.',
    filterBezeichnung: 'Filtro dalle statistiche',
    filterKopf: 'Dalle statistiche: «{quelle}» = «{gruppe}»',
    filterZurStatistik: 'Torna alle statistiche',
    filterAufheben: 'Rimuovi il filtro',
  },
  // Die Lasche „Style“: Stilpaket und Stiloptionen des Sets (#288).
  stil: {
    titel: 'Stile',
    vorschauTitel: 'Anteprima',
    vorschauLeer:
      'Questo set non ha ancora carte — creane una per vedere lo stile.',
    eigeneFarbe: 'colore personalizzato',
    vorgabe: 'Predefinito',
  },
  suche: {
    titel: 'Ricerca',
    eingabe: 'Cosa cerchi?',
    recherche: 'Ricerca …',
    feld: 'Solo in questo campo',
    alleFelder: 'Tutti i campi',
    set: 'Solo in questo set',
    alleSets: 'Tutti i set',
    leer: 'Non hai ancora cercato nulla. Scrivi qualcosa per trovare carte.',
    nichts: 'Nessun risultato. Prova un’altra parola?',
    gekuerzt:
      'I risultati sono più di quelli mostrati qui. Restringi la ricerca per vederli.',
    treffer: 'Risultati',
  },
  ausgeben,
  hilfe: {
    titel: 'Aiuto',
    erklaerung:
      'Cosa significano i termini con cui lavora questa applicazione — e cosa fare al primo avvio.',
    glossar: 'Termini',

    // **Die Beschriftung der Laschenleiste** — seit U11 sind es Laschen
    // und keine Sprungmarken mehr; der Name ist geblieben, weil er sagt,
    // was die Leiste tut.
    aufDieserSeite: 'In questa pagina',
    seitenuebersicht: 'Cosa fa ogni pagina',
  },

  bestandspflege,
  datenpflege,
  kosten,
  ruecksetzung,
  ruecksprung,

  einstellungen,

  nutzer,
} as const

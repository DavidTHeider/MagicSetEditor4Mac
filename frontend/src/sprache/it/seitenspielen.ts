// The texts of the **Play** area — work item G1 of phase 5.5
// (#1458).
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

  titel: 'Gioca',
  erklaerung:
    'Una partita è un tavolo con mazzi congelati. Ciò che accade qui è un flusso di eventi — ciascuno dipende dal precedente e nessuno viene sovrascritto. Se sposti una carta, vedi subito ciò che ne vedono gli altri.',
  anfangTitel: 'Inizia una partita',
  anfangErklaerung:
    'Ogni posto porta un mazzo, e questo viene congelato: modificare il mazzo in seguito non cambia questa partita. Il gioco viene dal primo mazzo scelto.',
  leerDecksTitel: 'Ancora nessun mazzo',
  leerDecks: 'Per giocare serve un mazzo. Creane uno in Mazzi.',
  sitz: 'Posto',
  sitzName: 'Nome al tavolo',
  sitzDeck: 'Mazzo',
  sitzLeihe: 'Preso in prestito da …{von} · {karten} carte',
  deckWaehlen: 'Scegli un mazzo',
  anschlussWartet:
    'Il mazzo “{name}” è arrivato con te e occupa il primo posto. Il secondo è libero.',
  sitzHinzu: 'Aggiungi un posto',
  sitzWeg: 'Rimuovi questo posto',
  sitzGruppe: 'Posto {nummer}',
  spiel: 'Gioco',
  anfangen: 'Inizia la partita',
  tischTitel: 'Al tavolo',
  partieGegen: 'Partita contro {namen}',
  partieZwischen: 'Partita tra {namen}',
  partieVon: 'Partita in solitaria: {namen}',
  fassung: 'Revisione',
  wirt: 'Host',
  wirtHinweis:
    'L’host detiene lo stato completo di questa partita e quindi vede più di quanto mostri la sua vista. Lo dice qui perché va saputo prima di partecipare, non dopo.',
  tischAufHaufen: '{zahl} in un mucchio',
  tischMarkeEine: '{zahl} segnalino {art}',
  tischMarkeViele: '{zahl} segnalini {art}',
  tischArt: {
    spielstein: 'Pedina',
    emblem: 'Emblema',
    faehigkeit: 'Abilità sulla pila',
  },

  sicht: 'Vista per',
  sichtZuschauer: 'Spettatore (senza posto)',
  sichtWechseln: 'Cambia posto',
  sichtLaedt: 'Caricamento della vista scelta …',
  verlassen: 'Chiudi la partita',
  zonen: 'Che cosa si trova dove',
  zoneBibliothek: 'Grimorio',
  zoneHand: 'Mano',
  zoneTisch: 'Campo di battaglia',
  zoneFriedhof: 'Cimitero',
  zoneExil: 'Esilio',
  zoneStapel: 'Pila',
  zoneKommando: 'Zona di comando',
  zonekurzBibliothek: 'Grim.',
  zonekurzFriedhof: 'Cim.',
  zonekurzExil: 'Esilio',
  zonekurzKommando: 'Com.',
  zoneLeer: 'vuoto',

  unbekannterTyp: 'Tipo sconosciuto: {woerter}',
  geheim: 'coperta',
  getappt: 'tappata',
  befehle: 'Azioni',
  regelnTitel: 'Che cosa viene applicato',
  regelwerk: 'Regolamento',
  wirksamAb: 'in vigore da',
  hoechsteStrenge: 'Rigore massimo',
  strenge: 'Rigore',
  andereSeite: 'Altro lato',
  grossAnsehen: 'Visualizza in grande',
  ersteSeite: 'Primo lato',
  stelle: 'Nella posizione',
  stelleOben: 'in cima',
  stelleUnten: 'in fondo',
  stelleZufaellig: 'rimescolata dentro',
  mischpflichtTitel: 'Rimescolamenti dovuti in sospeso',

  legendeTitel: 'Scelta aperta: regola delle leggende',
  legendeSatz:
    '{wer} deve scegliere quale permanente leggendario chiamato “{name}” rimane (CR 704.5j).',
  legendeBehalten: 'Tieni questo',

  reihungTitel: 'Scelta aperta: ordine delle abilità',
  reihungSatz:
    '{wer} controlla più abilità che si sono attivate contemporaneamente e stabilisce in quale ordine vanno sulla pila (CR 603.3b). Ciò che viene messo per primo finisce in fondo e si risolve per ultimo.',
  reihungLegen: 'Metti come prossima:',
  stand: 'Stato',
  standLaeuft: 'in corso',
  standPausiert: 'in pausa',
  standBeendet: 'terminata',
  pausieren: 'Metti in pausa',
  fortsetzen: 'Riprendi',
  beenden: 'Termina',
  beendenFrage:
    'Terminare questa partita? Terminare è definitivo — una partita terminata non può essere ripresa.',
  beendenJa: 'Termina',
  beendenNein: 'Annulla',
  ereignisse: 'Appena accaduto',
  ereignisseLeer: 'Ancora nulla.',
  ereignisseLuecke:
    'Qui potrebbero mancare mosse: mentre la pagina era nascosta sono arrivati più messaggi di quanti se ne possano recuperare.',
  ereignisPrivat: 'Solo per te',
  ereignisSatz: '{sitz}: {handlung} — {objekt}',
  ereignisSatzOhneKarte: '{sitz}: {handlung}',

  // **Die Rückschau** (G5, #1462).
  verlauf: 'Che cosa è successo finora',
  verlaufAnsehen: 'Mostra la cronologia',
  verlaufNachladen: 'Ricarica',
  verlaufLaeuft: 'Recupero in corso …',
  verlaufUngeholt:
    'La cronologia non è ancora stata recuperata. Mostra ogni mossa di questa partita, per quanto questo posto può vederla.',
  verlaufLeer: 'In questa partita non è ancora successo nulla.',
  verlaufGrenze:
    'Ciò che un altro posto ha visto solo per sé non è qui — nemmeno a posteriori.',

  // **Was von einer Partie übrig bleibt** (G5, #1462).
  protokoll: 'Che cosa resta di questa partita',
  protokollErklaerung:
    'Ogni mossa di questa partita è salvata su questo computer, e ogni posto ne vede la propria parte. Puoi rilasciare la tua parte per dopo — a un singolo avversario o a tutti.',
  protokollAnsehen: 'Mostra i rilasci',
  protokollNiemand: 'Nessuno ha rilasciato nulla.',
  protokollFuer: 'Rilasciato a',
  protokollAlle: 'tutti',
  protokollSitz: 'Posto',
  protokollEmpfaenger: 'Destinatari, separati da virgola — oppure * per tutti',
  protokollFreigeben: 'Rilascia',
  protokollZuruecknehmen: 'Ritira',
  protokollNurSpaeter:
    'Un rilascio ha effetto solo dopo la partita. Al tavolo ciascuno continua a vedere solo ciò che è suo.',
  protokollKeineRueckholung:
    'Ritirare chiude la porta — non recupera ciò che qualcuno ha già letto.',
  protokollAufbewahrung:
    'Un registro concluso viene proposto per la revisione nella pulizia dei dati dopo 90 giorni. Viene rimosso solo quando lo dici tu.',
  protokollLoeschen: 'Elimina registro',
  protokollLoeschenErklaerung:
    'Questo rimuove da questo computer stato, cronologia e rilasci di questa partita. Non dice nulla sulle copie altrove.',
  protokollGeloescht: 'Il registro è stato eliminato.',
  protokollLoeschenFrage: 'Eliminare il registro di questa partita?',
  protokollLoeschenJa: 'Elimina registro',

  // **Die Diagnose der Leitung** (G5, #1462).
  diagnose: 'Stato della connessione',
  diagnoseErklaerung:
    'Qui vedi su che cosa comunica questo tavolo e quanto è in attesa. Nulla di tutto ciò dice qualcosa sulle carte — né le tue né quelle altrui.',
  diagnoseAnsehen: 'Controlla la connessione',
  diagnoseStill: 'Al momento non c’è nessun tavolo attivo.',
  diagnoseAlsWirt: 'Questo computer ospita il tavolo.',
  diagnoseAlsGast: 'Questo computer è seduto al tavolo di qualcun altro.',
  diagnosePausiert: 'La connessione è caduta — il posto resta.',
  diagnoseWeg: 'Percorso',
  diagnoseGegenueber: 'Controparte',
  diagnoseSeit: 'Attiva da',
  diagnoseWartend: 'In attesa',
  diagnoseVerworfen: 'Scartati',
  diagnoseStand: 'Nel flusso a',
  diagnoseRunde: 'Ultimo andata e ritorno',
  diagnoseBildzeit: 'Intervallo tra i fotogrammi in questa finestra',
  diagnoseUnbekannt: 'sconosciuto',
  diagnoseOhneLeitung: 'Non c’è nessuna connessione attiva.',

  // **Die Freigabe für Gäste** (G5, #1462).
  freigabe: 'Fai giocare altri a questo tavolo',
  freigabeErklaerung:
    'Essere raggiungibili e far giocare altri sono due decisioni. Aprire il tavolo permette a una controparte di fare mosse in questa partita — i posti li assegni tu, non lei.',
  freigabeOffen: 'Questo tavolo è aperto.',
  freigabeZu: 'Questo tavolo non è aperto.',
  freigabeStarten: 'Apri il tavolo',
  freigabeBeenden: 'Chiudi il tavolo',
  freigabeNachsehen: 'Ricontrolla',
  freigabeSitze: 'Liberi per',
  freigabeNiemand: 'Non è ancora seduto nessuno.',
  freigabePartner: 'Si tratta di',
  freigabePartnerFremd: 'Questa controparte non è tra i tuoi contatti.',
  freigabeBesetzt: 'Seduti:',
  freigabeOhneHorchposten:
    'Perché una controparte possa arrivare, questo computer deve anche essere raggiungibile — lo trovi in Impostazioni.',

  // **Der gezeichnete Tisch** — Werkstück G3 von Phase 5.5
  // (#1460).
  tischBild: 'Il tavolo',
  tischListe: 'Come elenco',
  tischGrafisch: 'Come tavolo',
  tischDarstellung: 'Visualizzazione',
  tischEinpassen: 'Mostra tutto',
  tischNaeher: 'Ingrandisci',
  tischWeiter: 'Riduci',
  tischDrehen: 'Ruota il tavolo',
  tischAuswahl: 'Sulla selezione',
  tischEigenerSitz: 'Sul mio posto',
  tischBlick: 'Vista del tavolo',
  tischZiel: 'Vai a',
  tischOhneWebgl:
    'Questo computer non offre grafica accelerata. Il tavolo compare qui sotto come elenco, e tutto ciò che funziona sul tavolo disegnato funziona lì.',
  tischOhneRenderer:
    'Non è stato possibile disegnare il tavolo. L’elenco qui sotto contiene ogni carta, e lì funziona tutto il resto.',
  tischVerloren:
    'La connessione grafica è caduta. Il tavolo torna quando la finestra viene riaperta; fino ad allora l’elenco qui sotto contiene ogni carta.',
  tischHeimatlos:
    'Questi oggetti si trovano in una zona che questo programma non conosce e perciò non compaiono sul tavolo:',
  tischVerdeckt: 'coperta',
  tischGetappt: 'tappata',
  tischHaengtAn: 'assegnata a',
  tischbild: 'Sfondo',
  tischbildHolz: 'Legno',
  tischbildDunklesHolz: 'Legno scuro',
  tischbildPicknick: 'Picnic',
  tischbildDungeon: 'Dungeon',
  tischbildTaverne: 'Taverna',
  tischbildTurniertisch: 'Tavolo da torneo',
  tischEigenes: 'Immagine personale …',
  tischEigenesWeg: 'Rimuovi l’immagine personale',
  gespraech: 'Conversazione al tavolo',
  gespraechVerlauf: 'Che cosa è stato detto al tavolo',
  gespraechEingabe: 'Scrivi qualcosa',
  gespraechSenden: 'Invia',
  gespraechLeer: 'Nessuno ha ancora detto nulla.',
  gespraechNiemandZeigt: 'In questo momento nessuno indica nulla.',
  gespraechZeigt: '{sitz}: {tut}',
  gespraechZeigtAuf: '{sitz}: {tut} {objekt}',
  zeigerZeiger: 'indica',
  zeigerGreift: 'trascina',
  zeigerSchwebt: 'guarda',
  zeigerPing: 'chiama',
  spielart: 'Modalità',
  spielartGoldfish: 'Allenamento da soli',
  spielartSelfplay: 'Più posti',
  format: 'Formato',
  fehltDeck: 'Scegli un mazzo per ogni posto — il primo determina il gioco.',
  fehltEigenesDeck:
    'Un posto ha bisogno di un mazzo tuo: un prestito non porta un gioco, e il gioco viene dal primo mazzo tuo.',
  fehltName:
    'Ogni posto ha bisogno di un nome; al tavolo distingue i giocatori.',

  // **Die digitale Deckleihe** (G6, #1463).
  leihe: 'Presta o prendi in prestito un mazzo',
  leiheOeffnen: 'Presta o prendi in prestito …',
  leiheVerleihenTitel: 'Presta un mazzo',
  leiheAnnehmenTitel: 'Accetta un prestito',
  leiheListeTitel: 'I tuoi prestiti',
  leiheErklaerung:
    'Prestare permette a un avversario di giocare con un’istantanea congelata del tuo mazzo. Il mazzo resta qui e nessuna carta della tua collezione cambia.',
  leiheKeinEigentum:
    'È una concessione d’uso, non un trasferimento. Nulla cambia proprietario, né digitalmente né fisicamente.',
  leiheKeinKopierschutz:
    'E non è una protezione dalla copia: chi ha visto le carte le ha viste. Ciò che il prestito regola è con che cosa questo programma avvia una partita.',
  leiheAnsehen: 'Mostra i prestiti',
  leiheKeine: 'Non ci sono prestiti.',
  leiheEntleiher: 'A chi presti?',
  leiheEntleiherLeer:
    'Ancora nessuna controparte confermata. Aggiungine una nel riquadro “Aggiungi o conferma una controparte” qui sotto.',
  leiheDeck: 'Quale mazzo',
  leiheDeckWaehlen: 'Scegli un mazzo …',
  leiheKeinDeck: 'Ancora nessun mazzo — se ne crea uno in Mazzi.',
  leiheFachleute: 'Presta solo per una partita precisa',
  leiheFachleuteErklaerung:
    'Senza indicazione, il prestito vale per la prossima partita della controparte. Per legarlo a una sola, inserisci l’ID di quella partita — la controparte lo trova nel suo elenco di partite.',
  leiheWoran: 'ID della partita — vuoto significa la prossima',
  leiheStunden: 'Valido per quante ore',
  leiheZiel: 'Dove mettere la busta',
  leiheAnbieten: 'Presta e salva',
  leiheAbgelegt: 'La busta si trova qui:',
  leiheUmschlag: 'Busta di un prestito',
  leiheAnnehmen: 'Accetta il prestito',
  leiheWiderrufen: 'Revoca',
  leiheZurueckgeben: 'Restituisci',
  leiheVerliehen: 'prestato a',
  leiheGeliehen: 'preso in prestito da',
  leiheBis: 'valido fino a',
  leiheKarten: 'carte',
  leiheVerbraucht: 'riscattato per',
  leiheOffen: 'non ancora riscattato',
  leiheZurueckgenommen: 'revocato',
  leiheModusZuende: 'Una partita in corso prosegue fino alla fine.',
  leiheModusAbbruch: 'Una partita in corso viene interrotta.',
  leiheRuecknahmeGrenze:
    'Revocare impedisce nuovi avvii. Non rimuove nulla da una partita in corso e non elimina alcuna copia presso l’altra parte.',

  // **Wie die Partie ausgegangen ist** (M17, #1618).
  ausgangTitel: 'Il risultato',
  ausgangSieger: 'Ha vinto',
  ausgangUnentschieden:
    'Pareggio — tutti i giocatori rimasti hanno perso contemporaneamente.',
  grundAufgabe: 'ha concesso',
  standLeiste: 'Situazione dei giocatori',
  standDu: 'Tu',
  standLeben: 'Vita',
  standGift: 'Veleno',
  standMana: 'Mana',
  standStrahlung: 'Radiazione',
  standSteuer: 'Tassa',
  standCommanderschaden: '{name}: {schaden}',
  standCommanderTod: 'a 21 la partita è finita',
  standLeergezogen: 'ha pescato da un grimorio vuoto',
  grundLeben: 'era a zero punti vita o meno',
  grundBibliothek: 'doveva pescare da un grimorio vuoto',
  grundGift: 'aveva dieci segnalini veleno',
  aufgeben: 'Concedi',
  aufgebenFrage:
    'Concedere? La partita finisce subito per questo posto, e le sue carte se ne vanno con lui.',
  aufgebenJa: 'Sì, concedi',
  aufgebenNein: 'Continua a giocare',
  aufgebenHinweis:
    'Puoi concedere in qualsiasi momento e ha effetto subito. Lasci la partita e le tue carte se ne vanno con te.',

  // **Der Weg zurück in die eigene Partie** (#1600).

  // **Ein Sitz an einem Profil** (G8, #1465).
  sitzprofil: 'Profilo',
  sitzprofilKeines: 'nessun profilo',
  sitzprofilErklaerung:
    'Se tieni due profili, puoi assegnarne uno a ciascun posto. La partita inizia allora nella vista del posto che appartiene al tuo profilo attuale; agli altri arrivi dal selettore di vista al tavolo. Ciò che mostrava l’altra vista sparisce. Se non assegni un profilo a nessun posto, il tuo profilo attuale occupa il primo posto.',
  sitzprofilBefunde: 'C’è una nota al riguardo:',
}

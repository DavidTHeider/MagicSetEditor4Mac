// **Die Texte der KI** — Werkstück A6 von Phase 5.3
// (#1445),
// erweitert für A7
// (#1446).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als einer
// ankommt** (§9): `seiteneinstellungen.ts` stand mit ihnen an der
// Zeilengrenze aus §2.
export const ki = {
  kimodelleTitel: 'Modelli di IA',
  kimodelleKeine:
    'Nessun modello di IA è stato accertato. L’elenco dei modelli di un fornitore non è ancora una capacità dimostrata — qui compaiono solo i modelli che sono stati effettivamente verificati.',
  kiNurLokal:
    'Vale “solo in locale”: non viene interpellato alcun modello al di fuori di questo computer.',
  kiAuchCloud:
    'I modelli di fornitori terzi sono consentiti. Ciò che viene inviato loro lascia questo computer.',
  kiBeleg: '{faehigkeit} — verificato il {datum}: {beleg}',
  kiKontextfenster: 'Finestra di contesto: {tokens} token — {beleg}',
  kiKontextfensterUnbekannt: 'Finestra di contesto: non indicata',
  kiHerkunftUnbekannt:
    'Origine sconosciuta — non è accertato quale file giri effettivamente qui.',
  kiEinrichtungAus:
    'Non attivato. Non viene chiesto nulla e nulla lascia questo computer.',
  kiEinrichtungUnerreichbar:
    'All’indirizzo configurato nessuno risponde. Questa applicazione non avvia né installa nulla lì.',
  kiEinrichtungOhneModell:
    'Risponde ma non ha alcun modello. Questa applicazione non ne scarica.',
  kiEinrichtungOhneFaehigkeit:
    'Ci sono modelli, ma nessuno ha una capacità comprovata di cui qui ci si possa servire.',
  kiEinrichtungBereit:
    'Almeno un modello è pronto con una capacità comprovata.',
  kiLaufzeitOllama:
    'Se qui non risponde nulla, Ollama non è in esecuzione. Questa applicazione non lo avvia né lo installa — questo avviene in Ollama stesso.',
  kiLaufzeitTitel: 'IA locale',
  kiLaufzeitFassung: 'Runtime: {fassung}',
  kiAdresse: 'Viene interpellato {adresse}.',
  kiAdresseVorgabe: 'l’indirizzo predefinito su questo computer',
  kiCloudAus: 'Il trasferimento nel cloud di Ollama è disattivato.',
  kiCloudAn: 'Il trasferimento nel cloud di Ollama non è disattivato.',
  kiCloudUngewiss:
    'Da qui non si può accertare se il trasferimento nel cloud di Ollama sia disattivato.',
  kiCloudBeleg: 'Riscontro: {beleg}',
  kiVorschlagTitel: 'Suggerimento da un modello',
  kiVorschlagModell: 'Modello',
  kiVorschlagAuftrag: 'Compito',
  kiVorschlagAuftragHinweis: 'Che cosa deve fare il modello?',
  kiVorschlagen: 'Suggerisci',
  kiVorschlagLaeuft:
    'Il modello sta lavorando — la prima volta può richiedere un po’ perché viene caricato.',
  kiVorschlagAbbrechen: 'Annulla',
  kiVorschlagUebernehmen: 'Applica',
  kiVorschlagNichts:
    'Il modello non ha suggerito alcun campo che il modello di carta conosca.',
  kiVorschlagAendertNichts:
    'Un suggerimento non modifica la carta. Solo “Applica” scrive.',
  kiVorschlagUebernommen: '{felder} campi sono stati scritti.',
  kiVorschlagAblauf: 'Applicabile fino a {ablauf}.',
  // **The second provider** — A8, #1447.
  kiAnbieterZeile: 'Fornitore {anbieter}: {stand}',
  kiVorschlagAnbieter: 'Fornitore e modello',
  kiVorschlagModellWaehlen: 'Scegli',
  kiVorschlagOhneModell:
    'Nessun modello disponibile. Ciò che manca è elencato sotto “Modelli di IA”.',
  kiVorschlagAntwortVon: 'La risposta è arrivata da {anbieter}/{modell}.',
  // **Measuring what a model can do** — A9, #1448.
  kiMessbarTitel: 'Modelli senza descrizione',
  kiAdresseSocket:
    'Interpellato tramite un socket Unix proprio di questo programma — nessuna porta.',
  kiGeraetSitzung: 'Interpellare in questa sessione',
  kiGeraetSitzungHinweis:
    'Vale fino al prossimo riavvio; poi torna a valere l’impostazione «Modello sul dispositivo di Apple».',
  kiFeldVorschlagen: 'Suggerisci',
  kiFeldVorschlagenFuer: '{feld} — suggerimento',
  kiFeldAuftragName:
    'Suggerisci un nome proprio adatto a questa carta — non il tipo di carta né una descrizione delle sue regole. Rispondi in italiano, solo con il nome.',
  kiFeldAuftragFlavor:
    'Suggerisci un breve testo d’atmosfera per questa carta, una o due frasi: atmosfera o storia che non ripeta il testo delle regole. Rispondi in italiano.',
  kiFeldKarte: 'La carta finora:',
  kiFeldUebernehmen: 'Applica',
  kiFeldVon: 'da {anbieter} · {weg}',
  kiFeldWegGewaehlt: 'scelto',
  kiFeldWegVorrang: 'prioritario',
  kiFeldVorrangFragt: 'La priorità interroga: {reihe}',
  kiFeldWegRueckfall: 'ripiego',
  kiFeldAbgelehnt: 'Il modello ha rifiutato: {grund}',
  kiFeldNichts: 'Nessun suggerimento per questo campo.',
  kiMessbarSatz:
    '{anbieter} elenca questi modelli ma non dice nulla su di essi. Una misurazione invia un piccolo compito con uno schema fisso — costa quanto costa quel singolo compito.',
  kiMessen: 'Misura',
  kiMessungVergessen: 'Scarta la misurazione',
  kiMessungLaeuft: 'Misurazione in corso …',
  kiMessungGeglueckt:
    '{alias} ha risposto: {faehigkeiten} sono ora comprovate.',
  // **Selection, comparison and provenance** — A11, #1450.
  kiVorschlagAuswahl: 'Applica: {feld}',
  kiVorschlagBisher: 'Attuale: {wert}',
  kiVorschlagBisherLeer: 'Attualmente vuoto.',
  kiVorschlagNichtsGewaehlt:
    'Nessun campo è spuntato. Viene applicato ciò che è spuntato.',
  kiHerkunftTitel: 'Da dove provengono i valori',
  kiHerkunftZeile: '{feld}: {anbieter}/{modell} il {zeitpunkt}',
  kiHerkunftAufgabe: 'Il compito era: {aufgabe}',
  kiHerkunftVeraendert: 'Da allora modificato a mano.',
  kiHerkunftKeine: 'Nessun campo di questa carta proviene da un suggerimento.',
  // **The image proposal** — A12, #1451.
  kiBildTitel: 'Immagine da un modello',
  kiBildAendertNichts:
    'Generare non modifica la carta. Ciò che nasce è una proposta con scadenza — la guardi, la ritagli e solo dopo la inserisci.',
  kiBildFeld: 'Campo immagine',
  kiBildAufgabe: 'Che cosa deve mostrare?',
  kiBildAufgabeHinweis: 'Un fulmine su una vasta pianura, dipinto a olio',
  kiBildGroesse: 'Dimensione',
  kiBildGuete: 'Qualità',
  kiBildBearbeiten:
    'Modifica l’immagine esistente invece di generarne una nuova',
  kiBildErzeugen: 'Genera immagine',
  kiBildLaeuft:
    'L’immagine viene creata presso {anbieter}/{modell} — ci vogliono da secondi a minuti.',
  kiBildAbbrechen: 'Annulla',
  kiBildAbgebrochen:
    'Annullato. Non è stato acquisito nulla; i costi maturati fino a quel momento si trovano nel registro dei costi.',
  kiBildVorschau: 'Proposta per {feld}',
  kiBildMasse: '{breite} × {hoehe} pixel, {kilobytes} kB',
  kiBildAblauf: 'Valida fino a {ablauf}.',
  kiBildUmschrift: 'Ciò che è stato disegnato: {umschrift}',
  kiBildNachweisTitel: 'Da dove proviene questa immagine',
  kiBildAngabeGroesse: 'Dimensione',
  kiBildAngabeGuete: 'Qualità',
  kiBildAngabeUmschrift: 'Prompt rivisto',
  kiBildAngabeMedientyp: 'Formato del file',
  kiBildUebernehmen: 'Ritaglia e inserisci',
  kiBildVerwerfen: 'Scarta',
  kiBildVerworfen: 'La proposta non c’è più.',
  kiBildKosten:
    'Un ordine di immagine viene autorizzato separatamente — l’autorizzazione per il testo non vale per esso.',
  // **Approving a paid request** —
  // #1539.
  kiFreigabeTitel: 'Questa richiesta costa denaro',
  kiFreigabeSatz:
    'Il modello scelto viene eseguito presso {anbieter}. Autorizza un importo massimo — vale per questa singola richiesta e scade dopo cinque minuti.',
  kiFreigabeBetrag:
    'Importo massimo nell’unità più piccola (per EUR: centesimi)',
  kiFreigabeVerfuegbar: 'Disponibile oggi: {betrag}',
  kiFreigabeOhneBudget:
    'Per {anbieter} non è autorizzato nulla. Senza un budget autorizzato non parte alcuna richiesta a pagamento — il limite si imposta sotto “Costi”.',
  kiFreigabeErteilen: 'Autorizza e chiedi',
  kiFreigabeVerzichten: 'Non autorizzare',
  kiFreigabeLaeuft: 'Emissione dell’autorizzazione …',

  // **The evidence-backed assessment** — A13.2a
  // (#1452).
  kiEinschaetzungTitel: 'Richiedi una valutazione',
  kiEinschaetzungErklaerung:
    'Un modello valuta questa carta sulla base di passaggi tratti dall’indice di ricerca locale. Non modifica nulla — né la carta né la collezione.',
  kiEinschaetzungArt: 'Di che cosa si tratta',
  kiEinschaetzungArtRegel: 'Secondo le regole',
  kiEinschaetzungArtBalance: 'In confronto',
  kiEinschaetzungFrage: 'Che cosa valutare',
  kiEinschaetzungFragen: 'Ottieni valutazione',
  kiEinschaetzungLaeuft: 'Il modello sta leggendo i riscontri …',
  kiEinschaetzungLeer: 'Ancora nulla valutato.',
  kiEinschaetzungBelegt: 'Tutto ciò che è citato compare nei riscontri.',
  kiEinschaetzungUnbelegt:
    'Questa risposta cita qualcosa che non compare in nessun riscontro. Di cosa si tratta è indicato qui sotto.',
  kiEinschaetzungBefunde: 'Ciò che non è comprovato',
  kiEinschaetzungEinschraenkungen: 'Ciò che non è stato considerato',
  kiEinschaetzungBelege: 'I passaggi forniti',
  kiEinschaetzungStaende: 'Stato delle fonti',

  // **The balance analysis** — A13.2b
  // (#1452).
  kiEinschaetzungFormat: 'Formato',
  kiEinschaetzungOhneFormat: 'nessuno indicato',
  kiEinschaetzungVergleiche: 'Con cosa è stato confrontato',
  kiEinschaetzungStrategie: 'Trovato tramite',
  kiEinschaetzungOhneVergleich:
    'Questa risposta non cita alcuna carta di confronto. Un’affermazione su come si posiziona una carta è un confronto.',
  kiEinschaetzungSynergien: 'Quali sinergie sono state escluse',

  // **The batch and the atomic group** — A13.1b
  // (#1452).
  kiBatchTitel: 'Più carte contemporaneamente',
  kiBatchSatz:
    'Un modello risponde allo stesso compito per ogni carta selezionata. Per ora non viene scritto nulla — lo decide il secondo passaggio.',
  kiBatchZiele: '{gewaehlt} di {gesamt} carte selezionate',
  kiBatchOhneKarten: 'Questo set non ha ancora carte.',
  kiBatchOhneName: '(senza nome)',
  kiBatchAuftrag: 'Compito per ogni carta',
  kiBatchFahren: 'Ottieni suggerimenti',
  kiBatchLaeuft: 'In corso …',
  kiBatchLagen: 'Com’è andata',
  kiBatchGanzOderGar:
    'Viene applicato tutto insieme oppure niente. Se un obiettivo fallisce, ogni carta resta com’era — e i suggerimenti restano aperti.',
  kiBatchGruppe: 'Applica come gruppo ({karten})',
  kiBatchGruppeLaeuft: 'Applicazione in corso …',
  kiBatchGeschrieben: '{karten} carte scritte.',

  // **Card recognition from a photo** — A13.3
  // (#1452).
  kiErkennungTitel: 'Riconosci una carta da una foto',
  kiErkennungSatz:
    'Un modello legge ciò che è stampato sulla foto. Quale carta sia lo decide la stessa ricerca delle edizioni usata per un elenco digitato — e nella collezione non viene aggiunto nulla.',
  kiErkennungLesen: 'Scegli una foto e leggila',
  kiErkennungLaeuft: 'Il modello sta leggendo la foto …',
  kiErkennungLeer: 'Nessuna foto letta finora.',
  kiErkennungGelesen: 'Ciò che è stato letto',
  kiErkennungNichtGelesen: 'Nulla da dire su',
  kiErkennungUnlesbar: 'illeggibile',
  kiErkennungOhneWert: 'nulla letto',
  kiErkennungArten: {
    name: 'Nome della carta',
    setcode: 'Set',
    nummer: 'Numero da collezione',
    typzeile: 'Riga del tipo',
    layout: 'Layout',
    kunstwerk: 'Illustrazione',
  } as Record<string, string>,
  kiErkennungKandidaten: 'Quali edizioni corrispondono',
  kiErkennungDrucke: 'Corrispondono {drucke} edizioni.',
  kiErkennungEinDruck: 'Corrisponde esattamente un’edizione.',
  kiErkennungWeitere:
    'E ce ne sono altre oltre — è stata contata solo la prima pagina.',
  kiErkennungOhneKandidat:
    'Nessuna edizione corrisponde a ciò che è stato letto. Può dipendere dalla foto oppure dal fatto che la carta qui non è catalogata.',
  kiErkennungGestuetzt: 'Si basa su: {arten}',
  kiErkennungOhneStuetze:
    'Nulla di questa edizione si ritrova sulla foto — compare solo perché la ricerca l’ha trovata per il termine.',
  kiErkennungNichtsAngelegt:
    'Non viene aggiunto nulla. Per registrare un esemplare usa il modulo qui sopra — così è chiaro quale edizione si intende.',
  kiErkennungStand: 'Aggiornamento della ricerca delle edizioni: {alter}',
  kiErkennungModell: 'Letto da {modell} presso {anbieter}',

  // **The citation question** — A13.4
  // (#1452).
  kiVorschlagBelegfrage:
    'Includi passaggi dalla base di conoscenza (facoltativo)',
  kiVorschlagBelegfrageHinweis:
    'cosa cercare — vuoto significa: non cercare affatto',
} as const

// **Die Texte der KI** — Werkstück A6 von Phase 5.3
// (#1445),
// erweitert für A7
// (#1446).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als einer
// ankommt** (§9): `seiteneinstellungen.ts` stand mit ihnen an der
// Zeilengrenze aus §2.
export const ki = {
  kimodelleTitel: 'AI models',
  kimodelleKeine:
    'No AI model is established. A provider\u2019s model list is not yet a demonstrated capability — only models that were actually checked appear here.',
  kiNurLokal: 'Local only is in force: no model outside this machine is asked.',
  kiAuchCloud:
    'Models at third-party providers are allowed. Whatever goes to them leaves this machine.',
  kiBeleg: '{faehigkeit} — checked on {datum}: {beleg}',
  kiKontextfenster: 'Context window: {tokens} tokens — {beleg}',
  kiKontextfensterUnbekannt: 'Context window: not stated',
  kiHerkunftUnbekannt:
    'Provenance unknown — it is not established which file actually runs here.',
  kiEinrichtungAus:
    'Not switched on. Nothing is asked, and nothing leaves this machine.',
  kiEinrichtungUnerreichbar:
    'Nothing answers at the configured address. This application starts nothing there and installs nothing.',
  kiEinrichtungOhneModell:
    'It answers but has no model. This application does not download one.',
  kiEinrichtungOhneFaehigkeit:
    'There are models, but none has an evidenced capability that is of use here.',
  kiEinrichtungBereit:
    'At least one model is ready with an evidenced capability.',
  kiLaufzeitOllama:
    'If nothing answers here, Ollama is not running. This application neither starts nor installs it — that happens in Ollama itself.',
  kiLaufzeitTitel: 'Local AI',
  kiLaufzeitFassung: 'Runtime: {fassung}',
  kiAdresse: 'Asking {adresse}.',
  kiAdresseVorgabe: 'the default address on this machine',
  kiCloudAus: 'Offloading to Ollama’s cloud is switched off.',
  kiCloudAn: 'Offloading to Ollama’s cloud is not switched off.',
  kiCloudUngewiss:
    'Whether offloading to Ollama’s cloud is switched off cannot be established from here.',
  kiCloudBeleg: 'Evidence: {beleg}',
  kiVorschlagTitel: 'Suggestion from a model',
  kiVorschlagModell: 'Model',
  kiVorschlagAuftrag: 'Task',
  kiVorschlagAuftragHinweis: 'What should the model do?',
  kiVorschlagen: 'Suggest',
  kiVorschlagLaeuft:
    'The model is working — the first time this can take a while because it is being loaded.',
  kiVorschlagAbbrechen: 'Cancel',
  kiVorschlagUebernehmen: 'Apply',
  kiVorschlagNichts: 'The model suggested no field that the template knows.',
  kiVorschlagAendertNichts:
    'A suggestion does not change the card. Only “Apply” writes.',
  kiVorschlagUebernommen: '{felder} fields were written.',
  kiVorschlagAblauf: 'Applicable until {ablauf}.',
  // **The second provider** — A8, #1447.
  kiAnbieterZeile: 'Provider {anbieter}: {stand}',
  kiVorschlagAnbieter: 'Provider and model',
  kiVorschlagModellWaehlen: 'Please choose',
  kiVorschlagOhneModell:
    'No model is available. What is missing is listed under “AI models”.',
  kiVorschlagAntwortVon: 'The answer came from {anbieter}/{modell}.',
  // **Measuring what a model can do** — A9, #1448.
  kiMessbarTitel: 'Models without a description',
  kiAdresseSocket: 'Asked over a Unix socket of this program — no port.',
  kiGeraetSitzung: 'Ask in this session',
  kiGeraetSitzungHinweis:
    'Applies until the next restart; then the “Apple on-device model” setting applies again.',
  kiFeldVorschlagen: 'Suggest',
  kiFeldVorschlagenFuer: '{feld} — suggestion',
  kiFeldAuftragName:
    'Suggest a fitting proper name for this card — not the card type and not a description of its rules. Answer in English, with the name only.',
  kiFeldAuftragFlavor:
    'Suggest a short flavor text for this card, one or two sentences: mood or story that does not repeat the rules text. Answer in English.',
  kiFeldKarte: 'The card so far:',
  kiFeldUebernehmen: 'Apply',
  kiFeldVon: 'from {anbieter} · {weg}',
  kiFeldWegGewaehlt: 'chosen',
  kiFeldWegVorrang: 'preferred',
  kiFeldVorrangFragt: 'Priority asks: {reihe}',
  kiFeldWegRueckfall: 'fallback',
  kiFeldAbgelehnt: 'The model declined: {grund}',
  kiFeldNichts: 'No suggestion came for this field.',
  kiMessbarSatz:
    '{anbieter} lists these models but says nothing about them. A measurement sends a small task with a fixed schema — it costs what that one task costs.',
  kiMessen: 'Measure',
  kiMessungVergessen: 'Discard measurement',
  kiMessungLaeuft: 'Measuring …',
  kiMessungGeglueckt: '{alias} answered: {faehigkeiten} are now evidenced.',
  // **Selection, comparison and provenance** — A11, #1450.
  kiVorschlagAuswahl: 'Apply: {feld}',
  kiVorschlagBisher: 'Currently: {wert}',
  kiVorschlagBisherLeer: 'Currently empty.',
  kiVorschlagNichtsGewaehlt: 'No field is ticked. What is ticked gets applied.',
  kiHerkunftTitel: 'Where the values come from',
  kiHerkunftZeile: '{feld}: {anbieter}/{modell} on {zeitpunkt}',
  kiHerkunftAufgabe: 'The task was: {aufgabe}',
  kiHerkunftVeraendert: 'Edited by hand since.',
  kiHerkunftKeine: 'No field of this card comes from a suggestion.',
  // **The image proposal** — A12, #1451.
  kiBildTitel: 'Image from a model',
  kiBildAendertNichts:
    'Generating does not change the card. What appears is a proposal with an expiry — you look at it, crop it, and only then insert it.',
  kiBildFeld: 'Image field',
  kiBildAufgabe: 'What should it show?',
  kiBildAufgabeHinweis: 'A bolt of lightning over a wide plain, oil painting',
  kiBildGroesse: 'Size',
  kiBildGuete: 'Quality',
  kiBildBearbeiten: 'Edit the existing image instead of generating a new one',
  kiBildErzeugen: 'Generate image',
  kiBildLaeuft:
    'The image is being made at {anbieter}/{modell} — this takes seconds to minutes.',
  kiBildAbbrechen: 'Cancel',
  kiBildAbgebrochen:
    'Cancelled. Nothing was taken over; whatever it cost until then is in the cost journal.',
  kiBildVorschau: 'Proposal for {feld}',
  kiBildMasse: '{breite} × {hoehe} pixels, {kilobytes} kB',
  kiBildAblauf: 'Valid until {ablauf}.',
  kiBildUmschrift: 'What was drawn: {umschrift}',
  kiBildNachweisTitel: 'Where this image came from',
  kiBildAngabeGroesse: 'Size',
  kiBildAngabeGuete: 'Quality',
  kiBildAngabeUmschrift: 'Revised prompt',
  kiBildAngabeMedientyp: 'File format',
  kiBildUebernehmen: 'Crop and insert',
  kiBildVerwerfen: 'Discard',
  kiBildVerworfen: 'The proposal is gone.',
  kiBildKosten:
    'An image order is released separately — the release for text does not cover it.',
  // **Approving a paid request** —
  // #1539.
  kiFreigabeTitel: 'This request costs money',
  kiFreigabeSatz:
    'The chosen model runs at {anbieter}. Approve a maximum amount — it covers this one request and expires after five minutes.',
  kiFreigabeBetrag: 'Maximum amount in the smallest unit (cents for EUR)',
  kiFreigabeVerfuegbar: 'Available today: {betrag}',
  kiFreigabeOhneBudget:
    'Nothing is approved for {anbieter}. Without an approved budget no paid request goes out — the limit is set under “Costs”.',
  kiFreigabeErteilen: 'Approve and ask',
  kiFreigabeVerzichten: 'Do not approve',
  kiFreigabeLaeuft: 'Issuing the approval …',

  // **The evidence-backed assessment** — A13.2a
  // (#1452).
  kiEinschaetzungTitel: 'Ask for an assessment',
  kiEinschaetzungErklaerung:
    'A model assesses this card using passages from the local search index. It changes nothing — neither the card nor the collection.',
  kiEinschaetzungArt: 'What about',
  kiEinschaetzungArtRegel: 'By the rules',
  kiEinschaetzungArtBalance: 'In comparison',
  kiEinschaetzungFrage: 'What should be assessed',
  kiEinschaetzungFragen: 'Get assessment',
  kiEinschaetzungLaeuft: 'The model is reading the evidence …',
  kiEinschaetzungLeer: 'Nothing assessed yet.',
  kiEinschaetzungBelegt: 'Everything mentioned appears in the evidence.',
  kiEinschaetzungUnbelegt:
    'This answer mentions something no passage contains. What that is appears below.',
  kiEinschaetzungBefunde: 'What is not backed by evidence',
  kiEinschaetzungEinschraenkungen: 'What was not considered',
  kiEinschaetzungBelege: 'The passages provided',
  kiEinschaetzungStaende: 'Source state',

  // **The balance analysis** — A13.2b
  // (#1452).
  kiEinschaetzungFormat: 'Format',
  kiEinschaetzungOhneFormat: 'none given',
  kiEinschaetzungVergleiche: 'Compared against',
  kiEinschaetzungStrategie: 'Found via',
  kiEinschaetzungOhneVergleich:
    'This answer names no comparison card. A statement about how a card stands is a comparison.',
  kiEinschaetzungSynergien: 'Which synergies were left out',

  // **The batch and the atomic group** — A13.1b
  // (#1452).
  kiBatchTitel: 'Several cards at once',
  kiBatchSatz:
    'One model answers the same task for each selected card. Nothing is written yet — that is the second step.',
  kiBatchZiele: '{gewaehlt} of {gesamt} cards selected',
  kiBatchOhneKarten: 'This set has no cards yet.',
  kiBatchOhneName: '(unnamed)',
  kiBatchAuftrag: 'Task for each card',
  kiBatchFahren: 'Get suggestions',
  kiBatchLaeuft: 'Running …',
  kiBatchLagen: 'What came of it',
  kiBatchGanzOderGar:
    'Everything is adopted at once, or nothing. If one target fails, every card stays as it was — and the suggestions stay open.',
  kiBatchGruppe: 'Adopt as a group ({karten})',
  kiBatchGruppeLaeuft: 'Adopting …',
  kiBatchGeschrieben: 'Wrote {karten} cards.',

  // **Card recognition from a photo** — A13.3
  // (#1452).
  kiErkennungTitel: 'Recognise a card from a photo',
  kiErkennungSatz:
    'A model reads what is printed on the photo. Which card that is gets decided by the same print search a typed list uses — and nothing is added to the collection.',
  kiErkennungLesen: 'Choose a photo and read it',
  kiErkennungLaeuft: 'The model is reading the photo …',
  kiErkennungLeer: 'No photo read yet.',
  kiErkennungGelesen: 'What was read',
  kiErkennungNichtGelesen: 'Nothing to say about',
  kiErkennungUnlesbar: 'not readable',
  kiErkennungOhneWert: 'nothing read',
  kiErkennungArten: {
    name: 'Card name',
    setcode: 'Set',
    nummer: 'Collector number',
    typzeile: 'Type line',
    layout: 'Layout',
    kunstwerk: 'Artwork',
  } as Record<string, string>,
  kiErkennungKandidaten: 'Which printings match',
  kiErkennungDrucke: '{drucke} printings match.',
  kiErkennungEinDruck: 'Exactly one printing matches.',
  kiErkennungWeitere:
    'And there are more behind — only the first page was counted.',
  kiErkennungOhneKandidat:
    'No printing matches what was read. That may be the photo, or the card may not be listed here.',
  kiErkennungGestuetzt: 'Supported by: {arten}',
  kiErkennungOhneStuetze:
    'Nothing on this printing can be found again on the photo — it is listed only because the search found it for the term.',
  kiErkennungNichtsAngelegt:
    'Nothing is added. To record a copy, use the form above — then it is settled which printing is meant.',
  kiErkennungStand: 'Age of the print search: {alter}',
  kiErkennungModell: 'Read by {modell} at {anbieter}',

  // **The citation question** — A13.4
  // (#1452).
  kiVorschlagBelegfrage: 'Include passages from the knowledge base (optional)',
  kiVorschlagBelegfrageHinweis:
    'what to search for — empty means: do not search',
} as const

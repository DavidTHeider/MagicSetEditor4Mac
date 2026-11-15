// **Die Texte der KI** — Werkstück A6 von Phase 5.3
// (#1445),
// erweitert für A7
// (#1446).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als einer
// ankommt** (§9): `seiteneinstellungen.ts` stand mit ihnen an der
// Zeilengrenze aus §2.
export const ki = {
  kimodelleTitel: 'Modèles d’IA',
  kimodelleKeine:
    'Aucun modèle d’IA n’est établi. La liste de modèles d’un fournisseur n’est pas encore une capacité démontrée — seuls les modèles réellement vérifiés apparaissent ici.',
  kiNurLokal:
    'Le mode « local uniquement » est actif : aucun modèle en dehors de ce Mac n’est interrogé.',
  kiAuchCloud:
    'Les modèles de fournisseurs tiers sont autorisés. Tout ce qui leur est envoyé quitte ce Mac.',
  kiBeleg: '{faehigkeit} — vérifié le {datum} : {beleg}',
  kiKontextfenster: 'Fenêtre de contexte : {tokens} jetons — {beleg}',
  kiKontextfensterUnbekannt: 'Fenêtre de contexte : non indiquée',
  kiHerkunftUnbekannt:
    'Provenance inconnue — on ne sait pas quel fichier s’exécute réellement ici.',
  kiEinrichtungAus: 'Non activé. Rien n’est demandé, et rien ne quitte ce Mac.',
  kiEinrichtungUnerreichbar:
    'Personne ne répond à l’adresse configurée. Cette application n’y démarre rien et n’y installe rien.',
  kiEinrichtungOhneModell:
    'Il répond mais n’a aucun modèle. Cette application n’en télécharge aucun.',
  kiEinrichtungOhneFaehigkeit:
    'Il y a des modèles, mais aucun n’a de capacité démontrée qui serait utile ici.',
  kiEinrichtungBereit:
    'Au moins un modèle est prêt, avec une capacité démontrée.',
  kiLaufzeitOllama:
    'Si rien ne répond ici, Ollama n’est pas en cours d’exécution. Cette application ne le démarre pas et ne l’installe pas — cela se passe dans Ollama lui-même.',
  kiLaufzeitTitel: 'IA locale',
  kiLaufzeitFassung: 'Environnement d’exécution : {fassung}',
  kiAdresse: 'Requête envoyée à {adresse}.',
  kiAdresseVorgabe: 'l’adresse par défaut sur ce Mac',
  kiCloudAus: 'Le déport vers le cloud d’Ollama est désactivé.',
  kiCloudAn: 'Le déport vers le cloud d’Ollama n’est pas désactivé.',
  kiCloudUngewiss:
    'On ne peut pas établir d’ici si le déport vers le cloud d’Ollama est désactivé.',
  kiCloudBeleg: 'Preuve : {beleg}',
  kiVorschlagTitel: 'Suggestion d’un modèle',
  kiVorschlagModell: 'Modèle',
  kiVorschlagAuftrag: 'Tâche',
  kiVorschlagAuftragHinweis: 'Que doit faire le modèle ?',
  kiVorschlagen: 'Suggérer',
  kiVorschlagLaeuft:
    'Le modèle travaille — la première fois, cela peut prendre un moment car il doit être chargé.',
  kiVorschlagAbbrechen: 'Annuler',
  kiVorschlagUebernehmen: 'Appliquer',
  kiVorschlagNichts:
    'Le modèle n’a suggéré aucun champ que le modèle de carte connaisse.',
  kiVorschlagAendertNichts:
    'Une suggestion ne modifie pas la carte. Seul « Appliquer » écrit.',
  kiVorschlagUebernommen: '{felder} champs ont été écrits.',
  kiVorschlagAblauf: 'Applicable jusqu’à {ablauf}.',
  // **The second provider** — A8, #1447.
  kiAnbieterZeile: 'Fournisseur {anbieter} : {stand}',
  kiVorschlagAnbieter: 'Fournisseur et modèle',
  kiVorschlagModellWaehlen: 'Choisis',
  kiVorschlagOhneModell:
    'Aucun modèle n’est disponible. Ce qui manque est indiqué sous « Modèles d’IA ».',
  kiVorschlagAntwortVon: 'La réponse vient de {anbieter}/{modell}.',
  // **Measuring what a model can do** — A9, #1448.
  kiMessbarTitel: 'Modèles sans description',
  kiAdresseSocket:
    'Interrogé via un socket Unix propre à ce programme — aucun port.',
  kiGeraetSitzung: 'Interroger pendant cette session',
  kiGeraetSitzungHinweis:
    'Valable jusqu’au prochain redémarrage ; ensuite le réglage « Modèle sur l’appareil d’Apple » s’applique de nouveau.',
  kiFeldVorschlagen: 'Proposer',
  kiFeldVorschlagenFuer: '{feld} — proposition',
  kiFeldAuftragName:
    'Propose un nom propre adapté à cette carte — ni le type de carte, ni une description de ses règles. Réponds en français, uniquement avec le nom.',
  kiFeldAuftragFlavor:
    'Propose un court texte d’ambiance pour cette carte, une ou deux phrases : une atmosphère ou une histoire qui ne répète pas le texte de règles. Réponds en français.',
  kiFeldKarte: 'La carte jusqu’ici :',
  kiFeldUebernehmen: 'Reprendre',
  kiFeldVon: 'de {anbieter} · {weg}',
  kiFeldWegGewaehlt: 'choisi',
  kiFeldWegVorrang: 'prioritaire',
  kiFeldVorrangFragt: 'La priorité interroge : {reihe}',
  kiFeldWegRueckfall: 'repli',
  kiFeldAbgelehnt: 'Le modèle a refusé : {grund}',
  kiFeldNichts: 'Aucune proposition pour ce champ.',
  kiMessbarSatz:
    '{anbieter} liste ces modèles mais ne dit rien à leur sujet. Une mesure envoie une petite tâche à schéma fixe — elle coûte ce que coûte cette unique tâche.',
  kiMessen: 'Mesurer',
  kiMessungVergessen: 'Annuler la mesure',
  kiMessungLaeuft: 'Mesure en cours …',
  kiMessungGeglueckt:
    '{alias} a répondu : {faehigkeiten} sont désormais démontrées.',
  // **Selection, comparison and provenance** — A11, #1450.
  kiVorschlagAuswahl: 'Appliquer : {feld}',
  kiVorschlagBisher: 'Actuellement : {wert}',
  kiVorschlagBisherLeer: 'Actuellement vide.',
  kiVorschlagNichtsGewaehlt:
    'Aucun champ n’est coché. Ce qui est coché est appliqué.',
  kiHerkunftTitel: 'D’où viennent les valeurs',
  kiHerkunftZeile: '{feld} : {anbieter}/{modell} le {zeitpunkt}',
  kiHerkunftAufgabe: 'La tâche était : {aufgabe}',
  kiHerkunftVeraendert: 'Modifié à la main depuis.',
  kiHerkunftKeine: 'Aucun champ de cette carte ne provient d’une suggestion.',
  // **The image proposal** — A12, #1451.
  kiBildTitel: 'Image d’un modèle',
  kiBildAendertNichts:
    'La génération ne modifie pas la carte. Ce qui apparaît est une proposition avec une date d’expiration — tu la regardes, la recadres, et seulement ensuite l’insères.',
  kiBildFeld: 'Champ d’image',
  kiBildAufgabe: 'Que doit-elle montrer ?',
  kiBildAufgabeHinweis:
    'Un éclair au-dessus d’une vaste plaine, peinture à l’huile',
  kiBildGroesse: 'Taille',
  kiBildGuete: 'Qualité',
  kiBildBearbeiten:
    'Modifier l’image existante au lieu d’en générer une nouvelle',
  kiBildErzeugen: 'Générer l’image',
  kiBildLaeuft:
    'L’image est en cours de création chez {anbieter}/{modell} — cela prend de quelques secondes à quelques minutes.',
  kiBildAbbrechen: 'Annuler',
  kiBildAbgebrochen:
    'Annulé. Rien n’a été repris ; ce que cela a coûté jusque-là figure dans le journal des coûts.',
  kiBildVorschau: 'Proposition pour {feld}',
  kiBildMasse: '{breite} × {hoehe} pixels, {kilobytes} ko',
  kiBildAblauf: 'Valable jusqu’au {ablauf}.',
  kiBildUmschrift: 'Ce qui a été dessiné : {umschrift}',
  kiBildNachweisTitel: 'D’où vient cette image',
  kiBildAngabeGroesse: 'Taille',
  kiBildAngabeGuete: 'Qualité',
  kiBildAngabeUmschrift: 'Prompt révisé',
  kiBildAngabeMedientyp: 'Format de fichier',
  kiBildUebernehmen: 'Recadrer et insérer',
  kiBildVerwerfen: 'Rejeter',
  kiBildVerworfen: 'La proposition a disparu.',
  kiBildKosten:
    'Une commande d’image est autorisée séparément — l’autorisation pour le texte ne la couvre pas.',
  // **Approving a paid request** —
  // #1539.
  kiFreigabeTitel: 'Cette requête coûte de l’argent',
  kiFreigabeSatz:
    'Le modèle choisi s’exécute chez {anbieter}. Autorise un montant maximal — il couvre cette seule requête et expire au bout de cinq minutes.',
  kiFreigabeBetrag:
    'Montant maximal dans la plus petite unité (centimes pour l’EUR)',
  kiFreigabeVerfuegbar: 'Disponible aujourd’hui : {betrag}',
  kiFreigabeOhneBudget:
    'Rien n’est autorisé pour {anbieter}. Sans budget autorisé, aucune requête payante ne part — la limite se fixe sous « Coûts ».',
  kiFreigabeErteilen: 'Autoriser et demander',
  kiFreigabeVerzichten: 'Ne pas autoriser',
  kiFreigabeLaeuft: 'Émission de l’autorisation …',

  // **The evidence-backed assessment** — A13.2a
  // (#1452).
  kiEinschaetzungTitel: 'Demander une évaluation',
  kiEinschaetzungErklaerung:
    'Un modèle évalue cette carte à partir de passages de l’index de recherche local. Il ne modifie rien — ni la carte, ni la collection.',
  kiEinschaetzungArt: 'De quoi s’agit-il',
  kiEinschaetzungArtRegel: 'Selon les règles',
  kiEinschaetzungArtBalance: 'En comparaison',
  kiEinschaetzungFrage: 'Ce qui doit être évalué',
  kiEinschaetzungFragen: 'Obtenir l’évaluation',
  kiEinschaetzungLaeuft: 'Le modèle lit les preuves …',
  kiEinschaetzungLeer: 'Rien n’a encore été évalué.',
  kiEinschaetzungBelegt: 'Tout ce qui est mentionné figure dans les preuves.',
  kiEinschaetzungUnbelegt:
    'Cette réponse mentionne quelque chose qu’aucun passage ne contient. Ce dont il s’agit figure ci-dessous.',
  kiEinschaetzungBefunde: 'Ce qui n’est pas étayé',
  kiEinschaetzungEinschraenkungen: 'Ce qui n’a pas été pris en compte',
  kiEinschaetzungBelege: 'Les passages fournis',
  kiEinschaetzungStaende: 'État des sources',

  // **The balance analysis** — A13.2b
  // (#1452).
  kiEinschaetzungFormat: 'Format',
  kiEinschaetzungOhneFormat: 'aucun indiqué',
  kiEinschaetzungVergleiche: 'Comparé avec',
  kiEinschaetzungStrategie: 'Trouvé via',
  kiEinschaetzungOhneVergleich:
    'Cette réponse ne nomme aucune carte de comparaison. Une affirmation sur la position d’une carte est une comparaison.',
  kiEinschaetzungSynergien: 'Quelles synergies ont été laissées de côté',

  // **The batch and the atomic group** — A13.1b
  // (#1452).
  kiBatchTitel: 'Plusieurs cartes à la fois',
  kiBatchSatz:
    'Un modèle répond à la même tâche pour chaque carte sélectionnée. Rien n’est encore écrit — c’est la seconde étape.',
  kiBatchZiele: '{gewaehlt} cartes sélectionnées sur {gesamt}',
  kiBatchOhneKarten: 'Ce set ne contient pas encore de carte.',
  kiBatchOhneName: '(sans nom)',
  kiBatchAuftrag: 'Tâche pour chaque carte',
  kiBatchFahren: 'Obtenir des suggestions',
  kiBatchLaeuft: 'En cours …',
  kiBatchLagen: 'Ce qu’il en est résulté',
  kiBatchGanzOderGar:
    'Tout est repris en une fois, ou rien. Si une cible échoue, chaque carte reste comme elle était — et les suggestions restent ouvertes.',
  kiBatchGruppe: 'Reprendre en groupe ({karten})',
  kiBatchGruppeLaeuft: 'Reprise en cours …',
  kiBatchGeschrieben: '{karten} cartes écrites.',

  // **Card recognition from a photo** — A13.3
  // (#1452).
  kiErkennungTitel: 'Reconnaître une carte à partir d’une photo',
  kiErkennungSatz:
    'Un modèle lit ce qui est imprimé sur la photo. La carte dont il s’agit est déterminée par la même recherche d’impressions qu’une liste saisie — et rien n’est ajouté à la collection.',
  kiErkennungLesen: 'Choisir une photo et la lire',
  kiErkennungLaeuft: 'Le modèle lit la photo …',
  kiErkennungLeer: 'Aucune photo lue pour l’instant.',
  kiErkennungGelesen: 'Ce qui a été lu',
  kiErkennungNichtGelesen: 'Ce sur quoi il n’y avait rien à dire',
  kiErkennungUnlesbar: 'illisible',
  kiErkennungOhneWert: 'rien de lu',
  kiErkennungArten: {
    name: 'Nom de la carte',
    setcode: 'Set',
    nummer: 'Numéro de collector',
    typzeile: 'Ligne de type',
    layout: 'Disposition',
    kunstwerk: 'Illustration',
  } as Record<string, string>,
  kiErkennungKandidaten: 'Quelles impressions correspondent',
  kiErkennungDrucke: '{drucke} impressions correspondent.',
  kiErkennungEinDruck: 'Exactement une impression correspond.',
  kiErkennungWeitere:
    'Et il y en a d’autres derrière — seule la première page a été comptée.',
  kiErkennungOhneKandidat:
    'Aucune impression ne correspond à ce qui a été lu. Cela peut venir de la photo, ou la carte n’est pas répertoriée ici.',
  kiErkennungGestuetzt: 'Repose sur : {arten}',
  kiErkennungOhneStuetze:
    'Rien de cette impression ne se retrouve sur la photo — elle figure là uniquement parce que la recherche l’a trouvée pour le terme.',
  kiErkennungNichtsAngelegt:
    'Rien n’est ajouté. Pour enregistrer un exemplaire, utilise le formulaire ci-dessus — l’impression concernée y est alors précisée.',
  kiErkennungStand: 'Ancienneté de la recherche d’impressions : {alter}',
  kiErkennungModell: 'Lu par {modell} chez {anbieter}',

  // **The citation question** — A13.4
  // (#1452).
  kiVorschlagBelegfrage:
    'Joindre des passages de la base de connaissances (facultatif)',
  kiVorschlagBelegfrageHinweis:
    'ce qu’il faut rechercher — vide signifie : ne pas rechercher',
} as const

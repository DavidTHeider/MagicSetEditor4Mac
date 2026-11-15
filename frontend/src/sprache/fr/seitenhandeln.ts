// The texts of the **trading** area — Werkstück T9 von Phase 5.6
// (#1474).
//
// **Ein eigener Katalog und keine Zeilen mehr in `seitentechnik.ts`.** Der
// Abgleich, die Post und die Verbuchung standen dort, weil es den Bereich
// noch nicht gab — ihre eigenen Kommentare sagen es. Jetzt gibt es ihn, und
// ein Text über das Handeln unter „Technik" wäre eine Fundstelle, die
// niemand sucht.

import { handelsliste } from './handelsliste'
import { handelsassistent } from './handelsassistent'

export const handeln = {
  liste: handelsliste,
  assistent: handelsassistent,
  titel: 'Échanges',
  erklaerung:
    'Échange avec un correspondant que tu as confirmé. Tout est calculé sur cet ordinateur ; ce qui en sort relève de ta décision.',
  gebiete: 'Domaines des échanges',
  gebietPartner: 'Avec qui',
  gebietMarkt: 'Ce que cela vaut',
  neuerTausch: 'Nouvel échange …',
  neuerTauschTitel: 'Nouvel échange',
  partnerTitel: 'Correspondants confirmés',
  partnerErklaerung:
    'Tu n’échanges qu’avec un correspondant avec lequel tu as comparé un code. Choisis-en un et l’échange s’ouvre ; pour inviter et confirmer, utilise l’encadré ci-dessous.',
  partnerLeer:
    'Aucun correspondant confirmé pour l’instant. Échangez vos codes d’invitation dans l’encadré ci-dessous et comparez les chiffres.',
  partnerWaehlen: 'Ouvrir',
  partnerWaehlenName: 'Ouvrir {name}',
  partnerOffen: 'Ouvert',
  partnerOffenName: '{name} ouvert',
  partnerKeiner:
    'Choisis un échange à gauche — ou lances-en un avec « Nouvel échange … ».',
  marktTitel: 'Ce que vaut ta collection',
  marktErklaerung:
    'Une observation du marché n’est pas une promesse : elle indique ce qui a été demandé ailleurs, pas ce que tu obtiendras. L’origine de chaque chiffre et son ancienneté sont indiquées à côté ; là où il n’y en a pas, aucun chiffre n’est affiché.',
  vorgangTitel: 'Un échange, de l’ouverture à la remise',
  vorgangErklaerung:
    'C’est ici que se déroule le contrat : ouvrir, proposer, accepter, préparer, remettre. **Aucun stock ne bouge ici** — l’enregistrement se fait dans « Enregistrement », et ce programme te le redemande.',
  wahlTitel: 'Quel échange',
  wahlNeu: 'Un nouvel échange',
  wahlNeuZusatz: 'Le programme attribue son identifiant à l’ouverture.',
  wahlEintrag: '{zustand} · révision {revision} · {positionen}',
  wahlLeer: 'Il n’y a pas encore d’échange avec ce correspondant.',
  wahlFachleute: 'Ouvrir un échange par son identifiant',
  wahlFachleuteErklaerung:
    'Chaque échange a un **identifiant**. Le programme l’attribue à l’ouverture, et chaque message le concernant le porte. Si un échange ne figure pas encore ci-dessus — par exemple parce que son premier message est encore en route —, ouvre-le ici grâce à cet identifiant.',
  wahlKennung: 'Identifiant de l’échange',
  wahlOeffnen: 'Ouvrir',
  wahlGewaehlt: 'L’échange {kennung} est sélectionné.',
  vorgangOhneKennung: 'Choisis un échange ci-dessus — ou ouvres-en un nouveau.',
  vorgangAnsehen: 'Afficher l’état',
  vorgangFachleute: 'Afficher l’état avec version et empreinte',
  vorgangEroeffnen: 'Ouvrir',
  vorgangVorschlagen: 'Proposer à nouveau',
  vorgangRevision: 'Version',
  vorgangAbdruck: 'Empreinte',
  vorgangParteien: 'Parties',
  vorgangZustimmungen: 'Ont accepté jusqu’ici',
  vorgangKeine: 'Personne pour l’instant.',
  vorgangDu: 'toi',
  vorgangEnde: 'Il n’y a pas de suite.',
  vorgangArt: 'Type',
  arten: {
    teilen: 'Partage',
    deckleihe: 'Prêt de decks',
    tausch: 'Troc',
    markt: 'Marché',
  },
  artOhneBestand:
    'Cet échange ne déplace aucun stock — seul un troc est enregistré.',
  vorgangVereinbarung:
    'Vous êtes d’accord. C’est un **accord** : les cartes ne changent de mains qu’à la remise, et il n’est compté que lorsque les deux parties l’ont enregistré.',
  vorgangSchritte: 'Étapes',
  schrittAnbieten: 'Proposer',
  schrittGegenangebot: 'Contre-proposition',
  schrittZustimmen: 'Accepter',
  schrittVorbereiten: 'Préparer',
  schrittUebergeben: 'Confirmer la remise',
  schrittAblehnen: 'Refuser',
  schrittAbbrechen: 'Annuler',
  schrittAblaufen: 'Expiré',
  schrittKlaeren: 'Signaler pour clarification',
  positionTitel: 'Ce qui change de mains',
  positionVon: 'De qui',
  positionVonIch: 'De toi',
  positionVonGegenueber: 'De ton correspondant',
  positionPrinting: 'Impression',
  kennungVonHand: 'Saisir l’identifiant à la main',
  positionMenge: 'Nombre',
  positionGuete: 'État',
  positionSprache: 'Langue',
  positionFinish: 'Finition',
  positionHinzufuegen: 'Ajouter une ligne',
  anschlussWartet:
    '« {name} » vient de ta collection — vérifie la quantité et le côté, puis elle figure dans l’offre.',
  wunschTitel: 'Liste de souhaits issue d’un deck',
  wunschWartet: '{name}, issu d’un deck, pas encore saisi : {liste}.',
  wunschEintragen: 'Saisir comme souhait auprès du correspondant',
  anschlussOhnePartner:
    '« {name} » attend — choisis un correspondant sous « Nouvel échange … », puis elle passe sur sa ligne.',
  preise: {
    holen: 'Récupérer les prix',
    abbrechen: 'Annuler la récupération des prix',
    laeuft: 'Récupération des prix … {fertig} sur {gesamt}',
    fertig: 'Prix récupérés pour {zahl} impressions.',
    teils: '{zahl} impressions sur {gesamt} sans prix : {meldung}',
    abgebrochen:
      'Annulé après {fertig} sur {gesamt} — ce qui a été récupéré reste.',
  },
  markttisch: {
    neuTitel: 'Nouvel échange avec',
    neuMit: 'Nouvel échange avec {name}',
    meinFach: 'Mon casier',
    sammlung: 'Ma collection',
    wunschliste: 'Liste de souhaits',
    seinFach: 'Casier de {name}',
    freigegeben:
      'Ce que {name} partage pour l’échange — demandé à l’ouverture.',
    meinLeer:
      'Rien ici pour l’instant. Les exemplaires arrivent par la collection.',
    wunschLeer:
      'Aucun souhait. Une carte devient un souhait dans la collection.',
    seinLeer: 'Rien de partagé.',
    neuFragen: 'Redemander',
    hinweis: 'Glisser, ou Espace, flèches, Entrée. Sélection multiple avec ⌘.',
    laedt: 'Demande …',
    karte: '{zustand} · {sprache} · {menge}×',
    ansicht: 'Affichage',
    alsTisch: 'En table',
    alsListe: 'En liste',
    fachSuchen: 'Rechercher dans le casier',
    fachZaehlung: '{gezeigt} sur {gesamt}',
    fachMehr: 'Afficher plus',
    fachKeinTreffer: 'Aucune carte de ce casier ne porte ce nom.',
  },
  zug: {
    veraltet:
      'L’échange a changé entre-temps. Regarde le nouvel état et décide à nouveau.',
    vorbereitenUebergeben: 'Préparer et remettre',
    uebergebenVerbuchen: 'Remettre et enregistrer',
    anbietenZustimmen: 'Proposer et accepter',
    buchungHinweis:
      '« Enregistrer » signale ensuite le rapprochement et confirme ta clôture. La confirmation de {name}, tu la saisis à part.',
    festhaltenOffen:
      'On ne sait pas quel exemplaire part. Choisis-le sous « Réservations » et appuie encore une fois.',
    gescheitert: '{schritt} n’a pas abouti : {meldung}',
    lesen:
      'C’est enregistré, mais l’état n’a pas pu être lu ensuite : {meldung}',
    ausgang:
      'C’est enregistré chez toi, mais ce n’est pas arrivé dans la boîte d’envoi : {meldung}',
    unterwegs:
      'C’est enregistré chez toi ; {name} ne l’a pas encore : {meldung}',
    hinaus: 'Envoyé à {name}.',
  },
  sicht: {
    titel: 'Ce que ton partenaire partage',
    anfragen: 'Demander la vue partagée',
    laeuft: 'Demande par la connexion …',
    zeilen: '{zahl} lignes partagées, dont {wuensche} recherchées.',
    keinVorschlag:
      'Rien ne correspond : aucun souhait ici ne rencontre une offre là-bas.',
    gibst: 'Tu donnes',
    bekommst: 'Tu reçois',
    zeile: '{menge} × {name}',
    uebernehmen: 'Reprendre la proposition',
    uebernommen: 'Repris : {zahl} lignes figurent dans l’offre.',
  },
  druck: {
    eigen: 'impression personnelle · sans image',
    ohneNamen: 'Impression sans nom',
    unbekannt: 'Impression inconnue',
    bezug: '{code} #{nummer}',
    laedt: '…',
  },
  positionEntfernen: 'Supprimer la ligne',
  positionEntfernenName: 'Supprimer la ligne {nummer}',
  positionLeer:
    'Aucune ligne pour l’instant. Une offre sans position n’indique pas de quoi il s’agit.',
  positionVorbereitung:
    'Lequel de tes exemplaires part, tu l’indiques dans « Réservations » : une offre désigne une **impression**, pas une ligne de ton stock.',
  zusagenTitel: 'Quels exemplaires partent',
  zusagenErklaerung:
    'Une offre désigne une **impression**, pas une ligne de ton stock. Indique ici **quel** exemplaire part — sinon rien n’indique à l’enregistrement sur quoi prélever.',
  zusagenWahl: 'Lequel de tes exemplaires part',
  zusagenPosition: '{menge} × {name}',
  zusagenZeile: '{lagerort} · {zustand} · {sprache} · {finish} · {stueck}',
  zusagenOhneLagerort: 'sans emplacement',
  zusagenKeinExemplar: 'Rien de cela ne figure dans ta collection.',
  zusagenKeinePassende:
    'Aucune ligne de ta collection ne correspond à l’état, à la langue et à la finition de cette ligne. Ajoute-la à ta collection, ou ajoute de nouveau la ligne à l’échange avec des indications correspondantes.',
  zusagenFachleute: 'Réserver une ligne de collection par son identifiant',
  zusagenFachleuteErklaerung:
    'Une **ligne de collection** est une ligne de ta collection, et sa **version** compte chaque modification. Les deux figurent déjà dans le choix ci-dessus — ici, tu les saisis à la main, par exemple pour un exemplaire qui n’y figure pas.',
  zusagenExemplar: 'Quelle ligne de stock',
  zusagenMenge: 'Combien',
  zusagenFassung: 'Version consultée',
  zusagenFesthalten: 'Réserver',
  zusagenNachsehen: 'Afficher les réservations',
  zusagenOhneVorgang:
    'Choisis un échange ci-dessus ; ce que tu donnes en échange s’affiche alors ici.',
  zusagenLeer: 'Rien n’est réservé pour cet échange.',
  zusagenOffen: 'réservé',
  zusagenUebergeben: 'remis',
  ausgangTitel: 'Ce qui part vers ton correspondant',
  ausgangErklaerung:
    'L’état est signé et placé dans la **boîte d’envoi**. **Envoyer maintenant** ouvre une connexion vers ton correspondant et transmet ce qui attend — cela y reste malgré tout jusqu’à ce que ton correspondant le confirme.',
  ausgangSchicken: 'Placer dans la boîte d’envoi',
  ausgangNachricht: 'Identifiant du message',
  ausgangWegraeumen: 'Bien arrivé — supprimer',
  ausgangWegraeumenName: 'Bien arrivé — supprimer : {nachricht}',
  ausgangFachleute: 'Supprimer un message par son identifiant',
  ausgangFachleuteErklaerung:
    'Chaque message de la boîte d’envoi a un **identifiant**. Ce qui reste en attente après une interruption figure ci-dessus avec un bouton pour le supprimer ; ici, tu en supprimes un qui n’y figure pas.',
  ausgangBestaetigen: 'Supprimer comme bien arrivé',
  ausgangOhneVorgang: 'Choisis ci-dessus un échange qui doit partir.',
  ausgangOhneNachricht: 'Saisis d’abord l’identifiant du message.',
  ausgangWiederaufnahme: 'Vérifier après une déconnexion',
  ausgangZustellen: 'Envoyer maintenant',
  ausgangOhneGegenueber: 'Choisis d’abord à qui envoyer.',
  ausgangGetragen: '{zahl} envoyé(s).',
  ausgangKlaerung:
    'Deux historiques s’opposent. Cela ne peut pas se calculer — parlez-en ensemble.',
  ausgangNachzuliefern: 'Doit encore partir',
  ausgangLeer: 'La boîte d’envoi est vide.',
  ausgangFolge: 'Séquence',
  tauschTitel: 'Ce que vous pourriez échanger',
  tauschErklaerung:
    'Colle ce que ton correspondant t’a envoyé — ses doubles et ses souhaits, exactement tels que son programme les a partagés. La comparaison s’effectue sur cet ordinateur ; rien n’en sort.',
  tauschStand: 'Ce que ton correspondant a envoyé',
  tauschAbgleichen: 'Comparer',
  tauschAusUmschlag: 'Lire depuis une enveloppe',
  tauschUnlesbar: 'Ce n’est pas un état partagé.',
  tauschLeer: 'Rien ne correspond.',
  tauschGebe: 'Tu donnes',
  tauschBekomme: 'Tu reçois',
  tauschVorschlag: 'Une proposition',
  tauschOhneVorschlag:
    'Une proposition exige quelque chose de chaque côté — un échange où l’un ne donne rien est un cadeau.',
  tauschBefunde: 'Ce qui n’a pas pu être comparé',
  postLeer: 'Rien en transit.',
  postOffen: 'En transit',
  postVersuche: 'Tentatives',
  buchungTitel: 'Enregistrer un échange',
  buchungErklaerung:
    'Une fois que les cartes ont effectivement changé de mains, enregistre l’échange ici. Ta collection change, et le journal conserve la trace de ce qui est allé où.',
  buchungVerbuchen: 'Enregistrer',
  buchungAbgleichen: 'Signaler le rapprochement',
  buchungEigeneStimme: 'Confirmer la clôture',
  buchungBestaetigen: 'Autre partie confirmée',
  buchungNochNicht:
    'Rien à enregistrer pour l’instant : on enregistre une fois la remise confirmée.',
  buchungZustand: 'État',
  buchungStandEntwurf: 'Brouillon',
  buchungStandAngeboten: 'Proposé',
  buchungStandVerhandlung: 'En négociation',
  buchungStandAngenommen: 'Accepté',
  buchungStandVorbereitet: 'Préparé',
  buchungStandUebergabeOffen: 'Remise confirmée',
  buchungStandLokalVerbucht: 'Enregistré de ton côté',
  buchungStandAbgleichOffen: 'Rapprochement en cours',
  buchungStandBeidseitigVerbucht: 'Enregistré des deux côtés',
  buchungStandAbgelehnt: 'Refusé',
  buchungStandAbgebrochen: 'Annulé',
  buchungStandAbgelaufen: 'Expiré',
  buchungStandKlaerung: 'À clarifier',
  buchungBeleg: 'Justificatif',
  buchungEinseitig: 'Seule ta signature est présente.',
  buchungBeidseitig: 'Les deux parties ont signé.',
  buchungOhneGegenbeleg:
    'Aucun justificatif vérifié de l’autre partie pour l’instant.',
  buchungGegenbeleg: 'Justificatif vérifié de {wer}, reçu {wann}.',
  buchungGegenbelegAbgewiesen:
    'Un justificatif de {wer} est arrivé {wann}, mais il ne correspond pas au tien — {mangel}',
  buchungGegenbelegRat:
    'Rapproche l’échange avec {wer} ; un justificatif concordant remplace ce message.',
  gegenbelegMaengel: {
    unterschrift: 'la signature ne vient pas de l’expéditeur.',
    vorgang: 'il appartient à un autre échange.',
    revision: 'il appartient à un autre état de cet échange.',
    parteien: 'il nomme d’autres parties.',
    zeilen: 'les cartes comptabilisées diffèrent.',
  },
  buchungJournal: 'Ce qui a bougé',
  buchungJournalLeer: 'Rien dans le journal pour cet échange pour l’instant.',
  kontoTitel: 'Ton propre compte sur une place de marché',
  kontoErklaerung:
    'Ce que **tu** y proposes et ce que **tu** y as vendu ou acheté. Il est uniquement lu : aucune annonce n’est créée, rien n’est mis au panier, rien n’est acheté et rien n’est expédié. Cette fonction est **désactivée** au départ — sous Réglages → Services, en plus de la place de marché elle-même.',
  kontoNachsehen: 'Consulter',
  kontoAngebote: 'Annonces',
  kontoBestellungen: 'Commandes',
  kontoBewegt: 'a déplacé du stock',
  kontoBewegtNicht: 'n’a pas déplacé de stock',
} as const

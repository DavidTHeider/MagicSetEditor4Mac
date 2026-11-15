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

  titel: 'Jouer',
  erklaerung:
    'Une partie est une table avec des decks figés. Ce qui se passe ici est un flux d’événements — chacun dépend du précédent, et aucun n’est écrasé. Si tu déplaces une carte, tu vois aussitôt ce que les autres en voient.',
  anfangTitel: 'Commencer une partie',
  anfangErklaerung:
    'Chaque siège apporte un deck, et il est figé : modifier ton deck ensuite ne modifie pas cette partie. Le jeu est déterminé par le premier deck choisi.',
  leerDecksTitel: 'Pas encore de deck',
  leerDecks: 'Pour jouer, il faut un deck. Crées-en un sous Decks.',
  sitz: 'Siège',
  sitzName: 'Nom à la table',
  sitzDeck: 'Deck',
  sitzLeihe: 'Emprunté à …{von} · {karten} cartes',
  deckWaehlen: 'Choisir un deck',
  anschlussWartet:
    '« {name} » a été apporté et occupe le premier siège. Le deuxième est libre.',
  sitzHinzu: 'Ajouter un siège',
  sitzWeg: 'Retirer ce siège',
  sitzGruppe: 'Place {nummer}',
  spiel: 'Jeu',
  anfangen: 'Commencer la partie',
  tischTitel: 'À la table',
  partieGegen: 'Partie contre {namen}',
  partieZwischen: 'Partie entre {namen}',
  partieVon: 'Partie en solo : {namen}',
  fassung: 'Révision',
  wirt: 'Hôte',
  wirtHinweis:
    'L’hôte détient l’état complet de cette partie et voit donc plus que ce que sa propre vue montre. C’est indiqué ici, car cela doit être su avant de rejoindre la partie, et non après.',
  tischAufHaufen: '{zahl} en tas',
  tischMarkeEine: '{zahl} marqueur {art}',
  tischMarkeViele: '{zahl} marqueurs {art}',
  tischArt: {
    spielstein: 'Jeton',
    emblem: 'Emblème',
    faehigkeit: 'Capacité sur la pile',
  },

  sicht: 'Vue pour',
  sichtZuschauer: 'Spectateur (sans siège)',
  sichtWechseln: 'Changer de siège',
  sichtLaedt: 'Chargement de la vue choisie …',
  verlassen: 'Fermer la partie',
  zonen: 'Ce qui se trouve où',
  zoneBibliothek: 'Bibliothèque',
  zoneHand: 'Main',
  zoneTisch: 'Champ de bataille',
  zoneFriedhof: 'Cimetière',
  zoneExil: 'Exil',
  zoneStapel: 'Pile',
  zoneKommando: 'Zone de commandement',
  zonekurzBibliothek: 'Bibl.',
  zonekurzFriedhof: 'Cim.',
  zonekurzExil: 'Exil',
  zonekurzKommando: 'Cmd.',
  zoneLeer: 'vide',

  unbekannterTyp: 'Type inconnu : {woerter}',
  geheim: 'face cachée',
  getappt: 'engagé',
  befehle: 'Actions',
  regelnTitel: 'Ce qui est appliqué',
  regelwerk: 'Jeu de règles',
  wirksamAb: 'en vigueur',
  hoechsteStrenge: 'Rigueur maximale',
  strenge: 'Rigueur',
  andereSeite: 'Autre face',
  grossAnsehen: 'Voir en grand',
  ersteSeite: 'Première face',
  stelle: 'À la position',
  stelleOben: 'dessus',
  stelleUnten: 'dessous',
  stelleZufaellig: 'mélangée dedans',
  mischpflichtTitel: 'Mélanges à effectuer',

  legendeTitel: 'Choix en attente : la règle des légendes',
  legendeSatz:
    '{wer} doit choisir quel permanent légendaire nommé « {name} » reste (CR 704.5j).',
  legendeBehalten: 'Garder celui-ci',

  reihungTitel: 'Choix en attente : l’ordre des capacités',
  reihungSatz:
    '{wer} contrôle plusieurs capacités qui se sont déclenchées en même temps et indique dans quel ordre elles sont placées sur la pile (CR 603.3b). Ce qui est placé en premier se retrouve en bas et se résout en dernier.',
  reihungLegen: 'Placer ensuite :',
  stand: 'État',
  standLaeuft: 'en cours',
  standPausiert: 'en pause',
  standBeendet: 'terminée',
  pausieren: 'Mettre en pause',
  fortsetzen: 'Reprendre',
  beenden: 'Terminer',
  beendenFrage:
    'Terminer cette partie ? Terminer est définitif — une partie terminée ne peut pas être reprise.',
  beendenJa: 'Terminer',
  beendenNein: 'Annuler',
  ereignisse: 'Ce qui vient de se passer',
  ereignisseLeer: 'Rien pour l’instant.',
  ereignisseLuecke:
    'Des coups manquent peut-être ici : pendant que la page était masquée, il est arrivé plus de messages qu’il n’est possible d’en rattraper.',
  ereignisPrivat: 'Pour toi seulement',
  ereignisSatz: '{sitz} : {handlung} — {objekt}',
  ereignisSatzOhneKarte: '{sitz} : {handlung}',

  // **Die Rückschau** (G5, #1462).
  verlauf: 'L’histoire jusqu’ici',
  verlaufAnsehen: 'Afficher l’historique',
  verlaufNachladen: 'Recharger',
  verlaufLaeuft: 'Récupération en cours …',
  verlaufUngeholt:
    'L’historique n’a pas encore été récupéré. Il montre chaque coup de cette partie, dans la mesure où ce siège peut le voir.',
  verlaufLeer: 'Rien ne s’est encore passé dans cette partie.',
  verlaufGrenze:
    'Ce qu’un autre siège a vu pour lui seul ne figure pas ici — pas même après coup.',

  // **Was von einer Partie übrig bleibt** (G5, #1462).
  protokoll: 'Ce qu’il reste de cette partie',
  protokollErklaerung:
    'Chaque coup de cette partie est enregistré sur cet ordinateur, et chaque siège en voit sa part. Tu peux partager ta part pour plus tard — avec un seul adversaire ou avec tout le monde.',
  protokollAnsehen: 'Afficher les partages',
  protokollNiemand: 'Personne n’a rien partagé.',
  protokollFuer: 'Partagé avec',
  protokollAlle: 'tout le monde',
  protokollSitz: 'Siège',
  protokollEmpfaenger:
    'Destinataires, séparés par des virgules — ou * pour tout le monde',
  protokollFreigeben: 'Partager',
  protokollZuruecknehmen: 'Retirer',
  protokollNurSpaeter:
    'Un partage ne prend effet qu’après la partie. À la table, chacun ne voit toujours que ce qui lui appartient.',
  protokollKeineRueckholung:
    'Retirer ferme la porte — cela ne reprend pas ce que quelqu’un a déjà lu.',
  protokollAufbewahrung:
    'Un enregistrement terminé est proposé à l’examen dans la maintenance des données après 90 jours. Il n’est supprimé que si tu le décides.',
  protokollLoeschen: 'Supprimer l’enregistrement',
  protokollLoeschenErklaerung:
    'Cela supprime de cet ordinateur l’état, l’historique et les partages de cette partie. Cela ne dit rien des copies qui existent ailleurs.',
  protokollGeloescht: 'L’enregistrement a été supprimé.',
  protokollLoeschenFrage: 'Supprimer l’enregistrement de cette partie ?',
  protokollLoeschenJa: 'Supprimer l’enregistrement',

  // **Die Diagnose der Leitung** (G5, #1462).
  diagnose: 'État de la connexion',
  diagnoseErklaerung:
    'Cette page indique sur quoi cette table communique et combien de choses sont en attente. Rien de tout cela ne dit quoi que ce soit sur des cartes — les tiennes ou celles des autres.',
  diagnoseAnsehen: 'Vérifier la connexion',
  diagnoseStill: 'Aucune table n’est en cours.',
  diagnoseAlsWirt: 'Cet ordinateur héberge la table.',
  diagnoseAlsGast: 'Cet ordinateur est assis à la table de quelqu’un d’autre.',
  diagnosePausiert: 'La connexion a été interrompue — la place est conservée.',
  diagnoseWeg: 'Route',
  diagnoseGegenueber: 'Correspondant',
  diagnoseSeit: 'Actif depuis',
  diagnoseWartend: 'En attente',
  diagnoseVerworfen: 'Abandonnés',
  diagnoseStand: 'À la séquence',
  diagnoseRunde: 'Dernier aller-retour',
  diagnoseBildzeit: 'Intervalle d’images dans cette fenêtre',
  diagnoseUnbekannt: 'inconnu',
  diagnoseOhneLeitung: 'Aucune connexion n’est établie.',

  // **Die Freigabe für Gäste** (G5, #1462).
  freigabe: 'Laisser d’autres jouer à cette table',
  freigabeErklaerung:
    'Être joignable et laisser d’autres jouer sont deux décisions distinctes. Ouvrir la table permet à un correspondant de jouer des coups dans cette partie — c’est toi qui attribues les sièges, pas lui.',
  freigabeOffen: 'Cette table est ouverte.',
  freigabeZu: 'Cette table n’est pas ouverte.',
  freigabeStarten: 'Ouvrir la table',
  freigabeBeenden: 'Fermer la table',
  freigabeNachsehen: 'Vérifier de nouveau',
  freigabeSitze: 'Libre pour',
  freigabeNiemand: 'Personne n’est encore assis.',
  freigabePartner: 'Il s’agit de',
  freigabePartnerFremd: 'Ce correspondant ne figure pas dans tes contacts.',
  freigabeBesetzt: 'Sont assis :',
  freigabeOhneHorchposten:
    'Pour qu’un correspondant puisse arriver, cet ordinateur doit aussi être joignable — cela se règle dans les Réglages.',

  // **Der gezeichnete Tisch** — Werkstück G3 von Phase 5.5
  // (#1460).
  tischBild: 'La table',
  tischListe: 'En liste',
  tischGrafisch: 'En table',
  tischDarstellung: 'Affichage',
  tischEinpassen: 'Tout afficher',
  tischNaeher: 'Zoom avant',
  tischWeiter: 'Zoom arrière',
  tischDrehen: 'Faire pivoter la table',
  tischAuswahl: 'Vers la sélection',
  tischEigenerSitz: 'Vers mon siège',
  tischBlick: 'Vue de la table',
  tischZiel: 'Aller à',
  tischOhneWebgl:
    'Cet ordinateur ne fournit pas de graphismes accélérés. La table apparaît ci-dessous sous forme de liste, et tout ce qui fonctionne sur la table dessinée y fonctionne aussi.',
  tischOhneRenderer:
    'La table n’a pas pu être dessinée. La liste ci-dessous contient chaque carte, et tout le reste y fonctionne.',
  tischVerloren:
    'La connexion graphique a été perdue. La table reviendra dès que la fenêtre sera rouverte ; d’ici là, la liste ci-dessous contient chaque carte.',
  tischHeimatlos:
    'Ces objets se trouvent dans une zone que ce programme ne connaît pas et n’apparaissent donc pas sur la table :',
  tischVerdeckt: 'face cachée',
  tischGetappt: 'engagé',
  tischHaengtAn: 'attaché à',
  tischbild: 'Arrière-plan',
  tischbildHolz: 'Bois',
  tischbildDunklesHolz: 'Bois sombre',
  tischbildPicknick: 'Pique-nique',
  tischbildDungeon: 'Donjon',
  tischbildTaverne: 'Taverne',
  tischbildTurniertisch: 'Table de tournoi',
  tischEigenes: 'Image personnelle …',
  tischEigenesWeg: 'Retirer l’image personnelle',
  gespraech: 'Discussion à la table',
  gespraechVerlauf: 'Ce qui a été dit à la table',
  gespraechEingabe: 'Dire quelque chose',
  gespraechSenden: 'Dire',
  gespraechLeer: 'Personne n’a encore rien dit.',
  gespraechNiemandZeigt: 'Personne ne montre quoi que ce soit pour le moment.',
  gespraechZeigt: '{sitz} : {tut}',
  gespraechZeigtAuf: '{sitz} : {tut} {objekt}',
  zeigerZeiger: 'montre',
  zeigerGreift: 'fait glisser',
  zeigerSchwebt: 'regarde',
  zeigerPing: 'signale',
  spielart: 'Mode',
  spielartGoldfish: 'S’entraîner seul',
  spielartSelfplay: 'Plusieurs sièges',
  format: 'Format',
  fehltDeck: 'Choisis un deck pour chaque siège — le premier détermine le jeu.',
  fehltEigenesDeck:
    'Une place a besoin d’un deck à toi — un prêt n’apporte pas de jeu, et le jeu vient du premier deck à toi.',
  fehltName:
    'Chaque siège a besoin d’un nom ; à la table, il distingue les joueurs.',

  // **Die digitale Deckleihe** (G6, #1463).
  leihe: 'Prêter ou emprunter un deck',
  leiheOeffnen: 'Prêter ou emprunter …',
  leiheVerleihenTitel: 'Prêter un deck',
  leiheAnnehmenTitel: 'Accepter un prêt',
  leiheListeTitel: 'Tes prêts',
  leiheErklaerung:
    'Prêter permet à un adversaire de jouer avec un instantané figé de ton deck. Le deck reste ici, et pas une seule carte de ta collection ne change.',
  leiheKeinEigentum:
    'Il s’agit d’une autorisation d’usage, pas d’un transfert. Rien ne change de propriétaire, ni numériquement ni physiquement.',
  leiheKeinKopierschutz:
    'Et ce n’est pas une protection contre la copie : quiconque a vu les cartes les a vues. Ce que le prêt régit, c’est avec quoi ce programme démarre une partie.',
  leiheAnsehen: 'Afficher les prêts',
  leiheKeine: 'Il n’y a aucun prêt.',
  leiheEntleiher: 'À qui prêtes-tu ?',
  leiheEntleiherLeer:
    'Aucun correspondant confirmé pour l’instant. Ajoutes-en un dans l’encadré « Ajouter ou confirmer un correspondant » plus bas.',
  leiheDeck: 'Quel deck',
  leiheDeckWaehlen: 'Choisir un deck …',
  leiheKeinDeck: 'Pas encore de deck — on en crée un sous Decks.',
  leiheFachleute: 'Prêter pour une partie précise uniquement',
  leiheFachleuteErklaerung:
    'Sans indication, le prêt s’applique à la prochaine partie du correspondant. Pour le lier à une seule partie, saisis son identifiant — le correspondant le trouve dans sa liste de parties.',
  leiheWoran: 'Identifiant de la partie — vide signifie la prochaine',
  leiheStunden: 'Valable pendant combien d’heures',
  leiheZiel: 'Où déposer l’enveloppe',
  leiheAnbieten: 'Prêter et enregistrer',
  leiheAbgelegt: 'L’enveloppe se trouve ici :',
  leiheUmschlag: 'Enveloppe d’un prêt',
  leiheAnnehmen: 'Accepter le prêt',
  leiheWiderrufen: 'Révoquer',
  leiheZurueckgeben: 'Rendre',
  leiheVerliehen: 'prêté à',
  leiheGeliehen: 'emprunté à',
  leiheBis: 'valable jusqu’au',
  leiheKarten: 'cartes',
  leiheVerbraucht: 'utilisé pour',
  leiheOffen: 'pas encore utilisé',
  leiheZurueckgenommen: 'révoqué',
  leiheModusZuende: 'Une partie en cours se termine normalement.',
  leiheModusAbbruch: 'Une partie en cours est interrompue.',
  leiheRuecknahmeGrenze:
    'Révoquer empêche de nouveaux démarrages. Cela ne retire rien d’une partie en cours et ne supprime aucune copie chez le correspondant.',

  // **Wie die Partie ausgegangen ist** (M17, #1618).
  ausgangTitel: 'Le résultat',
  ausgangSieger: 'Le vainqueur est',
  ausgangUnentschieden:
    'Égalité — tous les joueurs restants ont perdu en même temps.',
  grundAufgabe: 'a abandonné',
  standLeiste: 'État des joueurs',
  standDu: 'Toi',
  standLeben: 'Vie',
  standGift: 'Poison',
  standMana: 'Mana',
  standStrahlung: 'Radiation',
  standSteuer: 'Taxe',
  standCommanderschaden: '{name} : {schaden}',
  standCommanderTod: 'à 21, la partie est terminée',
  standLeergezogen: 'a pioché dans une bibliothèque vide',
  grundLeben: 'était à zéro point de vie ou moins',
  grundBibliothek: 'devait piocher dans une bibliothèque vide',
  grundGift: 'avait dix marqueurs de poison',
  aufgeben: 'Abandonner',
  aufgebenFrage:
    'Abandonner ? La partie se termine aussitôt pour cette place, et ses cartes partent avec elle.',
  aufgebenJa: 'Oui, abandonner',
  aufgebenNein: 'Continuer',
  aufgebenHinweis:
    'Tu peux abandonner à tout moment, et c’est immédiat. Tu quittes la partie, et tes cartes partent avec toi.',

  // **Der Weg zurück in die eigene Partie** (#1600).

  // **Ein Sitz an einem Profil** (G8, #1465).
  sitzprofil: 'Profil',
  sitzprofilKeines: 'aucun profil',
  sitzprofilErklaerung:
    'Si tu gères deux profils, tu peux en attribuer un à chaque siège. La partie commence alors dans la vue du siège qui correspond à ton profil actuel ; tu accèdes aux autres depuis le sélecteur de vue à la table. Ce que l’autre vue montrait disparaît. Si tu n’attribues de profil à aucun siège, ton profil actuel prend le premier siège.',
  sitzprofilBefunde: 'Une remarque à ce sujet :',
}

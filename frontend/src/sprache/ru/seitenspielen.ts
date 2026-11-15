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

  titel: 'Игра',
  erklaerung:
    'Партия — это стол с зафиксированными колодами. Здесь происходит поток событий: каждое связано с предыдущим, и ни одно не перезаписывается. Перемести карту — и ты сразу увидишь, что видят о ней остальные.',
  anfangTitel: 'Начать партию',
  anfangErklaerung:
    'За каждое место приходит колода, и она зафиксирована: последующее изменение твоей колоды эту партию не меняет. Игра определяется первой выбранной колодой.',
  leerDecksTitel: 'Колод пока нет',
  leerDecks: 'Для игры нужна колода. Создай её в разделе «Колоды».',
  sitz: 'Место',
  sitzName: 'Имя за столом',
  sitzDeck: 'Колода',
  sitzLeihe: 'Взято у …{von} · {karten} карт',
  deckWaehlen: 'Выбери колоду',
  anschlussWartet:
    'Колода «{name}» перенесена и занимает первое место. Второе свободно.',
  sitzHinzu: 'Добавить место',
  sitzWeg: 'Убрать это место',
  sitzGruppe: 'Место {nummer}',
  spiel: 'Игра',
  anfangen: 'Начать партию',
  tischTitel: 'За столом',
  partieGegen: 'Партия против {namen}',
  partieZwischen: 'Партия между {namen}',
  partieVon: 'Одиночная партия: {namen}',
  fassung: 'Ревизия',
  wirt: 'Хост',
  wirtHinweis:
    'Хост хранит полное состояние этой партии и потому видит больше, чем показывает его собственный вид. Это указано здесь потому, что знать об этом нужно до присоединения к игре, а не после.',
  tischAufHaufen: 'В стопке: {zahl}',
  tischMarkeEine: 'Жетон {art}: {zahl}',
  tischMarkeViele: 'Жетоны {art}: {zahl}',
  tischArt: {
    spielstein: 'Фишка',
    emblem: 'Эмблема',
    faehigkeit: 'Способность в стеке',
  },

  sicht: 'Вид для',
  sichtZuschauer: 'Зритель (без места)',
  sichtWechseln: 'Сменить место',
  sichtLaedt: 'Загружается выбранный вид …',
  verlassen: 'Закрыть партию',
  zonen: 'Что где лежит',
  zoneBibliothek: 'Библиотека',
  zoneHand: 'Рука',
  zoneTisch: 'Поле битвы',
  zoneFriedhof: 'Кладбище',
  zoneExil: 'Изгнание',
  zoneStapel: 'Стек',
  zoneKommando: 'Командная зона',
  zonekurzBibliothek: 'Библ.',
  zonekurzFriedhof: 'Клад.',
  zonekurzExil: 'Изгн.',
  zonekurzKommando: 'Ком.',
  zoneLeer: 'пусто',

  unbekannterTyp: 'Неизвестный тип: {woerter}',
  geheim: 'рубашкой вверх',
  getappt: 'повёрнута',
  befehle: 'Действия',
  regelnTitel: 'Что соблюдается принудительно',
  regelwerk: 'Набор правил',
  wirksamAb: 'действует с',
  hoechsteStrenge: 'Наивысшая строгость',
  strenge: 'Строгость',
  andereSeite: 'Другая сторона',
  grossAnsehen: 'Посмотреть крупно',
  ersteSeite: 'Первая сторона',
  stelle: 'На позицию',
  stelleOben: 'наверх',
  stelleUnten: 'вниз',
  stelleZufaellig: 'с перетасовкой',
  mischpflichtTitel: 'Незавершённые обязательные тасовки',

  legendeTitel: 'Ожидается выбор: правило легенд',
  legendeSatz:
    '{wer} должен выбрать легендарный перманент с названием «{name}», который останется (CR 704.5j).',
  legendeBehalten: 'Оставить этот',

  reihungTitel: 'Ожидается выбор: порядок способностей',
  reihungSatz:
    '{wer} контролирует несколько способностей, сработавших одновременно, и определяет, в каком порядке они помещаются в стек (CR 603.3b). То, что помещено первым, оказывается в стеке ниже всех и разрешается последним.',
  reihungLegen: 'Поместить следующей:',
  stand: 'Состояние',
  standLaeuft: 'идёт',
  standPausiert: 'приостановлена',
  standBeendet: 'завершена',
  pausieren: 'Приостановить',
  fortsetzen: 'Продолжить',
  beenden: 'Завершить',
  beendenFrage:
    'Завершить эту партию? Завершение окончательно — завершённую партию нельзя продолжить.',
  beendenJa: 'Завершить',
  beendenNein: 'Отмена',
  ereignisse: 'Что только что произошло',
  ereignisseLeer: 'Пока ничего.',
  ereignisseLuecke:
    'Здесь могут отсутствовать ходы: пока страница была скрыта, пришло больше сообщений, чем можно восстановить.',
  ereignisPrivat: 'Только для тебя',
  ereignisSatz: '{sitz}: {handlung} — {objekt}',
  ereignisSatzOhneKarte: '{sitz}: {handlung}',

  // **Die Rückschau** (G5, #1462).
  verlauf: 'Что было до сих пор',
  verlaufAnsehen: 'Показать историю',
  verlaufNachladen: 'Загрузить заново',
  verlaufLaeuft: 'Загрузка …',
  verlaufUngeholt:
    'История ещё не загружена. Она показывает каждый ход этой партии в той мере, в какой это место вправе его видеть.',
  verlaufLeer: 'В этой партии пока ничего не происходило.',
  verlaufGrenze:
    'То, что другое место видело только для себя, здесь отсутствует — и задним числом тоже.',

  // **Was von einer Partie übrig bleibt** (G5, #1462).
  protokoll: 'Что остаётся от партии',
  protokollErklaerung:
    'Каждый ход этой партии сохраняется на этом компьютере, и каждое место видит свою часть. Свою часть ты можешь открыть на будущее — одному противнику или всем.',
  protokollAnsehen: 'Показать открытия доступа',
  protokollNiemand: 'Никто ничего не открыл.',
  protokollFuer: 'Открыто для',
  protokollAlle: 'всех',
  protokollSitz: 'Место',
  protokollEmpfaenger: 'Получатели через запятую — или * для всех',
  protokollFreigeben: 'Открыть доступ',
  protokollZuruecknehmen: 'Отозвать',
  protokollNurSpaeter:
    'Открытие доступа вступает в силу только после партии. За столом каждый по-прежнему видит только своё.',
  protokollKeineRueckholung:
    'Отзыв закрывает дверь — он не возвращает то, что кто-то уже прочитал.',
  protokollAufbewahrung:
    'Завершённая запись предлагается к проверке в разделе обслуживания данных через 90 дней. Удаляется она только тогда, когда ты этого захочешь.',
  protokollLoeschen: 'Удалить запись',
  protokollLoeschenErklaerung:
    'Это удаляет состояние, историю и открытия доступа этой партии с этого компьютера. О копиях в других местах это ничего не говорит.',
  protokollGeloescht: 'Запись удалена.',
  protokollLoeschenFrage: 'Удалить запись этой партии?',
  protokollLoeschenJa: 'Удалить запись',

  // **Die Diagnose der Leitung** (G5, #1462).
  diagnose: 'Состояние соединения',
  diagnoseErklaerung:
    'Здесь показано, по каким каналам этот стол общается и сколько ожидает отправки. Ничто из этого не говорит о картах — ни твоих, ни чужих.',
  diagnoseAnsehen: 'Проверить соединение',
  diagnoseStill: 'Ни один стол не запущен.',
  diagnoseAlsWirt: 'Этот компьютер является хостом стола.',
  diagnoseAlsGast: 'Этот компьютер сидит за чужим столом.',
  diagnosePausiert: 'Соединение оборвалось — место сохраняется.',
  diagnoseWeg: 'Маршрут',
  diagnoseGegenueber: 'Партнёр',
  diagnoseSeit: 'В сети уже',
  diagnoseWartend: 'Ожидает',
  diagnoseVerworfen: 'Отброшено',
  diagnoseStand: 'Номер последовательности',
  diagnoseRunde: 'Последняя круговая задержка',
  diagnoseBildzeit: 'Интервал кадров в этом окне',
  diagnoseUnbekannt: 'неизвестно',
  diagnoseOhneLeitung: 'Ни одно соединение не установлено.',

  // **Die Freigabe für Gäste** (G5, #1462).
  freigabe: 'Разрешить другим играть за этим столом',
  freigabeErklaerung:
    'Быть доступным и позволять другим играть — это два разных решения. Открыв стол, ты позволяешь партнёру делать ходы в этой партии — места раздаёшь ты, а не он.',
  freigabeOffen: 'Этот стол открыт.',
  freigabeZu: 'Этот стол не открыт.',
  freigabeStarten: 'Открыть стол',
  freigabeBeenden: 'Закрыть стол',
  freigabeNachsehen: 'Проверить ещё раз',
  freigabeSitze: 'Свободно для',
  freigabeNiemand: 'Пока никто не сел.',
  freigabePartner: 'Это',
  freigabePartnerFremd: 'Этого партнёра нет среди твоих контактов.',
  freigabeBesetzt: 'Занято:',
  freigabeOhneHorchposten:
    'Чтобы партнёр мог подключиться, этот компьютер тоже должен быть доступен — это настраивается в разделе «Настройки».',

  // **Der gezeichnete Tisch** — Werkstück G3 von Phase 5.5
  // (#1460).
  tischBild: 'Стол',
  tischListe: 'Списком',
  tischGrafisch: 'Столом',
  tischDarstellung: 'Отображение',
  tischEinpassen: 'Показать всё',
  tischNaeher: 'Приблизить',
  tischWeiter: 'Отдалить',
  tischDrehen: 'Повернуть стол',
  tischAuswahl: 'К выбранному',
  tischEigenerSitz: 'К моему месту',
  tischBlick: 'Вид стола',
  tischZiel: 'Перейти к',
  tischOhneWebgl:
    'На этом компьютере нет ускоренной графики. Стол показан ниже списком, и всё, что работает на нарисованном столе, работает и там.',
  tischOhneRenderer:
    'Стол не удалось нарисовать. Список ниже содержит каждую карту, и всё остальное работает там.',
  tischVerloren:
    'Графическое соединение потеряно. Стол вернётся после повторного открытия окна; до тех пор каждую карту содержит список ниже.',
  tischHeimatlos:
    'Эти объекты лежат в зоне, которую эта программа не знает, и поэтому не появляются на столе:',
  tischVerdeckt: 'рубашкой вверх',
  tischGetappt: 'повёрнута',
  tischHaengtAn: 'присоединена к',
  tischbild: 'Фон',
  tischbildHolz: 'Дерево',
  tischbildDunklesHolz: 'Тёмное дерево',
  tischbildPicknick: 'Пикник',
  tischbildDungeon: 'Подземелье',
  tischbildTaverne: 'Таверна',
  tischbildTurniertisch: 'Турнирный стол',
  tischEigenes: 'Своё изображение …',
  tischEigenesWeg: 'Удалить своё изображение',
  gespraech: 'Разговор за столом',
  gespraechVerlauf: 'Что было сказано за столом',
  gespraechEingabe: 'Сказать что-нибудь',
  gespraechSenden: 'Сказать',
  gespraechLeer: 'Пока никто ничего не сказал.',
  gespraechNiemandZeigt: 'Сейчас никто ни на что не указывает.',
  gespraechZeigt: '{sitz}: {tut}',
  gespraechZeigtAuf: '{sitz}: {tut} — {objekt}',
  zeigerZeiger: 'указывает',
  zeigerGreift: 'тащит',
  zeigerSchwebt: 'смотрит',
  zeigerPing: 'привлекает внимание',
  spielart: 'Режим',
  spielartGoldfish: 'Тренировка в одиночку',
  spielartSelfplay: 'Несколько мест',
  format: 'Формат',
  fehltDeck: 'Выбери колоду для каждого места — первая определяет игру.',
  fehltEigenesDeck:
    'Одному месту нужна своя колода: заём не несёт игры, и игра берётся из первой своей колоды.',
  fehltName:
    'У каждого места должно быть имя; за столом оно позволяет различать игроков.',

  // **Die digitale Deckleihe** (G6, #1463).
  leihe: 'Одолжить или взять колоду',
  leiheOeffnen: 'Одолжить или взять …',
  leiheVerleihenTitel: 'Одолжить колоду',
  leiheAnnehmenTitel: 'Принятие займа',
  leiheListeTitel: 'Твои займы',
  leiheErklaerung:
    'Одалживание позволяет противнику играть с зафиксированным снимком твоей колоды. Колода остаётся здесь, и ни одна карта твоей коллекции не меняется.',
  leiheKeinEigentum:
    'Это разрешение на использование, а не передача. Ничто не меняет владельца — ни цифровое, ни физическое.',
  leiheKeinKopierschutz:
    'И это не защита от копирования: кто видел карты, тот их видел. Заём определяет лишь то, с чем эта программа начнёт партию.',
  leiheAnsehen: 'Показать займы',
  leiheKeine: 'Займов нет.',
  leiheEntleiher: 'Кому ты одалживаешь?',
  leiheEntleiherLeer:
    'Подтверждённых партнёров пока нет. Добавь одного в блоке «Добавить или подтвердить партнёра» ниже.',
  leiheDeck: 'Какая колода',
  leiheDeckWaehlen: 'Выбери колоду …',
  leiheKeinDeck: 'Колод пока нет — их создают в разделе «Колоды».',
  leiheFachleute: 'Одолжить только для одной определённой партии',
  leiheFachleuteErklaerung:
    'Без этого заём действует для следующей партии партнёра. Чтобы привязать его ровно к одной партии, введи её идентификатор — партнёр найдёт его в своём списке партий.',
  leiheWoran: 'Идентификатор партии — пусто означает следующую',
  leiheStunden: 'Действует сколько часов',
  leiheZiel: 'Куда положить конверт',
  leiheAnbieten: 'Одолжить и сохранить',
  leiheAbgelegt: 'Конверт находится здесь:',
  leiheUmschlag: 'Конверт займа',
  leiheAnnehmen: 'Принять заём',
  leiheWiderrufen: 'Отозвать',
  leiheZurueckgeben: 'Вернуть',
  leiheVerliehen: 'одолжено',
  leiheGeliehen: 'взято у',
  leiheBis: 'действует до',
  leiheKarten: 'карт',
  leiheVerbraucht: 'использовано для',
  leiheOffen: 'ещё не использовано',
  leiheZurueckgenommen: 'отозвано',
  leiheModusZuende: 'Идущая партия доигрывается.',
  leiheModusAbbruch: 'Идущая партия останавливается.',
  leiheRuecknahmeGrenze:
    'Отзыв не даёт начать новые партии. Он ничего не убирает из идущей партии и не удаляет копию на другой стороне.',

  // **Wie die Partie ausgegangen ist** (M17, #1618).
  ausgangTitel: 'Итог',
  ausgangSieger: 'Победитель:',
  ausgangUnentschieden:
    'Ничья — все, кто ещё оставался в партии, проиграли одновременно.',
  grundAufgabe: 'сдался',
  standLeiste: 'Состояние игроков',
  standDu: 'Ты',
  standLeben: 'Жизнь',
  standGift: 'Яд',
  standMana: 'Мана',
  standStrahlung: 'Радиация',
  standSteuer: 'Налог',
  standCommanderschaden: '{name}: {schaden}',
  standCommanderTod: 'при 21 партия окончена',
  standLeergezogen: 'взял карту из пустой библиотеки',
  grundLeben: 'имел ноль жизней или меньше',
  grundBibliothek: 'должен был взять карту из пустой библиотеки',
  grundGift: 'имел десять ядовитых жетонов',
  aufgeben: 'Сдаться',
  aufgebenFrage:
    'Сдаться? Игра для этого места сразу закончится, и его карты уйдут вместе с ним.',
  aufgebenJa: 'Да, сдаться',
  aufgebenNein: 'Играть дальше',
  aufgebenHinweis:
    'Ты можешь сдаться в любой момент, и это вступает в силу сразу. Ты выходишь из партии, а твои карты уходят вместе с тобой.',

  // **Der Weg zurück in die eigene Partie** (#1600).

  // **Ein Sitz an einem Profil** (G8, #1465).
  sitzprofil: 'Профиль',
  sitzprofilKeines: 'без профиля',
  sitzprofilErklaerung:
    'Если у тебя два профиля, каждому месту можно назначить один. Тогда партия начинается с вида того места, которое относится к твоему текущему профилю; к остальным можно перейти через выбор вида за столом. То, что показывал другой вид, пропадает. Если не назначить профиль ни одному месту, твой текущий профиль занимает первое место.',
  sitzprofilBefunde: 'Об этом есть примечание:',
}

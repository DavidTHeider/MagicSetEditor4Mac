// **Die Texte der KI** — Werkstück A6 von Phase 5.3
// (#1445),
// erweitert für A7
// (#1446).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als einer
// ankommt** (§9): `seiteneinstellungen.ts` stand mit ihnen an der
// Zeilengrenze aus §2.
export const ki = {
  kimodelleTitel: 'Модели ИИ',
  kimodelleKeine:
    'Ни одна модель ИИ не подтверждена. Список моделей провайдера ещё не является подтверждённой возможностью — здесь появляются только модели, которые действительно проверены.',
  kiNurLokal:
    'Действует режим «только локально»: ни одна модель вне этого компьютера не запрашивается.',
  kiAuchCloud:
    'Модели у сторонних провайдеров разрешены. Всё, что им отправляется, покидает этот компьютер.',
  kiBeleg: '{faehigkeit} — проверено {datum}: {beleg}',
  kiKontextfenster: 'Окно контекста: {tokens} токенов — {beleg}',
  kiKontextfensterUnbekannt: 'Окно контекста: не указано',
  kiHerkunftUnbekannt:
    'Происхождение неизвестно — не установлено, какой именно файл здесь запускается.',
  kiEinrichtungAus:
    'Не включено. Ничего не запрашивается, и ничего не покидает этот компьютер.',
  kiEinrichtungUnerreichbar:
    'По настроенному адресу никто не отвечает. Это приложение ничего там не запускает и ничего не устанавливает.',
  kiEinrichtungOhneModell:
    'Ответ есть, но моделей нет. Это приложение их не загружает.',
  kiEinrichtungOhneFaehigkeit:
    'Модели есть, но ни у одной нет подтверждённой возможности, полезной здесь.',
  kiEinrichtungBereit:
    'Готова как минимум одна модель с подтверждённой возможностью.',
  kiLaufzeitOllama:
    'Если здесь никто не отвечает, значит, Ollama не запущена. Это приложение не запускает и не устанавливает её — это делается в самой Ollama.',
  kiLaufzeitTitel: 'Локальный ИИ',
  kiLaufzeitFassung: 'Среда выполнения: {fassung}',
  kiAdresse: 'Запрос отправляется на {adresse}.',
  kiAdresseVorgabe: 'адрес по умолчанию на этом компьютере',
  kiCloudAus: 'Передача вычислений в облако Ollama отключена.',
  kiCloudAn: 'Передача вычислений в облако Ollama не отключена.',
  kiCloudUngewiss:
    'Отсюда нельзя установить, отключена ли передача вычислений в облако Ollama.',
  kiCloudBeleg: 'Подтверждение: {beleg}',
  kiVorschlagTitel: 'Предложение от модели',
  kiVorschlagModell: 'Модель',
  kiVorschlagAuftrag: 'Задача',
  kiVorschlagAuftragHinweis: 'Что должна сделать модель?',
  kiVorschlagen: 'Предложить',
  kiVorschlagLaeuft:
    'Модель работает — в первый раз это может занять время, потому что она загружается.',
  kiVorschlagAbbrechen: 'Отмена',
  kiVorschlagUebernehmen: 'Применить',
  kiVorschlagNichts: 'Модель не предложила ни одного поля, известного шаблону.',
  kiVorschlagAendertNichts:
    'Предложение не изменяет карту. Записывает только «Применить».',
  kiVorschlagUebernommen: 'Записано полей: {felder}.',
  kiVorschlagAblauf: 'Можно применить до {ablauf}.',
  // **The second provider** — A8, #1447.
  kiAnbieterZeile: 'Провайдер {anbieter}: {stand}',
  kiVorschlagAnbieter: 'Провайдер и модель',
  kiVorschlagModellWaehlen: 'Выбери',
  kiVorschlagOhneModell:
    'Нет доступной модели. Чего не хватает, указано в разделе «Модели ИИ».',
  kiVorschlagAntwortVon: 'Ответ получен от {anbieter}/{modell}.',
  // **Measuring what a model can do** — A9, #1448.
  kiMessbarTitel: 'Модели без описания',
  kiAdresseSocket:
    'Запросы идут через собственный Unix-сокет этой программы — без порта.',
  kiGeraetSitzung: 'Использовать в этом сеансе',
  kiGeraetSitzungHinweis:
    'Действует до перезапуска; затем снова действует настройка «Модель Apple на устройстве».',
  kiFeldVorschlagen: 'Предложить',
  kiFeldVorschlagenFuer: '{feld} — предложение',
  kiFeldAuftragName:
    'Предложи подходящее имя собственное для этой карты — не тип карты и не описание её правил. Отвечай по-русски, только названием.',
  kiFeldAuftragFlavor:
    'Предложи короткий художественный текст для этой карты, одно-два предложения: настроение или история без повторения текста правил. Отвечай по-русски.',
  kiFeldKarte: 'Карта на данный момент:',
  kiFeldUebernehmen: 'Применить',
  kiFeldVon: 'от {anbieter} · {weg}',
  kiFeldWegGewaehlt: 'выбран',
  kiFeldWegVorrang: 'приоритет',
  kiFeldVorrangFragt: 'Приоритет спрашивает: {reihe}',
  kiFeldWegRueckfall: 'запасной',
  kiFeldAbgelehnt: 'Модель отказала: {grund}',
  kiFeldNichts: 'Для этого поля предложения нет.',
  kiMessbarSatz:
    '{anbieter} перечисляет эти модели, но ничего о них не сообщает. Измерение отправляет небольшую задачу с фиксированной схемой — оно стоит столько, сколько стоит эта одна задача.',
  kiMessen: 'Измерить',
  kiMessungVergessen: 'Отбросить результат измерения',
  kiMessungLaeuft: 'Измерение …',
  kiMessungGeglueckt: 'Ответ от {alias}. Теперь подтверждено: {faehigkeiten}.',
  // **Selection, comparison and provenance** — A11, #1450.
  kiVorschlagAuswahl: 'Применить: {feld}',
  kiVorschlagBisher: 'Сейчас: {wert}',
  kiVorschlagBisherLeer: 'Сейчас пусто.',
  kiVorschlagNichtsGewaehlt:
    'Ни одно поле не отмечено. Применяется то, что отмечено.',
  kiHerkunftTitel: 'Откуда взяты значения',
  kiHerkunftZeile: '{feld}: {anbieter}/{modell}, {zeitpunkt}',
  kiHerkunftAufgabe: 'Задача была такой: {aufgabe}',
  kiHerkunftVeraendert: 'С тех пор изменено вручную.',
  kiHerkunftKeine: 'Ни одно поле этой карты не взято из предложения.',
  // **The image proposal** — A12, #1451.
  kiBildTitel: 'Изображение от модели',
  kiBildAendertNichts:
    'Создание не изменяет карту. Результат — предложение со сроком действия: ты просматриваешь его, кадрируешь и только потом вставляешь.',
  kiBildFeld: 'Поле изображения',
  kiBildAufgabe: 'Что оно должно показывать?',
  kiBildAufgabeHinweis: 'Молния над широкой равниной, масляная живопись',
  kiBildGroesse: 'Размер',
  kiBildGuete: 'Качество',
  kiBildBearbeiten: 'Изменить существующее изображение, а не создавать новое',
  kiBildErzeugen: 'Создать изображение',
  kiBildLaeuft:
    'Изображение создаётся у {anbieter}/{modell} — это занимает от нескольких секунд до нескольких минут.',
  kiBildAbbrechen: 'Отмена',
  kiBildAbgebrochen:
    'Отменено. Ничего не перенесено; всё, что это стоило до этого момента, записано в журнал расходов.',
  kiBildVorschau: 'Предложение для {feld}',
  kiBildMasse: '{breite} × {hoehe} пикселей, {kilobytes} КБ',
  kiBildAblauf: 'Действительно до {ablauf}.',
  kiBildUmschrift: 'Что было нарисовано: {umschrift}',
  kiBildNachweisTitel: 'Откуда взято это изображение',
  kiBildAngabeGroesse: 'Размер',
  kiBildAngabeGuete: 'Качество',
  kiBildAngabeUmschrift: 'Уточнённый запрос',
  kiBildAngabeMedientyp: 'Формат файла',
  kiBildUebernehmen: 'Кадрировать и вставить',
  kiBildVerwerfen: 'Отбросить',
  kiBildVerworfen: 'Предложения больше нет.',
  kiBildKosten:
    'Заказ изображения разрешается отдельно — разрешение для текста на него не распространяется.',
  // **Approving a paid request** —
  // #1539.
  kiFreigabeTitel: 'Этот запрос стоит денег',
  kiFreigabeSatz:
    'Выбранная модель работает у {anbieter}. Разреши максимальную сумму — она покрывает один этот запрос и истекает через пять минут.',
  kiFreigabeBetrag: 'Максимальная сумма в наименьшей единице (центы для EUR)',
  kiFreigabeVerfuegbar: 'Доступно сегодня: {betrag}',
  kiFreigabeOhneBudget:
    'Для {anbieter} ничего не разрешено. Без разрешённого бюджета ни один платный запрос не уходит — лимит задаётся в разделе «Расходы».',
  kiFreigabeErteilen: 'Разрешить и запросить',
  kiFreigabeVerzichten: 'Не разрешать',
  kiFreigabeLaeuft: 'Выдача разрешения …',

  // **The evidence-backed assessment** — A13.2a
  // (#1452).
  kiEinschaetzungTitel: 'Запросить оценку',
  kiEinschaetzungErklaerung:
    'Модель оценивает эту карту по отрывкам из локального поискового индекса. Она ничего не меняет — ни карту, ни коллекцию.',
  kiEinschaetzungArt: 'О чём',
  kiEinschaetzungArtRegel: 'По правилам',
  kiEinschaetzungArtBalance: 'В сравнении',
  kiEinschaetzungFrage: 'Что оценить',
  kiEinschaetzungFragen: 'Получить оценку',
  kiEinschaetzungLaeuft: 'Модель читает подтверждения …',
  kiEinschaetzungLeer: 'Пока ничего не оценено.',
  kiEinschaetzungBelegt: 'Всё упомянутое есть в подтверждениях.',
  kiEinschaetzungUnbelegt:
    'В этом ответе упомянуто то, чего нет ни в одном отрывке. Что именно, показано ниже.',
  kiEinschaetzungBefunde: 'Что не подкреплено подтверждениями',
  kiEinschaetzungEinschraenkungen: 'Что не учитывалось',
  kiEinschaetzungBelege: 'Предоставленные отрывки',
  kiEinschaetzungStaende: 'Состояние источников',

  // **The balance analysis** — A13.2b
  // (#1452).
  kiEinschaetzungFormat: 'Формат',
  kiEinschaetzungOhneFormat: 'не указан',
  kiEinschaetzungVergleiche: 'Сравнение с',
  kiEinschaetzungStrategie: 'Найдено через',
  kiEinschaetzungOhneVergleich:
    'В этом ответе не названа ни одна карта для сравнения. Утверждение о положении карты — это сравнение.',
  kiEinschaetzungSynergien: 'Какие синергии не учтены',

  // **The batch and the atomic group** — A13.1b
  // (#1452).
  kiBatchTitel: 'Сразу несколько карт',
  kiBatchSatz:
    'Одна модель отвечает на одну и ту же задачу для каждой выбранной карты. Пока ничего не записывается — это второй шаг.',
  kiBatchZiele: 'Выбрано карт: {gewaehlt} из {gesamt}',
  kiBatchOhneKarten: 'В этом наборе пока нет карт.',
  kiBatchOhneName: '(без названия)',
  kiBatchAuftrag: 'Задача для каждой карты',
  kiBatchFahren: 'Получить предложения',
  kiBatchLaeuft: 'Выполняется …',
  kiBatchLagen: 'Что из этого вышло',
  kiBatchGanzOderGar:
    'Принимается либо всё сразу, либо ничего. Если одна цель не удалась, все карты остаются как были — а предложения остаются открытыми.',
  kiBatchGruppe: 'Принять группой ({karten})',
  kiBatchGruppeLaeuft: 'Применение …',
  kiBatchGeschrieben: 'Записано карт: {karten}.',

  // **Card recognition from a photo** — A13.3
  // (#1452).
  kiErkennungTitel: 'Распознать карту по фото',
  kiErkennungSatz:
    'Модель читает то, что напечатано на фото. Какая это карта, решает тот же поиск изданий, что и для набранного списка, — и в коллекцию ничего не добавляется.',
  kiErkennungLesen: 'Выбрать фото и прочитать',
  kiErkennungLaeuft: 'Модель читает фото …',
  kiErkennungLeer: 'Ни одно фото ещё не прочитано.',
  kiErkennungGelesen: 'Что прочитано',
  kiErkennungNichtGelesen: 'Нечего сказать о',
  kiErkennungUnlesbar: 'не читается',
  kiErkennungOhneWert: 'ничего не прочитано',
  kiErkennungArten: {
    name: 'Название карты',
    setcode: 'Набор',
    nummer: 'Коллекционный номер',
    typzeile: 'Строка типа',
    layout: 'Макет',
    kunstwerk: 'Иллюстрация',
  } as Record<string, string>,
  kiErkennungKandidaten: 'Какие издания подходят',
  kiErkennungDrucke: 'Подходящих изданий: {drucke}.',
  kiErkennungEinDruck: 'Подходит ровно одно издание.',
  kiErkennungWeitere: 'И дальше есть ещё — подсчитана только первая страница.',
  kiErkennungOhneKandidat:
    'Ни одно издание не подходит к прочитанному. Причина может быть в фото, или этой карты здесь нет в списке.',
  kiErkennungGestuetzt: 'Подтверждено: {arten}',
  kiErkennungOhneStuetze:
    'Ничто в этом издании нельзя найти на фото снова — оно показано только потому, что поиск нашёл его по запросу.',
  kiErkennungNichtsAngelegt:
    'Ничего не добавляется. Чтобы учесть экземпляр, воспользуйся формой выше — тогда будет определено, какое издание имеется в виду.',
  kiErkennungStand: 'Возраст поиска изданий: {alter}',
  kiErkennungModell: 'Прочитано: {modell} у {anbieter}',

  // **The citation question** — A13.4
  // (#1452).
  kiVorschlagBelegfrage: 'Добавить отрывки из базы знаний (необязательно)',
  kiVorschlagBelegfrageHinweis: 'что искать — пусто означает: не искать',
} as const

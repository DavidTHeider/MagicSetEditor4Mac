// **Die Texte der KI** — Werkstück A6 von Phase 5.3
// (#1445),
// erweitert für A7
// (#1446).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als einer
// ankommt** (§9): `seiteneinstellungen.ts` stand mit ihnen an der
// Zeilengrenze aus §2.
export const ki = {
  kimodelleTitel: 'AI 模型',
  kimodelleKeine:
    '尚未确认任何 AI 模型。提供方的模型列表并不等于已证实的能力——这里只会显示确实经过检查的模型。',
  kiNurLokal: '“仅限本机”已生效：不会向本机以外的任何模型发出请求。',
  kiAuchCloud: '允许使用第三方提供方的模型。发送给它们的内容都会离开这台电脑。',
  kiBeleg: '{faehigkeit}——于 {datum} 检查：{beleg}',
  kiKontextfenster: '上下文窗口：{tokens} 个令牌 — {beleg}',
  kiKontextfensterUnbekannt: '上下文窗口：未说明',
  kiHerkunftUnbekannt: '来源未知——无法确定这里实际运行的是哪个文件。',
  kiEinrichtungAus: '未开启。不会发出任何请求，也不会有任何内容离开这台电脑。',
  kiEinrichtungUnerreichbar:
    '所配置的地址没有响应。本应用不会在那里启动或安装任何东西。',
  kiEinrichtungOhneModell: '有响应，但没有模型。本应用不会下载模型。',
  kiEinrichtungOhneFaehigkeit: '有模型，但没有一个具备在此处可用的已证实能力。',
  kiEinrichtungBereit: '至少有一个模型已就绪，并具备已证实的能力。',
  kiLaufzeitOllama:
    '如果这里没有响应，说明 Ollama 未在运行。本应用既不会启动也不会安装它——这需要在 Ollama 中完成。',
  kiLaufzeitTitel: '本地 AI',
  kiLaufzeitFassung: '运行环境：{fassung}',
  kiAdresse: '正在请求 {adresse}。',
  kiAdresseVorgabe: '这台电脑上的默认地址',
  kiCloudAus: '向 Ollama 云端分流的功能已关闭。',
  kiCloudAn: '向 Ollama 云端分流的功能未关闭。',
  kiCloudUngewiss: '无法在此确定向 Ollama 云端分流的功能是否已关闭。',
  kiCloudBeleg: '依据：{beleg}',
  kiVorschlagTitel: '模型的建议',
  kiVorschlagModell: '模型',
  kiVorschlagAuftrag: '任务',
  kiVorschlagAuftragHinweis: '模型应该做什么？',
  kiVorschlagen: '给出建议',
  kiVorschlagLaeuft: '模型正在工作——首次可能需要一会儿，因为要先载入模型。',
  kiVorschlagAbbrechen: '取消',
  kiVorschlagUebernehmen: '应用',
  kiVorschlagNichts: '模型没有建议任何模板所认识的字段。',
  kiVorschlagAendertNichts: '建议不会更改卡牌。只有点按“应用”才会写入。',
  kiVorschlagUebernommen: '已写入 {felder} 个字段。',
  kiVorschlagAblauf: '可应用至 {ablauf}。',
  // **The second provider** — A8, #1447.
  kiAnbieterZeile: '提供方 {anbieter}：{stand}',
  kiVorschlagAnbieter: '提供方与模型',
  kiVorschlagModellWaehlen: '请选择',
  kiVorschlagOhneModell: '没有可用的模型。缺少的内容列在“AI 模型”下。',
  kiVorschlagAntwortVon: '回答来自 {anbieter}/{modell}。',
  // **Measuring what a model can do** — A9, #1448.
  kiMessbarTitel: '没有说明的模型',
  kiAdresseSocket: '通过本程序自己的 Unix 套接字访问——不开放端口。',
  kiGeraetSitzung: '在本次会话中使用',
  kiGeraetSitzungHinweis:
    '在下次重启前有效；之后重新按“Apple 设备端模型”设置执行。',
  kiFeldVorschlagen: '建议',
  kiFeldVorschlagenFuer: '{feld} — 建议',
  kiFeldAuftragName:
    '请为这张卡牌建议一个合适的专有名称——不要用卡牌类型，也不要描述规则。用中文回答，只给出名称。',
  kiFeldAuftragFlavor:
    '请为这张卡牌建议一段简短的风味文字，一到两句：写氛围或故事，不要重复规则文字。用中文回答。',
  kiFeldKarte: '目前的卡牌：',
  kiFeldUebernehmen: '采用',
  kiFeldVon: '来自 {anbieter} · {weg}',
  kiFeldWegGewaehlt: '已选择',
  kiFeldWegVorrang: '优先',
  kiFeldVorrangFragt: '优先询问：{reihe}',
  kiFeldWegRueckfall: '后备',
  kiFeldAbgelehnt: '模型拒绝了：{grund}',
  kiFeldNichts: '此字段没有收到建议。',
  kiMessbarSatz:
    '{anbieter} 列出了这些模型，但没有给出任何说明。测量会发送一个带固定结构的小任务——费用即这一个任务的费用。',
  kiMessen: '测量',
  kiMessungVergessen: '丢弃测量结果',
  kiMessungLaeuft: '正在测量…',
  kiMessungGeglueckt: '{alias} 已回答：{faehigkeiten} 现已得到证实。',
  // **Selection, comparison and provenance** — A11, #1450.
  kiVorschlagAuswahl: '应用：{feld}',
  kiVorschlagBisher: '当前：{wert}',
  kiVorschlagBisherLeer: '当前为空。',
  kiVorschlagNichtsGewaehlt: '没有勾选任何字段。只有勾选的内容才会被应用。',
  kiHerkunftTitel: '数值的来源',
  kiHerkunftZeile: '{feld}：{anbieter}/{modell}，{zeitpunkt}',
  kiHerkunftAufgabe: '任务是：{aufgabe}',
  kiHerkunftVeraendert: '此后已被手动修改。',
  kiHerkunftKeine: '这张卡牌没有任何字段来自建议。',
  // **The image proposal** — A12, #1451.
  kiBildTitel: '来自模型的图片',
  kiBildAendertNichts:
    '生成不会更改卡牌。出现的只是一个有时效的建议——你先查看、裁切，然后才插入。',
  kiBildFeld: '图片字段',
  kiBildAufgabe: '应该显示什么？',
  kiBildAufgabeHinweis: '一望无际的平原上空的一道闪电，油画',
  kiBildGroesse: '尺寸',
  kiBildGuete: '质量',
  kiBildBearbeiten: '编辑现有图片，而不是生成新图片',
  kiBildErzeugen: '生成图片',
  kiBildLaeuft: '正在通过 {anbieter}/{modell} 制作图片——需要几秒到几分钟。',
  kiBildAbbrechen: '取消',
  kiBildAbgebrochen:
    '已取消。没有采纳任何内容；到此为止产生的费用记录在费用日志中。',
  kiBildVorschau: '{feld} 的建议',
  kiBildMasse: '{breite} × {hoehe} 像素，{kilobytes} kB',
  kiBildAblauf: '有效期至 {ablauf}。',
  kiBildUmschrift: '所绘内容：{umschrift}',
  kiBildNachweisTitel: '这张图片的来源',
  kiBildAngabeGroesse: '尺寸',
  kiBildAngabeGuete: '质量',
  kiBildAngabeUmschrift: '改写后的提示词',
  kiBildAngabeMedientyp: '文件格式',
  kiBildUebernehmen: '裁切并插入',
  kiBildVerwerfen: '丢弃',
  kiBildVerworfen: '该建议已不存在。',
  kiBildKosten: '图片订单需单独批准——文本的批准并不涵盖它。',
  // **Approving a paid request** —
  // #1539.
  kiFreigabeTitel: '此请求需要付费',
  kiFreigabeSatz:
    '所选模型在 {anbieter} 上运行。请批准一个最高金额——它仅涵盖这一次请求，五分钟后失效。',
  kiFreigabeBetrag: '最高金额（以最小单位计，欧元为分）',
  kiFreigabeVerfuegbar: '今日可用：{betrag}',
  kiFreigabeOhneBudget:
    '尚未为 {anbieter} 批准任何预算。没有已批准的预算，就不会发出任何付费请求——限额在“费用”下设置。',
  kiFreigabeErteilen: '批准并请求',
  kiFreigabeVerzichten: '不批准',
  kiFreigabeLaeuft: '正在授予批准…',

  // **The evidence-backed assessment** — A13.2a
  // (#1452).
  kiEinschaetzungTitel: '请求评估',
  kiEinschaetzungErklaerung:
    '模型会依据本地搜索索引中的段落来评估这张卡牌。它不会更改任何内容——不论是卡牌还是收藏。',
  kiEinschaetzungArt: '评估什么',
  kiEinschaetzungArtRegel: '按规则',
  kiEinschaetzungArtBalance: '作比较',
  kiEinschaetzungFrage: '要评估什么',
  kiEinschaetzungFragen: '获取评估',
  kiEinschaetzungLaeuft: '模型正在阅读依据…',
  kiEinschaetzungLeer: '尚未评估。',
  kiEinschaetzungBelegt: '提到的所有内容都能在依据中找到。',
  kiEinschaetzungUnbelegt:
    '这个回答提到了没有任何段落包含的内容。具体是什么，见下方。',
  kiEinschaetzungBefunde: '没有依据支持的内容',
  kiEinschaetzungEinschraenkungen: '未考虑的内容',
  kiEinschaetzungBelege: '所提供的段落',
  kiEinschaetzungStaende: '来源状态',

  // **The balance analysis** — A13.2b
  // (#1452).
  kiEinschaetzungFormat: '赛制',
  kiEinschaetzungOhneFormat: '未指定',
  kiEinschaetzungVergleiche: '比较对象',
  kiEinschaetzungStrategie: '查找方式',
  kiEinschaetzungOhneVergleich:
    '这个回答没有列出任何作比较的卡牌。对一张卡牌所处位置的论断本身就是一种比较。',
  kiEinschaetzungSynergien: '被略去的协同效应',

  // **The batch and the atomic group** — A13.1b
  // (#1452).
  kiBatchTitel: '一次处理多张卡牌',
  kiBatchSatz:
    '同一个模型对每张所选卡牌回答同一个任务。此时尚未写入任何内容——那是第二步。',
  kiBatchZiele: '已选 {gewaehlt} / {gesamt} 张卡牌',
  kiBatchOhneKarten: '此系列还没有卡牌。',
  kiBatchOhneName: '（未命名）',
  kiBatchAuftrag: '针对每张卡牌的任务',
  kiBatchFahren: '获取建议',
  kiBatchLaeuft: '正在运行…',
  kiBatchLagen: '结果如何',
  kiBatchGanzOderGar:
    '要么全部采纳，要么一个都不采纳。只要有一个目标失败，所有卡牌都保持原样——建议也仍然保持打开。',
  kiBatchGruppe: '作为一组采纳（{karten}）',
  kiBatchGruppeLaeuft: '正在采纳…',
  kiBatchGeschrieben: '已写入 {karten} 张卡牌。',

  // **Card recognition from a photo** — A13.3
  // (#1452).
  kiErkennungTitel: '通过照片识别卡牌',
  kiErkennungSatz:
    '模型会读取照片上印刷的内容。具体是哪张卡牌，由与输入列表所用的同一套印刷版本搜索来决定——不会向收藏中添加任何内容。',
  kiErkennungLesen: '选取照片并读取',
  kiErkennungLaeuft: '模型正在读取照片…',
  kiErkennungLeer: '尚未读取照片。',
  kiErkennungGelesen: '读取到的内容',
  kiErkennungNichtGelesen: '无法给出结论的项目',
  kiErkennungUnlesbar: '无法辨认',
  kiErkennungOhneWert: '未读取到内容',
  kiErkennungArten: {
    name: '卡牌名称',
    setcode: '系列',
    nummer: '收藏编号',
    typzeile: '类别行',
    layout: '版式',
    kunstwerk: '插图',
  } as Record<string, string>,
  kiErkennungKandidaten: '哪些印刷版本相符',
  kiErkennungDrucke: '有 {drucke} 个印刷版本相符。',
  kiErkennungEinDruck: '恰好有一个印刷版本相符。',
  kiErkennungWeitere: '后面还有更多——只统计了第一页。',
  kiErkennungOhneKandidat:
    '没有印刷版本与读取到的内容相符。可能是照片的问题，也可能是这里没有收录这张卡牌。',
  kiErkennungGestuetzt: '依据：{arten}',
  kiErkennungOhneStuetze:
    '这个印刷版本上没有任何内容能在照片中再次找到——它被列出，只是因为搜索按该词找到了它。',
  kiErkennungNichtsAngelegt:
    '不会添加任何内容。要记录一件藏品，请使用上方的表单——这样才能确定指的是哪个印刷版本。',
  kiErkennungStand: '印刷版本搜索的时效：{alter}',
  kiErkennungModell: '由 {anbieter} 的 {modell} 读取',

  // **The citation question** — A13.4
  // (#1452).
  kiVorschlagBelegfrage: '包含知识库中的段落（可选）',
  kiVorschlagBelegfrageHinweis: '要搜索什么——留空表示：不搜索',
} as const

package database

import (
	"encoding/json"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	"mbti-backend/config"
	"mbti-backend/models"
	"mbti-backend/services"
	"mbti-backend/utils"
)

type seedQ struct {
	dim, text, aText, aPole, bText, bPole string
	sort                                  int
}

// 32 题：EI / SN / TF / JP 各 8 题
var seedQuestions = []seedQ{
	// E/I
	{"EI", "在一场聚会上，你更倾向于：", "主动结识新朋友，和很多人聊天", "E", "和少数几个熟人深入交谈", "I", 1},
	{"EI", "忙碌了一整天后，恢复精力的方式是：", "约朋友出去热闹一下", "E", "独自待着安静地放松", "I", 2},
	{"EI", "在团队讨论中，你通常：", "边说边想，乐于第一个发言", "E", "先在心里想清楚再开口", "I", 3},
	{"EI", "别人对你的第一印象更可能是：", "热情、容易接近", "E", "沉稳、需要慢慢熟悉", "I", 4},
	{"EI", "面对一个新项目，你更喜欢：", "和大家协作推进", "E", "独立负责一块再同步", "I", 5},
	{"EI", "周末没有安排时，你会：", "主动组局或接受邀约", "E", "享受一个人在家的时光", "I", 6},
	{"EI", "表达想法时，你更习惯：", "滔滔不绝、语速较快", "E", "字斟句酌、点到为止", "I", 7},
	{"EI", "被陌生人搭讪时，你会：", "自然地聊起来", "E", "有些警惕，简短回应", "I", 8},
	// S/N
	{"SN", "你更相信：", "亲身经历和可验证的事实", "S", "直觉和事物背后的可能性", "N", 1},
	{"SN", "学习新东西时，你偏好：", "按步骤实操，先看具体案例", "S", "先了解整体概念和理论", "N", 2},
	{"SN", "描述一件事时，你更常：", "原原本本讲清细节", "S", "打比方、讲它意味着什么", "N", 3},
	{"SN", "你更欣赏的人是：", "脚踏实地、把事做扎实的人", "S", "富有想象、能看见未来的人", "N", 4},
	{"SN", "面对任务说明，你关注：", "具体要交付什么、怎么做", "S", "为什么要做、还能怎么做", "N", 5},
	{"SN", "你更容易记住：", "当天发生的具体场景和对话", "S", "当时的感受与一闪而过的念头", "N", 6},
	{"SN", "读书时你更喜欢：", "纪实、历史、实用技巧类", "S", "科幻、哲学、脑洞创意类", "N", 7},
	{"SN", "工作中你更擅长：", "处理当下真实存在的问题", "S", "预判趋势、规划未来方向", "N", 8},
	// T/F
	{"TF", "做重要决定时，你首先考虑：", "客观利弊与逻辑是否成立", "T", "涉及的人会有什么感受", "F", 1},
	{"TF", "朋友向你诉苦，你更可能：", "帮 TA 分析问题、给出方案", "T", "先共情陪伴，让 TA 感觉被理解", "F", 2},
	{"TF", "评价一件事，你更看重：", "是否公平、标准是否一致", "T", "是否有人情味、关系是否和谐", "F", 3},
	{"TF", "被人批评时，你倾向于：", "就事论事地讨论对错", "T", "先在意对方的语气和态度", "F", 4},
	{"TF", "团队里出现争执，你会：", "坚持更合理的那个方案", "T", "寻找各方都能接受的折中", "F", 5},
	{"TF", "你认为自己更：", "理性、讲原则", "T", "体贴、重感情", "F", 6},
	{"TF", "看到朋友做了不理智的决定，你会：", "直接指出风险", "T", "委婉表达，怕伤了和气", "F", 7},
	{"TF", "夸奖别人时，你更常说：", "“这事你干得真漂亮”", "T", "“有你在真的太好了”", "F", 8},
	// J/P
	{"JP", "你的日常生活更接近：", "有计划、列清单、按日程走", "J", "随性灵活、走一步看一步", "P", 1},
	{"JP", "出门旅行前，你会：", "提前做好详细攻略和预订", "J", "定个大方向，到了再说", "P", 2},
	{"JP", "面对截止日期，你通常：", "尽早完成，心里才踏实", "J", "在压力下临近时才爆发", "P", 3},
	{"JP", "你的桌面/房间通常：", "整齐有序、各有归处", "J", "看似杂乱但我找得到", "P", 4},
	{"JP", "计划临时被改变，你会：", "有些不适，想重新安排", "J", "觉得没关系，随机应变", "P", 5},
	{"JP", "做一件事，你更在意：", "尽快做出决定并定下来", "J", "保留更多选择、再看看", "P", 6},
	{"JP", "对于规则和流程，你倾向于：", "遵守它，它让事情高效", "J", "觉得可以视情况灵活处理", "P", 7},
	{"JP", "同时有几个任务时，你会：", "排好优先级逐个完成", "J", "在不同任务间来回切换", "P", 8},
}

type seedProfile struct {
	code, nickname, group, groupKey, color, desc, traits, strengths, weaknesses, careers string
}

var seedProfiles = []seedProfile{
	{"INTJ", "建筑师", "分析家", "NT", "#6366f1", "富有远见的战略家，凡事皆可规划。擅长独立思考与系统构建，用理性和高标准推动想法落地。", "独立、战略性、理性、完美主义", "战略思维强、意志坚定、擅长长远规划", "不善表达情感、对他人要求过高、有时过于独断", "架构师、战略顾问、科学家、投资分析师"},
	{"INTP", "逻辑学家", "分析家", "NT", "#6366f1", "永不满足的理论探索者，痴迷于拆解万物运转的逻辑。思想自由，常沉浸在自己的思维世界。", "好奇、严谨、抽象、怀疑精神", "逻辑缜密、创造力强、客观中立", "拖延、忽视现实细节、不擅长情感表达", "研究员、程序员、数学家、哲学家"},
	{"ENTJ", "指挥官", "分析家", "NT", "#4f46e5", "天生的组织者与领导者，决断果敢，擅长调配资源把宏大目标变为现实。", "果断、领导欲、高效、目标导向", "组织能力强、自信、战略执行力出色", "急躁、缺乏耐心、显得强势不近人情", "企业高管、创业者、律师、项目总监"},
	{"ENTP", "辩论家", "分析家", "NT", "#818cf8", "机敏的思想探险家，热爱头脑风暴与智识交锋，总能从不同角度提出新点子。", "机敏、善辩、创新、精力充沛", "思维敏捷、适应力强、善于发现机会", "容易厌倦、爱抬杠、不重细节落地", "产品经理、创业者、市场策划、咨询顾问"},
	{"INFJ", "提倡者", "外交家", "NF", "#10b981", "安静而富有理想主义的引路人，深刻洞察人心，渴望让世界变得更好。最稀有的类型之一。", "理想主义、洞察、利他、坚定", "共情力强、有远见、坚守价值观", "过度内耗、完美主义、难以敞开心扉", "心理咨询师、作家、公益组织者、HR"},
	{"INFP", "调停者", "外交家", "NF", "#34d399", "浪漫真诚的理想主义者，内心有一片丰盈世界，追寻意义、真诚与自我实现。", "真诚、理想主义、创造力、温柔", "想象力丰富、忠诚、价值观坚定", "过于敏感、逃避现实、对自己苛刻", "作家、编辑、艺术创作者、社工"},
	{"ENFJ", "主人公", "外交家", "NF", "#059669", "热情有魅力的鼓舞者，天然关注他人成长，善于凝聚人心朝共同愿景前进。", "热情、感召力、利他、负责", "领导力强、善于沟通、富有同理心", "过度理想化、容易燃烧自己、太在意外界评价", "教师、培训师、团队负责人、公关"},
	{"ENFP", "竞选者", "外交家", "NF", "#6ee7b7", "热情洋溢的自由灵魂，对人和可能性充满好奇，是社交场合中带来活力的催化剂。", "热情、好奇、即兴、富感染力", "创意丰富、社交力强、适应力佳", "注意力分散、情绪化、不擅收尾", "自媒体人、记者、创意策划、用户运营"},
	{"ISTJ", "物流师", "守护者", "SJ", "#f59e0b", "务实可靠的执行者，重视责任、秩序与传统，承诺过的事一定会稳稳完成。", "严谨、负责、务实、守规矩", "可靠细致、执行力强、忠诚稳重", "保守、不够灵活、不擅表达情感", "会计、审计师、公务员、工程师"},
	{"ISFJ", "守卫者", "守护者", "SJ", "#fbbf24", "温暖尽责的守护者，默默关注身边人的需要，用细致与耐心维系着日常的安稳。", "细心、奉献、耐心、忠诚", "体贴周到、责任心强、记忆力好", "不善拒绝、压抑自身需求、抗拒变化", "护士、教师、行政、客户服务"},
	{"ESTJ", "总经理", "守护者", "SJ", "#d97706", "井井有条的组织者，信奉规则与秩序，擅长把人和事安排得明明白白。", "务实、秩序感、直率、组织力", "决断力强、执行力出色、可靠", "固执、不近人情、难容忍不同意见", "管理者、法官、财务主管、运营"},
	{"ESFJ", "执政官", "守护者", "SJ", "#fcd34d", "热心肠的社交黏合剂，重视和谐与归属，乐于照顾他人，是群体中最暖的存在。", "热心、合作、健谈、尽责", "善解人意、组织协调、忠诚可靠", "太在意他人评价、害怕冲突、依赖认可", "人力资源、教师、销售、社区运营"},
	{"ISTP", "鉴赏家", "探险家", "SP", "#f97316", "冷静灵活的动手大师，擅长在当下拆解问题、驾驭工具，危机中反而最沉着。", "冷静、务实、独立、爱探索", "动手能力强、临危不乱、效率高", "情感封闭、容易厌倦、抗拒承诺", "工程师、机械师、飞行员、数据分析师"},
	{"ISFP", "探险家", "探险家", "SP", "#fb923c", "安静敏感的生活艺术家，用审美和行动表达自我，追求真实自在的生活方式。", "温和、感性、艺术气质、随性", "审美出众、富有同理心、灵活", "极度不喜冲突、缺乏长远规划、怕压力", "设计师、摄影师、厨师、音乐人"},
	{"ESTP", "企业家", "探险家", "SP", "#ea580c", "精力旺盛的行动派，享受冒险与当下，反应快、胆子大，在真实世界里如鱼得水。", "大胆、现实、即兴、爱冒险", "行动力强、抗压、谈判高手", "冲动、缺乏耐心、易忽视长远后果", "销售、创业者、急救人员、运动员"},
	{"ESFP", "表演者", "探险家", "SP", "#fdba74", "天生的派对灵魂，热爱生活与人群，用热情和幽默把快乐带给身边每个人。", "热情、乐观、爱玩、感染力", "审美佳、善于社交、务实灵活", "难以专注、回避难题、计划性弱", "演员、导游、活动策划、导购"},
}

// Seed 幂等灌入题库、画像与演示数据
func Seed() {
	seedQuestionsData()
	seedProfilesData()
	seedDemoData()
}

func seedQuestionsData() {
	var cnt int64
	DB.Model(&models.Question{}).Count(&cnt)
	if cnt > 0 {
		return
	}
	for _, q := range seedQuestions {
		DB.Create(&models.Question{
			Dimension: q.dim, Text: q.text,
			OptionAText: q.aText, OptionAPole: q.aPole,
			OptionBText: q.bText, OptionBPole: q.bPole,
			Sort: q.sort,
		})
	}
	log.Println("已灌入 32 道测评题")
}

func seedProfilesData() {
	var cnt int64
	DB.Model(&models.TypeProfile{}).Count(&cnt)
	if cnt > 0 {
		return
	}
	for _, p := range seedProfiles {
		DB.Create(&models.TypeProfile{
			Code: p.code, Nickname: p.nickname, Group: p.group, GroupKey: p.groupKey,
			Color: p.color, Description: p.desc, Traits: p.traits,
			Strengths: p.strengths, Weaknesses: p.weaknesses, Careers: p.careers,
		})
	}
	log.Println("已灌入 16 型人格画像")
}

func dimsJSON(picks []string) string {
	_, dims := utils.ScoreMBTI(picks)
	b, _ := json.Marshal(dims)
	return string(b)
}

func seedDemoData() {
	var cnt int64
	DB.Model(&models.User{}).Count(&cnt)
	if cnt > 0 {
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)
	mk := func(username, nick, mbti string) *models.User {
		u := &models.User{
			Username: username, PasswordHash: string(hash), Nickname: nick,
			CoinBalance: config.InitialCoins, FrozenBalance: 0, MBTIType: mbti,
			CreatedAt: time.Now().Add(-72 * time.Hour),
		}
		if err := DB.Create(u).Error; err != nil {
			log.Printf("演示用户创建失败 %s: %v", username, err)
		}
		DB.Create(&models.CoinTransaction{
			UserID: u.ID, ChangeAmount: config.InitialCoins, BalanceAfter: config.InitialCoins,
			Type: models.TxnRegister, Remark: "注册赠送测评币", CreatedAt: u.CreatedAt,
		})
		return u
	}

	// 三个演示用户初始均为 200 币，后续冻结/结算全部走 service 事务，保证账目自洽
	lin := mk("linxia", "林小夏", "INFP")
	zhou := mk("zhouye", "周野", "ENTJ")
	chen := mk("chenan", "陈安", "ISFJ")

	// A 的测评结果 INFP（偏 I、N、F、P）
	infpDims := dimsJSON([]string{
		"I", "I", "I", "I", "I", "E", "I", "E", // EI: I 6 / E 2
		"N", "N", "N", "S", "N", "N", "N", "S", // SN: N 6 / S 2
		"F", "F", "F", "F", "F", "F", "T", "F", // TF: F 7 / T 1
		"P", "P", "J", "P", "P", "P", "P", "P", // JP: P 7 / J 1
	})
	resA := &models.AssessmentResult{
		UserID: lin.ID, TypeCode: "INFP", Dimensions: infpDims,
		CreatedAt: time.Now().Add(-50 * time.Hour),
	}
	DB.Create(resA)

	// B 的测评结果 ENTJ（偏 E、N、T、J）
	entjDims := dimsJSON([]string{
		"E", "E", "E", "E", "E", "E", "E", "I", // EI: E 7 / I 1
		"N", "S", "N", "N", "N", "S", "N", "N", // SN: N 6 / S 2
		"T", "T", "T", "F", "T", "T", "T", "T", // TF: T 7 / F 1
		"J", "J", "J", "J", "P", "J", "J", "J", // JP: J 7 / P 1
	})
	resB := &models.AssessmentResult{
		UserID: zhou.ID, TypeCode: "ENTJ", Dimensions: entjDims,
		CreatedAt: time.Now().Add(-49 * time.Hour),
	}
	DB.Create(resB)

	// C 的测评结果 ISFJ
	isfjDims := dimsJSON([]string{
		"I", "I", "E", "I", "I", "I", "I", "E",
		"S", "S", "S", "N", "S", "S", "N", "S",
		"F", "T", "F", "F", "F", "F", "F", "F",
		"J", "J", "J", "J", "J", "J", "P", "J",
	})
	resC := &models.AssessmentResult{
		UserID: chen.ID, TypeCode: "ISFJ", Dimensions: isfjDims,
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}
	DB.Create(resC)

	// 求解读 1：林小夏（open，悬赏 50），已有陈安的解读
	req1, err := services.CreateRequest(DB, lin.ID, resA.ID,
		"INFP 在职场总被说“太敏感”，我该改变吗？",
		"测评说我是典型 INFP，共情力强但内耗严重。团队里我常能察觉到别人没说出口的情绪，可表达不同意见时总怕伤害关系，会后又反复回想。希望有过来人帮我解读：这是性格缺陷还是可以利用的优势？",
		50)
	if err != nil {
		log.Printf("演示求解读1创建失败: %v", err)
	}

	i1 := &models.Interpretation{
		RequestID: req1.ID, UserID: chen.ID,
		Content:   "同为高敏型人格，我的体会是：共情不是缺陷，而是你读空气、做用户洞察的天赋。你需要补的不是“变钝”，而是给善良装上边界——表达异议时可以用“事实+感受+请求”的句式，既不攻击别人，也不委屈自己。内耗的本质常是把别人的情绪当成自己的责任，试着在每次反刍时问自己：这是谁的课题？",
		CreatedAt: time.Now().Add(-20 * time.Hour),
	}
	DB.Create(i1)

	// 求解读 2：周野（悬赏 80），陈安解读后被采纳结算
	req2, err := services.CreateRequest(DB, zhou.ID, resB.ID,
		"ENTJ 带团队效率很高，为什么朋友说我压迫感太强？",
		"我习惯目标导向、快速决策，团队业绩确实不错，但 360 评估里好几个人写“难以接近”。理性和效率难道不对吗？想听听不同视角的解读。",
		80)
	if err != nil {
		log.Printf("演示求解读2创建失败: %v", err)
	}
	i2 := &models.Interpretation{
		RequestID: req2.ID, UserID: chen.ID,
		Content:   "效率没有错，但你优化的是“事”，团队还需要被看见“人”。ENTJ 的压迫感常来自过快收口——别人还在消化情绪，你已经在布置下一步了。可以试试两个小动作：决策后多问一句“大家还有什么顾虑”，以及公开场合先肯定再指出问题。目标不变，只是让团队愿意跟你跑得更远。",
		CreatedAt: time.Now().Add(-30 * time.Hour),
	}
	DB.Create(i2)
	if err := services.AcceptInterpretation(DB, zhou.ID, req2.ID, i2.ID); err != nil {
		log.Printf("演示采纳结算失败: %v", err)
	}

	// 求解读 3：陈安（open，悬赏 30），暂无解读
	if _, err = services.CreateRequest(DB, chen.ID, resC.ID,
		"ISFJ 总是不懂拒绝，把自己累垮了怎么办？",
		"同事的请求我几乎都答应，怕拒绝会破坏关系，结果自己的活加班到很晚。测评显示我是 ISFJ 守卫者，难道这个类型注定讨好？",
		30); err != nil {
		log.Printf("演示求解读3创建失败: %v", err)
	}

	log.Println("已灌入演示数据：3 个用户 / 3 条测评 / 3 条求解读（登录密码 demo123）")
}

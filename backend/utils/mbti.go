package utils

// 四组维度，每个维度两个极点
var dimensions = [4][2]string{
	{"E", "I"},
	{"S", "N"},
	{"T", "F"},
	{"J", "P"},
}

// DimensionStat 单个维度的统计
type DimensionStat struct {
	Pair      string `json:"pair"`       // 如 "E/I"
	PoleA     string `json:"pole_a"`     // E
	PoleB     string `json:"pole_b"`     // I
	ScoreA    int    `json:"score_a"`    // A 极得分
	ScoreB    int    `json:"score_b"`    // B 极得分
	PercentA  int    `json:"percent_a"`  // A 极占比（0-100）
	PercentB  int    `json:"percent_b"`  // B 极占比
	Dominant  string `json:"dominant"`   // 占优极点
	LabelA    string `json:"label_a"`    // 中文释义
	LabelB    string `json:"label_b"`
}

var poleLabels = map[string]string{
	"E": "外向", "I": "内向",
	"S": "感觉", "N": "直觉",
	"T": "思考", "F": "情感",
	"J": "判断", "P": "感知",
}

// ScoreMBTI 根据每题选择的极点计算类型与四维占比
// picks: 题号顺序无所谓，元素为每题所选极点 E/I/S/N/T/F/J/P
// 平局时偏向 I/N/F/P（自我探索者更常见的一极）
func ScoreMBTI(picks []string) (string, []DimensionStat) {
	counts := map[string]int{}
	for _, p := range picks {
		counts[p]++
	}

	typeCode := ""
	stats := make([]DimensionStat, 0, 4)
	for _, pair := range dimensions {
		a, b := pair[0], pair[1]
		sa, sb := counts[a], counts[b]
		total := sa + sb
		if total == 0 {
			total = 1
		}
		pctA := sa * 100 / total
		pctB := 100 - pctA

		dominant := a
		if sb > sa {
			dominant = b
		} else if sa == sb {
			// 平局判定
			dominant = tieBreak(b)
		}
		typeCode += dominant
		stats = append(stats, DimensionStat{
			Pair:     a + "/" + b,
			PoleA:    a, PoleB: b,
			ScoreA: sa, ScoreB: sb,
			PercentA: pctA, PercentB: pctB,
			Dominant: dominant,
			LabelA:   poleLabels[a], LabelB: poleLabels[b],
		})
	}
	return typeCode, stats
}

// tieBreak 平局时的偏向：I、N、F、P
func tieBreak(pole string) string {
	switch pole {
	case "I", "N", "F", "P":
		return pole
	default:
		return oppositeOf(pole)
	}
}

func oppositeOf(pole string) string {
	switch pole {
	case "E":
		return "I"
	case "S":
		return "N"
	case "T":
		return "F"
	case "J":
		return "P"
	}
	return pole
}

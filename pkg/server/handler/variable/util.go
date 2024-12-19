package variable

import "kusionstack.io/kusion/pkg/domain/entity"

// CalculateLabelMatchScore returns the score of label matching.
func CalculateLabelMatchScore(variable *entity.Variable, variableLabels *entity.VariableLabels, matchedLabels map[string]string) int {
	labelScoreMap := make(map[string]int, len(variableLabels.Labels))
	for i, labelKey := range variableLabels.Labels {
		labelScoreMap[labelKey] = 1 << i
	}

	score := 0
	for labelKey, labelValue := range variable.Labels {
		if matchedValue, ok := matchedLabels[labelKey]; ok && matchedValue != labelValue {
			// Directly return 0 if any of the labels not matched.
			return 0
		}

		score += labelScoreMap[labelKey]
	}

	return score
}

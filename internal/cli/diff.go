package cli

import "github.com/shipquiz/shipquiz/internal/quiz"

func diffForQuiz() (quiz.Diff, error) {
	return quiz.WorkingDiff()
}

func parsePush(raw string) []quiz.Ref {
	return quiz.ParsePushRefs(raw)
}

func pushDiff(raw string) (quiz.Diff, error) {
	return quiz.PushDiff(quiz.ParsePushRefs(raw))
}

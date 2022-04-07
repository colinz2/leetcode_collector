package leetcode_cli

/*
Thanks to https://transform.tools/json-to-go
*/

type ProblemStatus struct {
	Stat struct {
		QuestionID          int    `json:"question_id"`
		QuestionTitle       string `json:"question__title"`
		QuestionTitleSlug   string `json:"question__title_slug"`
		QuestionHide        bool   `json:"question__hide"`
		TotalAcs            int    `json:"total_acs"`
		TotalSubmitted      int    `json:"total_submitted"`
		TotalColumnArticles int    `json:"total_column_articles"`
		FrontendQuestionID  string `json:"frontend_question_id"`
		IsNewQuestion       bool   `json:"is_new_question"`
	} `json:"stat"`
	Status     string `json:"status"`
	Difficulty struct {
		Level int `json:"level"`
	} `json:"difficulty"`
	PaidOnly  bool    `json:"paid_only"`
	IsFavor   bool    `json:"is_favor"`
	Frequency float64 `json:"frequency"`
	Progress  float64 `json:"progress"`
}

func (p ProblemStatus) IsAc() bool {
	return p.Status == "ac"
}

type AllProblemsResponse struct {
	UserName        string          `json:"user_name"`
	NumSolved       int             `json:"num_solved"`
	NumTotal        int             `json:"num_total"`
	AcEasy          int             `json:"ac_easy"`
	AcMedium        int             `json:"ac_medium"`
	AcHard          int             `json:"ac_hard"`
	StatStatusPairs []ProblemStatus `json:"stat_status_pairs"`
	FrequencyHigh   float64         `json:"frequency_high"`
	FrequencyMid    float64         `json:"frequency_mid"`
	CategorySlug    string          `json:"category_slug"`
}

type TopicTag struct {
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	TranslatedName string `json:"translatedName"`
}

type Question struct {
	Content           string     `json:"content"`
	Hints             []string   `json:"hints"`
	QuestionID        string     `json:"questionId"`
	SimilarQuestions  string     `json:"similarQuestions"`
	TopicTags         []TopicTag `json:"topicTags"`
	TranslatedContent string     `json:"translatedContent"`
	TranslatedTitle   string     `json:"translatedTitle"`
}

type QuestionDetailResponse struct {
	Question Question `json:"question"`
}

type Submission struct {
	Typename      string `json:"__typename"`
	ID            string `json:"id"`
	IsPending     string `json:"isPending"`
	Lang          string `json:"lang"`
	Memory        string `json:"memory"`
	Runtime       string `json:"runtime"`
	StatusDisplay string `json:"statusDisplay"`
	Timestamp     string `json:"timestamp"`
	URL           string `json:"url"`
}

type SubmissionsByQuestionResponse struct {
	SubmissionList struct {
		Typename    string       `json:"__typename"`
		HasNext     bool         `json:"hasNext"`
		LastKey     string       `json:"lastKey"`
		Submissions []Submission `json:"submissions"`
	} `json:"submissionList"`
}

type SubmissionDetail struct {
	Typename     string `json:"__typename"`
	Code         string `json:"code"`
	ID           string `json:"id"`
	Lang         string `json:"lang"`
	Memory       string `json:"memory"`
	OutputDetail struct {
		Typename       string `json:"__typename"`
		CodeOutput     string `json:"codeOutput"`
		CompileError   string `json:"compileError"`
		ExpectedOutput string `json:"expectedOutput"`
		Input          string `json:"input"`
		LastTestcase   string `json:"lastTestcase"`
		RuntimeError   string `json:"runtimeError"`
	} `json:"outputDetail"`
	PassedTestCaseCnt int `json:"passedTestCaseCnt"`
	Question          struct {
		Typename        string `json:"__typename"`
		QuestionID      string `json:"questionId"`
		Title           string `json:"title"`
		TitleSlug       string `json:"titleSlug"`
		TranslatedTitle string `json:"translatedTitle"`
	} `json:"question"`
	RawMemory         string      `json:"rawMemory"`
	Runtime           string      `json:"runtime"`
	SourceURL         string      `json:"sourceUrl"`
	StatusDisplay     string      `json:"statusDisplay"`
	SubmissionComment interface{} `json:"submissionComment"`
	Timestamp         int         `json:"timestamp"`
	TotalTestCaseCnt  int         `json:"totalTestCaseCnt"`
}

type SubmissionDetailResponse struct {
	SubmissionDetail *SubmissionDetail `json:"submissionDetail"`
}

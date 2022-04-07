package leetcode_cli

const (
	// QueryQuestion 通过题目的slug 获取题目的信息
	QueryQuestion = `
		query questionData($titleSlug: String!) {
			question(titleSlug: $titleSlug) {
				questionId
				content
				translatedTitle
				translatedContent
				similarQuestions
				topicTags {
					name
					slug
					translatedName
				}
				hints
			}
		}
`
	// QuerySubmissionDetail 获取一次提交的信息，获取源码
	QuerySubmissionDetail = `
		query mySubmissionDetail($id: ID!) {
		  submissionDetail(submissionId: $id) {
			id
			code
			runtime
			memory
			rawMemory
			statusDisplay
			timestamp
			lang
			passedTestCaseCnt
			totalTestCaseCnt
			sourceUrl
			question {
			  titleSlug
			  title
			  translatedTitle
			  questionId
			  __typename
			}
			... on GeneralSubmissionNode {
			  outputDetail {
				codeOutput
				expectedOutput
				input
				compileError
				runtimeError
				lastTestcase
				__typename
			  }
			  __typename
			}
			submissionComment {
			  comment
			  flagType
			  __typename
			}
			__typename
		  }
		}
`
	// QuerySubmissionByQuestionSlug 获取提交的 ID
	QuerySubmissionByQuestionSlug = `
	query submissions($offset: Int!, $limit: Int!, $lastKey: String, $questionSlug: String!) {
		submissionList(offset: $offset, limit: $limit, lastKey: $lastKey, questionSlug: $questionSlug) {
			lastKey
			hasNext
			submissions {
				id
				statusDisplay
				lang
				runtime
				timestamp
				url
				isPending
				memory
				__typename
			}
		__typename
		}
	}
`
)

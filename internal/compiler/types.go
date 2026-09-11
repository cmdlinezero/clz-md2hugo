package compiler

type Content struct {
	ID             string      `json:"id"`
	Source         string      `json:"source,omitempty"`
	Type           string      `json:"type"`
	Title          string      `json:"title"`
	Description    string      `json:"description,omitempty"`
	Slug           string      `json:"slug"`
	Date           string      `json:"date,omitempty"`
	Status         string      `json:"status"`
	Draft          bool        `json:"draft"`
	Kind           string      `json:"kind,omitempty"`
	Provider       string      `json:"provider,omitempty"`
	RatingID       string      `json:"rating_id,omitempty"`
	ProductID      string      `json:"product_id,omitempty"`
	Categories     []string    `json:"categories,omitempty"`
	Tags           []string    `json:"tags,omitempty"`
	Topics         []string    `json:"topics,omitempty"`
	Collections    []string    `json:"collections,omitempty"`
	Certifications []string    `json:"certifications,omitempty"`
	Prerequisites  []string    `json:"prerequisites,omitempty"`
	Children       []string    `json:"children,omitempty"`
	Activities     []string    `json:"activities,omitempty"`
	Options        []string    `json:"options,omitempty"`
	Answer         int         `json:"answer,omitempty"`
	Explanation    string      `json:"explanation,omitempty"`
	Duration       string      `json:"duration,omitempty"`
	HeroTitle      string      `json:"hero_title,omitempty"`
	HeroImage      string      `json:"hero_image,omitempty"`
	Issue          int         `json:"issue,omitempty"`
	Volume         int         `json:"volume,omitempty"`
	Special        bool        `json:"special_edition,omitempty"`
	Steps          []Step      `json:"steps,omitempty"`
	Conclusion     string      `json:"conclusion,omitempty"`
	Cards          []Flashcard `json:"cards,omitempty"`
	Challenges     []Challenge `json:"challenges,omitempty"`
}

type Step struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	Duration string  `json:"duration,omitempty"`
	Blocks   []Block `json:"blocks"`
}

type Block struct {
	Type       string            `json:"type"`
	Language   string            `json:"language,omitempty"`
	Content    string            `json:"content,omitempty"`
	Title      string            `json:"title,omitempty"`
	Level      string            `json:"level,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Copy       bool              `json:"copy,omitempty"`
	Terminal   bool              `json:"terminal,omitempty"`
	FileID     string            `json:"file_id,omitempty"`
	CmdID      string            `json:"cmd_id,omitempty"`
}

type ContentGraph struct {
	Version        int        `json:"version"`
	Generated      string     `json:"generated"`
	Collections    []Content  `json:"collections,omitempty"`
	LearningPaths  []Content  `json:"learningPaths,omitempty"`
	Certifications []Content  `json:"certifications,omitempty"`
	Tutorials      []Content  `json:"tutorials,omitempty"`
	Quizzes        []Content  `json:"quizzes,omitempty"`
	Questions      []Content  `json:"questions,omitempty"`
	Topics         []Content  `json:"topics,omitempty"`
	FlashcardDecks []Content  `json:"flashcardDecks,omitempty"`
	ChallengeLabs  []Content  `json:"challengeLabs,omitempty"`
	Other          []Content  `json:"other,omitempty"`
	Stats          BuildStats `json:"stats"`
}

type BuildStats struct {
	Documents      int `json:"documents"`
	Collections    int `json:"collections"`
	LearningPaths  int `json:"learningPaths"`
	Certifications int `json:"certifications"`
	Tutorials      int `json:"tutorials"`
	Quizzes        int `json:"quizzes"`
	Questions      int `json:"questions"`
	Topics         int `json:"topics"`
	FlashcardDecks int `json:"flashcardDecks"`
	ChallengeLabs  int `json:"challengeLabs"`
	Steps          int `json:"steps"`
	Blocks         int `json:"blocks"`
}

type Flashcard struct {
	ID          string   `json:"id"`
	Front       string   `json:"front"`
	Back        string   `json:"back"`
	Hint        string   `json:"hint,omitempty"`
	Explanation string   `json:"explanation,omitempty"`
	Topics      []string `json:"topics,omitempty"`
}

type Challenge struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Objective    string            `json:"objective,omitempty"`
	Instructions string            `json:"instructions,omitempty"`
	Hint         string            `json:"hint,omitempty"`
	Success      string            `json:"success,omitempty"`
	Command      string            `json:"command,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

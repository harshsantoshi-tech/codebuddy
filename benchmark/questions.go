package benchmark

// TestCase is one evaluation question with the file we expect to appear
// in top-5 results. This is how we measure retrieval precision.
type TestCase struct {
	Query         string
	ExpectedFiles []string // at least one of these should appear in top-5
	Category      string   // for grouping in the report
}

// TestSuite is our eval set — 20 questions covering different query types.
// Designed for the godotenv repo (github.com/joho/godotenv)
// Change ExpectedFiles if you're testing a different repo.
var TestSuite = []TestCase{
	// --- Feature understanding ---
	{
		Query:         "how are .env files parsed?",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "feature",
	},
	{
		Query:         "how do you load environment variables from a file?",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "feature",
	},
	{
		Query:         "how does the Read function work?",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "feature",
	},
	{
		Query:         "how are quoted strings handled in env files?",
		ExpectedFiles: []string{"godotenv.go", "parser.go"},
		Category:      "feature",
	},
	{
		Query:         "what happens when a variable is already set in the environment?",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "feature",
	},

	// --- Bug investigation ---
	{
		Query:         "where could parsing fail?",
		ExpectedFiles: []string{"godotenv.go", "parser.go"},
		Category:      "bug",
	},
	{
		Query:         "how are errors returned to the caller?",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "bug",
	},
	{
		Query:         "what happens with malformed env file lines?",
		ExpectedFiles: []string{"godotenv.go", "parser.go"},
		Category:      "bug",
	},
	{
		Query:         "how are comments handled?",
		ExpectedFiles: []string{"godotenv.go", "parser.go"},
		Category:      "bug",
	},
	{
		Query:         "what happens if the file does not exist?",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "bug",
	},

	// --- Architecture ---
	{
		Query:         "what are the main exported functions?",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "architecture",
	},
	{
		Query:         "how is the package structured?",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "architecture",
	},
	{
		Query:         "how does Overload differ from Load?",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "architecture",
	},
	{
		Query:         "how does the Write function work?",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "architecture",
	},
	{
		Query:         "how are multiple env files handled?",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "architecture",
	},

	// --- Vague / fuzzy queries (hardest for retrieval) ---
	{
		Query:         "export",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "fuzzy",
	},
	{
		Query:         "parse environment",
		ExpectedFiles: []string{"godotenv.go", "parser.go"},
		Category:      "fuzzy",
	},
	{
		Query:         "file not found",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "fuzzy",
	},
	{
		Query:         "overwrite existing",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "fuzzy",
	},
	{
		Query:         "write back to disk",
		ExpectedFiles: []string{"godotenv.go"},
		Category:      "fuzzy",
	},
}
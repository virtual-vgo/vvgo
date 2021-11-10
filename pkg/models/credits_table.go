package models

// CreditsTable is the credits for a project.
type CreditsTable []CreditsTopicRow

// CreditsTopicRow represents either all performers or all crew.
type CreditsTopicRow struct {
	Name   string
	Rows   []*CreditsTeamRow
	rowMap map[string]*CreditsTeamRow
}

// CreditsTeamRow represents either a production team or instrument section.
type CreditsTeamRow struct {
	Name string
	Rows []Credit
}

func BuildCreditsTable(credits Credits, project Project) CreditsTable {
	var rows []*CreditsTopicRow
	rowMap := make(map[string]*CreditsTopicRow)
	for _, projectCredit := range credits.ForProject(project.Name) {
		if rowMap[projectCredit.MajorCategory] == nil {
			rowMap[projectCredit.MajorCategory] = new(CreditsTopicRow)
			rowMap[projectCredit.MajorCategory].Name = projectCredit.MajorCategory
			rowMap[projectCredit.MajorCategory].rowMap = make(map[string]*CreditsTeamRow)
			rows = append(rows, rowMap[projectCredit.MajorCategory])
		}
		major := rowMap[projectCredit.MajorCategory]
		if major.rowMap[projectCredit.MinorCategory] == nil {
			major.rowMap[projectCredit.MinorCategory] = new(CreditsTeamRow)
			major.rowMap[projectCredit.MinorCategory].Name = projectCredit.MinorCategory
			major.Rows = append(major.Rows, major.rowMap[projectCredit.MinorCategory])
		}
		minor := major.rowMap[projectCredit.MinorCategory]
		minor.Rows = append(minor.Rows, projectCredit)
	}

	table := make(CreditsTable, len(rows))
	for i := range rows {
		table[i] = *rows[i]
	}
	return table
}

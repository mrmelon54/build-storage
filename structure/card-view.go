package structure

type CardView struct {
	Title    string
	PagePath string
	PageName string
	BasePath string
	Sections map[string]CardSection
}

type CardSection struct {
	Name  string
	Cards map[string]CardItem
}

type CardItem struct {
	Name string
	Icon string
}

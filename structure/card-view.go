package structure

type CardView struct {
	Title    string
	PagePath string
	BasePath string
	Sections []CardSection
}

type CardSection struct {
	Name  string
	Cards []CardItem
}

type CardItem struct {
	Name    string
	Icon    string
	Address string
}

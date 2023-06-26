package structure

type CardView struct {
	Title    string
	PagePath string
	BasePath string
	Login    string
	Icon     string
	Sections []CardSection
}

type CardSection struct {
	Name  string
	Style string
	Cards []CardItem
}

type CardItem struct {
	Name      string
	Icon      string
	Address   string
	CanUpload bool
	Sha256    string
	Sha512    string
}

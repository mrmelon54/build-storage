package structure

import (
	"log"
)

func GetUploadMeta(name string, parser ParserYaml) (projectName string, layers []string) {
	matches := parser.Exp.FindStringSubmatch(name)
	if len(matches) < 1 {
		log.Printf("Match failed: '%s' with '%s'\n", name, parser.Exp)
		return
	}
	projectName = matches[parser.Exp.SubexpIndex(parser.Name)]
	layers = make([]string, len(parser.Layers))
	for i := range parser.Layers {
		layers[i] = matches[parser.Exp.SubexpIndex(parser.Layers[i])]
	}
	return
}

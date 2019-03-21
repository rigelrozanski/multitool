package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"unicode"

	"github.com/rigelrozanski/common"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(CreateAlias)
}

// CreateAlias creates autogen code for exposed functions
var CreateAlias = &cobra.Command{
	Use: "create-alias",
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := ioutil.ReadDir("./")
		if err != nil {
			return err
		}

		if len(files) == 0 {
			return errors.New("no files")
		}

		thisPath, err := os.Getwd()
		srcPrefix := path.Join(os.ExpandEnv("$GOPATH"), "src") + "/"
		importDir := strings.TrimPrefix(thisPath, srcPrefix)
		importName := path.Base(importDir)

		if err != nil {
			return err
		}
		var funcNames, constNames, varNames, typeNames []string

		for _, f := range files {
			fname := f.Name()
			if path.Ext(fname) == ".go" {
				lines, err := common.ReadLines(fname)
				if err != nil {
					continue
				}

				// get the exported functions
				for _, line := range lines {
					sep := strings.Fields(line)
					if len(sep) > 2 && sep[0] == "func" {
						if sep[1] == "init()" {
							continue
						}
						if unicode.IsLetter(rune(sep[1][0])) && string(sep[1][0]) == strings.ToUpper(string(sep[1][0])) {
							funcNames = append(funcNames, strings.Split(sep[1], "(")[0])
						}
					}
				}

				// get the exported constants
				constNames = append(constNames, getVarOrConstBlocks(lines, "const")...)

				// get the exported variables
				varNames = append(varNames, getVarOrConstBlocks(lines, "var")...)

				// get the exported types
				for _, line := range lines {
					sep := strings.Fields(line)
					if len(sep) > 2 &&
						sep[0] == "type" &&
						unicode.IsLetter(rune(sep[1][0])) &&
						string(sep[1][0]) == strings.ToUpper(string(sep[1][0])) {

						typeNames = append(typeNames, sep[1])
					}
				}

			}
		}

		if len(funcNames)+len(varNames)+len(constNames)+len(typeNames) == 0 {
			return nil
		}
		out := fmt.Sprintf("import (\n\t%v\n)", importDir)

		if len(constNames) > 0 {
			out += fmt.Sprintf("\n\nconst (")
			for _, constName := range constNames {
				out += fmt.Sprintf("\n\t%v = %v.%v", constName, importName, constName)
			}
			out += fmt.Sprintf("\n)")
		}

		if len(funcNames)+len(varNames) > 0 {
			out += fmt.Sprintf("\n\nvar (")
			if len(funcNames) > 0 {
				out += fmt.Sprintf("\n\t// functions aliases")
				for _, funcName := range funcNames {
					out += fmt.Sprintf("\n\t%v = %v.%v", funcName, importName, funcName)
				}
			}

			if len(varNames) > 0 {
				out += fmt.Sprintf("\n\t// variable aliases")
				for _, varName := range varNames {
					out += fmt.Sprintf("\n\t%v = %v.%v", varName, importName, varName)
				}
			}
			out += fmt.Sprintf("\n)")
		}

		if len(typeNames) > 0 {
			out += fmt.Sprintf("\n\ntype (")
			for _, typeName := range typeNames {
				out += fmt.Sprintf("\n\t%v = %v.%v", typeName, importName, typeName)
			}
			out += fmt.Sprintf("\n)")
		}

		fmt.Println(out)
		return nil
	},
}

func getVarOrConstBlocks(lines []string, varOrConst string) (names []string) {

	withinBlock := false // within a "var (" or "const (" block

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, varOrConst+" (") && withinBlock == false:
			withinBlock = true
			continue
		case strings.HasPrefix(line, ")") && withinBlock == true:
			withinBlock = false
			continue

		case withinBlock:
			sep := strings.Fields(line)
			switch {
			case len(sep) < 2:
				continue
			case !unicode.IsLetter(rune(sep[0][0])):
				continue
			default:
				leftOfEq := strings.Split(line, "=")
				lcns := strings.Split(leftOfEq[0], ",")
				for _, lcn := range lcns {
					lcnT := strings.TrimSpace(lcn)
					lcnT = strings.Fields(lcn)[0]
					if unicode.IsLetter(rune(lcnT[0])) && string(lcnT[0]) == strings.ToUpper(string(lcnT[0])) {
						names = append(names, lcnT)
					}
				}
			}

		case strings.HasPrefix(line, varOrConst) && withinBlock == false:
			sep := strings.Fields(line)
			switch {
			case len(sep) < 2:
				continue
			case !unicode.IsLetter(rune(sep[1][0])):
				continue
			default:
				afterConst := strings.TrimPrefix(line, varOrConst+" ")
				leftOfEq := strings.Split(afterConst, "=")
				lcns := strings.Split(leftOfEq[0], ",")
				for _, lcn := range lcns {
					lcnT := strings.TrimSpace(lcn)
					lcnT = strings.Fields(lcn)[0]
					if unicode.IsLetter(rune(lcnT[0])) && string(lcnT[0]) == strings.ToUpper(string(lcnT[0])) {
						names = append(names, lcnT)
					}
				}
			}
		}
	}
	return names
}

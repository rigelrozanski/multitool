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
	RootCmd.AddCommand(UpdateAlias)
}

const (
	// keyword for alias gen
	AliasKeyword = "ALIASGEN:"
	// escape word to not include in alias file
	NoAliasEsc = "noalias"
)

// CreateAlias creates autogen code for exposed functions
var UpdateAlias = &cobra.Command{
	Use:  "update-alias [alias-file-dir]",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		lines, err := common.ReadLines(args[0])
		if err != nil {
			return err
		}

		srcPrefix := path.Join(os.ExpandEnv("$GOPATH"), "src") + "/"

		var importDirs, fullDirs []string
		for _, line := range lines {
			if strings.Contains(line, AliasKeyword) {
				importDir := strings.TrimSpace(strings.Split(line, AliasKeyword)[1])
				importDirs = append(importDirs, importDir)
				fullDir := path.Join(srcPrefix, importDir)
				fullDirs = append(fullDirs, fullDir)
			}
		}

		aliasFilePath := args[0]
		if !common.FileExists(aliasFilePath) {
			thisPath, err := os.Getwd()
			if err != nil {
				return err
			}

			aliasFilePath = path.Join(thisPath, args[0])
		}
		packageName := path.Base(path.Dir(aliasFilePath))
		return CreateAliasFromDirs(packageName, aliasFilePath, importDirs, fullDirs)
	},
}

// CreateAlias creates autogen code for exposed functions
var CreateAlias = &cobra.Command{
	Use:  "create-alias [dirs...]",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		thisPath, err := os.Getwd()
		if err != nil {
			return err
		}
		srcPrefix := path.Join(os.ExpandEnv("$GOPATH"), "src") + "/"
		baseImportDir := strings.TrimPrefix(thisPath, srcPrefix)

		var importDirs, fullDirs []string
		for _, dir := range args {
			importDir := path.Join(baseImportDir, dir)
			importDir = strings.TrimSuffix(importDir, "/")
			importDirs = append(importDirs, importDir)
			fullDirs = append(fullDirs, path.Join(srcPrefix, importDir))
		}

		packageName := path.Base(thisPath)
		aliasFilePath := path.Join(thisPath, "alias.go")
		return CreateAliasFromDirs(packageName, aliasFilePath, importDirs, fullDirs)
	},
}

// CreateAliasFromDirs
func CreateAliasFromDirs(packageName, aliasFilePath string, importDirs, fullDirs []string) error {
	var pas []PackageAlias
	for i := 0; i < len(importDirs); i++ {
		importDir := importDirs[i]
		fullDir := fullDirs[i]

		files, err := ioutil.ReadDir(fullDir)
		if err != nil {
			return err
		}

		if len(files) == 0 {
			return errors.New("no files")
		}

		pa, err := CreatePackageAlias(importDir, fullDir, files)
		if err != nil {
			return err
		}
		pas = append(pas, pa)
	}

	out := CompileOutput(packageName, pas)
	ioutil.WriteFile(aliasFilePath, []byte(out), 0644)

	command := fmt.Sprintf("go fmt %v", aliasFilePath)
	_, err := common.Execute(command)
	return err
}

// FileAliases - file alaises
type PackageAlias struct {
	Dir        string
	Name       string
	FuncNames  []string
	VarNames   []string
	ConstNames []string
	TypeNames  []string
}

// NewfileAliases creates a new fileAliases object
func CreatePackageAlias(importDir, fullDir string, files []os.FileInfo) (PackageAlias, error) {

	fa := PackageAlias{
		Dir:  importDir,
		Name: path.Base(importDir),
	}
	//localDir := path.Base(importDir) + "/"

	for _, f := range files {
		fname := f.Name()
		if path.Ext(fname) != ".go" || strings.HasSuffix(fname, "_test.go") {
			continue
		}
		filePath := path.Join(fullDir, fname)
		lines, err := common.ReadLines(filePath)
		if err != nil {
			continue
		}

		// top level escape
		if len(lines) > 0 && strings.Contains(lines[0], NoAliasEsc) {
			continue
		}

		// get the exported functions
		escNext := false
		for _, line := range lines {
			if escNext {
				escNext = false
				continue
			}
			if commentLineHasEsc(line) {
				escNext = true
				continue
			}
			if strings.Contains(line, NoAliasEsc) {
				continue
			}
			sep := strings.Fields(line)
			if len(sep) > 1 && sep[0] == "func" {
				if sep[1] == "init()" {
					continue
				}
				if firstCharIsUpperLetter(sep[1]) {
					fa.FuncNames = append(fa.FuncNames, strings.Split(sep[1], "(")[0])
				}
			}
		}

		// get the exported constants
		fa.ConstNames = append(fa.ConstNames, getDefinitionBlocks(lines, "const")...)

		// get the exported variables
		fa.VarNames = append(fa.VarNames, getDefinitionBlocks(lines, "var")...)

		// get the exported types defined in blocks
		fa.TypeNames = append(fa.TypeNames, getDefinitionBlocks(lines, "type")...)
	}

	return fa, nil
}

// indicatorWord is either: "var", "const", or "type"
func getDefinitionBlocks(lines []string, indicatorWord string) (names []string) {

	withinBlock := false // within a "var (" or "const (" or "type (" block
	withinItoa := false  // within an itoa section

	escNext := false
	for _, line := range lines {
		switch {
		case escNext:
			escNext = false
			continue
		case commentLineHasEsc(line):
			escNext = true
			continue
		case strings.Contains(line, NoAliasEsc):
			continue
		case strings.HasPrefix(line, indicatorWord+" (") && withinBlock == false:
			withinBlock = true
			continue
		case strings.HasPrefix(line, ")") && withinBlock == true:
			withinBlock = false
			withinItoa = false
			continue

		case withinBlock:
			sep := strings.Fields(line)
			switch {
			case strings.HasPrefix(line, "\t\t"): // more then two tabs means is a part of another structure
				continue
			case len(sep) == 0:
				withinItoa = false
			case len(sep) < 2 && !withinItoa:
				continue
			case !firstCharIsUpperLetter(sep[0]):
				continue
			default:
				eqSplit := strings.Split(line, "=")
				if len(eqSplit) >= 2 {
					if strings.TrimSpace(eqSplit[1]) == "iota" {
						withinItoa = true
					}
				}
				lcns := strings.Split(eqSplit[0], ",")
				for _, lcn := range lcns {
					lcnT := strings.TrimSpace(lcn)
					if fields := strings.Fields(lcn); len(fields) == 0 {
						continue
					}
					lcnT = strings.Fields(lcnT)[0]
					if firstCharIsUpperLetter(lcnT) {
						names = append(names, lcnT)
					}
				}
			}

		case strings.HasPrefix(line, indicatorWord) && withinBlock == false:
			sep := strings.Fields(line)
			switch {
			case strings.HasPrefix(line, "\t\t"): // more then two tabs means is a part of another structure
				continue
			case len(sep) < 2:
				continue
			case !firstCharIsUpperLetter(sep[1]):
				continue
			default:
				afterIndicator := strings.TrimPrefix(line, indicatorWord+" ")
				eqSplit := strings.Split(afterIndicator, "=")
				lcns := strings.Split(eqSplit[0], ",")
				for _, lcn := range lcns {
					lcnT := strings.TrimSpace(lcn)
					if fields := strings.Fields(lcnT); len(fields) == 0 {
						continue
					}
					lcnT = strings.Fields(lcnT)[0]
					if firstCharIsUpperLetter(lcnT) {
						names = append(names, lcnT)
					}
				}
			}
		}
	}
	return names
}

func firstCharIsUpperLetter(s string) bool {
	return firstCharIsLetter(s) && string(s[0]) == strings.ToUpper(string(s[0]))
}

func firstCharIsLetter(s string) bool {
	return unicode.IsLetter(rune(s[0]))
}

func commentLineHasEsc(line string) bool {
	sep := strings.Fields(line)
	if len(sep) > 2 && sep[0] == "//" && strings.Contains(line, NoAliasEsc) {
		return true
	}
	return false
}

// compile package aliases into output string
func CompileOutput(packageName string, pas []PackageAlias) string {

	out := fmt.Sprintf("// nolint")
	out += fmt.Sprintf("\n// autogenerated code using github.com/rigelrozanski/multitool")
	out += fmt.Sprintf("\n// aliases generated for the following subdirectories:")
	for _, alias := range pas {
		out += fmt.Sprintf("\n// %v %v", AliasKeyword, alias.Dir)
	}
	out += fmt.Sprintf("\npackage %s", packageName)

	out += fmt.Sprintf("\n\nimport (")
	for _, alias := range pas {
		out += fmt.Sprintf("\n\t\"%v\"", alias.Dir)
	}
	out += fmt.Sprintf("\n)")

	var constLines, funcLines, varLines, typeLines []string
	for _, pa := range pas {
		for _, constName := range pa.ConstNames {
			constLines = append(constLines, fmt.Sprintf("\n\t%v = %v.%v", constName, pa.Name, constName))
		}
		for _, funcName := range pa.FuncNames {
			funcLines = append(funcLines, fmt.Sprintf("\n\t%v = %v.%v", funcName, pa.Name, funcName))
		}
		for _, varName := range pa.VarNames {
			varLines = append(varLines, fmt.Sprintf("\n\t%v = %v.%v", varName, pa.Name, varName))
		}
		for _, typeName := range pa.TypeNames {
			typeLines = append(typeLines, fmt.Sprintf("\n\t%v = %v.%v", typeName, pa.Name, typeName))
		}
	}

	if len(constLines) > 0 {
		out += fmt.Sprintf("\n\nconst (")
		for _, constLine := range constLines {
			out += constLine
		}
		out += fmt.Sprintf("\n)")
	}

	if len(funcLines)+len(varLines) > 0 {
		out += fmt.Sprintf("\n\nvar (")
		if len(funcLines) > 0 {
			out += fmt.Sprintf("\n\t// functions aliases")
			for _, funcLine := range funcLines {
				out += funcLine
			}
		}

		if len(varLines) > 0 {
			out += fmt.Sprintf("\n\n\t// variable aliases")
			for _, varLine := range varLines {
				out += varLine
			}
		}
		out += fmt.Sprintf("\n)")
	}

	if len(typeLines) > 0 {
		out += fmt.Sprintf("\n\ntype (")
		for _, typeLine := range typeLines {
			out += typeLine
		}
		out += fmt.Sprintf("\n)")
	}
	return out
}

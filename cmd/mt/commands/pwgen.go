package commands

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	mathrand "math/rand"
	"strconv"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

// Lock2yamlCmd represents the lock2yaml command
var (
	PWGenCmd = &cobra.Command{
		Use:   "pw [length] [chset]",
		Short: "pw generator; chset can be \"simple\", or \"adv\"",
		RunE:  pwGenCmd,
	}
)

func init() {
	RootCmd.AddCommand(PWGenCmd)
}

func pwGenCmd(cmd *cobra.Command, args []string) error {

	chsetSimple := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y",
		"z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N",
		"O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "0", "1", "2",
		"3", "4", "5", "6", "7", "8", "9"}

	chsetAdv := []string{`!`, `@`, `#`, `$`, `%`, `^`, `&`, `*`, `(`, `)`, `-`,
		`_`, `+`, `=`, `[`, `]`, `{`, `}`, `<`, `>`, `?`, `,`, `.`, `~`}
	chsetAdv = append(chsetAdv, chsetSimple...)

	chset := chsetAdv
	if len(args) == 2 && args[1] == "simple" {
		chset = chsetSimple
	}

	if len(args) < 1 {
		return errors.New("must provide a length argument")
	}
	pwLen, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("bad pw length, error: %v", err)
	}

	var bz [8]byte
	_, err = cryptorand.Read(bz[:])
	if err != nil {
		return fmt.Errorf("error getting random seed: %v", err)
	}
	seed := int64(binary.LittleEndian.Uint64(bz[:]))
	mathrand.Seed(seed)

	// generate the pw
	maxN := len(chset) - 1
	out := ""
	for i := 0; i < pwLen; i++ {
		out += chset[mathrand.Intn(maxN)]
	}
	clipboard.WriteAll(out)
	fmt.Println("pw added to clipboard")
	return nil
}

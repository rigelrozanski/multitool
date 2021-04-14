package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/knetic/govaluate"
	"github.com/spf13/cobra"
)

// file commands
var (
	ConvertCmd = &cobra.Command{
		Use:   "convert [amount] [unit] to [unit] <of> <material>",
		Short: "convert arbitrary values between units",
		Args:  cobra.RangeArgs(4, 6),
		RunE:  convertCmd,
	}
)

var decimalPlacesFromFlag int

func init() {
	ConvertCmd.PersistentFlags().IntVarP(&decimalPlacesFromFlag, "decimals", "d", -1,
		"number of decimal places (-1 for auto)")
	RootCmd.AddCommand(ConvertCmd)
}

var (
	unitAlias = map[string]string{ // map[aliasUnit]unit
		"Gal":           "gal",
		"inch":          "in",
		"sqft":          "ft^2",
		"ft2":           "ft^2",
		"sqm":           "m^2",
		"m2":            "m^2",
		"Acre":          "acre",
		"ac":            "acre",
		"hec":           "hectare",
		"tsp":           "teaspoon",
		"tbls":          "tablespoon",
		"tbsp":          "tablespoon",
		"lemons":        "lemon",
		"baking-soda":   "bakingsoda",
		"baking-powder": "bakingpowder",
	}

	// TODO add with reverse function for standard multiplier conversions
	cvs = map[string]string{ // map[from_to_of]expr
		"C_F":                     "(a*1.8)+32",
		"F_C":                     "(a-32)/1.8",
		"g_ml_water":              "a",
		"g_L_water":               "a/1000",
		"acre_hectare":            "a*0.404686",
		"hectare_acre":            "a/0.404686",
		"acre_ft^2":               "a*43560.04",
		"ft^2_acre":               "a/43560.04",
		"hectare_sqft":            "a*107639.1",
		"ft^2_hectare":            "a/107639.1",
		"m_ft":                    "a*3.28084",
		"ft_m":                    "a/3.28084",
		"ft_in":                   "a*12",
		"in_ft":                   "a/12",
		"in_mm":                   "a*25.4",
		"mm_in":                   "a/25.4",
		"m^2_ft^2":                "a*10.76390999",
		"ft^2_m^2":                "a/10.76390999",
		"ft^2_L":                  "a*28.31685",
		"L_ft^2":                  "a/28.31685",
		"pint_cup":                "a*2",
		"cup_pint":                "a/2",
		"quart_cup":               "a*4",
		"cup_quart":               "a/4",
		"cup_L":                   "a*0.236587524",
		"L_cup":                   "a/0.236587524",
		"gal_L":                   "a*4.54609",
		"L_gal":                   "a/4.54609",
		"tablespoon_teaspoon":     "a*3",
		"teaspoon_tablespoon":     "a/3",
		"tablespoon_cup":          "a*0.0625",
		"cup_tablespoon":          "a/0.0625",
		"teaspoon_cup":            "a*0.02083333156038129",
		"cup_teaspoon":            "a/0.02083333156038129",
		"kg_pound":                "a*2.204623",
		"pound_kg":                "a/2.204623",
		"lemon_tablespoon":        "RANGE a*4 a*5",
		"tablespoon_lemon":        "RANGE a/4 a/5",
		"lemon_cup":               "RANGE a*1/4 a*1/3",
		"cup_lemon":               "RANGE a/(1/4) a/(1/3)",
		"cup_egg":                 "a*5",
		"egg_cup":                 "a/5",
		"bakingsoda_bakingpowder": "a*4",
		"bakingpowder_bakingsoda": "a/4",
	}
)

func convertCmd(cmd *cobra.Command, args []string) error {

	amountStr, unitFrom, toArg, unitTo := args[0], args[1], args[2], args[3]
	if toArg != "to" {
		return errors.New("the word \"to\" not in the correct place (3rd arg)")
	}
	if len(args) == 5 {
		return errors.New("invalid number of args")
	}
	material := ""
	if len(args) == 6 {
		material = args[5]
	}

	decimalPlaces := 2
	splt := strings.Split(amountStr, ".")
	if len(splt) == 2 {
		decimalPlaces = len(splt[1])
	}

	if decimalPlacesFromFlag != -1 {
		decimalPlaces = decimalPlacesFromFlag
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return err
	}

	uf, ok := unitAlias[unitFrom]
	if ok {
		unitFrom = uf
	}

	ut, ok := unitAlias[unitTo]
	if ok {
		unitTo = ut
	}

	lookup := unitFrom + "_" + unitTo
	if material != "" {
		lookup += "_" + material
	}
	convExpr, found := cvs[lookup]
	if !found {
		return errors.New("unknown conversion")
	}

	if strings.HasPrefix(convExpr, "RANGE") {
		splt := strings.Fields(convExpr)
		if len(splt) != 3 {
			panic(fmt.Sprintf("invalid range expression input: %v", convExpr))
		}
		convExpr1 := splt[1]
		convExpr2 := splt[2]

		expr1, err := govaluate.NewEvaluableExpression(convExpr1)
		if err != nil {
			return err
		}
		expr2, err := govaluate.NewEvaluableExpression(convExpr2)
		if err != nil {
			return err
		}
		param := map[string]interface{}{"a": amount}
		convertedAmt1, err := expr1.Evaluate(param)
		if err != nil {
			return err
		}
		convertedAmt2, err := expr2.Evaluate(param)
		if err != nil {
			return err
		}

		f1 := strconv.FormatFloat(convertedAmt1.(float64), 'f', decimalPlaces, 64)
		f2 := strconv.FormatFloat(convertedAmt2.(float64), 'f', decimalPlaces, 64)
		if f1 < f2 {
			fmt.Printf("between %v - %v %v\n", f1, f2, unitTo)
		} else {
			fmt.Printf("between %v - %v %v\n", f2, f1, unitTo)
		}
	} else {
		expr, err := govaluate.NewEvaluableExpression(convExpr)
		if err != nil {
			return err
		}

		param := map[string]interface{}{"a": amount}
		convertedAmt, err := expr.Evaluate(param)
		if err != nil {
			return err
		}

		f := strconv.FormatFloat(convertedAmt.(float64), 'f', decimalPlaces, 64)
		fmt.Printf("%v %v\n", f, unitTo)
	}
	return nil
}

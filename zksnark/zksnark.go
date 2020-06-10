package zksnark

import (
	"fmt"
	"github.com/arnaucube/go-snark"
	"github.com/arnaucube/go-snark/circuitcompiler"
	"math/big"
	"strings"
)

func Run() {
	flatCode := `
func exp3(private a):
	b = a * a
	c = a * b
	return c

func main(private s0, public s1):
    
	s3 = exp3(s0)
	s4 = s3 + s0
	s5 = s4 + 5
	equals(s1, s5)
`

	// parse the code
	parser := circuitcompiler.NewParser(strings.NewReader(flatCode))
	circuit, _ := parser.Parse()
	fmt.Println(circuit)

	b3 := big.NewInt(int64(3))
	privateInputs := []*big.Int{b3}
	b35 := big.NewInt(int64(35))
	publicSignals := []*big.Int{b35}

	// witness
	w, _ := circuit.CalculateWitness(privateInputs, publicSignals)
	fmt.Println("witness", w)

	// flat code to R1CS
	fmt.Println("generating R1CS from flat code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Printf("a:%v\n", a)
	fmt.Printf("b:%v\n", b)
	fmt.Printf("c:%v\n", c)

	alphas, betas, gammas, _ := snark.Utils.PF.R1CSToQAP(a, b, c)

	ax, bx, cx, px := snark.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	fmt.Printf("ax:%v\n", ax)
	fmt.Printf("bx:%v\n", bx)
	fmt.Printf("cx:%v\n", cx)
	//setup
	setup, _ := snark.GenerateTrustedSetup(len(w), *circuit, alphas, betas, gammas)

	hx := snark.Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	fmt.Printf("hx:%v\n", hx)

	proof, _ := snark.GenerateProofs(*circuit, setup.Pk, w, px)

	b35Verif := big.NewInt(int64(35))
	publicSignalsVerif := []*big.Int{b35Verif}
	fmt.Println(snark.VerifyProof(setup.Vk, proof, publicSignalsVerif, true))
}

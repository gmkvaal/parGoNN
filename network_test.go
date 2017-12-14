package main

import (
	"testing"

	//plr "github.com/gmkvaal/pythonlistreader"
	"fmt"
)

/*
// TestForwardFeed tests the forward feed algorithm
// by initiating with zero-weights and biases, hence making
// all z's zero and thus all activations 1/2 (given sigmoids)
func TestForwardFeed(t *testing.T) {
	nf = initNetworkForTesting()
	nf.weights = nf.cubicMatrix(zeroFunc())
	nf.biases = nf.squareMatrix(zeroFunc())
	l := len(nf.sizes) - 1

	x := make([]float64, 784, 784)
	y := []float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5}

	assert.Equal(t, nf.forwardFeed(x, l), y)
}

func TestOutputError(t *testing.T) {
	nf = initNetworkForTesting()
	l := len(nf.sizes) - 1

	y := []float64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	yNeg := []float64{-1, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	nf.outputError(y, l)
	assert.Equal(t, nf.delta[l-1], yNeg)
}

func TestBackPropShort(t *testing.T) {
	nf = initNetworkForTesting()
	l := len(nf.sizes) - 1

	nablaW := nf.cubicMatrix(zeroFunc())
	nablaB := nf.squareMatrix(zeroFunc())
	nf.delta[l-1] = []float64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	nablaB[l-1] = nf.delta[l-1]
	nablaW[l-1] = vectorMatrixProduct(nablaW[l-1], nf.delta[l-1], nf.activations[l-1])
	nablaW, nablaB = nf.backPropError(nablaW, nablaB, l)
	nablaB[l-1] = nf.delta[l-1]
	nablaW[l-1] = vectorMatrixProduct(nablaW[l-1], nf.delta[l-1], nf.activations[l-1])
	nablaW, nablaB = nf.backPropError(nablaW, nablaB, l)
}


func TestOutputGradients(t *testing.T) {
	nf = initNetworkForTesting()
	l := len(nf.sizes) - 1

	nablaW := nf.cubicMatrix(zeroFunc())
	nablaB := nf.squareMatrix(zeroFunc())

	nf.outputGradients(nablaW, nablaB, l)
}



// TestBackProp tests the back propagation algorithm
// by running through two mini batches of length two,
// and then comparing the weights and biases with the
// validated solver https://github.com/mnielsen/neural-networks-and-deep-learning/tree/master/src/network2
// Weights and biases are for the purpose of creating
// a deterministic result initiated as 1's.
func TestBackProp(t *testing.T) {
	nf := networkFormat{sizes: []int{784, 30, 10}}

	nf.weights = nf.cubicMatrix(oneFunc())
	nf.biases = nf.squareMatrix(oneFunc())
	nf.delta = nf.squareMatrix(zeroFunc())
	nf.z = nf.squareMatrix(zeroFunc())
	nf.activations = nf.squareMatrixFull(zeroFunc())

	nf.miniBatchSize = 2
	nf.n = 4
	nf.hp.eta = 1
	nf.hp.lambda = 5.0

	x1 := make([]float64, 784, 784)
	y1 := make([]float64, 10, 10)
	y1[0] = 1
	x1[1] = 1
	b1 := [][]float64{x1, y1}

	x2 := make([]float64, 784, 784)
	y2 := make([]float64, 10, 10)
	y2[1] = 1
	x2[2] = 1
	b2 := [][]float64{x2, y2}

	x3 := make([]float64, 784, 784)
	y3 := make([]float64, 10, 10)
	y3[2] = 1
	x3[3] = 1
	b3 := [][]float64{x3, y3}

	x4 := make([]float64, 784, 784)
	y4 := make([]float64, 10, 10)
	y4[3] = 1
	x4[4] = 1
	b4 := [][]float64{x4, y4}

	miniBatchA := [][][]float64{b1, b2}
	miniBatchB := [][][]float64{b3, b4}
	miniBatches := [][][][]float64{miniBatchA, miniBatchB}
	nf.data.miniBatches = miniBatches
	nf.updateMiniBatches()


	testData, err := plr.ReadFile("testdata/bias1.txt")
	if err != nil {
		t.Fatal(err)
	}

	bias1FromPy := plr.PythonFloatListParser(testData)

	testData, err = plr.ReadFile("testdata/bias0.txt")
	if err != nil {
		t.Fatal(err)
	}

	bias0FromPy := plr.PythonFloatListParser(testData)

	testData, err = plr.ReadFile("testdata/weights1.txt")
	if err != nil {
		t.Fatal(err)
	}

	weights1FromPy := plr.PythonNestedFloatListParser(testData)

	testData, err = plr.ReadFile("testdata/weights0.txt")
	if err != nil {
		t.Fatal(err)
	}

	weights0FromPy := plr.PythonNestedFloatListParser(testData)

	for idx := range bias1FromPy {
		if bias1FromPy[idx] - nf.biases[1][idx] > 1e-7 {
			t.Error("not equal", bias1FromPy[idx], nf.biases[1][idx])
		}
	}

	for idx := range bias0FromPy {
		if bias0FromPy[idx] - nf.biases[0][idx] > 1e-7 {
			t.Error("not equal", bias0FromPy[idx], nf.biases[0][idx])
		}
	}

	for idx1 := range weights1FromPy {
		for idx2 := range weights1FromPy[idx1] {
			if weights1FromPy[idx1][idx2] - nf.weights[1][idx1][idx2] > 1e-7 {
				t.Error("not equal", weights1FromPy[idx1][idx2], nf.weights[1][idx1][idx2])
			}
		}
	}

	for idx1 := range weights0FromPy {
		for idx2 := range weights0FromPy[idx1] {
			if weights0FromPy[idx1][idx2] - nf.weights[0][idx1][idx2] > 1e-7 {
				t.Error("not equal", weights0FromPy[idx1][idx2], nf.weights[0][idx1][idx2])
			}
		}
	}
}


*/

func vizNumber(s []float64) {
	for idx := range s {
		if s[idx] > 0 {
			s[idx] = 1
		}
	}
	counter := 0
	matrix := make([][]float64, 28, 28)
	for i := 0; i < 28; i++ {
		matrix[i] = make([]float64, 28, 28)
		for j := 0; j < 28; j++ {
			matrix[i][j] = s[counter]
			counter++
		}
	}

	for i := 0; i < 28; i++ {
		fmt.Println(matrix[i])
	}
}

func TestIt(t *testing.T) {
	fmt.Println()
	nf := networkFormat{sizes: []int{784, 100, 10}}
	nf.initNetwork()

	nf.hp.initHyperParameters(1,0.5)
	nf.data.formatData()

	//x2 := nf.miniBatches[1][0][0]
	//yReal2 := nf.data.miniBatches[1][0][1]


	//fmt.Println(nf.miniBatches)

	fmt.Println()
	//fmt.Println(yReal1, yReal2)


	//for i := 0; i < 10; i++ {
	//	for j := 0; j < 10; j++ {
	//		fmt.Println(nf.data.miniBatches[i][j][1])
	//	}
	//}


	nf.data.miniBatchGenerator(0, 1000, 10, true)
	//x2 := nf.miniBatches[1][0][0]
	//yReal2 := nf.data.miniBatches[1][0][1]



	l := len(nf.sizes) - 1 // last entry "layer-vise"
	for i := 0; i < 10; i++ {

		nf.data.miniBatchGenerator(0, 10000, 10, true)
		x1 := nf.miniBatches[0][0][0]
		yReal1 := nf.data.miniBatches[0][0][1]
		nf.updateMiniBatches()
		y1 := nf.forwardFeed(x1, l)

		fmt.Println(y1, yReal1)

		fmt.Println(checkIfEqual(y1, yReal1))
	}






	//inputData := [][]float64{nf.miniBatches[0][0][0], nf.miniBatches[0][1][0]}
	//outputData := [][]float64{nf.miniBatches[0][0][1], nf.miniBatches[0][1][1]}

	//nf.validate(inputData, outputData, 2)


}

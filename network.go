package main

import (
	"fmt"
	"github.com/petar/GoMNIST"
	"log"
)

// NOTE: Maybe not so smart with all these global variables?
// Perhaps put them in a struct and create an object and
// then use overwriting through pointers???

// Variables to hold the activation(s) for one node,
// nodes on one layer, and all layers, respectively
var zNode float64
var zLayer []float64
var zAllLayers [][]float64

// Variables to hold the sigmoid(s) for one node,
// nodes on a new layer, input layer, and every
// layer, respectively.
var activationNode float64
var activationNewLayer []float64
var activationLayer []float64
var activationAllLayers [][]float64

var delta [][]float64



var sigPrimeLayer []float64


// networkFormat contains the
// fields sizes, biases, and weights
type networkFormat struct {
	sizes   []int
	biases  [][]float64
	weights [][][]float64
	delta [][]float64
	nablaW [][][]float64
	nablaB [][]float64
	z [][]float64
	activations [][]float64

}

// initNetwork initiates the weights
// and biases with random numbers
func (nf *networkFormat) initNetwork() {
	nf.weights = nf.cubicMatrix(randomFunc())
	nf.biases = nf.squareMatrix(randomFunc())
	nf.delta = nf.squareMatrix(zeroFunc())
	nf.nablaW = nf.cubicMatrix(zeroFunc())
	nf.nablaB = nf.squareMatrix(zeroFunc())
	nf.z = nf.squareMatrix(zeroFunc())
	nf.activations = nf.squareMatrixFull(zeroFunc())
}



// squareMatrix creates a square matrix with entries from input function
func (nf networkFormat) squareMatrix(input func(int) float64) [][]float64 {
	b := make([][]float64, len(nf.sizes[1:]))

	for k := range b {
		b[k] = make([]float64, nf.sizes[k+1])

		for j := 0; j < nf.sizes[k+1]; j++ {
			b[k][j] = input(nf.sizes[k+1])
		}
	}

	return b
}

// squareMatrix creates a square matrix with entries from input function
func (nf networkFormat) squareMatrixFull(input func(int) float64) [][]float64 {
	b := make([][]float64, len(nf.sizes))

	for k := range b {
		b[k] = make([]float64, nf.sizes[k])

		for j := 0; j < nf.sizes[k]; j++ {
			b[k][j] = input(nf.sizes[k])
		}
	}

	return b
}

// cubicMatrix creates a square matrix with entries from input function
func (nf networkFormat) cubicMatrix(input func(int) float64) [][][]float64 {
	w := make([][][]float64, len(nf.sizes[1:]))

	for k := range w {
		w[k] = make([][]float64, nf.sizes[k+1])

		for j := 0; j < nf.sizes[k+1]; j++ {
			w[k][j] = make([]float64, nf.sizes[k])

			for i := 0; i < nf.sizes[k]; i++ {
				w[k][j][i] = input(nf.sizes[k])
			}
		}
	}

	return w
}

// backProp performs one iteration of the backpropagation algorithm
// for input x and training output y (one batch in a mini batch)
func (nf networkFormat) backProp(x []float64, y [10]float64) ([][]float64, [][][]float64) {

	// Initiating the gradient matrices
	l := len(nf.sizes) - 1// last entry "layer-vise"

	// Clearing / preparing the slices
	nf.activations[0] = x

	// Updating all neurons with the forwardFeed algorithm
	for k := 0; k < len(nf.biases); k++ {
		for i := range nf.weights[k] {
			nf.z[k][i] = dot(nf.activations[k], nf.weights[k][i]) + nf.biases[k][i]
			nf.activations[k+1][i] = sigmoid(nf.z[k][i])
		}
	}

	// Computing the output error (delta L).
	for i := range nf.activations[l] {
		nf.delta[l-1][i] = outputNeuronError(nf.z[l-1][i], nf.activations[l][i], y[i])
	}

	// Gradients at the output layer
	nf.nablaB[l-1] = nf.delta[l-1]
	nf.nablaW[l-1] = vectorMatrixProduct(nf.nablaW[l-1], nf.delta[l-1], nf.activations[l-1])

	// Backpropagating the error
	for k := 2; k < l + 1; k++ {
		for j := 0; j < nf.sizes[l+1-k]; j++ {
			for i := 0; i < nf.sizes[l+2-k]; i++ {
				nf.delta[l-k][j] += nf.weights[l+1-k][i][j]*nf.delta[l+1-k][i] * sigmoidPrime(nf.z[l+1-k][i])
			}

			nf.nablaB[l-k][j] = nf.delta[l-k][j]

			for i := 0; i < nf.sizes[l-k]; i++ {
				nf.nablaW[l-k][j][i] += nf.delta[l-k][j] * nf.activations[l-k][i]
			}
		}

	}

	return nf.nablaB, nf.nablaW

}




func main() {


	// Load files of type GoMNIST which is actually []byte, where the byte value
	train, _, err := GoMNIST.Load("/home/guttorm/xal/go/src/github.com/petar/GoMNIST/data")
	if err != nil {
		log.Fatal(err)
	}

	sweeper := train.Sweep()
	image, label, _ := sweeper.Next()

	labelArr, err := labelToArray(int(label))
	if err != nil {
		log.Fatal(err)
	}

	y := labelArr
	x := customSliceToFloat64Slice(image)

	fmt.Println("")

	nf := &networkFormat{sizes: []int{784, 30, 10}}
	nf.initNetwork()
	nf.nablaB, nf.nablaW = nf.backProp(x, y)

	//fmt.Println(nablaW)
}


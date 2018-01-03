package network

import (
	"fmt"
	"github.com/gonum/matrix/mat64"
	"log"
	"math"
	"regexp"
	"runtime"
	"sync"
	"time"
)

var wg sync.WaitGroup

// Network contains the
// fields Sizes, biases, and weights
type Network struct {
	Sizes  []int
	layer  layer
	layers []layer
	l      int
	nCores int
	hp     HyperParameters
	data
	NetworkMethods
	dataContainers
}

type layer struct {
	size int
	activationFunction
}

type dataContainers struct {
	weights     []*mat64.Dense
	biases      []*mat64.Dense
	nablaW      [][]*mat64.Dense
	nablaB      [][]*mat64.Dense
	deltaNablaW [][]*mat64.Dense
	deltaNablaB [][]*mat64.Dense
	delta       [][]*mat64.Dense
	z           [][]*mat64.Dense
	activations [][]*mat64.Dense
	sp          [][]*mat64.Dense
}

type activationFunction struct {
	function func(i, j int, v float64) float64
	prime    func(i, j int, v float64) float64
}

type NetworkMethods struct {
	outputErrorFunc  func(delta *mat64.Dense, a, y mat64.Matrix)
	validationMethod func(n *Network, inputData, outputData []*mat64.Dense) bool
}

type HyperParameters struct {
	eta    float64
	lambda float64
}

func (n *Network) AddLayer(layerSize int, activationFunction, activationPrime func(i, j int, v float64) float64) {
	n.layer.size = layerSize
	n.layer.function = activationFunction
	n.layer.prime = activationPrime

	n.layers = append(n.layers, n.layer)
}

func (n *Network) setSizes() {
	for idx := range n.layers {
		n.Sizes = append(n.Sizes, n.layers[idx].size)
	}
}

// initNetwork initiates the weights
// and biases with random numbers
func (n *Network) initDataContainers(nCores int) {
	n.setSizes()
	n.l = len(n.Sizes) - 1
	n.nCores = nCores
	n.weights = sliceWithGonumDense(len(n.Sizes[1:]), n.Sizes[:], n.Sizes[1:], randomFunc())
	n.biases = sliceWithGonumDense(len(n.Sizes[1:]), n.Sizes[1:], 1, randomFunc())
	for idx := 0; idx < n.nCores; idx++ {
		n.nablaW = append(n.nablaW, sliceWithGonumDense(len(n.Sizes[1:]), n.Sizes[:], n.Sizes[1:], zeroFunc()))
		n.nablaB = append(n.nablaB, sliceWithGonumDense(len(n.Sizes[1:]), n.Sizes[1:], 1, zeroFunc()))
		n.deltaNablaW = append(n.deltaNablaW, sliceWithGonumDense(len(n.Sizes[1:]), n.Sizes[:], n.Sizes[1:], zeroFunc()))
		n.deltaNablaB = append(n.deltaNablaB, sliceWithGonumDense(len(n.Sizes[1:]), n.Sizes[1:], 1, zeroFunc()))
		n.delta = append(n.delta, sliceWithGonumDense(len(n.Sizes[1:]), n.Sizes[1:], 1, zeroFunc()))
		n.z = append(n.z, sliceWithGonumDense(len(n.Sizes[1:]), n.Sizes[1:], 1, zeroFunc()))
		n.activations = append(n.activations, sliceWithGonumDense(len(n.Sizes[:]), n.Sizes[:], 1, zeroFunc()))
		n.sp = append(n.sp, sliceWithGonumDense(len(n.Sizes[1:]), n.Sizes[1:], 1, zeroFunc()))
	}
}

func (nm *NetworkMethods) InitNetworkMethods(outputError func(delta *mat64.Dense, a, y mat64.Matrix),
	validationMethod func(n *Network, inputData, outputData []*mat64.Dense) bool) {
	nm.outputErrorFunc = outputError
	nm.validationMethod = validationMethod
}

// setHyperParameters initiates the hyper parameters
func (hp *HyperParameters) InitHyperParameters(eta float64, lambda float64) {
	hp.eta = eta
	hp.lambda = lambda
}

// forwardFeed computes the z-s and activations at every neuron and returns the output layer
func (n *Network) forwardFeed(x mat64.Matrix, proc int) *mat64.Dense {
	n.activations[proc][0].Clone(x)
	for k := range n.Sizes[1:] {
		n.z[proc][k].Mul(n.weights[k].T(), n.activations[proc][k])
		n.z[proc][k].Add(n.z[proc][k], n.biases[k])
		n.activations[proc][k+1].Apply(n.layers[k].activationFunction.function, n.z[proc][k])
	}

	return n.activations[proc][n.l]
}

// outputError computes the error at the output neurons
func (n *Network) outputError(y mat64.Matrix, proc int) {
	n.outputErrorFunc(n.delta[proc][n.l-1], n.activations[proc][n.l], y)
}

// outputGradients computes the (delta) gradients at the output layer
func (n *Network) outputGradients(proc int) {
	n.deltaNablaB[proc][n.l-1].Clone(n.delta[proc][n.l-1])
	n.deltaNablaW[proc][n.l-1].Mul(n.activations[proc][n.l-1], n.delta[proc][n.l-1].T())
}

// backPropError backpropagates the error and computes the (delta) gradients
// at every layer
func (n *Network) backPropError(proc int) {
	for k := 2; k < n.l+1; k++ {
		n.sp[proc][n.l-k].Apply(n.layers[k].activationFunction.prime, n.z[proc][n.l-k])
		n.delta[proc][n.l-k].Mul(n.weights[n.l+1-k], n.delta[proc][n.l+1-k])
		n.delta[proc][n.l-k].MulElem(n.delta[proc][n.l-k], n.sp[proc][n.l-k])
		n.deltaNablaB[proc][n.l-k].Clone(n.delta[proc][n.l-k])
		n.deltaNablaW[proc][n.l-k].Mul(n.activations[proc][n.l-k], n.delta[proc][n.l-k].T())
	}
}

// backProp performs one iteration of the backpropagation algorithm
// for input x and training output y (one batch in a mini batch)
func (n *Network) BackPropAlgorithm(x, y *mat64.Dense, proc int) {
	defer wg.Done()

	// 1. Forward feed
	n.forwardFeed(x, proc)

	// 2. Computing the output error (delta L).
	n.outputError(y, proc)

	// 3. Gradients at the output layer
	n.outputGradients(proc)

	// 4. Backpropagating the error
	n.backPropError(proc)

	// 5. Adding delta gradients to gradients
	n.updateGradients(proc)
}

func (n *Network) mergeGradientsAtLayer(k int) {
	for proc := 1; proc < n.nCores; proc++ {
		n.nablaW[0][k].Add(n.nablaW[0][k], n.nablaW[proc][k])
		n.nablaB[0][k].Add(n.nablaB[0][k], n.nablaB[proc][k])
	}
}



// updateGradients adds the delta gradient matrices to the gradient matrices
func (n *Network) updateGradients(proc int) {
	for k := range n.Sizes[1:] {
		n.nablaW[proc][k].Add(n.nablaW[proc][k], n.deltaNablaW[proc][k])
		n.nablaB[proc][k].Add(n.nablaB[proc][k], n.deltaNablaB[proc][k])
	}
}

// updateWeightsAtLayer updates the weights at a given layer of the network
func (n *Network) updateWeightAtLayer(k int) {
	n.weights[k].Scale(1-n.hp.eta*(n.hp.lambda/n.data.n), n.weights[k])
	n.nablaW[0][k].Scale(n.hp.eta/n.data.miniBatchSize, n.nablaW[0][k])
	n.weights[k].Sub(n.weights[k], n.nablaW[0][k])
}

// updateWeightsAtLayer updates the biases at a given layer of the network
func (n *Network) updateBiasesAtLayer(k int) {
	n.nablaB[0][k].Scale(n.hp.eta/n.data.miniBatchSize, n.nablaB[0][k])
	n.biases[k].Sub(n.biases[k], n.nablaB[0][k])
}

// clearGradientsAtLayer sets the weight and bias gradients to zero
func (n *Network) clearGradientsAtLayer(k, proc int) {
	n.nablaW[proc][k].Scale(0, n.nablaW[proc][k])
	n.nablaB[proc][k].Scale(0, n.nablaB[proc][k])
}

// updateWeightsAndBiases updates the weights and biases
// at every layer of the network
func (n *Network) updateWeightsAndBiases() {
	for k := range n.Sizes[1:] {
		n.mergeGradientsAtLayer(k)
		n.updateWeightAtLayer(k)
		n.updateBiasesAtLayer(k)
		for proc := 0; proc < n.nCores; proc++ {
			n.clearGradientsAtLayer(k, proc)
		}
	}
}

// updateMiniBatches runs the stochastic gradient descent
// algorithm for a set of mini batches (e.g one epoch)
func (n *Network) updateMiniBatches() {
	defer TimeTrack(time.Now())

	for i := range n.data.miniBatches {
		for idx, dataSet := range n.data.miniBatches[i] {
			wg.Add(1)
			go n.BackPropAlgorithm(dataSet[0], dataSet[1], int(math.Mod(float64(idx), float64(n.nCores))))
		}

		wg.Wait()
		n.updateWeightsAndBiases()
	}

}

// trainNetwork trains the network with the parameters given as arguments
func (n *Network) TrainNetwork(epochs int, miniBatchSize int, eta, lambda float64, shuffle, validate bool, nCores int) {

	runtime.GOMAXPROCS(nCores)

	if len(n.trainingInput) == 0 || len(n.trainingOutput) == 0 {
		log.Fatal("Insufficient training data submitted")
	}

	if validate {
		if len(n.validationInput) == 0 || len(n.validationOutput) == 0 {
			log.Fatal("Insufficient validation data submitted")
		}
	}

	n.initDataContainers(nCores)
	n.hp.InitHyperParameters(eta, lambda)

	for i := 0; i < epochs; i++ {
		fmt.Println("Epoch", i, ":")

		n.data.miniBatchGenerator(miniBatchSize, shuffle)
		n.updateMiniBatches()

		if validate {
			n.validationMethod(n, n.data.validationInput, n.data.validationOutput)
		}

		//fmt.Println("Avg cost:", nf.totalCost(nf.data.validationInput[:dataCap], nf.data.validationInput[:dataCap]))
		fmt.Println("")
	}
}

func TimeTrack(start time.Time) {
	elapsed := time.Since(start)

	// Skip this function, and fetch the PC and file for its parent.
	pc, _, _, _ := runtime.Caller(1)

	// Retrieve a function object this functions parent.
	funcObj := runtime.FuncForPC(pc)

	// Regex to extract just the function name (and not the module path).
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	log.Println(fmt.Sprintf("%s took %s", name, elapsed))
}

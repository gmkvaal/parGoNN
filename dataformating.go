package network

import (
	"github.com/gonum/matrix/mat64"
	"math/rand"
	"time"
)

type data struct {
	trainingInput    []*mat64.Vector
	trainingOutput   []*mat64.Vector
	validationInput  []*mat64.Vector
	validationOutput []*mat64.Vector
	miniBatches      [][][]*mat64.Vector
	n                float64
	miniBatchSize    float64
}

func (data *data) LoadTrainingData(trainingInput, trainingOutput [][]float64) {
	for idx := range trainingInput {
		data.trainingInput = append(data.trainingInput,
			mat64.NewVector(len(trainingInput[idx]), trainingInput[idx]))
		data.trainingOutput = append(data.trainingOutput,
			mat64.NewVector(len(trainingOutput[idx]), trainingOutput[idx]))
	}
}

func (data *data) LoadValidationData(validationInput, validationOutput [][]float64) {
	for idx := range validationInput {
		data.validationInput = append(data.validationInput,
			mat64.NewVector(len(validationInput[idx]), validationInput[idx]))
		data.validationOutput = append(data.validationOutput,
			mat64.NewVector(len(validationOutput[idx]), validationOutput[idx]))
	}
}

// initSizes initiates the fields containing the size and length of the training set and mini batch
func (data *data) initSizes(trainingSetLength int, miniBatchSize int) {
	data.n = float64(trainingSetLength)
	data.miniBatchSize = float64(miniBatchSize)
}

// shuffleTrainingData shuffles the training data
func (data *data) shuffleTrainingData() {
	for i := len(data.trainingInput) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		data.trainingInput[i], data.trainingInput[j] = data.trainingInput[j], data.trainingInput[i]
		data.trainingOutput[i], data.trainingOutput[j] = data.trainingOutput[j], data.trainingOutput[i]
	}
}

// shuffleValidationData shuffles the validation data
func (data *data) shuffleValidationData() {
	for i := len(data.validationInput) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		data.validationInput[i], data.validationInput[j] = data.validationInput[j], data.validationInput[i]
		data.validationOutput[i], data.validationOutput[j] = data.validationOutput[j], data.validationOutput[i]
	}
}

// shuffleAllData sets a unique seed to initiate unique shuffles
// and calls shuffleTrainingData and suffleValidationData
func (data *data) shuffleAllData() {
	rand.Seed(time.Now().UnixNano())
	data.shuffleTrainingData()
	data.shuffleValidationData()
}

// miniBatchGenerator generates a new set of miniBatches from the training data.
// miniBatches contain (numberOfMiniBatches) number of mini batches, each of which contains (miniBatchSize) number
// of len 2 slices containing the trainingInput and trainingOutput at the respective entries.
func (data *data) miniBatchGenerator(miniBatchSize int, shuffle bool) {

	if shuffle {
		data.shuffleAllData()
	}

	trainingSetLength := len(data.trainingInput)
	numberOfMiniBatches := int(trainingSetLength / miniBatchSize)
	data.miniBatches = make([][][]*mat64.Vector, numberOfMiniBatches, numberOfMiniBatches)
	data.initSizes(trainingSetLength, miniBatchSize)

	for i := 0; i < numberOfMiniBatches; i++ {
		data.miniBatches[i] = make([][]*mat64.Vector, miniBatchSize, miniBatchSize)
		for j := 0; j < miniBatchSize; j++ {
			data.miniBatches[i][j] = []*mat64.Vector{data.trainingInput[i*miniBatchSize+j],
				data.trainingOutput[i*miniBatchSize+j]}
		}
	}
}


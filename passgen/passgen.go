package passgen

import (
  "bitbucket.org/gofd/gofd/core"
  "bitbucket.org/gofd/gofd/propagator"
  "bitbucket.org/gofd/gofd/labeling"
  "crypto/rand"
  "math/big"
  "strings"
)

type Generator struct {
  alphabets map[string]string
  minLength int
  maxLength int
  min map[string]int
  max map[string]int
  Exclude string
}

func New() *Generator {
  return &Generator {
    alphabets: make(map[string]string),
    minLength: 30,
    maxLength: 40,
    min: make(map[string]int),
    max: make(map[string]int),
    Exclude: "",
  }
}

func (this *Generator) SetLength(min, max int) {
  this.minLength = min
  this.maxLength = max
}

func (this *Generator) SetMin(name string, min int) {
  this.min[name] = min
}

func (this *Generator) SetMax(name string, max int) {
  this.max[name] = max
}

func (this *Generator) SetMinMax(name string, min, max int) {
  this.SetMin(name, min)
  this.SetMax(name, max)
}

func (this *Generator) SetRange(name string, min, max int) {
  this.SetMin(name, min)
  this.SetMax(name, max)
}

func (generator *Generator) AddAlphabet(name string, alphabet string) {
  generator.alphabets[name] = alphabet
}

func randInt(low, high int) int {
  num, err := rand.Int(rand.Reader, big.NewInt(int64(high - low + 1)))
  if err != nil {
    panic(err)
  }

  return int(num.Int64()) + low
}

func randBig(low, high *big.Int) *big.Int {
  randRange := high.Add(high, big.NewInt(0))
  randRange = randRange.Sub(randRange, low)
  randRange = randRange.Add(randRange, big.NewInt(1))

  num, err := rand.Int(rand.Reader, randRange)

  if err != nil {
    panic(err)
  }

  num = num.Add(num, low)
  return num
}

func shuffle(source []byte) {
  // Fisher-Yates shuffle.
  for i := len(source) - 1; i >= 1; i-- {
    j := randInt(0, i)
    source[i], source[j] = source[j], source[i]
  }
}

func randBytes(alphabet []byte, count int) []byte {
  bytes := make([]byte, count)

  for i := 0; i < count; i++ {
    index := randInt(0, len(alphabet) - 1)
    bytes[i] = alphabet[index]
  }

  return bytes
}

type choice struct {
  Weight *big.Int
  Item interface{}
}

func weightedRand(choices []*choice) *choice {
  sum := big.NewInt(0)
  for _, c := range choices {
    sum = sum.Add(sum, c.Weight)
  }

  max := sum.Add(sum, big.NewInt(0))
  max = max.Sub(max, big.NewInt(1))
  r := randBig(big.NewInt(0), max)

  for _, c := range choices {
    r = r.Sub(r, c.Weight)
    if r.Cmp(big.NewInt(0)) < 0 {
      return c
    }
  }

  panic("Error getting weighted random value.")
}

func fac(x int) *big.Int {
  i := big.NewInt(1)
  for j := 1; j <= x; j++ {
    i.Mul(i, big.NewInt(int64(j)))
  }

  return i
}

func pow(x int, y int) *big.Int {
  xi := big.NewInt(int64(x))
  yi := big.NewInt(int64(y))
  return xi.Exp(xi, yi, nil)
}

func (this *Generator) resultWeight(result map[core.VarId]int, store *core.Store, length int, alphabets map[string]string) *big.Int {
  // weight = (length! * product(i=1,n | len(alphabet_n)^slots_n)) / (product(i=1,n | slots_n!)

  counts := make(map[string]int)

  for varId, count := range result {
    counts[store.GetName(varId)] = count
  }

  num := fac(length)
  den := big.NewInt(1)

  for name, alphabet := range alphabets {
    count := counts[name]
    opts := pow(len(alphabet), count)
    num = num.Mul(num, opts)
    den = den.Mul(den, fac(count))
  }

  return num.Div(num, den)
}

func exclude(alphabet, exclusions string) string {
  for _, c := range exclusions {
    alphabet = strings.Replace(alphabet, string(c), "", -1)
  }

  return alphabet
}

func (this *Generator) Generate() string {
  // TODO: Handle errors:
  // No alphabets defined
  // All alphabets excluded

  alphabets := make(map[string]string)

  for name, alphabet := range this.alphabets {
    alphabets[name] = exclude(alphabet, this.Exclude)
  }

  chosenLength := randInt(this.minLength, this.maxLength)

  store := core.CreateStore()
  length := core.CreateIntVarFromTo("length", store, chosenLength, chosenLength)

  parts := make([]core.VarId, 0)

  for name, _ := range alphabets {
    var min, max int
    var ok bool

    if min, ok = this.min[name]; !ok {
      min = 0
    }

    if max, ok = this.max[name]; !ok {
      max = chosenLength
    }

    if max == -1 {
      max = chosenLength
    }

    intVar := core.CreateIntVarFromTo(name, store, min, max)
    parts = append(parts, intVar)
  }

  if len(parts) == 1 {
    eq := propagator.CreateXeqC(parts[0], chosenLength)
    store.AddPropagator(eq)
  } else {
    sum := propagator.CreateSum(store, length, parts)
    store.AddPropagator(sum)
  }

  query := labeling.CreateSearchAllQuery()
  solutionFound := labeling.Labeling(store, query, labeling.SmallestDomainFirst, labeling.InDomainMin)

  if !solutionFound {
    // TODO: Support error.
    panic("Solution not found.")
  }

  resultSet := query.GetResultSet()

  choices := make([]*choice, len(resultSet))
  for i, result := range resultSet {
    choices[i] = &choice {
      Weight: this.resultWeight(result, store, chosenLength, alphabets),
      Item: result,
    }
  }

  chosenResult := weightedRand(choices)
  result := chosenResult.Item.(map[core.VarId]int)

  counts := make(map[string]int)

  for varId, count := range result {
    counts[store.GetName(varId)] = count
  }

  bytes := make([]byte, 0)
  for name, count := range counts {
    if alphabet, ok := alphabets[name]; ok {
      bytes = append(bytes, randBytes([]byte(alphabet), count)...)
    }
  }

  shuffle(bytes)

  return string(bytes)
}
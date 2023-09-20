package file

type Option func(*options) error

type options struct {
	wordsReduceProbability        []string
	wordsIncreaseProbability      []string
	passwordsCompletelyIgnore     []string
	usernamesCompletelyIgnore     []string
	probabilityDecreaseMultiplier float64
	probabilityIncreaseMultiplier float64
	entropyThresholdMidpoint      int
	logisticGrowthRate            float64
}

func WithWordsReduceProbability(useDefault bool, names ...string) Option {
	return func(o *options) error {
		if useDefault {
			o.wordsReduceProbability = defaultWordsReduceProbability[:]
		}
		o.wordsReduceProbability = append(o.wordsReduceProbability, names...)
		return nil
	}
}

func WithWordsIncreaseProbability(useDefault bool, names ...string) Option {
	return func(o *options) error {
		if useDefault {
			o.wordsIncreaseProbability = defaultWordsIncreaseProbability[:]
		}
		o.wordsIncreaseProbability = append(o.wordsIncreaseProbability, names...)
		return nil
	}
}

func WithPasswordsCompletelyIgnore(useDefault bool, names ...string) Option {
	return func(o *options) error {
		if useDefault {
			o.passwordsCompletelyIgnore = defaultPasswordsCompletelyIgnore[:]
		}
		o.passwordsCompletelyIgnore = append(o.passwordsCompletelyIgnore, names...)
		return nil
	}
}

func WithUsernamesCompletelyIgnore(useDefault bool, names ...string) Option {
	return func(o *options) error {
		if useDefault {
			o.usernamesCompletelyIgnore = defaultUsernamesCompletelyIgnore[:]
		}
		o.usernamesCompletelyIgnore = append(o.usernamesCompletelyIgnore, names...)
		return nil
	}
}

func WithProbabilityDecreaseMultiplier(probabilityDecreaseMultiplier float64) Option {
	return func(o *options) error {
		o.probabilityDecreaseMultiplier = probabilityDecreaseMultiplier
		return nil
	}
}

func WithProbabilityIncreaseMultiplier(probabilityIncreaseMultiplier float64) Option {
	return func(o *options) error {
		o.probabilityIncreaseMultiplier = probabilityIncreaseMultiplier
		return nil
	}
}

func WithEntropyThresholdMidpoint(entropyThresholdMidpoint int) Option {
	return func(o *options) error {
		o.entropyThresholdMidpoint = entropyThresholdMidpoint
		return nil
	}
}

func WithLogisticGrowthRate(logisticGrowthRate float64) Option {
	return func(o *options) error {
		o.logisticGrowthRate = logisticGrowthRate
		return nil
	}
}

func makeOptions(opts ...Option) (*options, error) {
	o := &options{
		wordsReduceProbability:        defaultWordsReduceProbability[:],
		wordsIncreaseProbability:      defaultWordsIncreaseProbability[:],
		passwordsCompletelyIgnore:     defaultPasswordsCompletelyIgnore[:],
		usernamesCompletelyIgnore:     defaultUsernamesCompletelyIgnore[:],
		probabilityDecreaseMultiplier: defaultProbabilityDecreaseMultiplier,
		probabilityIncreaseMultiplier: defaultProbabilityIncreaseMultiplier,
		entropyThresholdMidpoint:      defaultEntropyThresholdMidpoint,
		logisticGrowthRate:            defaultLogisticGrowthRate,
	}
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

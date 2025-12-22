package kafka

import (
	"github.com/semanggilab/webcore-go/app/config"
	"github.com/semanggilab/webcore-go/app/loader"
)

type KafkaConsumerLoader struct {
	name string
}

func (a *KafkaConsumerLoader) SetClassName(name string) {
	a.name = name
}

func (a *KafkaConsumerLoader) ClassName() string {
	return a.name
}

func (l *KafkaConsumerLoader) Init(args ...any) (loader.Library, error) {
	config := args[0].(config.KafkaConfig)
	receiver := args[1].(KafkaReceiver)

	kc, err := NewKafkaConsumer(&config, receiver)
	if err != nil {
		return nil, err
	}

	err = kc.Install(args...)
	if err != nil {
		return nil, err
	}

	kc.Connect()
	return kc, nil
}

type KafkaProducerLoader struct {
	name string
}

func (a *KafkaProducerLoader) SetClassName(name string) {
	a.name = name
}

func (a *KafkaProducerLoader) ClassName() string {
	return a.name
}

func (l *KafkaProducerLoader) Init(args ...any) (loader.Library, error) {
	config := args[0].(config.KafkaConfig)

	kc, err := NewKafkaProducer(&config)
	if err != nil {
		return nil, err
	}

	err = kc.Install(args...)
	if err != nil {
		return nil, err
	}

	kc.Connect()
	return kc, nil
}

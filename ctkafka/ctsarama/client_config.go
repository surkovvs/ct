package ctsarama

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/IBM/sarama"
	"github.com/surkovvs/ct/ctifaces"
	"github.com/surkovvs/ct/internal/tools"
)

func newClientConfigDefault(cfg ctifaces.KafkaClientConfigurator) (*sarama.Config, error) {
	sConfig := sarama.NewConfig()

	sConfig.Version = sarama.V2_6_0_0
	sConfig.ClientID = cfg.GetCLientID()
	sConfig.Net.MaxOpenRequests = 1
	sConfig.Metadata.Full = false // Cache only necessary metadata.
	tools.SetIfNotNil(&sConfig.ChannelBufferSize, cfg.GetChannelBufferSize())

	if cfg.AuthNeeded() {
		sConfig.Net.SASL.Enable = true
		sConfig.Net.SASL.Handshake = true
		sConfig.Net.SASL.User = cfg.GetUserName()
		sConfig.Net.SASL.Password = cfg.GetPassword()
		sConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
	}

	if cfg.GetCertPath() != nil {
		certs := x509.NewCertPool()
		pemData, err := os.ReadFile(*cfg.GetCertPath())
		if err != nil {
			return nil, fmt.Errorf("cannot read cert path for kafka: %w", err)
		}
		certs.AppendCertsFromPEM(pemData)

		sConfig.Net.TLS.Enable = true
		sConfig.Net.TLS.Config = &tls.Config{
			RootCAs: certs,
		}
	}

	return sConfig, nil
}

type CfgOpt func(*sarama.Config)

func WithSCRAMClient(genFunc func() sarama.SCRAMClient) CfgOpt {
	return func(c *sarama.Config) {
		c.Net.SASL.SCRAMClientGeneratorFunc = genFunc
	}
}

func WithPartitioner(partitioner sarama.PartitionerConstructor) CfgOpt {
	return func(c *sarama.Config) {
		c.Producer.Partitioner = partitioner
	}
}

func WithProducerInterceptor(itc sarama.ProducerInterceptor) CfgOpt {
	return func(c *sarama.Config) {
		c.Producer.Interceptors = append(c.Producer.Interceptors, itc)
	}
}

func WithConsumerInterceptor(itc sarama.ConsumerInterceptor) CfgOpt {
	return func(c *sarama.Config) {
		c.Consumer.Interceptors = append(c.Consumer.Interceptors, itc)
	}
}

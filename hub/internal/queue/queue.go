// Package queue encapsula SNS+SQS Standard (ARQ-03), usados para o
// transporte ASYNC/AUTO e para a distribuicao de fatos (COM-01, COM-03).
// Na referencia local/dev, aponta para um LocalStack; em prd apontaria
// para SNS/SQS reais da conta/regiao aprovada em P-01 (nao resolvido
// nesta implementacao — placeholder aceito).
//
// O caminho SYNC nunca usa este pacote no percurso obrigatorio
// (COM-01/COM-06): SYNC e HTTPS direto entre Orbita e Cometa.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Client agrupa os dois clientes AWS necessarios ao dominio de
// mensageria do hub.
type Client struct {
	SQS *sqs.Client
	SNS *sns.Client
}

// Envelope e o envelope obrigatorio de mensagem (COM-03): event_id,
// tipo, versao do schema, produtor, tenant_id, protocol_id, datas,
// causation_id e versoes de configuracao. step_id/operation_id/
// attempt_id sao incluidos quando pertinentes.
type Envelope struct {
	EventID           string          `json:"event_id"`
	Type              string          `json:"type"`
	SchemaVersion     int             `json:"schema_version"`
	Producer          string          `json:"producer"`
	TenantID          string          `json:"tenant_id"`
	ProtocolID        string          `json:"protocol_id"`
	StepID            string          `json:"step_id,omitempty"`
	OperationID       string          `json:"operation_id,omitempty"`
	AttemptID         string          `json:"attempt_id,omitempty"`
	OccurredAt        time.Time       `json:"occurred_at"`
	RecordedAt        time.Time       `json:"recorded_at"`
	CausationID       string          `json:"causation_id,omitempty"`
	AggregateVersion  int             `json:"aggregate_version"`
	ConfigVersions     map[string]string `json:"config_versions,omitempty"`
	Payload           json.RawMessage `json:"payload"`
}

// New conecta ao endpoint informado (LocalStack em local/dev) usando
// credenciais estaticas de desenvolvimento — nunca usar em prd, onde a
// identidade de workload (SEG-01) resolveria as credenciais reais.
func New(ctx context.Context, endpoint, region string) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("local", "local", "")),
	)
	if err != nil {
		return nil, fmt.Errorf("queue: carregar config aws: %w", err)
	}

	resolver := func(o *sqs.Options) { o.BaseEndpoint = aws.String(endpoint) }
	snsResolver := func(o *sns.Options) { o.BaseEndpoint = aws.String(endpoint) }

	return &Client{
		SQS: sqs.NewFromConfig(cfg, resolver),
		SNS: sns.NewFromConfig(cfg, snsResolver),
	}, nil
}

// EnsureQueue cria a fila se ainda nao existir (idempotente o
// suficiente para bootstrap local/dev) e retorna sua URL.
func (c *Client) EnsureQueue(ctx context.Context, name string) (string, error) {
	out, err := c.SQS.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String(name)})
	if err != nil {
		return "", fmt.Errorf("queue: criar fila %s: %w", name, err)
	}
	return aws.ToString(out.QueueUrl), nil
}

// QueueARN retorna o ARN de uma fila a partir de sua URL — necessario
// para assinar a fila a um topico SNS.
func (c *Client) QueueARN(ctx context.Context, queueURL string) (string, error) {
	out, err := c.SQS.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameQueueArn},
	})
	if err != nil {
		return "", fmt.Errorf("queue: obter ARN da fila: %w", err)
	}
	return out.Attributes[string(sqstypes.QueueAttributeNameQueueArn)], nil
}

// EnsureTopic cria o topico SNS se ainda nao existir e retorna seu ARN.
func (c *Client) EnsureTopic(ctx context.Context, name string) (string, error) {
	out, err := c.SNS.CreateTopic(ctx, &sns.CreateTopicInput{Name: aws.String(name)})
	if err != nil {
		return "", fmt.Errorf("queue: criar topico %s: %w", name, err)
	}
	return aws.ToString(out.TopicArn), nil
}

// Subscribe assina uma fila SQS a um topico SNS (fan-out), retornando
// o ARN da subscricao.
func (c *Client) Subscribe(ctx context.Context, topicARN, queueARN string) error {
	_, err := c.SNS.Subscribe(ctx, &sns.SubscribeInput{
		TopicArn: aws.String(topicARN),
		Protocol: aws.String("sqs"),
		Endpoint: aws.String(queueARN),
	})
	if err != nil {
		return fmt.Errorf("queue: assinar fila no topico: %w", err)
	}
	return nil
}

// AllowSNSDelivery autoriza qualquer topico SNS a entregar mensagens
// nesta fila. So aceitavel na referencia local/dev (LocalStack); em
// prd a policy da fila restringiria a origem ao(s) ARN(s) de topico
// especifico(s) — placeholder de seguranca, ver auditoria.
func (c *Client) AllowSNSDelivery(ctx context.Context, queueURL, queueARN string) error {
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": "*",
			"Action": "sqs:SendMessage",
			"Resource": %q
		}]
	}`, queueARN)
	_, err := c.SQS.SetQueueAttributes(ctx, &sqs.SetQueueAttributesInput{
		QueueUrl:   aws.String(queueURL),
		Attributes: map[string]string{string(sqstypes.QueueAttributeNamePolicy): policy},
	})
	if err != nil {
		return fmt.Errorf("queue: autorizar entrega SNS na fila: %w", err)
	}
	return nil
}

// SendCommand publica um comando durável na fila (ASYNC/AUTO), nunca
// usado para o comando DIRECT de SYNC (EXE-15: dispatch_mode QUEUED
// apenas).
func (c *Client) SendCommand(ctx context.Context, queueURL string, env Envelope) error {
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("queue: serializar envelope: %w", err)
	}
	_, err = c.SQS.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return fmt.Errorf("queue: enviar comando: %w", err)
	}
	return nil
}

// PublishFact publica um fato de dominio (evento) no topico SNS, para
// fan-out a Orbita/Pulsar/Libra conforme COM-01.
func (c *Client) PublishFact(ctx context.Context, topicARN string, env Envelope) error {
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("queue: serializar envelope: %w", err)
	}
	_, err = c.SNS.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(topicARN),
		Message:  aws.String(string(body)),
	})
	if err != nil {
		return fmt.Errorf("queue: publicar fato: %w", err)
	}
	return nil
}

// ReceivedMessage e uma mensagem consumida de uma fila SQS, pronta
// para ser confirmada (ack) somente apos o commit do efeito/intencao
// local (COM-03: inbox antes do ack).
type ReceivedMessage struct {
	ReceiptHandle string
	Envelope      Envelope
	SNSWrapped    bool
}

// snsNotification e o envelope que o SQS recebe quando a fila esta
// assinada a um topico SNS (fan-out).
type snsNotification struct {
	Message string `json:"Message"`
}

// Receive consome ate max mensagens da fila, com long polling
// (waitSeconds), decodificando o Envelope de dominio (diretamente ou
// desembrulhado de uma notificacao SNS).
func (c *Client) Receive(ctx context.Context, queueURL string, max int32, waitSeconds int32) ([]ReceivedMessage, error) {
	out, err := c.SQS.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: max,
		WaitTimeSeconds:     waitSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("queue: receber mensagens: %w", err)
	}
	result := make([]ReceivedMessage, 0, len(out.Messages))
	for _, m := range out.Messages {
		body := aws.ToString(m.Body)
		var env Envelope
		wrapped := false
		if err := json.Unmarshal([]byte(body), &env); err != nil || env.EventID == "" {
			var sn snsNotification
			if err2 := json.Unmarshal([]byte(body), &sn); err2 == nil && sn.Message != "" {
				if err3 := json.Unmarshal([]byte(sn.Message), &env); err3 == nil {
					wrapped = true
				}
			}
		}
		result = append(result, ReceivedMessage{
			ReceiptHandle: aws.ToString(m.ReceiptHandle),
			Envelope:      env,
			SNSWrapped:    wrapped,
		})
	}
	return result, nil
}

// Delete confirma (ack) o processamento de uma mensagem — deve ser
// chamado somente depois do commit local do efeito/intencao (COM-03).
func (c *Client) Delete(ctx context.Context, queueURL, receiptHandle string) error {
	_, err := c.SQS.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		return fmt.Errorf("queue: confirmar mensagem: %w", err)
	}
	return nil
}

#!/bin/bash
echo "Criando fila SQS no LocalStack..."
awslocal sqs create-queue --queue-name user-events-queue
echo "Fila criada com sucesso!"
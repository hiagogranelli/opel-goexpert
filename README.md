# Opel GoExpert — Observabilidade com OpenTelemetry, Zipkin e Docker Compose

Este projeto contém dois serviços em Go instrumentados com OpenTelemetry, um coletor (OTel Collector) e o Zipkin para visualização de traces.

- `service-a`: expõe uma API HTTP pública na porta 8080. Recebe um CEP e consulta o `service-b`.
- `service-b`: consulta o ViaCEP (para obter a localidade) e o WeatherAPI (para obter a temperatura). Não é exposto externamente; é acessível apenas pela rede do Docker.
- `otel-collector`: recebe os traces via OTLP e exporta para o Zipkin.
- `zipkin`: interface web para visualizar traces em http://localhost:9411.

Estrutura (resumo):
- `config/otel-collector-config.yaml`: configuração do OpenTelemetry Collector.
- `service-a/` e `service-b/`: código fonte e Dockerfiles.
- `docker-compose.yaml`: orquestração dos serviços.

---

## Pré-requisitos

- Docker e Docker Compose instalados
- Uma chave de API do WeatherAPI (https://www.weatherapi.com/) para a variável `WEATHER_API_KEY`
- Portas livres localmente (principalmente 9411 para o Zipkin e 8080 para o `service-a`)

---

## Configuração do .env

Crie um arquivo `.env` na raiz do projeto com a sua chave do WeatherAPI:

    WEATHER_API_KEY=coloque_sua_chave_aqui

Observações:
- O `docker-compose.yaml` injeta essa variável no `service-b`.
- As variáveis de OpenTelemetry já estão definidas no Compose:
  - `OTEL_SERVICE_NAME` (ex.: `service-a`, `service-b`)
  - `OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4318` (via HTTP/OTLP)
  - `REQUEST_NAME_OTEL` (nome do span/operation, se aplicável)

---

## Como subir com Docker Compose

Na raiz do projeto:

1) Build e subir os serviços (em segundo plano):

    docker compose up -d --build

2) Ver logs (opcional):

    docker compose logs -f

3) Parar e remover tudo:

    docker compose down

Se quiser forçar rebuild sem cache:

    docker compose build --no-cache

---

## URLs e Portas

- Zipkin (UI): http://localhost:9411
- service-a (API pública): http://localhost:8080
- service-b: porta 8081 (apenas dentro da rede do Docker; não é exposto no host)

Pipeline de observabilidade:
service-a/service-b → otel-collector (OTLP 4318) → Zipkin (9411)

---

## Teste rápido

1) Suba o ambiente (veja seção “Como subir com Docker Compose”).
2) Faça uma requisição ao `service-a` (POST na raiz “/”) informando um CEP válido (8 dígitos). Exemplo:

    curl -s -X POST http://localhost:8080/ \
      -H "Content-Type: application/json" \
      -d '{"cep":"01001000"}'

Resposta esperada (exemplo de formato):

    {
      "city": "São Paulo",
      "temp_C": 28.5,
      "temp_F": 28.5,
      "temp_K": 28.5
    }

3) Abra o Zipkin e visualize os traces:
   - Acesse: http://localhost:9411
   - Procure pelos serviços “service-a” e “service-b”.

Dica: Gere algumas requisições para acumular dados no Zipkin antes de pesquisar.

---

## Endpoints

- service-a:
  - POST `http://localhost:8080/`
    - Body (JSON): { "cep": "12345678" }
    - Validação: CEP deve ter 8 dígitos numéricos
    - Resposta: temperaturas em Celsius, Fahrenheit e Kelvin

- service-b:
  - GET `/temperatura?cep=XXXXXXXX` (não exposto ao host; usado internamente pelo `service-a`)

---

## Observabilidade (OpenTelemetry + Zipkin)

- Os serviços enviam traces para o `otel-collector` em `otel-collector:4318` (OTLP/HTTP).
- O Collector exporta para o Zipkin, que fica disponível em http://localhost:9411.
- Para ver spans por serviço, use o filtro “Service Name” no Zipkin.

Arquivos relevantes:
- `config/otel-collector-config.yaml`: define o receiver OTLP (4318) e exporta para Zipkin.
- Variáveis de ambiente nos serviços controlam nome do serviço (`OTEL_SERVICE_NAME`) e endpoint do exporter (`OTEL_EXPORTER_OTLP_ENDPOINT`).

---

## Solução de problemas

- "cannot find zipcode" (404): CEP inválido ou não encontrado no ViaCEP.
- "error fetching weather data" (500): verifique se `WEATHER_API_KEY` está correto e se há acesso à internet para alcançar o WeatherAPI.
- Zipkin vazio: gere requisições ao `service-a` e verifique logs:
  
      docker compose logs -f service-a
      docker compose logs -f service-b
      docker compose logs -f otel-collector

- Conflito de portas: verifique se as portas 8080 e 9411 já não estão em uso.

---

## Comandos úteis

- Subir:

      docker compose up -d --build

- Logs:

      docker compose logs -f

- Derrubar:

      docker compose down

- Rebuild sem cache:

      docker compose build --no-cache

---

Bom uso!
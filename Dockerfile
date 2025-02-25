# Stage 1: Build the agent
FROM golang:1.21 AS agent-builder

WORKDIR /app

# Копируем go.work и go.work.sum
COPY go.work ./

# Копируем go.mod и go.sum для всех модулей
COPY agent/go.mod ./agent/
COPY orchestrator/go.mod ./orchestrator/
COPY pkg/go.mod ./pkg/

# Скачиваем зависимости для всех модулей
RUN go work sync

# Копируем исходный код для всех модулей
COPY agent ./agent
COPY orchestrator ./orchestrator
COPY pkg ./pkg

# Собираем agent
RUN cd agent && go build -o agent ./cmd

# Stage 2: Build the orchestrator
FROM golang:1.21 AS orchestrator-builder

WORKDIR /app

# Копируем go.work и go.work.sum
COPY go.work ./

# Копируем go.mod и go.sum для всех модулей
COPY agent/go.mod ./agent/
COPY orchestrator/go.mod ./orchestrator/
COPY pkg/go.mod ./pkg/

# Скачиваем зависимости для всех модулей
RUN go work sync

# Копируем исходный код для всех модулей
COPY agent ./agent
COPY orchestrator ./orchestrator
COPY pkg ./pkg

# Собираем orchestrator
RUN cd orchestrator && go build -o orchestrator ./cmd

# Stage 3: Create the final image for agent
FROM debian:bookworm AS agent

WORKDIR /app
COPY --from=agent-builder /app/agent/agent .
CMD ["./agent"]

# Stage 4: Create the final image for orchestrator
FROM debian:bookworm AS orchestrator

WORKDIR /app
COPY --from=orchestrator-builder /app/orchestrator/orchestrator .
CMD ["./orchestrator"]